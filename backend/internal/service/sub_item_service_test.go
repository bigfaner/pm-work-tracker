package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock repos using testify/mock for SubItemService tests
// ---------------------------------------------------------------------------

// mockSubItemRepoTM uses testify/mock to satisfy repository.SubItemRepo.
type mockSubItemRepoTM struct {
	mock.Mock
}

func (m *mockSubItemRepoTM) Create(ctx context.Context, item *model.SubItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *mockSubItemRepoTM) FindByID(ctx context.Context, id uint) (*model.SubItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SubItem), args.Error(1)
}

func (m *mockSubItemRepoTM) Update(ctx context.Context, item *model.SubItem, fields map[string]interface{}) error {
	args := m.Called(ctx, item, fields)
	return args.Error(0)
}

func (m *mockSubItemRepoTM) List(ctx context.Context, teamID uint, mainItemID uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	args := m.Called(ctx, teamID, mainItemID, filter, page)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PageResult[model.SubItem]), args.Error(1)
}

func (m *mockSubItemRepoTM) ListByMainItem(ctx context.Context, mainItemID uint) ([]*model.SubItem, error) {
	args := m.Called(ctx, mainItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SubItem), args.Error(1)
}

func (m *mockSubItemRepoTM) ListByTeam(ctx context.Context, teamID uint) ([]model.SubItem, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.SubItem), args.Error(1)
}

// mockMainItemSvcTM uses testify/mock to satisfy MainItemService.
type mockMainItemSvcTM struct {
	mock.Mock
}

func (m *mockMainItemSvcTM) Create(ctx context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error) {
	args := m.Called(ctx, teamID, pmID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MainItem), args.Error(1)
}

func (m *mockMainItemSvcTM) Update(ctx context.Context, teamID, itemID uint, req dto.MainItemUpdateReq) error {
	args := m.Called(ctx, teamID, itemID, req)
	return args.Error(0)
}

func (m *mockMainItemSvcTM) Archive(ctx context.Context, teamID, itemID uint) error {
	args := m.Called(ctx, teamID, itemID)
	return args.Error(0)
}

func (m *mockMainItemSvcTM) List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	args := m.Called(ctx, teamID, filter, page)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PageResult[model.MainItem]), args.Error(1)
}

func (m *mockMainItemSvcTM) Get(ctx context.Context, itemID uint) (*model.MainItem, error) {
	args := m.Called(ctx, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MainItem), args.Error(1)
}

func (m *mockMainItemSvcTM) RecalcCompletion(ctx context.Context, mainItemID uint) error {
	args := m.Called(ctx, mainItemID)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestSubItemCreate_Success(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(item *model.SubItem) bool {
		return item.TeamID == 1 && item.MainItemID == 5 && item.Title == "Sub task A" && item.Status == "pending"
	})).Return(nil)

	item, err := svc.Create(context.Background(), 1, 10, dto.SubItemCreateReq{
		AssigneeID:      42,
		MainItemID: 5,
		Title:      "Sub task A",
		Priority:   "P2",
	})
	require.NoError(t, err)
	assert.Equal(t, uint(1), item.TeamID)
	assert.Equal(t, uint(5), item.MainItemID)
	assert.Equal(t, "pending", item.Status)
	assert.Equal(t, "Sub task A", item.Title)
	assert.Equal(t, "P2", item.Priority)
	assert.NotNil(t, item.AssigneeID)
	assert.Equal(t, uint(42), *item.AssigneeID)

	repo.AssertExpectations(t)
}

