package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock repos for MilestoneService tests
// ---------------------------------------------------------------------------

type mockMilestoneRepo struct {
	item       *model.Milestone
	bizKeyItem *model.Milestone
	items      []model.Milestone
	findErr    error
	bizKeyErr  error
	createErr  error
	updateErr  error
	listErr    error
	// capture calls
	createdItem    *model.Milestone
	updatedID      uint
	updatedFields  map[string]interface{}
	softDeletedID  uint
	deleteByMapKey int64
}

func (m *mockMilestoneRepo) Create(_ context.Context, item *model.Milestone) error {
	m.createdItem = item
	if m.createErr != nil {
		return m.createErr
	}
	item.ID = 1
	return nil
}

func (m *mockMilestoneRepo) FindByID(_ context.Context, id uint) (*model.Milestone, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findErr
}

func (m *mockMilestoneRepo) FindByBizKey(_ context.Context, _ int64) (*model.Milestone, error) {
	if m.bizKeyItem != nil {
		return m.bizKeyItem, nil
	}
	return nil, m.bizKeyErr
}

func (m *mockMilestoneRepo) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.Milestone, error) {
	return nil, nil
}

func (m *mockMilestoneRepo) Update(_ context.Context, item *model.Milestone, fields map[string]interface{}) error {
	m.updatedID = item.ID
	m.updatedFields = fields
	return m.updateErr
}

func (m *mockMilestoneRepo) ListByMap(_ context.Context, _ int64) ([]model.Milestone, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
}

func (m *mockMilestoneRepo) ListByTeam(_ context.Context, _ int64, _ bool) ([]model.Milestone, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
}

func (m *mockMilestoneRepo) SoftDelete(_ context.Context, id uint) error {
	m.softDeletedID = id
	return nil
}

func (m *mockMilestoneRepo) DeleteByMap(_ context.Context, bizKey int64) error {
	m.deleteByMapKey = bizKey
	return nil
}

type mockMilestoneMapRepo struct {
	item      *model.MilestoneMap
	bizKeyErr error
}

func (m *mockMilestoneMapRepo) Create(_ context.Context, _ *model.MilestoneMap) error {
	return nil
}

func (m *mockMilestoneMapRepo) FindByID(_ context.Context, _ uint) (*model.MilestoneMap, error) {
	return nil, nil
}

func (m *mockMilestoneMapRepo) FindByBizKey(_ context.Context, _ int64) (*model.MilestoneMap, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.bizKeyErr
}

func (m *mockMilestoneMapRepo) Update(_ context.Context, _ *model.MilestoneMap, _ map[string]interface{}) error {
	return nil
}

func (m *mockMilestoneMapRepo) List(_ context.Context, _ int64, _ dto.MilestoneMapFilter, _ dto.Pagination) (*dto.PageResult[model.MilestoneMap], error) {
	return nil, nil
}

func (m *mockMilestoneMapRepo) SoftDelete(_ context.Context, _ uint) error {
	return nil
}

// msMockMainItemRepo for milestone-related tests
type msMockMainItemRepo struct {
	unbindErr           error
	calcCompletionVal   float64
	calcCompletionErr   error
	countByMilestone    int64
	countByMilestoneErr error
	unbindCalledWith    int64
}

func (m *msMockMainItemRepo) Create(_ context.Context, _ *model.MainItem) error { return nil }
func (m *msMockMainItemRepo) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *msMockMainItemRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *msMockMainItemRepo) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}
func (m *msMockMainItemRepo) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *msMockMainItemRepo) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *msMockMainItemRepo) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *msMockMainItemRepo) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *msMockMainItemRepo) NextCode(_ context.Context, _ int64) (string, error) {
	return "", nil
}
func (m *msMockMainItemRepo) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *msMockMainItemRepo) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}

func (m *msMockMainItemRepo) UnbindByMilestone(_ context.Context, milestoneBizKey int64) error {
	m.unbindCalledWith = milestoneBizKey
	return m.unbindErr
}

