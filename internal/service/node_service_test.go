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
	db.AutoMigrate(&model.User{}, &model.Node{}, &model.Post{}, &model.Mention{})
	log := zap.NewNop()
	nodeRepo := repository.NewNodeRepository(db)
	userRepo := repository.NewUserRepository(db)
	mentionRepo := repository.NewMentionRepository(db)
	return NewNodeService(nodeRepo, userRepo, mentionRepo, log), db
}

func seedNode(t *testing.T, db *gorm.DB) *model.Node {
	t.Helper()
	n := &model.Node{Name: "技术讨论", Slug: "tech", Desc: "技术相关", Color: "#9b8ec4", SortOrder: 1}
	if err := db.Create(n).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	return n
}

func seedUserForMention(t *testing.T, db *gorm.DB, username string) *model.User {
	t.Helper()
	u := &model.User{Username: username, Nickname: username, Password: "hash"}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("seed user %s: %v", username, err)
	}
	return u
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

// --- ParseAndSaveMentions ---

func TestNodeService_ParseAndSaveMentions(t *testing.T) {
	ctx := context.Background()
	svc, db := setupNodeTest(t)
	node := seedNode(t, db)
	author := seedUserForMention(t, db, "author")
	alice := seedUserForMention(t, db, "alice")
	_ = seedUserForMention(t, db, "bob")

	// Create a post for foreign key constraints
	post := &model.Post{UserID: author.ID, NodeID: node.ID, Title: "Test", Content: "Hello"}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("seed post: %v", err)
	}

	t.Run("no mentions", func(t *testing.T) {
		svc.ParseAndSaveMentions(ctx, post.ID, "just some text without mentions")
		var count int64
		db.Model(&model.Mention{}).Where("post_id = ?", post.ID).Count(&count)
		if count != 0 {
			t.Errorf("expected 0 mentions, got %d", count)
		}
	})

	t.Run("valid single mention", func(t *testing.T) {
		svc.ParseAndSaveMentions(ctx, post.ID, "hello @alice how are you?")
		var mentions []model.Mention
		db.Where("post_id = ?", post.ID).Find(&mentions)
		if len(mentions) != 1 {
			t.Fatalf("expected 1 mention, got %d", len(mentions))
		}
		if mentions[0].Username != "alice" {
			t.Errorf("expected username=alice, got %s", mentions[0].Username)
		}
		if mentions[0].UserID != alice.ID {
			t.Errorf("expected userID=%d, got %d", alice.ID, mentions[0].UserID)
		}
		// cleanup
		db.Exec("DELETE FROM mentions WHERE post_id = ?", post.ID)
	})

	t.Run("multiple mentions with dedup", func(t *testing.T) {
		svc.ParseAndSaveMentions(ctx, post.ID, "hey @alice and @bob and @alice again")
		var mentions []model.Mention
		db.Where("post_id = ?", post.ID).Find(&mentions)
		if len(mentions) != 2 {
			t.Fatalf("expected 2 deduplicated mentions, got %d", len(mentions))
		}
		// cleanup
		db.Exec("DELETE FROM mentions WHERE post_id = ?", post.ID)
	})

	t.Run("nonexistent user skipped", func(t *testing.T) {
		svc.ParseAndSaveMentions(ctx, post.ID, "hello @nonexistent_user and @alice")
		var mentions []model.Mention
		db.Where("post_id = ?", post.ID).Find(&mentions)
		if len(mentions) != 1 {
			t.Fatalf("expected 1 mention (nonexistent skipped), got %d", len(mentions))
		}
		if mentions[0].Username != "alice" {
			t.Errorf("expected username=alice, got %s", mentions[0].Username)
		}
		// cleanup
		db.Exec("DELETE FROM mentions WHERE post_id = ?", post.ID)
	})

	t.Run("too short mention ignored", func(t *testing.T) {
		// username must be 3-30 chars, @ab should be ignored
		svc.ParseAndSaveMentions(ctx, post.ID, "hi @ab @alice")
		var mentions []model.Mention
		db.Where("post_id = ?", post.ID).Find(&mentions)
		if len(mentions) != 1 {
			t.Fatalf("expected 1 mention (@ab too short), got %d", len(mentions))
		}
		// cleanup
		db.Exec("DELETE FROM mentions WHERE post_id = ?", post.ID)
	})
}

// --- GetMentions ---

func TestNodeService_GetMentions(t *testing.T) {
	ctx := context.Background()
	svc, db := setupNodeTest(t)
	node := seedNode(t, db)
	author := seedUserForMention(t, db, "author2")
	alice := seedUserForMention(t, db, "alice2")

	post := &model.Post{UserID: author.ID, NodeID: node.ID, Title: "Test", Content: "Hello"}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("seed post: %v", err)
	}

	t.Run("empty mentions", func(t *testing.T) {
		mentions, err := svc.GetMentions(ctx, post.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(mentions) != 0 {
			t.Errorf("expected 0 mentions, got %d", len(mentions))
		}
	})

	t.Run("with mentions", func(t *testing.T) {
		db.Create(&model.Mention{PostID: post.ID, UserID: alice.ID, Username: "alice2"})
		mentions, err := svc.GetMentions(ctx, post.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(mentions) != 1 {
			t.Fatalf("expected 1 mention, got %d", len(mentions))
		}
		if mentions[0].Username != "alice2" {
			t.Errorf("expected username=alice2, got %s", mentions[0].Username)
		}
	})
}

// --- extractMentions (pure function) ---

func TestExtractMentions(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{"no mentions", "hello world", nil},
		{"single mention", "hello @alice", []string{"alice"}},
		{"multiple mentions", "@alice and @bob", []string{"alice", "bob"}},
		{"duplicate mentions", "@alice @alice @alice", []string{"alice", "alice", "alice"}},
		{"too short", "@ab is too short", nil},
		{"with underscore", "@user_name works", []string{"user_name"}},
		{"max length 30", "@" + repeatChars('a', 30) + " is valid", []string{repeatChars('a', 30)}},
		{"over max length 31 matches first 30", "@" + repeatChars('a', 31) + " ignored", []string{repeatChars('a', 30)}},
		{"in parentheses", "(@alice)", []string{"alice"}},
		{"at end of sentence", "@alice.", []string{"alice"}},
		{"hyphen not matched", "@user-name should only match user", []string{"user"}},
		{"mixed valid invalid", "@ab @alice @bob", []string{"alice", "bob"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMentions(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected, result)
					return
				}
			}
		})
	}
}

func repeatChars(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}
