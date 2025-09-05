# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

本文件为 Claude Code (claude.ai/code) 在此代码仓库中工作时提供指导。

## 语言和回复风格
CLI 窗口的回答以及代码中的注释，可以添加部分 emoji，以中文的形式回答我。并且保持专业性，仔细思考我的上文，如果我设计的有问题，及时指出。并告诉我下一步应该做什么。

## 核心命令

### 开发工作流
```bash
make dev              # 开发模式（热重载，最重要）
make watch            # 替代热重载模式（如果 make dev 不可用）
make build           # 构建应用
make run             # 运行构建后的应用  
make test            # 运行测试
make lint            # 运行 golangci-lint（自动安装）
make fmt             # 格式化代码（gofmt + goimports）
```

### 开发模式热重载
开发时使用 `make dev` 启动热重载，它会：
- 自动检测 Go 文件变化
- 重新编译并重启服务
- 保持数据库连接和 Redis 缓存
- 支持调试模式和详细日志

如果 Air 工具未安装，`make dev` 会自动安装。

### 数据库操作
```bash
make migrate         # 执行数据库迁移
make db-reset        # 重置数据库（需确认）
make db-check        # 检查数据库连接
```

### 服务管理
```bash
make compose         # 通过 docker-compose 启动所有服务
make compose-down    # 停止所有服务
make logs           # 查看容器日志
make status         # 检查项目状态
```

### 代码质量
```bash
make lint           # 运行 golangci-lint（自动安装）
make fmt            # 格式化代码（gofmt + goimports）
make test           # 运行所有测试
make bench          # 运行基准测试
make install-tools  # 安装开发工具（air, golangci-lint, goimports）
```

### 单独测试命令
```bash
# 运行特定包的测试
go test -v ./internal/service/...
go test -v ./internal/repo/...

# 运行特定测试函数
go test -v ./internal/service/ -run TestLocalMovieService

# 运行基准测试
go test -bench=. -benchmem ./internal/service/
```

## 架构概览

### 整洁架构实现
本项目遵循整洁架构，具有清晰的关注点分离：

1. **处理器层** (`internal/api/handlers/`) - HTTP 请求/响应处理
2. **服务层** (`internal/service/`) - 业务逻辑层  
3. **仓储层** (`internal/repo/`) - 数据访问抽象
4. **模型层** (`internal/model/`) - GORM 领域实体

### 核心组件

**API 层 (`internal/api/`)**
- 使用 Gin 框架的 RESTful 设计
- 在 `internal/api/routes/` 中按功能分组路由
- 标准化错误响应和中间件

**数据层 (`internal/model/`)**
- GORM ORM 与 PostgreSQL 后端
- 复杂关系：影片 ↔ 演员、制片厂、系列、标签
- 自定义类型：StringArray, Int64Array 用于 PostgreSQL 数组
- 启用软删除以便数据恢复

**服务层 (`internal/service/`)**
- **扫描服务**：自动化本地媒体库扫描（15分钟间隔）
- **排行榜服务**：JAVDb 排行榜爬虫，定时调度
- **搜索服务**：JAVDb 在线搜索集成
- **配置服务**：动态配置管理，支持热重载

**爬虫引擎 (`internal/crawler/`)**
- 基于 Colly 框架的专业网页抓取
- 多站点支持（JAVDb, JAVLibrary 可扩展）
- 速率限制、代理支持、重试机制
- 定时任务：每日 12:00 排行榜爬取，每小时本地检查

### 数据库架构
- **主要模型**：Movie, Actress, Studio, Series, Tag, LocalMovie, Ranking
- **高级功能**：PostgreSQL 数组、JSONB 字段、全文搜索、审计字段
- **性能优化**：响应时间从 30秒 优化到 3毫秒（提升 10,000 倍）

## 配置管理

### 环境搭建要求
```bash
# 1. 复制配置模板
cp configs/config.example.yaml config.yaml

# 2. 启动服务（PostgreSQL, Redis 等）
docker compose up -d

# 3. 初始化数据库
make migrate

# 4. 运行开发服务器
make dev
```

