package gorm

import (
	"context"
	stderrors "errors"
	"strings"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

type teamRepo struct {
	db *gormlib.DB
}

// NewGormTeamRepo creates a GORM-backed TeamRepo.
func NewGormTeamRepo(db *gormlib.DB) repository.TeamRepo {
	return &teamRepo{db: db}
}

func (r *teamRepo) Create(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).Create(team).Error
}

func (r *teamRepo) FindByID(ctx context.Context, teamID uint) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).First(&team, teamID).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &team, nil
}

func (r *teamRepo) List(ctx context.Context) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.db.WithContext(ctx).Find(&teams).Error
	return teams, err
}

func (r *teamRepo) Update(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).Save(team).Error
}

func (r *teamRepo) Delete(ctx context.Context, teamID uint) error {
	return r.db.WithContext(ctx).Delete(&model.Team{}, teamID).Error
}

func (r *teamRepo) AddMember(ctx context.Context, member *model.TeamMember) error {
	err := r.db.WithContext(ctx).Create(member).Error
	if err != nil {
		// Check for unique constraint violation on (team_id, user_id)
		if isDuplicateKeyError(err) {
			return errors.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *teamRepo) RemoveMember(ctx context.Context, teamID, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&model.TeamMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *teamRepo) FindMember(ctx context.Context, teamID, userID uint) (*model.TeamMember, error) {
	var member model.TeamMember
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *teamRepo) ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error) {
	var results []*dto.TeamMemberDTO
	err := r.db.WithContext(ctx).
		Table("team_members").
		Select("team_members.id, team_members.team_id, team_members.user_id, team_members.role, team_members.joined_at, users.display_name, users.username").
		Joins("LEFT JOIN users ON users.id = team_members.user_id").
		Where("team_members.team_id = ?", teamID).
		Scan(&results).Error
	return results, err
}

func (r *teamRepo) UpdateMember(ctx context.Context, member *model.TeamMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *teamRepo) ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error) {
	var results []*dto.AdminTeamDTO
	err := r.db.WithContext(ctx).
		Table("teams").
		Select("teams.id, teams.name, users.display_name as pm_display_name, "+
			"(SELECT COUNT(*) FROM team_members WHERE team_members.team_id = teams.id) as member_count, "+
			"(SELECT COUNT(*) FROM main_items WHERE main_items.team_id = teams.id AND main_items.deleted_at IS NULL) as main_item_count, "+
			"teams.created_at").
		Joins("LEFT JOIN users ON users.id = teams.pm_id").
		Where("teams.deleted_at IS NULL").
		Scan(&results).Error
	return results, err
}

// isDuplicateKeyError checks if the error is a unique constraint violation.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "Duplicate entry")
}

func (r *teamRepo) FindTeamsByUserIDs(ctx context.Context, userIDs []uint) (map[uint][]dto.TeamSummary, error) {
	if len(userIDs) == 0 {
		return map[uint][]dto.TeamSummary{}, nil
	}

	type row struct {
		UserID  uint
		TeamID  uint
		Name    string
		Role    string
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("team_members").
		Select("team_members.user_id, team_members.team_id, teams.name, team_members.role").
		Joins("JOIN teams ON teams.id = team_members.team_id").
		Where("team_members.user_id IN ?", userIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint][]dto.TeamSummary)
	for _, r := range rows {
		result[r.UserID] = append(result[r.UserID], dto.TeamSummary{
			ID:   r.TeamID,
			Name: r.Name,
			Role: r.Role,
		})
	}
	return result, nil
}
