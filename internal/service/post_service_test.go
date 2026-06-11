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

func setupPostTest(t *testing.T) (*PostService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.Post{}, &model.Node{}, &model.Like{}, &model.NodeModerator{}, &model.Mention{})
	log := zap.NewNop()
	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	nodeModRepo := repository.NewNodeModeratorRepository(db)
	userRepo := repository.NewUserRepository(db)
	mentionRepo := repository.NewMentionRepository(db)
	likeSvc := NewLikeService(likeRepo, log)
	nodeSvc := NewNodeService(nodeRepo, userRepo, mentionRepo, log)
	return NewPostService(postRepo, likeRepo, likeSvc, nodeRepo, nodeModRepo, nodeSvc, db, log), db
}

func seedPostData(t *testing.T, db *gorm.DB) (*model.User, *model.Node) {
	t.Helper()
	u := &model.User{Username: "author", Nickname: "author", Password: "hash", Role: int(auth.RoleUser)}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	n := &model.Node{Name: "tech", Slug: "tech"}
	if err := db.Create(n).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	return u, n
}

func seedUserWithRole(t *testing.T, db *gorm.DB, name string, role auth.Role) *model.User {
	t.Helper()
	u := &model.User{Username: name, Nickname: name, Password: "hash", Role: int(role)}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("seed %s: %v", name, err)
	}
	return u
}

func adminUC(u *model.User) *auth.UserContext {
	return &auth.UserContext{UserID: u.ID, Username: u.Username, Role: auth.RoleAdmin, DeviceID: "test"}
}

func modUC(u *model.User) *auth.UserContext {
	return &auth.UserContext{UserID: u.ID, Username: u.Username, Role: auth.RoleModerator, DeviceID: "test"}
}

func TestPostService_CreatePost(t *testing.T) {
	ctx := context.Background()
	svc, db := setupPostTest(t)
	user, node := seedPostData(t, db)

	t.Run("success", func(t *testing.T) {
		post, err := svc.CreatePost(ctx, userUC(user), node.ID, "Title", "Content")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if post.ID == 0 {
			t.Error("expected post ID")
		}
		if post.Code == "" {
			t.Error("expected code")
		}
		if post.UserID != user.ID {
			t.Errorf("expected userID=%d, got %d", user.ID, post.UserID)
		}
	})

	t.Run("guest rejected", func(t *testing.T) {
		_, err := svc.CreatePost(ctx, guestUC(), node.ID, "T", "C")
		if err == nil {
			t.Fatal("expected error for guest")
		}
	})

	t.Run("node not found", func(t *testing.T) {
		_, err := svc.CreatePost(ctx, userUC(user), 9999, "T", "C")
		if err == nil {
			t.Fatal("expected error for nonexistent node")
		}
	})
}

func TestPostService_DeletePost(t *testing.T) {
	ctx := context.Background()
	svc, db := setupPostTest(t)
	user, node := seedPostData(t, db)
	admin := seedUserWithRole(t, db, "admin1", auth.RoleAdmin)
	mod := seedUserWithRole(t, db, "mod1", auth.RoleModerator)
	other := seedUserWithRole(t, db, "other1", auth.RoleUser)
	node2 := &model.Node{Name: "other", Slug: "other"}
	if err := db.Create(node2).Error; err != nil {
		t.Fatalf("seed node2: %v", err)
	}

	t.Run("author delete success", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		if err := svc.DeletePost(ctx, userUC(user), post.Code); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("admin delete other's post", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		if err := svc.DeletePost(ctx, adminUC(admin), post.Code); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("moderator with jurisdiction", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		db.Create(&model.NodeModerator{NodeID: node.ID, UserID: mod.ID})
		if err := svc.DeletePost(ctx, modUC(mod), post.Code); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("moderator without jurisdiction rejected", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node2.ID, "T", "C")
		err := svc.DeletePost(ctx, modUC(mod), post.Code)
		if err == nil {
			t.Fatal("expected error for mod without jurisdiction")
		}
	})

	t.Run("non-author non-admin rejected", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		err := svc.DeletePost(ctx, userUC(other), post.Code)
		if err == nil {
			t.Fatal("expected error for non-author")
		}
	})

	t.Run("post not found", func(t *testing.T) {
		err := svc.DeletePost(ctx, userUC(user), "noexist")
		if err == nil {
			t.Fatal("expected error for nonexistent post")
		}
	})
}

