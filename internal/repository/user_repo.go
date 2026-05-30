// Package repository 提供用户数据访问层抽象。
package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// UserRepository 用户数据访问接口。
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	// Update 按字段 map 部分更新，key 为列名。
	Update(ctx context.Context, id uint, fields map[string]any) error
	Delete(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)
	// FindPage 分页查询，返回当前页数据、总记录数。
	FindPage(ctx context.Context, page, size int) ([]model.User, int64, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
