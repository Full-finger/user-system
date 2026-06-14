package service

import (
	"context"
	"testing"

	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupNodeTest(t *testing.T) (*NodeService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.Node{}, &model.Post{}, &model.Mention{}, &model.NodeModerator{})
	log := zap.NewNop()
	nodeRepo := repository.NewNodeRepository(db)
	userRepo := repository.NewUserRepository(db)
	nodeModRepo := repository.NewNodeModeratorRepository(db)
	return NewNodeService(nodeRepo, userRepo, nodeModRepo, log), db
}

func seedNode(t *testing.T, db *gorm.DB) *model.Node {
	t.Helper()
	n := &model.Node{Name: "技术讨论", Slug: "tech", Desc: "技术相关", Color: "#9b8ec4", SortOrder: 1}
	if err := db.Create(n).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	return n
}

// --- ListNodes ---

func TestNodeService_ListNodes(t *testing.T) {
	ctx := context.Background()
	svc, db := setupNodeTest(t)

	t.Run("empty list", func(t *testing.T) {
		nodes, err := svc.ListNodes(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(nodes) != 0 {
			t.Errorf("expected 0 nodes, got %d", len(nodes))
		}
	})

	t.Run("with data", func(t *testing.T) {
		seedNode(t, db)
		nodes, err := svc.ListNodes(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(nodes) != 1 {
			t.Fatalf("expected 1 node, got %d", len(nodes))
		}
		if nodes[0].Slug != "tech" {
			t.Errorf("expected slug=tech, got %s", nodes[0].Slug)
		}
	})
}

// --- GetNode ---

func TestNodeService_GetNode(t *testing.T) {
	ctx := context.Background()
	svc, db := setupNodeTest(t)
	node := seedNode(t, db)

	t.Run("success", func(t *testing.T) {
		found, err := svc.GetNode(ctx, node.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found.Slug != "tech" {
			t.Errorf("expected slug=tech, got %s", found.Slug)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetNode(ctx, 9999)
		if err == nil {
			t.Fatal("expected error for nonexistent node")
		}
	})
}

// --- SeedNodes ---

func TestNodeService_SeedNodes(t *testing.T) {
	ctx := context.Background()
	svc, db := setupNodeTest(t)

	t.Run("first seed creates nodes", func(t *testing.T) {
		svc.SeedNodes(ctx)
		var count int64
		db.Model(&model.Node{}).Count(&count)
		if count == 0 {
			t.Error("expected nodes to be created")
		}
	})

	t.Run("second seed is idempotent", func(t *testing.T) {
		db.Exec("DELETE FROM nodes")
		svc.SeedNodes(ctx)
		var firstCount int64
		db.Model(&model.Node{}).Count(&firstCount)

		svc.SeedNodes(ctx)
		var secondCount int64
		db.Model(&model.Node{}).Count(&secondCount)

		if firstCount != secondCount {
			t.Errorf("expected idempotent seed: first=%d second=%d", firstCount, secondCount)
		}
	})
}
