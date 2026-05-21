package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/full-finger/user-system/internal/config"
	customPkg "github.com/full-finger/user-system/pkg"
	"github.com/redis/go-redis/v9"
)

type CaptchaService struct {
	rdb  *redis.Client
	cfg  *config.CaptchaConfig
	mail *customPkg.EmailSender
}

func NewCaptchaService(rdb *redis.Client, cfg *config.CaptchaConfig, mail *customPkg.EmailSender) *CaptchaService {
	return &CaptchaService{rdb: rdb, cfg: cfg, mail: mail}
}

func (s *CaptchaService) SendCode(email string) error {
	ctx := context.Background()

	// 检查发送频率
	rateKey := fmt.Sprintf("captcha:rate:%s", email)
	if s.rdb.Exists(ctx, rateKey).Val() > 0 {
		return errors.New("发送过于频繁，请稍后再试")
	}

	// 生成验证码
	code := s.generateCode()

	// 存储验证码
	codeKey := fmt.Sprintf("captcha:code:%s", email)
	if err := s.rdb.Set(ctx, codeKey, code, s.cfg.Expire).Err(); err != nil {
		return fmt.Errorf("存储验证码失败")
	}

	// 设置频率限制
	if err := s.rdb.Set(ctx, rateKey, "1", s.cfg.SendInterval).Err(); err != nil {
		return fmt.Errorf("设置频率限制失败")
	}

	// 发送邮件
	subject := "验证码"
	body := fmt.Sprintf("您的验证码是：%s，有效期 %s。", code, s.cfg.Expire)
	if err := s.mail.Send(email, subject, body); err != nil {
		// 发送失败，删除已存储的验证码和频率限制
		s.rdb.Del(ctx, codeKey, rateKey)
		return fmt.Errorf("发送邮件失败")
	}

	return nil
}

func (s *CaptchaService) VerifyCode(email, code string) error {
	ctx := context.Background()

	codeKey := fmt.Sprintf("captcha:code:%s", email)
	stored, err := s.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return errors.New("验证码已过期")
	}
	if err != nil {
		return fmt.Errorf("验证码校验失败")
	}

	if stored != code {
		return errors.New("验证码错误")
	}

	// 验证成功，删除验证码防止重复使用
	s.rdb.Del(ctx, codeKey)
	return nil
}

func (s *CaptchaService) generateCode() string {
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
	default: // number
		chars = "0123456789"
	}

	result := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}
