package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

func TestMain(m *testing.M) {
	_ = snowflake.Init(1)
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Mock repos for MainItemService tests
// ---------------------------------------------------------------------------

type mockMainItemRepo struct {
	item        *model.MainItem
	bizKeyItem  *model.MainItem
	items       []model.MainItem
	pageResult  *dto.PageResult[model.MainItem]
	nextCodeVal string
	// per-operation errors
	findErr              error
	bizKeyErr            error
	createErr            error
	updateErr            error
	listErr              error
	nextErr              error
	cascadeSoftDeleteErr error
	// capture calls
	createdItem              *model.MainItem
	updatedID                uint
	updatedFields            map[string]interface{}
	cascadeSoftDeleteCalled  bool
	cascadeSoftDeleteItemID  uint
	cascadeSoftDeleteSubIDs  []uint
	cascadeSoftDeleteHistory []model.StatusHistory
}

func (m *mockMainItemRepo) Create(_ context.Context, item *model.MainItem) error {
	m.createdItem = item
	if m.createErr != nil {
		return m.createErr
	}
	item.ID = 1
	return nil
}

func (m *mockMainItemRepo) FindByID(_ context.Context, id uint) (*model.MainItem, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findErr
}

func (m *mockMainItemRepo) Update(_ context.Context, item *model.MainItem, fields map[string]interface{}) error {
	m.updatedID = item.ID
	m.updatedFields = fields
	// Apply fields to the item so subsequent FindByID returns updated values
	if s, ok := fields["item_status"]; ok {
		item.ItemStatus = s.(string)
	}
	if c, ok := fields["completion_pct"]; ok {
		item.Completion = c.(float64)
	}
	return m.updateErr
}

func (m *mockMainItemRepo) List(_ context.Context, teamBizKey int64, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.pageResult != nil {
		return m.pageResult, nil
	}
	return &dto.PageResult[model.MainItem]{Items: m.items, Total: int64(len(m.items))}, nil
}

func (m *mockMainItemRepo) NextCode(_ context.Context, teamBizKey int64) (string, error) {
	return m.nextCodeVal, m.nextErr
}

func (m *mockMainItemRepo) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

func (m *mockMainItemRepo) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepo) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepo) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	if m.bizKeyItem != nil {
		return m.bizKeyItem, nil
	}
	return nil, m.bizKeyErr
}
func (m *mockMainItemRepo) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepo) SoftDelete(_ context.Context, _ uint) error {
	return nil
}
func (m *mockMainItemRepo) CascadeSoftDelete(_ context.Context, mainItemID uint, subItemIDs []uint, histories []model.StatusHistory) error {
	m.cascadeSoftDeleteCalled = true
	m.cascadeSoftDeleteItemID = mainItemID
	m.cascadeSoftDeleteSubIDs = subItemIDs
	m.cascadeSoftDeleteHistory = histories
	return m.cascadeSoftDeleteErr
}

func (m *mockMainItemRepo) FindByMilestoneKey(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepo) CountByMilestoneKey(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepo) ClearMilestoneKeyByMilestone(_ context.Context, _ int64) error {
	return nil
}
func (m *mockMainItemRepo) ClearMilestoneKeyByMap(_ context.Context, _ []int64) error {
	return nil
}

type mockSubItemRepo struct {
	subItems     []*model.SubItem
	teamSubItems []model.SubItem
	findErr      error
	teamListErr  error
}

func (m *mockSubItemRepo) Create(_ context.Context, item *model.SubItem) error {
	return nil
}

func (m *mockSubItemRepo) FindByID(_ context.Context, id uint) (*model.SubItem, error) {
	return nil, nil
}

func (m *mockSubItemRepo) Update(_ context.Context, item *model.SubItem, fields map[string]interface{}) error {
	return nil
}

func (m *mockSubItemRepo) List(_ context.Context, teamBizKey int64, mainItemBizKey int64, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}

func (m *mockSubItemRepo) ListByMainItem(_ context.Context, mainItemBizKey int64) ([]*model.SubItem, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.subItems, nil
}

func (m *mockSubItemRepo) ListByTeam(_ context.Context, _ int64) ([]model.SubItem, error) {
	if m.teamListErr != nil {
		return nil, m.teamListErr
	}
	return m.teamSubItems, nil
}

func (m *mockSubItemRepo) SoftDelete(_ context.Context, _ uint) error {
	return nil
}
func (m *mockSubItemRepo) FindByBizKey(_ context.Context, _ int64) (*model.SubItem, error) {
	return nil, nil
}
func (m *mockSubItemRepo) NextSubCode(_ context.Context, _ int64) (string, error) {
	return "", nil
}

type mockStatusHistorySvc struct {
	recorded *model.StatusHistory
	recordFn func(ctx context.Context, record *model.StatusHistory) error
}

func (m *mockStatusHistorySvc) Record(ctx context.Context, record *model.StatusHistory) error {
	if m.recordFn != nil {
		return m.recordFn(ctx, record)
	}
	m.recorded = record
	return nil
}

func (m *mockStatusHistorySvc) ListByItem(_ context.Context, _ string, _ uint, _ dto.Pagination) (*dto.PageResult[model.StatusHistory], error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestMainItemCreate_Success(t *testing.T) {
	mainRepo := &mockMainItemRepo{nextCodeVal: "MI-0001"}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	item, err := svc.Create(context.Background(), int64(1), 10, dto.MainItemCreateReq{
		Title:    "Feature A",
		Priority: "P0",
	})
	require.NoError(t, err)
	assert.Equal(t, "MI-0001", item.Code)
	assert.Equal(t, int64(10), item.ProposerKey)
	assert.Equal(t, "pending", item.ItemStatus)
	assert.Equal(t, int64(1), item.TeamKey)
	assert.Equal(t, "Feature A", item.Title)
	assert.Equal(t, "P0", item.Priority)
}

func TestMainItemCreate_NextCodeError(t *testing.T) {
	mainRepo := &mockMainItemRepo{nextErr: errors.New("db error")}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.Create(context.Background(), int64(1), 10, dto.MainItemCreateReq{Title: "Feature A"})
	assert.Error(t, err)
}

func TestMainItemCreate_RepoCreateError(t *testing.T) {
	mainRepo := &mockMainItemRepo{nextCodeVal: "MI-0001", createErr: errors.New("db error")}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.Create(context.Background(), int64(1), 10, dto.MainItemCreateReq{Title: "Feature A"})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
func TestMainItemCreate_SetsMilestoneKey(t *testing.T) {
	mainRepo := &mockMainItemRepo{nextCodeVal: "MI-0001"}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	msKey := "1234567890"
	item, err := svc.Create(context.Background(), int64(1), 10, dto.MainItemCreateReq{
		Title:        "Feature with milestone",
		Priority:     "P1",
		MilestoneKey: &msKey,
	})
	require.NoError(t, err)
	require.NotNil(t, item.MilestoneKey, "MilestoneKey should be set when provided in create request")
	parsedKey, _ := pkg.ParseID(msKey)
	assert.Equal(t, parsedKey, *item.MilestoneKey)
}

func TestMainItemCreate_NoMilestoneKey_Nil(t *testing.T) {
	mainRepo := &mockMainItemRepo{nextCodeVal: "MI-0001"}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	item, err := svc.Create(context.Background(), int64(1), 10, dto.MainItemCreateReq{
		Title:    "Feature without milestone",
		Priority: "P1",
	})
	require.NoError(t, err)
	assert.Nil(t, item.MilestoneKey, "MilestoneKey should be nil when not provided")
}

// Tests: Update
// ---------------------------------------------------------------------------

func TestMainItemUpdate_Success(t *testing.T) {
	existing := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1},
		TeamKey:   1,
		Title:     "Old Title",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Update(context.Background(), int64(1), 1, dto.MainItemUpdateReq{
		Title:    ptrStr("New Title"),
		Priority: ptrStr("P1"),
	})
	require.NoError(t, err)
	assert.Equal(t, uint(1), mainRepo.updatedID)
	assert.Equal(t, "New Title", mainRepo.updatedFields["title"])
	assert.Equal(t, "P1", mainRepo.updatedFields["priority"])
}

func TestMainItemUpdate_TeamMismatch(t *testing.T) {
	existing := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1},
		TeamKey:   2, // different team
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Update(context.Background(), int64(1), 1, dto.MainItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestMainItemUpdate_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Update(context.Background(), int64(1), 99, dto.MainItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

// ---------------------------------------------------------------------------
// Tests: Archive
// ---------------------------------------------------------------------------

func TestMainItemArchive_Success(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "completed",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), int64(1), 1)
	require.NoError(t, err)
	assert.NotNil(t, mainRepo.updatedFields["archived_at"])
}

func TestMainItemArchive_ClosedStatus(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "closed",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), int64(1), 1)
	require.NoError(t, err)
}

