package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Aggregator 统计聚合器（支持增量更新，内存友好）
type Aggregator struct {
	dailyStats   map[string]map[string]*DailyStats   // character -> date -> stats
	weeklyStats  map[string]map[string]*WeeklyStats  // character -> week -> stats
	monthlyStats map[string]map[string]*MonthlyStats // character -> month -> stats
	totalStats   map[string]*TotalStats              // character -> total
	keepRecords  bool                                // 是否保留详细记录
	recordCount  int                                 // 已处理记录数
}

// NewAggregator 创建聚合器
func NewAggregator(keepRecords bool) *Aggregator {
	return &Aggregator{
		dailyStats:   make(map[string]map[string]*DailyStats),
		weeklyStats:  make(map[string]map[string]*WeeklyStats),
		monthlyStats: make(map[string]map[string]*MonthlyStats),
		totalStats:   make(map[string]*TotalStats),
		keepRecords:  keepRecords,
		recordCount:  0,
	}
}

// AddRecord 增量添加记录（处理完可立即释放原始记录）
func (a *Aggregator) AddRecord(record DiamondRecord) {
	character := record.Character

	// 初始化角色统计
	if a.dailyStats[character] == nil {
		a.dailyStats[character] = make(map[string]*DailyStats)
		a.weeklyStats[character] = make(map[string]*WeeklyStats)
		a.monthlyStats[character] = make(map[string]*MonthlyStats)
		a.totalStats[character] = &TotalStats{Sources: make(map[string]SourceStats)}
	}

	// 解析日期（从时间戳提取 YYYY-MM-DD）
	date := record.Timestamp[:10]
	t, _ := time.Parse("2006-01-02", date)
	week := getWeekString(t)
	month := getMonthString(t)

	// 更新每日统计
	if a.dailyStats[character][date] == nil {
		a.dailyStats[character][date] = &DailyStats{
			Date:    date,
			Sources: make(map[string]SourceStats),
		}
	}
	a.updateDailyStats(a.dailyStats[character][date], record)

	// 更新每周统计
	if a.weeklyStats[character][week] == nil {
		a.weeklyStats[character][week] = &WeeklyStats{
			Week:    week,
			Sources: make(map[string]SourceStats),
		}
	}
	a.updateWeeklyStats(a.weeklyStats[character][week], record)

	// 更新每月统计
	if a.monthlyStats[character][month] == nil {
		a.monthlyStats[character][month] = &MonthlyStats{
			Month:   month,
			Sources: make(map[string]SourceStats),
		}
	}
	a.updateMonthlyStats(a.monthlyStats[character][month], record)

	// 更新总计
	a.updateTotalStats(a.totalStats[character], record)

	a.recordCount++
}

// updateDailyStats 更新每日统计
func (a *Aggregator) updateDailyStats(ds *DailyStats, record DiamondRecord) {
	if record.Type == "gain" {
		ds.Gain += record.Amount
		src := ds.Sources[record.Source]
		src.Gain += record.Amount
		ds.Sources[record.Source] = src
	} else {
		ds.Consume += record.Amount
		src := ds.Sources[record.Source]
		src.Consume += record.Amount
		ds.Sources[record.Source] = src
	}
	ds.NetChange = ds.Gain - ds.Consume

	// 只有 keepRecords=true 时才保留详细记录
	if a.keepRecords {
		ds.Records = append(ds.Records, record)
	}
}

// updateWeeklyStats 更新每周统计
func (a *Aggregator) updateWeeklyStats(ws *WeeklyStats, record DiamondRecord) {
	if record.Type == "gain" {
		ws.Gain += record.Amount
		src := ws.Sources[record.Source]
		src.Gain += record.Amount
		ws.Sources[record.Source] = src
	} else {
		ws.Consume += record.Amount
		src := ws.Sources[record.Source]
		src.Consume += record.Amount
		ws.Sources[record.Source] = src
	}
	ws.NetChange = ws.Gain - ws.Consume
}

