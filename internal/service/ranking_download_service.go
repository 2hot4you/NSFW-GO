package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"nsfw-go/internal/model"
	"nsfw-go/internal/repo"
	"gorm.io/gorm"
)

// RankingDownloadService 排行榜下载服务
type RankingDownloadService struct {
	taskRepo         repo.RankingDownloadTaskRepository
	subscriptionRepo repo.SubscriptionRepository
	rankingRepo      repo.RankingRepository
	localMovieRepo   repo.LocalMovieRepository
	torrentService   *TorrentService
	telegramService  *TelegramService
}

// NewRankingDownloadService 创建排行榜下载服务
func NewRankingDownloadService(
	taskRepo repo.RankingDownloadTaskRepository,
	subscriptionRepo repo.SubscriptionRepository,
	rankingRepo repo.RankingRepository,
	localMovieRepo repo.LocalMovieRepository,
	torrentService *TorrentService,
	telegramService *TelegramService,
) *RankingDownloadService {
	return &RankingDownloadService{
		taskRepo:         taskRepo,
		subscriptionRepo: subscriptionRepo,
		rankingRepo:      rankingRepo,
		localMovieRepo:   localMovieRepo,
		torrentService:   torrentService,
		telegramService:  telegramService,
	}
}

// StartDownloadTask 开始下载任务
func (s *RankingDownloadService) StartDownloadTask(code, title, source, rankType string) (*model.RankingDownloadTask, error) {
	// 检查是否已经在本地库中
	if localMovie, _ := s.localMovieRepo.SearchByCode(code); localMovie != nil {
		return nil, fmt.Errorf("影片 %s 已在本地库中", code)
	}
	
	// 检查是否有活跃的下载任务
	if existingTask, _ := s.taskRepo.GetActiveTaskByCode(code); existingTask != nil {
		return existingTask, nil
	}
	
	// 检查是否有历史任务（现在会正确排除软删除记录）
	if historyTask, _ := s.taskRepo.GetByCode(code); historyTask != nil {
		// 如果历史任务是失败/取消状态，允许重新创建
		if historyTask.Status == model.RankingDownloadStatusFailed || 
		   historyTask.Status == model.RankingDownloadStatusCancelled {
			// 软删除历史任务，为新任务让路
			if err := s.taskRepo.Delete(historyTask.ID); err != nil {
				return nil, fmt.Errorf("清理历史任务失败: %v", err)
			}
			log.Printf("[下载服务] 清理历史失败任务: %s (状态: %s)", code, historyTask.Status)
		} else {
			// 如果是已完成/进行中任务，返回现有任务或错误信息
			if historyTask.Status == model.RankingDownloadStatusCompleted {
				return nil, fmt.Errorf("番号 %s 已完成下载", code)
			}
			return historyTask, nil
		}
	}
	
	// 创建新的下载任务
	task := &model.RankingDownloadTask{
		Code:     code,
		Title:    title,
		Status:   model.RankingDownloadStatusPending,
		Source:   source,
		RankType: rankType,
	}
	
	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("创建下载任务失败: %v", err)
	}
	
	// 异步开始下载流程
	go s.executeDownload(task)
	
	return task, nil
}

// executeDownload 执行下载流程
func (s *RankingDownloadService) executeDownload(task *model.RankingDownloadTask) {
	// 更新状态为搜索中
	task.Status = model.RankingDownloadStatusSearching
	task.StartedAt = &[]time.Time{time.Now()}[0]
	s.taskRepo.Update(task)
	
	// 搜索种子
	log.Printf("[下载服务] 开始搜索种子: %s", task.Code)
	
	torrents, err := s.torrentService.SearchTorrentsForCode(task.Code)
	if err != nil || len(torrents) == 0 {
		s.markTaskFailed(task, "未找到可用种子")
		return
	}
	
	// 选择最优种子（第一个，已按优先级排序）
	bestTorrent := torrents[0]
	task.TorrentURL = bestTorrent.Link
	task.TorrentHash = bestTorrent.InfoHash
	task.FileSize = int64(bestTorrent.Size)
	task.Status = model.RankingDownloadStatusFound
	s.taskRepo.Update(task)
	
	log.Printf("[下载服务] 找到种子: %s (%s)", task.Code, bestTorrent.SizeFormatted)
	
	// 添加到 qBittorrent
	err = s.torrentService.DownloadTorrent(bestTorrent.Link)
	if err != nil {
		s.markTaskFailed(task, fmt.Sprintf("添加到下载器失败: %v", err))
		return
	}
	
	task.Status = model.RankingDownloadStatusStarted
	s.taskRepo.Update(task)
	
	log.Printf("[下载服务] 已添加到下载器: %s", task.Code)
	
	// 发送通知
	if s.telegramService != nil {
		message := fmt.Sprintf("🚀 开始下载: %s\n📁 %s\n💾 %s", 
			task.Code, task.Title, bestTorrent.SizeFormatted)
		s.telegramService.sendMessage(message)
	}
}

