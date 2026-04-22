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

func (m *mockSubItemRepoTM) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSubItemRepoTM) NextSubCode(ctx context.Context, mainItemID uint) (string, error) {
	args := m.Called(ctx, mainItemID)
	return args.String(0), args.Error(1)
}
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

func (m *mockMainItemSvcTM) ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*model.MainItem, error) {
	args := m.Called(ctx, teamID, callerID, itemID, newStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MainItem), args.Error(1)
}

func (m *mockMainItemSvcTM) AvailableTransitions(ctx context.Context, teamID, callerID, itemID uint) ([]string, error) {
	args := m.Called(ctx, teamID, callerID, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockMainItemSvcTM) EvaluateLinkage(ctx context.Context, mainItemID uint, changedBy uint) (*LinkageResult, error) {
	args := m.Called(ctx, mainItemID, changedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LinkageResult), args.Error(1)
}

// mockStatusHistorySvcTM uses testify/mock to satisfy StatusHistoryService.
type mockStatusHistorySvcTM struct {
	mock.Mock
}

func (m *mockStatusHistorySvcTM) Record(ctx context.Context, record *model.StatusHistory) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockStatusHistorySvcTM) ListByItem(ctx context.Context, itemType string, itemID uint, page dto.Pagination) (*dto.PageResult[model.StatusHistory], error) {
	args := m.Called(ctx, itemType, itemID, page)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PageResult[model.StatusHistory]), args.Error(1)
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestSubItemCreate_Success(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(item *model.SubItem) bool {
		return item.TeamID == 1 && item.MainItemID == 5 && item.Title == "Sub task A" && item.Status == "pending"
	})).Return(nil)
	repo.On("NextSubCode", mock.Anything, uint(5)).Return("FEAT-00001-01", nil)
	mainSvc.On("Get", mock.Anything, uint(5)).Return(&model.MainItem{BaseModel: model.BaseModel{ID: 5}, Status: "pending"}, nil)
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Return(nil, nil)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
	repo.On("NextSubCode", mock.Anything, uint(5)).Return("FEAT-00001-01", nil)
	mainSvc.On("Get", mock.Anything, uint(5)).Return(&model.MainItem{BaseModel: model.BaseModel{ID: 5}, Status: "pending"}, nil)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	// Update should NOT be called when no fields are provided.

	err := svc.Update(context.Background(), 1, 1, dto.SubItemUpdateReq{})
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — valid transitions using new status codes
// ---------------------------------------------------------------------------

func TestChangeStatus_PendingToProgressing(t *testing.T) {
	testValidTransitionTM(t, "pending", "progressing")
}

func TestChangeStatus_PendingToClosed(t *testing.T) {
	testValidTransitionTM(t, "pending", "closed")
}

func TestChangeStatus_ProgressingToBlocking(t *testing.T) {
	testValidTransitionTM(t, "progressing", "blocking")
}

func TestChangeStatus_ProgressingToPausing(t *testing.T) {
	testValidTransitionTM(t, "progressing", "pausing")
}

func TestChangeStatus_ProgressingToCompleted(t *testing.T) {
	testValidTransitionTM(t, "progressing", "completed")
}

func TestChangeStatus_ProgressingToClosed(t *testing.T) {
	testValidTransitionTM(t, "progressing", "closed")
}

func TestChangeStatus_BlockingToProgressing(t *testing.T) {
	testValidTransitionTM(t, "blocking", "progressing")
}

func TestChangeStatus_PausingToProgressing(t *testing.T) {
	testValidTransitionTM(t, "pausing", "progressing")
}

func TestChangeStatus_PausingToClosed(t *testing.T) {
	testValidTransitionTM(t, "pausing", "closed")
}

func testValidTransitionTM(t *testing.T, from, to string) {
	t.Helper()
	existing := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     from,
	}
	updated := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     to,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.Anything, existing, mock.MatchedBy(func(fields map[string]interface{}) bool {
		if fields["status"] != to {
			return false
		}
		if to == "completed" || to == "closed" {
			return fields["completion"] == float64(100) && fields["actual_end_date"] != nil
		}
		return true
	})).Return(nil)
	repo.On("FindByID", mock.Anything, uint(1)).Return(updated, nil).Once()

	// If transitioning to a terminal status, RecalcCompletion will be called.
	if to == "completed" || to == "closed" {
		mainSvc.On("RecalcCompletion", mock.Anything, uint(5)).Return(nil)
	}

	// Status history is always recorded
	historySvc.On("Record", mock.Anything, mock.MatchedBy(func(record *model.StatusHistory) bool {
		return record.ItemID == 1 && record.FromStatus == from && record.ToStatus == to && record.IsAuto == false
	})).Return(nil)

	// EvaluateLinkage is always called after status change
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Return(nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, to)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.SubItem)
	assert.Equal(t, to, result.SubItem.Status)
	assert.Nil(t, result.LinkageResult)

	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
	historySvc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — invalid transitions