// updateMonthlyStats 更新每月统计
func (a *Aggregator) updateMonthlyStats(ms *MonthlyStats, record DiamondRecord) {
	if record.Type == "gain" {
		ms.Gain += record.Amount
		src := ms.Sources[record.Source]
		src.Gain += record.Amount
		ms.Sources[record.Source] = src
	} else {
		ms.Consume += record.Amount
		src := ms.Sources[record.Source]
		src.Consume += record.Amount
		ms.Sources[record.Source] = src
	}
	ms.NetChange = ms.Gain - ms.Consume
}

// updateTotalStats 更新总计
func (a *Aggregator) updateTotalStats(ts *TotalStats, record DiamondRecord) {
	if record.Type == "gain" {
		ts.Gain += record.Amount
		src := ts.Sources[record.Source]
		src.Gain += record.Amount
		ts.Sources[record.Source] = src
	} else {
		ts.Consume += record.Amount
		src := ts.Sources[record.Source]
		src.Consume += record.Amount
		ts.Sources[record.Source] = src
	}
	ts.NetChange = ts.Gain - ts.Consume
}

// LoadExistingStats 加载现有统计数据（保留原有统计，用于增量合并）
func (a *Aggregator) LoadExistingStats(existingStats map[string]map[string]any) {
	if existingStats == nil {
		return
	}

	// 加载已有统计（直接合并到内存）
	for character, charData := range existingStats {
		// 初始化角色统计
		if a.dailyStats[character] == nil {
			a.dailyStats[character] = make(map[string]*DailyStats)
			a.weeklyStats[character] = make(map[string]*WeeklyStats)
			a.monthlyStats[character] = make(map[string]*MonthlyStats)
			a.totalStats[character] = &TotalStats{Sources: make(map[string]SourceStats)}
		}

		// 加载每日统计
		if dailyData, ok := charData["daily"].(map[string]any); ok {
			for date, dayData := range dailyData {
				if dayMap, ok := dayData.(map[string]any); ok {
					ds := &DailyStats{
						Date:      date,
						Gain:      getInt(dayMap, "gain"),
						Consume:   getInt(dayMap, "consume"),
						NetChange: getInt(dayMap, "net_change"),
						Sources:   make(map[string]SourceStats),
					}
					if a.keepRecords {
						ds.Records = make([]DiamondRecord, 0)
					}
					if sources, ok := dayMap["sources"].(map[string]any); ok {
						for srcName, srcData := range sources {
							if srcMap, ok := srcData.(map[string]any); ok {
								ds.Sources[srcName] = SourceStats{
									Gain:    getInt(srcMap, "gain"),
									Consume: getInt(srcMap, "consume"),
								}
							}
						}
					}
					a.dailyStats[character][date] = ds
				}
			}
		}

		// 加载每周统计
		if weeklyData, ok := charData["weekly"].(map[string]any); ok {
			for week, weekData := range weeklyData {
				if weekMap, ok := weekData.(map[string]any); ok {
					ws := &WeeklyStats{
						Week:      week,
						Gain:      getInt(weekMap, "gain"),
						Consume:   getInt(weekMap, "consume"),
						NetChange: getInt(weekMap, "net_change"),
						Sources:   make(map[string]SourceStats),
					}
					if sources, ok := weekMap["sources"].(map[string]any); ok {
						for srcName, srcData := range sources {
							if srcMap, ok := srcData.(map[string]any); ok {
								ws.Sources[srcName] = SourceStats{
									Gain:    getInt(srcMap, "gain"),
									Consume: getInt(srcMap, "consume"),
								}
							}
						}
					}
					a.weeklyStats[character][week] = ws
				}
			}
		}

		// 加载每月统计
		if monthlyData, ok := charData["monthly"].(map[string]any); ok {
			for month, monthData := range monthlyData {
				if monthMap, ok := monthData.(map[string]any); ok {
					ms := &MonthlyStats{
						Month:     month,
						Gain:      getInt(monthMap, "gain"),
						Consume:   getInt(monthMap, "consume"),
						NetChange: getInt(monthMap, "net_change"),
						Sources:   make(map[string]SourceStats),
					}
					if sources, ok := monthMap["sources"].(map[string]any); ok {
						for srcName, srcData := range sources {
							if srcMap, ok := srcData.(map[string]any); ok {
								ms.Sources[srcName] = SourceStats{
									Gain:    getInt(srcMap, "gain"),
									Consume: getInt(srcMap, "consume"),
								}
							}
						}
					}
					a.monthlyStats[character][month] = ms
				}
			}
		}

		// 加载总计
		if totalData, ok := charData["total"].(map[string]any); ok {
			ts := &TotalStats{
				Gain:      getInt(totalData, "gain"),
				Consume:   getInt(totalData, "consume"),
				NetChange: getInt(totalData, "net_change"),
				Sources:   make(map[string]SourceStats),
			}
			if sources, ok := totalData["sources"].(map[string]any); ok {
				for srcName, srcData := range sources {
					if srcMap, ok := srcData.(map[string]any); ok {
						ts.Sources[srcName] = SourceStats{
							Gain:    getInt(srcMap, "gain"),
							Consume: getInt(srcMap, "consume"),
						}
					}
				}
			}
			a.totalStats[character] = ts
		}
	}
}

