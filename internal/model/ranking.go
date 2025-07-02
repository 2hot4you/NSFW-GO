package model

import (
	"time"
)

// Ranking 排行榜模型
type Ranking struct {
	BaseModel
	Code        string     `gorm:"size:50;not null;index" json:"code"`
	Title       string     `gorm:"size:500;not null" json:"title"`
	CoverURL    string     `gorm:"size:1000" json:"cover_url"`
	RankType    string     `gorm:"size:20;not null;index" json:"rank_type"` // daily, weekly, monthly
	Position    int        `gorm:"not null;index" json:"position"`          // 排名位置
	CrawledAt   time.Time  `gorm:"not null;index" json:"crawled_at"`        // 爬取时间
	LocalExists bool       `gorm:"default:false;index" json:"local_exists"` // 是否在本地存在
	LastChecked *time.Time `json:"last_checked"`                            // 最后检查时间
}

// TableName 表名
func (Ranking) TableName() string {
	return "rankings"
}

// RankingType 排行榜类型常量
const (
	RankTypeDaily   = "daily"
	RankTypeWeekly  = "weekly"
	RankTypeMonthly = "monthly"
)
