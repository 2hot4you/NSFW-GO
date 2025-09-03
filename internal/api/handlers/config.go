package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL驱动

	"nsfw-go/internal/model"
	"nsfw-go/internal/service"
)

// ConfigHandler 配置管理处理器
type ConfigHandler struct {
	configService      *service.ConfigService
	configStoreService *service.ConfigStoreService
	telegramService    *service.TelegramService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(configService *service.ConfigService, telegramService *service.TelegramService) *ConfigHandler {
	return &ConfigHandler{
		configService:      configService,
		configStoreService: service.NewConfigStoreService(),
		telegramService:    telegramService,
	}
}

// GetConfig 获取系统配置
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	// 优先从数据库获取配置
	dbConfigs, err := h.configStoreService.GetAllConfigs()
	if err == nil && len(dbConfigs) > 0 {
		// 将扁平的数据库配置转换为嵌套结构
		nestedConfig := h.convertToNestedConfig(dbConfigs)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nestedConfig,
			"source":  "database",
		})
		return
	}

	// 如果数据库配置不可用，回退到文件配置
	config, err := h.configService.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("获取配置失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
		"source":  "file",
	})
}

// SaveConfig 保存系统配置到数据库
func (h *ConfigHandler) SaveConfig(c *gin.Context) {
	var requestData map[string]interface{}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("配置格式错误: %s", err.Error()),
		})
		return
	}

	// 创建备份
	backupName := fmt.Sprintf("api_backup_%d", time.Now().Unix())
	if err := h.configStoreService.CreateBackup(backupName, "API配置更新前备份", "api"); err != nil {
		// 备份失败不阻止保存，只记录警告
		fmt.Printf("Warning: Failed to create backup: %v\n", err)
	}

	// 将嵌套的配置结构扁平化并保存到数据库
	flatConfigs := h.flattenConfig(requestData, "")
	
	// 批量更新配置到数据库
	for key, value := range flatConfigs {
		valueStr, valueType := h.serializeConfigValue(value)
		category := h.extractCategory(key)
		
		if err := h.configStoreService.SetConfig(key, valueStr, valueType, category, "", false); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": fmt.Sprintf("保存配置失败: %s", err.Error()),
			})
			return
		}
	}

	// 同时保存到文件（作为备份）
	var config model.SystemConfig
	configBytes, _ := json.Marshal(requestData)
	if err := json.Unmarshal(configBytes, &config); err == nil {
		// 尝试保存到文件，失败不影响响应
		h.configService.SaveConfig(&config)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置保存成功",
		"backup":  backupName,
	})
}

// TestConnection 测试连接
func (h *ConfigHandler) TestConnection(c *gin.Context) {
	var request model.ConfigTestRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("请求格式错误: %s", err.Error()),
		})
		return
	}

	var result *model.ConnectionTestResult

	switch request.Type {
	case "database":
		configBytes, _ := json.Marshal(request.Data)
		var dbConfig model.DatabaseConfig
		if err := json.Unmarshal(configBytes, &dbConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "数据库配置格式错误",
			})
			return
		}
		result = h.configService.TestDatabaseConnection(dbConfig)

	case "redis":
		configBytes, _ := json.Marshal(request.Data)
		var redisConfig model.RedisConfig
		if err := json.Unmarshal(configBytes, &redisConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Redis配置格式错误",
			})
			return
		}
		result = h.configService.TestRedisConnection(redisConfig)

	case "telegram":
		configBytes, _ := json.Marshal(request.Data)
		var botConfig model.BotConfig
		if err := json.Unmarshal(configBytes, &botConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Telegram配置格式错误",
			})
			return
		}
		result = h.configService.TestTelegramConnection(botConfig)

	case "email":
		configBytes, _ := json.Marshal(request.Data)
		var emailConfig model.EmailNotificationConfig
		if err := json.Unmarshal(configBytes, &emailConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "邮件配置格式错误",
			})
			return
		}
		result = h.configService.TestEmailConnection(emailConfig)

	case "jackett":
		configBytes, _ := json.Marshal(request.Data)
		var jackettConfig model.JackettConfig
		if err := json.Unmarshal(configBytes, &jackettConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Jackett配置格式错误",
			})
			return
		}
		result = h.configService.TestJackettConnection(jackettConfig)

	case "qbittorrent":
		configBytes, _ := json.Marshal(request.Data)
		var qbittorrentConfig model.QBittorrentConfig
		if err := json.Unmarshal(configBytes, &qbittorrentConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "qBittorrent配置格式错误",
			})
			return
		}
		result = h.configService.TestQBittorrentConnection(qbittorrentConfig)

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的连接类型",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetConfigByCategory 根据分类获取配置
func (h *ConfigHandler) GetConfigByCategory(c *gin.Context) {
	category := c.Param("category")
	
	configs, err := h.configStoreService.GetConfigsByCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("获取配置失败: %s", err.Error()),
		})
		return
	}

	// 转换为嵌套结构
	result := make(map[string]interface{})
	for _, cfg := range configs {
		// 移除分类前缀
		key := strings.TrimPrefix(cfg.Key, category+".")
		value := h.parseConfigValue(cfg.Value, cfg.Type)
		result[key] = value
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"category": category,
	})
}

