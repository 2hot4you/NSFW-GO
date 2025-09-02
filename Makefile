# NSFW-Go Makefile
# 用于简化开发、测试和部署流程

# 变量定义
APP_NAME = nsfw-go
VERSION = v1.0.0
BUILD_TIME = $(shell date +%Y%m%d_%H%M%S)
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION = $(shell go version | grep -o 'go[0-9]\+\.[0-9]\+\.[0-9]\+')

# 构建标志
LDFLAGS = -X main.Version=$(VERSION) \
          -X main.BuildTime=$(BUILD_TIME) \
          -X main.GitCommit=$(GIT_COMMIT) \
          -X main.GoVersion=$(GO_VERSION)

# 默认目标
.PHONY: all
all: clean build

# 帮助信息
.PHONY: help
help:
	@echo "NSFW-Go Makefile 命令:"
	@echo ""
	@echo "开发命令:"
	@echo "  make deps        安装依赖"
	@echo "  make build       构建应用"
	@echo "  make run         运行应用"
	@echo "  make dev         开发模式运行"
	@echo "  make test        运行测试"
	@echo "  make lint        代码检查"
	@echo "  make fmt         代码格式化"
	@echo ""
	@echo "数据库命令:"
	@echo "  make migrate     执行数据库迁移"
	@echo "  make db-reset    重置数据库"
	@echo "  make db-check    检查数据库连接"
	@echo ""
	@echo "配置管理:"
	@echo "  make config-sync     同步配置到数据库"
	@echo "  make config-show     显示数据库配置"
	@echo "  make config-backup   备份数据库配置"
	@echo ""
	@echo "部署命令:"
	@echo "  make docker      构建Docker镜像"
	@echo "  make compose     使用docker-compose启动"
	@echo "  make release     构建发布版本"
	@echo ""
	@echo "清理命令:"
	@echo "  make clean       清理构建文件"
	@echo "  make clean-all   清理所有文件"

# 安装依赖
.PHONY: deps
deps:
	@echo "正在安装Go依赖..."
	go mod download
	go mod tidy
	@echo "✓ 依赖安装完成"

# 构建应用
.PHONY: build
build: deps
	@echo "正在构建 $(APP_NAME)..."
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-api cmd/api/main.go
	@echo "✓ 构建完成: bin/$(APP_NAME)-api"

# 构建所有组件
.PHONY: build-all
build-all: deps
	@echo "正在构建所有组件..."
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-api cmd/api/main.go
	# go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-bot cmd/bot/main.go
	# go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-crawler cmd/crawler/main.go
	@echo "✓ 构建完成"

# 运行应用
.PHONY: run
run: build
	@echo "正在启动应用..."
	./bin/$(APP_NAME)-api -config config.yaml

# 开发模式运行
.PHONY: dev
dev:
	@echo "正在以开发模式启动..."
	go run cmd/api/main.go -config config.yaml

# 运行测试
.PHONY: test
test:
	@echo "正在运行测试..."
	go test -v ./...
	@echo "✓ 测试完成"

# 运行基准测试
.PHONY: bench
bench:
	@echo "正在运行基准测试..."
	go test -bench=. -benchmem ./...

# 代码检查
.PHONY: lint
lint:
	@echo "正在进行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，正在安装..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
		golangci-lint run; \
	fi

# 代码格式化
.PHONY: fmt
fmt:
	@echo "正在格式化代码..."
	go fmt ./...
	goimports -w .
	@echo "✓ 代码格式化完成"

# 执行数据库迁移
.PHONY: migrate
migrate: build
	@echo "正在执行数据库迁移..."
	./bin/$(APP_NAME)-api -migrate -config config.yaml
	@echo "✓ 数据库迁移完成"

# 重置数据库
.PHONY: db-reset
db-reset:
	@echo "警告: 这将删除所有数据!"
	@read -p "确认重置数据库? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		echo "正在重置数据库..."; \
		if command -v docker-compose >/dev/null 2>&1; then \
			docker-compose exec postgres psql -U nsfw -d nsfw_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
		else \
			docker compose exec postgres psql -U nsfw -d nsfw_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
		fi; \
		make migrate; \
	else \
		echo "操作已取消"; \
	fi

# 检查数据库连接
.PHONY: db-check
db-check:
	@echo "正在检查数据库连接..."
	@if (command -v docker-compose >/dev/null 2>&1 && docker-compose ps postgres | grep -q "Up") || (docker compose ps postgres | grep -q "Up"); then \
		echo "✓ PostgreSQL 容器正在运行"; \
		if command -v docker-compose >/dev/null 2>&1; then \
			docker-compose exec postgres pg_isready -U nsfw; \
		else \
			docker compose exec postgres pg_isready -U nsfw; \
		fi; \
	else \
		echo "✗ PostgreSQL 容器未运行"; \
		exit 1; \
	fi

