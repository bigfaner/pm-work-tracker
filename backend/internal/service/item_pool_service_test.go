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

func uintPtr(v uint) *uint { return &v }

// ---------------------------------------------------------------------------
// Mock repos for ItemPoolService tests
// ---------------------------------------------------------------------------

type mockItemPoolRepo struct {
	item       *model.ItemPool
	items      []model.ItemPool
	pageResult *dto.PageResult[model.ItemPool]

	findErr   error
	createErr error
	updateErr error
	listErr   error

	createdItem   *model.ItemPool
	updatedItem   *model.ItemPool
	updatedFields map[string]interface{}
}

func (m *mockItemPoolRepo) Create(_ context.Context, item *model.ItemPool) error {
	m.createdItem = item
	if m.createErr != nil {
		return m.createErr
	}
	item.ID = 1
	return nil
}

func (m *mockItemPoolRepo) FindByID(_ context.Context, id uint) (*model.ItemPool, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findErr
}

func (m *mockItemPoolRepo) FindByBizKey(_ context.Context, bizKey int64) (*model.ItemPool, error) {
	if m.item != nil && m.item.BizKey == bizKey {
		return m.item, nil
	}
	return nil, m.findErr
}

func (m *mockItemPoolRepo) Update(_ context.Context, item *model.ItemPool, fields map[string]interface{}) error {
	m.updatedItem = item
	m.updatedFields = fields
	return m.updateErr
}

func (m *mockItemPoolRepo) List(_ context.Context, teamBizKey int64, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.pageResult != nil {
		return m.pageResult, nil
	}
	return &dto.PageResult[model.ItemPool]{Items: m.items, Total: int64(len(m.items))}, nil
}

// mockSubItemRepoForPool captures Create calls from Assign.
type mockSubItemRepoForPool struct {
	createdItem *model.SubItem
	createErr   error
}

func (m *mockSubItemRepoForPool) Create(_ context.Context, item *model.SubItem) error {
	m.createdItem = item
	if m.createErr != nil {
		return m.createErr
	}
	item.ID = 10
	item.BizKey = 100
	return nil
}
func (m *mockSubItemRepoForPool) FindByID(_ context.Context, id uint) (*model.SubItem, error) {
	return nil, nil
}
func (m *mockSubItemRepoForPool) Update(_ context.Context, item *model.SubItem, fields map[string]interface{}) error {
	return nil
}
func (m *mockSubItemRepoForPool) List(_ context.Context, teamBizKey int64, mainItemID uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}
func (m *mockSubItemRepoForPool) ListByMainItem(_ context.Context, mainItemID uint) ([]*model.SubItem, error) {
	return nil, nil
}
func (m *mockSubItemRepoForPool) ListByTeam(_ context.Context, _ int64) ([]model.SubItem, error) {
	return nil, nil
}
func (m *mockSubItemRepoForPool) SoftDelete(_ context.Context, _ uint) error {
	return nil
}
func (m *mockSubItemRepoForPool) FindByBizKey(_ context.Context, _ int64) (*model.SubItem, error) {
	return nil, nil
}
func (m *mockSubItemRepoForPool) NextSubCode(_ context.Context, _ uint) (string, error) {
	return "", nil
}

// mockMainItemRepoForPool captures FindByID for Assign validation.
type mockMainItemRepoForPool struct {
	item         *model.MainItem
	findErr      error
	nextCodeVal  string
	nextCodeErr  error
	createdItem  *model.MainItem
}

func (m *mockMainItemRepoForPool) Create(_ context.Context, item *model.MainItem) error {
	m.createdItem = item
	item.ID = 2
	item.BizKey = 999
	return nil
}
func (m *mockMainItemRepoForPool) FindByID(_ context.Context, id uint) (*model.MainItem, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findErr
}
func (m *mockMainItemRepoForPool) Update(_ context.Context, item *model.MainItem, fields map[string]interface{}) error {
	return nil
}
func (m *mockMainItemRepoForPool) List(_ context.Context, teamBizKey int64, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemRepoForPool) NextCode(_ context.Context, teamBizKey int64) (string, error) {
	return m.nextCodeVal, m.nextCodeErr
}

