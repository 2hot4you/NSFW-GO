package service

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"nsfw-go/internal/model"
	"nsfw-go/internal/repo"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	ScanInterval = 15 * time.Minute // 15分钟扫描一次
)

// ScannerService 扫描服务
type ScannerService struct {
	localMovieRepo   repo.LocalMovieRepository
	mediaLibraryPath string
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewScannerService 创建扫描服务
func NewScannerService(localMovieRepo repo.LocalMovieRepository, mediaLibraryPath string) *ScannerService {
	ctx, cancel := context.WithCancel(context.Background())
	return &ScannerService{
		localMovieRepo:   localMovieRepo,
		mediaLibraryPath: mediaLibraryPath,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start 启动定时扫描
func (s *ScannerService) Start() {
	if s.mediaLibraryPath == "" {
		log.Println("📂 媒体库路径未配置，跳过本地影片扫描")
		return
	}
	log.Printf("🔍 启动本地影片扫描服务，媒体库路径: %s，每15分钟扫描一次", s.mediaLibraryPath)

	// 立即执行一次扫描
	go s.scanAndStore()

	// 启动定时扫描
	ticker := time.NewTicker(ScanInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.scanAndStore()
			case <-s.ctx.Done():
				ticker.Stop()
				log.Println("📴 本地影片扫描服务已停止")
				return
			}
		}
	}()
}

// Stop 停止扫描服务
func (s *ScannerService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// scanAndStore 扫描并存储到数据库
func (s *ScannerService) scanAndStore() {
	log.Println("🔍 开始扫描本地影片库...")
	startTime := time.Now()

	// 扫描文件系统
	movies, err := s.scanDirectory(s.mediaLibraryPath)
	if err != nil {
		log.Printf("❌ 扫描失败: %v", err)
		return
	}

	// 清空旧数据
	if err := s.localMovieRepo.Clear(); err != nil {
		log.Printf("❌ 清空旧数据失败: %v", err)
		return
	}

	// 逐个插入新数据（避免重复路径错误）
	if len(movies) > 0 {
		successCount := 0
		skipCount := 0
		for _, movie := range movies {
			err := s.localMovieRepo.Create(movie)
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key value") {
					skipCount++
					continue
				}
				log.Printf("⚠️ 插入影片失败 [%s]: %v", movie.Path, err)
				continue
			}
			successCount++
		}
		log.Printf("✅ 成功插入 %d 部影片，跳过重复 %d 部，总计扫描 %d 部", successCount, skipCount, len(movies))
	}

	duration := time.Since(startTime)
	log.Printf("✅ 扫描完成，共找到 %d 部影片，耗时 %v", len(movies), duration)
}

// ForceRescan 强制重新扫描
func (s *ScannerService) ForceRescan() error {
	log.Println("🔄 手动触发重新扫描...")
	s.scanAndStore()
	return nil
}

// scanDirectory 扫描指定目录
func (s *ScannerService) scanDirectory(rootPath string) ([]*model.LocalMovie, error) {
	var movies []*model.LocalMovie

	// 检查目录是否存在
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return movies, fmt.Errorf("媒体库目录不存在: %s", rootPath)
	}

	// 遍历女优目录
	actressDirs, err := os.ReadDir(rootPath)
	if err != nil {
		return movies, err
	}

	for _, actressDir := range actressDirs {
		if !actressDir.IsDir() || strings.HasPrefix(actressDir.Name(), ".") {
			continue
		}

		actressPath := filepath.Join(rootPath, actressDir.Name())
		actressName := actressDir.Name()

		// 遍历女优目录下的影片
		movieDirs, err := os.ReadDir(actressPath)
		if err != nil {
			continue
		}

		for _, movieDir := range movieDirs {
			if !movieDir.IsDir() || strings.HasPrefix(movieDir.Name(), ".") {
				continue
			}

			moviePath := filepath.Join(actressPath, movieDir.Name())

			// 扫描影片目录中的视频文件
			videoFiles, err := s.findVideoFiles(moviePath)
			if err != nil {
				continue
			}

			for _, videoFile := range videoFiles {
				movie := s.parseMovieInfo(videoFile, actressName, movieDir.Name())
				if movie != nil {
					movies = append(movies, movie)
				}
			}
		}
	}

	return movies, nil
}

