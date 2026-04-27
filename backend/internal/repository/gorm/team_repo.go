package gorm

import (
	"context"
	stderrors "errors"
	"strings"
	"time"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
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
	err := r.db.WithContext(ctx).Create(team).Error
	if err != nil && isDuplicateKeyError(err) {
		return errors.ErrTeamCodeDuplicate
	}
	return err
}

func (r *teamRepo) FindByID(ctx context.Context, teamID uint) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("id = ?", teamID).First(&team).Error
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
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Find(&teams).Error
	return teams, err
}

func (r *teamRepo) ListFiltered(ctx context.Context, search string, offset, limit int) ([]*model.Team, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Team{}).Scopes(NotDeleted)
	if search != "" {
		q = q.Where("team_name LIKE ? OR team_code LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var teams []*model.Team
	err := q.Order("create_time DESC").Offset(offset).Limit(limit).Find(&teams).Error
	return teams, total, err
}

func (r *teamRepo) Update(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).Save(team).Error
}

func (r *teamRepo) SoftDelete(ctx context.Context, teamID uint) error {
	return r.db.WithContext(ctx).Model(&model.Team{}).
		Where("id = ?", teamID).
		Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error
}

func (r *teamRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key = ?", bizKey).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
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
		Where("team_key = ? AND user_key = ?", teamID, userID).
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
		Scopes(NotDeleted).
		Where("team_key = ? AND user_key = ?", teamID, userID).
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
	type scanRow struct {
		BizKey      int64
		TeamKey     int64
		UserKey     int64
		Role        string
		JoinedAt    string
		DisplayName string
		Username    string
	}
	var rows []scanRow
	err := r.db.WithContext(ctx).
		Table("pmw_team_members").
		Select("pmw_team_members.biz_key, pmw_teams.biz_key as team_key, pmw_users.biz_key as user_key, "+
			"CASE WHEN pmw_roles.role_name IS NOT NULL THEN pmw_roles.role_name "+
			"     WHEN pmw_team_members.user_key = pmw_teams.pm_key THEN 'pm' "+
			"     ELSE 'member' END as role, "+
			"pmw_team_members.joined_at, pmw_users.display_name, pmw_users.username").
		Joins("LEFT JOIN pmw_users ON pmw_users.id = pmw_team_members.user_key").
		Joins("LEFT JOIN pmw_roles ON pmw_roles.id = pmw_team_members.role_key").
		Joins("LEFT JOIN pmw_teams ON pmw_teams.id = pmw_team_members.team_key").
		Scopes(NotDeletedTable("pmw_users")).
		Where("pmw_team_members.team_key = ?", teamID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	results := make([]*dto.TeamMemberDTO, len(rows))
	for i, row := range rows {
		results[i] = &dto.TeamMemberDTO{
			BizKey:      pkg.FormatID(row.BizKey),
			TeamKey:     pkg.FormatID(row.TeamKey),
			UserKey:     pkg.FormatID(row.UserKey),
			Role:        row.Role,
			JoinedAt:    row.JoinedAt,
			DisplayName: row.DisplayName,
			Username:    row.Username,
		}
	}
	return results, nil
}

func (r *teamRepo) CountMembers(ctx context.Context, teamID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("pmw_team_members").
		Scopes(NotDeleted).
		Where("team_key = ?", teamID).
		Count(&count).Error
	return count, err
}

func (r *teamRepo) UpdateMember(ctx context.Context, member *model.TeamMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *teamRepo) FindPMMembers(ctx context.Context, teamIDs []uint) (map[uint]string, error) {
	if len(teamIDs) == 0 {
		return map[uint]string{}, nil
	}

	type row struct {
		TeamKey     uint
		DisplayName string
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("pmw_team_members").
		Select("pmw_team_members.team_key, pmw_users.display_name").
		Joins("JOIN pmw_users ON pmw_users.id = pmw_team_members.user_key").
		Joins("JOIN pmw_roles ON pmw_roles.id = pmw_team_members.role_key").
		Scopes(NotDeletedTable("pmw_users")).
		Where("pmw_team_members.team_key IN ? AND pmw_roles.role_name = ?", teamIDs, "pm").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]string, len(rows))
	for _, r := range rows {
		result[r.TeamKey] = r.DisplayName
	}
	return result, nil
}

func (r *teamRepo) ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error) {
	type scanRow struct {
		BizKey        int64
		Name          string
		PMDisplayName string
		MemberCount   int
		MainItemCount int
		CreatedAt     string
	}
	var rows []scanRow
	err := r.db.WithContext(ctx).
		Table("pmw_teams").
		Select("pmw_teams.biz_key, pmw_teams.team_name as name, pmw_users.display_name as pm_display_name, "+
			"(SELECT COUNT(*) FROM pmw_team_members WHERE pmw_team_members.team_key = pmw_teams.id) as member_count, "+
			"(SELECT COUNT(*) FROM pmw_main_items WHERE pmw_main_items.team_key = pmw_teams.id AND pmw_main_items.deleted_flag = 0) as main_item_count, "+
			"pmw_teams.create_time as created_at").
		Joins("LEFT JOIN pmw_users ON pmw_users.id = pmw_teams.pm_key").
		Scopes(NotDeletedTable("pmw_teams")).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	results := make([]*dto.AdminTeamDTO, len(rows))
	for i, row := range rows {
		results[i] = &dto.AdminTeamDTO{
			BizKey:        pkg.FormatID(row.BizKey),
			Name:          row.Name,
			PMDisplayName: row.PMDisplayName,
			MemberCount:   row.MemberCount,
			MainItemCount: row.MainItemCount,
			CreatedAt:     row.CreatedAt,
		}
	}
	return results, nil
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
		UserID uint
		TeamID uint
		BizKey int64
		Name   string
		Role   string
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("pmw_team_members").
		Select("pmw_team_members.user_key as user_id, pmw_team_members.team_key as team_id, pmw_teams.biz_key, pmw_teams.team_name as name, "+
			"CASE WHEN pmw_roles.role_name IS NOT NULL THEN pmw_roles.role_name "+
			"     WHEN pmw_team_members.user_key = pmw_teams.pm_key THEN 'pm' "+
			"     ELSE 'member' END as role").
		Joins("JOIN pmw_teams ON pmw_teams.id = pmw_team_members.team_key").
		Joins("LEFT JOIN pmw_roles ON pmw_roles.id = pmw_team_members.role_key").
		Scopes(NotDeletedTable("pmw_teams")).
		Where("pmw_team_members.user_key IN ?", userIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint][]dto.TeamSummary)
	for _, r := range rows {
		result[r.UserID] = append(result[r.UserID], dto.TeamSummary{
			BizKey: pkg.FormatID(r.BizKey),
			TeamID: r.TeamID,
			Name:   r.Name,
			Role:   r.Role,
		})
	}
	return result, nil
}
