# ⚡ NSFW-GO 快速参考卡

> 常用命令和操作速查表

## 🚀 一键部署

```bash
git clone https://github.com/your-repo/NSFW-GO.git
cd NSFW-GO
chmod +x deploy.sh
./deploy.sh
```

访问：**http://localhost:8080**

---

## 📱 Web 界面快速访问

| 页面 | URL | 功能 |
|------|-----|------|
| 🏠 主页 | http://localhost:8080/ | 系统概览 |
| 🎬 本地影片 | http://localhost:8080/local-movies.html | 管理本地影片 |
| 🔍 搜索 | http://localhost:8080/search.html | 搜索影片 |
| 📈 排行榜 | http://localhost:8080/rankings.html | 热门排行 |
| 💾 下载 | http://localhost:8080/downloads.html | 下载管理 |
| ⚙️ 配置 | http://localhost:8080/config.html | 系统配置 |
| 📝 日志 | http://localhost:8080/logs.html | 系统日志 |

---

## 🛠️ Docker 常用命令

### 服务管理

```bash
# 启动所有服务
docker compose -f docker-compose.prod.yml up -d

# 停止所有服务
docker compose -f docker-compose.prod.yml down

# 重启服务
docker compose -f docker-compose.prod.yml restart

# 查看服务状态
docker compose -f docker-compose.prod.yml ps

# 查看日志
docker compose -f docker-compose.prod.yml logs -f api
```

### 启动可选服务

```bash
# 启动管理工具（pgAdmin + Redis Commander）
docker compose -f docker-compose.prod.yml --profile admin up -d

# 启动监控（Prometheus + Grafana）
docker compose -f docker-compose.prod.yml --profile monitoring up -d

# 启动所有服务
docker compose -f docker-compose.prod.yml \
  --profile admin \
  --profile monitoring \
  up -d
```

---

## 🗄️ 数据库操作

```bash
# 进入数据库
docker exec -it nsfw-postgres psql -U nsfw -d nsfw_db

# 备份数据库
docker exec nsfw-postgres pg_dump -U nsfw nsfw_db > backup_$(date +%Y%m%d).sql

# 恢复数据库
docker exec -i nsfw-postgres psql -U nsfw -d nsfw_db < backup_20250101.sql

# 查看数据库大小
docker exec nsfw-postgres psql -U nsfw -d nsfw_db \
  -c "SELECT pg_size_pretty(pg_database_size('nsfw_db'));"
```

---

## 🔧 配置文件

### 必须修改的配置（.env）

```bash
# 数据库密码（必须改！）
POSTGRES_PASSWORD=your_secure_password_here

# 媒体库路径（必须改！）
MEDIA_BASE_PATH=/path/to/your/media
```

### 常用配置项

```bash
# 应用端口
API_PORT=8080

# 扫描间隔
MEDIA_SCAN_INTERVAL=15m

# 支持的视频格式
MEDIA_SUPPORTED_FORMATS=mp4,mkv,avi,mov,wmv

# 日志级别
LOG_LEVEL=info

# 代理设置
CRAWLER_PROXY=http://proxy:8080
```

---

## 📊 常用 API 端点

```bash
# 健康检查
curl http://localhost:8080/health

# 获取本地影片（分页）
curl "http://localhost:8080/api/v1/local/movies?page=1&limit=20"

# 获取统计信息
curl http://localhost:8080/api/v1/local/stats

# 触发手动扫描
curl -X POST http://localhost:8080/api/v1/local/scan

# 搜索 JAVDb
curl "http://localhost:8080/api/v1/search/javdb?q=STARS-123"

# 获取排行榜
curl "http://localhost:8080/api/v1/rankings?type=daily&page=1"

# 获取系统配置
curl http://localhost:8080/api/v1/config

# 获取日志
curl "http://localhost:8080/api/v1/logs?level=error&limit=100"
```

---

## 🔍 日志查看

```bash
# API 服务日志
docker compose -f docker-compose.prod.yml logs -f api

# 数据库日志
docker compose -f docker-compose.prod.yml logs -f postgres

# Redis 日志
docker compose -f docker-compose.prod.yml logs -f redis

# 所有服务日志
docker compose -f docker-compose.prod.yml logs -f

# 查看最近 100 行
docker compose -f docker-compose.prod.yml logs --tail=100 api

# 导出日志
docker compose -f docker-compose.prod.yml logs api > api_logs.txt
```

---

## 🐛 故障排除

### 服务无法启动

```bash
# 查看详细日志
docker compose -f docker-compose.prod.yml logs api

# 检查端口占用
sudo netstat -tulpn | grep -E '8080|5433|6380'

# 重启服务
docker compose -f docker-compose.prod.yml restart
```

### 数据库连接失败

