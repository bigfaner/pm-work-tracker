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
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.TeamMember{}))
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
	team := model.Team{Name: "Team A", PmID: pm.ID}
	require.NoError(t, repo.Create(ctx, &team))
	assert.NotZero(t, team.ID)
}

func TestTeamRepo_FindByID(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm2")
	team := model.Team{Name: "Team B", PmID: pm.ID}
	require.NoError(t, repo.Create(ctx, &team))

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, team.ID)
		require.NoError(t, err)
		assert.Equal(t, "Team B", found.Name)
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
	require.NoError(t, repo.Create(ctx, &model.Team{Name: "T1", PmID: pm.ID}))
	require.NoError(t, repo.Create(ctx, &model.Team{Name: "T2", PmID: pm.ID}))

	teams, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, teams, 2)
}

func TestTeamRepo_Update(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm4")
	team := model.Team{Name: "Old Name", PmID: pm.ID}
	require.NoError(t, repo.Create(ctx, &team))

	team.Name = "New Name"
	require.NoError(t, repo.Update(ctx, &team))

	found, err := repo.FindByID(ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", found.Name)
}

func TestTeamRepo_Delete_SoftDelete(t *testing.T) {
	db := setupTeamTestDB(t)
	repo := gormrepo.NewGormTeamRepo(db)
	ctx := context.Background()

	pm := seedUser(t, db, "pm5")
	team := model.Team{Name: "ToDelete", PmID: pm.ID}
	require.NoError(t, repo.Create(ctx, &team))

	require.NoError(t, repo.Delete(ctx, team.ID))

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
	team := model.Team{Name: "Team M", PmID: pm.ID}
	require.NoError(t, repo.Create(ctx, &team))

	t.Run("success", func(t *testing.T) {
		tm := model.TeamMember{
			TeamID:   team.ID,
			UserID:   member.ID,
			Role:     "member",
			JoinedAt: time.Now(),
		}
		require.NoError(t, repo.AddMember(ctx, &tm))
		assert.NotZero(t, tm.ID)
	})

	t.Run("duplicate_returns_ErrAlreadyExists", func(t *testing.T) {
		tm := model.TeamMember{
			TeamID:   team.ID,
			UserID:   member.ID,
			Role:     "pm",
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
	team := model.Team{Name: "Team R", PmID: pm.ID}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: "member", JoinedAt: time.Now()}
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
	team := model.Team{Name: "Team F", PmID: pm.ID}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: "member", JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindMember(ctx, team.ID, member.ID)
		require.NoError(t, err)
		assert.Equal(t, "member", found.Role)
		assert.Equal(t, team.ID, found.TeamID)
		assert.Equal(t, member.ID, found.UserID)
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

	pm := seedUser(t, db, "pm9")
	m1 := seedUser(t, db, "m1")
	m2 := seedUser(t, db, "m2")
	team := model.Team{Name: "Team LM", PmID: pm.ID}
	require.NoError(t, repo.Create(ctx, &team))

	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamID: team.ID, UserID: m1.ID, Role: "pm", JoinedAt: time.Now(),
	}))
	require.NoError(t, repo.AddMember(ctx, &model.TeamMember{
		TeamID: team.ID, UserID: m2.ID, Role: "member", JoinedAt: time.Now(),
	}))

	results, err := repo.ListMembers(ctx, team.ID)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Verify joined user display name and username are present
	usernames := map[string]bool{}
	for _, dto := range results {
		assert.NotEmpty(t, dto.DisplayName, "display_name should be populated from join")
		assert.NotEmpty(t, dto.Username, "username should be populated from join")
		usernames[dto.Username] = true
	}
	assert.True(t, usernames["m1"], "should contain m1")
	assert.True(t, usernames["m2"], "should contain m2")
}

func TestTeamRepo_ListMembers_EmptyTeam(t *testing.T) {
	db := setupTeamTestDB(t)
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
	team := model.Team{Name: "Team UM", PmID: pm.ID}
	require.NoError(t, repo.Create(ctx, &team))

	tm := model.TeamMember{TeamID: team.ID, UserID: member.ID, Role: "member", JoinedAt: time.Now()}
	require.NoError(t, repo.AddMember(ctx, &tm))

	tm.Role = "pm"
	require.NoError(t, repo.UpdateMember(ctx, &tm))

	found, err := repo.FindMember(ctx, team.ID, member.ID)
	require.NoError(t, err)
	assert.Equal(t, "pm", found.Role)
}
