package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pm-work-tracker/backend/internal/model"
)

func TestRole_TableName(t *testing.T) {
	r := model.Role{}
	assert.Equal(t, "pmw_roles", r.TableName())
}

func TestRole_AutoMigrateCreatesCorrectSchema(t *testing.T) {
	db := setupTestDB(t)

	err := db.AutoMigrate(&model.Role{})
	require.NoError(t, err)

	r1 := model.Role{Name: "pm", Description: "Project Manager", IsPreset: true}
	require.NoError(t, db.Create(&r1).Error)

	// Duplicate name should fail (unique constraint)
	r2 := model.Role{Name: "pm", Description: "Another PM"}
	err = db.Create(&r2).Error
	assert.Error(t, err, "duplicate role name should be rejected")
}

func TestRole_HasSoftDelete(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Role{})
	require.NoError(t, err)

	r := model.Role{Name: "to-delete", Description: "will be deleted"}
	require.NoError(t, db.Create(&r).Error)

	// Role uses hard delete (no soft delete needed per design)
	require.NoError(t, db.Delete(&r).Error)

	// Should NOT be found after hard delete
	var found model.Role
	err = db.First(&found, "role_name = ?", "to-delete").Error
	assert.Error(t, err, "deleted role should not be found")
}

func TestRole_Defaults(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Role{})
	require.NoError(t, err)

	r := model.Role{Name: "custom", Description: "A custom role"}
	require.NoError(t, db.Create(&r).Error)

	var fetched model.Role
	db.First(&fetched, "role_name = ?", "custom")
	assert.Equal(t, "custom", fetched.Name)
	assert.Equal(t, "A custom role", fetched.Description)
	assert.False(t, fetched.IsPreset, "IsPreset should default to false")
}

func TestRolePermission_TableName(t *testing.T) {
	rp := model.RolePermission{}
	assert.Equal(t, "pmw_role_permissions", rp.TableName())
}

func TestRolePermission_UniqueConstraint(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Role{}, &model.RolePermission{})
	require.NoError(t, err)

	r := model.Role{Name: "test-role", Description: "test"}
	require.NoError(t, db.Create(&r).Error)

	// First permission OK
	rp1 := model.RolePermission{RoleKey: r.BizKey, PermissionCode: "team:create"}
	require.NoError(t, db.Create(&rp1).Error)

	// Duplicate (role_key, permission_code) should fail
	rp2 := model.RolePermission{RoleKey: r.BizKey, PermissionCode: "team:create"}
	err = db.Create(&rp2).Error
	assert.Error(t, err, "duplicate (role_key, permission_code) should be rejected")
}

func TestRolePermission_HasSoftDelete(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Role{}, &model.RolePermission{})
	require.NoError(t, err)

	r := model.Role{Name: "soft-del-role", Description: "test"}
	require.NoError(t, db.Create(&r).Error)

	rp := model.RolePermission{RoleKey: r.BizKey, PermissionCode: "team:create"}
	require.NoError(t, db.Create(&rp).Error)

	// Soft-delete
	require.NoError(t, db.Model(&rp).Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error)

	// Same (role_key, permission_code) can be re-created after soft-delete
	rp2 := model.RolePermission{RoleKey: r.BizKey, PermissionCode: "team:create"}
	require.NoError(t, db.Create(&rp2).Error, "should allow re-creating after soft-delete")
}

func TestTeamMember_HasRoleID(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.Role{}, &model.TeamMember{})
	require.NoError(t, err)

	u := model.User{Username: "tmuser", DisplayName: "TM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)

	team := model.Team{TeamName: "TMTeam", PmKey: int64(u.ID)}
	require.NoError(t, db.Create(&team).Error)

	r := model.Role{Name: "member-role", Description: "member role"}
	require.NoError(t, db.Create(&r).Error)

	// Create TeamMember with RoleBizKey
	tm := model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(u.ID), RoleKey: &r.BizKey}
	require.NoError(t, db.Create(&tm).Error)

	var fetched model.TeamMember
	db.First(&fetched, "team_key = ? AND user_key = ?", team.ID, u.ID)
	assert.NotNil(t, fetched.RoleKey, "RoleKey should be set")
	assert.Equal(t, r.BizKey, *fetched.RoleKey, "RoleKey should match role BizKey")
}

func TestTeamMember_RoleIDNullable(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.Role{}, &model.TeamMember{})
	require.NoError(t, err)

	u := model.User{Username: "tmnull", DisplayName: "TMN", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)

	team := model.Team{TeamName: "NullTeam", PmKey: int64(u.ID)}
	require.NoError(t, db.Create(&team).Error)

	// Create TeamMember without RoleID (nullable)
	tm := model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(u.ID)}
	require.NoError(t, db.Create(&tm).Error)

	var fetched model.TeamMember
	db.First(&fetched, "team_key = ? AND user_key = ?", team.ID, u.ID)
	assert.Nil(t, fetched.RoleKey, "RoleID should be nil when not set")
}
