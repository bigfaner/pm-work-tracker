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
	FindByBizKey(ctx context.Context, bizKey int64) (*model.Team, error)
	List(ctx context.Context) ([]*model.Team, error)
	Update(ctx context.Context, team *model.Team) error
	SoftDelete(ctx context.Context, id uint) error

	// TeamMember operations
	AddMember(ctx context.Context, member *model.TeamMember) error
	RemoveMember(ctx context.Context, teamBizKey, userBizKey int64) error
	FindMember(ctx context.Context, teamBizKey, userBizKey int64) (*model.TeamMember, error)
	ListMembers(ctx context.Context, teamBizKey int64) ([]*dto.TeamMemberDTO, error)
	CountMembers(ctx context.Context, teamBizKey int64) (int64, error)
	UpdateMember(ctx context.Context, member *model.TeamMember) error
	FindPMMembers(ctx context.Context, teamBizKeys []int64) (map[int64]string, error)

	// Paginated list with optional search
	ListFiltered(ctx context.Context, search string, offset, limit int) ([]*model.Team, int64, error)

	// Admin operations
	ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error)
	FindTeamsByUserIDs(ctx context.Context, userIDs []uint) (map[uint][]dto.TeamSummary, error)
	FindTeamsByUserBizKeys(ctx context.Context, userBizKeys []int64) (map[int64][]dto.TeamSummary, error)
}

// TeamMemberRepo defines persistence operations for TeamMember entities.
type TeamMemberRepo interface {
	AddMember(ctx context.Context, member *model.TeamMember) error
	RemoveMember(ctx context.Context, teamBizKey, userBizKey int64) error
	FindMember(ctx context.Context, teamBizKey, userBizKey int64) (*model.TeamMember, error)
	ListMembers(ctx context.Context, teamBizKey int64) ([]*dto.TeamMemberDTO, error)
	CountMembers(ctx context.Context, teamBizKey int64) (int64, error)
	UpdateMember(ctx context.Context, member *model.TeamMember) error
	FindPMMembers(ctx context.Context, teamBizKeys []int64) (map[int64]string, error)
	SoftDelete(ctx context.Context, id uint) error
}
