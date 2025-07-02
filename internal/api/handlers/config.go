package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"nsfw-go/internal/model"
	"nsfw-go/internal/service"
)

// ConfigHandler 配置管理处理器
type ConfigHandler struct {
	configService *service.ConfigService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

// GetConfig 获取系统配置
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	config, err := h.configService.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("获取配置失败: %s", err.Error()),
		})
		return
	}

	// 隐藏敏感信息
	sanitizedConfig := h.sanitizeConfig(config)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sanitizedConfig,
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

// sanitizeConfig 隐藏敏感配置信息
func (h *ConfigHandler) sanitizeConfig(config *model.SystemConfig) *model.SystemConfig {
	// 创建配置副本
	sanitized := *config

	// 隐藏敏感信息
	if sanitized.Database.Password != "" {
		sanitized.Database.Password = "****"
	}

	if sanitized.Redis.Password != "" {
		sanitized.Redis.Password = "****"
	}

	if sanitized.Bot.Token != "" {
		sanitized.Bot.Token = "****"
	}

	if sanitized.Security.JWTSecret != "" {
		sanitized.Security.JWTSecret = "****"
	}

	if sanitized.Security.PasswordSalt != "" {
		sanitized.Security.PasswordSalt = "****"
	}

	if sanitized.Notifications.Email.Password != "" {
		sanitized.Notifications.Email.Password = "****"
	}

	return &sanitized
}
