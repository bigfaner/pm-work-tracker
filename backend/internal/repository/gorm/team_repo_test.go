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
	"pm-work-tracker/backend/internal/pkg/snowflake"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

func setupTeamTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.Role{}, &model.TeamMember{}))
	// Create the composite unique index that matches the real schema DDL.
	// GORM AutoMigrate can't express a unique index spanning BaseModel fields + model fields,
	// so we create it manually here to match the production migration.
	require.NoError(t, db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uk_team_user_deleted ON pmw_team_members(team_key, user_key, deleted_flag, deleted_time)").Error)
	return db
}

func seedUser(t *testing.T, db *gormlib.DB, username string) *model.User {
	t.Helper()
	u := model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Username: username, DisplayName: username + "_display", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	return &u
}

// --- Team CRUD ---

func TestTeamRepo_Create(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm1")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team A", PmKey: pm.BizKey, Code: "TAMA"}
	require.NoError(t, repo.Create(ctx, &team))
	assert.NotZero(t, team.ID)
}

func TestTeamRepo_Create_DuplicateCode(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm_dc")
	require.NoError(t, repo.Create(ctx, &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team X", PmKey: pm.BizKey, Code: "DUPX"}))

	err := repo.Create(ctx, &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team Y", PmKey: pm.BizKey, Code: "DUPX"})
	assert.ErrorIs(t, err, pkgerrors.ErrTeamCodeDuplicate)
}

func TestTeamRepo_FindByID(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm2")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team B", PmKey: pm.BizKey, Code: "TAMB"}
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
	require.NoError(t, repo.Create(ctx, &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "T1", PmKey: pm.BizKey, Code: "TT01"}))
	require.NoError(t, repo.Create(ctx, &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "T2", PmKey: pm.BizKey, Code: "TT02"}))

	teams, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, teams, 2)
}

func TestTeamRepo_Update(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm4")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Old Name", PmKey: pm.BizKey, Code: "OLDN"}
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
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "ToDelete", PmKey: pm.BizKey, Code: "TDEL"}
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
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team M", PmKey: pm.BizKey, Code: "TAMM"}
	require.NoError(t, repo.Create(ctx, &team))

	t.Run("success", func(t *testing.T) {
		tm := model.TeamMember{
			TeamKey: team.BizKey,
			UserKey: member.BizKey,
			JoinedAt: time.Now(),
		}
		require.NoError(t, repo.AddMember(ctx, &tm))
		assert.NotZero(t, tm.ID)
	})

	t.Run("duplicate_returns_ErrAlreadyExists", func(t *testing.T) {
		tm := model.TeamMember{
			TeamKey: team.BizKey,
			UserKey: member.BizKey,
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
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team R", PmKey: pm.BizKey, Code: "TAMR"}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamKey: team.BizKey, UserKey: member.BizKey,  JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	t.Run("success", func(t *testing.T) {
		require.NoError(t, repo.RemoveMember(ctx, team.BizKey, member.BizKey))
		// Verify member is gone
		_, err := repo.FindMember(ctx, team.BizKey, member.BizKey)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})

	t.Run("not_found_returns_error", func(t *testing.T) {
		err := repo.RemoveMember(ctx, team.BizKey, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

func TestTeamRepo_FindMember(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm8")
	member := seedUser(t, db, "member8")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team F", PmKey: pm.BizKey, Code: "TAMF"}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamKey: team.BizKey, UserKey: member.BizKey,  JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindMember(ctx, team.BizKey, member.BizKey)
		require.NoError(t, err)
		assert.Equal(t, team.BizKey, found.TeamKey)
		assert.Equal(t, member.BizKey, found.UserKey)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindMember(ctx, team.BizKey, 9999)
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
	pmRole.BizKey = snowflake.Generate()
	memberRole.BizKey = snowflake.Generate()
	require.NoError(t, db.Save(&pmRole).Error)
	require.NoError(t, db.Save(&memberRole).Error)

	pm := seedUser(t, db, "pm9")
	m1 := seedUser(t, db, "m1")
	m2 := seedUser(t, db, "m2")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team LM", PmKey: pm.BizKey, Code: "TALM"}
	require.NoError(t, repo.Create(ctx, &team))

	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamKey: team.BizKey, UserKey: m1.BizKey, RoleKey: &pmRole.BizKey, JoinedAt: time.Now(),
	}))
	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamKey: team.BizKey, UserKey: m2.BizKey, RoleKey: &memberRole.BizKey, JoinedAt: time.Now(),
	}))

	results, err := repo.ListMembers(ctx, team.BizKey)
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
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team Null", PmKey: pmUser.BizKey, Code: "TNUL"}
	require.NoError(t, repo.Create(ctx, &team))

	// Add members with no RoleID (legacy data)
	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamKey: team.BizKey, UserKey: pmUser.BizKey, JoinedAt: time.Now(),
	}))
	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamKey: team.BizKey, UserKey: memberUser.BizKey, JoinedAt: time.Now(),
	}))

	results, err := repo.ListMembers(ctx, team.BizKey)
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

	results, err := repo.ListMembers(ctx, int64(9999))
	require.NoError(t, err)
	assert.Empty(t, results)
}

// --- Soft-delete exclusion tests ---

