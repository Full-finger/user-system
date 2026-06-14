package service

import (
	"context"
	"errors"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *UserService) CheckUsername(ctx context.Context, username string) error {
	exists, err := s.repo.ExistsByUsername(ctx, username)
	if err != nil {
		s.log.Error("查询用户名失败", zap.Error(err))
		return apperror.Internal("查询失败")
	}
	if exists {
		return apperror.BadRequest("用户名已被占用")
	}
	return nil
}

func (s *UserService) Register(ctx context.Context, in RegisterInput) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal("密码加密失败")
	}

	nickname := in.Nickname
	if nickname == "" {
		nickname = in.Username
	}

	user := &model.User{
		Username: in.Username,
		Nickname: nickname,
		Password: string(hashed),
		Role:     int(auth.RoleUser),
	}
	if in.Email != "" {
		user.Email = &in.Email
	}
	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperror.BadRequest("用户名或邮箱已存在")
		}
		s.log.Error("注册失败", zap.Error(err))
		return nil, apperror.Internal("注册失败")
	}
	return user, nil
}

// Login 支持用户名或邮箱登录，返回 JWT token。
func (s *UserService) Login(ctx context.Context, in LoginInput) (string, error) {
	user, err := s.repo.FindByUsername(ctx, in.Username)
	if err != nil {
		// 用户名查不到时回退到邮箱
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user, err = s.repo.FindByEmail(ctx, in.Username)
		}
		if err != nil {
			// 区分「用户不存在」与「数据库故障」
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 用户不存在：用假 hash 比较消耗时间，防时序攻击
				bcrypt.CompareHashAndPassword([]byte("$2a$10$fakehashfakehashfakehashfakeha"), []byte(in.Password))
				return "", apperror.Unauthorized("用户名或密码错误")
			}
			// 真正的数据库错误：记录日志并返回内部错误
			s.log.Error("登录查询用户失败", zap.Error(err))
			return "", apperror.Internal("登录失败")
		}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return "", apperror.Unauthorized("用户名或密码错误")
	}
	return s.generateToken(user)
}

func (s *UserService) LoginByEmail(ctx context.Context, email string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", apperror.Unauthorized("验证码无效或用户不存在")
	}
	return s.generateToken(user)
}

// SeedAdmin 当数据库中不存在任何 super_admin 用户时，使用配置信息创建一个。
func (s *UserService) SeedAdmin(ctx context.Context, cfg *config.AdminConfig) {
	if cfg == nil || cfg.Username == "" || cfg.Password == "" {
		return
	}
	exists, err := s.repo.ExistsByRole(ctx, int(auth.RoleSuperAdmin))
	if err != nil {
		s.log.Error("检查 super_admin 用户失败", zap.Error(err))
		return
	}
	if exists {
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("admin 密码加密失败", zap.Error(err))
		return
	}
	user := &model.User{
		Username: cfg.Username,
		Nickname: cfg.Username,
		Password: string(hashed),
		Role:     int(auth.RoleSuperAdmin),
	}
	if cfg.Email != "" {
		user.Email = &cfg.Email
	}
	if err := s.repo.Create(ctx, user); err != nil {
		s.log.Error("创建 super_admin 用户失败", zap.Error(err))
		return
	}
	s.log.Info("已自动创建 super_admin 用户", zap.String("username", cfg.Username))
}

// generateToken 签发 HS256 JWT，payload 包含 user_id、username、role(int)。
func (s *UserService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      jwt.NewNumericDate(time.Now().Add(s.cfg.Expire)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}
