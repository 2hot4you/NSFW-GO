package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nsfw-go/internal/api/handlers"
	"nsfw-go/internal/crawler"
	"nsfw-go/internal/repo"
	"nsfw-go/internal/service"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL驱动
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	"gorm.io/gorm"
)

// LocalMovieAdapter 适配器，让repo.LocalMovieRepository兼容service.LocalMovieRepository
type LocalMovieAdapter struct {
	repo repo.LocalMovieRepository
}

func (a *LocalMovieAdapter) SearchByCode(code string) (*service.LocalMovie, error) {
	movie, err := a.repo.SearchByCode(code)
	if err != nil {
		return nil, err
	}

	// 转换为service包的LocalMovie结构
	return &service.LocalMovie{
		ID:    movie.ID,
		Code:  movie.Code,
		Title: movie.Title,
	}, nil
}

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// 清空数据库中的模拟数据
	clearDatabaseData(db)

	// 创建仓库
	localMovieRepo := repo.NewLocalMovieRepository(db)
	rankingRepo := repo.NewRankingRepository(db)
	rankingDownloadTaskRepo := repo.NewRankingDownloadTaskRepository(db)
	subscriptionRepo := repo.NewSubscriptionRepository(db)

	// 创建爬虫配置
	crawlerConfig := &crawler.CrawlerConfig{
		UserAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		},
		ProxyEnabled:  false,
		RequestDelay:  2 * time.Second,
		RetryCount:    3,
		Timeout:       30 * time.Second,
		ConcurrentMax: 3,
	}

	// 稍后创建排行榜服务（需要等待 logService 创建）

	// 稍后创建JAVDb搜索服务（需要等待 logService 创建）

	// 创建配置服务
	configService := service.NewConfigService("config.yaml")

	// 创建配置存储服务来读取数据库配置
	configStoreService := service.NewConfigStoreService()
	
	log.Printf("🔧 开始从数据库加载服务配置...")

	// 从数据库获取 Telegram 配置并创建服务
	var telegramService *service.TelegramService
	if botEnabled, err := configStoreService.GetConfig("bot.enabled"); err == nil {
		if enabled := botEnabled.Bool(); enabled {
			if botToken, err := configStoreService.GetConfig("bot.token"); err == nil {
				token := botToken.String()
				// 去除可能的双引号
				token = strings.Trim(token, "\"")
				if token != "" {
					// 获取管理员ID列表
					var chatID string
					if adminIds, err := configStoreService.GetConfig("bot.admin_ids"); err == nil {
						adminIdsStr := adminIds.String()
						// 去除外层双引号
						adminIdsStr = strings.Trim(adminIdsStr, "\"")
						log.Printf("🔍 解析管理员ID字符串: %s", adminIdsStr)
						
						var ids []float64
						if err := json.Unmarshal([]byte(adminIdsStr), &ids); err == nil && len(ids) > 0 {
							chatID = fmt.Sprintf("%.0f", ids[0])
							log.Printf("🆔 解析得到聊天ID: %s", chatID)
						} else {
							log.Printf("❌ 解析管理员ID失败: %v", err)
						}
					}
					telegramService = service.NewTelegramService(token, chatID, enabled)
					log.Printf("✅ Telegram 服务已创建，Token: %s..., 聊天ID: %s", token[:10], chatID)
				} else {
					log.Printf("⚠️  Telegram token 为空，跳过服务创建")
				}
			}
		} else {
			log.Printf("⚠️  Telegram 服务未启用")
		}
	}

	// 从数据库获取种子下载配置并创建服务
	localMovieAdapter := &LocalMovieAdapter{repo: localMovieRepo}
	
	// 获取 Jackett 配置
	jackettHost := "http://your-jackett-server:9117"
	jackettAPIKey := "your_jackett_api_key"
	if config, err := configStoreService.GetConfig("torrent.jackett.host"); err == nil {
		jackettHost = strings.Trim(config.String(), "\"")
	}
	if config, err := configStoreService.GetConfig("torrent.jackett.api_key"); err == nil {
		jackettAPIKey = strings.Trim(config.String(), "\"")
	}
	
	// 获取 qBittorrent 配置
	qbittorrentHost := "http://your-qbittorrent-server:8080"
	qbittorrentUser := "admin"
	qbittorrentPass := "adminadmin"
	if config, err := configStoreService.GetConfig("torrent.qbittorrent.host"); err == nil {
		qbittorrentHost = strings.Trim(config.String(), "\"")
	}
	if config, err := configStoreService.GetConfig("torrent.qbittorrent.username"); err == nil {
		qbittorrentUser = strings.Trim(config.String(), "\"")
	}
	if config, err := configStoreService.GetConfig("torrent.qbittorrent.password"); err == nil {
		qbittorrentPass = strings.Trim(config.String(), "\"")
	}
	
	torrentService := service.NewTorrentService(
		jackettHost,
		jackettAPIKey,
		qbittorrentHost,
		qbittorrentUser,
		qbittorrentPass,
		localMovieAdapter,
	)
	
	log.Printf("🔧 种子服务已创建 - Jackett: %s, qBittorrent: %s", jackettHost, qbittorrentHost)
	
	// 注入Telegram服务到种子下载服务
	if telegramService != nil {
		torrentService.SetTelegramService(telegramService)
	}

	// 创建日志服务（从配置中获取日志目录路径）
	logDir := "logs"
	if config, err := configStoreService.GetConfig("log.directory"); err == nil {
		logDir = strings.Trim(config.String(), "\"")
	}
	logService := service.NewLogService(logDir)
	log.Printf("📝 日志服务已创建，日志目录: %s", logDir)

	// 创建排行榜下载服务
	rankingDownloadService := service.NewRankingDownloadService(
		rankingDownloadTaskRepo,
		subscriptionRepo,
		rankingRepo,
		localMovieRepo,
		torrentService,
		telegramService,
		logService,
	)
	log.Printf("📥 排行榜下载服务已创建")

	// 记录系统启动相关日志
	logService.LogInfo("system", "routes", "开始初始化服务路由")
	logService.LogInfo("system", "database", "数据库连接已建立")
	logService.LogInfo("system", "config", "配置服务已初始化")

	// 创建扫描服务（从配置中获取媒体库路径）
	mediaLibraryPath := "/media/default"
	if config, err := configStoreService.GetConfig("media.base_path"); err == nil {
		mediaLibraryPath = strings.Trim(config.String(), "\"")
	}
	scannerService := service.NewScannerService(localMovieRepo, mediaLibraryPath, logService)

	// 创建排行榜服务（现在 logService 已经创建）
	rankingService := service.NewRankingService(crawlerConfig, rankingRepo, localMovieRepo, logService)

	// 创建JAVDb搜索服务（现在 logService 已经创建）
	javdbSearchService := service.NewJAVDbSearchService(crawlerConfig, logService)

	// 启动服务
	logService.LogInfo("scanner", "media-scan", "启动媒体库扫描服务，路径: "+mediaLibraryPath)
	scannerService.Start()

	logService.LogInfo("crawler", "ranking", "启动排行榜爬虫服务")
	rankingService.Start()

	// 创建处理器
	logService.LogInfo("system", "handlers", "初始化API处理器")
	localHandler := handlers.NewLocalHandler(localMovieRepo, scannerService, mediaLibraryPath)
	statsHandler := handlers.NewStatsHandler(localMovieRepo, rankingRepo)
	rankingHandler := handlers.NewRankingHandler(rankingService)
	rankingDownloadHandler := handlers.NewRankingDownloadHandler(rankingDownloadService)
	searchHandler := handlers.NewSearchHandler(localMovieRepo, rankingRepo)
	javdbSearchHandler := handlers.NewJAVDbSearchHandler(javdbSearchService)
	configHandler := handlers.NewConfigHandler(configService, telegramService)
	configStoreHandler := handlers.NewConfigStoreHandler()
	torrentHandler := handlers.NewTorrentHandler(torrentService)
	systemHandler := handlers.NewSystemHandler()
	logsHandler := handlers.NewLogsHandler(logService)

	// 记录各种服务状态
	if telegramService != nil {
		logService.LogInfo("system", "telegram", "Telegram通知服务已启用")
	} else {
		logService.LogWarn("system", "telegram", "Telegram通知服务未配置或已禁用")
	}
	logService.LogInfo("torrent", "jackett", "Jackett搜索服务配置: "+jackettHost)
	logService.LogInfo("torrent", "qbittorrent", "qBittorrent下载服务配置: "+qbittorrentHost)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "nsfw-go",
		})
	})

	// API 路由组
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{

			// 本地影片相关路由
			local := v1.Group("/local")
			{
				local.POST("/scan", localHandler.ScanLocalMovies)      // 手动扫描本地影片库
				local.GET("/movies", localHandler.GetLocalMovies)      // 获取本地影片列表
				local.GET("/search", localHandler.SearchLocalMovies)   // 搜索本地影片
				local.GET("/stats", localHandler.GetLocalMovieStats)   // 获取本地影片统计
				local.GET("/image/*filepath", localHandler.ServeImage) // 提供图片服务
			}

			// 排行榜相关路由
			rankings := v1.Group("/rankings")
			{
				rankings.GET("", rankingHandler.GetRankings)           // 获取排行榜
				rankings.GET("/stats", rankingHandler.GetRankingStats) // 获取排行榜统计
				rankings.GET("/local", rankingHandler.GetLocalExists)  // 获取本地已存在的排行榜影片
				rankings.POST("/crawl", rankingHandler.TriggerCrawl)   // 手动触发爬取
				rankings.POST("/check", rankingHandler.TriggerCheck)   // 手动触发本地检查

				// 下载任务相关
				rankings.POST("/download", rankingDownloadHandler.StartDownload)                        // 开始下载任务
				rankings.GET("/download-status/:code", rankingDownloadHandler.GetDownloadStatus)        // 获取下载状态
				rankings.GET("/download-tasks", rankingDownloadHandler.GetDownloadTasks)                // 获取任务列表
				rankings.DELETE("/download-tasks/:id", rankingDownloadHandler.CancelTask)               // 取消任务
				rankings.POST("/download-tasks/:id/retry", rankingDownloadHandler.RetryTask)            // 重试任务
				rankings.GET("/download-stats", rankingDownloadHandler.GetTaskStats)                    // 获取任务统计
				rankings.PUT("/download-tasks/:code/progress", rankingDownloadHandler.UpdateTaskProgress) // 更新任务进度

				// 订阅下载相关
				rankings.GET("/subscriptions", rankingDownloadHandler.GetSubscriptions)                       // 获取所有订阅配置
				rankings.GET("/subscription/:rank_type", rankingDownloadHandler.GetSubscriptionStatus)        // 获取订阅状态
				rankings.PUT("/subscription/:rank_type", rankingDownloadHandler.UpdateSubscription)           // 更新订阅配置
				rankings.POST("/subscription/:rank_type/run", rankingDownloadHandler.RunSubscriptionDownload) // 执行订阅下载
			}

			// 统计信息路由
			v1.GET("/stats", statsHandler.GetSystemStats)

			// 搜索功能
			search := v1.Group("/search")
			{
				search.GET("/", searchHandler.Search)
				search.GET("/suggestions", searchHandler.GetSuggestions)
				search.GET("/javdb", javdbSearchHandler.SearchJAVDb) // JAVDb搜索
			}

			// 配置管理
			config := v1.Group("/config")
			{
				config.GET("", configHandler.GetConfig)                            // 获取系统配置（从数据库）
				config.POST("", configHandler.SaveConfig)                          // 保存系统配置（到数据库）
				config.POST("/test", configHandler.TestConnection)                 // 测试连接
				config.POST("/test-notification", configHandler.TestNotification)  // 测试通知发送
				config.POST("/validate", configHandler.ValidateConfig)             // 验证配置
				config.GET("/categories", configHandler.GetConfigCategories)       // 获取配置分类
				config.GET("/category/:category", configHandler.GetConfigByCategory) // 按分类获取配置
				config.GET("/backups", configHandler.GetConfigBackups)             // 获取备份列表
				config.POST("/restore/:id", configHandler.RestoreConfigBackup)     // 恢复备份

				// 数据库配置管理
				store := config.Group("/store")
				{
					store.GET("", configStoreHandler.GetAllConfigs)                       // 获取所有配置
					store.GET("/category", configStoreHandler.GetConfigsByCategory)       // 按分类获取配置
					store.GET("/categories", configStoreHandler.GetConfigCategories)      // 获取配置分类
					store.GET("/:key", configStoreHandler.GetConfig)                      // 获取单个配置
					store.POST("", configStoreHandler.SetConfig)                          // 设置配置
					store.POST("/batch", configStoreHandler.BatchSetConfigs)              // 批量设置配置
					store.DELETE("/:key", configStoreHandler.DeleteConfig)                // 删除配置
					store.POST("/save-current", configStoreHandler.SaveCurrentConfigToDB) // 保存当前配置到数据库
					store.POST("/migrate", configStoreHandler.MigrateFileConfigToDB)      // 迁移配置文件到数据库

					// 配置备份管理
					backup := store.Group("/backup")
					{
						backup.POST("", configStoreHandler.CreateBackup)                  // 创建备份
						backup.GET("", configStoreHandler.GetBackups)                     // 获取备份列表
						backup.POST("/:id/restore", configStoreHandler.RestoreFromBackup) // 恢复备份
						backup.DELETE("/:id", configStoreHandler.DeleteBackup)            // 删除备份
					}
				}
			}

			// 种子下载功能
			log.Println("正在注册种子下载路由...")
			torrents := v1.Group("/torrents")
			{
				torrents.GET("/search", torrentHandler.SearchTorrents)             // 基础搜索（支持任意关键词）
				torrents.GET("/search/code", torrentHandler.SearchTorrentsForCode) // 按番号搜索（检查本地是否存在）
				torrents.GET("/best", torrentHandler.GetBestTorrentForCode)        // 获取番号最佳种子（最大文件）
				torrents.POST("/download", torrentHandler.DownloadTorrent)         // 下载种子
				torrents.POST("/download/best", torrentHandler.DownloadBestTorrentForCode) // 下载番号最佳种子
				torrents.GET("/list", torrentHandler.GetTorrentList)               // 获取下载列表
				torrents.GET("/status", torrentHandler.GetDownloadStatus)          // 获取下载状态统计
			}
			log.Println("种子下载路由注册完成。")

			// 系统管理功能
			system := v1.Group("/system")
			{
				system.POST("/restart", systemHandler.RestartServer) // 重启服务器
				system.GET("/info", systemHandler.GetSystemInfo)     // 获取系统信息
			}

			// 日志管理功能
			logs := v1.Group("/logs")
			{
				logs.GET("", logsHandler.GetLogs)           // 获取日志列表
				logs.DELETE("", logsHandler.ClearLogs)      // 清空日志
				logs.GET("/stats", logsHandler.GetLogStats) // 获取日志统计
				logs.POST("/test", logsHandler.CreateTestLogs) // 创建测试日志
			}
		}
	}

	// 记录路由初始化完成
	logService.LogInfo("system", "routes", "所有API路由注册完成")
	logService.LogInfo("system", "web", "静态文件服务配置完成")

	// 静态文件服务（前端）
	r.Static("/static", "./web/dist/static")
	r.StaticFile("/", "./web/dist/index.html")
	r.StaticFile("/local-movies.html", "./web/dist/local-movies.html")
	r.StaticFile("/rankings.html", "./web/dist/rankings.html")
	r.StaticFile("/search.html", "./web/dist/search.html")
	r.StaticFile("/config.html", "./web/dist/config.html")
	r.StaticFile("/downloads.html", "./web/dist/downloads.html")
	r.StaticFile("/logs.html", "./web/dist/logs.html")
	r.StaticFile("/javdb-search-test.html", "./web/dist/javdb-search-test.html")
	r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	// Swagger 文档路由（必须在 NoRoute 之前注册）
	url := ginSwagger.URL("doc.json") // 使用相对路径，与Swagger UI位于同一路径下
	swaggerGroup := r.Group("/swagger")
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	log.Printf("📚 Swagger 文档: http://localhost:8080/swagger/index.html")

	// 处理前端路由（SPA）- 只对非API路径生效
	r.NoRoute(func(c *gin.Context) {
		// 如果是API路径，返回404
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "NOT_FOUND",
				"message": "API路径不存在",
			})
			return
		}

		// 其他路径返回主页
		c.File("./web/dist/index.html")
	})
}

// clearDatabaseData 清空数据库中的模拟数据
func clearDatabaseData(db *gorm.DB) {
	// 清空所有模拟数据表
	db.Exec("DELETE FROM movies")
	db.Exec("DELETE FROM actresses")
	db.Exec("DELETE FROM studios")
	db.Exec("DELETE FROM series")
	db.Exec("DELETE FROM tags")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM watch_history")
	db.Exec("DELETE FROM favorites")
	db.Exec("DELETE FROM movie_actresses")
	db.Exec("DELETE FROM movie_tags")

	// 重置自增ID
	db.Exec("ALTER SEQUENCE movies_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE actresses_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE studios_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE series_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE tags_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")

	// 注意：不清空 local_movies 和 rankings 表，因为这是我们的缓存数据
}

