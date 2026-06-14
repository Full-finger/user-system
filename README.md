# user-system

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white) ![Vue](https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vuedotjs&logoColor=white) ![Echo](https://img.shields.io/badge/Echo-4.x-000000?logo=labstack&logoColor=white) ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white) ![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis&logoColor=white)

Go 用户系统，包含用户认证、帖子、节点、关注、点赞等功能。

## 前置依赖

- Go 1.26+
- Node.js 18+
- Docker & Docker Compose（用于 PostgreSQL + Redis）

## 快速启动

```bash
git clone git@github.com:Full-finger/user-system.git
cd user-system

make init          # 生成随机密钥的配置文件
make docker-up     # 启动 PostgreSQL + Redis
make run           # 启动后端（默认 :1323）

# 前端（另开终端）
make web           # 安装依赖（首次）
make web-dev       # 启动开发服务器（默认 :5173）
```

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

见 [docs/openapi.yaml](docs/openapi.yaml)，可用 Swagger UI 等工具查看。