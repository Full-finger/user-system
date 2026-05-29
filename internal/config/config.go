package config

import (
	"fmt"
	"os"
	"time"

	"github.com/full-finger/user-system/pkg/logger"
	"gopkg.in/yaml.v3"
)

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
	Port string `yaml:"port"`
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

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