// markTaskFailed 标记任务失败
func (s *RankingDownloadService) markTaskFailed(task *model.RankingDownloadTask, errorMsg string) {
	task.Status = model.RankingDownloadStatusFailed
	task.ErrorMsg = errorMsg
	task.CompletedAt = &[]time.Time{time.Now()}[0]
	s.taskRepo.Update(task)
	
	log.Printf("[下载服务] 任务失败: %s - %s", task.Code, errorMsg)
	
	// 发送失败通知
	if s.telegramService != nil {
		message := fmt.Sprintf("❌ 下载失败: %s\n📁 %s\n🚫 %s", 
			task.Code, task.Title, errorMsg)
		s.telegramService.sendMessage(message)
	}
}

// GetTaskByCode 根据番号获取任务状态
func (s *RankingDownloadService) GetTaskByCode(code string) (*model.RankingDownloadTask, error) {
	return s.taskRepo.GetByCode(code)
}

// GetTasks 获取下载任务列表
func (s *RankingDownloadService) GetTasks(status, source, rankType string, limit, offset int) ([]*model.RankingDownloadTask, int64, error) {
	if status != "" || source != "" || rankType != "" {
		return s.taskRepo.GetTasksWithFilter(status, source, rankType, limit, offset)
	}
	return s.taskRepo.GetTasks(limit, offset)
}

// GetTaskStats 获取任务统计
func (s *RankingDownloadService) GetTaskStats() (*repo.TaskStats, error) {
	return s.taskRepo.GetTaskStats()
}

// CancelTask 取消任务
func (s *RankingDownloadService) CancelTask(id uint) error {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		return err
	}
	
	if !task.IsActive() {
		return fmt.Errorf("任务 %s 无法取消，当前状态: %s", task.Code, task.Status)
	}
	
	task.Status = model.RankingDownloadStatusCancelled
	task.CompletedAt = &[]time.Time{time.Now()}[0]
	return s.taskRepo.Update(task)
}

// RetryTask 重试失败的任务
func (s *RankingDownloadService) RetryTask(id uint) error {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		return err
	}
	
	if task.Status != model.RankingDownloadStatusFailed {
		return fmt.Errorf("只能重试失败的任务，当前状态: %s", task.Status)
	}
	
	// 重置任务状态
	task.Status = model.RankingDownloadStatusPending
	task.ErrorMsg = ""
	task.Progress = 0
	task.StartedAt = nil
	task.CompletedAt = nil
	
	if err := s.taskRepo.Update(task); err != nil {
		return err
	}
	
	// 异步重新执行
	go s.executeDownload(task)
	
	return nil
}

// UpdateTaskProgress 更新任务进度（由外部调用，如定时任务）
func (s *RankingDownloadService) UpdateTaskProgress(code string, progress float64) error {
	task, err := s.taskRepo.GetByCode(code)
	if err != nil {
		return err
	}
	
	task.Progress = progress
	if progress >= 1.0 {
		task.Status = model.RankingDownloadStatusCompleted
		task.CompletedAt = &[]time.Time{time.Now()}[0]
		
		// 发送完成通知
		if s.telegramService != nil {
			message := fmt.Sprintf("✅ 下载完成: %s\n📁 %s\n🎉 已保存到本地库", 
				task.Code, task.Title)
			s.telegramService.sendMessage(message)
		}
	} else if progress > 0 {
		task.Status = model.RankingDownloadStatusProgress
	}
	
	return s.taskRepo.Update(task)
}

// CleanupOldTasks 清理旧任务
func (s *RankingDownloadService) CleanupOldTasks(days int) error {
	return s.taskRepo.CleanupOldTasks(days)
}

// 订阅下载相关方法

// GetSubscriptions 获取所有订阅配置
func (s *RankingDownloadService) GetSubscriptions() ([]*model.Subscription, error) {
	return s.subscriptionRepo.GetAll()
}

