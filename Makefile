# ── 项目配置 ────────────────────────────────────────────────
APP_NAME    := user-system
CMD_DIR     := ./cmd/$(APP_NAME)
BIN_DIR     := ./bin
BIN         := $(BIN_DIR)/$(APP_NAME)
CONFIG      := ./configs/config.yaml

# 前端
WEB_DIR     := ./web

# Docker
COMPOSE_FILE := ./deployments/docker-compose.yml

# Go
GOFMT       := gofmt -s -l
GOLINT      := go vet

# ── 颜色输出 ────────────────────────────────────────────────
CYAN  := \033[36m
GREEN := \033[32m
YELLOW:= \033[33m
RED   := \033[31m
RESET := \033[0m

define log
	@printf "$(CYAN)==> $(GREEN)$(1)$(RESET)\n"
endef

# ── Phony 声明 ─────────────────────────────────────────────
.PHONY: help run build dev init test lint fmt clean \
        web web-dev web-build web-clean \
        docker-up docker-down docker-logs \
        all

# ── 默认目标 ────────────────────────────────────────────────
help: ## 显示帮助信息
	@printf "$(CYAN)$(APP_NAME)$(RESET)  可用命令：\n\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-14s$(RESET) %s\n", $$1, $$2}'

# ── 后端 ────────────────────────────────────────────────────
run: $(CONFIG) ## 启动后端服务
	$(call log,启动后端服务...)
	go run $(CMD_DIR)/

build: ## 编译后端二进制
	$(call log,编译 $(APP_NAME)...)
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN) $(CMD_DIR)/

dev: $(CONFIG) ## 热重载开发模式（需要 air）
	$(call log,启动热重载开发模式...)
	@if command -v air > /dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "$(RED)air 未安装，请运行：go install github.com/air-verse/air@latest$(RESET)"; \
		exit 1; \
	fi

init: ## 生成安全随机配置文件
	$(call log,生成配置文件...)
	bash scripts/init.sh

test: ## 运行测试
	$(call log,运行测试...)
	go test -v -count=1 ./...

lint: ## 代码检查 (vet + fmt)
	$(call log,代码检查...)
	@$(GOLINT) ./...
	@files=$$($(GOFMT) .); \
	if [ -n "$$files" ]; then \
		echo "$(RED)以下文件需要格式化：$(RESET)"; \
		echo "$$files"; \
		exit 1; \
	fi
	@printf "$(GREEN)  ✓ lint 通过$(RESET)\n"

fmt: ## 格式化代码
	$(call log,格式化代码...)
	gofmt -s -w .

# ── 前端 ────────────────────────────────────────────────────
web: ## 安装前端依赖
	$(call log,安装前端依赖...)
	cd $(WEB_DIR) && npm install

web-dev: ## 启动前端开发服务器
	$(call log,启动前端开发服务器...)
	cd $(WEB_DIR) && npx vite --host

web-build: ## 构建前端生产包
	$(call log,构建前端生产包...)
	cd $(WEB_DIR) && npx vite build

web-clean: ## 清理前端构建产物
	$(call log,清理前端构建产物...)
	rm -rf $(WEB_DIR)/dist $(WEB_DIR)/node_modules

# ── Docker ──────────────────────────────────────────────────
docker-up: ## 启动 Docker 基础设施 (PostgreSQL + Redis)
	$(call log,启动 Docker 基础设施...)
	docker compose -f $(COMPOSE_FILE) up -d
	@printf "$(GREEN)  ✓ 服务已启动$(RESET)\n"

docker-down: ## 停止 Docker 基础设施
	$(call log,停止 Docker 基础设施...)
	docker compose -f $(COMPOSE_FILE) down

docker-logs: ## 查看 Docker 日志
	docker compose -f $(COMPOSE_FILE) logs -f

# ── 清理 ────────────────────────────────────────────────────
clean: ## 清理后端构建产物
	$(call log,清理构建产物...)
	rm -rf $(BIN_DIR)

cleanall: clean web-clean ## 清理全部构建产物
	$(call log,全部清理完毕)

# ── 组合 ────────────────────────────────────────────────────
all: lint build web-build ## 完整构建（lint + 后端编译 + 前端构建）