// ---------------------------------------------------------------------------

func TestChangeStatus_Invalid_CompletedToAnything(t *testing.T) {
	for _, target := range []string{"pending", "progressing", "blocking", "pausing", "closed"} {
		t.Run("completed->"+target, func(t *testing.T) {
			existing := &model.SubItem{
				BaseModel:  model.BaseModel{ID: 1},
				TeamID: 1,
				Status: "completed",
			}
			repo := new(mockSubItemRepoTM)
			mainSvc := new(mockMainItemSvcTM)
			historySvc := new(mockStatusHistorySvcTM)
			svc := NewSubItemService(repo, mainSvc, historySvc)

			repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

			result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, target)
			assert.ErrorIs(t, err, apperrors.ErrInvalidStatus, "from completed to %s should be invalid", target)
			assert.Nil(t, result)

			repo.AssertExpectations(t)
		})
	}
}

func TestChangeStatus_Invalid_ClosedToAnything(t *testing.T) {
	for _, target := range []string{"pending", "progressing", "blocking", "pausing", "completed"} {
		t.Run("closed->"+target, func(t *testing.T) {
			existing := &model.SubItem{
				BaseModel:  model.BaseModel{ID: 1},
				TeamID: 1,
				Status: "closed",
			}
			repo := new(mockSubItemRepoTM)
			mainSvc := new(mockMainItemSvcTM)
			historySvc := new(mockStatusHistorySvcTM)
			svc := NewSubItemService(repo, mainSvc, historySvc)

			repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

			result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, target)
			assert.ErrorIs(t, err, apperrors.ErrInvalidStatus, "from closed to %s should be invalid", target)
			assert.Nil(t, result)

			repo.AssertExpectations(t)
		})
	}
}

func TestChangeStatus_Invalid_PendingToCompleted(t *testing.T) {
	testInvalidTransitionTM(t, "pending", "completed")
}

func TestChangeStatus_Invalid_ProgressingToPending(t *testing.T) {
	testInvalidTransitionTM(t, "progressing", "pending")
}

func TestChangeStatus_Invalid_BlockingToPausing(t *testing.T) {
	testInvalidTransitionTM(t, "blocking", "pausing")
}

func TestChangeStatus_SameStatus(t *testing.T) {
	testInvalidTransitionTM(t, "progressing", "progressing")
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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, to)
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus, "from %s to %s should be invalid", from, to)
	assert.Nil(t, result)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — completed sets completion=100 and ActualEndDate
// ---------------------------------------------------------------------------

func TestChangeStatus_Completed_SetsCompletionAndActualEndDate(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     "progressing",
	}
	updated := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     "completed",
		Completion: 100,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.Anything, existing, mock.MatchedBy(func(fields map[string]interface{}) bool {
		return fields["status"] == "completed" &&
			fields["completion"] == float64(100) &&
			fields["actual_end_date"] != nil
	})).Return(nil)
	repo.On("FindByID", mock.Anything, uint(1)).Return(updated, nil).Once()
	mainSvc.On("RecalcCompletion", mock.Anything, uint(5)).Return(nil)
	historySvc.On("Record", mock.Anything, mock.Anything).Return(nil)
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Return(nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "completed")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "completed", result.SubItem.Status)

	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
	historySvc.AssertExpectations(t)
}

func TestChangeStatus_Closed_SetsCompletionAndActualEndDate(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     "progressing",
	}
	updated := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     "closed",
		Completion: 100,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.Anything, existing, mock.MatchedBy(func(fields map[string]interface{}) bool {
		return fields["status"] == "closed" &&
			fields["completion"] == float64(100) &&
			fields["actual_end_date"] != nil
	})).Return(nil)
	repo.On("FindByID", mock.Anything, uint(1)).Return(updated, nil).Once()
	mainSvc.On("RecalcCompletion", mock.Anything, uint(5)).Return(nil)
	historySvc.On("Record", mock.Anything, mock.Anything).Return(nil)
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Return(nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "closed")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "closed", result.SubItem.Status)

	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
	historySvc.AssertExpectations(t)
}

func TestChangeStatus_Completed_RecalcCompletion_CalledWithCorrectMainItemID(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 42,
		Status:     "progressing",
	}
	updated := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 42,
		Status:     "completed",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("FindByID", mock.Anything, uint(1)).Return(updated, nil).Once()
	mainSvc.On("RecalcCompletion", mock.Anything, uint(42)).Return(nil)
	historySvc.On("Record", mock.Anything, mock.Anything).Return(nil)
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(42), uint(10)).Return(nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "completed")
	require.NoError(t, err)
	require.NotNil(t, result)

	mainSvc.AssertCalled(t, "RecalcCompletion", mock.Anything, uint(42))
	mainSvc.AssertExpectations(t)
}

