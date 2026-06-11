package gorm_test

import (
	"context"
	"strconv"
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

func seedMilestoneMapData(t *testing.T, db *gormlib.DB) (*model.User, *model.Team) {
	t.Helper()
	u := model.User{Username: "mm_user", DisplayName: "MM User", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "MM Team", PmKey: int64(u.ID), Code: "MMTE"}
	require.NoError(t, db.Create(&team).Error)
	team.BizKey = int64(team.ID)
	require.NoError(t, db.Save(&team).Error)
	return &u, &team
}

func createMilestoneMap(t *testing.T, db *gormlib.DB, teamBizKey, creatorKey, assigneeKey int64, name, status string) *model.MilestoneMap {
	t.Helper()
	m := model.MilestoneMap{
		TeamKey:     teamBizKey,
		CreatorKey:  creatorKey,
		AssigneeKey: assigneeKey,
		MapName:     name,
		MapDesc:     "desc",
		MapStatus:   status,
	}
	require.NoError(t, db.Create(&m).Error)
	return &m
}

func strPtr(s string) *string { return &s }

// --- Create ---

func TestMilestoneMapRepo_Create(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	u, team := seedMilestoneMapData(t, db)
	now := time.Now()
	m := &model.MilestoneMap{
		TeamKey:         team.BizKey,
		CreatorKey:      int64(u.ID),
		AssigneeKey:     int64(u.ID),
		MapName:         "Q3 Plan",
		MapDesc:         "Quarterly plan",
		MapStatus:       "planning",
		PlanStartDate:   &now,
		ExpectedEndDate: &now,
	}
	require.NoError(t, repo.Create(ctx, m))
	assert.NotZero(t, m.ID)

	// Verify field mapping
	found, err := repo.FindByID(ctx, m.ID)
	require.NoError(t, err)
	assert.Equal(t, "Q3 Plan", found.MapName)
	assert.Equal(t, "Quarterly plan", found.MapDesc)
	assert.Equal(t, "planning", found.MapStatus)
	assert.Equal(t, team.BizKey, found.TeamKey)
	assert.Equal(t, int64(u.ID), found.CreatorKey)
	assert.Equal(t, int64(u.ID), found.AssigneeKey)
}

// --- FindByID ---

func TestMilestoneMapRepo_FindByID(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	u, team := seedMilestoneMapData(t, db)
	m := createMilestoneMap(t, db, team.BizKey, int64(u.ID), int64(u.ID), "Find Me", "planning")

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, m.ID)
		require.NoError(t, err)
		assert.Equal(t, "Find Me", found.MapName)
		assert.Equal(t, team.BizKey, found.TeamKey)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

// --- FindByBizKey ---

func TestMilestoneMapRepo_FindByBizKey(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	u, team := seedMilestoneMapData(t, db)
	m := createMilestoneMap(t, db, team.BizKey, int64(u.ID), int64(u.ID), "BizKey Test", "planning")

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByBizKey(ctx, m.BizKey)
		require.NoError(t, err)
		assert.Equal(t, "BizKey Test", found.MapName)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByBizKey(ctx, 99999)
		assert.ErrorIs(t, err, gormlib.ErrRecordNotFound)
	})
}

// --- Update ---

func TestMilestoneMapRepo_Update(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	u, team := seedMilestoneMapData(t, db)
	m := createMilestoneMap(t, db, team.BizKey, int64(u.ID), int64(u.ID), "Update Me", "planning")

	fields := map[string]interface{}{
		"map_name":   "Updated Name",
		"map_status": "reviewed",
	}
	require.NoError(t, repo.Update(ctx, m, fields))

	found, err := repo.FindByID(ctx, m.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.MapName)
	assert.Equal(t, "reviewed", found.MapStatus)
}

