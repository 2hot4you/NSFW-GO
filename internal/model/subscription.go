package model

import (
	"time"
)

// Subscription 订阅下载配置模型
type Subscription struct {
	BaseModel
	RankType     string    `gorm:"size:20;not null;uniqueIndex" json:"rank_type"`   // 排行榜类型: daily, weekly, monthly
	Enabled      bool      `gorm:"default:false" json:"enabled"`                    // 是否启用
	HourlyLimit  int       `gorm:"default:10" json:"hourly_limit"`                  // 每小时下载限制
	DailyLimit   int       `gorm:"default:50" json:"daily_limit"`                   // 每日下载限制  
	LastRunAt    *time.Time `json:"last_run_at"`                                    // 上次运行时间
	LastCheckAt  *time.Time `json:"last_check_at"`                                  // 上次检查时间
	TotalDownloads int     `gorm:"default:0" json:"total_downloads"`                // 总下载数量
	SuccessDownloads int   `gorm:"default:0" json:"success_downloads"`              // 成功下载数量
}

// TableName 表名
func (Subscription) TableName() string {
	return "subscriptions"
}

// SubscriptionLimit 订阅限制记录
type SubscriptionLimit struct {
	BaseModel
	RankType    string    `gorm:"size:20;not null;index" json:"rank_type"`         // 排行榜类型
	LimitType   string    `gorm:"size:20;not null;index" json:"limit_type"`        // 限制类型: hourly, daily
	Count       int       `gorm:"default:0" json:"count"`                          // 当前计数
	PeriodStart time.Time `gorm:"not null;index" json:"period_start"`              // 周期开始时间
	PeriodEnd   time.Time `gorm:"not null;index" json:"period_end"`                // 周期结束时间
}

// TableName 表名
func (SubscriptionLimit) TableName() string {
	return "subscription_limits"
}

// 限制类型常量
const (
	LimitTypeHourly = "hourly" // 每小时限制
	LimitTypeDaily  = "daily"  // 每日限制
)

// IsExpired 限制是否过期
func (sl *SubscriptionLimit) IsExpired() bool {
	return time.Now().After(sl.PeriodEnd)
}

// CanDownload 是否可以下载（未达到限制）
func (sl *SubscriptionLimit) CanDownload(limit int) bool {
	if sl.IsExpired() {
		return true
	}
	return sl.Count < limit
}