package gorm_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	pkgerrors "pm-work-tracker/backend/internal/pkg/errors"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

func setupDecisionLogTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.DecisionLog{}))
	return db
}

func createTestDecisionLog(t *testing.T, db *gormlib.DB, mainItemKey, teamKey int64, createdBy int64, status string, content string, createdAt time.Time) *model.DecisionLog {
	t.Helper()
	log := model.DecisionLog{
		BizKey:      int64(time.Now().UnixNano()),
		MainItemKey: mainItemKey,
		TeamKey:     teamKey,
		Category:    "tech",
		Tags:        `["go","gorm"]`,
		Content:     content,
		LogStatus:   status,
		CreatedBy:   createdBy,
		CreateTime:  createdAt,
	}
	require.NoError(t, db.Create(&log).Error)
	return &log
}

// --- Create ---

func TestDecisionLogRepo_Create(t *testing.T) {
	db := setupDecisionLogTestDB(t)
	repo := gormrepo.NewGormDecisionLogRepo(db)
	ctx := context.Background()

	log := &model.DecisionLog{
		BizKey:      1001,
		MainItemKey: 100,
		TeamKey:     10,
		Category:    "tech",
		Tags:        `["api"]`,
		Content:     "Use GORM for DB access",
		LogStatus:   "draft",
		CreatedBy:   1,
	}
	require.NoError(t, repo.Create(ctx, log))
	assert.NotZero(t, log.ID)
}

// --- FindByID ---

func TestDecisionLogRepo_FindByID(t *testing.T) {
	db := setupDecisionLogTestDB(t)
	repo := gormrepo.NewGormDecisionLogRepo(db)
	ctx := context.Background()

	log := createTestDecisionLog(t, db, 100, 10, 1, "published", "Some decision", time.Now())

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, log.ID)
		require.NoError(t, err)
		assert.Equal(t, log.BizKey, found.BizKey)
		assert.Equal(t, "Some decision", found.Content)
		assert.Equal(t, "published", found.LogStatus)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

// --- FindByBizKey ---

func TestDecisionLogRepo_FindByBizKey(t *testing.T) {
	db := setupDecisionLogTestDB(t)
	repo := gormrepo.NewGormDecisionLogRepo(db)
	ctx := context.Background()

	log := createTestDecisionLog(t, db, 100, 10, 1, "draft", "Draft decision", time.Now())

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByBizKey(ctx, log.BizKey)
		require.NoError(t, err)
		assert.Equal(t, log.ID, found.ID)
		assert.Equal(t, "Draft decision", found.Content)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByBizKey(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

// --- ListByItem ---

func TestDecisionLogRepo_ListByItem(t *testing.T) {
	db := setupDecisionLogTestDB(t)
	repo := gormrepo.NewGormDecisionLogRepo(db)
	ctx := context.Background()

	mainItemKey := int64(100)
	teamKey := int64(10)
	userA := int64(1)
	userB := int64(2)

	base := time.Now()

	// User A creates published and draft
	pub1 := createTestDecisionLog(t, db, mainItemKey, teamKey, userA, "published", "Published by A", base.Add(-3*time.Hour))
	_ = pub1
	draft1 := createTestDecisionLog(t, db, mainItemKey, teamKey, userA, "draft", "Draft by A", base.Add(-2*time.Hour))
	_ = draft1
	// User B creates published and draft
	createTestDecisionLog(t, db, mainItemKey, teamKey, userB, "published", "Published by B", base.Add(-1*time.Hour))
	draft2 := createTestDecisionLog(t, db, mainItemKey, teamKey, userB, "draft", "Draft by B", base)
	_ = draft2

	t.Run("user_a_sees_own_drafts_and_all_published", func(t *testing.T) {
		logs, total, err := repo.ListByItem(ctx, uint(mainItemKey), uint(userA), 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // pub1 + draft1 + pub2 (not draft2)
		assert.Len(t, logs, 3)
		// Ordered by create_time DESC
		assert.Equal(t, "Published by B", logs[0].Content)
		assert.Equal(t, "Draft by A", logs[1].Content)
		assert.Equal(t, "Published by A", logs[2].Content)
	})

	t.Run("user_b_sees_own_drafts_and_all_published", func(t *testing.T) {
		logs, total, err := repo.ListByItem(ctx, uint(mainItemKey), uint(userB), 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // pub1 + pub2 + draft2 (not draft1)
		assert.Len(t, logs, 3)
	})

	t.Run("third_user_sees_only_published", func(t *testing.T) {
		logs, total, err := repo.ListByItem(ctx, uint(mainItemKey), 999, 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total) // pub1 + pub2
		assert.Len(t, logs, 2)
		for _, l := range logs {
			assert.Equal(t, "published", l.LogStatus)
		}
	})

	t.Run("pagination_offset_and_limit", func(t *testing.T) {
		logs, total, err := repo.ListByItem(ctx, uint(mainItemKey), uint(userA), 0, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // total reflects all matching records
		assert.Len(t, logs, 2)           // but only 2 returned

		// Second page
		logs2, total2, err := repo.ListByItem(ctx, uint(mainItemKey), uint(userA), 2, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total2)
		assert.Len(t, logs2, 1)
	})

	t.Run("empty_when_no_match", func(t *testing.T) {
		logs, total, err := repo.ListByItem(ctx, 9999, uint(userA), 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, logs)
	})
}

// --- Update ---

func TestDecisionLogRepo_Update(t *testing.T) {
	db := setupDecisionLogTestDB(t)
	repo := gormrepo.NewGormDecisionLogRepo(db)
	ctx := context.Background()

	log := createTestDecisionLog(t, db, 100, 10, 1, "draft", "Original content", time.Now())

	log.Content = "Updated content"
	log.LogStatus = "published"
	require.NoError(t, repo.Update(ctx, log))

	found, err := repo.FindByID(ctx, log.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated content", found.Content)
	assert.Equal(t, "published", found.LogStatus)
}
