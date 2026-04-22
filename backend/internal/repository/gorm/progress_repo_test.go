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

func setupProgressTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.ProgressRecord{},
	))
	return db
}

func seedProgressData(t *testing.T, db *gormlib.DB) (*model.User, *model.Team, *model.MainItem, *model.SubItem) {
	t.Helper()
	u := model.User{Username: "pr_pm", DisplayName: "PR PM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "PR Team", PmID: u.ID, Code: "PRTE"}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamID: team.ID, Code: "MI-PR01", Title: "Main", Priority: "P1", ProposerID: u.ID, Status: "pending"}
	require.NoError(t, db.Create(&mi).Error)
	si := model.SubItem{TeamID: team.ID, MainItemID: mi.ID, Title: "Sub", Priority: "P1", Status: "progressing"}
	require.NoError(t, db.Create(&si).Error)
	return &u, &team, &mi, &si
}

func createProgressRecord(t *testing.T, db *gormlib.DB, subItemID, teamID, authorID uint, completion float64, achievement string, createdAt time.Time) *model.ProgressRecord {
	t.Helper()
	record := model.ProgressRecord{
		SubItemID:   subItemID,
		TeamID:      teamID,
		AuthorID:    authorID,
		Completion:  completion,
		Achievement: achievement,
		CreatedAt:   createdAt,
	}
	require.NoError(t, db.Create(&record).Error)
	return &record
}

// --- Create ---

func TestProgressRepo_Create(t *testing.T) {
	db := setupProgressTestDB(t)
	repo := gormrepo.NewGormProgressRepo(db)
	ctx := context.Background()

	u, team, _, si := seedProgressData(t, db)
	record := &model.ProgressRecord{
		SubItemID:   si.ID,
		TeamID:      team.ID,
		AuthorID:    u.ID,
		Completion:  30.0,
		Achievement: "Started work",
		CreatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, record))
	assert.NotZero(t, record.ID)
}

// --- FindByID ---

func TestProgressRepo_FindByID(t *testing.T) {
	db := setupProgressTestDB(t)
	repo := gormrepo.NewGormProgressRepo(db)
	ctx := context.Background()

	u, team, _, si := seedProgressData(t, db)
	record := createProgressRecord(t, db, si.ID, team.ID, u.ID, 50.0, "Half done", time.Now())

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, record.ID)
		require.NoError(t, err)
		assert.Equal(t, 50.0, found.Completion)
		assert.Equal(t, "Half done", found.Achievement)
		assert.Equal(t, team.ID, found.TeamID)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

// --- ListBySubItem ---

func TestProgressRepo_ListBySubItem(t *testing.T) {
	db := setupProgressTestDB(t)
	repo := gormrepo.NewGormProgressRepo(db)
	ctx := context.Background()

	u, team, _, si := seedProgressData(t, db)

	base := time.Now()
	createProgressRecord(t, db, si.ID, team.ID, u.ID, 20.0, "First", base.Add(-2*time.Hour))
	createProgressRecord(t, db, si.ID, team.ID, u.ID, 50.0, "Second", base.Add(-1*time.Hour))
	createProgressRecord(t, db, si.ID, team.ID, u.ID, 80.0, "Third", base)

	t.Run("returns_ordered_by_created_at_asc", func(t *testing.T) {
		records, err := repo.ListBySubItem(ctx, team.ID, si.ID)
		require.NoError(t, err)
		assert.Len(t, records, 3)
		assert.Equal(t, 20.0, records[0].Completion)
		assert.Equal(t, 50.0, records[1].Completion)
		assert.Equal(t, 80.0, records[2].Completion)
	})

	t.Run("team_isolation", func(t *testing.T) {
		// Different team should see nothing
		u2 := model.User{Username: "other_pr_pm", DisplayName: "OP", PasswordHash: "h"}
		require.NoError(t, db.Create(&u2).Error)
		team2 := model.Team{Name: "PR Team2", PmID: u2.ID, Code: "PRT2"}
		require.NoError(t, db.Create(&team2).Error)
		records, err := repo.ListBySubItem(ctx, team2.ID, si.ID)
		require.NoError(t, err)
		assert.Empty(t, records)
	})

	t.Run("empty_when_none", func(t *testing.T) {
		records, err := repo.ListBySubItem(ctx, team.ID, 9999)
		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

// --- LatestBySubItem ---

func TestProgressRepo_LatestBySubItem(t *testing.T) {
	db := setupProgressTestDB(t)
	repo := gormrepo.NewGormProgressRepo(db)
	ctx := context.Background()

	u, team, _, si := seedProgressData(t, db)

	t.Run("nil_when_none", func(t *testing.T) {
		found, err := repo.LatestBySubItem(ctx, si.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("returns_latest", func(t *testing.T) {
		base := time.Now()
		createProgressRecord(t, db, si.ID, team.ID, u.ID, 20.0, "First", base.Add(-2*time.Hour))
		createProgressRecord(t, db, si.ID, team.ID, u.ID, 80.0, "Latest", base)

		found, err := repo.LatestBySubItem(ctx, si.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, 80.0, found.Completion)
		assert.Equal(t, "Latest", found.Achievement)
	})
}

// --- UpdateCompletion ---

func TestProgressRepo_UpdateCompletion(t *testing.T) {
	db := setupProgressTestDB(t)
	repo := gormrepo.NewGormProgressRepo(db)
	ctx := context.Background()

	u, team, _, si := seedProgressData(t, db)
	record := createProgressRecord(t, db, si.ID, team.ID, u.ID, 50.0, "Original", time.Now())

	require.NoError(t, repo.UpdateCompletion(ctx, record.ID, 75.0))

	found, err := repo.FindByID(ctx, record.ID)
	require.NoError(t, err)
	assert.Equal(t, 75.0, found.Completion)
	assert.True(t, found.IsPMCorrect)
	// Achievement should not change
	assert.Equal(t, "Original", found.Achievement)
}

func TestProgressRepo_UpdateCompletion_NotFound(t *testing.T) {
	db := setupProgressTestDB(t)
	repo := gormrepo.NewGormProgressRepo(db)
	ctx := context.Background()

	err := repo.UpdateCompletion(ctx, 9999, 75.0)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}
