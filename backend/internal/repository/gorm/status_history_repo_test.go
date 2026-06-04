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

func setupStatusHistoryTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.StatusHistory{}))
	return db
}

func createTestStatusHistory(t *testing.T, db *gormlib.DB, itemType string, itemID uint, from, to string, changedBy uint) *model.StatusHistory {
	t.Helper()
	record := &model.StatusHistory{
		ItemType:   itemType,
		ItemKey:    int64(itemID),
		FromStatus: from,
		ToStatus:   to,
		ChangedBy:  int64(changedBy),
	}
	require.NoError(t, db.Create(record).Error)
	return record
}

// --- Constructor ---

func TestNewGormStatusHistoryRepo_NilDB(t *testing.T) {
	assert.Panics(t, func() {
		gormrepo.NewGormStatusHistoryRepo(nil)
	})
}

// --- Create ---

func TestStatusHistoryRepo_Create(t *testing.T) {
	db := setupStatusHistoryTestDB(t)
	repo := gormrepo.NewGormStatusHistoryRepo(db)
	ctx := context.Background()

	record := &model.StatusHistory{
		ItemType:   "sub_item",
		ItemKey:    int64(1),
		FromStatus: "pending",
		ToStatus:   "progressing",
		ChangedBy:  10,
	}
	require.NoError(t, repo.Create(ctx, record))
	assert.NotZero(t, record.ID)
}

// --- FindByID ---

func TestStatusHistoryRepo_FindByID_Found(t *testing.T) {
	db := setupStatusHistoryTestDB(t)
	repo := gormrepo.NewGormStatusHistoryRepo(db)
	ctx := context.Background()

	created := createTestStatusHistory(t, db, "sub_item", 1, "pending", "progressing", 10)

	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "sub_item", found.ItemType)
	assert.Equal(t, uint(1), uint(found.ItemKey))
	assert.Equal(t, "pending", found.FromStatus)
	assert.Equal(t, "progressing", found.ToStatus)
	assert.Equal(t, int64(10), found.ChangedBy)
}

func TestStatusHistoryRepo_FindByID_NotFound(t *testing.T) {
	db := setupStatusHistoryTestDB(t)
	repo := gormrepo.NewGormStatusHistoryRepo(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 9999)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- ListByItem ---

func TestStatusHistoryRepo_ListByItem(t *testing.T) {
	db := setupStatusHistoryTestDB(t)
	repo := gormrepo.NewGormStatusHistoryRepo(db)
	ctx := context.Background()

	// Create records for item 1
	createTestStatusHistory(t, db, "sub_item", 1, "pending", "progressing", 10)
	createTestStatusHistory(t, db, "sub_item", 1, "progressing", "blocking", 10)
	createTestStatusHistory(t, db, "sub_item", 1, "blocking", "progressing", 20)
	// Create record for a different item
	createTestStatusHistory(t, db, "sub_item", 2, "pending", "progressing", 10)
	// Create record for a different item type
	createTestStatusHistory(t, db, "main_item", 1, "pending", "progressing", 10)

	t.Run("returns only matching item", func(t *testing.T) {
		result, err := repo.ListByItem(ctx, "sub_item", 1, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 3)
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := repo.ListByItem(ctx, "sub_item", 1, dto.Pagination{Page: 1, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 2, result.Size)
	})

	t.Run("second_page", func(t *testing.T) {
		result, err := repo.ListByItem(ctx, "sub_item", 1, dto.Pagination{Page: 2, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, 2, result.Page)
	})

	t.Run("empty_result", func(t *testing.T) {
		result, err := repo.ListByItem(ctx, "main_item", 99, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
		assert.Empty(t, result.Items)
	})

	t.Run("defaults_page_and_pagesize", func(t *testing.T) {
		result, err := repo.ListByItem(ctx, "sub_item", 1, dto.Pagination{})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.Size)
	})
}

// --- ListByItemKeysInRange ---

func TestStatusHistoryRepo_ListByItemKeysInRange(t *testing.T) {
	db := setupStatusHistoryTestDB(t)
	repo := gormrepo.NewGormStatusHistoryRepo(db)
	ctx := context.Background()

	start := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	// Create records with explicit timestamps
	r1 := &model.StatusHistory{ItemType: "main_item", ItemKey: 1, FromStatus: "pending", ToStatus: "progressing", ChangedBy: 10}
	r1.CreateTime = time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(r1).Error)

	r2 := &model.StatusHistory{ItemType: "main_item", ItemKey: 2, FromStatus: "pending", ToStatus: "closed", ChangedBy: 10}
	r2.CreateTime = time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(r2).Error)

	// Outside range
	r3 := &model.StatusHistory{ItemType: "main_item", ItemKey: 1, FromStatus: "progressing", ToStatus: "completed", ChangedBy: 10}
	r3.CreateTime = time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(r3).Error)

	// Different item type
	r4 := &model.StatusHistory{ItemType: "sub_item", ItemKey: 1, FromStatus: "pending", ToStatus: "progressing", ChangedBy: 10}
	r4.CreateTime = time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(r4).Error)

	t.Run("returns matching records in range", func(t *testing.T) {
		records, err := repo.ListByItemKeysInRange(ctx, "main_item", []int64{1, 2}, start, end)
		require.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("filters by item type", func(t *testing.T) {
		records, err := repo.ListByItemKeysInRange(ctx, "sub_item", []int64{1}, start, end)
		require.NoError(t, err)
		assert.Len(t, records, 1)
	})

	t.Run("filters by item keys", func(t *testing.T) {
		records, err := repo.ListByItemKeysInRange(ctx, "main_item", []int64{1}, start, end)
		require.NoError(t, err)
		assert.Len(t, records, 1)
	})

	t.Run("empty item keys returns nil", func(t *testing.T) {
		records, err := repo.ListByItemKeysInRange(ctx, "main_item", nil, start, end)
		require.NoError(t, err)
		assert.Nil(t, records)
	})
}
