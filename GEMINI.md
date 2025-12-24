# GEMINI.md

本文件为 Gemini 在此代码仓库中工作时提供指导。

## 语言和回复风格
CLI 窗口的回答以及代码中的注释，请使用**中文**。可以适当添加 Emoji 以增强可读性，但需保持专业性。请仔细思考上下文，如果发现设计或指令有问题，请及时指出并提出改进建议。

## 核心命令

### 开发工作流
```bash
make dev              # 开发模式（热重载，最重要）
make watch            # 替代热重载模式（如果 make dev 不可用）
make build            # 构建应用
make run              # 运行构建后的应用  
make test             # 运行测试
make lint             # 运行 golangci-lint（自动安装）
make fmt              # 格式化代码（gofmt + goimports）
make help             # 显示所有可用命令帮助
```

### 数据库操作
```bash
make migrate          # 执行数据库迁移
make db-reset         # 重置数据库（需确认，慎用）
make db-check         # 检查数据库连接
```

### 服务管理
```bash
make compose          # 通过 docker-compose 启动所有服务
make compose-down     # 停止所有服务
make logs             # 查看容器日志
make status           # 检查项目状态
```

## 架构概览

本项目遵循整洁架构（Clean Architecture），关注点分离：

1.  **处理器层** (`internal/api/handlers/`) - 处理 HTTP 请求/响应。
2.  **服务层** (`internal/service/`) - 核心业务逻辑。
3.  **仓储层** (`internal/repo/`) - 数据访问抽象。
4.  **模型层** (`internal/model/`) - GORM 领域实体。

### 核心组件

*   **API 层**: 基于 Gin 框架，RESTful 设计。
*   **数据层**: PostgreSQL + GORM，支持数组和 JSONB 字段。
*   **服务层**:
    *   **扫描服务**: 自动扫描本地媒体库（15分钟间隔）。
    *   **排行榜服务**: JAVDb 爬虫，定时调度。
    *   **搜索服务**: JAVDb 在线搜索集成。
    *   **配置服务**: 动态配置管理。
*   **爬虫引擎**: 基于 Colly，支持多站点（JAVDb 等），具备速率限制和代理支持。

## 配置管理

### 环境搭建
1.  复制配置模板: `cp configs/config.example.yaml config.yaml`
2.  启动服务: `docker compose up -d`
3.  初始化数据库: `make migrate`
4.  运行开发服务器: `make dev`

### 关键配置
*   **数据库**: PostgreSQL 连接设置。
*   **媒体**: 本地库路径（如 `/MediaCenter/NSFW/Hub/#Done`）。
*   **爬虫**: JAVDb URLs，调度，代理。
*   **服务**: Redis 缓存，API 超时。

**注意**: 数据库中的配置优先于 `config.yaml` 文件配置。

## 开发指南

### 添加新功能步骤
1.  **模型层**: 在 `internal/model/` 定义数据结构。
2.  **仓储层**: 在 `internal/repo/` 实现数据存取。
3.  **服务层**: 在 `internal/service/` 实现业务逻辑。
4.  **处理器层**: 在 `internal/api/handlers/` 实现 HTTP 接口。
5.  **路由层**: 在 `internal/api/routes/routes.go` 注册路由。

### 端口配置
*   **API 服务器**: 8080
*   **PostgreSQL**: 5433
*   **Redis**: 6380
*   **Nginx**: 80/443 (可选)

## 前端架构
*   **技术栈**: 原生 JavaScript + Tailwind CSS + FontAwesome。
*   **页面**:
    *   `web/dist/index.html`: 主仪表板
    *   `web/dist/search.html`: 搜索界面
    *   `web/dist/local-movies.html`: 本地媒体库
    *   `web/dist/rankings.html`: 排行榜
    *   `web/dist/config.html`: 配置面板

## 常见问题排查

*   **图片 404**: 检查 URL 编码和媒体库路径配置。
*   **种子下载权限错误**: 确认 qBittorrent（如在 NAS 上）能访问配置的下载路径。
*   **配置不生效**: 检查是否被数据库配置覆盖，使用 API 查看当前生效配置。
*   **数据库连接失败**: 检查端口 5433 是否正常监听。

## 提交前检查
```bash
make fmt
make lint
make test
```
