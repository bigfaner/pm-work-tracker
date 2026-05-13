package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/repo"
)

// mmMockRepo is a testify/mock implementation of repository.MilestoneMapRepo for service tests.
type mmMockRepo struct {
	mock.Mock
}

func (m *mmMockRepo) Create(ctx context.Context, mm *model.MilestoneMap) error {
	args := m.Called(ctx, mm)
	return args.Error(0)
}

func (m *mmMockRepo) FindByID(ctx context.Context, id uint) (*model.MilestoneMap, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MilestoneMap), args.Error(1)
}

func (m *mmMockRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error) {
	args := m.Called(ctx, bizKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MilestoneMap), args.Error(1)
}

func (m *mmMockRepo) Update(ctx context.Context, mm *model.MilestoneMap, fields map[string]interface{}) error {
	args := m.Called(ctx, mm, fields)
	return args.Error(0)
}

func (m *mmMockRepo) List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error) {
	args := m.Called(ctx, teamBizKey, filter, page)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PageResult[model.MilestoneMap]), args.Error(1)
}

func (m *mmMockRepo) SoftDelete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// mmMockMilestoneRepo is a minimal mock for repository.MilestoneRepo.
type mmMockMilestoneRepo struct{}

func (m *mmMockMilestoneRepo) Create(_ context.Context, _ *model.Milestone) error { return nil }
func (m *mmMockMilestoneRepo) FindByID(_ context.Context, _ uint) (*model.Milestone, error) {
	return nil, nil
}
func (m *mmMockMilestoneRepo) FindByBizKey(_ context.Context, _ int64) (*model.Milestone, error) {
	return nil, nil
}
func (m *mmMockMilestoneRepo) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.Milestone, error) {
	return nil, nil
}
func (m *mmMockMilestoneRepo) Update(_ context.Context, _ *model.Milestone, _ map[string]interface{}) error {
	return nil
}
func (m *mmMockMilestoneRepo) ListByMap(_ context.Context, _ int64) ([]model.Milestone, error) {
	return nil, nil
}
func (m *mmMockMilestoneRepo) ListByTeam(_ context.Context, _ int64, _ bool) ([]model.Milestone, error) {
	return nil, nil
}
func (m *mmMockMilestoneRepo) SoftDelete(_ context.Context, _ uint) error { return nil }
func (m *mmMockMilestoneRepo) DeleteByMap(_ context.Context, _ int64) error { return nil }

// mmMockMainItemRepo is a minimal mock for repository.MainItemRepo.
type mmMockMainItemRepo struct{}

func (m *mmMockMainItemRepo) Create(_ context.Context, _ *model.MainItem) error             { return nil }
func (m *mmMockMainItemRepo) FindByID(_ context.Context, _ uint) (*model.MainItem, error)   { return nil, nil }
func (m *mmMockMainItemRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mmMockMainItemRepo) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) { return nil, nil }
func (m *mmMockMainItemRepo) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mmMockMainItemRepo) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *mmMockMainItemRepo) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mmMockMainItemRepo) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mmMockMainItemRepo) NextCode(_ context.Context, _ int64) (string, error) { return "M001", nil }
func (m *mmMockMainItemRepo) CountByTeam(_ context.Context, _ int64) (int64, error) { return 0, nil }
func (m *mmMockMainItemRepo) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mmMockMainItemRepo) UnbindByMilestone(_ context.Context, _ int64) error { return nil }
func (m *mmMockMainItemRepo) UnbindByMap(_ context.Context, _ int64) error       { return nil }
func (m *mmMockMainItemRepo) CalcCompletionByMilestone(_ context.Context, _ int64) (float64, error) {
	return 0, nil
}
func (m *mmMockMainItemRepo) CountByMilestone(_ context.Context, _ int64) (int64, error) { return 0, nil }
func (m *mmMockMainItemRepo) CalcCompletionByMap(_ context.Context, _ int64) (float64, error) {
	return 0, nil
}
func (m *mmMockMainItemRepo) CountByMap(_ context.Context, _ int64) (int64, error) { return 0, nil }

type mmMockDBTx struct{}

func (m *mmMockDBTx) Transaction(fc func(tx *gorm.DB) error, _ ...*sql.TxOptions) error {
	return fc(nil)
}

var _ repo.DBTransactor = (*mmMockDBTx)(nil)

// mmSampleMap creates a sample MilestoneMap model.
func mmSampleMap(id uint, teamKey int64, name string) *model.MilestoneMap {
	return &model.MilestoneMap{
		BaseModel: model.BaseModel{ID: id, BizKey: 1234567890},
		TeamKey:   teamKey,
		MapName:   name,
		MapDesc:   "test description",
		MapStatus: "planning",
	}
}

func TestMilestoneMapService_Create(t *testing.T) {
	ctx := context.Background()
	repo := new(mmMockRepo)
	svc := NewMilestoneMapService(repo, &mmMockMilestoneRepo{}, &mmMockMainItemRepo{}, &mmMockDBTx{})

	t.Run("success", func(t *testing.T) {
		req := dto.MilestoneMapCreateReq{MapName: "Test Map", MapDesc: "desc"}
		repo.On("Create", ctx, mock.MatchedBy(func(m *model.MilestoneMap) bool {
			return m.MapName == "Test Map" && m.MapStatus == "planning" && m.TeamKey == 100
		})).Return(nil).Once()

		result, err := svc.Create(ctx, 100, req)
		require.NoError(t, err)
		assert.Equal(t, "Test Map", result.MapName)
		assert.Equal(t, "planning", result.MapStatus)
		assert.Equal(t, int64(100), result.TeamKey)
		assert.NotZero(t, result.BizKey)
		repo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		req := dto.MilestoneMapCreateReq{MapName: "Test Map"}
		repo.On("Create", ctx, mock.Anything).Return(assert.AnError).Once()

		_, err := svc.Create(ctx, 100, req)
		assert.ErrorIs(t, err, assert.AnError)
		repo.AssertExpectations(t)
	})
}