func (m *mockMainItemRepoForPool) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

func (m *mockMainItemRepoForPool) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForPool) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForPool) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForPool) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findErr
}
func (m *mockMainItemRepoForPool) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}

// mockDBTx captures transaction callback execution.
type mockDBTx struct {
	txFunc func(fc func(tx *gorm.DB) error) error
}

func (m *mockDBTx) Transaction(fc func(tx *gorm.DB) error, _ ...*sql.TxOptions) error {
	return m.txFunc(fc)
}

// ---------------------------------------------------------------------------
// Tests: Submit
// ---------------------------------------------------------------------------

func TestItemPoolSubmit_Success(t *testing.T) {
	poolRepo := &mockItemPoolRepo{}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	item, err := svc.Submit(context.Background(), int64(1), 10, dto.SubmitItemPoolReq{
		Title:          "Optimize homepage",
		Background:     "Users complain about slow load",
		ExpectedOutput: "LCP < 1.5s",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), poolRepo.createdItem.TeamKey)
	assert.Equal(t, int64(10), poolRepo.createdItem.SubmitterKey)
	assert.Equal(t, "pending", poolRepo.createdItem.PoolStatus)
	assert.Equal(t, "Optimize homepage", poolRepo.createdItem.Title)
	assert.Equal(t, "Users complain about slow load", poolRepo.createdItem.Background)
	assert.Equal(t, "LCP < 1.5s", poolRepo.createdItem.ExpectedOutput)
	assert.Equal(t, uint(1), item.ID)
}

