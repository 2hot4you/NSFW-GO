#!/bin/bash

# NSFW-Go 服务启动脚本
# 确保所有依赖服务按正确顺序启动

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "🚀 启动 NSFW-Go 服务..."

# 检查 Docker 服务
echo "📦 检查 Docker 服务..."
if ! systemctl is-active --quiet docker; then
    echo "⚠️ Docker 服务未运行，尝试启动..."
    sudo systemctl start docker
fi

# 启动数据库服务
echo "🗄️ 启动数据库服务..."
sudo systemctl start nsfw-postgres
sleep 5

# 启动 Redis 服务
echo "🔴 启动 Redis 服务..."
sudo systemctl start nsfw-redis
sleep 3

# 等待服务就绪
echo "⏳ 等待数据库就绪..."
for i in {1..30}; do
    if docker exec nsfw-postgres pg_isready -U nsfw -d nsfw_db > /dev/null 2>&1; then
        echo "✅ 数据库已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ 数据库启动超时"
        exit 1
    fi
    sleep 1
done

# 运行数据库迁移
echo "🔄 执行数据库迁移..."
cd "$PROJECT_DIR"
make migrate

# 启动主应用
echo "🎬 启动主应用..."
sudo systemctl start nsfw-go

echo "✅ 所有服务已启动完成！"
echo "🌐 访问地址: http://localhost:8080"