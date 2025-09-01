package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL驱动

	"nsfw-go/internal/model"
	"nsfw-go/internal/service"
)

// ConfigHandler 配置管理处理器
type ConfigHandler struct {
	configService      *service.ConfigService
	configStoreService *service.ConfigStoreService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService:      configService,
		configStoreService: service.NewConfigStoreService(),
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

// SaveConfig 保存系统配置
func (h *ConfigHandler) SaveConfig(c *gin.Context) {
	var config model.SystemConfig

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("配置格式错误: %s", err.Error()),
		})
		return
	}

	// 验证配置
	if errors := h.configService.ValidateConfig(&config); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "配置验证失败",
			"errors":  errors,
		})
		return
	}

	if err := h.configService.SaveConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("保存配置失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置保存成功",
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

// GetConfigBackups 获取配置备份列表
func (h *ConfigHandler) GetConfigBackups(c *gin.Context) {
	backups, err := h.configService.GetConfigBackups()
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
	backupName := c.Param("backup")
	if backupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "备份文件名不能为空",
		})
		return
	}

	if err := h.configService.RestoreConfigBackup(backupName); err != nil {
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
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
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
