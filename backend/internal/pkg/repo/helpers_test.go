package repo

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlib "gorm.io/gorm"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/model"
)

// setupTestDB creates an in-memory SQLite database and auto-migrates the schema.
func setupTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(
		&model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.User{},
		&model.ProgressRecord{}, &model.StatusHistory{}, &model.Role{},
	)
	require.NoError(t, err)
	return db
}

// --- FindByID tests ---

func TestFindByID_Found(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create a team first (MainItem has team_id FK)
	db.Create(&model.Team{Code: "TST", TeamName: "Test"})

	item := &model.MainItem{ItemStatus: "open", Code: "TST-00001", Title: "test item"}
	require.NoError(t, db.Create(item).Error)

	found, err := FindByID[model.MainItem](db, ctx, item.ID)
	require.NoError(t, err)
	assert.Equal(t, item.ID, found.ID)
	assert.Equal(t, "test item", found.Title)
}

func TestFindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	_, err := FindByID[model.MainItem](db, ctx, 9999)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestFindByID_DBError(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Close the DB to force an error
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	_, err = FindByID[model.MainItem](db, ctx, 1)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, apperrors.ErrNotFound)
}

// --- FindByIDs tests ---

func TestFindByIDs_EmptyIDs(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	result, err := FindByIDs[model.MainItem](db, ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, result)

	result, err = FindByIDs[model.MainItem](db, ctx, []uint{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFindByIDs_Found(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})

	item1 := &model.MainItem{ItemStatus: "open", Code: "TST-00001"}
	item2 := &model.MainItem{ItemStatus: "open", Code: "TST-00002"}
	require.NoError(t, db.Create(item1).Error)
	require.NoError(t, db.Create(item2).Error)

	result, err := FindByIDs[model.MainItem](db, ctx, []uint{item1.ID, item2.ID})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, item1.ID)
	assert.Contains(t, result, item2.ID)
}

func TestFindByIDs_PartialResults(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})

	item := &model.MainItem{ItemStatus: "open", Code: "TST-00001", Title: "test item"}
	require.NoError(t, db.Create(item).Error)

	result, err := FindByIDs[model.MainItem](db, ctx, []uint{item.ID, 9999})
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, item.ID)
	assert.NotContains(t, result, uint(9999))
}

func TestFindByIDs_ZeroResults(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	result, err := FindByIDs[model.MainItem](db, ctx, []uint{9999, 8888})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFindByIDs_UserType(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	user1 := &model.User{Username: "user1", PasswordHash: "h", DisplayName: "One"}
	user2 := &model.User{Username: "user2", PasswordHash: "h", DisplayName: "Two"}
	require.NoError(t, db.Create(user1).Error)
	require.NoError(t, db.Create(user2).Error)

	result, err := FindByIDs[model.User](db, ctx, []uint{user1.ID, user2.ID})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "One", result[user1.ID].DisplayName)
	assert.Equal(t, "Two", result[user2.ID].DisplayName)
}

func TestFindByIDs_SubItemType(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})
	db.Create(&model.MainItem{ItemStatus: "open", Code: "TST-00001"})

	sub1 := &model.SubItem{TeamKey: 1, MainItemKey: int64(1), Title: "sub1", ItemStatus: "open"}
	sub2 := &model.SubItem{TeamKey: 1, MainItemKey: int64(1), Title: "sub2", ItemStatus: "pending"}
	require.NoError(t, db.Create(sub1).Error)
	require.NoError(t, db.Create(sub2).Error)

	result, err := FindByIDs[model.SubItem](db, ctx, []uint{sub1.ID, sub2.ID})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "sub1", result[sub1.ID].Title)
	assert.Equal(t, "sub2", result[sub2.ID].Title)
}

func TestFindByIDs_DBError(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	_, err = FindByIDs[model.MainItem](db, ctx, []uint{1, 2})
	assert.Error(t, err)
}

// --- isSoftDeletable tests ---

func TestIsSoftDeletable_SoftDeletableTypes(t *testing.T) {
	// Types that embed BaseModel should return true
	assert.True(t, isSoftDeletable[model.User]())
	assert.True(t, isSoftDeletable[model.Team]())
	assert.True(t, isSoftDeletable[model.MainItem]())
	assert.True(t, isSoftDeletable[model.SubItem]())
	assert.True(t, isSoftDeletable[model.ItemPool]())
	assert.True(t, isSoftDeletable[model.Role]())
}

func TestIsSoftDeletable_NonSoftDeletableTypes(t *testing.T) {
	// Types without BaseModel should return false
	assert.False(t, isSoftDeletable[model.ProgressRecord]())
	assert.False(t, isSoftDeletable[model.StatusHistory]())
}

// --- FindByID soft-delete tests ---

func TestFindByID_SoftDeletedUser_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	user := &model.User{Username: "deleted_user", PasswordHash: "h", DisplayName: "Deleted"}
	require.NoError(t, db.Create(user).Error)

	// Soft-delete the user
	db.Model(user).Update("deleted_flag", 1)

	_, err := FindByID[model.User](db, ctx, user.ID)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestFindByID_SoftDeletedMainItem_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})

	item := &model.MainItem{ItemStatus: "open", Code: "TST-00001", Title: "deleted item"}
	require.NoError(t, db.Create(item).Error)

	// Soft-delete the item
	db.Model(item).Update("deleted_flag", 1)

	_, err := FindByID[model.MainItem](db, ctx, item.ID)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestFindByID_SoftDeletedSubItem_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})
	db.Create(&model.MainItem{ItemStatus: "open", Code: "TST-00001"})

	sub := &model.SubItem{TeamKey: 1, MainItemKey: 1, Title: "sub", ItemStatus: "open"}
	require.NoError(t, db.Create(sub).Error)

	// Soft-delete the sub item
	db.Model(sub).Update("deleted_flag", 1)

	_, err := FindByID[model.SubItem](db, ctx, sub.ID)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// --- FindByIDs soft-delete tests ---

