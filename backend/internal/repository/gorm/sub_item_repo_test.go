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
	team := model.Team{TeamName: "SI Team", PmKey: int64(u.ID), Code: "SITE"}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamKey: int64(team.ID), Code: "MI-SI01", ItemStatus: "pending", Priority: "P1", Title: "SI01"}
	require.NoError(t, db.Create(&mi).Error)
	return &u, &team, &mi
}

func createSubItem(t *testing.T, db *gormlib.DB, teamID, mainItemID uint, title, priority, status string) *model.SubItem {
	t.Helper()
	item := model.SubItem{
		TeamKey: int64(teamID),
		MainItemKey: int64(mainItemID),
		Title:      title,
		Priority:   priority,
		ItemStatus: status,
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
		TeamKey: int64(team.ID),
		MainItemKey: int64(mi.ID),
		Title:      "Sub 1",
		Priority:   "P1",
		ItemStatus: "pending",
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
	item := createSubItem(t, db, team.ID, mi.ID, "Find Me", "P1", "pending")

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, item.ID)
		require.NoError(t, err)
		assert.Equal(t, "Find Me", found.Title)
		assert.Equal(t, int64(team.ID), found.TeamKey)
		assert.Equal(t, mi.ID, uint(found.MainItemKey))
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
	item := createSubItem(t, db, team.ID, mi.ID, "Update Me", "P1", "pending")

	fields := map[string]interface{}{
		"title":       "Updated Sub",
		"item_status": "progressing",
		"priority":    "P2",
	}
	require.NoError(t, repo.Update(ctx, item, fields))

	found, err := repo.FindByID(ctx, item.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Sub", found.Title)
	assert.Equal(t, "progressing", found.ItemStatus)
	assert.Equal(t, "P2", found.Priority)
}

func TestSubItemRepo_Update_NotFound(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	_, team, _ := seedSubItemData(t, db)
	fakeItem := &model.SubItem{BaseModel: model.BaseModel{ID: 9999}, TeamKey: int64(team.ID)}
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

	createSubItem(t, db, team.ID, mi.ID, "Sub A", "P1", "pending")
	createSubItem(t, db, team.ID, mi.ID, "Sub B", "P2", "progressing")
	createSubItem(t, db, team.ID, mi.ID, "Sub C", "P1", "completed")

	// Create another main item with sub in same team
	mi2 := model.MainItem{ItemStatus: "pending", Code: "MI2"}
	require.NoError(t, db.Create(&mi2).Error)
	createSubItem(t, db, team.ID, mi2.ID, "Sub Other", "P1", "pending")

	t.Run("all_for_main_item", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("filter_by_status", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{Status: "progressing"}, dto.Pagination{Page: 1, PageSize: 10})
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
		result, err := repo.List(ctx, team.ID, mi.ID, dto.SubItemFilter{Status: "pending", Priority: "P1"}, dto.Pagination{Page: 1, PageSize: 10})
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
		team2 := model.Team{TeamName: "Other SI Team", PmKey: int64(u2.ID), Code: "OSIT"}
		require.NoError(t, db.Create(&team2).Error)
		mi3 := model.MainItem{ItemStatus: "pending", Code: "MI3"}
		require.NoError(t, db.Create(&mi3).Error)
		createSubItem(t, db, team2.ID, mi3.ID, "Other Sub", "P1", "pending")

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

	item := createSubItem(t, db, team.ID, mi.ID, "Assigned Sub", "P1", "pending")
	require.NoError(t, db.Model(item).Update("assignee_key", u.ID).Error)

	createSubItem(t, db, team.ID, mi.ID, "Unassigned Sub", "P2", "pending")

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

	item := createSubItem(t, db, team.ID, mi.ID, "Key Sub", "P1", "pending")
	require.NoError(t, db.Model(item).Update("is_key_item", true).Error)

	createSubItem(t, db, team.ID, mi.ID, "Normal Sub", "P2", "pending")

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

	createSubItem(t, db, team.ID, mi.ID, "Sub 1", "P1", "pending")
	createSubItem(t, db, team.ID, mi.ID, "Sub 2", "P2", "progressing")
	createSubItem(t, db, team.ID, mi.ID, "Sub 3", "P1", "completed")

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

// --- NextSubCode ---

func TestNextSubCode(t *testing.T) {
	db := setupSubItemTestDB(t)
	repo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	_, team, mi := seedSubItemData(t, db) // mi.Code = "MI-SI01"

	t.Run("first_sub_code", func(t *testing.T) {
		code, err := repo.NextSubCode(ctx, mi.ID)
		require.NoError(t, err)
		assert.Equal(t, "MI-SI01-01", code)
	})

	t.Run("sequential", func(t *testing.T) {
		sub := model.SubItem{TeamKey: int64(team.ID), MainItemKey: int64(mi.ID), Title: "S1", Priority: "P1", ItemStatus: "pending", Code: "MI-SI01-01"}
		require.NoError(t, db.Create(&sub).Error)

		code, err := repo.NextSubCode(ctx, mi.ID)
		require.NoError(t, err)
		assert.Equal(t, "MI-SI01-02", code)
	})

	t.Run("skips_gaps", func(t *testing.T) {
		sub := model.SubItem{TeamKey: int64(team.ID), MainItemKey: int64(mi.ID), Title: "S5", Priority: "P1", ItemStatus: "pending", Code: "MI-SI01-05"}
		require.NoError(t, db.Create(&sub).Error)

		code, err := repo.NextSubCode(ctx, mi.ID)
		require.NoError(t, err)
		assert.Equal(t, "MI-SI01-06", code)
	})

	t.Run("main_item_isolation", func(t *testing.T) {
		mi2 := model.MainItem{TeamKey: int64(team.ID), Code: "MI-SI02", ItemStatus: "pending", Priority: "P1", Title: "SI02"}
		require.NoError(t, db.Create(&mi2).Error)

		code, err := repo.NextSubCode(ctx, mi2.ID)
		require.NoError(t, err)
		assert.Equal(t, "MI-SI02-01", code)
	})
}