### 关键配置节
- **数据库**：PostgreSQL 连接设置，连接池配置
- **媒体**：本地库路径 `/MediaCenter/NSFW/Hub/#Done`，扫描间隔
- **爬虫**：JAVDb URLs，调度设置，速率限制，代理设置
- **服务**：Redis 缓存，API 超时，CORS 设置

## 开发模式

### 添加新功能
1. **模型层**：在 `internal/model/` 中添加 GORM 标签
2. **仓储层**：在 `internal/repo/` 中实现数据访问
3. **服务层**：在 `internal/service/` 中实现业务逻辑
4. **处理器层**：在 `internal/api/handlers/` 中实现 HTTP 层
5. **路由层**：在 `internal/api/routes/routes.go` 中注册

### 数据库变更
- 将迁移文件添加到 `migrations/` 目录
- 运行 `make migrate` 应用变更
- 使用 `make db-reset` 完全重置（需确认）

### 测试策略
- 为服务和仓储编写单元测试
- 为 API 端点编写集成测试
- 使用 `make test` 运行所有测试
- 可用 `make bench` 进行基准测试

## 端口配置
**重要**：本应用运行在 8080 端口。所有服务端口配置：
- **API 服务器**：8080（主要入口）
- **PostgreSQL**：5433（避免与现有服务冲突）
- **Redis**：6380（避免与现有服务冲突）
- **Nginx**：80/443（反向代理，可选）
- **Web 界面**：http://localhost:8080

## 当前状态

### 已完成功能
- 核心基础设施（Go 1.23, PostgreSQL, Redis, Docker）
- 本地媒体库管理，支持自动化扫描（15分钟间隔）
- JAVDb 集成（搜索、排行榜、定时爬取）
- 现代化 Web 界面（深色主题、响应式设计）
- 性能优化（API 响应：30秒 → 3毫秒，提升 10,000 倍）
- 种子下载集成（Jackett + qBittorrent 支持）

### 开发中功能
- 增强的种子搜索和下载管理
- 数据库存储的配置管理系统
- 高级筛选和搜索功能

### 重要 API 端点
```bash
# 本地媒体
GET /api/v1/local/movies    # 分页获取本地影片列表
GET /api/v1/local/stats     # 本地统计信息
POST /api/v1/local/scan     # 触发手动扫描

# JAVDb 集成  
GET /api/v1/search/javdb    # 在线搜索 JAVDb
GET /api/v1/rankings        # 获取排行榜数据
POST /api/v1/rankings/crawl # 触发手动爬取

# 种子下载集成
GET /api/v1/torrents/search # 通过 Jackett 搜索种子
POST /api/v1/torrents/download # 添加种子到 qBittorrent

# 配置管理
GET /api/v1/config          # 获取当前配置
PUT /api/v1/config          # 更新配置
POST /api/v1/config/test    # 测试配置设置

# 系统
GET /api/v1/stats           # 系统级统计信息
GET /health                 # 健康检查
```

## 前端架构

### 技术栈
- 原生 JavaScript（无构建过程）
- Tailwind CSS 样式框架
- FontAwesome 图标库
- 深色主题配玻璃态效果

### 页面组织
- `web/dist/index.html` - 主仪表板
- `web/dist/search.html` - 统一搜索界面（本地 + JAVDb）
- `web/dist/local-movies.html` - 本地媒体库管理
- `web/dist/rankings.html` - JAVDb 排行榜显示
- `web/dist/config.html` - 系统配置面板
- `web/dist/downloads.html` - 种子下载管理
- `web/dist/test-connections.html` - 配置测试工具

## 性能特征

### 优化成果
- API 响应时间：30秒 → 3毫秒（提升 10,000 倍）
- 文件系统扫描：缓存数据 vs 实时扫描
- 数据库查询：高级索引和查询优化
- 缓存：Redis 集成，频繁访问数据缓存

### 可扩展性特征
- 连接池（可配置的数据库连接）
- 并发爬取（可配置的并行抓取）
- 速率限制（尊重外部 API 使用）
- 后台任务（通过 goroutines 的定时操作）

## Docker 服务