func TestMainItemArchive_NotAllowed_InProgress(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "in_progress",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), int64(1), 1)
	assert.ErrorIs(t, err, apperrors.ErrArchiveNotAllowed)
}

func TestMainItemArchive_NotAllowed_Pending(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "pending",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), int64(1), 1)
	assert.ErrorIs(t, err, apperrors.ErrArchiveNotAllowed)
}

func TestMainItemArchive_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), int64(1), 99)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestMainItemList_Success(t *testing.T) {
	items := []model.MainItem{
		{BaseModel: model.BaseModel{ID: 1}, Title: "Item 1"},
		{BaseModel: model.BaseModel{ID: 2}, Title: "Item 2"},
	}
	mainRepo := &mockMainItemRepo{items: items}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, matchInfo, err := svc.List(context.Background(), int64(1), dto.MainItemFilter{}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Nil(t, matchInfo, "no matchInfo when no filter active")
}

func TestMainItemList_RepoError(t *testing.T) {
	mainRepo := &mockMainItemRepo{listErr: errors.New("db error")}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, _, err := svc.List(context.Background(), int64(1), dto.MainItemFilter{}, dto.Pagination{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestMainItemGet_Success(t *testing.T) {
	existing := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1},
		TeamKey:   1,
		Title:     "Item 1",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	item, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Item 1", item.Title)
}

func TestMainItemGet_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.Get(context.Background(), 99)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

// ---------------------------------------------------------------------------
// Tests: GetByBizKey
// ---------------------------------------------------------------------------

func TestMainItemGetByBizKey_Success(t *testing.T) {
	existing := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1, BizKey: 123456},
		TeamKey:   1,
		Title:     "Item 1",
	}
	mainRepo := &mockMainItemRepo{}
	mainRepo.bizKeyItem = existing
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	item, err := svc.GetByBizKey(context.Background(), 123456)
	require.NoError(t, err)
	assert.Equal(t, "Item 1", item.Title)
}

func TestMainItemGetByBizKey_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{bizKeyErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.GetByBizKey(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

// ---------------------------------------------------------------------------
// Tests: RecalcCompletion
// ---------------------------------------------------------------------------

func TestRecalcCompletion_ZeroSubItems(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		Completion: 50,
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: existing}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.RecalcCompletion(context.Background(), 1)
	require.NoError(t, err)
	assert.InDelta(t, float64(0), mainRepo.updatedFields["completion_pct"], 0.001)
}

func TestRecalcCompletion_OneSubItem(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		Completion: 0,
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: existing}
	subRepo := &mockSubItemRepo{
		subItems: []*model.SubItem{
			{Completion: 60, Weight: 1.0},
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.RecalcCompletion(context.Background(), 1)
	require.NoError(t, err)
	assert.InDelta(t, float64(60), mainRepo.updatedFields["completion_pct"], 0.001)
}

func TestRecalcCompletion_MultipleSubItems_EqualWeights(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		Completion: 0,
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: existing}
	subRepo := &mockSubItemRepo{
		subItems: []*model.SubItem{
			{Completion: 30, Weight: 1.0},
			{Completion: 60, Weight: 1.0},
			{Completion: 90, Weight: 1.0},
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.RecalcCompletion(context.Background(), 1)
	require.NoError(t, err)
	// Simple average with equal weights: (30+60+90)/3 = 60
	assert.InDelta(t, float64(60), mainRepo.updatedFields["completion_pct"], 0.001)
}

func TestRecalcCompletion_AllZeroWeights_FallbackSimpleAvg(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		Completion: 0,
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: existing}
	subRepo := &mockSubItemRepo{
		subItems: []*model.SubItem{
			{Completion: 50, Weight: 0},
			{Completion: 80, Weight: 0},
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.RecalcCompletion(context.Background(), 1)
	require.NoError(t, err)
	// Simple average: (50+80)/2 = 65
	assert.InDelta(t, float64(65), mainRepo.updatedFields["completion_pct"], 0.001)
}

func TestRecalcCompletion_VaryingWeights(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		Completion: 0,
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: existing}
	subRepo := &mockSubItemRepo{
		subItems: []*model.SubItem{
			{Completion: 100, Weight: 3.0},
			{Completion: 50, Weight: 1.0},
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.RecalcCompletion(context.Background(), 1)
	require.NoError(t, err)
	// Weighted: (100*3 + 50*1) / (3+1) = 350/4 = 87.5
	assert.InDelta(t, 87.5, mainRepo.updatedFields["completion_pct"], 0.001)
}

func TestRecalcCompletion_ItemNotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{bizKeyErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.RecalcCompletion(context.Background(), 99)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func ptrStr(s string) *string {
	return &s
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus
// ---------------------------------------------------------------------------

func TestChangeStatus_AllValidTransitions(t *testing.T) {
	// All 10 legal MainItem transitions (AC-2)
	validTransitions := []struct {
		from, to string
	}{
		{"pending", "progressing"},
		{"pending", "closed"},
		{"progressing", "blocking"},
		{"progressing", "pausing"},
		{"progressing", "reviewing"},
		{"progressing", "closed"},
		{"blocking", "progressing"},
		{"pausing", "progressing"},
		{"pausing", "closed"},
		{"reviewing", "completed"},
		{"reviewing", "progressing"},
	}

	for _, tt := range validTransitions {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			item := &model.MainItem{
				BaseModel:   model.BaseModel{ID: 1},
				TeamKey:     1,
				ItemStatus:  tt.from,
				ProposerKey: int64(10), // PM
			}
			mainRepo := &mockMainItemRepo{item: item}
			historySvc := &mockStatusHistorySvc{}
			svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, historySvc)

			updated, err := svc.ChangeStatus(context.Background(), int64(1), 10, 1, tt.to)
			require.NoError(t, err)
			assert.Equal(t, tt.to, mainRepo.updatedFields["item_status"])
			assert.Equal(t, tt.to, updated.ItemStatus)

			// Verify status history recorded
			assert.NotNil(t, historySvc.recorded)
			assert.Equal(t, "main_item", historySvc.recorded.ItemType)
			assert.Equal(t, tt.from, historySvc.recorded.FromStatus)
			assert.Equal(t, tt.to, historySvc.recorded.ToStatus)
			assert.Equal(t, int64(10), historySvc.recorded.ChangedBy)
			assert.Equal(t, 0, historySvc.recorded.IsAuto)
		})
	}
}

func TestChangeStatus_SelfTransition(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamKey:     1,
		ItemStatus:  "pending",
		ProposerKey: int64(10),
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.ChangeStatus(context.Background(), int64(1), 10, 1, "pending")
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
}

func TestChangeStatus_InvalidTransitions(t *testing.T) {
	invalidTransitions := []struct {
		from, to string
	}{
		// Terminal states
		{"completed", "pending"},
		{"completed", "progressing"},
		{"closed", "pending"},
		// Non-adjacent
		{"pending", "reviewing"},
		{"pending", "completed"},
		{"blocking", "closed"},
		{"blocking", "pausing"},
		// Non-existent statuses
		{"pending", "nonexistent"},
	}

	for _, tt := range invalidTransitions {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			item := &model.MainItem{
				BaseModel:   model.BaseModel{ID: 1},
				TeamKey:     1,
				ItemStatus:  tt.from,
				ProposerKey: int64(10),
			}
			mainRepo := &mockMainItemRepo{item: item}
			svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

			_, err := svc.ChangeStatus(context.Background(), int64(1), 10, 1, tt.to)
			assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
		})
	}
}

func TestChangeStatus_PMOnly_ReviewingToCompleted(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamKey:     1,
		ItemStatus:  "reviewing",
		ProposerKey: int64(10), // PM is user 10
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	// Non-PM caller should be forbidden
	_, err := svc.ChangeStatus(context.Background(), int64(1), 99, 1, "completed")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestChangeStatus_PMOnly_ReviewingToProgressing(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamKey:     1,
		ItemStatus:  "reviewing",
		ProposerKey: int64(10),
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	// Non-PM caller should be forbidden
	_, err := svc.ChangeStatus(context.Background(), int64(1), 99, 1, "progressing")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)

	// PM caller should succeed
	_, err = svc.ChangeStatus(context.Background(), int64(1), 10, 1, "progressing")
	require.NoError(t, err)
}

func TestChangeStatus_TerminalSideEffects(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus string
		newStatus  string
	}{
		{"completed sets completion=100 and actual_end_date", "reviewing", "completed"},
		{"closed sets completion=100 and actual_end_date", "pending", "closed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &model.MainItem{
				BaseModel:   model.BaseModel{ID: 1},
				TeamKey:     1,
				ItemStatus:  tt.fromStatus,
				ProposerKey: int64(10),
				Completion:  50,
			}
			mainRepo := &mockMainItemRepo{item: item}
			svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

			_, err := svc.ChangeStatus(context.Background(), int64(1), 10, 1, tt.newStatus)
			require.NoError(t, err)

			assert.Equal(t, float64(100), mainRepo.updatedFields["completion_pct"])
			assert.NotNil(t, mainRepo.updatedFields["actual_end_date"])
		})
	}
}

