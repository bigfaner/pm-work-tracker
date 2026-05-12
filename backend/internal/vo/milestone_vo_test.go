package vo

import (
	"testing"
	"time"

	"pm-work-tracker/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestNewMilestoneMapVO_PopulatesFields(t *testing.T) {
	now := time.Now()
	m := &model.MilestoneMap{
		BaseModel: model.BaseModel{ID: 1, BizKey: 123456, CreateTime: now, DbUpdateTime: now},
		TeamKey:   100,
		MapName:   "Q1 Roadmap",
		MapDesc:   "First quarter plan",
		MapStatus: "planning",
	}

	result := NewMilestoneMapVO(m)

	assert.Equal(t, "123456", result.BizKey)
	assert.Equal(t, "100", result.TeamKey)
	assert.Equal(t, "Q1 Roadmap", result.MapName)
	assert.Equal(t, "First quarter plan", result.MapDesc)
	assert.Equal(t, "planning", result.MapStatus)
	assert.Equal(t, "规划中", result.StatusName)
}

func TestNewMilestoneMapVO_UnknownStatus(t *testing.T) {
	now := time.Now()
	m := &model.MilestoneMap{
		BaseModel: model.BaseModel{ID: 2, BizKey: 999, CreateTime: now, DbUpdateTime: now},
		MapStatus: "unknown",
	}

	result := NewMilestoneMapVO(m)

	assert.Equal(t, "unknown", result.MapStatus)
	assert.Equal(t, "", result.StatusName)
}

func TestNewMilestoneVO_PopulatesFields(t *testing.T) {
	now := time.Now()
	endDate := now.Add(30 * 24 * time.Hour)
	m := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 1, BizKey: 234567, CreateTime: now, DbUpdateTime: now},
		TeamKey:         100,
		MilestoneMapKey: 123456,
		MilestoneName:   "MVP Release",
		ExpectedEndDate: &endDate,
		MilestoneStatus: "in_progress",
	}

	result := NewMilestoneVO(m)

	assert.Equal(t, "234567", result.BizKey)
	assert.Equal(t, "100", result.TeamKey)
	assert.Equal(t, "123456", result.MilestoneMapKey)
	assert.Equal(t, "MVP Release", result.MilestoneName)
	assert.NotNil(t, result.ExpectedEndDate)
	assert.Equal(t, "in_progress", result.MilestoneStatus)
	assert.Equal(t, "进行中", result.StatusName)
}

func TestNewMilestoneVO_NilEndDate(t *testing.T) {
	now := time.Now()
	m := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 3, BizKey: 345678, CreateTime: now, DbUpdateTime: now},
		MilestoneStatus: "not_started",
	}

	result := NewMilestoneVO(m)

	assert.Nil(t, result.ExpectedEndDate)
	assert.Equal(t, "未开始", result.StatusName)
}

func TestNewMilestoneVO_UnknownStatus(t *testing.T) {
	now := time.Now()
	m := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 4, CreateTime: now, DbUpdateTime: now},
		MilestoneStatus: "unknown",
	}

	result := NewMilestoneVO(m)

	assert.Equal(t, "unknown", result.MilestoneStatus)
	assert.Equal(t, "", result.StatusName)
}

func TestNewMilestoneMapVO_ComputedFieldsDefault(t *testing.T) {
	now := time.Now()
	m := &model.MilestoneMap{
		BaseModel: model.BaseModel{ID: 5, CreateTime: now, DbUpdateTime: now},
	}

	result := NewMilestoneMapVO(m)

	assert.Equal(t, 0, result.MilestoneCount)
	assert.Equal(t, float64(0), result.OverallProgress)
}

func TestNewMilestoneVO_ComputedFieldsDefault(t *testing.T) {
	now := time.Now()
	m := &model.Milestone{
		BaseModel: model.BaseModel{ID: 6, CreateTime: now, DbUpdateTime: now},
	}

	result := NewMilestoneVO(m)

	assert.Equal(t, float64(0), result.Completion)
	assert.Equal(t, int64(0), result.ItemCount)
}
