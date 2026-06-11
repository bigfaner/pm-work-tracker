//nolint:misspell // "cancelled" is a domain status value per PRD/API contract, not a misspelling.
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
// Mock repos for MilestoneMapService tests
// ---------------------------------------------------------------------------

type mockMilestoneMapRepo struct {
	item       *model.MilestoneMap
	items      []model.MilestoneMap
	pageResult *dto.PageResult[model.MilestoneMap]

	findByIDErr   error
	findByBizErr  error
	createErr     error
	updateErr     error
	listErr       error
	softDeleteErr error

	createdItem   *model.MilestoneMap
	updatedItem   *model.MilestoneMap
	updatedFields map[string]interface{}
	deletedID     *uint
}

func (m *mockMilestoneMapRepo) Create(_ context.Context, item *model.MilestoneMap) error {
	m.createdItem = item
	if m.createErr != nil {
		return m.createErr
	}
	item.ID = 1
	return nil
}

func (m *mockMilestoneMapRepo) FindByID(_ context.Context, id uint) (*model.MilestoneMap, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findByIDErr
}

func (m *mockMilestoneMapRepo) FindByBizKey(_ context.Context, bizKey int64) (*model.MilestoneMap, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findByBizErr
}

func (m *mockMilestoneMapRepo) Update(_ context.Context, item *model.MilestoneMap, fields map[string]interface{}) error {
	m.updatedItem = item
	m.updatedFields = fields
	return m.updateErr
}

func (m *mockMilestoneMapRepo) List(_ context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.pageResult != nil {
		return m.pageResult, nil
	}
	return &dto.PageResult[model.MilestoneMap]{Items: m.items, Total: int64(len(m.items))}, nil
}

func (m *mockMilestoneMapRepo) SoftDelete(_ context.Context, id uint) error {
	m.deletedID = &id
	return m.softDeleteErr
}

// mockMilestoneRepoForMap captures calls from MilestoneMapService.
type mockMilestoneRepoForMap struct {
	milestones    []model.Milestone
	findByIDItem  *model.Milestone
	findByIDErr   error
	listByMapErr  error
	updateErr     error
	softDeleteErr error
	existsResult  bool
	existsErr     error

	updatedItems  []*model.Milestone
	updatedFields map[string]interface{}
	deletedByMap  bool
	createdItem   *model.Milestone
}

func (m *mockMilestoneRepoForMap) Create(_ context.Context, ms *model.Milestone) error {
	m.createdItem = ms
	ms.ID = 10
	return nil
}

func (m *mockMilestoneRepoForMap) FindByID(_ context.Context, id uint) (*model.Milestone, error) {
	if m.findByIDItem != nil {
		return m.findByIDItem, nil
	}
	return nil, m.findByIDErr
}

func (m *mockMilestoneRepoForMap) FindByBizKey(_ context.Context, bizKey int64) (*model.Milestone, error) {
	if m.findByIDItem != nil {
		return m.findByIDItem, nil
	}
	return nil, m.findByIDErr
}

func (m *mockMilestoneRepoForMap) FindBatchByBizKeys(_ context.Context, _ []int64) (map[int64]*model.Milestone, error) {
	return nil, nil
}

func (m *mockMilestoneRepoForMap) Update(_ context.Context, ms *model.Milestone, fields map[string]interface{}) error {
	m.updatedItems = append(m.updatedItems, ms)
	m.updatedFields = fields
	return m.updateErr
}

func (m *mockMilestoneRepoForMap) ListByMap(_ context.Context, _ int64) ([]model.Milestone, error) {
	if m.listByMapErr != nil {
		return nil, m.listByMapErr
	}
	return m.milestones, nil
}

func (m *mockMilestoneRepoForMap) ListByTeam(_ context.Context, _ int64, _ dto.MilestoneTeamFilter) ([]model.Milestone, error) {
	return nil, nil
}

func (m *mockMilestoneRepoForMap) SoftDelete(_ context.Context, _ uint) error {
	return m.softDeleteErr
}

func (m *mockMilestoneRepoForMap) SoftDeleteByMap(_ context.Context, _ int64) error {
	m.deletedByMap = true
	return m.softDeleteErr
}

