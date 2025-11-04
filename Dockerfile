# ============================================
# 多阶段构建 Dockerfile for NSFW-Go
# ============================================

# ============================================
# 阶段 1: 构建器
# ============================================
FROM golang:1.23-alpine AS builder

# 构建参数
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# 安装必要的工具
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    make \
    upx

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件（利用 Docker 缓存）
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建 API 服务
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix cgo \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o bin/nsfw-go-api \
    cmd/api/main.go

# 压缩二进制文件（可选，可减少 50% 大小）
RUN upx --best --lzma bin/nsfw-go-api || true

# 构建 Bot 服务 (如果存在)
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/nsfw-go-bot cmd/bot/main.go

# ============================================
# 阶段 2: 开发环境镜像
# ============================================
FROM alpine:latest AS development

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    wget

# 设置时区（可通过环境变量覆盖）
ARG TZ=Asia/Shanghai
RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo "${TZ}" > /etc/timezone

WORKDIR /app

# 从构建器复制二进制文件
COPY --from=builder /app/bin/nsfw-go-api .

# 复制 Web 静态文件
COPY --from=builder /app/web/dist ./web/dist

# 复制配置文件模板
COPY --from=builder /app/configs ./configs

# 创建必要的目录
RUN mkdir -p logs media

# 设置文件权限
RUN chmod +x ./nsfw-go-api

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动命令
CMD ["./nsfw-go-api"]

# Bot 服务镜像
FROM alpine:latest AS bot

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从构建器复制二进制文件（如果 bot 存在的话）
# COPY --from=builder /app/bin/nsfw-go-bot .

# 创建必要的目录
RUN mkdir -p logs

# 启动命令（暂时使用 sleep，因为 bot 可能还没实现）
CMD ["sleep", "infinity"]

# ============================================
# 阶段 4: 生产环境镜像
# ============================================
FROM alpine:latest AS production

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    wget

# 设置时区（可通过环境变量覆盖）
ARG TZ=Asia/Shanghai
RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo "${TZ}" > /etc/timezone

# 创建非 root 用户
RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup -h /app

WORKDIR /app

# 从构建器复制二进制文件
COPY --from=builder /app/bin/nsfw-go-api .

# 复制 Web 静态文件
COPY --from=builder /app/web/dist ./web/dist

# 复制配置文件模板
COPY --from=builder /app/configs ./configs

# 创建必要的目录并设置权限
RUN mkdir -p logs media && \
    chown -R appuser:appgroup /app && \
    chmod +x ./nsfw-go-api

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 元数据标签
LABEL maintainer="NSFW-GO Team" \
      description="NSFW-GO API Service - Production Image" \
      version="${VERSION}"

# 启动命令
CMD ["./nsfw-go-api"]
