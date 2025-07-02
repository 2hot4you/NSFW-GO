package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 系统配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Bot      BotConfig      `mapstructure:"bot"`
	Crawler  CrawlerConfig  `mapstructure:"crawler"`
	Media    MediaConfig    `mapstructure:"media"`
	Security SecurityConfig `mapstructure:"security"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release, test
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	EnableCORS   bool          `mapstructure:"enable_cors"`
	EnableSwagger bool         `mapstructure:"enable_swagger"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"` // 秒
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// BotConfig Telegram Bot配置
type BotConfig struct {
	Token      string  `mapstructure:"token"`
	WebhookURL string  `mapstructure:"webhook_url"`
	AdminIDs   []int64 `mapstructure:"admin_ids"`
	Enabled    bool    `mapstructure:"enabled"`
}

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	UserAgents    []string      `mapstructure:"user_agents"`
	ProxyEnabled  bool          `mapstructure:"proxy_enabled"`
	ProxyList     []string      `mapstructure:"proxy_list"`
	RequestDelay  time.Duration `mapstructure:"request_delay"`
	RetryCount    int           `mapstructure:"retry_count"`
	Timeout       time.Duration `mapstructure:"timeout"`
	ConcurrentMax int           `mapstructure:"concurrent_max"`
}

// MediaConfig 媒体库配置
type MediaConfig struct {
	BasePath      string   `mapstructure:"base_path"`
	ScanInterval  int      `mapstructure:"scan_interval"` // 小时
	SupportedExts []string `mapstructure:"supported_exts"`
	MinFileSize   int64    `mapstructure:"min_file_size"` // MB
	MaxFileSize   int64    `mapstructure:"max_file_size"` // MB
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret     string        `mapstructure:"jwt_secret"`
	JWTExpiry     time.Duration `mapstructure:"jwt_expiry"`
	PasswordSalt  string        `mapstructure:"password_salt"`
	RateLimitRPS  int           `mapstructure:"rate_limit_rps"`
	AllowedIPs    []string      `mapstructure:"allowed_ips"`
	EnableAuth    bool          `mapstructure:"enable_auth"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, text
	Output     string `mapstructure:"output"` // stdout, file
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`     // 天
	Compress   bool   `mapstructure:"compress"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("/etc/nsfw-go")
	}

	// 设置环境变量前缀
	viper.SetEnvPrefix("NSFW")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.enable_cors", true)
	viper.SetDefault("server.enable_swagger", true)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "nsfw")
	viper.SetDefault("database.password", "nsfw123")
	viper.SetDefault("database.dbname", "nsfw_db")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_lifetime", 3600)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)

	// Bot defaults
	viper.SetDefault("bot.enabled", false)
	viper.SetDefault("bot.admin_ids", []int64{})

	// Crawler defaults
	viper.SetDefault("crawler.user_agents", []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
	})
	viper.SetDefault("crawler.proxy_enabled", false)
	viper.SetDefault("crawler.request_delay", "2s")
	viper.SetDefault("crawler.retry_count", 3)
	viper.SetDefault("crawler.timeout", "30s")
	viper.SetDefault("crawler.concurrent_max", 5)

	// Media defaults
	viper.SetDefault("media.base_path", "/MediaCenter")
	viper.SetDefault("media.scan_interval", 24)
	viper.SetDefault("media.supported_exts", []string{".mp4", ".mkv", ".avi", ".mov", ".wmv"})
	viper.SetDefault("media.min_file_size", 100) // 100MB
	viper.SetDefault("media.max_file_size", 10240) // 10GB

	// Security defaults
	viper.SetDefault("security.jwt_secret", "your-secret-key-change-it")
	viper.SetDefault("security.jwt_expiry", "24h")
	viper.SetDefault("security.password_salt", "nsfw-salt")
	viper.SetDefault("security.rate_limit_rps", 100)
	viper.SetDefault("security.enable_auth", false)

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.filename", "logs/nsfw-go.log")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 7)
	viper.SetDefault("log.max_age", 30)
	viper.SetDefault("log.compress", true)
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetAddr 获取服务器地址
func (c *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
} 