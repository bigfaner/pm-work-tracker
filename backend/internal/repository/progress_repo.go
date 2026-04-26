package repository

import (
	"context"
	"time"

	"pm-work-tracker/backend/internal/model"
)

// ProgressRepo defines persistence operations for ProgressRecord entities.
// Records are append-only: no Delete, only Create and PM completion correction.
type ProgressRepo interface {
	Create(ctx context.Context, record *model.ProgressRecord) error
	FindByID(ctx context.Context, id uint) (*model.ProgressRecord, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.ProgressRecord, error)
	ListBySubItem(ctx context.Context, teamID uint, subItemID uint) ([]model.ProgressRecord, error)
	LatestBySubItem(ctx context.Context, subItemID uint) (*model.ProgressRecord, error)
	UpdateCompletion(ctx context.Context, recordID uint, completion float64) error
	ListByTeamInRange(ctx context.Context, teamID uint, start, end time.Time) ([]model.ProgressRecord, error)
}
