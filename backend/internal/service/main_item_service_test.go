package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock repos for MainItemService tests
// ---------------------------------------------------------------------------

type mockMainItemRepo struct {
	item        *model.MainItem
	items       []model.MainItem
	pageResult  *dto.PageResult[model.MainItem]
	nextCodeVal string
	// per-operation errors
	findErr   error
	createErr error
	updateErr error
	listErr   error
	nextErr   error
	// capture calls
	createdItem *model.MainItem
	updatedID   uint
	updatedFields map[string]interface{}
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
	if s, ok := fields["status"]; ok {
		item.Status = s.(string)
	}
	if c, ok := fields["completion"]; ok {
		item.Completion = c.(float64)
	}
	return m.updateErr
}

func (m *mockMainItemRepo) List(_ context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.pageResult != nil {
		return m.pageResult, nil
	}
	return &dto.PageResult[model.MainItem]{Items: m.items, Total: int64(len(m.items))}, nil
}

func (m *mockMainItemRepo) NextCode(_ context.Context, teamID uint) (string, error) {
	return m.nextCodeVal, m.nextErr
}

func (m *mockMainItemRepo) CountByTeam(_ context.Context, _ uint) (int64, error) {
	return 0, nil
}

func (m *mockMainItemRepo) ListNonArchivedByTeam(_ context.Context, _ uint) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepo) ListByTeamAndStatus(_ context.Context, _ uint, _ string) ([]model.MainItem, error) {
	return nil, nil
}

type mockSubItemRepo struct {
	subItems []*model.SubItem
	findErr  error
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

func (m *mockSubItemRepo) List(_ context.Context, teamID uint, mainItemID uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}

func (m *mockSubItemRepo) ListByMainItem(_ context.Context, mainItemID uint) ([]*model.SubItem, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.subItems, nil
}

func (m *mockSubItemRepo) ListByTeam(_ context.Context, _ uint) ([]model.SubItem, error) {
	return nil, nil
}

func (m *mockSubItemRepo) Delete(_ context.Context, _ uint) error {
	return nil
}
func (m *mockSubItemRepo) NextSubCode(_ context.Context, _ uint) (string, error) {
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

	item, err := svc.Create(context.Background(), 1, 10, dto.MainItemCreateReq{
		Title:    "Feature A",
		Priority: "P0",
	})
	require.NoError(t, err)
	assert.Equal(t, "MI-0001", item.Code)
	assert.Equal(t, uint(10), item.ProposerID)
	assert.Equal(t, "pending", item.Status)
	assert.Equal(t, uint(1), item.TeamID)
	assert.Equal(t, "Feature A", item.Title)
	assert.Equal(t, "P0", item.Priority)
}

func TestMainItemCreate_NextCodeError(t *testing.T) {
	mainRepo := &mockMainItemRepo{nextErr: errors.New("db error")}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.Create(context.Background(), 1, 10, dto.MainItemCreateReq{Title: "Feature A"})
	assert.Error(t, err)
}

func TestMainItemCreate_RepoCreateError(t *testing.T) {
	mainRepo := &mockMainItemRepo{nextCodeVal: "MI-0001", createErr: errors.New("db error")}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.Create(context.Background(), 1, 10, dto.MainItemCreateReq{Title: "Feature A"})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestMainItemUpdate_Success(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		Title:  "Old Title",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
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
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 2, // different team
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestMainItemUpdate_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Update(context.Background(), 1, 99, dto.MainItemUpdateReq{
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
		TeamID: 1,
		Status: "completed",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.NotNil(t, mainRepo.updatedFields["archived_at"])
}

func TestMainItemArchive_ClosedStatus(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		Status: "closed",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), 1, 1)
	require.NoError(t, err)
}

func TestMainItemArchive_NotAllowed_InProgress(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		Status: "in_progress",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), 1, 1)
	assert.ErrorIs(t, err, apperrors.ErrArchiveNotAllowed)
}

func TestMainItemArchive_NotAllowed_Pending(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		Status: "pending",
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), 1, 1)
	assert.ErrorIs(t, err, apperrors.ErrArchiveNotAllowed)
}

