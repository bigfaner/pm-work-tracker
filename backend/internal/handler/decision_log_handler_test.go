package handler

import (
	"context"
	"testing"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Mock DecisionLogService for handler tests
// ---------------------------------------------------------------------------

type mockDecisionLogService struct{}

func (m *mockDecisionLogService) Create(_ context.Context, _ uint, _ uint, _ dto.DecisionLogCreateReq) (*model.DecisionLog, error) {
	return nil, nil
}
func (m *mockDecisionLogService) Update(_ context.Context, _ int64, _ uint, _ dto.DecisionLogUpdateReq) (*model.DecisionLog, error) {
	return nil, nil
}
func (m *mockDecisionLogService) Publish(_ context.Context, _ int64, _ uint) (*model.DecisionLog, error) {
	return nil, nil
}
func (m *mockDecisionLogService) List(_ context.Context, _ uint, _ uint, _ dto.Pagination) (*dto.PageResult[model.DecisionLog], error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Mock MainItemRepo for decision log handler tests
// ---------------------------------------------------------------------------

type mockMainItemRepoForDecisionLog struct{}

func (m *mockMainItemRepoForDecisionLog) Create(_ context.Context, _ *model.MainItem) error {
	return nil
}
func (m *mockMainItemRepoForDecisionLog) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockMainItemRepoForDecisionLog) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) NextCode(_ context.Context, _ int64) (string, error) {
	return "", nil
}
func (m *mockMainItemRepoForDecisionLog) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepoForDecisionLog) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}

// Shared mock instances for constructor tests
var (
	mockDecisionLogSvc = &mockDecisionLogService{}
	mockUserRepo       = &mockUserRepoForHandler{}
	mockMainItemRepo   = &mockMainItemRepoForDecisionLog{}
)

func TestNewDecisionLogHandler_NilDependencies(t *testing.T) {
	t.Run("panics on nil service", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDecisionLogHandler(nil, mockUserRepo, mockMainItemRepo)
		})
	})
	t.Run("panics on nil userRepo", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDecisionLogHandler(mockDecisionLogSvc, nil, mockMainItemRepo)
		})
	})
	t.Run("panics on nil mainItemRepo", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDecisionLogHandler(mockDecisionLogSvc, mockUserRepo, nil)
		})
	})
	t.Run("succeeds with all dependencies", func(t *testing.T) {
		h := NewDecisionLogHandler(mockDecisionLogSvc, mockUserRepo, mockMainItemRepo)
		assert.NotNil(t, h)
	})
}
