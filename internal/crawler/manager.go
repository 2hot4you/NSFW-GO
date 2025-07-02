package crawler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Manager 爬虫管理器实现
type Manager struct {
	crawlers map[string]Crawler
	config   *CrawlerConfig
	mu       sync.RWMutex
}

// NewManager 创建新的爬虫管理器
func NewManager(config *CrawlerConfig) *Manager {
	return &Manager{
		crawlers: make(map[string]Crawler),
		config:   config,
	}
}

// RegisterCrawler 注册爬虫
func (m *Manager) RegisterCrawler(name string, crawler Crawler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.crawlers[name] = crawler
	log.Printf("[爬虫管理器] 注册爬虫: %s", name)
}

// GetCrawler 获取指定爬虫
func (m *Manager) GetCrawler(name string) (Crawler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	crawler, exists := m.crawlers[name]
	return crawler, exists
}

// GetAllCrawlers 获取所有爬虫
func (m *Manager) GetAllCrawlers() map[string]Crawler {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Crawler)
	for name, crawler := range m.crawlers {
		result[name] = crawler
	}
	return result
}

// CrawlMovieByCode 使用所有可用爬虫搜索影片
func (m *Manager) CrawlMovieByCode(ctx context.Context, code string) (*MovieData, error) {
	m.mu.RLock()
	crawlers := make([]Crawler, 0, len(m.crawlers))
	for _, crawler := range m.crawlers {
		crawlers = append(crawlers, crawler)
	}
	m.mu.RUnlock()

	if len(crawlers) == 0 {
		return nil, fmt.Errorf("没有可用的爬虫")
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	// 使用channel收集结果
	type result struct {
		data    *MovieData
		crawler string
		err     error
	}

	resultChan := make(chan result, len(crawlers))

	// 并发执行所有爬虫
	for _, crawler := range crawlers {
		go func(c Crawler) {
			data, err := c.GetMovieByCode(ctx, code)
			resultChan <- result{
				data:    data,
				crawler: c.GetName(),
				err:     err,
			}
		}(crawler)
	}

	// 收集结果，返回第一个成功的结果
	var lastErr error
	for i := 0; i < len(crawlers); i++ {
		select {
		case res := <-resultChan:
			if res.err == nil && res.data != nil {
				log.Printf("[爬虫管理器] 成功获取影片 %s (来源: %s)", code, res.crawler)
				return res.data, nil
			}
			if res.err != nil {
				log.Printf("[爬虫管理器] 爬虫 %s 失败: %v", res.crawler, res.err)
				lastErr = res.err
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("爬虫任务超时")
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("所有爬虫都失败了，最后错误: %v", lastErr)
	}

	return nil, fmt.Errorf("未找到影片 %s", code)
}

// SearchMovies 搜索影片
func (m *Manager) SearchMovies(ctx context.Context, keyword string) ([]SearchResult, error) {
	m.mu.RLock()
	crawlers := make([]Crawler, 0, len(m.crawlers))
	for _, crawler := range m.crawlers {
		crawlers = append(crawlers, crawler)
	}
	m.mu.RUnlock()

	if len(crawlers) == 0 {
		return nil, fmt.Errorf("没有可用的爬虫")
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	// 使用channel收集结果
	type result struct {
		results []SearchResult
		crawler string
		err     error
	}

	resultChan := make(chan result, len(crawlers))

	// 并发执行所有爬虫
	for _, crawler := range crawlers {
		go func(c Crawler) {
			results, err := c.Search(ctx, keyword)
			resultChan <- result{
				results: results,
				crawler: c.GetName(),
				err:     err,
			}
		}(crawler)
	}

	// 收集所有结果
	var allResults []SearchResult
	var errors []error

	for i := 0; i < len(crawlers); i++ {
		select {
		case res := <-resultChan:
			if res.err == nil {
				log.Printf("[爬虫管理器] 爬虫 %s 搜索到 %d 个结果", res.crawler, len(res.results))
				allResults = append(allResults, res.results...)
			} else {
				log.Printf("[爬虫管理器] 爬虫 %s 搜索失败: %v", res.crawler, res.err)
				errors = append(errors, res.err)
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("搜索任务超时")
		}
	}

	if len(allResults) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("所有爬虫搜索都失败了")
	}

	// 去重处理
	uniqueResults := m.deduplicateSearchResults(allResults)

	log.Printf("[爬虫管理器] 搜索关键词 '%s' 共找到 %d 个唯一结果", keyword, len(uniqueResults))
	return uniqueResults, nil
}

// deduplicateSearchResults 去重搜索结果
func (m *Manager) deduplicateSearchResults(results []SearchResult) []SearchResult {
	seen := make(map[string]bool)
	var unique []SearchResult

	for _, result := range results {
		key := result.Code
		if key == "" {
			key = result.Title + result.DetailURL
		}

		if !seen[key] {
			seen[key] = true
			unique = append(unique, result)
		}
	}

	return unique
}

// CheckCrawlerHealth 检查所有爬虫的健康状态
func (m *Manager) CheckCrawlerHealth(ctx context.Context) map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthStatus := make(map[string]bool)

	for name, crawler := range m.crawlers {
		// 为每个健康检查设置较短的超时时间
		checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		healthy := crawler.IsHealthy(checkCtx)
		cancel()

		healthStatus[name] = healthy
		log.Printf("[爬虫管理器] 爬虫 %s 健康状态: %v", name, healthy)
	}

	return healthStatus
}

// GetConfig 获取配置
func (m *Manager) GetConfig() *CrawlerConfig {
	return m.config
}

// UpdateConfig 更新配置
func (m *Manager) UpdateConfig(config *CrawlerConfig) {
	m.config = config
	log.Printf("[爬虫管理器] 配置已更新")
}
