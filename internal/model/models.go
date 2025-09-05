package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// StringArray 字符串数组类型，用于PostgreSQL的TEXT[]类型
type StringArray []string

func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "{}", nil
	}
	data, err := json.Marshal(sa)
	return string(data), err
}

func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = StringArray{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, sa)
	case string:
		return json.Unmarshal([]byte(v), sa)
	default:
		return fmt.Errorf("无法扫描 %T 到 StringArray", value)
	}
}

// Int64Array 整数数组类型
type Int64Array []int64

func (ia Int64Array) Value() (driver.Value, error) {
	if len(ia) == 0 {
		return "{}", nil
	}
	data, err := json.Marshal(ia)
	return string(data), err
}

func (ia *Int64Array) Scan(value interface{}) error {
	if value == nil {
		*ia = Int64Array{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ia)
	case string:
		return json.Unmarshal([]byte(v), ia)
	default:
		return fmt.Errorf("无法扫描 %T 到 Int64Array", value)
	}
}

// BaseModel 基础模型，包含公共字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Actress 演员模型
type Actress struct {
	BaseModel
	Name        string      `gorm:"size:100;not null;index" json:"name"`
	Alias       StringArray `gorm:"type:text[]" json:"alias"`
	AvatarURL   string      `gorm:"size:500" json:"avatar_url"`
	Description string      `gorm:"type:text" json:"description"`

	// 关联关系
	Movies []Movie `gorm:"many2many:movie_actresses;" json:"movies,omitempty"`
}

// Studio 制作商模型
type Studio struct {
	BaseModel
	Name    string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	LogoURL string `gorm:"size:500" json:"logo_url"`

	// 关联关系
	Movies []Movie  `json:"movies,omitempty"`
	Series []Series `json:"series,omitempty"`
}

// Series 系列模型
type Series struct {
	BaseModel
	Name     string `gorm:"size:100;not null" json:"name"`
	StudioID uint   `gorm:"index" json:"studio_id"`

	// 关联关系
	Studio Studio  `json:"studio,omitempty"`
	Movies []Movie `json:"movies,omitempty"`
}

// Tag 标签模型
type Tag struct {
	BaseModel
	Name     string `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Category string `gorm:"size:20;index" json:"category"` // genre, quality, etc

	// 关联关系
	Movies []Movie `gorm:"many2many:movie_tags;" json:"movies,omitempty"`
}

// Movie 影片模型
type Movie struct {
	BaseModel
	Code              string      `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Title             string      `gorm:"size:500;not null" json:"title"`
	ReleaseDate       *time.Time  `json:"release_date"`
	Duration          int         `json:"duration"` // 分钟
	StudioID          *uint       `gorm:"index" json:"studio_id"`
	SeriesID          *uint       `gorm:"index" json:"series_id"`
	Description       string      `gorm:"type:text" json:"description"`
	Rating            float32     `gorm:"type:decimal(3,1)" json:"rating"`
	CoverURL          string      `gorm:"size:500" json:"cover_url"`
	FanartURL         string      `gorm:"size:500" json:"fanart_url"`
	TrailerURL        string      `gorm:"size:500" json:"trailer_url"`
	LocalPath         string      `gorm:"size:1000" json:"local_path"`
	FileSize          int64       `json:"file_size"`
	FileFormat        string      `gorm:"size:20" json:"file_format"`
	Quality           string      `gorm:"size:20" json:"quality"`
	HasSubtitle       bool        `gorm:"default:false" json:"has_subtitle"`
	SubtitleLanguages StringArray `gorm:"type:text[]" json:"subtitle_languages"`
	IsDownloaded      bool        `gorm:"default:false;index" json:"is_downloaded"`
	DownloadStatus    string      `gorm:"size:20;default:pending;index" json:"download_status"`
	DownloadProgress  int         `gorm:"default:0" json:"download_progress"`
	LastWatched       *time.Time  `json:"last_watched"`
	WatchCount        int         `gorm:"default:0" json:"watch_count"`

	// 关联关系
	Studio       *Studio        `json:"studio,omitempty"`
	Series       *Series        `json:"series,omitempty"`
	Actresses    []Actress      `gorm:"many2many:movie_actresses;" json:"actresses,omitempty"`
	Tags         []Tag          `gorm:"many2many:movie_tags;" json:"tags,omitempty"`
	Downloads    []DownloadTask `json:"downloads,omitempty"`
	Favorites    []Favorite     `json:"favorites,omitempty"`
	WatchHistory []WatchHistory `json:"watch_history,omitempty"`
}

