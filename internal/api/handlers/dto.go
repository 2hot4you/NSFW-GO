package handlers

// Response 通用API响应结构
type Response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ListResponse 列表响应结构
type ListResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int64       `json:"total_pages"`
}

// CreateMovieRequest 创建影片请求
type CreateMovieRequest struct {
	Code        string  `json:"code" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	StudioID    *uint   `json:"studio_id"`
	SeriesID    *uint   `json:"series_id"`
	CoverURL    string  `json:"cover_url"`
	Rating      float32 `json:"rating"`
}

// UpdateMovieRequest 更新影片请求
type UpdateMovieRequest struct {
	Title       *string  `json:"title,omitempty" example:"更新的影片标题"`
	Description *string  `json:"description,omitempty" example:"更新的影片描述"`
	StudioID    *uint    `json:"studio_id,omitempty" example:"2"`
	SeriesID    *uint    `json:"series_id,omitempty" example:"2"`
	CoverURL    *string  `json:"cover_url,omitempty" example:"https://example.com/new_cover.jpg"`
	Rating      *float32 `json:"rating,omitempty" example:"9.0"`
}

// CreateActressRequest 创建女优请求
type CreateActressRequest struct {
	Name        string `json:"name" binding:"required" example:"女优名字"`
	Avatar      string `json:"avatar" example:"https://example.com/avatar.jpg"`
	Nationality string `json:"nationality" example:"日本"`
	Birthday    string `json:"birthday" example:"1995-01-01"`
	Height      int    `json:"height" example:"165"`
	Cup         string `json:"cup" example:"C"`
	Debut       string `json:"debut" example:"2020-01-01"`
}

// UpdateActressRequest 更新女优请求
type UpdateActressRequest struct {
	Name        *string `json:"name,omitempty" example:"更新的女优名字"`
	Avatar      *string `json:"avatar,omitempty" example:"https://example.com/new_avatar.jpg"`
	Nationality *string `json:"nationality,omitempty" example:"中国"`
	Birthday    *string `json:"birthday,omitempty" example:"1996-01-01"`
	Height      *int    `json:"height,omitempty" example:"170"`
	Cup         *string `json:"cup,omitempty" example:"D"`
	Debut       *string `json:"debut,omitempty" example:"2021-01-01"`
}

// CreateStudioRequest 创建制作商请求
type CreateStudioRequest struct {
	Name    string `json:"name" binding:"required" example:"制作商名称"`
	Website string `json:"website" example:"https://studio-website.com"`
	Logo    string `json:"logo" example:"https://example.com/logo.jpg"`
}

// CreateSeriesRequest 创建系列请求
type CreateSeriesRequest struct {
	Name        string `json:"name" binding:"required" example:"系列名称"`
	Description string `json:"description" example:"系列描述"`
	StudioID    uint   `json:"studio_id" binding:"required" example:"1"`
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name     string `json:"name" binding:"required" example:"标签名称"`
	Category string `json:"category" binding:"required" example:"类型"`
	Color    string `json:"color" example:"#FF5722"`
}

// CreateDownloadTaskRequest 创建下载任务请求
type CreateDownloadTaskRequest struct {
	MovieID     uint   `json:"movie_id" binding:"required" example:"1"`
	URL         string `json:"url" binding:"required" example:"https://download-url.com/movie.mp4"`
	Quality     string `json:"quality" example:"1080p"`
	Priority    int    `json:"priority" example:"1"`
	SavePath    string `json:"save_path" example:"/MediaCenter/NSFW/Hub/#Done/女优名字/[番号]标题/"`
	Subtitles   string `json:"subtitles" example:"字幕文件URL"`
	ExtractCode string `json:"extract_code" example:"解压密码"`
}

// StatsResponse 统计信息响应
type StatsResponse struct {
	TotalMovies     int64 `json:"total_movies"`
	TotalActresses  int64 `json:"total_actresses"`
	TotalStudios    int64 `json:"total_studios"`
	TotalTags       int64 `json:"total_tags"`
	DownloadedCount int64 `json:"downloaded_count"`
	PendingCount    int64 `json:"pending_count"`
	StorageUsed     int64 `json:"storage_used"`  // 字节
	StorageTotal    int64 `json:"storage_total"` // 字节
}

// CrawlerStatsResponse 爬虫统计响应
type CrawlerStatsResponse struct {
	TotalCrawled   int64  `json:"total_crawled"`
	LastCrawlTime  string `json:"last_crawl_time"`
	CrawlStatus    string `json:"crawl_status"`
	ErrorCount     int64  `json:"error_count"`
	SuccessRate    string `json:"success_rate"`
	NewMoviesCount int64  `json:"new_movies_count"`
	UpdatedCount   int64  `json:"updated_count"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Services map[string]string `json:"services"`
	Uptime   string            `json:"uptime"`
}
