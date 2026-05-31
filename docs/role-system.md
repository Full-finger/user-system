# 角色系统重构设计文档

## 1. 背景与动机

### 1.1 现状

当前项目的角色系统极为简陋：

- `model.User.Role` 为 `string` 类型，仅有 `"admin"` 和 `"user"` 两个值
- 权限检查分散在三层：`AdminOnly()` 中间件做一层、Controller 里 `role == "admin"` 判断做一层、Service 层通过 `isAdmin bool` 参数再传一层
- Controller 中到处是 `c.Get("user_id").(uint)` / `c.Get("role").(string)` 的手动提取
- 游客（未登录用户）无任何身份标识，`OptionalJWTMiddleware` 直接放行不注入信息
- 限流纯靠 `IP + 路径`，无法按用户或设备维度限流
- 无 `device_id` 概念，无法跨请求追踪同一游客

### 1.2 目标

- 建立 6 级角色层级体系，支持未来扩展
- 引入统一 `UserContext`，贯穿中间件 → Controller → Service 全链路
- 权限判断全部下沉到 Service 层，中间件只负责"你是谁"
- 游客通过 `device_id` 获得可追踪身份，可选签发 Guest JWT
- 限流支持 `user_id > device_id > IP` 优先级

## 2. 角色体系

### 2.1 角色定义

| 角色 | 常量 | iota 值 | 说明 |
|------|------|---------|------|
| 游客 | `RoleGuest` | 0 | 未注册用户，通过 device_id 标识 |
| 普通用户 | `RoleUser` | 1 | 注册后的默认角色 |
| 认证用户 | `RoleVerifiedUser` | 2 | 预留角色，后续接入经验系统后实现自动升级 |
| 版主 | `RoleModerator` | 3 | 按节点划分，仅能管理自己负责的节点 |
| 管理员 | `RoleAdmin` | 4 | 全局管理权限，但不能修改超级管理员 |
| 超级管理员（站长） | `RoleSuperAdmin` | 5 | 最高权限，唯一或少数几个 |

### 2.2 层级继承

高角色**自动继承**低角色的所有权限。权限判断基于 `Level()` 比较：

```go
// 示例：删帖权限 — 作者本人 或 RoleModerator 及以上
if post.UserID != uc.UserID && uc.Role.Level() < RoleModerator.Level() {
    return apperror.Forbidden("无权删除此帖子")
}
```

### 2.3 特殊说明

- **VerifiedUser**：当前阶段不实际赋予任何用户，仅作为类型预留。后续接入经验系统（如发帖数、注册时长等条件）后，由后台任务自动升级。
- **Moderator**：版主权限**按节点划分**，需要在数据库中维护 `node_moderators` 关联表。版主只能对自己负责的节点执行管理操作（如删帖、置顶），对非管辖节点仅拥有普通用户权限。
- **SuperAdmin vs Admin**：SuperAdmin 是站长级别，可以管理管理员账号；Admin 不能修改 SuperAdmin 的角色和信息。

## 3. 统一 UserContext

### 3.1 结构体定义

```go
// internal/auth/context.go

type UserContext struct {
    Role     Role
    UserID   uint   // 游客为 0
    DeviceID string // 所有请求都有
    Username string // 游客为空
}

// IsGuest 便捷判断
func (uc *UserContext) IsGuest() bool {
    return uc.Role == RoleGuest
}

// IsAuthenticated 是否为已登录用户（非游客）
func (uc *UserContext) IsAuthenticated() bool {
    return uc.Role > RoleGuest
}

// RequireRole 当角色级别不足时返回 Forbidden 错误，否则返回 nil
func (uc *UserContext) RequireRole(minRole Role) error {
    if uc.Role.Level() < minRole.Level() {
        return apperror.Forbidden("权限不足")
    }
    return nil
}
```

### 3.2 与 Echo Context 的桥接

```go
const userContextKey = "user_context"

// GetUserContext 从 echo.Context 提取 UserContext，永远不为 nil
func GetUserContext(c echo.Context) *UserContext {
    if uc, ok := c.Get(userContextKey).(*UserContext); ok {
        return uc
    }
    // 兜底：不应发生，但防止 panic
    return &UserContext{Role: RoleGuest}
}

// SetUserContext 注入 UserContext 到 echo.Context
func SetUserContext(c echo.Context, uc *UserContext) {
    c.Set(userContextKey, uc)
}
```

