package service

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL驱动
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

	// 验证必要参数
	if config.Host == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "数据库主机地址不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	if config.User == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "数据库用户名不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	if config.DBName == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "数据库名称不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("创建数据库连接失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer db.Close()

	// 设置较短的超时时间进行真实连接测试
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		// 解析具体的错误信息
		errMsg := err.Error()
		if strings.Contains(errMsg, "connection refused") {
			errMsg = "连接被拒绝，请检查主机地址和端口"
		} else if strings.Contains(errMsg, "password authentication failed") {
			errMsg = "密码认证失败，请检查用户名和密码"
		} else if strings.Contains(errMsg, "database") && strings.Contains(errMsg, "does not exist") {
			errMsg = "数据库不存在，请检查数据库名称"
		} else if strings.Contains(errMsg, "timeout") {
			errMsg = "连接超时，请检查网络和防火墙设置"
		}

		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("数据库连接失败: %s", errMsg),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 执行一个简单的查询来确保连接正常工作
	var result int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("数据库查询测试失败: %s", err.Error()),
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

	// 验证必要参数
	if config.Host == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Redis主机地址不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// 首先测试TCP连接
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "connection refused") {
			errMsg = "连接被拒绝，请检查Redis服务是否启动以及主机地址和端口"
		} else if strings.Contains(errMsg, "timeout") {
			errMsg = "连接超时，请检查网络和防火墙设置"
		} else if strings.Contains(errMsg, "no route to host") {
			errMsg = "无法到达主机，请检查主机地址"
		}

		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Redis TCP连接失败: %s", errMsg),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer conn.Close()

	// 设置读写超时
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// 发送Redis PING命令
	if config.Password != "" {
		// 如果有密码，先发送AUTH命令
		authCmd := fmt.Sprintf("AUTH %s\r\n", config.Password)
		_, err = conn.Write([]byte(authCmd))
		if err != nil {
			return &model.ConnectionTestResult{
				Success: false,
				Message: fmt.Sprintf("发送AUTH命令失败: %s", err.Error()),
				Latency: time.Since(start).Milliseconds(),
			}
		}

		// 读取AUTH响应
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			return &model.ConnectionTestResult{
				Success: false,
				Message: fmt.Sprintf("读取AUTH响应失败: %s", err.Error()),
				Latency: time.Since(start).Milliseconds(),
			}
		}

		response := string(buffer[:n])
		if strings.Contains(response, "-ERR") {
			if strings.Contains(response, "invalid password") || strings.Contains(response, "WRONGPASS") {
				return &model.ConnectionTestResult{
					Success: false,
					Message: "Redis密码认证失败，请检查密码",
					Latency: time.Since(start).Milliseconds(),
				}
			}
			return &model.ConnectionTestResult{
				Success: false,
				Message: fmt.Sprintf("Redis认证失败: %s", strings.TrimSpace(response)),
				Latency: time.Since(start).Milliseconds(),
			}
		}
	}

	// 发送PING命令
	_, err = conn.Write([]byte("PING\r\n"))
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("发送PING命令失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 读取PING响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("读取PING响应失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	response := string(buffer[:n])
	if !strings.Contains(response, "+PONG") {
		if strings.Contains(response, "-NOAUTH") {
			return &model.ConnectionTestResult{
				Success: false,
				Message: "Redis需要密码认证，请设置正确的密码",
				Latency: time.Since(start).Milliseconds(),
			}
		}

		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Redis PING测试失败，响应: %s", strings.TrimSpace(response)),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 如果指定了数据库编号，测试SELECT命令
	if config.DB > 0 {
		selectCmd := fmt.Sprintf("SELECT %d\r\n", config.DB)
		_, err = conn.Write([]byte(selectCmd))
		if err != nil {
			return &model.ConnectionTestResult{
				Success: false,
				Message: fmt.Sprintf("发送SELECT命令失败: %s", err.Error()),
				Latency: time.Since(start).Milliseconds(),
			}
		}

		n, err = conn.Read(buffer)
		if err != nil {
			return &model.ConnectionTestResult{
				Success: false,
				Message: fmt.Sprintf("读取SELECT响应失败: %s", err.Error()),
				Latency: time.Since(start).Milliseconds(),
			}
		}

		response = string(buffer[:n])
		if strings.Contains(response, "-ERR") {
			return &model.ConnectionTestResult{
				Success: false,
				Message: fmt.Sprintf("Redis数据库选择失败: %s", strings.TrimSpace(response)),
				Latency: time.Since(start).Milliseconds(),
			}
		}
	}

	return &model.ConnectionTestResult{
		Success: true,
		Message: "Redis连接成功",
		Latency: time.Since(start).Milliseconds(),
	}
}

// TestTelegramConnection 测试Telegram Bot连接
func (s *ConfigService) TestTelegramConnection(config model.BotConfig) *model.ConnectionTestResult {
	start := time.Now()

	if config.Token == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Bot Token不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 验证Token格式：应该是 "数字:字符串" 的格式
	parts := strings.Split(config.Token, ":")
	if len(parts) != 2 || len(parts[0]) < 8 || len(parts[1]) < 35 {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Bot Token格式无效，应为 'botId:authToken' 格式",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 实际调用Telegram API测试Token有效性
	client := &http.Client{Timeout: 10 * time.Second}
	testURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", config.Token)

	resp, err := client.Get(testURL)
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("无法连接到Telegram API: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Bot Token无效或已过期",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	if resp.StatusCode != 200 {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Telegram API返回错误: HTTP %d", resp.StatusCode),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &model.ConnectionTestResult{
		Success: true,
		Message: "Telegram Bot连接成功",
		Latency: time.Since(start).Milliseconds(),
	}
}

// TestEmailConnection 测试邮件连接
func (s *ConfigService) TestEmailConnection(config model.EmailNotificationConfig) *model.ConnectionTestResult {
	start := time.Now()

	if config.SMTPHost == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "SMTP服务器地址不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	if config.Username == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "邮箱用户名不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	if config.SMTPPort <= 0 || config.SMTPPort > 65535 {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "SMTP端口号无效，应在1-65535之间",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	// 首先测试TCP连接
	conn, err := net.DialTimeout("tcp", addr, 8*time.Second)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "connection refused") {
			errMsg = "连接被拒绝，请检查SMTP服务器地址和端口"
		} else if strings.Contains(errMsg, "timeout") {
			errMsg = "连接超时，请检查网络和防火墙设置"
		} else if strings.Contains(errMsg, "no route to host") {
			errMsg = "无法到达主机，请检查SMTP服务器地址"
		}

		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("SMTP TCP连接失败: %s", errMsg),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	conn.Close()

	// 尝试SMTP连接和认证
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)

	client, err := smtp.Dial(addr)
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("连接SMTP服务器失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer client.Close()

	// 如果端口是465或587，尝试StartTLS
	if config.SMTPPort == 587 || config.SMTPPort == 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(nil); err != nil {
				return &model.ConnectionTestResult{
					Success: false,
					Message: fmt.Sprintf("启动TLS失败: %s", err.Error()),
					Latency: time.Since(start).Milliseconds(),
				}
			}
		}
	}

	// 测试认证
	if err := client.Auth(auth); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "authentication failed") || strings.Contains(errMsg, "535") {
			errMsg = "SMTP认证失败，请检查用户名和密码"
		} else if strings.Contains(errMsg, "530") {
			errMsg = "SMTP服务器要求认证，请检查配置"
		}

		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("SMTP认证失败: %s", errMsg),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &model.ConnectionTestResult{
		Success: true,
		Message: "邮件服务器连接成功",
		Latency: time.Since(start).Milliseconds(),
	}
}

