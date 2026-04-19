package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// ---------------------------------------------------------------------------
// Mock repos for ProgressService tests
// ---------------------------------------------------------------------------

type mockProgressRepo struct {
	latest            *model.ProgressRecord
	latestErr         error
	created           *model.ProgressRecord
	createErr         error
	records           []model.ProgressRecord
	listErr           error
	updatedID         uint
	updatedCompletion float64
	updateErr         error
}

func (m *mockProgressRepo) Create(_ context.Context, record *model.ProgressRecord) error {
	m.created = record
	if m.createErr != nil {
		return m.createErr
	}
	record.ID = 1
	return nil
}

func (m *mockProgressRepo) FindByID(_ context.Context, id uint) (*model.ProgressRecord, error) {
	for _, r := range m.records {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockProgressRepo) ListBySubItem(_ context.Context, teamID uint, subItemID uint) ([]model.ProgressRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.records, nil
}

func (m *mockProgressRepo) LatestBySubItem(_ context.Context, subItemID uint) (*model.ProgressRecord, error) {
	if m.latestErr != nil {
		return nil, m.latestErr
	}
	return m.latest, nil
}

func (m *mockProgressRepo) UpdateCompletion(_ context.Context, recordID uint, completion float64) error {
	m.updatedID = recordID
	m.updatedCompletion = completion
	return m.updateErr
}

func (m *mockProgressRepo) ListByTeamInRange(_ context.Context, _ uint, _, _ time.Time) ([]model.ProgressRecord, error) {
	return nil, nil
}

// compile-time checks
var _ repository.ProgressRepo = (*mockProgressRepo)(nil)

// mockSubItemRepoForProgress is a minimal SubItemRepo mock for progress service tests.
type mockSubItemRepoForProgress struct {
	item          *model.SubItem
	findErr       error
	updatedID     uint
	updatedFields map[string]interface{}
	updateErr     error
}

func (m *mockSubItemRepoForProgress) Create(_ context.Context, item *model.SubItem) error {
	return nil
}

func (m *mockSubItemRepoForProgress) FindByID(_ context.Context, id uint) (*model.SubItem, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.item, nil
}

func (m *mockSubItemRepoForProgress) Update(_ context.Context, item *model.SubItem, fields map[string]interface{}) error {
	m.updatedID = item.ID
	m.updatedFields = fields
	return m.updateErr
}

func (m *mockSubItemRepoForProgress) List(_ context.Context, teamID uint, mainItemID uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}

func (m *mockSubItemRepoForProgress) ListByMainItem(_ context.Context, mainItemID uint) ([]*model.SubItem, error) {
	return nil, nil
}

func (m *mockSubItemRepoForProgress) ListByTeam(_ context.Context, _ uint) ([]model.SubItem, error) {
	return nil, nil
}

var _ repository.SubItemRepo = (*mockSubItemRepoForProgress)(nil)

// mockMainItemSvcForProgress captures RecalcCompletion calls.
type mockMainItemSvcForProgress struct {
	recalcCalled bool
	recalcID     uint
	recalcErr    error
}

func (m *mockMainItemSvcForProgress) Create(_ context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error) {
	return nil, nil
}

func (m *mockMainItemSvcForProgress) Update(_ context.Context, teamID, itemID uint, req dto.MainItemUpdateReq) error {
	return nil
}

func (m *mockMainItemSvcForProgress) Archive(_ context.Context, teamID, itemID uint) error {
	return nil
}

func (m *mockMainItemSvcForProgress) List(_ context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}

func (m *mockMainItemSvcForProgress) Get(_ context.Context, itemID uint) (*model.MainItem, error) {
	return nil, nil
}

func (m *mockMainItemSvcForProgress) RecalcCompletion(_ context.Context, mainItemID uint) error {
	m.recalcCalled = true
	m.recalcID = mainItemID
	return m.recalcErr
}

var _ MainItemService = (*mockMainItemSvcForProgress)(nil)

// ---------------------------------------------------------------------------
// Tests: Append
// ---------------------------------------------------------------------------

func TestProgressAppend_FirstRecord_NoRegression(t *testing.T) {
	progressRepo := &mockProgressRepo{latest: nil}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 0},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	record, err := svc.Append(context.Background(), 1, 2, 5, 30.0, "achievement", "blocker", "lesson", false)
	require.NoError(t, err)
	assert.Equal(t, uint(5), record.SubItemID)
	assert.Equal(t, float64(30.0), record.Completion)
	assert.Equal(t, uint(2), record.AuthorID)
	assert.False(t, record.IsPMCorrect)
}

func TestProgressAppend_RegressionDetected(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 50.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 30.0, "", "", "", false)
	assert.ErrorIs(t, err, apperrors.ErrProgressRegression)
}

func TestProgressAppend_EqualCompletion_NoRegression(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 50.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	record, err := svc.Append(context.Background(), 1, 2, 5, 50.0, "", "", "", false)
	require.NoError(t, err)
	assert.Equal(t, float64(50.0), record.Completion)
}

func TestProgressAppend_HigherCompletion_Passes(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 50.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	record, err := svc.Append(context.Background(), 1, 2, 5, 75.0, "", "", "", false)
	require.NoError(t, err)
	assert.Equal(t, float64(75.0), record.Completion)
}

func TestProgressAppend_PMCanBypassRegression(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 80.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 80},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	record, err := svc.Append(context.Background(), 1, 2, 5, 30.0, "", "", "", true)
	require.NoError(t, err)
	assert.Equal(t, float64(30.0), record.Completion)
}

