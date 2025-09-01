# 🛠️ NSFW-Go 开发模式指南

## 📋 开发环境配置

### 快速启动开发模式

**方法一：使用管理脚本（推荐）**
```bash
# 启动所有开发服务
./nsfw-go.sh start

# 检查服务状态
./nsfw-go.sh status

# 查看日志
./nsfw-go.sh logs

# 停止所有服务
./nsfw-go.sh stop
```

**方法二：手动启动**
1. **启动基础服务**（仅数据库和Redis）
   ```bash
   docker compose -f docker-compose.dev.yml up -d
   ```

2. **启动 Go 后端开发模式**（热重载）
   ```bash
   make dev
   ```

3. **前端文件直接修改**
   - 前端文件位于 `web/dist/` 目录
   - 修改后直接刷新浏览器即可看到效果
   - 无需重新构建 Docker 镜像

### 服务端口配置

| 服务 | 端口 | 说明 |
|------|------|------|
| **API 服务** | 8080 | Go 后端应用，支持热重载 |
| **PostgreSQL** | 5433 | 数据库服务（避免与系统端口冲突） |
| **Redis** | 6380 | 缓存服务（避免与系统端口冲突） |

### 访问地址

- 🌐 **主页**: http://localhost:8080
- 🔍 **搜索页面**: http://localhost:8080/search.html（已优化布局）
- 📁 **本地影片**: http://localhost:8080/local-movies.html
- 🏆 **排行榜**: http://localhost:8080/rankings.html
- ⚙️ **配置**: http://localhost:8080/config.html
- 📊 **API 统计**: http://localhost:8080/api/v1/stats
- ✅ **健康检查**: http://localhost:8080/health

## 🔧 开发工作流

### 修改前端代码
```bash
# 直接编辑前端文件
vim web/dist/search.html
vim web/dist/static/js/search.js
vim web/dist/static/css/style.css

# 刷新浏览器即可看到效果
```

### 修改后端代码
```bash
# 后端代码会自动热重载
vim internal/api/handlers/search.go
vim internal/service/javdb_service.go

# 代码会自动重新编译和重启
```

### 常用命令
```bash
# 重新构建
make build

# 格式化代码
make fmt

# 代码检查
make lint

# 运行测试
make test

# 查看开发服务日志（如果需要）
# 开发模式已在前台运行，直接查看终端输出即可
```

## 🚀 生产环境打包

当开发完成后，需要构建 Docker 镜像用于生产环境：

```bash
# 停止开发环境
Ctrl+C  # 停止 make dev
docker compose -f docker-compose.dev.yml down

# 构建生产镜像
docker compose build --no-cache

# 启动生产环境
docker compose up -d
```

## 📝 注意事项

1. **端口冲突处理**: 开发环境使用 5433(PostgreSQL) 和 6380(Redis) 避免与系统服务冲突
2. **热重载**: 后端 Go 代码修改后自动重启，前端文件修改后刷新浏览器即可
3. **数据持久化**: 数据库和Redis数据通过Docker volume持久化
4. **配置文件**: `config.yaml` 配置开发环境的数据库和Redis端口
5. **日志查看**: 开发模式在前台运行，所有日志直接输出到终端

## 🐛 故障排除

### 端口占用问题
```bash
# 检查端口占用
netstat -tlnp | grep -E ':(8080|5433|6380)'

# 如果端口被占用，修改配置文件中的端口号
vim config.yaml
vim docker-compose.dev.yml
```

### 数据库连接问题
```bash
# 检查数据库容器状态
docker compose -f docker-compose.dev.yml ps

# 测试数据库连接
docker compose -f docker-compose.dev.yml exec postgres psql -U nsfw -d nsfw_db -c "SELECT 1;"
```

### 热重载不工作
```bash
# 检查 Air 工具是否安装
make install-tools

# 手动重启开发服务
# Ctrl+C 停止当前服务，然后重新运行
make dev
```

---

**🎯 开发环境已就绪！现在可以高效地进行前后端开发了！**