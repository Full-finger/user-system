package repository

import (
	"context"
	"strings"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/gorm"
)

type userRepoGorm struct {
	db *gorm.DB
}

// NewUserRepository 创建基于 GORM 的 UserRepository 实现。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepoGorm{db: db}
}

func (r *userRepoGorm) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepoGorm) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepoGorm) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepoGorm) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepoGorm) Update(ctx context.Context, id uint, upd UserUpdate) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(upd).Error
}

func (r *userRepoGorm) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *userRepoGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepoGorm) FindPage(ctx context.Context, page, size int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("id desc").Offset(offset).Limit(size).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepoGorm) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepoGorm) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepoGorm) ExistsByRole(ctx context.Context, role int) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("role = ?", role).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepoGorm) FindByIDs(ctx context.Context, ids []uint) ([]model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []model.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepoGorm) FindByRoleGTE(ctx context.Context, minRole int) ([]model.User, error) {
	var users []model.User
	if err := r.db.WithContext(ctx).Where("role >= ?", minRole).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepoGorm) FindByUsernames(ctx context.Context, usernames []string) ([]model.User, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	// 使用 LOWER() 实现大小写不敏感匹配，兼容 PostgreSQL（默认大小写敏感）。
	lower := make([]string, len(usernames))
	for i, u := range usernames {
		lower[i] = strings.ToLower(u)
	}
	var users []model.User
	if err := r.db.WithContext(ctx).Where("LOWER(username) IN ?", lower).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
