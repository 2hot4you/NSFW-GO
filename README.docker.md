# 🐳 NSFW-GO Docker 部署指南

快速、简单地使用 Docker Compose 部署 NSFW-GO 应用。

## 📋 目录

- [系统要求](#系统要求)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [高级配置](#高级配置)
- [管理命令](#管理命令)
- [故障排除](#故障排除)
- [升级指南](#升级指南)

---

## 🔧 系统要求

### 最低配置
- **操作系统**: Linux / macOS / Windows (带 WSL2)
- **Docker**: 20.10.0 或更高版本
- **Docker Compose**: 2.0.0 或更高版本
- **内存**: 最低 2GB，推荐 4GB+
- **磁盘**: 10GB+ 可用空间（不包括媒体文件）

### 推荐配置
- **CPU**: 4 核心或更多
- **内存**: 8GB 或更多
- **磁盘**: SSD，20GB+ 可用空间
- **网络**: 稳定的互联网连接（用于爬虫功能）

### 安装 Docker 和 Docker Compose

**Linux (Ubuntu/Debian):**
```bash
# 安装 Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 安装 Docker Compose
sudo apt-get update
sudo apt-get install docker-compose-plugin

# 验证安装
docker --version
docker compose version
```

**macOS / Windows:**
- 下载并安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/)

---

## 🚀 快速开始

### 方法一：使用一键部署脚本（推荐）

```bash
# 1. 克隆仓库
git clone https://github.com/your-repo/NSFW-GO.git
cd NSFW-GO

# 2. 运行一键部署脚本
chmod +x deploy.sh
./deploy.sh

# 脚本会自动：
# - 检查 Docker 环境
# - 复制并引导配置 .env 文件
# - 启动所有服务
# - 显示访问信息
```

### 方法二：手动部署

```bash
# 1. 克隆仓库
git clone https://github.com/your-repo/NSFW-GO.git
cd NSFW-GO

# 2. 复制并编辑环境变量文件
cp .env.example .env
nano .env  # 或使用你喜欢的编辑器

# 3. 修改关键配置（必须！）
# - POSTGRES_PASSWORD: 设置强数据库密码
# - MEDIA_BASE_PATH: 设置你的媒体库路径

# 4. 启动服务
docker compose -f docker-compose.prod.yml up -d

# 5. 查看日志
docker compose -f docker-compose.prod.yml logs -f api

# 6. 等待服务启动完成
# 访问 http://localhost:8080
```

### 验证部署

```bash
# 检查服务状态
docker compose -f docker-compose.prod.yml ps

# 所有服务应该显示为 "Up (healthy)"

# 测试 API 健康检查
curl http://localhost:8080/health

# 应该返回: {"status":"ok"}
```

---

## ⚙️ 配置说明

### 必需配置

在 `.env` 文件中，以下配置是**必须**设置的：

```bash
# 数据库密码（请设置强密码！）
POSTGRES_PASSWORD=your_secure_password_here

# 媒体库路径（绝对路径）
MEDIA_BASE_PATH=/path/to/your/media/library
```

### 重要配置

```bash
# 应用运行模式
APP_MODE=release          # debug, release, test

# 时区设置
TZ=Asia/Shanghai

# API 端口
API_PORT=8080

# 日志级别
LOG_LEVEL=info           # debug, info, warn, error

# 媒体扫描间隔
MEDIA_SCAN_INTERVAL=15m  # 15分钟扫描一次
```

### 可选配置

#### 种子下载配置

如果你想使用种子下载功能，需要配置 Jackett 和 qBittorrent：

```bash
# Jackett 配置
JACKETT_URL=http://your-jackett-host:9117
JACKETT_API_KEY=your_jackett_api_key

# qBittorrent 配置
QBITTORRENT_URL=http://your-qbittorrent-host:8085
QBITTORRENT_USERNAME=admin
QBITTORRENT_PASSWORD=adminpass
QBITTORRENT_DOWNLOAD_DIR=/downloads
```

#### 代理配置

如果需要使用代理访问外部网站（如 JAVDb）：

```bash
# HTTP 代理
CRAWLER_PROXY=http://proxy.example.com:8080

# SOCKS5 代理
CRAWLER_PROXY=socks5://proxy.example.com:1080
```

#### 爬虫计划配置

```bash
# 排行榜爬取计划（Cron 表达式）
CRAWLER_RANKING_SCHEDULE=0 12 * * *    # 每天 12:00

# 本地匹配检查计划
CRAWLER_LOCAL_CHECK_SCHEDULE=0 * * * * # 每小时
```

---

## 🎯 高级配置

### 启用管理工具

使用 Docker Compose Profiles 来启用可选服务：

#### pgAdmin（PostgreSQL 管理界面）

```bash
# 在 .env 中配置
ENABLE_PGADMIN=true
PGADMIN_PORT=5050
PGADMIN_EMAIL=admin@nsfw.local
PGADMIN_PASSWORD=admin123

# 启动时添加 profile
docker compose -f docker-compose.prod.yml --profile admin up -d

# 访问 http://localhost:5050
```

#### Redis Commander（Redis 管理界面）

```bash
# 在 .env 中配置
ENABLE_REDIS_COMMANDER=true
REDIS_COMMANDER_PORT=8081

# 启动时添加 profile
docker compose -f docker-compose.prod.yml --profile admin up -d

# 访问 http://localhost:8081
```

#### Nginx 反向代理

```bash
# 在 .env 中配置
ENABLE_NGINX=true
NGINX_HTTP_PORT=80
NGINX_HTTPS_PORT=443

# 启动时添加 profile
docker compose -f docker-compose.prod.yml --profile nginx up -d

# 访问 http://localhost
```

#### 监控系统（Prometheus + Grafana）

```bash
# 在 .env 中配置
ENABLE_MONITORING=true
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
GRAFANA_PASSWORD=admin123

# 启动时添加 profile
docker compose -f docker-compose.prod.yml --profile monitoring up -d

# 访问:
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3000 (admin/admin123)
```

#### Telegram Bot

```bash
# 在 .env 中配置
ENABLE_BOT=true
TELEGRAM_BOT_TOKEN=your_telegram_bot_token

# 启动时添加 profile
docker compose -f docker-compose.prod.yml --profile bot up -d
```

### 同时启用多个 Profile

```bash
# 启用管理工具和监控
docker compose -f docker-compose.prod.yml --profile admin --profile monitoring up -d

# 启用所有可选服务
docker compose -f docker-compose.prod.yml \
  --profile admin \
  --profile monitoring \
  --profile nginx \
  --profile bot \
  up -d
```

---

## 🛠️ 管理命令

### 基础操作

```bash
# 启动所有服务
docker compose -f docker-compose.prod.yml up -d

# 停止所有服务
docker compose -f docker-compose.prod.yml down

# 重启服务
docker compose -f docker-compose.prod.yml restart

# 查看服务状态
docker compose -f docker-compose.prod.yml ps

# 查看实时日志
docker compose -f docker-compose.prod.yml logs -f

# 查看特定服务日志
docker compose -f docker-compose.prod.yml logs -f api
docker compose -f docker-compose.prod.yml logs -f postgres
```

### 服务管理

```bash
# 重启单个服务
docker compose -f docker-compose.prod.yml restart api

# 停止单个服务
docker compose -f docker-compose.prod.yml stop api

# 启动单个服务
docker compose -f docker-compose.prod.yml start api

# 重新构建并启动
docker compose -f docker-compose.prod.yml up -d --build
```

### 数据库操作

```bash
# 进入数据库容器
docker exec -it nsfw-postgres psql -U nsfw -d nsfw_db

# 备份数据库
docker exec nsfw-postgres pg_dump -U nsfw nsfw_db > backup_$(date +%Y%m%d_%H%M%S).sql

# 恢复数据库
docker exec -i nsfw-postgres psql -U nsfw -d nsfw_db < backup_20250101_120000.sql

# 查看数据库大小
docker exec nsfw-postgres psql -U nsfw -d nsfw_db -c "SELECT pg_size_pretty(pg_database_size('nsfw_db'));"
```

### 日志管理

```bash
# 查看最近 100 行日志
docker compose -f docker-compose.prod.yml logs --tail=100 api

# 导出日志到文件
docker compose -f docker-compose.prod.yml logs api > api_logs.txt

# 清理日志（通过重新创建容器）
docker compose -f docker-compose.prod.yml up -d --force-recreate
```

### 资源清理

```bash
# 停止并删除所有容器
docker compose -f docker-compose.prod.yml down

# 停止并删除所有容器和卷（警告：会删除所有数据！）
docker compose -f docker-compose.prod.yml down -v

# 清理未使用的 Docker 资源
docker system prune -a

# 查看 Docker 磁盘占用
docker system df
```

---

## 🔍 故障排除

### 服务无法启动

**问题**: 容器启动后立即退出

```bash
# 查看详细日志
docker compose -f docker-compose.prod.yml logs api

# 常见原因：
# 1. .env 文件配置错误
# 2. 端口被占用
# 3. 数据库连接失败
```

**解决方案**:
```bash
# 检查端口占用
sudo netstat -tulpn | grep -E '8080|5433|6380'

# 检查配置文件
cat .env | grep -v '^#' | grep -v '^$'

# 重新启动
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d
```

### 数据库连接失败

**问题**: `connection refused` 或 `password authentication failed`

```bash
# 检查数据库容器状态
docker compose -f docker-compose.prod.yml ps postgres

# 查看数据库日志
docker compose -f docker-compose.prod.yml logs postgres

# 测试数据库连接
docker exec nsfw-postgres pg_isready -U nsfw
```

**解决方案**:
```bash
# 确保 .env 中的密码正确
# 重启数据库服务
docker compose -f docker-compose.prod.yml restart postgres

# 如果密码确实错误，需要重建数据库
docker compose -f docker-compose.prod.yml down -v
docker compose -f docker-compose.prod.yml up -d
```

### 媒体文件无法访问

**问题**: API 返回 404 或 权限拒绝

```bash
# 检查挂载路径
docker exec nsfw-api ls -la /app/media

# 检查宿主机路径
ls -la /path/to/your/media

# 查看权限
docker exec nsfw-api id
```

**解决方案**:
```bash
# 确保 .env 中的 MEDIA_BASE_PATH 正确
# 确保路径存在且有读取权限
sudo chmod -R 755 /path/to/your/media

# 重启 API 服务
docker compose -f docker-compose.prod.yml restart api
```

### 爬虫无法访问外部网站

**问题**: `timeout` 或 `connection refused`

**解决方案**:
```bash
# 方案 1: 配置代理（在 .env 中）
CRAWLER_PROXY=http://your-proxy:8080

# 方案 2: 检查防火墙
sudo ufw status

# 方案 3: 测试网络连接
docker exec nsfw-api wget -O- https://javdb.com
```

### 容器占用太多磁盘空间

```bash
# 查看磁盘占用
docker system df -v

# 清理未使用的镜像
docker image prune -a

# 清理未使用的卷
docker volume prune

# 清理构建缓存
docker builder prune
```

### 性能问题

**问题**: API 响应慢，CPU/内存占用高

```bash
# 查看资源占用
docker stats

# 增加资源限制（在 docker-compose.prod.yml 中）
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          memory: 1G
```

---

## 📈 升级指南

### 升级到新版本

```bash
# 1. 备份数据库
docker exec nsfw-postgres pg_dump -U nsfw nsfw_db > backup_before_upgrade.sql

# 2. 停止服务
docker compose -f docker-compose.prod.yml down

# 3. 拉取最新代码
git pull origin main

# 4. 更新镜像
docker compose -f docker-compose.prod.yml pull

# 5. 重新构建
docker compose -f docker-compose.prod.yml build --no-cache

# 6. 启动服务
docker compose -f docker-compose.prod.yml up -d

# 7. 查看日志确认
docker compose -f docker-compose.prod.yml logs -f api
```

### 回滚到旧版本

```bash
# 1. 停止服务
docker compose -f docker-compose.prod.yml down

# 2. 切换到旧版本
git checkout <old-version-tag>

# 3. 恢复数据库（如果需要）
docker exec -i nsfw-postgres psql -U nsfw -d nsfw_db < backup_before_upgrade.sql

# 4. 启动服务
docker compose -f docker-compose.prod.yml up -d
```

---

## 📊 监控和维护

### 健康检查

```bash
# API 健康检查
curl http://localhost:8080/health

# 数据库健康检查
docker exec nsfw-postgres pg_isready -U nsfw

# Redis 健康检查
docker exec nsfw-redis redis-cli ping

# 所有服务健康状态
docker compose -f docker-compose.prod.yml ps
```

### 定期维护任务

```bash
# 每周备份数据库
0 2 * * 0 docker exec nsfw-postgres pg_dump -U nsfw nsfw_db > /backups/weekly_$(date +\%Y\%m\%d).sql

# 每月清理 Docker 资源
0 3 1 * * docker system prune -af --volumes

# 每天查看日志大小
0 0 * * * docker system df
```

### 性能监控

访问监控界面（如果启用了 `monitoring` profile）：

- **Prometheus**: http://localhost:9090
  - 查看指标和告警

- **Grafana**: http://localhost:3000
  - 用户名: admin
  - 密码: 在 `.env` 中的 `GRAFANA_PASSWORD`

---

## 🔐 安全建议

1. **修改默认密码**: 必须修改 `.env` 中的所有默认密码
2. **使用防火墙**: 限制外部访问，只开放必要端口
3. **启用 HTTPS**: 在生产环境使用 SSL/TLS 证书
4. **定期备份**: 设置自动备份任务
5. **更新依赖**: 定期更新 Docker 镜像
6. **日志审计**: 定期检查日志，发现异常行为
7. **限制访问**: 使用 Nginx 添加访问控制和速率限制

---

## 📞 获取帮助

- **GitHub Issues**: [提交问题](https://github.com/your-repo/NSFW-GO/issues)
- **文档**: 查看 [CLAUDE.md](./CLAUDE.md) 和 [README.md](./README.md)
- **日志**: 优先查看容器日志定位问题

---

## 📝 常见配置示例

### 示例 1: 最小化部署（只有核心服务）

```bash
# .env 配置
APP_MODE=release
POSTGRES_PASSWORD=your_password
MEDIA_BASE_PATH=/media/NSFW
API_PORT=8080

# 启动
docker compose -f docker-compose.prod.yml up -d
```

### 示例 2: 完整部署（所有功能）

```bash
# .env 配置
APP_MODE=release
POSTGRES_PASSWORD=your_password
MEDIA_BASE_PATH=/media/NSFW

# 种子下载
JACKETT_URL=http://jackett:9117
JACKETT_API_KEY=xxx
QBITTORRENT_URL=http://10.10.10.200:8085
QBITTORRENT_USERNAME=admin
QBITTORRENT_PASSWORD=pass

# 管理工具
ENABLE_PGADMIN=true
ENABLE_MONITORING=true

# 启动所有服务
docker compose -f docker-compose.prod.yml \
  --profile admin \
  --profile monitoring \
  --profile nginx \
  up -d
```

### 示例 3: Synology NAS 部署

```bash
# .env 配置
POSTGRES_EXTERNAL_PORT=5433
REDIS_EXTERNAL_PORT=6380
API_PORT=8080

# 使用 NAS 路径
MEDIA_BASE_PATH=/volume1/media/NSFW
QBITTORRENT_DOWNLOAD_DIR=/volume1/Downloads

# 启动
docker compose -f docker-compose.prod.yml up -d
```

---

## 🎉 部署完成

恭喜！你已经成功部署 NSFW-GO。

**访问应用**:
- Web 界面: http://localhost:8080
- API 文档: http://localhost:8080/swagger (如果启用)
- 健康检查: http://localhost:8080/health

**下一步**:
1. 浏览 Web 界面，熟悉功能
2. 配置媒体库扫描
3. 设置爬虫计划
4. 配置种子下载（可选）
5. 启用监控和管理工具（可选）

享受使用 NSFW-GO！🚀
