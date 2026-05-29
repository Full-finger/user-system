package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Mailer interface {
	Send(to, subject, body string) error
}

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
		s.rdb.Del(ctx, codeKey, rateKey)
		s.log.Error("发送邮件失败", zap.Error(err))
		return apperror.Internal("发送邮件失败")
	}
	return nil
}

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
			attempts, _ := s.rdb.Get(ctx, attemptsKey).Int()
			if attempts >= s.cfg.MaxAttempts-1 {
				s.rdb.Del(ctx, codeKey, attemptsKey)
				return apperror.BadRequest("验证码已失效，请重新获取")
			}
			s.rdb.Incr(ctx, attemptsKey)
			s.rdb.Expire(ctx, attemptsKey, s.cfg.Expire)
		}
		return apperror.BadRequest("验证码错误")
	}

	delKeys := []string{codeKey}
	if s.cfg.MaxAttempts > 0 {
		delKeys = append(delKeys, fmt.Sprintf("captcha:attempts:%s", email))
	}
	s.rdb.Del(ctx, delKeys...)
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
