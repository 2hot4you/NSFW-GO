package repo

import (
	"fmt"
	"nsfw-go/internal/model"
	"strings"

	"gorm.io/gorm"
)

// MovieRepository 影片仓库接口
type MovieRepository interface {
	Create(movie *model.Movie) error
	GetByID(id uint) (*model.Movie, error)
	GetByCode(code string) (*model.Movie, error)
	List(offset, limit int, filters MovieFilter) ([]*model.Movie, int64, error)
	Update(movie *model.Movie) error
	Delete(id uint) error
	Search(query string, offset, limit int) ([]*model.Movie, int64, error)
	GetRecentlyAdded(limit int) ([]*model.Movie, error)
	GetPopular(limit int) ([]*model.Movie, error)
	Count() (int64, error)
}

// MovieFilter 影片筛选条件
type MovieFilter struct {
	StudioID     *uint    `json:"studio_id"`
	SeriesID     *uint    `json:"series_id"`
	ActressIDs   []uint   `json:"actress_ids"`
	TagIDs       []uint   `json:"tag_ids"`
	Quality      string   `json:"quality"`
	HasSubtitle  *bool    `json:"has_subtitle"`
	IsDownloaded *bool    `json:"is_downloaded"`
	MinRating    *float32 `json:"min_rating"`
	MaxRating    *float32 `json:"max_rating"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	SortBy       string   `json:"sort_by"`    // created_at, rating, watch_count, release_date
	SortOrder    string   `json:"sort_order"` // asc, desc
}

// movieRepository 影片仓库实现
type movieRepository struct {
	db *gorm.DB
}

// NewMovieRepository 创建影片仓库
func NewMovieRepository(db *gorm.DB) MovieRepository {
	return &movieRepository{db: db}
}

// Create 创建影片
func (r *movieRepository) Create(movie *model.Movie) error {
	return r.db.Create(movie).Error
}

// GetByID 根据ID获取影片
func (r *movieRepository) GetByID(id uint) (*model.Movie, error) {
	var movie model.Movie
	err := r.db.Preload("Studio").
		Preload("Series").
		Preload("Actresses").
		Preload("Tags").
		First(&movie, id).Error
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

// GetByCode 根据番号获取影片
func (r *movieRepository) GetByCode(code string) (*model.Movie, error) {
	var movie model.Movie
	err := r.db.Preload("Studio").
		Preload("Series").
		Preload("Actresses").
		Preload("Tags").
		Where("code = ?", strings.ToUpper(code)).
		First(&movie).Error
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

// List 获取影片列表
func (r *movieRepository) List(offset, limit int, filters MovieFilter) ([]*model.Movie, int64, error) {
	query := r.db.Model(&model.Movie{}).
		Preload("Studio").
		Preload("Series").
		Preload("Actresses").
		Preload("Tags")

	// 应用筛选条件
	query = r.applyFilters(query, filters)

	// 计算总数
	var total int64
	countQuery := r.db.Model(&model.Movie{})
	countQuery = r.applyFilters(countQuery, filters)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	if filters.SortBy != "" {
		order := fmt.Sprintf("%s %s", filters.SortBy, r.getSortOrder(filters.SortOrder))
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	// 分页
	var movies []*model.Movie
	err := query.Offset(offset).Limit(limit).Find(&movies).Error
	if err != nil {
		return nil, 0, err
	}

	return movies, total, nil
}

// Update 更新影片
func (r *movieRepository) Update(movie *model.Movie) error {
	return r.db.Save(movie).Error
}

// Delete 删除影片
func (r *movieRepository) Delete(id uint) error {
	return r.db.Delete(&model.Movie{}, id).Error
}

// Search 搜索影片
func (r *movieRepository) Search(query string, offset, limit int) ([]*model.Movie, int64, error) {
	searchQuery := r.db.Model(&model.Movie{}).
		Preload("Studio").
		Preload("Series").
		Preload("Actresses").
		Preload("Tags")

	// 使用全文搜索
	if query != "" {
		searchCondition := fmt.Sprintf("code ILIKE ? OR title ILIKE ?")
		searchValue := "%" + query + "%"
		searchQuery = searchQuery.Where(searchCondition, searchValue, searchValue)
	}

	// 计算总数
	var total int64
	countQuery := r.db.Model(&model.Movie{})
	if query != "" {
		searchCondition := fmt.Sprintf("code ILIKE ? OR title ILIKE ?")
		searchValue := "%" + query + "%"
		countQuery = countQuery.Where(searchCondition, searchValue, searchValue)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var movies []*model.Movie
	err := searchQuery.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&movies).Error
	if err != nil {
		return nil, 0, err
	}

	return movies, total, nil
}

// GetRecentlyAdded 获取最近添加的影片
func (r *movieRepository) GetRecentlyAdded(limit int) ([]*model.Movie, error) {
	var movies []*model.Movie
	err := r.db.Preload("Studio").
		Preload("Series").
		Preload("Actresses").
		Preload("Tags").
		Order("created_at DESC").
		Limit(limit).
		Find(&movies).Error
	return movies, err
}

// GetPopular 获取热门影片
func (r *movieRepository) GetPopular(limit int) ([]*model.Movie, error) {
	var movies []*model.Movie
	err := r.db.Preload("Studio").
		Preload("Series").
		Preload("Actresses").
		Preload("Tags").
		Order("watch_count DESC, rating DESC").
		Limit(limit).
		Find(&movies).Error
	return movies, err
}

// applyFilters 应用筛选条件
func (r *movieRepository) applyFilters(query *gorm.DB, filters MovieFilter) *gorm.DB {
	if filters.StudioID != nil {
		query = query.Where("studio_id = ?", *filters.StudioID)
	}

	if filters.SeriesID != nil {
		query = query.Where("series_id = ?", *filters.SeriesID)
	}

	if len(filters.ActressIDs) > 0 {
		query = query.Joins("JOIN movie_actresses ON movies.id = movie_actresses.movie_id").
			Where("movie_actresses.actress_id IN ?", filters.ActressIDs)
	}

	if len(filters.TagIDs) > 0 {
		query = query.Joins("JOIN movie_tags ON movies.id = movie_tags.movie_id").
			Where("movie_tags.tag_id IN ?", filters.TagIDs)
	}

	if filters.Quality != "" {
		query = query.Where("quality = ?", filters.Quality)
	}

	if filters.HasSubtitle != nil {
		query = query.Where("has_subtitle = ?", *filters.HasSubtitle)
	}

	if filters.IsDownloaded != nil {
		query = query.Where("is_downloaded = ?", *filters.IsDownloaded)
	}

	if filters.MinRating != nil {
		query = query.Where("rating >= ?", *filters.MinRating)
	}

	if filters.MaxRating != nil {
		query = query.Where("rating <= ?", *filters.MaxRating)
	}

	if filters.StartDate != "" {
		query = query.Where("release_date >= ?", filters.StartDate)
	}

	if filters.EndDate != "" {
		query = query.Where("release_date <= ?", filters.EndDate)
	}

	return query
}

// getSortOrder 获取排序方向
func (r *movieRepository) getSortOrder(order string) string {
	if strings.ToLower(order) == "asc" {
		return "ASC"
	}
	return "DESC"
}

// Count 获取影片总数
func (r *movieRepository) Count() (int64, error) {
	var total int64
	err := r.db.Model(&model.Movie{}).Count(&total).Error
	return total, err
}
