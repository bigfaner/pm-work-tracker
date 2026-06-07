//nolint:misspell // "cancelled" is a domain status value per PRD/API contract, not a misspelling.
package service

import (
	"context"
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
// Helper: create a sample Milestone
// ---------------------------------------------------------------------------

func sampleMilestone(msStatus string) *model.Milestone {
	return &model.Milestone{
		BaseModel:       model.BaseModel{ID: 10, BizKey: 200},
		TeamKey:         1,
		MilestoneMapKey: 100,
		MilestoneName:   "Test Milestone",
		MilestoneDesc:   "A test milestone",
		MilestoneStatus: msStatus,
	}
}

// ---------------------------------------------------------------------------
// Mock: MilestoneRepo for MilestoneService (primary repo)
// ---------------------------------------------------------------------------

type mockMilestoneRepo struct {
	item          *model.Milestone
	items         []model.Milestone
	findByIDErr   error
	findByBizErr  error
	createErr     error
	updateErr     error
	listByMapErr  error
	listByTeamErr error
	softDeleteErr error
	existsResult  bool
	existsErr     error

	createdItem   *model.Milestone
	updatedItem   *model.Milestone
	updatedFields map[string]interface{}
	deletedID     *uint
}

func (m *mockMilestoneRepo) Create(_ context.Context, item *model.Milestone) error {
	m.createdItem = item
	if m.createErr != nil {
		return m.createErr
	}
	item.ID = 10
	return nil
}

func (m *mockMilestoneRepo) FindByID(_ context.Context, id uint) (*model.Milestone, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findByIDErr
}

func (m *mockMilestoneRepo) FindByBizKey(_ context.Context, bizKey int64) (*model.Milestone, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findByBizErr
}

func (m *mockMilestoneRepo) FindBatchByBizKeys(_ context.Context, _ []int64) (map[int64]*model.Milestone, error) {
	return nil, nil
}

func (m *mockMilestoneRepo) Update(_ context.Context, item *model.Milestone, fields map[string]interface{}) error {
	m.updatedItem = item
	m.updatedFields = fields
	return m.updateErr
}

func (m *mockMilestoneRepo) ListByMap(_ context.Context, _ int64) ([]model.Milestone, error) {
	if m.listByMapErr != nil {
		return nil, m.listByMapErr
	}
	return m.items, nil
}

func (m *mockMilestoneRepo) ListByTeam(_ context.Context, _ int64, _ dto.MilestoneTeamFilter) ([]model.Milestone, error) {
	if m.listByTeamErr != nil {
		return nil, m.listByTeamErr
	}
	return m.items, nil
}

func (m *mockMilestoneRepo) SoftDelete(_ context.Context, id uint) error {
	m.deletedID = &id
	return m.softDeleteErr
}

func (m *mockMilestoneRepo) SoftDeleteByMap(_ context.Context, _ int64) error {
	return nil
}

func (m *mockMilestoneRepo) ExistsByNameAndMap(_ context.Context, _ int64, _ string, _ *uint) (bool, error) {
	return m.existsResult, m.existsErr
}

// ---------------------------------------------------------------------------
// Mock: MilestoneMapRepo for MilestoneService (secondary, used for BR-5)
// ---------------------------------------------------------------------------

type mockMapRepoForMilestone struct {
	item         *model.MilestoneMap
	findByBizErr error
}

func (m *mockMapRepoForMilestone) Create(_ context.Context, _ *model.MilestoneMap) error {
	return nil
}

func (m *mockMapRepoForMilestone) FindByID(_ context.Context, _ uint) (*model.MilestoneMap, error) {
	return nil, nil
}

func (m *mockMapRepoForMilestone) FindByBizKey(_ context.Context, _ int64) (*model.MilestoneMap, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findByBizErr
}

func (m *mockMapRepoForMilestone) Update(_ context.Context, _ *model.MilestoneMap, _ map[string]interface{}) error {
	return nil
}

func (m *mockMapRepoForMilestone) List(_ context.Context, _ int64, _ dto.MilestoneMapFilter, _ dto.Pagination) (*dto.PageResult[model.MilestoneMap], error) {
	return nil, nil
}

func (m *mockMapRepoForMilestone) SoftDelete(_ context.Context, _ uint) error {
	return nil
}

// ---------------------------------------------------------------------------
// Mock: MainItemRepo for MilestoneService
// ---------------------------------------------------------------------------

type mockMainItemRepoForMs struct {
	items                 []model.MainItem
	findByMilestoneKeyErr error
	clearByMilestoneErr   error
	countByMilestoneKey   int64
	countErr              error

	clearedBizKey int64
}

func (m *mockMainItemRepoForMs) Create(_ context.Context, _ *model.MainItem) error { return nil }
func (m *mockMainItemRepoForMs) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMs) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMs) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMs) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMs) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockMainItemRepoForMs) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemRepoForMs) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMs) SoftDelete(_ context.Context, _ uint) error { return nil }
func (m *mockMainItemRepoForMs) CascadeSoftDelete(_ context.Context, _ uint, _ []uint, _ []model.StatusHistory) error {
	return nil
}
func (m *mockMainItemRepoForMs) NextCode(_ context.Context, _ int64) (string, error) {
	return "T001", nil
}
func (m *mockMainItemRepoForMs) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepoForMs) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}

