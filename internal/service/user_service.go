// Package service 实现用户业务逻辑。
package service

import (
	"context"
	"errors"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user, err = s.repo.FindByEmail(ctx, in.Username)
		}
		if err != nil {
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

func (s *UserService) BindEmail(ctx context.Context, uc *auth.UserContext, email string) error {
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return err
	}
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		s.log.Error("查询邮箱失败", zap.Error(err))
		return apperror.Internal("绑定邮箱失败")
	}
	if exists {
		return apperror.BadRequest("该邮箱已被绑定")
	}
	if err := s.repo.Update(ctx, uc.UserID, repository.UserUpdate{Email: &email}); err != nil {
		s.log.Error("绑定邮箱失败", zap.Error(err))
		return apperror.Internal("绑定邮箱失败")
	}
	return nil
}

// FindByID 按 ID 查找用户，无权限检查，供管理员查看他人资料使用。
func (s *UserService) FindByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("用户不存在")
		}
		s.log.Error("查询用户失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return user, nil
}

func (s *UserService) GetProfile(ctx context.Context, uc *auth.UserContext) (*model.User, error) {
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return nil, err
	}
	user, err := s.repo.FindByID(ctx, uc.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("用户不存在")
		}
		s.log.Error("查询用户失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, uc *auth.UserContext, in ProfileUpdateInput) (*model.User, error) {
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return nil, err
	}
	user, err := s.repo.FindByID(ctx, uc.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("用户不存在")
		}
		s.log.Error("查询用户失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return s.updateUser(ctx, user, UpdateInput{Password: in.Password, Nickname: in.Nickname})
}

func (s *UserService) ListUsers(ctx context.Context, uc *auth.UserContext, page, pageSize int) ([]model.User, int64, error) {
	if err := uc.RequireRole(auth.RoleAdmin); err != nil {
		return nil, 0, err
	}
	users, total, err := s.repo.FindPage(ctx, page, pageSize)
	if err != nil {
		s.log.Error("查询用户列表失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return users, total, nil
}

func (s *UserService) UpdateUser(ctx context.Context, uc *auth.UserContext, id uint, in UpdateInput) (*model.User, error) {
	if err := uc.RequireRole(auth.RoleAdmin); err != nil {
		return nil, err
	}
	target, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("用户不存在")
		}
		s.log.Error("查询用户失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	// SuperAdmin 不可被 Admin 修改
	if auth.Role(target.Role) == auth.RoleSuperAdmin && uc.Role != auth.RoleSuperAdmin {
		return nil, apperror.Forbidden("无权修改超级管理员")
	}
	return s.updateUser(ctx, target, in)
}

func (s *UserService) DeleteUser(ctx context.Context, uc *auth.UserContext, id uint) error {
	if err := uc.RequireRole(auth.RoleAdmin); err != nil {
		return err
	}
	target, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("用户不存在")
		}
		s.log.Error("查询用户失败", zap.Error(err))
		return apperror.Internal("查询失败")
	}
	if auth.Role(target.Role) == auth.RoleSuperAdmin && uc.Role != auth.RoleSuperAdmin {
		return apperror.Forbidden("无权删除超级管理员")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("删除用户失败", zap.Error(err))
		return apperror.Internal("删除失败")
	}
	return nil
}

func (s *UserService) updateUser(ctx context.Context, user *model.User, in UpdateInput) (*model.User, error) {
	upd := repository.UserUpdate{}
	if in.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, apperror.Internal("密码加密失败")
		}
		p := string(hashed)
		upd.Password = &p
	}
	if in.Nickname != "" {
		upd.Nickname = &in.Nickname
	}
	if in.Role != "" {
		r := auth.ParseRole(in.Role)
		if r != auth.RoleUser && r != auth.RoleModerator {
			return nil, apperror.BadRequest("只能设置为 user 或 moderator")
		}
		ri := int(r)
		upd.Role = &ri
	}
	if upd.Email == nil && upd.Nickname == nil && upd.Password == nil && upd.Role == nil {
		return user, nil
	}
	if err := s.repo.Update(ctx, user.ID, upd); err != nil {
		s.log.Error("更新用户失败", zap.Error(err))
		return nil, apperror.Internal("更新失败")
	}
	return s.repo.FindByID(ctx, user.ID)
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
