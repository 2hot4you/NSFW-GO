package model

import (
	"time"

	"gorm.io/gorm"
)

// LocalMovie 本地影片数据库模型
type LocalMovie struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Title       string         `gorm:"not null;index" json:"title"`
	Code        string         `gorm:"index" json:"code"`
	Actress     string         `gorm:"not null;index" json:"actress"`
	Path        string         `gorm:"not null;unique" json:"path"`
	Size        int64          `gorm:"not null" json:"size"`
	Modified    time.Time      `gorm:"not null" json:"modified"`
	Format      string         `gorm:"not null" json:"format"`
	FanartPath  string         `json:"fanart_path"`
	FanartURL   string         `json:"fanart_url"`
	HasFanart   bool           `gorm:"default:false" json:"has_fanart"`
	LastScanned time.Time      `gorm:"not null" json:"last_scanned"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (LocalMovie) TableName() string {
	return "local_movies"
}
