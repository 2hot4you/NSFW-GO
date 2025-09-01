package api

import (
	"nsfw-go/internal/api/handlers"
	"nsfw-go/internal/service"

	"github.com/gin-gonic/gin"
)

// Router API路由
type Router struct {
	engine   *gin.Engine
	services *Services
}

// Services 服务集合
type Services struct {
	TorrentService *service.TorrentService
}

// NewRouter 创建路由
func NewRouter(engine *gin.Engine, services *Services) *Router {
	return &Router{
		engine:   engine,
		services: services,
	}
}

// RegisterRoutes 注册所有路由
func (r *Router) RegisterRoutes() {
	v1 := r.engine.Group("/api/v1")

	// 种子下载相关路由
	torrentHandler := handlers.NewTorrentHandler(r.services.TorrentService)
	
	// 基础搜索（支持任意关键词）
	v1.GET("/torrents/search", torrentHandler.SearchTorrents)
	
	// 按番号搜索（检查本地是否存在）
	v1.GET("/torrents/search/code", torrentHandler.SearchTorrentsForCode)
	
	// 下载种子
	v1.POST("/torrents/download", torrentHandler.DownloadTorrent)
	
	// 获取下载列表
	v1.GET("/torrents/list", torrentHandler.GetTorrentList)
	
	// 获取下载状态统计
	v1.GET("/torrents/status", torrentHandler.GetDownloadStatus)
}
