package handlers

import (
	"net/http"
	"nsfw-go/internal/repo"
	"strconv"
	"strings"

	"nsfw-go/internal/model"

	"github.com/gin-gonic/gin"
)

// SearchHandler 搜索处理器
type SearchHandler struct {
	localMovieRepo repo.LocalMovieRepository
	rankingRepo    repo.RankingRepository
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler(localMovieRepo repo.LocalMovieRepository, rankingRepo repo.RankingRepository) *SearchHandler {
	return &SearchHandler{
		localMovieRepo: localMovieRepo,
		rankingRepo:    rankingRepo,
	}
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query   string `form:"q" json:"q"`
	Type    string `form:"type" json:"type"` // local, ranking, all
	Page    int    `form:"page" json:"page"`
	Limit   int    `form:"limit" json:"limit"`
	Actress string `form:"actress" json:"actress"`
	Code    string `form:"code" json:"code"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	LocalMovies []LocalMovieResult `json:"local_movies"`
	Rankings    []RankingResult    `json:"rankings"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	Limit       int                `json:"limit"`
	Query       string             `json:"query"`
}

// LocalMovieResult 本地影片搜索结果
type LocalMovieResult struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Code      string `json:"code"`
	Actress   string `json:"actress"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Format    string `json:"format"`
	HasFanart bool   `json:"has_fanart"`
	FanartURL string `json:"fanart_url"`
}

// RankingResult 排行榜搜索结果
type RankingResult struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Title       string `json:"title"`
	CoverURL    string `json:"cover_url"`
	RankType    string `json:"rank_type"`
	Position    int    `json:"position"`
	LocalExists bool   `json:"local_exists"`
}

// Search 执行搜索
func (h *SearchHandler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "搜索参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Type == "" {
		req.Type = "all"
	}

	response := SearchResponse{
		LocalMovies: []LocalMovieResult{},
		Rankings:    []RankingResult{},
		Page:        req.Page,
		Limit:       req.Limit,
		Query:       req.Query,
	}

	// 如果没有搜索词，返回空结果
	if req.Query == "" && req.Code == "" && req.Actress == "" {
		c.JSON(http.StatusOK, Response{
			Code:    "SUCCESS",
			Message: "搜索成功",
			Data:    response,
		})
		return
	}

	// 搜索本地影片
	if req.Type == "local" || req.Type == "all" {
		localResults, err := h.searchLocalMovies(req)
		if err == nil {
			response.LocalMovies = localResults
		}
	}

	// 搜索排行榜
	if req.Type == "ranking" || req.Type == "all" {
		rankingResults, err := h.searchRankings(req)
		if err == nil {
			response.Rankings = rankingResults
		}
	}

	response.Total = len(response.LocalMovies) + len(response.Rankings)

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "搜索成功",
		Data:    response,
	})
}

// searchLocalMovies 搜索本地影片
func (h *SearchHandler) searchLocalMovies(req SearchRequest) ([]LocalMovieResult, error) {
	offset := (req.Page - 1) * req.Limit
	var movies []*model.LocalMovie
	var err error

	// 根据搜索条件选择不同的查询方法
	if req.Code != "" {
		// 按番号搜索
		movie, err := h.localMovieRepo.SearchByCode(req.Code)
		if err != nil {
			return []LocalMovieResult{}, nil // 没找到不算错误
		}
		movies = []*model.LocalMovie{movie}
	} else if req.Actress != "" {
		// 按女优搜索
		movies, _, err = h.localMovieRepo.SearchByActress(req.Actress, offset, req.Limit)
	} else if req.Query != "" {
		// 综合搜索
		movies, _, err = h.localMovieRepo.Search(req.Query, offset, req.Limit)
	} else {
		return []LocalMovieResult{}, nil
	}

	if err != nil {
		return []LocalMovieResult{}, err
	}

	// 转换为返回格式
	results := make([]LocalMovieResult, 0, len(movies))
	for _, movie := range movies {
		result := LocalMovieResult{
			ID:        movie.ID,
			Title:     movie.Title,
			Code:      movie.Code,
			Actress:   movie.Actress,
			Path:      movie.Path,
			Size:      movie.Size,
			Format:    movie.Format,
			HasFanart: movie.HasFanart,
			FanartURL: movie.FanartURL,
		}
		results = append(results, result)
	}

	return results, nil
}

// searchRankings 搜索排行榜
func (h *SearchHandler) searchRankings(req SearchRequest) ([]RankingResult, error) {
	offset := (req.Page - 1) * req.Limit
	var rankings []*model.Ranking
	var err error

	// 根据搜索条件选择不同的查询方法
	if req.Code != "" {
		// 按番号搜索
		rankings, _, err = h.rankingRepo.SearchByCode(req.Code, offset, req.Limit)
	} else if req.Query != "" {
		// 综合搜索
		rankings, _, err = h.rankingRepo.Search(req.Query, offset, req.Limit)
	} else {
		return []RankingResult{}, nil
	}

	if err != nil {
		return []RankingResult{}, err
	}

	// 转换为返回格式
	results := make([]RankingResult, 0, len(rankings))
	for _, ranking := range rankings {
		result := RankingResult{
			ID:          ranking.ID,
			Code:        ranking.Code,
			Title:       ranking.Title,
			CoverURL:    ranking.CoverURL,
			RankType:    ranking.RankType,
			Position:    ranking.Position,
			LocalExists: ranking.LocalExists,
		}
		results = append(results, result)
	}

	return results, nil
}

// GetSuggestions 获取搜索建议
func (h *SearchHandler) GetSuggestions(c *gin.Context) {
	query := c.Query("q")
	limit := 10

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	suggestions := []string{}

	// 如果查询为空，返回热门搜索建议
	if query == "" {
		suggestions = []string{
			"河北彩花", "新ありな", "小湊よつ葉", "相沢みなみ", "浅野こころ",
			"水卜さくら", "田中レモン", "菜月アンナ", "岬ななみ", "青空ひかり",
		}
	} else {
		// 基于输入提供建议（这里可以从数据库获取匹配的女优名或番号）
		suggestions = h.generateSuggestions(query, limit)
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取建议成功",
		Data: gin.H{
			"suggestions": suggestions,
			"query":       query,
		},
	})
}

// generateSuggestions 生成搜索建议
func (h *SearchHandler) generateSuggestions(query string, limit int) []string {
	suggestions := []string{}
	suggestionsSet := make(map[string]bool) // 用于去重

	query = strings.ToLower(query)

	// 1. 从本地影片中查找番号匹配
	if localMovies, _, err := h.localMovieRepo.Search(query, 0, limit); err == nil {
		for _, movie := range localMovies {
			if len(suggestions) >= limit {
				break
			}
			// 添加番号建议
			if movie.Code != "" && strings.Contains(strings.ToLower(movie.Code), query) {
				if !suggestionsSet[movie.Code] {
					suggestions = append(suggestions, movie.Code)
					suggestionsSet[movie.Code] = true
				}
			}
			// 添加女优建议
			if movie.Actress != "" && strings.Contains(strings.ToLower(movie.Actress), query) {
				if !suggestionsSet[movie.Actress] {
					suggestions = append(suggestions, movie.Actress)
					suggestionsSet[movie.Actress] = true
				}
			}
		}
	}

	// 2. 从排行榜中查找匹配
	if len(suggestions) < limit {
		if rankings, _, err := h.rankingRepo.Search(query, 0, limit-len(suggestions)); err == nil {
			for _, ranking := range rankings {
				if len(suggestions) >= limit {
					break
				}
				// 添加番号建议
				if ranking.Code != "" && strings.Contains(strings.ToLower(ranking.Code), query) {
					if !suggestionsSet[ranking.Code] {
						suggestions = append(suggestions, ranking.Code)
						suggestionsSet[ranking.Code] = true
					}
				}
			}
		}
	}

	// 3. 如果还没有足够的建议，尝试按番号精确搜索
	if len(suggestions) < limit {
		// 尝试按番号搜索本地影片
		if movie, err := h.localMovieRepo.SearchByCode(query); err == nil && movie != nil {
			if !suggestionsSet[movie.Code] {
				suggestions = append(suggestions, movie.Code)
				suggestionsSet[movie.Code] = true
			}
			if movie.Actress != "" && !suggestionsSet[movie.Actress] {
				suggestions = append(suggestions, movie.Actress)
				suggestionsSet[movie.Actress] = true
			}
		}
	}

	// 4. 如果仍然没有建议，提供一些热门搜索
	if len(suggestions) == 0 {
		fallbackSuggestions := []string{
			"河北彩花", "新ありな", "小湊よつ葉", "相沢みなみ", "浅野こころ",
			"水卜さくら", "田中レモン", "菜月アンナ", "岬ななみ", "青空ひかり",
		}

		for _, actress := range fallbackSuggestions {
			if len(suggestions) >= limit {
				break
			}
			if strings.Contains(strings.ToLower(actress), query) {
				suggestions = append(suggestions, actress)
			}
		}
	}

	return suggestions
}
