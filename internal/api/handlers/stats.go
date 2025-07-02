package handlers

import (
	"log"
	"net/http"
	"nsfw-go/internal/repo"

	"github.com/gin-gonic/gin"
)

// StatsHandler 统计信息处理器
type StatsHandler struct {
	localMovieRepo repo.LocalMovieRepository
	rankingRepo    repo.RankingRepository
}

// NewStatsHandler 创建统计信息处理器
func NewStatsHandler(localMovieRepo repo.LocalMovieRepository, rankingRepo repo.RankingRepository) *StatsHandler {
	return &StatsHandler{
		localMovieRepo: localMovieRepo,
		rankingRepo:    rankingRepo,
	}
}

// GetSystemStats 获取系统统计信息
func (h *StatsHandler) GetSystemStats(c *gin.Context) {
	log.Println("开始获取系统统计信息")

	// 获取本地影片数量
	localMovieCount, err := h.localMovieRepo.Count()
	if err != nil {
		log.Printf("获取本地影片数量失败: %v", err)
		localMovieCount = 0
	}
	log.Printf("本地影片数量: %d", localMovieCount)

	// 获取排行榜统计信息
	rankingStats, err := h.rankingRepo.GetStatsByType()
	if err != nil {
		log.Printf("获取排行榜统计失败: %v", err)
		c.Header("X-Debug-Ranking-Error", err.Error())
		rankingStats = make(map[string]map[string]int64)
	}
	log.Printf("排行榜统计数据: %+v", rankingStats)

	// 计算各类型统计
	dailyTotal := int64(0)
	dailyLocal := int64(0)
	weeklyTotal := int64(0)
	weeklyLocal := int64(0)
	monthlyTotal := int64(0)
	monthlyLocal := int64(0)

	if daily, exists := rankingStats["daily"]; exists {
		dailyTotal = daily["total"]
		dailyLocal = daily["local"]
	}
	if weekly, exists := rankingStats["weekly"]; exists {
		weeklyTotal = weekly["total"]
		weeklyLocal = weekly["local"]
	}
	if monthly, exists := rankingStats["monthly"]; exists {
		monthlyTotal = monthly["total"]
		monthlyLocal = monthly["local"]
	}

	// 获取本地影片的最后扫描时间
	lastScanTime, err := h.localMovieRepo.GetLastScanTime()
	lastScanTimeStr := "从未扫描"
	if err == nil && !lastScanTime.IsZero() {
		lastScanTimeStr = lastScanTime.Format("2006-01-02 15:04:05")
	}

	responseData := gin.H{
		"local_movies":   localMovieCount,
		"daily_total":    dailyTotal,
		"daily_local":    dailyLocal,
		"weekly_total":   weeklyTotal,
		"weekly_local":   weeklyLocal,
		"monthly_total":  monthlyTotal,
		"monthly_local":  monthlyLocal,
		"health_status":  "online",
		"api_status":     "healthy",
		"last_scan_time": lastScanTimeStr,
	}

	log.Printf("响应数据: %+v", responseData)

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "统计信息获取成功",
		Data:    responseData,
	})
}
