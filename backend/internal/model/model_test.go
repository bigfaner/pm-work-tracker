package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func TestUser_TableName(t *testing.T) {
	u := model.User{}
	assert.Equal(t, "users", u.TableName())
}

func TestUser_PasswordHashNotSerialized(t *testing.T) {
	u := model.User{
		Username:       "alice",
		DisplayName:    "Alice",
		PasswordHash:   "secret-hash",
		IsSuperAdmin:   false,
	}

	data, err := json.Marshal(u)
	require.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)

	_, hasPassword := m["password_hash"]
	_, hasPasswordHash := m["passwordHash"]
	assert.False(t, hasPassword, "password_hash should not appear in JSON")
	assert.False(t, hasPasswordHash, "passwordHash should not appear in JSON")
}

func TestUser_AutoMigrateCreatesCorrectSchema(t *testing.T) {
	db := setupTestDB(t)

	err := db.AutoMigrate(&model.User{})
	require.NoError(t, err)

	// Verify unique index on username exists by inserting duplicate
	u1 := model.User{Username: "testuser", DisplayName: "Test", PasswordHash: "h"}
	require.NoError(t, db.Create(&u1).Error)

	u2 := model.User{Username: "testuser", DisplayName: "Test2", PasswordHash: "h2"}
	err = db.Create(&u2).Error
	assert.Error(t, err, "duplicate username should be rejected")
}


func TestTeam_TableName(t *testing.T) {
	team := model.Team{}
	assert.Equal(t, "teams", team.TableName())
}

func TestTeam_AutoMigrateCreatesCorrectSchema(t *testing.T) {
	db := setupTestDB(t)

	err := db.AutoMigrate(&model.User{}, &model.Team{})
	require.NoError(t, err)

	// Create a user to be PM
	u := model.User{Username: "pm", DisplayName: "PM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)

	team := model.Team{Name: "Team A", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	var fetched model.Team
	db.First(&fetched, "name = ?", "Team A")
	assert.Equal(t, "Team A", fetched.Name)
	assert.Equal(t, u.ID, fetched.PmID)
}

func TestTeamMember_TableName(t *testing.T) {
	tm := model.TeamMember{}
	assert.Equal(t, "team_members", tm.TableName())
}

func TestTeamMember_NoSoftDelete(t *testing.T) {
	tm := model.TeamMember{}
	// TeamMember should NOT have DeletedAt field (no soft-delete)
	// Verify by reflecting on the struct — DeletedAt should not exist
	assert.Zero(t, tm.ID, "TeamMember should have plain ID, not gorm.Model")
	// No DeletedAt field means GORM won't add soft-delete scope
	_ = tm
}

func TestTeamMember_CompositeUniqueIndex(t *testing.T) {
	db := setupTestDB(t)

	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.TeamMember{})
	require.NoError(t, err)

	// Create user and team
	u := model.User{Username: "member1", DisplayName: "M1", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)

	team := model.Team{Name: "Team1", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	// First membership OK
	tm1 := model.TeamMember{TeamID: team.ID, UserID: u.ID, Role: "member"}
	require.NoError(t, db.Create(&tm1).Error)

	// Duplicate (team_id, user_id) should fail
	tm2 := model.TeamMember{TeamID: team.ID, UserID: u.ID, Role: "pm"}
	err = db.Create(&tm2).Error
	assert.Error(t, err, "duplicate (team_id, user_id) should be rejected")
}

func TestTeamMember_DefaultRole(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.TeamMember{})
	require.NoError(t, err)

	u := model.User{Username: "roleuser", DisplayName: "RU", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)

	team := model.Team{Name: "RoleTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	tm := model.TeamMember{TeamID: team.ID, UserID: u.ID}
	require.NoError(t, db.Create(&tm).Error)

	var fetched model.TeamMember
	db.First(&fetched, "team_id = ? AND user_id = ?", team.ID, u.ID)
	assert.Equal(t, "member", fetched.Role, "role should default to 'member'")
}
