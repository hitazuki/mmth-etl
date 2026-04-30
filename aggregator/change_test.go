package aggregator

import (
	"fmt"
	"mmth-etl/i18n"
	"mmth-etl/types"
	"testing"
)

func TestRewardMissionCompositeSourcesAggregateByTextResourceID(t *testing.T) {
	aggregator := NewChangeAggregator()

	aggregator.AddRecord(types.ChangeRecord{
		Character: "test",
		Timestamp: "2026-04-30T10:00:00+08:00",
		Amount:    20,
		Source:    "id:23214000020",
		SourceID:  int(i18n.RewardMissionSourceID(i18n.MissionGroupDailyID, 20)),
	})
	aggregator.AddRecord(types.ChangeRecord{
		Character: "test",
		Timestamp: "2026-04-30T11:00:00+08:00",
		Amount:    60,
		Source:    "id:23214000060",
		SourceID:  int(i18n.RewardMissionSourceID(i18n.MissionGroupDailyID, 60)),
	})

	aggregateID := int(i18n.MissionGroupDailyID) * int(i18n.RewardMissionCompositeFactor)
	sourceKey := fmt.Sprintf("id:%d", aggregateID)
	sources := aggregator.totalStats["test"].Sources
	if len(sources) != 1 {
		t.Fatalf("len(sources) = %d, want 1: %#v", len(sources), sources)
	}

	source := sources[sourceKey]
	if source.Gain != 80 {
		t.Fatalf("source.Gain = %d, want 80", source.Gain)
	}
	if source.SourceID != aggregateID {
		t.Fatalf("source.SourceID = %d, want %d", source.SourceID, aggregateID)
	}
}