func TestMainItemArchive_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.Archive(context.Background(), 1, 99)
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

	result, err := svc.List(context.Background(), 1, dto.MainItemFilter{}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)
}

func TestMainItemList_RepoError(t *testing.T) {
	mainRepo := &mockMainItemRepo{listErr: errors.New("db error")}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.List(context.Background(), 1, dto.MainItemFilter{}, dto.Pagination{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestMainItemGet_Success(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		Title:  "Item 1",
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
// Tests: RecalcCompletion
// ---------------------------------------------------------------------------

func TestRecalcCompletion_ZeroSubItems(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		Completion: 50,
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.RecalcCompletion(context.Background(), 1)
	require.NoError(t, err)
	assert.InDelta(t, float64(0), mainRepo.updatedFields["completion"], 0.001)
}

func TestRecalcCompletion_OneSubItem(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		Completion: 0,
	}
	mainRepo := &mockMainItemRepo{item: existing}
	subRepo := &mockSubItemRepo{
		subItems: []*model.SubItem{
			{Completion: 60, Weight: 1.0},
		},
	}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	err := svc.RecalcCompletion(context.Background(), 1)
	require.NoError(t, err)
	assert.InDelta(t, float64(60), mainRepo.updatedFields["completion"], 0.001)
}

func TestRecalcCompletion_MultipleSubItems_EqualWeights(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		Completion: 0,
	}
	mainRepo := &mockMainItemRepo{item: existing}
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
	assert.InDelta(t, float64(60), mainRepo.updatedFields["completion"], 0.001)
}

func TestRecalcCompletion_AllZeroWeights_FallbackSimpleAvg(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		Completion: 0,
	}
	mainRepo := &mockMainItemRepo{item: existing}
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
	assert.InDelta(t, float64(65), mainRepo.updatedFields["completion"], 0.001)
}

func TestRecalcCompletion_VaryingWeights(t *testing.T) {
	existing := &model.MainItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		Completion: 0,
	}
	mainRepo := &mockMainItemRepo{item: existing}
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
	assert.InDelta(t, 87.5, mainRepo.updatedFields["completion"], 0.001)
}

func TestRecalcCompletion_ItemNotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
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
				TeamID:      1,
				Status:      tt.from,
				ProposerID:  10, // PM
			}
			mainRepo := &mockMainItemRepo{item: item}
			historySvc := &mockStatusHistorySvc{}
			svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, historySvc)

			updated, err := svc.ChangeStatus(context.Background(), 1, 10, 1, tt.to)
			require.NoError(t, err)
			assert.Equal(t, tt.to, mainRepo.updatedFields["status"])
			assert.Equal(t, tt.to, updated.Status)

			// Verify status history recorded
			assert.NotNil(t, historySvc.recorded)
			assert.Equal(t, "main_item", historySvc.recorded.ItemType)
			assert.Equal(t, tt.from, historySvc.recorded.FromStatus)
			assert.Equal(t, tt.to, historySvc.recorded.ToStatus)
			assert.Equal(t, uint(10), historySvc.recorded.ChangedBy)
			assert.False(t, historySvc.recorded.IsAuto)
		})
	}
}

func TestChangeStatus_SelfTransition(t *testing.T) {
	item := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID:     1,
		Status:     "pending",
		ProposerID: 10,
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "pending")
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
				TeamID:      1,
				Status:      tt.from,
				ProposerID:  10,
			}
			mainRepo := &mockMainItemRepo{item: item}
			svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

			_, err := svc.ChangeStatus(context.Background(), 1, 10, 1, tt.to)
			assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
		})
	}
}

