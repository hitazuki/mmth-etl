package main

import (
	"fmt"
	"os"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: diamond_tracker <日志文件路径>")
		os.Exit(1)
	}
	inputLogPath := os.Args[1]

	// 配置
	outputJSONPath := "./data/diamond_stats.json"
	stateFilePath := "./data/mmth_etl_state.json"

	processor := &LogProcessor{
		outputJSONPath: outputJSONPath,
		stateFilePath:  stateFilePath,
		inputLogPath:   inputLogPath,
	}

	// 加载上次记录的时间戳
	processor.loadCheckpoint()

	// 创建聚合器（不保留详细记录以节省内存）
	// 如果需要保留 records，改为 NewAggregator(true)
	agg := NewAggregator(false)

	// 加载已有统计（增量处理）
	if existingStats := loadExistingStats(outputJSONPath); existingStats != nil {
		agg.LoadExistingStats(existingStats)
		fmt.Printf("已加载现有统计数据\n")
	}

	// 流式处理日志（内存友好）
	lastLogTime := processor.processStream(agg)

	// 检查是否有新记录
	if agg.RecordCount() == 0 {
		fmt.Println("没有新的钻石记录需要处理")
		return
	}

	fmt.Printf("新增 %d 条钻石记录\n", agg.RecordCount())

	// 输出统计结果
	stats := agg.ToMap()
	SaveStats(stats, outputJSONPath)

	// 更新状态
	if lastLogTime != "" {
		processor.checkpoint = lastLogTime
		processor.saveCheckpoint()
	}

	fmt.Printf("处理完成\n")
}
