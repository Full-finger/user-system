package repository

import (
	"context"
	"testing"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFollowRepoTest(t *testing.T) FollowRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.Follow{})

	// 创建测试用户：id=1 (follower), id=2 (following), id=3 (另一个用户)
	db.Create(&model.User{Username: "alice", Nickname: "Alice", Password: "hashed"})
	db.Create(&model.User{Username: "bob", Nickname: "Bob", Password: "hashed"})
	db.Create(&model.User{Username: "charlie", Nickname: "Charlie", Password: "hashed"})

	return NewFollowRepository(db)
}

func TestFollowRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := setupFollowRepoTest(t)

	t.Run("success", func(t *testing.T) {
		follow := &model.Follow{FollowerID: 1, FollowingID: 2}
		err := repo.Create(ctx, follow)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if follow.ID == 0 {
			t.Error("expected ID to be set")
		}
	})

	t.Run("duplicate ignored (OnConflict DoNothing)", func(t *testing.T) {
		// 重复创建不应报错（OnConflict DoNothing）
		follow2 := &model.Follow{FollowerID: 1, FollowingID: 2}
		err := repo.Create(ctx, follow2)
		if err != nil {
			t.Fatalf("expected no error on duplicate, got %v", err)
		}
		// ID 不应被设置（DoNothing 不插入）
		if follow2.ID != 0 {
			t.Error("expected ID=0 for duplicate DoNothing")
		}
	})
}

func TestFollowRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := setupFollowRepoTest(t)
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})

	err := repo.Delete(ctx, 1, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	exists, _ := repo.Exists(ctx, 1, 2)
	if exists {
		t.Error("expected follow to be deleted")
	}
}

func TestFollowRepository_Exists(t *testing.T) {
	ctx := context.Background()
	repo := setupFollowRepoTest(t)
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})

	t.Run("exists", func(t *testing.T) {
		ok, err := repo.Exists(ctx, 1, 2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !ok {
			t.Error("expected true")
		}
	})

	t.Run("not exists", func(t *testing.T) {
		ok, err := repo.Exists(ctx, 2, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if ok {
			t.Error("expected false")
		}
	})
}

func TestFollowRepository_CountFollowers(t *testing.T) {
	ctx := context.Background()
	repo := setupFollowRepoTest(t)
	// bob(id=2) 有两个粉丝：alice(1) 和 charlie(3)
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	_ = repo.Create(ctx, &model.Follow{FollowerID: 3, FollowingID: 2})

	count, err := repo.CountFollowers(ctx, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 followers, got %d", count)
	}
}

func TestFollowRepository_CountFollowings(t *testing.T) {
	ctx := context.Background()
	repo := setupFollowRepoTest(t)
	// alice(id=1) 关注了 bob(2) 和 charlie(3)
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 3})

	count, err := repo.CountFollowings(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 followings, got %d", count)
	}
}

func TestFollowRepository_FindFollowers(t *testing.T) {
	ctx := context.Background()
	repo := setupFollowRepoTest(t)
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	_ = repo.Create(ctx, &model.Follow{FollowerID: 3, FollowingID: 2})

	t.Run("found", func(t *testing.T) {
		follows, total, err := repo.FindFollowers(ctx, 2, 1, 10)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(follows) != 2 {
			t.Errorf("expected 2 follows, got %d", len(follows))
		}
		// 验证 Preload Follower
		if follows[0].Follower.Username == "" {
			t.Error("expected Follower to be preloaded")
		}
	})

	t.Run("no followers", func(t *testing.T) {
		follows, total, err := repo.FindFollowers(ctx, 1, 1, 10)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(follows) != 0 {
			t.Errorf("expected 0 follows, got %d", len(follows))
		}
	})
}

func TestFollowRepository_FindFollowings(t *testing.T) {
	ctx := context.Background()
	repo := setupFollowRepoTest(t)
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 3})

	follows, total, err := repo.FindFollowings(ctx, 1, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(follows) != 2 {
		t.Errorf("expected 2 follows, got %d", len(follows))
	}
	// 验证 Preload Following
	if follows[0].Following.Username == "" {
		t.Error("expected Following to be preloaded")
	}
}

func TestFollowRepository_FollowingIDs(t *testing.T) {
	ctx := context.Background()
	repo := setupFollowRepoTest(t)
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 3})

	t.Run("with data", func(t *testing.T) {
		ids, err := repo.FollowingIDs(ctx, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}
	})

	t.Run("empty", func(t *testing.T) {
		ids, err := repo.FollowingIDs(ctx, 2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("expected 0 ids, got %d", len(ids))
		}
	})
}

func TestFollowRepository_FindFollowedUserIDs(t *testing.T) {
	ctx := context.Background()
	repo := setupFollowRepoTest(t)
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	_ = repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 3})

	t.Run("partial match", func(t *testing.T) {
		m, err := repo.FindFollowedUserIDs(ctx, 1, []uint{2, 99})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(m) != 1 {
			t.Errorf("expected 1 followed, got %d", len(m))
		}
		if !m[2] {
			t.Error("expected user 2 to be followed")
		}
		if m[99] {
			t.Error("expected user 99 to NOT be followed")
		}
	})

	t.Run("empty userIDs", func(t *testing.T) {
		m, err := repo.FindFollowedUserIDs(ctx, 1, []uint{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(m) != 0 {
			t.Errorf("expected empty map, got %d", len(m))
		}
	})

	t.Run("all followed", func(t *testing.T) {
		m, err := repo.FindFollowedUserIDs(ctx, 1, []uint{2, 3})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(m) != 2 {
			t.Errorf("expected 2 followed, got %d", len(m))
		}
	})
}
