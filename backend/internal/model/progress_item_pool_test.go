package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/model"
)

// --- ProgressRecord tests ---

func TestProgressRecord_TableName(t *testing.T) {
	pr := model.ProgressRecord{}
	assert.Equal(t, "pmw_progress_records", pr.TableName())
}

func TestProgressRecord_NoUpdatedAtOrDeletedAt(t *testing.T) {
	pr := model.ProgressRecord{}
	// ProgressRecord should only have ID and CreatedAt, no UpdatedAt/DeletedAt
	assert.Zero(t, pr.ID, "should have plain ID field")
	// Verify via AutoMigrate that the schema has no updated_at/deleted_at columns
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.ProgressRecord{})
	require.NoError(t, err)

	// Check columns: should NOT have updated_at or deleted_at
	type columnInfo struct {
		Name string
	}
	var columns []columnInfo
	db.Raw("PRAGMA table_info(pmw_progress_records)").Scan(&columns)
	colNames := map[string]bool{}
	for _, c := range columns {
		colNames[c.Name] = true
	}
	assert.False(t, colNames["updated_at"], "progress_records should not have updated_at column")
	assert.False(t, colNames["deleted_at"], "progress_records should not have deleted_at column")
	assert.True(t, colNames["create_time"], "progress_records should have created_at column")
}

func TestProgressRecord_Defaults(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.ProgressRecord{})
	require.NoError(t, err)

	u := model.User{Username: "pr_author", DisplayName: "Author", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "PRTeam", PmKey: int64(u.ID), Code: "PRTE"}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamKey: int64(team.ID), Code: "PRTE-00001", Title: "Main", Priority: "P1", ProposerKey: int64(u.ID)}
	require.NoError(t, db.Create(&mi).Error)
	si := model.SubItem{TeamKey: int64(team.ID), MainItemKey: int64(mi.ID), Title: "Sub", Priority: "P1"}
	require.NoError(t, db.Create(&si).Error)
	si.BizKey = 1001
	require.NoError(t, db.Save(&si).Error)

	pr := model.ProgressRecord{
		SubItemKey: si.BizKey,
		TeamKey:    int64(team.ID),
		AuthorKey:  int64(u.ID),
		Completion: 50.0,
	}
	require.NoError(t, db.Create(&pr).Error)

	var fetched model.ProgressRecord
	db.First(&fetched, "sub_item_key = ?", si.BizKey)
	assert.Equal(t, float64(50.0), fetched.Completion)
	assert.Equal(t, 0, fetched.IsPmCorrect, "is_pm_correct should default to false")
	assert.False(t, fetched.CreateTime.IsZero(), "created_at should be set")
}

func TestProgressRecord_InsertAndQuery(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.ProgressRecord{})
	require.NoError(t, err)

	u := model.User{Username: "pr_q_author", DisplayName: "QAuthor", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "PRQTeam", PmKey: int64(u.ID), Code: "PRQT"}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamKey: int64(team.ID), Code: "PRQT-00001", Title: "Main", Priority: "P1", ProposerKey: int64(u.ID)}
	require.NoError(t, db.Create(&mi).Error)
	si := model.SubItem{TeamKey: int64(team.ID), MainItemKey: int64(mi.ID), Title: "Sub", Priority: "P1"}
	require.NoError(t, db.Create(&si).Error)
	si.BizKey = 1002
	require.NoError(t, db.Save(&si).Error)

	records := []model.ProgressRecord{
		{SubItemKey: si.BizKey, TeamKey: int64(team.ID), AuthorKey: int64(u.ID), Completion: 30.0, Achievement: "did stuff"},
		{SubItemKey: si.BizKey, TeamKey: int64(team.ID), AuthorKey: int64(u.ID), Completion: 60.0, Blocker: "blocked"},
		{SubItemKey: si.BizKey, TeamKey: int64(team.ID), AuthorKey: int64(u.ID), Completion: 90.0, Lesson: "learned"},
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	var fetched []model.ProgressRecord
	db.Where("sub_item_key = ? AND team_key = ?", si.BizKey, team.ID).Order("create_time").Find(&fetched)
	assert.Len(t, fetched, 3)
	assert.Equal(t, float64(30.0), fetched[0].Completion)
	assert.Equal(t, "did stuff", fetched[0].Achievement)
	assert.Equal(t, "blocked", fetched[1].Blocker)
	assert.Equal(t, "learned", fetched[2].Lesson)
}

// --- ItemPool tests ---

func TestItemPool_TableName(t *testing.T) {
	ip := model.ItemPool{}
	assert.Equal(t, "pmw_item_pools", ip.TableName())
}

func TestItemPool_HasSoftDelete(t *testing.T) {
	ip := model.ItemPool{}
	// ItemPool embeds gorm.Model which has DeletedAt
	assert.False(t, ip.DeletedFlag != 0, "DeletedAt should exist but be zero")
}