func (m *msMockMainItemRepo) CalcCompletionByMilestone(_ context.Context, _ int64) (float64, error) {
	return m.calcCompletionVal, m.calcCompletionErr
}

func (m *msMockMainItemRepo) CountByMilestone(_ context.Context, _ int64) (int64, error) {
	return m.countByMilestone, m.countByMilestoneErr
}

// mockTransactor implements repo.DBTransactor for tests
type mockTransactor struct {
	txErr  error
	called bool
}

func (m *mockTransactor) Transaction(fc func(tx *gorm.DB) error, _ ...*sql.TxOptions) error {
	m.called = true
	if m.txErr != nil {
		return m.txErr
	}
	return fc(nil)
}

// ---------------------------------------------------------------------------
// Helper: new service with default mocks
// ---------------------------------------------------------------------------

func newTestMilestoneService(milestoneRepo *mockMilestoneRepo, mapRepo *mockMilestoneMapRepo, mainItemRepo *msMockMainItemRepo, tx *mockTransactor) MilestoneService {
	return NewMilestoneService(milestoneRepo, mapRepo, mainItemRepo, tx)
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestMilestoneCreate_Success(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{item: &model.MilestoneMap{}}
	milestoneRepo := &mockMilestoneRepo{}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, mapRepo, mainItemRepo, tx)

	dateStr := "2026-06-01"
	m, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "M1",
		ExpectedEndDate: &dateStr,
	})
	require.NoError(t, err)
	assert.Equal(t, "M1", m.MilestoneName)
	assert.Equal(t, "not_started", m.MilestoneStatus)
	assert.Equal(t, int64(1), m.TeamKey)
	assert.Equal(t, int64(100), m.MilestoneMapKey)
	assert.NotNil(t, m.ExpectedEndDate)
}

func TestMilestoneCreate_ParentMapNotFound(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{bizKeyErr: apperrors.ErrMilestoneMapNotFound}
	milestoneRepo := &mockMilestoneRepo{}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, mapRepo, mainItemRepo, tx)

	dateStr := "2026-06-01"
	_, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "M1",
		ExpectedEndDate: &dateStr,
	})
	assert.ErrorIs(t, err, apperrors.ErrMilestoneMapNotFound)
}

func TestMilestoneCreate_InvalidDate(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{item: &model.MilestoneMap{}}
	milestoneRepo := &mockMilestoneRepo{}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, mapRepo, mainItemRepo, tx)

	badDate := "not-a-date"
	_, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "M1",
		ExpectedEndDate: &badDate,
	})
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestMilestoneCreate_RepoCreateError(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{item: &model.MilestoneMap{}}
	milestoneRepo := &mockMilestoneRepo{createErr: errors.New("db error")}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, mapRepo, mainItemRepo, tx)

	dateStr := "2026-06-01"
	_, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "M1",
		ExpectedEndDate: &dateStr,
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestMilestoneGet_Success(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:     model.BaseModel{ID: 1},
		MilestoneName: "M1",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	m, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "M1", m.MilestoneName)
}

func TestMilestoneGet_NotFound(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{findErr: gorm.ErrRecordNotFound}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	_, err := svc.Get(context.Background(), 99)
	assert.ErrorIs(t, err, apperrors.ErrMilestoneNotFound)
}

// ---------------------------------------------------------------------------
// Tests: GetByBizKey
// ---------------------------------------------------------------------------

func TestMilestoneGetByBizKey_Success(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:     model.BaseModel{ID: 1, BizKey: 123456},
		MilestoneName: "M1",
	}
	milestoneRepo := &mockMilestoneRepo{bizKeyItem: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	m, err := svc.GetByBizKey(context.Background(), 123456)
	require.NoError(t, err)
	assert.Equal(t, "M1", m.MilestoneName)
}

func TestMilestoneGetByBizKey_NotFound(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{bizKeyErr: gorm.ErrRecordNotFound}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	_, err := svc.GetByBizKey(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrMilestoneNotFound)
}

