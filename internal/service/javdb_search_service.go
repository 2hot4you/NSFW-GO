package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nsfw-go/internal/crawler"

	"github.com/gocolly/colly/v2"
)

// JAVDbSearchService JAVDb搜索服务
type JAVDbSearchService struct {
	baseURL    string
	config     *crawler.CrawlerConfig
	logService *LogService
}

// NewJAVDbSearchService 创建JAVDb搜索服务
func NewJAVDbSearchService(config *crawler.CrawlerConfig, logService *LogService) *JAVDbSearchService {
	return &JAVDbSearchService{
		baseURL:    "https://javdb.com",
		config:     config,
		logService: logService,
	}
}

// MovieSearchResult 影片搜索结果
type MovieSearchResult struct {
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	CoverURL    string    `json:"cover_url"`
	Rating      float32   `json:"rating"`
	ReleaseDate time.Time `json:"release_date"`
	DetailURL   string    `json:"detail_url"`
}

// ActressSearchResult 演员搜索结果
type ActressSearchResult struct {
	Name       string `json:"name"`
	AvatarURL  string `json:"avatar_url"`
	DetailURL  string `json:"detail_url"`
	MovieCount int    `json:"movie_count"`
}

// ActressMovieResult 演员作品结果
type ActressMovieResult struct {
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	CoverURL    string    `json:"cover_url"`
	ReleaseDate time.Time `json:"release_date"`
	Rating      float32   `json:"rating"`
	DetailURL   string    `json:"detail_url"`
}

// SearchMovieByCode 根据番号搜索影片
func (s *JAVDbSearchService) SearchMovieByCode(ctx context.Context, code string) (*MovieSearchResult, error) {
	searchURL := fmt.Sprintf("%s/search?q=%s", s.baseURL, url.QueryEscape(code))

	c := colly.NewCollector()
	c.SetRequestTimeout(s.config.Timeout)

	// 设置User-Agent和其他请求头
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", s.config.UserAgents[0])
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
	})

	// 添加响应处理
	c.OnResponse(func(r *colly.Response) {
		if s.logService != nil {
		s.logService.LogInfo("crawler", "javdb-search", fmt.Sprintf("访问 %s 成功，状态码: %d", r.Request.URL, r.StatusCode))
	}
	})

	var result *MovieSearchResult
	var searchErr error

	// 查找完全匹配的番号
	c.OnHTML(".movie-list .item", func(e *colly.HTMLElement) {
		if result != nil {
			return // 已找到结果，跳过
		}

		// 获取标题和链接
		titleEl := e.DOM.Find(".video-title")
		if titleEl.Length() == 0 {
			return
		}

		title := strings.TrimSpace(titleEl.Text())

		// 查找链接（可能在父级或其他元素中）
		var detailURL string
		if linkEl := e.DOM.Find("a"); linkEl.Length() > 0 {
			if href, exists := linkEl.Attr("href"); exists {
				detailURL = href
			}
		}

		if detailURL == "" {
			if s.logService != nil {
			s.logService.LogWarn("crawler", "javdb-search", fmt.Sprintf("未找到链接，标题: %s", title))
		}
			return
		}

		// 构建完整URL
		if fullURL, err := s.buildURL(detailURL); err == nil {
			detailURL = fullURL
		}

		// 从标题中提取番号
		extractedCode := s.extractMovieCode(title)
		normalizedExtracted := s.normalizeCode(extractedCode)
		normalizedInput := s.normalizeCode(code)

		// 检查番号是否完全匹配
		if normalizedExtracted == normalizedInput {
			movieResult := &MovieSearchResult{
				Code:      extractedCode,
				Title:     title,
				DetailURL: detailURL,
			}

			// 获取封面图片
			imgEl := e.DOM.Find(".cover img")
			if imgEl.Length() > 0 {
				if src, exists := imgEl.Attr("src"); exists {
					if fullURL, err := s.buildURL(src); err == nil {
						movieResult.CoverURL = fullURL
					}
				}
				// 尝试data-src（懒加载）
				if movieResult.CoverURL == "" {
					if dataSrc, exists := imgEl.Attr("data-src"); exists {
						if fullURL, err := s.buildURL(dataSrc); err == nil {
							movieResult.CoverURL = fullURL
						}
					}
				}
			}

			// 获取评分
			ratingEl := e.DOM.Find(".score")
			if ratingEl.Length() > 0 {
				ratingText := strings.TrimSpace(ratingEl.Text())
				// 从文本中提取评分数字，如"4.5分, 由2355人評價" -> "4.5"
				re := regexp.MustCompile(`(\d+\.?\d*)分`)
				if matches := re.FindStringSubmatch(ratingText); len(matches) > 1 {
					if rating, err := strconv.ParseFloat(matches[1], 32); err == nil {
						movieResult.Rating = float32(rating)
					}
				}
			}

			// 获取发布日期
			dateEl := e.DOM.Find(".meta")
			if dateEl.Length() > 0 {
				dateText := strings.TrimSpace(dateEl.Text())
				movieResult.ReleaseDate = s.parseReleaseDate(dateText)
			}

			result = movieResult
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		if s.logService != nil {
		s.logService.LogError("crawler", "javdb-search", fmt.Sprintf("访问失败: URL=%s, 状态码=%d, 错误=%v", r.Request.URL, r.StatusCode, err))
	}
		searchErr = fmt.Errorf("搜索失败: %v", err)
	})

	// 执行搜索
	err := c.Visit(searchURL)
	if err != nil {
		return nil, fmt.Errorf("访问搜索页面失败: %v", err)
	}

	c.Wait()

	if searchErr != nil {
		return nil, searchErr
	}

	if result == nil {
		return nil, fmt.Errorf("未找到番号为 %s 的影片", code)
	}

	if s.logService != nil {
		s.logService.LogInfo("crawler", "javdb-search", fmt.Sprintf("找到番号 %s 的影片: %s", code, result.Title))
	}
	return result, nil
}

