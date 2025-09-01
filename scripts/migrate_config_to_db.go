package main

import (
	"fmt"
	"log"
	"nsfw-go/internal/config"
	"nsfw-go/internal/database"
	"nsfw-go/internal/service"
	"os"
)

// MigrateConfigToDB 将配置文件迁移到数据库
func main() {
	// 获取配置文件路径
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()

	// 创建配置服务
	configStoreService := service.NewConfigStoreService()

	// 将配置结构体保存到数据库
	fmt.Println("开始将配置迁移到数据库...")

	if err := configStoreService.SaveStructToConfig(cfg); err != nil {
		log.Fatalf("保存配置到数据库失败: %v", err)
	}

	fmt.Println("✓ 配置迁移完成")

	// 创建初始备份
	fmt.Println("创建初始备份...")
	if err := configStoreService.CreateBackup(
		"Initial Migration",
		"从配置文件自动迁移的初始配置备份",
		"system",
	); err != nil {
		log.Printf("创建初始备份失败: %v", err)
	} else {
		fmt.Println("✓ 初始备份创建完成")
	}

	// 验证迁移结果
	fmt.Println("验证迁移结果...")
	configs, err := configStoreService.GetAllConfigs()
	if err != nil {
		log.Fatalf("获取配置列表失败: %v", err)
	}

	fmt.Printf("✓ 成功迁移 %d 个配置项\n", len(configs))

	// 显示配置分类统计
	categoryCount := make(map[string]int)
	for _, config := range configs {
		categoryCount[config.Category]++
	}

	fmt.Println("配置分类统计:")
	for category, count := range categoryCount {
		fmt.Printf("  %s: %d 项\n", category, count)
	}

	fmt.Println("迁移完成！现在可以通过 API 管理配置了。")
}
