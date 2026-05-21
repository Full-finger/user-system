package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTest(t *testing.T) *UserService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.AutoMigrate(&model.User{})
	cfg := &config.JWTConfig{
		Secret: "test-secret",
		Expire: time.Hour,
	}
	return NewUserService(db, cfg)
}

func TestUserService_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := setupUserTest(t)
		user, err := svc.Register(RegisterInput{Username: "alice", Password: "123456"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.ID == 0 {
			t.Error("expected user ID to be set")
		}
		if user.Username != "alice" {
			t.Errorf("expected username alice, got %s", user.Username)
		}
		if user.Role != "user" {
			t.Errorf("expected role user, got %s", user.Role)
		}
		if user.Password == "123456" {
			t.Error("password should be hashed")
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		svc := setupUserTest(t)
		_, _ = svc.Register(RegisterInput{Username: "bob", Password: "123456"})
		_, err := svc.Register(RegisterInput{Username: "bob", Password: "654321"})
		if err == nil {
			t.Fatal("expected error for duplicate username")
		}
	})
}

func TestUserService_Login(t *testing.T) {
	svc := setupUserTest(t)
	_, _ = svc.Register(RegisterInput{Username: "alice", Password: "123456"})

	t.Run("success with username", func(t *testing.T) {
		token, err := svc.Login(LoginInput{Username: "alice", Password: "123456"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if token == "" {
			t.Error("expected token")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login(LoginInput{Username: "alice", Password: "wrong"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("nonexistent user", func(t *testing.T) {
		_, err := svc.Login(LoginInput{Username: "nobody", Password: "123456"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUserService_GetProfile(t *testing.T) {
	svc := setupUserTest(t)
	user, _ := svc.Register(RegisterInput{Username: "alice", Password: "123456"})

	t.Run("found", func(t *testing.T) {
		got, err := svc.GetProfile(user.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Username != "alice" {
			t.Errorf("expected alice, got %s", got.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetProfile(999)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	svc := setupUserTest(t)
	user, _ := svc.Register(RegisterInput{Username: "alice", Password: "123456"})

	t.Run("update password", func(t *testing.T) {
		updated, err := svc.UpdateUser(user.ID, UpdateInput{Password: "newpass"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Password == user.Password {
			t.Error("expected password to change")
		}
	})

	t.Run("update role", func(t *testing.T) {
		updated, err := svc.UpdateUser(user.ID, UpdateInput{Role: "admin"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Role != "admin" {
			t.Errorf("expected admin, got %s", updated.Role)
		}
	})

	t.Run("empty update does nothing", func(t *testing.T) {
		updated, err := svc.UpdateUser(user.ID, UpdateInput{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Username != "alice" {
			t.Errorf("expected alice, got %s", updated.Username)
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	svc := setupUserTest(t)
	user, _ := svc.Register(RegisterInput{Username: "alice", Password: "123456"})

	err := svc.DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = svc.GetProfile(user.ID)
	if err == nil {
		t.Fatal("expected user to be deleted")
	}
}

func TestUserService_BindEmail(t *testing.T) {
	svc := setupUserTest(t)
	user, _ := svc.Register(RegisterInput{Username: "alice", Password: "123456"})
	user2, _ := svc.Register(RegisterInput{Username: "bob", Password: "123456"})

	t.Run("success", func(t *testing.T) {
		err := svc.BindEmail(user.ID, "alice@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		_ = svc.BindEmail(user2.ID, "bob@example.com")
		err := svc.BindEmail(user.ID, "bob@example.com")
		if err == nil {
			t.Fatal("expected error for duplicate email")
		}
	})
}

func TestUserService_ListUsers(t *testing.T) {
	svc := setupUserTest(t)
	for i := 0; i < 5; i++ {
		_, _ = svc.Register(RegisterInput{
			Username: fmt.Sprintf("user%d", i),
			Password: "123456",
		})
	}

	users, total, err := svc.ListUsers(1, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}
