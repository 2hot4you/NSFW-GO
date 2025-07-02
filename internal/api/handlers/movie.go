package handlers

import (
	"net/http"
	"strconv"

	"nsfw-go/internal/model"
	"nsfw-go/internal/repo"

	"github.com/gin-gonic/gin"
)

// MovieHandler 影片处理器
type MovieHandler struct {
	movieRepo repo.MovieRepository
}

// NewMovieHandler 创建影片处理器
func NewMovieHandler(movieRepo repo.MovieRepository) *MovieHandler {
	return &MovieHandler{
		movieRepo: movieRepo,
	}
}

// ListMovies 获取影片列表
// @Summary 获取影片列表
// @Description 获取影片列表，支持分页和筛选
// @Tags movies
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param studio_id query int false "制作商ID"
// @Param quality query string false "画质"
// @Param has_subtitle query bool false "是否有字幕"
// @Param is_downloaded query bool false "是否已下载"
// @Param sort_by query string false "排序字段" Enums(created_at,rating,watch_count,release_date)
// @Param sort_order query string false "排序方向" Enums(asc,desc)
// @Success 200 {object} Response{data=ListResponse{items=[]model.Movie}}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/movies [get]
func (h *MovieHandler) ListMovies(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// 解析筛选条件
	filters := repo.MovieFilter{
		Quality:   c.Query("quality"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	if studioID := c.Query("studio_id"); studioID != "" {
		if id, err := strconv.ParseUint(studioID, 10, 32); err == nil {
			uid := uint(id)
			filters.StudioID = &uid
		}
	}

	if hasSubtitle := c.Query("has_subtitle"); hasSubtitle != "" {
		if val, err := strconv.ParseBool(hasSubtitle); err == nil {
			filters.HasSubtitle = &val
		}
	}

	if isDownloaded := c.Query("is_downloaded"); isDownloaded != "" {
		if val, err := strconv.ParseBool(isDownloaded); err == nil {
			filters.IsDownloaded = &val
		}
	}

	// 查询数据
	movies, total, err := h.movieRepo.List(offset, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "获取影片列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取成功",
		Data: ListResponse{
			Items:      movies,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetMovieByID 根据ID获取影片详情
// @Summary 获取影片详情
// @Description 根据影片ID获取详细信息
// @Tags movies
// @Accept json
// @Produce json
// @Param id path int true "影片ID"
// @Success 200 {object} Response{data=model.Movie}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/movies/{id} [get]
func (h *MovieHandler) GetMovieByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PARAMETER",
			Message: "无效的影片ID",
			Error:   err.Error(),
		})
		return
	}

	movie, err := h.movieRepo.GetByID(uint(id))
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "MOVIE_NOT_FOUND",
				Message: "影片不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "获取影片详情失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取成功",
		Data:    movie,
	})
}

// GetMovieByCode 根据番号获取影片详情
// @Summary 根据番号获取影片详情
// @Description 根据影片番号获取详细信息
// @Tags movies
// @Accept json
// @Produce json
// @Param code path string true "影片番号"
// @Success 200 {object} Response{data=model.Movie}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/movies/code/{code} [get]
func (h *MovieHandler) GetMovieByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PARAMETER",
			Message: "番号不能为空",
		})
		return
	}

	movie, err := h.movieRepo.GetByCode(code)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "MOVIE_NOT_FOUND",
				Message: "影片不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "获取影片详情失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取成功",
		Data:    movie,
	})
}

// SearchMovies 搜索影片
// @Summary 搜索影片
// @Description 根据关键词搜索影片
// @Tags movies
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} Response{data=ListResponse{items=[]model.Movie}}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/movies/search [get]
func (h *MovieHandler) SearchMovies(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PARAMETER",
			Message: "搜索关键词不能为空",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	movies, total, err := h.movieRepo.Search(query, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "搜索影片失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "搜索成功",
		Data: ListResponse{
			Items:      movies,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateMovie 创建影片
// @Summary 创建影片
// @Description 创建新的影片记录
// @Tags movies
// @Accept json
// @Produce json
// @Param movie body CreateMovieRequest true "影片信息"
// @Success 201 {object} Response{data=model.Movie}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/movies [post]
func (h *MovieHandler) CreateMovie(c *gin.Context) {
	var req CreateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	// 检查番号是否已存在
	if existingMovie, _ := h.movieRepo.GetByCode(req.Code); existingMovie != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MOVIE_EXISTS",
			Message: "该番号的影片已存在",
		})
		return
	}

	movie := &model.Movie{
		Code:        req.Code,
		Title:       req.Title,
		Description: req.Description,
		StudioID:    req.StudioID,
		SeriesID:    req.SeriesID,
		CoverURL:    req.CoverURL,
		Rating:      req.Rating,
	}

	if err := h.movieRepo.Create(movie); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "创建影片失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    "SUCCESS",
		Message: "创建成功",
		Data:    movie,
	})
}

// GetRecentMovies 获取最近添加的影片
// @Summary 获取最近添加的影片
// @Description 获取最近添加的影片列表
// @Tags movies
// @Accept json
// @Produce json
// @Param limit query int false "数量限制" default(10)
// @Success 200 {object} Response{data=[]model.Movie}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/movies/recent [get]
func (h *MovieHandler) GetRecentMovies(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	movies, err := h.movieRepo.GetRecentlyAdded(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "获取最近影片失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取成功",
		Data:    movies,
	})
}

// GetPopularMovies 获取热门影片
// @Summary 获取热门影片
// @Description 获取热门影片列表
// @Tags movies
// @Accept json
// @Produce json
// @Param limit query int false "数量限制" default(10)
// @Success 200 {object} Response{data=[]model.Movie}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/movies/popular [get]
func (h *MovieHandler) GetPopularMovies(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	movies, err := h.movieRepo.GetPopular(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: "获取热门影片失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取成功",
		Data:    movies,
	})
} 