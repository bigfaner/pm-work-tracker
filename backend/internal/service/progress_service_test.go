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

func (m *mockProgressRepo) ListBySubItem(_ context.Context, teamID uint, subItemBizKey int64) ([]model.ProgressRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.records, nil
}

func (m *mockProgressRepo) LatestBySubItem(_ context.Context, subItemBizKey int64) (*model.ProgressRecord, error) {
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

func (m *mockProgressRepo) FindByBizKey(_ context.Context, bizKey int64) (*model.ProgressRecord, error) {
	for _, r := range m.records {
		if r.BizKey == bizKey {
			return &r, nil
		}
	}
	return nil, apperrors.ErrNotFound
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

func (m *mockSubItemRepoForProgress) SoftDelete(_ context.Context, _ uint) error {
	return nil
}
func (m *mockSubItemRepoForProgress) FindByBizKey(_ context.Context, _ int64) (*model.SubItem, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.item, nil
}
func (m *mockSubItemRepoForProgress) NextSubCode(_ context.Context, _ uint) (string, error) {
	return "", nil
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

func (m *mockMainItemSvcForProgress) ChangeStatus(_ context.Context, _, _, _ uint, _ string) (*model.MainItem, error) {
	return nil, nil
}

func (m *mockMainItemSvcForProgress) AvailableTransitions(_ context.Context, _, _, _ uint) ([]string, error) {
	return nil, nil
}

func (m *mockMainItemSvcForProgress) EvaluateLinkage(_ context.Context, _ uint, _ uint) (*LinkageResult, error) {
	return nil, nil
}

func (m *mockMainItemSvcForProgress) GetByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}

var _ MainItemService = (*mockMainItemSvcForProgress)(nil)

// mockStatusHistorySvcForProgress captures Record calls.
type mockStatusHistorySvcForProgress struct {
	recorded []*model.StatusHistory
	recordErr error
}

func (m *mockStatusHistorySvcForProgress) Record(_ context.Context, record *model.StatusHistory) error {
	m.recorded = append(m.recorded, record)
	return m.recordErr
}

func (m *mockStatusHistorySvcForProgress) ListByItem(_ context.Context, _ string, _ uint, _ dto.Pagination) (*dto.PageResult[model.StatusHistory], error) {
	return nil, nil
}

var _ StatusHistoryService = (*mockStatusHistorySvcForProgress)(nil)

// ---------------------------------------------------------------------------
// Tests: Append
// ---------------------------------------------------------------------------

func TestProgressAppend_FirstRecord_NoRegression(t *testing.T) {
	progressRepo := &mockProgressRepo{latest: nil}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 0},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	const teamBizKey int64 = 123456789012345678
	record, err := svc.Append(context.Background(), teamBizKey, 2, 5, 30.0, "achievement", "blocker", "lesson", false)
	require.NoError(t, err)
	assert.Equal(t, int64(5), record.SubItemKey)
	assert.Equal(t, float64(30.0), record.Completion)
	assert.Equal(t, int64(2), record.AuthorKey)
	assert.Equal(t, 0, record.IsPmCorrect)
	// TeamKey must store the snowflake bizKey, not the internal auto-increment ID
	assert.Equal(t, teamBizKey, record.TeamKey)
}

func TestProgressAppend_RegressionDetected(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 50.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	_, err := svc.Append(context.Background(), 1, 2, 5, 30.0, "", "", "", false)
	assert.ErrorIs(t, err, apperrors.ErrProgressRegression)
}

func TestProgressAppend_EqualCompletion_NoRegression(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 50.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	record, err := svc.Append(context.Background(), 1, 2, 5, 50.0, "", "", "", false)
	require.NoError(t, err)
	assert.Equal(t, float64(50.0), record.Completion)
}

func TestProgressAppend_HigherCompletion_Passes(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 50.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	record, err := svc.Append(context.Background(), 1, 2, 5, 75.0, "", "", "", false)
	require.NoError(t, err)
	assert.Equal(t, float64(75.0), record.Completion)
}

func TestProgressAppend_PMCanBypassRegression(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 80.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 80},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	record, err := svc.Append(context.Background(), 1, 2, 5, 30.0, "", "", "", true)
	require.NoError(t, err)
	assert.Equal(t, float64(30.0), record.Completion)
}