// ToMap 转换为输出格式
func (a *Aggregator) ToMap() map[string]map[string]any {
	stats := make(map[string]map[string]any)

	// 按角色名排序
	characterKeys := make([]string, 0, len(a.dailyStats))
	for k := range a.dailyStats {
		characterKeys = append(characterKeys, k)
	}
	sort.Strings(characterKeys)

	for _, character := range characterKeys {
		characterData := make(map[string]any)

		// 保存每日统计
		dailyData := make(map[string]*DailyStats)
		var dateKeys []string
		for k := range a.dailyStats[character] {
			dateKeys = append(dateKeys, k)
		}
		sort.Strings(dateKeys)
		for _, k := range dateKeys {
			dailyData[k] = a.dailyStats[character][k]
		}
		characterData["daily"] = dailyData

		// 保存每周统计
		weeklyData := make(map[string]*WeeklyStats)
		var weekKeys []string
		for k := range a.weeklyStats[character] {
			weekKeys = append(weekKeys, k)
		}
		sort.Strings(weekKeys)
		for _, k := range weekKeys {
			weeklyData[k] = a.weeklyStats[character][k]
		}
		characterData["weekly"] = weeklyData

		// 保存每月统计
		monthlyData := make(map[string]*MonthlyStats)
		var monthKeys []string
		for k := range a.monthlyStats[character] {
			monthKeys = append(monthKeys, k)
		}
		sort.Strings(monthKeys)
		for _, k := range monthKeys {
			monthlyData[k] = a.monthlyStats[character][k]
		}
		characterData["monthly"] = monthlyData

		// 保存总统计
		characterData["total"] = a.totalStats[character]

		stats[character] = characterData
	}

	return stats
}

// RecordCount 返回已处理的记录数
func (a *Aggregator) RecordCount() int {
	return a.recordCount
}

// 辅助函数
func getInt(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

// getWeekString 获取周字符串（年-周）
func getWeekString(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

// getMonthString 获取月字符串（年-月）
func getMonthString(t time.Time) string {
	year := t.Year()
	month := t.Month()
	return fmt.Sprintf("%d-%02d", year, month)
}

// loadExistingStats 从文件加载现有统计数据
func loadExistingStats(filePath string) map[string]map[string]any {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var stats map[string]map[string]any
	if err := json.NewDecoder(file).Decode(&stats); err != nil {
		return nil
	}
	return stats
}

// SaveStats 保存统计结果到文件
func SaveStats(stats map[string]map[string]any, logFilePath string) {
	// 确保目录存在
	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}

	// 写入文件
	file, err := os.Create(logFilePath)
	if err != nil {
		log.Fatalf("创建统计文件失败: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(stats); err != nil {
		log.Fatalf("写入统计文件失败: %v", err)
	}

	fmt.Printf("统计结果已保存到: %s\n", logFilePath)
}
