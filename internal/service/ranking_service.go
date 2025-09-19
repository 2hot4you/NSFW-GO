package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nsfw-go/internal/crawler"
	"nsfw-go/internal/model"
	"nsfw-go/internal/repo"
)

// RankingService 排行榜服务
type RankingService struct {
	rankingCrawler *crawler.RankingCrawler
	rankingRepo    repo.RankingRepository
	localMovieRepo repo.LocalMovieRepository
	logService     *LogService
	stopChan       chan struct{}
	crawlScheduled bool
	checkScheduled bool
}

// NewRankingService 创建排行榜服务
func NewRankingService(
	config *crawler.CrawlerConfig,
	rankingRepo repo.RankingRepository,
	localMovieRepo repo.LocalMovieRepository,
	logService *LogService,
) *RankingService {
	return &RankingService{
		rankingCrawler: crawler.NewRankingCrawler(config),
		rankingRepo:    rankingRepo,
		localMovieRepo: localMovieRepo,
		logService:     logService,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动服务
func (rs *RankingService) Start() {
	if rs.logService != nil {
		rs.logService.LogInfo("crawler", "ranking-service", "排行榜服务启动")
	}

	// 启动定时爬取任务（每天中午12:00）
	if !rs.crawlScheduled {
		go rs.startCrawlScheduler()
		rs.crawlScheduled = true
	}

	// 启动定时检查任务（每小时）
	if !rs.checkScheduled {
		go rs.startCheckScheduler()
		rs.checkScheduled = true
	}

	// 立即执行一次爬取（如果今天还没有爬取过）
	go func() {
		ctx := context.Background()
		if rs.shouldCrawlToday() {
			if rs.logService != nil {
				rs.logService.LogInfo("crawler", "ranking-service", "执行初始爬取")
			}
			rs.CrawlAndSaveRankings(ctx)
		}

		// 执行一次检查
		if rs.logService != nil {
			rs.logService.LogInfo("crawler", "ranking-service", "执行初始本地检查")
		}
		rs.CheckLocalExists(ctx, 100)
	}()
}

// Stop 停止服务
func (rs *RankingService) Stop() {
	if rs.logService != nil {
		rs.logService.LogInfo("crawler", "ranking-service", "排行榜服务停止")
	}
	close(rs.stopChan)
}

// startCrawlScheduler 启动爬取调度器
func (rs *RankingService) startCrawlScheduler() {
	ticker := time.NewTicker(time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			// 检查是否是中午12:00
			if now.Hour() == 12 && now.Minute() == 0 {
				if rs.logService != nil {
					rs.logService.LogInfo("crawler", "ranking-service", "定时爬取开始")
				}
				ctx := context.Background()
				rs.CrawlAndSaveRankings(ctx)
			}
		case <-rs.stopChan:
			return
		}
	}
}

// startCheckScheduler 启动检查调度器
func (rs *RankingService) startCheckScheduler() {
	ticker := time.NewTicker(time.Hour) // 每小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if rs.logService != nil {
				rs.logService.LogInfo("crawler", "ranking-service", "定时检查本地存在开始")
			}
			ctx := context.Background()
			rs.CheckLocalExists(ctx, 100)
		case <-rs.stopChan:
			return
		}
	}
}

