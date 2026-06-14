package service

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
)

// mentionRegex 匹配 @username（与注册规则一致：字母/数字/下划线，3-30字符）。
var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9_]{3,30})`)

// MentionSuggestionUser @补全候选用户简要信息。
type MentionSuggestionUser struct {
	ID       uint
	Username string
	Nickname string
}

// MentionService @提及补全 + 提及解析相关业务。
type MentionService struct {
	userRepo    repository.UserRepository
	followRepo  repository.FollowRepository
	nodeModRepo repository.NodeModeratorRepository
	mentionRepo repository.MentionRepository
	log         *zap.Logger
}

func NewMentionService(userRepo repository.UserRepository, followRepo repository.FollowRepository, nodeModRepo repository.NodeModeratorRepository, mentionRepo repository.MentionRepository, log *zap.Logger) *MentionService {
	return &MentionService{userRepo: userRepo, followRepo: followRepo, nodeModRepo: nodeModRepo, mentionRepo: mentionRepo, log: log}
}

// ParseAndSaveMentions 解析帖子内容中的 @username，查找对应用户并批量保存。
func (s *MentionService) ParseAndSaveMentions(ctx context.Context, postID uint, content string) {
	s.parseAndSaveMentions(ctx, postID, nil, content)
}

// ParseAndSaveCommentMentions 解析评论内容中的 @username，查找对应用户并批量保存。
func (s *MentionService) ParseAndSaveCommentMentions(ctx context.Context, postID, commentID uint, content string) {
	s.parseAndSaveMentions(ctx, postID, &commentID, content)
}

func (s *MentionService) parseAndSaveMentions(ctx context.Context, postID uint, commentID *uint, content string) {
	usernames := ExtractMentions(content)
	if len(usernames) == 0 {
		return
	}

	seen := make(map[string]bool)
	var unique []string
	for _, u := range usernames {
		if !seen[u] {
			seen[u] = true
			unique = append(unique, u)
		}
	}

	// 批量查询所有提及的用户，避免 N+1 查询
	users, err := s.userRepo.FindByUsernames(ctx, unique)
	if err != nil {
		s.log.Error("批量查询提及用户失败", zap.Error(err))
		return
	}

	var mentions []model.Mention
	for _, user := range users {
		mentions = append(mentions, model.Mention{
			PostID:    postID,
			CommentID: commentID,
			UserID:    user.ID,
			Username:  user.Username,
		})
	}
	if len(mentions) == 0 {
		return
	}
	if err := s.mentionRepo.CreateBatch(ctx, mentions); err != nil {
		s.log.Error("保存提及记录失败", zap.Error(err))
	}
}

// GetMentions 获取帖子的提及列表。
func (s *MentionService) GetMentions(ctx context.Context, postID uint) ([]model.Mention, error) {
	return s.mentionRepo.FindByPostID(ctx, postID)
}

// ExtractMentions 从内容中提取 @username 列表（去重前）。
func ExtractMentions(content string) []string {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	var usernames []string
	for _, m := range matches {
		// TODO: 过滤掉代码块中的提及（当前未实现，代码块中的 @xxx 也会被匹配）
		name := strings.ToLower(m[1])
		usernames = append(usernames, name)
	}
	return usernames
}

// GetMentionCache 获取当前用户可用于 @补全的用户列表。
func (s *MentionService) GetMentionCache(ctx context.Context, uc *auth.UserContext, sources []string, nodeIDStr string) ([]MentionSuggestionUser, error) {
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return nil, err
	}

	seen := make(map[uint]struct{})
	var allIDs []uint

	for _, src := range sources {
		var ids []uint
		switch src {
		case "following":
			followingIDs, err := s.followRepo.FollowingIDs(ctx, uc.UserID)
			if err != nil {
				return nil, apperror.Internal("查询失败")
			}
			ids = followingIDs
		case "followers":
			followerIDs, err := s.followRepo.FollowerIDs(ctx, uc.UserID)
			if err != nil {
				return nil, apperror.Internal("查询失败")
			}
			ids = followerIDs
		case "admins":
			admins, err := s.userRepo.FindByRoleGTE(ctx, int(auth.RoleAdmin))
			if err != nil {
				return nil, apperror.Internal("查询失败")
			}
			for _, a := range admins {
				ids = append(ids, a.ID)
			}
		case "moderators":
			if nodeIDStr != "" {
				nodeID, err := strconv.ParseUint(nodeIDStr, 10, 64)
				if err == nil {
					modIDs, err := s.nodeModRepo.FindUserIDsByNodeID(ctx, uint(nodeID))
					if err != nil {
						s.log.Warn("查询节点版主失败", zap.Uint("nodeID", uint(nodeID)), zap.Error(err))
					}
					ids = modIDs
				}
			}
		}
		for _, id := range ids {
			if id == uc.UserID {
				continue
			}
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				allIDs = append(allIDs, id)
			}
		}
	}

	if len(allIDs) == 0 {
		return []MentionSuggestionUser{}, nil
	}

	users, err := s.userRepo.FindByIDs(ctx, allIDs)
	if err != nil {
		return nil, apperror.Internal("查询失败")
	}

	// 保持 allIDs 的顺序（去重后的合并顺序）
	orderMap := make(map[uint]int, len(allIDs))
	for i, id := range allIDs {
		orderMap[id] = i
	}
	sorted := make([]MentionSuggestionUser, len(allIDs))
	for _, u := range users {
		idx, ok := orderMap[u.ID]
		if ok {
			nickname := u.Nickname
			if nickname == "" {
				nickname = u.Username
			}
			sorted[idx] = MentionSuggestionUser{
				ID:       u.ID,
				Username: u.Username,
				Nickname: nickname,
			}
		}
	}
	// 过滤零值
	result := make([]MentionSuggestionUser, 0, len(sorted))
	for _, s := range sorted {
		if s.ID != 0 {
			result = append(result, s)
		}
	}
	return result, nil
}
