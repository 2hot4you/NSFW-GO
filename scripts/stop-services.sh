#!/bin/bash

# NSFW-Go 服务停止脚本

set -e

echo "🛑 停止 NSFW-Go 服务..."

# 停止主应用
echo "⏹️ 停止主应用..."
sudo systemctl stop nsfw-go || true

# 停止 Redis
echo "🔴 停止 Redis..."
sudo systemctl stop nsfw-redis || true

# 停止数据库
echo "🗄️ 停止数据库..."
sudo systemctl stop nsfw-postgres || true

echo "✅ 所有服务已停止！"