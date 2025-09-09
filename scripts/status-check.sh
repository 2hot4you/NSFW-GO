#!/bin/bash

# NSFW-Go 服务状态检查脚本

echo "📊 NSFW-Go 服务状态检查"
echo "=========================="

# 检查 Docker 状态
echo "📦 Docker 服务:"
systemctl is-active docker && echo "  ✅ 运行中" || echo "  ❌ 已停止"

# 检查数据库服务
echo "🗄️ PostgreSQL 服务:"
if systemctl is-active --quiet nsfw-postgres; then
    echo "  ✅ 服务运行中"
    if docker exec nsfw-postgres pg_isready -U nsfw -d nsfw_db > /dev/null 2>&1; then
        echo "  ✅ 数据库连接正常"
    else
        echo "  ⚠️ 数据库连接异常"
    fi
else
    echo "  ❌ 服务已停止"
fi

# 检查 Redis 服务
echo "🔴 Redis 服务:"
if systemctl is-active --quiet nsfw-redis; then
    echo "  ✅ 服务运行中"
    if docker exec nsfw-redis redis-cli ping > /dev/null 2>&1; then
        echo "  ✅ Redis 连接正常"
    else
        echo "  ⚠️ Redis 连接异常"
    fi
else
    echo "  ❌ 服务已停止"
fi

# 检查主应用服务
echo "🎬 NSFW-Go 主应用:"
if systemctl is-active --quiet nsfw-go; then
    echo "  ✅ 服务运行中"
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health | grep -q "200"; then
        echo "  ✅ API 响应正常"
    else
        echo "  ⚠️ API 响应异常"
    fi
else
    echo "  ❌ 服务已停止"
fi

echo "=========================="

# 显示服务启动状态
echo "🔄 开机自启状态:"
systemctl is-enabled nsfw-postgres 2>/dev/null && echo "  ✅ PostgreSQL 自启已启用" || echo "  ❌ PostgreSQL 自启未启用"
systemctl is-enabled nsfw-redis 2>/dev/null && echo "  ✅ Redis 自启已启用" || echo "  ❌ Redis 自启未启用"  
systemctl is-enabled nsfw-go 2>/dev/null && echo "  ✅ NSFW-Go 自启已启用" || echo "  ❌ NSFW-Go 自启未启用"

# 显示资源使用情况
echo "💾 资源使用情况:"
if systemctl is-active --quiet nsfw-go; then
    PID=$(systemctl show -p MainPID nsfw-go | cut -d= -f2)
    if [ "$PID" != "0" ] && [ -n "$PID" ]; then
        MEM=$(ps -p $PID -o rss= 2>/dev/null | awk '{print int($1/1024)"MB"}')
        CPU=$(ps -p $PID -o pcpu= 2>/dev/null | awk '{print $1"%"}')
        echo "  📊 CPU: $CPU, 内存: $MEM"
    fi
fi