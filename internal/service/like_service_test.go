package service

import (
	"context"
	"testing"

	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLikeTest(t *testing.T) (*LikeService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.Node{}, &model.Post{}, &model.Like{})
	likeRepo := repository.NewLikeRepository(db)
	log := zap.NewNop()
	return NewLikeService(likeRepo, log), db
}

func TestLikeService_FindLikedPostIDs(t *testing.T) {
	ctx := context.Background()
	svc, db := setupLikeTest(t)

	// 种子数据：用户 + 节点 + 帖子
	user := &model.User{Username: "alice", Nickname: "alice", Password: "hash", Role: int(auth.RoleUser)}
	db.Create(user)
	node := &model.Node{Slug: "test", Name: "test"}
	db.Create(node)
	p1 := &model.Post{Code: "a1", UserID: user.ID, NodeID: node.ID, Title: "p1", Content: "c1"}
	p2 := &model.Post{Code: "b2", UserID: user.ID, NodeID: node.ID, Title: "p2", Content: "c2"}
	p3 := &model.Post{Code: "c3", UserID: user.ID, NodeID: node.ID, Title: "p3", Content: "c3"}
	db.Create(p1)
	db.Create(p2)
	db.Create(p3)

	uc := &auth.UserContext{UserID: user.ID, Username: user.Username, Role: auth.RoleUser}

	t.Run("guest returns empty", func(t *testing.T) {
		m, err := svc.FindLikedPostIDs(ctx, guestUC(), []uint{p1.ID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("expected empty, got %v", m)
		}
	})

	t.Run("empty post ids", func(t *testing.T) {
		m, err := svc.FindLikedPostIDs(ctx, uc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("expected empty, got %v", m)
		}
	})

	t.Run("no likes", func(t *testing.T) {
		m, err := svc.FindLikedPostIDs(ctx, uc, []uint{p1.ID, p2.ID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("expected empty, got %v", m)
		}
	})

	t.Run("mixed liked and unliked", func(t *testing.T) {
		// 点赞 p1 和 p3
		db.Create(&model.Like{UserID: user.ID, PostID: p1.ID})
		db.Create(&model.Like{UserID: user.ID, PostID: p3.ID})

		m, err := svc.FindLikedPostIDs(ctx, uc, []uint{p1.ID, p2.ID, p3.ID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !m[p1.ID] {
			t.Error("expected p1 liked")
		}
		if m[p2.ID] {
			t.Error("expected p2 not liked")
		}
		if !m[p3.ID] {
			t.Error("expected p3 liked")
		}
	})
}