func TestChangeStatus_Completed_RecalcError(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 1},
		TeamID:     1,
		MainItemID: 5,
		Status:     "progressing",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mainSvc.On("RecalcCompletion", mock.Anything, uint(5)).Return(errors.New("recalc failed"))

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "completed")
	assert.Error(t, err)
	assert.Nil(t, result)

	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — status history recorded
// ---------------------------------------------------------------------------

func TestChangeStatus_RecordsHistory(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 7},
		TeamID:     1,
		MainItemID: 5,
		Status:     "pending",
	}
	updated := &model.SubItem{
		BaseModel:      model.BaseModel{ID: 7},
		TeamID:     1,
		MainItemID: 5,
		Status:     "progressing",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(7)).Return(existing, nil).Once()
	repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("FindByID", mock.Anything, uint(7)).Return(updated, nil).Once()
	historySvc.On("Record", mock.Anything, mock.MatchedBy(func(record *model.StatusHistory) bool {
		return record.ItemType == "sub_item" &&
			record.ItemID == 7 &&
			record.FromStatus == "pending" &&
			record.ToStatus == "progressing" &&
			record.ChangedBy == 10 &&
			record.IsAuto == false
	})).Return(nil)
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Return(nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 7, "progressing")
	require.NoError(t, err)
	require.NotNil(t, result)

	historySvc.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — not found / team mismatch
// ---------------------------------------------------------------------------

func TestChangeStatus_NotFound(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 99, "progressing")
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
	assert.Nil(t, result)

	repo.AssertExpectations(t)
}

func TestChangeStatus_TeamMismatch(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 2,
		Status: "pending",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "progressing")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.Nil(t, result)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: AvailableTransitions
// ---------------------------------------------------------------------------

func TestSubItemAvailableTransitions_Pending(t *testing.T) {
	existing := &model.SubItem{
		BaseModel: model.BaseModel{ID: 1},
		TeamID:    1,
		Status:    "pending",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"progressing", "closed"}, transitions)

	repo.AssertExpectations(t)
}

func TestSubItemAvailableTransitions_Progressing(t *testing.T) {
	existing := &model.SubItem{
		BaseModel: model.BaseModel{ID: 1},
		TeamID:    1,
		Status:    "progressing",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"blocking", "pausing", "completed", "closed"}, transitions)

	repo.AssertExpectations(t)
}

