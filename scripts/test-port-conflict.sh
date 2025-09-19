#!/bin/bash

# 测试端口冲突检测功能的脚本

echo "🔍 端口冲突检测功能测试"
echo "================================"
echo ""

# 检查当前端口占用情况
echo "📊 当前端口占用情况:"
echo ""

ports=(5433 6380 8080)
for port in "${ports[@]}"; do
    if netstat -tlnp 2>/dev/null | grep -q ":$port "; then
        pid=$(netstat -tlnp 2>/dev/null | grep ":$port " | awk '{print $7}' | cut -d'/' -f1 | head -1)
        process=$(ps -p $pid -o comm= 2>/dev/null || echo "unknown")
        echo "✅ 端口 $port 被占用: PID $pid ($process)"
    else
        echo "❌ 端口 $port 空闲"
    fi
done

echo ""
echo "🚀 运行开发脚本测试端口冲突检测..."
echo ""

# 如果要测试端口冲突检测，需要启动另一个实例
# 这里只是演示如何检测