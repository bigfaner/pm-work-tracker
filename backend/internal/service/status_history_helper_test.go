package service

import (
	"context"
	"testing"

	"pm-work-tracker/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

type mockStatusHistoryRecorder struct {
	recordFn func(ctx context.Context, history *model.StatusHistory) error
	called   bool
	lastArg  *model.StatusHistory
}

func (m *mockStatusHistoryRecorder) Record(ctx context.Context, history *model.StatusHistory) error {
	m.called = true
	m.lastArg = history
	if m.recordFn != nil {
		return m.recordFn(ctx, history)
	}
	return nil
}

func TestRecordStatusChange_NilRecorder(t *testing.T) {
	err := RecordStatusChange(nil, context.Background(), "main_item", 1, "pending", "progressing", 1, 0, "")
	assert.NoError(t, err)
}

func TestRecordStatusChange_WithRecorder(t *testing.T) {
	recorder := &mockStatusHistoryRecorder{}
	err := RecordStatusChange(recorder, context.Background(), "sub_item", 42, "progressing", "completed", 5, 0, "")
	assert.NoError(t, err)
	assert.True(t, recorder.called)
	assert.Equal(t, "sub_item", recorder.lastArg.ItemType)
	assert.Equal(t, int64(42), recorder.lastArg.ItemKey)
	assert.Equal(t, "progressing", recorder.lastArg.FromStatus)
	assert.Equal(t, "completed", recorder.lastArg.ToStatus)
	assert.Equal(t, int64(5), recorder.lastArg.ChangedBy)
	assert.Equal(t, 0, recorder.lastArg.IsAuto)
}

func TestRecordStatusChange_WithRemark(t *testing.T) {
	recorder := &mockStatusHistoryRecorder{}
	err := RecordStatusChange(recorder, context.Background(), "main_item", 1, "pending", "blocked", 1, 1, "auto-transition failed")
	assert.NoError(t, err)
	assert.Equal(t, "auto-transition failed", recorder.lastArg.Remark)
	assert.Equal(t, 1, recorder.lastArg.IsAuto)
}

func TestRecordStatusChange_RecorderError(t *testing.T) {
	recorder := &mockStatusHistoryRecorder{
		recordFn: func(_ context.Context, _ *model.StatusHistory) error {
			return assert.AnError
		},
	}
	err := RecordStatusChange(recorder, context.Background(), "main_item", 1, "pending", "progressing", 1, 0, "")
	assert.ErrorIs(t, err, assert.AnError)
}
