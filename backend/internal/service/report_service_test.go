package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Tests: ReportService.Preview
// ---------------------------------------------------------------------------

func TestReportService_Preview_NoProgressRecords_ReturnsNoData(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Completion: 0},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", ItemStatus: "pending", Completion: 0},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	_, err := svc.Preview(context.Background(), 1, monday)
	assert.ErrorIs(t, err, apperrors.ErrNoData)
}

func TestReportService_Preview_WithProgressRecords(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Completion: 60},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", ItemStatus: "progressing", Completion: 60},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{
				ID:          100,
				SubItemKey:  10,
				TeamKey: 1,
				Completion:  60,
				Achievement: "完成了前端开发",
				Blocker:     "",
				CreateTime:  time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	result, err := svc.Preview(context.Background(), 1, monday)
	require.NoError(t, err)

	assert.Equal(t, "2026-04-13", result.WeekStart)
	assert.Equal(t, "2026-04-19", result.WeekEnd)
	require.Len(t, result.Sections, 1)
	section := result.Sections[0]
	assert.Equal(t, "1", section.MainItem.BizKey)
	assert.Equal(t, "Main 1", section.MainItem.Title)
	assert.Equal(t, 60.0, section.MainItem.Completion)

	require.Len(t, section.SubItems, 1)
	sub := section.SubItems[0]
	assert.Equal(t, "10", sub.BizKey)
	assert.Equal(t, "Sub A", sub.Title)
	assert.Equal(t, 60.0, sub.Completion)
	require.Len(t, sub.Achievements, 1)
	assert.Equal(t, "完成了前端开发", sub.Achievements[0])
	assert.Empty(t, sub.Blockers)
}

func TestReportService_Preview_AchievementsAndBlockers(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Completion: 40},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", ItemStatus: "progressing", Completion: 40},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{
				ID:          100,
				SubItemKey:  10,
				TeamKey: 1,
				Completion:  20,
				Achievement: "完成了设计",
				Blocker:     "接口未就绪",
				CreateTime:  time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC),
			},
			{
				ID:          101,
				SubItemKey:  10,
				TeamKey: 1,
				Completion:  40,
				Achievement: "完成了前端页面",
				Blocker:     "",
				CreateTime:  time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	result, err := svc.Preview(context.Background(), 1, monday)
	require.NoError(t, err)

	require.Len(t, result.Sections, 1)
	sub := result.Sections[0].SubItems[0]
	// Only non-empty achievement strings
	require.Len(t, sub.Achievements, 2)
	assert.Equal(t, "完成了设计", sub.Achievements[0])
	assert.Equal(t, "完成了前端页面", sub.Achievements[1])
	// Only non-empty blocker strings
	require.Len(t, sub.Blockers, 1)
	assert.Equal(t, "接口未就绪", sub.Blockers[0])
}

func TestReportService_Preview_GroupsByMainItem(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Completion: 50},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Title: "Main 2", Completion: 30},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", Completion: 50},
			{BaseModel: model.BaseModel{ID: 20, BizKey: 20}, TeamKey: 1, MainItemKey: int64(2), Title: "Sub B", Completion: 30},
			{BaseModel: model.BaseModel{ID: 21, BizKey: 21}, TeamKey: 1, MainItemKey: int64(2), Title: "Sub C", Completion: 30},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 50, Achievement: "A done", CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
			{ID: 200, SubItemKey: 20, TeamKey: 1, Completion: 30, Achievement: "B done", CreateTime: time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)},
			{ID: 201, SubItemKey: 21, TeamKey: 1, Completion: 30, Achievement: "C done", CreateTime: time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	result, err := svc.Preview(context.Background(), 1, monday)
	require.NoError(t, err)

	require.Len(t, result.Sections, 2)
	// Main 1 has 1 sub-item
	assert.Equal(t, "1", result.Sections[0].MainItem.BizKey)
	require.Len(t, result.Sections[0].SubItems, 1)
	// Main 2 has 2 sub-items
	assert.Equal(t, "2", result.Sections[1].MainItem.BizKey)
	require.Len(t, result.Sections[1].SubItems, 2)
}

func TestReportService_Preview_SubItemNoProgressThisWeek_NotIncluded(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", Completion: 50},
			{BaseModel: model.BaseModel{ID: 11, BizKey: 11}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub B", Completion: 0},
		},
	}
	// Only Sub A has progress this week; Sub B has none
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 50, Achievement: "A done", CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	result, err := svc.Preview(context.Background(), 1, monday)
	require.NoError(t, err)

	require.Len(t, result.Sections, 1)
	// Only sub-items with progress records this week should appear
	require.Len(t, result.Sections[0].SubItems, 1)
	assert.Equal(t, "10", result.Sections[0].SubItems[0].BizKey)
}

