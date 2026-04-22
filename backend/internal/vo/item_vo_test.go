package vo

import (
	"testing"
	"time"

	"pm-work-tracker/backend/internal/model"

	"github.com/stretchr/testify/assert"
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
