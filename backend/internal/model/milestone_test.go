package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/model"
)

func TestMilestoneMap_TableName(t *testing.T) {
	m := model.MilestoneMap{}
	assert.Equal(t, "pmw_milestone_maps", m.TableName())
}

func TestMilestone_TableName(t *testing.T) {
	m := model.Milestone{}
	assert.Equal(t, "pmw_milestones", m.TableName())
}

func TestMilestoneMap_AutoMigrate(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.MilestoneMap{})
	require.NoError(t, err)
}

func TestMilestone_AutoMigrate(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Milestone{})
	require.NoError(t, err)
}

func TestMilestoneMap_Defaults(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.MilestoneMap{})
	require.NoError(t, err)

	m := model.MilestoneMap{
		TeamKey:     100,
		CreatorKey:  200,
		AssigneeKey: 300,
		MapName:     "Test Map",
	}
	require.NoError(t, db.Create(&m).Error)

	var fetched model.MilestoneMap
	require.NoError(t, db.First(&fetched, m.ID).Error)
	assert.Equal(t, "planning", fetched.MapStatus, "map_status should default to 'planning'")
	assert.Equal(t, "", fetched.MapDesc, "map_desc should default to empty string")
}

func TestMilestone_Defaults(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Milestone{})
	require.NoError(t, err)

	ms := model.Milestone{
		TeamKey:         100,
		MilestoneMapKey: 200,
		MilestoneName:   "M1",
	}
	require.NoError(t, db.Create(&ms).Error)

	var fetched model.Milestone
	require.NoError(t, db.First(&fetched, ms.ID).Error)
	assert.Equal(t, "not_started", fetched.MilestoneStatus, "milestone_status should default to 'not_started'")
	assert.Equal(t, "", fetched.MilestoneDesc, "milestone_desc should default to empty string")
}

// AC3 & AC4: uk_milestone_maps_team_name_deleted enforces team-scoped name uniqueness with soft-delete support
func TestMilestoneMap_TeamScopedNameUnique(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.MilestoneMap{})
	require.NoError(t, err)

	m1 := model.MilestoneMap{
		TeamKey:     1,
		CreatorKey:  10,
		AssigneeKey: 20,
		MapName:     "Map A",
	}
	require.NoError(t, db.Create(&m1).Error)

	// Same team, same name -> should fail
	m2 := model.MilestoneMap{
		TeamKey:     1,
		CreatorKey:  10,
		AssigneeKey: 20,
		MapName:     "Map A",
	}
	err = db.Create(&m2).Error
	assert.Error(t, err, "duplicate team_key + map_name should be rejected")
}

func TestMilestoneMap_TeamScopedNameAllowsDifferentTeams(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.MilestoneMap{})
	require.NoError(t, err)

	m1 := model.MilestoneMap{
		TeamKey:     1,
		CreatorKey:  10,
		AssigneeKey: 20,
		MapName:     "Shared Name",
	}
	require.NoError(t, db.Create(&m1).Error)

	// Different team, same name -> should succeed
	m2 := model.MilestoneMap{
		TeamKey:     2,
		CreatorKey:  10,
		AssigneeKey: 20,
		MapName:     "Shared Name",
	}
	require.NoError(t, db.Create(&m2).Error, "different team should allow same map_name")
}

// AC4: soft-delete allows recreating with same name (requires full DDL index including deleted_flag, deleted_time)
func TestMilestoneMap_SoftDeleteAllowsRecreate(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.MilestoneMap{})
	require.NoError(t, err)

	// Drop the GORM-generated partial unique index and recreate with deleted_flag, deleted_time
	db.Exec("DROP INDEX IF EXISTS uk_milestone_maps_team_name_deleted")
	db.Exec("CREATE UNIQUE INDEX uk_milestone_maps_team_name_deleted ON pmw_milestone_maps(team_key, map_name, deleted_flag, deleted_time)")

	m1 := model.MilestoneMap{
		TeamKey:     1,
		CreatorKey:  10,
		AssigneeKey: 20,
		MapName:     "Map B",
	}
	require.NoError(t, db.Create(&m1).Error)

	// Soft-delete the first record
	now := time.Now()
	err = db.Model(&model.MilestoneMap{}).Where("id = ?", m1.ID).Updates(map[string]interface{}{
		"deleted_flag": 1,
		"deleted_time": now,
	}).Error
	require.NoError(t, err)

	// Recreate with same team_key + map_name -> should succeed
	m2 := model.MilestoneMap{
		TeamKey:     1,
		CreatorKey:  10,
		AssigneeKey: 20,
		MapName:     "Map B",
	}
	require.NoError(t, db.Create(&m2).Error, "should allow recreating after soft-delete")
}

// AC3 & AC4: uk_milestones_map_name_deleted enforces map-scoped name uniqueness
func TestMilestone_MapScopedNameUnique(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Milestone{})
	require.NoError(t, err)

	ms1 := model.Milestone{
		TeamKey:         1,
		MilestoneMapKey: 100,
		MilestoneName:   "Phase 1",
	}
	require.NoError(t, db.Create(&ms1).Error)

	// Same map, same name -> should fail
	ms2 := model.Milestone{
		TeamKey:         1,
		MilestoneMapKey: 100,
		MilestoneName:   "Phase 1",
	}
	err = db.Create(&ms2).Error
	assert.Error(t, err, "duplicate milestone_map_key + milestone_name should be rejected")
}

func TestMilestone_MapScopedNameAllowsDifferentMaps(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Milestone{})
	require.NoError(t, err)

	ms1 := model.Milestone{
		TeamKey:         1,
		MilestoneMapKey: 100,
		MilestoneName:   "Phase 1",
	}
	require.NoError(t, db.Create(&ms1).Error)

	// Different map, same name -> should succeed
	ms2 := model.Milestone{
		TeamKey:         1,
		MilestoneMapKey: 200,
		MilestoneName:   "Phase 1",
	}
	require.NoError(t, db.Create(&ms2).Error, "different map should allow same milestone_name")
}

// AC3: idx_milestone_maps_team_status composite index works
func TestMilestoneMap_TeamStatusIndex(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.MilestoneMap{})
	require.NoError(t, err)

	for i, status := range []string{"planning", "executing", "completed"} {
		m := model.MilestoneMap{
			TeamKey:     1,
			CreatorKey:  10,
			AssigneeKey: 20,
			MapName:     "Map " + string(rune('A'+i)),
			MapStatus:   status,
		}
		require.NoError(t, db.Create(&m).Error)
	}

	var results []model.MilestoneMap
	err = db.Where("team_key = ? AND map_status = ?", 1, "executing").Find(&results).Error
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

// AC3: idx_milestones_team_status composite index works
func TestMilestone_TeamStatusIndex(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Milestone{})
	require.NoError(t, err)

	for i, status := range []string{"not_started", "in_progress", "completed"} {
		ms := model.Milestone{
			TeamKey:         1,
			MilestoneMapKey: 100,
			MilestoneName:   "MS " + string(rune('A'+i)),
			MilestoneStatus: status,
		}
		require.NoError(t, db.Create(&ms).Error)
	}

	var results []model.Milestone
	err = db.Where("team_key = ? AND milestone_status = ?", 1, "in_progress").Find(&results).Error
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}