// GetConfigCategories 获取所有配置分类
func (h *ConfigHandler) GetConfigCategories(c *gin.Context) {
	configs, err := h.configStoreService.GetAllConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("获取配置失败: %s", err.Error()),
		})
		return
	}

	// 提取所有分类
	categoryMap := make(map[string]int)
	for _, cfg := range configs {
		categoryMap[cfg.Category]++
	}

	categories := []map[string]interface{}{}
	for cat, count := range categoryMap {
		categories = append(categories, map[string]interface{}{
			"name":  cat,
			"count": count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

// GetConfigBackups 获取配置备份列表
func (h *ConfigHandler) GetConfigBackups(c *gin.Context) {
	// 从数据库获取备份列表
	backups, err := h.configStoreService.GetBackups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("获取备份列表失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    backups,
	})
}

// RestoreConfigBackup 恢复配置备份
func (h *ConfigHandler) RestoreConfigBackup(c *gin.Context) {
	backupID := c.Param("id")
	if backupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "备份ID不能为空",
		})
		return
	}

	// 将字符串ID转换为uint
	id, err := strconv.ParseUint(backupID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的备份ID",
		})
		return
	}

	if err := h.configStoreService.RestoreFromBackup(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("恢复备份失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置恢复成功",
	})
}

// ValidateConfig 验证配置
func (h *ConfigHandler) ValidateConfig(c *gin.Context) {
	var config model.SystemConfig

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("配置格式错误: %s", err.Error()),
		})
		return
	}

	errors := h.configService.ValidateConfig(&config)

	c.JSON(http.StatusOK, gin.H{
		"success": len(errors) == 0,
		"errors":  errors,
	})
}

// convertToNestedConfig 将扁平的配置转换为嵌套结构
func (h *ConfigHandler) convertToNestedConfig(configs []model.ConfigStore) map[string]interface{} {
	result := make(map[string]interface{})

	for _, cfg := range configs {
		keys := strings.Split(cfg.Key, ".")
		current := result

		// 遍历键路径，创建嵌套结构
		for i, key := range keys {
			if i == len(keys)-1 {
				// 最后一个键，设置值
				value := h.parseConfigValue(cfg.Value, cfg.Type)
				current[key] = value
			} else {
				// 中间键，确保存在嵌套map
				if _, exists := current[key]; !exists {
					current[key] = make(map[string]interface{})
				}
				current = current[key].(map[string]interface{})
			}
		}
	}

	return result
}

// parseConfigValue 根据类型解析配置值
func (h *ConfigHandler) parseConfigValue(value, valueType string) interface{} {
	switch valueType {
	case "int":
		var intVal int
		if err := json.Unmarshal([]byte(value), &intVal); err == nil {
			return intVal
		}
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
		return 0
	case "bool":
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
		return false
	case "float":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
		return 0.0
	case "array", "json":
		var result interface{}
		if err := json.Unmarshal([]byte(value), &result); err == nil {
			// 特殊处理时间相关的配置
			if num, ok := result.(float64); ok {
				// 检查是否是时间纳秒值，转换为字符串格式
				if num > 1000000000 { // 大于1秒的纳秒值
					return fmt.Sprintf("%ds", int(num/1000000000))
				}
			}
			return result
		}
		return value
	case "string":
		// 去掉JSON字符串的引号
		var stringVal string
		if err := json.Unmarshal([]byte(value), &stringVal); err == nil {
			return stringVal
		}
		return value
	default:
		return value
	}
}

// flattenConfig 将嵌套的配置结构扁平化
func (h *ConfigHandler) flattenConfig(config map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range config {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// 递归处理嵌套map
			nested := h.flattenConfig(v, fullKey)
			for k, val := range nested {
				result[k] = val
			}
		default:
			// 直接存储值
			result[fullKey] = value
		}
	}

	return result
}

// serializeConfigValue 序列化配置值
func (h *ConfigHandler) serializeConfigValue(value interface{}) (string, string) {
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v), "bool"
	case int:
		return strconv.Itoa(v), "int"
	case int64:
		return strconv.FormatInt(v, 10), "int"
	case float64:
		// 检查是否是整数
		if float64(int(v)) == v {
			return strconv.Itoa(int(v)), "int"
		}
		return strconv.FormatFloat(v, 'f', -1, 64), "float"
	case string:
		// 字符串需要JSON编码以保持引号
		jsonStr, _ := json.Marshal(v)
		return string(jsonStr), "string"
	case []interface{}, map[string]interface{}:
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes), "json"
	default:
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes), "json"
	}
}

// extractCategory 从配置键提取分类
func (h *ConfigHandler) extractCategory(key string) string {
	parts := strings.Split(key, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "general"
}

// TestNotification 测试通知发送
func (h *ConfigHandler) TestNotification(c *gin.Context) {
	var req struct {
		Type    string `json:"type"`
		ChatID  string `json:"chat_id"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
			"error":   err.Error(),
		})
		return
	}

	// 只支持Telegram通知测试
	if req.Type != "telegram" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的通知类型",
		})
		return
	}

	// 检查Telegram服务是否初始化
	if h.telegramService == nil {
		// 尝试重新初始化
		config, err := h.configService.GetConfig()
		if err != nil || config == nil || !config.Bot.Enabled {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Telegram Bot未启用或配置不正确",
			})
			return
		}
		// 使用空的默认聊天ID，测试时会指定
		h.telegramService = service.NewTelegramService(config.Bot.Token, "", config.Bot.Enabled)
	}

	// 发送测试通知
	err := h.telegramService.SendTestNotification(req.ChatID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "发送通知失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "测试通知已发送",
	})
}
