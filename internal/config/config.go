// Package config 负责加载和校验 YAML 配置。
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/full-finger/user-system/pkg/logger"
	"gopkg.in/yaml.v3"
)

// ErrInvalidConfig 配置校验失败时的哨兵错误。
var ErrInvalidConfig = fmt.Errorf("配置校验失败")

// Config 顶层配置结构。
type Config struct {
	Server    ServerConfig     `yaml:"server"`
	Database  DatabaseConfig   `yaml:"database"`
	Redis     RedisConfig      `yaml:"redis"`
	JWT       JWTConfig        `yaml:"jwt"`
	SMTP      SMTPConfig       `yaml:"smtp"`
	Captcha   CaptchaConfig    `yaml:"captcha"`
	RateLimit RateLimitConfig  `yaml:"rate_limit"`
	Log       logger.LogConfig `yaml:"log"`
}

type ServerConfig struct {
	Port        string   `yaml:"port"`
	Env         string   `yaml:"env"`
	CORSOrigins []string `yaml:"cors_origins"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Port     int    `yaml:"port"`
	SSLMode  string `yaml:"sslmode"`
	Timezone string `yaml:"timezone"`
}

// DSN 生成 PostgreSQL 连接字符串。
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		d.Host, d.User, d.Password, d.DBName, d.Port, d.SSLMode, d.Timezone)
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret string        `yaml:"secret"`
	Expire time.Duration `yaml:"expire"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	TLS      bool   `yaml:"tls"`
	Auth     bool   `yaml:"auth"`
}

type CaptchaConfig struct {
	Length       int           `yaml:"length"`
	Type         string        `yaml:"type"`
	Expire       time.Duration `yaml:"expire"`
	SendInterval time.Duration `yaml:"send_interval"`
	MaxAttempts  int           `yaml:"max_attempts"`
}

type RateLimitConfig struct {
	Window     time.Duration `yaml:"window"`
	MaxRequest int           `yaml:"max_requests"`
}

// Load 从 YAML 文件加载配置并校验。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate 校验关键配置项的合法性。
func (c *Config) Validate() error {
	if len(c.JWT.Secret) < 16 {
		return fmt.Errorf("%w: JWT secret 长度不能少于 16", ErrInvalidConfig)
	}
	if c.JWT.Expire <= 0 {
		return fmt.Errorf("%w: JWT expire 必须大于 0", ErrInvalidConfig)
	}
	if c.Captcha.Expire <= 0 {
		return fmt.Errorf("%w: captcha expire 必须大于 0", ErrInvalidConfig)
	}
	if c.RateLimit.Window <= 0 {
		return fmt.Errorf("%w: rate_limit window 必须大于 0", ErrInvalidConfig)
	}
	if c.RateLimit.MaxRequest <= 0 {
		return fmt.Errorf("%w: rate_limit max_requests 必须大于 0", ErrInvalidConfig)
	}
	return nil
}
