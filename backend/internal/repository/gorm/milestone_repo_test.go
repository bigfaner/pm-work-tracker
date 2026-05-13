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

func setupMilestoneTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.MilestoneMap{}, &model.Milestone{}))
	return db
}

func seedMilestoneTeam(t *testing.T, db *gormlib.DB) *model.Team {
	t.Helper()
	u := model.User{Username: "ms_pm", DisplayName: "MS PM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "MS Team", PmKey: int64(u.ID), Code: "MSTT"}
	require.NoError(t, db.Create(&team).Error)
	team.BizKey = int64(team.ID)
	require.NoError(t, db.Save(&team).Error)
	return &team
}

func seedMilestoneMap(t *testing.T, db *gormlib.DB, teamBizKey int64, name string) *model.MilestoneMap {
	t.Helper()
	m := model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: time.Now().UnixNano()},
		TeamKey:   teamBizKey,
		MapName:   name,
		MapStatus: "draft",
	}
	require.NoError(t, db.Create(&m).Error)
	return &m
}

func createMilestone(t *testing.T, db *gormlib.DB, teamBizKey, mapBizKey int64, name, status string) *model.Milestone {
	t.Helper()
	m := model.Milestone{
		BaseModel:       model.BaseModel{BizKey: time.Now().UnixNano()},
		TeamKey:         teamBizKey,
		MilestoneMapKey: mapBizKey,
		MilestoneName:   name,
		MilestoneStatus: status,
	}
	require.NoError(t, db.Create(&m).Error)
	return &m
}

// --- Create ---

func TestMilestoneRepo_Create(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Q1 Plan")

	m := &model.Milestone{
		TeamKey:         team.BizKey,
		MilestoneMapKey: mmap.BizKey,
		MilestoneName:   "Phase 1",
		MilestoneStatus: "not_started",
	}
	err := repo.Create(ctx, m)
	require.NoError(t, err)
	assert.NotZero(t, m.ID)
}

// --- FindByID ---

func TestMilestoneRepo_FindByID(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Q1 Plan")
	created := createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase 1", "not_started")

	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Phase 1", found.MilestoneName)
}

func TestMilestoneRepo_FindByID_NotFound(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 99999)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- FindByBizKey ---

func TestMilestoneRepo_FindByBizKey(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Q1 Plan")
	created := createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase 1", "in_progress")

	found, err := repo.FindByBizKey(ctx, created.BizKey)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Phase 1", found.MilestoneName)
}

func TestMilestoneRepo_FindByBizKey_NotFound(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	_, err := repo.FindByBizKey(ctx, 123456)
	assert.ErrorIs(t, err, pkgerrors.ErrMilestoneNotFound)
}

// --- FindByBizKeys ---

func TestMilestoneRepo_FindByBizKeys(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Q1 Plan")
	ms1 := createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase 1", "not_started")
	ms2 := createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase 2", "not_started")

	result, err := repo.FindByBizKeys(ctx, []int64{ms1.BizKey, ms2.BizKey})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, ms1.BizKey)
	assert.Contains(t, result, ms2.BizKey)
}

