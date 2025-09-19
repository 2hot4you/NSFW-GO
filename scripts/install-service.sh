#!/bin/bash

# NSFW-Go SystemD 服务安装脚本
# 用于将 NSFW-Go 部署为系统服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICE_NAME="nsfw-go"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
BINARY_PATH="/projects/NSFW-GO/bin/nsfw-go"
CONFIG_PATH="/projects/NSFW-GO/config.yaml"

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
    echo -e "${PURPLE} 🚀 NSFW-Go 服务安装脚本${NC}"
    echo -e "${PURPLE}================================${NC}"
}

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "请使用 sudo 运行此脚本"
        exit 1
    fi
}

# 检查依赖
check_dependencies() {
    local missing_deps=()

    if ! command -v systemctl >/dev/null 2>&1; then
        missing_deps+=("systemd")
    fi

    if ! command -v netstat >/dev/null 2>&1; then
        missing_deps+=("netstat")
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "缺少必要依赖: ${missing_deps[*]}"
        exit 1
    fi
}

# 检查端口冲突
check_port_conflicts() {
    local conflicts=()
    local api_port=8080
    local postgres_port=5433
    local redis_port=6380

    print_info "检查端口冲突..."

    if netstat -tlnp 2>/dev/null | grep -q ":$api_port "; then
        local pid=$(netstat -tlnp 2>/dev/null | grep ":$api_port " | awk '{print $7}' | cut -d'/' -f1 | head -1)
        local process=$(ps -p $pid -o comm= 2>/dev/null || echo "unknown")
        conflicts+=("API服务器:$api_port:$pid:$process")
    fi

    if netstat -tlnp 2>/dev/null | grep -q ":$postgres_port "; then
        local pid=$(netstat -tlnp 2>/dev/null | grep ":$postgres_port " | awk '{print $7}' | cut -d'/' -f1 | head -1)
        local process=$(ps -p $pid -o comm= 2>/dev/null || echo "unknown")
        conflicts+=("PostgreSQL:$postgres_port:$pid:$process")
    fi

    if netstat -tlnp 2>/dev/null | grep -q ":$redis_port "; then
        local pid=$(netstat -tlnp 2>/dev/null | grep ":$redis_port " | awk '{print $7}' | cut -d'/' -f1 | head -1)
        local process=$(ps -p $pid -o comm= 2>/dev/null || echo "unknown")
        conflicts+=("Redis:$redis_port:$pid:$process")
    fi

    if [ ${#conflicts[@]} -eq 0 ]; then
        print_success "所有端口可用"
        return 0
    fi

    print_warning "发现端口冲突:"
    for conflict in "${conflicts[@]}"; do
        IFS=':' read -r service port pid process <<< "$conflict"
        echo -e "  • $service (端口 $port): PID $pid ($process)"
    done
    echo ""

    echo -e "${YELLOW}选择处理方式:${NC}"
    echo "  1) 自动终止所有冲突进程"
    echo "  2) 忽略冲突，继续安装"
    echo "  3) 退出安装"
    echo ""
    read -p "请选择 (1-3): " choice

    case $choice in
        1)
            print_info "自动处理端口冲突..."
            for conflict in "${conflicts[@]}"; do
                IFS=':' read -r service port pid process <<< "$conflict"
                print_info "终止 $service 进程: PID $pid ($process)"
                if kill $pid 2>/dev/null; then
                    sleep 1
                    if kill -0 $pid 2>/dev/null; then
                        kill -9 $pid 2>/dev/null
                    fi
                    print_success "$service 进程已终止"
                else
                    print_error "无法终止 $service 进程"
                fi
            done
            return 0
            ;;
        2)
            print_warning "忽略端口冲突，继续安装"
            return 0
            ;;
        3)
            print_info "退出安装"
            exit 0
            ;;
        *)
            print_error "无效选择，退出安装"
            exit 1
            ;;
    esac
}

# 构建二进制文件
build_binary() {
    print_info "构建生产版本二进制文件..."

    cd "$PROJECT_ROOT"

    # 创建 bin 目录
    mkdir -p bin

    # 构建二进制文件
    if ! make build; then
        print_error "构建失败"
        exit 1
    fi

    # 检查二进制文件是否存在
    if [ ! -f "$BINARY_PATH" ]; then
        print_error "二进制文件未找到: $BINARY_PATH"
        exit 1
    fi

    # 设置执行权限
    chmod +x "$BINARY_PATH"

    print_success "二进制文件构建完成: $BINARY_PATH"
}

