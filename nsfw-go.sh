#!/bin/bash

# NSFW-Go 项目管理脚本
# 作者: NSFW-Go Team
# 版本: 1.0.0

set -e

# 自动检测项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"

# 切换到项目目录
cd "$PROJECT_ROOT"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置项
POSTGRES_PORT=5433
REDIS_PORT=6380
API_PORT=8080
PROJECT_NAME="NSFW-Go"

# 进程ID文件
PIDFILE="/tmp/nsfw-go-dev.pid"

# 函数：打印带颜色的信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${PURPLE}================================${NC}"
    echo -e "${PURPLE} 🚀 $PROJECT_NAME 管理脚本${NC}"
    echo -e "${PURPLE}================================${NC}"
}

# 函数：检查端口是否被占用
check_port() {
    local port=$1
    local service=$2
    
    if netstat -tlnp 2>/dev/null | grep -q ":$port "; then
        local pid=$(netstat -tlnp 2>/dev/null | grep ":$port " | awk '{print $7}' | cut -d'/' -f1 | head -1)
        if [ ! -z "$pid" ]; then
            echo -e "${GREEN}✅${NC} $service (端口 $port) - 运行中 [PID: $pid]"
            return 0
        else
            echo -e "${YELLOW}⚠️${NC}  $service (端口 $port) - 端口被占用但无法获取进程信息"
            return 1
        fi
    else
        echo -e "${RED}❌${NC} $service (端口 $port) - 未运行"
        return 1
    fi
}

# 函数：检查Docker容器状态
check_docker_container() {
    local container_name=$1
    local service_name=$2
    
    if docker ps --filter "name=$container_name" --filter "status=running" --format "table {{.Names}}" | grep -q "$container_name"; then
        local status=$(docker ps --filter "name=$container_name" --format "table {{.Status}}" | tail -1)
        echo -e "${GREEN}✅${NC} $service_name 容器 - 运行中 ($status)"
        return 0
    elif docker ps -a --filter "name=$container_name" --format "table {{.Names}}" | grep -q "$container_name"; then
        local status=$(docker ps -a --filter "name=$container_name" --format "table {{.Status}}" | tail -1)
        echo -e "${RED}❌${NC} $service_name 容器 - 已停止 ($status)"
        return 1
    else
        echo -e "${YELLOW}⚠️${NC}  $service_name 容器 - 不存在"
        return 1
    fi
}