func TestMilestoneRepo_FindByBizKeys_Empty(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	result, err := repo.FindByBizKeys(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestMilestoneRepo_FindByBizKeys_PartialMatch(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Q1 Plan")
	ms1 := createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase 1", "not_started")

	result, err := repo.FindByBizKeys(ctx, []int64{ms1.BizKey, 99999})
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, ms1.BizKey)
}

// --- Update ---

func TestMilestoneRepo_Update(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Q1 Plan")
	created := createMilestone(t, db, team.BizKey, mmap.BizKey, "Old Name", "not_started")

	err := repo.Update(ctx, created, map[string]interface{}{
		"milestone_name": "New Name",
	})
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", found.MilestoneName)
}

// --- ListByMap ---

func TestMilestoneRepo_ListByMap(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap1 := seedMilestoneMap(t, db, team.BizKey, "Map 1")
	mmap2 := seedMilestoneMap(t, db, team.BizKey, "Map 2")

	createMilestone(t, db, team.BizKey, mmap1.BizKey, "Phase A", "not_started")
	createMilestone(t, db, team.BizKey, mmap1.BizKey, "Phase B", "in_progress")
	createMilestone(t, db, team.BizKey, mmap2.BizKey, "Phase C", "not_started")

	items, err := repo.ListByMap(ctx, mmap1.BizKey)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestMilestoneRepo_ListByMap_ExcludesDeleted(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Map 1")
	createMilestone(t, db, team.BizKey, mmap.BizKey, "Active", "not_started")
	deleted := createMilestone(t, db, team.BizKey, mmap.BizKey, "Deleted", "not_started")

	err := repo.SoftDelete(ctx, deleted.ID)
	require.NoError(t, err)

	items, err := repo.ListByMap(ctx, mmap.BizKey)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Active", items[0].MilestoneName)
}

// --- ListByTeam ---

func TestMilestoneRepo_ListByTeam(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Map 1")

	createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase A", "not_started")
	createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase B", "cancelled")
	createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase C", "in_progress")

	// Without excludeCancelled: returns all 3
	items, err := repo.ListByTeam(ctx, team.BizKey, false)
	require.NoError(t, err)
	assert.Len(t, items, 3)

	// With excludeCancelled: returns 2
	items, err = repo.ListByTeam(ctx, team.BizKey, true)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestMilestoneRepo_ListByTeam_TeamScoped(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team1 := seedMilestoneTeam(t, db)
	// Create second team
	u2 := model.User{Username: "ms_pm2", DisplayName: "MS PM2", PasswordHash: "h"}
	require.NoError(t, db.Create(&u2).Error)
	team2 := model.Team{TeamName: "MS Team2", PmKey: int64(u2.ID), Code: "MST2"}
	require.NoError(t, db.Create(&team2).Error)
	team2.BizKey = int64(team2.ID)
	require.NoError(t, db.Save(&team2).Error)

	mmap1 := seedMilestoneMap(t, db, team1.BizKey, "Map 1")
	mmap2 := seedMilestoneMap(t, db, team2.BizKey, "Map 2")

	createMilestone(t, db, team1.BizKey, mmap1.BizKey, "Team1 MS", "not_started")
	createMilestone(t, db, team2.BizKey, mmap2.BizKey, "Team2 MS", "not_started")

	items, err := repo.ListByTeam(ctx, team1.BizKey, false)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Team1 MS", items[0].MilestoneName)
}

// --- SoftDelete ---

func TestMilestoneRepo_SoftDelete(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Map 1")
	created := createMilestone(t, db, team.BizKey, mmap.BizKey, "To Delete", "not_started")

	err := repo.SoftDelete(ctx, created.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, created.ID)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

func TestMilestoneRepo_SoftDelete_NotFound(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	err := repo.SoftDelete(ctx, 99999)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- DeleteByMap ---

func TestMilestoneRepo_DeleteByMap(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team := seedMilestoneTeam(t, db)
	mmap := seedMilestoneMap(t, db, team.BizKey, "Map 1")
	mmap2 := seedMilestoneMap(t, db, team.BizKey, "Map 2")

	ms1 := createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase A", "not_started")
	ms2 := createMilestone(t, db, team.BizKey, mmap.BizKey, "Phase B", "in_progress")
	ms3 := createMilestone(t, db, team.BizKey, mmap2.BizKey, "Phase C", "not_started")

	err := repo.DeleteByMap(ctx, mmap.BizKey)
	require.NoError(t, err)

	// Milestones under mmap should be deleted
	_, err = repo.FindByID(ctx, ms1.ID)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	_, err = repo.FindByID(ctx, ms2.ID)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)

	// Milestones under mmap2 should still exist
	found, err := repo.FindByID(ctx, ms3.ID)
	require.NoError(t, err)
	assert.Equal(t, "Phase C", found.MilestoneName)
}
