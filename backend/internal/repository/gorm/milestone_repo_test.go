//nolint:misspell // "cancelled" is a domain status value per PRD/API contract.
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

func setupMilestoneTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.MilestoneMap{}, &model.Milestone{}))
	return db
}

func seedMilestoneData(t *testing.T, db *gormlib.DB) (*model.Team, *model.MilestoneMap) {
	t.Helper()
	u := model.User{Username: "ms_user", DisplayName: "MS User", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "MS Team", PmKey: int64(u.ID), Code: "MSTE"}
	require.NoError(t, db.Create(&team).Error)
	team.BizKey = int64(team.ID)
	require.NoError(t, db.Save(&team).Error)

	mm := model.MilestoneMap{
		TeamKey:     team.BizKey,
		CreatorKey:  int64(u.ID),
		AssigneeKey: int64(u.ID),
		MapName:     "Test Map",
		MapStatus:   "planning",
	}
	require.NoError(t, db.Create(&mm).Error)
	mm.BizKey = int64(mm.ID)
	require.NoError(t, db.Save(&mm).Error)
	return &team, &mm
}

func createMilestone(t *testing.T, db *gormlib.DB, teamBizKey, mapBizKey int64, name, status string, expectedEndDate *time.Time) *model.Milestone {
	t.Helper()
	m := model.Milestone{
		TeamKey:         teamBizKey,
		MilestoneMapKey: mapBizKey,
		MilestoneName:   name,
		MilestoneDesc:   "desc",
		MilestoneStatus: status,
		ExpectedEndDate: expectedEndDate,
	}
	require.NoError(t, db.Create(&m).Error)
	m.BizKey = int64(m.ID)
	require.NoError(t, db.Save(&m).Error)
	return &m
}

func boolPtr(b bool) *bool { return &b }
func uintPtr(u uint) *uint { return &u }

// --- Create ---

func TestMilestoneRepo_Create(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)
	endDate := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	m := &model.Milestone{
		TeamKey:         team.BizKey,
		MilestoneMapKey: mm.BizKey,
		MilestoneName:   "Sprint 1",
		MilestoneDesc:   "First sprint",
		MilestoneStatus: "not_started",
		ExpectedEndDate: &endDate,
	}
	require.NoError(t, repo.Create(ctx, m))
	assert.NotZero(t, m.ID)

	// Verify field mapping
	found, err := repo.FindByID(ctx, m.ID)
	require.NoError(t, err)
	assert.Equal(t, "Sprint 1", found.MilestoneName)
	assert.Equal(t, "First sprint", found.MilestoneDesc)
	assert.Equal(t, "not_started", found.MilestoneStatus)
	assert.Equal(t, team.BizKey, found.TeamKey)
	assert.Equal(t, mm.BizKey, found.MilestoneMapKey)
}

// --- FindByID ---

func TestMilestoneRepo_FindByID(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)
	m := createMilestone(t, db, team.BizKey, mm.BizKey, "Find Me", "not_started", nil)

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, m.ID)
		require.NoError(t, err)
		assert.Equal(t, "Find Me", found.MilestoneName)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

// --- FindByBizKey ---

func TestMilestoneRepo_FindByBizKey(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)
	m := createMilestone(t, db, team.BizKey, mm.BizKey, "BizKey Test", "not_started", nil)

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByBizKey(ctx, m.BizKey)
		require.NoError(t, err)
		assert.Equal(t, "BizKey Test", found.MilestoneName)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByBizKey(ctx, 99999)
		assert.ErrorIs(t, err, gormlib.ErrRecordNotFound)
	})
}

// --- FindBatchByBizKeys ---

