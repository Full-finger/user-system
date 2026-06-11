package repository

import (
	"context"
	"testing"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPostRepoTest(t *testing.T) PostRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.Node{}, &model.Post{})

	// 创建测试用户
	db.Create(&model.User{Username: "author", Nickname: "Author", Password: "hashed"})
	// 创建测试节点
	db.Create(&model.Node{Name: "test", Slug: "test", Desc: "test node"})

	return NewPostRepository(db)
}

func createTestPost(t *testing.T, repo PostRepository, ctx context.Context, code string) *model.Post {
	t.Helper()
	post := &model.Post{
		Code:    code,
		UserID:  1,
		NodeID:  1,
		Title:   "Test Post " + code,
		Content: "Content for " + code,
	}
	err := repo.Create(ctx, post)
	if err != nil {
		t.Fatalf("failed to create test post: %v", err)
	}
	return post
}

func TestPostRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)

	t.Run("success", func(t *testing.T) {
		post := createTestPost(t, repo, ctx, "abc001")
		if post.ID == 0 {
			t.Error("expected ID to be set after create")
		}
		if post.Code != "abc001" {
			t.Errorf("expected code abc001, got %s", post.Code)
		}
	})
}

func TestPostRepository_FindByCode(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	_ = createTestPost(t, repo, ctx, "abc002")

	t.Run("found", func(t *testing.T) {
		post, err := repo.FindByCode(ctx, "abc002")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if post.Title != "Test Post abc002" {
			t.Errorf("expected 'Test Post abc002', got %s", post.Title)
		}
		// 验证 Preload
		if post.User.Username != "author" {
			t.Errorf("expected preloaded user 'author', got %s", post.User.Username)
		}
		if post.Node.Name != "test" {
			t.Errorf("expected preloaded node 'test', got %s", post.Node.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByCode(ctx, "notexist")
		if err == nil {
			t.Fatal("expected error for not found")
		}
	})
}

func TestPostRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	post := createTestPost(t, repo, ctx, "abc003")

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, post.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Code != "abc003" {
			t.Errorf("expected abc003, got %s", got.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 999)
		if err == nil {
			t.Fatal("expected error for not found")
		}
	})
}

func TestPostRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	post := createTestPost(t, repo, ctx, "abc004")

	err := repo.Delete(ctx, post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.FindByID(ctx, post.ID)
	if err == nil {
		t.Fatal("expected post to be deleted")
	}
}

func TestPostRepository_IncrLikeCount(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	post := createTestPost(t, repo, ctx, "abc005")

	err := repo.IncrLikeCount(ctx, post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := repo.FindByID(ctx, post.ID)
	if updated.LikeCount != 1 {
		t.Errorf("expected like_count 1, got %d", updated.LikeCount)
	}
}

func TestPostRepository_DecrLikeCount(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	post := createTestPost(t, repo, ctx, "abc006")

	// 先加再减
	_ = repo.IncrLikeCount(ctx, post.ID)
	err := repo.DecrLikeCount(ctx, post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := repo.FindByID(ctx, post.ID)
	if updated.LikeCount != 0 {
		t.Errorf("expected like_count 0, got %d", updated.LikeCount)
	}

	// 从 0 再减不应变为负数（SQL 条件 like_count > 0）
	_ = repo.DecrLikeCount(ctx, post.ID)
	updated2, _ := repo.FindByID(ctx, post.ID)
	if updated2.LikeCount != 0 {
		t.Errorf("expected like_count to stay 0, got %d", updated2.LikeCount)
	}
}

func TestPostRepository_IncrViewCount(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	post := createTestPost(t, repo, ctx, "abc007")

	err := repo.IncrViewCount(ctx, post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := repo.FindByID(ctx, post.ID)
	if updated.ViewCount != 1 {
		t.Errorf("expected view_count 1, got %d", updated.ViewCount)
	}
}

func TestPostRepository_FindByUserID(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	_ = createTestPost(t, repo, ctx, "uid001")
	_ = createTestPost(t, repo, ctx, "uid002")

	t.Run("found", func(t *testing.T) {
		posts, total, err := repo.FindByUserID(ctx, 1, 1, 10)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(posts) != 2 {
			t.Errorf("expected 2 posts, got %d", len(posts))
		}
	})

	t.Run("no posts for user", func(t *testing.T) {
		posts, total, err := repo.FindByUserID(ctx, 999, 1, 10)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(posts) != 0 {
			t.Errorf("expected 0 posts, got %d", len(posts))
		}
	})
}

func TestPostRepository_FindPage(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	_ = createTestPost(t, repo, ctx, "pg001")
	_ = createTestPost(t, repo, ctx, "pg002")
	_ = createTestPost(t, repo, ctx, "pg003")

	t.Run("page 1 size 2", func(t *testing.T) {
		posts, total, err := repo.FindPage(ctx, 1, 2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(posts) != 2 {
			t.Errorf("expected 2 posts, got %d", len(posts))
		}
	})
}

func TestPostRepository_FindByNodeID(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	_ = createTestPost(t, repo, ctx, "nid001")
	_ = createTestPost(t, repo, ctx, "nid002")

	t.Run("found", func(t *testing.T) {
		posts, total, err := repo.FindByNodeID(ctx, 1, 1, 10, "time")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(posts) != 2 {
			t.Errorf("expected 2 posts, got %d", len(posts))
		}
	})

	t.Run("sort by replies", func(t *testing.T) {
		posts, _, err := repo.FindByNodeID(ctx, 1, 1, 10, "replies")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(posts) != 2 {
			t.Errorf("expected 2 posts, got %d", len(posts))
		}
	})
}

func TestPostRepository_CountByUserID(t *testing.T) {
	ctx := context.Background()
	repo := setupPostRepoTest(t)
	_ = createTestPost(t, repo, ctx, "cnt001")
	_ = createTestPost(t, repo, ctx, "cnt002")

	t.Run("count", func(t *testing.T) {
		count, err := repo.CountByUserID(ctx, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2, got %d", count)
		}
	})

	t.Run("no posts", func(t *testing.T) {
		count, err := repo.CountByUserID(ctx, 999)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	})
}
