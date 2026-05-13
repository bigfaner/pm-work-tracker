package gorm_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	pkgerrors "pm-work-tracker/backend/internal/pkg/errors"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

func setupMilestoneMapTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.MilestoneMap{}))
	return db
}

func seedMilestoneMapTeam(t *testing.T, db *gormlib.DB) *model.Team {
	t.Helper()
	u := model.User{Username: "mm_pm", DisplayName: "MM PM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "MM Team", PmKey: int64(u.ID), Code: "MMAP"}
	require.NoError(t, db.Create(&team).Error)
	team.BizKey = int64(team.ID)
	require.NoError(t, db.Save(&team).Error)
	return &team
}

func createMilestoneMap(t *testing.T, db *gormlib.DB, teamBizKey int64, name, status string) *model.MilestoneMap {
	t.Helper()
	m := model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: time.Now().UnixNano()},
		TeamKey:   teamBizKey,
		MapName:   name,
		MapStatus: status,
	}
	require.NoError(t, db.Create(&m).Error)
	return &m
}

// --- Create ---

func TestMilestoneMapRepo_Create(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	team := seedMilestoneMapTeam(t, db)

	m := &model.MilestoneMap{
		TeamKey:   team.BizKey,
		MapName:   "Q1 Plan",
		MapStatus: "draft",
	}
	err := repo.Create(ctx, m)
	require.NoError(t, err)
	assert.NotZero(t, m.ID)
}

// --- FindByID ---

func TestMilestoneMapRepo_FindByID(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	team := seedMilestoneMapTeam(t, db)
	created := createMilestoneMap(t, db, team.BizKey, "Q1 Plan", "draft")

	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Q1 Plan", found.MapName)
}

func TestMilestoneMapRepo_FindByID_NotFound(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 99999)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- FindByBizKey ---

func TestMilestoneMapRepo_FindByBizKey(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	team := seedMilestoneMapTeam(t, db)
	created := createMilestoneMap(t, db, team.BizKey, "Q2 Plan", "active")

	found, err := repo.FindByBizKey(ctx, created.BizKey)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Q2 Plan", found.MapName)
}

func TestMilestoneMapRepo_FindByBizKey_NotFound(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	_, err := repo.FindByBizKey(ctx, 123456)
	assert.ErrorIs(t, err, pkgerrors.ErrMilestoneMapNotFound)
}

// --- Update ---

func TestMilestoneMapRepo_Update(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	team := seedMilestoneMapTeam(t, db)
	created := createMilestoneMap(t, db, team.BizKey, "Old Name", "draft")

	err := repo.Update(ctx, created, map[string]interface{}{
		"map_name": "New Name",
	})
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", found.MapName)
}

// --- List ---

func TestMilestoneMapRepo_List(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	team := seedMilestoneMapTeam(t, db)
	createMilestoneMap(t, db, team.BizKey, "Map A", "draft")
	createMilestoneMap(t, db, team.BizKey, "Map B", "active")

	result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Items, 2)
}

func TestMilestoneMapRepo_List_WithStatusFilter(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	team := seedMilestoneMapTeam(t, db)
	createMilestoneMap(t, db, team.BizKey, "Map A", "draft")
	createMilestoneMap(t, db, team.BizKey, "Map B", "active")

	activeStatus := "active"
	result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{Status: &activeStatus}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "active", result.Items[0].MapStatus)
}

func TestMilestoneMapRepo_List_TeamScoped(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	team1 := seedMilestoneMapTeam(t, db)
	// Create second team
	u2 := model.User{Username: "mm_pm2", DisplayName: "MM PM2", PasswordHash: "h"}
	require.NoError(t, db.Create(&u2).Error)
	team2 := model.Team{TeamName: "MM Team2", PmKey: int64(u2.ID), Code: "MMB2"}
	require.NoError(t, db.Create(&team2).Error)
	team2.BizKey = int64(team2.ID)
	require.NoError(t, db.Save(&team2).Error)

	createMilestoneMap(t, db, team1.BizKey, "Team1 Map", "draft")
	createMilestoneMap(t, db, team2.BizKey, "Team2 Map", "draft")

	result, err := repo.List(ctx, team1.BizKey, dto.MilestoneMapFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "Team1 Map", result.Items[0].MapName)
}

// --- SoftDelete ---

func TestMilestoneMapRepo_SoftDelete(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	team := seedMilestoneMapTeam(t, db)
	created := createMilestoneMap(t, db, team.BizKey, "To Delete", "draft")

	err := repo.SoftDelete(ctx, created.ID)
	require.NoError(t, err)

	// Should not be found via FindByID (which applies NotDeleted scope)
	_, err = repo.FindByID(ctx, created.ID)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

func TestMilestoneMapRepo_SoftDelete_NotFound(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	err := repo.SoftDelete(ctx, 99999)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}
