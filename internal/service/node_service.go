package service

import (
	"context"
	"errors"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NodeService 节点业务服务。
type NodeService struct {
	nodeRepo    repository.NodeRepository
	userRepo    repository.UserRepository
	nodeModRepo repository.NodeModeratorRepository
	log         *zap.Logger
}

func NewNodeService(nodeRepo repository.NodeRepository, userRepo repository.UserRepository, nodeModRepo repository.NodeModeratorRepository, log *zap.Logger) *NodeService {
	return &NodeService{nodeRepo: nodeRepo, userRepo: userRepo, nodeModRepo: nodeModRepo, log: log}
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

// GetNodeModerators 返回指定节点的版主用户列表。
func (s *NodeService) GetNodeModerators(ctx context.Context, nodeID uint) ([]model.User, error) {
	userIDs, err := s.nodeModRepo.FindUserIDsByNodeID(ctx, nodeID)
	if err != nil || len(userIDs) == 0 {
		return nil, err
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		s.log.Error("批量查询版主用户失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return users, nil
}

// CreateNode 管理员创建节点。
func (s *NodeService) CreateNode(ctx context.Context, uc *auth.UserContext, name, slug, desc, color, icon string, sortOrder int) (*model.Node, error) {
	if err := uc.RequireRole(auth.RoleAdmin); err != nil {
		return nil, err
	}
	if name == "" || slug == "" {
		return nil, apperror.BadRequest("节点名称和标识不能为空")
	}
	if _, err := s.nodeRepo.FindBySlug(ctx, slug); err == nil {
		return nil, apperror.BadRequest("节点标识已存在")
	}
	node := &model.Node{
		Name:      name,
		Slug:      slug,
		Desc:      desc,
		Color:     color,
		Icon:      icon,
		SortOrder: sortOrder,
	}
	if err := s.nodeRepo.Create(ctx, node); err != nil {
		s.log.Error("创建节点失败", zap.Error(err))
		return nil, apperror.Internal("创建节点失败")
	}
	return node, nil
}

// UpdateNode 管理员更新节点。
func (s *NodeService) UpdateNode(ctx context.Context, uc *auth.UserContext, id uint, name, slug, desc, color, icon *string, sortOrder *int) (*model.Node, error) {
	if err := uc.RequireRole(auth.RoleAdmin); err != nil {
		return nil, err
	}
	if _, err := s.nodeRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("节点不存在")
		}
		return nil, apperror.Internal("查询失败")
	}
	upd := repository.NodeUpdate{
		Name:      name,
		Slug:      slug,
		Desc:      desc,
		Color:     color,
		Icon:      icon,
		SortOrder: sortOrder,
	}
	if err := s.nodeRepo.Update(ctx, id, upd); err != nil {
		s.log.Error("更新节点失败", zap.Error(err))
		return nil, apperror.Internal("更新节点失败")
	}
	return s.nodeRepo.FindByID(ctx, id)
}

// DeleteNode 管理员删除节点（仅允许删除空节点）。
func (s *NodeService) DeleteNode(ctx context.Context, uc *auth.UserContext, id uint) error {
	if err := uc.RequireRole(auth.RoleAdmin); err != nil {
		return err
	}
	node, err := s.nodeRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("节点不存在")
		}
		return apperror.Internal("查询失败")
	}
	if node.PostCount > 0 {
		return apperror.BadRequest("该节点下还有帖子，无法删除")
	}
	if err := s.nodeRepo.Delete(ctx, id); err != nil {
		s.log.Error("删除节点失败", zap.Error(err))
		return apperror.Internal("删除节点失败")
	}
	return nil
}

// CountNodes 返回节点总数。
func (s *NodeService) CountNodes(ctx context.Context) (int64, error) {
	return s.nodeRepo.Count(ctx)
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