# 配置管理命令
.PHONY: config-sync
config-sync:
	@echo "正在构建配置同步工具..."
	@go build -o bin/config-sync cmd/config-sync/main.go
	@echo "正在同步配置到数据库..."
	@./bin/config-sync -config config.yaml -mode sync -backup
	@echo "✓ 配置同步完成"

.PHONY: config-show
config-show:
	@echo "正在构建配置同步工具..."
	@go build -o bin/config-sync cmd/config-sync/main.go
	@echo "数据库配置项:"
	@./bin/config-sync -config config.yaml -mode show

.PHONY: config-backup
config-backup:
	@echo "正在构建配置同步工具..."
	@go build -o bin/config-sync cmd/config-sync/main.go
	@echo "正在备份数据库配置..."
	@./bin/config-sync -config config.yaml -backup
	@echo "✓ 配置备份完成"

# 构建Docker镜像
.PHONY: docker
docker:
	@echo "正在构建Docker镜像..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "✓ Docker镜像构建完成"

# 使用docker-compose启动
.PHONY: compose
compose:
	@echo "正在启动docker-compose..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose up -d; \
	else \
		docker compose up -d; \
	fi
	@echo "✓ 服务已启动"
	@echo "等待服务就绪..."
	@sleep 5
	@make db-check

# 停止docker-compose
.PHONY: compose-down
compose-down:
	@echo "正在停止服务..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose down; \
	else \
		docker compose down; \
	fi
	@echo "✓ 服务已停止"

# 查看日志
.PHONY: logs
logs:
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose logs -f; \
	else \
		docker compose logs -f; \
	fi

# 构建发布版本
.PHONY: release
release: clean
	@echo "正在构建发布版本..."
	mkdir -p release
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o release/$(APP_NAME)-linux-amd64 cmd/api/main.go
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS) -s -w" -o release/$(APP_NAME)-linux-arm64 cmd/api/main.go
	
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o release/$(APP_NAME)-windows-amd64.exe cmd/api/main.go
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o release/$(APP_NAME)-darwin-amd64 cmd/api/main.go
	
	# macOS arm64
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS) -s -w" -o release/$(APP_NAME)-darwin-arm64 cmd/api/main.go
	
	# 复制配置文件
	cp config.yaml release/
	cp README.md release/
	
	@echo "✓ 发布版本构建完成，文件位于 release/ 目录"

# 清理构建文件
.PHONY: clean
clean:
	@echo "正在清理构建文件..."
	rm -rf bin/
	rm -rf release/
	go clean -cache
	@echo "✓ 清理完成"

# 清理所有文件
.PHONY: clean-all
clean-all: clean
	@echo "正在清理所有文件..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose down -v; \
	else \
		docker compose down -v; \
	fi
	docker rmi $(APP_NAME):latest $(APP_NAME):$(VERSION) 2>/dev/null || true
	rm -rf logs/
	@echo "✓ 所有文件已清理"

# 生成API文档
.PHONY: docs
docs:
	@echo "正在生成API文档..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs/swagger; \
	else \
		echo "swag 未安装，正在安装..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g cmd/api/main.go -o docs/swagger; \
	fi
	@echo "✓ API文档生成完成"

# 安装开发工具
.PHONY: install-tools
install-tools:
	@echo "正在安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✓ 开发工具安装完成"

# 监控模式运行（需要air工具）
.PHONY: watch
watch:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air 未安装，正在安装..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# 性能分析
.PHONY: profile
profile:
	@echo "启动性能分析服务器..."
	go tool pprof http://localhost:8080/debug/pprof/profile

# 内存分析
.PHONY: memprofile
memprofile:
	@echo "启动内存分析..."
	go tool pprof http://localhost:8080/debug/pprof/heap

# 显示项目状态
.PHONY: status
status:
	@echo "检查项目状态..."
	go run scripts/check_status.go

# 显示项目信息
.PHONY: info
info:
	@echo "========================================="
	@echo "           项目信息"
	@echo "========================================="
	@echo "项目名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo "Go版本: $(GO_VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo ""
	@echo "Docker容器状态:"
	@docker-compose ps 2>/dev/null || echo "Docker Compose 未运行"
	@echo ""
	@echo "文件大小:"
	@if [ -f "bin/$(APP_NAME)-api" ]; then ls -lh bin/$(APP_NAME)-api; else echo "二进制文件未构建"; fi 