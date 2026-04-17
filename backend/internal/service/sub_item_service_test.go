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
// Mock repos for SubItemService tests
// ---------------------------------------------------------------------------

type mockSubItemRepoForSubSvc struct {
	item       *model.SubItem
	items      []model.SubItem
	pageResult *dto.PageResult[model.SubItem]

	findErr   error
	createErr error
	updateErr error
	listErr   error

	createdItem   *model.SubItem
	updatedItem   *model.SubItem
	updatedFields map[string]interface{}
}

func (m *mockSubItemRepoForSubSvc) Create(_ context.Context, item *model.SubItem) error {
	m.createdItem = item
	if m.createErr != nil {
		return m.createErr
	}
	item.ID = 1
	return nil
}

func (m *mockSubItemRepoForSubSvc) FindByID(_ context.Context, id uint) (*model.SubItem, error) {
	if m.item != nil {
		return m.item, nil
	}
	return nil, m.findErr
}

func (m *mockSubItemRepoForSubSvc) Update(_ context.Context, item *model.SubItem, fields map[string]interface{}) error {
	m.updatedItem = item
	m.updatedFields = fields
	return m.updateErr
}

func (m *mockSubItemRepoForSubSvc) List(_ context.Context, teamID uint, mainItemID uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.pageResult != nil {
		return m.pageResult, nil
	}
	return &dto.PageResult[model.SubItem]{Items: m.items, Total: int64(len(m.items))}, nil
}

func (m *mockSubItemRepoForSubSvc) ListByMainItem(_ context.Context, mainItemID uint) ([]*model.SubItem, error) {
	return nil, nil
}

// mockMainItemSvcForSubSvc captures RecalcCompletion calls.
type mockMainItemSvcForSubSvc struct {
	recalcCalledWith uint
	recalcErr        error
}

func (m *mockMainItemSvcForSubSvc) Create(_ context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemSvcForSubSvc) Update(_ context.Context, teamID, itemID uint, req dto.MainItemUpdateReq) error {
	return nil
}
func (m *mockMainItemSvcForSubSvc) Archive(_ context.Context, teamID, itemID uint) error { return nil }
func (m *mockMainItemSvcForSubSvc) List(_ context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemSvcForSubSvc) Get(_ context.Context, itemID uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemSvcForSubSvc) RecalcCompletion(_ context.Context, mainItemID uint) error {
	m.recalcCalledWith = mainItemID
	return m.recalcErr
}

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestSubItemCreate_Success(t *testing.T) {
	repo := &mockSubItemRepoForSubSvc{}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	item, err := svc.Create(context.Background(), 1, 10, dto.SubItemCreateReq{
		MainItemID: 5,
		Title:      "Sub task A",
		Priority:   "P2",
	})
	require.NoError(t, err)
	assert.Equal(t, uint(1), item.TeamID)
	assert.Equal(t, uint(5), item.MainItemID)
	assert.Equal(t, "待开始", item.Status)
	assert.Equal(t, "Sub task A", item.Title)
	assert.Equal(t, "P2", item.Priority)
	assert.Nil(t, repo.createdItem.AssigneeID)
}

