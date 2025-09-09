#!/bin/bash

# NSFW-Go 服务管理脚本
# 使用方法: ./nsfw-manager.sh [start|stop|restart|status|logs|install]

set -e

# 获取脚本实际路径（处理符号链接）
SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]}")"
SCRIPT_DIR="$(dirname "$SCRIPT_PATH")"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

show_usage() {
    echo "📋 NSFW-Go 服务管理器"
    echo "使用方法: $0 [命令]"
    echo ""
    echo "可用命令:"
    echo "  start    - 启动所有服务"
    echo "  stop     - 停止所有服务"  
    echo "  restart  - 重启所有服务"
    echo "  status   - 显示服务状态"
    echo "  logs     - 查看服务日志"
    echo "  install  - 安装并启用开机自启"
    echo "  build    - 重新构建应用"
    echo "  migrate  - 执行数据库迁移"
    echo "  help     - 显示帮助信息"
}

install_services() {
    echo "🔧 安装 NSFW-Go 服务..."
    
    # 构建应用
    echo "📦 构建应用..."
    cd "$PROJECT_DIR"
    make build
    
    # 重载 systemd
    echo "🔄 重载 systemd 配置..."
    sudo systemctl daemon-reload
    
    # 启用服务
    echo "🚀 启用开机自启..."
    sudo systemctl enable nsfw-postgres nsfw-redis nsfw-go
    
    # 启动服务
    echo "▶️ 启动服务..."
    start_services
    
    echo "✅ 安装完成！服务已启动并设置为开机自启"
}

start_services() {
    echo "🚀 启动 NSFW-Go 服务..."
    
    sudo systemctl start docker || echo "⚠️ Docker 可能已在运行"
    sleep 2
    
    sudo systemctl start nsfw-postgres
    echo "⏳ 等待数据库启动..."
    sleep 8
    
    sudo systemctl start nsfw-redis
    sleep 3
    
    # 检查数据库连接
    for i in {1..30}; do
        if docker exec nsfw-postgres pg_isready -U nsfw -d nsfw_db > /dev/null 2>&1; then
            echo "✅ 数据库已就绪"
            break
        fi
        if [ $i -eq 30 ]; then
            echo "❌ 数据库连接超时"
            exit 1
        fi
        sleep 1
    done
    
    sudo systemctl start nsfw-go
    sleep 3
    
    echo "✅ 所有服务已启动！"
    echo "🌐 访问地址: http://localhost:8080"
}

stop_services() {
    echo "🛑 停止 NSFW-Go 服务..."
    sudo systemctl stop nsfw-go || true
    sudo systemctl stop nsfw-redis || true  
    sudo systemctl stop nsfw-postgres || true
    echo "✅ 所有服务已停止"
}

restart_services() {
    echo "🔄 重启 NSFW-Go 服务..."
    stop_services
    sleep 3
    start_services
}

show_status() {
    "$SCRIPT_DIR/status-check.sh"
}

show_logs() {
    echo "📋 选择要查看的日志:"
    echo "1) NSFW-Go 主应用"
    echo "2) PostgreSQL"
    echo "3) Redis"
    echo "4) 全部服务"
    read -p "请选择 [1-4]: " choice
    
    case $choice in
        1)
            echo "📄 NSFW-Go 主应用日志:"
            journalctl -u nsfw-go -f --no-pager
            ;;
        2)
            echo "📄 PostgreSQL 日志:"
            journalctl -u nsfw-postgres -f --no-pager
            ;;
        3)
            echo "📄 Redis 日志:"
            journalctl -u nsfw-redis -f --no-pager
            ;;
        4)
            echo "📄 所有服务日志:"
            journalctl -u nsfw-go -u nsfw-postgres -u nsfw-redis -f --no-pager
            ;;
        *)
            echo "❌ 无效选择"
            exit 1
            ;;
    esac
}

build_app() {
    echo "🔨 重新构建应用..."
    cd "$PROJECT_DIR"
    make build
    echo "✅ 构建完成"
    
    echo "🔄 重启服务以使用新版本..."
    sudo systemctl restart nsfw-go
    echo "✅ 服务已重启"
}

migrate_db() {
    echo "🗄️ 执行数据库迁移..."
    cd "$PROJECT_DIR"
    make migrate
    echo "✅ 迁移完成"
}

# 主逻辑
case "${1:-}" in
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    install)
        install_services
        ;;
    build)
        build_app
        ;;
    migrate)
        migrate_db
        ;;
    help|--help|-h)
        show_usage
        ;;
    "")
        show_usage
        ;;
    *)
        echo "❌ 未知命令: $1"
        show_usage
        exit 1
        ;;
esac