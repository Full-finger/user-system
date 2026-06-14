package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Mailer 邮件发送接口，由 pkg/email 实现。
type Mailer interface {
	Send(to, subject, body string) error
}

// CaptchaService 邮箱验证码服务，基于 Redis 存储和频率限制。
type CaptchaService struct {
	rdb  *redis.Client
	cfg  *config.CaptchaConfig
	mail Mailer
	log  *zap.Logger
}

func NewCaptchaService(rdb *redis.Client, cfg *config.CaptchaConfig, mail Mailer, log *zap.Logger) *CaptchaService {
	return &CaptchaService{rdb: rdb, cfg: cfg, mail: mail, log: log}
}

func (s *CaptchaService) SendCode(ctx context.Context, email string) error {
	rateKey := fmt.Sprintf("captcha:rate:%s", email)
	ok, err := s.rdb.SetNX(ctx, rateKey, "1", s.cfg.SendInterval).Result()
	if err != nil {
		s.log.Error("设置频率限制失败", zap.Error(err))
		return apperror.Internal("设置频率限制失败")
	}
	if !ok {
		return apperror.TooMany("发送过于频繁，请稍后再试")
	}

	code, err := s.generateCode()
	if err != nil {
		return apperror.Internal("生成验证码失败")
	}

	codeKey := fmt.Sprintf("captcha:code:%s", email)
	if err := s.rdb.Set(ctx, codeKey, code, s.cfg.Expire).Err(); err != nil {
		s.log.Error("存储验证码失败", zap.Error(err))
		return apperror.Internal("存储验证码失败")
	}

	subject := "验证码"
	body := fmt.Sprintf("您的验证码是：%s，有效期 %s。", code, s.cfg.Expire)
	if err := s.mail.Send(email, subject, body); err != nil {
		if err := s.rdb.Del(ctx, codeKey, rateKey).Err(); err != nil {
			s.log.Warn("清理验证码 key 失败", zap.Error(err))
		}
		s.log.Error("发送邮件失败", zap.Error(err))
		return apperror.Internal("发送邮件失败")
	}
	return nil
}

// VerifyCode 校验验证码，正确后立即失效。超过 MaxAttempts 次错误则提前失效。
func (s *CaptchaService) VerifyCode(ctx context.Context, email, code string) error {
	codeKey := fmt.Sprintf("captcha:code:%s", email)
	stored, err := s.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return apperror.BadRequest("验证码已过期")
	}
	if err != nil {
		s.log.Error("验证码校验失败", zap.Error(err))
		return apperror.Internal("验证码校验失败")
	}

	if stored != code {
		if s.cfg.MaxAttempts > 0 {
			attemptsKey := fmt.Sprintf("captcha:attempts:%s", email)
			if err := s.incAttempts(ctx, attemptsKey); err != nil {
				if err := s.rdb.Del(ctx, codeKey, attemptsKey).Err(); err != nil {
					s.log.Warn("清理验证码 key 失败", zap.Error(err))
				}
				return apperror.BadRequest("验证码已失效，请重新获取")
			}
		}
		return apperror.BadRequest("验证码错误")
	}

	delKeys := []string{codeKey}
	if s.cfg.MaxAttempts > 0 {
		delKeys = append(delKeys, fmt.Sprintf("captcha:attempts:%s", email))
	}
	if err := s.rdb.Del(ctx, delKeys...).Err(); err != nil {
		s.log.Warn("清理验证码 key 失败", zap.Error(err))
	}
	return nil
}

// incAttempts 原子递增错误次数，达到上限时返回错误。
func (s *CaptchaService) incAttempts(ctx context.Context, key string) error {
	attempts, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if err := s.rdb.Expire(ctx, key, s.cfg.Expire).Err(); err != nil {
		s.log.Warn("设置 attempts TTL 失败", zap.Error(err))
	}
	if attempts > int64(s.cfg.MaxAttempts) {
		return errors.New("exceeded")
	}
	return nil
}

func (s *CaptchaService) generateCode() (string, error) {
	length := s.cfg.Length
	if length <= 0 {
		length = 6
	}
	var chars string
	switch s.cfg.Type {
	case "alpha":
		chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case "alphanumeric":
		chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	default:
		chars = "0123456789"
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[n.Int64()]
	}
	return string(result), nil
}
