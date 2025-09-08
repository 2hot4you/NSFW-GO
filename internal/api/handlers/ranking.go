package handlers

import (
	"net/http"
	"strconv"

	"nsfw-go/internal/model"
	"nsfw-go/internal/service"

	"github.com/gin-gonic/gin"
)

// RankingHandler 排行榜处理器
type RankingHandler struct {
	rankingService *service.RankingService
}

// NewRankingHandler 创建排行榜处理器
func NewRankingHandler(rankingService *service.RankingService) *RankingHandler {
	return &RankingHandler{
		rankingService: rankingService,
	}
}

// GetRankings 获取排行榜
// @Summary 获取排行榜列表
// @Description 获取JAVDb排行榜数据，支持日榜、周榜、月榜
// @Tags rankings
// @Accept json
// @Produce json
// @Param type query string false "排行榜类型" default(daily) Enums(daily, weekly, monthly)
// @Param limit query int false "数量限制" default(50)
// @Success 200 {object} Response{data=[]model.Ranking} "排行榜列表"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "获取失败"
// @Router /rankings [get]
func (h *RankingHandler) GetRankings(c *gin.Context) {
	// 兼容 type 和 period 参数
	rankType := c.DefaultQuery("type", "")
	if rankType == "" {
		rankType = c.DefaultQuery("period", model.RankTypeDaily)
	}
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的limit参数",
		})
		return
	}

	// 验证排行榜类型
	if rankType != model.RankTypeDaily && rankType != model.RankTypeWeekly && rankType != model.RankTypeMonthly {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的排行榜类型，支持: daily, weekly, monthly",
		})
		return
	}

	rankings, err := h.rankingService.GetRankings(rankType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取排行榜失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      rankings,
		"rank_type": rankType,
		"count":     len(rankings),
	})
}

// GetRankingStats 获取排行榜统计信息
func (h *RankingHandler) GetRankingStats(c *gin.Context) {
	stats, err := h.rankingService.GetRankingStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取排行榜统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// TriggerCrawl 手动触发爬取
func (h *RankingHandler) TriggerCrawl(c *gin.Context) {
	err := h.rankingService.TriggerManualCrawl(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "触发爬取失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "爬取任务已启动",
	})
}

// TriggerCheck 手动触发本地检查
func (h *RankingHandler) TriggerCheck(c *gin.Context) {
	err := h.rankingService.TriggerManualCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "触发本地检查失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "本地检查任务已启动",
	})
}

// GetLocalExists 获取本地已存在的排行榜影片
func (h *RankingHandler) GetLocalExists(c *gin.Context) {
	rankType := c.DefaultQuery("type", "all")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的limit参数",
		})
		return
	}

	var allRankings []*model.Ranking

	if rankType == "all" {
		// 获取所有类型的排行榜
		rankTypes := []string{model.RankTypeDaily, model.RankTypeWeekly, model.RankTypeMonthly}
		for _, rt := range rankTypes {
			rankings, err := h.rankingService.GetRankings(rt, limit/3)
			if err != nil {
				continue
			}
			allRankings = append(allRankings, rankings...)
		}
	} else {
		rankings, err := h.rankingService.GetRankings(rankType, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "获取排行榜失败: " + err.Error(),
			})
			return
		}
		allRankings = rankings
	}

	// 过滤出本地存在的影片
	var localExists []*model.Ranking
	for _, ranking := range allRankings {
		if ranking.LocalExists {
			localExists = append(localExists, ranking)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        localExists,
		"rank_type":   rankType,
		"total_count": len(allRankings),
		"local_count": len(localExists),
	})
}
