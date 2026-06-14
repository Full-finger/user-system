# syntax=docker/dockerfile:1
# ── 后端镜像：Go 多阶段构建 ─────────────────────────────────
# 构建：docker build -t user-system-app .
# 国内网络默认走 goproxy.cn；国外可：
#   docker build --build-arg GOPROXY=https://proxy.golang.org,direct -t user-system-app .

# ── Stage 1: 构建 ───────────────────────────────────────────
FROM golang:1.26-alpine AS builder

# git：go mod 可能需要；tzdata：保证编译期时区数据库可用
RUN apk add --no-cache git tzdata

WORKDIR /src

# Go 模块代理（国内构建默认走 goproxy.cn；可 --build-arg 覆盖）
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

# 先拷依赖，利用层缓存
COPY go.mod go.sum ./
RUN go mod download

# 拷源码
COPY . .

# 静态编译（CGO_ENABLED=0），去掉调试信息（-s -w）
# 生产只用 PostgreSQL，不需要 SQLite/cgo
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/user-system \
    ./cmd/user-system

# ── Stage 2: 运行 ───────────────────────────────────────────
FROM alpine:3

# ca-certificates：HTTPS（SMTP/TLS 等）；tzdata：时区；wget：健康检查
RUN apk add --no-cache ca-certificates tzdata wget \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 非 root 用户
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# 拷贝二进制
COPY --from=builder /out/user-system /app/user-system

# 兜底默认配置（实际运行请通过 volume 挂载覆盖）
COPY configs/config.yaml.example /app/configs/config.yaml

# 日志目录
RUN mkdir -p /app/logs && chown -R app:app /app
USER app

EXPOSE 1323

# 健康检查（/health 端点，由 router.go 提供）
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:1323/health || exit 1

ENTRYPOINT ["/app/user-system"]
CMD ["-config", "/app/configs/config.yaml"]