func (m *mockMilestoneRepoForMap) ExistsByNameAndMap(_ context.Context, _ int64, _ string, _ *uint) (bool, error) {
	return m.existsResult, m.existsErr
}

// mockMainItemRepoForMap captures calls from MilestoneMapService.
type mockMainItemRepoForMap struct {
	items                    []model.MainItem
	findByMilestoneKeyErr    error
	clearByMilestoneErr      error
	clearByMapErr            error
	clearedByMilestoneBizKey int64
	clearedByMapBizKeys      []int64
}

func (m *mockMainItemRepoForMap) Create(_ context.Context, _ *model.MainItem) error { return nil }
func (m *mockMainItemRepoForMap) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMap) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMap) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMap) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMap) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockMainItemRepoForMap) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemRepoForMap) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMap) SoftDelete(_ context.Context, _ uint) error { return nil }
func (m *mockMainItemRepoForMap) CascadeSoftDelete(_ context.Context, _ uint, _ []uint, _ []model.StatusHistory) error {
	return nil
}
func (m *mockMainItemRepoForMap) NextCode(_ context.Context, _ int64) (string, error) {
	return "T001", nil
}
func (m *mockMainItemRepoForMap) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepoForMap) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}

func (m *mockMainItemRepoForMap) FindByMilestoneKey(_ context.Context, bizKey int64) ([]model.MainItem, error) {
	if m.findByMilestoneKeyErr != nil {
		return nil, m.findByMilestoneKeyErr
	}
	return m.items, nil
}

func (m *mockMainItemRepoForMap) CountByMilestoneKey(_ context.Context, _ int64) (int64, error) {
	return int64(len(m.items)), nil
}

func (m *mockMainItemRepoForMap) ClearMilestoneKeyByMilestone(_ context.Context, bizKey int64) error {
	m.clearedByMilestoneBizKey = bizKey
	return m.clearByMilestoneErr
}

func (m *mockMainItemRepoForMap) ClearMilestoneKeyByMap(_ context.Context, bizKeys []int64) error {
	m.clearedByMapBizKeys = bizKeys
	return m.clearByMapErr
}

// mockDBTxForMap captures transaction callback execution.
type mockDBTxForMap struct {
	txFunc func(fc func(tx *gorm.DB) error) error
}

func (m *mockDBTxForMap) Transaction(fc func(tx *gorm.DB) error, _ ...*sql.TxOptions) error {
	return m.txFunc(fc)
}

// ---------------------------------------------------------------------------
// Helper: create a sample MilestoneMap
// ---------------------------------------------------------------------------

func sampleMilestoneMap(status string) *model.MilestoneMap {
	return &model.MilestoneMap{
		BaseModel:   model.BaseModel{ID: 1, BizKey: 100},
		TeamKey:     1,
		CreatorKey:  10,
		AssigneeKey: 20,
		MapName:     "Test Map",
		MapDesc:     "A test map",
		MapStatus:   status,
	}
}

func msStrPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestMilestoneMapCreate_Success(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	result, err := svc.Create(context.Background(), 1, 10, dto.MilestoneMapCreateReq{
		MapName:        "Release v2",
		MapDesc:        "Second release",
		AssigneeBizKey: "20",
	})
	require.NoError(t, err)
	assert.Equal(t, "Release v2", mapRepo.createdItem.MapName)
	assert.Equal(t, "planning", mapRepo.createdItem.MapStatus)
	assert.Equal(t, int64(1), mapRepo.createdItem.TeamKey)
	assert.Equal(t, int64(10), mapRepo.createdItem.CreatorKey)
	assert.Equal(t, int64(20), mapRepo.createdItem.AssigneeKey)
	assert.NotZero(t, mapRepo.createdItem.BizKey)
	assert.Equal(t, uint(1), result.ID)
}

