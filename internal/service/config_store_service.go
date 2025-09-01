package service

import (
	"encoding/json"
	"fmt"
	"nsfw-go/internal/database"
	"nsfw-go/internal/model"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ConfigStoreService struct {
	db *gorm.DB
}

func NewConfigStoreService() *ConfigStoreService {
	return &ConfigStoreService{
		db: database.DB,
	}
}

// GetConfig 获取配置值
func (s *ConfigStoreService) GetConfig(key string) (model.ConfigValue, error) {
	var config model.ConfigStore
	err := s.db.Where("key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return model.NewConfigValue(config.Value, config.Type), nil
}

// SetConfig 设置配置值
func (s *ConfigStoreService) SetConfig(key, value, vtype, category, description string, isSecret bool) error {
	var config model.ConfigStore
	err := s.db.Where("key = ?", key).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新配置
		config = model.ConfigStore{
			Key:         key,
			Value:       value,
			Type:        vtype,
			Category:    category,
			Description: description,
			IsSecret:    isSecret,
			Version:     1,
		}
		return s.db.Create(&config).Error
	} else if err != nil {
		return err
	}

	// 更新现有配置
	config.Value = value
	config.Type = vtype
	config.Category = category
	config.Description = description
	config.IsSecret = isSecret
	config.Version++
	config.UpdatedAt = time.Now()

	return s.db.Save(&config).Error
}

// GetAllConfigs 获取所有配置
func (s *ConfigStoreService) GetAllConfigs() ([]model.ConfigStore, error) {
	var configs []model.ConfigStore
	err := s.db.Order("category, key").Find(&configs).Error
	return configs, err
}

// GetConfigsByCategory 根据分类获取配置
func (s *ConfigStoreService) GetConfigsByCategory(category string) ([]model.ConfigStore, error) {
	var configs []model.ConfigStore
	err := s.db.Where("category = ?", category).Order("key").Find(&configs).Error
	return configs, err
}

// DeleteConfig 删除配置
func (s *ConfigStoreService) DeleteConfig(key string) error {
	return s.db.Where("key = ?", key).Delete(&model.ConfigStore{}).Error
}

// BatchSetConfigs 批量设置配置
func (s *ConfigStoreService) BatchSetConfigs(configs map[string]interface{}, category string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range configs {
			jsonValue, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal config %s: %w", key, err)
			}

			vtype := s.getValueType(value)

			var config model.ConfigStore
			err = tx.Where("key = ?", key).First(&config).Error

			if err == gorm.ErrRecordNotFound {
				config = model.ConfigStore{
					Key:      key,
					Value:    string(jsonValue),
					Type:     vtype,
					Category: category,
					Version:  1,
				}
				if err := tx.Create(&config).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				config.Value = string(jsonValue)
				config.Type = vtype
				config.Category = category
				config.Version++
				config.UpdatedAt = time.Now()
				if err := tx.Save(&config).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// LoadConfigToStruct 将数据库配置加载到结构体
func (s *ConfigStoreService) LoadConfigToStruct(configStruct interface{}) error {
	configs, err := s.GetAllConfigs()
	if err != nil {
		return err
	}

	configMap := make(map[string]string)
	for _, config := range configs {
		configMap[config.Key] = config.Value
	}

	return s.mapToStruct(configMap, configStruct)
}

// SaveStructToConfig 将结构体配置保存到数据库
func (s *ConfigStoreService) SaveStructToConfig(configStruct interface{}) error {
	configMap, err := s.structToMap(configStruct, "")
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range configMap {
			category := s.extractCategory(key)

			jsonValue, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal config %s: %w", key, err)
			}

			vtype := s.getValueType(value)

			var config model.ConfigStore
			err = tx.Where("key = ?", key).First(&config).Error

			if err == gorm.ErrRecordNotFound {
				config = model.ConfigStore{
					Key:      key,
					Value:    string(jsonValue),
					Type:     vtype,
					Category: category,
					Version:  1,
				}
				if err := tx.Create(&config).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				config.Value = string(jsonValue)
				config.Type = vtype
				config.Category = category
				config.Version++
				config.UpdatedAt = time.Now()
				if err := tx.Save(&config).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// CreateBackup 创建配置备份
func (s *ConfigStoreService) CreateBackup(name, description, createdBy string) error {
	configs, err := s.GetAllConfigs()
	if err != nil {
		return err
	}

	configData, err := json.Marshal(configs)
	if err != nil {
		return err
	}

	backup := model.ConfigBackup{
		Name:        name,
		Description: description,
		ConfigData:  string(configData),
		Version:     "1.0",
		CreatedBy:   createdBy,
	}

	return s.db.Create(&backup).Error
}

// RestoreFromBackup 从备份恢复配置
func (s *ConfigStoreService) RestoreFromBackup(backupID uint) error {
	var backup model.ConfigBackup
	if err := s.db.First(&backup, backupID).Error; err != nil {
		return err
	}

	var configs []model.ConfigStore
	if err := json.Unmarshal([]byte(backup.ConfigData), &configs); err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 清空现有配置
		if err := tx.Delete(&model.ConfigStore{}, "1 = 1").Error; err != nil {
			return err
		}

		// 恢复配置
		for _, config := range configs {
			config.ID = 0 // 重置ID让数据库自动生成
			config.CreatedAt = time.Now()
			config.UpdatedAt = time.Now()
			if err := tx.Create(&config).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetBackups 获取配置备份列表
func (s *ConfigStoreService) GetBackups() ([]model.ConfigBackup, error) {
	var backups []model.ConfigBackup
	err := s.db.Order("created_at DESC").Find(&backups).Error
	return backups, err
}

// DeleteBackup 删除配置备份
func (s *ConfigStoreService) DeleteBackup(backupID uint) error {
	return s.db.Delete(&model.ConfigBackup{}, backupID).Error
}

// 辅助方法

func (s *ConfigStoreService) getValueType(value interface{}) string {
	switch value.(type) {
	case bool:
		return "bool"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "int"
	case float32, float64:
		return "float"
	case string:
		return "string"
	case []interface{}, []string, []int:
		return "array"
	default:
		return "json"
	}
}

func (s *ConfigStoreService) extractCategory(key string) string {
	parts := strings.Split(key, ".")
	if len(parts) > 1 {
		return parts[0]
	}
	return "general"
}

func (s *ConfigStoreService) structToMap(obj interface{}, prefix string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过未导出的字段
		if !fieldValue.CanInterface() {
			continue
		}

		// 获取字段名
		fieldName := field.Name
		if tag := field.Tag.Get("mapstructure"); tag != "" && tag != "-" {
			fieldName = tag
		}

		// 构建完整的键名
		var key string
		if prefix != "" {
			key = prefix + "." + strings.ToLower(fieldName)
		} else {
			key = strings.ToLower(fieldName)
		}

		// 处理不同类型的字段
		if fieldValue.Kind() == reflect.Struct {
			// 递归处理嵌套结构体
			nestedMap, err := s.structToMap(fieldValue.Interface(), key)
			if err != nil {
				return nil, err
			}
			for k, v := range nestedMap {
				result[k] = v
			}
		} else {
			result[key] = fieldValue.Interface()
		}
	}

	return result, nil
}

func (s *ConfigStoreService) mapToStruct(configMap map[string]string, configStruct interface{}) error {
	v := reflect.ValueOf(configStruct)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("configStruct must be a pointer to struct")
	}

	v = v.Elem()
	t := v.Type()

	return s.setStructFields(v, t, configMap, "")
}

func (s *ConfigStoreService) setStructFields(v reflect.Value, t reflect.Type, configMap map[string]string, prefix string) error {
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		fieldName := field.Name
		if tag := field.Tag.Get("mapstructure"); tag != "" && tag != "-" {
			fieldName = tag
		}

		var key string
		if prefix != "" {
			key = prefix + "." + strings.ToLower(fieldName)
		} else {
			key = strings.ToLower(fieldName)
		}

		if fieldValue.Kind() == reflect.Struct {
			// 递归处理嵌套结构体
			err := s.setStructFields(fieldValue, fieldValue.Type(), configMap, key)
			if err != nil {
				return err
			}
		} else {
			// 设置字段值
			if value, exists := configMap[key]; exists {
				if err := s.setFieldValue(fieldValue, value); err != nil {
					return fmt.Errorf("failed to set field %s: %w", key, err)
				}
			}
		}
	}

	return nil
}

func (s *ConfigStoreService) setFieldValue(fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		var str string
		if err := json.Unmarshal([]byte(value), &str); err != nil {
			fieldValue.SetString(value)
		} else {
			fieldValue.SetString(str)
		}
	case reflect.Bool:
		var b bool
		if err := json.Unmarshal([]byte(value), &b); err != nil {
			b, _ = strconv.ParseBool(value)
		}
		fieldValue.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		if err := json.Unmarshal([]byte(value), &i); err != nil {
			i, _ = strconv.ParseInt(value, 10, 64)
		}
		fieldValue.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var u uint64
		if err := json.Unmarshal([]byte(value), &u); err != nil {
			u, _ = strconv.ParseUint(value, 10, 64)
		}
		fieldValue.SetUint(u)
	case reflect.Float32, reflect.Float64:
		var f float64
		if err := json.Unmarshal([]byte(value), &f); err != nil {
			f, _ = strconv.ParseFloat(value, 64)
		}
		fieldValue.SetFloat(f)
	case reflect.Slice:
		return json.Unmarshal([]byte(value), fieldValue.Addr().Interface())
	case reflect.Map:
		return json.Unmarshal([]byte(value), fieldValue.Addr().Interface())
	default:
		return json.Unmarshal([]byte(value), fieldValue.Addr().Interface())
	}

	return nil
}