// ---------------------------------------------------------------------------
// Tests: ListByMap
// ---------------------------------------------------------------------------

func TestMilestoneListByMap_Success(t *testing.T) {
	items := []model.Milestone{
		{BaseModel: model.BaseModel{ID: 1}, MilestoneName: "M1"},
		{BaseModel: model.BaseModel{ID: 2}, MilestoneName: "M2"},
	}
	milestoneRepo := &mockMilestoneRepo{items: items}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	result, err := svc.ListByMap(context.Background(), 100)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestMilestoneListByMap_RepoError(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{listErr: errors.New("db error")}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	_, err := svc.ListByMap(context.Background(), 100)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: ListByTeam
// ---------------------------------------------------------------------------

func TestMilestoneListByTeam_Success(t *testing.T) {
	items := []model.Milestone{
		{BaseModel: model.BaseModel{ID: 1}, MilestoneName: "M1"},
	}
	milestoneRepo := &mockMilestoneRepo{items: items}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	result, err := svc.ListByTeam(context.Background(), 1, true)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestMilestoneListByTeam_RepoError(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{listErr: errors.New("db error")}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	_, err := svc.ListByTeam(context.Background(), 1, true)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestMilestoneUpdate_Success(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	existing := &model.Milestone{
		BaseModel:     model.BaseModel{ID: 1, DbUpdateTime: now},
		MilestoneName: "Old Name",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	updated, err := svc.Update(context.Background(), 1, dto.MilestoneUpdateReq{
		MilestoneName: ptrStr("New Name"),
		DbUpdateTime:  now.Format(time.RFC3339),
	})
	require.NoError(t, err)
	assert.Equal(t, "New Name", milestoneRepo.updatedFields["milestone_name"])
	assert.Equal(t, uint(1), milestoneRepo.updatedID)
	// Since FindByID returns the same item, the returned value reflects the update
	assert.NotNil(t, updated)
}

func TestMilestoneUpdate_NotFound(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{findErr: gorm.ErrRecordNotFound}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	_, err := svc.Update(context.Background(), 99, dto.MilestoneUpdateReq{
		MilestoneName: ptrStr("New Name"),
		DbUpdateTime:  time.Now().Format(time.RFC3339),
	})
	assert.ErrorIs(t, err, apperrors.ErrMilestoneNotFound)
}

func TestMilestoneUpdate_ConcurrentEdit(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	past := now.Add(-time.Hour)
	existing := &model.Milestone{
		BaseModel:     model.BaseModel{ID: 1, DbUpdateTime: now},
		MilestoneName: "Old Name",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	_, err := svc.Update(context.Background(), 1, dto.MilestoneUpdateReq{
		MilestoneName: ptrStr("New Name"),
		DbUpdateTime:  past.Format(time.RFC3339), // stale timestamp
	})
	assert.ErrorIs(t, err, apperrors.ErrConcurrentEdit)
}

func TestMilestoneUpdate_InvalidDbUpdateTime(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:     model.BaseModel{ID: 1},
		MilestoneName: "Old Name",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	_, err := svc.Update(context.Background(), 1, dto.MilestoneUpdateReq{
		MilestoneName: ptrStr("New Name"),
		DbUpdateTime:  "not-a-time",
	})
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestMilestoneUpdate_NoFields_NoOp(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	existing := &model.Milestone{
		BaseModel:     model.BaseModel{ID: 1, DbUpdateTime: now},
		MilestoneName: "Old Name",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	updated, err := svc.Update(context.Background(), 1, dto.MilestoneUpdateReq{
		DbUpdateTime: now.Format(time.RFC3339),
	})
	require.NoError(t, err)
	assert.Equal(t, "Old Name", updated.MilestoneName)
	assert.Nil(t, milestoneRepo.updatedFields)
}

func TestMilestoneUpdate_UpdateExpectedEndDate(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	existing := &model.Milestone{
		BaseModel:     model.BaseModel{ID: 1, DbUpdateTime: now},
		MilestoneName: "M1",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	newDate := "2026-07-01"
	_, err := svc.Update(context.Background(), 1, dto.MilestoneUpdateReq{
		ExpectedEndDate: &newDate,
		DbUpdateTime:    now.Format(time.RFC3339),
	})
	require.NoError(t, err)
	assert.Equal(t, newDate, milestoneRepo.updatedFields["expected_end_date"])
}

// ---------------------------------------------------------------------------
// Tests: Delete
// ---------------------------------------------------------------------------

func TestMilestoneDelete_Success(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:     model.BaseModel{ID: 10, BizKey: 555},
		MilestoneName: "M1",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	err := svc.Delete(context.Background(), 10)
	require.NoError(t, err)
	assert.True(t, tx.called)
	assert.Equal(t, int64(555), mainItemRepo.unbindCalledWith)
	assert.Equal(t, uint(10), milestoneRepo.softDeletedID)
}

func TestMilestoneDelete_NotFound(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{findErr: gorm.ErrRecordNotFound}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	err := svc.Delete(context.Background(), 99)
	assert.ErrorIs(t, err, apperrors.ErrMilestoneNotFound)
}

func TestMilestoneDelete_UnbindError(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:     model.BaseModel{ID: 10, BizKey: 555},
		MilestoneName: "M1",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{unbindErr: errors.New("db error")}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	err := svc.Delete(context.Background(), 10)
	assert.Error(t, err)
}

func TestMilestoneDelete_SoftDeleteError(t *testing.T) {
	existing := &model.Milestone{
		BaseModel: model.BaseModel{ID: 10, BizKey: 555},
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{txErr: errors.New("tx error")}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	err := svc.Delete(context.Background(), 10)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus
// ---------------------------------------------------------------------------

func TestMilestoneChangeStatus_NotStartedToInProgress(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10, BizKey: 555},
		MilestoneStatus: "not_started",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	updated, err := svc.ChangeStatus(context.Background(), 10, "in_progress")
	require.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "in_progress", milestoneRepo.updatedFields["milestone_status"])
}

func TestMilestoneChangeStatus_ToCancelled_AutoUnbinds(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10, BizKey: 555},
		MilestoneStatus: "not_started",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	updated, err := svc.ChangeStatus(context.Background(), 10, "cancelled")
	require.NoError(t, err)
	assert.NotNil(t, updated)
	assert.True(t, tx.called)
	assert.Equal(t, int64(555), mainItemRepo.unbindCalledWith)
	assert.Equal(t, "cancelled", milestoneRepo.updatedFields["milestone_status"])
}

func TestMilestoneChangeStatus_InProgressToCompleted(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10, BizKey: 555},
		MilestoneStatus: "in_progress",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	updated, err := svc.ChangeStatus(context.Background(), 10, "completed")
	require.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "completed", milestoneRepo.updatedFields["milestone_status"])
}

func TestMilestoneChangeStatus_InvalidTransition(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10},
		MilestoneStatus: "not_started",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	_, err := svc.ChangeStatus(context.Background(), 10, "completed")
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
}

func TestMilestoneChangeStatus_InvalidStatus(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10},
		MilestoneStatus: "not_started",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	_, err := svc.ChangeStatus(context.Background(), 10, "unknown_status")
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
}

func TestMilestoneChangeStatus_SelfTransition(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10},
		MilestoneStatus: "not_started",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	_, err := svc.ChangeStatus(context.Background(), 10, "not_started")
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
}

func TestMilestoneChangeStatus_NotFound(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{findErr: gorm.ErrRecordNotFound}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	_, err := svc.ChangeStatus(context.Background(), 99, "in_progress")
	assert.ErrorIs(t, err, apperrors.ErrMilestoneNotFound)
}

func TestMilestoneChangeStatus_CompletedToCancelled(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10, BizKey: 555},
		MilestoneStatus: "completed",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	updated, err := svc.ChangeStatus(context.Background(), 10, "cancelled")
	require.NoError(t, err)
	assert.NotNil(t, updated)
	assert.True(t, tx.called)
	assert.Equal(t, int64(555), mainItemRepo.unbindCalledWith)
}

func TestMilestoneChangeStatus_CancelledIsTerminal(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10},
		MilestoneStatus: "cancelled",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	mainItemRepo := &msMockMainItemRepo{}
	tx := &mockTransactor{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, tx)

	_, err := svc.ChangeStatus(context.Background(), 10, "in_progress")
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
}

// ---------------------------------------------------------------------------
// Tests: AvailableTransitions
// ---------------------------------------------------------------------------

func TestMilestoneAvailableTransitions_NotStarted(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10},
		MilestoneStatus: "not_started",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 10)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"in_progress", "cancelled"}, transitions)
}

