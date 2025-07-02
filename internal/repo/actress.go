package repo

import (
	"nsfw-go/internal/model"

	"gorm.io/gorm"
)

type ActressRepository struct {
	db *gorm.DB
}

func NewActressRepository(db *gorm.DB) *ActressRepository {
	return &ActressRepository{db: db}
}

// Create 创建女优
func (r *ActressRepository) Create(actress *model.Actress) error {
	return r.db.Create(actress).Error
}

// GetByID 根据ID获取女优
func (r *ActressRepository) GetByID(id uint) (*model.Actress, error) {
	var actress model.Actress
	if err := r.db.Preload("Movies").First(&actress, id).Error; err != nil {
		return nil, err
	}
	return &actress, nil
}

// GetByName 根据姓名获取女优
func (r *ActressRepository) GetByName(name string) (*model.Actress, error) {
	var actress model.Actress
	if err := r.db.Where("name = ?", name).Preload("Movies").First(&actress).Error; err != nil {
		return nil, err
	}
	return &actress, nil
}

// Update 更新女优
func (r *ActressRepository) Update(actress *model.Actress) error {
	return r.db.Save(actress).Error
}

// Delete 删除女优
func (r *ActressRepository) Delete(id uint) error {
	return r.db.Delete(&model.Actress{}, id).Error
}

// List 获取女优列表
func (r *ActressRepository) List(offset, limit int) ([]model.Actress, int64, error) {
	var actresses []model.Actress
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Actress{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.Offset(offset).Limit(limit).Find(&actresses).Error; err != nil {
		return nil, 0, err
	}

	return actresses, total, nil
}

// Search 搜索女优
func (r *ActressRepository) Search(keyword string, offset, limit int) ([]model.Actress, int64, error) {
	var actresses []model.Actress
	var total int64

	query := r.db.Model(&model.Actress{}).Where("name ILIKE ?", "%"+keyword+"%")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := query.Offset(offset).Limit(limit).Find(&actresses).Error; err != nil {
		return nil, 0, err
	}

	return actresses, total, nil
}

// GetMovies 获取女优的影片列表
func (r *ActressRepository) GetMovies(actressID uint, offset, limit int) ([]model.Movie, int64, error) {
	var movies []model.Movie
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Movie{}).
		Joins("JOIN movie_actresses ON movies.id = movie_actresses.movie_id").
		Where("movie_actresses.actress_id = ?", actressID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.Preload("Studio").Preload("Series").Preload("Actresses").Preload("Tags").
		Joins("JOIN movie_actresses ON movies.id = movie_actresses.movie_id").
		Where("movie_actresses.actress_id = ?", actressID).
		Offset(offset).Limit(limit).
		Find(&movies).Error; err != nil {
		return nil, 0, err
	}

	return movies, total, nil
}

// GetPopular 获取热门女优
func (r *ActressRepository) GetPopular(limit int) ([]model.Actress, error) {
	var actresses []model.Actress

	if err := r.db.
		Select("actresses.*, COUNT(movie_actresses.actress_id) as movie_count").
		Joins("LEFT JOIN movie_actresses ON actresses.id = movie_actresses.actress_id").
		Group("actresses.id").
		Order("movie_count DESC").
		Limit(limit).
		Find(&actresses).Error; err != nil {
		return nil, err
	}

	return actresses, nil
}

// GetTotal 获取女优总数
func (r *ActressRepository) GetTotal() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Actress{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Count 获取女优总数（别名方法）
func (r *ActressRepository) Count() (int64, error) {
	return r.GetTotal()
}