### 3.3 Controller 使用模式

```go
// 之前
func (ctrl *PostController) CreatePost(c echo.Context) error {
    userID, ok := c.Get("user_id").(uint)
    if !ok {
        return apperror.Unauthorized("未认证")
    }
    // ...
}

// 之后
func (ctrl *PostController) CreatePost(c echo.Context) error {
    uc := auth.GetUserContext(c)
    // ...
    post, err := ctrl.postSvc.CreatePost(c.Request().Context(), uc, req.NodeID, req.Title, req.Content)
}
```

## 4. 中间件重构

### 4.1 统一 AuthMiddleware

替代现有的 `JWTMiddleware` + `OptionalJWTMiddleware` + `AdminOnly` 三个中间件。

**处理流程**：

```
请求进入
  │
  ├─ 有 Authorization: Bearer xxx 头？
  │    ├─ 是 → 解析 JWT
  │    │    ├─ 含 user_id → 构建用户 UserContext（Role 从 JWT claims 读取）
  │    │    ├─ 仅含 device_id（Guest JWT）→ 构建游客 UserContext
  │    │    └─ 解析失败 → 构建纯 Guest UserContext（device_id 从 X-Device-ID 头读取）
  │    │
  │    └─ 否 → 构建 Guest UserContext
  │             ├─ 读 X-Device-ID 头
  │             └─ 头为空 → 生成随机 device_id
  │
  ├─ 将 UserContext 注入 echo.Context
  ├─ 限流检查（优先 user_id，fallback device_id，最后 fallback IP）
  └─ next()
```

**关键设计决策**：
- 无 Token 的请求**不拦截**，自动赋予 Guest 身份。如果未来需要强制要求 Token，可以在网关层拦截无 Token 请求，返回特定错误码让前端去换取 Guest JWT 后重试。
- JWT 解析失败时**降级为 Guest** 而非返回 401，保证健壮性。

### 4.2 限流升级

```go
// 限流 key 优先级
func rateLimitKey(uc *UserContext, path string, ip string) string {
    if uc.IsAuthenticated() {
        return fmt.Sprintf("ratelimit:user:%d:%s", uc.UserID, path)
    }
    if uc.DeviceID != "" {
        return fmt.Sprintf("ratelimit:device:%s:%s", uc.DeviceID, path)
    }
    return fmt.Sprintf("ratelimit:ip:%s:%s", ip, path)
}
```

### 4.3 移除的中间件

- `JWTMiddleware` → 被 `AuthMiddleware` 替代
- `OptionalJWTMiddleware` → 被 `AuthMiddleware` 替代
- `AdminOnly` → 删除，权限判断下沉到 Service 层

## 5. Guest JWT

### 5.1 签发端点

```
POST /api/guest-token
Content-Type: application/json

{
  "device_id": "abc123def456..."
}

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### 5.2 Guest JWT Payload

```json
{
  "device_id": "abc123def456...",
  "role": "guest",
  "exp": 1700000000
}
```

- **不包含** `user_id` 和 `username`
- 过期时间建议 7 天（可配置），比用户 JWT 短
- 使用独立的 `guest_jwt.secret`（可与主 JWT 共用，但建议分开以便独立轮换）

### 5.3 两种游客识别方式

前端可自由选择：

| 方式 | 优点 | 缺点 |
|------|------|------|
| (a) 调 `/api/guest-token` 获取 Guest JWT | 标准 Bearer 流程，中间件逻辑统一 | 多一次请求 |
| (b) 直接带 `X-Device-ID` 请求头 | 零额外请求 | 需要中间件特殊处理无 Token 情况 |

两种方式中间件都支持。

### 5.4 Device ID 生成

**由前端负责生成**（使用 fingerprintjs 等库），通过请求头或接口参数传递。后端**不信任也不生成** device_id，仅在无 device_id 时生成随机值作为兜底（但此时跨会话追踪不可靠）。

## 6. Service 层鉴权模式

### 6.1 核心原则

> **中间件负责"你是谁"，Service 负责"你能看什么/做什么"。**

Service 方法签名统一改为接收 `*UserContext`：

```go
// 之前
func (s *PostService) DeletePost(ctx context.Context, userID uint, code string, isAdmin bool) error

