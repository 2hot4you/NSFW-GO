# 🚀 NSFW-Go 部署指南

## 📋 概述

NSFW-Go 现在提供两种运行方式：

1. **开发模式**: 使用 `nsfw-dev.sh` 脚本进行开发
2. **生产部署**: 使用 SystemD 服务进行生产环境部署

## 🛠️ 开发模式

### 使用 nsfw-dev.sh

```bash
# 启动开发环境（PostgreSQL + Redis + Go热重载）
./nsfw-dev.sh start

# 查看服务状态
./nsfw-dev.sh status

# 查看服务日志
./nsfw-dev.sh logs

# 停止开发环境
./nsfw-dev.sh stop

# 重启开发环境
./nsfw-dev.sh restart
```

### 开发模式特点
- ✅ 自动热重载 (Air)
- ✅ 开发调试日志
- ✅ Docker 服务自动管理
- ✅ 实时代码变更检测

## 🏭 生产部署

### 使用 SystemD 服务 (推荐)

#### 1. 安装服务

```bash
# 使用 root 权限安装
sudo ./scripts/install-service.sh install
```

安装过程会自动：
- 构建生产版本二进制文件
- 创建 SystemD 服务文件
- 设置目录权限
- 启动服务并设置开机自启

#### 2. 服务管理

```bash
# 启动服务
sudo systemctl start nsfw-go

# 停止服务
sudo systemctl stop nsfw-go

# 重启服务
sudo systemctl restart nsfw-go

# 查看服务状态
sudo systemctl status nsfw-go

# 查看实时日志
sudo journalctl -u nsfw-go -f

# 查看历史日志
sudo journalctl -u nsfw-go

# 禁用开机自启
sudo systemctl disable nsfw-go

# 启用开机自启
sudo systemctl enable nsfw-go
```

#### 3. 卸载服务

```bash
sudo ./scripts/install-service.sh uninstall
```

### 生产部署特点
- ✅ SystemD 管理，系统级服务
- ✅ 开机自启动
- ✅ 自动重启（故障恢复）
- ✅ 系统日志集成
- ✅ 安全权限控制
- ✅ 资源限制保护

## 🔧 配置管理

### 配置文件位置
- 开发模式: `config.yaml`
- 生产模式: `/projects/NSFW-GO/config.yaml`

### 数据库配置优先级
1. **数据库存储配置** (最高优先级)
2. 配置文件 (`config.yaml`)
3. 默认配置

### 配置同步命令
```bash
# 同步配置到数据库
make config-sync

# 显示数据库配置
make config-show

# 备份数据库配置
make config-backup
```

## 🌐 访问地址

无论开发还是生产模式，访问地址相同：

- **主页**: http://localhost:8080
- **搜索页**: http://localhost:8080/search.html
- **本地影片**: http://localhost:8080/local-movies.html
- **排行榜**: http://localhost:8080/rankings.html
- **配置页**: http://localhost:8080/config.html
- **API 统计**: http://localhost:8080/api/v1/stats
- **健康检查**: http://localhost:8080/health

## 📊 服务依赖

### 必需服务
- **PostgreSQL**: 端口 5433
- **Redis**: 端口 6380

### 外部服务 (可选)
- **qBittorrent**: http://10.10.10.200:8085
- **Jackett**: http://10.10.10.200:9117
- **Telegram Bot**: 用于通知

## 🔍 故障排查

### 开发模式问题
```bash
# 检查服务状态
./nsfw-dev.sh status

# 查看详细日志
./nsfw-dev.sh logs

# 重启所有服务
./nsfw-dev.sh restart
```

### 生产模式问题
```bash
# 检查服务状态
sudo systemctl status nsfw-go

# 查看错误日志
sudo journalctl -u nsfw-go --since "10 minutes ago"

# 检查配置文件
sudo /projects/NSFW-GO/bin/nsfw-go -config /projects/NSFW-GO/config.yaml -check
```

### 常见问题

1. **端口被占用**
   ```bash
   # 查看端口占用
   sudo netstat -tlnp | grep :8080
   sudo lsof -i :8080
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker ps | grep postgres
   make db-check
   ```

3. **权限问题**
   ```bash
   # 修复文件权限
   sudo chown -R root:root /projects/NSFW-GO
   sudo chmod +x /projects/NSFW-GO/bin/nsfw-go
   ```

## 🎯 推荐使用方式

### 开发环境
使用 `nsfw-dev.sh` 进行日常开发，享受热重载和调试功能。

### 生产环境
使用 SystemD 服务部署，获得企业级的稳定性和可维护性。

### 快速切换
```bash
# 停止开发模式
./nsfw-dev.sh stop

# 安装生产服务
sudo ./scripts/install-service.sh install

# 或者相反：卸载生产服务，启动开发模式
sudo ./scripts/install-service.sh uninstall
./nsfw-dev.sh start
```

## 📝 注意事项

1. **服务冲突**: 开发模式和生产模式不能同时运行（端口冲突）
2. **权限管理**: 生产模式需要 root 权限进行服务管理
3. **数据持久化**: 两种模式共享相同的数据库和配置
4. **日志轮转**: 生产模式使用 systemd 日志，自动轮转
5. **资源监控**: 生产模式有内置的资源限制和监控

通过这种方式，你可以在开发时享受便捷的热重载，在生产环境获得企业级的服务管理能力！🎉