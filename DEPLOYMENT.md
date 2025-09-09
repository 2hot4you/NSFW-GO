# 🚀 NSFW-Go 长期运行部署指南

## 📋 概览

本文档介绍如何在当前环境中长期运行 NSFW-Go 系统，无需 Docker Compose，使用 systemd 管理服务。

## 🛠️ 已完成的部署

### SystemD 服务配置
- ✅ `nsfw-postgres.service` - PostgreSQL 数据库服务
- ✅ `nsfw-redis.service` - Redis 缓存服务  
- ✅ `nsfw-go.service` - 主应用服务
- ✅ 所有服务已设置开机自启

### 管理脚本
- ✅ `scripts/nsfw-manager.sh` - 综合服务管理脚本
- ✅ `scripts/start-services.sh` - 服务启动脚本
- ✅ `scripts/stop-services.sh` - 服务停止脚本
- ✅ `scripts/status-check.sh` - 状态检查脚本
- ✅ 全局命令 `nsfw` 可用

## 🎯 使用方法

### 基础命令
```bash
# 使用全局命令
nsfw start    # 启动所有服务
nsfw stop     # 停止所有服务  
nsfw restart  # 重启所有服务
nsfw status   # 查看服务状态
nsfw logs     # 查看服务日志
nsfw build    # 重新构建应用
nsfw migrate  # 执行数据库迁移

# 或直接使用脚本
/root/nsfw-go/scripts/nsfw-manager.sh status
```

### SystemD 原生命令
```bash
# 服务控制
sudo systemctl start nsfw-go
sudo systemctl stop nsfw-go
sudo systemctl restart nsfw-go
sudo systemctl status nsfw-go

# 查看日志
journalctl -u nsfw-go -f
journalctl -u nsfw-postgres -f
journalctl -u nsfw-redis -f

# 开机自启管理
sudo systemctl enable nsfw-go
sudo systemctl disable nsfw-go
```

## 🔄 服务启动顺序

系统会自动按以下顺序启动：
1. **Docker** → 2. **PostgreSQL** → 3. **Redis** → 4. **NSFW-Go 主应用**

依赖关系已在 systemd 配置中正确设置。

## 📊 当前状态

运行 `nsfw status` 查看详细状态：
- 📦 **Docker**: 运行中
- 🗄️ **PostgreSQL**: 运行中 (端口 5432)
- 🔴 **Redis**: 运行中 (端口 6379)  
- 🎬 **NSFW-Go**: 运行中 (端口 8080)
- 🔄 **开机自启**: 全部已启用

## 🚀 访问应用

- **主界面**: http://localhost:8080
- **健康检查**: http://localhost:8080/health
- **API 文档**: http://localhost:8080/swagger (如果启用)

## 🔧 故障排除

### 服务无法启动
```bash
# 查看具体错误
journalctl -u nsfw-go --no-pager -l

# 重新加载配置
sudo systemctl daemon-reload

# 手动启动调试
/root/nsfw-go/bin/nsfw-go-api
```

### 数据库连接问题
```bash
# 检查数据库状态
docker exec nsfw-postgres pg_isready -U nsfw -d nsfw_db

# 重启数据库服务
sudo systemctl restart nsfw-postgres
```

### 端口冲突
```bash
# 检查端口占用
sudo netstat -tulpn | grep :8080
sudo netstat -tulpn | grep :5432
sudo netstat -tulpn | grep :6379
```

## 📁 重要文件位置

### 配置文件
- 主配置: `/root/nsfw-go/config.yaml`
- SystemD 服务: `/etc/systemd/system/nsfw-*.service`

### 日志文件
- SystemD 日志: `journalctl -u nsfw-go`
- 应用日志: 根据 config.yaml 中的 log 配置

### 数据持久化
- PostgreSQL 数据: Docker volume `nsfw_postgres_data`
- Redis 数据: Docker volume `nsfw_redis_data`  
- 媒体文件: `/MediaCenter/NSFW/Hub/#Done` (可配置)

## 🔐 安全注意事项

- 服务以 root 用户运行 (生产环境建议使用专用用户)
- 数据库密码在 config.yaml 中明文存储
- 默认无认证访问 (可通过 config.yaml 启用)
- Redis 无密码访问 (可配置密码)

## 📈 性能监控

### 资源使用
```bash
# 查看服务资源占用
nsfw status

# 系统资源监控
htop
iotop
nethogs
```

### 应用监控
- 健康检查: `curl http://localhost:8080/health`
- 统计信息: `curl http://localhost:8080/api/v1/stats`

## 🔄 更新部署

### 更新应用代码
```bash
cd /root/nsfw-go
git pull origin main
nsfw build    # 重新构建
nsfw restart  # 重启服务
```

### 数据库迁移
```bash
nsfw migrate  # 执行新的迁移
```

## 🆘 紧急恢复

### 完全重启
```bash
nsfw stop
sudo systemctl restart docker
nsfw start
```

### 重置数据库 (⚠️ 会丢失数据)
```bash
nsfw stop
docker volume rm nsfw_postgres_data
nsfw start
nsfw migrate
```

---

## 📞 下一步建议

1. **监控设置**: 考虑设置 Prometheus + Grafana 监控
2. **备份策略**: 定期备份数据库和配置文件  
3. **日志轮转**: 配置 logrotate 管理应用日志
4. **安全加固**: 设置防火墙规则，启用认证
5. **性能调优**: 根据实际负载调整配置参数

系统现在已经完全配置好，可以长期稳定运行！🎉