// TestJackettConnection 测试Jackett连接
func (s *ConfigService) TestJackettConnection(config model.JackettConfig) *model.ConnectionTestResult {
	start := time.Now()

	if config.Host == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Jackett主机地址不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	if config.APIKey == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "Jackett API密钥不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 测试Jackett API连接 - 使用搜索API进行测试（更可靠）
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 不自动跟随重定向
		},
	}
	
	// 使用搜索API端点进行测试，搜索一个简单的关键词
	testURL := fmt.Sprintf("%s/api/v2.0/indexers/all/results?apikey=%s&Query=test", config.Host, config.APIKey)

	resp, err := client.Get(testURL)
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("连接Jackett失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	// 200 OK 或 302 重定向都表示API密钥有效
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		// 如果是401或403，说明API密钥无效
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return &model.ConnectionTestResult{
				Success: false,
				Message: "Jackett API密钥无效，请检查密钥是否正确",
				Latency: time.Since(start).Milliseconds(),
			}
		}
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Jackett API响应错误: HTTP %d", resp.StatusCode),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &model.ConnectionTestResult{
		Success: true,
		Message: "Jackett连接成功",
		Latency: time.Since(start).Milliseconds(),
	}
}

// TestQBittorrentConnection 测试qBittorrent连接
func (s *ConfigService) TestQBittorrentConnection(config model.QBittorrentConfig) *model.ConnectionTestResult {
	start := time.Now()

	if config.Host == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "qBittorrent主机地址不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	if config.Username == "" {
		return &model.ConnectionTestResult{
			Success: false,
			Message: "qBittorrent用户名不能为空",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 测试qBittorrent API连接
	client := &http.Client{Timeout: 10 * time.Second}
	loginURL := fmt.Sprintf("%s/api/v2/auth/login", config.Host)

	resp, err := client.PostForm(loginURL, map[string][]string{
		"username": {config.Username},
		"password": {config.Password},
	})
	if err != nil {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("连接qBittorrent失败: %s", err.Error()),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &model.ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("qBittorrent登录失败: HTTP %d", resp.StatusCode),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &model.ConnectionTestResult{
		Success: true,
		Message: "qBittorrent连接成功",
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