func TestTeamRepo_FindByBizKey_ExcludesSoftDeleted(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm_sd_biz")
	team := model.Team{BaseModel: model.BaseModel{BizKey: 12345}, TeamName: "SoftDel BizKey", PmKey: pm.BizKey, Code: "SD01"}
	require.NoError(t, repo.Create(ctx, &team))
	require.NoError(t, repo.SoftDelete(ctx, team.ID))

	_, err := repo.FindByBizKey(ctx, team.BizKey)
	assert.ErrorIs(t, err, gormlib.ErrRecordNotFound, "soft-deleted team should not be found by biz_key")
}

func TestTeamRepo_List_ExcludesSoftDeleted(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm_sd_list")
	require.NoError(t, repo.Create(ctx, &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Active Team", PmKey: pm.BizKey, Code: "SD02"}))
	deleted := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Deleted Team", PmKey: pm.BizKey, Code: "SD03"}
	require.NoError(t, repo.Create(ctx, &deleted))
	require.NoError(t, repo.SoftDelete(ctx, deleted.ID))

	teams, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, teams, 1)
	assert.Equal(t, "Active Team", teams[0].TeamName)
}

func TestTeamRepo_FindMember_ExcludesSoftDeleted(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm_sd_fm")
	member := seedUser(t, db, "member_sd_fm")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team SD FM", PmKey: pm.BizKey, Code: "SDFM"}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamKey: team.BizKey, UserKey: member.BizKey, JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	// Soft-delete the member record directly
	require.NoError(t, db.Model(&model.TeamMember{}).Where("id = ?", tm.ID).
		Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error)

	_, err := repo.FindMember(ctx, team.BizKey, member.BizKey)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound, "soft-deleted member should not be found by FindMember")
}

func TestTeamRepo_CountMembers_ExcludesSoftDeleted(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm_sd_cm")
	m1 := seedUser(t, db, "m1_sd_cm")
	m2 := seedUser(t, db, "m2_sd_cm")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team SD CM", PmKey: pm.BizKey, Code: "SDCM"}
	require.NoError(t, repo.Create(ctx, &team))

	tm1 := model.TeamMember{TeamKey: team.BizKey, UserKey: m1.BizKey, JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm1))
	tm2 := model.TeamMember{TeamKey: team.BizKey, UserKey: m2.BizKey, JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm2))

	// Soft-delete one member
	require.NoError(t, db.Model(&model.TeamMember{}).Where("id = ?", tm1.ID).
		Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error)

	count, err := repo.CountMembers(ctx, team.BizKey)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "CountMembers should exclude soft-deleted members")
}

func TestTeamRepo_UpdateMember(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm10")
	member := seedUser(t, db, "member10")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team UM", PmKey: pm.BizKey, Code: "TAUM"}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamKey: team.BizKey, UserKey: member.BizKey,  JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	require.NoError(t, repo.UpdateMember(ctx, &tm))

	found, err := repo.FindMember(ctx, team.BizKey, member.BizKey)
	require.NoError(t, err)
	assert.Equal(t, team.BizKey, found.TeamKey)
	assert.Equal(t, member.BizKey, found.UserKey)
}

// --- RemoveMember soft-delete tests ---

func TestTeamRepo_RemoveMember_SoftDeletes(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm_sd_rm")
	member := seedUser(t, db, "member_sd_rm")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team SD RM", PmKey: pm.BizKey, Code: "SDRM"}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamKey: team.BizKey, UserKey: member.BizKey, JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	require.NoError(t, repo.RemoveMember(ctx, team.BizKey, member.BizKey))

	// Record still exists in DB with deleted_flag=1
	var count int64
	db.Unscoped().Model(&model.TeamMember{}).Where("id = ?", tm.ID).Count(&count)
	assert.Equal(t, int64(1), count, "team member should still exist as soft-deleted")

	var found model.TeamMember
	require.NoError(t, db.Unscoped().First(&found, tm.ID).Error)
	assert.Equal(t, 1, found.DeletedFlag, "deleted_flag should be 1 after RemoveMember")

	// FindMember should not find the removed member
	_, err := repo.FindMember(ctx, team.BizKey, member.BizKey)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

func TestTeamRepo_RemoveMember_NotFound(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm_sd_nf")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team SD NF", PmKey: pm.BizKey, Code: "SDNF"}
	require.NoError(t, repo.Create(ctx, &team))

	err := repo.RemoveMember(ctx, team.BizKey, 9999)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

func TestTeamRepo_RemoveMember_ReaddAfterRemove(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm_sd_ra")
	member := seedUser(t, db, "member_sd_ra")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team SD RA", PmKey: pm.BizKey, Code: "SDRA"}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamKey: team.BizKey, UserKey: member.BizKey, JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	// Soft-delete the member
	require.NoError(t, repo.RemoveMember(ctx, team.BizKey, member.BizKey))

	// Re-add the same member (should create a new row, not error)
	newTm := model.TeamMember{TeamKey: team.BizKey, UserKey: member.BizKey, JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &newTm))
	assert.NotZero(t, newTm.ID)
	assert.NotEqual(t, tm.ID, newTm.ID, "re-added member should have a new ID")

	// FindMember should find the new record
	found, err := repo.FindMember(ctx, team.BizKey, member.BizKey)
	require.NoError(t, err)
	assert.Equal(t, newTm.ID, found.ID)
}
