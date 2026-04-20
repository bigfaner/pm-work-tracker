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

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestMainItemCreate_Success(t *testing.T) {
	mainRepo := &mockMainItemRepo{nextCodeVal: "MI-0001"}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

	_, err := svc.Create(context.Background(), 1, 10, dto.MainItemCreateReq{Title: "Feature A"})
	assert.Error(t, err)
}

func TestMainItemCreate_RepoCreateError(t *testing.T) {
	mainRepo := &mockMainItemRepo{nextCodeVal: "MI-0001", createErr: errors.New("db error")}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

	err := svc.Update(context.Background(), 1, 1, dto.MainItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestMainItemUpdate_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

	err := svc.Archive(context.Background(), 1, 1)
	assert.ErrorIs(t, err, apperrors.ErrArchiveNotAllowed)
}

func TestMainItemArchive_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

	result, err := svc.List(context.Background(), 1, dto.MainItemFilter{}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)
}

func TestMainItemList_RepoError(t *testing.T) {
	mainRepo := &mockMainItemRepo{listErr: errors.New("db error")}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

	item, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Item 1", item.Title)
}

func TestMainItemGet_NotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

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
	svc := NewMainItemService(mainRepo, subRepo)

	err := svc.RecalcCompletion(context.Background(), 1)
	require.NoError(t, err)
	// Weighted: (100*3 + 50*1) / (3+1) = 350/4 = 87.5
	assert.InDelta(t, 87.5, mainRepo.updatedFields["completion"], 0.001)
}

func TestRecalcCompletion_ItemNotFound(t *testing.T) {
	mainRepo := &mockMainItemRepo{findErr: gorm.ErrRecordNotFound}
	subRepo := &mockSubItemRepo{}
	svc := NewMainItemService(mainRepo, subRepo)

	err := svc.RecalcCompletion(context.Background(), 99)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func ptrStr(s string) *string {
	return &s
}
