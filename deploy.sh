#!/bin/bash

# ============================================
# NSFW-GO 一键部署脚本
# ============================================
# 用途: 快速部署 NSFW-GO Docker 容器化应用
# 作者: NSFW-GO Team
# ============================================

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 图标定义
CHECK="✓"
CROSS="✗"
ARROW="➜"
WARN="⚠"
INFO="ℹ"

# ============================================
# 工具函数
# ============================================

print_header() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}     NSFW-GO Docker 一键部署工具      ${BLUE}║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}${CHECK}${NC} $1"
}

print_error() {
    echo -e "${RED}${CROSS}${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}${WARN}${NC} $1"
}

print_info() {
    echo -e "${BLUE}${INFO}${NC} $1"
}

print_step() {
    echo -e "${BLUE}${ARROW}${NC} $1"
}

# ============================================
# 检查函数
# ============================================

check_docker() {
    print_step "检查 Docker 环境..."

    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装"
        echo ""
        echo "请先安装 Docker："
        echo "  Ubuntu/Debian: curl -fsSL https://get.docker.com | sh"
        echo "  macOS/Windows: 下载 Docker Desktop"
        echo ""
        exit 1
    fi

    if ! docker ps &> /dev/null; then
        print_error "Docker 服务未运行或权限不足"
        echo ""
        echo "请启动 Docker 服务："
        echo "  Linux: sudo systemctl start docker"
        echo "  macOS/Windows: 启动 Docker Desktop"
        echo ""
        echo "如果是权限问题，请运行："
        echo "  sudo usermod -aG docker $USER"
        echo "  然后重新登录系统"
        echo ""
        exit 1
    fi

    print_success "Docker 环境正常 ($(docker --version | cut -d' ' -f3 | tr -d ','))"
}

check_docker_compose() {
    print_step "检查 Docker Compose..."

    if docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
    elif command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        print_error "Docker Compose 未安装"
        echo ""
        echo "请先安装 Docker Compose："
        echo "  Ubuntu/Debian: sudo apt-get install docker-compose-plugin"
        echo "  macOS/Windows: 已包含在 Docker Desktop 中"
        echo ""
        exit 1
    fi

    print_success "Docker Compose 已安装 ($($COMPOSE_CMD version --short))"
}

check_ports() {
    print_step "检查端口占用..."

    local ports=(8080 5433 6380)
    local port_names=("API" "PostgreSQL" "Redis")
    local occupied=false

    for i in "${!ports[@]}"; do
        local port="${ports[$i]}"
        local name="${port_names[$i]}"

        if netstat -tuln 2>/dev/null | grep -q ":$port " || \
           ss -tuln 2>/dev/null | grep -q ":$port " || \
           lsof -i ":$port" 2>/dev/null | grep -q LISTEN; then
            print_warning "端口 $port ($name) 已被占用"
            occupied=true
        fi
    done

    if [ "$occupied" = true ]; then
        echo ""
        read -p "是否继续部署？某些服务可能无法启动 (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        print_success "所有必需端口可用"
    fi
}

# ============================================
# 配置函数
# ============================================

