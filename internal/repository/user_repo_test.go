package repository

import (
	"context"
	"testing"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserRepoTest(t *testing.T) (UserRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.AutoMigrate(&model.User{})
	repo := NewUserRepository(db)
	return repo, db
}

func TestUserRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)

	t.Run("success", func(t *testing.T) {
		user := &model.User{Username: "alice", Nickname: "Alice", Password: "hashed"}
		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.ID == 0 {
			t.Error("expected ID to be set after create")
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		_, _ = setupUserRepoTest(t)
		repo, _ := setupUserRepoTest(t)
		_ = repo.Create(ctx, &model.User{Username: "bob", Password: "p"})
		err := repo.Create(ctx, &model.User{Username: "bob", Password: "p2"})
		if err == nil {
			t.Fatal("expected error for duplicate username")
		}
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)
	created := &model.User{Username: "alice", Password: "hashed"}
	_ = repo.Create(ctx, created)

	t.Run("found", func(t *testing.T) {
		user, err := repo.FindByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Username != "alice" {
			t.Errorf("expected alice, got %s", user.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 999)
		if err == nil {
			t.Fatal("expected error for not found")
		}
	})
}

func TestUserRepository_FindByUsername(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)
	_ = repo.Create(ctx, &model.User{Username: "alice", Password: "hashed"})

	t.Run("found", func(t *testing.T) {
		user, err := repo.FindByUsername(ctx, "alice")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Username != "alice" {
			t.Errorf("expected alice, got %s", user.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByUsername(ctx, "nobody")
		if err == nil {
			t.Fatal("expected error for not found")
		}
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)
	email := "alice@example.com"
	_ = repo.Create(ctx, &model.User{Username: "alice", Password: "hashed", Email: &email})

	t.Run("found", func(t *testing.T) {
		user, err := repo.FindByEmail(ctx, "alice@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Username != "alice" {
			t.Errorf("expected alice, got %s", user.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByEmail(ctx, "nobody@example.com")
		if err == nil {
			t.Fatal("expected error for not found")
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)
	_ = repo.Create(ctx, &model.User{Username: "alice", Password: "hashed", Nickname: "Alice"})

	t.Run("update nickname", func(t *testing.T) {
		nick := "新昵称"
		err := repo.Update(ctx, 1, UserUpdate{Nickname: &nick})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		user, _ := repo.FindByID(ctx, 1)
		if user.Nickname != nick {
			t.Errorf("expected %s, got %s", nick, user.Nickname)
		}
	})

	t.Run("update email", func(t *testing.T) {
		email := "new@example.com"
		err := repo.Update(ctx, 1, UserUpdate{Email: &email})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		user, _ := repo.FindByID(ctx, 1)
		if user.Email == nil || *user.Email != email {
			t.Errorf("expected %s, got %v", email, user.Email)
		}
	})

	t.Run("update password", func(t *testing.T) {
		pw := "newhashed"
		err := repo.Update(ctx, 1, UserUpdate{Password: &pw})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		user, _ := repo.FindByID(ctx, 1)
		if user.Password != pw {
			t.Errorf("expected password updated")
		}
	})
}

func TestUserRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)
	user := &model.User{Username: "alice", Password: "hashed"}
	_ = repo.Create(ctx, user)

	err := repo.Delete(ctx, user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.FindByID(ctx, user.ID)
	if err == nil {
		t.Fatal("expected user to be deleted (soft delete)")
	}
}

func TestUserRepository_Count(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)

	t.Run("empty", func(t *testing.T) {
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	})

	t.Run("with data", func(t *testing.T) {
		_ = repo.Create(ctx, &model.User{Username: "alice", Password: "p"})
		_ = repo.Create(ctx, &model.User{Username: "bob", Password: "p"})
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2, got %d", count)
		}
	})
}

func TestUserRepository_FindPage(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)
	for i := 0; i < 5; i++ {
		username := []string{"a", "b", "c", "d", "e"}[i]
		_ = repo.Create(ctx, &model.User{Username: username, Password: "p"})
	}

	t.Run("page 1 size 3", func(t *testing.T) {
		users, total, err := repo.FindPage(ctx, 1, 3)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
		if len(users) != 3 {
			t.Errorf("expected 3 users, got %d", len(users))
		}
	})

	t.Run("page 3 out of range", func(t *testing.T) {
		users, total, err := repo.FindPage(ctx, 3, 3)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
		if len(users) != 0 {
			t.Errorf("expected 0 users, got %d", len(users))
		}
	})
}

func TestUserRepository_ExistsByUsername(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)
	_ = repo.Create(ctx, &model.User{Username: "alice", Password: "p"})

	t.Run("exists", func(t *testing.T) {
		ok, err := repo.ExistsByUsername(ctx, "alice")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !ok {
			t.Error("expected true")
		}
	})

	t.Run("not exists", func(t *testing.T) {
		ok, err := repo.ExistsByUsername(ctx, "nobody")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if ok {
			t.Error("expected false")
		}
	})
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)
	email := "alice@example.com"
	_ = repo.Create(ctx, &model.User{Username: "alice", Password: "p", Email: &email})

	t.Run("exists", func(t *testing.T) {
		ok, err := repo.ExistsByEmail(ctx, "alice@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !ok {
			t.Error("expected true")
		}
	})

	t.Run("not exists", func(t *testing.T) {
		ok, err := repo.ExistsByEmail(ctx, "nobody@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if ok {
			t.Error("expected false")
		}
	})
}

func TestUserRepository_ExistsByRole(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupUserRepoTest(t)
	_ = repo.Create(ctx, &model.User{Username: "alice", Password: "p", Role: 1})

	t.Run("exists", func(t *testing.T) {
		ok, err := repo.ExistsByRole(ctx, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !ok {
			t.Error("expected true")
		}
	})

	t.Run("not exists", func(t *testing.T) {
		ok, err := repo.ExistsByRole(ctx, 99)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if ok {
			t.Error("expected false")
		}
	})
}
