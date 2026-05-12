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

func TestMilestoneMap_Defaults(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.MilestoneMap{})
	require.NoError(t, err)

	m := model.MilestoneMap{
		TeamKey:   12345,
		MapName:   "Q2 Roadmap",
		MapDesc:   "Quarterly plan",
		MapStatus: "planning",
	}
	require.NoError(t, db.Create(&m).Error)

	var fetched model.MilestoneMap
	db.First(&fetched, "map_name = ?", "Q2 Roadmap")
	assert.Equal(t, "Q2 Roadmap", fetched.MapName)
	assert.Equal(t, "planning", fetched.MapStatus)
	assert.Equal(t, int64(12345), fetched.TeamKey)
}

func TestMilestoneMap_DescNullable(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.MilestoneMap{})
	require.NoError(t, err)

	m := model.MilestoneMap{
		TeamKey:   1,
		MapName:   "No Desc Map",
		MapStatus: "planning",
	}
	require.NoError(t, db.Create(&m).Error)

	var fetched model.MilestoneMap
	db.First(&fetched, "map_name = ?", "No Desc Map")
	assert.Equal(t, "", fetched.MapDesc, "map_desc should default to empty")
}

func TestMilestone_TableName(t *testing.T) {
	m := model.Milestone{}
	assert.Equal(t, "pmw_milestones", m.TableName())
}

func TestMilestone_Defaults(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Milestone{})
	require.NoError(t, err)

	m := model.Milestone{
		TeamKey:         12345,
		MilestoneMapKey: 67890,
		MilestoneName:   "Alpha Release",
		MilestoneStatus: "not_started",
	}
	require.NoError(t, db.Create(&m).Error)

	var fetched model.Milestone
	db.First(&fetched, "milestone_name = ?", "Alpha Release")
	assert.Equal(t, "Alpha Release", fetched.MilestoneName)
	assert.Equal(t, "not_started", fetched.MilestoneStatus)
	assert.Equal(t, int64(12345), fetched.TeamKey)
	assert.Equal(t, int64(67890), fetched.MilestoneMapKey)
	assert.Nil(t, fetched.ExpectedEndDate, "expected_end_date should default to nil")
}

func TestMilestone_ExpectedEndDateSet(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Milestone{})
	require.NoError(t, err)

	future := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	m := model.Milestone{
		TeamKey:         1,
		MilestoneMapKey: 10,
		MilestoneName:   "With Date",
		MilestoneStatus: "not_started",
		ExpectedEndDate: &future,
	}
	require.NoError(t, db.Create(&m).Error)

	var fetched model.Milestone
	db.First(&fetched, "milestone_name = ?", "With Date")
	assert.NotNil(t, fetched.ExpectedEndDate)
	assert.Equal(t, future.Year(), fetched.ExpectedEndDate.Year())
}

func TestMainItem_MilestoneKey(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{})
	require.NoError(t, err)

	u := model.User{Username: "mkuser", DisplayName: "MK", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "MKT", PmKey: int64(u.ID), Code: "MKT1"}
	require.NoError(t, db.Create(&team).Error)

	mk := int64(99999)
	m := model.MainItem{
		TeamKey:      int64(team.ID),
		Code:         "MKT1-00001",
		Title:        "Item with milestone",
		Priority:     "P1",
		ProposerKey:  int64(u.ID),
		MilestoneKey: &mk,
	}
	require.NoError(t, db.Create(&m).Error)

	var fetched model.MainItem
	db.First(&fetched, "item_code = ?", "MKT1-00001")
	assert.NotNil(t, fetched.MilestoneKey)
	assert.Equal(t, int64(99999), *fetched.MilestoneKey)
}

func TestMainItem_MilestoneKeyNull(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{})
	require.NoError(t, err)

	u := model.User{Username: "mknull", DisplayName: "MN", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "MKN", PmKey: int64(u.ID), Code: "MKN1"}
	require.NoError(t, db.Create(&team).Error)

	m := model.MainItem{
		TeamKey:     int64(team.ID),
		Code:        "MKN1-00001",
		Title:       "Item no milestone",
		Priority:    "P2",
		ProposerKey: int64(u.ID),
	}
	require.NoError(t, db.Create(&m).Error)

	var fetched model.MainItem
	db.First(&fetched, "item_code = ?", "MKN1-00001")
	assert.Nil(t, fetched.MilestoneKey, "milestone_key should default to nil")
}
