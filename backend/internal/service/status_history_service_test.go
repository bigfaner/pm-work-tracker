package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
)

// mockStatusHistoryRepo uses testify/mock to satisfy repository.StatusHistoryRepo.
type mockStatusHistoryRepo struct {
	mock.Mock
}

func (m *mockStatusHistoryRepo) Create(ctx context.Context, record *model.StatusHistory) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockStatusHistoryRepo) FindByID(ctx context.Context, id uint) (*model.StatusHistory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StatusHistory), args.Error(1)
}

func (m *mockStatusHistoryRepo) ListByItem(ctx context.Context, itemType string, itemID uint, page dto.Pagination) (*dto.PageResult[model.StatusHistory], error) {
	args := m.Called(ctx, itemType, itemID, page)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PageResult[model.StatusHistory]), args.Error(1)
}

// --- Constructor ---

func TestNewStatusHistoryService_NilRepo(t *testing.T) {
	assert.Panics(t, func() {
		NewStatusHistoryService(nil)
	})
}

// --- Record ---

func TestStatusHistoryService_Record_Success(t *testing.T) {
	repo := new(mockStatusHistoryRepo)
	svc := NewStatusHistoryService(repo)

	record := &model.StatusHistory{
		ItemType:   "sub_item",
		ItemID:     1,
		FromStatus: "pending",
		ToStatus:   "progressing",
		ChangedBy:  10,
	}
	repo.On("Create", mock.Anything, record).Return(nil)

	err := svc.Record(context.Background(), record)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestStatusHistoryService_Record_RepoError(t *testing.T) {
	repo := new(mockStatusHistoryRepo)
	svc := NewStatusHistoryService(repo)

	record := &model.StatusHistory{
		ItemType:   "sub_item",
		ItemID:     1,
		FromStatus: "pending",
		ToStatus:   "progressing",
		ChangedBy:  10,
	}
	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	err := svc.Record(context.Background(), record)
	assert.Error(t, err)

	repo.AssertExpectations(t)
}

// --- ListByItem ---

func TestStatusHistoryService_ListByItem_Success(t *testing.T) {
	repo := new(mockStatusHistoryRepo)
	svc := NewStatusHistoryService(repo)

	records := []model.StatusHistory{
		{ID: 1, ItemType: "sub_item", ItemID: 1, FromStatus: "pending", ToStatus: "progressing"},
		{ID: 2, ItemType: "sub_item", ItemID: 1, FromStatus: "progressing", ToStatus: "blocking"},
	}
	repo.On("ListByItem", mock.Anything, "sub_item", uint(1), dto.Pagination{Page: 1, PageSize: 20}).
		Return(&dto.PageResult[model.StatusHistory]{Items: records, Total: 2}, nil)

	result, err := svc.ListByItem(context.Background(), "sub_item", 1, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)

	repo.AssertExpectations(t)
}

func TestStatusHistoryService_ListByItem_RepoError(t *testing.T) {
	repo := new(mockStatusHistoryRepo)
	svc := NewStatusHistoryService(repo)

	repo.On("ListByItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	_, err := svc.ListByItem(context.Background(), "sub_item", 1, dto.Pagination{Page: 1, PageSize: 20})
	assert.Error(t, err)

	repo.AssertExpectations(t)
}

func TestStatusHistoryService_ListByItem_Empty(t *testing.T) {
	repo := new(mockStatusHistoryRepo)
	svc := NewStatusHistoryService(repo)

	repo.On("ListByItem", mock.Anything, "sub_item", uint(99), mock.Anything).
		Return(&dto.PageResult[model.StatusHistory]{Items: []model.StatusHistory{}, Total: 0}, nil)

	result, err := svc.ListByItem(context.Background(), "sub_item", 99, dto.Pagination{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)

	repo.AssertExpectations(t)
}
