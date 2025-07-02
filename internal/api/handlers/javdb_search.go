package handlers

import (
	"context"
	"net/http"
	"nsfw-go/internal/service"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// JAVDbSearchHandler JAVDb搜索处理器
type JAVDbSearchHandler struct {
	searchService *service.JAVDbSearchService
}

// NewJAVDbSearchHandler 创建JAVDb搜索处理器
func NewJAVDbSearchHandler(searchService *service.JAVDbSearchService) *JAVDbSearchHandler {
	return &JAVDbSearchHandler{
		searchService: searchService,
	}
}

// JAVDbSearchRequest 搜索请求
type JAVDbSearchRequest struct {
	Query      string `form:"q" json:"q" binding:"required"`
	SearchType string `form:"type" json:"type"` // movie, actress
}

// JAVDbMovieSearchResponse 影片搜索响应
type JAVDbMovieSearchResponse struct {
	Code        string  `json:"code"`
	Title       string  `json:"title"`
	CoverURL    string  `json:"cover_url"`
	Rating      float32 `json:"rating"`
	ReleaseDate string  `json:"release_date"`
	DetailURL   string  `json:"detail_url"`
}

// JAVDbActressSearchResponse 演员搜索响应
type JAVDbActressSearchResponse struct {
	Name       string                      `json:"name"`
	AvatarURL  string                      `json:"avatar_url"`
	DetailURL  string                      `json:"detail_url"`
	MovieCount int                         `json:"movie_count"`
	Movies     []JAVDbActressMovieResponse `json:"movies,omitempty"`
}

// JAVDbActressMovieResponse 演员作品响应
type JAVDbActressMovieResponse struct {
	Code        string  `json:"code"`
	Title       string  `json:"title"`
	CoverURL    string  `json:"cover_url"`
	ReleaseDate string  `json:"release_date"`
	Rating      float32 `json:"rating"`
	DetailURL   string  `json:"detail_url"`
}

// SearchJAVDb 执行JAVDb搜索
func (h *JAVDbSearchHandler) SearchJAVDb(c *gin.Context) {
	var req JAVDbSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "搜索参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 清理查询词
	query := strings.TrimSpace(req.Query)
	if query == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "搜索关键词不能为空",
		})
		return
	}

	// 设置默认搜索类型
	if req.SearchType == "" {
		// 根据查询词判断搜索类型
		if h.isLikelyMovieCode(query) {
			req.SearchType = "movie"
		} else {
			req.SearchType = "actress"
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch req.SearchType {
	case "movie":
		h.searchMovie(c, ctx, query)
	case "actress":
		h.searchActress(c, ctx, query)
	default:
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "无效的搜索类型，支持: movie, actress",
		})
	}
}

// searchMovie 搜索影片
func (h *JAVDbSearchHandler) searchMovie(c *gin.Context, ctx context.Context, code string) {
	result, err := h.searchService.SearchMovieByCode(ctx, code)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    "NOT_FOUND",
			Message: err.Error(),
		})
		return
	}

	response := JAVDbMovieSearchResponse{
		Code:        result.Code,
		Title:       result.Title,
		CoverURL:    result.CoverURL,
		Rating:      result.Rating,
		DetailURL:   result.DetailURL,
		ReleaseDate: "",
	}

	if !result.ReleaseDate.IsZero() {
		response.ReleaseDate = result.ReleaseDate.Format("2006-01-02")
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "搜索成功",
		Data:    response,
	})
}

// searchActress 搜索演员
func (h *JAVDbSearchHandler) searchActress(c *gin.Context, ctx context.Context, actressName string) {
	// 首先搜索演员信息
	actressResult, err := h.searchService.SearchActressByName(ctx, actressName)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    "NOT_FOUND",
			Message: err.Error(),
		})
		return
	}

	response := JAVDbActressSearchResponse{
		Name:       actressResult.Name,
		AvatarURL:  actressResult.AvatarURL,
		DetailURL:  actressResult.DetailURL,
		MovieCount: actressResult.MovieCount,
	}

	// 如果有演员详情页面URL，获取演员的作品列表
	if actressResult.DetailURL != "" {
		// 使用新的context，设置较短的超时时间
		moviesCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if movies, err := h.searchService.GetActressMovies(moviesCtx, actressResult.DetailURL); err == nil {
			response.Movies = make([]JAVDbActressMovieResponse, len(movies))
			for i, movie := range movies {
				movieResponse := JAVDbActressMovieResponse{
					Code:      movie.Code,
					Title:     movie.Title,
					CoverURL:  movie.CoverURL,
					Rating:    movie.Rating,
					DetailURL: movie.DetailURL,
				}
				if !movie.ReleaseDate.IsZero() {
					movieResponse.ReleaseDate = movie.ReleaseDate.Format("2006-01-02")
				}
				response.Movies[i] = movieResponse
			}
			response.MovieCount = len(movies)
		}
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "搜索成功",
		Data:    response,
	})
}

// isLikelyMovieCode 判断是否像影片番号
func (h *JAVDbSearchHandler) isLikelyMovieCode(query string) bool {
	query = strings.ToUpper(strings.TrimSpace(query))

	// 常见番号模式
	patterns := []string{
		`^[A-Z]{2,10}-\d{3,5}$`, // ABC-123, ABCD-1234
		`^[A-Z]{2,10}\d{3,5}$`,  // ABC123, ABCD1234
		`^\d{6}_\d{3}$`,         // 123456_789
		`^[A-Z]+\s*\d+$`,        // ABC 123, ABC123
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, query); matched {
			return true
		}
	}

	return false
}
