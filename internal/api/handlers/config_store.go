package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"nsfw-go/internal/service"

	"github.com/gin-gonic/gin"
)

// ConfigStoreHandler 数据库配置管理处理器
type ConfigStoreHandler struct {
	configStoreService *service.ConfigStoreService
}

// NewConfigStoreHandler 创建数据库配置处理器
func NewConfigStoreHandler() *ConfigStoreHandler {
	return &ConfigStoreHandler{
		configStoreService: service.NewConfigStoreService(),
	}
}

// GetAllConfigs 获取所有配置
func (h *ConfigStoreHandler) GetAllConfigs(c *gin.Context) {
	configs, err := h.configStoreService.GetAllConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("获取配置失败: %s", err.Error()),
		})
		return
	}

	// 隐藏敏感配置的值
	for i := range configs {
		if configs[i].IsSecret {
			configs[i].Value = "****"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// GetConfigsByCategory 根据分类获取配置
func (h *ConfigStoreHandler) GetConfigsByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "分类参数不能为空",
		})
		return
	}

	configs, err := h.configStoreService.GetConfigsByCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("获取配置失败: %s", err.Error()),
		})
		return
	}

	// 隐藏敏感配置的值
	for i := range configs {
		if configs[i].IsSecret {
			configs[i].Value = "****"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// GetConfig 获取单个配置
func (h *ConfigStoreHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "配置键不能为空",
		})
		return
	}

	configValue, err := h.configStoreService.GetConfig(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": fmt.Sprintf("配置不存在: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"key":   key,
			"value": configValue.String(),
		},
	})
}

// SetConfig 设置配置
func (h *ConfigStoreHandler) SetConfig(c *gin.Context) {
	var req struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value"`
		Type        string `json:"type"`
		Category    string `json:"category"`
		Description string `json:"description"`
		IsSecret    bool   `json:"is_secret"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("请求参数错误: %s", err.Error()),
		})
		return
	}

	// 设置默认类型
	if req.Type == "" {
		req.Type = "string"
	}

	err := h.configStoreService.SetConfig(
		req.Key,
		req.Value,
		req.Type,
		req.Category,
		req.Description,
		req.IsSecret,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("设置配置失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置设置成功",
	})
}

// BatchSetConfigs 批量设置配置
func (h *ConfigStoreHandler) BatchSetConfigs(c *gin.Context) {
	var req struct {
		Category string                 `json:"category" binding:"required"`
		Configs  map[string]interface{} `json:"configs" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("请求参数错误: %s", err.Error()),
		})
		return
	}

	err := h.configStoreService.BatchSetConfigs(req.Configs, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("批量设置配置失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "批量设置配置成功",
	})
}

// DeleteConfig 删除配置
func (h *ConfigStoreHandler) DeleteConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "配置键不能为空",
		})
		return
	}

	err := h.configStoreService.DeleteConfig(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("删除配置失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置删除成功",
	})
}

// SaveCurrentConfigToDB 将当前配置保存到数据库
func (h *ConfigStoreHandler) SaveCurrentConfigToDB(c *gin.Context) {
	// 这里可以从全局配置或服务中获取当前配置
	// 暂时返回成功，实际实现需要根据具体的配置管理方式来做
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置保存功能需要与现有配置系统集成",
	})
}

// MigrateFileConfigToDB 将配置文件迁移到数据库
func (h *ConfigStoreHandler) MigrateFileConfigToDB(c *gin.Context) {
	// 创建一些示例配置数据来演示功能
	sampleConfigs := map[string]interface{}{
		"server.host":         "0.0.0.0",
		"server.port":         8080,
		"server.mode":         "debug",
		"database.host":       "postgres",
		"database.port":       5432,
		"database.user":       "nsfw",
		"database.dbname":     "nsfw_db",
		"redis.host":          "redis",
		"redis.port":          6379,
		"media.base_path":     "/MediaCenter/NSFW/Hub/#Done",
		"media.scan_interval": 24,
	}

	err := h.configStoreService.BatchSetConfigs(sampleConfigs, "migrated")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("迁移配置失败: %s", err.Error()),
		})
		return
	}

	// 创建备份
	err = h.configStoreService.CreateBackup(
		"Initial Migration",
		"从配置文件自动迁移的初始配置备份",
		"system",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("创建备份失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "配置迁移完成",
		"migrated_count": len(sampleConfigs),
	})
}

// CreateBackup 创建配置备份
func (h *ConfigStoreHandler) CreateBackup(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("请求参数错误: %s", err.Error()),
		})
		return
	}

	// 获取创建者信息，这里可以从认证上下文中获取
	createdBy := "system" // 可以从JWT或session中获取真实用户

	err := h.configStoreService.CreateBackup(req.Name, req.Description, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("创建备份失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置备份创建成功",
	})
}

// GetBackups 获取配置备份列表
func (h *ConfigStoreHandler) GetBackups(c *gin.Context) {
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

// RestoreFromBackup 从备份恢复配置
func (h *ConfigStoreHandler) RestoreFromBackup(c *gin.Context) {
	backupIDStr := c.Param("id")
	backupID, err := strconv.ParseUint(backupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "备份ID格式错误",
		})
		return
	}

	err = h.configStoreService.RestoreFromBackup(uint(backupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("恢复配置失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置恢复成功",
	})
}

// DeleteBackup 删除配置备份
func (h *ConfigStoreHandler) DeleteBackup(c *gin.Context) {
	backupIDStr := c.Param("id")
	backupID, err := strconv.ParseUint(backupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "备份ID格式错误",
		})
		return
	}

	err = h.configStoreService.DeleteBackup(uint(backupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("删除备份失败: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "备份删除成功",
	})
}

// GetConfigCategories 获取配置分类
func (h *ConfigStoreHandler) GetConfigCategories(c *gin.Context) {
	// 这里可以从数据库获取配置分类，暂时返回硬编码的分类
	categories := []gin.H{
		{"name": "server", "display_name": "服务器配置", "description": "HTTP服务器相关配置"},
		{"name": "database", "display_name": "数据库配置", "description": "数据库连接和性能配置"},
		{"name": "redis", "display_name": "Redis配置", "description": "Redis缓存配置"},
		{"name": "media", "display_name": "媒体库配置", "description": "本地媒体库扫描配置"},
		{"name": "crawler", "display_name": "爬虫配置", "description": "网站爬取相关配置"},
		{"name": "security", "display_name": "安全配置", "description": "安全和认证相关配置"},
		{"name": "bot", "display_name": "Bot配置", "description": "Telegram Bot配置"},
		{"name": "torrent", "display_name": "种子下载配置", "description": "种子搜索和下载配置"},
		{"name": "notifications", "display_name": "通知配置", "description": "邮件和消息通知配置"},
		{"name": "log", "display_name": "日志配置", "description": "日志记录相关配置"},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}
