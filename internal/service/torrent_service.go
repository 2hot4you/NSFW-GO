package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// TorrentService 种子下载服务
type TorrentService struct {
	jackettHost     string
	jackettAPIKey   string
	qbittorrentHost string
	qbittorrentUser string
	qbittorrentPass string
	maxResults      int
	minSeeders      int
	sortBySize      bool
	timeout         time.Duration
	localMovieRepo  LocalMovieRepository
	telegramService *TelegramService
}

// LocalMovieRepository 本地影片仓库接口（定义在这里避免循环依赖）
type LocalMovieRepository interface {
	SearchByCode(code string) (*LocalMovie, error)
}

// LocalMovie 本地电影结构（简化版，避免循环依赖）
type LocalMovie struct {
	ID    uint   `json:"id"`
	Code  string `json:"code"`
	Title string `json:"title"`
}

// NewTorrentService 创建种子下载服务
func NewTorrentService(jackettHost, jackettAPIKey, qbittorrentHost, qbittorrentUser, qbittorrentPass string, localMovieRepo LocalMovieRepository) *TorrentService {
	return &TorrentService{
		jackettHost:     jackettHost,
		jackettAPIKey:   jackettAPIKey,
		qbittorrentHost: qbittorrentHost,
		qbittorrentUser: qbittorrentUser,
		qbittorrentPass: qbittorrentPass,
		maxResults:      20,
		minSeeders:      1,
		sortBySize:      true,
		timeout:         30 * time.Second,
		localMovieRepo:  localMovieRepo,
		telegramService: nil, // 将在路由设置中注入
	}
}

// SetTelegramService 设置 Telegram 服务（依赖注入）
func (s *TorrentService) SetTelegramService(telegramService *TelegramService) {
	s.telegramService = telegramService
}

// JackettResult Jackett搜索结果
type JackettResult struct {
	Title         string `json:"title"`
	Link          string `json:"link"`
	Size          int64  `json:"size"`
	SizeFormatted string `json:"sizeFormatted"`
	Seeders       int    `json:"seeders"`
	Leechers      int    `json:"leechers"`
	PublishDate   string `json:"publishDate"`
	MagnetURI     string `json:"magnetUri"`
	InfoHash      string `json:"infoHash"`
	Tracker       string `json:"tracker"`
	Category      string `json:"category"`
}

// JackettResponse Jackett API响应结构
type JackettResponse struct {
	Results []struct {
		Title       string `json:"Title"`
		TrackerID   string `json:"TrackerId"`
		Tracker     string `json:"Tracker"`
		CategoryID  int    `json:"CategoryId"`
		Category    []int  `json:"Category"` // 修复: Category 是数组而不是字符串
		Size        int64  `json:"Size"`
		Files       int    `json:"Files"`
		Grabs       int    `json:"Grabs"`
		Description string `json:"Description"`
		Link        string `json:"Link"`
		Comments    string `json:"Comments"`
		PublishDate string `json:"PublishDate"`
		Seeders     int    `json:"Seeders"`
		Peers       int    `json:"Peers"`
		InfoHash    string `json:"InfoHash"`
		MagnetURI   string `json:"MagnetUri"`
		MinSeedTime int    `json:"MinSeedTime"`
	} `json:"Results"`
}

// SearchTorrents 搜索种子（按文件大小排序）
func (s *TorrentService) SearchTorrents(keyword string) ([]JackettResult, error) {
	// 构建Jackett API URL，使用成人内容的分类ID
	categories := []string{
		"6000", "6010", "6060", "6080",
		"100431", "100437", "100410", "100424", 
		"100432", "100426", "100429", "100430",
		"100436", "100433", "100425",
	}

	// 构建分类参数
	categoryParams := ""
	for _, cat := range categories {
		categoryParams += "&Category%5B%5D=" + cat
	}

	apiURL := fmt.Sprintf("%s/api/v2.0/indexers/all/results?apikey=%s&Query=%s%s",
		s.jackettHost, s.jackettAPIKey, url.QueryEscape(keyword), categoryParams)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: s.timeout,
	}

	// 发送请求到Jackett
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("调用Jackett API失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jackett API返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var jackettResp JackettResponse
	if err := json.NewDecoder(resp.Body).Decode(&jackettResp); err != nil {
		return nil, fmt.Errorf("解析Jackett响应失败: %v", err)
	}

	// 转换为标准格式并过滤
	var results []JackettResult
	for _, item := range jackettResp.Results {
		// 过滤掉做种数不足的
		if item.Seeders < s.minSeeders {
			continue
		}

		// 格式化文件大小
		sizeFormatted := formatFileSize(item.Size)

		result := JackettResult{
			Title:         item.Title,
			Link:          item.Link,
			Size:          item.Size,
			SizeFormatted: sizeFormatted,
			Seeders:       item.Seeders,
			Leechers:      item.Peers,
			PublishDate:   item.PublishDate,
			MagnetURI:     item.MagnetURI,
			InfoHash:      item.InfoHash,
			Tracker:       item.Tracker,
			Category:      fmt.Sprintf("%v", item.Category), // 将数组转换为字符串
		}
		results = append(results, result)
	}

	// 按文件大小排序（从大到小）
	if s.sortBySize {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Size > results[j].Size
		})
	}

	// 限制结果数量
	if len(results) > s.maxResults {
		results = results[:s.maxResults]
	}

	return results, nil
}

