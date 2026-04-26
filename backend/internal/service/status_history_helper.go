package service

import (
	"context"

	"pm-work-tracker/backend/internal/model"
)

// StatusHistoryRecorder records status change history entries.
type StatusHistoryRecorder interface {
	Record(ctx context.Context, history *model.StatusHistory) error
}

// RecordStatusChange creates a StatusHistory record if the recorder is non-nil.
// Returns the error from the recorder so callers can propagate DB failures.
func RecordStatusChange(
	recorder StatusHistoryRecorder,
	ctx context.Context,
	itemType string,
	itemKey int64,
	fromStatus, toStatus string,
	changedBy uint,
	isAuto int,
	remark string,
) error {
	if recorder == nil {
		return nil
	}
	return recorder.Record(ctx, &model.StatusHistory{
		ItemType:   itemType,
		ItemKey:    itemKey,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		ChangedBy:  int64(changedBy),
		IsAuto:     isAuto,
		Remark:     remark,
	})
}