func TestMilestoneMapCreate_InvalidAssignee(t *testing.T) {
	svc := NewMilestoneMapService(&mockMilestoneMapRepo{}, nil, nil, nil)

	_, err := svc.Create(context.Background(), 1, 10, dto.MilestoneMapCreateReq{
		MapName:        "Release v2",
		AssigneeBizKey: "invalid",
	})
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestMilestoneMapCreate_WithDates(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	planStart := "2026-01-01"
	expectedEnd := "2026-06-30"
	_, err := svc.Create(context.Background(), 1, 10, dto.MilestoneMapCreateReq{
		MapName:         "Dated Map",
		AssigneeBizKey:  "20",
		PlanStartDate:   &planStart,
		ExpectedEndDate: &expectedEnd,
	})
	require.NoError(t, err)
	assert.NotNil(t, mapRepo.createdItem.PlanStartDate)
	assert.NotNil(t, mapRepo.createdItem.ExpectedEndDate)
}

func TestMilestoneMapCreate_RepoError(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{createErr: errors.New("db error")}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	_, err := svc.Create(context.Background(), 1, 10, dto.MilestoneMapCreateReq{
		MapName:        "Test",
		AssigneeBizKey: "20",
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestMilestoneMapUpdate_Success(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Update(context.Background(), 1, 1, dto.MilestoneMapUpdateReq{
		MapName: msStrPtr("Updated Name"),
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", mapRepo.updatedFields["map_name"])
}

func TestMilestoneMapUpdate_PartialNilSkipped(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Update(context.Background(), 1, 1, dto.MilestoneMapUpdateReq{
		MapDesc: msStrPtr("New desc"),
	})
	require.NoError(t, err)
	assert.Nil(t, mapRepo.updatedFields["map_name"])
	assert.Equal(t, "New desc", mapRepo.updatedFields["map_desc"])
}

func TestMilestoneMapUpdate_NoFields(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Update(context.Background(), 1, 1, dto.MilestoneMapUpdateReq{})
	require.NoError(t, err)
	assert.Nil(t, mapRepo.updatedFields)
}

func TestMilestoneMapUpdate_NotFound(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Update(context.Background(), 1, 999, dto.MilestoneMapUpdateReq{MapName: msStrPtr("X")})
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestMilestoneMapUpdate_WrongTeam(t *testing.T) {
	m := sampleMilestoneMap("planning")
	m.TeamKey = 99
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Update(context.Background(), 1, 1, dto.MilestoneMapUpdateReq{MapName: msStrPtr("X")})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestMilestoneMapUpdate_DateRangeInvalid(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	m := sampleMilestoneMap("planning")
	m.PlanStartDate = &start
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	// Expected end date before existing plan start date
	endBefore := "2026-05-01"
	err := svc.Update(context.Background(), 1, 1, dto.MilestoneMapUpdateReq{
		ExpectedEndDate: &endBefore,
	})
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestMilestoneMapUpdate_BothDatesInRange(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	planStart := "2026-01-01"
	expectedEnd := "2026-06-30"
	err := svc.Update(context.Background(), 1, 1, dto.MilestoneMapUpdateReq{
		PlanStartDate:   &planStart,
		ExpectedEndDate: &expectedEnd,
	})
	require.NoError(t, err)
	assert.NotNil(t, mapRepo.updatedFields["plan_start_date"])
	assert.NotNil(t, mapRepo.updatedFields["expected_end_date"])
}

// ---------------------------------------------------------------------------
// Tests: Get / GetByBizKey
// ---------------------------------------------------------------------------

func TestMilestoneMapGet_Success(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	result, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, m, result)
}

func TestMilestoneMapGet_NotFound(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	_, err := svc.Get(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestMilestoneMapGetByBizKey_Success(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	result, err := svc.GetByBizKey(context.Background(), 100)
	require.NoError(t, err)
	assert.Equal(t, m, result)
}

func TestMilestoneMapGetByBizKey_NotFound(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{findByBizErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	_, err := svc.GetByBizKey(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestMilestoneMapList_Success(t *testing.T) {
	items := []model.MilestoneMap{*sampleMilestoneMap("planning"), *sampleMilestoneMap("reviewed")}
	mapRepo := &mockMilestoneMapRepo{items: items}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	result, err := svc.List(context.Background(), 1, dto.MilestoneMapFilter{}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
}

// ---------------------------------------------------------------------------
// Tests: Delete (BR-4)
// ---------------------------------------------------------------------------

func TestMilestoneMapDelete_Planning(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{}
	miRepo := &mockMainItemRepoForMap{}
	tx := &mockDBTxForMap{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewMilestoneMapService(mapRepo, msRepo, miRepo, tx)

	err := svc.Delete(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.True(t, msRepo.deletedByMap)
	assert.NotNil(t, mapRepo.deletedID)
}

func TestMilestoneMapDelete_Reviewed(t *testing.T) {
	m := sampleMilestoneMap("reviewed")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{}
	miRepo := &mockMainItemRepoForMap{}
	tx := &mockDBTxForMap{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewMilestoneMapService(mapRepo, msRepo, miRepo, tx)

	err := svc.Delete(context.Background(), 1, 1)
	require.NoError(t, err)
}

func TestMilestoneMapDelete_Ready(t *testing.T) {
	m := sampleMilestoneMap("ready")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{}
	miRepo := &mockMainItemRepoForMap{}
	tx := &mockDBTxForMap{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewMilestoneMapService(mapRepo, msRepo, miRepo, tx)

	err := svc.Delete(context.Background(), 1, 1)
	require.NoError(t, err)
}

func TestMilestoneMapDelete_ExecutingBlocked(t *testing.T) {
	m := sampleMilestoneMap("executing")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.ErrorIs(t, err, apperrors.ErrMapCannotDelete)
}

func TestMilestoneMapDelete_CompletedBlocked(t *testing.T) {
	m := sampleMilestoneMap("completed")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.ErrorIs(t, err, apperrors.ErrMapCannotDelete)
}

func TestMilestoneMapDelete_CancelledBlocked(t *testing.T) {
	m := sampleMilestoneMap("cancelled")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.ErrorIs(t, err, apperrors.ErrMapCannotDelete)
}

func TestMilestoneMapDelete_WrongTeam(t *testing.T) {
	m := sampleMilestoneMap("planning")
	m.TeamKey = 99
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestMilestoneMapDelete_CascadeClearsMIs(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{
		milestones: []model.Milestone{
			{BaseModel: model.BaseModel{BizKey: 200}, MilestoneMapKey: 100, MilestoneStatus: "not_started"},
		},
	}
	miRepo := &mockMainItemRepoForMap{}
	tx := &mockDBTxForMap{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewMilestoneMapService(mapRepo, msRepo, miRepo, tx)

	err := svc.Delete(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, []int64{200}, miRepo.clearedByMapBizKeys)
	assert.True(t, msRepo.deletedByMap)
}

func TestMilestoneMapDelete_NotFound(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	err := svc.Delete(context.Background(), 1, 999)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus (BR-2, BR-6)
// ---------------------------------------------------------------------------

func TestMilestoneMapChangeStatus_ValidTransition(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 1, "reviewed")
	require.NoError(t, err)
	assert.Equal(t, "reviewed", mapRepo.updatedFields["map_status"])
	assert.NotNil(t, result)
}

func TestMilestoneMapChangeStatus_InvalidTransition(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 1, "completed")
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
}

func TestMilestoneMapChangeStatus_SelfTransition(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 1, "planning")
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
}

func TestMilestoneMapChangeStatus_Completed_AllMilestonesTerminal(t *testing.T) {
	m := sampleMilestoneMap("executing")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{
		milestones: []model.Milestone{
			{BaseModel: model.BaseModel{BizKey: 200}, MilestoneStatus: "completed"},
			{BaseModel: model.BaseModel{BizKey: 201}, MilestoneStatus: "cancelled"},
		},
	}
	svc := NewMilestoneMapService(mapRepo, msRepo, nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 1, "completed")
	require.NoError(t, err)
	assert.Equal(t, "completed", mapRepo.updatedFields["map_status"])
	assert.NotNil(t, result)
}

func TestMilestoneMapChangeStatus_Completed_NonTerminalMilestones(t *testing.T) {
	m := sampleMilestoneMap("executing")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{
		milestones: []model.Milestone{
			{BaseModel: model.BaseModel{BizKey: 200}, MilestoneStatus: "completed"},
			{BaseModel: model.BaseModel{BizKey: 201}, MilestoneStatus: "in_progress"},
		},
	}
	svc := NewMilestoneMapService(mapRepo, msRepo, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 1, "completed")
	assert.ErrorIs(t, err, apperrors.ErrMapHasNonTerminalMilestones)
}

func TestMilestoneMapChangeStatus_Completed_EmptyMap(t *testing.T) {
	m := sampleMilestoneMap("executing")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{milestones: []model.Milestone{}}
	svc := NewMilestoneMapService(mapRepo, msRepo, nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 1, "completed")
	require.NoError(t, err)
	assert.Equal(t, "completed", mapRepo.updatedFields["map_status"])
	assert.NotNil(t, result)
}

func TestMilestoneMapChangeStatus_Cancelled_Cascades(t *testing.T) {
	m := sampleMilestoneMap("executing")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{
		milestones: []model.Milestone{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 200}, MilestoneMapKey: 100, MilestoneStatus: "in_progress"},
			{BaseModel: model.BaseModel{ID: 11, BizKey: 201}, MilestoneMapKey: 100, MilestoneStatus: "completed"},
		},
	}
	miRepo := &mockMainItemRepoForMap{}
	tx := &mockDBTxForMap{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewMilestoneMapService(mapRepo, msRepo, miRepo, tx)

	result, err := svc.ChangeStatus(context.Background(), 1, 1, "cancelled")
	require.NoError(t, err)
	assert.NotNil(t, result)
	// Non-terminal milestone should be cancelled
	assert.Equal(t, "cancelled", msRepo.updatedFields["milestone_status"])
	// All milestone bizKeys should have MI keys cleared
	assert.Equal(t, []int64{200, 201}, miRepo.clearedByMapBizKeys)
}

func TestMilestoneMapChangeStatus_WrongTeam(t *testing.T) {
	m := sampleMilestoneMap("planning")
	m.TeamKey = 99
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 1, "reviewed")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestMilestoneMapChangeStatus_NotFound(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	_, err := svc.ChangeStatus(context.Background(), 1, 999, "reviewed")
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Tests: AvailableTransitions (BR-2 filtering)
// ---------------------------------------------------------------------------

func TestMilestoneMapAvailableTransitions_Planning(t *testing.T) {
	m := sampleMilestoneMap("planning")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{milestones: []model.Milestone{}}
	svc := NewMilestoneMapService(mapRepo, msRepo, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"reviewed", "cancelled"}, transitions)
}

func TestMilestoneMapAvailableTransitions_Executing_AllTerminal(t *testing.T) {
	m := sampleMilestoneMap("executing")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{
		milestones: []model.Milestone{
			{MilestoneStatus: "completed"},
		},
	}
	svc := NewMilestoneMapService(mapRepo, msRepo, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Contains(t, transitions, "completed")
}

func TestMilestoneMapAvailableTransitions_Executing_NonTerminal_FilteredCompleted(t *testing.T) {
	m := sampleMilestoneMap("executing")
	mapRepo := &mockMilestoneMapRepo{item: m}
	msRepo := &mockMilestoneRepoForMap{
		milestones: []model.Milestone{
			{MilestoneStatus: "in_progress"},
		},
	}
	svc := NewMilestoneMapService(mapRepo, msRepo, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.NotContains(t, transitions, "completed")
	assert.Contains(t, transitions, "ready")
	assert.Contains(t, transitions, "cancelled")
}

func TestMilestoneMapAvailableTransitions_Terminal(t *testing.T) {
	m := sampleMilestoneMap("completed")
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Empty(t, transitions)
}

func TestMilestoneMapAvailableTransitions_WrongTeam(t *testing.T) {
	m := sampleMilestoneMap("planning")
	m.TeamKey = 99
	mapRepo := &mockMilestoneMapRepo{item: m}
	svc := NewMilestoneMapService(mapRepo, nil, nil, nil)

	_, err := svc.AvailableTransitions(context.Background(), 1, 1)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}
