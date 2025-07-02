package repo

import (
	"nsfw-go/internal/model"

	"gorm.io/gorm"
)

// SeriesRepository 系列仓库接口
type SeriesRepository interface {
	Create(series *model.Series) error
	GetByID(id uint) (*model.Series, error)
	GetByName(name string) (*model.Series, error)
	List(offset, limit int) ([]*model.Series, int64, error)
	Update(series *model.Series) error
	Delete(id uint) error
	Count() (int64, error)
}

// seriesRepository 系列仓库实现
type seriesRepository struct {
	db *gorm.DB
}

// NewSeriesRepository 创建系列仓库
func NewSeriesRepository(db *gorm.DB) SeriesRepository {
	return &seriesRepository{db: db}
}

// Create 创建系列
func (r *seriesRepository) Create(series *model.Series) error {
	return r.db.Create(series).Error
}

// GetByID 根据ID获取系列
func (r *seriesRepository) GetByID(id uint) (*model.Series, error) {
	var series model.Series
	err := r.db.Preload("Studio").Preload("Movies").First(&series, id).Error
	if err != nil {
		return nil, err
	}
	return &series, nil
}

// GetByName 根据名称获取系列
func (r *seriesRepository) GetByName(name string) (*model.Series, error) {
	var series model.Series
	err := r.db.Where("name = ?", name).First(&series).Error
	if err != nil {
		return nil, err
	}
	return &series, nil
}

// List 获取系列列表
func (r *seriesRepository) List(offset, limit int) ([]*model.Series, int64, error) {
	var series []*model.Series
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Series{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.Preload("Studio").Offset(offset).Limit(limit).Find(&series).Error; err != nil {
		return nil, 0, err
	}

	return series, total, nil
}

// Update 更新系列
func (r *seriesRepository) Update(series *model.Series) error {
	return r.db.Save(series).Error
}

// Delete 删除系列
func (r *seriesRepository) Delete(id uint) error {
	return r.db.Delete(&model.Series{}, id).Error
}

// Count 获取系列总数
func (r *seriesRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Series{}).Count(&count).Error
	return count, err
}