func TestChangeStatus_PMOnly_ReviewingToCompleted(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamID:      1,
		Status:      "reviewing",
		ProposerID:  10, // PM is user 10
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	// Non-PM caller should be forbidden
	_, err := svc.ChangeStatus(context.Background(), 1, 99, 1, "completed")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestChangeStatus_PMOnly_ReviewingToProgressing(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamID:      1,
		Status:      "reviewing",
		ProposerID:  10,
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	// Non-PM caller should be forbidden
	_, err := svc.ChangeStatus(context.Background(), 1, 99, 1, "progressing")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)

	// PM caller should succeed
	_, err = svc.ChangeStatus(context.Background(), 1, 10, 1, "progressing")
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
				TeamID:      1,
				Status:      tt.fromStatus,
				ProposerID:  10,
				Completion:  50,
			}
			mainRepo := &mockMainItemRepo{item: item}
			svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

			_, err := svc.ChangeStatus(context.Background(), 1, 10, 1, tt.newStatus)
			require.NoError(t, err)

			assert.Equal(t, float64(100), mainRepo.updatedFields["completion"])
			assert.NotNil(t, mainRepo.updatedFields["actual_end_date"])
		})
	}
}

func TestChangeStatus_NonTerminal_NoSideEffects(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamID:      1,
		Status:      "pending",
		ProposerID:  10,
		Completion:  30,
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "progressing")
	require.NoError(t, err)

	assert.Equal(t, "progressing", mainRepo.updatedFields["status"])
	_, hasCompletion := mainRepo.updatedFields["completion"]
	assert.False(t, hasCompletion, "non-terminal transition should not set completion")
	_, hasEndDate := mainRepo.updatedFields["actual_end_date"]
	assert.False(t, hasEndDate, "non-terminal transition should not set actual_end_date")
}

func TestChangeStatus_ItemNotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, 999, "progressing")
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestMainItemChangeStatus_TeamMismatch(t *testing.T) {
	item := &model.MainItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID:     2,
		Status:     "pending",
		ProposerID: 10,
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "progressing")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestChangeStatus_StatusHistoryRecorded(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamID:      1,
		Status:      "pending",
		ProposerID:  10,
	}
	mainRepo := &mockMainItemRepo{item: item}
	historySvc := &mockStatusHistorySvc{}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, historySvc)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "progressing")
	require.NoError(t, err)

	require.NotNil(t, historySvc.recorded)
	assert.Equal(t, "main_item", historySvc.recorded.ItemType)
	assert.Equal(t, uint(1), historySvc.recorded.ItemID)
	assert.Equal(t, "pending", historySvc.recorded.FromStatus)
	assert.Equal(t, "progressing", historySvc.recorded.ToStatus)
	assert.Equal(t, uint(10), historySvc.recorded.ChangedBy)
	assert.False(t, historySvc.recorded.IsAuto)
}

// ---------------------------------------------------------------------------
// Tests: AvailableTransitions
// ---------------------------------------------------------------------------

func TestAvailableTransitions_Success(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		callerID   uint
		proposerID uint
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
				TeamID:      1,
				Status:      tt.status,
				ProposerID:  tt.proposerID,
			}
			mainRepo := &mockMainItemRepo{item: item}
			svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

			transitions, err := svc.AvailableTransitions(context.Background(), 1, tt.callerID, 1)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, transitions)
		})
	}
}

func TestAvailableTransitions_NonPMReviewing_FiltersCompletedProgressing(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamID:      1,
		Status:      "reviewing",
		ProposerID:  10, // PM is user 10
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	// Non-PM caller should not see completed/progressing
	transitions, err := svc.AvailableTransitions(context.Background(), 1, 99, 1)
	require.NoError(t, err)
	assert.Empty(t, transitions, "non-PM should see no transitions from reviewing since all require PM")
}

func TestAvailableTransitions_ItemNotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.AvailableTransitions(context.Background(), 1, 10, 999)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestAvailableTransitions_TeamMismatch(t *testing.T) {
	item := &model.MainItem{
		BaseModel:   model.BaseModel{ID: 1},
		TeamID:      2,
		Status:      "pending",
		ProposerID:  10,
	}
	mainRepo := &mockMainItemRepo{item: item}
	svc := NewMainItemService(mainRepo, &mockSubItemRepo{}, nil)

	_, err := svc.AvailableTransitions(context.Background(), 1, 10, 1)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: EvaluateLinkage
// ---------------------------------------------------------------------------

func TestEvaluateLinkage_NoSubItems_NoLinkageTriggered(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1},
		TeamID:    1,
		Status:    "pending",
	}
	mainRepo := &mockMainItemRepo{item: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEvaluateLinkage_MainItemNotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.EvaluateLinkage(context.Background(), 999, 10)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestEvaluateLinkage_SubItemRepoError(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1},
		Status:    "pending",
	}
	mainRepo := &mockMainItemRepo{item: mainItem}
	subRepo := &mockSubItemRepo{findErr: errors.New("db error")}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	_, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	assert.Error(t, err)
}

