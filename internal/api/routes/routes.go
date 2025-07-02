package routes

import (
	"net/http"
	"nsfw-go/internal/api/handlers"
	"nsfw-go/internal/crawler"
	"nsfw-go/internal/repo"
	"nsfw-go/internal/service"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// 清空数据库中的模拟数据
	clearDatabaseData(db)

	// 创建仓库
	localMovieRepo := repo.NewLocalMovieRepository(db)
	rankingRepo := repo.NewRankingRepository(db)

	// 创建扫描服务
	scannerService := service.NewScannerService(localMovieRepo)

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
	db.Exec("DELETE FROM watch_histories")
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
