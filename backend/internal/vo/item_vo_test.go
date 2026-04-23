package vo

import (
	"testing"
	"time"

	"pm-work-tracker/backend/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubItemVO_PopulatesStatusName(t *testing.T) {
	now := time.Now()
	subItem := &model.SubItem{
		BaseModel:   model.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
		TeamID:      1,
		MainItemID:  10,
		Title:       "Test SubItem",
		Description: "desc",
		Priority:    "P0",
		Status:      "progressing",
		Completion:  50.0,
		IsKeyItem:   false,
		Weight:      1.0,
	}

	result := NewSubItemVO(subItem)

	assert.Equal(t, "progressing", result.Status)
	assert.Equal(t, "进行中", result.StatusName)
}

func TestNewSubItemVO_UnknownStatus(t *testing.T) {
	now := time.Now()
	subItem := &model.SubItem{
		BaseModel: model.BaseModel{ID: 2, CreatedAt: now, UpdatedAt: now},
		Status:    "unknown_status",
	}

	result := NewSubItemVO(subItem)

	assert.Equal(t, "unknown_status", result.Status)
	assert.Equal(t, "", result.StatusName)
}

func TestNewSubItemVO_CodePropagated(t *testing.T) {
	now := time.Now()
	subItem := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 3, CreatedAt: now, UpdatedAt: now},
		Code:       "ALPHA-00001-02",
		MainItemID: 1,
		Status:     "pending",
	}

	result := NewSubItemVO(subItem)

	assert.Equal(t, "ALPHA-00001-02", result.Code)
}

func TestNewSubItemSummaryVOs_CodePropagated(t *testing.T) {
	now := time.Now()
	items := []*model.SubItem{
		{BaseModel: model.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now}, Code: "TEAM1-00001-01", Status: "pending"},
		{BaseModel: model.BaseModel{ID: 2, CreatedAt: now, UpdatedAt: now}, Code: "TEAM1-00001-02", Status: "progressing"},
	}

	result := NewSubItemSummaryVOs(items)

	require.Len(t, result, 2)
	assert.Equal(t, "TEAM1-00001-01", result[0].Code)
	assert.Equal(t, "TEAM1-00001-02", result[1].Code)
}
