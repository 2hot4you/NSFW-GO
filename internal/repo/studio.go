package repo

import (
	"nsfw-go/internal/model"

	"gorm.io/gorm"
)

// StudioRepository 制作商仓库接口
type StudioRepository interface {
	Create(studio *model.Studio) error
	GetByID(id uint) (*model.Studio, error)
	GetByName(name string) (*model.Studio, error)
	List(offset, limit int) ([]*model.Studio, int64, error)
	Update(studio *model.Studio) error
	Delete(id uint) error
	Count() (int64, error)
}

// studioRepository 制作商仓库实现
type studioRepository struct {
	db *gorm.DB
}

// NewStudioRepository 创建制作商仓库
func NewStudioRepository(db *gorm.DB) StudioRepository {
	return &studioRepository{db: db}
}

// Create 创建制作商
func (r *studioRepository) Create(studio *model.Studio) error {
	return r.db.Create(studio).Error
}

// GetByID 根据ID获取制作商
func (r *studioRepository) GetByID(id uint) (*model.Studio, error) {
	var studio model.Studio
	err := r.db.Preload("Movies").Preload("Series").First(&studio, id).Error
	if err != nil {
		return nil, err
	}
	return &studio, nil
}

// GetByName 根据名称获取制作商
func (r *studioRepository) GetByName(name string) (*model.Studio, error) {
	var studio model.Studio
	err := r.db.Where("name = ?", name).First(&studio).Error
	if err != nil {
		return nil, err
	}
	return &studio, nil
}

// List 获取制作商列表
func (r *studioRepository) List(offset, limit int) ([]*model.Studio, int64, error) {
	var studios []*model.Studio
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Studio{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.Offset(offset).Limit(limit).Find(&studios).Error; err != nil {
		return nil, 0, err
	}

	return studios, total, nil
}

// Update 更新制作商
func (r *studioRepository) Update(studio *model.Studio) error {
	return r.db.Save(studio).Error
}

// Delete 删除制作商
func (r *studioRepository) Delete(id uint) error {
	return r.db.Delete(&model.Studio{}, id).Error
}

// Count 获取制作商总数
func (r *studioRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Studio{}).Count(&count).Error
	return count, err
}