func TestSubItemCreate_RepoError(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, err := svc.Create(context.Background(), 1, 10, dto.SubItemCreateReq{
		MainItemID: 5,
		Title:      "Sub task",
		Priority:   "P2",
	})
	assert.Error(t, err)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestSubItemUpdate_Success(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		Title:  "Old Title",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	repo.On("Update", mock.Anything, existing, mock.MatchedBy(func(fields map[string]interface{}) bool {
		return fields["title"] == "New Title"
	})).Return(nil)

	err := svc.Update(context.Background(), 1, 1, dto.SubItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestSubItemUpdate_TeamMismatch(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 2,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	err := svc.Update(context.Background(), 1, 1, dto.SubItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)

	repo.AssertExpectations(t)
}

func TestSubItemUpdate_NotFound(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.Update(context.Background(), 1, 99, dto.SubItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)

	repo.AssertExpectations(t)
}

func TestSubItemUpdate_NoFields_Noop(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	// Update should NOT be called when no fields are provided.

	err := svc.Update(context.Background(), 1, 1, dto.SubItemUpdateReq{})
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — valid transitions
// ---------------------------------------------------------------------------

func TestChangeStatus_待开始_to_进行中(t *testing.T) {
	testValidTransitionTM(t, "待开始", "进行中")
}

func TestChangeStatus_待开始_to_已关闭(t *testing.T) {
	testValidTransitionTM(t, "待开始", "已关闭")
}

func TestChangeStatus_进行中_to_阻塞中(t *testing.T) {
	testValidTransitionTM(t, "进行中", "阻塞中")
}

func TestChangeStatus_进行中_to_挂起(t *testing.T) {
	testValidTransitionTM(t, "进行中", "挂起")
}

func TestChangeStatus_进行中_to_待验收(t *testing.T) {
	testValidTransitionTM(t, "进行中", "待验收")
}

func TestChangeStatus_进行中_to_已延期(t *testing.T) {
	testValidTransitionTM(t, "进行中", "已延期")
}

func TestChangeStatus_进行中_to_已关闭(t *testing.T) {
	testValidTransitionTM(t, "进行中", "已关闭")
}

func TestChangeStatus_阻塞中_to_进行中(t *testing.T) {
	testValidTransitionTM(t, "阻塞中", "进行中")
}

func TestChangeStatus_挂起_to_进行中(t *testing.T) {
	testValidTransitionTM(t, "挂起", "进行中")
}

func TestChangeStatus_挂起_to_已关闭(t *testing.T) {
	testValidTransitionTM(t, "挂起", "已关闭")
}

func TestChangeStatus_已延期_to_进行中(t *testing.T) {
	testValidTransitionTM(t, "已延期", "进行中")
}

func TestChangeStatus_待验收_to_已完成(t *testing.T) {
	testValidTransitionTM(t, "待验收", "已完成")
}

func TestChangeStatus_待验收_to_进行中(t *testing.T) {
	testValidTransitionTM(t, "待验收", "进行中")
}

func testValidTransitionTM(t *testing.T, from, to string) {
	t.Helper()
	existing := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     from,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	repo.On("Update", mock.Anything, existing, mock.MatchedBy(func(fields map[string]interface{}) bool {
		return fields["status"] == to
	})).Return(nil)

	// If transitioning to 已完成, RecalcCompletion will be called.
	if to == "已完成" {
		mainSvc.On("RecalcCompletion", mock.Anything, uint(5)).Return(nil)
	}

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, to)
	require.NoError(t, err)

	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — invalid transitions
// ---------------------------------------------------------------------------

func TestChangeStatus_Invalid_已完成_to_anything(t *testing.T) {
	for _, target := range []string{"待开始", "进行中", "阻塞中", "挂起", "已延期", "待验收", "已关闭"} {
		t.Run("已完成->"+target, func(t *testing.T) {
			existing := &model.SubItem{
				BaseModel:  model.BaseModel{ID: 1},
				TeamID: 1,
				Status: "已完成",
			}
			repo := new(mockSubItemRepoTM)
			mainSvc := new(mockMainItemSvcTM)
			svc := NewSubItemService(repo, mainSvc)

			repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

			err := svc.ChangeStatus(context.Background(), 1, 10, 1, target)
			assert.ErrorIs(t, err, apperrors.ErrInvalidStatus, "from 已完成 to %s should be invalid", target)

			repo.AssertExpectations(t)
		})
	}
}

func TestChangeStatus_Invalid_已关闭_to_anything(t *testing.T) {
	for _, target := range []string{"待开始", "进行中", "阻塞中", "挂起", "已延期", "待验收", "已完成"} {
		t.Run("已关闭->"+target, func(t *testing.T) {
			existing := &model.SubItem{
				BaseModel:  model.BaseModel{ID: 1},
				TeamID: 1,
				Status: "已关闭",
			}
			repo := new(mockSubItemRepoTM)
			mainSvc := new(mockMainItemSvcTM)
			svc := NewSubItemService(repo, mainSvc)

			repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

			err := svc.ChangeStatus(context.Background(), 1, 10, 1, target)
			assert.ErrorIs(t, err, apperrors.ErrInvalidStatus, "from 已关闭 to %s should be invalid", target)

			repo.AssertExpectations(t)
		})
	}
}

func TestChangeStatus_Invalid_待开始_to_待验收(t *testing.T) {
	testInvalidTransitionTM(t, "待开始", "待验收")
}

func TestChangeStatus_Invalid_待开始_to_已延期(t *testing.T) {
	testInvalidTransitionTM(t, "待开始", "已延期")
}

func TestChangeStatus_Invalid_进行中_to_待开始(t *testing.T) {
	testInvalidTransitionTM(t, "进行中", "待开始")
}

func TestChangeStatus_Invalid_阻塞中_to_挂起(t *testing.T) {
	testInvalidTransitionTM(t, "阻塞中", "挂起")
}

func TestChangeStatus_SameStatus(t *testing.T) {
	testInvalidTransitionTM(t, "进行中", "进行中")
}

func testInvalidTransitionTM(t *testing.T, from, to string) {
	t.Helper()
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		Status: from,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, to)
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus, "from %s to %s should be invalid", from, to)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — 已完成 sets ActualEndDate and calls RecalcCompletion
// ---------------------------------------------------------------------------

func TestChangeStatus_已完成_SetsActualEndDate(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     "待验收",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	repo.On("Update", mock.Anything, existing, mock.MatchedBy(func(fields map[string]interface{}) bool {
		return fields["status"] == "已完成" && fields["actual_end_date"] != nil
	})).Return(nil)
	mainSvc.On("RecalcCompletion", mock.Anything, uint(5)).Return(nil)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "已完成")
	require.NoError(t, err)

	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
}

func TestChangeStatus_已完成_RecalcCompletion_CalledWithCorrectMainItemID(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 42,
		Status:     "待验收",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// Verify RecalcCompletion is called with the correct MainItemID.
	mainSvc.On("RecalcCompletion", mock.Anything, uint(42)).Return(nil)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "已完成")
	require.NoError(t, err)

	mainSvc.AssertCalled(t, "RecalcCompletion", mock.Anything, uint(42))
	mainSvc.AssertExpectations(t)
}

func TestChangeStatus_已完成_RecalcError(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     "待验收",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mainSvc.On("RecalcCompletion", mock.Anything, uint(5)).Return(errors.New("recalc failed"))

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "已完成")
	assert.Error(t, err)

	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — not found / team mismatch
// ---------------------------------------------------------------------------

func TestChangeStatus_NotFound(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.ChangeStatus(context.Background(), 1, 10, 99, "进行中")
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)

	repo.AssertExpectations(t)
}