// SearchTorrentsForCode 为特定番号搜索种子
func (s *TorrentService) SearchTorrentsForCode(code string) ([]JackettResult, error) {
	// 检查本地是否已存在该番号
	_, err := s.localMovieRepo.SearchByCode(code)
	if err == nil {
		// 找到了记录，说明已存在
		return nil, fmt.Errorf("番号 %s 已存在于本地影视库", code)
	}

	// 如果是 gorm.ErrRecordNotFound 或者其他"未找到"错误，继续搜索
	// 否则返回查询错误
	if err.Error() != "record not found" && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("检查本地电影失败: %v", err)
	}

	// 搜索种子
	return s.SearchTorrents(code)
}

// DownloadTorrentWithNotification 带通知的完整下载流程
func (s *TorrentService) DownloadTorrentWithNotification(magnetURI, downloadURI, code, title, tracker string, size int64) error {
	// 优先使用磁力链接，如果没有则使用HTTP下载链接
	var actualURI string
	var uriType string
	
	if magnetURI != "" {
		actualURI = magnetURI
		uriType = "磁力链接"
	} else if downloadURI != "" {
		actualURI = downloadURI
		uriType = "HTTP链接"
	} else {
		err := fmt.Errorf("没有可用的下载链接")
		if s.telegramService != nil {
			s.telegramService.SendNotification("error", map[string]interface{}{
				"error":     err.Error(),
				"component": "TorrentService",
				"code":      code,
			})
		}
		return err
	}

	// 添加种子到下载器
	err := s.DownloadTorrent(actualURI)
	if err != nil {
		// 如果HTTP链接失败且有磁力链接，尝试使用磁力链接
		if uriType == "HTTP链接" && magnetURI != "" {
			fmt.Printf("HTTP下载失败，尝试使用磁力链接: %s\n", err.Error())
			err = s.DownloadTorrent(magnetURI)
			if err == nil {
				actualURI = magnetURI
				uriType = "磁力链接(备用)"
			}
		}
		
		if err != nil {
			// 发送错误通知
			if s.telegramService != nil {
				s.telegramService.SendNotification("error", map[string]interface{}{
					"error":     fmt.Sprintf("下载失败 (%s): %s", uriType, err.Error()),
					"component": "TorrentService",
					"code":      code,
				})
			}
			return err
		}
	}

	// 发送成功通知
	if s.telegramService != nil {
		s.telegramService.SendNotification("download_started", map[string]interface{}{
			"code":     code,
			"title":    title,
			"size":     size,
			"tracker":  tracker,
			"uri_type": uriType,
		})
	}

	return nil
}

