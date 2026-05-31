package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
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
	repo := repository.NewUserRepository(db)
	cfg := &config.JWTConfig{
		Secret: "test-secret-key-16chars!",
		Expire: time.Hour,
	}
	log := zap.NewNop()
	return NewUserService(repo, cfg, log)
}

func superAdminUC() *auth.UserContext {
	return &auth.UserContext{
		UserID:   1,
		Username: "superadmin",
		Role:     auth.RoleSuperAdmin,
		DeviceID: "test-device",
	}
}

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		svc := setupUserTest(t)
		user, err := svc.Register(ctx, RegisterInput{Username: "alice", Password: "alice123"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.ID == 0 {
			t.Error("expected user ID to be set")
		}
		if user.Username != "alice" {
			t.Errorf("expected username alice, got %s", user.Username)
		}
		if user.Nickname != "alice" {
			t.Errorf("expected default nickname alice, got %s", user.Nickname)
		}
		if user.Role != int(auth.RoleUser) {
			t.Errorf("expected role %d, got %d", int(auth.RoleUser), user.Role)
		}
		if user.Password == "alice123" {
			t.Error("password should be hashed")
		}
	})

	t.Run("custom nickname", func(t *testing.T) {
		svc := setupUserTest(t)
		user, err := svc.Register(ctx, RegisterInput{Username: "bob", Password: "bob12345", Nickname: "鲍勃"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Nickname != "鲍勃" {
			t.Errorf("expected nickname 鲍勃, got %s", user.Nickname)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		svc := setupUserTest(t)
		_, _ = svc.Register(ctx, RegisterInput{Username: "bob", Password: "bob12345"})
		_, err := svc.Register(ctx, RegisterInput{Username: "bob", Password: "bob54321"})
		if err == nil {
			t.Fatal("expected error for duplicate username")
		}
	})
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()
	svc := setupUserTest(t)
	_, _ = svc.Register(ctx, RegisterInput{Username: "alice", Password: "alice123"})

	t.Run("success with username", func(t *testing.T) {
		token, err := svc.Login(ctx, LoginInput{Username: "alice", Password: "alice123"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if token == "" {
			t.Error("expected token")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login(ctx, LoginInput{Username: "alice", Password: "wrong123"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("nonexistent user", func(t *testing.T) {
		_, err := svc.Login(ctx, LoginInput{Username: "nobody", Password: "nobody123"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUserService_GetProfile(t *testing.T) {
	ctx := context.Background()
	svc := setupUserTest(t)
	user, _ := svc.Register(ctx, RegisterInput{Username: "alice", Password: "alice123"})
	uc := &auth.UserContext{UserID: user.ID, Username: user.Username, Role: auth.RoleUser, DeviceID: "test"}

	t.Run("found", func(t *testing.T) {
		got, err := svc.GetProfile(ctx, uc)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Username != "alice" {
			t.Errorf("expected alice, got %s", got.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		badUC := &auth.UserContext{UserID: 999, Username: "nobody", Role: auth.RoleUser, DeviceID: "test"}
		_, err := svc.GetProfile(ctx, badUC)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	svc := setupUserTest(t)
	user, _ := svc.Register(ctx, RegisterInput{Username: "alice", Password: "alice123"})
	adminUC := superAdminUC()

	t.Run("update password", func(t *testing.T) {
		updated, err := svc.UpdateUser(ctx, adminUC, user.ID, UpdateInput{Password: "newpass88"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Password == user.Password {
			t.Error("expected password to change")
		}
	})

	t.Run("update nickname", func(t *testing.T) {
		updated, err := svc.UpdateUser(ctx, adminUC, user.ID, UpdateInput{Nickname: "爱丽丝"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Nickname != "爱丽丝" {
			t.Errorf("expected nickname 爱丽丝, got %s", updated.Nickname)
		}
	})

	t.Run("update role to moderator", func(t *testing.T) {
		updated, err := svc.UpdateUser(ctx, adminUC, user.ID, UpdateInput{Role: "moderator"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Role != int(auth.RoleModerator) {
			t.Errorf("expected role %d, got %d", int(auth.RoleModerator), updated.Role)
		}
	})

	t.Run("empty update does nothing", func(t *testing.T) {
		updated, err := svc.UpdateUser(ctx, adminUC, user.ID, UpdateInput{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Username != "alice" {
			t.Errorf("expected alice, got %s", updated.Username)
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	svc := setupUserTest(t)
	user, _ := svc.Register(ctx, RegisterInput{Username: "alice", Password: "alice123"})
	adminUC := superAdminUC()

	err := svc.DeleteUser(ctx, adminUC, user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	uc := &auth.UserContext{UserID: user.ID, Username: "alice", Role: auth.RoleUser, DeviceID: "test"}
	_, err = svc.GetProfile(ctx, uc)
	if err == nil {
		t.Fatal("expected user to be deleted")
	}
}

func TestUserService_BindEmail(t *testing.T) {
	ctx := context.Background()
	svc := setupUserTest(t)
	user, _ := svc.Register(ctx, RegisterInput{Username: "alice", Password: "alice123"})
	user2, _ := svc.Register(ctx, RegisterInput{Username: "bob", Password: "bob12345"})
	uc1 := &auth.UserContext{UserID: user.ID, Username: user.Username, Role: auth.RoleUser, DeviceID: "test"}
	uc2 := &auth.UserContext{UserID: user2.ID, Username: user2.Username, Role: auth.RoleUser, DeviceID: "test"}

	t.Run("success", func(t *testing.T) {
		err := svc.BindEmail(ctx, uc1, "alice@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		_ = svc.BindEmail(ctx, uc2, "bob@example.com")
		err := svc.BindEmail(ctx, uc1, "bob@example.com")
		if err == nil {
			t.Fatal("expected error for duplicate email")
		}
	})
}

func TestUserService_ListUsers(t *testing.T) {
	ctx := context.Background()
	svc := setupUserTest(t)
	for i := 0; i < 5; i++ {
		_, _ = svc.Register(ctx, RegisterInput{
			Username: fmt.Sprintf("user%d", i),
			Password: fmt.Sprintf("user%dpass", i),
		})
	}
	adminUC := superAdminUC()

	users, total, err := svc.ListUsers(ctx, adminUC, 1, 3)
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
