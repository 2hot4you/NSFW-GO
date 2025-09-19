package handlers

import (
	"net/http"
	"strconv"

	"nsfw-go/internal/service"

	"github.com/gin-gonic/gin"
)

// LogsHandler 日志处理器
type LogsHandler struct {
	logService *service.LogService
}

// NewLogsHandler 创建日志处理器
func NewLogsHandler(logService *service.LogService) *LogsHandler {
	return &LogsHandler{
		logService: logService,
	}
}

// GetLogs 获取日志列表
// @Summary 获取系统日志
// @Description 获取系统日志列表，支持分类和级别过滤
// @Tags logs
// @Accept json
// @Produce json
// @Param category query string false "日志分类" Enums(all,system,crawler,scanner,torrent,config)
// @Param level query string false "日志级别" Enums(all,info,warn,error,debug)
// @Param limit query int false "每页数量" default(50)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} APIResponse
// @Router /api/v1/logs [get]
func (h *LogsHandler) GetLogs(c *gin.Context) {
	// 获取查询参数
	category := c.DefaultQuery("category", "all")
	level := c.DefaultQuery("level", "all")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	// 解析分页参数
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000 // 限制最大返回数量
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 获取日志数据
	logs, err := h.logService.GetLogs(category, level, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取日志失败: " + err.Error(),
		})
		return
	}

	// 获取统计信息
	stats, _ := h.logService.GetLogStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取日志成功",
		"data":    logs,
		"meta": gin.H{
			"total":  len(logs),
			"limit":  limit,
			"offset": offset,
			"stats":  stats,
		},
	})
}

// ClearLogs 清空日志
// @Summary 清空系统日志
// @Description 清空指定分类的系统日志
// @Tags logs
// @Accept json
// @Produce json
// @Param category query string false "日志分类" Enums(all,system,crawler,scanner,torrent,config)
// @Success 200 {object} APIResponse
// @Router /api/v1/logs [delete]
func (h *LogsHandler) ClearLogs(c *gin.Context) {
	category := c.DefaultQuery("category", "all")

	err := h.logService.ClearLogs(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "清空日志失败: " + err.Error(),
		})
		return
	}

	// 记录清空操作日志
	h.logService.LogInfo("system", "logs-api", "日志已清空，分类: "+category)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "日志已清空",
	})
}

// GetLogStats 获取日志统计
// @Summary 获取日志统计信息
// @Description 获取各级别日志的数量统计
// @Tags logs
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Router /api/v1/logs/stats [get]
func (h *LogsHandler) GetLogStats(c *gin.Context) {
	stats, err := h.logService.GetLogStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取日志统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取统计成功",
		"data":    stats,
	})
}

// CreateTestLogs 创建测试日志
// @Summary 创建测试日志
// @Description 创建一些测试日志数据（开发调试用）
// @Tags logs
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Router /api/v1/logs/test [post]
func (h *LogsHandler) CreateTestLogs(c *gin.Context) {
	// 创建一些测试日志
	h.logService.LogInfo("system", "test", "这是一条测试信息日志")
	h.logService.LogWarn("system", "test", "这是一条测试警告日志")
	h.logService.LogError("system", "test", "这是一条测试错误日志")
	h.logService.LogDebug("system", "test", "这是一条测试调试日志")

	h.logService.LogInfo("crawler", "javdb", "开始爬取JAVDb排行榜")
	h.logService.LogInfo("crawler", "javdb", "爬取完成，共获取123条数据")
	h.logService.LogWarn("crawler", "javdb", "遇到反爬虫限制，等待重试")

	h.logService.LogInfo("scanner", "media-scan", "开始扫描本地媒体库")
	h.logService.LogInfo("scanner", "media-scan", "发现新文件：SONE-123.mp4")
	h.logService.LogInfo("scanner", "media-scan", "扫描完成，共处理456个文件")

	h.logService.LogInfo("torrent", "jackett", "种子搜索完成")
	h.logService.LogInfo("torrent", "qbittorrent", "添加下载任务成功")
	h.logService.LogError("torrent", "qbittorrent", "连接qBittorrent失败")

	h.logService.LogInfo("config", "config-service", "配置已更新")
	h.logService.LogInfo("config", "config-service", "数据库配置备份完成")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "测试日志已创建",
	})
}