// findVideoFiles 查找视频文件
func (s *ScannerService) findVideoFiles(dirPath string) ([]string, error) {
	var videoFiles []string
	videoExtensions := []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".m4v", ".flv", ".webm"}

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续处理
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, validExt := range videoExtensions {
			if ext == validExt {
				videoFiles = append(videoFiles, path)
				break
			}
		}

		return nil
	})

	return videoFiles, err
}

// parseMovieInfo 解析影片信息
func (s *ScannerService) parseMovieInfo(filePath, actress, dirName string) *model.LocalMovie {
	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil
	}

	// 提取番号和标题
	code, title := s.extractCodeAndTitle(dirName, filepath.Base(filePath))

	// 查找fanart图片
	fanartPath, fanartURL, hasFanart := s.findFanart(filepath.Dir(filePath))

	return &model.LocalMovie{
		Title:       title,
		Code:        code,
		Actress:     actress,
		Path:        filePath,
		Size:        fileInfo.Size(),
		Modified:    fileInfo.ModTime(),
		Format:      strings.ToUpper(strings.TrimPrefix(filepath.Ext(filePath), ".")),
		FanartPath:  fanartPath,
		FanartURL:   fanartURL,
		HasFanart:   hasFanart,
		LastScanned: time.Now(),
	}
}

// extractCodeAndTitle 从目录名或文件名中提取番号和标题
func (s *ScannerService) extractCodeAndTitle(dirName, fileName string) (string, string) {
	// 优先从目录名提取
	if code, title := s.parseNameForCode(dirName); code != "" {
		return code, title
	}

	// 如果目录名没有番号，从文件名提取
	if code, title := s.parseNameForCode(fileName); code != "" {
		return code, title
	}

	// 如果都没有，使用目录名作为标题
	return "", dirName
}

// parseNameForCode 解析名称中的番号
func (s *ScannerService) parseNameForCode(name string) (string, string) {
	// 匹配 [CODE-123] 格式
	re1 := regexp.MustCompile(`\[([A-Z]+[-_]?\d+)\](.*)`)
	if matches := re1.FindStringSubmatch(name); len(matches) >= 2 {
		code := strings.ToUpper(matches[1])
		title := strings.TrimSpace(matches[2])
		if title == "" {
			title = name
		}
		return code, title
	}

	// 匹配 CODE-123 格式（不在括号中）
	re2 := regexp.MustCompile(`^([A-Z]+[-_]?\d+)\s*(.*)`)
	if matches := re2.FindStringSubmatch(name); len(matches) >= 2 {
		code := strings.ToUpper(matches[1])
		title := strings.TrimSpace(matches[2])
		if title == "" {
			title = name
		}
		return code, title
	}

	return "", name
}

// findFanart 查找fanart图片
func (s *ScannerService) findFanart(movieDir string) (string, string, bool) {
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".webp", ".bmp"}
	fanartNames := []string{"fanart", "poster", "cover", "thumb", "thumbnail"}

	// 遍历目录查找fanart图片
	files, err := os.ReadDir(movieDir)
	if err != nil {
		return "", "", false
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := strings.ToLower(file.Name())
		fileExt := filepath.Ext(fileName)
		fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)

		// 检查是否是图片文件
		isImage := false
		for _, ext := range imageExtensions {
			if fileExt == ext {
				isImage = true
				break
			}
		}

		if !isImage {
			continue
		}

		// 检查是否是fanart命名
		for _, fanartName := range fanartNames {
			if fileNameWithoutExt == fanartName {
				fullPath := filepath.Join(movieDir, file.Name())
				// 生成相对于媒体库的URL路径
				relPath, err := filepath.Rel(s.mediaLibraryPath, fullPath)
				if err != nil {
					continue
				}
				// 将路径转换为URL格式并进行编码
				urlPath := "/api/v1/local/image/" + url.PathEscape(strings.ReplaceAll(relPath, "\\", "/"))
				return fullPath, urlPath, true
			}
		}
	}

	return "", "", false
}
