package handlers

import (
	"net/http"
	"strconv"

	"nsfw-go/internal/model"
	"nsfw-go/internal/repo"

	"github.com/gin-gonic/gin"
)

type ActressHandler struct {
	repo *repo.ActressRepository
}

func NewActressHandler(repo *repo.ActressRepository) *ActressHandler {
	return &ActressHandler{repo: repo}
}

// CreateActress 创建女优
func (h *ActressHandler) CreateActress(c *gin.Context) {
	var req CreateActressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	actress := &model.Actress{
		Name:        req.Name,
		AvatarURL:   req.Avatar,
		Description: req.Nationality + " " + req.Birthday, // 临时存储到描述字段
	}

	if err := h.repo.Create(actress); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CREATE_FAILED",
			Message: "创建女优失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    "SUCCESS",
		Message: "女优创建成功",
		Data:    actress,
	})
}

// GetActressByID 根据ID获取女优
func (h *ActressHandler) GetActressByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "无效的女优ID",
			Error:   err.Error(),
		})
		return
	}

	actress, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "女优不存在",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取女优成功",
		Data:    actress,
	})
}

// UpdateActress 更新女优信息
func (h *ActressHandler) UpdateActress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "无效的女优ID",
			Error:   err.Error(),
		})
		return
	}

	actress, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "女优不存在",
			Error:   err.Error(),
		})
		return
	}

	var req UpdateActressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	// 更新字段
	if req.Name != nil {
		actress.Name = *req.Name
	}
	if req.Avatar != nil {
		actress.AvatarURL = *req.Avatar
	}
	// 临时将其他信息存储到描述字段
	if req.Nationality != nil || req.Birthday != nil {
		desc := actress.Description
		if req.Nationality != nil {
			desc = *req.Nationality + " "
		}
		if req.Birthday != nil {
			desc += *req.Birthday
		}
		actress.Description = desc
	}

	if err := h.repo.Update(actress); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_FAILED",
			Message: "更新女优失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "女优更新成功",
		Data:    actress,
	})
}

// DeleteActress 删除女优
func (h *ActressHandler) DeleteActress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "无效的女优ID",
			Error:   err.Error(),
		})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_FAILED",
			Message: "删除女优失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "女优删除成功",
	})
}

// ListActresses 获取女优列表
func (h *ActressHandler) ListActresses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var actresses []model.Actress
	var total int64
	var err error

	if search != "" {
		actresses, total, err = h.repo.Search(search, (page-1)*limit, limit)
	} else {
		actresses, total, err = h.repo.List((page-1)*limit, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "QUERY_FAILED",
			Message: "查询女优列表失败",
			Error:   err.Error(),
		})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取女优列表成功",
		Data: ListResponse{
			Items:      actresses,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// GetActressByName 根据姓名获取女优
func (h *ActressHandler) GetActressByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_NAME",
			Message: "女优姓名不能为空",
		})
		return
	}

	actress, err := h.repo.GetByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "女优不存在",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取女优成功",
		Data:    actress,
	})
}

// GetActressMovies 获取女优的影片列表
func (h *ActressHandler) GetActressMovies(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "无效的女优ID",
			Error:   err.Error(),
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

	movies, total, err := h.repo.GetMovies(uint(id), (page-1)*limit, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "QUERY_FAILED",
			Message: "查询女优影片失败",
			Error:   err.Error(),
		})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取女优影片成功",
		Data: ListResponse{
			Items:      movies,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}
