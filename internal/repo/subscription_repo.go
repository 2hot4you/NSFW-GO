package repo

import (
	"time"

	"nsfw-go/internal/model"
	"gorm.io/gorm"
)

// SubscriptionRepository 订阅配置仓储接口
type SubscriptionRepository interface {
	// 基础 CRUD
	GetByRankType(rankType string) (*model.Subscription, error)
	Create(subscription *model.Subscription) error
	Update(subscription *model.Subscription) error
	GetAll() ([]*model.Subscription, error)
	
	// 订阅限制管理
	GetCurrentLimit(rankType, limitType string) (*model.SubscriptionLimit, error)
	CreateOrUpdateLimit(limit *model.SubscriptionLimit) error
	IncrementLimitCount(rankType, limitType string) error
	ResetExpiredLimits() error
	CanDownload(rankType string, hourlyLimit, dailyLimit int) (bool, *LimitStatus, error)
}

// LimitStatus 限制状态
type LimitStatus struct {
	HourlyUsed  int `json:"hourly_used"`
	HourlyLimit int `json:"hourly_limit"`
	DailyUsed   int `json:"daily_used"`
	DailyLimit  int `json:"daily_limit"`
	CanDownload bool `json:"can_download"`
}

// subscriptionRepo 订阅配置仓储实现
type subscriptionRepo struct {
	db *gorm.DB
}

// NewSubscriptionRepository 创建订阅配置仓储
func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepo{
		db: db,
	}
}

// GetByRankType 根据排行榜类型获取订阅配置
func (r *subscriptionRepo) GetByRankType(rankType string) (*model.Subscription, error) {
	var subscription model.Subscription
	err := r.db.Where("rank_type = ?", rankType).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// Create 创建订阅配置
func (r *subscriptionRepo) Create(subscription *model.Subscription) error {
	return r.db.Create(subscription).Error
}

// Update 更新订阅配置
func (r *subscriptionRepo) Update(subscription *model.Subscription) error {
	return r.db.Save(subscription).Error
}

// GetAll 获取所有订阅配置
func (r *subscriptionRepo) GetAll() ([]*model.Subscription, error) {
	var subscriptions []*model.Subscription
	err := r.db.Order("rank_type").Find(&subscriptions).Error
	return subscriptions, err
}

// GetCurrentLimit 获取当前限制记录
func (r *subscriptionRepo) GetCurrentLimit(rankType, limitType string) (*model.SubscriptionLimit, error) {
	var limit model.SubscriptionLimit
	now := time.Now()
	
	err := r.db.Where("rank_type = ? AND limit_type = ? AND period_start <= ? AND period_end > ?",
		rankType, limitType, now, now).First(&limit).Error
	
	if err == gorm.ErrRecordNotFound {
		// 创建新的限制记录
		var periodStart, periodEnd time.Time
		
		if limitType == model.LimitTypeHourly {
			periodStart = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
			periodEnd = periodStart.Add(time.Hour)
		} else { // daily
			periodStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			periodEnd = periodStart.AddDate(0, 0, 1)
		}
		
		limit = model.SubscriptionLimit{
			RankType:    rankType,
			LimitType:   limitType,
			Count:       0,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
		}
		
		if err := r.db.Create(&limit).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	
	return &limit, nil
}

// CreateOrUpdateLimit 创建或更新限制记录
func (r *subscriptionRepo) CreateOrUpdateLimit(limit *model.SubscriptionLimit) error {
	return r.db.Save(limit).Error
}

// IncrementLimitCount 增加限制计数
func (r *subscriptionRepo) IncrementLimitCount(rankType, limitType string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 获取或创建限制记录
		limit, err := r.GetCurrentLimit(rankType, limitType)
		if err != nil {
			return err
		}
		
		// 增加计数
		limit.Count++
		return tx.Save(limit).Error
	})
}

// ResetExpiredLimits 重置过期的限制记录
func (r *subscriptionRepo) ResetExpiredLimits() error {
	now := time.Now()
	return r.db.Where("period_end <= ?", now).Delete(&model.SubscriptionLimit{}).Error
}

// CanDownload 检查是否可以下载
func (r *subscriptionRepo) CanDownload(rankType string, hourlyLimit, dailyLimit int) (bool, *LimitStatus, error) {
	// 获取小时限制
	hourlyLimitRecord, err := r.GetCurrentLimit(rankType, model.LimitTypeHourly)
	if err != nil {
		return false, nil, err
	}
	
	// 获取日限制
	dailyLimitRecord, err := r.GetCurrentLimit(rankType, model.LimitTypeDaily)
	if err != nil {
		return false, nil, err
	}
	
	// 检查是否超过限制
	canDownload := hourlyLimitRecord.Count < hourlyLimit && dailyLimitRecord.Count < dailyLimit
	
	status := &LimitStatus{
		HourlyUsed:  hourlyLimitRecord.Count,
		HourlyLimit: hourlyLimit,
		DailyUsed:   dailyLimitRecord.Count,
		DailyLimit:  dailyLimit,
		CanDownload: canDownload,
	}
	
	return canDownload, status, nil
}