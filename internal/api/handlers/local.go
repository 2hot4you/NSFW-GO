package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"nsfw-go/internal/repo"
	"nsfw-go/internal/service"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// LocalMovie 本地影片结构（用于API响应）
type LocalMovie struct {
	Title      string `json:"title"`
	Code       string `json:"code"`
	Actress    string `json:"actress"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	Modified   string `json:"modified"`
	Format     string `json:"format"`
	FanartPath string `json:"fanart_path"`
	FanartURL  string `json:"fanart_url"`
	HasFanart  bool   `json:"has_fanart"`
}

// LocalHandler 本地影片处理器
type LocalHandler struct {
	localMovieRepo   repo.LocalMovieRepository
	scannerService   *service.ScannerService
	mediaLibraryPath string
}

// NewLocalHandler 创建本地影片处理器
func NewLocalHandler(localMovieRepo repo.LocalMovieRepository, scannerService *service.ScannerService, mediaLibraryPath string) *LocalHandler {
	return &LocalHandler{
		localMovieRepo:   localMovieRepo,
		scannerService:   scannerService,
		mediaLibraryPath: mediaLibraryPath,
	}
}

// ScanLocalMovies 触发手动扫描本地影片库
func (h *LocalHandler) ScanLocalMovies(c *gin.Context) {
	err := h.scannerService.ForceRescan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: fmt.Sprintf("扫描失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "手动扫描已完成",
	})
}

// GetLocalMovies 获取本地影片列表（从数据库读取）
func (h *LocalHandler) GetLocalMovies(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	actress := c.Query("actress")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 100
	}

	offset := (page - 1) * limit

	// 从数据库获取数据
	dbMovies, total, err := h.localMovieRepo.List(offset, limit, actress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: fmt.Sprintf("获取影片列表失败: %v", err),
		})
		return
	}

	// 转换为API响应格式
	movies := make([]LocalMovie, len(dbMovies))
	for i, movie := range dbMovies {
		movies[i] = LocalMovie{
			Title:      movie.Title,
			Code:       movie.Code,
			Actress:    movie.Actress,
			Path:       movie.Path,
			Size:       movie.Size,
			Modified:   movie.Modified.Format("2006-01-02 15:04:05"),
			Format:     movie.Format,
			FanartPath: movie.FanartPath,
			FanartURL:  movie.FanartURL,
			HasFanart:  movie.HasFanart,
		}
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取成功",
		Data: gin.H{
			"items": movies,
			"count": len(movies),
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetLocalMovieStats 获取本地影片统计信息
func (h *LocalHandler) GetLocalMovieStats(c *gin.Context) {
	// 获取总数
	total, err := h.localMovieRepo.Count()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: fmt.Sprintf("获取统计信息失败: %v", err),
		})
		return
	}

	// 获取按女优统计
	actressCounts, err := h.localMovieRepo.CountByActress()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: fmt.Sprintf("获取女优统计失败: %v", err),
		})
		return
	}

	// 获取最后扫描时间
	lastScanTime, err := h.localMovieRepo.GetLastScanTime()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: fmt.Sprintf("获取扫描时间失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取成功",
		Data: gin.H{
			"total_movies":   total,
			"actress_counts": actressCounts,
			"actress_total":  len(actressCounts),
			"last_scan_time": lastScanTime.Format("2006-01-02 15:04:05"),
		},
	})
}

// ServeImage 提供图片服务
func (h *LocalHandler) ServeImage(c *gin.Context) {
	imagePath := c.Param("filepath")
	if imagePath == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "图片路径不能为空",
		})
		return
	}

	// URL解码
	decodedPath, err := url.PathUnescape(imagePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "路径解码失败",
		})
		return
	}

	// 构建完整的文件路径
	fullPath := filepath.Join(h.mediaLibraryPath, decodedPath)

	// 安全检查：确保路径在媒体库目录内
	if !strings.HasPrefix(fullPath, h.mediaLibraryPath) {
		c.JSON(http.StatusForbidden, Response{
			Code:    "ERROR",
			Message: "访问被拒绝",
		})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, Response{
			Code:    "ERROR",
			Message: "图片文件不存在",
		})
		return
	}

	// 设置适当的内容类型
	ext := strings.ToLower(filepath.Ext(fullPath))
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".webp":
		contentType = "image/webp"
	case ".bmp":
		contentType = "image/bmp"
	default:
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=3600") // 缓存1小时
	c.File(fullPath)
}