func TestMilestoneAvailableTransitions_InProgress(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10},
		MilestoneStatus: "in_progress",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 10)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"completed", "cancelled"}, transitions)
}

func TestMilestoneAvailableTransitions_CancelledEmpty(t *testing.T) {
	existing := &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10},
		MilestoneStatus: "cancelled",
	}
	milestoneRepo := &mockMilestoneRepo{item: existing}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 10)
	require.NoError(t, err)
	assert.Empty(t, transitions)
}

func TestMilestoneAvailableTransitions_NotFound(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{findErr: gorm.ErrRecordNotFound}
	svc := newTestMilestoneService(milestoneRepo, nil, nil, nil)

	_, err := svc.AvailableTransitions(context.Background(), 99)
	assert.ErrorIs(t, err, apperrors.ErrMilestoneNotFound)
}

// ---------------------------------------------------------------------------
// Tests: CalcCompletion
// ---------------------------------------------------------------------------

func TestMilestoneCalcCompletion_WithValue(t *testing.T) {
	mainItemRepo := &msMockMainItemRepo{calcCompletionVal: 75.5}
	milestoneRepo := &mockMilestoneRepo{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, nil)

	result := svc.CalcCompletion(context.Background(), 555)
	assert.Equal(t, 75.5, result)
}

func TestMilestoneCalcCompletion_NoMIs(t *testing.T) {
	mainItemRepo := &msMockMainItemRepo{calcCompletionVal: 0}
	milestoneRepo := &mockMilestoneRepo{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, nil)

	result := svc.CalcCompletion(context.Background(), 555)
	assert.Equal(t, float64(0), result)
}

