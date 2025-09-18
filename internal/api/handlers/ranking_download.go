package handlers

import (
	"net/http"
	"strconv"

	"nsfw-go/internal/model"
	"nsfw-go/internal/service"

	"github.com/gin-gonic/gin"
)

// RankingDownloadHandler 排行榜下载处理器
type RankingDownloadHandler struct {
	downloadService *service.RankingDownloadService
}

// NewRankingDownloadHandler 创建排行榜下载处理器
func NewRankingDownloadHandler(downloadService *service.RankingDownloadService) *RankingDownloadHandler {
	return &RankingDownloadHandler{
		downloadService: downloadService,
	}
}

// StartDownloadRequest 开始下载请求
type StartDownloadRequest struct {
	Code     string `json:"code" binding:"required"`
	Title    string `json:"title"`
	CoverURL string `json:"cover_url"`
	Source   string `json:"source"`
	RankType string `json:"rank_type"`
}

// StartDownload 开始下载任务
func (h *RankingDownloadHandler) StartDownload(c *gin.Context) {
	var req StartDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Source == "" {
		req.Source = model.DownloadSourceManual
	}
	if req.RankType == "" {
		req.RankType = model.RankTypeDaily
	}

	task, err := h.downloadService.StartDownloadTask(req.Code, req.Title, req.CoverURL, req.Source, req.RankType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "下载任务已启动",
		"data":    task,
	})
}

// GetDownloadStatus 获取下载状态
func (h *RankingDownloadHandler) GetDownloadStatus(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "番号不能为空",
		})
		return
	}

	task, err := h.downloadService.GetTaskByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到下载任务",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

// GetDownloadTasks 获取下载任务列表
func (h *RankingDownloadHandler) GetDownloadTasks(c *gin.Context) {
	status := c.Query("status")
	source := c.Query("source")
	rankType := c.Query("rank_type")
	
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的limit参数",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的offset参数",
		})
		return
	}

	tasks, total, err := h.downloadService.GetTasks(status, source, rankType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取任务列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks":  tasks,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// CancelTask 取消任务
func (h *RankingDownloadHandler) CancelTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的任务ID",
		})
		return
	}

	err = h.downloadService.CancelTask(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务已取消",
	})
}

// RetryTask 重试任务
func (h *RankingDownloadHandler) RetryTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的任务ID",
		})
		return
	}

	err = h.downloadService.RetryTask(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务已重新启动",
	})
}

// GetTaskStats 获取任务统计
func (h *RankingDownloadHandler) GetTaskStats(c *gin.Context) {
	stats, err := h.downloadService.GetTaskStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取统计信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSubscriptions 获取所有订阅配置
func (h *RankingDownloadHandler) GetSubscriptions(c *gin.Context) {
	subscriptions, err := h.downloadService.GetSubscriptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取订阅配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    subscriptions,
	})
}

// GetSubscriptionStatus 获取订阅状态
func (h *RankingDownloadHandler) GetSubscriptionStatus(c *gin.Context) {
	rankType := c.Param("rank_type")
	if rankType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "排行榜类型不能为空",
		})
		return
	}

	status, err := h.downloadService.GetSubscriptionStatus(rankType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// UpdateSubscriptionRequest 更新订阅请求
type UpdateSubscriptionRequest struct {
	Enabled     bool `json:"enabled"`
	HourlyLimit int  `json:"hourly_limit"`
	DailyLimit  int  `json:"daily_limit"`
}

// UpdateSubscription 更新订阅配置
func (h *RankingDownloadHandler) UpdateSubscription(c *gin.Context) {
	rankType := c.Param("rank_type")
	if rankType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "排行榜类型不能为空",
		})
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	err := h.downloadService.UpdateSubscription(rankType, req.Enabled, req.HourlyLimit, req.DailyLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订阅配置已更新",
	})
}

// RunSubscriptionDownload 执行订阅下载
func (h *RankingDownloadHandler) RunSubscriptionDownload(c *gin.Context) {
	rankType := c.Param("rank_type")
	if rankType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "排行榜类型不能为空",
		})
		return
	}

	// 验证排行榜类型
	if rankType != model.RankTypeDaily && rankType != model.RankTypeWeekly && rankType != model.RankTypeMonthly {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的排行榜类型，支持: daily, weekly, monthly",
		})
		return
	}

	err := h.downloadService.ExecuteSubscriptionDownload(rankType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订阅下载已启动",
	})
}

// UpdateTaskProgress 更新任务进度（用于外部调用，如定时任务）
func (h *RankingDownloadHandler) UpdateTaskProgress(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "番号不能为空",
		})
		return
	}

	progressStr := c.Query("progress")
	progress, err := strconv.ParseFloat(progressStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的进度值",
		})
		return
	}

	err = h.downloadService.UpdateTaskProgress(code, progress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "进度已更新",
	})
}