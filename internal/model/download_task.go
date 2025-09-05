package model

import (
	"time"
)

// RankingDownloadTask 排行榜下载任务模型
type RankingDownloadTask struct {
	BaseModel
	Code         string    `gorm:"size:50;not null;uniqueIndex" json:"code"`        // 影片番号
	Title        string    `gorm:"size:500" json:"title"`                           // 影片标题  
	Status       string    `gorm:"size:20;not null;default:'pending'" json:"status"` // 下载状态
	TorrentURL   string    `gorm:"size:2000" json:"torrent_url"`                    // 种子链接
	TorrentHash  string    `gorm:"size:100" json:"torrent_hash"`                    // 种子哈希
	Progress     float64   `gorm:"default:0" json:"progress"`                       // 下载进度 (0-1)
	ErrorMsg     string    `gorm:"size:1000" json:"error_msg"`                      // 错误信息
	StartedAt    *time.Time `json:"started_at"`                                     // 开始时间
	CompletedAt  *time.Time `json:"completed_at"`                                   // 完成时间
	FileSize     int64     `gorm:"default:0" json:"file_size"`                      // 文件大小(字节)
	DownloadedSize int64   `gorm:"default:0" json:"downloaded_size"`                // 已下载大小
	Source       string    `gorm:"size:50;default:'manual'" json:"source"`          // 下载来源: manual, subscription
	RankType     string    `gorm:"size:20" json:"rank_type"`                        // 排行榜类型(用于订阅下载)
}

// TableName 表名
func (RankingDownloadTask) TableName() string {
	return "ranking_download_tasks"
}

// 排行榜下载状态常量
const (
	RankingDownloadStatusPending    = "pending"     // 等待中
	RankingDownloadStatusSearching  = "searching"   // 搜索种子中
	RankingDownloadStatusFound      = "found"       // 已找到种子
	RankingDownloadStatusStarted    = "started"     // 已开始下载
	RankingDownloadStatusProgress   = "progress"    // 下载中
	RankingDownloadStatusCompleted  = "completed"   // 下载完成
	RankingDownloadStatusFailed     = "failed"      // 下载失败
	RankingDownloadStatusCancelled  = "cancelled"   // 已取消
)

// 下载来源常量
const (
	DownloadSourceManual       = "manual"       // 手动下载
	DownloadSourceSubscription = "subscription" // 订阅下载
)

// IsCompleted 是否已完成
func (dt *RankingDownloadTask) IsCompleted() bool {
	return dt.Status == RankingDownloadStatusCompleted
}

// IsFailed 是否失败
func (dt *RankingDownloadTask) IsFailed() bool {
	return dt.Status == RankingDownloadStatusFailed
}

// IsActive 是否正在活跃（下载中）
func (dt *RankingDownloadTask) IsActive() bool {
	return dt.Status == RankingDownloadStatusPending ||
		   dt.Status == RankingDownloadStatusSearching ||
		   dt.Status == RankingDownloadStatusFound ||
		   dt.Status == RankingDownloadStatusStarted ||
		   dt.Status == RankingDownloadStatusProgress
}

// GetProgressPercent 获取进度百分比
func (dt *RankingDownloadTask) GetProgressPercent() int {
	if dt.Progress >= 1.0 {
		return 100
	}
	return int(dt.Progress * 100)
}