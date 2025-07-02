package repo

import (
	"nsfw-go/internal/model"

	"gorm.io/gorm"
)

// TagRepository 标签仓库接口
type TagRepository interface {
	Create(tag *model.Tag) error
	GetByID(id uint) (*model.Tag, error)
	GetByName(name string) (*model.Tag, error)
	List(offset, limit int) ([]*model.Tag, int64, error)
	Update(tag *model.Tag) error
	Delete(id uint) error
	Count() (int64, error)
}

// tagRepository 标签仓库实现
type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓库
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// Create 创建标签
func (r *tagRepository) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

// GetByID 根据ID获取标签
func (r *tagRepository) GetByID(id uint) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Preload("Movies").First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetByName 根据名称获取标签
func (r *tagRepository) GetByName(name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// List 获取标签列表
func (r *tagRepository) List(offset, limit int) ([]*model.Tag, int64, error) {
	var tags []*model.Tag
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Tag{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := r.db.Offset(offset).Limit(limit).Find(&tags).Error; err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// Update 更新标签
func (r *tagRepository) Update(tag *model.Tag) error {
	return r.db.Save(tag).Error
}

// Delete 删除标签
func (r *tagRepository) Delete(id uint) error {
	return r.db.Delete(&model.Tag{}, id).Error
}

// Count 获取标签总数
func (r *tagRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Tag{}).Count(&count).Error
	return count, err
}