func TestMilestoneCalcCompletion_Error(t *testing.T) {
	mainItemRepo := &msMockMainItemRepo{calcCompletionErr: errors.New("db error")}
	milestoneRepo := &mockMilestoneRepo{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, nil)

	result := svc.CalcCompletion(context.Background(), 555)
	assert.Equal(t, float64(0), result)
}

// ---------------------------------------------------------------------------
// Tests: CountRelatedMIs
// ---------------------------------------------------------------------------

func TestMilestoneCountRelatedMIs_WithValue(t *testing.T) {
	mainItemRepo := &msMockMainItemRepo{countByMilestone: 5}
	milestoneRepo := &mockMilestoneRepo{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, nil)

	result := svc.CountRelatedMIs(context.Background(), 555)
	assert.Equal(t, int64(5), result)
}

func TestMilestoneCountRelatedMIs_Zero(t *testing.T) {
	mainItemRepo := &msMockMainItemRepo{countByMilestone: 0}
	milestoneRepo := &mockMilestoneRepo{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, nil)

	result := svc.CountRelatedMIs(context.Background(), 555)
	assert.Equal(t, int64(0), result)
}

func TestMilestoneCountRelatedMIs_Error(t *testing.T) {
	mainItemRepo := &msMockMainItemRepo{countByMilestoneErr: errors.New("db error")}
	milestoneRepo := &mockMilestoneRepo{}
	svc := newTestMilestoneService(milestoneRepo, nil, mainItemRepo, nil)

	result := svc.CountRelatedMIs(context.Background(), 555)
	assert.Equal(t, int64(0), result)
}
