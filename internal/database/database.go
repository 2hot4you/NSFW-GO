package database

import (
	"fmt"
	"log"
	"nsfw-go/internal/config"
	"nsfw-go/internal/model"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize 初始化数据库连接
func Initialize(cfg *config.Config) error {
	// 构建DSN
	dsn := cfg.Database.GetDSN()

	// 配置GORM
	gormConfig := &gorm.Config{}

	// 根据调试模式设置日志级别
	if cfg.Server.Mode == "debug" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取SQL DB实例失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.MaxLifetime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 自动迁移模型（可选，主要用于开发环境）
	if cfg.Server.Mode == "debug" {
		if err := autoMigrate(db); err != nil {
			log.Printf("警告：自动迁移失败: %v", err)
		}
	}

	DB = db
	log.Println("✓ 数据库连接初始化成功")
	return nil
}

// autoMigrate 自动迁移数据库模式
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Movie{},
		&model.Actress{},
		&model.Studio{},
		&model.Series{},
		&model.Tag{},
		&model.DownloadTask{},
		&model.CrawlTask{},
		&model.User{},
		&model.WatchHistory{},
		&model.Favorite{},
		&model.LocalMovie{},
	)
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