// UpdateSubscription 更新订阅配置
func (s *RankingDownloadService) UpdateSubscription(rankType string, enabled bool, hourlyLimit, dailyLimit int) error {
	subscription, err := s.subscriptionRepo.GetByRankType(rankType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果订阅不存在，创建新的订阅配置
			subscription = &model.Subscription{
				RankType:    rankType,
				Enabled:     enabled,
				HourlyLimit: hourlyLimit,
				DailyLimit:  dailyLimit,
			}
			return s.subscriptionRepo.Create(subscription)
		}
		return err
	}
	
	subscription.Enabled = enabled
	subscription.HourlyLimit = hourlyLimit
	subscription.DailyLimit = dailyLimit
	
	return s.subscriptionRepo.Update(subscription)
}

// ExecuteSubscriptionDownload 执行订阅下载
func (s *RankingDownloadService) ExecuteSubscriptionDownload(rankType string) error {
	// 获取订阅配置
	subscription, err := s.subscriptionRepo.GetByRankType(rankType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("订阅 %s 不存在，请先配置订阅设置", rankType)
		}
		return fmt.Errorf("获取订阅配置失败: %v", err)
	}
	
	if !subscription.Enabled {
		return fmt.Errorf("订阅 %s 未启用", rankType)
	}
	
	// 检查限制
	canDownload, limitStatus, err := s.subscriptionRepo.CanDownload(rankType, subscription.HourlyLimit, subscription.DailyLimit)
	if err != nil {
		return fmt.Errorf("检查下载限制失败: %v", err)
	}
	
	if !canDownload {
		return fmt.Errorf("已达到下载限制 - 小时: %d/%d, 日: %d/%d", 
			limitStatus.HourlyUsed, limitStatus.HourlyLimit,
			limitStatus.DailyUsed, limitStatus.DailyLimit)
	}
	
	// 获取排行榜中未在本地的影片
	rankings, err := s.rankingRepo.GetByRankType(rankType, 50)
	if err != nil {
		return fmt.Errorf("获取排行榜失败: %v", err)
	}
	
	downloadCount := 0
	maxDownloads := subscription.HourlyLimit - limitStatus.HourlyUsed
	if dailyRemaining := subscription.DailyLimit - limitStatus.DailyUsed; dailyRemaining < maxDownloads {
		maxDownloads = dailyRemaining
	}
	
	for _, ranking := range rankings {
		if downloadCount >= maxDownloads {
			break
		}
		
		// 跳过已在本地的影片
		if ranking.LocalExists {
			continue
		}
		
		// 检查是否已有下载任务
		if existingTask, _ := s.taskRepo.GetActiveTaskByCode(ranking.Code); existingTask != nil {
			continue
		}
		
		// 开始下载任务
		_, err := s.StartDownloadTask(ranking.Code, ranking.Title, model.DownloadSourceSubscription, rankType)
		if err != nil {
			log.Printf("[订阅下载] 启动任务失败 %s: %v", ranking.Code, err)
			continue
		}
		
		// 增加限制计数
		s.subscriptionRepo.IncrementLimitCount(rankType, model.LimitTypeHourly)
		s.subscriptionRepo.IncrementLimitCount(rankType, model.LimitTypeDaily)
		
		downloadCount++
		time.Sleep(2 * time.Second) // 避免过于频繁的请求
	}
	
	// 更新订阅运行时间
	subscription.LastRunAt = &[]time.Time{time.Now()}[0]
	subscription.TotalDownloads += downloadCount
	s.subscriptionRepo.Update(subscription)
	
	log.Printf("[订阅下载] %s 执行完成，启动了 %d 个下载任务", rankType, downloadCount)
	
	// 发送通知
	if s.telegramService != nil && downloadCount > 0 {
		message := fmt.Sprintf("📋 订阅下载完成\n📊 %s 榜单\n🚀 启动了 %d 个下载任务", 
			rankType, downloadCount)
		s.telegramService.sendMessage(message)
	}
	
	return nil
}

// GetSubscriptionStatus 获取订阅状态
func (s *RankingDownloadService) GetSubscriptionStatus(rankType string) (*repo.LimitStatus, error) {
	subscription, err := s.subscriptionRepo.GetByRankType(rankType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果订阅不存在，创建默认订阅配置
			subscription = &model.Subscription{
				RankType:    rankType,
				Enabled:     false,
				HourlyLimit: 10,
				DailyLimit:  50,
			}
			if err := s.subscriptionRepo.Create(subscription); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	
	_, status, err := s.subscriptionRepo.CanDownload(rankType, subscription.HourlyLimit, subscription.DailyLimit)
	return status, err
}