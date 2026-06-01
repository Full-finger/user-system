package controller

import (
	"strconv"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

// NodeController 节点相关接口的处理器。
type NodeController struct {
	nodeSvc *service.NodeService
	postSvc *service.PostService
}

func NewNodeController(nodeSvc *service.NodeService, postSvc *service.PostService) *NodeController {
	return &NodeController{nodeSvc: nodeSvc, postSvc: postSvc}
}

// ListNodes 获取所有节点。
func (ctrl *NodeController) ListNodes(c echo.Context) error {
	nodes, err := ctrl.nodeSvc.ListNodes(c.Request().Context())
	if err != nil {
		return err
	}
	items := make([]param.NodeResponse, 0, len(nodes))
	for i := range nodes {
		items = append(items, param.ToNodeResponse(&nodes[i]))
	}
	return success(c, param.NodeListResponse{Nodes: items})
}

// GetNode 获取单个节点。
func (ctrl *NodeController) GetNode(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的节点ID")
	}
	node, err := ctrl.nodeSvc.GetNode(c.Request().Context(), uint(id))
	if err != nil {
		return err
	}
	return success(c, param.ToNodeResponse(node))
}

// ListNodePosts 按节点查看帖子，sort=time|replies。
func (ctrl *NodeController) ListNodePosts(c echo.Context) error {
	uc := auth.GetUserContext(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的节点ID")
	}
	page, size := parsePage(c)
	sort := c.QueryParam("sort")
	posts, total, likedMap, err := ctrl.postSvc.ListPostsByNode(c.Request().Context(), uc, uint(id), page, size, sort)
	if err != nil {
		return err
	}
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}