func TestChangeStatus_NonTerminal_NoSideEffects(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamKey:     1,
		ItemStatus:  "pending",
		ProposerKey: int64(10),
		Completion:  30,
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.ChangeStatus(context.Background(), int64(1), 10, 1, "progressing")
	require.NoError(t, err)

	assert.Equal(t, "progressing", mainRepo.updatedFields["item_status"])
	_, hasCompletion := mainRepo.updatedFields["completion_pct"]
	assert.False(t, hasCompletion, "non-terminal transition should not set completion")
	_, hasEndDate := mainRepo.updatedFields["actual_end_date"]
	assert.False(t, hasEndDate, "non-terminal transition should not set actual_end_date")
}

func TestChangeStatus_ItemNotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.ChangeStatus(context.Background(), int64(1), 10, 999, "progressing")
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestMainItemChangeStatus_TeamMismatch(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamKey:     2,
		ItemStatus:  "pending",
		ProposerKey: int64(10),
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.ChangeStatus(context.Background(), int64(1), 10, 1, "progressing")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestChangeStatus_StatusHistoryRecorded(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1, BizKey: 100},
		TeamKey:     1,
		ItemStatus:  "pending",
		ProposerKey: int64(10),
	}
	mainRepo := &mockMainItemRepo{item: item}
	historySvc := &mockStatusHistorySvc{}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, historySvc)

	_, err := svc.ChangeStatus(context.Background(), int64(1), 10, 1, "progressing")
	require.NoError(t, err)

	require.NotNil(t, historySvc.recorded)
	assert.Equal(t, "main_item", historySvc.recorded.ItemType)
	assert.Equal(t, int64(100), historySvc.recorded.ItemKey)
	assert.Equal(t, "pending", historySvc.recorded.FromStatus)
	assert.Equal(t, "progressing", historySvc.recorded.ToStatus)
	assert.Equal(t, int64(10), historySvc.recorded.ChangedBy)
	assert.Equal(t, 0, historySvc.recorded.IsAuto)
}

// ---------------------------------------------------------------------------
// Tests: AvailableTransitions
// ---------------------------------------------------------------------------

func TestAvailableTransitions_Success(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		callerID   int64
		proposerID int64
		expected   []string
	}{
		{"pending returns progressing,closed", "pending", 10, 10, []string{"progressing", "closed"}},
		{"progressing returns 4 targets", "progressing", 10, 10, []string{"blocking", "pausing", "reviewing", "closed"}},
		{"blocking returns progressing", "blocking", 10, 10, []string{"progressing"}},
		{"pausing returns progressing,closed", "pausing", 10, 10, []string{"progressing", "closed"}},
		{"reviewing as PM returns completed,progressing", "reviewing", 10, 10, []string{"completed", "progressing"}},
		{"completed (terminal) returns empty", "completed", 10, 10, []string{}},
		{"closed (terminal) returns empty", "closed", 10, 10, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &model.MainItem{
				BaseModel:   model.BaseModel{ID: 1},
				TeamKey:     1,
				ItemStatus:  tt.status,
				ProposerKey: tt.proposerID,
			}
			mainRepo := &mockMainItemRepo{item: item}
			svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

			transitions, err := svc.AvailableTransitions(context.Background(), int64(1), tt.callerID, 1)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, transitions)
		})
	}
}

