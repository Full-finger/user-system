package controller

import (
	"strconv"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

// resolveUsername 从路由参数中解析用户名并查找用户，失败时返回错误响应。
func (ctrl *PostController) resolveUsername(c echo.Context) (*model.User, error) {
	username := c.Param("username")
	if username == "" {
		return nil, apperror.BadRequest("缺少用户名")
	}
	return ctrl.followSvc.ResolveUsername(c.Request().Context(), username)
}

// PostController 帖子相关接口的处理器。
type PostController struct {
	postSvc   *service.PostService
	nodeSvc   *service.NodeService
	followSvc *service.FollowService
	likeSvc   *service.LikeService
}

func NewPostController(postSvc *service.PostService, nodeSvc *service.NodeService, followSvc *service.FollowService, likeSvc *service.LikeService) *PostController {
	return &PostController{postSvc: postSvc, nodeSvc: nodeSvc, followSvc: followSvc, likeSvc: likeSvc}
}

func parsePage(c echo.Context) (page, size int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	size, _ = strconv.Atoi(c.QueryParam("page_size"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return
}

// optionalUserID 尝试从 context 获取当前登录用户 ID，未登录时返回 (0, false)。
func optionalUserID(c echo.Context) (uint, bool) {
	id, ok := c.Get("user_id").(uint)
	return id, ok
}

// ── 节点 ───────────────────────────────────────────────────

// ListNodes 获取所有节点。
func (ctrl *PostController) ListNodes(c echo.Context) error {
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
func (ctrl *PostController) GetNode(c echo.Context) error {
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
func (ctrl *PostController) ListNodePosts(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的节点ID")
	}
	page, size := parsePage(c)
	sort := c.QueryParam("sort")
	posts, total, err := ctrl.postSvc.ListPostsByNode(c.Request().Context(), uint(id), page, size, sort)
	if err != nil {
		return err
	}
	likedMap := ctrl.buildLikedMap(c, posts)
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}

// ── 帖子 ───────────────────────────────────────────────────

// CreatePost 发帖。
func (ctrl *PostController) CreatePost(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	var req param.CreatePostRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	post, err := ctrl.postSvc.CreatePost(c.Request().Context(), userID, req.NodeID, req.Title, req.Content)
	if err != nil {
		return err
	}
	mentions, _ := ctrl.nodeSvc.GetMentions(c.Request().Context(), post.ID)
	return success(c, param.ToPostResponse(post, mentions, nil))
}

// DeletePost 删帖。
func (ctrl *PostController) DeletePost(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	role, _ := c.Get("role").(string)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	if err := ctrl.postSvc.DeletePost(c.Request().Context(), userID, uint(id), role == "admin"); err != nil {
		return err
	}
	return success(c, nil)
}

// GetPost 查看帖子详情。
func (ctrl *PostController) GetPost(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	post, mentions, err := ctrl.postSvc.GetPost(c.Request().Context(), uint(id))
	if err != nil {
		return err
	}
	likedMap := ctrl.buildLikedMapForPosts(c, []uint{post.ID})
	return success(c, param.ToPostResponse(post, mentions, likedMap))
}

// ListPosts 全站帖子列表。
func (ctrl *PostController) ListPosts(c echo.Context) error {
	page, size := parsePage(c)
	posts, total, err := ctrl.postSvc.ListPosts(c.Request().Context(), page, size)
	if err != nil {
		return err
	}
	likedMap := ctrl.buildLikedMap(c, posts)
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}

// ListUserPosts 某用户的帖子列表。
func (ctrl *PostController) ListUserPosts(c echo.Context) error {
	target, err := ctrl.resolveUsername(c)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	posts, total, err := ctrl.postSvc.ListUserPosts(c.Request().Context(), target.ID, page, size)
	if err != nil {
		return err
	}
	likedMap := ctrl.buildLikedMap(c, posts)
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}

// ListFeed 关注用户的帖子（时间线）。
func (ctrl *PostController) ListFeed(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	page, size := parsePage(c)
	ids, err := ctrl.followSvc.FollowingIDs(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	posts, total, err := ctrl.postSvc.ListFeed(c.Request().Context(), ids, page, size)
	if err != nil {
		return err
	}
	likedMap := ctrl.buildLikedMap(c, posts)
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}

// ── 点赞 ───────────────────────────────────────────────────

// ToggleLike 点赞/取消点赞。
func (ctrl *PostController) ToggleLike(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	liked, err := ctrl.postSvc.ToggleLike(c.Request().Context(), userID, uint(id))
	if err != nil {
		return err
	}
	return success(c, map[string]bool{"liked": liked})
}

// ListLikedPosts 某用户点赞过的帖子。
func (ctrl *PostController) ListLikedPosts(c echo.Context) error {
	target, err := ctrl.resolveUsername(c)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	likes, total, err := ctrl.postSvc.ListLikedPosts(c.Request().Context(), target.ID, page, size)
	if err != nil {
		return err
	}
	postIDs := make([]uint, 0, len(likes))
	for i := range likes {
		postIDs = append(postIDs, likes[i].PostID)
	}
	likedMap := ctrl.buildLikedMapForPosts(c, postIDs)
	return success(c, param.ToLikedPostListResponse(likes, total, page, size, likedMap))
}

// ── 关注 ───────────────────────────────────────────────────

// ToggleFollow 关注/取消关注。
func (ctrl *PostController) ToggleFollow(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	target, err := ctrl.resolveUsername(c)
	if err != nil {
		return err
	}
	followed, err := ctrl.followSvc.ToggleFollow(c.Request().Context(), userID, target.ID)
	if err != nil {
		return err
	}
	return success(c, map[string]bool{"followed": followed})
}

// GetFollowers 某用户的粉丝列表。
func (ctrl *PostController) GetFollowers(c echo.Context) error {
	target, err := ctrl.resolveUsername(c)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	follows, total, err := ctrl.followSvc.GetFollowers(c.Request().Context(), target.ID, page, size)
	if err != nil {
		return err
	}
	userIDs := make([]uint, 0, len(follows))
	for i := range follows {
		userIDs = append(userIDs, follows[i].FollowerID)
	}
	followedMap := ctrl.buildFollowedMap(c, userIDs)
	return success(c, param.ToFollowerListResponse(follows, total, page, size, followedMap))
}

// GetFollowings 某用户的关注列表。
func (ctrl *PostController) GetFollowings(c echo.Context) error {
	target, err := ctrl.resolveUsername(c)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	follows, total, err := ctrl.followSvc.GetFollowings(c.Request().Context(), target.ID, page, size)
	if err != nil {
		return err
	}
	userIDs := make([]uint, 0, len(follows))
	for i := range follows {
		userIDs = append(userIDs, follows[i].FollowingID)
	}
	followedMap := ctrl.buildFollowedMap(c, userIDs)
	return success(c, param.ToFollowingListResponse(follows, total, page, size, followedMap))
}

// GetUserProfile 查看其他用户信息（含统计）。
func (ctrl *PostController) GetUserProfile(c echo.Context) error {
	target, err := ctrl.resolveUsername(c)
	if err != nil {
		return err
	}
	user, postCount, followerCount, followingCount, err := ctrl.followSvc.GetUserProfile(c.Request().Context(), target.ID)
	if err != nil {
		return err
	}
	followed := false
	if currentID, ok := optionalUserID(c); ok && currentID != target.ID {
		followed, _ = ctrl.followSvc.IsFollowing(c.Request().Context(), currentID, target.ID)
	}
	return success(c, param.ToUserProfileResponse(user, postCount, followerCount, followingCount, followed))
}

// ── 辅助方法 ───────────────────────────────────────────────

// buildLikedMap 从帖子列表构建 likedMap，未登录时返回 nil。
func (ctrl *PostController) buildLikedMap(c echo.Context, posts []model.Post) map[uint]bool {
	if len(posts) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(posts))
	for i := range posts {
		ids = append(ids, posts[i].ID)
	}
	return ctrl.buildLikedMapForPosts(c, ids)
}

// buildLikedMapForPosts 从帖子 ID 列表构建 likedMap，未登录时返回 nil。
func (ctrl *PostController) buildLikedMapForPosts(c echo.Context, postIDs []uint) map[uint]bool {
	userID, ok := optionalUserID(c)
	if !ok || len(postIDs) == 0 {
		return nil
	}
	m, _ := ctrl.likeSvc.FindLikedPostIDs(c.Request().Context(), userID, postIDs)
	return m
}

// buildFollowedMap 从用户 ID 列表构建 followedMap，未登录时返回 nil。
func (ctrl *PostController) buildFollowedMap(c echo.Context, userIDs []uint) map[uint]bool {
	userID, ok := optionalUserID(c)
	if !ok || len(userIDs) == 0 {
		return nil
	}
	m, _ := ctrl.followSvc.FindFollowedUserIDs(c.Request().Context(), userID, userIDs)
	return m
}
