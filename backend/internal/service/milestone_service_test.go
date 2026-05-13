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
	createdItem   *model.Milestone
	updatedID     uint
	updatedFields map[string]interface{}
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

func (m *mockMilestoneRepo) SoftDelete(_ context.Context, _ uint) error {
	return nil
}

func (m *mockMilestoneRepo) DeleteByMap(_ context.Context, _ int64) error {
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

// ---------------------------------------------------------------------------
// Tests: Create
// ---------------------------------------------------------------------------

func TestMilestoneCreate_Success(t *testing.T) {
	mapRepo := &mockMilestoneMapRepo{item: &model.MilestoneMap{}}
	milestoneRepo := &mockMilestoneRepo{}
	svc := NewMilestoneService(milestoneRepo, mapRepo)

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
	svc := NewMilestoneService(milestoneRepo, mapRepo)

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
	svc := NewMilestoneService(milestoneRepo, mapRepo)

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
	svc := NewMilestoneService(milestoneRepo, mapRepo)

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
	svc := NewMilestoneService(milestoneRepo, nil)

	m, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "M1", m.MilestoneName)
}

func TestMilestoneGet_NotFound(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{findErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneService(milestoneRepo, nil)

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
	svc := NewMilestoneService(milestoneRepo, nil)

	m, err := svc.GetByBizKey(context.Background(), 123456)
	require.NoError(t, err)
	assert.Equal(t, "M1", m.MilestoneName)
}

func TestMilestoneGetByBizKey_NotFound(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{bizKeyErr: gorm.ErrRecordNotFound}
	svc := NewMilestoneService(milestoneRepo, nil)

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
	svc := NewMilestoneService(milestoneRepo, nil)

	result, err := svc.ListByMap(context.Background(), 100)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestMilestoneListByMap_RepoError(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{listErr: errors.New("db error")}
	svc := NewMilestoneService(milestoneRepo, nil)

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
	svc := NewMilestoneService(milestoneRepo, nil)

	result, err := svc.ListByTeam(context.Background(), 1, true)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestMilestoneListByTeam_RepoError(t *testing.T) {
	milestoneRepo := &mockMilestoneRepo{listErr: errors.New("db error")}
	svc := NewMilestoneService(milestoneRepo, nil)

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
	svc := NewMilestoneService(milestoneRepo, nil)

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
	svc := NewMilestoneService(milestoneRepo, nil)

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
	svc := NewMilestoneService(milestoneRepo, nil)

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
	svc := NewMilestoneService(milestoneRepo, nil)

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
	svc := NewMilestoneService(milestoneRepo, nil)

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
	svc := NewMilestoneService(milestoneRepo, nil)

	newDate := "2026-07-01"
	_, err := svc.Update(context.Background(), 1, dto.MilestoneUpdateReq{
		ExpectedEndDate: &newDate,
		DbUpdateTime:    now.Format(time.RFC3339),
	})
	require.NoError(t, err)
	assert.Equal(t, newDate, milestoneRepo.updatedFields["expected_end_date"])
}