// CrawlAndSaveRankings 爬取并保存排行榜
func (rs *RankingService) CrawlAndSaveRankings(ctx context.Context) error {
	if rs.logService != nil {
		rs.logService.LogInfo("crawler", "ranking-service", "开始爬取所有排行榜")
	}

	// 爬取所有排行榜
	rankings, err := rs.rankingCrawler.CrawlAllRankings(ctx)
	if err != nil {
		if rs.logService != nil {
			rs.logService.LogError("crawler", "ranking-service", fmt.Sprintf("爬取失败: %v", err))
		}
		return err
	}

	crawledAt := time.Now()
	totalSaved := 0

	// 保存每种类型的排行榜
	for rankType, items := range rankings {
		if len(items) == 0 {
			continue
		}

		// 清理该类型今天的旧数据（避免重复数据）
		todayStart := time.Date(crawledAt.Year(), crawledAt.Month(), crawledAt.Day(), 0, 0, 0, 0, crawledAt.Location())
		if err := rs.rankingRepo.ClearOldRankings(rankType, todayStart.Add(24*time.Hour)); err != nil {
			if rs.logService != nil {
				rs.logService.LogError("crawler", "ranking-service", fmt.Sprintf("清理 %s 今天的旧数据失败: %v", rankType, err))
			}
		}

		// 同时清理过期数据（保留最近7天的数据）
		keepTime := crawledAt.AddDate(0, 0, -7)
		if err := rs.rankingRepo.ClearOldRankings(rankType, keepTime); err != nil {
			if rs.logService != nil {
				rs.logService.LogError("crawler", "ranking-service", fmt.Sprintf("清理 %s 过期数据失败: %v", rankType, err))
			}
		}

		// 转换为数据库模型
		var rankingModels []model.Ranking
		for _, item := range items {
			ranking := model.Ranking{
				Code:      rs.normalizeCode(item.Code),
				Title:     item.Title,
				CoverURL:  item.CoverURL,
				RankType:  rankType,
				Position:  item.Position,
				CrawledAt: crawledAt,
			}
			rankingModels = append(rankingModels, ranking)
		}

		// 批量保存
		if err := rs.rankingRepo.BatchCreate(rankingModels); err != nil {
			if rs.logService != nil {
				rs.logService.LogError("crawler", "ranking-service", fmt.Sprintf("保存 %s 排行榜失败: %v", rankType, err))
			}
			continue
		}

		totalSaved += len(rankingModels)
		if rs.logService != nil {
			rs.logService.LogInfo("crawler", "ranking-service", fmt.Sprintf("保存 %s 排行榜成功，共 %d 条", rankType, len(rankingModels)))
		}
	}

	if rs.logService != nil {
		rs.logService.LogInfo("crawler", "ranking-service", fmt.Sprintf("爬取完成，共保存 %d 条记录", totalSaved))
	}
	return nil
}

// CheckLocalExists 检查本地存在状态
func (rs *RankingService) CheckLocalExists(ctx context.Context, batchSize int) error {
	if rs.logService != nil {
		rs.logService.LogInfo("crawler", "ranking-service", "开始检查本地存在状态")
	}

	// 获取需要检查的记录
	pendingRankings, err := rs.rankingRepo.GetPendingCheck(batchSize)
	if err != nil {
		if rs.logService != nil {
			rs.logService.LogError("crawler", "ranking-service", fmt.Sprintf("获取待检查记录失败: %v", err))
		}
		return err
	}

	if len(pendingRankings) == 0 {
		if rs.logService != nil {
			rs.logService.LogInfo("crawler", "ranking-service", "没有需要检查的记录")
		}
		return nil
	}

	checkedCount := 0
	existsCount := 0

	for _, ranking := range pendingRankings {
		// 检查本地是否存在
		exists := rs.checkLocalMovieExists(ranking.Code)

		// 更新状态
		if err := rs.rankingRepo.UpdateLocalExists(ranking.ID, exists); err != nil {
			if rs.logService != nil {
				rs.logService.LogError("crawler", "ranking-service", fmt.Sprintf("更新 %s 本地存在状态失败: %v", ranking.Code, err))
			}
			continue
		}

		checkedCount++
		if exists {
			existsCount++
		}

		// 添加小延时
		time.Sleep(10 * time.Millisecond)
	}

	if rs.logService != nil {
		rs.logService.LogInfo("crawler", "ranking-service", fmt.Sprintf("检查完成，共检查 %d 条，本地存在 %d 条", checkedCount, existsCount))
	}
	return nil
}

// checkLocalMovieExists 检查本地影片是否存在
func (rs *RankingService) checkLocalMovieExists(code string) bool {
	if code == "" {
		return false
	}

	// 标准化番号
	normalizedCode := rs.normalizeCode(code)

	// 从本地影视库检查
	localMovies, _, err := rs.localMovieRepo.List(0, 1000, "") // 获取所有本地影片，不按女优筛选
	if err != nil {
		if rs.logService != nil {
			rs.logService.LogError("crawler", "ranking-service", fmt.Sprintf("获取本地影片列表失败: %v", err))
		}
		return false
	}

	for _, localMovie := range localMovies {
		// 从文件路径中提取番号进行比较
		if rs.extractCodeFromPath(localMovie.Path) == normalizedCode {
			return true
		}

		// 也可以从标题中提取
		if rs.extractCodeFromFilename(localMovie.Title) == normalizedCode {
			return true
		}
	}

	return false
}

