package main

import (
	"log"
	"time"

	"nsfw-go/internal/config"
	"nsfw-go/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	log.Println("开始添加示例数据...")

	// 添加女优数据
	actresses := []model.Actress{
		{
			BaseModel: model.BaseModel{ID: 1},
			Name:      "三上悠亚",
			AvatarURL: "https://example.com/mikami_yua.jpg",
		},
		{
			BaseModel: model.BaseModel{ID: 2},
			Name:      "葵つかさ",
			AvatarURL: "https://example.com/tsukasa_aoi.jpg",
		},
		{
			BaseModel: model.BaseModel{ID: 3},
			Name:      "桥本有菜",
			AvatarURL: "https://example.com/hashimoto_arina.jpg",
		},
		{
			BaseModel: model.BaseModel{ID: 4},
			Name:      "白石茉莉奈",
			AvatarURL: "https://example.com/shiraishi_marina.jpg",
		},
		{
			BaseModel: model.BaseModel{ID: 5},
			Name:      "深田咏美",
			AvatarURL: "https://example.com/fukata_eimi.jpg",
		},
	}

	for _, actress := range actresses {
		result := db.Where("name = ?", actress.Name).FirstOrCreate(&actress)
		if result.Error != nil {
			log.Printf("添加女优 %s 失败: %v", actress.Name, result.Error)
		} else {
			log.Printf("✓ 添加女优: %s", actress.Name)
		}
	}

	// 添加系列数据
	series := []model.Series{
		{
			BaseModel: model.BaseModel{ID: 1},
			Name:      "新人",
			StudioID:  1, // S1 No.1 Style
		},
		{
			BaseModel: model.BaseModel{ID: 2},
			Name:      "专属",
			StudioID:  1,
		},
		{
			BaseModel: model.BaseModel{ID: 3},
			Name:      "姐姐系列",
			StudioID:  2, // Prestige
		},
	}

	for _, s := range series {
		result := db.Where("name = ? AND studio_id = ?", s.Name, s.StudioID).FirstOrCreate(&s)
		if result.Error != nil {
			log.Printf("添加系列 %s 失败: %v", s.Name, result.Error)
		} else {
			log.Printf("✓ 添加系列: %s", s.Name)
		}
	}

	// 添加影片数据
	releaseDate1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	releaseDate2 := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	releaseDate3 := time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)

	movies := []model.Movie{
		{
			Code:         "SSIS-001",
			Title:        "三上悠亚的专属新作品",
			ReleaseDate:  &releaseDate1,
			Duration:     120,
			StudioID:     uintPtr(1), // S1 No.1 Style
			SeriesID:     uintPtr(1), // 新人
			Description:  "这是一部精彩的作品，展现了三上悠亚的魅力。",
			Rating:       8.5,
			CoverURL:     "https://example.com/ssis001_cover.jpg",
			Quality:      "1080p",
			HasSubtitle:  true,
			IsDownloaded: true,
			WatchCount:   150,
		},
		{
			Code:         "ABP-999",
			Title:        "葵つかさ 最新作品",
			ReleaseDate:  &releaseDate2,
			Duration:     135,
			StudioID:     uintPtr(2), // Prestige
			SeriesID:     uintPtr(3), // 姐姐系列
			Description:  "葵つかさ的最新力作，不容错过。",
			Rating:       9.0,
			CoverURL:     "https://example.com/abp999_cover.jpg",
			Quality:      "4K",
			HasSubtitle:  false,
			IsDownloaded: false,
			WatchCount:   89,
		},
		{
			Code:         "IPX-789",
			Title:        "桥本有菜 特别企划",
			ReleaseDate:  &releaseDate3,
			Duration:     140,
			StudioID:     uintPtr(3), // IdeaPocket
			SeriesID:     nil,
			Description:  "桥本有菜参与的特别企划作品。",
			Rating:       8.8,
			CoverURL:     "https://example.com/ipx789_cover.jpg",
			Quality:      "1080p",
			HasSubtitle:  true,
			IsDownloaded: true,
			WatchCount:   234,
		},
		{
			Code:         "MIDE-123",
			Title:        "白石茉莉奈 经典回归",
			ReleaseDate:  &releaseDate1,
			Duration:     125,
			StudioID:     uintPtr(4), // MOODYZ
			Description:  "白石茉莉奈的经典回归之作。",
			Rating:       8.2,
			CoverURL:     "https://example.com/mide123_cover.jpg",
			Quality:      "720p",
			HasSubtitle:  true,
			IsDownloaded: false,
			WatchCount:   67,
		},
		{
			Code:         "STARS-456",
			Title:        "深田咏美 突破极限",
			ReleaseDate:  &releaseDate2,
			Duration:     150,
			StudioID:     uintPtr(5), // SOD Create
			Description:  "深田咏美挑战自我极限的作品。",
			Rating:       9.2,
			CoverURL:     "https://example.com/stars456_cover.jpg",
			Quality:      "4K",
			HasSubtitle:  false,
			IsDownloaded: true,
			WatchCount:   312,
		},
	}

	for _, movie := range movies {
		result := db.Where("code = ?", movie.Code).FirstOrCreate(&movie)
		if result.Error != nil {
			log.Printf("添加影片 %s 失败: %v", movie.Code, result.Error)
		} else {
			log.Printf("✓ 添加影片: %s - %s", movie.Code, movie.Title)
		}
	}

	// 添加影片-女优关联关系
	movieActressRelations := []struct {
		MovieCode   string
		ActressName string
	}{
		{"SSIS-001", "三上悠亚"},
		{"ABP-999", "葵つかさ"},
		{"IPX-789", "桥本有菜"},
		{"MIDE-123", "白石茉莉奈"},
		{"STARS-456", "深田咏美"},
	}

	for _, relation := range movieActressRelations {
		var movie model.Movie
		var actress model.Actress

		if err := db.Where("code = ?", relation.MovieCode).First(&movie).Error; err != nil {
			log.Printf("找不到影片: %s", relation.MovieCode)
			continue
		}

		if err := db.Where("name = ?", relation.ActressName).First(&actress).Error; err != nil {
			log.Printf("找不到女优: %s", relation.ActressName)
			continue
		}

		if err := db.Model(&movie).Association("Actresses").Append(&actress); err != nil {
			log.Printf("关联影片和女优失败: %v", err)
		} else {
			log.Printf("✓ 关联 %s - %s", relation.MovieCode, relation.ActressName)
		}
	}

	log.Println("✓ 示例数据添加完成！")
}

func uintPtr(u uint) *uint {
	return &u
} 