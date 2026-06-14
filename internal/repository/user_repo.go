// Package repository 提供用户数据访问层抽象。
package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// UserUpdate 用户更新字段，nil 表示不更新。
type UserUpdate struct {
	Email      *string
	Nickname   *string
	Password   *string // 调用方需传入已哈希的密码
	Role       *int
	CoverTheme *string
	Motto      *string // nil 不更新；"" 清空；非空为座右铭内容
}

// UserRepository 用户数据访问接口。
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	// Update 按非 nil 字段部分更新。
	Update(ctx context.Context, id uint, upd UserUpdate) error
	Delete(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)
	// FindPage 分页查询，返回当前页数据、总记录数。
	FindPage(ctx context.Context, page, size int) ([]model.User, int64, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByRole(ctx context.Context, role int) (bool, error)
	// FindByIDs 按 ID 列表批量查询用户。
	FindByIDs(ctx context.Context, ids []uint) ([]model.User, error)
	// FindByRoleGTE 查询角色 >= minRole 的所有用户。
	FindByRoleGTE(ctx context.Context, minRole int) ([]model.User, error)
	// FindByUsernames 按用户名列表批量查询用户。
	FindByUsernames(ctx context.Context, usernames []string) ([]model.User, error)
}
