package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// --- Mock repos ---

type mockDecisionLogRepo struct {
	createFn       func(ctx context.Context, log *model.DecisionLog) error
	findByBizKeyFn func(ctx context.Context, bizKey int64) (*model.DecisionLog, error)
	listByItemFn   func(ctx context.Context, mainItemID uint, userID uint, offset, limit int) ([]model.DecisionLog, int64, error)
	updateFn       func(ctx context.Context, log *model.DecisionLog) error
}

func (m *mockDecisionLogRepo) Create(ctx context.Context, log *model.DecisionLog) error {
	return m.createFn(ctx, log)
}

func (m *mockDecisionLogRepo) FindByID(ctx context.Context, id uint) (*model.DecisionLog, error) {
	return nil, nil
}

func (m *mockDecisionLogRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
	return m.findByBizKeyFn(ctx, bizKey)
}

func (m *mockDecisionLogRepo) ListByItem(ctx context.Context, mainItemID uint, userID uint, offset, limit int) ([]model.DecisionLog, int64, error) {
	return m.listByItemFn(ctx, mainItemID, userID, offset, limit)
}

func (m *mockDecisionLogRepo) Update(ctx context.Context, log *model.DecisionLog) error {
	return m.updateFn(ctx, log)
}

type mockMainItemRepoForDL struct {
	findByBizKeyFn func(ctx context.Context, bizKey int64) (*model.MainItem, error)
}

func (m *mockMainItemRepoForDL) Create(ctx context.Context, item *model.MainItem) error { return nil }
func (m *mockMainItemRepoForDL) FindByID(ctx context.Context, id uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDL) FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDL) FindByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error) {
	return m.findByBizKeyFn(ctx, bizKey)
}
func (m *mockMainItemRepoForDL) FindByBizKeys(ctx context.Context, bizKeys []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDL) Update(ctx context.Context, item *model.MainItem, fields map[string]interface{}) error {
	return nil
}
func (m *mockMainItemRepoForDL) NextCode(ctx context.Context, teamBizKey int64) (string, error) {
	return "", nil
}
func (m *mockMainItemRepoForDL) List(ctx context.Context, teamBizKey int64, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemRepoForDL) ListByTeamAndStatus(ctx context.Context, teamBizKey int64, status string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDL) CountByTeam(ctx context.Context, teamBizKey int64) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepoForDL) ListNonArchivedByTeam(ctx context.Context, teamBizKey int64) ([]model.MainItem, error) {
	return nil, nil
}

// Verify interfaces at compile time
var _ repository.DecisionLogRepo = (*mockDecisionLogRepo)(nil)
var _ repository.MainItemRepo = (*mockMainItemRepoForDL)(nil)

// --- Tests ---

