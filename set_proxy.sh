#!/bin/bash

# 🌐 代理设置脚本
# 用于设置系统代理环境变量

# 代理服务器配置
PROXY_SERVER="http://naffly:pangmaom1@home.cloudbp.com:17890/"

echo "🚀 正在设置系统代理..."

# 设置代理环境变量
export http_proxy="$PROXY_SERVER"
export https_proxy="$PROXY_SERVER"
export ftp_proxy="$PROXY_SERVER"
export no_proxy="127.0.0.1,localhost,.local"

# 同时设置大写版本（某些应用需要）
export HTTP_PROXY="$PROXY_SERVER"
export HTTPS_PROXY="$PROXY_SERVER"
export FTP_PROXY="$PROXY_SERVER"
export NO_PROXY="127.0.0.1,localhost,.local"

echo "✅ 代理设置完成！"
echo "📍 代理服务器: $PROXY_SERVER"
echo "🚫 无代理地址: 127.0.0.1,localhost,.local"

# 验证代理设置
echo ""
echo "🔍 当前代理设置："
echo "HTTP_PROXY: $HTTP_PROXY"
echo "HTTPS_PROXY: $HTTPS_PROXY"
echo "FTP_PROXY: $FTP_PROXY"
echo "NO_PROXY: $NO_PROXY"

echo ""
echo "💡 使用方法："
echo "   source ./set_proxy.sh    # 在当前终端生效"
echo "   ./set_proxy.sh          # 仅在脚本执行时生效"