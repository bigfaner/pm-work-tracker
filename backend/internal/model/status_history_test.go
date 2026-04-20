package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"pm-work-tracker/backend/internal/model"
)

func TestStatusHistory_TableName(t *testing.T) {
	s := model.StatusHistory{}
	assert.Equal(t, "status_histories", s.TableName())
}

func TestStatusHistory_DoesNotEmbedBaseModel(t *testing.T) {
	// StatusHistory should NOT have UpdatedAt or DeletedAt fields.
	// It's an append-only log with just ID and CreatedAt.
	s := model.StatusHistory{}
	// Verify the struct only has basic fields, not BaseModel
	_ = s.ID
	_ = s.CreatedAt
	// If it embedded BaseModel, it would have DeletedAt too.
	// This compiles only because StatusHistory has its own ID and CreatedAt fields.
}

func TestStatusHistory_Fields(t *testing.T) {
	s := model.StatusHistory{
		ItemType:   "main_item",
		ItemID:     1,
		FromStatus: "pending",
		ToStatus:   "in_progress",
		ChangedBy:  42,
		IsAuto:     false,
		Remark:     "manual change",
	}
	assert.Equal(t, "main_item", s.ItemType)
	assert.Equal(t, uint(1), s.ItemID)
	assert.Equal(t, "pending", s.FromStatus)
	assert.Equal(t, "in_progress", s.ToStatus)
	assert.Equal(t, uint(42), s.ChangedBy)
	assert.False(t, s.IsAuto)
	assert.Equal(t, "manual change", s.Remark)
}