func TestSubItemAvailableTransitions_TerminalStatus(t *testing.T) {
	existing := &model.SubItem{
		BaseModel: model.BaseModel{ID: 1},
		TeamID:    1,
		Status:    "completed",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	transitions, err := svc.AvailableTransitions(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Empty(t, transitions)

	repo.AssertExpectations(t)
}

func TestSubItemAvailableTransitions_NotFound(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.AvailableTransitions(context.Background(), 1, 99)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)

	repo.AssertExpectations(t)
}

func TestSubItemAvailableTransitions_TeamMismatch(t *testing.T) {
	existing := &model.SubItem{
		BaseModel: model.BaseModel{ID: 1},
		TeamID:    2,
		Status:    "pending",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	_, err := svc.AvailableTransitions(context.Background(), 1, 1)
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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	item, err := svc.Get(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, "Sub 1", item.Title)

	repo.AssertExpectations(t)
}

func TestSubItemGet_NotFound(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

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
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	_, err := svc.List(context.Background(), 1, nil, dto.SubItemFilter{}, dto.Pagination{})
	assert.Error(t, err)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: Create — triggers linkage
// ---------------------------------------------------------------------------

func TestSubItemCreate_TriggersLinkage(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("NextSubCode", mock.Anything, uint(5)).Return("FEAT-00001-01", nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mainSvc.On("Get", mock.Anything, uint(5)).Return(&model.MainItem{BaseModel: model.BaseModel{ID: 5}, Status: "pending"}, nil)
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Return(nil, nil)

	_, err := svc.Create(context.Background(), 1, 10, dto.SubItemCreateReq{
		MainItemID: 5,
		Title:      "Sub task",
		Priority:   "P2",
	})
	require.NoError(t, err)

	mainSvc.AssertCalled(t, "EvaluateLinkage", mock.Anything, uint(5), uint(10))
	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: Delete — triggers linkage
// ---------------------------------------------------------------------------

func TestSubItemDelete_TriggersLinkage(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		MainItemID: 5,
		Status: "pending",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	repo.On("Delete", mock.Anything, uint(1)).Return(nil)
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Return(nil, nil)

	err := svc.Delete(context.Background(), 1, 10, 1)
	require.NoError(t, err)

	mainSvc.AssertCalled(t, "EvaluateLinkage", mock.Anything, uint(5), uint(10))
	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
}

func TestSubItemDelete_NotFound(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.Delete(context.Background(), 1, 10, 99)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)

	repo.AssertExpectations(t)
}

func TestSubItemDelete_TeamMismatch(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 2,
		Status: "pending",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)

	err := svc.Delete(context.Background(), 1, 10, 1)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)

	repo.AssertExpectations(t)
}

func TestSubItemDelete_RepoError(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		MainItemID: 5,
		Status: "pending",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil)
	repo.On("Delete", mock.Anything, uint(1)).Return(errors.New("db error"))

	err := svc.Delete(context.Background(), 1, 10, 1)
	assert.Error(t, err)

	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — returns linkage result
// ---------------------------------------------------------------------------

func TestChangeStatus_ReturnsLinkageResult(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		MainItemID: 5,
		Status: "pending",
	}
	updated := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		MainItemID: 5,
		Status: "progressing",
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("FindByID", mock.Anything, uint(1)).Return(updated, nil).Once()
	historySvc.On("Record", mock.Anything, mock.Anything).Return(nil)

	linkageResult := &LinkageResult{
		Triggered:    true,
		Success:      false,
		TargetStatus: "reviewing",
		Remark:       "blocking→reviewing 不允许",
	}
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Return(linkageResult, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "progressing")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "progressing", result.SubItem.Status)
	require.NotNil(t, result.LinkageResult)
	assert.Equal(t, "主事项状态联动失败：blocking→reviewing 不允许", result.LinkageResult.Warning())

	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
	historySvc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Tests: Create — calls NextSubCode and assigns Code
// ---------------------------------------------------------------------------

func TestSubItemCreate_AssignsCode(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("NextSubCode", mock.Anything, uint(5)).Return("FEAT-00001-01", nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(item *model.SubItem) bool {
		return item.Code == "FEAT-00001-01"
	})).Return(nil)
	mainSvc.On("Get", mock.Anything, uint(5)).Return(&model.MainItem{BaseModel: model.BaseModel{ID: 5}, Status: "pending"}, nil)
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Return(nil, nil)

	item, err := svc.Create(context.Background(), 1, 10, dto.SubItemCreateReq{
		MainItemID: 5,
		Title:      "Sub task",
		Priority:   "P2",
	})
	require.NoError(t, err)
	assert.Equal(t, "FEAT-00001-01", item.Code)

	repo.AssertExpectations(t)
}

func TestSubItemCreate_NextSubCodeError_ReturnsError(t *testing.T) {
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	repo.On("NextSubCode", mock.Anything, uint(5)).Return("", errors.New("code gen failed"))
	mainSvc.On("Get", mock.Anything, uint(5)).Return(&model.MainItem{BaseModel: model.BaseModel{ID: 5}, Status: "pending"}, nil)

	_, err := svc.Create(context.Background(), 1, 10, dto.SubItemCreateReq{
		MainItemID: 5,
		Title:      "Sub task",
		Priority:   "P2",
	})
	assert.Error(t, err)
	repo.AssertExpectations(t)
	// Create should NOT have been called
	repo.AssertNotCalled(t, "Create")
}

// TestChangeStatus_RecalcCompletionBeforeLinkage tests AC-13:
// RecalcCompletion runs before EvaluateLinkage on SubItem completed.
func TestChangeStatus_RecalcCompletionBeforeLinkage(t *testing.T) {
	existing := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		MainItemID: 5,
		Status: "progressing",
	}
	updated := &model.SubItem{
		BaseModel:  model.BaseModel{ID: 1},
		TeamID: 1,
		MainItemID: 5,
		Status: "completed",
		Completion: 100,
	}
	repo := new(mockSubItemRepoTM)
	mainSvc := new(mockMainItemSvcTM)
	historySvc := new(mockStatusHistorySvcTM)
	svc := NewSubItemService(repo, mainSvc, historySvc)

	var callOrder []string
	repo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("FindByID", mock.Anything, uint(1)).Return(updated, nil).Once()
	historySvc.On("Record", mock.Anything, mock.Anything).Return(nil)

	mainSvc.On("RecalcCompletion", mock.Anything, uint(5)).Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "recalc")
	}).Return(nil)
	mainSvc.On("EvaluateLinkage", mock.Anything, uint(5), uint(10)).Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "linkage")
	}).Return(nil, nil)

	result, err := svc.ChangeStatus(context.Background(), 1, 10, 1, "completed")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify order: RecalcCompletion before EvaluateLinkage
	assert.Equal(t, []string{"recalc", "linkage"}, callOrder)

	repo.AssertExpectations(t)
	mainSvc.AssertExpectations(t)
	historySvc.AssertExpectations(t)
}
