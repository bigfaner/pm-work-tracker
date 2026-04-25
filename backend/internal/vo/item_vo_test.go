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
		BaseModel:   model.BaseModel{ID: 1, CreateTime: now, DbUpdateTime: now},
		TeamKey: 1,
		MainItemKey: int64(10),
		Title:       "Test SubItem",
		ItemDesc:    "desc",
		Priority:    "P0",
		ItemStatus: "progressing",
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
		BaseModel:  model.BaseModel{ID: 2, CreateTime: now, DbUpdateTime: now},
		ItemStatus: "unknown_status",
	}

	result := NewSubItemVO(subItem)

	assert.Equal(t, "unknown_status", result.Status)
	assert.Equal(t, "", result.StatusName)
}

func TestNewSubItemVO_CodePropagated(t *testing.T) {
	now := time.Now()
	subItem := &model.SubItem{
		BaseModel:   model.BaseModel{ID: 3, CreateTime: now, DbUpdateTime: now},
		Code:        "ALPHA-00001-02",
		MainItemKey: int64(1),
		ItemStatus: "pending",
	}

	result := NewSubItemVO(subItem)

	assert.Equal(t, "ALPHA-00001-02", result.Code)
}

func TestNewSubItemSummaryVOs_CodePropagated(t *testing.T) {
	now := time.Now()
	items := []*model.SubItem{
		{BaseModel: model.BaseModel{ID: 1, CreateTime: now, DbUpdateTime: now}, Code: "TEAM1-00001-01", ItemStatus: "pending"},
		{BaseModel: model.BaseModel{ID: 2, CreateTime: now, DbUpdateTime: now}, Code: "TEAM1-00001-02", ItemStatus: "progressing"},
	}

	result := NewSubItemSummaryVOs(items)

	require.Len(t, result, 2)
	assert.Equal(t, "TEAM1-00001-01", result[0].Code)
	assert.Equal(t, "TEAM1-00001-02", result[1].Code)
}
