package migration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

func setupBackfillTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = snowflake.Init(1)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.ProgressRecord{}))
	return db
}

func TestBackfillBizKeys_Idempotent(t *testing.T) {
	db := setupBackfillTestDB(t)

	sub := &model.SubItem{
		BaseModel:   model.BaseModel{BizKey: 0},
		TeamKey:     1,
		MainItemKey: 1,
		Code:        "T-01",
		Title:       "test sub",
		Priority:    "P1",
		ItemStatus:  "pending",
	}
	require.NoError(t, db.Create(sub).Error)
	require.Equal(t, int64(0), sub.BizKey)

	require.NoError(t, BackfillBizKeys(db))

	var updated model.SubItem
	require.NoError(t, db.First(&updated, sub.ID).Error)
	require.NotEqual(t, int64(0), updated.BizKey)

	require.NoError(t, BackfillBizKeys(db))
}

func TestBackfillBizKeys_MultipleRows(t *testing.T) {
	db := setupBackfillTestDB(t)

	subs := []*model.SubItem{
		{BaseModel: model.BaseModel{BizKey: 0}, TeamKey: 1, MainItemKey: 1, Code: "A-01", Title: "sub1", Priority: "P1", ItemStatus: "pending"},
		{BaseModel: model.BaseModel{BizKey: 0}, TeamKey: 1, MainItemKey: 1, Code: "A-02", Title: "sub2", Priority: "P1", ItemStatus: "pending"},
	}
	for _, s := range subs {
		require.NoError(t, db.Create(s).Error)
	}

	require.NoError(t, BackfillBizKeys(db))

	var results []model.SubItem
	require.NoError(t, db.Where("main_item_key = 1").Find(&results).Error)
	require.Len(t, results, 2)
	keys := map[int64]bool{}
	for _, r := range results {
		require.NotEqual(t, int64(0), r.BizKey)
		keys[r.BizKey] = true
	}
	require.Len(t, keys, 2, "biz_keys must be unique")
}

func TestBackfillBizKeys_SkipsNonZeroBizKeys(t *testing.T) {
	db := setupBackfillTestDB(t)

	existingKey := int64(12345)
	sub := &model.SubItem{
		BaseModel:   model.BaseModel{BizKey: existingKey},
		TeamKey:     1,
		MainItemKey: 1,
		Code:        "B-01",
		Title:       "sub",
		Priority:    "P1",
		ItemStatus:  "pending",
	}
	require.NoError(t, db.Create(sub).Error)

	require.NoError(t, BackfillBizKeys(db))

	var updated model.SubItem
	require.NoError(t, db.First(&updated, sub.ID).Error)
	require.Equal(t, existingKey, updated.BizKey, "existing biz_key should not be changed")
}