func TestPostService_ToggleLike(t *testing.T) {
	ctx := context.Background()
	svc, db := setupPostTest(t)
	user, node := seedPostData(t, db)

	t.Run("like success", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		liked, err := svc.ToggleLike(ctx, userUC(user), post.Code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !liked {
			t.Error("expected liked=true")
		}
	})

	t.Run("unlike", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		svc.ToggleLike(ctx, userUC(user), post.Code)
		liked, err := svc.ToggleLike(ctx, userUC(user), post.Code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if liked {
			t.Error("expected liked=false after unlike")
		}
	})

	t.Run("guest rejected", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		_, err := svc.ToggleLike(ctx, guestUC(), post.Code)
		if err == nil {
			t.Fatal("expected error for guest")
		}
	})

	t.Run("post not found", func(t *testing.T) {
		_, err := svc.ToggleLike(ctx, userUC(user), "noexist")
		if err == nil {
			t.Fatal("expected error for nonexistent post")
		}
	})
}

func TestPostService_GetPost(t *testing.T) {
	ctx := context.Background()
	svc, db := setupPostTest(t)
	user, node := seedPostData(t, db)

	t.Run("success with view count", func(t *testing.T) {
		created, _ := svc.CreatePost(ctx, userUC(user), node.ID, "Title", "Content")
		post, _, _, err := svc.GetPost(ctx, userUC(user), created.Code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if post.ViewCount != 1 {
			t.Errorf("expected viewCount=1, got %d", post.ViewCount)
		}
	})

	t.Run("post not found", func(t *testing.T) {
		_, _, _, err := svc.GetPost(ctx, userUC(user), "noexist")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestPostService_ListPosts(t *testing.T) {
	ctx := context.Background()
	svc, db := setupPostTest(t)
	user, node := seedPostData(t, db)
	svc.CreatePost(ctx, userUC(user), node.ID, "T1", "C1")
	svc.CreatePost(ctx, userUC(user), node.ID, "T2", "C2")

	posts, total, likedMap, err := svc.ListPosts(ctx, userUC(user), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}
	if likedMap == nil {
		t.Error("expected non-nil likedMap")
	}
}

func TestPostService_AdminDeletePost(t *testing.T) {
	ctx := context.Background()
	svc, db := setupPostTest(t)
	user, node := seedPostData(t, db)
	node2 := &model.Node{Name: "other2", Slug: "other2"}
	if err := db.Create(node2).Error; err != nil {
		t.Fatalf("seed node2: %v", err)
	}
	admin := seedUserWithRole(t, db, "admin2", auth.RoleAdmin)
	mod := seedUserWithRole(t, db, "mod2", auth.RoleModerator)
	normalUser := seedUserWithRole(t, db, "normal2", auth.RoleUser)

	t.Run("admin success", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		if err := svc.AdminDeletePost(ctx, adminUC(admin), post.Code); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("moderator with jurisdiction", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		db.Create(&model.NodeModerator{NodeID: node.ID, UserID: mod.ID})
		if err := svc.AdminDeletePost(ctx, modUC(mod), post.Code); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("moderator without jurisdiction rejected", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node2.ID, "T", "C")
		err := svc.AdminDeletePost(ctx, modUC(mod), post.Code)
		if err == nil {
			t.Fatal("expected error for mod without jurisdiction")
		}
	})

	t.Run("normal user rejected", func(t *testing.T) {
		post, _ := svc.CreatePost(ctx, userUC(user), node.ID, "T", "C")
		err := svc.AdminDeletePost(ctx, userUC(normalUser), post.Code)
		if err == nil {
			t.Fatal("expected error for normal user")
		}
	})
}