func TestMilestoneMapRepo_Update_NotFound(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	_, team := seedMilestoneMapData(t, db)
	fakeItem := &model.MilestoneMap{BaseModel: model.BaseModel{ID: 9999}, TeamKey: team.BizKey}
	fields := map[string]interface{}{"map_status": "reviewed"}
	err := repo.Update(ctx, fakeItem, fields)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- List with filters ---

func TestMilestoneMapRepo_List(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	u, team := seedMilestoneMapData(t, db)

	createMilestoneMap(t, db, team.BizKey, int64(u.ID), int64(u.ID), "Alpha Plan", "planning")
	createMilestoneMap(t, db, team.BizKey, int64(u.ID), int64(u.ID), "Beta Plan", "executing")
	createMilestoneMap(t, db, team.BizKey, int64(u.ID), int64(u.ID), "Gamma Review", "reviewed")

	// Another team - should not appear
	u2 := model.User{Username: "mm_other", DisplayName: "MM Other", PasswordHash: "h"}
	require.NoError(t, db.Create(&u2).Error)
	team2 := model.Team{TeamName: "MM Team2", PmKey: int64(u2.ID), Code: "MMT2"}
	require.NoError(t, db.Create(&team2).Error)
	team2.BizKey = int64(team2.ID)
	require.NoError(t, db.Save(&team2).Error)
	createMilestoneMap(t, db, team2.BizKey, int64(u2.ID), int64(u2.ID), "Other Team Map", "planning")

	t.Run("all_for_team", func(t *testing.T) {
		result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("filter_by_name_like", func(t *testing.T) {
		result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{Name: strPtr("Plan")}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		for _, item := range result.Items {
			assert.Contains(t, item.MapName, "Plan")
		}
	})

	t.Run("filter_by_status", func(t *testing.T) {
		result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{Status: strPtr("executing")}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Beta Plan", result.Items[0].MapName)
	})

	t.Run("filter_by_assignee", func(t *testing.T) {
		assigneeStr := strconv.FormatInt(int64(u.ID), 10)
		result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{AssigneeKey: &assigneeStr}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("filter_by_assignee_invalid", func(t *testing.T) {
		result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{AssigneeKey: strPtr("notanumber")}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total) // invalid key returns empty
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{}, dto.Pagination{Page: 1, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 2, result.Size)
	})

	t.Run("pagination_page2", func(t *testing.T) {
		result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{}, dto.Pagination{Page: 2, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 1)
	})

	t.Run("default_pagination", func(t *testing.T) {
		result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{}, dto.Pagination{})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.Size)
	})

	t.Run("team_isolation", func(t *testing.T) {
		result, err := repo.List(ctx, team2.BizKey, dto.MilestoneMapFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Other Team Map", result.Items[0].MapName)
	})
}

// --- SoftDelete ---

func TestMilestoneMapRepo_SoftDelete(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	u, team := seedMilestoneMapData(t, db)
	m := createMilestoneMap(t, db, team.BizKey, int64(u.ID), int64(u.ID), "Delete Me", "planning")

	require.NoError(t, repo.SoftDelete(ctx, m.ID))

	// FindByID should return not-found (excludes soft-deleted)
	_, err := repo.FindByID(ctx, m.ID)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)

	// FindByBizKey should also exclude
	_, err = repo.FindByBizKey(ctx, m.BizKey)
	assert.ErrorIs(t, err, gormlib.ErrRecordNotFound)

	// Verify deleted_flag was set directly in DB
	var raw model.MilestoneMap
	require.NoError(t, db.Unscoped().First(&raw, m.ID).Error)
	assert.Equal(t, 1, raw.DeletedFlag)
	assert.False(t, raw.DeletedTime.IsZero())
}

func TestMilestoneMapRepo_SoftDelete_ListExcludesDeleted(t *testing.T) {
	db := setupMilestoneMapTestDB(t)
	repo := gormrepo.NewGormMilestoneMapRepo(db)
	ctx := context.Background()

	u, team := seedMilestoneMapData(t, db)
	active := createMilestoneMap(t, db, team.BizKey, int64(u.ID), int64(u.ID), "Active Map", "planning")
	deleted := createMilestoneMap(t, db, team.BizKey, int64(u.ID), int64(u.ID), "Deleted Map", "planning")

	require.NoError(t, repo.SoftDelete(ctx, deleted.ID))

	result, err := repo.List(ctx, team.BizKey, dto.MilestoneMapFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, active.ID, result.Items[0].ID)
}
