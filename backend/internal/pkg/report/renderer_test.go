package report

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
)

func TestRenderMarkdown_Header(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	preview := &dto.ReportPreview{
		WeekStart: "2026-04-13",
		WeekEnd:   "2026-04-19",
		Sections:  []dto.ReportSectionDTO{},
	}

	data := RenderMarkdown(preview, monday)
	md := string(data)
	assert.Contains(t, md, "# 周报 2026-W16")
}

func TestRenderMarkdown_KeyItem(t *testing.T) {
	preview := &dto.ReportPreview{
		Sections: []dto.ReportSectionDTO{
			{
				MainItem: dto.MainItemSummaryDTO{BizKey: "1", Title: "重点项目", IsKeyItem: true},
				SubItems: []dto.ReportSubItemDTO{
					{BizKey: "10", Title: "Sub", Completion: 50, Achievements: []string{"done"}},
				},
			},
		},
	}

	data := RenderMarkdown(preview, time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))
	assert.Contains(t, string(data), "## [重点] 重点项目")
}

func TestRenderMarkdown_NoKeyItem(t *testing.T) {
	preview := &dto.ReportPreview{
		Sections: []dto.ReportSectionDTO{
			{
				MainItem: dto.MainItemSummaryDTO{BizKey: "1", Title: "普通项目"},
				SubItems: []dto.ReportSubItemDTO{
					{BizKey: "10", Title: "Sub", Completion: 50, Achievements: []string{"done"}},
				},
			},
		},
	}

	data := RenderMarkdown(preview, time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))
	assert.NotContains(t, string(data), "[重点]")
	assert.Contains(t, string(data), "## 普通项目")
}

func TestRenderMarkdown_CompletionFormatting(t *testing.T) {
	preview := &dto.ReportPreview{
		Sections: []dto.ReportSectionDTO{
			{
				MainItem: dto.MainItemSummaryDTO{BizKey: "1", Title: "M"},
				SubItems: []dto.ReportSubItemDTO{
					{BizKey: "10", Title: "S", Completion: 75.5, Achievements: []string{"done"}},
				},
			},
		},
	}

	data := RenderMarkdown(preview, time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))
	// Should format as integer percentage (76% due to %.0f)
	assert.Contains(t, string(data), "完成度: 76%")
}

func TestRenderMarkdown_NoAchievementsOrBlockers(t *testing.T) {
	preview := &dto.ReportPreview{
		Sections: []dto.ReportSectionDTO{
			{
				MainItem: dto.MainItemSummaryDTO{BizKey: "1", Title: "M"},
				SubItems: []dto.ReportSubItemDTO{
					{BizKey: "10", Title: "S", Completion: 10},
				},
			},
		},
	}

	data := RenderMarkdown(preview, time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))
	md := string(data)
	assert.NotContains(t, md, "**成果:**")
	assert.NotContains(t, md, "**卡点:**")
}

func TestRenderMarkdown_MultipleSubItems(t *testing.T) {
	preview := &dto.ReportPreview{
		Sections: []dto.ReportSectionDTO{
			{
				MainItem: dto.MainItemSummaryDTO{BizKey: "1", Title: "M"},
				SubItems: []dto.ReportSubItemDTO{
					{BizKey: "10", Title: "S1", Completion: 50, Achievements: []string{"a1"}, Blockers: []string{"b1"}},
					{BizKey: "11", Title: "S2", Completion: 80, Achievements: []string{"a2"}},
				},
			},
		},
	}

	data := RenderMarkdown(preview, time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))
	md := string(data)
	assert.Contains(t, md, "### S1")
	assert.Contains(t, md, "### S2")
	assert.Contains(t, md, "- a1")
	assert.Contains(t, md, "- b1")
	assert.Contains(t, md, "- a2")
}

func TestRenderMarkdown_EmptyPreview(t *testing.T) {
	preview := &dto.ReportPreview{
		WeekStart: "2026-04-13",
		WeekEnd:   "2026-04-19",
		Sections:  []dto.ReportSectionDTO{},
	}

	data := RenderMarkdown(preview, time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))
	require.True(t, strings.HasPrefix(string(data), "# 周报"))
}