func TestAvailableTransitions_NonPMReviewing_FiltersCompletedProgressing(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamKey:     1,
		ItemStatus:  "reviewing",
		ProposerKey: int64(10), // PM is user 10
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	// Non-PM caller should not see completed/progressing
	transitions, err := svc.AvailableTransitions(context.Background(), int64(1), 99, 1)
	require.NoError(t, err)
	assert.Empty(t, transitions, "non-PM should see no transitions from reviewing since all require PM")
}

func TestAvailableTransitions_ItemNotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.AvailableTransitions(context.Background(), int64(1), 10, 999)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestAvailableTransitions_TeamMismatch(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamKey:     2,
		ItemStatus:  "pending",
		ProposerKey: int64(10),
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.AvailableTransitions(context.Background(), int64(1), 10, 1)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: EvaluateLinkage
// ---------------------------------------------------------------------------

func TestEvaluateLinkage_NoSubItems_NoLinkageTriggered(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "pending",
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEvaluateLinkage_MainItemNotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{bizKeyErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.EvaluateLinkage(context.Background(), 999, 10)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestEvaluateLinkage_SubItemRepoError(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		ItemStatus: "pending",
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{findErr: errors.New("db error")}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	assert.Error(t, err)
}

// TestEvaluateLinkage_Priority1_AllCompletedOrClosed tests Priority 1:
// all completed/closed + at least one completed -> reviewing
func TestEvaluateLinkage_Priority1_AllCompletedOrClosed(t *testing.T) {
	tests := []struct {
		name  string
		items []*model.SubItem
	}{
		{
			"all completed -> reviewing",
			[]*model.SubItem{
				{ItemStatus: "completed"},
				{ItemStatus: "completed"},
			},
		},
		{
			"mixed completed+closed -> reviewing",
			[]*model.SubItem{
				{ItemStatus: "completed"},
				{ItemStatus: "closed"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainItem := &model.MainItem{
				BaseModel:  model.BaseModel{ID: 1},
				ItemStatus: "progressing",
			}
			mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
			subRepo := &mockSubItemRepo{subItems: tt.items}
			historySvc := &mockStatusHistorySvc{}
			svc := NewMainItemService(mainRepo, subRepo, historySvc)

			result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.Triggered)
			assert.True(t, result.Success)
			assert.Equal(t, "reviewing", result.TargetStatus)
			assert.Equal(t, "reviewing", mainRepo.updatedFields["item_status"])

			// StatusHistory should be recorded with is_auto=true
			assert.NotNil(t, historySvc.recorded)
			assert.Equal(t, 1, historySvc.recorded.IsAuto)
		})
	}
}

// TestEvaluateLinkage_Priority2_AllClosed tests Priority 2:
// all closed -> closed
func TestEvaluateLinkage_Priority2_AllClosed(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		ItemStatus: "pending",
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{ItemStatus: "closed"},
		{ItemStatus: "closed"},
	}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "closed", result.TargetStatus)
	// Terminal side effects
	assert.Equal(t, float64(100), mainRepo.updatedFields["completion_pct"])
	assert.NotNil(t, mainRepo.updatedFields["actual_end_date"])
}

// TestEvaluateLinkage_Priority3_AllPausing tests Priority 3:
// all pausing (or pausing + closed) -> pausing
func TestEvaluateLinkage_Priority3_AllPausing(t *testing.T) {
	tests := []struct {
		name  string
		items []*model.SubItem
	}{
		{
			"all pausing -> pausing",
			[]*model.SubItem{
				{ItemStatus: "pausing"},
				{ItemStatus: "pausing"},
			},
		},
		{
			"pausing + closed -> pausing",
			[]*model.SubItem{
				{ItemStatus: "pausing"},
				{ItemStatus: "closed"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainItem := &model.MainItem{
				BaseModel:  model.BaseModel{ID: 1},
				ItemStatus: "progressing",
			}
			mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
			subRepo := &mockSubItemRepo{subItems: tt.items}
			svc := NewMainItemService(mainRepo, subRepo, nil)

			result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.Success)
			assert.Equal(t, "pausing", result.TargetStatus)
		})
	}
}

// TestEvaluateLinkage_Priority4_AnyBlocking tests Priority 4:
// any blocking (not all terminal) -> blocking (only from progressing, since pending->blocking is not a valid transition)
func TestEvaluateLinkage_Priority4_AnyBlocking(t *testing.T) {
	tests := []struct {
		name        string
		mainStatus  string
		items       []*model.SubItem
		wantTarget  string // expected target status, empty means no linkage
		wantSuccess bool   // whether linkage should succeed
	}{
		{
			"progressing + blocking sub -> blocking (success)",
			"progressing",
			[]*model.SubItem{{ItemStatus: "blocking"}, {ItemStatus: "pending"}},
			"blocking",
			true,
		},
		{
			"pending + blocking sub -> blocking (fails: pending->blocking not valid)",
			"pending",
			[]*model.SubItem{{ItemStatus: "blocking"}, {ItemStatus: "pending"}},
			"blocking",
			false, // pending->blocking is not a valid MainItem transition
		},
		{
			"reviewing + blocking sub -> progressing (via AC-9 revert)",
			"reviewing",
			[]*model.SubItem{{ItemStatus: "blocking"}, {ItemStatus: "pending"}},
			"progressing",
			true, // reviewing->progressing is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainItem := &model.MainItem{
				BaseModel:  model.BaseModel{ID: 1},
				ItemStatus: tt.mainStatus,
			}
			mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
			subRepo := &mockSubItemRepo{subItems: tt.items}
			svc := NewMainItemService(mainRepo, subRepo, nil)

			result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantTarget, result.TargetStatus)
			assert.Equal(t, tt.wantSuccess, result.Success)
		})
	}
}

