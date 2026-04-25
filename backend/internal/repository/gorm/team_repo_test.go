package gorm_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	pkgerrors "pm-work-tracker/backend/internal/pkg/errors"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

func setupTeamTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.Role{}, &model.TeamMember{}))
	return db
}

func seedUser(t *testing.T, db *gormlib.DB, username string) *model.User {
	t.Helper()
	u := model.User{Username: username, DisplayName: username + "_display", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	return &u
}

// --- Team CRUD ---

func TestTeamRepo_Create(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm1")
	team := model.Team{TeamName: "Team A", PmKey: int64(pm.ID), Code: "TAMA"}
	require.NoError(t, repo.Create(ctx, &team))
	assert.NotZero(t, team.ID)
}

func TestTeamRepo_Create_DuplicateCode(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm_dc")
	require.NoError(t, repo.Create(ctx, &model.Team{TeamName: "Team X", PmKey: int64(pm.ID), Code: "DUPX"}))

	err := repo.Create(ctx, &model.Team{TeamName: "Team Y", PmKey: int64(pm.ID), Code: "DUPX"})
	assert.ErrorIs(t, err, pkgerrors.ErrTeamCodeDuplicate)
}

func TestTeamRepo_FindByID(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm2")
	team := model.Team{TeamName: "Team B", PmKey: int64(pm.ID), Code: "TAMB"}
	require.NoError(t, repo.Create(ctx, &team))

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, team.ID)
		require.NoError(t, err)
		assert.Equal(t, "Team B", found.TeamName)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

func TestTeamRepo_List(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm3")
	require.NoError(t, repo.Create(ctx, &model.Team{TeamName: "T1", PmKey: int64(pm.ID), Code: "TT01"}))
	require.NoError(t, repo.Create(ctx, &model.Team{TeamName: "T2", PmKey: int64(pm.ID), Code: "TT02"}))

	teams, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, teams, 2)
}

func TestTeamRepo_Update(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm4")
	team := model.Team{TeamName: "Old Name", PmKey: int64(pm.ID), Code: "OLDN"}
	require.NoError(t, repo.Create(ctx, &team))

	team.TeamName = "New Name"
	require.NoError(t, repo.Update(ctx, &team))

	found, err := repo.FindByID(ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", found.TeamName)
}

func TestTeamRepo_Delete_SoftDelete(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm5")
	team := model.Team{TeamName: "ToDelete", PmKey: int64(pm.ID), Code: "TDEL"}
	require.NoError(t, repo.Create(ctx, &team))

	require.NoError(t, repo.SoftDelete(ctx, team.ID))

	// FindByID should return ErrNotFound (soft-deleted rows excluded by GORM)
	_, err := repo.FindByID(ctx, team.ID)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)

	// Verify the record still exists in DB (soft-deleted, not hard-deleted)
	var count int64
	db.Unscoped().Model(&model.Team{}).Where("id = ?", team.ID).Count(&count)
	assert.Equal(t, int64(1), count, "team should still exist as soft-deleted")
}

// --- TeamMember operations ---

func TestTeamRepo_AddMember(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm6")
	member := seedUser(t, db, "member6")
	team := model.Team{TeamName: "Team M", PmKey: int64(pm.ID), Code: "TAMM"}
	require.NoError(t, repo.Create(ctx, &team))

	t.Run("success", func(t *testing.T) {
		tm := model.TeamMember{
			TeamKey: int64(team.ID),
			UserKey: int64(member.ID),
			JoinedAt: time.Now(),
		}
		require.NoError(t, repo.AddMember(ctx, &tm))
		assert.NotZero(t, tm.ID)
	})

	t.Run("duplicate_returns_ErrAlreadyExists", func(t *testing.T) {
		tm := model.TeamMember{
			TeamKey: int64(team.ID),
			UserKey: int64(member.ID),
			JoinedAt: time.Now(),
		}
		err := repo.AddMember(ctx, &tm)
		assert.ErrorIs(t, err, pkgerrors.ErrAlreadyExists)
	})
}

