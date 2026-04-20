package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
)

// ChallengeAggregator 挑战聚合器
type ChallengeAggregator struct {
	stats       map[string]*ChallengeStats // character -> stats
	recordCount int
}

// NewChallengeAggregator 创建挑战聚合器
func NewChallengeAggregator() *ChallengeAggregator {
	return &ChallengeAggregator{
		stats: make(map[string]*ChallengeStats),
	}
}

// getOrCreateCharacterStats 获取或创建角色统计
func (a *ChallengeAggregator) getOrCreateCharacterStats(character string) *ChallengeStats {
	if a.stats[character] == nil {
		a.stats[character] = &ChallengeStats{
			Quest:  make(map[string]*ChallengeLevelStats),
			Towers: make(map[TowerType]map[string]*ChallengeLevelStats),
		}
	}
	return a.stats[character]
}

// getOrCreateTowerStats 获取或创建塔统计
func (a *ChallengeAggregator) getOrCreateTowerStats(charStats *ChallengeStats, towerType TowerType) map[string]*ChallengeLevelStats {
	if charStats.Towers[towerType] == nil {
		charStats.Towers[towerType] = make(map[string]*ChallengeLevelStats)
	}
	return charStats.Towers[towerType]
}

// updateLevelStats 更新关卡统计
func (a *ChallengeAggregator) updateLevelStats(levelMap map[string]*ChallengeLevelStats, level string, status ChallengeStatus, timestamp string) {
	if levelMap[level] == nil {
		levelMap[level] = &ChallengeLevelStats{Level: level, Success: false}
	}
	levelMap[level].Attempts++
	levelMap[level].LastTime = timestamp
	if status == ChallengeStatusSuccess {
		levelMap[level].Success = true
	}
}

// AddRecord 添加挑战记录
func (a *ChallengeAggregator) AddRecord(record ChallengeRecord) {
	charStats := a.getOrCreateCharacterStats(record.Character)

	switch record.Type {
	case ChallengeTypeQuest:
		a.updateLevelStats(charStats.Quest, record.Level, record.Status, record.Timestamp)
	case ChallengeTypeTower:
		towerStats := a.getOrCreateTowerStats(charStats, record.TowerType)
		a.updateLevelStats(towerStats, record.Level, record.Status, record.Timestamp)
	}
	a.recordCount++
}

// LoadExistingStats 加载现有统计数据（用于增量合并）
func (a *ChallengeAggregator) LoadExistingStats(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	var existingStats map[string]*ChallengeStats
	if err := json.NewDecoder(file).Decode(&existingStats); err != nil {
		return
	}

	for character, stats := range existingStats {
		if a.stats[character] == nil {
			a.stats[character] = &ChallengeStats{
				Quest:  make(map[string]*ChallengeLevelStats),
				Towers: make(map[TowerType]map[string]*ChallengeLevelStats),
			}
		}

		// 合并主线统计
		for level, levelStats := range stats.Quest {
			a.stats[character].Quest[level] = levelStats
		}

		// 合并塔统计
		for towerType, towerStats := range stats.Towers {
			if a.stats[character].Towers[towerType] == nil {
				a.stats[character].Towers[towerType] = make(map[string]*ChallengeLevelStats)
			}
			for level, levelStats := range towerStats {
				a.stats[character].Towers[towerType][level] = levelStats
			}
		}
	}
}

// ToMap 转换为输出格式
func (a *ChallengeAggregator) ToMap() map[string]*ChallengeStats {
	result := make(map[string]*ChallengeStats)

	// 按角色名排序
	characterKeys := make([]string, 0, len(a.stats))
	for k := range a.stats {
		characterKeys = append(characterKeys, k)
	}
	sort.Strings(characterKeys)

	for _, character := range characterKeys {
		result[character] = a.stats[character]
	}

	return result
}

// RecordCount 返回已处理的记录数
func (a *ChallengeAggregator) RecordCount() int {
	return a.recordCount
}

// SaveChallengeStats 保存挑战统计结果到文件
func SaveChallengeStats(stats map[string]*ChallengeStats, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("创建挑战统计文件失败: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(stats); err != nil {
		log.Printf("写入挑战统计文件失败: %v", err)
		return
	}

	log.Printf("挑战统计结果已保存到: %s", filePath)
}
