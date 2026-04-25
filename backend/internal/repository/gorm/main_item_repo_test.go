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

func setupMainItemTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}))
	return db
}

func seedMainItemTeam(t *testing.T, db *gormlib.DB) (*model.User, *model.Team) {
	t.Helper()
	u := model.User{Username: "mi_pm", DisplayName: "MI PM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "MI Team", PmKey: int64(u.ID), Code: "FEAT"}
	require.NoError(t, db.Create(&team).Error)
	return &u, &team
}

func createMainItem(t *testing.T, db *gormlib.DB, teamID, proposerID uint, code, title, priority, status string) *model.MainItem {
	t.Helper()
	item := model.MainItem{
		TeamID: teamID,
		Code:       code,
		Title:      title,
		Priority:   priority,
		ProposerKey: int64(proposerID),
		ItemStatus: status,
	}
	require.NoError(t, db.Create(&item).Error)
	return &item
}

// --- Create ---

func TestMainItemRepo_Create(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db)
	item := &model.MainItem{
		TeamID: team.ID,
		Code:       "FEAT-00001",
		Title:      "Test Item",
		Priority:   "P1",
		ProposerKey: int64(u.ID),
		ItemStatus: "pending",
	}
	require.NoError(t, repo.Create(ctx, item))
	assert.NotZero(t, item.ID)
}

// --- FindByID ---

func TestMainItemRepo_FindByID(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db)
	item := createMainItem(t, db, team.ID, u.ID, "FEAT-00001", "Find Me", "P1", "pending")

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, item.ID)
		require.NoError(t, err)
		assert.Equal(t, "Find Me", found.Title)
		assert.Equal(t, team.ID, found.TeamID)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

// --- Update ---

func TestMainItemRepo_Update(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db)
	item := createMainItem(t, db, team.ID, u.ID, "FEAT-00001", "Update Me", "P1", "pending")

	fields := map[string]interface{}{
		"title":       "Updated Title",
		"item_status": "progressing",
		"priority":    "P2",
	}
	require.NoError(t, repo.Update(ctx, item, fields))

	found, err := repo.FindByID(ctx, item.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", found.Title)
	assert.Equal(t, "progressing", found.ItemStatus)
	assert.Equal(t, "P2", found.Priority)
}

func TestMainItemRepo_Update_NotFound(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	_, team := seedMainItemTeam(t, db)
	fakeItem := &model.MainItem{BaseModel: model.BaseModel{ID: 9999}, TeamID: team.ID}
	fields := map[string]interface{}{"title": "Nope"}
	err := repo.Update(ctx, fakeItem, fields)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- NextCode ---

func TestMainItemRepo_NextCode(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db) // team.Code = "FEAT"

	t.Run("first_code", func(t *testing.T) {
		code, err := repo.NextCode(ctx, team.ID)
		require.NoError(t, err)
		assert.Equal(t, "FEAT-00001", code)
	})

	t.Run("sequential", func(t *testing.T) {
		createMainItem(t, db, team.ID, u.ID, "FEAT-00001", "First", "P1", "pending")

		code, err := repo.NextCode(ctx, team.ID)
		require.NoError(t, err)
		assert.Equal(t, "FEAT-00002", code)
	})

	t.Run("skips_gaps", func(t *testing.T) {
		createMainItem(t, db, team.ID, u.ID, "FEAT-00005", "Fifth", "P1", "pending")

		code, err := repo.NextCode(ctx, team.ID)
		require.NoError(t, err)
		assert.Equal(t, "FEAT-00006", code)
	})

	t.Run("team_isolation", func(t *testing.T) {
		// Different team should start fresh with its own code prefix
		u2 := model.User{Username: "pm_other", DisplayName: "Other PM", PasswordHash: "h"}
		require.NoError(t, db.Create(&u2).Error)
		team2 := model.Team{TeamName: "Other Team", PmKey: int64(u2.ID), Code: "OTHR"}
		require.NoError(t, db.Create(&team2).Error)

		code, err := repo.NextCode(ctx, team2.ID)
		require.NoError(t, err)
		assert.Equal(t, "OTHR-00001", code, "new team should get OTHR-00001")
	})
}

// --- List with filters and pagination ---

func TestMainItemRepo_List(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db)

	// Create several items
	createMainItem(t, db, team.ID, u.ID, "FEAT-00001", "Item A", "P1", "pending")
	createMainItem(t, db, team.ID, u.ID, "FEAT-00002", "Item B", "P2", "progressing")
	createMainItem(t, db, team.ID, u.ID, "FEAT-00003", "Item C", "P1", "completed")
	// Create one in another team — should not appear
	u2 := model.User{Username: "pm_other2", DisplayName: "P2", PasswordHash: "h"}
	require.NoError(t, db.Create(&u2).Error)
	team2 := model.Team{TeamName: "Team2", PmKey: int64(u2.ID)}
	require.NoError(t, db.Create(&team2).Error)
	createMainItem(t, db, team2.ID, u2.ID, "FEAT-00004", "Other Team Item", "P1", "pending")

	t.Run("all_items_for_team", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.MainItemFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 3)
	})

	t.Run("filter_by_status", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.MainItemFilter{Status: "progressing"}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Item B", result.Items[0].Title)
	})

	t.Run("filter_by_priority", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.MainItemFilter{Priority: "P1"}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
	})

	t.Run("filter_by_status_and_priority", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.MainItemFilter{Status: "pending", Priority: "P1"}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Item A", result.Items[0].Title)
	})

	t.Run("pagination_page1", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.MainItemFilter{}, dto.Pagination{Page: 1, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 2, result.Size)
	})

	t.Run("pagination_page2", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.MainItemFilter{}, dto.Pagination{Page: 2, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 1)
	})

	t.Run("default_pagination", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.MainItemFilter{}, dto.Pagination{})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.Size)
	})
}

