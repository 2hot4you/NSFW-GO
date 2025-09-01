# 多阶段构建 Dockerfile for NSFW-Go
FROM golang:1.23-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建 API 服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/nsfw-go-api cmd/api/main.go

# 构建 Bot 服务 (如果存在)
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/nsfw-go-bot cmd/bot/main.go

# 开发环境镜像
FROM alpine:latest AS development

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata curl

# 设置时区
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从构建器复制二进制文件
COPY --from=builder /app/bin/nsfw-go-api .
COPY --from=builder /app/web/dist ./web/dist
COPY --from=builder /app/configs ./configs

# 创建必要的目录
RUN mkdir -p logs media

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

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

# 生产环境镜像
FROM alpine:latest AS production

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata curl

# 设置时区
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

# 创建非root用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# 从构建器复制二进制文件
COPY --from=builder /app/bin/nsfw-go-api .
COPY --from=builder /app/web/dist ./web/dist
COPY --from=builder /app/configs ./configs

# 创建必要的目录并设置权限
RUN mkdir -p logs media && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# 启动命令
CMD ["./nsfw-go-api"]
