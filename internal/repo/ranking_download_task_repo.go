package repo

import (
	"time"

	"nsfw-go/internal/model"
	"gorm.io/gorm"
)

// RankingDownloadTaskRepository 排行榜下载任务仓储接口
type RankingDownloadTaskRepository interface {
	// 基础 CRUD
	Create(task *model.RankingDownloadTask) error
	GetByID(id uint) (*model.RankingDownloadTask, error)
	GetByCode(code string) (*model.RankingDownloadTask, error)
	Update(task *model.RankingDownloadTask) error
	Delete(id uint) error
	HardDelete(id uint) error
	
	// 查询方法
	GetActiveTaskByCode(code string) (*model.RankingDownloadTask, error)
	GetTasksByStatus(status string) ([]*model.RankingDownloadTask, error)
	GetTasksBySource(source string) ([]*model.RankingDownloadTask, error)
	GetTasksByRankType(rankType string) ([]*model.RankingDownloadTask, error)
	
	// 分页查询
	GetTasks(limit, offset int) ([]*model.RankingDownloadTask, int64, error)
	GetTasksWithFilter(status, source, rankType string, limit, offset int) ([]*model.RankingDownloadTask, int64, error)
	
	// 统计方法
	CountTasksByStatus(status string) (int64, error)
	CountTasksBySource(source string) (int64, error)
	GetTaskStats() (*TaskStats, error)
	
	// 批量操作
	BatchUpdateStatus(ids []uint, status string, errorMsg string) error
	BatchDelete(ids []uint) error
	CleanupOldTasks(days int) error
}

// TaskStats 任务统计信息
type TaskStats struct {
	Total      int64 `json:"total"`
	Pending    int64 `json:"pending"`
	Searching  int64 `json:"searching"`
	Progress   int64 `json:"progress"`
	Completed  int64 `json:"completed"`
	Failed     int64 `json:"failed"`
	Cancelled  int64 `json:"cancelled"`
	Manual     int64 `json:"manual"`
	Subscription int64 `json:"subscription"`
}

// rankingDownloadTaskRepo 排行榜下载任务仓储实现
type rankingDownloadTaskRepo struct {
	db *gorm.DB
}

// NewRankingDownloadTaskRepository 创建排行榜下载任务仓储
func NewRankingDownloadTaskRepository(db *gorm.DB) RankingDownloadTaskRepository {
	return &rankingDownloadTaskRepo{
		db: db,
	}
}

// Create 创建下载任务
func (r *rankingDownloadTaskRepo) Create(task *model.RankingDownloadTask) error {
	return r.db.Create(task).Error
}