func TestDecisionLogService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success_with_draft", func(t *testing.T) {
		mainItemRepo := &mockMainItemRepoForDL{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.MainItem, error) {
				return &model.MainItem{BaseModel: model.BaseModel{ID: 1, BizKey: bizKey}}, nil
			},
		}
		dlRepo := &mockDecisionLogRepo{
			createFn: func(ctx context.Context, log *model.DecisionLog) error {
				assert.Equal(t, "draft", log.LogStatus)
				assert.Equal(t, int64(100), log.MainItemKey)
				assert.Equal(t, int64(10), log.CreatedBy)
				assert.Equal(t, "technical", log.Category)
				assert.Equal(t, "test content", log.Content)

				// Verify tags are serialized to JSON
				var tags []string
				require.NoError(t, json.Unmarshal([]byte(log.Tags), &tags))
				assert.Equal(t, []string{"tag1", "tag2"}, tags)

				assert.NotZero(t, log.BizKey) // snowflake generated
				return nil
			},
		}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		req := dto.DecisionLogCreateReq{
			Category:  "technical",
			Tags:      []string{"tag1", "tag2"},
			Content:   "test content",
			LogStatus: "draft",
		}
		result, err := svc.Create(ctx, 100, 10, req)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("success_with_published", func(t *testing.T) {
		mainItemRepo := &mockMainItemRepoForDL{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.MainItem, error) {
				return &model.MainItem{}, nil
			},
		}
		var captured *model.DecisionLog
		dlRepo := &mockDecisionLogRepo{
			createFn: func(ctx context.Context, log *model.DecisionLog) error {
				captured = log
				return nil
			},
		}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		req := dto.DecisionLogCreateReq{
			Category:  "risk",
			Content:   "published content",
			LogStatus: "published",
		}
		result, err := svc.Create(ctx, 100, 10, req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "published", captured.LogStatus)
	})

	t.Run("main_item_not_found", func(t *testing.T) {
		mainItemRepo := &mockMainItemRepoForDL{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.MainItem, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		dlRepo := &mockDecisionLogRepo{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		req := dto.DecisionLogCreateReq{
			Category:  "technical",
			Content:   "test",
			LogStatus: "draft",
		}
		_, err := svc.Create(ctx, 999, 10, req)
		assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
	})

	t.Run("sets_team_key_from_main_item", func(t *testing.T) {
		mainItemRepo := &mockMainItemRepoForDL{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.MainItem, error) {
				return &model.MainItem{BaseModel: model.BaseModel{BizKey: 100}, TeamKey: 50}, nil
			},
		}
		var captured *model.DecisionLog
		dlRepo := &mockDecisionLogRepo{
			createFn: func(ctx context.Context, log *model.DecisionLog) error {
				captured = log
				return nil
			},
		}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		req := dto.DecisionLogCreateReq{
			Category:  "technical",
			Content:   "test",
			LogStatus: "draft",
		}
		_, err := svc.Create(ctx, 100, 10, req)
		require.NoError(t, err)
		assert.Equal(t, int64(50), captured.TeamKey)
	})
}

