//nolint:misspell // "cancelled" is a domain status value per PRD/API contract.
package vo

import (
	"testing"
	"time"

	"pm-work-tracker/backend/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMilestoneMapVO_BasicFields(t *testing.T) {
	now := time.Now()
	m := &model.MilestoneMap{
		BaseModel:       model.BaseModel{ID: 1, BizKey: int64(100001), CreateTime: now, DbUpdateTime: now},
		TeamKey:         int64(200001),
		CreatorKey:      int64(300001),
		AssigneeKey:     int64(400001),
		MapName:         "Q3 Roadmap",
		MapDesc:         "Quarterly plan",
		MapStatus:       "planning",
		PlanStartDate:   nil,
		ExpectedEndDate: nil,
	}

	result := NewMilestoneMapVO(m)

	assert.Equal(t, "100001", result.BizKey)
	assert.Equal(t, "200001", result.TeamKey)
	assert.Equal(t, "300001", result.CreatorKey)
	assert.Equal(t, "400001", result.AssigneeKey)
	assert.Equal(t, "Q3 Roadmap", result.MapName)
	assert.Equal(t, "Quarterly plan", result.MapDesc)
	assert.Equal(t, "planning", result.MapStatus)
	assert.Equal(t, "规划中", result.StatusName)
	assert.Nil(t, result.PlanStartDate)
	assert.Nil(t, result.ExpectedEndDate)
	// Computed fields default to zero — caller enriches
	assert.Equal(t, 0, result.MilestoneCount)
	assert.Equal(t, 0, result.ItemCount)
	assert.Equal(t, float64(0), result.OverallProgress)
	assert.Equal(t, "", result.CreatorName)
	assert.Equal(t, "", result.AssigneeName)
}

func TestNewMilestoneMapVO_StatusNames(t *testing.T) {
	tests := []struct {
		status     string
		statusName string
	}{
		{"planning", "规划中"},
		{"reviewed", "已评审"},
		{"ready", "待实施"},
		{"executing", "实施中"},
		{"completed", "已完成"},
		{"cancelled", "已取消"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			m := &model.MilestoneMap{
				BaseModel: model.BaseModel{ID: 1, BizKey: 1, CreateTime: time.Now(), DbUpdateTime: time.Now()},
				MapStatus: tt.status,
			}
			result := NewMilestoneMapVO(m)
			assert.Equal(t, tt.statusName, result.StatusName)
		})
	}
}

func TestNewMilestoneMapVO_UnknownStatus(t *testing.T) {
	m := &model.MilestoneMap{
		BaseModel: model.BaseModel{ID: 1, BizKey: 1, CreateTime: time.Now(), DbUpdateTime: time.Now()},
		MapStatus: "unknown",
	}
	result := NewMilestoneMapVO(m)
	assert.Equal(t, "", result.StatusName)
}

func TestNewMilestoneMapVO_DateFields(t *testing.T) {
	planStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	m := &model.MilestoneMap{
		BaseModel:       model.BaseModel{ID: 1, BizKey: 1, CreateTime: now, DbUpdateTime: now},
		PlanStartDate:   &planStart,
		ExpectedEndDate: &expectedEnd,
	}

	result := NewMilestoneMapVO(m)

	require.NotNil(t, result.PlanStartDate)
	assert.Equal(t, "2026-07-01", *result.PlanStartDate)
	require.NotNil(t, result.ExpectedEndDate)
	assert.Equal(t, "2026-09-30", *result.ExpectedEndDate)
}

func TestNewMilestoneVO_BasicFields(t *testing.T) {
	now := time.Now()
	m := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 1, BizKey: int64(500001), CreateTime: now, DbUpdateTime: now},
		TeamKey:         int64(200001),
		MilestoneMapKey: int64(100001),
		MilestoneName:   "Phase 1",
		MilestoneDesc:   "First phase",
		MilestoneStatus: "not_started",
	}

	result := NewMilestoneVO(m)

	assert.Equal(t, "500001", result.BizKey)
	assert.Equal(t, "200001", result.TeamKey)
	assert.Equal(t, "100001", result.MilestoneMapKey)
	assert.Equal(t, "Phase 1", result.MilestoneName)
	assert.Equal(t, "First phase", result.MilestoneDesc)
	assert.Equal(t, "not_started", result.MilestoneStatus)
	assert.Equal(t, "未开始", result.StatusName)
	// Computed fields default to zero — caller enriches
	assert.Equal(t, float64(0), result.Completion)
	assert.Equal(t, 0, result.RelatedMICount)
}

func TestNewMilestoneVO_StatusNames(t *testing.T) {
	tests := []struct {
		status     string
		statusName string
	}{
		{"not_started", "未开始"},
		{"in_progress", "进行中"},
		{"completed", "已完成"},
		{"cancelled", "已取消"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			m := &model.Milestone{
				BaseModel:       model.BaseModel{ID: 1, BizKey: 1, CreateTime: time.Now(), DbUpdateTime: time.Now()},
				MilestoneStatus: tt.status,
			}
			result := NewMilestoneVO(m)
			assert.Equal(t, tt.statusName, result.StatusName)
		})
	}
}

func TestNewMilestoneVO_UnknownStatus(t *testing.T) {
	m := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 1, BizKey: 1, CreateTime: time.Now(), DbUpdateTime: time.Now()},
		MilestoneStatus: "unknown",
	}
	result := NewMilestoneVO(m)
	assert.Equal(t, "", result.StatusName)
}

func TestNewMilestoneVO_DateField(t *testing.T) {
	expectedEnd := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	m := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 1, BizKey: 1, CreateTime: now, DbUpdateTime: now},
		ExpectedEndDate: &expectedEnd,
	}

	result := NewMilestoneVO(m)

	require.NotNil(t, result.ExpectedEndDate)
	assert.Equal(t, "2026-08-15", *result.ExpectedEndDate)
}

func TestNewMilestoneVO_NilEndDate(t *testing.T) {
	m := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 1, BizKey: 1, CreateTime: time.Now(), DbUpdateTime: time.Now()},
		ExpectedEndDate: nil,
	}

	result := NewMilestoneVO(m)
	assert.Nil(t, result.ExpectedEndDate)
}
