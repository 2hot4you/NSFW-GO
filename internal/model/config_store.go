package model

import (
	"encoding/json"
	"time"
)

// ConfigStore 配置存储表
type ConfigStore struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"uniqueIndex;not null;size:100" json:"key"`
	Value       string    `gorm:"type:text" json:"value"`
	Type        string    `gorm:"size:20;not null" json:"type"` // string, int, bool, json, array
	Description string    `gorm:"size:500" json:"description"`
	Category    string    `gorm:"size:50;index" json:"category"`
	IsSecret    bool      `gorm:"default:false" json:"is_secret"`
	Version     int       `gorm:"default:1" json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConfigCategory 配置分类
type ConfigCategory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:50" json:"name"`
	DisplayName string    `gorm:"not null;size:100" json:"display_name"`
	Description string    `gorm:"size:500" json:"description"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConfigTemplate 配置模板（用于定义配置项的结构和验证规则）
type ConfigTemplate struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Key            string    `gorm:"uniqueIndex;not null;size:100" json:"key"`
	Name           string    `gorm:"not null;size:100" json:"name"`
	Description    string    `gorm:"size:500" json:"description"`
	Type           string    `gorm:"size:20;not null" json:"type"`
	DefaultValue   string    `gorm:"type:text" json:"default_value"`
	Category       string    `gorm:"size:50;index" json:"category"`
	Required       bool      `gorm:"default:false" json:"required"`
	IsSecret       bool      `gorm:"default:false" json:"is_secret"`
	ValidationRule string    `gorm:"type:text" json:"validation_rule"` // JSON格式的验证规则
	Options        string    `gorm:"type:text" json:"options"`         // JSON格式的选项列表
	SortOrder      int       `gorm:"default:0" json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ConfigValue 配置值的通用接口
type ConfigValue interface {
	String() string
	Int() int
	Bool() bool
	Float64() float64
	JSON(v interface{}) error
}

// configValue 实现 ConfigValue 接口
type configValue struct {
	value string
	vtype string
}

func NewConfigValue(value, vtype string) ConfigValue {
	return &configValue{value: value, vtype: vtype}
}

func (c *configValue) String() string {
	return c.value
}

func (c *configValue) Int() int {
	var result int
	json.Unmarshal([]byte(c.value), &result)
	return result
}

func (c *configValue) Bool() bool {
	var result bool
	json.Unmarshal([]byte(c.value), &result)
	return result
}

func (c *configValue) Float64() float64 {
	var result float64
	json.Unmarshal([]byte(c.value), &result)
	return result
}

func (c *configValue) JSON(v interface{}) error {
	return json.Unmarshal([]byte(c.value), v)
}

// ConfigBackup 配置备份
type ConfigBackup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;size:100" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	ConfigData  string    `gorm:"type:longtext" json:"config_data"` // JSON格式的完整配置
	Version     string    `gorm:"size:20" json:"version"`
	CreatedBy   string    `gorm:"size:50" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (ConfigStore) TableName() string {
	return "config_store"
}

func (ConfigCategory) TableName() string {
	return "config_categories"
}

func (ConfigTemplate) TableName() string {
	return "config_templates"
}

func (ConfigBackup) TableName() string {
	return "config_backups"
}
