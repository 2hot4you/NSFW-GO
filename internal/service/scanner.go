package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
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

// NFO文件结构
type NFOMovie struct {
	XMLName xml.Name `xml:"movie"`
	Title   string   `xml:"title"`
	Code    string   `xml:"num"`
	Year    string   `xml:"year"`
	Studio  string   `xml:"studio"`
	Plot    string   `xml:"plot"`
}

// ScannerService 扫描服务
type ScannerService struct {
	localMovieRepo   repo.LocalMovieRepository
	mediaLibraryPath string
	ctx              context.Context
	cancel           context.CancelFunc
	logService       *LogService
}

// NewScannerService 创建扫描服务
func NewScannerService(localMovieRepo repo.LocalMovieRepository, mediaLibraryPath string, logService *LogService) *ScannerService {
	ctx, cancel := context.WithCancel(context.Background())
	return &ScannerService{
		localMovieRepo:   localMovieRepo,
		mediaLibraryPath: mediaLibraryPath,
		ctx:              ctx,
		cancel:           cancel,
		logService:       logService,
	}
}

// Start 启动定时扫描
func (s *ScannerService) Start() {
	if s.mediaLibraryPath == "" {
		if s.logService != nil {
			s.logService.LogWarn("scanner", "media-scan", "媒体库路径未配置，跳过本地影片扫描")
		}
		return
	}
	if s.logService != nil {
		s.logService.LogInfo("scanner", "media-scan", fmt.Sprintf("启动本地影片扫描服务，媒体库路径: %s，每15分钟扫描一次", s.mediaLibraryPath))
	}

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
				if s.logService != nil {
					s.logService.LogInfo("scanner", "media-scan", "本地影片扫描服务已停止")
				}
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
	if s.logService != nil {
		s.logService.LogInfo("scanner", "media-scan", "开始扫描本地影片库...")
	}
	startTime := time.Now()

	// 扫描文件系统
	movies, err := s.scanDirectory(s.mediaLibraryPath)
	if err != nil {
		if s.logService != nil {
			s.logService.LogError("scanner", "media-scan", fmt.Sprintf("扫描失败: %v", err))
		}
		return
	}

	// 清空旧数据
	if err := s.localMovieRepo.Clear(); err != nil {
		if s.logService != nil {
			s.logService.LogError("scanner", "media-scan", fmt.Sprintf("清空旧数据失败: %v", err))
		}
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
				if s.logService != nil {
					s.logService.LogWarn("scanner", "media-scan", fmt.Sprintf("插入影片失败 [%s]: %v", movie.Path, err))
				}
				continue
			}
			successCount++
		}
		if s.logService != nil {
			s.logService.LogInfo("scanner", "media-scan", fmt.Sprintf("成功插入 %d 部影片，跳过重复 %d 部，总计扫描 %d 部", successCount, skipCount, len(movies)))
		}
	}

	duration := time.Since(startTime)
	if s.logService != nil {
		s.logService.LogInfo("scanner", "media-scan", fmt.Sprintf("扫描完成，共找到 %d 部影片，耗时 %v", len(movies), duration))
	}
}