// GetByID 根据ID获取任务
func (r *rankingDownloadTaskRepo) GetByID(id uint) (*model.RankingDownloadTask, error) {
	var task model.RankingDownloadTask
	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByCode 根据番号获取任务（不包括软删除）
func (r *rankingDownloadTaskRepo) GetByCode(code string) (*model.RankingDownloadTask, error) {
	var task model.RankingDownloadTask
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetActiveTaskByCode 获取指定番号的活跃任务（非完成、非失败）
func (r *rankingDownloadTaskRepo) GetActiveTaskByCode(code string) (*model.RankingDownloadTask, error) {
	var task model.RankingDownloadTask
	err := r.db.Where("code = ? AND status NOT IN (?, ?, ?) AND deleted_at IS NULL", 
		code, 
		model.RankingDownloadStatusCompleted, 
		model.RankingDownloadStatusFailed,
		model.RankingDownloadStatusCancelled,
	).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update 更新任务
func (r *rankingDownloadTaskRepo) Update(task *model.RankingDownloadTask) error {
	return r.db.Save(task).Error
}

// Delete 删除任务（软删除）
func (r *rankingDownloadTaskRepo) Delete(id uint) error {
	return r.db.Delete(&model.RankingDownloadTask{}, id).Error
}

// HardDelete 硬删除任务（彻底删除记录）
func (r *rankingDownloadTaskRepo) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&model.RankingDownloadTask{}, id).Error
}

// GetTasksByStatus 根据状态获取任务
func (r *rankingDownloadTaskRepo) GetTasksByStatus(status string) ([]*model.RankingDownloadTask, error) {
	var tasks []*model.RankingDownloadTask
	err := r.db.Where("status = ?", status).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetTasksBySource 根据来源获取任务
func (r *rankingDownloadTaskRepo) GetTasksBySource(source string) ([]*model.RankingDownloadTask, error) {
	var tasks []*model.RankingDownloadTask
	err := r.db.Where("source = ?", source).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetTasksByRankType 根据排行榜类型获取任务
func (r *rankingDownloadTaskRepo) GetTasksByRankType(rankType string) ([]*model.RankingDownloadTask, error) {
	var tasks []*model.RankingDownloadTask
	err := r.db.Where("rank_type = ?", rankType).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetTasks 获取任务列表（分页）
func (r *rankingDownloadTaskRepo) GetTasks(limit, offset int) ([]*model.RankingDownloadTask, int64, error) {
	var tasks []*model.RankingDownloadTask
	var count int64
	
	// 统计总数
	if err := r.db.Model(&model.RankingDownloadTask{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error
	return tasks, count, err
}

// GetTasksWithFilter 带筛选条件的任务列表（分页）
func (r *rankingDownloadTaskRepo) GetTasksWithFilter(status, source, rankType string, limit, offset int) ([]*model.RankingDownloadTask, int64, error) {
	var tasks []*model.RankingDownloadTask
	var count int64
	
	query := r.db.Model(&model.RankingDownloadTask{})
	
	// 添加筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if rankType != "" {
		query = query.Where("rank_type = ?", rankType)
	}
	
	// 统计总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error
	return tasks, count, err
}

// CountTasksByStatus 统计指定状态的任务数量
func (r *rankingDownloadTaskRepo) CountTasksByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.RankingDownloadTask{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountTasksBySource 统计指定来源的任务数量
func (r *rankingDownloadTaskRepo) CountTasksBySource(source string) (int64, error) {
	var count int64
	err := r.db.Model(&model.RankingDownloadTask{}).Where("source = ?", source).Count(&count).Error
	return count, err
}

// GetTaskStats 获取任务统计信息
func (r *rankingDownloadTaskRepo) GetTaskStats() (*TaskStats, error) {
	stats := &TaskStats{}
	
	// 总数
	if err := r.db.Model(&model.RankingDownloadTask{}).Count(&stats.Total).Error; err != nil {
		return nil, err
	}
	
	// 按状态统计
	statusCounts := []struct {
		Status string
		Count  int64
	}{}
	
	err := r.db.Model(&model.RankingDownloadTask{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, sc := range statusCounts {
		switch sc.Status {
		case model.RankingDownloadStatusPending:
			stats.Pending = sc.Count
		case model.RankingDownloadStatusSearching:
			stats.Searching = sc.Count
		case model.RankingDownloadStatusProgress:
			stats.Progress = sc.Count
		case model.RankingDownloadStatusCompleted:
			stats.Completed = sc.Count
		case model.RankingDownloadStatusFailed:
			stats.Failed = sc.Count
		case model.RankingDownloadStatusCancelled:
			stats.Cancelled = sc.Count
		}
	}
	
	// 按来源统计
	sourceCounts := []struct {
		Source string
		Count  int64
	}{}
	
	err = r.db.Model(&model.RankingDownloadTask{}).
		Select("source, count(*) as count").
		Group("source").
		Scan(&sourceCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, sc := range sourceCounts {
		switch sc.Source {
		case model.DownloadSourceManual:
			stats.Manual = sc.Count
		case model.DownloadSourceSubscription:
			stats.Subscription = sc.Count
		}
	}
	
	return stats, nil
}

// BatchUpdateStatus 批量更新任务状态
func (r *rankingDownloadTaskRepo) BatchUpdateStatus(ids []uint, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
		"updated_at": time.Now(),
	}
	
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	
	if status == model.RankingDownloadStatusCompleted {
		updates["completed_at"] = time.Now()
	}
	
	return r.db.Model(&model.RankingDownloadTask{}).
		Where("id IN ?", ids).
		Updates(updates).Error
}

// BatchDelete 批量删除任务
func (r *rankingDownloadTaskRepo) BatchDelete(ids []uint) error {
	return r.db.Delete(&model.RankingDownloadTask{}, ids).Error
}

// CleanupOldTasks 清理旧任务（超过指定天数的已完成/失败任务）
func (r *rankingDownloadTaskRepo) CleanupOldTasks(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	
	return r.db.Where("status IN (?, ?) AND updated_at < ?", 
		model.RankingDownloadStatusCompleted, 
		model.RankingDownloadStatusFailed,
		cutoffTime).
		Delete(&model.RankingDownloadTask{}).Error
}