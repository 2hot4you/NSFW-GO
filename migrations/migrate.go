package migrations

import (
	"fmt"
	"log"

	"nsfw-go/internal/model"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Migrate 执行数据库迁移
func Migrate(dsn string) error {
	// 配置GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 启用UUID扩展
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("创建UUID扩展失败（可能已存在）: %v", err)
	}

	// 自动迁移所有模型
	models := model.GetAllModels()
	
	log.Println("开始数据库迁移...")
	
	for _, modelPtr := range models {
		if err := db.AutoMigrate(modelPtr); err != nil {
			return fmt.Errorf("迁移模型失败 %T: %w", modelPtr, err)
		}
		log.Printf("✓ 迁移模型: %T", modelPtr)
	}

	// 创建索引
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	// 插入初始数据
	if err := seedData(db); err != nil {
		return fmt.Errorf("插入初始数据失败: %w", err)
	}

	log.Println("✓ 数据库迁移完成")
	return nil
}

// createIndexes 创建额外的索引
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		// 影片相关索引
		"CREATE INDEX IF NOT EXISTS idx_movies_code_gin ON movies USING gin(code gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_movies_title_gin ON movies USING gin(title gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_movies_release_date ON movies(release_date)",
		"CREATE INDEX IF NOT EXISTS idx_movies_rating ON movies(rating)",
		"CREATE INDEX IF NOT EXISTS idx_movies_created_at ON movies(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_movies_watch_count ON movies(watch_count)",
		
		// 演员相关索引
		"CREATE INDEX IF NOT EXISTS idx_actresses_name_gin ON actresses USING gin(name gin_trgm_ops)",
		
		// 下载任务索引
		"CREATE INDEX IF NOT EXISTS idx_download_tasks_status ON download_tasks(status)",
		"CREATE INDEX IF NOT EXISTS idx_download_tasks_created_at ON download_tasks(created_at)",
		
		// 爬虫任务索引
		"CREATE INDEX IF NOT EXISTS idx_crawl_tasks_type_status ON crawl_tasks(type, status)",
		"CREATE INDEX IF NOT EXISTS idx_crawl_tasks_created_at ON crawl_tasks(created_at)",
		
		// 观看历史索引
		"CREATE INDEX IF NOT EXISTS idx_watch_history_watched_at ON watch_history(watched_at)",
	}

	// 启用pg_trgm扩展（用于模糊搜索）
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		log.Printf("创建pg_trgm扩展失败（可能已存在）: %v", err)
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("创建索引失败: %s, 错误: %v", indexSQL, err)
		} else {
			log.Printf("✓ 创建索引: %s", indexSQL)
		}
	}

	return nil
}

// seedData 插入初始数据
func seedData(db *gorm.DB) error {
	// 检查是否已有数据
	var count int64
	db.Model(&model.Tag{}).Count(&count)
	if count > 0 {
		log.Println("数据库已有数据，跳过初始化")
		return nil
	}

	log.Println("插入初始数据...")

	// 插入默认标签
	defaultTags := []model.Tag{
		{Name: "高清", Category: model.TagCategoryQuality},
		{Name: "4K", Category: model.TagCategoryQuality},
		{Name: "1080p", Category: model.TagCategoryQuality},
		{Name: "720p", Category: model.TagCategoryQuality},
		{Name: "中文字幕", Category: model.TagCategoryOther},
		{Name: "无码", Category: model.TagCategoryGenre},
		{Name: "有码", Category: model.TagCategoryGenre},
		{Name: "素人", Category: model.TagCategoryGenre},
		{Name: "偶像", Category: model.TagCategoryGenre},
		{Name: "制服", Category: model.TagCategoryGenre},
		{Name: "OL", Category: model.TagCategoryGenre},
		{Name: "学生", Category: model.TagCategoryGenre},
		{Name: "熟女", Category: model.TagCategoryGenre},
		{Name: "人妻", Category: model.TagCategoryGenre},
		{Name: "巨乳", Category: model.TagCategoryGenre},
		{Name: "美乳", Category: model.TagCategoryGenre},
		{Name: "美少女", Category: model.TagCategoryGenre},
		{Name: "单体作品", Category: model.TagCategoryGenre},
		{Name: "合集", Category: model.TagCategoryGenre},
		{Name: "VR", Category: model.TagCategoryOther},
	}

	for _, tag := range defaultTags {
		if err := db.FirstOrCreate(&tag, model.Tag{Name: tag.Name}).Error; err != nil {
			log.Printf("插入标签失败: %s, 错误: %v", tag.Name, err)
		}
	}

	// 插入一些知名制作商
	defaultStudios := []model.Studio{
		{Name: "SOD Create"},
		{Name: "Prestige"},
		{Name: "S1 No.1 Style"},
		{Name: "MOODYZ"},
		{Name: "Idea Pocket"},
		{Name: "E-BODY"},
		{Name: "FALENO"},
		{Name: "Alice JAPAN"},
		{Name: "Kawaii"},
		{Name: "Madonna"},
		{Name: "Premium"},
		{Name: "MUTEKI"},
		{Name: "Natural High"},
		{Name: "Fitch"},
		{Name: "Wanz Factory"},
	}

	for _, studio := range defaultStudios {
		if err := db.FirstOrCreate(&studio, model.Studio{Name: studio.Name}).Error; err != nil {
			log.Printf("插入制作商失败: %s, 错误: %v", studio.Name, err)
		}
	}

	log.Println("✓ 初始数据插入完成")
	return nil
}

// DropAllTables 删除所有表（用于重置数据库）
func DropAllTables(dsn string) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取所有表名
	var tables []string
	db.Raw(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename NOT LIKE 'pg_%'
	`).Scan(&tables)

	// 删除所有表
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Printf("删除表失败: %s, 错误: %v", table, err)
		} else {
			log.Printf("✓ 删除表: %s", table)
		}
	}

	log.Println("✓ 所有表已删除")
	return nil
}

// CheckConnection 检查数据库连接
func CheckConnection(dsn string) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库ping失败: %w", err)
	}

	log.Println("✓ 数据库连接正常")
	return nil
} 