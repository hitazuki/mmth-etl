package parser

import (
	"mmth-etl/i18n"
	"mmth-etl/types"
	"strings"
)

// sourceTableCache holds the built lookup tables for a specific language.
type sourceTableCache struct {
	lang           i18n.Language
	entries        []i18n.SourceEntry
	lookup         map[string]i18n.SourceEntry
	rewardPatterns []i18n.RewardMissionPattern
}

var currentCache *sourceTableCache

func buildSourceCache(mgr *i18n.Manager, lang i18n.Language) *sourceTableCache {
	entries := mgr.GetSources(lang)
	patterns := mgr.GetRewardMissionPatterns(lang)

	lookup := make(map[string]i18n.SourceEntry, len(entries))
	for _, entry := range entries {
		lookup[entry.Text] = entry
	}

	return &sourceTableCache{
		lang:           lang,
		entries:        entries,
		lookup:         lookup,
		rewardPatterns: patterns,
	}
}

func ensureCache() {
	mgr := types.GetI18nManager()
	currentLang := mgr.CurrentLanguage()

	if currentCache == nil || currentCache.lang != currentLang {
		currentCache = buildSourceCache(mgr, currentLang)
	}
}

// findSource finds a source by exact or prefix match.
// Returns the entry and true if found.
func findSource(source string) (i18n.SourceEntry, bool) {
	ensureCache()

	// Exact match
	if entry, ok := currentCache.lookup[source]; ok {
		return entry, true
	}

	// Prefix match
	for _, entry := range currentCache.entries {
		if strings.HasPrefix(source, entry.Text) {
			return entry, true
		}
	}

	// Reward mission patterns
	for _, pattern := range currentCache.rewardPatterns {
		if strings.HasPrefix(source, pattern.Prefix) {
			return i18n.SourceEntry{
				ID:    pattern.SourceID,
				Alias: pattern.Alias,
				Text:  pattern.Prefix,
			}, true
		}
	}

	return i18n.SourceEntry{}, false
}

// GetSourceID returns the source ID for a given source string.
func GetSourceID(source string) i18n.SourceID {
	if entry, ok := findSource(source); ok {
		return entry.ID
	}
	return 0
}

// GetSourceAlias returns a friendly alias for a source string.
func GetSourceAlias(source string) string {
	if entry, ok := findSource(source); ok {
		return entry.Alias
	}
	return source
}

// MapSourceWithID maps source to alias and returns the source ID.
func MapSourceWithID(source string) (alias string, sourceID i18n.SourceID) {
	// Gacha: extract gacha name
	if gachaAlias, ok := extractGacha(source); ok {
		return gachaAlias, i18n.SourceIDGacha
	}

	// Open: extract item name
	if openAlias, ok := extractOpen(source); ok {
		return openAlias, i18n.SourceIDOpen
	}

	// Check mapping table
	if entry, ok := findSource(source); ok {
		return entry.Alias, entry.ID
	}

	return source, 0
}

// extractGacha extracts gacha name if source is a gacha log.
// Returns "Gacha <name>" format, with quantity suffix removed.
// Pattern: "抽卡 <name> <count> 次" -> "抽卡 <name>"
// Cuts at the second space in the full string (first space is after prefix).
func extractGacha(source string) (string, bool) {
	mgr := types.GetI18nManager()
	prefix := mgr.GetCurrentGachaPrefix()

	if strings.HasPrefix(source, prefix) {
		// Count spaces in full string, cut at second space
		// "抽卡 黒葬武具ガチャ 5 次" -> second space is before "5"
		spaceCount := 0
		for i, r := range source {
			if r == ' ' {
				spaceCount++
				if spaceCount == 2 {
					return strings.TrimSpace(source[:i]), true
				}
			}
		}
		// No second space found, return full source
		return strings.TrimSpace(source), true
	}
	return "", false
}

// extractOpen extracts item name if source is an open log.
// Returns "Open <name>" format, with quantity suffix removed.
// Pattern: "開啟 <name> x 5" -> "開啟 <name>"
// Cuts at " x" pattern.
func extractOpen(source string) (string, bool) {
	mgr := types.GetI18nManager()
	prefix := mgr.GetCurrentOpenPrefix()

	if strings.HasPrefix(source, prefix) {
		content := source[len(prefix):]
		// Find " x" to cut before the quantity
		// "上級封印寶箱 x 5" -> "上級封印寶箱"
		if idx := strings.Index(content, " x"); idx != -1 {
			return prefix + strings.TrimSpace(content[:idx]), true
		}
		return prefix + strings.TrimSpace(content), true
	}
	return "", false
}

// InvalidateSourceCache clears the source cache.
func InvalidateSourceCache() {
	currentCache = nil
}
