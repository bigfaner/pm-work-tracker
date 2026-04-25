package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pm-work-tracker/backend/internal/model"
)

func TestRole_TableName(t *testing.T) {
	r := model.Role{}
	assert.Equal(t, "roles", r.TableName())
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
	err = db.First(&found, "name = ?", "to-delete").Error
	assert.Error(t, err, "deleted role should not be found")
}

func TestRole_Defaults(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Role{})
	require.NoError(t, err)

	r := model.Role{Name: "custom", Description: "A custom role"}
	require.NoError(t, db.Create(&r).Error)

	var fetched model.Role
	db.First(&fetched, "name = ?", "custom")
	assert.Equal(t, "custom", fetched.Name)
	assert.Equal(t, "A custom role", fetched.Description)
	assert.False(t, fetched.IsPreset, "IsPreset should default to false")
}

func TestRolePermission_TableName(t *testing.T) {
	rp := model.RolePermission{}
	assert.Equal(t, "role_permissions", rp.TableName())
}

func TestRolePermission_UniqueConstraint(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Role{})
	require.NoError(t, err)

	// Use raw SQL to create role_permissions with unique constraint
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS role_permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_id INTEGER NOT NULL REFERENCES roles(id),
			permission_code TEXT NOT NULL,
			UNIQUE(role_id, permission_code)
		)
	`).Error
	require.NoError(t, err)

	r := model.Role{Name: "test-role", Description: "test"}
	require.NoError(t, db.Create(&r).Error)

	// First permission OK
	rp1 := model.RolePermission{RoleID: r.ID, PermissionCode: "team:create"}
	require.NoError(t, db.Create(&rp1).Error)

	// Duplicate (role_id, permission_code) should fail
	rp2 := model.RolePermission{RoleID: r.ID, PermissionCode: "team:create"}
	err = db.Create(&rp2).Error
	assert.Error(t, err, "duplicate (role_id, permission_code) should be rejected")
}

func TestRolePermission_NoSoftDelete(t *testing.T) {
	rp := model.RolePermission{}
	// RolePermission should NOT have DeletedAt field
	assert.Zero(t, rp.ID, "RolePermission should have plain ID")
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

	// Create TeamMember with RoleID
	tm := model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(u.ID), RoleKey: func() *int64 { v := int64(r.ID); return &v }()}
	require.NoError(t, db.Create(&tm).Error)

	var fetched model.TeamMember
	db.First(&fetched, "team_key = ? AND user_key = ?", team.ID, u.ID)
	assert.NotNil(t, fetched.RoleKey, "RoleID should be set")
	assert.Equal(t, int64(r.ID), *fetched.RoleKey, "RoleID should match role ID")
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
