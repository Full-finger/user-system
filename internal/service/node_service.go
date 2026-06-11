package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NodeService 节点业务服务。
type NodeService struct {
	nodeRepo    repository.NodeRepository
	userRepo    repository.UserRepository
	mentionRepo repository.MentionRepository
	log         *zap.Logger
}

func NewNodeService(nodeRepo repository.NodeRepository, userRepo repository.UserRepository, mentionRepo repository.MentionRepository, log *zap.Logger) *NodeService {
	return &NodeService{nodeRepo: nodeRepo, userRepo: userRepo, mentionRepo: mentionRepo, log: log}
}

// ListNodes 返回所有节点。
func (s *NodeService) ListNodes(ctx context.Context) ([]model.Node, error) {
	nodes, err := s.nodeRepo.FindAll(ctx)
	if err != nil {
		s.log.Error("查询节点列表失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return nodes, nil
}

// GetNode 返回单个节点。
func (s *NodeService) GetNode(ctx context.Context, id uint) (*model.Node, error) {
	node, err := s.nodeRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("节点不存在")
		}
		s.log.Error("查询节点失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return node, nil
}

// DefaultNodes 返回前端 ExploreView 对应的默认节点。
func DefaultNodes() []model.Node {
	return []model.Node{
		{Name: "技术讨论", Slug: "tech", Desc: "编程语言、框架、工具的深度技术讨论", Color: "#9b8ec4", Icon: "PhCode", SortOrder: 1},
		{Name: "项目展示", Slug: "showcase", Desc: "分享你的开源项目和作品", Color: "#6db89a", Icon: "PhRocketLaunch", SortOrder: 2},
		{Name: "新手求助", Slug: "help", Desc: "遇到问题？在这里寻求帮助", Color: "#7ba4d4", Icon: "PhQuestion", SortOrder: 3},
		{Name: "资源分享", Slug: "resources", Desc: "优质教程、工具、资源推荐", Color: "#d4a07a", Icon: "PhGift", SortOrder: 4},
		{Name: "公告/官方", Slug: "announcements", Desc: "社区公告和官方信息", Color: "#d4b85a", Icon: "PhMegaphone", SortOrder: 5},
		{Name: "闲聊灌水", Slug: "offtopic", Desc: "轻松闲聊，分享日常", Color: "#c47a99", Icon: "PhChatsCircle", SortOrder: 6},
		{Name: "求职招聘", Slug: "jobs", Desc: "工作机会和求职信息", Color: "#8bb8a8", Icon: "PhBriefcase", SortOrder: 7},
		{Name: "Bug 反馈", Slug: "bug-feedback", Desc: "反馈社区平台的 Bug 和建议", Color: "#c4987a", Icon: "PhBug", SortOrder: 8},
	}
}

// SeedNodes 将默认节点写入数据库（已存在则跳过）。
func (s *NodeService) SeedNodes(ctx context.Context) {
	for _, n := range DefaultNodes() {
		if _, err := s.nodeRepo.FindBySlug(ctx, n.Slug); err != nil {
			if err := s.nodeRepo.Create(ctx, &n); err != nil {
				s.log.Error("种子节点创建失败", zap.String("slug", n.Slug), zap.Error(err))
			}
		}
	}
}

// mentionRegex 匹配 @username（与注册规则一致：字母/数字/下划线，3-30字符）。
var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9_]{3,30})`)

// ParseAndSaveMentions 解析帖子内容中的 @username，查找对应用户并批量保存。
func (s *NodeService) ParseAndSaveMentions(ctx context.Context, postID uint, content string) {
	usernames := extractMentions(content)
	if len(usernames) == 0 {
		return
	}

	// 去重
	seen := make(map[string]bool)
	var unique []string
	for _, u := range usernames {
		if !seen[u] {
			seen[u] = true
			unique = append(unique, u)
		}
	}

	var mentions []model.Mention
	for _, name := range unique {
		user, err := s.userRepo.FindByUsername(ctx, name)
		if err != nil {
			continue // 用户不存在就跳过
		}
		mentions = append(mentions, model.Mention{
			PostID:   postID,
			UserID:   user.ID,
			Username: user.Username,
		})
	}
	if err := s.mentionRepo.CreateBatch(ctx, mentions); err != nil {
		s.log.Error("保存提及记录失败", zap.Error(err))
	}
}

// GetMentions 获取帖子的提及列表。
func (s *NodeService) GetMentions(ctx context.Context, postID uint) ([]model.Mention, error) {
	return s.mentionRepo.FindByPostID(ctx, postID)
}

func extractMentions(content string) []string {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	var usernames []string
	for _, m := range matches {
		// TODO: 过滤掉代码块中的提及（当前未实现，代码块中的 @xxx 也会被匹配）
		name := strings.ToLower(m[1])
		usernames = append(usernames, name)
	}
	return usernames
}
