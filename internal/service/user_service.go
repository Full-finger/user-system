// Package service 实现用户业务逻辑。
package service

import (
	"context"
	"errors"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户业务服务。
type UserService struct {
	repo        repository.UserRepository
	nodeModRepo repository.NodeModeratorRepository
	cfg         *config.JWTConfig
	txRunner    TransactionRunner
	log         *zap.Logger
}

func NewUserService(repo repository.UserRepository, nodeModRepo repository.NodeModeratorRepository, cfg *config.JWTConfig, txRunner TransactionRunner, log *zap.Logger) *UserService {
	return &UserService{repo: repo, nodeModRepo: nodeModRepo, cfg: cfg, txRunner: txRunner, log: log}
}

// Count 返回用户总数。
func (s *UserService) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
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

// FindByID 管理员按 ID 查找用户，需要 Admin 及以上权限。
func (s *UserService) FindByID(ctx context.Context, uc *auth.UserContext, id uint) (*model.User, error) {
	if err := uc.RequireRole(auth.RoleAdmin); err != nil {
		return nil, err
	}
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
	return s.updateUser(ctx, user, UpdateInput{Password: in.Password, Nickname: in.Nickname, CoverTheme: in.CoverTheme, Motto: in.Motto})
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
	// 不允许修改自身角色
	if id == uc.UserID && in.Role != "" {
		return nil, apperror.BadRequest("不能修改自身角色")
	}
	// 不允许将任何人设为 super_admin
	if in.Role != "" && auth.ParseRole(in.Role) == auth.RoleSuperAdmin {
		return nil, apperror.BadRequest("不能设置为超级管理员")
	}
	// 不允许通过 UpdateUser 设置 moderator，需使用 AppointModerator 接口
	if in.Role != "" && auth.ParseRole(in.Role) == auth.RoleModerator {
		return nil, apperror.BadRequest("不能通过此接口设置版主，请使用任命版主接口")
	}
	// admin 不能设 admin，只有 super_admin 能
	if in.Role != "" && auth.ParseRole(in.Role) == auth.RoleAdmin && uc.Role != auth.RoleSuperAdmin {
		return nil, apperror.Forbidden("无权将用户设为管理员")
	}
	// SuperAdmin 不可被 Admin 修改
	if auth.Role(target.Role) == auth.RoleSuperAdmin && uc.Role != auth.RoleSuperAdmin {
		return nil, apperror.Forbidden("无权修改超级管理员")
	}
	// 版主被降级时，在事务中同时清理节点绑定和更新角色
	if in.Role != "" && auth.Role(target.Role) == auth.RoleModerator && auth.ParseRole(in.Role) != auth.RoleModerator {
		return s.updateWithModeratorCleanup(ctx, target, in)
	}
	return s.updateUser(ctx, target, in)
}

// AppointModerator 任命版主：将 user/verified_user 升为 moderator 并绑定节点。
func (s *UserService) AppointModerator(ctx context.Context, uc *auth.UserContext, userID uint, nodeIDs []uint) (*model.User, error) {
	if err := uc.RequireRole(auth.RoleAdmin); err != nil {
		return nil, err
	}
	target, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("用户不存在")
		}
		s.log.Error("查询用户失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	targetRole := auth.Role(target.Role)
	if targetRole != auth.RoleUser && targetRole != auth.RoleVerifiedUser {
		return nil, apperror.BadRequest("只能任命普通用户或认证用户为版主")
	}
	// 事务：设置角色 + 绑定节点
	if err := s.txRunner.RunInTransaction(ctx, func(tx *gorm.DB) error {
		ri := int(auth.RoleModerator)
		if err := tx.Model(&model.User{}).Where("id = ?", target.ID).Update("role", ri).Error; err != nil {
			return err
		}
		mods := make([]repository.NodeModeratorCreate, len(nodeIDs))
		for i, nid := range nodeIDs {
			mods[i] = repository.NodeModeratorCreate{NodeID: nid, UserID: target.ID}
		}
		return tx.Create(&mods).Error
	}); err != nil {
		s.log.Error("任命版主事务失败", zap.Error(err))
		return nil, apperror.Internal("任命版主失败")
	}
	return s.repo.FindByID(ctx, target.ID)
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

// updateWithModeratorCleanup 在事务中同时更新用户信息和清理版主节点绑定。
func (s *UserService) updateWithModeratorCleanup(ctx context.Context, user *model.User, in UpdateInput) (*model.User, error) {
	if err := s.txRunner.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&model.NodeModerator{}).Error; err != nil {
			return err
		}
		upd := repository.UserUpdate{}
		if in.Password != "" {
			hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			p := string(hashed)
			upd.Password = &p
		}
		if in.Nickname != "" {
			upd.Nickname = &in.Nickname
		}
		if in.Role != "" {
			r := auth.ParseRole(in.Role)
			ri := int(r)
			upd.Role = &ri
		}
		cover, err := normalizeCoverTheme(in.CoverTheme)
		if err != nil {
			return err
		}
		upd.CoverTheme = cover
		motto, err := normalizeMotto(in.Motto)
		if err != nil {
			return err
		}
		upd.Motto = motto
		if upd.Nickname != nil || upd.Password != nil || upd.Role != nil || upd.CoverTheme != nil || upd.Motto != nil {
			return tx.Model(&model.User{}).Where("id = ?", user.ID).Updates(upd).Error
		}
		return nil
	}); err != nil {
		s.log.Error("更新用户（含版主清理）事务失败", zap.Error(err))
		return nil, apperror.Internal("更新失败")
	}
	return s.repo.FindByID(ctx, user.ID)
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
		if r == auth.RoleSuperAdmin || r == auth.RoleGuest {
			return nil, apperror.BadRequest("无效的角色值")
		}
		ri := int(r)
		upd.Role = &ri
	}
	cover, err := normalizeCoverTheme(in.CoverTheme)
	if err != nil {
		return nil, err
	}
	upd.CoverTheme = cover
	motto, err := normalizeMotto(in.Motto)
	if err != nil {
		return nil, err
	}
	upd.Motto = motto
	if upd.Nickname == nil && upd.Password == nil && upd.Role == nil && upd.CoverTheme == nil && upd.Motto == nil {
		return user, nil
	}
	if err := s.repo.Update(ctx, user.ID, upd); err != nil {
		s.log.Error("更新用户失败", zap.Error(err))
		return nil, apperror.Internal("更新失败")
	}
	return s.repo.FindByID(ctx, user.ID)
}