func TestDecisionLogService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("success_owner_edits_draft", func(t *testing.T) {
		existing := &model.DecisionLog{
			ID:        1,
			BizKey:    2001,
			LogStatus: "draft",
			CreatedBy: 10,
		}
		dlRepo := &mockDecisionLogRepo{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
				return existing, nil
			},
			updateFn: func(ctx context.Context, log *model.DecisionLog) error {
				assert.Equal(t, "risk", log.Category)
				assert.Equal(t, "updated content", log.Content)
				return nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		req := dto.DecisionLogUpdateReq{
			Category: "risk",
			Tags:     []string{"new"},
			Content:  "updated content",
		}
		result, err := svc.Update(ctx, 2001, 10, req)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("forbidden_if_not_draft", func(t *testing.T) {
		existing := &model.DecisionLog{
			ID:        1,
			BizKey:    2001,
			LogStatus: "published",
			CreatedBy: 10,
		}
		dlRepo := &mockDecisionLogRepo{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
				return existing, nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		req := dto.DecisionLogUpdateReq{
			Category: "technical",
			Content:  "attempt update",
		}
		_, err := svc.Update(ctx, 2001, 10, req)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})

	t.Run("forbidden_if_not_owner", func(t *testing.T) {
		existing := &model.DecisionLog{
			ID:        1,
			BizKey:    2001,
			LogStatus: "draft",
			CreatedBy: 10,
		}
		dlRepo := &mockDecisionLogRepo{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
				return existing, nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		req := dto.DecisionLogUpdateReq{
			Category: "technical",
			Content:  "attempt update",
		}
		_, err := svc.Update(ctx, 2001, 99, req) // userID=99, owner=10
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})

	t.Run("not_found", func(t *testing.T) {
		dlRepo := &mockDecisionLogRepo{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		req := dto.DecisionLogUpdateReq{
			Category: "technical",
			Content:  "test",
		}
		_, err := svc.Update(ctx, 9999, 10, req)
		assert.ErrorIs(t, err, apperrors.ErrDecisionLogNotFound)
	})

	t.Run("serializes_tags_on_update", func(t *testing.T) {
		existing := &model.DecisionLog{
			ID:        1,
			BizKey:    2001,
			LogStatus: "draft",
			CreatedBy: 10,
		}
		var captured *model.DecisionLog
		dlRepo := &mockDecisionLogRepo{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
				return existing, nil
			},
			updateFn: func(ctx context.Context, log *model.DecisionLog) error {
				captured = log
				return nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		req := dto.DecisionLogUpdateReq{
			Category: "technical",
			Tags:     []string{"a", "b"},
			Content:  "test",
		}
		_, err := svc.Update(ctx, 2001, 10, req)
		require.NoError(t, err)
		var tags []string
		require.NoError(t, json.Unmarshal([]byte(captured.Tags), &tags))
		assert.Equal(t, []string{"a", "b"}, tags)
	})
}

func TestDecisionLogService_Publish(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		existing := &model.DecisionLog{
			ID:        1,
			BizKey:    3001,
			LogStatus: "draft",
			CreatedBy: 10,
		}
		var captured *model.DecisionLog
		dlRepo := &mockDecisionLogRepo{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
				return existing, nil
			},
			updateFn: func(ctx context.Context, log *model.DecisionLog) error {
				captured = log
				return nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		result, err := svc.Publish(ctx, 3001, 10)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "published", captured.LogStatus)
	})

	t.Run("forbidden_if_not_draft", func(t *testing.T) {
		existing := &model.DecisionLog{
			ID:        1,
			BizKey:    3001,
			LogStatus: "published",
			CreatedBy: 10,
		}
		dlRepo := &mockDecisionLogRepo{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
				return existing, nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		_, err := svc.Publish(ctx, 3001, 10)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})

	t.Run("forbidden_if_not_owner", func(t *testing.T) {
		existing := &model.DecisionLog{
			ID:        1,
			BizKey:    3001,
			LogStatus: "draft",
			CreatedBy: 10,
		}
		dlRepo := &mockDecisionLogRepo{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
				return existing, nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		_, err := svc.Publish(ctx, 3001, 99)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})

	t.Run("not_found", func(t *testing.T) {
		dlRepo := &mockDecisionLogRepo{
			findByBizKeyFn: func(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		_, err := svc.Publish(ctx, 9999, 10)
		assert.ErrorIs(t, err, apperrors.ErrDecisionLogNotFound)
	})
}

func TestDecisionLogService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("returns_paginated_results", func(t *testing.T) {
		logs := []model.DecisionLog{
			{ID: 1, BizKey: 100, Content: "first"},
			{ID: 2, BizKey: 101, Content: "second"},
		}
		dlRepo := &mockDecisionLogRepo{
			listByItemFn: func(ctx context.Context, mainItemID uint, userID uint, offset, limit int) ([]model.DecisionLog, int64, error) {
				assert.Equal(t, uint(100), mainItemID)
				assert.Equal(t, uint(10), userID)
				assert.Equal(t, 0, offset)
				assert.Equal(t, 20, limit)
				return logs, int64(2), nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		result, err := svc.List(ctx, 100, 10, dto.Pagination{Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Equal(t, 2, len(result.Items))
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.Size)
	})

	t.Run("applies_default_pagination", func(t *testing.T) {
		dlRepo := &mockDecisionLogRepo{
			listByItemFn: func(ctx context.Context, mainItemID uint, userID uint, offset, limit int) ([]model.DecisionLog, int64, error) {
				assert.Equal(t, 0, offset) // default page=1 → offset=0
				assert.Equal(t, 20, limit) // default pageSize=20
				return []model.DecisionLog{}, int64(0), nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		result, err := svc.List(ctx, 100, 10, dto.Pagination{})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.Size)
	})

	t.Run("page2_with_custom_page_size", func(t *testing.T) {
		dlRepo := &mockDecisionLogRepo{
			listByItemFn: func(ctx context.Context, mainItemID uint, userID uint, offset, limit int) ([]model.DecisionLog, int64, error) {
				assert.Equal(t, 5, offset) // (2-1)*5=5
				assert.Equal(t, 5, limit)
				return []model.DecisionLog{}, int64(10), nil
			},
		}
		mainItemRepo := &mockMainItemRepoForDL{}
		svc := NewDecisionLogService(dlRepo, mainItemRepo)

		result, err := svc.List(ctx, 100, 10, dto.Pagination{Page: 2, PageSize: 5})
		require.NoError(t, err)
		assert.Equal(t, 2, result.Page)
		assert.Equal(t, 5, result.Size)
		assert.Equal(t, int64(10), result.Total)
	})
}