func TestMilestoneRepo_FindBatchByBizKeys(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)
	m1 := createMilestone(t, db, team.BizKey, mm.BizKey, "M1", "not_started", nil)
	m2 := createMilestone(t, db, team.BizKey, mm.BizKey, "M2", "in_progress", nil)
	_ = createMilestone(t, db, team.BizKey, mm.BizKey, "M3", "completed", nil)

	t.Run("batch_lookup", func(t *testing.T) {
		result, err := repo.FindBatchByBizKeys(ctx, []int64{m1.BizKey, m2.BizKey, 99999})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, m1.BizKey)
		assert.Contains(t, result, m2.BizKey)
		assert.Equal(t, "M1", result[m1.BizKey].MilestoneName)
	})

	t.Run("empty_keys", func(t *testing.T) {
		result, err := repo.FindBatchByBizKeys(ctx, []int64{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

// --- Update ---

func TestMilestoneRepo_Update(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)
	m := createMilestone(t, db, team.BizKey, mm.BizKey, "Update Me", "not_started", nil)

	fields := map[string]interface{}{
		"milestone_name":   "Updated Name",
		"milestone_status": "in_progress",
	}
	require.NoError(t, repo.Update(ctx, m, fields))

	found, err := repo.FindByID(ctx, m.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.MilestoneName)
	assert.Equal(t, "in_progress", found.MilestoneStatus)
}

func TestMilestoneRepo_Update_NotFound(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, _ := seedMilestoneData(t, db)
	fakeItem := &model.Milestone{BaseModel: model.BaseModel{ID: 9999}, TeamKey: team.BizKey}
	fields := map[string]interface{}{"milestone_status": "in_progress"}
	err := repo.Update(ctx, fakeItem, fields)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- ListByMap ---

func TestMilestoneRepo_ListByMap(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)

	endDate1 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate2 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	endDate3 := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	createMilestone(t, db, team.BizKey, mm.BizKey, "August MS", "not_started", &endDate1)
	createMilestone(t, db, team.BizKey, mm.BizKey, "July MS", "in_progress", &endDate2)
	createMilestone(t, db, team.BizKey, mm.BizKey, "September MS", "completed", &endDate3)

	t.Run("returns_sorted_by_expected_end_date_asc", func(t *testing.T) {
		items, err := repo.ListByMap(ctx, mm.BizKey)
		require.NoError(t, err)
		assert.Len(t, items, 3)
		assert.Equal(t, "July MS", items[0].MilestoneName)
		assert.Equal(t, "August MS", items[1].MilestoneName)
		assert.Equal(t, "September MS", items[2].MilestoneName)
	})

	t.Run("excludes_soft_deleted", func(t *testing.T) {
		deleted := createMilestone(t, db, team.BizKey, mm.BizKey, "Deleted MS", "not_started", nil)
		require.NoError(t, repo.SoftDelete(ctx, deleted.ID))

		items, err := repo.ListByMap(ctx, mm.BizKey)
		require.NoError(t, err)
		assert.Len(t, items, 3) // original 3 only
		for _, item := range items {
			assert.NotEqual(t, "Deleted MS", item.MilestoneName)
		}
	})

	t.Run("empty_for_nonexistent_map", func(t *testing.T) {
		items, err := repo.ListByMap(ctx, 99999)
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}

// --- ListByTeam ---

func TestMilestoneRepo_ListByTeam(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)

	createMilestone(t, db, team.BizKey, mm.BizKey, "Alpha MS", "not_started", nil)
	createMilestone(t, db, team.BizKey, mm.BizKey, "Beta MS", "in_progress", nil)
	createMilestone(t, db, team.BizKey, mm.BizKey, "Gamma MS", "cancelled", nil)
	createMilestone(t, db, team.BizKey, mm.BizKey, "Delta MS", "completed", nil)

	t.Run("all_for_team", func(t *testing.T) {
		items, err := repo.ListByTeam(ctx, team.BizKey, dto.MilestoneTeamFilter{})
		require.NoError(t, err)
		assert.Len(t, items, 4)
	})

	t.Run("filter_by_name", func(t *testing.T) {
		items, err := repo.ListByTeam(ctx, team.BizKey, dto.MilestoneTeamFilter{Name: strPtr("Alpha")})
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, "Alpha MS", items[0].MilestoneName)
	})

	t.Run("filter_by_status", func(t *testing.T) {
		items, err := repo.ListByTeam(ctx, team.BizKey, dto.MilestoneTeamFilter{Status: strPtr("in_progress")})
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, "Beta MS", items[0].MilestoneName)
	})

	t.Run("exclude_cancelled", func(t *testing.T) {
		items, err := repo.ListByTeam(ctx, team.BizKey, dto.MilestoneTeamFilter{ExcludeCancelled: boolPtr(true)})
		require.NoError(t, err)
		assert.Len(t, items, 3)
		for _, item := range items {
			assert.NotEqual(t, "cancelled", item.MilestoneStatus)
		}
	})

	t.Run("combined_filters", func(t *testing.T) {
		items, err := repo.ListByTeam(ctx, team.BizKey, dto.MilestoneTeamFilter{
			Status:           strPtr("not_started"),
			ExcludeCancelled: boolPtr(true),
		})
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, "Alpha MS", items[0].MilestoneName)
	})

	t.Run("team_isolation", func(t *testing.T) {
		// Create another team with no milestones
		u2 := model.User{Username: "ms_other", DisplayName: "MS Other", PasswordHash: "h"}
		require.NoError(t, db.Create(&u2).Error)
		team2 := model.Team{TeamName: "MS Team2", PmKey: int64(u2.ID), Code: "MST2"}
		require.NoError(t, db.Create(&team2).Error)
		team2.BizKey = int64(team2.ID)
		require.NoError(t, db.Save(&team2).Error)

		items, err := repo.ListByTeam(ctx, team2.BizKey, dto.MilestoneTeamFilter{})
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}

// --- SoftDelete ---

func TestMilestoneRepo_SoftDelete(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)
	m := createMilestone(t, db, team.BizKey, mm.BizKey, "Delete Me", "not_started", nil)

	require.NoError(t, repo.SoftDelete(ctx, m.ID))

	// FindByID should return not-found
	_, err := repo.FindByID(ctx, m.ID)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)

	// FindByBizKey should also exclude
	_, err = repo.FindByBizKey(ctx, m.BizKey)
	assert.ErrorIs(t, err, gormlib.ErrRecordNotFound)

	// Verify deleted_flag was set directly in DB
	var raw model.Milestone
	require.NoError(t, db.Unscoped().First(&raw, m.ID).Error)
	assert.Equal(t, 1, raw.DeletedFlag)
	assert.False(t, raw.DeletedTime.IsZero())
}

func TestMilestoneRepo_SoftDelete_ListByMapExcludesDeleted(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)
	active := createMilestone(t, db, team.BizKey, mm.BizKey, "Active MS", "not_started", nil)
	deleted := createMilestone(t, db, team.BizKey, mm.BizKey, "Deleted MS", "not_started", nil)

	require.NoError(t, repo.SoftDelete(ctx, deleted.ID))

	items, err := repo.ListByMap(ctx, mm.BizKey)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, active.ID, items[0].ID)
}

