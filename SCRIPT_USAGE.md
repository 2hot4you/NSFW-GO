# 🚀 NSFW-Go 管理脚本使用指南

## 📋 脚本功能概述

`nsfw-go.sh` 是一个功能强大的开发环境管理脚本，提供了启动、停止、状态检查和日志查看等功能。

## 🎯 主要特性

- ✅ **智能服务检测** - 自动检测 PostgreSQL、Redis、Go 开发服务器状态
- ✅ **端口状态监控** - 实时检查端口占用和进程信息
- ✅ **Docker 容器管理** - 管理开发环境所需的容器服务
- ✅ **多源日志查看** - 支持查看不同服务的日志信息
- ✅ **彩色输出界面** - 清晰的颜色编码状态提示
- ✅ **跨目录执行** - 可从任意目录执行，自动定位项目根目录

## 📖 使用方法

### 基本命令格式
```bash
./nsfw-go.sh {start|stop|restart|status|logs|help}
```

### 详细命令说明

#### 1. 启动开发环境
```bash
./nsfw-go.sh start
```
**功能：**
- 启动 PostgreSQL 和 Redis Docker 容器
- 启动 Go 开发服务器（热重载模式）
- 自动检查服务启动状态
- 显示所有访问地址

**输出示例：**
```
================================
 🚀 NSFW-Go 管理脚本
================================

[INFO] 启动开发环境...
[INFO] 启动 PostgreSQL 和 Redis 服务...
[SUCCESS] Docker服务启动成功
[INFO] 启动 Go 开发服务器...
[SUCCESS] Go 开发服务器启动成功

🎉 所有服务启动完成！
```

#### 2. 检查服务状态
```bash
./nsfw-go.sh status
```
**功能：**
- 检查 Docker 容器运行状态
- 监控端口监听情况
- 显示进程 PID 信息
- 列出所有访问地址

**输出示例：**
```
📦 Docker 服务状态:
✅ PostgreSQL 容器 - 运行中 (Up 10 minutes (healthy))
✅ Redis 容器 - 运行中 (Up 10 minutes (healthy))

🔌 端口监听状态:
✅ PostgreSQL (端口 5433) - 运行中 [PID: 12345]
✅ Redis (端口 6380) - 运行中 [PID: 12346]
✅ API服务器 (端口 8080) - 运行中 [PID: 12347]

🚀 应用服务状态:
✅ Go 开发服务器 - 运行中 [PID: 12348] (make dev)
```

#### 3. 查看服务日志
```bash
./nsfw-go.sh logs
```
**功能：**
- 交互式选择要查看的日志类型
- 支持多种日志源
- 显示最近 50 行日志内容

**日志选项：**
1. Go 开发服务器日志
2. PostgreSQL 日志
3. Redis 日志
4. 所有 Docker 服务日志
5. 实时跟踪 Go 开发日志

#### 4. 停止开发环境
```bash
./nsfw-go.sh stop
```
**功能：**
- 停止 Go 开发服务器进程
- 停止 Air 热重载进程
- 停止 Docker 容器
- 清理临时文件

#### 5. 重启开发环境
```bash
./nsfw-go.sh restart
```
**功能：**
- 相当于执行 stop + start
- 完全重启所有服务
- 适用于配置更新后的重启

#### 6. 显示帮助信息
```bash
./nsfw-go.sh help
```
**功能：**
- 显示所有可用命令
- 列出端口配置信息
- 提供使用示例

## 🔧 服务端口配置

| 服务 | 端口 | 说明 |
|------|------|------|
| **PostgreSQL** | 5433 | 数据库服务（避免与系统端口冲突） |
| **Redis** | 6380 | 缓存服务（避免与系统端口冲突） |
| **API服务器** | 8080 | Go 后端应用，支持热重载 |

## 🌐 访问地址

启动后可通过以下地址访问：
- **主页**: http://localhost:8080
- **搜索页面**: http://localhost:8080/search.html
- **本地影片**: http://localhost:8080/local-movies.html
- **排行榜**: http://localhost:8080/rankings.html
- **配置**: http://localhost:8080/config.html
- **API 统计**: http://localhost:8080/api/v1/stats
- **健康检查**: http://localhost:8080/health

## 💡 使用技巧

### 1. 全局使用
可以创建符号链接以便全局使用：
```bash
# 创建符号链接到 /usr/local/bin
sudo ln -sf /projects/NSFW-GO/nsfw-go.sh /usr/local/bin/nsfw-go

# 然后可以在任意目录使用
nsfw-go status
nsfw-go start
```

### 2. 状态监控
使用 watch 命令实时监控服务状态：
```bash
watch -n 2 './nsfw-go.sh status'
```

### 3. 快速重启
当修改了配置文件后需要重启：
```bash
./nsfw-go.sh restart
```

### 4. 日志调试
查看实时日志进行问题调试：
```bash
./nsfw-go.sh logs
# 选择选项 5 进行实时跟踪
```

## 🐛 故障排除

### 端口占用问题
如果遇到端口被占用的错误：
```bash
# 检查端口占用情况
netstat -tlnp | grep -E ':(8080|5433|6380)'

# 或使用脚本检查
./nsfw-go.sh status
```

### 服务启动失败
```bash
# 检查 Docker 服务状态
docker compose -f docker-compose.dev.yml ps

# 查看详细错误日志
./nsfw-go.sh logs
```

### 权限问题
```bash
# 确保脚本有执行权限
chmod +x nsfw-go.sh
```

## 📋 依赖检查

脚本会自动检查以下必要依赖：
- `docker` - Docker 容器管理
- `netstat` - 端口状态检查
- `make` - 构建工具

如果缺少依赖，脚本会显示错误并提示安装。

## 🎉 总结

`nsfw-go.sh` 管理脚本大大简化了开发环境的管理工作，提供了：
- 一键启动/停止功能
- 实时状态监控
- 便捷的日志查看
- 智能错误检测
- 跨平台兼容性

使用这个脚本，你可以专注于代码开发，而不用担心环境配置的复杂性！