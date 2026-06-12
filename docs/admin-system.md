# 管理后台重构需求文档

> 创建时间：2026-06-12
> 状态：规划中

## 1. 概述

将现有的单页面管理后台（仅用户管理）重构为多 Tab 页管理面板，支持数据概览、用户管理、帖子管理、评论管理、节点管理等功能模块。

## 2. 前端 Tab 页设计

### 2.1 Tab 列表

| Tab | 图标 | 说明 |
|-----|------|------|
| 数据概览 | PhChartBar | 统计面板：用户数、帖子数、节点数、评论数 |
| 用户管理 | PhUsers | 用户列表、搜索、编辑角色、删除 |
| 帖子管理 | PhArticleMedium | 帖子列表、搜索、查看详情、删除 |
| 评论管理 | PhChatCircleDots | 评论列表、搜索、删除 |
| 节点管理 | PhFolders | 创建/编辑/删除节点 |

### 2.2 复用组件

- 使用已有的 `tab-bar` / `tab-btn` CSS 组件（`components.css`）
- 复用已有的 `card`、`btn`、`input`、`pill`、`modal-overlay` 等样式

### 2.3 文件结构

```
web/src/views/AdminView.vue              — Tab 布局容器
web/src/components/admin/AdminStatsTab.vue    — 数据概览
web/src/components/admin/AdminUsersTab.vue    — 用户管理（从原 AdminView 提取）
web/src/components/admin/AdminPostsTab.vue    — 帖子管理
web/src/components/admin/AdminCommentsTab.vue — 评论管理
web/src/components/admin/AdminNodesTab.vue    — 节点管理
```

## 3. 后端 API 设计

### 3.1 已有的 Admin API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/users` | 用户列表 |
| GET | `/admin/users/:id` | 用户详情 |
| PUT | `/admin/users/:id` | 编辑用户 |
| DELETE | `/admin/users/:id` | 删除用户 |
| POST | `/admin/moderators` | 任命版主 |
| DELETE | `/admin/posts/:code` | 管理员删帖 |

### 3.2 需新增的 API

#### 数据概览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/stats` | 返回统计数据（用户数/帖子数/节点数/评论数） |

响应示例：
```json
{
  "code": 200,
  "data": {
    "user_count": 100,
    "post_count": 500,
    "node_count": 8,
    "comment_count": 1200
  }
}
```

#### 帖子管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/posts` | 管理员帖子列表（支持分页、搜索关键词、按节点筛选） |

查询参数：`page`, `page_size`, `keyword`(标题搜索), `node_id`(节点筛选)

响应与 `GET /posts` 类似，但额外包含作者信息和节点信息。

#### 评论管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/comments` | 管理员评论列表（支持分页、搜索） |
| DELETE | `/admin/comments/:id` | 管理员删除评论 |

查询参数：`page`, `page_size`, `keyword`(内容搜索)

#### 节点管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/admin/nodes` | 创建节点 |
| PUT | `/admin/nodes/:id` | 编辑节点 |
| DELETE | `/admin/nodes/:id` | 删除节点（仅允许删除空节点） |

创建/编辑请求体：
```json
{
  "name": "节点名称",
  "slug": "node-slug",
  "desc": "节点描述",
  "color": "#9b8ec4",
  "icon": "PhCode",
  "sort_order": 1
}
```

## 4. 后端实现细节

### 4.1 权限控制

所有 `/admin/*` 接口需要在 Service 层通过 `uc.RequireRole(auth.RoleAdmin)` 校验管理员权限。

### 4.2 新增/修改的文件

#### Repository 层

- `internal/repository/node_repo.go` — 新增 `Update`, `Delete` 接口方法
- `internal/repository/node_repo_gorm.go` — 实现 `Update`, `Delete`
- `internal/repository/comment_repo.go` — 新增 `FindPage` (管理后台评论分页), `Delete` 接口方法
- `internal/repository/comment_repo_gorm.go` — 实现上述方法
- `internal/repository/post_repo.go` — 新增 `FindByKeyword` (搜索), `FindByNodeID` 已有
- `internal/repository/user_repo.go` — 新增 `Count` 方法

#### Service 层

- `internal/service/node_service.go` — 新增 `CreateNode`, `UpdateNode`, `DeleteNode`
- `internal/service/comment_service.go` — 新增 `AdminListComments`, `AdminDeleteComment`
- `internal/service/post_service.go` — 新增 `AdminListPosts`

#### Controller 层

- `internal/controller/node_controller.go` — 新增 `CreateNode`, `UpdateNode`, `DeleteNode`
- `internal/controller/post_controller.go` — 新增 `AdminListPosts`
- `internal/controller/comment_controller.go` — 新增 `AdminListComments`, `AdminDeleteComment`
- `internal/controller/user_controller.go` — 新增 `AdminStats`

#### Param 层

- `internal/controller/param/request.go` — 新增 `CreateNodeRequest`, `UpdateNodeRequest`
- `internal/controller/param/node_response.go` — 新增 `ToNodeDetailResponse`（含 SortOrder）
- `internal/controller/param/comment_response.go` — 新增管理后台评论响应

#### 路由

- `internal/router/router.go` — 注册新路由

## 5. 实施顺序

1. ✅ 编写需求文档
2. ⬜ 后端：Repository 层新增方法
3. ⬜ 后端：Service 层新增方法
4. ⬜ 后端：Controller/Param 层新增方法
5. ⬜ 后端：Router 注册新路由
6. ⬜ 前端：api/index.js 补充新接口
7. ⬜ 前端：重构 AdminView.vue 为 Tab 布局
8. ⬜ 前端：实现 AdminStatsTab
9. ⬜ 前端：提取 AdminUsersTab（从现有代码迁移）
10. ⬜ 前端：实现 AdminPostsTab
11. ⬜ 前端：实现 AdminCommentsTab
12. ⬜ 前端：实现 AdminNodesTab
13. ⬜ 编译测试