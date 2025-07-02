package repo

import (
	"nsfw-go/internal/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

// LocalMovieRepository 本地影片仓库接口
type LocalMovieRepository interface {
	Create(movie *model.LocalMovie) error
	Update(movie *model.LocalMovie) error
	Delete(id uint) error
	GetByPath(path string) (*model.LocalMovie, error)
	List(offset, limit int, actress string) ([]*model.LocalMovie, int64, error)
	Count() (int64, error)
	CountByActress() (map[string]int64, error)
	Clear() error // 清空所有本地影片记录
	BulkCreate(movies []*model.LocalMovie) error
	GetLastScanTime() (time.Time, error)
	UpdateLastScanTime() error
	Search(query string, offset, limit int) ([]*model.LocalMovie, int64, error)
	SearchByActress(actress string, offset, limit int) ([]*model.LocalMovie, int64, error)
	SearchByCode(code string) (*model.LocalMovie, error)
}

// localMovieRepository 本地影片仓库实现
type localMovieRepository struct {
	db *gorm.DB
}

// NewLocalMovieRepository 创建本地影片仓库
func NewLocalMovieRepository(db *gorm.DB) LocalMovieRepository {
	return &localMovieRepository{db: db}
}

// Create 创建本地影片
func (r *localMovieRepository) Create(movie *model.LocalMovie) error {
	movie.LastScanned = time.Now()
	return r.db.Create(movie).Error
}

// Update 更新本地影片
func (r *localMovieRepository) Update(movie *model.LocalMovie) error {
	movie.LastScanned = time.Now()
	return r.db.Save(movie).Error
}

// Delete 删除本地影片
func (r *localMovieRepository) Delete(id uint) error {
	return r.db.Delete(&model.LocalMovie{}, id).Error
}

// GetByPath 根据路径获取本地影片
func (r *localMovieRepository) GetByPath(path string) (*model.LocalMovie, error) {
	var movie model.LocalMovie
	err := r.db.Where("path = ?", path).First(&movie).Error
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

// List 获取本地影片列表
func (r *localMovieRepository) List(offset, limit int, actress string) ([]*model.LocalMovie, int64, error) {
	query := r.db.Model(&model.LocalMovie{})

	if actress != "" {
		query = query.Where("actress = ?", actress)
	}

	// 计算总数
	var total int64
	countQuery := r.db.Model(&model.LocalMovie{})
	if actress != "" {
		countQuery = countQuery.Where("actress = ?", actress)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var movies []*model.LocalMovie
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&movies).Error
	if err != nil {
		return nil, 0, err
	}

	return movies, total, nil
}

// Count 获取本地影片总数
func (r *localMovieRepository) Count() (int64, error) {
	var total int64
	err := r.db.Model(&model.LocalMovie{}).Count(&total).Error
	return total, err
}

// CountByActress 按女优统计影片数量
func (r *localMovieRepository) CountByActress() (map[string]int64, error) {
	var results []struct {
		Actress string
		Count   int64
	}

	err := r.db.Model(&model.LocalMovie{}).
		Select("actress, count(*) as count").
		Group("actress").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	countMap := make(map[string]int64)
	for _, result := range results {
		countMap[result.Actress] = result.Count
	}

	return countMap, nil
}

// Clear 清空所有本地影片记录（物理删除）
func (r *localMovieRepository) Clear() error {
	return r.db.Unscoped().Where("1 = 1").Delete(&model.LocalMovie{}).Error
}

// BulkCreate 批量创建本地影片
func (r *localMovieRepository) BulkCreate(movies []*model.LocalMovie) error {
	now := time.Now()
	for _, movie := range movies {
		movie.LastScanned = now
	}
	return r.db.CreateInBatches(movies, 100).Error
}

// GetLastScanTime 获取最后扫描时间
func (r *localMovieRepository) GetLastScanTime() (time.Time, error) {
	var movie model.LocalMovie
	err := r.db.Order("last_scanned DESC").First(&movie).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil // 返回零值时间表示从未扫描过
		}
		return time.Time{}, err
	}
	return movie.LastScanned, nil
}

// UpdateLastScanTime 更新最后扫描时间
func (r *localMovieRepository) UpdateLastScanTime() error {
	now := time.Now()
	return r.db.Model(&model.LocalMovie{}).Where("1 = 1").Update("last_scanned", now).Error
}

// Search 综合搜索本地影片
func (r *localMovieRepository) Search(query string, offset, limit int) ([]*model.LocalMovie, int64, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []*model.LocalMovie{}, 0, nil
	}

	// 构建搜索条件
	searchPattern := "%" + query + "%"

	// 计算总数
	var total int64
	countQuery := r.db.Model(&model.LocalMovie{}).Where(
		"title ILIKE ? OR code ILIKE ? OR actress ILIKE ?",
		searchPattern, searchPattern, searchPattern,
	)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var movies []*model.LocalMovie
	query_db := r.db.Where(
		"title ILIKE ? OR code ILIKE ? OR actress ILIKE ?",
		searchPattern, searchPattern, searchPattern,
	).Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	err := query_db.Find(&movies).Error
	if err != nil {
		return nil, 0, err
	}

	return movies, total, nil
}

// SearchByActress 按女优搜索本地影片
func (r *localMovieRepository) SearchByActress(actress string, offset, limit int) ([]*model.LocalMovie, int64, error) {
	actress = strings.TrimSpace(actress)
	if actress == "" {
		return []*model.LocalMovie{}, 0, nil
	}

	// 计算总数
	var total int64
	countQuery := r.db.Model(&model.LocalMovie{}).Where("actress ILIKE ?", "%"+actress+"%")
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var movies []*model.LocalMovie
	err := r.db.Where("actress ILIKE ?", "%"+actress+"%").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&movies).Error
	if err != nil {
		return nil, 0, err
	}

	return movies, total, nil
}

// SearchByCode 按番号搜索本地影片
func (r *localMovieRepository) SearchByCode(code string) (*model.LocalMovie, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var movie model.LocalMovie
	err := r.db.Where("code ILIKE ?", "%"+code+"%").First(&movie).Error
	if err != nil {
		return nil, err
	}
	return &movie, nil
}