// SearchActressByName 根据名字搜索演员
func (s *JAVDbSearchService) SearchActressByName(ctx context.Context, actressName string) (*ActressSearchResult, error) {
	searchURL := fmt.Sprintf("%s/search?q=%s&f=actor", s.baseURL, url.QueryEscape(actressName))

	c := colly.NewCollector()
	c.SetRequestTimeout(s.config.Timeout)

	// 设置User-Agent和其他请求头
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", s.config.UserAgents[0])
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
	})

	// 添加响应处理
	c.OnResponse(func(r *colly.Response) {
		if s.logService != nil {
		s.logService.LogInfo("crawler", "javdb-actress-search", fmt.Sprintf("访问 %s 成功，状态码: %d", r.Request.URL, r.StatusCode))
	}
	})

	var result *ActressSearchResult
	var searchErr error

	// 查找演员信息
	c.OnHTML(".actor-list .item", func(e *colly.HTMLElement) {
		if result != nil {
			return // 已找到结果，跳过
		}

		// 获取演员名字和链接
		nameEl := e.DOM.Find(".actor-name a")
		if nameEl.Length() == 0 {
			nameEl = e.DOM.Find(".actor-name")
		}

		if nameEl.Length() == 0 {
			return
		}

		name := strings.TrimSpace(nameEl.Text())

		// 检查名字是否匹配（支持部分匹配）
		if !strings.Contains(strings.ToLower(name), strings.ToLower(actressName)) {
			return
		}

		// 获取演员详情页面链接
		var detailURL string
		if linkEl := e.DOM.Find("a"); linkEl.Length() > 0 {
			if href, exists := linkEl.Attr("href"); exists {
				if fullURL, err := s.buildURL(href); err == nil {
					detailURL = fullURL
				}
			}
		}

		actressResult := &ActressSearchResult{
			Name:      name,
			DetailURL: detailURL,
		}

		// 获取头像
		imgEl := e.DOM.Find("img")
		if imgEl.Length() > 0 {
			if src, exists := imgEl.Attr("src"); exists {
				if fullURL, err := s.buildURL(src); err == nil {
					actressResult.AvatarURL = fullURL
				}
			}
			// 尝试data-src（懒加载）
			if actressResult.AvatarURL == "" {
				if dataSrc, exists := imgEl.Attr("data-src"); exists {
					if fullURL, err := s.buildURL(dataSrc); err == nil {
						actressResult.AvatarURL = fullURL
					}
				}
			}
		}

		// 获取作品数量（如果页面有显示）
		countEl := e.DOM.Find(".movie-count")
		if countEl.Length() > 0 {
			countText := strings.TrimSpace(countEl.Text())
			// 提取数字
			re := regexp.MustCompile(`\d+`)
			if matches := re.FindString(countText); matches != "" {
				if count, err := strconv.Atoi(matches); err == nil {
					actressResult.MovieCount = count
				}
			}
		}

		result = actressResult
	})

	c.OnError(func(r *colly.Response, err error) {
		if s.logService != nil {
		s.logService.LogError("crawler", "javdb-actress-search", fmt.Sprintf("访问失败: URL=%s, 状态码=%d, 错误=%v", r.Request.URL, r.StatusCode, err))
	}
		searchErr = fmt.Errorf("搜索演员失败: %v", err)
	})

	// 执行搜索
	err := c.Visit(searchURL)
	if err != nil {
		return nil, fmt.Errorf("访问演员搜索页面失败: %v", err)
	}

	c.Wait()

	if searchErr != nil {
		return nil, searchErr
	}

	if result == nil {
		return nil, fmt.Errorf("未找到演员: %s", actressName)
	}

	if s.logService != nil {
		s.logService.LogInfo("crawler", "javdb-actress-search", fmt.Sprintf("找到演员: %s", result.Name))
	}
	return result, nil
}

