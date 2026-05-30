package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/full-finger/user-system/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type mockMailer struct {
	lastTo      string
	lastSubject string
	lastBody    string
	err         error
}

func (m *mockMailer) Send(to, subject, body string) error {
	m.lastTo = to
	m.lastSubject = subject
	m.lastBody = body
	return m.err
}

func setupCaptchaTest(t *testing.T, cfg *config.CaptchaConfig) (*CaptchaService, *miniredis.Miniredis, *mockMailer) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mailer := &mockMailer{}
	svc := NewCaptchaService(rdb, cfg, mailer, zap.NewNop())
	return svc, mr, mailer
}

func TestCaptchaService_SendCode(t *testing.T) {
	ctx := context.Background()
	cfg := &config.CaptchaConfig{
		Length:       6,
		Type:         "number",
		Expire:       5 * time.Minute,
		SendInterval: time.Minute,
		MaxAttempts:  0,
	}

	t.Run("success", func(t *testing.T) {
		svc, mr, mailer := setupCaptchaTest(t, cfg)
		err := svc.SendCode(ctx, "test@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if mailer.lastTo != "test@example.com" {
			t.Errorf("expected mail to test@example.com, got %s", mailer.lastTo)
		}
		val, err := mr.Get("captcha:code:test@example.com")
		if err != nil || val == "" {
			t.Error("expected code to be stored in redis")
		}
	})

	t.Run("rate limited", func(t *testing.T) {
		svc, _, _ := setupCaptchaTest(t, cfg)
		_ = svc.SendCode(ctx, "test@example.com")
		err := svc.SendCode(ctx, "test@example.com")
		if err == nil {
			t.Fatal("expected rate limit error")
		}
	})

	t.Run("email failure cleans up", func(t *testing.T) {
		svc, mr, mailer := setupCaptchaTest(t, cfg)
		mailer.err = fmt.Errorf("smtp error")
		err := svc.SendCode(ctx, "test@example.com")
		if err == nil {
			t.Fatal("expected error")
		}
		if _, e := mr.Get("captcha:code:test@example.com"); e == nil {
			t.Error("expected code to be cleaned up after email failure")
		}
	})
}

func TestCaptchaService_VerifyCode(t *testing.T) {
	ctx := context.Background()
	cfg := &config.CaptchaConfig{
		Length:       6,
		Type:         "number",
		Expire:       5 * time.Minute,
		SendInterval: time.Minute,
		MaxAttempts:  3,
	}

	t.Run("success", func(t *testing.T) {
		svc, mr, _ := setupCaptchaTest(t, cfg)
		mr.Set("captcha:code:test@example.com", "123456")
		err := svc.VerifyCode(ctx, "test@example.com", "123456")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, e := mr.Get("captcha:code:test@example.com"); e == nil {
			t.Error("expected code to be deleted after verification")
		}
	})

	t.Run("expired", func(t *testing.T) {
		svc, _, _ := setupCaptchaTest(t, cfg)
		err := svc.VerifyCode(ctx, "test@example.com", "123456")
		if err == nil {
			t.Fatal("expected error for expired code")
		}
	})

	t.Run("wrong code increments attempts", func(t *testing.T) {
		svc, mr, _ := setupCaptchaTest(t, cfg)
		mr.Set("captcha:code:test@example.com", "123456")
		err := svc.VerifyCode(ctx, "test@example.com", "654321")
		if err == nil {
			t.Fatal("expected error for wrong code")
		}
		val, _ := mr.Get("captcha:attempts:test@example.com")
		if val != "1" {
			t.Errorf("expected 1 attempt, got %s", val)
		}
	})

	t.Run("max attempts exceeded invalidates code", func(t *testing.T) {
		svc, mr, _ := setupCaptchaTest(t, cfg)
		mr.Set("captcha:code:test@example.com", "123456")
		mr.Set("captcha:attempts:test@example.com", "3")
		err := svc.VerifyCode(ctx, "test@example.com", "wrong")
		if err == nil {
			t.Fatal("expected error")
		}
		if _, e := mr.Get("captcha:code:test@example.com"); e == nil {
			t.Error("expected code to be cleaned up after max attempts")
		}
	})

	t.Run("no limit when max_attempts is 0", func(t *testing.T) {
		noLimitCfg := &config.CaptchaConfig{
			Length:       6,
			Type:         "number",
			Expire:       5 * time.Minute,
			SendInterval: time.Minute,
			MaxAttempts:  0,
		}
		svc, mr, _ := setupCaptchaTest(t, noLimitCfg)
		mr.Set("captcha:code:test@example.com", "123456")
		for i := 0; i < 10; i++ {
			_ = svc.VerifyCode(ctx, "test@example.com", "wrong")
		}
		if _, e := mr.Get("captcha:code:test@example.com"); e != nil {
			t.Error("expected code to still exist when no limit")
		}
	})
}

func TestCaptchaService_generateCode(t *testing.T) {
	cfg := &config.CaptchaConfig{
		Length:       6,
		Type:         "number",
		Expire:       5 * time.Minute,
		SendInterval: time.Minute,
	}
	svc, _, _ := setupCaptchaTest(t, cfg)

	t.Run("number type", func(t *testing.T) {
		code, err := svc.generateCode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 6 {
			t.Errorf("expected 6 chars, got %d", len(code))
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Errorf("expected digit, got %c", c)
			}
		}
	})

	t.Run("alpha type", func(t *testing.T) {
		alphaCfg := &config.CaptchaConfig{Length: 8, Type: "alpha"}
		svc, _, _ := setupCaptchaTest(t, alphaCfg)
		code, err := svc.generateCode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 8 {
			t.Errorf("expected 8 chars, got %d", len(code))
		}
	})

	t.Run("alphanumeric type", func(t *testing.T) {
		anCfg := &config.CaptchaConfig{Length: 10, Type: "alphanumeric"}
		svc, _, _ := setupCaptchaTest(t, anCfg)
		code, err := svc.generateCode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 10 {
			t.Errorf("expected 10 chars, got %d", len(code))
		}
	})

	t.Run("default length when 0", func(t *testing.T) {
		zeroCfg := &config.CaptchaConfig{Length: 0, Type: "number"}
		svc, _, _ := setupCaptchaTest(t, zeroCfg)
		code, err := svc.generateCode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 6 {
			t.Errorf("expected default 6 chars, got %d", len(code))
		}
	})
}
