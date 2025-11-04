# 🚀 NSFW-GO 快速部署指南

> 5 分钟快速部署 NSFW-GO 到你的服务器

## ⚡ 快速开始（推荐）

### 前提条件
- ✅ 已安装 Docker 和 Docker Compose
- ✅ 有媒体库目录访问权限
- ✅ 至少 2GB 可用内存

### 一键部署

```bash
# 1. 克隆项目
git clone https://github.com/your-repo/NSFW-GO.git
cd NSFW-GO

# 2. 运行部署脚本
chmod +x deploy.sh
./deploy.sh

# 3. 按照提示配置
# - 设置数据库密码
# - 设置媒体库路径
# - 选择可选功能

# 4. 访问应用
# http://localhost:8080
```

就这么简单！🎉

---

## 📝 手动部署

如果你想手动控制每一步：

### 步骤 1: 配置环境变量

```bash
# 复制配置模板
cp .env.example .env

# 编辑配置文件
nano .env
```

**必须修改的配置**：
```bash
# 数据库密码（必须修改！）
POSTGRES_PASSWORD=your_secure_password_here

# 媒体库路径（必须修改！）
MEDIA_BASE_PATH=/path/to/your/media
```

### 步骤 2: 启动服务

```bash
# 启动所有核心服务
docker compose -f docker-compose.prod.yml up -d

# 查看启动日志
docker compose -f docker-compose.prod.yml logs -f
```

### 步骤 3: 验证部署

```bash
# 检查服务状态
docker compose -f docker-compose.prod.yml ps

# 测试 API
curl http://localhost:8080/health
```

---

## 🎯 常见部署场景

### 场景 1: 本地开发测试

```bash
# 使用默认配置
cp .env.example .env
docker compose -f docker-compose.prod.yml up -d
```

访问: http://localhost:8080

### 场景 2: Synology NAS 部署

```env
# .env 配置
POSTGRES_PASSWORD=your_password
MEDIA_BASE_PATH=/volume1/media/NSFW
QBITTORRENT_URL=http://10.10.10.200:8085
QBITTORRENT_DOWNLOAD_DIR=/volume1/Downloads
```

```bash
docker compose -f docker-compose.prod.yml up -d
```

### 场景 3: 带管理工具的完整部署

```bash
# 启动核心服务 + 管理工具
docker compose -f docker-compose.prod.yml --profile admin up -d
```

访问:
- 应用: http://localhost:8080
- pgAdmin: http://localhost:5050
- Redis Commander: http://localhost:8081

### 场景 4: 生产环境（带监控）

```bash
# 启动所有服务
docker compose -f docker-compose.prod.yml \
  --profile admin \
  --profile monitoring \
  --profile nginx \
  up -d
```

访问:
- 应用: http://localhost (通过 Nginx)
- Grafana: http://localhost:3000
- Prometheus: http://localhost:9090

---

## 🔧 常用命令

```bash
# 查看服务状态
docker compose -f docker-compose.prod.yml ps

# 查看日志
docker compose -f docker-compose.prod.yml logs -f api

# 重启服务
docker compose -f docker-compose.prod.yml restart

# 停止服务
docker compose -f docker-compose.prod.yml down

# 停止并删除所有数据（危险！）
docker compose -f docker-compose.prod.yml down -v
```

---

## ❓ 常见问题

### Q: 端口被占用怎么办？

修改 `.env` 文件中的端口：

```env
API_PORT=8090
POSTGRES_EXTERNAL_PORT=5434
REDIS_EXTERNAL_PORT=6381
```

### Q: 数据库连接失败？

检查密码配置：

```bash
# 查看配置
cat .env | grep POSTGRES_PASSWORD

# 重启数据库
docker compose -f docker-compose.prod.yml restart postgres
```

### Q: 媒体文件无法访问？

检查路径和权限：

```bash
# 检查路径存在
ls -la /path/to/your/media

# 检查容器内挂载
docker exec nsfw-api ls -la /app/media
```

### Q: 如何升级到新版本？

```bash
# 1. 备份数据库
docker exec nsfw-postgres pg_dump -U nsfw nsfw_db > backup.sql

# 2. 拉取新版本
git pull

# 3. 重新构建
docker compose -f docker-compose.prod.yml build --no-cache

# 4. 重启服务
docker compose -f docker-compose.prod.yml up -d
```

---

## 📊 性能优化建议

### 小型部署（< 1000 影片）
```env
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5
MEDIA_SCAN_INTERVAL=30m
```

### 中型部署（1000-10000 影片）
```env
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
MEDIA_SCAN_INTERVAL=15m
```

### 大型部署（> 10000 影片）
```env
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=20
MEDIA_SCAN_INTERVAL=5m
```

---

## 🔐 安全检查清单

- [ ] 修改了所有默认密码
- [ ] 使用强密码（至少 12 位，包含大小写字母、数字、符号）
- [ ] 限制了外部访问（防火墙规则）
- [ ] 启用了 HTTPS（生产环境）
- [ ] 定期备份数据库
- [ ] 定期更新镜像

---

## 📚 更多信息

- **完整文档**: [README.docker.md](./README.docker.md)
- **开发文档**: [CLAUDE.md](./CLAUDE.md)
- **问题反馈**: [GitHub Issues](https://github.com/your-repo/NSFW-GO/issues)

---

## 🎉 部署成功后做什么？

1. **浏览 Web 界面**
   访问 http://localhost:8080，熟悉各个功能

2. **配置媒体扫描**
   在配置页面设置媒体库路径和扫描间隔

3. **设置爬虫**
   配置 JAVDb 爬虫，自动获取影片信息

4. **配置下载器**（可选）
   连接 Jackett 和 qBittorrent，实现种子下载

5. **启用监控**（可选）
   配置 Grafana 仪表板，监控系统运行状态

---

**祝你使用愉快！** 🚀

如有问题，请查看 [README.docker.md](./README.docker.md) 获取详细帮助。