func TestMilestoneMapService_Get(t *testing.T) {
	ctx := context.Background()
	repo := new(mmMockRepo)
	svc := NewMilestoneMapService(repo, &mmMockMilestoneRepo{}, &mmMockMainItemRepo{}, &mmMockDBTx{})

	t.Run("success", func(t *testing.T) {
		expected := mmSampleMap(1, 100, "Test")
		repo.On("FindByID", ctx, uint(1)).Return(expected, nil).Once()

		result, err := svc.Get(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "Test", result.MapName)
		repo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		repo.On("FindByID", ctx, uint(999)).Return(nil, apperrors.ErrNotFound).Once()

		_, err := svc.Get(ctx, 999)
		assert.Equal(t, apperrors.ErrMilestoneMapNotFound, err)
		repo.AssertExpectations(t)
	})
}

func TestMilestoneMapService_GetByBizKey(t *testing.T) {
	ctx := context.Background()
	repo := new(mmMockRepo)
	svc := NewMilestoneMapService(repo, &mmMockMilestoneRepo{}, &mmMockMainItemRepo{}, &mmMockDBTx{})

	t.Run("success", func(t *testing.T) {
		expected := mmSampleMap(1, 100, "Test")
		repo.On("FindByBizKey", ctx, int64(1234567890)).Return(expected, nil).Once()

		result, err := svc.GetByBizKey(ctx, 1234567890)
		require.NoError(t, err)
		assert.Equal(t, "Test", result.MapName)
		repo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		repo.On("FindByBizKey", ctx, int64(999)).Return(nil, apperrors.ErrNotFound).Once()

		_, err := svc.GetByBizKey(ctx, 999)
		assert.Equal(t, apperrors.ErrMilestoneMapNotFound, err)
		repo.AssertExpectations(t)
	})
}

func TestMilestoneMapService_List(t *testing.T) {
	ctx := context.Background()
	repo := new(mmMockRepo)
	svc := NewMilestoneMapService(repo, &mmMockMilestoneRepo{}, &mmMockMainItemRepo{}, &mmMockDBTx{})

	t.Run("success", func(t *testing.T) {
		items := []model.MilestoneMap{*mmSampleMap(1, 100, "Map A"), *mmSampleMap(2, 100, "Map B")}
		expected := &dto.PageResult[model.MilestoneMap]{Items: items, Total: 2, Page: 1, Size: 20}
		filter := dto.MilestoneMapFilter{}
		page := dto.Pagination{Page: 1, PageSize: 20}

		repo.On("List", ctx, int64(100), filter, page).Return(expected, nil).Once()

		result, err := svc.List(ctx, 100, filter, page)
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Len(t, result.Items, 2)
		repo.AssertExpectations(t)
	})

	t.Run("with status filter", func(t *testing.T) {
		status := "planning"
		filter := dto.MilestoneMapFilter{Status: &status}
		page := dto.Pagination{Page: 1, PageSize: 20}
		expected := &dto.PageResult[model.MilestoneMap]{Items: []model.MilestoneMap{}, Total: 0, Page: 1, Size: 20}

		repo.On("List", ctx, int64(100), filter, page).Return(expected, nil).Once()

		result, err := svc.List(ctx, 100, filter, page)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
		repo.AssertExpectations(t)
	})
}

func TestMilestoneMapService_Update(t *testing.T) {
	ctx := context.Background()
	repo := new(mmMockRepo)
	svc := NewMilestoneMapService(repo, &mmMockMilestoneRepo{}, &mmMockMainItemRepo{}, &mmMockDBTx{})

	t.Run("update name only", func(t *testing.T) {
		existing := mmSampleMap(1, 100, "Old Name")
		repo.On("FindByID", ctx, uint(1)).Return(existing, nil).Once()
		repo.On("Update", ctx, existing, map[string]interface{}{"map_name": "New Name"}).Return(nil).Once()

		newName := "New Name"
		err := svc.Update(ctx, 100, 1, dto.MilestoneMapUpdateReq{MapName: &newName})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("update desc only", func(t *testing.T) {
		existing := mmSampleMap(1, 100, "Test")
		repo.On("FindByID", ctx, uint(1)).Return(existing, nil).Once()
		repo.On("Update", ctx, existing, map[string]interface{}{"map_desc": "new desc"}).Return(nil).Once()

		desc := "new desc"
		err := svc.Update(ctx, 100, 1, dto.MilestoneMapUpdateReq{MapDesc: &desc})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("no fields to update", func(t *testing.T) {
		existing := mmSampleMap(1, 100, "Test")
		repo.On("FindByID", ctx, uint(1)).Return(existing, nil).Once()

		err := svc.Update(ctx, 100, 1, dto.MilestoneMapUpdateReq{})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		repo.On("FindByID", ctx, uint(999)).Return(nil, apperrors.ErrNotFound).Once()

		name := "x"
		err := svc.Update(ctx, 100, 999, dto.MilestoneMapUpdateReq{MapName: &name})
		assert.Equal(t, apperrors.ErrMilestoneMapNotFound, err)
		repo.AssertExpectations(t)
	})

	t.Run("wrong team", func(t *testing.T) {
		existing := mmSampleMap(1, 200, "Test")
		repo.On("FindByID", ctx, uint(1)).Return(existing, nil).Once()

		name := "x"
		err := svc.Update(ctx, 100, 1, dto.MilestoneMapUpdateReq{MapName: &name})
		assert.Equal(t, apperrors.ErrForbidden, err)
		repo.AssertExpectations(t)
	})
}
