// Package i18n provides internationalization support for log parsing.
package i18n

import (
	"regexp"
)

// Language represents a supported language for log parsing.
type Language string

const (
	// LangEn is English.
	LangEn Language = "en"
	// LangTw is Traditional Chinese.
	LangTw Language = "tw"
	// LangJa is Japanese.
	LangJa Language = "ja"
	// LangKo is Korean.
	LangKo Language = "ko"
)

// PatternSet contains all regex patterns for a specific language.
type PatternSet struct {
	// Item change patterns
	Diamond        *regexp.Regexp
	RuneTicket     *regexp.Regexp
	UpgradePanacea *regexp.Regexp

	// Cave patterns
	CaveEnter  *regexp.Regexp
	CaveFinish *regexp.Regexp

	// Challenge patterns
	ChallengeQuest   *regexp.Regexp
	ChallengeTower   *regexp.Regexp
	ChallengeSuccess *regexp.Regexp
	ChallengeFailed  *regexp.Regexp
}

// Manager manages multi-language patterns for log parsing.
type Manager struct {
	currentLang Language
	patterns    map[Language]*PatternSet
}

// NewManager creates a new i18n manager with patterns for all supported languages.
func NewManager() *Manager {
	m := &Manager{
		patterns: make(map[Language]*PatternSet),
	}

	// Pre-compile patterns for all languages
	for _, lang := range []Language{LangEn, LangTw, LangJa, LangKo} {
		m.patterns[lang] = buildPatternSet(lang)
	}

	// Default to English
	m.currentLang = LangEn

	return m
}

// SetLanguage sets the current language for pattern matching.
func (m *Manager) SetLanguage(lang Language) {
	m.currentLang = lang
}

// CurrentLanguage returns the current language.
func (m *Manager) CurrentLanguage() Language {
	return m.currentLang
}

// CurrentPatterns returns the pattern set for the current language.
func (m *Manager) CurrentPatterns() *PatternSet {
	return m.patterns[m.currentLang]
}

// GetPatterns returns the pattern set for a specific language.
func (m *Manager) GetPatterns(lang Language) *PatternSet {
	if patterns, ok := m.patterns[lang]; ok {
		return patterns
	}
	return m.patterns[LangEn]
}

// IsNameLabelLine checks if the body is an item change log (starts with Name: label).
// Item change logs cannot be used as source context.
func (m *Manager) IsNameLabelLine(body string) bool {
	def := languageDefinitions[m.currentLang]
	prefix := def.NameLabel + ":"
	return len(body) >= len(prefix) && body[:len(prefix)] == prefix
}

// IsChallengeLine checks if the body is a challenge log (starts with Challenge keyword).
// Challenge logs cannot be used as source context.
func (m *Manager) IsChallengeLine(body string) bool {
	def := languageDefinitions[m.currentLang]
	prefix := def.ChallengeKeyword + " "
	return len(body) >= len(prefix) && body[:len(prefix)] == prefix
}

// GetVictoryPhrases returns all victory phrases in all languages.
// Used for source mapping (e.g., "You have triumphed." in English).
func (m *Manager) GetVictoryPhrases() map[string]string {
	result := make(map[string]string)
	for _, lang := range []Language{LangEn, LangTw, LangJa, LangKo} {
		def := languageDefinitions[lang]
		result[def.SuccessKeyword] = def.SuccessKeyword
	}
	return result
}

// GetAllGachaPrefixes returns all Gacha prefixes for matching.
// EN/JA: Gacha, TW/CN: 抽卡, KO: 가챠
func (m *Manager) GetAllGachaPrefixes() []string {
	return []string{"Gacha ", "抽卡 ", "가챠 "}
}

// GetAllOpenPrefixes returns all Open prefixes for matching.
// EN: Open, TW: 開啟, CN: 开启, JA: 開く, KO: 열기
func (m *Manager) GetAllOpenPrefixes() []string {
	return []string{"Open ", "開啟 ", "开启 ", "開く ", "열기 "}
}