func TestItemPoolSubmit_RepoError(t *testing.T) {
	poolRepo := &mockItemPoolRepo{createErr: errors.New("db error")}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.Submit(context.Background(), int64(1), 10, dto.SubmitItemPoolReq{
		Title: "Test",
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Assign — success
// ---------------------------------------------------------------------------

func TestItemPoolAssign_Success(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:   model.BaseModel{ID: 5},
		TeamKey: 1,
		Title:   "Pool item",
		PoolStatus: "pending",
		SubmitterKey: 10,
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	subRepo := &mockSubItemRepoForPool{}
	mainRepo := &mockMainItemRepoForPool{item: &model.MainItem{BaseModel: model.BaseModel{ID: 20, BizKey: 200}, TeamKey: 1}}

	txExecuted := false
	dbtx := &mockDBTx{txFunc: func(fc func(tx *gorm.DB) error) error {
		txExecuted = true
		return fc(nil)
	}}
	svc := NewItemPoolService(poolRepo, subRepo, mainRepo, dbtx)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	require.NoError(t, err)
	assert.True(t, txExecuted, "should execute within a transaction")

	// Verify pool item was updated
	assert.Equal(t, "assigned", poolRepo.updatedFields["pool_status"])
	assert.Equal(t, int64(200), poolRepo.updatedFields["assigned_main_key"])
	assert.Equal(t, int64(100), poolRepo.updatedFields["assigned_sub_key"]) // SubItem.BizKey set to 100 by mock
	assert.Equal(t, func() *int64 { v := int64(30); return &v }(), poolRepo.updatedFields["assignee_key"])
	assert.Equal(t, uint(100), poolRepo.updatedFields["reviewer_key"])
	assert.NotNil(t, poolRepo.updatedFields["reviewed_at"])

	// Verify SubItem was created
	assert.Equal(t, int64(1), subRepo.createdItem.TeamKey)
	assert.Equal(t, uint(20), uint(subRepo.createdItem.MainItemKey))
	assert.Equal(t, "Pool item", subRepo.createdItem.Title)
	assert.Equal(t, "pending", subRepo.createdItem.ItemStatus)
	assert.Equal(t, int64(30), *subRepo.createdItem.AssigneeKey)
}

func TestItemPoolAssign_SubItemInheritsBackgroundAsDescription(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:      model.BaseModel{ID: 5},
		TeamKey: 1,
		Title:      "Pool item",
		Background: "some background info",
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	subRepo := &mockSubItemRepoForPool{}
	mainRepo := &mockMainItemRepoForPool{item: &model.MainItem{BaseModel: model.BaseModel{ID: 20, BizKey: 200}, TeamKey: 1}}
	dbtx := &mockDBTx{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewItemPoolService(poolRepo, subRepo, mainRepo, dbtx)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	require.NoError(t, err)
	assert.Equal(t, "some background info", subRepo.createdItem.ItemDesc)
}

// ---------------------------------------------------------------------------
// Tests: Assign — errors
// ---------------------------------------------------------------------------

func TestItemPoolAssign_PoolItemNotFound(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Assign(context.Background(), int64(1), 100, 99, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestItemPoolAssign_AlreadyProcessed_AlreadyAssigned(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		PoolStatus: "assigned",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemAlreadyProcessed)
}

func TestItemPoolAssign_AlreadyProcessed_Rejected(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		PoolStatus: "rejected",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemAlreadyProcessed)
}

func TestItemPoolAssign_MainItemNotFound(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	mainRepo := &mockMainItemRepoForPool{findErr: gorm.ErrRecordNotFound}
	dbtx := &mockDBTx{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewItemPoolService(poolRepo, nil, mainRepo, dbtx)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "99",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestItemPoolAssign_TeamMismatch(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 2,
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: Assign — rollback on SubItem creation failure
// ---------------------------------------------------------------------------

func TestItemPoolAssign_RollbackOnSubItemCreateError(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:   model.BaseModel{ID: 5},
		TeamKey: 1,
		Title:   "Pool item",
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	subRepo := &mockSubItemRepoForPool{createErr: errors.New("sub item creation failed")}
	mainRepo := &mockMainItemRepoForPool{item: &model.MainItem{BaseModel: model.BaseModel{ID: 20, BizKey: 200}, TeamKey: 1}}

	txCallbackExecuted := false
	dbtx := &mockDBTx{txFunc: func(fc func(tx *gorm.DB) error) error {
		txCallbackExecuted = true
		return fc(nil)
	}}
	svc := NewItemPoolService(poolRepo, subRepo, mainRepo, dbtx)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.Error(t, err)
	assert.True(t, txCallbackExecuted, "should have attempted the transaction")
	assert.Contains(t, err.Error(), "sub item creation failed")
}

// ---------------------------------------------------------------------------
// Tests: Reject — success
// ---------------------------------------------------------------------------

func TestItemPoolReject_Success(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:   model.BaseModel{ID: 5},
		TeamKey: 1,
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Reject(context.Background(), int64(1), 100, 5, "Not enough priority")
	require.NoError(t, err)

	assert.Equal(t, "rejected", poolRepo.updatedFields["pool_status"])
	assert.Equal(t, "Not enough priority", poolRepo.updatedFields["reject_reason"])
	assert.Equal(t, uint(100), poolRepo.updatedFields["reviewer_key"])
	assert.NotNil(t, poolRepo.updatedFields["reviewed_at"])
}

func TestItemPoolReject_AlreadyProcessed(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		PoolStatus: "assigned",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Reject(context.Background(), int64(1), 100, 5, "reason")
	assert.ErrorIs(t, err, apperrors.ErrItemAlreadyProcessed)
}

func TestItemPoolReject_NotFound(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Reject(context.Background(), int64(1), 100, 99, "reason")
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestItemPoolReject_TeamMismatch(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 2,
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Reject(context.Background(), int64(1), 100, 5, "reason")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestItemPoolGet_Success(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		Title:  "My Pool Item",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	item, err := svc.Get(context.Background(), int64(1), 5)
	require.NoError(t, err)
	assert.Equal(t, "My Pool Item", item.Title)
}

func TestItemPoolGet_NotFound(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.Get(context.Background(), int64(1), 99)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestItemPoolGet_TeamMismatch(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 2,
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.Get(context.Background(), int64(1), 5)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestItemPoolList_Success(t *testing.T) {
	items := []model.ItemPool{
		{BaseModel: model.BaseModel{ID: 1}, Title: "Pool 1"},
		{BaseModel: model.BaseModel{ID: 2}, Title: "Pool 2"},
	}
	poolRepo := &mockItemPoolRepo{items: items}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	result, err := svc.List(context.Background(), int64(1), dto.ItemPoolFilter{}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)
}

func TestItemPoolList_WithStatusFilter(t *testing.T) {
	items := []model.ItemPool{
		{BaseModel: model.BaseModel{ID: 1}, Title: "Pool 1", PoolStatus: "pending"},
	}
	poolRepo := &mockItemPoolRepo{items: items}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	result, err := svc.List(context.Background(), int64(1), dto.ItemPoolFilter{Status: "pending"}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
}

func TestItemPoolList_RepoError(t *testing.T) {
	poolRepo := &mockItemPoolRepo{listErr: errors.New("db error")}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.List(context.Background(), int64(1), dto.ItemPoolFilter{}, dto.Pagination{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Assign — reviewed_at is set to current time
// ---------------------------------------------------------------------------

func TestItemPoolAssign_ReviewedAtIsSet(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		Title:  "Pool item",
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	subRepo := &mockSubItemRepoForPool{}
	mainRepo := &mockMainItemRepoForPool{item: &model.MainItem{BaseModel: model.BaseModel{ID: 20, BizKey: 200}, TeamKey: 1}}
	dbtx := &mockDBTx{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewItemPoolService(poolRepo, subRepo, mainRepo, dbtx)

	before := time.Now()
	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	require.NoError(t, err)

	reviewedAt, ok := poolRepo.updatedFields["reviewed_at"].(time.Time)
	require.True(t, ok, "reviewed_at should be time.Time")
	assert.False(t, reviewedAt.Before(before), "reviewed_at should be >= time before call")
}

// ---------------------------------------------------------------------------
// Tests: Reject — reviewed_at is set to current time
// ---------------------------------------------------------------------------

func TestItemPoolReject_ReviewedAtIsSet(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	before := time.Now()
	err := svc.Reject(context.Background(), int64(1), 100, 5, "reason")
	require.NoError(t, err)

	reviewedAt, ok := poolRepo.updatedFields["reviewed_at"].(time.Time)
	require.True(t, ok, "reviewed_at should be time.Time")
	assert.False(t, reviewedAt.Before(before), "reviewed_at should be >= time before call")
}

// ---------------------------------------------------------------------------
// Tests: Assign — main item team mismatch
// ---------------------------------------------------------------------------

func TestItemPoolAssign_MainItemTeamMismatch(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	mainRepo := &mockMainItemRepoForPool{item: &model.MainItem{BaseModel: model.BaseModel{ID: 20, BizKey: 200}, TeamKey: 99}}
	dbtx := &mockDBTx{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewItemPoolService(poolRepo, nil, mainRepo, dbtx)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

// ---------------------------------------------------------------------------
// Tests: error mapping — non-ErrRecordNotFound errors pass through
// ---------------------------------------------------------------------------

func TestItemPoolAssign_PoolItemGenericError(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: errors.New("generic db error")}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.Equal(t, "generic db error", err.Error())
}

func TestItemPoolAssign_MainItemGenericError(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	mainRepo := &mockMainItemRepoForPool{findErr: errors.New("main item db error")}
	dbtx := &mockDBTx{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewItemPoolService(poolRepo, nil, mainRepo, dbtx)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.Equal(t, "main item db error", err.Error())
}

func TestItemPoolGet_GenericFindError(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: errors.New("generic db error")}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.Get(context.Background(), 1, 5)
	assert.Equal(t, "generic db error", err.Error())
}

func TestItemPoolReject_GenericFindError(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: errors.New("generic db error")}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Reject(context.Background(), 1, 100, 5, "reason")
	assert.Equal(t, "generic db error", err.Error())
}

// ---------------------------------------------------------------------------
// Tests: error mapping — ErrNotFound from repo
// ---------------------------------------------------------------------------

func TestItemPoolAssign_PoolItemErrNotFound(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: apperrors.ErrNotFound}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestItemPoolAssign_MainItemErrNotFound(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey: 1,
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	mainRepo := &mockMainItemRepoForPool{findErr: apperrors.ErrNotFound}
	dbtx := &mockDBTx{txFunc: func(fc func(tx *gorm.DB) error) error { return fc(nil) }}
	svc := NewItemPoolService(poolRepo, nil, mainRepo, dbtx)

	err := svc.Assign(context.Background(), int64(1), 100, 5, dto.AssignItemPoolReq{
		MainItemKey: "20",
		AssigneeKey: strPtr("30"),
			StartDate: strPtr("2024-01-01"),
			ExpectedEndDate: strPtr("2024-03-01"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestItemPoolGet_ErrNotFound(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: apperrors.ErrNotFound}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.Get(context.Background(), 1, 5)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestItemPoolReject_ErrNotFound(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: apperrors.ErrNotFound}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	err := svc.Reject(context.Background(), 1, 100, 5, "reason")
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

// ---------------------------------------------------------------------------
// Tests: ConvertToMain
// ---------------------------------------------------------------------------

func TestItemPoolConvertToMain_Success(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:   model.BaseModel{ID: 5},
		TeamKey:     1,
		Title:       "Pool item",
		PoolStatus:  "pending",
		SubmitterKey: 10,
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	mainRepo := &mockMainItemRepoForPool{nextCodeVal: "MI-0001"}

	txExecuted := false
	dbtx := &mockDBTx{txFunc: func(fc func(tx *gorm.DB) error) error {
		txExecuted = true
		return fc(nil)
	}}
	svc := NewItemPoolService(poolRepo, nil, mainRepo, dbtx)

	result, err := svc.ConvertToMain(context.Background(), int64(1), 100, 5, dto.ConvertToMainItemReq{
		Priority: "P1",
	})
	require.NoError(t, err)
	assert.True(t, txExecuted, "should execute within a transaction")
	require.NotNil(t, result)
	assert.Equal(t, "MI-0001", result.Code)
	assert.Equal(t, "Pool item", result.Title)
	assert.Equal(t, "P1", result.Priority)
	assert.Equal(t, int64(1), result.TeamKey)
	assert.Equal(t, "pending", result.ItemStatus)

	// Verify pool item was updated to "assigned"
	assert.Equal(t, "assigned", poolRepo.updatedFields["pool_status"])

	// Verify main item was created
	assert.NotNil(t, mainRepo.createdItem)
}

func TestItemPoolConvertToMain_PoolItemNotFound(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.ConvertToMain(context.Background(), int64(1), 100, 99, dto.ConvertToMainItemReq{
		Priority: "P1",
	})
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestItemPoolConvertToMain_TeamMismatch(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey:    2,
		PoolStatus: "pending",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.ConvertToMain(context.Background(), int64(1), 100, 5, dto.ConvertToMainItemReq{
		Priority: "P1",
	})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestItemPoolConvertToMain_AlreadyProcessed(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel:  model.BaseModel{ID: 5},
		TeamKey:    1,
		PoolStatus: "assigned",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.ConvertToMain(context.Background(), int64(1), 100, 5, dto.ConvertToMainItemReq{
		Priority: "P1",
	})
	assert.ErrorIs(t, err, apperrors.ErrItemAlreadyProcessed)
}

// ---------------------------------------------------------------------------
// Tests: GetByBizKey
// ---------------------------------------------------------------------------

func TestItemPoolGetByBizKey_Found(t *testing.T) {
	poolItem := &model.ItemPool{
		BaseModel: model.BaseModel{BizKey: 123456},
		Title:     "My Pool Item",
	}
	poolRepo := &mockItemPoolRepo{item: poolItem}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	item, err := svc.GetByBizKey(context.Background(), 123456)
	require.NoError(t, err)
	assert.Equal(t, "My Pool Item", item.Title)
}

func TestItemPoolGetByBizKey_NotFound(t *testing.T) {
	poolRepo := &mockItemPoolRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewItemPoolService(poolRepo, nil, nil, nil)

	_, err := svc.GetByBizKey(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

