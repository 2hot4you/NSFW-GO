package crawler

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// RankingItem 排行榜项目
type RankingItem struct {
	Code     string `json:"code"`
	Title    string `json:"title"`
	CoverURL string `json:"cover_url"`
	Position int    `json:"position"`
}

// RankingCrawler 排行榜爬虫
type RankingCrawler struct {
	*BaseCrawler
	baseURL string
}

// NewRankingCrawler 创建排行榜爬虫
func NewRankingCrawler(config *CrawlerConfig) *RankingCrawler {
	baseCrawler := NewBaseCrawler("JAVDb-Ranking", config)

	return &RankingCrawler{
		BaseCrawler: baseCrawler,
		baseURL:     "https://javdb.com",
	}
}

// CrawlRanking 爬取排行榜
func (rc *RankingCrawler) CrawlRanking(ctx context.Context, rankType string) ([]RankingItem, error) {
	var items []RankingItem
	var crawlErr error

	// 构建排行榜URL
	rankingURL := fmt.Sprintf("%s/rankings/movies?p=%s&t=censored", rc.baseURL, rankType)

	c := rc.GetCollector().Clone()

	// 解析排行榜页面
	c.OnHTML(".movie-list .item", func(e *colly.HTMLElement) {
		item := RankingItem{}

		// 获取标题和链接
		titleEl := e.DOM.Find(".video-title")
		if titleEl.Length() > 0 {
			item.Title = rc.CleanText(titleEl.Text())

			// 从链接中提取番号
			linkEl := titleEl.Find("a")
			if linkEl.Length() > 0 {
				if href, exists := linkEl.Attr("href"); exists {
					// 从URL路径中提取番号，例如: /v/abc123 -> abc123
					parts := strings.Split(href, "/")
					if len(parts) >= 3 {
						item.Code = strings.ToUpper(parts[len(parts)-1])
					}
				}
			}
		}

		// 如果从链接提取失败，尝试从标题提取
		if item.Code == "" {
			item.Code = rc.ExtractMovieCode(item.Title)
		}

		// 获取封面图片
		imgEl := e.DOM.Find(".cover img")
		if imgEl.Length() > 0 {
			if src, exists := imgEl.Attr("src"); exists {
				if fullURL, err := rc.BuildURL(rc.baseURL, src); err == nil {
					item.CoverURL = fullURL
				}
			}
			// 尝试data-src属性（懒加载）
			if item.CoverURL == "" {
				if dataSrc, exists := imgEl.Attr("data-src"); exists {
					if fullURL, err := rc.BuildURL(rc.baseURL, dataSrc); err == nil {
						item.CoverURL = fullURL
					}
				}
			}
		}

		// 获取排名位置
		posEl := e.DOM.Find(".rank-number")
		if posEl.Length() > 0 {
			posText := rc.CleanText(posEl.Text())
			if pos, err := strconv.Atoi(posText); err == nil {
				item.Position = pos
			}
		}

		// 如果没有明确的排名元素，使用当前索引+1
		if item.Position == 0 {
			item.Position = len(items) + 1
		}

		// 验证必要字段
		if item.Code != "" && item.Title != "" {
			items = append(items, item)
		} else {
			log.Printf("[排行榜爬虫] 跳过无效项目: Code=%s, Title=%s", item.Code, item.Title)
		}
	})

	// 错误处理
	c.OnError(func(r *colly.Response, err error) {
		crawlErr = fmt.Errorf("爬取排行榜失败: %v", err)
	})

	// 访问排行榜页面
	log.Printf("[排行榜爬虫] 开始爬取 %s 排行榜: %s", rankType, rankingURL)
	err := c.Visit(rankingURL)
	if err != nil {
		return nil, fmt.Errorf("访问排行榜页面失败: %v", err)
	}

	c.Wait()

	if crawlErr != nil {
		return nil, crawlErr
	}

	log.Printf("[排行榜爬虫] %s 排行榜爬取完成，共 %d 个项目", rankType, len(items))
	return items, nil
}

// CrawlAllRankings 爬取所有类型的排行榜
func (rc *RankingCrawler) CrawlAllRankings(ctx context.Context) (map[string][]RankingItem, error) {
	rankTypes := []string{"daily", "weekly", "monthly"}
	results := make(map[string][]RankingItem)

	for _, rankType := range rankTypes {
		items, err := rc.CrawlRanking(ctx, rankType)
		if err != nil {
			log.Printf("[排行榜爬虫] 爬取 %s 排行榜失败: %v", rankType, err)
			continue
		}

		results[rankType] = items

		// 添加延时，避免请求过于频繁
		time.Sleep(2 * time.Second)
	}

	return results, nil
}

// IsHealthy 检查爬虫健康状态
func (rc *RankingCrawler) IsHealthy(ctx context.Context) bool {
	c := rc.GetCollector().Clone()

	var isHealthy bool

	// 尝试访问排行榜页面
	testURL := fmt.Sprintf("%s/rankings/movies?p=daily&t=censored", rc.baseURL)

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode == 200 {
			isHealthy = true
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("[排行榜爬虫] 健康检查失败: %v", err)
		isHealthy = false
	})

	// 设置较短的超时时间
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := c.Visit(testURL)
	if err != nil {
		log.Printf("[排行榜爬虫] 健康检查访问失败: %v", err)
		return false
	}

	c.Wait()

	log.Printf("[排行榜爬虫] 健康检查结果: %v", isHealthy)
	return isHealthy
}