// DownloadTask 下载任务模型
type DownloadTask struct {
	BaseModel
	MovieID        *uint      `gorm:"index" json:"movie_id"`
	URL            string     `gorm:"size:1000;not null" json:"url"`
	Status         string     `gorm:"size:20;default:pending;index" json:"status"`
	Progress       int        `gorm:"default:0" json:"progress"`
	Speed          int64      `gorm:"default:0" json:"speed"` // B/s
	TotalSize      int64      `gorm:"default:0" json:"total_size"`
	DownloadedSize int64      `gorm:"default:0" json:"downloaded_size"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`

	// 关联关系
	Movie *Movie `json:"movie,omitempty"`
}

// CrawlTask 爬虫任务模型
type CrawlTask struct {
	BaseModel
	URL          string                 `gorm:"size:1000;not null" json:"url"`
	Type         string                 `gorm:"size:20;not null;index" json:"type"`
	Status       string                 `gorm:"size:20;default:pending;index" json:"status"`
	Result       map[string]interface{} `gorm:"type:jsonb" json:"result"`
	ErrorMessage string                 `gorm:"type:text" json:"error_message"`
	CompletedAt  *time.Time             `json:"completed_at"`
}

// Favorite 用户收藏模型
type Favorite struct {
	BaseModel
	UserID  uint `gorm:"index" json:"user_id"` // 预留用户系统
	MovieID uint `gorm:"index" json:"movie_id"`

	// 关联关系
	Movie Movie `json:"movie,omitempty"`
}

// WatchHistory 观看历史模型
type WatchHistory struct {
	BaseModel
	UserID        uint      `gorm:"index" json:"user_id"` // 预留用户系统
	MovieID       uint      `gorm:"index" json:"movie_id"`
	WatchPosition int       `gorm:"default:0" json:"watch_position"` // 秒
	WatchedAt     time.Time `gorm:"index" json:"watched_at"`

	// 关联关系
	Movie Movie `json:"movie,omitempty"`
}

// User 用户模型（预留）
type User struct {
	BaseModel
	Username     string     `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Email        string     `gorm:"size:100;uniqueIndex" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	IsAdmin      bool       `gorm:"default:false" json:"is_admin"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`

	// 关联关系
	Favorites    []Favorite     `json:"favorites,omitempty"`
	WatchHistory []WatchHistory `json:"watch_history,omitempty"`
}

// MovieActor 影片-演员关联表（可以包含额外信息）
type MovieActress struct {
	MovieID   uint   `gorm:"primarykey"`
	ActressID uint   `gorm:"primarykey"`
	Role      string `gorm:"size:50"` // 角色类型：主演、配角等
}

// MovieTag 影片-标签关联表
type MovieTag struct {
	MovieID uint `gorm:"primarykey"`
	TagID   uint `gorm:"primarykey"`
}

// DownloadStatus 下载状态常量
const (
	DownloadStatusPending     = "pending"
	DownloadStatusDownloading = "downloading"
	DownloadStatusCompleted   = "completed"
	DownloadStatusFailed      = "failed"
	DownloadStatusPaused      = "paused"
	DownloadStatusCanceled    = "canceled"
)

// CrawlTaskType 爬虫任务类型常量
const (
	CrawlTaskTypeTrending    = "trending"
	CrawlTaskTypeSearch      = "search"
	CrawlTaskTypeMovieDetail = "movie_detail"
	CrawlTaskTypeActress     = "actress"
)

// CrawlTaskStatus 爬虫任务状态常量
const (
	CrawlTaskStatusPending   = "pending"
	CrawlTaskStatusRunning   = "running"
	CrawlTaskStatusCompleted = "completed"
	CrawlTaskStatusFailed    = "failed"
)

// TagCategory 标签分类常量
const (
	TagCategoryGenre   = "genre"
	TagCategoryQuality = "quality"
	TagCategoryStudio  = "studio"
	TagCategorySeries  = "series"
	TagCategoryOther   = "other"
)

// GetAllModels 返回所有需要迁移的模型
func GetAllModels() []interface{} {
	return []interface{}{
		&User{},
		&Actress{},
		&Studio{},
		&Series{},
		&Tag{},
		&Movie{},
		&DownloadTask{},
		&CrawlTask{},
		&Favorite{},
		&WatchHistory{},
		&Ranking{},
		&RankingDownloadTask{},
		&Subscription{},
		&SubscriptionLimit{},
	}
}

// TableName 方法用于自定义表名
func (Movie) TableName() string {
	return "movies"
}

func (Actress) TableName() string {
	return "actresses"
}

func (Studio) TableName() string {
	return "studios"
}

func (Series) TableName() string {
	return "series"
}

func (Tag) TableName() string {
	return "tags"
}

func (DownloadTask) TableName() string {
	return "download_tasks"
}

func (CrawlTask) TableName() string {
	return "crawl_tasks"
}

func (Favorite) TableName() string {
	return "favorites"
}

func (WatchHistory) TableName() string {
	return "watch_history"
}

func (User) TableName() string {
	return "users"
}
