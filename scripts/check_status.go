package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"nsfw-go/internal/config"
	"nsfw-go/internal/database"
)

func main() {
	fmt.Println("🔍 NSFW-Go 项目状态检查")
	fmt.Println("===============================")

	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		log.Printf("❌ 配置加载失败: %v", err)
		return
	}

	fmt.Println("✅ 配置文件加载成功")

	// 检查API服务
	checkAPIService(cfg)

	// 检查数据库连接
	checkDatabase(cfg)

	// 检查Redis连接
	checkRedis(cfg)

	// 检查数据库数据
	checkDatabaseData(cfg)

	fmt.Println("\n===============================")
	fmt.Println("📋 项目状态检查完成")
	fmt.Println("\n🎯 下一步开发建议:")
	fmt.Println("1. 运行 'make dev' 启动开发环境")
	fmt.Println("2. 访问 http://localhost:8080 查看Web界面")
	fmt.Println("3. 开始 Phase 2: 爬虫系统开发")
}

func checkAPIService(cfg *config.Config) {
	url := fmt.Sprintf("http://localhost:%d/health", cfg.Server.Port)
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ API服务检查失败: %v\n", err)
		fmt.Printf("   请运行: make run 或 ./bin/nsfw-go-api\n")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✅ API服务运行正常 (端口: %d)\n", cfg.Server.Port)
	} else {
		fmt.Printf("⚠️  API服务响应异常 (状态码: %d)\n", resp.StatusCode)
	}
}

func checkDatabase(cfg *config.Config) {
	err := database.Initialize(cfg)
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		fmt.Println("   请检查PostgreSQL服务状态: make services-status")
		return
	}
	defer database.Close()

	fmt.Printf("✅ PostgreSQL连接正常 (%s:%d)\n", cfg.Database.Host, cfg.Database.Port)
}

func checkRedis(cfg *config.Config) {
	// 简化Redis检查，不依赖redis包
	fmt.Printf("⚠️  Redis检查跳过 (未集成redis包)\n")
	fmt.Printf("   Redis地址: %s:%d\n", cfg.Redis.Host, cfg.Redis.Port)
}

func checkDatabaseData(cfg *config.Config) {
	db := database.GetDB()
	if db == nil {
		fmt.Println("   数据库未初始化")
		return
	}

	// 获取底层sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	// 检查各表的数据量
	tables := map[string]string{
		"movies":    "影片",
		"actresses": "女优",
		"studios":   "制作商",
		"tags":      "标签",
	}

	fmt.Println("\n📊 数据库数据统计:")
	for table, desc := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", table)
		err := sqlDB.QueryRow(query).Scan(&count)
		if err != nil {
			fmt.Printf("   %s: 查询失败\n", desc)
			continue
		}
		fmt.Printf("   %s: %d 条记录\n", desc, count)
	}
} 