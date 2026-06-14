package service

import (
	"context"

	"gorm.io/gorm"
)

// TransactionRunner 事务执行抽象，解除 Service 对 *gorm.DB 的直接依赖。
type TransactionRunner interface {
	RunInTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// GormTransactionRunner 基于 GORM 的事务执行器。
type GormTransactionRunner struct {
	db *gorm.DB
}

// NewGormTransactionRunner 创建 GORM 事务执行器。
func NewGormTransactionRunner(db *gorm.DB) *GormTransactionRunner {
	return &GormTransactionRunner{db: db}
}

// RunInTransaction 在事务中执行 fn，自动提交或回滚。
func (r *GormTransactionRunner) RunInTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// Compile-time check: GormTransactionRunner satisfies TransactionRunner.
var _ TransactionRunner = (*GormTransactionRunner)(nil)