# 函数：检查Go开发服务器
check_go_dev() {
    if [ -f "$PIDFILE" ]; then
        local pid=$(cat "$PIDFILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "${GREEN}✅${NC} Go 开发服务器 - 运行中 [PID: $pid]"
            return 0
        else
            echo -e "${RED}❌${NC} Go 开发服务器 - PID文件存在但进程已死"
            rm -f "$PIDFILE"
            return 1
        fi
    else
        # 检查Air进程
        if pgrep -f "air" >/dev/null 2>&1; then
            local pid=$(pgrep -f "air" | head -1)
            echo -e "${GREEN}✅${NC} Go 开发服务器 - 运行中 [PID: $pid] (Air热重载)"
            return 0
        # 检查make dev进程
        elif pgrep -f "make dev" >/dev/null 2>&1; then
            local pid=$(pgrep -f "make dev" | head -1)
            echo -e "${GREEN}✅${NC} Go 开发服务器 - 运行中 [PID: $pid] (make dev)"
            return 0
        # 检查nsfw-go进程
        elif pgrep -f "nsfw-go" >/dev/null 2>&1; then
            local pid=$(pgrep -f "nsfw-go" | head -1)
            echo -e "${GREEN}✅${NC} Go 开发服务器 - 运行中 [PID: $pid]"
            return 0
        # 最后检查端口
        elif check_port $API_PORT "API服务器" >/dev/null 2>&1; then
            echo -e "${GREEN}✅${NC} Go 开发服务器 - 运行中 (端口监听检测)"
            return 0
        else
            echo -e "${RED}❌${NC} Go 开发服务器 - 未运行"
            return 1
        fi
    fi
}

# 函数：显示服务状态
status() {
    print_header
    echo ""
    print_info "检查服务状态..."
    echo ""
    
    # 检查Docker服务
    echo -e "${CYAN}📦 Docker 服务状态:${NC}"
    check_docker_container "nsfw-postgres-dev" "PostgreSQL"
    check_docker_container "nsfw-redis-dev" "Redis"
    echo ""
    
    # 检查端口状态
    echo -e "${CYAN}🔌 端口监听状态:${NC}"
    check_port $POSTGRES_PORT "PostgreSQL"
    check_port $REDIS_PORT "Redis"
    check_port $API_PORT "API服务器"
    echo ""
    
    # 检查Go开发服务器
    echo -e "${CYAN}🚀 应用服务状态:${NC}"
    check_go_dev
    echo ""
    
    # 显示访问地址
    echo -e "${CYAN}🌐 访问地址:${NC}"
    echo -e "  • 主页: ${GREEN}http://localhost:$API_PORT${NC}"
    echo -e "  • 搜索页面: ${GREEN}http://localhost:$API_PORT/search.html${NC}"
    echo -e "  • 本地影片: ${GREEN}http://localhost:$API_PORT/local-movies.html${NC}"
    echo -e "  • 排行榜: ${GREEN}http://localhost:$API_PORT/rankings.html${NC}"
    echo -e "  • 配置页面: ${GREEN}http://localhost:$API_PORT/config.html${NC}"
    echo -e "  • API统计: ${GREEN}http://localhost:$API_PORT/api/v1/stats${NC}"
    echo -e "  • 健康检查: ${GREEN}http://localhost:$API_PORT/health${NC}"
    echo ""
}

# 函数：启动服务
start() {
    print_header
    echo ""
    print_info "启动开发环境..."
    
    # 启动Docker服务
    print_info "启动 PostgreSQL 和 Redis 服务..."
    if ! docker compose -f docker-compose.dev.yml up -d; then
        print_error "Docker服务启动失败"
        exit 1
    fi
    
    # 等待服务就绪
    print_info "等待服务启动完成..."
    sleep 5
    
    # 检查Docker服务状态
    if ! check_docker_container "nsfw-postgres-dev" "PostgreSQL" >/dev/null; then
        print_error "PostgreSQL 启动失败"
        exit 1
    fi
    
    if ! check_docker_container "nsfw-redis-dev" "Redis" >/dev/null; then
        print_error "Redis 启动失败"
        exit 1
    fi
    
    print_success "Docker服务启动成功"
    
    # 启动Go开发服务器
    if check_go_dev >/dev/null 2>&1; then
        print_warning "Go 开发服务器已在运行"
    else
        print_info "启动 Go 开发服务器..."
        nohup make dev > /tmp/nsfw-go-dev.log 2>&1 &
        echo $! > "$PIDFILE"
        
        # 等待API服务器启动
        print_info "等待 API 服务器启动..."
        for i in {1..30}; do
            if check_port $API_PORT "API服务器" >/dev/null 2>&1; then
                print_success "Go 开发服务器启动成功"
                break
            fi
            sleep 2
            if [ $i -eq 30 ]; then
                print_error "Go 开发服务器启动超时"
                stop_dev_server
                exit 1
            fi
        done
    fi
    
    echo ""
    print_success "🎉 所有服务启动完成！"
    echo ""
    status
}

# 函数：停止开发服务器
stop_dev_server() {
    if [ -f "$PIDFILE" ]; then
        local pid=$(cat "$PIDFILE")
        if kill -0 "$pid" 2>/dev/null; then
            print_info "停止 Go 开发服务器 [PID: $pid]..."
            kill "$pid"
            rm -f "$PIDFILE"
        else
            rm -f "$PIDFILE"
        fi
    fi
    
    # 强制停止Air进程
    if pgrep -f "air" >/dev/null 2>&1; then
        print_info "停止 Air 热重载进程..."
        pkill -f "air"
    fi
    
    # 强制停止可能的API服务器进程
    if pgrep -f "nsfw-go" >/dev/null 2>&1; then
        print_info "停止 API 服务器进程..."
        pkill -f "nsfw-go"
    fi
}

# 函数：停止服务
stop() {
    print_header
    echo ""
    print_info "停止开发环境..."
    
    # 停止Go开发服务器
    stop_dev_server
    
    # 停止Docker服务
    print_info "停止 Docker 服务..."
    docker compose -f docker-compose.dev.yml down
    
    print_success "✅ 所有服务已停止"
}

# 函数：重启服务
restart() {
    print_header
    echo ""
    print_info "重启开发环境..."
    
    stop
    echo ""
    sleep 2
    start
}

# 函数：显示日志
logs() {
    print_header
    echo ""
    print_info "显示服务日志..."
    echo ""
    
    # 显示选项
    echo "选择要查看的日志:"
    echo "  1) Go 开发服务器日志"
    echo "  2) PostgreSQL 日志"
    echo "  3) Redis 日志"
    echo "  4) 所有 Docker 服务日志"
    echo "  5) 实时跟踪 Go 开发日志"
    echo ""
    
    read -p "请输入选项 (1-5): " choice
    echo ""
    
    case $choice in
        1)
            if [ -f "/tmp/nsfw-go-dev.log" ]; then
                print_info "显示 Go 开发服务器日志:"
                echo "----------------------------------------"
                tail -50 /tmp/nsfw-go-dev.log
            elif pgrep -f "make dev" >/dev/null 2>&1; then
                print_info "显示 Go 开发服务器实时日志 (最近50行):"
                echo "----------------------------------------"
                print_warning "开发模式正在前台运行，显示进程实时输出..."
                echo "提示: 可以通过 Ctrl+Z 暂停进程，然后用 'fg' 命令恢复前台运行"
                echo ""
                # 显示make dev进程的输出 (如果可能的话)
                print_info "当前开发服务器正在运行，PID: $(pgrep -f 'make dev' | head -1)"
            else
                print_warning "Go 开发服务器日志文件不存在且服务器未运行"
            fi
            ;;
        2)
            print_info "显示 PostgreSQL 日志:"
            echo "----------------------------------------"
            docker logs nsfw-postgres-dev --tail=50
            ;;
        3)
            print_info "显示 Redis 日志:"
            echo "----------------------------------------"
            docker logs nsfw-redis-dev --tail=50
            ;;
        4)
            print_info "显示所有 Docker 服务日志:"
            echo "----------------------------------------"
            docker compose -f docker-compose.dev.yml logs --tail=50
            ;;
        5)
            if [ -f "/tmp/nsfw-go-dev.log" ]; then
                print_info "实时跟踪 Go 开发日志 (Ctrl+C 退出):"
                echo "----------------------------------------"
                tail -f /tmp/nsfw-go-dev.log
            else
                print_warning "Go 开发服务器日志文件不存在"
            fi
            ;;
        *)
            print_error "无效选项"
            exit 1
            ;;
    esac
}

