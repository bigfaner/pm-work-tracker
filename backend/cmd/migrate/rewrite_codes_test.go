package main

import (
	"context"
	"testing"

	"pm-work-tracker/backend/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Team{}, &model.MainItem{}, &model.SubItem{}))
	return db
}

func seedTeam(t *testing.T, db *gorm.DB, code string) model.Team {
	t.Helper()
	team := model.Team{TeamName: "Team " + code, Code: code, PmKey: 1}
	require.NoError(t, db.Create(&team).Error)
	return team
}

func seedMainItem(t *testing.T, db *gorm.DB, teamID uint, oldCode string) model.MainItem {
	t.Helper()
	item := model.MainItem{TeamKey: int64(teamID), Code: oldCode, Title: "item", Priority: "P1", ProposerKey: 1}
	require.NoError(t, db.Create(&item).Error)
	return item
}

func seedSubItem(t *testing.T, db *gorm.DB, teamID, mainItemID uint) model.SubItem {
	t.Helper()
	sub := model.SubItem{TeamKey: int64(teamID), MainItemKey: int64(mainItemID), Title: "sub", Priority: "P1"}
	require.NoError(t, db.Create(&sub).Error)
	return sub
}

// TestMigrate_RewritesMainItemCodes verifies main_items.code is rewritten to {TEAM_CODE}-{seq:05d}.
func TestMigrate_RewritesMainItemCodes(t *testing.T) {
	db := setupTestDB(t)
	team := seedTeam(t, db, "ALPHA")
	item1 := seedMainItem(t, db, team.ID, "MI-0001")
	item2 := seedMainItem(t, db, team.ID, "MI-0002")

	require.NoError(t, migrate(context.Background(), db, false))

	var got1, got2 model.MainItem
	require.NoError(t, db.First(&got1, item1.ID).Error)
	require.NoError(t, db.First(&got2, item2.ID).Error)

	assert.Equal(t, "ALPHA-00001", got1.Code)
	assert.Equal(t, "ALPHA-00002", got2.Code)
}

// TestMigrate_GroupsByTeam verifies items from different teams get their own sequences.
func TestMigrate_GroupsByTeam(t *testing.T) {
	db := setupTestDB(t)
	teamA := seedTeam(t, db, "AAA")
	teamB := seedTeam(t, db, "BBB")
	itemA := seedMainItem(t, db, teamA.ID, "MI-0001")
	itemB := seedMainItem(t, db, teamB.ID, "MI-0002")

	require.NoError(t, migrate(context.Background(), db, false))

	var gotA, gotB model.MainItem
	require.NoError(t, db.First(&gotA, itemA.ID).Error)
	require.NoError(t, db.First(&gotB, itemB.ID).Error)

	assert.Equal(t, "AAA-00001", gotA.Code)
	assert.Equal(t, "BBB-00001", gotB.Code)
}

// TestMigrate_GeneratesSubItemCodes verifies sub_items.code is generated as {main_code}-{seq:02d}.
func TestMigrate_GeneratesSubItemCodes(t *testing.T) {
	db := setupTestDB(t)
	team := seedTeam(t, db, "BETA")
	item := seedMainItem(t, db, team.ID, "MI-0001")
	sub1 := seedSubItem(t, db, team.ID, item.ID)
	sub2 := seedSubItem(t, db, team.ID, item.ID)

	require.NoError(t, migrate(context.Background(), db, false))

	var gotSub1, gotSub2 model.SubItem
	require.NoError(t, db.First(&gotSub1, sub1.ID).Error)
	require.NoError(t, db.First(&gotSub2, sub2.ID).Error)

	assert.Equal(t, "BETA-00001-01", gotSub1.Code)
	assert.Equal(t, "BETA-00001-02", gotSub2.Code)
}