func TestProgressAppend_UpdatesSubItemCompletion(t *testing.T) {
	progressRepo := &mockProgressRepo{}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 0},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	require.NoError(t, err)
	assert.Equal(t, uint(5), subItemRepo.updatedID)
	assert.InDelta(t, 60.0, subItemRepo.updatedFields["completion"], 0.001)
}

func TestProgressAppend_TriggersRecalcCompletion(t *testing.T) {
	progressRepo := &mockProgressRepo{}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 0},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	require.NoError(t, err)
	assert.True(t, mainItemSvc.recalcCalled)
	assert.Equal(t, uint(10), mainItemSvc.recalcID)
}

func TestProgressAppend_SubItemNotFound(t *testing.T) {
	progressRepo := &mockProgressRepo{}
	subItemRepo := &mockSubItemRepoForProgress{findErr: apperrors.ErrNotFound}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	assert.Error(t, err)
}

func TestProgressAppend_LatestBySubItemError(t *testing.T) {
	progressRepo := &mockProgressRepo{latestErr: errors.New("db error")}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	assert.Error(t, err)
}

func TestProgressAppend_CreateError(t *testing.T) {
	progressRepo := &mockProgressRepo{createErr: errors.New("db error")}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: CorrectCompletion
// ---------------------------------------------------------------------------

func TestProgressCorrectCompletion_Success(t *testing.T) {
	record := &model.ProgressRecord{
		ID:         100,
		SubItemID:  5,
		TeamID:     1,
		AuthorID:   2,
		Completion: 50.0,
		CreatedAt:  time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records: []model.ProgressRecord{*record},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	require.NoError(t, err)
	assert.Equal(t, uint(100), progressRepo.updatedID)
	assert.InDelta(t, 80.0, progressRepo.updatedCompletion, 0.001)
}

func TestProgressCorrectCompletion_IsLatest_SyncsSubItem(t *testing.T) {
	record := &model.ProgressRecord{
		ID:         100,
		SubItemID:  5,
		TeamID:     1,
		Completion: 50.0,
		CreatedAt:  time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records: []model.ProgressRecord{*record},
		latest:  record,
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	require.NoError(t, err)
	assert.Equal(t, uint(5), subItemRepo.updatedID)
	assert.InDelta(t, 80.0, subItemRepo.updatedFields["completion"], 0.001)
}

func TestProgressCorrectCompletion_NotLatest_SyncsToLatest(t *testing.T) {
	record := &model.ProgressRecord{
		ID:         100,
		SubItemID:  5,
		TeamID:     1,
		Completion: 50.0,
		CreatedAt:  time.Now().Add(-2 * time.Hour),
	}
	latestRecord := &model.ProgressRecord{
		ID:         200,
		SubItemID:  5,
		Completion: 90.0,
		CreatedAt:  time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records: []model.ProgressRecord{*record, *latestRecord},
		latest:  latestRecord,
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 90},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	require.NoError(t, err)
	// SubItem should be synced to latest record's completion (90), not corrected (80)
	assert.Equal(t, uint(5), subItemRepo.updatedID)
	assert.InDelta(t, 90.0, subItemRepo.updatedFields["completion"], 0.001)
}

func TestProgressCorrectCompletion_TriggersRecalc(t *testing.T) {
	record := &model.ProgressRecord{
		ID:         100,
		SubItemID:  5,
		TeamID:     1,
		Completion: 50.0,
		CreatedAt:  time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records: []model.ProgressRecord{*record},
		latest:  record,
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10, Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	require.NoError(t, err)
	assert.True(t, mainItemSvc.recalcCalled)
	assert.Equal(t, uint(10), mainItemSvc.recalcID)
}

func TestProgressCorrectCompletion_RecordNotFound(t *testing.T) {
	progressRepo := &mockProgressRepo{records: []model.ProgressRecord{}}
	subItemRepo := &mockSubItemRepoForProgress{}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	err := svc.CorrectCompletion(context.Background(), 1, 999, 80.0)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestProgressCorrectCompletion_UpdateError(t *testing.T) {
	record := &model.ProgressRecord{
		ID:        100,
		SubItemID: 5,
		TeamID:    1,
		CreatedAt: time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records:   []model.ProgressRecord{*record},
		updateErr: errors.New("db error"),
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5}, MainItemID: 10},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestProgressList_Success(t *testing.T) {
	records := []model.ProgressRecord{
		{ID: 1, SubItemID: 5, Completion: 30.0, CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: 2, SubItemID: 5, Completion: 60.0, CreatedAt: time.Now().Add(-1 * time.Hour)},
		{ID: 3, SubItemID: 5, Completion: 90.0, CreatedAt: time.Now()},
	}
	progressRepo := &mockProgressRepo{records: records}
	subItemRepo := &mockSubItemRepoForProgress{}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	result, err := svc.List(context.Background(), 1, 5)
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, float64(30.0), result[0].Completion)
	assert.Equal(t, float64(60.0), result[1].Completion)
	assert.Equal(t, float64(90.0), result[2].Completion)
}

func TestProgressList_Empty(t *testing.T) {
	progressRepo := &mockProgressRepo{records: []model.ProgressRecord{}}
	subItemRepo := &mockSubItemRepoForProgress{}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	result, err := svc.List(context.Background(), 1, 5)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestProgressList_RepoError(t *testing.T) {
	progressRepo := &mockProgressRepo{listErr: errors.New("db error")}
	subItemRepo := &mockSubItemRepoForProgress{}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc)

	_, err := svc.List(context.Background(), 1, 5)
	assert.Error(t, err)
}