// TestEvaluateLinkage_Priority5_AnyProgressing tests Priority 5:
// any progressing -> progressing (only from pending)
func TestEvaluateLinkage_Priority5_AnyProgressing(t *testing.T) {
	tests := []struct {
		name            string
		mainStatus      string
		items           []*model.SubItem
		wantProgressing bool
	}{
		{
			"pending + progressing sub -> progressing",
			"pending",
			[]*model.SubItem{{ItemStatus: "progressing"}, {ItemStatus: "pending"}},
			true,
		},
		{
			"progressing main + progressing sub -> no change (same status)",
			"progressing",
			[]*model.SubItem{{ItemStatus: "progressing"}, {ItemStatus: "pending"}},
			false,
		},
		{
			"blocking main + progressing sub -> no linkage",
			"blocking",
			[]*model.SubItem{{ItemStatus: "progressing"}, {ItemStatus: "pending"}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainItem := &model.MainItem{
				BaseModel:  model.BaseModel{ID: 1},
				ItemStatus: tt.mainStatus,
			}
			mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
			subRepo := &mockSubItemRepo{subItems: tt.items}
			svc := NewMainItemService(mainRepo, subRepo, nil)

			result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
			require.NoError(t, err)
			if tt.wantProgressing {
				require.NotNil(t, result)
				assert.True(t, result.Success)
				assert.Equal(t, "progressing", result.TargetStatus)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// TestEvaluateLinkage_ReviewingAndNewPending tests AC-9:
// reviewing + new pending subitem -> MainItem reverts to progressing
func TestEvaluateLinkage_ReviewingAndNewPending(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		ItemStatus: "reviewing",
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{ItemStatus: "completed"},
		{ItemStatus: "pending"},
	}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "progressing", result.TargetStatus)
	assert.Equal(t, "progressing", mainRepo.updatedFields["item_status"])
}

// TestEvaluateLinkage_Failure_TransitionNotAllowed tests AC-12:
// linkage failure when transition not allowed
func TestEvaluateLinkage_Failure_TransitionNotAllowed(t *testing.T) {
	// When all sub-items are completed/closed, target is "reviewing"
	// But if main is in a state that can't transition to "reviewing",
	// linkage should fail and record intent in status history.
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		ItemStatus: "blocking", // blocking -> reviewing is not valid
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{ItemStatus: "completed"},
		{ItemStatus: "completed"},
	}}
	historySvc := &mockStatusHistorySvc{}
	svc := NewMainItemService(mainRepo, subRepo, historySvc)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Triggered)
	assert.False(t, result.Success)
	assert.Equal(t, "reviewing", result.TargetStatus)
	assert.Contains(t, result.Warning(), "主事项状态联动失败")
	assert.Contains(t, result.Remark, "blocking→reviewing 不允许")

	// Status history should record the intent
	assert.NotNil(t, historySvc.recorded)
	assert.Equal(t, 1, historySvc.recorded.IsAuto)
	assert.Equal(t, "blocking", historySvc.recorded.FromStatus)
	assert.Equal(t, "reviewing", historySvc.recorded.ToStatus)
	assert.Contains(t, historySvc.recorded.Remark, "不允许")
}

// TestEvaluateLinkage_SameStatus_NoTransition tests that no linkage is triggered
// when the target status matches current status.
func TestEvaluateLinkage_SameStatus_NoTransition(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		ItemStatus: "reviewing", // All completed would target reviewing -> same status
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{ItemStatus: "completed"},
		{ItemStatus: "completed"},
	}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Nil(t, result)
}

// TestEvaluateLinkage_TerminalSideEffects tests that terminal side effects
// are applied when linkage transitions to a terminal status.
func TestEvaluateLinkage_TerminalSideEffects(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		ItemStatus: "pending",
		Completion: 30,
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{ItemStatus: "closed"},
		{ItemStatus: "closed"},
	}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "closed", result.TargetStatus)
	assert.Equal(t, float64(100), mainRepo.updatedFields["completion_pct"])
	assert.NotNil(t, mainRepo.updatedFields["actual_end_date"])
}

// TestEvaluateLinkage_StatusHistoryIsAuto tests AC-15:
// is_auto=true for linkage transitions.
func TestEvaluateLinkage_StatusHistoryIsAuto(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		ItemStatus: "progressing",
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{ItemStatus: "completed"},
	}}
	historySvc := &mockStatusHistorySvc{}
	svc := NewMainItemService(mainRepo, subRepo, historySvc)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)

	require.NotNil(t, historySvc.recorded)
	assert.Equal(t, 1, historySvc.recorded.IsAuto, "linkage status history should have is_auto=true")
	assert.Equal(t, int64(10), historySvc.recorded.ChangedBy)
	assert.Equal(t, "progressing", historySvc.recorded.FromStatus)
	assert.Equal(t, "reviewing", historySvc.recorded.ToStatus)
}

// TestLinkageResult_Warning tests the Warning() method.
func TestLinkageResult_Warning(t *testing.T) {
	tests := []struct {
		name   string
		result *LinkageResult
		want   string
	}{
		{"nil result", nil, ""},
		{"not triggered", &LinkageResult{Triggered: false}, ""},
		{"triggered and succeeded", &LinkageResult{Triggered: true, Success: true}, ""},
		{"triggered and failed", &LinkageResult{Triggered: true, Success: false, Remark: "test reason"}, "主事项状态联动失败：test reason"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.result.Warning())
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: getLinkageMutex LRU eviction
// ---------------------------------------------------------------------------

func TestGetLinkageMutex_BasicAccess(t *testing.T) {
	resetLinkageMuMap()
	defer resetLinkageMuMap()

	mu := getLinkageMutex(1)
	assert.NotNil(t, mu)
	// Same ID returns same mutex
	mu2 := getLinkageMutex(1)
	assert.Same(t, mu, mu2)
}

func TestGetLinkageMutex_LRUEviction(t *testing.T) {
	resetLinkageMuMap()
	defer resetLinkageMuMap()

	// Fill up to capacity
	for i := 0; i < maxLinkageMuMapSize; i++ {
		getLinkageMutex(int64(i))
	}
	assert.Len(t, linkageMuMap, maxLinkageMuMapSize)

	// Access ID 0 to make it recently used
	getLinkageMutex(0)

	// Adding one more should evict the LRU entry (ID 1, since ID 0 was just accessed)
	getLinkageMutex(int64(maxLinkageMuMapSize))
	assert.Len(t, linkageMuMap, maxLinkageMuMapSize)

	// ID 0 should still exist (was accessed recently)
	_, ok := linkageMuMap[0]
	assert.True(t, ok, "ID 0 should still exist (recently accessed)")

	// ID 1 should have been evicted (oldest unused)
	_, ok = linkageMuMap[1]
	assert.False(t, ok, "ID 1 should have been evicted as LRU")
}

func TestGetLinkageMutex_CapacityBounded(t *testing.T) {
	resetLinkageMuMap()
	defer resetLinkageMuMap()

	// Add far more than capacity
	for i := 0; i < maxLinkageMuMapSize+100; i++ {
		getLinkageMutex(int64(i))
	}
	assert.Len(t, linkageMuMap, maxLinkageMuMapSize)
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestMainItemDelete_Success_NoSubItems(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1, BizKey: 100},
		ItemStatus: "progressing",
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Delete(context.Background(), 1, 100, 10)
	require.NoError(t, err)

	assert.True(t, mainRepo.cascadeSoftDeleteCalled)
	assert.Equal(t, uint(1), mainRepo.cascadeSoftDeleteItemID)
	assert.Empty(t, mainRepo.cascadeSoftDeleteSubIDs)
	// Should have 1 history record for the main item
	require.Len(t, mainRepo.cascadeSoftDeleteHistory, 1)
	assert.Equal(t, "main_item", mainRepo.cascadeSoftDeleteHistory[0].ItemType)
	assert.Equal(t, int64(100), mainRepo.cascadeSoftDeleteHistory[0].ItemKey)
	assert.Equal(t, "progressing", mainRepo.cascadeSoftDeleteHistory[0].FromStatus)
	assert.Equal(t, "deleted", mainRepo.cascadeSoftDeleteHistory[0].ToStatus)
	assert.Equal(t, int64(10), mainRepo.cascadeSoftDeleteHistory[0].ChangedBy)
}

func TestMainItemDelete_Success_WithSubItems(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1, BizKey: 100},
		ItemStatus: "progressing",
	}
	subItems := []*model.SubItem{
		{BaseModel: model.BaseModel{ID: 2, BizKey: 200}, ItemStatus: "pending"},
		{BaseModel: model.BaseModel{ID: 3, BizKey: 300}, ItemStatus: "progressing"},
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{subItems: subItems}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Delete(context.Background(), 1, 100, 10)
	require.NoError(t, err)

	assert.True(t, mainRepo.cascadeSoftDeleteCalled)
	assert.Equal(t, uint(1), mainRepo.cascadeSoftDeleteItemID)
	assert.Equal(t, []uint{2, 3}, mainRepo.cascadeSoftDeleteSubIDs)
	// Should have 3 history records: 1 main item + 2 sub items
	require.Len(t, mainRepo.cascadeSoftDeleteHistory, 3)
	assert.Equal(t, "main_item", mainRepo.cascadeSoftDeleteHistory[0].ItemType)
	assert.Equal(t, "sub_item", mainRepo.cascadeSoftDeleteHistory[1].ItemType)
	assert.Equal(t, int64(200), mainRepo.cascadeSoftDeleteHistory[1].ItemKey)
	assert.Equal(t, "sub_item", mainRepo.cascadeSoftDeleteHistory[2].ItemType)
	assert.Equal(t, int64(300), mainRepo.cascadeSoftDeleteHistory[2].ItemKey)
}

func TestMainItemDelete_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{bizKeyErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Delete(context.Background(), 1, 999, 10)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
	assert.False(t, mainRepo.cascadeSoftDeleteCalled)
}

func TestMainItemDelete_SubItemListError(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1, BizKey: 100},
		ItemStatus: "progressing",
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem}
	subRepo := &mockSubItemRepo{findErr: errors.New("db error")}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Delete(context.Background(), 1, 100, 10)
	assert.Error(t, err)
	assert.False(t, mainRepo.cascadeSoftDeleteCalled)
}

