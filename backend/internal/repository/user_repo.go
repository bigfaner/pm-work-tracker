package repository

import (
	"context"

	"pm-work-tracker/backend/internal/model"
)

// UserRepo defines persistence operations for User entities.
type UserRepo interface {
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.User, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context) ([]*model.User, error)
	ListFiltered(ctx context.Context, search string, offset, limit int) ([]*model.User, int64, error)
	SearchAvailable(ctx context.Context, teamBizKey int64, search string, limit int) ([]*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	SoftDelete(ctx context.Context, user *model.User) error
}
