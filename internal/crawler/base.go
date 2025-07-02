package crawler

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

// BaseCrawler 基础爬虫类
type BaseCrawler struct {
	name      string
	config    *CrawlerConfig
	collector *colly.Collector
	client    *http.Client
}

// NewBaseCrawler 创建基础爬虫
func NewBaseCrawler(name string, config *CrawlerConfig) *BaseCrawler {
	// 设置默认User-Agent
	defaultUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	userAgent := defaultUA
	if len(config.UserAgents) > 0 {
		userAgent = config.UserAgents[0]
	}

	// 创建Colly收集器
	c := colly.NewCollector(
		colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent(userAgent),
	)

	// 配置限制
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.ConcurrentMax,
		Delay:       config.RequestDelay,
	})

	// 设置超时
	c.SetRequestTimeout(config.Timeout)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: config.Timeout,
	}

	crawler := &BaseCrawler{
		name:      name,
		config:    config,
		collector: c,
		client:    client,
	}

	// 设置回调函数
	crawler.setupCallbacks()

	return crawler
}

// GetName 返回爬虫名称
func (bc *BaseCrawler) GetName() string {
	return bc.name
}

// setupCallbacks 设置回调函数
func (bc *BaseCrawler) setupCallbacks() {
	// 请求前回调
	bc.collector.OnRequest(func(r *colly.Request) {
		// 随机选择User-Agent
		if len(bc.config.UserAgents) > 1 {
			userAgent := bc.config.UserAgents[rand.Intn(len(bc.config.UserAgents))]
			r.Headers.Set("User-Agent", userAgent)
		}

		// 设置通用请求头
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")

		log.Printf("[%s] 请求: %s", bc.name, r.URL.String())
	})

	// 响应回调
	bc.collector.OnResponse(func(r *colly.Response) {
		log.Printf("[%s] 响应: %s [%d]", bc.name, r.Request.URL.String(), r.StatusCode)
	})

	// 错误回调
	bc.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("[%s] 错误: %s - %v", bc.name, r.Request.URL.String(), err)
	})
}

// Visit 访问URL
func (bc *BaseCrawler) Visit(ctx context.Context, url string) error {
	// 创建带取消的上下文
	_, cancel := context.WithCancel(ctx)
	defer cancel()

	return bc.collector.Visit(url)
}