func TestProgressAppend_UpdatesSubItemCompletion(t *testing.T) {
	progressRepo := &mockProgressRepo{}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 0},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	require.NoError(t, err)
	assert.Equal(t, uint(5), subItemRepo.updatedID)
	assert.InDelta(t, 60.0, subItemRepo.updatedFields["completion_pct"], 0.001)
}

func TestProgressAppend_TriggersRecalcCompletion(t *testing.T) {
	progressRepo := &mockProgressRepo{}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 0},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	require.NoError(t, err)
	assert.True(t, mainItemSvc.recalcCalled)
	assert.Equal(t, uint(10), mainItemSvc.recalcID)
}

func TestProgressAppend_SubItemNotFound(t *testing.T) {
	progressRepo := &mockProgressRepo{}
	subItemRepo := &mockSubItemRepoForProgress{findErr: apperrors.ErrNotFound}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	assert.Error(t, err)
}

func TestProgressAppend_LatestBySubItemError(t *testing.T) {
	progressRepo := &mockProgressRepo{latestErr: errors.New("db error")}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10)},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	assert.Error(t, err)
}

func TestProgressAppend_CreateError(t *testing.T) {
	progressRepo := &mockProgressRepo{createErr: errors.New("db error")}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10)},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: CorrectCompletion
// ---------------------------------------------------------------------------

func TestProgressCorrectCompletion_Success(t *testing.T) {
	record := &model.ProgressRecord{
		ID:         100,
		SubItemKey:  5,
		TeamKey: 1,
		AuthorKey:   2,
		Completion: 50.0,
		CreateTime:  time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records: []model.ProgressRecord{*record},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	require.NoError(t, err)
	assert.Equal(t, uint(100), progressRepo.updatedID)
	assert.InDelta(t, 80.0, progressRepo.updatedCompletion, 0.001)
}

func TestProgressCorrectCompletion_IsLatest_SyncsSubItem(t *testing.T) {
	record := &model.ProgressRecord{
		ID:         100,
		SubItemKey:  5,
		TeamKey: 1,
		Completion: 50.0,
		CreateTime:  time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records: []model.ProgressRecord{*record},
		latest:  record,
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	require.NoError(t, err)
	assert.Equal(t, uint(5), subItemRepo.updatedID)
	assert.InDelta(t, 80.0, subItemRepo.updatedFields["completion_pct"], 0.001)
}

func TestProgressCorrectCompletion_NotLatest_SyncsToLatest(t *testing.T) {
	record := &model.ProgressRecord{
		ID:         100,
		SubItemKey:  5,
		TeamKey: 1,
		Completion: 50.0,
		CreateTime:  time.Now().Add(-2 * time.Hour),
	}
	latestRecord := &model.ProgressRecord{
		ID:         200,
		SubItemKey:  5,
		Completion: 90.0,
		CreateTime:  time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records: []model.ProgressRecord{*record, *latestRecord},
		latest:  latestRecord,
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 90},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	require.NoError(t, err)
	// SubItem should be synced to latest record's completion (90), not corrected (80)
	assert.Equal(t, uint(5), subItemRepo.updatedID)
	assert.InDelta(t, 90.0, subItemRepo.updatedFields["completion_pct"], 0.001)
}

func TestProgressCorrectCompletion_TriggersRecalc(t *testing.T) {
	record := &model.ProgressRecord{
		ID:         100,
		SubItemKey:  5,
		TeamKey: 1,
		Completion: 50.0,
		CreateTime:  time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records: []model.ProgressRecord{*record},
		latest:  record,
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), Completion: 50},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	require.NoError(t, err)
	assert.True(t, mainItemSvc.recalcCalled)
	assert.Equal(t, uint(10), mainItemSvc.recalcID)
}

func TestProgressCorrectCompletion_RecordNotFound(t *testing.T) {
	progressRepo := &mockProgressRepo{records: []model.ProgressRecord{}}
	subItemRepo := &mockSubItemRepoForProgress{}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	err := svc.CorrectCompletion(context.Background(), 1, 999, 80.0)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestProgressCorrectCompletion_UpdateError(t *testing.T) {
	record := &model.ProgressRecord{
		ID:        100,
		SubItemKey: 5,
		TeamKey: 1,
		CreateTime: time.Now(),
	}
	progressRepo := &mockProgressRepo{
		records:   []model.ProgressRecord{*record},
		updateErr: errors.New("db error"),
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10)},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	err := svc.CorrectCompletion(context.Background(), 1, 100, 80.0)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestProgressList_Success(t *testing.T) {
	records := []model.ProgressRecord{
		{ID: 1, SubItemKey: 5, Completion: 30.0, CreateTime: time.Now().Add(-2 * time.Hour)},
		{ID: 2, SubItemKey: 5, Completion: 60.0, CreateTime: time.Now().Add(-1 * time.Hour)},
		{ID: 3, SubItemKey: 5, Completion: 90.0, CreateTime: time.Now()},
	}
	progressRepo := &mockProgressRepo{records: records}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10)},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	result, err := svc.List(context.Background(), 1, 5)
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, float64(30.0), result[0].Completion)
	assert.Equal(t, float64(60.0), result[1].Completion)
	assert.Equal(t, float64(90.0), result[2].Completion)
}

