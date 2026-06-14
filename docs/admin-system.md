# 管理后台功能文档

> 创建时间：2026-06-12
> 最后更新：2026-06-14
> 状态：已完成

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

前端实际采用 `views/admin/` 下分页面组件实现（非 `components/admin/*Tab`）：

```
web/src/views/AdminView.vue              — Tab 布局容器
web/src/views/admin/AdminDashboard.vue   — 数据概览
web/src/views/admin/AdminUsers.vue       — 用户管理（含编辑角色、删除）
web/src/views/admin/AdminPosts.vue       — 帖子管理（搜索、按节点筛选、删除）
web/src/views/admin/AdminComments.vue    — 评论管理（搜索、删除）
web/src/views/admin/AdminNodes.vue       — 节点管理（创建/编辑/删除）
web/src/api/index.js                     — 后台 API 封装
```

## 3. 后端 API 设计

### 3.1 Admin API 全量

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/stats` | 数据概览（用户/帖子/节点/评论数） |
| GET | `/admin/users` | 用户列表（分页） |
| GET | `/admin/users/:id` | 用户详情 |
| PUT | `/admin/users/:id` | 编辑用户（含角色） |
| DELETE | `/admin/users/:id` | 删除用户（软删除） |
| POST | `/admin/moderators` | 任命版主 |
| GET | `/admin/posts` | 帖子列表（支持 keyword、node_id 筛选） |
| DELETE | `/admin/posts/:code` | 删除帖子（管理员全局 / 版主仅管辖节点） |
| GET | `/admin/comments` | 评论列表（支持 keyword 搜索） |
| DELETE | `/admin/comments/:id` | 删除评论（软删除） |
| POST | `/admin/nodes` | 创建节点 |
| PUT | `/admin/nodes/:id` | 更新节点 |
| DELETE | `/admin/nodes/:id` | 删除节点（仅空节点） |

### 3.2 各接口请求/响应说明

#### 数据概览 `GET /admin/stats`

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

#### 帖子管理 `GET /admin/posts`

查询参数：`page`, `page_size`, `keyword`(标题搜索), `node_id`(节点筛选)

响应与 `GET /posts` 一致（`PostListResponse`）。

#### 评论管理

- `GET /admin/comments` — 查询参数：`page`, `page_size`, `keyword`(内容搜索)
- `DELETE /admin/comments/:id` — 软删除

#### 节点管理

- `POST /admin/nodes` — 创建节点
- `PUT /admin/nodes/:id` — 编辑节点（所有字段可选）
- `DELETE /admin/nodes/:id` — 删除节点（仅允许删除空节点）

创建请求体（`CreateNodeRequest`）：
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

更新请求体（`UpdateNodeRequest`）：同上，但所有字段均可选（`*string` / `*int`）。

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

## 5. 实施记录

1. ✅ 编写需求文档
2. ✅ 后端：Repository 层新增方法
3. ✅ 后端：Service 层新增方法
4. ✅ 后端：Controller/Param 层新增方法
5. ✅ 后端：Router 注册新路由
6. ✅ 前端：api/index.js 补充新接口
7. ✅ 前端：重构 AdminView.vue 为 Tab 布局
8. ✅ 前端：实现 AdminDashboard（数据概览）
9. ✅ 前端：迁移 AdminUsers（用户管理）
10. ✅ 前端：实现 AdminPosts（帖子管理）
11. ✅ 前端：实现 AdminComments（评论管理）
12. ✅ 前端：实现 AdminNodes（节点管理）
13. ✅ 编译测试通过