// 之后
func (s *PostService) DeletePost(ctx context.Context, uc *auth.UserContext, code string) error
```

### 6.2 权限判断示例

```go
// 发帖 — 需要登录
func (s *PostService) CreatePost(ctx context.Context, uc *auth.UserContext, ...) (*model.Post, error) {
    if err := uc.RequireRole(auth.RoleUser); err != nil {
        return nil, err  // Guest 会得到 403 Forbidden
    }
    // ...
}

// 删帖 — 作者本人 或 Moderator 及以上
func (s *PostService) DeletePost(ctx context.Context, uc *auth.UserContext, code string) error {
    // ...
    if post.UserID != uc.UserID {
        if uc.Role.Level() < auth.RoleModerator.Level() {
            return apperror.Forbidden("无权删除此帖子")
        }
        // 版主需要检查是否管辖该节点
        if uc.Role == auth.RoleModerator {
            if !s.nodeModRepo.IsModerator(ctx, uc.UserID, post.NodeID) {
                return apperror.Forbidden("无权管理该节点的帖子")
            }
        }
    }
    // ...
}

// 管理用户 — Admin 及以上
func (s *UserService) UpdateUser(ctx context.Context, uc *auth.UserContext, id uint, in UpdateInput) (*model.User, error) {
    if err := uc.RequireRole(auth.RoleAdmin); err != nil {
        return nil, err
    }
    // SuperAdmin 不能被 Admin 修改
    target, err := s.GetProfile(ctx, id)
    if err != nil {
        return nil, err
    }
    if target.Role == auth.RoleSuperAdmin && uc.Role != auth.RoleSuperAdmin {
        return nil, apperror.Forbidden("无权修改超级管理员")
    }
    // ...
}

// 浏览帖子 — 所有人可访问（包括 Guest）
func (s *PostService) GetPost(ctx context.Context, uc *auth.UserContext, code string) (*model.Post, []model.Mention, error) {
    // 无需权限检查
    // 但可根据 uc.IsGuest() 决定是否返回个性化数据
    // ...
}
```

### 6.3 数据差异组装

Service 层根据 UserContext 决定返回哪些数据：

```go
// 构建帖子响应时
func (s *PostService) enrichPosts(ctx context.Context, uc *auth.UserContext, posts []model.Post) (*PostListResult, error) {
    result := &PostListResult{Posts: posts}

    if uc.IsAuthenticated() {
        // 已登录：查询 liked/followed 状态
        result.LikedMap, _ = s.likeSvc.FindLikedPostIDs(ctx, uc.UserID, postIDs)
        result.FollowedMap, _ = s.followSvc.FindFollowedUserIDs(ctx, uc.UserID, userIDs)
    }
    // Guest：LikedMap/FollowedMap 为 nil，前端按 nil 渲染默认状态

    return result, nil
}
```

## 7. 数据库变更

### 7.1 users 表

```sql
-- 角色字段从 varchar 改为 int
ALTER TABLE users ALTER COLUMN role TYPE integer USING 
  CASE role 
    WHEN 'admin' THEN 4 
    WHEN 'user' THEN 1 
    ELSE 1 
  END;

-- 添加检查约束
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role >= 0 AND role <= 5);

-- 修改默认值
ALTER TABLE users ALTER COLUMN role SET DEFAULT 1;
```

### 7.2 新增 node_moderators 表

```sql
CREATE TABLE node_moderators (
    node_id  INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    user_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (node_id, user_id)
);

