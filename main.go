package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"mmth-etl/aggregator"
	"mmth-etl/i18n"
	"mmth-etl/storage"
	"mmth-etl/types"
	"os"
	"path/filepath"
)

// 版本信息，构建时通过 ldflags 注入
var Version = "dev"

// 类型别名，语义化
type DiamondAggregator = aggregator.ChangeAggregator
type RuneTicketAggregator = aggregator.ChangeAggregator
type UpgradePanaceaAggregator = aggregator.ChangeAggregator

func main() {
	// 显示版本信息
	fmt.Printf("MMTH ETL v%s\n", Version)

	// 命令行参数
	outputDir := flag.String("output", "./data", "输出目录路径")
	langFlag := flag.String("lang", "dynamic", "日志语言 (en, tw, ja, ko, auto, dynamic)")
	flag.Parse()

	// 检查日志文件参数
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("用法: mmth_etl [-output <输出目录>] [-lang <语言>] <日志文件路径>")
		fmt.Println("  -lang: en (英文), tw (繁中), ja (日文), ko (韩文), auto (启动时自动检测), dynamic (运行时动态检测)")
		os.Exit(1)
	}
	inputLogPath := args[0]

	// 初始化 i18n 管理器
	i18nMgr := i18n.NewManager()

	// 动态语言检测配置
	dynamicCfg := DynamicLanguageConfig{
		Enabled:         false,
		WindowSize:      100,
		SwitchThreshold: 5,
	}

	// 设置语言
	switch *langFlag {
	case "dynamic":
		// 动态检测模式：在运行时持续检测语言变化
		dynamicCfg.Enabled = true
		// 初始使用英文，运行时会根据检测结果切换
		i18nMgr.SetLanguage(i18n.LangEn)
		log.Printf("启用动态语言检测模式 (窗口: %d, 切换阈值: %d)", dynamicCfg.WindowSize, dynamicCfg.SwitchThreshold)
	case "auto":
		// 自动检测语言（启动时一次性检测）
		detector := i18n.NewDetector()
		sampleLines := readSampleLines(inputLogPath, 500)
		detectedLang := detector.Detect(sampleLines)
		i18nMgr.SetLanguage(detectedLang)
		log.Printf("检测到日志语言: %s", detectedLang)
	default:
		i18nMgr.SetLanguage(i18n.Language(*langFlag))
		log.Printf("使用指定语言: %s", *langFlag)
	}

	// 初始化全局模式
	types.InitI18n(i18nMgr)

	// 配置路径
	diamondJSONPath := filepath.Join(*outputDir, "diamond_stats.json")
	caveJSONPath := filepath.Join(*outputDir, "cave_stats.json")
	challengeJSONPath := filepath.Join(*outputDir, "challenge_stats.json")
	runeTicketJSONPath := filepath.Join(*outputDir, "rune_ticket_stats.json")
	upgradePanaceaJSONPath := filepath.Join(*outputDir, "upgrade_panacea_stats.json")
	stateFilePath := filepath.Join(*outputDir, "mmth_etl_state.json")

	// 加载检查点
	checkpointMgr := storage.NewCheckpointManager(stateFilePath)
	checkpoint := checkpointMgr.Load()

	// 创建处理器
	processor := NewLogProcessor(inputLogPath, checkpoint, i18nMgr, dynamicCfg)

	// 创建聚合器
	diamondAgg := aggregator.NewChangeAggregator(false)
	caveAgg := aggregator.NewCaveAggregator()
	challengeAgg := aggregator.NewChallengeAggregator()
	runeTicketAgg := aggregator.NewChangeAggregator(false)
	upgradePanaceaAgg := aggregator.NewChangeAggregator(false)

	// 加载已有统计（增量处理）
	diamondAgg.LoadExistingStats(diamondJSONPath)
	caveAgg.LoadExistingStats(caveJSONPath)
	challengeAgg.LoadExistingStats(challengeJSONPath)
	runeTicketAgg.LoadExistingStats(runeTicketJSONPath)
	upgradePanaceaAgg.LoadExistingStats(upgradePanaceaJSONPath)

	log.Println("已加载现有统计数据")

	// 流式处理日志
	lastLogTime := processor.Process(diamondAgg, caveAgg, challengeAgg, runeTicketAgg, upgradePanaceaAgg)

	// 检查是否有新记录
	if diamondAgg.RecordCount() == 0 && caveAgg.RecordCount() == 0 && challengeAgg.RecordCount() == 0 && runeTicketAgg.RecordCount() == 0 && upgradePanaceaAgg.RecordCount() == 0 {
		fmt.Println("没有新的记录需要处理")
		return
	}

	fmt.Printf("新增 %d 条钻石记录\n", diamondAgg.RecordCount())
	fmt.Printf("新增 %d 条洞穴记录\n", caveAgg.RecordCount())
	fmt.Printf("新增 %d 条挑战记录\n", challengeAgg.RecordCount())
	fmt.Printf("新增 %d 条饼干记录\n", runeTicketAgg.RecordCount())
	fmt.Printf("新增 %d 条红水记录\n", upgradePanaceaAgg.RecordCount())

	// 输出统计结果
	if diamondAgg.RecordCount() > 0 {
		storage.SaveStats(diamondAgg.ToMap(), diamondJSONPath)
	}

	if caveAgg.RecordCount() > 0 {
		SaveCaveStats(caveAgg.ToMap(), caveJSONPath)
	}

	if challengeAgg.RecordCount() > 0 {
		SaveChallengeStats(challengeAgg.ToMap(), challengeJSONPath)
	}

	if runeTicketAgg.RecordCount() > 0 {
		storage.SaveStats(runeTicketAgg.ToMap(), runeTicketJSONPath)
	}

	if upgradePanaceaAgg.RecordCount() > 0 {
		storage.SaveStats(upgradePanaceaAgg.ToMap(), upgradePanaceaJSONPath)
	}

	// 更新检查点
	if lastLogTime != "" {
		checkpointMgr.Save(lastLogTime)
	}

	fmt.Printf("处理完成\n")
}

// readSampleLines reads a sample of lines from a file for language detection.
func readSampleLines(filePath string, maxLines int) []string {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() && len(lines) < maxLines {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}
