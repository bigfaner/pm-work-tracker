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
	assert.Equal(t, "progress_records", pr.TableName())
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
	db.Raw("PRAGMA table_info(progress_records)").Scan(&columns)
	colNames := map[string]bool{}
	for _, c := range columns {
		colNames[c.Name] = true
	}
	assert.False(t, colNames["updated_at"], "progress_records should not have updated_at column")
	assert.False(t, colNames["deleted_at"], "progress_records should not have deleted_at column")
	assert.True(t, colNames["created_at"], "progress_records should have created_at column")
}

func TestProgressRecord_Defaults(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.ProgressRecord{})
	require.NoError(t, err)

	u := model.User{Username: "pr_author", DisplayName: "Author", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "PRTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamID: team.ID, Code: "MI-PR1", Title: "Main", Priority: "P1", ProposerID: u.ID}
	require.NoError(t, db.Create(&mi).Error)
	si := model.SubItem{TeamID: team.ID, MainItemID: mi.ID, Title: "Sub", Priority: "P1"}
	require.NoError(t, db.Create(&si).Error)

	pr := model.ProgressRecord{
		SubItemID:  si.ID,
		TeamID:     team.ID,
		AuthorID:   u.ID,
		Completion: 50.0,
	}
	require.NoError(t, db.Create(&pr).Error)

	var fetched model.ProgressRecord
	db.First(&fetched, "sub_item_id = ?", si.ID)
	assert.Equal(t, float64(50.0), fetched.Completion)
	assert.False(t, fetched.IsPMCorrect, "is_pm_correct should default to false")
	assert.False(t, fetched.CreatedAt.IsZero(), "created_at should be set")
}

func TestProgressRecord_InsertAndQuery(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.ProgressRecord{})
	require.NoError(t, err)

	u := model.User{Username: "pr_q_author", DisplayName: "QAuthor", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "PRQTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamID: team.ID, Code: "MI-PRQ1", Title: "Main", Priority: "P1", ProposerID: u.ID}
	require.NoError(t, db.Create(&mi).Error)
	si := model.SubItem{TeamID: team.ID, MainItemID: mi.ID, Title: "Sub", Priority: "P1"}
	require.NoError(t, db.Create(&si).Error)

	records := []model.ProgressRecord{
		{SubItemID: si.ID, TeamID: team.ID, AuthorID: u.ID, Completion: 30.0, Achievement: "did stuff"},
		{SubItemID: si.ID, TeamID: team.ID, AuthorID: u.ID, Completion: 60.0, Blocker: "blocked"},
		{SubItemID: si.ID, TeamID: team.ID, AuthorID: u.ID, Completion: 90.0, Lesson: "learned"},
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	var fetched []model.ProgressRecord
	db.Where("sub_item_id = ? AND team_id = ?", si.ID, team.ID).Order("created_at").Find(&fetched)
	assert.Len(t, fetched, 3)
	assert.Equal(t, float64(30.0), fetched[0].Completion)
	assert.Equal(t, "did stuff", fetched[0].Achievement)
	assert.Equal(t, "blocked", fetched[1].Blocker)
	assert.Equal(t, "learned", fetched[2].Lesson)
}

// --- ItemPool tests ---

func TestItemPool_TableName(t *testing.T) {
	ip := model.ItemPool{}
	assert.Equal(t, "item_pools", ip.TableName())
}

func TestItemPool_HasSoftDelete(t *testing.T) {
	ip := model.ItemPool{}
	// ItemPool embeds gorm.Model which has DeletedAt
	assert.False(t, ip.DeletedAt.Valid, "DeletedAt should exist but be zero")
}

func TestItemPool_DefaultStatus(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.ItemPool{})
	require.NoError(t, err)

	u := model.User{Username: "ip_submitter", DisplayName: "Submitter", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "IPTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	ip := model.ItemPool{
		TeamID:      team.ID,
		Title:       "Proposed Item",
		SubmitterID: u.ID,
	}
	require.NoError(t, db.Create(&ip).Error)

	var fetched model.ItemPool
	db.First(&fetched, "title = ?", "Proposed Item")
	assert.Equal(t, "pending", fetched.Status, "status should default to pending")
}

func TestItemPool_AllFields(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}, &model.ItemPool{})
	require.NoError(t, err)

	u := model.User{Username: "ip_all", DisplayName: "AllFields", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "IPAllTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamID: team.ID, Code: "MI-IP1", Title: "Main", Priority: "P1", ProposerID: u.ID}
	require.NoError(t, db.Create(&mi).Error)
	si := model.SubItem{TeamID: team.ID, MainItemID: mi.ID, Title: "Sub", Priority: "P1"}
	require.NoError(t, db.Create(&si).Error)

	now := time.Now()
	reviewerID := u.ID
	mainID := mi.ID
	subID := si.ID
	assigneeID := u.ID

	ip := model.ItemPool{
		TeamID:         team.ID,
		Title:          "Full Item",
		Background:     "some background",
		ExpectedOutput: "some output",
		SubmitterID:    u.ID,
		Status:         "assigned",
		AssignedMainID: &mainID,
		AssignedSubID:  &subID,
		AssigneeID:     &assigneeID,
		RejectReason:   "",
		ReviewedAt:     &now,
		ReviewerID:     &reviewerID,
	}
	require.NoError(t, db.Create(&ip).Error)

	var fetched model.ItemPool
	db.First(&fetched, "title = ?", "Full Item")
	assert.Equal(t, "assigned", fetched.Status)
	assert.NotNil(t, fetched.AssignedMainID)
	assert.Equal(t, mainID, *fetched.AssignedMainID)
	assert.NotNil(t, fetched.AssignedSubID)
	assert.Equal(t, subID, *fetched.AssignedSubID)
	assert.NotNil(t, fetched.AssigneeID)
	assert.Equal(t, assigneeID, *fetched.AssigneeID)
	assert.NotNil(t, fetched.ReviewedAt)
	assert.NotNil(t, fetched.ReviewerID)
}

func TestItemPool_StatusValues(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.ItemPool{})
	require.NoError(t, err)

	u := model.User{Username: "ip_status", DisplayName: "Status", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "IPStatusTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	statuses := []string{"pending", "assigned", "rejected"}
	for i, s := range statuses {
		ip := model.ItemPool{
			TeamID:      team.ID,
			Title:       "Item " + string(rune('A'+i)),
			SubmitterID: u.ID,
			Status:      s,
		}
		require.NoError(t, db.Create(&ip).Error)
	}

	var items []model.ItemPool
	db.Where("team_id = ?", team.ID).Order("title").Find(&items)
	assert.Len(t, items, 3)
	assert.Equal(t, "pending", items[0].Status)
	assert.Equal(t, "assigned", items[1].Status)
	assert.Equal(t, "rejected", items[2].Status)
}

func TestItemPool_TeamStatusCompositeIndex(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.ItemPool{})
	require.NoError(t, err)

	u := model.User{Username: "ip_idx", DisplayName: "Idx", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "IPIdxTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	for _, s := range []string{"pending", "assigned", "rejected"} {
		ip := model.ItemPool{
			TeamID:      team.ID,
			Title:       "Pool " + s,
			SubmitterID: u.ID,
			Status:      s,
		}
		require.NoError(t, db.Create(&ip).Error)
	}

	var items []model.ItemPool
	err = db.Where("team_id = ? AND status = ?", team.ID, "assigned").Find(&items).Error
	assert.NoError(t, err)
	assert.Len(t, items, 1)
}
