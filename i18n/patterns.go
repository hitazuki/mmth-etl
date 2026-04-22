package i18n

import (
	"fmt"
	"regexp"
)

// patternDefinition contains language-specific string values for pattern building.
type patternDefinition struct {
	// Name label (e.g., "Name" in English, "名称" in Chinese)
	NameLabel string

	// Item names
	Diamond        string
	RuneTicket     string
	UpgradePanacea string

	// Cave patterns
	CaveEnter  string
	CaveFinish string

	// Challenge patterns
	ChallengeKeyword string
	SuccessKeyword   string
	FailedKeyword    string

	// Tower names (for pattern building)
	TowerInfinity string
	TowerAzure    string
	TowerCrimson  string
	TowerEmerald  string
	TowerAmber    string
}

// languageDefinitions maps languages to their pattern definitions.
var languageDefinitions = map[Language]patternDefinition{
	LangEn: {
		NameLabel:        "Name",
		Diamond:          "Diamonds",
		RuneTicket:       "Rune Ticket",
		UpgradePanacea:   "Upgrade Panacea",
		CaveEnter:        "Enter Cave of Space-Time",
		CaveFinish:       "Cave of Space-Time Finished",
		ChallengeKeyword: "Challenge",
		SuccessKeyword:   "triumphed",
		FailedKeyword:    "failed",
		TowerInfinity:    "Infinity",
		TowerAzure:       "Azure",
		TowerCrimson:     "Crimson",
		TowerEmerald:     "Emerald",
		TowerAmber:       "Amber",
	},
	LangTw: {
		NameLabel:        "名称",
		Diamond:          "鑽石",
		RuneTicket:       "符石兌換券",
		UpgradePanacea:   "強化秘藥",
		CaveEnter:        "进入 時空洞窟",
		CaveFinish:       "時空洞窟已完成",
		ChallengeKeyword: "挑战",
		SuccessKeyword:   "勝利",
		FailedKeyword:    "敗北",
		TowerInfinity:    "無窮之塔",
		TowerAzure:       "憂藍之塔",
		TowerCrimson:     "業紅之塔",
		TowerEmerald:     "蒼翠之塔",
		TowerAmber:       "流金之塔",
	},
	LangJa: {
		NameLabel:        "名称",
		Diamond:          "ダイヤ",
		RuneTicket:       "ルーンチケット",
		UpgradePanacea:   "強化秘薬",
		CaveEnter:        "時空の洞窟に入る",
		CaveFinish:       "時空の洞窟完了",
		ChallengeKeyword: "挑戦",
		SuccessKeyword:   "勝利",
		FailedKeyword:    "敗北",
		TowerInfinity:    "無窮の塔",
		TowerAzure:       "藍の塔",
		TowerCrimson:     "紅の塔",
		TowerEmerald:     "翠の塔",
		TowerAmber:       "黄の塔",
	},
	LangKo: {
		NameLabel:        "이름",
		Diamond:          "다이아",
		RuneTicket:       "룬 티켓",
		UpgradePanacea:   "강화의 비약",
		CaveEnter:        "시공의 동굴 입장",
		CaveFinish:       "시공의 동굴 완료",
		ChallengeKeyword: "도전",
		SuccessKeyword:   "승리",
		FailedKeyword:    "패배",
		TowerInfinity:    "무한의 탑",
		TowerAzure:       "남청의 탑",
		TowerCrimson:     "홍염의 탑",
		TowerEmerald:     "비취의 탑",
		TowerAmber:       "황철의 탑",
	},
}

// buildPatternSet creates a PatternSet for the given language.
func buildPatternSet(lang Language) *PatternSet {
	def := languageDefinitions[lang]

	// Build tower name pattern (union of all tower names)
	towerPattern := fmt.Sprintf("%s|%s|%s|%s|%s",
		regexp.QuoteMeta(def.TowerInfinity),
		regexp.QuoteMeta(def.TowerAzure),
		regexp.QuoteMeta(def.TowerCrimson),
		regexp.QuoteMeta(def.TowerEmerald),
		regexp.QuoteMeta(def.TowerAmber),
	)

	return &PatternSet{
		// Item change patterns: Name: ItemName(Quality) × N
		Diamond: regexp.MustCompile(fmt.Sprintf(
			`^%s: %s\(None\) × (-?\d+)`,
			regexp.QuoteMeta(def.NameLabel),
			regexp.QuoteMeta(def.Diamond),
		)),
		RuneTicket: regexp.MustCompile(fmt.Sprintf(
			`^%s: %s\([A-Z]+\) × (-?\d+)`,
			regexp.QuoteMeta(def.NameLabel),
			regexp.QuoteMeta(def.RuneTicket),
		)),
		UpgradePanacea: regexp.MustCompile(fmt.Sprintf(
			`^%s: %s\([A-Z]+\) × (-?\d+)`,
			regexp.QuoteMeta(def.NameLabel),
			regexp.QuoteMeta(def.UpgradePanacea),
		)),

		// Cave patterns
		CaveEnter:  regexp.MustCompile(regexp.QuoteMeta(def.CaveEnter)),
		CaveFinish: regexp.MustCompile(regexp.QuoteMeta(def.CaveFinish)),

		// Challenge patterns
		// Quest: Challenge 36-13 boss (English) / 挑战 36-13 boss (Chinese)
		ChallengeQuest: regexp.MustCompile(fmt.Sprintf(
			`^%s (\d+-\d+) boss`,
			regexp.QuoteMeta(def.ChallengeKeyword),
		)),
		// Tower: Challenge Tower of Crimson 800 layer (English) / 挑战 業紅之塔 800 层一次 (Chinese)
		ChallengeTower: regexp.MustCompile(fmt.Sprintf(
			`^%s (%s) (\d+)`,
			regexp.QuoteMeta(def.ChallengeKeyword),
			towerPattern,
		)),
		ChallengeSuccess: regexp.MustCompile(regexp.QuoteMeta(def.SuccessKeyword)),
		ChallengeFailed:  regexp.MustCompile(regexp.QuoteMeta(def.FailedKeyword)),
	}
}

// TowerNameToType normalizes a language-specific tower name to a canonical tower type.
// Returns empty string if the name is not recognized.
func (m *Manager) TowerNameToType(name string) string {
	// Check all languages for the tower name
	for _, lang := range []Language{LangEn, LangTw, LangJa, LangKo} {
		def := languageDefinitions[lang]
		switch name {
		case def.TowerInfinity:
			return "Infinity"
		case def.TowerAzure:
			return "Azure"
		case def.TowerCrimson:
			return "Crimson"
		case def.TowerEmerald:
			return "Emerald"
		case def.TowerAmber:
			return "Amber"
		}
	}
	return ""
}