# 安装服务
install_service() {
    print_info "安装 SystemD 服务..."

    # 检查配置文件
    if [ ! -f "$CONFIG_PATH" ]; then
        print_warning "配置文件不存在: $CONFIG_PATH"
        print_info "复制示例配置文件..."
        cp "${PROJECT_ROOT}/configs/config.example.yaml" "$CONFIG_PATH"
    fi

    # 复制服务文件
    cp "${PROJECT_ROOT}/scripts/nsfw-go.service" "$SERVICE_FILE"

    # 重新加载 systemd
    systemctl daemon-reload

    # 启用服务（开机自启）
    systemctl enable "$SERVICE_NAME"

    print_success "服务安装完成"
}

# 创建必要的目录和权限
setup_directories() {
    print_info "设置目录权限..."

    # 创建日志目录
    mkdir -p "${PROJECT_ROOT}/logs"
    mkdir -p "${PROJECT_ROOT}/data"

    # 设置权限（根据需要调整）
    chown -R root:root "$PROJECT_ROOT"
    chmod -R 755 "$PROJECT_ROOT"
    chmod -R 644 "${PROJECT_ROOT}/logs"
    chmod -R 644 "${PROJECT_ROOT}/data"

    print_success "目录权限设置完成"
}

# 启动服务
start_service() {
    print_info "启动 NSFW-Go 服务..."

    if systemctl start "$SERVICE_NAME"; then
        print_success "服务启动成功"

        # 检查服务状态
        sleep 3
        if systemctl is-active --quiet "$SERVICE_NAME"; then
            print_success "服务运行正常"
        else
            print_error "服务启动后状态异常"
            print_info "查看服务状态: sudo systemctl status $SERVICE_NAME"
            print_info "查看服务日志: sudo journalctl -u $SERVICE_NAME"
        fi
    else
        print_error "服务启动失败"
        exit 1
    fi
}

# 显示服务信息
show_service_info() {
    print_info "服务管理命令:"
    echo -e "  启动服务: ${GREEN}sudo systemctl start $SERVICE_NAME${NC}"
    echo -e "  停止服务: ${GREEN}sudo systemctl stop $SERVICE_NAME${NC}"
    echo -e "  重启服务: ${GREEN}sudo systemctl restart $SERVICE_NAME${NC}"
    echo -e "  查看状态: ${GREEN}sudo systemctl status $SERVICE_NAME${NC}"
    echo -e "  查看日志: ${GREEN}sudo journalctl -u $SERVICE_NAME -f${NC}"
    echo -e "  禁用服务: ${GREEN}sudo systemctl disable $SERVICE_NAME${NC}"
    echo ""

    print_info "服务访问地址:"
    echo -e "  主页: ${GREEN}http://localhost:8080${NC}"
    echo -e "  API: ${GREEN}http://localhost:8080/api/v1/stats${NC}"
    echo -e "  健康检查: ${GREEN}http://localhost:8080/health${NC}"
}

# 卸载服务
uninstall_service() {
    print_header
    print_info "卸载 NSFW-Go 服务..."

    # 停止服务
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        print_info "停止服务..."
        systemctl stop "$SERVICE_NAME"
    fi

    # 禁用服务
    if systemctl is-enabled --quiet "$SERVICE_NAME"; then
        print_info "禁用服务..."
        systemctl disable "$SERVICE_NAME"
    fi

    # 删除服务文件
    if [ -f "$SERVICE_FILE" ]; then
        print_info "删除服务文件..."
        rm -f "$SERVICE_FILE"
    fi

    # 重新加载 systemd
    systemctl daemon-reload

    print_success "服务卸载完成"
}

# 主安装流程
install() {
    print_header
    echo ""

    check_root
    check_dependencies
    check_port_conflicts

    build_binary
    setup_directories
    install_service
    start_service

    echo ""
    print_success "🎉 NSFW-Go 服务安装完成！"
    echo ""
    show_service_info
}

# 显示帮助
show_help() {
    print_header
    echo ""
    echo "用法: $0 {install|uninstall|help}"
    echo ""
    echo "命令说明:"
    echo "  install    - 安装并启动 NSFW-Go 服务"
    echo "  uninstall  - 停止并卸载 NSFW-Go 服务"
    echo "  help       - 显示此帮助信息"
    echo ""
    echo "安装后管理:"
    echo "  sudo systemctl start nsfw-go     # 启动服务"
    echo "  sudo systemctl stop nsfw-go      # 停止服务"
    echo "  sudo systemctl restart nsfw-go   # 重启服务"
    echo "  sudo systemctl status nsfw-go    # 查看状态"
    echo "  sudo journalctl -u nsfw-go -f    # 实时查看日志"
}

# 主函数
main() {
    case "${1:-help}" in
        install)
            install
            ;;
        uninstall)
            check_root
            uninstall_service
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi