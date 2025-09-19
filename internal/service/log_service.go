package service

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// LogEntry 日志条目结构
type LogEntry struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Category  string    `json:"category"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// LogService 日志服务
type LogService struct {
	logDir    string
	mu        sync.RWMutex
	logFiles  map[string]*os.File
	loggers   map[string]*log.Logger
	counter   int
}

// NewLogService 创建日志服务
func NewLogService(logDir string) *LogService {
	// 确保日志目录存在
	os.MkdirAll(logDir, 0755)

	service := &LogService{
		logDir:   logDir,
		logFiles: make(map[string]*os.File),
		loggers:  make(map[string]*log.Logger),
		counter:  0,
	}

	// 初始化各类日志文件
	categories := []string{"system", "crawler", "scanner", "torrent", "config"}
	for _, category := range categories {
		service.initCategoryLogger(category)
	}

	// 记录服务启动日志
	service.LogInfo("system", "log-service", "日志服务已启动，日志目录: "+logDir)

	return service
}

// initCategoryLogger 初始化分类日志记录器
func (s *LogService) initCategoryLogger(category string) {
	filename := filepath.Join(s.logDir, fmt.Sprintf("%s.log", category))

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("创建日志文件失败 %s: %v", filename, err)
		return
	}

	logger := log.New(file, "", 0) // 不使用默认前缀，我们自己格式化

	s.logFiles[category] = file
	s.loggers[category] = logger
}

// LogInfo 记录信息日志
func (s *LogService) LogInfo(category, source, message string) {
	s.writeLog("INFO", category, source, message)
}

// LogWarn 记录警告日志
func (s *LogService) LogWarn(category, source, message string) {
	s.writeLog("WARN", category, source, message)
}

// LogError 记录错误日志
func (s *LogService) LogError(category, source, message string) {
	s.writeLog("ERROR", category, source, message)
}

// LogDebug 记录调试日志
func (s *LogService) LogDebug(category, source, message string) {
	s.writeLog("DEBUG", category, source, message)
}

// writeLog 写入日志
func (s *LogService) writeLog(level, category, source, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	timestamp := time.Now()

	// 格式化日志行
	logLine := fmt.Sprintf("[%s] [%s] [%s] [%s] %s",
		timestamp.Format("2006-01-02 15:04:05"),
		level,
		category,
		source,
		message,
	)

	// 写入到对应分类的日志文件
	if logger, exists := s.loggers[category]; exists {
		logger.Println(logLine)
	} else {
		// 如果分类不存在，写入到系统日志
		if systemLogger, exists := s.loggers["system"]; exists {
			systemLogger.Printf("[%s] %s", category, logLine)
		}
	}

	// 同时写入到总日志文件
	s.writeToMasterLog(logLine)
}

// writeToMasterLog 写入主日志文件
func (s *LogService) writeToMasterLog(logLine string) {
	masterFile := filepath.Join(s.logDir, "app.log")

	file, err := os.OpenFile(masterFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(logLine + "\n")
}

// GetLogs 获取日志条目
func (s *LogService) GetLogs(category string, level string, limit int, offset int) ([]*LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var logs []*LogEntry
	var filenames []string

	// 确定要读取的日志文件
	if category == "" || category == "all" {
		// 读取所有日志文件
		for cat := range s.logFiles {
			filenames = append(filenames, filepath.Join(s.logDir, fmt.Sprintf("%s.log", cat)))
		}
	} else {
		// 读取指定分类的日志文件
		filename := filepath.Join(s.logDir, fmt.Sprintf("%s.log", category))
		filenames = append(filenames, filename)
	}

	// 读取并解析日志文件
	for _, filename := range filenames {
		entries, err := s.parseLogFile(filename)
		if err != nil {
			continue // 忽略读取错误的文件
		}
		logs = append(logs, entries...)
	}

	// 按时间戳排序（最新的在前）
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	// 应用级别过滤
	if level != "" && level != "all" {
		var filteredLogs []*LogEntry
		for _, entry := range logs {
			if strings.EqualFold(entry.Level, level) {
				filteredLogs = append(filteredLogs, entry)
			}
		}
		logs = filteredLogs
	}

	// 应用分页
	if offset >= len(logs) {
		return []*LogEntry{}, nil
	}

	end := offset + limit
	if end > len(logs) {
		end = len(logs)
	}

	return logs[offset:end], nil
}

// parseLogFile 解析日志文件
func (s *LogService) parseLogFile(filename string) ([]*LogEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []*LogEntry
	scanner := bufio.NewScanner(file)
	id := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		entry := s.parseLogLine(line, id)
		if entry != nil {
			entries = append(entries, entry)
			id++
		}
	}

	return entries, scanner.Err()
}

// parseLogLine 解析日志行
func (s *LogService) parseLogLine(line string, id int) *LogEntry {
	// 解析格式: [2006-01-02 15:04:05] [LEVEL] [CATEGORY] [SOURCE] MESSAGE
	if !strings.HasPrefix(line, "[") {
		return nil
	}

	parts := strings.SplitN(line, "] ", 5)
	if len(parts) < 5 {
		return nil
	}

	// 解析时间戳
	timestampStr := strings.TrimPrefix(parts[0], "[")
	timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		timestamp = time.Now()
	}

	// 解析其他字段
	level := strings.Trim(parts[1], "[]")
	category := strings.Trim(parts[2], "[]")
	source := strings.Trim(parts[3], "[]")
	message := parts[4]

	return &LogEntry{
		ID:        id,
		Timestamp: timestamp,
		Level:     level,
		Category:  category,
		Message:   message,
		Source:    source,
	}
}

// ClearLogs 清空日志文件
func (s *LogService) ClearLogs(category string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if category == "" || category == "all" {
		// 清空所有日志文件
		for cat, file := range s.logFiles {
			file.Close()
			filename := filepath.Join(s.logDir, fmt.Sprintf("%s.log", cat))
			os.Truncate(filename, 0)
			s.initCategoryLogger(cat)
		}
		// 清空主日志文件
		masterFile := filepath.Join(s.logDir, "app.log")
		os.Truncate(masterFile, 0)
	} else {
		// 清空指定分类的日志文件
		if file, exists := s.logFiles[category]; exists {
			file.Close()
			filename := filepath.Join(s.logDir, fmt.Sprintf("%s.log", category))
			os.Truncate(filename, 0)
			s.initCategoryLogger(category)
		}
	}

	s.counter = 0
	return nil
}

// Close 关闭日志服务
func (s *LogService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, file := range s.logFiles {
		file.Close()
	}
}

// GetLogStats 获取日志统计
func (s *LogService) GetLogStats() (map[string]int, error) {
	stats := make(map[string]int)

	logs, err := s.GetLogs("", "", 10000, 0) // 获取最近10000条日志用于统计
	if err != nil {
		return stats, err
	}

	stats["total"] = len(logs)
	for _, log := range logs {
		stats[strings.ToLower(log.Level)]++
	}

	return stats, nil
}