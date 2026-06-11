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

func setupFollowTest(t *testing.T) (*FollowService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.Follow{}, &model.Node{}, &model.Post{})
	followRepo := repository.NewFollowRepository(db)
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	log := zap.NewNop()
	return NewFollowService(followRepo, userRepo, postRepo, log), db
}

func seedUsers(t *testing.T, db *gorm.DB) (alice, bob *model.User) {
	t.Helper()
	alice = &model.User{Username: "alice", Nickname: "alice", Password: "hash1", Role: int(auth.RoleUser)}
	bob = &model.User{Username: "bob", Nickname: "bob", Password: "hash2", Role: int(auth.RoleUser)}
	if err := db.Create(alice).Error; err != nil {
		t.Fatalf("seed alice: %v", err)
	}
	if err := db.Create(bob).Error; err != nil {
		t.Fatalf("seed bob: %v", err)
	}
	return
}

func userUC(u *model.User) *auth.UserContext {
	return &auth.UserContext{UserID: u.ID, Username: u.Username, Role: auth.RoleUser, DeviceID: "test"}
}

func guestUC() *auth.UserContext {
	return &auth.UserContext{Role: auth.RoleGuest, DeviceID: "dev-1"}
}

func TestFollowService_ToggleFollow(t *testing.T) {
	ctx := context.Background()
	svc, db := setupFollowTest(t)
	alice, bob := seedUsers(t, db)

	t.Run("follow success", func(t *testing.T) {
		followed, err := svc.ToggleFollow(ctx, userUC(alice), bob.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !followed {
			t.Error("expected followed=true")
		}
	})

	t.Run("unfollow", func(t *testing.T) {
		followed, err := svc.ToggleFollow(ctx, userUC(alice), bob.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if followed {
			t.Error("expected followed=false after unfollow")
		}
	})

	t.Run("follow self", func(t *testing.T) {
		_, err := svc.ToggleFollow(ctx, userUC(alice), alice.ID)
		if err == nil {
			t.Fatal("expected error for self-follow")
		}
	})

	t.Run("follow nonexistent user", func(t *testing.T) {
		_, err := svc.ToggleFollow(ctx, userUC(alice), 9999)
		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})

	t.Run("guest rejected", func(t *testing.T) {
		_, err := svc.ToggleFollow(ctx, guestUC(), bob.ID)
		if err == nil {
			t.Fatal("expected error for guest")
		}
	})
}

func TestFollowService_GetUserProfile(t *testing.T) {
	ctx := context.Background()
	svc, db := setupFollowTest(t)
	alice, bob := seedUsers(t, db)

	t.Run("basic profile", func(t *testing.T) {
		user, pc, fc, fic, followed, err := svc.GetUserProfile(ctx, userUC(alice), bob.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Username != "bob" {
			t.Errorf("expected bob, got %s", user.Username)
		}
		if pc != 0 || fc != 0 || fic != 0 {
			t.Errorf("expected all zeros, got posts=%d followers=%d followings=%d", pc, fc, fic)
		}
		if followed {
			t.Error("expected followed=false")
		}
	})

	t.Run("viewing self", func(t *testing.T) {
		_, _, _, _, followed, err := svc.GetUserProfile(ctx, userUC(alice), alice.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if followed {
			t.Error("expected followed=false when viewing self")
		}
	})

	t.Run("after follow", func(t *testing.T) {
		_, _ = svc.ToggleFollow(ctx, userUC(alice), bob.ID)
		_, _, _, _, followed, err := svc.GetUserProfile(ctx, userUC(alice), bob.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !followed {
			t.Error("expected followed=true after following")
		}
	})

	t.Run("guest views profile", func(t *testing.T) {
		_, _, _, _, followed, err := svc.GetUserProfile(ctx, guestUC(), bob.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if followed {
			t.Error("guest should always see followed=false")
		}
	})

	t.Run("nonexistent user", func(t *testing.T) {
		_, _, _, _, _, err := svc.GetUserProfile(ctx, userUC(alice), 9999)
		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})
}

func TestFollowService_GetFollowers_GetFollowings(t *testing.T) {
	ctx := context.Background()
	svc, db := setupFollowTest(t)
	alice, bob := seedUsers(t, db)
	_, _ = svc.ToggleFollow(ctx, userUC(alice), bob.ID)

	t.Run("bob has one follower", func(t *testing.T) {
		follows, total, _, err := svc.GetFollowers(ctx, userUC(alice), bob.ID, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total=1, got %d", total)
		}
		if len(follows) != 1 {
			t.Fatalf("expected 1 follow, got %d", len(follows))
		}
		if follows[0].FollowerID != alice.ID {
			t.Errorf("expected followerID=%d, got %d", alice.ID, follows[0].FollowerID)
		}
	})

	t.Run("alice has one following", func(t *testing.T) {
		follows, total, _, err := svc.GetFollowings(ctx, userUC(alice), alice.ID, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total=1, got %d", total)
		}
		if len(follows) != 1 {
			t.Fatalf("expected 1 follow, got %d", len(follows))
		}
		if follows[0].FollowingID != bob.ID {
			t.Errorf("expected followingID=%d, got %d", bob.ID, follows[0].FollowingID)
		}
	})
}

func TestFollowService_FollowingIDs_IsFollowing(t *testing.T) {
	ctx := context.Background()
	svc, db := setupFollowTest(t)
	alice, bob := seedUsers(t, db)

	t.Run("FollowingIDs guest rejected", func(t *testing.T) {
		_, err := svc.FollowingIDs(ctx, guestUC())
		if err == nil {
			t.Fatal("expected error for guest")
		}
	})

	t.Run("FollowingIDs empty", func(t *testing.T) {
		ids, err := svc.FollowingIDs(ctx, userUC(alice))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("expected 0 ids, got %d", len(ids))
		}
	})

	t.Run("FollowingIDs after follow", func(t *testing.T) {
		_, _ = svc.ToggleFollow(ctx, userUC(alice), bob.ID)
		ids, err := svc.FollowingIDs(ctx, userUC(alice))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ids) != 1 || ids[0] != bob.ID {
			t.Errorf("expected [%d], got %v", bob.ID, ids)
		}
	})

	t.Run("IsFollowing guest", func(t *testing.T) {
		followed, err := svc.IsFollowing(ctx, guestUC(), bob.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if followed {
			t.Error("guest should return false")
		}
	})
}

func TestFollowService_FindFollowedUserIDs(t *testing.T) {
	ctx := context.Background()
	svc, db := setupFollowTest(t)
	alice, bob := seedUsers(t, db)

	t.Run("empty ids", func(t *testing.T) {
		m, err := svc.FindFollowedUserIDs(ctx, userUC(alice), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("expected empty map, got %v", m)
		}
	})

	t.Run("guest returns empty", func(t *testing.T) {
		m, err := svc.FindFollowedUserIDs(ctx, guestUC(), []uint{bob.ID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("expected empty map for guest, got %v", m)
		}
	})

	t.Run("after follow", func(t *testing.T) {
		_, _ = svc.ToggleFollow(ctx, userUC(alice), bob.ID)
		m, err := svc.FindFollowedUserIDs(ctx, userUC(alice), []uint{bob.ID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !m[bob.ID] {
			t.Error("expected bob to be followed")
		}
	})
}