// DownloadTorrent 添加种子到qBittorrent (支持磁力链接和HTTP下载链接)
func (s *TorrentService) DownloadTorrent(downloadURI string) error {
	if downloadURI == "" {
		return fmt.Errorf("下载链接不能为空")
	}

	client := &http.Client{
		Timeout: s.timeout,
	}

	// 登录qBittorrent
	loginURL := fmt.Sprintf("%s/api/v2/auth/login", s.qbittorrentHost)
	loginData := url.Values{
		"username": {s.qbittorrentUser},
		"password": {s.qbittorrentPass},
	}

	resp, err := client.PostForm(loginURL, loginData)
	if err != nil {
		return fmt.Errorf("登录qBittorrent失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("登录qBittorrent失败，状态码: %d", resp.StatusCode)
	}

	// 获取Cookie
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		return fmt.Errorf("未获取到qBittorrent登录Cookie")
	}

	// 检查种子是否已存在
	existingTorrents, err := s.getTorrentListWithCookies(client, cookies)
	if err != nil {
		fmt.Printf("⚠️  无法检查现有种子列表: %v\n", err)
	} else {
		// 检查重复种子
		for _, torrent := range existingTorrents {
			if magnetURI, ok := torrent["magnet_uri"].(string); ok {
				if magnetURI == downloadURI {
					name := "未知"
					if n, ok := torrent["name"].(string); ok {
						name = n
					}
					return fmt.Errorf("种子 '%s' 已存在于下载列表中，无法重复添加", name)
				}
			}
		}
	}

	// 获取下载目录配置
	configStoreService := NewConfigStoreService()
	downloadPath := "/media/PornDB/Downloads" // 默认路径
	if config, err := configStoreService.GetConfig("torrent.download_path"); err == nil {
		downloadPath = strings.Trim(config.String(), "\"")
	}

	// 添加种子 - 支持磁力链接和HTTP下载链接
	addURL := fmt.Sprintf("%s/api/v2/torrents/add", s.qbittorrentHost)
	addData := url.Values{
		"urls":        {downloadURI}, // qBittorrent 的 urls 参数可以同时处理磁力链接和HTTP链接
		"savepath":    {downloadPath}, // 从配置获取下载目录
		"tags":        {"NSFW"}, // 添加NSFW标签
		"category":    {"NSFW"}, // 设置分类
		"paused":      {"false"}, // 立即开始下载
		"root_folder": {"false"}, // 不创建根文件夹
		"rename":      {""}, // 不重命名
		"upLimit":     {""}, // 无上传限制
		"dlLimit":     {""}, // 无下载限制
	}
	
	// 记录请求详情用于调试
	fmt.Printf("🔧 qBittorrent API 请求参数:\n")
	fmt.Printf("   URL: %s\n", addURL)
	fmt.Printf("   下载路径: %s\n", downloadPath)
	fmt.Printf("   下载URI: %s\n", downloadURI)
	fmt.Printf("   分类: NSFW\n")
	fmt.Printf("   标签: NSFW\n")

	req, err := http.NewRequest("POST", addURL, strings.NewReader(addData.Encode()))
	if err != nil {
		return fmt.Errorf("创建添加种子请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("添加种子到qBittorrent失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ qBittorrent API 错误响应:\n")
		fmt.Printf("   状态码: %d\n", resp.StatusCode)
		fmt.Printf("   响应内容: %s\n", string(body))
		return fmt.Errorf("添加种子失败，状态码: %d，响应: %s", resp.StatusCode, string(body))
	}

	// 成功响应
	body, _ := io.ReadAll(resp.Body)
	responseText := strings.TrimSpace(string(body))
	fmt.Printf("✅ qBittorrent API 成功响应: %s\n", responseText)
	
	// 检查是否是重复种子
	if responseText == "Fails." {
		return fmt.Errorf("检测到尝试添加重复 Torrent 文件，qBittorrent 已拒绝添加")
	}
	
	fmt.Printf("✅ 种子已添加到 qBittorrent，应保存至: %s\n", downloadPath)

	return nil
}

// GetTorrentList 获取qBittorrent中的种子列表
func (s *TorrentService) GetTorrentList() ([]map[string]interface{}, error) {
	client := &http.Client{
		Timeout: s.timeout,
	}

	// 登录qBittorrent
	loginURL := fmt.Sprintf("%s/api/v2/auth/login", s.qbittorrentHost)
	loginData := url.Values{
		"username": {s.qbittorrentUser},
		"password": {s.qbittorrentPass},
	}

	resp, err := client.PostForm(loginURL, loginData)
	if err != nil {
		return nil, fmt.Errorf("登录qBittorrent失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("登录qBittorrent失败，状态码: %d", resp.StatusCode)
	}

	// 获取Cookie
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		return nil, fmt.Errorf("未获取到qBittorrent登录Cookie")
	}

	// 获取种子列表
	listURL := fmt.Sprintf("%s/api/v2/torrents/info", s.qbittorrentHost)
	req, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建获取种子列表请求失败: %v", err)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取种子列表失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取种子列表失败，状态码: %d", resp.StatusCode)
	}

	var torrents []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, fmt.Errorf("解析种子列表失败: %v", err)
	}

	return torrents, nil
}

// getTorrentListWithCookies 使用现有cookies获取种子列表（内部辅助方法）
func (s *TorrentService) getTorrentListWithCookies(client *http.Client, cookies []*http.Cookie) ([]map[string]interface{}, error) {
	// 获取种子列表
	listURL := fmt.Sprintf("%s/api/v2/torrents/info", s.qbittorrentHost)
	req, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建获取种子列表请求失败: %v", err)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取种子列表失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取种子列表失败，状态码: %d", resp.StatusCode)
	}

	var torrents []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, fmt.Errorf("解析种子列表失败: %v", err)
	}

	return torrents, nil
}

// formatFileSize 格式化文件大小
func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB"}

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), sizes[exp+1])
}
