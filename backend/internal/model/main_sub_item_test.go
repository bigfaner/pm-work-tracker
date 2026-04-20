package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/model"
)

func TestMainItem_TableName(t *testing.T) {
	m := model.MainItem{}
	assert.Equal(t, "main_items", m.TableName())
}

func TestMainItem_CodeUniqueIndex(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{})
	require.NoError(t, err)

	u := model.User{Username: "pm1", DisplayName: "PM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "T1", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	m1 := model.MainItem{TeamID: team.ID, Code: "MI-0001", Title: "Item 1", Priority: "P1", ProposerID: u.ID, Status: "待开始"}
	require.NoError(t, db.Create(&m1).Error)

	m2 := model.MainItem{TeamID: team.ID, Code: "MI-0001", Title: "Item 2", Priority: "P2", ProposerID: u.ID, Status: "待开始"}
	err = db.Create(&m2).Error
	assert.Error(t, err, "duplicate code should be rejected")
}

func TestMainItem_Defaults(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{})
	require.NoError(t, err)

	u := model.User{Username: "proposer", DisplayName: "P", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "T2", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	m := model.MainItem{TeamID: team.ID, Code: "MI-0002", Title: "Item", Priority: "P2", ProposerID: u.ID}
	require.NoError(t, db.Create(&m).Error)

	var fetched model.MainItem
	db.First(&fetched, "code = ?", "MI-0002")
	assert.Equal(t, "pending", fetched.Status, "status should default to pending")
	assert.Equal(t, float64(0), fetched.Completion, "completion should default to 0")
	assert.False(t, fetched.IsKeyItem, "is_key_item should default to false")
}

func TestMainItem_ArchivedAt(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{})
	require.NoError(t, err)

	u := model.User{Username: "archiver", DisplayName: "A", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "T3", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	now := time.Now()
	m := model.MainItem{
		TeamID:     team.ID,
		Code:       "MI-0003",
		Title:      "Archived Item",
		Priority:   "P3",
		ProposerID: u.ID,
		ArchivedAt: &now,
	}
	require.NoError(t, db.Create(&m).Error)

	var fetched model.MainItem
	db.First(&fetched, "code = ?", "MI-0003")
	assert.NotNil(t, fetched.ArchivedAt, "archived_at should be set")
}

func TestSubItem_TableName(t *testing.T) {
	s := model.SubItem{}
	assert.Equal(t, "sub_items", s.TableName())
}

func TestSubItem_DefaultStatus(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{})
	require.NoError(t, err)

	u := model.User{Username: "subpm", DisplayName: "SP", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "ST1", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	mi := model.MainItem{TeamID: team.ID, Code: "MI-S1", Title: "Main", Priority: "P1", ProposerID: u.ID}
	require.NoError(t, db.Create(&mi).Error)

	s := model.SubItem{
		TeamID:     team.ID,
		MainItemID: mi.ID,
		Title:      "Sub 1",
		Priority:   "P2",
	}
	require.NoError(t, db.Create(&s).Error)

	var fetched model.SubItem
	db.First(&fetched, "title = ?", "Sub 1")
	assert.Equal(t, "pending", fetched.Status, "status should default to pending")
	assert.Equal(t, float64(0), fetched.Completion, "completion should default to 0")
	assert.Equal(t, float64(1), fetched.Weight, "weight should default to 1.0")
	assert.False(t, fetched.IsKeyItem, "is_key_item should default to false")
}

func TestSubItem_WeightCanBeCustom(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{})
	require.NoError(t, err)

	u := model.User{Username: "weightuser", DisplayName: "WU", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "WT1", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	mi := model.MainItem{TeamID: team.ID, Code: "MI-W1", Title: "Main", Priority: "P1", ProposerID: u.ID}
	require.NoError(t, db.Create(&mi).Error)

	s := model.SubItem{
		TeamID:     team.ID,
		MainItemID: mi.ID,
		Title:      "Weighted Sub",
		Priority:   "P1",
		Weight:     2.5,
	}
	require.NoError(t, db.Create(&s).Error)

	var fetched model.SubItem
	db.First(&fetched, "title = ?", "Weighted Sub")
	assert.Equal(t, 2.5, fetched.Weight, "weight should be 2.5")
}

func TestMainItem_TeamStatusAndPriorityIndexes(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{})
	require.NoError(t, err)

	u := model.User{Username: "idxuser", DisplayName: "IU", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "IdxTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	// Insert some items to exercise the indexes
	for i := 0; i < 3; i++ {
		statuses := []string{"待开始", "进行中", "已完成"}
		priorities := []string{"P1", "P2", "P3"}
		m := model.MainItem{
			TeamID:     team.ID,
			Code:       "MI-IDX" + string(rune('A'+i)),
			Title:      "Idx Item",
			Priority:   priorities[i],
			ProposerID: u.ID,
			Status:     statuses[i],
		}
		require.NoError(t, db.Create(&m).Error)
	}

	// Query using the composite indexes
	var items []model.MainItem
	err = db.Where("team_id = ? AND status = ?", team.ID, "进行中").Find(&items).Error
	assert.NoError(t, err)
	assert.Len(t, items, 1)

	err = db.Where("team_id = ? AND priority = ?", team.ID, "P1").Find(&items).Error
	assert.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestSubItem_TeamStatusAndPriorityIndexes(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{})
	require.NoError(t, err)

	u := model.User{Username: "subidx", DisplayName: "SI", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "SubIdxTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	mi := model.MainItem{TeamID: team.ID, Code: "MI-SIDX", Title: "Main", Priority: "P1", ProposerID: u.ID}
	require.NoError(t, db.Create(&mi).Error)

	for i := 0; i < 3; i++ {
		statuses := []string{"待开始", "进行中", "已完成"}
		priorities := []string{"P1", "P2", "P3"}
		s := model.SubItem{
			TeamID:     team.ID,
			MainItemID: mi.ID,
			Title:      "Idx Sub",
			Priority:   priorities[i],
			Status:     statuses[i],
		}
		require.NoError(t, db.Create(&s).Error)
	}

	var items []model.SubItem
	err = db.Where("team_id = ? AND status = ?", team.ID, "进行中").Find(&items).Error
	assert.NoError(t, err)
	assert.Len(t, items, 1)

	err = db.Where("team_id = ? AND priority = ?", team.ID, "P1").Find(&items).Error
	assert.NoError(t, err)
	assert.Len(t, items, 1)
}
