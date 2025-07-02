package crawler

import (
	"context"
	"time"
)

// CrawlResult 爬虫结果
type CrawlResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// MovieData 影片数据结构
type MovieData struct {
	Code              string        `json:"code"`
	Title             string        `json:"title"`
	ReleaseDate       time.Time     `json:"release_date"`
	Duration          int           `json:"duration"`
	Description       string        `json:"description"`
	Rating            float32       `json:"rating"`
	CoverURL          string        `json:"cover_url"`
	FanartURL         string        `json:"fanart_url"`
	TrailerURL        string        `json:"trailer_url"`
	Quality           string        `json:"quality"`
	HasSubtitle       bool          `json:"has_subtitle"`
	SubtitleLanguages []string      `json:"subtitle_languages"`
	Studio            *StudioData   `json:"studio,omitempty"`
	Series            *SeriesData   `json:"series,omitempty"`
	Actresses         []ActressData `json:"actresses,omitempty"`
	Tags              []TagData     `json:"tags,omitempty"`
}

// ActressData 女优数据结构
type ActressData struct {
	Name        string   `json:"name"`
	Alias       []string `json:"alias"`
	AvatarURL   string   `json:"avatar_url"`
	Description string   `json:"description"`
}

// StudioData 制作商数据结构
type StudioData struct {
	Name    string `json:"name"`
	LogoURL string `json:"logo_url"`
}

// SeriesData 系列数据结构
type SeriesData struct {
	Name   string `json:"name"`
	Studio string `json:"studio"`
}

// TagData 标签数据结构
type TagData struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	CoverURL    string    `json:"cover_url"`
	ReleaseDate time.Time `json:"release_date"`
	Rating      float32   `json:"rating"`
	DetailURL   string    `json:"detail_url"`
}

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	UserAgents    []string      `json:"user_agents"`
	ProxyEnabled  bool          `json:"proxy_enabled"`
	ProxyList     []string      `json:"proxy_list"`
	RequestDelay  time.Duration `json:"request_delay"`
	RetryCount    int           `json:"retry_count"`
	Timeout       time.Duration `json:"timeout"`
	ConcurrentMax int           `json:"concurrent_max"`
}

// Crawler 爬虫接口
type Crawler interface {
	// GetName 返回爬虫名称
	GetName() string

	// Search 搜索影片
	Search(ctx context.Context, keyword string) ([]SearchResult, error)

	// GetMovieByCode 根据番号获取影片详情
	GetMovieByCode(ctx context.Context, code string) (*MovieData, error)

	// GetMovieByURL 根据URL获取影片详情
	GetMovieByURL(ctx context.Context, url string) (*MovieData, error)

	// GetActressInfo 获取女优信息
	GetActressInfo(ctx context.Context, actressName string) (*ActressData, error)

	// IsHealthy 检查爬虫健康状态
	IsHealthy(ctx context.Context) bool
}

// CrawlerManager 爬虫管理器接口
type CrawlerManager interface {
	// RegisterCrawler 注册爬虫
	RegisterCrawler(name string, crawler Crawler)

	// GetCrawler 获取指定爬虫
	GetCrawler(name string) (Crawler, bool)

	// GetAllCrawlers 获取所有爬虫
	GetAllCrawlers() map[string]Crawler

	// CrawlMovieByCode 使用所有可用爬虫搜索影片
	CrawlMovieByCode(ctx context.Context, code string) (*MovieData, error)

	// SearchMovies 搜索影片
	SearchMovies(ctx context.Context, keyword string) ([]SearchResult, error)
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeMovieDetail TaskType = "movie_detail"
	TaskTypeMovieSearch TaskType = "movie_search"
	TaskTypeActressInfo TaskType = "actress_info"
	TaskTypeHealthCheck TaskType = "health_check"
)