func (m *mockMainItemRepoForMs) FindByMilestoneKey(_ context.Context, _ int64) ([]model.MainItem, error) {
	if m.findByMilestoneKeyErr != nil {
		return nil, m.findByMilestoneKeyErr
	}
	return m.items, nil
}

func (m *mockMainItemRepoForMs) CountByMilestoneKey(_ context.Context, _ int64) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	if m.countByMilestoneKey > 0 {
		return m.countByMilestoneKey, nil
	}
	return int64(len(m.items)), nil
}

func (m *mockMainItemRepoForMs) ClearMilestoneKeyByMilestone(_ context.Context, bizKey int64) error {
	m.clearedBizKey = bizKey
	return m.clearByMilestoneErr
}

func (m *mockMainItemRepoForMs) ClearMilestoneKeyByMap(_ context.Context, _ []int64) error {
	return nil
}

// ---------------------------------------------------------------------------
// Tests: Create (AC-1)
// ---------------------------------------------------------------------------

func TestMilestoneCreate_Success(t *testing.T) {
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "planning",
	}}
	msRepo := &mockMilestoneRepo{}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	result, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "Phase 1",
		MilestoneDesc:   "First phase",
		ExpectedEndDate: "2026-06-30",
	})
	require.NoError(t, err)
	assert.Equal(t, "Phase 1", msRepo.createdItem.MilestoneName)
	assert.Equal(t, "not_started", msRepo.createdItem.MilestoneStatus)
	assert.Equal(t, int64(1), msRepo.createdItem.TeamKey)
	assert.Equal(t, int64(100), msRepo.createdItem.MilestoneMapKey)
	assert.NotZero(t, msRepo.createdItem.BizKey)
	assert.NotNil(t, msRepo.createdItem.ExpectedEndDate)
	assert.Equal(t, uint(10), result.ID)
}

func TestMilestoneCreate_BR5_TerminalMap(t *testing.T) {
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "completed",
	}}
	svc := NewMilestoneService(nil, mapRepo, nil, nil)

	_, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "Phase 1",
		ExpectedEndDate: "2026-06-30",
	})
	assert.ErrorIs(t, err, apperrors.ErrMapIsTerminal)
}

func TestMilestoneCreate_BR5_CancelledMap(t *testing.T) {
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "cancelled",
	}}
	svc := NewMilestoneService(nil, mapRepo, nil, nil)

	_, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "Phase 1",
		ExpectedEndDate: "2026-06-30",
	})
	assert.ErrorIs(t, err, apperrors.ErrMapIsTerminal)
}

func TestMilestoneCreate_DuplicateName(t *testing.T) {
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "planning",
	}}
	msRepo := &mockMilestoneRepo{existsResult: true}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	_, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "Phase 1",
		ExpectedEndDate: "2026-06-30",
	})
	assert.ErrorIs(t, err, apperrors.ErrDuplicateMilestoneName)
}