func TestFindByIDs_SoftDeletedUser_ExcludedFromResults(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	user1 := &model.User{Username: "active", PasswordHash: "h", DisplayName: "Active"}
	user2 := &model.User{Username: "deleted", PasswordHash: "h", DisplayName: "Deleted"}
	require.NoError(t, db.Create(user1).Error)
	require.NoError(t, db.Create(user2).Error)

	// Soft-delete user2
	db.Model(user2).Update("deleted_flag", 1)

	result, err := FindByIDs[model.User](db, ctx, []uint{user1.ID, user2.ID})
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, user1.ID)
	assert.NotContains(t, result, user2.ID)
}

// --- Non-soft-deletable type tests (no deleted_flag column) ---

func TestFindByID_ProgressRecord_WorksNormal(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	record := &model.ProgressRecord{SubItemKey: 1, TeamKey: 1, AuthorKey: 1, Completion: 50.0}
	require.NoError(t, db.Create(record).Error)

	found, err := FindByID[model.ProgressRecord](db, ctx, record.ID)
	require.NoError(t, err)
	assert.Equal(t, record.ID, found.ID)
	assert.Equal(t, float64(50.0), found.Completion)
}

func TestFindByIDs_ProgressRecord_WorksNormal(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	r1 := &model.ProgressRecord{SubItemKey: 1, TeamKey: 1, AuthorKey: 1, Completion: 50.0}
	r2 := &model.ProgressRecord{SubItemKey: 2, TeamKey: 1, AuthorKey: 1, Completion: 75.0}
	require.NoError(t, db.Create(r1).Error)
	require.NoError(t, db.Create(r2).Error)

	result, err := FindByIDs[model.ProgressRecord](db, ctx, []uint{r1.ID, r2.ID})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, r1.ID)
	assert.Contains(t, result, r2.ID)
}

// --- UpdateFields tests ---

func TestUpdateFields_MainItem_ValidFields(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})

	item := &model.MainItem{ItemStatus: "open", Code: "TST-00001", Title: "test item"}
	require.NoError(t, db.Create(item).Error)

	fields := map[string]any{"title": "updated", "priority": "P0"}
	err := UpdateFields[model.MainItem](db, ctx, item, item.TeamKey, fields)
	require.NoError(t, err)

	var updated model.MainItem
	db.First(&updated, item.ID)
	assert.Equal(t, "updated", updated.Title)
	assert.Equal(t, "P0", updated.Priority)
}

func TestUpdateFields_SubItem_ValidFields(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})
	db.Create(&model.MainItem{ItemStatus: "open", Code: "TST-00001"})

	item := &model.SubItem{TeamKey: 1, MainItemKey: int64(1), Title: "sub1", ItemStatus: "open"}
	require.NoError(t, db.Create(item).Error)

	fields := map[string]any{"title": "updated sub", "item_status": "in_progress"}
	err := UpdateFields[model.SubItem](db, ctx, item, item.TeamKey, fields)
	require.NoError(t, err)

	var updated model.SubItem
	db.First(&updated, item.ID)
	assert.Equal(t, "updated sub", updated.Title)
	assert.Equal(t, "in_progress", updated.ItemStatus)
}

func TestUpdateFields_User_ValidFields(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	user := &model.User{Username: "testuser", PasswordHash: "hash", DisplayName: "Original"}
	require.NoError(t, db.Create(user).Error)

	fields := map[string]any{"display_name": "Updated Name"}
	err := UpdateFields[model.User](db, ctx, user, 0, fields)
	require.NoError(t, err)

	var updated model.User
	db.First(&updated, user.ID)
	assert.Equal(t, "Updated Name", updated.DisplayName)
}

func TestUpdateFields_InvalidField(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})

	item := &model.MainItem{ItemStatus: "open", Code: "TST-00001", Title: "test item"}
	require.NoError(t, db.Create(item).Error)

	fields := map[string]any{"title": "ok", "evil_field": "bad"}
	err := UpdateFields[model.MainItem](db, ctx, item, item.TeamKey, fields)
	assert.ErrorIs(t, err, apperrors.ErrInvalidField)
}

func TestUpdateFields_EmptyFields(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})

	item := &model.MainItem{ItemStatus: "open", Code: "TST-00001", Title: "test item"}
	require.NoError(t, db.Create(item).Error)

	err := UpdateFields[model.MainItem](db, ctx, item, item.TeamKey, map[string]any{})
	require.NoError(t, err)

	// Verify no changes
	var unchanged model.MainItem
	db.First(&unchanged, item.ID)
	assert.Equal(t, "test item", unchanged.Title)
}

func TestUpdateFields_TeamIDMismatch(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})
	db.Create(&model.Team{Code: "OTH", TeamName: "Other"})

	item := &model.MainItem{ItemStatus: "open", Code: "TST-00001", Title: "test item"}
	require.NoError(t, db.Create(item).Error)

	fields := map[string]any{"title": "hacked"}
	err := UpdateFields[model.MainItem](db, ctx, item, 2, fields) // teamID 2 != item.TeamKey 1
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestUpdateFields_UnsupportedType(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.Create(&model.Team{Code: "TST", TeamName: "Test"})

	item := &model.MainItem{ItemStatus: "open", Code: "TST-00001", Title: "test item"}
	require.NoError(t, db.Create(item).Error)

	fields := map[string]any{"title": "ok"}
	err := UpdateFields[model.Team](db, ctx, &model.Team{Code: "TST"}, 1, fields)
	assert.ErrorIs(t, err, apperrors.ErrInvalidField)
}