// TestEvaluateLinkage_Priority1_AllCompletedOrClosed tests Priority 1:
// all completed/closed + at least one completed -> reviewing
func TestEvaluateLinkage_Priority1_AllCompletedOrClosed(t *testing.T) {
	tests := []struct {
		name   string
		items  []*model.SubItem
	}{
		{
			"all completed -> reviewing",
			[]*model.SubItem{
				{Status: "completed"},
				{Status: "completed"},
			},
		},
		{
			"mixed completed+closed -> reviewing",
			[]*model.SubItem{
				{Status: "completed"},
				{Status: "closed"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainItem := &model.MainItem{
				BaseModel: model.BaseModel{ID: 1},
				Status:    "progressing",
			}
			mainRepo := &mockMainItemRepo{item: mainItem}
			subRepo := &mockSubItemRepo{subItems: tt.items}
			historySvc := &mockStatusHistorySvc{}
			svc := NewMainItemService(mainRepo, subRepo, historySvc)

			result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.Triggered)
			assert.True(t, result.Success)
			assert.Equal(t, "reviewing", result.TargetStatus)
			assert.Equal(t, "reviewing", mainRepo.updatedFields["status"])

			// StatusHistory should be recorded with is_auto=true
			assert.NotNil(t, historySvc.recorded)
			assert.True(t, historySvc.recorded.IsAuto)
		})
	}
}

