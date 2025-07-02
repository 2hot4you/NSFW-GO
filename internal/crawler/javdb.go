package crawler

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

// JAVDbCrawler JAVDb爬虫
type JAVDbCrawler struct {
	*BaseCrawler
	baseURL string
}

// NewJAVDbCrawler 创建JAVDb爬虫
func NewJAVDbCrawler(config *CrawlerConfig) *JAVDbCrawler {
	baseCrawler := NewBaseCrawler("JAVDb", config)

	return &JAVDbCrawler{
		BaseCrawler: baseCrawler,
		baseURL:     "https://javdb.com",
	}
}

// Search 搜索影片
func (jc *JAVDbCrawler) Search(ctx context.Context, keyword string) ([]SearchResult, error) {
	var results []SearchResult
	var searchErr error

	// 构建搜索URL
	searchURL := fmt.Sprintf("%s/search?q=%s&f=all", jc.baseURL, url.QueryEscape(keyword))

	c := jc.GetCollector().Clone()

	// 解析搜索结果
	c.OnHTML(".movie-list .item", func(e *colly.HTMLElement) {
		result := SearchResult{}

		// 获取标题和详情链接
		titleEl := e.DOM.Find("a")
		if titleEl.Length() > 0 {
			result.Title = jc.CleanText(titleEl.Text())
			if href, exists := titleEl.Attr("href"); exists {
				if fullURL, err := jc.BuildURL(jc.baseURL, href); err == nil {
					result.DetailURL = fullURL
				}
			}
		}

		// 获取封面图片
		imgEl := e.DOM.Find("img")
		if imgEl.Length() > 0 {
			if src, exists := imgEl.Attr("src"); exists {
				if fullURL, err := jc.BuildURL(jc.baseURL, src); err == nil {
					result.CoverURL = fullURL
				}
			}
		}

		// 提取番号
		result.Code = jc.ExtractMovieCode(result.Title)

		// 获取评分
		ratingEl := e.DOM.Find(".score")
		if ratingEl.Length() > 0 {
			result.Rating = jc.ParseRating(ratingEl.Text())
		}

		// 获取发布日期
		dateEl := e.DOM.Find(".meta")
		if dateEl.Length() > 0 {
			result.ReleaseDate = jc.ParseReleaseDate(dateEl.Text())
		}

		if result.Title != "" && result.DetailURL != "" {
			results = append(results, result)
		}
	})

	// 错误处理
	c.OnError(func(r *colly.Response, err error) {
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

	log.Printf("[JAVDb] 搜索 '%s' 找到 %d 个结果", keyword, len(results))
	return results, nil
}

// GetMovieByCode 根据番号获取影片详情
func (jc *JAVDbCrawler) GetMovieByCode(ctx context.Context, code string) (*MovieData, error) {
	// 先搜索找到详情页面URL
	searchResults, err := jc.Search(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("搜索影片失败: %v", err)
	}

	// 找到匹配的影片
	normalizedCode := jc.NormalizeMovieCode(code)
	for _, result := range searchResults {
		resultCode := jc.NormalizeMovieCode(result.Code)
		if resultCode == normalizedCode {
			return jc.GetMovieByURL(ctx, result.DetailURL)
		}
	}

	return nil, fmt.Errorf("未找到影片: %s", code)
}

// GetMovieByURL 根据URL获取影片详情
func (jc *JAVDbCrawler) GetMovieByURL(ctx context.Context, movieURL string) (*MovieData, error) {
	var movieData *MovieData
	var crawlErr error

	c := jc.GetCollector().Clone()

	// 解析影片详情页面
	c.OnHTML(".video-detail", func(e *colly.HTMLElement) {
		movie := &MovieData{
			Actresses: []ActressData{},
			Tags:      []TagData{},
		}

		// 获取标题
		titleEl := e.DOM.Find("h2 strong")
		if titleEl.Length() > 0 {
			movie.Title = jc.CleanText(titleEl.Text())
		}

		// 提取番号
		movie.Code = jc.ExtractMovieCode(movie.Title)
		if movie.Code == "" {
			// 尝试从页面其他位置提取
			codeEl := e.DOM.Find(".video-meta-panel .panel-block").First()
			if codeEl.Length() > 0 {
				movie.Code = jc.ExtractMovieCode(codeEl.Text())
			}
		}

		// 标准化番号
		movie.Code = jc.NormalizeMovieCode(movie.Code)

		// 获取封面图片
		posterEl := e.DOM.Find(".video-cover img")
		if posterEl.Length() > 0 {
			if src, exists := posterEl.Attr("src"); exists {
				if fullURL, err := jc.BuildURL(jc.baseURL, src); err == nil {
					movie.CoverURL = fullURL
				}
			}
		}

		// 解析影片信息面板
		e.DOM.Find(".video-meta-panel .panel-block").Each(func(i int, s *goquery.Selection) {
			blockText := jc.CleanText(s.Text())

			// 发布日期
			if strings.Contains(blockText, "日期") {
				parts := strings.Split(blockText, ":")
				if len(parts) > 1 {
					dateText := strings.TrimSpace(parts[1])
					movie.ReleaseDate = jc.ParseReleaseDate(dateText)
				}
			}

			// 时长
			if strings.Contains(blockText, "时长") {
				parts := strings.Split(blockText, ":")
				if len(parts) > 1 {
					durationText := strings.TrimSpace(parts[1])
					movie.Duration = jc.ParseDuration(durationText)
				}
			}

			// 制作商
			if strings.Contains(blockText, "片商") {
				parts := strings.Split(blockText, ":")
				if len(parts) > 1 {
					studioText := strings.TrimSpace(parts[1])
					if studioText != "" {
						movie.Studio = &StudioData{
							Name: studioText,
						}
					}
				}
			}

			// 系列
			if strings.Contains(blockText, "系列") {
				parts := strings.Split(blockText, ":")
				if len(parts) > 1 {
					seriesText := strings.TrimSpace(parts[1])
					if seriesText != "" {
						movie.Series = &SeriesData{
							Name: seriesText,
						}
					}
				}
			}
		})

		// 获取评分
		ratingEl := e.DOM.Find(".video-meta .score")
		if ratingEl.Length() > 0 {
			movie.Rating = jc.ParseRating(ratingEl.Text())
		}

		// 获取女优信息
		e.DOM.Find(".performers .performer").Each(func(i int, s *goquery.Selection) {
			actress := ActressData{}

			// 女优名字
			nameEl := s.Find("a")
			if nameEl.Length() > 0 {
				actress.Name = jc.CleanText(nameEl.Text())
			}

			// 女优头像
			imgEl := s.Find("img")
			if imgEl.Length() > 0 {
				if src, exists := imgEl.Attr("src"); exists {
					if fullURL, err := jc.BuildURL(jc.baseURL, src); err == nil {
						actress.AvatarURL = fullURL
					}
				}
			}

			if actress.Name != "" {
				movie.Actresses = append(movie.Actresses, actress)
			}
		})

		// 获取标签
		e.DOM.Find(".tags .tag").Each(func(i int, s *goquery.Selection) {
			tagText := jc.CleanText(s.Text())
			if tagText != "" {
				tag := TagData{
					Name:     tagText,
					Category: "genre", // 默认分类
				}
				movie.Tags = append(movie.Tags, tag)
			}
		})

		// 获取简介
		descEl := e.DOM.Find(".video-detail .content p")
		if descEl.Length() > 0 {
			movie.Description = jc.CleanText(descEl.Text())
		}

		movieData = movie
	})

	// 错误处理
	c.OnError(func(r *colly.Response, err error) {
		crawlErr = fmt.Errorf("获取影片详情失败: %v", err)
	})

	// 访问影片详情页面
	err := c.Visit(movieURL)
	if err != nil {
		return nil, fmt.Errorf("访问影片页面失败: %v", err)
	}

	c.Wait()

	if crawlErr != nil {
		return nil, crawlErr
	}

	if movieData == nil {
		return nil, fmt.Errorf("未能解析影片数据")
	}

	log.Printf("[JAVDb] 成功获取影片信息: %s - %s", movieData.Code, movieData.Title)
	return movieData, nil
}

// GetActressInfo 获取女优信息
func (jc *JAVDbCrawler) GetActressInfo(ctx context.Context, actressName string) (*ActressData, error) {
	var actressData *ActressData
	var crawlErr error

	// 构建女优搜索URL
	searchURL := fmt.Sprintf("%s/search?q=%s&f=actor", jc.baseURL, url.QueryEscape(actressName))

	c := jc.GetCollector().Clone()

	// 查找女优页面
	c.OnHTML(".actor-list .item", func(e *colly.HTMLElement) {
		if actressData != nil {
			return // 已找到，跳过
		}

		nameEl := e.DOM.Find(".actor-name")
		if nameEl.Length() > 0 {
			name := jc.CleanText(nameEl.Text())
			if strings.Contains(strings.ToLower(name), strings.ToLower(actressName)) {
				actress := &ActressData{
					Name: name,
				}

				// 获取头像
				imgEl := e.DOM.Find("img")
				if imgEl.Length() > 0 {
					if src, exists := imgEl.Attr("src"); exists {
						if fullURL, err := jc.BuildURL(jc.baseURL, src); err == nil {
							actress.AvatarURL = fullURL
						}
					}
				}

				actressData = actress
			}
		}
	})

	// 错误处理
	c.OnError(func(r *colly.Response, err error) {
		crawlErr = fmt.Errorf("获取女优信息失败: %v", err)
	})

	// 执行搜索
	err := c.Visit(searchURL)
	if err != nil {
		return nil, fmt.Errorf("访问女优搜索页面失败: %v", err)
	}

	c.Wait()

	if crawlErr != nil {
		return nil, crawlErr
	}

	if actressData == nil {
		return nil, fmt.Errorf("未找到女优: %s", actressName)
	}

	log.Printf("[JAVDb] 成功获取女优信息: %s", actressData.Name)
	return actressData, nil
}

// IsHealthy 检查爬虫健康状态
func (jc *JAVDbCrawler) IsHealthy(ctx context.Context) bool {
	c := jc.GetCollector().Clone()

	var isHealthy bool

	// 尝试访问首页
	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode == 200 {
			isHealthy = true
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("[JAVDb] 健康检查失败: %v", err)
		isHealthy = false
	})

	// 设置较短的超时时间
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := c.Visit(jc.baseURL)
	if err != nil {
		log.Printf("[JAVDb] 健康检查访问失败: %v", err)
		return false
	}

	c.Wait()

	log.Printf("[JAVDb] 健康检查结果: %v", isHealthy)
	return isHealthy
}