func TestMainItemRepo_List_ArchiveFilter(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db)

	// Active item
	createMainItem(t, db, team.ID, u.ID, "FEAT-00001", "Active", "P1", "pending")

	// Archived item
	archived := createMainItem(t, db, team.ID, u.ID, "FEAT-00002", "Archived", "P1", "completed")
	now := time.Now()
	require.NoError(t, db.Model(archived).Update("archived_at", &now).Error)

	t.Run("excludes_archived_by_default", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.MainItemFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Active", result.Items[0].Title)
	})

	t.Run("includes_archived_when_filter_set", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.MainItemFilter{Archived: true}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
	})
}

func TestMainItemRepo_List_FilterByKeyItem(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db)

	item := createMainItem(t, db, team.ID, u.ID, "FEAT-00001", "Key Item", "P1", "pending")
	require.NoError(t, db.Model(item).Update("is_key_item", true).Error)

	createMainItem(t, db, team.ID, u.ID, "FEAT-00002", "Normal Item", "P2", "pending")

	isKey := true
	result, err := repo.List(ctx, team.ID, dto.MainItemFilter{IsKeyItem: &isKey}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "Key Item", result.Items[0].Title)
}

func TestMainItemRepo_List_FilterByAssignee(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db)

	item := createMainItem(t, db, team.ID, u.ID, "FEAT-00001", "Assigned", "P1", "pending")
	require.NoError(t, db.Model(item).Update("assignee_key", u.ID).Error)

	createMainItem(t, db, team.ID, u.ID, "FEAT-00002", "Unassigned", "P2", "pending")

	assigneeID := u.ID
	result, err := repo.List(ctx, team.ID, dto.MainItemFilter{AssigneeID: &assigneeID}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "Assigned", result.Items[0].Title)
}

// --- FindByIDs ---

func TestMainItemRepo_FindByIDs(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db)
	item1 := createMainItem(t, db, team.ID, u.ID, "FEAT-00001", "Item 1", "P1", "pending")
	item2 := createMainItem(t, db, team.ID, u.ID, "FEAT-00002", "Item 2", "P2", "progressing")

	t.Run("returns_map_for_existing_ids", func(t *testing.T) {
		result, err := repo.FindByIDs(ctx, []uint{item1.ID, item2.ID})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Item 1", result[item1.ID].Title)
		assert.Equal(t, "Item 2", result[item2.ID].Title)
	})

	t.Run("empty_input_returns_empty_map", func(t *testing.T) {
		result, err := repo.FindByIDs(ctx, []uint{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("partial_results_return_found_keys_only", func(t *testing.T) {
		result, err := repo.FindByIDs(ctx, []uint{item1.ID, 9999})
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Item 1", result[item1.ID].Title)
		_, hasMissing := result[9999]
		assert.False(t, hasMissing)
	})
}

// --- ListByTeamAndStatus ---

func TestMainItemRepo_ListByTeamAndStatus(t *testing.T) {
	db := setupMainItemTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u, team := seedMainItemTeam(t, db)
	createMainItem(t, db, team.ID, u.ID, "FEAT-00001", "Pending 1", "P1", "pending")
	createMainItem(t, db, team.ID, u.ID, "FEAT-00002", "Progressing 1", "P2", "progressing")
	createMainItem(t, db, team.ID, u.ID, "FEAT-00003", "Pending 2", "P1", "pending")

	t.Run("returns_items_matching_status", func(t *testing.T) {
		items, err := repo.ListByTeamAndStatus(ctx, team.ID, "pending")
		require.NoError(t, err)
		assert.Len(t, items, 2)
	})

	t.Run("no_match_returns_empty", func(t *testing.T) {
		items, err := repo.ListByTeamAndStatus(ctx, team.ID, "completed")
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("team_isolation", func(t *testing.T) {
		u2 := model.User{Username: "other_pm", DisplayName: "Other PM", PasswordHash: "h"}
		require.NoError(t, db.Create(&u2).Error)
		team2 := model.Team{TeamName: "Team2", PmKey: int64(u2.ID), Code: "T2"}
		require.NoError(t, db.Create(&team2).Error)
		createMainItem(t, db, team2.ID, u2.ID, "T2-00001", "Other Pending", "P1", "pending")

		items, err := repo.ListByTeamAndStatus(ctx, team.ID, "pending")
		require.NoError(t, err)
		assert.Len(t, items, 2, "should not include items from other teams")
	})
}