func TestTeamRepo_RemoveMember(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm7")
	member := seedUser(t, db, "member7")
	team := model.Team{TeamName: "Team R", PmKey: int64(pm.ID), Code: "TAMR"}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(member.ID),  JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	t.Run("success", func(t *testing.T) {
		require.NoError(t, repo.RemoveMember(ctx, team.ID, member.ID))
		// Verify member is gone
		_, err := repo.FindMember(ctx, team.ID, member.ID)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})

	t.Run("not_found_returns_error", func(t *testing.T) {
		err := repo.RemoveMember(ctx, team.ID, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

func TestTeamRepo_FindMember(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm8")
	member := seedUser(t, db, "member8")
	team := model.Team{TeamName: "Team F", PmKey: int64(pm.ID), Code: "TAMF"}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(member.ID),  JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindMember(ctx, team.ID, member.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(team.ID), found.TeamKey)
		assert.Equal(t, member.ID, uint(found.UserKey))
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindMember(ctx, team.ID, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

func TestTeamRepo_ListMembers(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	// Seed roles so the join returns a non-empty role name
	pmRole := model.Role{Name: "pm", Description: "PM"}
	memberRole := model.Role{Name: "member", Description: "Member"}
	require.NoError(t, db.Create(&pmRole).Error)
	require.NoError(t, db.Create(&memberRole).Error)

	pm := seedUser(t, db, "pm9")
	m1 := seedUser(t, db, "m1")
	m2 := seedUser(t, db, "m2")
	team := model.Team{TeamName: "Team LM", PmKey: int64(pm.ID), Code: "TALM"}
	require.NoError(t, repo.Create(ctx, &team))

	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamKey: int64(team.ID), UserKey: int64(m1.ID), RoleKey: func() *int64 { v := int64(pmRole.ID); return &v }(), JoinedAt: time.Now(),
	}))
	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamKey: int64(team.ID), UserKey: int64(m2.ID), RoleKey: func() *int64 { v := int64(memberRole.ID); return &v }(), JoinedAt: time.Now(),
	}))

	results, err := repo.ListMembers(ctx, team.ID)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Verify joined user display name, username, and role are all populated
	roleByUsername := map[string]string{}
	for _, dto := range results {
		assert.NotEmpty(t, dto.DisplayName, "display_name should be populated from join")
		assert.NotEmpty(t, dto.Username, "username should be populated from join")
		assert.NotEmpty(t, dto.Role, "role should be populated from roles join")
		roleByUsername[dto.Username] = dto.Role
	}
	assert.True(t, roleByUsername["m1"] == "pm", "m1 should have role pm")
	assert.True(t, roleByUsername["m2"] == "member", "m2 should have role member")
}

// TestTeamRepo_ListMembers_NullRoleID verifies that members with NULL role_id
// (created before the RoleID migration) still get a correct role via fallback:
// the team PM gets "pm", everyone else gets "member".
func TestTeamRepo_ListMembers_NullRoleID(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pmUser := seedUser(t, db, "pm_null")
	memberUser := seedUser(t, db, "member_null")
	team := model.Team{TeamName: "Team Null", PmKey: int64(pmUser.ID), Code: "TNUL"}
	require.NoError(t, repo.Create(ctx, &team))

	// Add members with no RoleID (legacy data)
	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamKey: int64(team.ID), UserKey: int64(pmUser.ID), JoinedAt: time.Now(),
	}))
	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamKey: int64(team.ID), UserKey: int64(memberUser.ID), JoinedAt: time.Now(),
	}))

	results, err := repo.ListMembers(ctx, team.ID)
	require.NoError(t, err)
	require.Len(t, results, 2)

	roleByUsername := map[string]string{}
	for _, r := range results {
		assert.NotEmpty(t, r.Role, "role must not be empty even with NULL role_id")
		roleByUsername[r.Username] = r.Role
	}
	assert.Equal(t, "pm", roleByUsername["pm_null"], "team PM should fall back to 'pm'")
	assert.Equal(t, "member", roleByUsername["member_null"], "non-PM should fall back to 'member'")
}

func TestTeamRepo_ListMembers_EmptyTeam(t *testing.T) {	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	results, err := repo.ListMembers(ctx, 9999)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestTeamRepo_UpdateMember(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm10")
	member := seedUser(t, db, "member10")
	team := model.Team{TeamName: "Team UM", PmKey: int64(pm.ID), Code: "TAUM"}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(member.ID),  JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	require.NoError(t, repo.UpdateMember(ctx, &tm))

	found, err := repo.FindMember(ctx, team.ID, member.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(team.ID), found.TeamKey)
	assert.Equal(t, member.ID, uint(found.UserKey))
}
