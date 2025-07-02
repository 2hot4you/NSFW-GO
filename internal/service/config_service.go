package service

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"nsfw-go/internal/model"
)

// ConfigService 配置管理服务
type ConfigService struct {
	configPath string
}

// NewConfigService 创建配置服务
func NewConfigService(configPath string) *ConfigService {
	return &ConfigService{
		configPath: configPath,
	}
}

// GetConfig 获取系统配置
func (s *ConfigService) GetConfig() (*model.SystemConfig, error) {
	var config model.SystemConfig

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// SaveConfig 保存系统配置
func (s *ConfigService) SaveConfig(config *model.SystemConfig) error {
	// 备份原配置文件
	backupPath := s.configPath + ".backup." + time.Now().Format("20060102150405")
	if err := s.copyFile(s.configPath, backupPath); err != nil {
		return fmt.Errorf("备份配置文件失败: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 添加配置文件头部注释
	header := `# NSFW-Go 配置文件
# 这是一个自动生成的配置文件，包含了所有可配置的选项
# 修改时间: ` + time.Now().Format("2006-01-02 15:04:05") + `

`
	fullData := []byte(header + string(data))

	if err := os.WriteFile(s.configPath, fullData, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// TestDatabaseConnection 测试数据库连接
func (s *ConfigService) TestDatabaseConnection(config model.DatabaseConfig) *model.ConnectionTestResult {
	start := time.Now()

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("连接失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("ping失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &model.ConnectionTestResult{
		Success: true,
		Message: "数据库连接成功",
		Latency: time.Since(start).Milliseconds(),
	}
}

// TestRedisConnection 测试Redis连接
func (s *ConfigService) TestRedisConnection(config model.RedisConfig) *model.ConnectionTestResult {
	start := time.Now()

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("连接失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer conn.Close()

	return &model.ConnectionTestResult{
		Success: true,
		Message: "Redis端口连接成功（需要完整Redis库进行完整测试）",
		Latency: time.Since(start).Milliseconds(),
	}
}

// TestTelegramConnection 测试Telegram Bot连接
func (s *ConfigService) TestTelegramConnection(config model.BotConfig) *model.ConnectionTestResult {
	start := time.Now()

	if config.Token == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Bot Token为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 简单验证Token格式
	if len(config.Token) < 10 {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Bot Token格式无效",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &model.ConnectionTestResult{
		Success: true,
		Message: "Telegram Bot Token格式有效（需要完整Telegram库进行完整测试）",
		Latency: time.Since(start).Milliseconds(),
	}
}

// TestEmailConnection 测试邮件连接
func (s *ConfigService) TestEmailConnection(config model.EmailNotificationConfig) *model.ConnectionTestResult {
	start := time.Now()

	if config.SMTPHost == "" || config.Username == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "SMTP配置不完整",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	client, err := smtp.Dial(addr)
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("连接SMTP服务器失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("SMTP认证失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &model.ConnectionTestResult{
		Success: true,
		Message: "邮件服务器连接成功",
		Latency: time.Since(start).Milliseconds(),
	}
}

// TestTelegramNotification 测试Telegram通知
func (s *ConfigService) TestTelegramNotification(telegramConfig model.TelegramNotificationConfig, botToken string) *model.ConnectionTestResult {
	start := time.Now()

	if botToken == "" || telegramConfig.ChatID == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Bot Token或Chat ID为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 简单验证Token和ChatID格式
	if len(botToken) < 10 {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Bot Token格式无效",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	if telegramConfig.ChatID == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Chat ID不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &model.ConnectionTestResult{
		Success: true,
		Message: "Telegram通知配置格式有效（需要完整Telegram库进行完整测试）",
		Latency: time.Since(start).Milliseconds(),
	}
}

// ValidateConfig 验证配置的有效性
func (s *ConfigService) ValidateConfig(config *model.SystemConfig) []string {
	var errors []string

	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		errors = append(errors, "无效的服务器端口")
	}

	if config.Server.Mode != "debug" && config.Server.Mode != "release" && config.Server.Mode != "test" {
		errors = append(errors, "无效的服务器模式")
	}

	// 验证数据库配置
	if config.Database.Host == "" {
		errors = append(errors, "数据库主机不能为空")
	}

	if config.Database.User == "" {
		errors = append(errors, "数据库用户名不能为空")
	}

	if config.Database.DBName == "" {
		errors = append(errors, "数据库名称不能为空")
	}

	// 验证Redis配置
	if config.Redis.Host == "" {
		errors = append(errors, "Redis主机不能为空")
	}

	// 验证JWT密钥
	if config.Security.JWTSecret == "" || config.Security.JWTSecret == "your-secret-key-change-it" {
		errors = append(errors, "请设置安全的JWT密钥")
	}

	// 验证媒体库路径
	if config.Media.BasePath != "" {
		if _, err := os.Stat(config.Media.BasePath); os.IsNotExist(err) {
			errors = append(errors, "媒体库路径不存在")
		}
	}

	return errors
}

// copyFile 复制文件
func (s *ConfigService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// GetConfigBackups 获取配置备份列表
func (s *ConfigService) GetConfigBackups() ([]string, error) {
	dir := filepath.Dir(s.configPath)
	filename := filepath.Base(s.configPath)

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var backups []string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > len(filename)+8 &&
			file.Name()[:len(filename)+8] == filename+".backup." {
			backups = append(backups, file.Name())
		}
	}

	return backups, nil
}

// RestoreConfigBackup 恢复配置备份
func (s *ConfigService) RestoreConfigBackup(backupName string) error {
	dir := filepath.Dir(s.configPath)
	backupPath := filepath.Join(dir, backupName)

	return s.copyFile(backupPath, s.configPath)
}