setup_env() {
    print_step "配置环境变量..."

    if [ -f .env ]; then
        print_warning ".env 文件已存在"
        read -p "是否覆盖现有配置？(y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "跳过环境配置"
            return
        fi
    fi

    # 复制模板
    cp .env.example .env
    print_success "已创建 .env 文件"

    echo ""
    echo -e "${YELLOW}请配置以下必需项：${NC}"
    echo ""

    # 数据库密码
    read -p "设置数据库密码 [nsfw123]: " db_password
    db_password=${db_password:-nsfw123}
    sed -i.bak "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=$db_password/" .env

    # 媒体库路径
    echo ""
    echo "媒体库路径示例："
    echo "  - Linux: /media/NSFW"
    echo "  - Synology: /volume1/media/NSFW"
    echo "  - macOS: /Users/username/Movies/NSFW"
    echo "  - Windows (WSL): /mnt/c/Users/username/Videos/NSFW"
    echo ""

    read -p "设置媒体库路径 [./media]: " media_path
    media_path=${media_path:-./media}

    # 创建媒体目录（如果不存在）
    if [ ! -d "$media_path" ]; then
        mkdir -p "$media_path"
        print_info "已创建媒体目录: $media_path"
    fi

    sed -i.bak "s|MEDIA_BASE_PATH=.*|MEDIA_BASE_PATH=$media_path|" .env

    # 时区设置
    echo ""
    read -p "设置时区 [Asia/Shanghai]: " timezone
    timezone=${timezone:-Asia/Shanghai}
    sed -i.bak "s|TZ=.*|TZ=$timezone|" .env

    # 删除备份文件
    rm -f .env.bak

    echo ""
    print_success "环境变量配置完成"

    # 询问是否配置可选功能
    echo ""
    read -p "是否配置种子下载功能？(y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        configure_torrent
    fi
}

configure_torrent() {
    echo ""
    echo -e "${BLUE}配置种子下载功能：${NC}"
    echo ""

    read -p "Jackett URL [http://jackett:9117]: " jackett_url
    jackett_url=${jackett_url:-http://jackett:9117}
    sed -i.bak "s|JACKETT_URL=.*|JACKETT_URL=$jackett_url|" .env

    read -p "Jackett API Key: " jackett_key
    if [ -n "$jackett_key" ]; then
        sed -i.bak "s|JACKETT_API_KEY=.*|JACKETT_API_KEY=$jackett_key|" .env
    fi

    read -p "qBittorrent URL [http://qbittorrent:8085]: " qb_url
    qb_url=${qb_url:-http://qbittorrent:8085}
    sed -i.bak "s|QBITTORRENT_URL=.*|QBITTORRENT_URL=$qb_url|" .env

    read -p "qBittorrent 用户名 [admin]: " qb_user
    qb_user=${qb_user:-admin}
    sed -i.bak "s|QBITTORRENT_USERNAME=.*|QBITTORRENT_USERNAME=$qb_user|" .env

    read -sp "qBittorrent 密码: " qb_pass
    echo ""
    if [ -n "$qb_pass" ]; then
        sed -i.bak "s|QBITTORRENT_PASSWORD=.*|QBITTORRENT_PASSWORD=$qb_pass|" .env
    fi

    rm -f .env.bak
    print_success "种子下载配置完成"
}

# ============================================
# 部署函数
# ============================================

pull_images() {
    print_step "拉取 Docker 镜像..."

    $COMPOSE_CMD -f docker-compose.prod.yml pull 2>&1 | grep -E "Pulling|Downloaded|Digest" || true

    print_success "镜像拉取完成"
}

build_images() {
    print_step "构建应用镜像..."

    # 设置构建参数
    export VERSION=$(git describe --tags --always 2>/dev/null || echo "latest")
    export BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    $COMPOSE_CMD -f docker-compose.prod.yml build --no-cache 2>&1 | \
        grep -E "building|Successfully|Sending build context" || true

    print_success "镜像构建完成"
}

start_services() {
    print_step "启动服务..."

    # 询问是否启用可选服务
    echo ""
    echo "可选服务："
    echo "  1. 仅核心服务（API + 数据库 + Redis）"
    echo "  2. 核心 + 管理工具（pgAdmin, Redis Commander）"
    echo "  3. 核心 + 监控（Prometheus, Grafana）"
    echo "  4. 全部服务"
    echo ""

    read -p "选择启动模式 [1]: " mode
    mode=${mode:-1}

    local profiles=""
    case $mode in
        2)
            profiles="--profile admin"
            ;;
        3)
            profiles="--profile monitoring"
            ;;
        4)
            profiles="--profile admin --profile monitoring --profile nginx"
            ;;
    esac

    $COMPOSE_CMD -f docker-compose.prod.yml $profiles up -d

    print_success "服务启动命令已执行"
}