// --- SoftDeleteByMap ---

func TestMilestoneRepo_SoftDeleteByMap(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)
	createMilestone(t, db, team.BizKey, mm.BizKey, "MS1", "not_started", nil)
	createMilestone(t, db, team.BizKey, mm.BizKey, "MS2", "in_progress", nil)
	createMilestone(t, db, team.BizKey, mm.BizKey, "MS3", "completed", nil)

	require.NoError(t, repo.SoftDeleteByMap(ctx, mm.BizKey))

	items, err := repo.ListByMap(ctx, mm.BizKey)
	require.NoError(t, err)
	assert.Empty(t, items)

	// Verify all are soft-deleted in DB
	var count int64
	require.NoError(t, db.Model(&model.Milestone{}).Unscoped().Where("milestone_map_key = ? AND deleted_flag = 1", mm.BizKey).Count(&count).Error)
	assert.Equal(t, int64(3), count)
}

func TestMilestoneRepo_SoftDeleteByMap_DoesNotAffectOtherMaps(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm1 := seedMilestoneData(t, db)

	// Create second map
	mm2 := model.MilestoneMap{
		TeamKey:     team.BizKey,
		CreatorKey:  mm1.CreatorKey,
		AssigneeKey: mm1.AssigneeKey,
		MapName:     "Second Map",
		MapStatus:   "planning",
	}
	require.NoError(t, db.Create(&mm2).Error)

	createMilestone(t, db, team.BizKey, mm1.BizKey, "Map1 MS", "not_started", nil)
	createMilestone(t, db, team.BizKey, mm2.BizKey, "Map2 MS", "not_started", nil)

	require.NoError(t, repo.SoftDeleteByMap(ctx, mm1.BizKey))

	// mm2's milestones should still be present
	items, err := repo.ListByMap(ctx, mm2.BizKey)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Map2 MS", items[0].MilestoneName)
}

// --- ExistsByNameAndMap ---

func TestMilestoneRepo_ExistsByNameAndMap(t *testing.T) {
	db := setupMilestoneTestDB(t)
	repo := gormrepo.NewGormMilestoneRepo(db)
	ctx := context.Background()

	team, mm := seedMilestoneData(t, db)
	m := createMilestone(t, db, team.BizKey, mm.BizKey, "Unique Name", "not_started", nil)

	t.Run("exists", func(t *testing.T) {
		exists, err := repo.ExistsByNameAndMap(ctx, mm.BizKey, "Unique Name", nil)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("not_exists_different_name", func(t *testing.T) {
		exists, err := repo.ExistsByNameAndMap(ctx, mm.BizKey, "Different Name", nil)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("not_exists_different_map", func(t *testing.T) {
		exists, err := repo.ExistsByNameAndMap(ctx, 99999, "Unique Name", nil)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("exclude_self", func(t *testing.T) {
		exists, err := repo.ExistsByNameAndMap(ctx, mm.BizKey, "Unique Name", uintPtr(m.ID))
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("does_not_exclude_other", func(t *testing.T) {
		exists, err := repo.ExistsByNameAndMap(ctx, mm.BizKey, "Unique Name", uintPtr(9999))
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("excludes_soft_deleted", func(t *testing.T) {
		deleted := createMilestone(t, db, team.BizKey, mm.BizKey, "Deleted Name", "not_started", nil)
		require.NoError(t, repo.SoftDelete(ctx, deleted.ID))

		exists, err := repo.ExistsByNameAndMap(ctx, mm.BizKey, "Deleted Name", nil)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
