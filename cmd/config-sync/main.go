package main

import (
	"flag"
	"fmt"
	"log"
	"nsfw-go/internal/config"
	"nsfw-go/internal/database"
	"nsfw-go/internal/model"
	"nsfw-go/internal/service"
	"os"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
	mode       = flag.String("mode", "sync", "操作模式: sync(同步到数据库) | export(导出到文件) | show(显示配置)")
	category   = flag.String("category", "", "指定配置分类")
	backup     = flag.Bool("backup", false, "同步前创建备份")
	restore    = flag.Uint("restore", 0, "从指定备份ID恢复配置")
	listBackup = flag.Bool("list-backup", false, "列出所有备份")
	help       = flag.Bool("help", false, "显示帮助信息")
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// 加载配置文件
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化数据库
	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()

	// 创建配置存储服务
	configStore := service.NewConfigStoreService()

	// 处理备份相关操作
	if *listBackup {
		listBackups(configStore)
		return
	}

	if *restore > 0 {
		restoreBackup(configStore, *restore)
		return
	}

	// 根据模式执行操作
	switch *mode {
	case "sync":
		syncToDatabase(cfg, configStore)
	case "export":
		exportToFile(configStore)
	case "show":
		showConfigs(configStore, *category)
	default:
		log.Fatalf("无效的操作模式: %s", *mode)
	}
}

func syncToDatabase(cfg *config.Config, configStore *service.ConfigStoreService) {
	log.Println("开始将配置同步到数据库...")

	// 如果需要备份
	if *backup {
		err := configStore.CreateBackup(
			fmt.Sprintf("同步前备份_%s", *configPath),
			"配置文件同步前的自动备份",
			"config-sync",
		)
		if err != nil {
			log.Printf("创建备份失败: %v", err)
		} else {
			log.Println("✓ 已创建配置备份")
		}
	}

	// 将整个配置结构保存到数据库
	if err := configStore.SaveStructToConfig(cfg); err != nil {
		log.Fatalf("同步配置到数据库失败: %v", err)
	}

	log.Println("✓ 配置已成功同步到数据库")

	// 显示同步的配置数量
	configs, err := configStore.GetAllConfigs()
	if err == nil {
		log.Printf("共同步 %d 个配置项", len(configs))
	}
}

func exportToFile(configStore *service.ConfigStoreService) {
	log.Println("从数据库导出配置...")

	// 创建一个新的配置结构
	cfg := &config.Config{}

	// 从数据库加载配置到结构体
	if err := configStore.LoadConfigToStruct(cfg); err != nil {
		log.Fatalf("从数据库加载配置失败: %v", err)
	}

	// TODO: 实现导出到YAML文件的逻辑
	log.Println("导出功能暂未实现，配置已加载到内存")
	fmt.Printf("%+v\n", cfg)
}

func showConfigs(configStore *service.ConfigStoreService, category string) {
	var configs []model.ConfigStore
	var err error

	if category == "" {
		// 获取所有配置
		configs, err = configStore.GetAllConfigs()
	} else {
		// 获取指定分类的配置
		configs, err = configStore.GetConfigsByCategory(category)
	}

	if err != nil {
		log.Fatalf("获取配置失败: %v", err)
	}

	if len(configs) == 0 {
		log.Println("没有找到配置项")
		return
	}

	// 显示配置
	fmt.Println("\n=== 数据库配置项 ===")
	fmt.Printf("%-30s %-15s %-10s %s\n", "键", "类型", "分类", "值")
	fmt.Println(string(make([]byte, 80)))

	for _, cfg := range configs {
		value := cfg.Value
		if cfg.IsSecret {
			value = "******"
		}
		// 截断过长的值
		if len(value) > 40 {
			value = value[:37] + "..."
		}
		fmt.Printf("%-30s %-15s %-10s %s\n", cfg.Key, cfg.Type, cfg.Category, value)
	}
}

func listBackups(configStore *service.ConfigStoreService) {
	backups, err := configStore.GetBackups()
	if err != nil {
		log.Fatalf("获取备份列表失败: %v", err)
	}

	if len(backups) == 0 {
		log.Println("没有找到配置备份")
		return
	}

	fmt.Println("\n=== 配置备份列表 ===")
	fmt.Printf("%-5s %-30s %-20s %s\n", "ID", "名称", "创建者", "创建时间")
	fmt.Println(string(make([]byte, 80)))

	for _, backup := range backups {
		fmt.Printf("%-5d %-30s %-20s %s\n",
			backup.ID,
			backup.Name,
			backup.CreatedBy,
			backup.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
}

func restoreBackup(configStore *service.ConfigStoreService, backupID uint) {
	log.Printf("开始从备份 #%d 恢复配置...", backupID)

	if err := configStore.RestoreFromBackup(backupID); err != nil {
		log.Fatalf("恢复配置失败: %v", err)
	}

	log.Printf("✓ 已成功从备份 #%d 恢复配置", backupID)
}

func showHelp() {
	fmt.Printf(`配置同步工具 - 在文件和数据库之间同步配置

使用方法:
  %s [选项]

选项:
  -config string     配置文件路径 (默认: config.yaml)
  -mode string       操作模式 (默认: sync)
                     sync   - 将配置文件同步到数据库
                     export - 从数据库导出配置到文件
                     show   - 显示数据库中的配置
  -category string   指定配置分类 (用于show模式)
  -backup           同步前创建备份
  -restore uint     从指定备份ID恢复配置
  -list-backup      列出所有配置备份
  -help             显示此帮助信息

示例:
  # 将config.yaml同步到数据库
  %s -config config.yaml -mode sync

  # 同步前创建备份
  %s -config config.yaml -mode sync -backup

  # 显示所有数据库配置
  %s -mode show

  # 显示特定分类的配置
  %s -mode show -category database

  # 列出所有备份
  %s -list-backup

  # 从备份恢复
  %s -restore 1

注意:
  - sync模式会覆盖数据库中的现有配置
  - 建议在同步前使用-backup选项创建备份
  - 数据库配置优先级高于文件配置
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}