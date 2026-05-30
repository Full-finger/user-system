// Package service 实现用户业务逻辑。
package service

import (
	"context"
	"errors"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户业务服务。
type UserService struct {
	repo repository.UserRepository
	cfg  *config.JWTConfig
	log  *zap.Logger
}

func NewUserService(repo repository.UserRepository, cfg *config.JWTConfig, log *zap.Logger) *UserService {
	return &UserService{repo: repo, cfg: cfg, log: log}
}

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

	user := &model.User{
		Username: in.Username,
		Password: string(hashed),
		Role:     "user",
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
	// 先按用户名查，找不到则 fallback 到邮箱
	user, err := s.repo.FindByUsername(ctx, in.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user, err = s.repo.FindByEmail(ctx, in.Username)
		}
		if err != nil {
			// 用户不存在时执行一次假 bcrypt，防止通过响应时间枚举用户名
			bcrypt.CompareHashAndPassword([]byte("$2a$10$fakehashfakehashfakehashfakeha"), []byte(in.Password))
			return "", apperror.Unauthorized("用户名或密码错误")
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

func (s *UserService) BindEmail(ctx context.Context, userID uint, email string) error {
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		s.log.Error("查询邮箱失败", zap.Error(err))
		return apperror.Internal("绑定邮箱失败")
	}
	if exists {
		return apperror.BadRequest("该邮箱已被绑定")
	}
	if err := s.repo.Update(ctx, userID, map[string]any{"email": email}); err != nil {
		s.log.Error("绑定邮箱失败", zap.Error(err))
		return apperror.Internal("绑定邮箱失败")
	}
	return nil
}

func (s *UserService) GetProfile(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("用户不存在")
		}
		s.log.Error("查询用户失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, in ProfileUpdateInput) (*model.User, error) {
	user, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.updateUser(ctx, user, UpdateInput{Password: in.Password})
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	users, total, err := s.repo.FindPage(ctx, page, pageSize)
	if err != nil {
		s.log.Error("查询用户列表失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return users, total, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uint, in UpdateInput) (*model.User, error) {
	user, err := s.GetProfile(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.updateUser(ctx, user, in)
}

func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("删除用户失败", zap.Error(err))
		return apperror.Internal("删除失败")
	}
	return nil
}

func (s *UserService) updateUser(ctx context.Context, user *model.User, in UpdateInput) (*model.User, error) {
	updates := map[string]any{}
	if in.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, apperror.Internal("密码加密失败")
		}
		updates["password"] = string(hashed)
	}
	if in.Role != "" {
		if in.Role != "admin" && in.Role != "user" {
			return nil, apperror.BadRequest("无效的角色")
		}
		updates["role"] = in.Role
	}
	if len(updates) == 0 {
		return user, nil
	}
	if err := s.repo.Update(ctx, user.ID, updates); err != nil {
		s.log.Error("更新用户失败", zap.Error(err))
		return nil, apperror.Internal("更新失败")
	}
	return s.GetProfile(ctx, user.ID)
}

// generateToken 签发 HS256 JWT，payload 包含 user_id、username、role。
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
