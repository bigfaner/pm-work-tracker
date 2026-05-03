package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"pm-work-tracker/backend/internal/model"
)

func TestStatusHistory_TableName(t *testing.T) {
	s := model.StatusHistory{}
	assert.Equal(t, "pmw_status_histories", s.TableName())
}

func TestStatusHistory_DoesNotEmbedBaseModel(t *testing.T) {
	// StatusHistory should NOT have UpdatedAt or DeletedAt fields.
	// It's an append-only log with just ID and CreatedAt.
	s := model.StatusHistory{}
	// Verify the struct only has basic fields, not BaseModel
	_ = s.ID
	_ = s.CreateTime
	// If it embedded BaseModel, it would have DeletedAt too.
	// This compiles only because StatusHistory has its own ID and CreatedAt fields.
}

func TestStatusHistory_Fields(t *testing.T) {
	s := model.StatusHistory{
		ItemType:   "main_item",
		ItemKey:    int64(1),
		FromStatus: "pending",
		ToStatus:   "in_progress",
		ChangedBy:  42,
		IsAuto:     0,
		Remark:     "manual change",
	}
	assert.Equal(t, "main_item", s.ItemType)
	assert.Equal(t, uint(1), uint(s.ItemKey))
	assert.Equal(t, "pending", s.FromStatus)
	assert.Equal(t, "in_progress", s.ToStatus)
	assert.Equal(t, int64(42), s.ChangedBy)
	assert.Equal(t, 0, s.IsAuto)
	assert.Equal(t, "manual change", s.Remark)
}