```bash
# 检查数据库状态
docker compose -f docker-compose.prod.yml ps postgres

# 测试数据库连接
docker exec nsfw-postgres pg_isready -U nsfw

# 重启数据库
docker compose -f docker-compose.prod.yml restart postgres
```

### 媒体文件无法访问

```bash
# 检查容器内路径
docker exec nsfw-api ls -la /app/media

# 检查宿主机路径
ls -la /path/to/your/media

# 检查权限
sudo chmod -R 755 /path/to/your/media
```

### 清理和重启

```bash
# 停止并删除容器（保留数据）
docker compose -f docker-compose.prod.yml down

# 停止并删除容器和数据（危险！）
docker compose -f docker-compose.prod.yml down -v

# 清理 Docker 缓存
docker system prune -a

# 重新构建并启动
docker compose -f docker-compose.prod.yml build --no-cache
docker compose -f docker-compose.prod.yml up -d
```

---

## 🔄 升级流程

```bash
# 1. 备份数据
docker exec nsfw-postgres pg_dump -U nsfw nsfw_db > backup.sql

# 2. 停止服务
docker compose -f docker-compose.prod.yml down

# 3. 拉取更新
git pull origin main

# 4. 重新构建
docker compose -f docker-compose.prod.yml build --no-cache

# 5. 启动服务
docker compose -f docker-compose.prod.yml up -d

# 6. 检查状态
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f api
```

---

## 📈 性能优化

### 小型部署（< 1000 影片）

```bash
# .env 配置
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5
MEDIA_SCAN_INTERVAL=30m
```

### 中型部署（1000-10000 影片）

```bash
# .env 配置
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
MEDIA_SCAN_INTERVAL=15m
```

### 大型部署（> 10000 影片）

```bash
# .env 配置
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=20
MEDIA_SCAN_INTERVAL=5m
```

---

## 🔐 安全检查

```bash
# 检查密码强度
cat .env | grep PASSWORD

# 检查开放端口
sudo ufw status
sudo netstat -tulpn | grep LISTEN

# 更新所有镜像
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

---

## 📊 系统监控

```bash
# 查看容器资源占用
docker stats

# 查看磁盘占用
docker system df

# 查看特定容器资源
docker stats nsfw-api

# 查看数据库连接数
docker exec nsfw-postgres psql -U nsfw -d nsfw_db \
  -c "SELECT count(*) FROM pg_stat_activity;"
```

---

## 🎯 常见使用场景

### 场景 1：添加新影片

```bash
# 1. 复制影片到媒体库
cp movie.mp4 /path/to/your/media/

# 2. 触发手动扫描
curl -X POST http://localhost:8080/api/v1/local/scan

# 或访问 Web 界面点击"刷新"
```

### 场景 2：搜索并下载

```bash
# 1. 搜索影片
curl "http://localhost:8080/api/v1/search/javdb?q=STARS-123"

# 2. 访问 Web 界面下载种子
# http://localhost:8080/search.html
```

### 场景 3：查看排行榜

```bash
# 1. 获取排行榜
curl "http://localhost:8080/api/v1/rankings?type=daily"

# 2. 访问 Web 界面
# http://localhost:8080/rankings.html
```

---

## 🔗 快速链接

| 资源 | 链接 |
|------|------|
| 📖 用户手册 | [USER_MANUAL.md](USER_MANUAL.md) |
| 🚀 部署指南 | [DEPLOY_GUIDE.md](DEPLOY_GUIDE.md) |
| 📚 完整文档 | [README.docker.md](README.docker.md) |
| 💻 开发文档 | [CLAUDE.md](CLAUDE.md) |
| 🐛 问题反馈 | [GitHub Issues](https://github.com/your-repo/NSFW-GO/issues) |

---

## 🆘 紧急情况

### 服务完全无法访问

```bash
# 完全重启
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d

# 检查所有服务
docker compose -f docker-compose.prod.yml ps

# 查看错误日志
docker compose -f docker-compose.prod.yml logs --tail=50
```

### 数据库损坏

```bash
# 使用备份恢复
docker compose -f docker-compose.prod.yml down
docker volume rm nsfw-go_postgres_data
docker compose -f docker-compose.prod.yml up -d postgres
docker exec -i nsfw-postgres psql -U nsfw -d nsfw_db < backup.sql
docker compose -f docker-compose.prod.yml up -d
```

### 磁盘空间不足

```bash
# 清理 Docker 资源
docker system prune -a --volumes

# 清理日志
docker compose -f docker-compose.prod.yml down
rm -rf logs/*
docker compose -f docker-compose.prod.yml up -d

# 清理数据库日志
docker exec nsfw-postgres psql -U nsfw -d nsfw_db \
  -c "TRUNCATE TABLE logs;"
```

---

**保存此页面以便快速查阅！** 📌