// normalizeCode 标准化番号
func (rs *RankingService) normalizeCode(code string) string {
	if code == "" {
		return ""
	}

	// 转换为大写
	code = strings.ToUpper(strings.TrimSpace(code))

	// 移除特殊字符，保留字母数字和连字符
	var result strings.Builder
	for _, r := range code {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// extractCodeFromPath 从路径中提取番号
func (rs *RankingService) extractCodeFromPath(path string) string {
	// 从路径中提取番号，例如: /path/to/[SSIS-001]Title/video.mp4
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			start := strings.Index(part, "[")
			end := strings.Index(part, "]")
			if start >= 0 && end > start {
				code := part[start+1 : end]
				return rs.normalizeCode(code)
			}
		}
	}

	// 如果没有找到中括号格式，尝试正则表达式
	baseCrawler := crawler.NewBaseCrawler("temp", &crawler.CrawlerConfig{})
	return baseCrawler.ExtractMovieCode(path)
}

// extractCodeFromFilename 从文件名中提取番号
func (rs *RankingService) extractCodeFromFilename(filename string) string {
	baseCrawler := crawler.NewBaseCrawler("temp", &crawler.CrawlerConfig{})
	return rs.normalizeCode(baseCrawler.ExtractMovieCode(filename))
}

// shouldCrawlToday 检查今天是否应该爬取
func (rs *RankingService) shouldCrawlToday() bool {
	// 检查每种类型的最新爬取时间
	rankTypes := []string{model.RankTypeDaily, model.RankTypeWeekly, model.RankTypeMonthly}

	today := time.Now().Truncate(24 * time.Hour)

	for _, rankType := range rankTypes {
		lastCrawl, err := rs.rankingRepo.GetLatestCrawlTime(rankType)
		if err != nil || lastCrawl == nil {
			// 如果没有爬取记录，应该爬取
			return true
		}

		lastCrawlDay := lastCrawl.Truncate(24 * time.Hour)
		if lastCrawlDay.Before(today) {
			// 如果最后一次爬取不是今天，应该爬取
			return true
		}
	}

	return false
}

// GetRankings 获取排行榜数据
func (rs *RankingService) GetRankings(rankType string, limit int) ([]*model.Ranking, error) {
	return rs.rankingRepo.GetByRankType(rankType, limit)
}

// GetRankingStats 获取排行榜统计信息
func (rs *RankingService) GetRankingStats() (map[string]interface{}, error) {
	total, err := rs.rankingRepo.Count()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_rankings": total,
		"rank_types":     make(map[string]interface{}),
	}

	rankTypes := []string{model.RankTypeDaily, model.RankTypeWeekly, model.RankTypeMonthly}
	for _, rankType := range rankTypes {
		rankings, err := rs.rankingRepo.GetByRankType(rankType, 50)
		if err != nil {
			continue
		}

		localExists := 0
		for _, ranking := range rankings {
			if ranking.LocalExists {
				localExists++
			}
		}

		lastCrawl, _ := rs.rankingRepo.GetLatestCrawlTime(rankType)

		stats["rank_types"].(map[string]interface{})[rankType] = map[string]interface{}{
			"total":        len(rankings),
			"local_exists": localExists,
			"last_crawl":   lastCrawl,
		}
	}

	return stats, nil
}

// TriggerManualCrawl 手动触发爬取
func (rs *RankingService) TriggerManualCrawl(ctx context.Context) error {
	if rs.logService != nil {
		rs.logService.LogInfo("crawler", "ranking-service", "手动触发爬取")
	}
	return rs.CrawlAndSaveRankings(ctx)
}

// TriggerManualCheck 手动触发本地检查
func (rs *RankingService) TriggerManualCheck(ctx context.Context) error {
	if rs.logService != nil {
		rs.logService.LogInfo("crawler", "ranking-service", "手动触发本地检查")
	}
	return rs.CheckLocalExists(ctx, 200)
}