func TestMainItemDelete_CascadeError(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1, BizKey: 100},
		ItemStatus: "progressing",
	}
	mainRepo := &mockMainItemRepo{bizKeyItem: mainItem, cascadeSoftDeleteErr: errors.New("tx error")}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Delete(context.Background(), 1, 100, 10)
	assert.Error(t, err)
	assert.True(t, mainRepo.cascadeSoftDeleteCalled)
}

// ---------------------------------------------------------------------------
// Tests: Filter Penetration (List with assigneeKey)
// ---------------------------------------------------------------------------

// helper: string pointer for filter tests
func filterStrPtr(v string) *string { return &v }

func TestList_StatusOnly_ReturnsDirect(t *testing.T) {
	// AC-1: Only status filter -> returns status-matched items, all direct matchType
	// The repo mock returns pre-filtered items (simulating SQL WHERE status IN (...))
	assignee := int64(100)
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing", AssigneeKey: &assignee},
			{BaseModel: model.BaseModel{BizKey: 3}, ItemStatus: "progressing"},
		},
	}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, matchInfo, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{Statuses: []string{"progressing"}},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.NotNil(t, matchInfo)
	assert.Equal(t, "direct", matchInfo[1].MatchType)
	assert.Equal(t, "direct", matchInfo[3].MatchType)
	assert.Nil(t, matchInfo[1].MatchedSubItemIds)
}

func TestList_AssigneeOnly_DirectMatch(t *testing.T) {
	// AC-2: Only assignee -> returns direct match
	assignee := int64(100)
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing", AssigneeKey: &assignee},
			{BaseModel: model.BaseModel{BizKey: 2}, ItemStatus: "pending"},
		},
	}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, matchInfo, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{AssigneeKey: filterStrPtr("100")},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, int64(1), result.Items[0].BizKey)
	assert.NotNil(t, matchInfo)
	assert.Equal(t, "direct", matchInfo[1].MatchType)
}

func TestList_AssigneeOnly_IndirectMatch(t *testing.T) {
	// AC-2: Only assignee -> returns indirect match via sub-items
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing"},
		},
	}
	subAssignee := int64(100)
	subRepo := &mockSubItemRepo{
		teamSubItems: []model.SubItem{
			{BaseModel: model.BaseModel{BizKey: 10}, MainItemKey: 1, AssigneeKey: &subAssignee},
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, matchInfo, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{AssigneeKey: filterStrPtr("100")},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, int64(1), result.Items[0].BizKey)
	assert.Equal(t, "indirect", matchInfo[1].MatchType)
	assert.Equal(t, []string{"10"}, matchInfo[1].MatchedSubItemIds)
}

func TestList_StatusAndAssignee_ANDLogic(t *testing.T) {
	// AC-3: Both status and assignee -> AND logic
	assignee := int64(100)
	subAssignee := int64(100)
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing", AssigneeKey: &assignee}, // matches both
			{BaseModel: model.BaseModel{BizKey: 2}, ItemStatus: "pending", AssigneeKey: &assignee},     // wrong status
			{BaseModel: model.BaseModel{BizKey: 3}, ItemStatus: "progressing"},                         // no assignee
			{BaseModel: model.BaseModel{BizKey: 4}, ItemStatus: "blocking"},                            // wrong status + sub match
		},
	}
	subRepo := &mockSubItemRepo{
		teamSubItems: []model.SubItem{
			{BaseModel: model.BaseModel{BizKey: 10}, MainItemKey: 3, AssigneeKey: &subAssignee}, // main 3 matches status + sub assignee
			{BaseModel: model.BaseModel{BizKey: 11}, MainItemKey: 4, AssigneeKey: &subAssignee}, // main 4 wrong status
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, matchInfo, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{Statuses: []string{"progressing"}, AssigneeKey: filterStrPtr("100")},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	// MainItem 1: direct match (status + assignee)
	assert.Equal(t, "direct", matchInfo[1].MatchType)
	// MainItem 3: indirect match (status + sub-item assignee)
	assert.Equal(t, "indirect", matchInfo[3].MatchType)
	assert.Equal(t, []string{"10"}, matchInfo[3].MatchedSubItemIds)
}

func TestList_MatchedSubItemIds_OnlyIncludesMatching(t *testing.T) {
	// AC-4: matchedSubItemIds only contains matching sub-items
	subAssignee := int64(100)
	otherAssignee := int64(200)
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing"},
		},
	}
	subRepo := &mockSubItemRepo{
		teamSubItems: []model.SubItem{
			{BaseModel: model.BaseModel{BizKey: 10}, MainItemKey: 1, AssigneeKey: &subAssignee},   // matches
			{BaseModel: model.BaseModel{BizKey: 11}, MainItemKey: 1, AssigneeKey: &otherAssignee}, // no match
			{BaseModel: model.BaseModel{BizKey: 12}, MainItemKey: 1, AssigneeKey: &subAssignee},   // matches
			{BaseModel: model.BaseModel{BizKey: 13}, MainItemKey: 1},                              // no assignee
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, matchInfo, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{AssigneeKey: filterStrPtr("100")},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, []string{"10", "12"}, matchInfo[1].MatchedSubItemIds)
}