通过 `docker compose up -d` 可用的服务：
- **PostgreSQL 15**：主数据库（端口 5433）
- **Redis 7**：缓存层（端口 6380）
- **API 服务**：Go 应用程序（端口 8080）
- **Nginx**：反向代理（端口 80/443）
- **pgAdmin**：数据库管理界面（localhost:5050，使用 profile）
- **Redis Commander**：Redis 管理界面（localhost:8081，使用 profile）
- **Prometheus/Grafana**：监控栈（使用 profile）

## 开发工具

### 自动安装工具
Makefile 自动安装所需工具：
- **Air**：开发模式热重载
- **golangci-lint**：代码检查
- **goimports**：导入格式化

### 监控和调试
- 内置 pprof 支持（`make profile`, `make memprofile`）
- 结构化输出的全面日志记录
- 监控用的健康检查端点
- 实时统计和状态指示器

## 关键开发说明

### 数据库迁移
- 新的迁移应按顺序编号添加到 `migrations/` 目录
- 提交前务必用 `make migrate` 测试迁移
- 谨慎使用 `make db-reset` - 会删除所有数据

### 服务依赖
- 运行 `make dev` 前务必先用 `make compose` 或 `docker compose up -d` 启动服务
- 使用 `make db-check` 检查数据库连接状态
- API 服务器启动前必须保证数据库和 Redis 运行
- 如果端口冲突，已调整为：PostgreSQL(5433), Redis(6380), API(8080)

### 配置管理
- 系统现在支持基于文件（config.yaml）和数据库存储的配置
- 数据库配置优先于文件配置
- 应用前使用配置测试端点验证设置
- 配置更新 API：
  ```bash
  # 获取当前配置（包含数据库和文件合并结果）
  curl -s "http://localhost:8080/api/v1/config"
  
  # 更新配置（会保存到数据库并创建备份）
  curl -s -X POST "http://localhost:8080/api/v1/config" \
    -H "Content-Type: application/json" \
    -d '{"torrent": {"download_path": "/volume1/media/PornDB/Downloads"}}'
  ```

### 种子下载集成
- 需要外部 Jackett 和 qBittorrent 实例
- **重要**：qBittorrent 运行在 Synology NAS (10.10.10.200:8085)，不是本地容器
- 在 config.yaml 的 torrent 节中配置端点，但**数据库配置优先于文件配置**
- 使用前通过配置面板测试连接性
- 下载路径必须是 Synology 可访问的路径（如 `/volume1/media/PornDB/Downloads`）
- 避免使用本地路径（如 `/media/PornDB/Downloads`）否则会出现权限拒绝错误

### 本地媒体扫描
- 自动扫描每 15 分钟运行一次
- 可通过 API 触发手动扫描或启动时运行
- 媒体路径可通过 config.yaml media.base_path 配置

### 代码提交前检查
在提交代码前，请确保运行：
```bash
make fmt     # 格式化代码
make lint    # 代码质量检查
make test    # 运行所有测试
```

### 性能监控
- 使用 `make profile` 进行 CPU 性能分析
- 使用 `make memprofile` 进行内存分析
- 生产环境可访问 `/debug/pprof/` 端点进行实时性能监控

## 常见问题排查

### 种子下载问题
**权限拒绝错误**：`file_open error: Permission denied`
- **原因**：qBittorrent 在 Synology NAS 上，无法访问本地路径
- **解决**：确保下载路径使用 Synology 路径格式（如 `/volume1/media/PornDB/Downloads`）
- **检查**：通过 `/api/v1/config` 确认 `torrent.download_path` 和 `torrent.qbittorrent.download_dir` 配置正确

**种子添加失败**：`远程内容在服务器上未找到（404）`
- **原因**：Jackett 种子链接有时效性，可能已过期
- **解决**：重新搜索获取最新链接，或检查 Jackett 服务状态
- **API**：`GET /api/v1/torrents/search?q=番号` 获取新链接

### 服务启动问题
**数据库连接失败**
- **检查**：确保 PostgreSQL 在端口 5433 运行
- **命令**：`make db-check` 验证连接
- **重启**：`make compose` 重新启动所有服务

**配置不生效**
- **原因**：数据库配置覆盖文件配置
- **查看**：`curl -s "http://localhost:8080/api/v1/config"` 查看生效配置
- **更新**：通过 API 而不是修改 config.yaml 文件