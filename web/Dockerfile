# syntax=docker/dockerfile:1
# ── 前端镜像：Vite 构建 + nginx 托管 ────────────────────────
# 构建：docker build -t user-system-web ./web
# 注意：构建上下文为 web/ 目录（见 docker-compose 中的 context）

# ── Stage 1: 构建 ───────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /build

# 先拷依赖，利用层缓存
COPY package.json package-lock.json ./
RUN npm ci

# 拷源码并构建
COPY . .
RUN npm run build

# ── Stage 2: 运行 ───────────────────────────────────────────
FROM nginx:alpine

# 时区
RUN apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 拷贝自定义 nginx 配置（反代 /api + SPA fallback）
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 拷贝构建产物
COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1/ || exit 1