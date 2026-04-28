package repository

import (
	"context"

	"pm-work-tracker/backend/internal/model"
)

// RoleRepo defines persistence operations for Role and RolePermission entities.
type RoleRepo interface {
	List(ctx context.Context) ([]model.Role, error)
	FindByID(ctx context.Context, id uint) (*model.Role, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.Role, error)
	FindByName(ctx context.Context, name string) (*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uint) error

	// Permission bindings
	ListPermissions(ctx context.Context, roleID uint) ([]string, error)
	SetPermissions(ctx context.Context, roleID uint, codes []string) error

	// Usage count
	CountMembersByRoleID(ctx context.Context, roleID uint) (int64, error)

	// Non-team-context: check if any of user's roles has the given permission code
	HasPermission(ctx context.Context, userID uint, code string) (bool, error)

	// GetUserTeamPermissions returns teamBizKey -> permission codes for all teams the user belongs to.
	GetUserTeamPermissions(ctx context.Context, userID uint) (map[int64][]string, error)
}