// ForceRescan 强制重新扫描
func (s *ScannerService) ForceRescan() error {
	if s.logService != nil {
		s.logService.LogInfo("scanner", "media-scan", "手动触发重新扫描...")
	}
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

// findVideoFiles 查找视频文件（只查找主视频，排除花絮等）
func (s *ScannerService) findVideoFiles(dirPath string) ([]string, error) {
	var videoFiles []string
	videoExtensions := []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".m4v", ".flv", ".webm"}
	
	// 需要排除的目录名
	excludeDirs := []string{"behind the scenes", "extrafanart", "trailers", "extras", "sample", "samples"}
	
	// 最小文件大小（100MB，排除预览等小文件）
	minFileSize := int64(100 * 1024 * 1024)

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续处理
		}

		// 检查是否在排除的目录中
		for _, excludeDir := range excludeDirs {
			if strings.Contains(strings.ToLower(path), excludeDir) {
				if d.IsDir() {
					return filepath.SkipDir // 跳过整个目录
				}
				return nil // 跳过文件
			}
		}

		if d.IsDir() {
			return nil
		}

		// 获取文件信息
		fileInfo, err := d.Info()
		if err != nil {
			return nil
		}
		
		// 排除小文件
		if fileInfo.Size() < minFileSize {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		fileName := strings.ToLower(filepath.Base(path))
		
		// 排除包含 sample、trailer、preview 等关键词的文件
		if strings.Contains(fileName, "sample") || 
		   strings.Contains(fileName, "trailer") || 
		   strings.Contains(fileName, "preview") ||
		   strings.Contains(fileName, "fanart") {
			return nil
		}
		
		for _, validExt := range videoExtensions {
			if ext == validExt {
				// 优先匹配番号格式的文件名（如 START-395.mp4）
				if matched, _ := regexp.MatchString(`^[A-Z]+-\d+\.[a-z]+$`, fileName); matched {
					// 这是主视频文件，添加到列表前面
					videoFiles = append([]string{path}, videoFiles...)
				} else {
					// 其他视频文件添加到后面
					videoFiles = append(videoFiles, path)
				}
				break
			}
		}

		return nil
	})
	
	// 如果找到了主视频文件（番号格式），只返回第一个
	if len(videoFiles) > 0 {
		// 检查第一个是否是番号格式
		fileName := filepath.Base(videoFiles[0])
		if matched, _ := regexp.MatchString(`^[A-Z]+-\d+\.[a-z]+$`, fileName); matched {
			return []string{videoFiles[0]}, err
		}
		// 否则返回第一个大文件作为主视频
		return []string{videoFiles[0]}, err
	}

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
	
	// 尝试从NFO文件读取详细信息
	movieDir := filepath.Dir(filePath)
	if nfoTitle, nfoCode := s.readNFOFile(movieDir); nfoTitle != "" {
		title = nfoTitle
		if nfoCode != "" && code == "" {
			code = nfoCode
		}
	}

	// 查找fanart图片
	fanartPath, fanartURL, hasFanart := s.findFanart(movieDir)

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

// findFanart 查找fanart图片（优先fanart.jpg）
func (s *ScannerService) findFanart(movieDir string) (string, string, bool) {
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".webp", ".bmp"}
	// 按优先级排序：fanart.jpg 优先级最高
	fanartNames := []string{"fanart", "poster", "thumb", "cover", "thumbnail"}

	// 遍历目录查找fanart图片
	files, err := os.ReadDir(movieDir)
	if err != nil {
		return "", "", false
	}

	// 优先查找完全匹配的文件
	for _, fanartName := range fanartNames {
		for _, ext := range imageExtensions {
			targetFile := fanartName + ext
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				
				if strings.ToLower(file.Name()) == targetFile {
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
	}

	// 如果没找到精确匹配的，再查找包含关键词的图片
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := strings.ToLower(file.Name())
		fileExt := filepath.Ext(fileName)

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

		// 检查是否包含fanart相关命名
		for _, fanartName := range fanartNames {
			if strings.Contains(fileName, fanartName) {
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

// readNFOFile 读取NFO文件获取影片信息
func (s *ScannerService) readNFOFile(movieDir string) (string, string) {
	// 查找NFO文件
	files, err := os.ReadDir(movieDir)
	if err != nil {
		return "", ""
	}
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		// 检查是否是NFO文件
		if strings.ToLower(filepath.Ext(file.Name())) == ".nfo" {
			nfoPath := filepath.Join(movieDir, file.Name())
			
			// 读取NFO文件
			nfoFile, err := os.Open(nfoPath)
			if err != nil {
				continue
			}
			defer nfoFile.Close()
			
			// 读取文件内容
			data, err := io.ReadAll(nfoFile)
			if err != nil {
				continue
			}
			
			// 解析XML
			var nfoMovie NFOMovie
			err = xml.Unmarshal(data, &nfoMovie)
			if err != nil {
				// 如果XML解析失败，尝试提取CDATA中的内容
				titlePattern := regexp.MustCompile(`<title><!\[CDATA\[(.*?)\]\]></title>`)
				if matches := titlePattern.FindSubmatch(data); len(matches) > 1 {
					return string(matches[1]), ""
				}
				continue
			}
			
			return nfoMovie.Title, nfoMovie.Code
		}
	}
	
	return "", ""
}