// TestMigrate_PreConditionFails verifies migration aborts when a team has no code.
func TestMigrate_PreConditionFails(t *testing.T) {
	db := setupTestDB(t)
	// Insert a team with empty code directly via raw SQL to bypass model constraints.
	require.NoError(t, db.Exec("INSERT INTO pmw_teams (team_name, code, pm_key, biz_key, create_time, db_update_time, deleted_flag, deleted_time) VALUES ('NoCode', '', 1, 0, datetime('now'), datetime('now'), 0, '1970-01-01 08:00:00')").Error)

	err := migrate(context.Background(), db, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pre-condition failed")
}

// TestMigrate_DryRunDoesNotCommit verifies dry-run leaves data unchanged.
func TestMigrate_DryRunDoesNotCommit(t *testing.T) {
	db := setupTestDB(t)
	team := seedTeam(t, db, "GAMMA")
	item := seedMainItem(t, db, team.ID, "MI-0001")

	require.NoError(t, migrate(context.Background(), db, true))

	var got model.MainItem
	require.NoError(t, db.First(&got, item.ID).Error)
	assert.Equal(t, "MI-0001", got.Code, "dry-run must not commit changes")
}

// TestMigrate_ValidationPassesAfterMigration verifies no MI- codes remain and no empty sub codes.
func TestMigrate_ValidationPassesAfterMigration(t *testing.T) {
	db := setupTestDB(t)
	team := seedTeam(t, db, "DELTA")
	item := seedMainItem(t, db, team.ID, "MI-0001")
	seedSubItem(t, db, team.ID, item.ID)

	require.NoError(t, migrate(context.Background(), db, false))

	var oldFmt int64
	require.NoError(t, db.Model(&model.MainItem{}).Where("code LIKE 'MI-%'").Count(&oldFmt).Error)
	assert.Equal(t, int64(0), oldFmt)

	var emptySub int64
	require.NoError(t, db.Model(&model.SubItem{}).Where("code IS NULL OR code = ''").Count(&emptySub).Error)
	assert.Equal(t, int64(0), emptySub)
}

// TestMigrate_EmptyDB verifies migration succeeds with no data.
func TestMigrate_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, migrate(context.Background(), db, false))
}

// TestMigrate_ValidationFailsOldFormatRemains verifies the post-migration check catches
// a main_item with MI- code that was not processed (orphaned team_id not in teams table).
func TestMigrate_ValidationFailsOldFormatRemains(t *testing.T) {
	db := setupTestDB(t)
	// Seed a real team so pre-condition passes.
	seedTeam(t, db, "ECHO")
	// Insert an orphaned main_item with a non-existent team_key — migration loop won't touch it.
	require.NoError(t, db.Exec(
		"INSERT INTO pmw_main_items (team_key, code, title, priority, proposer_key, item_desc, item_status, completion, is_key_item, biz_key, create_time, db_update_time, deleted_flag, deleted_time) VALUES (9999, 'MI-0001', 'orphan', 'P1', 1, '', 'pending', 0, 0, 0, datetime('now'), datetime('now'), 0, '1970-01-01 08:00:00')",
	).Error)

	err := migrate(context.Background(), db, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "main_item(s) still have MI- prefix")
}

// TestMigrate_ValidationFailsEmptySubCode verifies the post-migration check catches
// a sub_item with empty code that was not processed (orphaned main_item_id).
func TestMigrate_ValidationFailsEmptySubCode(t *testing.T) {
	db := setupTestDB(t)
	team := seedTeam(t, db, "FOXT")
	item := seedMainItem(t, db, team.ID, "MI-0001")
	// Insert an orphaned sub_item with empty code — migration loop won't touch it.
	require.NoError(t, db.Exec(
		"INSERT INTO pmw_sub_items (team_key, main_item_key, code, title, priority, item_status, completion, is_key_item, weight, biz_key, create_time, db_update_time, deleted_flag, deleted_time) VALUES (?, 9999, '', 'orphan', 'P1', 'pending', 0, 0, 1, 0, datetime('now'), datetime('now'), 0, '1970-01-01 08:00:00')",
		team.ID,
	).Error)
	// Also seed a normal item so the main_item loop runs.
	seedSubItem(t, db, team.ID, item.ID)

	err := migrate(context.Background(), db, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sub_item(s) have empty code")
}