wait_for_services() {
    print_step "等待服务就绪..."

    echo -n "等待 API 服务"

    local max_wait=60
    local count=0

    while [ $count -lt $max_wait ]; do
        if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
            echo ""
            print_success "API 服务已就绪"
            return 0
        fi

        echo -n "."
        sleep 2
        count=$((count + 2))
    done

    echo ""
    print_warning "服务启动超时，请查看日志"
    return 1
}

show_status() {
    echo ""
    print_step "服务状态："
    echo ""

    $COMPOSE_CMD -f docker-compose.prod.yml ps
}

show_info() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║${NC}         部署完成！访问信息           ${GREEN}║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    echo ""

    echo -e "${BLUE}Web 界面:${NC}"
    echo -e "  ${ARROW} http://localhost:8080"
    echo ""

    echo -e "${BLUE}API 健康检查:${NC}"
    echo -e "  ${ARROW} http://localhost:8080/health"
    echo ""

    if [ -f .env ] && grep -q "ENABLE_PGADMIN=true" .env; then
        local pgadmin_port=$(grep PGADMIN_PORT .env | cut -d'=' -f2)
        echo -e "${BLUE}pgAdmin:${NC}"
        echo -e "  ${ARROW} http://localhost:${pgadmin_port:-5050}"
        echo ""
    fi

    if [ -f .env ] && grep -q "ENABLE_MONITORING=true" .env; then
        local grafana_port=$(grep GRAFANA_PORT .env | cut -d'=' -f2)
        echo -e "${BLUE}Grafana:${NC}"
        echo -e "  ${ARROW} http://localhost:${grafana_port:-3000}"
        echo ""
    fi

    echo -e "${BLUE}常用命令:${NC}"
    echo -e "  ${ARROW} 查看日志: docker compose -f docker-compose.prod.yml logs -f"
    echo -e "  ${ARROW} 停止服务: docker compose -f docker-compose.prod.yml down"
    echo -e "  ${ARROW} 重启服务: docker compose -f docker-compose.prod.yml restart"
    echo ""

    echo -e "${YELLOW}提示:${NC}"
    echo -e "  - 详细文档请查看 ${BLUE}README.docker.md${NC}"
    echo -e "  - 配置文件位于 ${BLUE}.env${NC}"
    echo -e "  - 如有问题请查看日志或提交 Issue"
    echo ""
}

# ============================================
# 清理函数
# ============================================

cleanup() {
    print_step "清理临时文件..."

    # 清理可能的备份文件
    rm -f .env.bak

    print_success "清理完成"
}

# ============================================
# 主函数
# ============================================

main() {
    # 打印标题
    print_header

    # 检查环境
    check_docker
    check_docker_compose
    check_ports

    echo ""

    # 配置环境
    setup_env

    echo ""

    # 询问是否继续
    read -p "开始部署？(Y/n): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        print_info "部署已取消"
        exit 0
    fi

    echo ""

    # 拉取和构建
    pull_images
    echo ""
    build_images

    echo ""

    # 启动服务
    start_services

    echo ""

    # 等待服务就绪
    if wait_for_services; then
        show_status
        show_info
    else
        echo ""
        print_error "服务可能未正常启动"
        echo ""
        echo "请运行以下命令查看日志："
        echo "  docker compose -f docker-compose.prod.yml logs"
        echo ""
        exit 1
    fi

    # 清理
    cleanup
}

# ============================================
# 错误处理
# ============================================

trap 'print_error "部署过程中发生错误"; cleanup; exit 1' ERR

# ============================================
# 脚本入口
# ============================================

# 检查是否在项目根目录
if [ ! -f "docker-compose.prod.yml" ]; then
    print_error "请在项目根目录运行此脚本"
    exit 1
fi

# 运行主函数
main

exit 0