func TestChangeStatus_TeamMismatch(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 2,
		Status: "待开始",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "进行中")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: Assign
// ---------------------------------------------------------------------------

func TestSubItemAssign_Success(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	var assigneeID uint = 42
	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	repo.On("Update", mock.Anything, existing, mock.MatchedBy(func(fields map[string]interface{}) bool {
		return fields["assignee_id"] == assigneeID
	})).Return(nil)

	err := svc.Assign(context.Background(), 1, 10, 1, assigneeID)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestSubItemAssign_NotFound(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.Assign(context.Background(), 1, 10, 99, 42)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)

	repo.AssertExpectations(t)
}

func TestSubItemAssign_TeamMismatch(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 2,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	err := svc.Assign(context.Background(), 1, 10, 1, 42)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestSubItemGet_Success(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		Title:  "Sub 1",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	item, err := svc.Get(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, "Sub 1", item.Title)

	repo.AssertExpectations(t)
}

func TestSubItemGet_NotFound(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("FindByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.Get(context.Background(), 1, 99)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestSubItemList_Success(t *testing.T) {
	items := []model.SubItem{
		{BaseModel: model.BaseModel{ID: 1}, Title: "Sub 1"},
		{BaseModel: model.BaseModel{ID: 2}, Title: "Sub 2"},
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("List", mock.Anything, uint(1), uint(0), mock.Anything, mock.Anything).
		Return(&dto.PageResult[model.SubItem]{Items: items, Total: 2}, nil)

	result, err := svc.List(context.Background(), 1, nil, dto.SubItemFilter{}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)

	repo.AssertExpectations(t)
}

func TestSubItemList_WithMainItemFilter(t *testing.T) {
	items := []model.SubItem{
		{BaseModel: model.BaseModel{ID: 1}, Title: "Sub 1"},
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	mainID := uint(5)
	repo.On("List", mock.Anything, uint(1), uint(5), mock.Anything, mock.Anything).
		Return(&dto.PageResult[model.SubItem]{Items: items, Total: 1}, nil)

	result, err := svc.List(context.Background(), 1, &mainID, dto.SubItemFilter{}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)

	repo.AssertExpectations(t)
}

func TestSubItemList_RepoError(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	svc := NewSubItemService(repo, mainSvc)

	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	_, err := svc.List(context.Background(), 1, nil, dto.SubItemFilter{}, dto.Pagination{})
	assert.Error(t, err)

	repo.AssertExpectations(t)
}
