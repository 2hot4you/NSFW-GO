package handlers

import (
	"net/http"
	"nsfw-go/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

// TorrentHandler 种子下载处理器
type TorrentHandler struct {
	torrentService *service.TorrentService
}

// NewTorrentHandler 创建种子下载处理器
func NewTorrentHandler(torrentService *service.TorrentService) *TorrentHandler {
	return &TorrentHandler{
		torrentService: torrentService,
	}
}

// SearchTorrents 搜索种子
func (h *TorrentHandler) SearchTorrents(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "搜索关键词不能为空",
		})
		return
	}

	results, err := h.torrentService.SearchTorrents(keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "搜索成功",
		Data: map[string]interface{}{
			"keyword": keyword,
			"count":   len(results),
			"results": results,
		},
	})
}

// SearchTorrentsForCode 为特定番号搜索种子（检查本地是否存在）
func (h *TorrentHandler) SearchTorrentsForCode(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "番号不能为空",
		})
		return
	}

	results, err := h.torrentService.SearchTorrentsForCode(code)
	if err != nil {
		// 如果是已存在的错误，返回特殊状态
		if err.Error() == "番号 "+code+" 已存在于本地影视库" {
			c.JSON(http.StatusConflict, Response{
				Code:    "EXISTS",
				Message: err.Error(),
				Data: map[string]interface{}{
					"code":        code,
					"exists":      true,
					"canDownload": false,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "搜索成功",
		Data: map[string]interface{}{
			"code":        code,
			"exists":      false,
			"canDownload": true,
			"count":       len(results),
			"results":     results,
		},
	})
}

// DownloadTorrent 下载种子
func (h *TorrentHandler) DownloadTorrent(c *gin.Context) {
	var request struct {
		MagnetURI   string `json:"magnet_uri" form:"magnet_uri"`
		DownloadURI string `json:"download_uri" form:"download_uri"` // 新增：支持HTTP下载链接
		Link        string `json:"link" form:"link"`                 // 新增：兼容link字段
		Code        string `json:"code" form:"code"`
		Title       string `json:"title" form:"title"`
		Size        int64  `json:"size" form:"size"`
		Tracker     string `json:"tracker" form:"tracker"`
	}

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 确定下载链接 - 优先级：magnet_uri > download_uri > link
	var downloadURI string
	if request.MagnetURI != "" {
		downloadURI = request.MagnetURI
	} else if request.DownloadURI != "" {
		downloadURI = request.DownloadURI
	} else if request.Link != "" {
		downloadURI = request.Link
	}

	if downloadURI == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    "ERROR",
			Message: "下载链接不能为空 (需要 magnet_uri、download_uri 或 link 参数)",
		})
		return
	}

	// 区分磁力链接和HTTP下载链接
	var magnetURI, httpDownloadURI string
	if strings.HasPrefix(downloadURI, "magnet:") {
		magnetURI = downloadURI
	} else {
		httpDownloadURI = downloadURI
	}

	// 使用带通知的下载方法
	err := h.torrentService.DownloadTorrentWithNotification(
		magnetURI,
		httpDownloadURI,
		request.Code,
		request.Title,
		request.Tracker,
		request.Size,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "添加下载任务成功",
		Data: map[string]interface{}{
			"download_uri": downloadURI,
			"code":         request.Code,
			"title":        request.Title,
			"size":         request.Size,
			"tracker":      request.Tracker,
		},
	})
}

// GetTorrentList 获取下载列表
func (h *TorrentHandler) GetTorrentList(c *gin.Context) {
	torrents, err := h.torrentService.GetTorrentList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取下载列表成功",
		Data: map[string]interface{}{
			"count":    len(torrents),
			"torrents": torrents,
		},
	})
}

// GetDownloadStatus 获取下载状态统计
func (h *TorrentHandler) GetDownloadStatus(c *gin.Context) {
	torrents, err := h.torrentService.GetTorrentList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    "ERROR",
			Message: err.Error(),
		})
		return
	}

	// 统计各种状态的数量
	stats := map[string]int{
		"total":       len(torrents),
		"downloading": 0,
		"completed":   0,
		"paused":      0,
		"error":       0,
	}

	for _, torrent := range torrents {
		if state, ok := torrent["state"].(string); ok {
			switch state {
			case "downloading", "metaDL", "allocating":
				stats["downloading"]++
			case "uploading", "stalledUP", "queuedUP":
				stats["completed"]++
			case "pausedDL", "pausedUP":
				stats["paused"]++
			case "error", "missingFiles":
				stats["error"]++
			}
		}
	}

	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Message: "获取下载状态成功",
		Data:    stats,
	})
}