func TestProgressList_Empty(t *testing.T) {
	progressRepo := &mockProgressRepo{records: []model.ProgressRecord{}}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	result, err := svc.List(context.Background(), 1, 5)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestProgressList_RepoError(t *testing.T) {
	progressRepo := &mockProgressRepo{listErr: errors.New("db error")}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, &mockStatusHistorySvcForProgress{})

	_, err := svc.List(context.Background(), 1, 5)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Auto-status-transition on Append
// ---------------------------------------------------------------------------

func TestProgressAppend_Pending_FirstProgress_TransitionsToProgressing(t *testing.T) {
	progressRepo := &mockProgressRepo{latest: nil}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), ItemStatus: "pending", Completion: 0},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}
	historySvc := &mockStatusHistorySvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, historySvc)

	record, err := svc.Append(context.Background(), 1, 2, 5, 30.0, "did stuff", "", "", false)
	require.NoError(t, err)
	assert.Equal(t, float64(30.0), record.Completion)

	// Verify status transitioned to "progressing"
	assert.Equal(t, "progressing", subItemRepo.updatedFields["item_status"])
	assert.InDelta(t, 30.0, subItemRepo.updatedFields["completion_pct"], 0.001)

	// Verify status history was recorded
	require.Len(t, historySvc.recorded, 1)
	assert.Equal(t, "pending", historySvc.recorded[0].FromStatus)
	assert.Equal(t, "progressing", historySvc.recorded[0].ToStatus)
	assert.Equal(t, int64(2), historySvc.recorded[0].ChangedBy)
	assert.Equal(t, 1, historySvc.recorded[0].IsAuto)
}

func TestProgressAppend_Progressing_FirstProgressDoesNotApply(t *testing.T) {
	progressRepo := &mockProgressRepo{latest: nil}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), ItemStatus: "progressing", Completion: 0},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}
	historySvc := &mockStatusHistorySvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, historySvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 30.0, "", "", "", false)
	require.NoError(t, err)

	// No status change
	assert.Nil(t, subItemRepo.updatedFields["item_status"])
	assert.Empty(t, historySvc.recorded)
}

func TestProgressAppend_Pending_NotFirstProgress_NoAutoTransition(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 20.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), ItemStatus: "pending", Completion: 20},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}
	historySvc := &mockStatusHistorySvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, historySvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 40.0, "", "", "", false)
	require.NoError(t, err)

	// No status change (not first progress)
	assert.Nil(t, subItemRepo.updatedFields["item_status"])
	assert.Empty(t, historySvc.recorded)
}

func TestProgressAppend_Progressing_100Percent_TransitionsToCompleted(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 80.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), ItemStatus: "progressing", Completion: 80},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}
	historySvc := &mockStatusHistorySvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, historySvc)

	record, err := svc.Append(context.Background(), 1, 2, 5, 100.0, "all done", "", "", false)
	require.NoError(t, err)
	assert.Equal(t, float64(100.0), record.Completion)

	// Verify status transitioned to "completed"
	assert.Equal(t, "completed", subItemRepo.updatedFields["item_status"])
	assert.InDelta(t, 100.0, subItemRepo.updatedFields["completion_pct"], 0.001)
	assert.NotNil(t, subItemRepo.updatedFields["actual_end_date"])

	// Verify status history was recorded
	require.Len(t, historySvc.recorded, 1)
	assert.Equal(t, "progressing", historySvc.recorded[0].FromStatus)
	assert.Equal(t, "completed", historySvc.recorded[0].ToStatus)
	assert.Equal(t, 1, historySvc.recorded[0].IsAuto)

	// Verify recalc was triggered (needed for completed status)
	assert.True(t, mainItemSvc.recalcCalled)
}

