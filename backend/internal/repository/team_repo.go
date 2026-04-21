package repository

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
)

// TeamRepo defines persistence operations for Team and TeamMember entities.
type TeamRepo interface {
	Create(ctx context.Context, team *model.Team) error
	FindByID(ctx context.Context, teamID uint) (*model.Team, error)
	List(ctx context.Context) ([]*model.Team, error)
	Update(ctx context.Context, team *model.Team) error
	Delete(ctx context.Context, teamID uint) error

	// TeamMember operations
	AddMember(ctx context.Context, member *model.TeamMember) error
	RemoveMember(ctx context.Context, teamID, userID uint) error
	FindMember(ctx context.Context, teamID, userID uint) (*model.TeamMember, error)
	ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error)
	UpdateMember(ctx context.Context, member *model.TeamMember) error
	FindPMMembers(ctx context.Context, teamIDs []uint) (map[uint]string, error)

	// Admin operations
	ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error)
	FindTeamsByUserIDs(ctx context.Context, userIDs []uint) (map[uint][]dto.TeamSummary, error)
}
