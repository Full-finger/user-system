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

func setupMentionTest(t *testing.T) (*MentionService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.Node{}, &model.Post{}, &model.Mention{})
	log := zap.NewNop()
	userRepo := repository.NewUserRepository(db)
	mentionRepo := repository.NewMentionRepository(db)
	return NewMentionService(userRepo, nil, nil, mentionRepo, log), db
}

func seedMentionUser(t *testing.T, db *gorm.DB, username string) *model.User {
	t.Helper()
	u := &model.User{Username: username, Nickname: username, Password: "hash"}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("seed user %s: %v", username, err)
	}
	return u
}

func seedMentionNode(t *testing.T, db *gorm.DB) *model.Node {
	t.Helper()
	n := &model.Node{Name: "tech", Slug: "tech", Desc: "test", Color: "#9b8ec4", SortOrder: 1}
	if err := db.Create(n).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	return n
}

func seedMentionPost(t *testing.T, db *gorm.DB, userID, nodeID uint) *model.Post {
	t.Helper()
	p := &model.Post{UserID: userID, NodeID: nodeID, Title: "Test", Content: "Hello"}
	if err := db.Create(p).Error; err != nil {
		t.Fatalf("seed post: %v", err)
	}
	return p
}

// --- ParseAndSaveMentions ---

func TestMentionService_ParseAndSaveMentions(t *testing.T) {
	ctx := context.Background()
	svc, db := setupMentionTest(t)
	node := seedMentionNode(t, db)
	author := seedMentionUser(t, db, "author")
	alice := seedMentionUser(t, db, "alice")
	_ = seedMentionUser(t, db, "bob")
	post := seedMentionPost(t, db, author.ID, node.ID)

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
		db.Exec("DELETE FROM mentions WHERE post_id = ?", post.ID)
	})

	t.Run("multiple mentions with dedup", func(t *testing.T) {
		svc.ParseAndSaveMentions(ctx, post.ID, "hey @alice and @bob and @alice again")
		var mentions []model.Mention
		db.Where("post_id = ?", post.ID).Find(&mentions)
		if len(mentions) != 2 {
			t.Fatalf("expected 2 deduplicated mentions, got %d", len(mentions))
		}
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
		db.Exec("DELETE FROM mentions WHERE post_id = ?", post.ID)
	})

	t.Run("case insensitive match", func(t *testing.T) {
		// 注册一个大写用户名用户，用小写 @mention 提及时应能匹配
		bigAlice := seedMentionUser(t, db, "BigAlice")
		svc.ParseAndSaveMentions(ctx, post.ID, "hello @bigalice how are you?")
		var mentions []model.Mention
		db.Where("post_id = ?", post.ID).Find(&mentions)
		if len(mentions) != 1 {
			t.Fatalf("expected 1 mention (case insensitive), got %d", len(mentions))
		}
		if mentions[0].UserID != bigAlice.ID {
			t.Errorf("expected userID=%d, got %d", bigAlice.ID, mentions[0].UserID)
		}
		if mentions[0].Username != "BigAlice" {
			t.Errorf("expected original username=BigAlice, got %s", mentions[0].Username)
		}
		db.Exec("DELETE FROM mentions WHERE post_id = ?", post.ID)
	})

	t.Run("too short mention ignored", func(t *testing.T) {
		svc.ParseAndSaveMentions(ctx, post.ID, "hi @ab @alice")
		var mentions []model.Mention
		db.Where("post_id = ?", post.ID).Find(&mentions)
		if len(mentions) != 1 {
			t.Fatalf("expected 1 mention (@ab too short), got %d", len(mentions))
		}
		db.Exec("DELETE FROM mentions WHERE post_id = ?", post.ID)
	})
}

// --- GetMentions ---

func TestMentionService_GetMentions(t *testing.T) {
	ctx := context.Background()
	svc, db := setupMentionTest(t)
	node := seedMentionNode(t, db)
	author := seedMentionUser(t, db, "author2")
	alice := seedMentionUser(t, db, "alice2")
	post := seedMentionPost(t, db, author.ID, node.ID)

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

// --- ExtractMentions (pure function) ---

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
			result := ExtractMentions(tt.content)
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
