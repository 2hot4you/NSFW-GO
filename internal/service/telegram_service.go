package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type TelegramService struct {
	token   string
	chatID  string
	enabled bool
}

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func NewTelegramService(token, chatID string, enabled bool) *TelegramService {
	return &TelegramService{
		token:   token,
		chatID:  chatID,
		enabled: enabled,
	}
}

func (s *TelegramService) SendNotification(messageType string, data map[string]interface{}) error {
	if !s.enabled || s.token == "" || s.chatID == "" {
		return nil
	}

	message := s.formatMessage(messageType, data)
	return s.sendMessage(message)
}

func (s *TelegramService) formatMessage(messageType string, data map[string]interface{}) string {
	var message strings.Builder
	
	switch messageType {
	case "scan_complete":
		message.WriteString("📽️ *本地影片扫描完成*\n\n")
		if count, ok := data["count"].(int); ok {
			message.WriteString(fmt.Sprintf("✅ 扫描到 %d 部影片\n", count))
		}
		if newCount, ok := data["new_count"].(int); ok && newCount > 0 {
			message.WriteString(fmt.Sprintf("🆕 新增 %d 部影片\n", newCount))
		}
		if duration, ok := data["duration"].(time.Duration); ok {
			message.WriteString(fmt.Sprintf("⏱️ 耗时: %s\n", duration))
		}
		
	case "ranking_update":
		message.WriteString("🏆 *排行榜更新*\n\n")
		if rankType, ok := data["type"].(string); ok {
			switch rankType {
			case "daily":
				message.WriteString("📅 日榜更新\n")
			case "weekly":
				message.WriteString("📆 周榜更新\n")
			case "monthly":
				message.WriteString("🗓️ 月榜更新\n")
			}
		}
		if count, ok := data["count"].(int); ok {
			message.WriteString(fmt.Sprintf("📊 共 %d 部影片\n", count))
		}
		
	case "new_movie":
		message.WriteString("🎬 *发现新影片*\n\n")
		if code, ok := data["code"].(string); ok {
			message.WriteString(fmt.Sprintf("🔖 番号: `%s`\n", code))
		}
		if title, ok := data["title"].(string); ok {
			message.WriteString(fmt.Sprintf("📝 标题: %s\n", title))
		}
		if actress, ok := data["actress"].(string); ok {
			message.WriteString(fmt.Sprintf("👤 演员: %s\n", actress))
		}
		
	case "download_started":
		message.WriteString("📥 *开始下载*\n\n")
		if code, ok := data["code"].(string); ok {
			message.WriteString(fmt.Sprintf("🔖 番号: `%s`\n", code))
		}
		if title, ok := data["title"].(string); ok {
			message.WriteString(fmt.Sprintf("📝 标题: %s\n", title))
		}
		if size, ok := data["size"].(int64); ok && size > 0 {
			message.WriteString(fmt.Sprintf("💾 大小: %.2f GB\n", float64(size)/1024/1024/1024))
		}
		if tracker, ok := data["tracker"].(string); ok && tracker != "" {
			message.WriteString(fmt.Sprintf("🌐 站点: %s\n", tracker))
		}
		if uriType, ok := data["uri_type"].(string); ok && uriType != "" {
			message.WriteString(fmt.Sprintf("🔗 方式: %s\n", uriType))
		}
		message.WriteString("\n🚀 任务已添加到下载器")
		
	case "download_complete":
		message.WriteString("✅ *下载完成*\n\n")
		if name, ok := data["name"].(string); ok {
			message.WriteString(fmt.Sprintf("📁 文件: %s\n", name))
		}
		if size, ok := data["size"].(int64); ok {
			message.WriteString(fmt.Sprintf("💾 大小: %.2f GB\n", float64(size)/1024/1024/1024))
		}
		
	case "error":
		message.WriteString("❌ *错误通知*\n\n")
		if err, ok := data["error"].(string); ok {
			message.WriteString(fmt.Sprintf("⚠️ %s\n", err))
		}
		if component, ok := data["component"].(string); ok {
			message.WriteString(fmt.Sprintf("🔧 组件: %s\n", component))
		}
		
	case "test":
		message.WriteString("🔔 *测试通知*\n\n")
		message.WriteString("✅ Telegram 通知配置成功！\n")
		message.WriteString(fmt.Sprintf("🕐 时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
		message.WriteString("\n_这是一条测试消息_")
		
	default:
		message.WriteString("📢 *系统通知*\n\n")
		for k, v := range data {
			message.WriteString(fmt.Sprintf("%s: %v\n", k, v))
		}
	}
	
	return message.String()
}

func (s *TelegramService) sendMessage(text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.token)
	
	msg := TelegramMessage{
		ChatID:    s.chatID,
		Text:      text,
		ParseMode: "Markdown",
	}
	
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}
	
	return nil
}

func (s *TelegramService) TestConnection() error {
	return s.SendNotification("test", nil)
}

// SendTestNotification 发送测试通知到指定聊天
func (s *TelegramService) SendTestNotification(chatID string, message string) error {
	if s.token == "" {
		return fmt.Errorf("telegram bot token未配置")
	}

	if chatID == "" {
		return fmt.Errorf("聊天ID不能为空")
	}

	if message == "" {
		message = "🎉 测试通知！\n\n如果您收到了这条消息，说明Telegram通知配置成功。"
	}

	// 构建API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.token)

	// 构建请求参数
	params := map[string]interface{}{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("序列化请求参数失败: %v", err)
	}

	// 发送请求
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("telegram API返回错误 (状态码: %d): %s", resp.StatusCode, string(body))
	}

	return nil
}