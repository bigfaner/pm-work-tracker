package gorm

import (
	"context"
	stderrors "errors"
	"time"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

type roleRepo struct {
	db *gormlib.DB
}

// NewGormRoleRepo creates a GORM-backed RoleRepo.
func NewGormRoleRepo(db *gormlib.DB) repository.RoleRepo {
	return &roleRepo{db: db}
}

func (r *roleRepo) List(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Order("create_time ASC").Find(&roles).Error
	return roles, err
}

func (r *roleRepo) FindByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("id = ?", id).First(&role).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key = ?", bizKey).First(&role).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) FindByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("role_name = ?", name).First(&role).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepo) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Role{}).
		Where("id = ?", id).
		Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error
}

func (r *roleRepo) ListPermissions(ctx context.Context, roleKey int64) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Model(&model.RolePermission{}).
		Where("role_key = ? AND deleted_flag = 0", roleKey).
		Pluck("permission_code", &codes).Error
	return codes, err
}

func (r *roleRepo) SetPermissions(ctx context.Context, roleKey int64, codes []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gormlib.DB) error {
		// Soft-delete all existing active permissions for this role
		if err := tx.Model(&model.RolePermission{}).
			Where("role_key = ? AND deleted_flag = 0", roleKey).
			Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error; err != nil {
			return err
		}
		// Insert new permissions
		for _, code := range codes {
			rp := model.RolePermission{RoleKey: roleKey, PermissionCode: code}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *roleRepo) CountMembersByRoleKey(ctx context.Context, roleKey int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TeamMember{}).
		Scopes(NotDeletedTable("pmw_team_members")).
		Joins("JOIN pmw_roles ON pmw_roles.biz_key = pmw_team_members.role_key AND pmw_roles.deleted_flag = 0").
		Where("pmw_roles.biz_key = ?", roleKey).
		Count(&count).Error
	return count, err
}

func (r *roleRepo) HasPermission(ctx context.Context, userBizKey int64, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("pmw_team_members").
		Scopes(NotDeletedTable("pmw_team_members")).
		Joins("JOIN pmw_roles ON pmw_roles.biz_key = pmw_team_members.role_key AND pmw_roles.deleted_flag = 0").
		Joins("JOIN pmw_role_permissions ON pmw_role_permissions.role_key = pmw_roles.biz_key AND pmw_role_permissions.deleted_flag = 0").
		Where("pmw_team_members.user_key = ? AND pmw_role_permissions.permission_code = ?", userBizKey, code).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// teamPermRow is a helper struct for scanning the GetUserTeamPermissions join query.
type teamPermRow struct {
	TeamBizKey     int64
	PermissionCode string
}

func (r *roleRepo) GetUserTeamPermissions(ctx context.Context, userBizKey int64) (map[int64][]string, error) {
	var rows []teamPermRow
	err := r.db.WithContext(ctx).
		Table("pmw_team_members").
		Scopes(NotDeletedTable("pmw_team_members")).
		Select("pmw_teams.biz_key as team_biz_key, pmw_role_permissions.permission_code").
		Joins("JOIN pmw_roles ON pmw_roles.biz_key = pmw_team_members.role_key AND pmw_roles.deleted_flag = 0").
		Joins("JOIN pmw_role_permissions ON pmw_role_permissions.role_key = pmw_roles.biz_key AND pmw_role_permissions.deleted_flag = 0").
		Joins("JOIN pmw_teams ON pmw_teams.biz_key = pmw_team_members.team_key AND pmw_teams.deleted_flag = 0").
		Where("pmw_team_members.user_key = ?", userBizKey).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64][]string)
	for _, row := range rows {
		result[row.TeamBizKey] = append(result[row.TeamBizKey], row.PermissionCode)
	}
	return result, nil
}