// VisitWithRetry 带重试的访问
func (bc *BaseCrawler) VisitWithRetry(ctx context.Context, url string, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[%s] 重试访问 %s (第%d次)", bc.name, url, attempt)
			// 等待一段时间再重试
			select {
			case <-time.After(time.Duration(attempt) * bc.config.RequestDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := bc.Visit(ctx, url)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("访问失败，已重试%d次: %v", maxRetries, lastErr)
}

// IsHealthy 健康检查
func (bc *BaseCrawler) IsHealthy(ctx context.Context) bool {
	// 默认实现，子类可以重写
	return true
}

// ExtractMovieCode 从字符串中提取影片番号
func (bc *BaseCrawler) ExtractMovieCode(text string) string {
	// 常见的番号格式匹配
	patterns := []string{
		`([A-Z]{2,5}-?\d{3,4})`, // SSIS-001, ABP999, MIDE-123等
		`([A-Z]{1,3}\d{2,4})`,   // T28-123, S2M-123等
		`(\d{6}[-_]\d{3})`,      // 数字型番号
		`([A-Z]+\d+[A-Z]*)`,     // 其他变体
	}

	text = strings.ToUpper(strings.TrimSpace(text))

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// NormalizeMovieCode 标准化影片番号
func (bc *BaseCrawler) NormalizeMovieCode(code string) string {
	if code == "" {
		return ""
	}

	// 转换为大写并去除空格
	code = strings.ToUpper(strings.TrimSpace(code))

	// 移除多余的分隔符
	code = strings.ReplaceAll(code, "--", "-")
	code = strings.ReplaceAll(code, "__", "_")

	// 标准化常见格式
	re := regexp.MustCompile(`([A-Z]+)[-_]?(\d+)`)
	if matches := re.FindStringSubmatch(code); len(matches) == 3 {
		prefix := matches[1]
		number := matches[2]

		// 补零到3位数
		for len(number) < 3 {
			number = "0" + number
		}

		return prefix + "-" + number
	}

	return code
}

// ParseRating 解析评分
func (bc *BaseCrawler) ParseRating(ratingText string) float32 {
	if ratingText == "" {
		return 0.0
	}

	// 移除所有非数字和小数点的字符
	re := regexp.MustCompile(`[^\d.]`)
	cleanText := re.ReplaceAllString(ratingText, "")

	// 尝试解析浮点数
	var rating float64
	if _, err := fmt.Sscanf(cleanText, "%f", &rating); err == nil {
		// 确保评分在合理范围内
		if rating > 10.0 {
			rating = rating / 10.0 // 可能是100分制，转换为10分制
		}
		if rating > 10.0 {
			rating = 10.0
		}
		if rating < 0.0 {
			rating = 0.0
		}
		return float32(rating)
	}

	return 0.0
}

// ParseDuration 解析时长
func (bc *BaseCrawler) ParseDuration(durationText string) int {
	if durationText == "" {
		return 0
	}

	// 匹配各种时长格式
	patterns := []string{
		`(\d+)\s*分`,    // 120分
		`(\d+)\s*min`,  // 120min
		`(\d+)\s*mins`, // 120mins
		`(\d+):\d+`,    // 2:00 (小时:分钟)
		`(\d+)`,        // 纯数字
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(durationText); len(matches) > 1 {
			var duration int
			if _, err := fmt.Sscanf(matches[1], "%d", &duration); err == nil {
				// 如果是小时:分钟格式，需要特殊处理
				if strings.Contains(pattern, ":") {
					// 这里应该是小时数，转换为分钟
					duration = duration * 60
					// 还需要加上分钟部分
					timeRe := regexp.MustCompile(`\d+:(\d+)`)
					if timeMatches := timeRe.FindStringSubmatch(durationText); len(timeMatches) > 1 {
						var minutes int
						if _, err := fmt.Sscanf(timeMatches[1], "%d", &minutes); err == nil {
							duration += minutes
						}
					}
				}
				return duration
			}
		}
	}

	return 0
}

// ParseReleaseDate 解析发布日期
func (bc *BaseCrawler) ParseReleaseDate(dateText string) time.Time {
	if dateText == "" {
		return time.Time{}
	}

	// 尝试各种日期格式
	formats := []string{
		"2006-01-02",      // 2023-12-01
		"2006/01/02",      // 2023/12/01
		"2006年01月02日",     // 2023年12月01日
		"01/02/2006",      // 12/01/2023
		"02-01-2006",      // 01-12-2023
		"Jan 2, 2006",     // Dec 1, 2023
		"January 2, 2006", // December 1, 2023
	}

	for _, format := range formats {
		if t, err := time.Parse(format, strings.TrimSpace(dateText)); err == nil {
			return t
		}
	}

	return time.Time{}
}

// CleanText 清理文本
func (bc *BaseCrawler) CleanText(text string) string {
	// 移除多余的空白字符
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// 移除HTML实体
	replacements := map[string]string{
		"&nbsp;": " ",
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
	}

	for entity, replacement := range replacements {
		text = strings.ReplaceAll(text, entity, replacement)
	}

	return text
}

// BuildURL 构建完整URL
func (bc *BaseCrawler) BuildURL(baseURL, path string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	relative, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(relative).String(), nil
}

// GetCollector 获取Colly收集器
func (bc *BaseCrawler) GetCollector() *colly.Collector {
	return bc.collector
}

// GetClient 获取HTTP客户端
func (bc *BaseCrawler) GetClient() *http.Client {
	return bc.client
}