func TestSubItemCreate_RepoError(t *testing.T) {
	repo := &mockSubItemRepoForSubSvc{createErr: errors.New("db error")}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	_, err := svc.Create(context.Background(), 1, 10, dto.SubItemCreateReq{
		MainItemID: 5,
		Title:      "Sub task",
		Priority:   "P2",
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestSubItemUpdate_Success(t *testing.T) {
	existing := &model.SubItem{
		Model:  gorm.Model{ID: 1},
		TeamID: 1,
		Title:  "Old Title",
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.Update(context.Background(), 1, 1, dto.SubItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	require.NoError(t, err)
	assert.Equal(t, "New Title", repo.updatedFields["title"])
}

func TestSubItemUpdate_TeamMismatch(t *testing.T) {
	existing := &model.SubItem{
		Model:  gorm.Model{ID: 1},
		TeamID: 2,
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.Update(context.Background(), 1, 1, dto.SubItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestSubItemUpdate_NotFound(t *testing.T) {
	repo := &mockSubItemRepoForSubSvc{findErr: gorm.ErrRecordNotFound}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.Update(context.Background(), 1, 99, dto.SubItemUpdateReq{
		Title: ptrStr("New Title"),
	})
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestSubItemUpdate_NoFields_Noop(t *testing.T) {
	existing := &model.SubItem{
		Model:  gorm.Model{ID: 1},
		TeamID: 1,
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.Update(context.Background(), 1, 1, dto.SubItemUpdateReq{})
	require.NoError(t, err)
	assert.Nil(t, repo.updatedFields)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — valid transitions
// ---------------------------------------------------------------------------

func TestChangeStatus_待开始_to_进行中(t *testing.T) {
	testValidTransition(t, "待开始", "进行中")
}

func TestChangeStatus_待开始_to_已关闭(t *testing.T) {
	testValidTransition(t, "待开始", "已关闭")
}

func TestChangeStatus_进行中_to_阻塞中(t *testing.T) {
	testValidTransition(t, "进行中", "阻塞中")
}

func TestChangeStatus_进行中_to_挂起(t *testing.T) {
	testValidTransition(t, "进行中", "挂起")
}

func TestChangeStatus_进行中_to_待验收(t *testing.T) {
	testValidTransition(t, "进行中", "待验收")
}

func TestChangeStatus_进行中_to_已延期(t *testing.T) {
	testValidTransition(t, "进行中", "已延期")
}

func TestChangeStatus_进行中_to_已关闭(t *testing.T) {
	testValidTransition(t, "进行中", "已关闭")
}

func TestChangeStatus_阻塞中_to_进行中(t *testing.T) {
	testValidTransition(t, "阻塞中", "进行中")
}

func TestChangeStatus_挂起_to_进行中(t *testing.T) {
	testValidTransition(t, "挂起", "进行中")
}

func TestChangeStatus_挂起_to_已关闭(t *testing.T) {
	testValidTransition(t, "挂起", "已关闭")
}

func TestChangeStatus_已延期_to_进行中(t *testing.T) {
	testValidTransition(t, "已延期", "进行中")
}

func TestChangeStatus_待验收_to_已完成(t *testing.T) {
	testValidTransition(t, "待验收", "已完成")
}

func TestChangeStatus_待验收_to_进行中(t *testing.T) {
	testValidTransition(t, "待验收", "进行中")
}

func testValidTransition(t *testing.T, from, to string) {
	t.Helper()
	existing := &model.SubItem{
		Model:     gorm.Model{ID: 1},
		TeamID:    1,
		Status:    from,
		DelayCount: 0,
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, to)
	require.NoError(t, err)
	assert.Equal(t, to, repo.updatedFields["status"])
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — invalid transitions
// ---------------------------------------------------------------------------

func TestChangeStatus_Invalid_已完成_to_anything(t *testing.T) {
	for _, target := range []string{"待开始", "进行中", "阻塞中", "挂起", "已延期", "待验收", "已关闭"} {
		existing := &model.SubItem{
			Model:  gorm.Model{ID: 1},
			TeamID: 1,
			Status: "已完成",
		}
		repo := &mockSubItemRepoForSubSvc{item: existing}
		mainSvc := &mockMainItemSvcForSubSvc{}
		svc := NewSubItemService(repo, mainSvc)

		err := svc.ChangeStatus(context.Background(), 1, 10, 1, target)
		assert.ErrorIs(t, err, apperrors.ErrInvalidStatus, "from 已完成 to %s should be invalid", target)
	}
}

func TestChangeStatus_Invalid_已关闭_to_anything(t *testing.T) {
	for _, target := range []string{"待开始", "进行中", "阻塞中", "挂起", "已延期", "待验收", "已完成"} {
		existing := &model.SubItem{
			Model:  gorm.Model{ID: 1},
			TeamID: 1,
			Status: "已关闭",
		}
		repo := &mockSubItemRepoForSubSvc{item: existing}
		mainSvc := &mockMainItemSvcForSubSvc{}
		svc := NewSubItemService(repo, mainSvc)

		err := svc.ChangeStatus(context.Background(), 1, 10, 1, target)
		assert.ErrorIs(t, err, apperrors.ErrInvalidStatus, "from 已关闭 to %s should be invalid", target)
	}
}

func TestChangeStatus_Invalid_待开始_to_待验收(t *testing.T) {
	testInvalidTransition(t, "待开始", "待验收")
}

func TestChangeStatus_Invalid_待开始_to_已延期(t *testing.T) {
	testInvalidTransition(t, "待开始", "已延期")
}

func TestChangeStatus_Invalid_进行中_to_待开始(t *testing.T) {
	testInvalidTransition(t, "进行中", "待开始")
}

func TestChangeStatus_Invalid_阻塞中_to_挂起(t *testing.T) {
	testInvalidTransition(t, "阻塞中", "挂起")
}

func TestChangeStatus_SameStatus(t *testing.T) {
	testInvalidTransition(t, "进行中", "进行中")
}

func testInvalidTransition(t *testing.T, from, to string) {
	t.Helper()
	existing := &model.SubItem{
		Model:  gorm.Model{ID: 1},
		TeamID: 1,
		Status: from,
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, to)
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus, "from %s to %s should be invalid", from, to)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — delay count and auto-upgrade
// ---------------------------------------------------------------------------

func TestChangeStatus_已延期_IncrementsDelayCount(t *testing.T) {
	existing := &model.SubItem{
		Model:      gorm.Model{ID: 1},
		TeamID:     1,
		Status:     "进行中",
		DelayCount: 0,
		Priority:   "P3",
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "已延期")
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updatedFields["delay_count"])
	// DelayCount=1 should NOT trigger auto-upgrade
	assert.Nil(t, repo.updatedFields["is_key_item"])
	assert.Nil(t, repo.updatedFields["priority"])
}

func TestChangeStatus_已延期_AutoUpgradeAtCount2(t *testing.T) {
	existing := &model.SubItem{
		Model:      gorm.Model{ID: 1},
		TeamID:     1,
		Status:     "进行中",
		DelayCount: 1,
		Priority:   "P3",
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "已延期")
	require.NoError(t, err)
	assert.Equal(t, 2, repo.updatedFields["delay_count"])
	assert.Equal(t, true, repo.updatedFields["is_key_item"])
	assert.Equal(t, "P1", repo.updatedFields["priority"])
}

func TestChangeStatus_已延期_AutoUpgradeAtCount3(t *testing.T) {
	existing := &model.SubItem{
		Model:      gorm.Model{ID: 1},
		TeamID:     1,
		Status:     "进行中",
		DelayCount: 2,
		Priority:   "P2",
		IsKeyItem:  true,
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "已延期")
	require.NoError(t, err)
	assert.Equal(t, 3, repo.updatedFields["delay_count"])
	assert.Equal(t, true, repo.updatedFields["is_key_item"])
	assert.Equal(t, "P1", repo.updatedFields["priority"])
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — 已完成 sets ActualEndDate and calls RecalcCompletion
// ---------------------------------------------------------------------------

func TestChangeStatus_已完成_SetsActualEndDate(t *testing.T) {
	existing := &model.SubItem{
		Model:     gorm.Model{ID: 1},
		TeamID:    1,
		MainItemID: 5,
		Status:    "待验收",
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "已完成")
	require.NoError(t, err)
	assert.NotNil(t, repo.updatedFields["actual_end_date"])
	assert.Equal(t, uint(5), mainSvc.recalcCalledWith)
}

func TestChangeStatus_已完成_RecalcError(t *testing.T) {
	existing := &model.SubItem{
		Model:     gorm.Model{ID: 1},
		TeamID:    1,
		MainItemID: 5,
		Status:    "待验收",
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{recalcErr: errors.New("recalc failed")}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "已完成")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: ChangeStatus — not found / team mismatch
// ---------------------------------------------------------------------------

func TestChangeStatus_NotFound(t *testing.T) {
	repo := &mockSubItemRepoForSubSvc{findErr: gorm.ErrRecordNotFound}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.ChangeStatus(context.Background(), 1, 10, 99, "进行中")
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestChangeStatus_TeamMismatch(t *testing.T) {
	existing := &model.SubItem{
		Model:  gorm.Model{ID: 1},
		TeamID: 2,
		Status: "待开始",
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.ChangeStatus(context.Background(), 1, 10, 1, "进行中")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: Assign
// ---------------------------------------------------------------------------

func TestSubItemAssign_Success(t *testing.T) {
	existing := &model.SubItem{
		Model:  gorm.Model{ID: 1},
		TeamID: 1,
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	var assigneeID uint = 42
	err := svc.Assign(context.Background(), 1, 10, 1, assigneeID)
	require.NoError(t, err)
	assert.Equal(t, assigneeID, repo.updatedFields["assignee_id"])
}

func TestSubItemAssign_NotFound(t *testing.T) {
	repo := &mockSubItemRepoForSubSvc{findErr: gorm.ErrRecordNotFound}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.Assign(context.Background(), 1, 10, 99, 42)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

func TestSubItemAssign_TeamMismatch(t *testing.T) {
	existing := &model.SubItem{
		Model:  gorm.Model{ID: 1},
		TeamID: 2,
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	err := svc.Assign(context.Background(), 1, 10, 1, 42)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: Get
// ---------------------------------------------------------------------------

func TestSubItemGet_Success(t *testing.T) {
	existing := &model.SubItem{
		Model:  gorm.Model{ID: 1},
		TeamID: 1,
		Title:  "Sub 1",
	}
	repo := &mockSubItemRepoForSubSvc{item: existing}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	item, err := svc.Get(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, "Sub 1", item.Title)
}

func TestSubItemGet_NotFound(t *testing.T) {
	repo := &mockSubItemRepoForSubSvc{findErr: gorm.ErrRecordNotFound}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	_, err := svc.Get(context.Background(), 1, 99)
	assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}

// ---------------------------------------------------------------------------
// Tests: List
// ---------------------------------------------------------------------------

func TestSubItemList_Success(t *testing.T) {
	items := []model.SubItem{
		{Model: gorm.Model{ID: 1}, Title: "Sub 1"},
		{Model: gorm.Model{ID: 2}, Title: "Sub 2"},
	}
	repo := &mockSubItemRepoForSubSvc{items: items}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	result, err := svc.List(context.Background(), 1, nil, dto.SubItemFilter{}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)
}

func TestSubItemList_WithMainItemFilter(t *testing.T) {
	items := []model.SubItem{
		{Model: gorm.Model{ID: 1}, Title: "Sub 1"},
	}
	repo := &mockSubItemRepoForSubSvc{items: items}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	mainID := uint(5)
	result, err := svc.List(context.Background(), 1, &mainID, dto.SubItemFilter{}, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
}

func TestSubItemList_RepoError(t *testing.T) {
	repo := &mockSubItemRepoForSubSvc{listErr: errors.New("db error")}
	mainSvc := &mockMainItemSvcForSubSvc{}
	svc := NewSubItemService(repo, mainSvc)

	_, err := svc.List(context.Background(), 1, nil, dto.SubItemFilter{}, dto.Pagination{})
	assert.Error(t, err)
}