// TestEvaluateLinkage_Priority2_AllClosed tests Priority 2:
// all closed -> closed
func TestEvaluateLinkage_Priority2_AllClosed(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1},
		Status:    "pending",
	}
	mainRepo := &mockMainItemRepo{item: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{Status: "closed"},
		{Status: "closed"},
	}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "closed", result.TargetStatus)
	// Terminal side effects
	assert.Equal(t, float64(100), mainRepo.updatedFields["completion"])
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
				{Status: "pausing"},
				{Status: "pausing"},
			},
		},
		{
			"pausing + closed -> pausing",
			[]*model.SubItem{
				{Status: "pausing"},
				{Status: "closed"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainItem := &model.MainItem{
				BaseModel: model.BaseModel{ID: 1},
				Status:    "progressing",
			}
			mainRepo := &mockMainItemRepo{item: mainItem}
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
		name         string
		mainStatus   string
		items        []*model.SubItem
		wantTarget   string // expected target status, empty means no linkage
		wantSuccess  bool   // whether linkage should succeed
	}{
		{
			"progressing + blocking sub -> blocking (success)",
			"progressing",
			[]*model.SubItem{{Status: "blocking"}, {Status: "pending"}},
			"blocking",
			true,
		},
		{
			"pending + blocking sub -> blocking (fails: pending->blocking not valid)",
			"pending",
			[]*model.SubItem{{Status: "blocking"}, {Status: "pending"}},
			"blocking",
			false, // pending->blocking is not a valid MainItem transition
		},
		{
			"reviewing + blocking sub -> progressing (via AC-9 revert)",
			"reviewing",
			[]*model.SubItem{{Status: "blocking"}, {Status: "pending"}},
			"progressing",
			true, // reviewing->progressing is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainItem := &model.MainItem{
				BaseModel: model.BaseModel{ID: 1},
				Status:    tt.mainStatus,
			}
			mainRepo := &mockMainItemRepo{item: mainItem}
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
		name             string
		mainStatus       string
		items            []*model.SubItem
		wantProgressing  bool
	}{
		{
			"pending + progressing sub -> progressing",
			"pending",
			[]*model.SubItem{{Status: "progressing"}, {Status: "pending"}},
			true,
		},
		{
			"progressing main + progressing sub -> no change (same status)",
			"progressing",
			[]*model.SubItem{{Status: "progressing"}, {Status: "pending"}},
			false,
		},
		{
			"blocking main + progressing sub -> no linkage",
			"blocking",
			[]*model.SubItem{{Status: "progressing"}, {Status: "pending"}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainItem := &model.MainItem{
				BaseModel: model.BaseModel{ID: 1},
				Status:    tt.mainStatus,
			}
			mainRepo := &mockMainItemRepo{item: mainItem}
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
		BaseModel: model.BaseModel{ID: 1},
		Status:    "reviewing",
	}
	mainRepo := &mockMainItemRepo{item: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{Status: "completed"},
		{Status: "pending"},
	}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "progressing", result.TargetStatus)
	assert.Equal(t, "progressing", mainRepo.updatedFields["status"])
}

// TestEvaluateLinkage_Failure_TransitionNotAllowed tests AC-12:
// linkage failure when transition not allowed
func TestEvaluateLinkage_Failure_TransitionNotAllowed(t *testing.T) {
	// When all sub-items are completed/closed, target is "reviewing"
	// But if main is in a state that can't transition to "reviewing",
	// linkage should fail and record intent in status history.
	mainItem := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1},
		Status:    "blocking", // blocking -> reviewing is not valid
	}
	mainRepo := &mockMainItemRepo{item: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{Status: "completed"},
		{Status: "completed"},
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
	assert.True(t, historySvc.recorded.IsAuto)
	assert.Equal(t, "blocking", historySvc.recorded.FromStatus)
	assert.Equal(t, "reviewing", historySvc.recorded.ToStatus)
	assert.Contains(t, historySvc.recorded.Remark, "不允许")
}

// TestEvaluateLinkage_SameStatus_NoTransition tests that no linkage is triggered
// when the target status matches current status.
func TestEvaluateLinkage_SameStatus_NoTransition(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1},
		Status:    "reviewing", // All completed would target reviewing -> same status
	}
	mainRepo := &mockMainItemRepo{item: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{Status: "completed"},
		{Status: "completed"},
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
		BaseModel: model.BaseModel{ID: 1},
		Status:    "pending",
		Completion: 30,
	}
	mainRepo := &mockMainItemRepo{item: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{Status: "closed"},
		{Status: "closed"},
	}}
	svc := NewMainItemService(mainRepo, subRepo, nil)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "closed", result.TargetStatus)
	assert.Equal(t, float64(100), mainRepo.updatedFields["completion"])
	assert.NotNil(t, mainRepo.updatedFields["actual_end_date"])
}

// TestEvaluateLinkage_StatusHistoryIsAuto tests AC-15:
// is_auto=true for linkage transitions.
func TestEvaluateLinkage_StatusHistoryIsAuto(t *testing.T) {
	mainItem := &model.MainItem{
		BaseModel: model.BaseModel{ID: 1},
		Status:    "progressing",
	}
	mainRepo := &mockMainItemRepo{item: mainItem}
	subRepo := &mockSubItemRepo{subItems: []*model.SubItem{
		{Status: "completed"},
	}}
	historySvc := &mockStatusHistorySvc{}
	svc := NewMainItemService(mainRepo, subRepo, historySvc)

	result, err := svc.EvaluateLinkage(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)

	require.NotNil(t, historySvc.recorded)
	assert.True(t, historySvc.recorded.IsAuto, "linkage status history should have is_auto=true")
	assert.Equal(t, uint(10), historySvc.recorded.ChangedBy)
	assert.Equal(t, "progressing", historySvc.recorded.FromStatus)
	assert.Equal(t, "reviewing", historySvc.recorded.ToStatus)
}

// TestLinkageResult_Warning tests the Warning() method.
func TestLinkageResult_Warning(t *testing.T) {
	tests := []struct {
		name    string
		result  *LinkageResult
		want    string
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
