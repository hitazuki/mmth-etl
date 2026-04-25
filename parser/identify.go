package parser

import (
	"mmth-etl/i18n"
	"mmth-etl/types"
)

// LogType 日志类型枚举
type LogType int

const (
	LogTypeNone           LogType = iota // 未知类型
	LogTypeDiamond                       // 钻石记录
	LogTypeCave                          // 洞穴记录
	LogTypeChallenge                     // 挑战记录
	LogTypeRuneTicket                    // 饼干记录
	LogTypeUpgradePanacea                // 红水记录
	LogTypeGacha                         // 抽卡来源
	LogTypeOpen                          // 开启来源
	LogTypeSystemError                   // 系统/错误日志（清空来源上下文）
	LogTypeNameLabel                     // 以 Name: 开头但未匹配已知物品类型
)

// IsNameLabelPrefix checks if body starts with any language's Name: prefix
func IsNameLabelPrefix(body string) bool {
	for _, prefix := range i18n.GetAllNameLabels() {
		if len(body) >= len(prefix) && body[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// IdentifyLogType 识别日志类型（一次扫描确定类型）
func IdentifyLogType(body string) LogType {
	// Step 1: Check if it's an item change log (Name: prefix in any language)
	isItemChangeLog := IsNameLabelPrefix(body)

	// Step 2: If item change log, check specific item types
	if isItemChangeLog {
		if types.DiamondRegex().MatchString(body) {
			return LogTypeDiamond
		}
		if types.RuneTicketRegex().MatchString(body) {
			return LogTypeRuneTicket
		}
		if types.UpgradePanaceaRegex().MatchString(body) {
			return LogTypeUpgradePanacea
		}
		// Name: prefix but not a tracked item type
		return LogTypeNameLabel
	}

	// 检查洞穴记录
	if types.CaveEnterRegex().MatchString(body) || types.CaveFinishRegex().MatchString(body) || types.CaveErrorRegex.MatchString(body) {
		return LogTypeCave
	}

	// 检查挑战记录
	// 必须同时匹配：挑战模式（塔/任务）+ 结果关键词（成功/失败）
	// EN/TW: 以 Challenge/挑战 开头
	// JA/KO: 包含 挑戦/도전 关键词
	if types.ChallengeQuestRegex().MatchString(body) || types.ChallengeTowerRegex().MatchString(body) {
		if types.ChallengeSuccessRegex().MatchString(body) || types.ChallengeFailedRegex().MatchString(body) {
			return LogTypeChallenge
		}
	}

	// 检查系统/错误日志（清空来源上下文）
	if types.SystemErrorRegex.MatchString(body) {
		return LogTypeSystemError
	}

	// 检查来源上下文日志
	if types.GachaPrefixRegex().MatchString(body) {
		return LogTypeGacha
	}
	if types.OpenPrefixRegex().MatchString(body) {
		return LogTypeOpen
	}

	return LogTypeNone
}
