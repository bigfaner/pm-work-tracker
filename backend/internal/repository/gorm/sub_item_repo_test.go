package gorm_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	pkgerrors "pm-work-tracker/backend/internal/pkg/errors"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

func setupSubItemTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}))
	return db
}

func seedSubItemData(t *testing.T, db *gormlib.DB) (*model.User, *model.Team, *model.MainItem) {
	t.Helper()
	u := model.User{Username: "si_pm", DisplayName: "SI PM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "SI Team", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamID: team.ID, Code: "MI-SI01", Title: "Main", Priority: "P1", ProposerID: u.ID, Status: "待开始"}
	require.NoError(t, db.Create(&mi).Error)
	return &u, &team, &mi
}

func createSubItem(t *testing.T, db *gormlib.DB, teamID, mainItemID uint, title, priority, status string) *model.SubItem {
	t.Helper()
	item := model.SubItem{
		TeamID:     teamID,
		MainItemID: mainItemID,
		Title:      title,
		Priority:   priority,
		Status:     status,
	}
	require.NoError(t, db.Create(&item).Error)
	return &item
}

// --- Create ---

func TestSubItemRepo_Create(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	_, team, mi := seedSubItemData(t, db)
	item := &model.SubItem{
		TeamID:     team.ID,
		MainItemID: mi.ID,
		Title:      "Sub 1",
		Priority:   "P1",
		Status:     "待开始",
	}
	require.NoError(t, repo.Create(ctx, item))
	assert.NotZero(t, item.ID)
}

// --- FindByID ---

func TestSubItemRepo_FindByID(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	_, team, mi := seedSubItemData(t, db)
	item := createSubItem(t, db, team.ID, mi.ID, "Find Me", "P1", "待开始")

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, item.ID)
		require.NoError(t, err)
		assert.Equal(t, "Find Me", found.Title)
		assert.Equal(t, team.ID, found.TeamID)
		assert.Equal(t, mi.ID, found.MainItemID)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

// --- Update ---

func TestSubItemRepo_Update(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	_, team, mi := seedSubItemData(t, db)
	item := createSubItem(t, db, team.ID, mi.ID, "Update Me", "P1", "待开始")

	fields := map[string]interface{}{
		"title":    "Updated Sub",
		"status":   "进行中",
		"priority": "P2",
	}
	require.NoError(t, repo.Update(ctx, item, fields))

	found, err := repo.FindByID(ctx, item.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Sub", found.Title)
	assert.Equal(t, "进行中", found.Status)
	assert.Equal(t, "P2", found.Priority)
}

func TestSubItemRepo_Update_NotFound(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	_, team, _ := seedSubItemData(t, db)
	fakeItem := &model.SubItem{BaseModel: model.BaseModel{ID: 9999}, TeamID: team.ID}
	fields := map[string]interface{}{"title": "Nope"}
	err := repo.Update(ctx, fakeItem, fields)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- List ---

func TestSubItemRepo_List(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	_, team, mi := seedSubItemData(t, db)

	createSubItem(t, db, team.ID, mi.ID, "Sub A", "P1", "待开始")
	createSubItem(t, db, team.ID, mi.ID, "Sub B", "P2", "进行中")
	createSubItem(t, db, team.ID, mi.ID, "Sub C", "P1", "已完成")

	// Create another main item with sub in same team
	mi2 := model.MainItem{TeamID: team.ID, Code: "MI-SI02", Title: "Main2", Priority: "P1", ProposerID: team.PmID, Status: "待开始"}
	require.NoError(t, db.Create(&mi2).Error)
	createSubItem(t, db, team.ID, mi2.ID, "Sub Other", "P1", "待开始")

	t.Run("all_for_main_item", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("filter_by_status", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{Status: "进行中"}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Sub B", result.Items[0].Title)
	})

	t.Run("filter_by_priority", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{Priority: "P1"}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
	})

	t.Run("filter_by_status_and_priority", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{Status: "待开始", Priority: "P1"}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Sub A", result.Items[0].Title)
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{}, dto.Pagination{Page: 1, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)
	})

	t.Run("without_main_item_filter", func(t *testing.T) {
		// mainItemID=0 means no main_item_id filter
		result, err := repo.List(ctx, team.ID, 0, dto.SubItemFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(4), result.Total, "should return all sub items for the team")
	})

	t.Run("team_isolation", func(t *testing.T) {
		u2 := model.User{Username: "other_pm", DisplayName: "OP", PasswordHash: "h"}
		require.NoError(t, db.Create(&u2).Error)
		team2 := model.Team{Name: "Other SI Team", PmID: u2.ID}
		require.NoError(t, db.Create(&team2).Error)
		mi3 := model.MainItem{TeamID: team2.ID, Code: "MI-SI03", Title: "M3", Priority: "P1", ProposerID: u2.ID, Status: "待开始"}
		require.NoError(t, db.Create(&mi3).Error)
		createSubItem(t, db, team2.ID, mi3.ID, "Other Sub", "P1", "待开始")

		result, err := repo.List(ctx, team2.ID, mi3.ID, dto.SubItemFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Other Sub", result.Items[0].Title)
	})
}

func TestSubItemRepo_List_FilterByAssignee(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	u, team, mi := seedSubItemData(t, db)

	item := createSubItem(t, db, team.ID, mi.ID, "Assigned Sub", "P1", "待开始")
	require.NoError(t, db.Model(item).Update("assignee_id", u.ID).Error)

	createSubItem(t, db, team.ID, mi.ID, "Unassigned Sub", "P2", "待开始")

	assigneeID := u.ID
	result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{AssigneeID: &assigneeID}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "Assigned Sub", result.Items[0].Title)
}

func TestSubItemRepo_List_FilterByKeyItem(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	_, team, mi := seedSubItemData(t, db)

	item := createSubItem(t, db, team.ID, mi.ID, "Key Sub", "P1", "待开始")
	require.NoError(t, db.Model(item).Update("is_key_item", true).Error)

	createSubItem(t, db, team.ID, mi.ID, "Normal Sub", "P2", "待开始")

	isKey := true
	result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{IsKeyItem: &isKey}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "Key Sub", result.Items[0].Title)
}

// --- ListByMainItem ---

func TestSubItemRepo_ListByMainItem(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	_, team, mi := seedSubItemData(t, db)

	createSubItem(t, db, team.ID, mi.ID, "Sub 1", "P1", "待开始")
	createSubItem(t, db, team.ID, mi.ID, "Sub 2", "P2", "进行中")
	createSubItem(t, db, team.ID, mi.ID, "Sub 3", "P1", "已完成")

	items, err := repo.ListByMainItem(ctx, mi.ID)
	require.NoError(t, err)
	assert.Len(t, items, 3)

	titles := map[string]bool{}
	for _, item := range items {
		titles[item.Title] = true
	}
	assert.True(t, titles["Sub 1"])
	assert.True(t, titles["Sub 2"])
	assert.True(t, titles["Sub 3"])
}

func TestSubItemRepo_ListByMainItem_Empty(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	items, err := repo.ListByMainItem(ctx, 9999)
	require.NoError(t, err)
	assert.Empty(t, items)
}
