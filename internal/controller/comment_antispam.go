package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

// getDelScript 原子 GET + DEL，防止 nonce 竞态条件。
var getDelScript = redis.NewScript(`
	local val = redis.call("GET", KEYS[1])
	if val then
		redis.call("DEL", KEYS[1])
	end
	return val
`)

// verifyPoW 验证 PoW：SHA256(nonce:timestamp:suffix) 前缀是否满足 difficulty 个 0。
func verifyPoW(nonce string, ts int64, suffix string, difficulty int) bool {
	if suffix == "" {
		return false
	}
	data := fmt.Sprintf("%s:%d:%s", nonce, ts, suffix)
	h := sha256SumHex(data)
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(h, prefix)
}

// sha256SumHex 计算 SHA256 并返回完整 hex 编码。
func sha256SumHex(data string) string {
	h := sha256Sum(data)
	return hex.EncodeToString(h[:])
}

// checkAntispamWithHoneypot 执行反 bot 检查。返回 (isSpam, isBot)。
func (ctrl *CommentController) checkAntispamWithHoneypot(c echo.Context, req *param.CreateCommentRequest, honeypotValue string) (bool, bool) {
	logger := ctrl.log.With(zap.String("ip", c.RealIP()))

	// 1. 蜜罐检查
	if honeypotValue != "" {
		logger.Warn("评论被蜜罐拦截",
			zap.String("honeypot", honeypotValue),
			zap.String("content_preview", truncateRunes(req.Content, 100)),
		)
		return true, true
	}

	// 2. 时间戳检查
	now := time.Now().UnixMilli()
	elapsed := now - req.Ts
	if elapsed < 3000 {
		logger.Warn("评论提交过快",
			zap.Int64("elapsed_ms", elapsed),
			zap.String("content_preview", truncateRunes(req.Content, 100)),
		)
		return true, true
	}
	if elapsed > 24*60*60*1000 {
		logger.Warn("评论时间戳过期",
			zap.Int64("elapsed_ms", elapsed),
			zap.String("content_preview", truncateRunes(req.Content, 100)),
		)
		return true, true
	}

	// 3. PoW Challenge 检查
	if req.Nonce == "" || req.Suffix == "" {
		logger.Warn("评论缺少 challenge 字段",
			zap.String("nonce", req.Nonce),
			zap.String("suffix", req.Suffix),
		)
		return true, true
	}

	// 原子 GET+DEL：防止 nonce 重放和竞态条件
	key := fmt.Sprintf("comment:challenge:%s", req.Nonce)
	val, err := getDelScript.Run(c.Request().Context(), ctrl.rdb, []string{key}).Text()
	if err != nil || val == "" {
		logger.Warn("评论 challenge nonce 无效或已使用",
			zap.String("nonce", req.Nonce),
			zap.Error(err),
		)
		return true, true
	}

	// 比对服务端下发的 timestamp 与客户端提交的 timestamp
	storedTs, parseErr := strconv.ParseInt(val, 10, 64)
	if parseErr != nil || storedTs != req.Ts {
		logger.Warn("评论 challenge timestamp 不匹配",
			zap.String("nonce", req.Nonce),
			zap.Int64("stored_ts", storedTs),
			zap.Int64("client_ts", req.Ts),
		)
		return true, true
	}

	// 验证 PoW
	if !verifyPoW(req.Nonce, req.Ts, req.Suffix, ctrl.antispam.Difficulty) {
		logger.Warn("评论 PoW 验证失败",
			zap.String("nonce", req.Nonce),
			zap.Int("difficulty", ctrl.antispam.Difficulty),
		)
		return true, true
	}

	return false, false
}

// checkContentFilter 内容过滤检查。返回错误信息，空字符串表示通过。
func (ctrl *CommentController) checkContentFilter(c echo.Context, req *param.CreateCommentRequest) string {
	content := req.Content
	uc := auth.GetUserContext(c)

	// 1. 内容长度校验（使用配置值）
	if utf8.RuneCountInString(content) > ctrl.antispam.MaxContentLength {
		return fmt.Sprintf("评论内容不能超过 %d 个字符", ctrl.antispam.MaxContentLength)
	}

	// 2. 关键词过滤（NFKC 标准化 + 去空白 + 大小写不敏感）
	normalized := normalizeForMatch(content)
	for _, kw := range ctrl.antispam.BlockedKeywords {
		normKw := normalizeForMatch(kw)
		if strings.Contains(normalized, normKw) {
			ctrl.log.Warn("评论触发关键词过滤",
				zap.String("keyword", kw),
				zap.String("content_preview", truncateRunes(content, 100)),
			)
			return "评论包含不允许的内容"
		}
	}

	// 3. 重复评论检测
	if ctrl.antispam.DupWindow > 0 && uc.IsAuthenticated() {
		dupKey := fmt.Sprintf("antispam:dup:%d:%s", uc.UserID, contentHash(content))
		set, err := ctrl.rdb.SetNX(c.Request().Context(), dupKey, "1", ctrl.antispam.DupWindow).Result()
		if err == nil && !set {
			ctrl.log.Warn("重复评论检测",
				zap.Uint("user_id", uc.UserID),
				zap.String("content_preview", truncateRunes(content, 100)),
			)
			return "请勿重复提交相同评论"
		}
	}

	// 4. 评论间隔限制
	if ctrl.antispam.CommentInterval > 0 && uc.IsAuthenticated() {
		intervalKey := fmt.Sprintf("antispam:interval:%d", uc.UserID)
		set, err := ctrl.rdb.SetNX(c.Request().Context(), intervalKey, "1", ctrl.antispam.CommentInterval).Result()
		if err == nil && !set {
			ctrl.log.Warn("评论间隔过短",
				zap.Uint("user_id", uc.UserID),
			)
			return "评论过于频繁，请稍后再试"
		}
	}

	return ""
}

// normalizeForMatch 对文本做 NFKC 标准化 + 去空白 + 小写，用于关键词匹配。
func normalizeForMatch(s string) string {
	s = string(norm.NFKC.Bytes([]byte(s)))
	s = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\u3000' || r == '\u00A0' {
			return -1
		}
		return r
	}, s)
	return cases.Lower(language.Und).String(s)
}

// sha256Sum 计算 SHA-256。
func sha256Sum(data string) [32]byte {
	return sha256.Sum256([]byte(data))
}

// contentHash 对评论内容做简单哈希，用于重复检测。
func contentHash(content string) string {
	h := sha256Sum(content)
	return hex.EncodeToString(h[:])[:32]
}

// truncateRunes 按 rune 截断字符串用于日志，避免截断中文字符。
func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n]) + "..."
}