func TestList_StatusAndAssignee_StatusBlocksIndirect(t *testing.T) {
	// AC-3 strict AND: main item status must match, even if sub-items match assignee
	subAssignee := int64(100)
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "pending"},     // wrong status, has matching sub
			{BaseModel: model.BaseModel{BizKey: 2}, ItemStatus: "progressing"}, // correct status, has matching sub
		},
	}
	subRepo := &mockSubItemRepo{
		teamSubItems: []model.SubItem{
			{BaseModel: model.BaseModel{BizKey: 10}, MainItemKey: 1, AssigneeKey: &subAssignee},
			{BaseModel: model.BaseModel{BizKey: 11}, MainItemKey: 2, AssigneeKey: &subAssignee},
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, _, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{Statuses: []string{"progressing"}, AssigneeKey: filterStrPtr("100")},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, int64(2), result.Items[0].BizKey) // only main 2 passes AND filter
}

func TestList_NoFilter_ReturnsAllNoMatchInfo(t *testing.T) {
	// AC: No filter -> all items, no matchType
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing"},
			{BaseModel: model.BaseModel{BizKey: 2}, ItemStatus: "pending"},
		},
	}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, matchInfo, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Nil(t, matchInfo, "no matchInfo when no filter active")
}

func TestList_AssigneeMultiStatus_ANDLogic(t *testing.T) {
	// Multiple statuses with assignee -> OR across statuses, AND with assignee
	assignee := int64(100)
	subAssignee := int64(100)
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing", AssigneeKey: &assignee},
			{BaseModel: model.BaseModel{BizKey: 2}, ItemStatus: "blocking", AssigneeKey: &assignee},
			{BaseModel: model.BaseModel{BizKey: 3}, ItemStatus: "pending", AssigneeKey: &assignee}, // excluded by status
		},
	}
	subRepo := &mockSubItemRepo{
		teamSubItems: []model.SubItem{
			{BaseModel: model.BaseModel{BizKey: 10}, MainItemKey: 3, AssigneeKey: &subAssignee}, // main 3 excluded by status
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, _, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{Statuses: []string{"progressing", "blocking"}, AssigneeKey: filterStrPtr("100")},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2) // main 1 and 2 match, main 3 excluded
}

func TestList_DirectAndIndirectBothMatch(t *testing.T) {
	// When main item's assignee matches AND it has matching sub-items
	assignee := int64(100)
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing", AssigneeKey: &assignee},
		},
	}
	subRepo := &mockSubItemRepo{
		teamSubItems: []model.SubItem{
			{BaseModel: model.BaseModel{BizKey: 10}, MainItemKey: 1, AssigneeKey: &assignee},
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, matchInfo, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{AssigneeKey: filterStrPtr("100")},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, "direct", matchInfo[1].MatchType)
	assert.Equal(t, []string{"10"}, matchInfo[1].MatchedSubItemIds)
}

func TestList_InvalidAssigneeKey_ReturnsEmpty(t *testing.T) {
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing"},
		},
	}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, _, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{AssigneeKey: filterStrPtr("invalid")},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
}

// ---------------------------------------------------------------------------
// Tests: Terminal Sort (Interface 6: #11)
// ---------------------------------------------------------------------------

func TestList_TerminalSort_MixedStatus_TerminalSinksToBottom(t *testing.T) {
	// AC-1: Mixed status list — terminal items sink to bottom
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "completed"},
			{BaseModel: model.BaseModel{BizKey: 2}, ItemStatus: "progressing"},
			{BaseModel: model.BaseModel{BizKey: 3}, ItemStatus: "closed"},
			{BaseModel: model.BaseModel{BizKey: 4}, ItemStatus: "pending"},
		},
	}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, _, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, result.Items, 4)

	// Non-terminal items come first (order preserved)
	assert.Equal(t, int64(2), result.Items[0].BizKey)
	assert.Equal(t, "progressing", result.Items[0].ItemStatus)
	assert.Equal(t, int64(4), result.Items[1].BizKey)
	assert.Equal(t, "pending", result.Items[1].ItemStatus)
	// Terminal items sink to bottom (relative order preserved)
	assert.Equal(t, int64(1), result.Items[2].BizKey)
	assert.Equal(t, "completed", result.Items[2].ItemStatus)
	assert.Equal(t, int64(3), result.Items[3].BizKey)
	assert.Equal(t, "closed", result.Items[3].ItemStatus)
}

func TestList_TerminalSort_AllTerminal_RelativeOrderPreserved(t *testing.T) {
	// AC-2: All terminal — relative order preserved
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "completed"},
			{BaseModel: model.BaseModel{BizKey: 2}, ItemStatus: "closed"},
			{BaseModel: model.BaseModel{BizKey: 3}, ItemStatus: "completed"},
		},
	}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, _, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, result.Items, 3)

	// All terminal — original relative order preserved
	assert.Equal(t, int64(1), result.Items[0].BizKey)
	assert.Equal(t, int64(2), result.Items[1].BizKey)
	assert.Equal(t, int64(3), result.Items[2].BizKey)
}

func TestList_TerminalSort_NoTerminal_Unchanged(t *testing.T) {
	// AC-3: No terminal items — result unchanged
	mainRepo := &mockMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{BizKey: 1}, ItemStatus: "progressing"},
			{BaseModel: model.BaseModel{BizKey: 2}, ItemStatus: "pending"},
			{BaseModel: model.BaseModel{BizKey: 3}, ItemStatus: "blocking"},
		},
	}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, _, err := svc.List(context.Background(), 1,
		dto.MainItemFilter{},
		dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, result.Items, 3)

	// No terminal — original order preserved
	assert.Equal(t, int64(1), result.Items[0].BizKey)
	assert.Equal(t, int64(2), result.Items[1].BizKey)
	assert.Equal(t, int64(3), result.Items[2].BizKey)
}

// ---------------------------------------------------------------------------
// Tests: Milestone Binding Validation (BR-3, BR-5)
// ---------------------------------------------------------------------------