# 函数：显示帮助信息
help() {
    print_header
    echo ""
    echo "用法: $0 {start|stop|restart|status|logs|help}"
    echo ""
    echo "命令说明:"
    echo "  start    - 启动所有开发服务 (PostgreSQL, Redis, Go开发服务器)"
    echo "  stop     - 停止所有开发服务"
    echo "  restart  - 重启所有开发服务"
    echo "  status   - 显示服务状态和端口信息"
    echo "  logs     - 显示服务日志 (交互式选择)"
    echo "  help     - 显示此帮助信息"
    echo ""
    echo "端口配置:"
    echo "  • PostgreSQL: $POSTGRES_PORT"
    echo "  • Redis: $REDIS_PORT"
    echo "  • API服务器: $API_PORT"
    echo ""
    echo "示例:"
    echo "  $0 start     # 启动开发环境"
    echo "  $0 status    # 检查服务状态"
    echo "  $0 logs      # 查看日志"
    echo ""
}

# 主函数
main() {
    case "${1:-help}" in
        start)
            start
            ;;
        stop)
            stop
            ;;
        restart)
            restart
            ;;
        status)
            status
            ;;
        logs)
            logs
            ;;
        help|--help|-h)
            help
            ;;
        *)
            print_error "未知命令: $1"
            echo ""
            help
            exit 1
            ;;
    esac
}

# 检查依赖命令
check_dependencies() {
    local missing_deps=()
    
    if ! command -v docker >/dev/null 2>&1; then
        missing_deps+=("docker")
    fi
    
    if ! command -v netstat >/dev/null 2>&1; then
        missing_deps+=("netstat")
    fi
    
    if ! command -v make >/dev/null 2>&1; then
        missing_deps+=("make")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "缺少必要依赖: ${missing_deps[*]}"
        echo "请安装缺少的命令后再运行此脚本"
        exit 1
    fi
}

# 脚本入口
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    check_dependencies
    main "$@"
fi