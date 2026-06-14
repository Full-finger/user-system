# user-system

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white) ![Vue](https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vuedotjs&logoColor=white) ![Echo](https://img.shields.io/badge/Echo-4.x-000000?logo=labstack&logoColor=white) ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white) ![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis&logoColor=white)

Go 用户系统，包含用户认证、帖子、节点、关注、点赞、评论、@提及、版主与管理后台等功能。

## 功能概览

- **用户认证**：用户名/邮箱密码登录、邮箱验证码登录、注册、绑定邮箱、游客 JWT
- **帖子与节点**：发帖、节点分类、点赞、关注时间线（Feed）
- **评论与回复**：楼中楼回复、评论点赞、@提及补全、PoW 反垃圾
- **社交关系**：关注/取消关注、粉丝列表、用户主页
- **角色体系**：6 级角色（Guest → User → VerifiedUser → Moderator → Admin → SuperAdmin），权限判断下沉到 Service 层
- **管理后台**：数据概览、用户/帖子/评论/节点管理、任命版主

> 角色体系与鉴权流程详见 [docs/role-system.md](docs/role-system.md)。

### 鉴权与限流

- 所有 API 统一走 `AuthMiddleware`：无 Token 自动赋予 Guest 身份，解析失败降级为 Guest。
- 权限判断全部在 Service 层通过 `UserContext.RequireRole()` 完成，路由层不再做权限区分。
- 限流优先级：`user_id > device_id > IP`（基于 Redis），游客可通过 `X-Device-ID` 头或 `POST /api/guest-token` 获取 Guest JWT 实现跨请求追踪。

## 前置依赖

- Go 1.26+
- Node.js 18+
- Docker & Docker Compose（用于 PostgreSQL + Redis）

## 快速启动（本地开发）

```bash
git clone git@github.com:Full-finger/user-system.git
cd user-system

make init          # 生成随机密钥的配置文件
# 仅启动基础设施 db + redis（全栈 docker-up 见下方"Docker 部署"）
docker compose -f deployments/docker-compose.yml up -d db redis
make run           # 启动后端（默认 :1323）

# 前端（另开终端）
make web           # 安装依赖（首次）
make web-dev       # 启动开发服务器（默认 :5173）
```

## Docker 部署（全栈一键）

适用于生产或完整联调，一键拉起 `app`(后端) + `web`(nginx) + `db`(PostgreSQL) + `redis`(Redis)。

```bash
# 1. 准备配置（从模板复制）
cp deployments/docker-compose.yml.example deployments/docker-compose.yml
cp deployments/config.yaml.example       deployments/config.yaml

# 2. 修改 deployments/config.yaml 的所有 CHANGE_ME_*（JWT/SMTP/Admin 等只此一处）
$EDITOR deployments/config.yaml

# 3. 修改 deployments/docker-compose.yml 的 DB/Redis 凭据，
#    且必须与 config.yaml 中的 database / redis 段落完全一致，否则 app 连不上
$EDITOR deployments/docker-compose.yml

# 4. 启动（首次会自动构建后端、前端镜像）
make docker-up

# 5. 访问
#    前端入口: http://localhost:8080
#    健康检查: curl http://localhost:8080/health

# 常用维护
make docker-logs      # 查看日志
make docker-down      # 停止
```

说明：
- ⚠️ **DB/Redis 凭据在 `docker-compose.yml` 和 `config.yaml` 中各有一份，两处必须完全一致**，否则 `app` 容器会因连不上数据库而反复重启。
- 前端 nginx 反代 `/api` 到后端 `app:1323`，前后端**同源**，无需配置 CORS。
- 后端 `:1323`、数据库 `:5432`、Redis `:6379` 默认只在 compose 内部网络通信；如需从宿主机调试，解开 `deployments/docker-compose.yml` 中对应 `ports` 注释。
- `server.env: production` 时后端会跳过 `AutoMigrate`。首次部署如需自动建表，可临时改为 `development` 启动一次，或在 `deployments/init-scripts/` 放置 SQL 脚本（compose 已配置挂载点，默认注释）。
- 仅构建镜像不启动：`make docker-build`（或 `make docker-app` / `make docker-web` 分开构建）。

### 镜像结构

| 镜像 | Dockerfile | 说明 |
|------|-----------|------|
| `user-system-app` | `Dockerfile`（根目录） | Go 多阶段构建，`CGO_ENABLED=0` 静态编译，alpine runtime，非 root 用户 |
| `user-system-web` | `web/Dockerfile` | Vite 构建 → nginx 托管，含 `/api` 反代与 SPA fallback（`web/nginx.conf`） |

## 常用命令

```bash
make help          # 查看所有可用命令
make dev           # 热重载开发模式（需要 air）
make test          # 运行测试
make lint          # 代码检查（go vet + gofmt）
make arch-check    # 分层架构合规性检查（Semgrep + Bash）
make semgrep       # 仅运行 Semgrep 规则
make build         # 编译后端
make all           # 完整构建（lint + 后端 + 前端）
make cleanall      # 清理全部构建产物
```

## 架构合规性检查

```bash
make arch-check    # 生成 docs/arch-report.md 报告
```

工具链分三层：
- **Layer 1 — Semgrep**：AST 级模式匹配（`scripts/semgrep-rules/*.yml`）
- **Layer 2 — Bash**：结构性检查（返回值计数、文件名匹配、多写事务等）
- **Layer 3 — go vet**：编译器级检查（`make lint`）

> 依赖：`pip install semgrep` + `sudo apt install jq`

## API 文档

见 [docs/openapi.yaml](docs/openapi.yaml)（OpenAPI 3.0，覆盖 40+ 端点），可用 Swagger UI、Redoc 等工具查看。

设计文档：
- [角色系统](docs/role-system.md) — 6 级角色层级、UserContext、Guest JWT、鉴权流程
- [管理后台](docs/admin-system.md) — 后台 Tab 页与 Admin API 设计