func TestItemPool_DefaultStatus(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.ItemPool{})
	require.NoError(t, err)

	u := model.User{Username: "ip_submitter", DisplayName: "Submitter", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "IPTeam", PmKey: int64(u.ID)}
	require.NoError(t, db.Create(&team).Error)

	ip := model.ItemPool{
		TeamKey: int64(team.ID),
		Title:       "Proposed Item",
		SubmitterKey: int64(u.ID),
	}
	require.NoError(t, db.Create(&ip).Error)

	var fetched model.ItemPool
	db.First(&fetched, "title = ?", "Proposed Item")
	assert.Equal(t, "pending", fetched.PoolStatus, "status should default to pending")
}

func TestItemPool_AllFields(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.ItemPool{})
	require.NoError(t, err)

	u := model.User{Username: "ip_all", DisplayName: "AllFields", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "IPAllTeam", PmKey: int64(u.ID), Code: "IPAL"}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamKey: int64(team.ID), Code: "IPAL-00001", Title: "Main", Priority: "P1", ProposerKey: int64(u.ID)}
	require.NoError(t, db.Create(&mi).Error)
	si := model.SubItem{TeamKey: int64(team.ID), MainItemKey: int64(mi.ID), Title: "Sub", Priority: "P1"}
	require.NoError(t, db.Create(&si).Error)

	now := time.Now()
	reviewerID := u.ID
	mainID := mi.ID
	subID := si.ID
	assigneeID := u.ID

	ip := model.ItemPool{
		TeamKey: int64(team.ID),
		Title:          "Full Item",
		Background:     "some background",
		ExpectedOutput: "some output",
		SubmitterKey: int64(u.ID),
		PoolStatus: "assigned",
		AssignedMainKey: func() *int64 { v := int64(mainID); return &v }(),
		AssignedSubKey: func() *int64 { v := int64(subID); return &v }(),
		AssigneeKey: func() *int64 { v := int64(assigneeID); return &v }(),
		RejectReason:   "",
		ReviewedAt:     &now,
		ReviewerKey: func() *int64 { v := int64(reviewerID); return &v }(),
	}
	require.NoError(t, db.Create(&ip).Error)

	var fetched model.ItemPool
	db.First(&fetched, "title = ?", "Full Item")
	assert.Equal(t, "assigned", fetched.PoolStatus)
	assert.NotNil(t, fetched.AssignedMainKey)
	assert.Equal(t, int64(mainID), *fetched.AssignedMainKey)
	assert.NotNil(t, fetched.AssignedSubKey)
	assert.Equal(t, int64(subID), *fetched.AssignedSubKey)
	assert.NotNil(t, fetched.AssigneeKey)
	assert.Equal(t, int64(assigneeID), *fetched.AssigneeKey)
	assert.NotNil(t, fetched.ReviewedAt)
	assert.NotNil(t, fetched.ReviewerKey)
}

func TestItemPool_StatusValues(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.ItemPool{})
	require.NoError(t, err)

	u := model.User{Username: "ip_status", DisplayName: "Status", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "IPStatusTeam", PmKey: int64(u.ID)}
	require.NoError(t, db.Create(&team).Error)

	statuses := []string{"pending", "assigned", "rejected"}
	for i, s := range statuses {
		ip := model.ItemPool{
			TeamKey: int64(team.ID),
			Title:       "Item " + string(rune('A'+i)),
			SubmitterKey: int64(u.ID),
			PoolStatus:      s,
		}
		require.NoError(t, db.Create(&ip).Error)
	}

	var items []model.ItemPool
	db.Where("team_key = ?", team.ID).Order("title").Find(&items)
	assert.Len(t, items, 3)
	assert.Equal(t, "pending", items[0].PoolStatus)
	assert.Equal(t, "assigned", items[1].PoolStatus)
	assert.Equal(t, "rejected", items[2].PoolStatus)
}

func TestItemPool_TeamStatusCompositeIndex(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.ItemPool{})
	require.NoError(t, err)

	u := model.User{Username: "ip_idx", DisplayName: "Idx", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{TeamName: "IPIdxTeam", PmKey: int64(u.ID)}
	require.NoError(t, db.Create(&team).Error)

	for _, s := range []string{"pending", "assigned", "rejected"} {
		ip := model.ItemPool{
			TeamKey: int64(team.ID),
			Title:       "Pool " + s,
			SubmitterKey: int64(u.ID),
			PoolStatus:      s,
		}
		require.NoError(t, db.Create(&ip).Error)
	}

	var items []model.ItemPool
	err = db.Where("team_key = ? AND pool_status = ?", team.ID, "assigned").Find(&items).Error
	assert.NoError(t, err)
	assert.Len(t, items, 1)
}
