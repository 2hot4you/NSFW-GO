package routes

import (
	"log"
	"net/http"
	"nsfw-go/internal/api/handlers"
	"nsfw-go/internal/crawler"
	"nsfw-go/internal/model"
	"nsfw-go/internal/repo"
	"nsfw-go/internal/service"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL驱动
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

	// 创建排行榜服务
	rankingService := service.NewRankingService(crawlerConfig, rankingRepo, localMovieRepo)

	// 创建JAVDb搜索服务
	javdbSearchService := service.NewJAVDbSearchService(crawlerConfig)

	// 创建配置服务
	configService := service.NewConfigService("config.yaml")

	// 加载系统配置
	systemConfig, err := configService.GetConfig()
	if err != nil {
		// 如果加载配置失败，使用默认配置
		systemConfig = &model.SystemConfig{
			Torrent: model.TorrentConfig{
				Jackett: model.JackettConfig{
					Host:       "http://your-jackett-server:9117",
					APIKey:     "your_jackett_api_key",
					Timeout:    "30s",
					RetryCount: 3,
				},
				QBittorrent: model.QBittorrentConfig{
					Host:        "http://your-qbittorrent-server:8080",
					Username:    "admin",
					Password:    "adminadmin",
					Timeout:     "30s",
					RetryCount:  3,
					DownloadDir: "/downloads",
				},
			},
		}
	}

	// 创建种子下载服务
	localMovieAdapter := &LocalMovieAdapter{repo: localMovieRepo}
	torrentService := service.NewTorrentService(
		systemConfig.Torrent.Jackett.Host,
		systemConfig.Torrent.Jackett.APIKey,
		systemConfig.Torrent.QBittorrent.Host,
		systemConfig.Torrent.QBittorrent.Username,
		systemConfig.Torrent.QBittorrent.Password,
		localMovieAdapter,
	)

	// 创建扫描服务（从配置中获取媒体库路径）
	mediaLibraryPath := ""
	if systemConfig != nil {
		mediaLibraryPath = systemConfig.Media.BasePath
	}
	scannerService := service.NewScannerService(localMovieRepo, mediaLibraryPath)

	// 启动服务
	scannerService.Start()
	rankingService.Start()

	// 创建处理器
	localHandler := handlers.NewLocalHandler(localMovieRepo, scannerService)
	statsHandler := handlers.NewStatsHandler(localMovieRepo, rankingRepo)
	rankingHandler := handlers.NewRankingHandler(rankingService)
	searchHandler := handlers.NewSearchHandler(localMovieRepo, rankingRepo)
	javdbSearchHandler := handlers.NewJAVDbSearchHandler(javdbSearchService)
	configHandler := handlers.NewConfigHandler(configService)
	configStoreHandler := handlers.NewConfigStoreHandler()
	torrentHandler := handlers.NewTorrentHandler(torrentService)
	systemHandler := handlers.NewSystemHandler()

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
				config.GET("", configHandler.GetConfig)                            // 获取系统配置
				config.POST("", configHandler.SaveConfig)                          // 保存系统配置
				config.POST("/test", configHandler.TestConnection)                 // 测试连接
				config.POST("/validate", configHandler.ValidateConfig)             // 验证配置
				config.GET("/backups", configHandler.GetConfigBackups)             // 获取备份列表
				config.POST("/restore/:backup", configHandler.RestoreConfigBackup) // 恢复备份

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
				torrents.POST("/download", torrentHandler.DownloadTorrent)         // 下载种子
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
		}
	}

	// 静态文件服务（前端）
	r.Static("/static", "./web/dist/static")
	r.StaticFile("/", "./web/dist/index.html")
	r.StaticFile("/local-movies.html", "./web/dist/local-movies.html")
	r.StaticFile("/rankings.html", "./web/dist/rankings.html")
	r.StaticFile("/search.html", "./web/dist/search.html")
	r.StaticFile("/config.html", "./web/dist/config.html")
	r.StaticFile("/downloads.html", "./web/dist/downloads.html")
	r.StaticFile("/javdb-search-test.html", "./web/dist/javdb-search-test.html")
	r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

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