// GetActressMovies 获取演员的所有作品
func (s *JAVDbSearchService) GetActressMovies(ctx context.Context, actressURL string) ([]ActressMovieResult, error) {
	c := colly.NewCollector()
	c.SetRequestTimeout(s.config.Timeout)

	// 设置User-Agent和其他请求头
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", s.config.UserAgents[0])
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
	})

	var results []ActressMovieResult
	var searchErr error

	// 解析演员作品页面
	c.OnHTML(".movie-list .item", func(e *colly.HTMLElement) {
		// 获取标题和链接
		titleEl := e.DOM.Find(".video-title a")
		if titleEl.Length() == 0 {
			return
		}

		title := strings.TrimSpace(titleEl.Text())
		detailURL, exists := titleEl.Attr("href")
		if !exists {
			return
		}

		// 构建完整URL
		if fullURL, err := s.buildURL(detailURL); err == nil {
			detailURL = fullURL
		}

		// 提取番号
		code := s.extractMovieCode(title)

		movieResult := ActressMovieResult{
			Code:      code,
			Title:     title,
			DetailURL: detailURL,
		}

		// 获取封面图片
		imgEl := e.DOM.Find(".cover img")
		if imgEl.Length() > 0 {
			if src, exists := imgEl.Attr("src"); exists {
				if fullURL, err := s.buildURL(src); err == nil {
					movieResult.CoverURL = fullURL
				}
			}
			// 尝试data-src（懒加载）
			if movieResult.CoverURL == "" {
				if dataSrc, exists := imgEl.Attr("data-src"); exists {
					if fullURL, err := s.buildURL(dataSrc); err == nil {
						movieResult.CoverURL = fullURL
					}
				}
			}
		}

		// 获取发布日期
		dateEl := e.DOM.Find(".meta")
		if dateEl.Length() > 0 {
			dateText := strings.TrimSpace(dateEl.Text())
			movieResult.ReleaseDate = s.parseReleaseDate(dateText)
		}

		// 获取评分
		ratingEl := e.DOM.Find(".score")
		if ratingEl.Length() > 0 {
			ratingText := strings.TrimSpace(ratingEl.Text())
			if rating, err := strconv.ParseFloat(ratingText, 32); err == nil {
				movieResult.Rating = float32(rating)
			}
		}

		results = append(results, movieResult)
	})

	c.OnError(func(r *colly.Response, err error) {
		if s.logService != nil {
		s.logService.LogError("crawler", "javdb-actress-movies", fmt.Sprintf("访问失败: URL=%s, 状态码=%d, 错误=%v", r.Request.URL, r.StatusCode, err))
	}
		searchErr = fmt.Errorf("获取演员作品失败: %v", err)
	})

	// 执行访问
	err := c.Visit(actressURL)
	if err != nil {
		return nil, fmt.Errorf("访问演员页面失败: %v", err)
	}

	c.Wait()

	if searchErr != nil {
		return nil, searchErr
	}

	if s.logService != nil {
		s.logService.LogInfo("crawler", "javdb-actress-movies", fmt.Sprintf("获取演员作品完成，共 %d 部", len(results)))
	}
	return results, nil
}

// 辅助方法

// buildURL 构建完整URL
func (s *JAVDbSearchService) buildURL(path string) (string, error) {
	if strings.HasPrefix(path, "http") {
		return path, nil
	}

	base, err := url.Parse(s.baseURL)
	if err != nil {
		return "", err
	}

	relative, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(relative).String(), nil
}

// extractMovieCode 从标题中提取番号
func (s *JAVDbSearchService) extractMovieCode(title string) string {
	// 常见的番号格式正则表达式
	patterns := []string{
		`([A-Z]{2,10}-\d{3,5})`, // ABC-123, ABCD-1234
		`([A-Z]{2,10}\d{3,5})`,  // ABC123, ABCD1234
		`(\d{6}_\d{3})`,         // 123456_789
		`([A-Z]+\s*\d+)`,        // ABC 123, ABC123
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(title)
		if len(matches) > 1 {
			code := strings.ReplaceAll(matches[1], " ", "")
			return strings.ToUpper(code)
		}
	}

	return ""
}

// normalizeCode 标准化番号格式
func (s *JAVDbSearchService) normalizeCode(code string) string {
	if code == "" {
		return ""
	}

	// 转换为大写
	code = strings.ToUpper(strings.TrimSpace(code))

	// 移除特殊字符，保留字母数字和连字符
	var result strings.Builder
	for _, r := range code {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// parseReleaseDate 解析发布日期
func (s *JAVDbSearchService) parseReleaseDate(dateText string) time.Time {
	if dateText == "" {
		return time.Time{}
	}

	// 尝试各种日期格式
	formats := []string{
		"2006-01-02",  // 2023-12-01
		"2006/01/02",  // 2023/12/01
		"2006年01月02日", // 2023年12月01日
		"01/02/2006",  // 12/01/2023
		"02-01-2006",  // 01-12-2023
	}

	for _, format := range formats {
		if t, err := time.Parse(format, strings.TrimSpace(dateText)); err == nil {
			return t
		}
	}

	return time.Time{}
}
