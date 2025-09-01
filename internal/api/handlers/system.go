package handlers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// SystemHandler 系统操作处理器
type SystemHandler struct{}

// NewSystemHandler 创建系统处理器
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// RestartServer 重启服务器
func (h *SystemHandler) RestartServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "重启命令已接收，服务器将在2秒后重启",
	})

	// 使用goroutine异步执行重启，避免阻塞响应
	go func() {
		// 等待2秒让响应完全发送
		time.Sleep(2 * time.Second)
		
		// 简单的重启策略：退出程序，让进程管理器重启
		// 在生产环境中，通常由systemd、supervisor等进程管理器监控并重启
		os.Exit(0)
	}()
}

// GetSystemInfo 获取系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
		},
	})
}