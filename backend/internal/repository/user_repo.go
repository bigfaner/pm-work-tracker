package repository

import (
	"context"

	"pm-work-tracker/backend/internal/model"
)

// UserRepo defines persistence operations for User entities.
type UserRepo interface {
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context) ([]*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
}
