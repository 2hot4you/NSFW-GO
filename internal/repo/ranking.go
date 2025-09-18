package repo

import (
	"nsfw-go/internal/model"
	"time"

	"gorm.io/gorm"
)

// RankingRepository 排行榜仓库接口
type RankingRepository interface {
	Create(ranking *model.Ranking) error
	BatchCreate(rankings []model.Ranking) error
	GetByRankType(rankType string, limit int) ([]*model.Ranking, error)
	GetLatestCrawlTime(rankType string) (*time.Time, error)
	ClearOldRankings(rankType string, keepTime time.Time) error
	GetPendingCheck(limit int) ([]*model.Ranking, error)
	UpdateLocalExists(id uint, exists bool) error
	GetByCode(code string) (*model.Ranking, error)
	GetByCodeAndType(code, rankType string) (*model.Ranking, error)
	Count() (int64, error)
	Search(query string, offset, limit int) ([]*model.Ranking, int64, error)
	SearchByCode(code string, offset, limit int) ([]*model.Ranking, int64, error)
	CountByType(rankType string) (int64, error)
	CountLocalExistsByType(rankType string) (int64, error)
	GetStatsByType() (map[string]map[string]int64, error)
}

// rankingRepository 排行榜仓库实现
type rankingRepository struct {
	db *gorm.DB
}

// NewRankingRepository 创建排行榜仓库
func NewRankingRepository(db *gorm.DB) RankingRepository {
	return &rankingRepository{db: db}
}

// Create 创建排行榜记录
func (r *rankingRepository) Create(ranking *model.Ranking) error {
	return r.db.Create(ranking).Error
}

// BatchCreate 批量创建排行榜记录
func (r *rankingRepository) BatchCreate(rankings []model.Ranking) error {
	if len(rankings) == 0 {
		return nil
	}
	return r.db.CreateInBatches(rankings, 50).Error
}

// GetByRankType 根据排行榜类型获取记录
func (r *rankingRepository) GetByRankType(rankType string, limit int) ([]*model.Ranking, error) {
	var rankings []*model.Ranking
	err := r.db.Where("rank_type = ?", rankType).
		Order("position ASC").
		Limit(limit).
		Find(&rankings).Error
	return rankings, err
}

// GetLatestCrawlTime 获取最新爬取时间
func (r *rankingRepository) GetLatestCrawlTime(rankType string) (*time.Time, error) {
	var ranking model.Ranking
	err := r.db.Where("rank_type = ?", rankType).
		Order("crawled_at DESC").
		First(&ranking).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &ranking.CrawledAt, nil
}

// ClearOldRankings 清理旧的排行榜记录
func (r *rankingRepository) ClearOldRankings(rankType string, keepTime time.Time) error {
	return r.db.Where("rank_type = ? AND crawled_at < ?", rankType, keepTime).
		Delete(&model.Ranking{}).Error
}

// GetPendingCheck 获取需要检查的记录
func (r *rankingRepository) GetPendingCheck(limit int) ([]*model.Ranking, error) {
	var rankings []*model.Ranking

	// 获取从未检查过的记录，或者1小时前检查过的记录
	oneHourAgo := time.Now().Add(-time.Hour)

	err := r.db.Where("last_checked IS NULL OR last_checked < ?", oneHourAgo).
		Order("last_checked ASC NULLS FIRST").
		Limit(limit).
		Find(&rankings).Error

	return rankings, err
}

// UpdateLocalExists 更新本地存在状态
func (r *rankingRepository) UpdateLocalExists(id uint, exists bool) error {
	now := time.Now()
	return r.db.Model(&model.Ranking{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"local_exists": exists,
			"last_checked": &now,
		}).Error
}

// GetByCode 根据番号获取记录
func (r *rankingRepository) GetByCode(code string) (*model.Ranking, error) {
	var ranking model.Ranking
	err := r.db.Where("code = ?", code).First(&ranking).Error
	if err != nil {
		return nil, err
	}
	return &ranking, nil
}

// GetByCodeAndType 根据番号和排行榜类型获取记录
func (r *rankingRepository) GetByCodeAndType(code, rankType string) (*model.Ranking, error) {
	var ranking model.Ranking
	err := r.db.Where("code = ? AND rank_type = ?", code, rankType).
		Order("crawled_at DESC").
		First(&ranking).Error
	if err != nil {
		return nil, err
	}
	return &ranking, nil
}

// Count 获取排行榜记录总数
func (r *rankingRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Ranking{}).Count(&count).Error
	return count, err
}

// Search 搜索排行榜记录
func (r *rankingRepository) Search(query string, offset, limit int) ([]*model.Ranking, int64, error) {
	var rankings []*model.Ranking
	var total int64
	err := r.db.Model(&model.Ranking{}).
		Where("rank_type LIKE ? OR code LIKE ? OR title LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Where("rank_type LIKE ? OR code LIKE ? OR title LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Order("position ASC").
		Offset(offset).
		Limit(limit).
		Find(&rankings).Error
	return rankings, total, err
}

// CountByType 按类型统计排行榜数量
func (r *rankingRepository) CountByType(rankType string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Ranking{}).Where("rank_type = ?", rankType).Count(&count).Error
	return count, err
}

// CountLocalExistsByType 按类型统计本地存在的排行榜数量
func (r *rankingRepository) CountLocalExistsByType(rankType string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Ranking{}).
		Where("rank_type = ? AND local_exists = ?", rankType, true).
		Count(&count).Error
	return count, err
}

// GetStatsByType 获取按类型的统计信息
func (r *rankingRepository) GetStatsByType() (map[string]map[string]int64, error) {
	stats := make(map[string]map[string]int64)

	// 定义排行榜类型
	rankTypes := []string{"daily", "weekly", "monthly"}

	for _, rankType := range rankTypes {
		// 获取该类型的总数
		total, err := r.CountByType(rankType)
		if err != nil {
			return nil, err
		}

		// 获取该类型的本地存在数量
		local, err := r.CountLocalExistsByType(rankType)
		if err != nil {
			return nil, err
		}

		stats[rankType] = map[string]int64{
			"total": total,
			"local": local,
		}
	}

	return stats, nil
}

// SearchByCode 按番号搜索排行榜记录
func (r *rankingRepository) SearchByCode(code string, offset, limit int) ([]*model.Ranking, int64, error) {
	var rankings []*model.Ranking
	var total int64
	err := r.db.Model(&model.Ranking{}).
		Where("code = ?", code).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Where("code = ?", code).
		Order("position ASC").
		Offset(offset).
		Limit(limit).
		Find(&rankings).Error
	return rankings, total, err
}