func TestReportService_Preview_MainItemWithNoSubItemProgress_Omitted(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Completion: 50},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Title: "Main 2", Completion: 0},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", Completion: 50},
			{BaseModel: model.BaseModel{ID: 20, BizKey: 20}, TeamKey: 1, MainItemKey: int64(2), Title: "Sub B", Completion: 0},
		},
	}
	// Only Sub A (MainItem 1) has progress
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 50, Achievement: "A done", CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	result, err := svc.Preview(context.Background(), 1, monday)
	require.NoError(t, err)

	// Main 2 has no sub-items with progress, so it should be omitted from sections
	require.Len(t, result.Sections, 1)
	assert.Equal(t, "1", result.Sections[0].MainItem.BizKey)
}

func TestReportService_Preview_RepoError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{listErr: errors.New("db error")}
	svc := NewReportService(mainRepo, &mockViewSubItemRepo{}, &mockViewProgressRepo{})
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	_, err := svc.Preview(context.Background(), 1, monday)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: ReportService.ExportMarkdown
// ---------------------------------------------------------------------------

func TestReportService_ExportMarkdown_NoData(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	svc := NewReportService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	_, err := svc.ExportMarkdown(context.Background(), 1, monday)
	assert.ErrorIs(t, err, apperrors.ErrNoData)
}

func TestReportService_ExportMarkdown_Format(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "主事项A", Completion: 60, IsKeyItem: true},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "子事项A1", Completion: 60},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{
				ID:          100,
				SubItemKey:  10,
				TeamKey: 1,
				Completion:  60,
				Achievement: "完成了前端页面",
				Blocker:     "接口未就绪",
				CreateTime:  time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	data, err := svc.ExportMarkdown(context.Background(), 1, monday)
	require.NoError(t, err)

	md := string(data)

	// Check header
	assert.Contains(t, md, "# 周报 2026-W16")

	// Check key item prefix
	assert.Contains(t, md, "## [重点] 主事项A")

	// Check sub-item section
	assert.Contains(t, md, "### 子事项A1")
	assert.Contains(t, md, "完成度: 60%")

	// Check achievements and blockers as bullet lists
	assert.Contains(t, md, "- 完成了前端页面")
	assert.Contains(t, md, "- 接口未就绪")
}

func TestReportService_ExportMarkdown_NoKeyItem_NoPrefix(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "普通事项", Completion: 30, IsKeyItem: false},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "子事项", Completion: 30},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 30, Achievement: "进展中", CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	data, err := svc.ExportMarkdown(context.Background(), 1, monday)
	require.NoError(t, err)

	md := string(data)
	assert.NotContains(t, md, "[重点]")
	assert.Contains(t, md, "## 普通事项")
}

func TestReportService_ExportMarkdown_EmptyAchievementsAndBlockers(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Completion: 10},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", Completion: 10},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 10, Achievement: "", Blocker: "", CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	data, err := svc.ExportMarkdown(context.Background(), 1, monday)
	require.NoError(t, err)

	md := string(data)
	assert.Contains(t, md, "### Sub A")
	// No bullet list items for empty achievements/blockers
	assert.NotContains(t, md, "**成果:**")
	assert.NotContains(t, md, "**卡点:**")
}

func TestReportService_ExportMarkdown_MultipleSections(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Completion: 50},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Title: "Main 2", Completion: 80, IsKeyItem: true},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", Completion: 50},
			{BaseModel: model.BaseModel{ID: 20, BizKey: 20}, TeamKey: 1, MainItemKey: int64(2), Title: "Sub B", Completion: 80},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 50, Achievement: "A done", CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
			{ID: 200, SubItemKey: 20, TeamKey: 1, Completion: 80, Achievement: "B done", CreateTime: time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	data, err := svc.ExportMarkdown(context.Background(), 1, monday)
	require.NoError(t, err)

	md := string(data)
	// Both main items should appear
	assert.Contains(t, md, "## Main 1")
	assert.Contains(t, md, "## [重点] Main 2")
	assert.Contains(t, md, "### Sub A")
	assert.Contains(t, md, "### Sub B")
}

func TestReportService_ExportMarkdown_FilenameFormat(t *testing.T) {
	// Test the week number calculation: 2026-04-13 is ISO week 16
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub", Completion: 50},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 50, Achievement: "done", CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	data, err := svc.ExportMarkdown(context.Background(), 1, monday)
	require.NoError(t, err)

	md := string(data)
	// Header should contain year and ISO week number
	assert.True(t, strings.Contains(md, "# 周报 2026-W16"), "expected header with year-week, got markdown starting with: %s", md[:50])
}

func TestReportService_ExportMarkdown_BlockersSectionLabel(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Completion: 30},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", Completion: 30},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 30, Achievement: "设计完成", Blocker: "等待审批", CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
		},
	}

	svc := NewReportService(mainRepo, subRepo, progressRepo)
	data, err := svc.ExportMarkdown(context.Background(), 1, monday)
	require.NoError(t, err)

	md := string(data)
	assert.Contains(t, md, "**成果:**")
	assert.Contains(t, md, "**卡点:**")
}
