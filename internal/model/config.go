package model

// SystemConfig 系统配置结构
type SystemConfig struct {
	Server        ServerConfig        `yaml:"server" json:"server"`
	Database      DatabaseConfig      `yaml:"database" json:"database"`
	Redis         RedisConfig         `yaml:"redis" json:"redis"`
	Bot           BotConfig           `yaml:"bot" json:"bot"`
	Crawler       CrawlerConfig       `yaml:"crawler" json:"crawler"`
	Media         MediaConfig         `yaml:"media" json:"media"`
	Security      SecurityConfig      `yaml:"security" json:"security"`
	Log           LogConfig           `yaml:"log" json:"log"`
	Dev           DevConfig           `yaml:"dev" json:"dev"`
	Sites         SitesConfig         `yaml:"sites" json:"sites"`
	Download      DownloadConfig      `yaml:"download" json:"download"`
	Notifications NotificationsConfig `yaml:"notifications" json:"notifications"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host          string `yaml:"host" json:"host"`
	Port          int    `yaml:"port" json:"port"`
	Mode          string `yaml:"mode" json:"mode"`
	ReadTimeout   string `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout  string `yaml:"write_timeout" json:"write_timeout"`
	EnableCors    bool   `yaml:"enable_cors" json:"enable_cors"`
	EnableSwagger bool   `yaml:"enable_swagger" json:"enable_swagger"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `yaml:"host" json:"host"`
	Port         int    `yaml:"port" json:"port"`
	User         string `yaml:"user" json:"user"`
	Password     string `yaml:"password" json:"password"`
	DBName       string `yaml:"dbname" json:"dbname"`
	SSLMode      string `yaml:"sslmode" json:"sslmode"`
	MaxOpenConns int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxLifetime  int    `yaml:"max_lifetime" json:"max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `yaml:"host" json:"host"`
	Port         int    `yaml:"port" json:"port"`
	Password     string `yaml:"password" json:"password"`
	DB           int    `yaml:"db" json:"db"`
	PoolSize     int    `yaml:"pool_size" json:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns" json:"min_idle_conns"`
}

// BotConfig Telegram Bot配置
type BotConfig struct {
	Enabled    bool    `yaml:"enabled" json:"enabled"`
	Token      string  `yaml:"token" json:"token"`
	WebhookURL string  `yaml:"webhook_url" json:"webhook_url"`
	AdminIDs   []int64 `yaml:"admin_ids" json:"admin_ids"`
}

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	UserAgents    []string `yaml:"user_agents" json:"user_agents"`
	ProxyEnabled  bool     `yaml:"proxy_enabled" json:"proxy_enabled"`
	ProxyList     []string `yaml:"proxy_list" json:"proxy_list"`
	RequestDelay  string   `yaml:"request_delay" json:"request_delay"`
	RetryCount    int      `yaml:"retry_count" json:"retry_count"`
	Timeout       string   `yaml:"timeout" json:"timeout"`
	ConcurrentMax int      `yaml:"concurrent_max" json:"concurrent_max"`
}

// MediaConfig 媒体库配置
type MediaConfig struct {
	BasePath      string   `yaml:"base_path" json:"base_path"`
	ScanInterval  int      `yaml:"scan_interval" json:"scan_interval"`
	SupportedExts []string `yaml:"supported_exts" json:"supported_exts"`
	MinFileSize   int      `yaml:"min_file_size" json:"min_file_size"`
	MaxFileSize   int      `yaml:"max_file_size" json:"max_file_size"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret    string   `yaml:"jwt_secret" json:"jwt_secret"`
	JWTExpiry    string   `yaml:"jwt_expiry" json:"jwt_expiry"`
	PasswordSalt string   `yaml:"password_salt" json:"password_salt"`
	RateLimitRPS int      `yaml:"rate_limit_rps" json:"rate_limit_rps"`
	AllowedIPs   []string `yaml:"allowed_ips" json:"allowed_ips"`
	EnableAuth   bool     `yaml:"enable_auth" json:"enable_auth"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	Output     string `yaml:"output" json:"output"`
	Filename   string `yaml:"filename" json:"filename"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// DevConfig 开发环境配置
type DevConfig struct {
	EnableDebugRoutes bool `yaml:"enable_debug_routes" json:"enable_debug_routes"`
	EnableProfiling   bool `yaml:"enable_profiling" json:"enable_profiling"`
	AutoReload        bool `yaml:"auto_reload" json:"auto_reload"`
}

// SitesConfig 站点特定配置
type SitesConfig struct {
	JAVDb      SiteConfig `yaml:"javdb" json:"javdb"`
	JAVLibrary SiteConfig `yaml:"javlibrary" json:"javlibrary"`
	JAVBus     SiteConfig `yaml:"javbus" json:"javbus"`
}

// SiteConfig 单个站点配置
type SiteConfig struct {
	BaseURL        string `yaml:"base_url" json:"base_url"`
	SearchPath     string `yaml:"search_path" json:"search_path"`
	DetailSelector string `yaml:"detail_selector,omitempty" json:"detail_selector,omitempty"`
	RateLimit      string `yaml:"rate_limit" json:"rate_limit"`
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	MaxConcurrent int    `yaml:"max_concurrent" json:"max_concurrent"`
	RetryCount    int    `yaml:"retry_count" json:"retry_count"`
	RetryDelay    string `yaml:"retry_delay" json:"retry_delay"`
	SpeedLimit    int    `yaml:"speed_limit" json:"speed_limit"`
	TempDir       string `yaml:"temp_dir" json:"temp_dir"`
	CompletedDir  string `yaml:"completed_dir" json:"completed_dir"`
}

// NotificationsConfig 通知配置
type NotificationsConfig struct {
	Telegram TelegramNotificationConfig `yaml:"telegram" json:"telegram"`
	Email    EmailNotificationConfig    `yaml:"email" json:"email"`
}

// TelegramNotificationConfig Telegram通知配置
type TelegramNotificationConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	ChatID  string `yaml:"chat_id" json:"chat_id"`
}

// EmailNotificationConfig 邮件通知配置
type EmailNotificationConfig struct {
	Enabled  bool     `yaml:"enabled" json:"enabled"`
	SMTPHost string   `yaml:"smtp_host" json:"smtp_host"`
	SMTPPort int      `yaml:"smtp_port" json:"smtp_port"`
	Username string   `yaml:"username" json:"username"`
	Password string   `yaml:"password" json:"password"`
	From     string   `yaml:"from" json:"from"`
	To       []string `yaml:"to" json:"to"`
}

// ConnectionTestResult 连接测试结果
type ConnectionTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Latency int64  `json:"latency"` // 毫秒
}

// ConfigTestRequest 配置测试请求
type ConfigTestRequest struct {
	Type string      `json:"type"` // database, redis, telegram, email
	Data interface{} `json:"data"`
}