func TestProgressAppend_Pending_FirstProgress_100Percent_TransitionsToCompleted(t *testing.T) {
	progressRepo := &mockProgressRepo{latest: nil}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), ItemStatus: "pending", Completion: 0},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}
	historySvc := &mockStatusHistorySvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, historySvc)

	record, err := svc.Append(context.Background(), 1, 2, 5, 100.0, "done immediately", "", "", false)
	require.NoError(t, err)
	assert.Equal(t, float64(100.0), record.Completion)

	// Both rules apply: pending->progressing (rule 1), then progressing->completed (rule 2)
	// End result: "completed"
	assert.Equal(t, "completed", subItemRepo.updatedFields["item_status"])
	assert.InDelta(t, 100.0, subItemRepo.updatedFields["completion_pct"], 0.001)
	assert.NotNil(t, subItemRepo.updatedFields["actual_end_date"])

	// Status history should record the overall transition from pending to completed
	require.Len(t, historySvc.recorded, 1)
	assert.Equal(t, "pending", historySvc.recorded[0].FromStatus)
	assert.Equal(t, "completed", historySvc.recorded[0].ToStatus)
}

func TestProgressAppend_100Percent_Blocked_NoTransition(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 80.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), ItemStatus: "blocking", Completion: 80},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}
	historySvc := &mockStatusHistorySvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, historySvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 100.0, "", "", "", false)
	require.NoError(t, err)

	// "blocking" -> "completed" is NOT a valid transition
	assert.Nil(t, subItemRepo.updatedFields["item_status"])
	assert.Empty(t, historySvc.recorded)
}

func TestProgressAppend_100Percent_Pausing_NoTransition(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 80.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), ItemStatus: "pausing", Completion: 80},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}
	historySvc := &mockStatusHistorySvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, historySvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 100.0, "", "", "", false)
	require.NoError(t, err)

	// "pausing" -> "completed" is NOT a valid transition
	assert.Nil(t, subItemRepo.updatedFields["item_status"])
	assert.Empty(t, historySvc.recorded)
}

func TestProgressAppend_Progressing_LessThan100_NoStatusChange(t *testing.T) {
	progressRepo := &mockProgressRepo{
		latest: &model.ProgressRecord{Completion: 30.0},
	}
	subItemRepo := &mockSubItemRepoForProgress{
		item: &model.SubItem{BaseModel: model.BaseModel{ID: 5, BizKey: 5}, MainItemKey: int64(10), ItemStatus: "progressing", Completion: 30},
	}
	mainItemSvc := &mockMainItemSvcForProgress{}
	historySvc := &mockStatusHistorySvcForProgress{}

	svc := NewProgressService(progressRepo, subItemRepo, mainItemSvc, historySvc)

	_, err := svc.Append(context.Background(), 1, 2, 5, 60.0, "", "", "", false)
	require.NoError(t, err)

	// Already progressing, not 100%, no transition
	assert.Nil(t, subItemRepo.updatedFields["item_status"])
	assert.Empty(t, historySvc.recorded)
}

// ---------------------------------------------------------------------------
// Tests: GetByBizKey
// ---------------------------------------------------------------------------

func TestProgressGetByBizKey_Found(t *testing.T) {
	records := []model.ProgressRecord{
		{BizKey: 123456, Completion: 60.0, Achievement: "did stuff"},
	}
	progressRepo := &mockProgressRepo{records: records}
	svc := NewProgressService(progressRepo, &mockSubItemRepoForProgress{}, &mockMainItemSvcForProgress{}, &mockStatusHistorySvcForProgress{})

	record, err := svc.GetByBizKey(context.Background(), 123456)
	require.NoError(t, err)
	assert.Equal(t, float64(60.0), record.Completion)
	assert.Equal(t, "did stuff", record.Achievement)
}

func TestProgressGetByBizKey_NotFound(t *testing.T) {
	progressRepo := &mockProgressRepo{records: []model.ProgressRecord{}}
	svc := NewProgressService(progressRepo, &mockSubItemRepoForProgress{}, &mockMainItemSvcForProgress{}, &mockStatusHistorySvcForProgress{})

	_, err := svc.GetByBizKey(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}
