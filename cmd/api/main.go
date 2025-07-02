package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"nsfw-go/internal/api/routes"
	"nsfw-go/internal/config"
	"nsfw-go/internal/database"
	"nsfw-go/migrations"

	"github.com/gin-gonic/gin"
)

var (
	configPath = flag.String("config", "", "配置文件路径")
	migrate    = flag.Bool("migrate", false, "执行数据库迁移")
	version    = flag.Bool("version", false, "显示版本信息")
	help       = flag.Bool("help", false, "显示帮助信息")
)

const (
	AppName    = "NSFW-Go"
	AppVersion = "v1.0.0"
	AppDesc    = "智能成人影视库管理系统"
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		showVersion()
		return
	}

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 如果指定了migrate参数，执行数据库迁移
	if *migrate {
		log.Println("开始数据库迁移...")
		dsn := cfg.Database.GetDSN()
		
		// 检查数据库连接
		if err := migrations.CheckConnection(dsn); err != nil {
			log.Fatalf("数据库连接失败: %v", err)
		}
		
		// 执行迁移
		if err := migrations.Migrate(dsn); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
		
		log.Println("✓ 数据库迁移完成")
		return
	}

	// 初始化数据库
	log.Println("初始化数据库连接...")
	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 设置路由
	routes.SetupRoutes(r, database.DB)

	// 启动服务器
	port := ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("🚀 %s %s 启动完成", AppName, AppVersion)
	log.Printf("🌐 服务器地址: http://localhost%s", port)
	log.Printf("📋 健康检查: http://localhost%s/health", port)
	log.Printf("📊 API统计: http://localhost%s/api/v1/stats", port)
	
	if err := r.Run(port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

func showHelp() {
	fmt.Printf(`%s %s - %s

使用方法:
  %s [选项]

选项:
  -config string    配置文件路径 (可选)
  -migrate          执行数据库迁移
  -version          显示版本信息
  -help             显示此帮助信息

示例:
  %s -config config.yaml              # 使用指定配置文件启动
  %s -migrate                         # 执行数据库迁移
  %s -migrate -config config.yaml     # 使用指定配置执行迁移

环境变量:
  NSFW_DATABASE_HOST=localhost         # 数据库主机
  NSFW_DATABASE_PORT=5432              # 数据库端口
  NSFW_DATABASE_USER=nsfw              # 数据库用户
  NSFW_DATABASE_PASSWORD=nsfw123       # 数据库密码
  NSFW_DATABASE_DBNAME=nsfw_db         # 数据库名称
  NSFW_REDIS_HOST=localhost            # Redis主机
  NSFW_REDIS_PORT=6379                 # Redis端口
  NSFW_SERVER_PORT=8080                # 服务器端口

配置文件:
  系统会按以下顺序查找配置文件:
  1. 命令行指定的路径
  2. ./config.yaml
  3. ./configs/config.yaml
  4. /etc/nsfw-go/config.yaml

更多信息请查看: https://github.com/your-repo/nsfw-go
`, AppName, AppVersion, AppDesc, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func showVersion() {
	fmt.Printf(`%s %s
%s

构建信息:
  Go版本: %s
  操作系统: %s
  架构: %s

作者: Your Name
项目: https://github.com/your-repo/nsfw-go
许可: MIT License
`, AppName, AppVersion, AppDesc, "go1.21+", "linux", "amd64")
}

func showConfig(cfg *config.Config) {
	fmt.Printf(`
========================================
         配置信息
========================================
服务器:
  地址: %s
  模式: %s
  CORS: %v
  Swagger: %v

数据库:
  主机: %s:%d
  数据库: %s
  用户: %s
  最大连接: %d

Redis:
  地址: %s
  数据库: %d
  连接池: %d

媒体库:
  基础路径: %s
  扫描间隔: %d 小时
  支持格式: %v

爬虫:
  代理启用: %v
  请求延时: %v
  重试次数: %d
  并发数: %d

Telegram Bot:
  启用: %v
  管理员ID: %v

安全:
  认证启用: %v
  限流: %d RPS
  JWT过期: %v

日志:
  级别: %s
  格式: %s
  输出: %s
========================================
`, 
		cfg.Server.GetAddr(),
		cfg.Server.Mode,
		cfg.Server.EnableCORS,
		cfg.Server.EnableSwagger,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.User,
		cfg.Database.MaxOpenConns,
		cfg.Redis.GetAddr(),
		cfg.Redis.DB,
		cfg.Redis.PoolSize,
		cfg.Media.BasePath,
		cfg.Media.ScanInterval,
		cfg.Media.SupportedExts,
		cfg.Crawler.ProxyEnabled,
		cfg.Crawler.RequestDelay,
		cfg.Crawler.RetryCount,
		cfg.Crawler.ConcurrentMax,
		cfg.Bot.Enabled,
		cfg.Bot.AdminIDs,
		cfg.Security.EnableAuth,
		cfg.Security.RateLimitRPS,
		cfg.Security.JWTExpiry,
		cfg.Log.Level,
		cfg.Log.Format,
		cfg.Log.Output,
	)
}
 