CREATE INDEX idx_node_moderators_user_id ON node_moderators(user_id);
```

### 7.3 GORM Model

```go
// internal/model/node_moderator.go
type NodeModerator struct {
    NodeID    uint      `json:"node_id" gorm:"primaryKey"`
    UserID    uint      `json:"user_id" gorm:"primaryKey"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 7.4 迁移脚本

迁移 SQL 将放在 `init-scripts/migrate-role-system.sql`，支持从现有结构平滑升级。

## 8. 路由变更

### 8.1 之前

```go
public  → IP限流                        // 注册/登录/公开节点
optAuth → IP限流 + OptionalJWT          // 帖子列表（可选个性化）
auth    → JWTMiddleware                 // 发帖/点赞/关注
admin   → JWTMiddleware + AdminOnly     // 管理接口
```

### 8.2 之后

```go
api → AuthMiddleware（统一身份解析 + 限流）
  ├── POST /guest-token                 // 获取 Guest JWT
  ├── GET  /check-username              // 公开
  ├── POST /register                    // 公开
  ├── POST /login                       // 公开
  ├── POST /send-code                   // 公开
  ├── POST /code-login                  // 公开
  ├── GET  /nodes                       // 公开
  ├── GET  /nodes/:id                   // 公开
  ├── GET  /nodes/:id/posts             // Service 层按 uc 返回差异化数据
  ├── GET  /posts                       // 同上
  ├── GET  /posts/:id                   // 同上
  ├── GET  /users/:username/posts       // 同上
  ├── GET  /users/:username/likes       // 同上
  ├── GET  /users/:username/followers   // 同上
  ├── GET  /users/:username/followings  // 同上
  ├── GET  /users/:username             // 同上
  ├── GET  /profile                     // Service: RequireRole(User)
  ├── PUT  /profile                     // Service: RequireRole(User)
  ├── PUT  /profile/email               // Service: RequireRole(User)
  ├── POST /posts                       // Service: RequireRole(User)
  ├── DELETE /posts/:id                 // Service: 作者 or RequireRole(Moderator) + 节点管辖
  ├── PUT  /posts/:id/like              // Service: RequireRole(User)
  ├── GET  /feed                        // Service: RequireRole(User)
  ├── PUT  /users/:username/follow      // Service: RequireRole(User)
  ├── GET  /admin/users                 // Service: RequireRole(Admin)
  ├── GET  /admin/users/:id             // Service: RequireRole(Admin)
  ├── PUT  /admin/users/:id             // Service: RequireRole(Admin) + SuperAdmin 保护
  ├── DELETE /admin/users/:id           // Service: RequireRole(Admin) + SuperAdmin 保护
  └── DELETE /admin/posts/:id           // Service: RequireRole(Moderator) + 节点管辖
```

**核心变化**：路由层不再做权限区分，所有路由统一走 `AuthMiddleware`，权限判断全部在 Service 内部完成。

## 9. 配置变更

### 9.1 新增配置项

```yaml
# configs/config.yaml.example 新增
guest_jwt:
  secret: "guest-jwt-secret-at-least-16-chars"
  expire: 168h  # 7 天
```

### 9.2 Config 结构体

```go
type Config struct {
    // ...existing fields...
    GuestJWT GuestJWTConfig `yaml:"guest_jwt"`
}

type GuestJWTConfig struct {
    Secret string        `yaml:"secret"`
    Expire time.Duration `yaml:"expire"`
}
```

### 9.3 校验规则

```go
// 新增校验
if len(c.GuestJWT.Secret) < 16 {
    return fmt.Errorf("%w: guest_jwt secret 长度不能少于 16", ErrInvalidConfig)
}
if c.GuestJWT.Expire <= 0 {
    return fmt.Errorf("%w: guest_jwt expire 必须大于 0", ErrInvalidConfig)
}
```

## 10. 受影响文件清单

### 10.1 新增文件

| 文件 | 说明 |
|------|------|
| `internal/auth/role.go` | Role 类型定义、常量、方法 |
| `internal/auth/context.go` | UserContext 定义、Get/Set 工具函数 |
| `internal/auth/guest.go` | Guest JWT 签发逻辑 |
| `internal/model/node_moderator.go` | NodeModerator 模型 |
| `internal/repository/node_moderator_repo.go` | NodeModerator 仓库接口 |
| `internal/repository/node_moderator_repo_gorm.go` | NodeModerator GORM 实现 |
| `init-scripts/migrate-role-system.sql` | 数据库迁移脚本 |

### 10.2 重构文件

| 文件 | 变更内容 |
|------|---------|
| `internal/model/user.go` | `Role string` → `Role int` |
| `internal/config/config.go` | 新增 `GuestJWTConfig` |
| `internal/middleware/jwt.go` | 重写为 `AuthMiddleware`，删除 `JWTMiddleware`/`OptionalJWTMiddleware`/`AdminOnly` |
| `internal/middleware/ratelimit.go` | 支持 user_id/device_id 限流 |
| `internal/router/router.go` | 路由分组简化，新增 `/guest-token` |
| `internal/controller/user_controller.go` | 用 `GetUserContext()` 替代手动取值，新增 `GuestToken` handler |
| `internal/controller/post_controller.go` | 同上，传递 `UserContext` 到 Service |
| `internal/controller/follow_controller.go` | 同上 |
| `internal/controller/node_controller.go` | 同上 |
| `internal/controller/helpers.go` | 移除 `optionalUserID()`，改用 UserContext |
| `internal/controller/param/user_response.go` | Role 序列化适配 |
| `internal/service/user_service.go` | 接收 UserContext，角色校验，generateToken 适配 |
| `internal/service/user_input.go` | UpdateInput.Role 类型变更 |
| `internal/service/post_service.go` | 接收 UserContext，权限下沉 |
| `internal/service/follow_service.go` | 接收 UserContext |
| `internal/service/like_service.go` | 接收 UserContext |
| `internal/service/node_service.go` | 可能新增版主管理方法 |
| `cmd/user-system/main.go` | 依赖注入调整（新增 GuestJWT 配置、NodeModeratorRepo） |
| `configs/config.yaml.example` | 新增 `guest_jwt` 段 |

## 11. 实施阶段划分

### Phase 1: 基础设施

- 新建 `internal/auth/` 包：`role.go`、`context.go`、`guest.go`
- 新建 `internal/model/node_moderator.go`
- 新建 `internal/repository/node_moderator_repo.go` + gorm 实现
- 修改 `internal/config/config.go` 新增 GuestJWT 配置
- 修改 `configs/config.yaml.example`
- **此阶段不改变任何现有行为，仅新增代码**

### Phase 2: 中间件重构 + 路由简化

- 重写 `internal/middleware/jwt.go` → `AuthMiddleware`
- 重构 `internal/middleware/ratelimit.go`
- 简化 `internal/router/router.go`
- 新增 `POST /api/guest-token` 端点
- **此阶段改变中间件行为，需要同步调整 Controller 的取值方式**

### Phase 3: Service 层改造

- 所有 Service 方法签名改为接收 `*auth.UserContext`
- 权限判断从 Controller/中间件下沉到 Service
- 数据差异组装逻辑移入 Service
- **此阶段是核心改动，需要逐个 Service 仔细改造**

### Phase 4: Controller 层适配

- 所有 Controller 用 `auth.GetUserContext(c)` 替代手动取值
- 移除 Controller 中的权限判断逻辑
- 适配 param 层的 Role 序列化
- **此阶段是机械性改动，风险较低**

### Phase 5: 数据库迁移 + Model 适配 + 测试

- `model.User.Role` 类型变更
- 编写并测试迁移脚本
- 更新 `cmd/user-system/main.go` 依赖注入
- AutoMigrate 新增 NodeModerator
- 全量回归测试
- **此阶段涉及数据迁移，需要最谨慎**

## 12. 注意事项

### 12.1 Breaking Change

本次重构为 **Breaking Change**：
- JWT payload 中 `role` 字段从字符串变为数字
- 所有已签发的用户 JWT 在部署后失效，用户需要重新登录
- API 响应中 `role` 字段从字符串变为数字（或保持字符串兼容，看前端偏好）

### 12.2 向后兼容策略（可选）

如果不想强制所有用户重新登录，可以在 `AuthMiddleware` 中兼容旧格式 JWT：
- 解析 `role` 字段时，如果是 string 则尝试 `ParseRole()` 转换
- 如果是 float64 则直接转 int
- 这种兼容逻辑可在上线一段时间后移除

### 12.3 前端适配

- 登录/注册流程不变
- 需要实现 device_id 生成逻辑（fingerprintjs）
- 需要处理 `403` 错误码（之前主要是 `401`）
- 帖子列表中 `liked`/`followed` 字段：Guest 时为 `null`/`false`