func TestMilestoneCreate_MapNotFound(t *testing.T) {
	mapRepo := &mockMapRepoForMilestone{findByBizErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneService(nil, mapRepo, nil, nil)

	_, err := svc.Create(context.Background(), 1, 999, dto.MilestoneCreateReq{
		MilestoneName:   "Phase 1",
		ExpectedEndDate: "2026-06-30",
	})
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestMilestoneCreate_InvalidDate(t *testing.T) {
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "planning",
	}}
	svc := NewMilestoneService(&mockMilestoneRepo{}, mapRepo, nil, nil)

	_, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "Phase 1",
		ExpectedEndDate: "not-a-date",
	})
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestMilestoneCreate_GeneratesBizKey(t *testing.T) {
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "planning",
	}}
	msRepo := &mockMilestoneRepo{}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	result, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "Phase 1",
		ExpectedEndDate: "2026-06-30",
	})
	require.NoError(t, err)
	assert.NotZero(t, result.BizKey)
}

// ---------------------------------------------------------------------------
// Tests: Update (AC-2)
// ---------------------------------------------------------------------------

func TestMilestoneUpdate_Success(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "planning",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	err := svc.Update(context.Background(), 1, 10, dto.MilestoneUpdateReq{
		MilestoneName: msStrPtr("Updated Name"),
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", msRepo.updatedFields["milestone_name"])
}

func TestMilestoneUpdate_PartialDescOnly(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	err := svc.Update(context.Background(), 1, 10, dto.MilestoneUpdateReq{
		MilestoneDesc: msStrPtr("New desc"),
	})
	require.NoError(t, err)
	assert.Equal(t, "New desc", msRepo.updatedFields["milestone_desc"])
	// No name change, no BR-5 check needed
	assert.Nil(t, msRepo.updatedFields["milestone_name"])
}

func TestMilestoneUpdate_NoFields(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	err := svc.Update(context.Background(), 1, 10, dto.MilestoneUpdateReq{})
	require.NoError(t, err)
	assert.Nil(t, msRepo.updatedFields)
}

func TestMilestoneUpdate_BR5_TerminalMap(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "completed",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	err := svc.Update(context.Background(), 1, 10, dto.MilestoneUpdateReq{
		MilestoneName: msStrPtr("New Name"),
	})
	assert.ErrorIs(t, err, apperrors.ErrMapIsTerminal)
}

func TestMilestoneUpdate_DuplicateName(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms, existsResult: true}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "planning",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	err := svc.Update(context.Background(), 1, 10, dto.MilestoneUpdateReq{
		MilestoneName: msStrPtr("Existing Name"),
	})
	assert.ErrorIs(t, err, apperrors.ErrDuplicateMilestoneName)
}

func TestMilestoneUpdate_NotFound(t *testing.T) {
	msRepo := &mockMilestoneRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	err := svc.Update(context.Background(), 1, 999, dto.MilestoneUpdateReq{MilestoneName: msStrPtr("X")})
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestMilestoneUpdate_WrongTeam(t *testing.T) {
	ms := sampleMilestone("not_started")
	ms.TeamKey = 99
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	err := svc.Update(context.Background(), 1, 10, dto.MilestoneUpdateReq{MilestoneName: msStrPtr("X")})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestMilestoneUpdate_InvalidDate(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "planning",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	err := svc.Update(context.Background(), 1, 10, dto.MilestoneUpdateReq{
		ExpectedEndDate: msStrPtr("bad-date"),
	})
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

// ---------------------------------------------------------------------------
// Tests: Get / GetByBizKey
// ---------------------------------------------------------------------------

func TestMilestoneGet_Success(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	result, err := svc.Get(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, ms, result)
}

func TestMilestoneGet_NotFound(t *testing.T) {
	msRepo := &mockMilestoneRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	_, err := svc.Get(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestMilestoneGetByBizKey_Success(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	result, err := svc.GetByBizKey(context.Background(), 200)
	require.NoError(t, err)
	assert.Equal(t, ms, result)
}

func TestMilestoneGetByBizKey_NotFound(t *testing.T) {
	msRepo := &mockMilestoneRepo{findByBizErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	_, err := svc.GetByBizKey(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Tests: ListByMap / ListByTeam
// ---------------------------------------------------------------------------

func TestMilestoneListByMap_Success(t *testing.T) {
	items := []model.Milestone{*sampleMilestone("not_started"), *sampleMilestone("in_progress")}
	msRepo := &mockMilestoneRepo{items: items}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	result, err := svc.ListByMap(context.Background(), 100)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestMilestoneListByTeam_Success(t *testing.T) {
	items := []model.Milestone{*sampleMilestone("not_started")}
	msRepo := &mockMilestoneRepo{items: items}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	result, err := svc.ListByTeam(context.Background(), 1, dto.MilestoneTeamFilter{})
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

// ---------------------------------------------------------------------------
// Tests: Delete (AC-3)
// ---------------------------------------------------------------------------

func TestMilestoneDelete_NotStarted(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	miRepo := &mockMainItemRepoForMs{}
	tx := &mockDBTxForMap{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewMilestoneService(msRepo, nil, miRepo, tx)

	err := svc.Delete(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(200), miRepo.clearedBizKey)
	assert.NotNil(t, msRepo.deletedID)
}

func TestMilestoneDelete_Cancelled(t *testing.T) {
	ms := sampleMilestone("cancelled")
	msRepo := &mockMilestoneRepo{item: ms}
	miRepo := &mockMainItemRepoForMs{}
	tx := &mockDBTxForMap{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewMilestoneService(msRepo, nil, miRepo, tx)

	err := svc.Delete(context.Background(), 1, 10)
	require.NoError(t, err)
}

func TestMilestoneDelete_InProgressBlocked(t *testing.T) {
	ms := sampleMilestone("in_progress")
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	err := svc.Delete(context.Background(), 1, 10)
	assert.ErrorIs(t, err, apperrors.ErrMilestoneCannotDelete)
}

func TestMilestoneDelete_CompletedBlocked(t *testing.T) {
	ms := sampleMilestone("completed")
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	err := svc.Delete(context.Background(), 1, 10)
	assert.ErrorIs(t, err, apperrors.ErrMilestoneCannotDelete)
}

func TestMilestoneDelete_WrongTeam(t *testing.T) {
	ms := sampleMilestone("not_started")
	ms.TeamKey = 99
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	err := svc.Delete(context.Background(), 1, 10)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestMilestoneDelete_NotFound(t *testing.T) {
	msRepo := &mockMilestoneRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	err := svc.Delete(context.Background(), 1, 999)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestMilestoneDelete_UnbindsMIsInTransaction(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	miRepo := &mockMainItemRepoForMs{}
	tx := &mockDBTxForMap{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewMilestoneService(msRepo, nil, miRepo, tx)

	err := svc.Delete(context.Background(), 1, 10)
	require.NoError(t, err)
	// MI keys should be cleared with the milestone's BizKey
	assert.Equal(t, int64(200), miRepo.clearedBizKey)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus (AC-4)
// ---------------------------------------------------------------------------

func TestMilestoneChangeStatus_ValidTransition(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, "in_progress")
	require.NoError(t, err)
	assert.Equal(t, "in_progress", msRepo.updatedFields["milestone_status"])
	assert.NotNil(t, result)
}

func TestMilestoneChangeStatus_InvalidTransition(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, "completed")
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
}

func TestMilestoneChangeStatus_BR5_TerminalMap(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "completed",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, "in_progress")
	assert.ErrorIs(t, err, apperrors.ErrMapIsTerminal)
}

func TestMilestoneChangeStatus_BR1_Completed_AllMIsTerminal(t *testing.T) {
	ms := sampleMilestone("in_progress")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	miRepo := &mockMainItemRepoForMs{
		items: []model.MainItem{
			{ItemStatus: "completed"},
			{ItemStatus: "closed"},
		},
	}
	svc := NewMilestoneService(msRepo, mapRepo, miRepo, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, "completed")
	require.NoError(t, err)
	assert.Equal(t, "completed", msRepo.updatedFields["milestone_status"])
	assert.NotNil(t, result)
}

func TestMilestoneChangeStatus_BR1_Completed_NonTerminalMIs(t *testing.T) {
	ms := sampleMilestone("in_progress")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	miRepo := &mockMainItemRepoForMs{
		items: []model.MainItem{
			{ItemStatus: "completed"},
			{ItemStatus: "pending"},
		},
	}
	svc := NewMilestoneService(msRepo, mapRepo, miRepo, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, "completed")
	assert.ErrorIs(t, err, apperrors.ErrMilestoneHasNonTerminalItems)
}

func TestMilestoneChangeStatus_BR1_Completed_EmptyMilestone(t *testing.T) {
	ms := sampleMilestone("in_progress")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	miRepo := &mockMainItemRepoForMs{items: []model.MainItem{}}
	svc := NewMilestoneService(msRepo, mapRepo, miRepo, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, "completed")
	require.NoError(t, err)
	assert.Equal(t, "completed", msRepo.updatedFields["milestone_status"])
	assert.NotNil(t, result)
}

func TestMilestoneChangeStatus_Cancelled_AutoUnbinds(t *testing.T) {
	ms := sampleMilestone("in_progress")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	miRepo := &mockMainItemRepoForMs{
		items: []model.MainItem{
			{ItemStatus: "pending"},
			{ItemStatus: "in_progress"},
		},
	}
	tx := &mockDBTxForMap{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewMilestoneService(msRepo, mapRepo, miRepo, tx)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, "cancelled")
	require.NoError(t, err)
	assert.Equal(t, "cancelled", msRepo.updatedFields["milestone_status"])
	assert.Equal(t, int64(200), miRepo.clearedBizKey)
	assert.NotNil(t, result)
}

func TestMilestoneChangeStatus_CompletedToInProgress(t *testing.T) {
	ms := sampleMilestone("completed")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, "in_progress")
	require.NoError(t, err)
	assert.Equal(t, "in_progress", msRepo.updatedFields["milestone_status"])
	assert.NotNil(t, result)
}

func TestMilestoneChangeStatus_WrongTeam(t *testing.T) {
	ms := sampleMilestone("not_started")
	ms.TeamKey = 99
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, "in_progress")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestMilestoneChangeStatus_NotFound(t *testing.T) {
	msRepo := &mockMilestoneRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 999, "in_progress")
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Tests: AvailableTransitions (AC-5)
// ---------------------------------------------------------------------------

func TestMilestoneAvailableTransitions_NotStarted(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"in_progress", "cancelled"}, transitions)
}

func TestMilestoneAvailableTransitions_BR5_TerminalMap(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "completed",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Empty(t, transitions)
}

func TestMilestoneAvailableTransitions_BR5_CancelledMap(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "cancelled",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Empty(t, transitions)
}

func TestMilestoneAvailableTransitions_BR1_AllMIsTerminal(t *testing.T) {
	ms := sampleMilestone("in_progress")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	miRepo := &mockMainItemRepoForMs{
		items: []model.MainItem{
			{ItemStatus: "completed"},
		},
	}
	svc := NewMilestoneService(msRepo, mapRepo, miRepo, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Contains(t, transitions, "completed")
}

func TestMilestoneAvailableTransitions_BR1_NonTerminalMIs_FilteredCompleted(t *testing.T) {
	ms := sampleMilestone("in_progress")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	miRepo := &mockMainItemRepoForMs{
		items: []model.MainItem{
			{ItemStatus: "pending"},
		},
	}
	svc := NewMilestoneService(msRepo, mapRepo, miRepo, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.NotContains(t, transitions, "completed")
	assert.Contains(t, transitions, "cancelled")
}

func TestMilestoneAvailableTransitions_TerminalSelf(t *testing.T) {
	ms := sampleMilestone("cancelled")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "executing",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Empty(t, transitions)
}

func TestMilestoneAvailableTransitions_WrongTeam(t *testing.T) {
	ms := sampleMilestone("not_started")
	ms.TeamKey = 99
	msRepo := &mockMilestoneRepo{item: ms}
	svc := NewMilestoneService(msRepo, nil, nil, nil)

	_, err := svc.AvailableTransitions(context.Background(), 1, 10)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: calcCompletion (AC-6)
// ---------------------------------------------------------------------------

func TestCalcCompletion_EmptyMilestone(t *testing.T) {
	miRepo := &mockMainItemRepoForMs{items: []model.MainItem{}}
	svc := NewMilestoneService(nil, nil, miRepo, nil).(*milestoneService)

	result := svc.calcCompletion(context.Background(), 200)
	assert.Equal(t, float64(0), result)
}

func TestCalcCompletion_SingleMI(t *testing.T) {
	miRepo := &mockMainItemRepoForMs{
		items: []model.MainItem{
			{Completion: 75},
		},
	}
	svc := NewMilestoneService(nil, nil, miRepo, nil).(*milestoneService)

	result := svc.calcCompletion(context.Background(), 200)
	assert.Equal(t, float64(75), result)
}

func TestCalcCompletion_MultipleMIs(t *testing.T) {
	miRepo := &mockMainItemRepoForMs{
		items: []model.MainItem{
			{Completion: 50},
			{Completion: 100},
		},
	}
	svc := NewMilestoneService(nil, nil, miRepo, nil).(*milestoneService)

	result := svc.calcCompletion(context.Background(), 200)
	assert.Equal(t, float64(75), result)
}

// ---------------------------------------------------------------------------
// Tests: countRelatedMIs (AC-6)
// ---------------------------------------------------------------------------

func TestCountRelatedMIs_Empty(t *testing.T) {
	miRepo := &mockMainItemRepoForMs{items: []model.MainItem{}}
	svc := NewMilestoneService(nil, nil, miRepo, nil).(*milestoneService)

	count := svc.countRelatedMIs(context.Background(), 200)
	assert.Equal(t, int64(0), count)
}

func TestCountRelatedMIs_WithItems(t *testing.T) {
	miRepo := &mockMainItemRepoForMs{
		items: []model.MainItem{
			{ItemStatus: "pending"},
			{ItemStatus: "completed"},
			{ItemStatus: "in_progress"},
		},
	}
	svc := NewMilestoneService(nil, nil, miRepo, nil).(*milestoneService)

	count := svc.countRelatedMIs(context.Background(), 200)
	assert.Equal(t, int64(3), count)
}

func TestCountRelatedMIs_ExplicitCount(t *testing.T) {
	miRepo := &mockMainItemRepoForMs{countByMilestoneKey: 5}
	svc := NewMilestoneService(nil, nil, miRepo, nil).(*milestoneService)

	count := svc.countRelatedMIs(context.Background(), 200)
	assert.Equal(t, int64(5), count)
}

// ---------------------------------------------------------------------------
// Tests: MapNotFound integration edge case
// ---------------------------------------------------------------------------

func TestMilestoneChangeStatus_MapNotFound(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{findByBizErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 10, "in_progress")
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Additional edge: Ensure time parsing in update works
// ---------------------------------------------------------------------------

func TestMilestoneUpdate_ValidDate(t *testing.T) {
	ms := sampleMilestone("not_started")
	msRepo := &mockMilestoneRepo{item: ms}
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "planning",
	}}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	err := svc.Update(context.Background(), 1, 10, dto.MilestoneUpdateReq{
		ExpectedEndDate: msStrPtr("2026-12-31"),
	})
	require.NoError(t, err)
	dateVal, ok := msRepo.updatedFields["expected_end_date"]
	assert.True(t, ok)
	timeVal, ok := dateVal.(*time.Time)
	assert.True(t, ok)
	assert.Equal(t, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), *timeVal)
}

func TestMilestoneCreate_RepoError(t *testing.T) {
	mapRepo := &mockMapRepoForMilestone{item: &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: 100}, MapStatus: "planning",
	}}
	msRepo := &mockMilestoneRepo{createErr: errors.New("db error")}
	svc := NewMilestoneService(msRepo, mapRepo, nil, nil)

	_, err := svc.Create(context.Background(), 1, 100, dto.MilestoneCreateReq{
		MilestoneName:   "Phase 1",
		ExpectedEndDate: "2026-06-30",
	})
	assert.Error(t, err)
}