//nolint:misspell // "cancelled" is a domain status value per PRD/API contract, not a misspelling.
func TestUpdate_MilestoneBinding_TerminalItemCannotMove_BR3(t *testing.T) {
	// BR-3: terminal MI (completed/closed) cannot change milestone_key
	tests := []struct {
		name   string
		status string
	}{
		{"completed", "completed"},
		{"closed", "closed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msKey := int64(500)
			existing := &model.MainItem{
				BaseModel:  model.BaseModel{ID: 1},
				TeamKey:    1,
				ItemStatus: tt.status,
			}
			mainRepo := &mockMainItemRepo{item: existing}
			subRepo := &mockSubItemRepo{}
			msRepo := &mockMilestoneRepo{
				item: &model.Milestone{MilestoneStatus: "not_started"},
			}
			mapRepo := &mockMilestoneMapRepo{}
			svc := NewMainItemService(mainRepo, subRepo, nil)
			svc = WithMilestoneRepos(svc, msRepo, mapRepo)

			err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
				MilestoneKey: ptrStr(pkg.FormatID(msKey)),
			})
			assert.ErrorIs(t, err, apperrors.ErrTerminalItemCannotMove)
		})
	}
}

//nolint:misspell // "cancelled" is a domain status value per PRD/API contract, not a misspelling.
func TestUpdate_MilestoneBinding_TerminalMilestoneCannotReceive_BR3(t *testing.T) {
	// BR-3: target milestone in terminal state (cancelled) cannot receive MI
	tests := []struct {
		name   string
		status string
	}{
		{"cancelled milestone", "cancelled"},
		{"completed milestone", "completed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msKey := int64(500)
			existing := &model.MainItem{
				BaseModel:  model.BaseModel{ID: 1},
				TeamKey:    1,
				ItemStatus: "progressing",
			}
			mainRepo := &mockMainItemRepo{item: existing}
			subRepo := &mockSubItemRepo{}
			msRepo := &mockMilestoneRepo{
				item: &model.Milestone{
					BaseModel:       model.BaseModel{BizKey: msKey},
					MilestoneStatus: tt.status,
					MilestoneMapKey: 100,
				},
			}
			mapRepo := &mockMilestoneMapRepo{
				item: &model.MilestoneMap{MapStatus: "executing"},
			}
			svc := NewMainItemService(mainRepo, subRepo, nil)
			svc = WithMilestoneRepos(svc, msRepo, mapRepo)

			err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
				MilestoneKey: ptrStr(pkg.FormatID(msKey)),
			})
			assert.ErrorIs(t, err, apperrors.ErrTerminalMilestoneCannotReceive)
		})
	}
}

//nolint:misspell // "cancelled" is a domain status value per PRD/API contract, not a misspelling.
func TestUpdate_MilestoneBinding_ValidationOrder_BR3(t *testing.T) {
	// BR-3: check MI status first, then target milestone
	// When MI is terminal AND target milestone is terminal, should return ErrTerminalItemCannotMove
	msKey := int64(500)
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "completed", // terminal MI
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	msRepo := &mockMilestoneRepo{
		item: &model.Milestone{
			MilestoneStatus: "cancelled", // terminal milestone
		},
	}
	mapRepo := &mockMilestoneMapRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)
	svc = WithMilestoneRepos(svc, msRepo, mapRepo)

	err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
		MilestoneKey: ptrStr(pkg.FormatID(msKey)),
	})
	// Should hit terminal MI check first
	assert.ErrorIs(t, err, apperrors.ErrTerminalItemCannotMove)
}

//nolint:misspell // "cancelled" is a domain status value per PRD/API contract, not a misspelling.
func TestUpdate_MilestoneBinding_ParentMapTerminal_BR5(t *testing.T) {
	// BR-5: parent milestone map in terminal state prevents milestone_key changes
	tests := []struct {
		name      string
		mapStatus string
	}{
		{"completed map", "completed"},
		{"cancelled map", "cancelled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msKey := int64(500)
			existing := &model.MainItem{
				BaseModel:  model.BaseModel{ID: 1},
				TeamKey:    1,
				ItemStatus: "progressing",
			}
			mainRepo := &mockMainItemRepo{item: existing}
			subRepo := &mockSubItemRepo{}
			msRepo := &mockMilestoneRepo{
				item: &model.Milestone{
					BaseModel:       model.BaseModel{BizKey: msKey},
					MilestoneStatus: "in_progress", // non-terminal milestone
					MilestoneMapKey: 100,
				},
			}
			mapRepo := &mockMilestoneMapRepo{
				item: &model.MilestoneMap{MapStatus: tt.mapStatus},
			}
			svc := NewMainItemService(mainRepo, subRepo, nil)
			svc = WithMilestoneRepos(svc, msRepo, mapRepo)

			err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
				MilestoneKey: ptrStr(pkg.FormatID(msKey)),
			})
			assert.ErrorIs(t, err, apperrors.ErrMapIsTerminal)
		})
	}
}

func TestUpdate_MilestoneBinding_Success_BindToMilestone(t *testing.T) {
	// Successful bind: non-terminal MI + non-terminal milestone + non-terminal map
	msKey := int64(500)
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "progressing",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	msRepo := &mockMilestoneRepo{
		item: &model.Milestone{
			BaseModel:       model.BaseModel{BizKey: msKey},
			MilestoneStatus: "in_progress",
			MilestoneMapKey: 100,
		},
	}
	mapRepo := &mockMilestoneMapRepo{
		item: &model.MilestoneMap{MapStatus: "executing"},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)
	svc = WithMilestoneRepos(svc, msRepo, mapRepo)

	err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
		MilestoneKey: ptrStr(pkg.FormatID(msKey)),
	})
	require.NoError(t, err)
	assert.Equal(t, msKey, mainRepo.updatedFields["milestone_key"])
}

func TestUpdate_MilestoneBinding_Success_Unbind(t *testing.T) {
	// Unbind: set milestone_key to empty string
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "progressing",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	msRepo := &mockMilestoneRepo{}
	mapRepo := &mockMilestoneMapRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)
	svc = WithMilestoneRepos(svc, msRepo, mapRepo)

	err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
		MilestoneKey: ptrStr(""),
	})
	require.NoError(t, err)
	assert.Nil(t, mainRepo.updatedFields["milestone_key"])
}

func TestUpdate_MilestoneBinding_NoMilestoneKey_NoValidation(t *testing.T) {
	// When MilestoneKey is nil in the request, no milestone validation happens
	// (existing behavior: just update other fields)
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "completed", // terminal, but MilestoneKey is nil so no validation
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	// Should fail on ErrTerminalMainItem (existing guard), NOT milestone-related error
	assert.ErrorIs(t, err, apperrors.ErrTerminalMainItem)
}

func TestUpdate_MilestoneBinding_TargetMilestoneNotFound(t *testing.T) {
	// Target milestone not found -> should return not-found error
	msKey := int64(500)
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamKey:    1,
		ItemStatus: "progressing",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	msRepo := &mockMilestoneRepo{findByBizErr: gorm.ErrRecordNotFound}
	mapRepo := &mockMilestoneMapRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)
	svc = WithMilestoneRepos(svc, msRepo, mapRepo)

	err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
		MilestoneKey: ptrStr(pkg.FormatID(msKey)),
	})
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}
