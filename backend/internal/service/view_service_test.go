package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock repos for ViewService tests
// ---------------------------------------------------------------------------

type mockViewMainItemRepo struct {
	items    []model.MainItem
	listErr  error
	findErr  error
	findItem *model.MainItem

	// ListByTeamAndStatus tracking
	listByTeamAndStatusCalled bool
	listByTeamAndStatusTeamID int64
	listByTeamAndStatusStatus string
	listByTeamAndStatusResult []model.MainItem
	listByTeamAndStatusErr    error

	// ListNonArchivedByTeam tracking
	listNonArchivedCalled bool
}

func (m *mockViewMainItemRepo) Create(_ context.Context, _ *model.MainItem) error {
	return nil
}
func (m *mockViewMainItemRepo) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	if m.findItem != nil {
		return m.findItem, nil
	}
	return nil, m.findErr
}
func (m *mockViewMainItemRepo) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockViewMainItemRepo) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockViewMainItemRepo) NextCode(_ context.Context, _ int64) (string, error) {
	return "", nil
}
func (m *mockViewMainItemRepo) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockViewMainItemRepo) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	m.listNonArchivedCalled = true
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
}
func (m *mockViewMainItemRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockViewMainItemRepo) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockViewMainItemRepo) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockViewMainItemRepo) ListByTeamAndStatus(_ context.Context, teamBizKey int64, status string) ([]model.MainItem, error) {
	m.listByTeamAndStatusCalled = true
	m.listByTeamAndStatusTeamID = teamBizKey
	m.listByTeamAndStatusStatus = status
	if m.listByTeamAndStatusErr != nil {
		return nil, m.listByTeamAndStatusErr
	}
	if m.listByTeamAndStatusResult != nil {
		return m.listByTeamAndStatusResult, nil
	}
	return m.items, nil
}

type mockViewSubItemRepo struct {
	items   []model.SubItem
	listErr error
}

func (m *mockViewSubItemRepo) Create(_ context.Context, _ *model.SubItem) error {
	return nil
}
func (m *mockViewSubItemRepo) FindByID(_ context.Context, _ uint) (*model.SubItem, error) {
	return nil, nil
}
func (m *mockViewSubItemRepo) Update(_ context.Context, _ *model.SubItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockViewSubItemRepo) List(_ context.Context, _ int64, _ uint, _ dto.SubItemFilter, _ dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}
func (m *mockViewSubItemRepo) ListByMainItem(_ context.Context, _ uint) ([]*model.SubItem, error) {
	return nil, nil
}
func (m *mockViewSubItemRepo) ListByTeam(_ context.Context, _ int64) ([]model.SubItem, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
}
func (m *mockViewSubItemRepo) SoftDelete(_ context.Context, _ uint) error {
	return nil
}
func (m *mockViewSubItemRepo) FindByBizKey(_ context.Context, _ int64) (*model.SubItem, error) {
	return nil, nil
}
func (m *mockViewSubItemRepo) NextSubCode(_ context.Context, _ uint) (string, error) {
	return "", nil
}

type mockViewProgressRepo struct {
	records []model.ProgressRecord
	listErr error
}

func (m *mockViewProgressRepo) Create(_ context.Context, _ *model.ProgressRecord) error {
	return nil
}
func (m *mockViewProgressRepo) FindByID(_ context.Context, _ uint) (*model.ProgressRecord, error) {
	return nil, nil
}
func (m *mockViewProgressRepo) ListBySubItem(_ context.Context, _ int64, _ int64) ([]model.ProgressRecord, error) {
	return nil, nil
}
func (m *mockViewProgressRepo) LatestBySubItem(_ context.Context, _ int64) (*model.ProgressRecord, error) {
	return nil, nil
}
func (m *mockViewProgressRepo) UpdateCompletion(_ context.Context, _ uint, _ float64) error {
	return nil
}
func (m *mockViewProgressRepo) ListByTeamInRange(_ context.Context, _ int64, _, _ time.Time) ([]model.ProgressRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.records, nil
}
func (m *mockViewProgressRepo) FindByBizKey(_ context.Context, _ int64) (*model.ProgressRecord, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Tests: NewViewService constructor
// ---------------------------------------------------------------------------

func TestNewViewService_WithoutUserRepo(t *testing.T) {
	svc := NewViewService(&mockViewMainItemRepo{}, &mockViewSubItemRepo{}, &mockViewProgressRepo{})
	vs := svc.(*viewService)
	assert.Nil(t, vs.userRepo, "userRepo should be nil when not provided")
}

func TestNewViewService_WithUserRepo(t *testing.T) {
	ur := &mockViewUserRepo{}
	svc := NewViewService(&mockViewMainItemRepo{}, &mockViewSubItemRepo{}, &mockViewProgressRepo{}, ur)
	vs := svc.(*viewService)
	assert.NotNil(t, vs.userRepo, "userRepo should be set when provided")
}

// ---------------------------------------------------------------------------
// Tests: WeeklyComparison
// ---------------------------------------------------------------------------

func TestWeeklyComparison_RejectsNonMonday(t *testing.T) {
	svc := NewViewService(&mockViewMainItemRepo{}, &mockViewSubItemRepo{}, &mockViewProgressRepo{})

	tuesday := time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC)
	_, err := svc.WeeklyComparison(context.Background(), 1, tuesday)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestWeeklyComparison_RejectsFutureWeek(t *testing.T) {
	svc := NewViewService(&mockViewMainItemRepo{}, &mockViewSubItemRepo{}, &mockViewProgressRepo{})

	// A future Monday
	futureMonday := time.Date(2099, 1, 5, 0, 0, 0, 0, time.UTC)
	_, err := svc.WeeklyComparison(context.Background(), 1, futureMonday)
	assert.ErrorIs(t, err, apperrors.ErrFutureWeekNotAllowed)
}

func TestWeeklyComparison_AcceptsMonday(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	result, err := svc.WeeklyComparison(context.Background(), 1, monday)
	require.NoError(t, err)
	assert.Equal(t, "2026-04-13", result.WeekStart)
	assert.Equal(t, "2026-04-19", result.WeekEnd)
}

func TestWeeklyComparison_EmptyTeam_NoGroups(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	result, err := svc.WeeklyComparison(context.Background(), 1, monday)
	require.NoError(t, err)
	assert.Empty(t, result.Groups)
	assert.Equal(t, dto.WeeklyStats{}, result.Stats)
}

func TestWeeklyComparison_StatsCounts(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	completedDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "In Progress", ItemStatus: "progressing", Completion: 60},
			{BaseModel: model.BaseModel{ID: 11, BizKey: 11}, TeamKey: 1, MainItemKey: int64(1), Title: "Blocked", ItemStatus: "blocking", Completion: 30},
			{BaseModel: model.BaseModel{ID: 12, BizKey: 12}, TeamKey: 1, MainItemKey: int64(1), Title: "Completed", ItemStatus: "completed", Completion: 100, ActualEndDate: &completedDate},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 60, CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
			{ID: 101, SubItemKey: 12, TeamKey: 1, Completion: 100, CreateTime: time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	assert.Equal(t, 3, result.Stats.ActiveSubItems) // 2 non-completed + 1 just-completed
	assert.Equal(t, 1, result.Stats.NewlyCompleted)  // sub 12
	assert.Equal(t, 1, result.Stats.InProgress)      // sub 10
	assert.Equal(t, 1, result.Stats.Blocked)         // sub 11
}

func TestWeeklyComparison_DeltaComputation(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	lastWeekStart := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", ItemStatus: "progressing", Completion: 70},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			// Last week: completion was 40
			{ID: 90, SubItemKey: 10, TeamKey: 1, Completion: 40, CreateTime: lastWeekStart.AddDate(0, 0, 2)},
			// This week: completion is 70
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 70, CreateTime: weekStart.AddDate(0, 0, 1)},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]

	// Should have lastWeek snapshot
	require.Len(t, group.LastWeek, 1)
	assert.Equal(t, 40.0, group.LastWeek[0].Completion)

	// Should have thisWeek snapshot with delta
	require.Len(t, group.ThisWeek, 1)
	assert.Equal(t, 70.0, group.ThisWeek[0].Completion)
	assert.Equal(t, 30.0, group.ThisWeek[0].Delta) // 70 - 40
	assert.False(t, group.ThisWeek[0].IsNew)
	assert.False(t, group.ThisWeek[0].JustCompleted)
}

func TestWeeklyComparison_IsNew(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "New Sub", ItemStatus: "pending", Completion: 0},
		},
	}
	// Only this week progress, no last week progress
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 0, CreateTime: weekStart.AddDate(0, 0, 1)},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]

	require.Len(t, group.ThisWeek, 1)
	assert.True(t, group.ThisWeek[0].IsNew)
	assert.Empty(t, group.LastWeek)
}

func TestWeeklyComparison_JustCompleted(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	completedDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:         model.BaseModel{ID: 10, BizKey: 10},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:         "Just Done",
				ItemStatus: "completed",
				Completion:    100,
				ActualEndDate: &completedDate,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 100, CreateTime: completedDate},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]

	require.Len(t, group.ThisWeek, 1)
	assert.True(t, group.ThisWeek[0].JustCompleted)
}

func TestWeeklyComparison_CompletedNoChange(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	// Completed before this week
	oldEndDate := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:         model.BaseModel{ID: 10, BizKey: 10},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:         "Old Completed",
				ItemStatus: "completed",
				Completion:    100,
				ActualEndDate: &oldEndDate,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]

	assert.Empty(t, group.ThisWeek)
	require.Len(t, group.CompletedNoChange, 1)
	assert.Equal(t, "10", group.CompletedNoChange[0].BizKey)
}

func TestWeeklyComparison_GroupsSortedByPriority(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "P3 Item", Priority: "P3", Completion: 10},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Title: "P1 Item", Priority: "P1", Completion: 50},
			{BaseModel: model.BaseModel{ID: 3, BizKey: 3}, TeamKey: 1, Title: "P2 Item", Priority: "P2", Completion: 30},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub 1", ItemStatus: "progressing", Completion: 10},
			{BaseModel: model.BaseModel{ID: 20, BizKey: 20}, TeamKey: 1, MainItemKey: int64(2), Title: "Sub 2", ItemStatus: "progressing", Completion: 50},
			{BaseModel: model.BaseModel{ID: 30, BizKey: 30}, TeamKey: 1, MainItemKey: int64(3), Title: "Sub 3", ItemStatus: "progressing", Completion: 30},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 3)
	assert.Equal(t, "P1 Item", result.Groups[0].MainItem.Title)
	assert.Equal(t, "P2 Item", result.Groups[1].MainItem.Title)
	assert.Equal(t, "P3 Item", result.Groups[2].MainItem.Title)
}

func TestWeeklyComparison_MainItemSummary(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Title:           "Main 1",
				Priority:        "P1",
				PlanStartDate: &startDate,
				ExpectedEndDate: &endDate,
				Completion:      58,
			},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", ItemStatus: "progressing", Completion: 60},
			{BaseModel: model.BaseModel{ID: 11, BizKey: 11}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub B", ItemStatus: "pending", Completion: 0},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	mi := result.Groups[0].MainItem
	assert.Equal(t, "1", mi.BizKey)
	assert.Equal(t, "Main 1", mi.Title)
	assert.Equal(t, "P1", mi.Priority)
	assert.Equal(t, "2026-04-01", mi.StartDate)
	assert.Equal(t, "2026-04-30", mi.ExpectedEndDate)
	assert.Equal(t, 58.0, mi.Completion)
	assert.Equal(t, 2, mi.SubItemCount)
}

func TestWeeklyComparison_MainItemWithNoSubItems_Omitted(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Empty Main", Completion: 0},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)
	assert.Empty(t, result.Groups)
}

func TestWeeklyComparison_ProgressDescriptionFromLatestRecord(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", ItemStatus: "progressing", Completion: 60},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{
				ID:         100,
				SubItemKey:  10,
				TeamKey: 1,
				Completion: 60,
				Achievement: "Token sign done",
				Blocker:     "Blacklist WIP",
				CreateTime:   weekStart.AddDate(0, 0, 1),
			},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	require.Len(t, result.Groups[0].ThisWeek, 1)
	assert.Equal(t, "Token sign done; Blacklist WIP", result.Groups[0].ThisWeek[0].ProgressDescription)
}

func TestWeeklyComparison_SnapshotCodePropagated(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	lastWeekStart := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main", Priority: "P1"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Code: "ALPHA-00001-01", Title: "Sub", ItemStatus: "progressing", Completion: 60},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 1, SubItemKey: 10, TeamKey: 1, Completion: 40, CreateTime: lastWeekStart.AddDate(0, 0, 1)},
			{ID: 2, SubItemKey: 10, TeamKey: 1, Completion: 60, CreateTime: weekStart.AddDate(0, 0, 1)},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	require.Len(t, group.ThisWeek, 1)
	assert.Equal(t, "ALPHA-00001-01", group.ThisWeek[0].Code)
	require.Len(t, group.LastWeek, 1)
	assert.Equal(t, "ALPHA-00001-01", group.LastWeek[0].Code)
}

func TestWeeklyComparison_RepoErrors(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	// MainItemRepo error
	svc := NewViewService(&mockViewMainItemRepo{listErr: errors.New("db error")}, &mockViewSubItemRepo{}, &mockViewProgressRepo{})
	_, err := svc.WeeklyComparison(context.Background(), 1, monday)
	assert.Error(t, err)

	// SubItemRepo error
	svc = NewViewService(
		&mockViewMainItemRepo{items: []model.MainItem{{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1"}}},
		&mockViewSubItemRepo{listErr: errors.New("db error")},
		&mockViewProgressRepo{},
	)
	_, err = svc.WeeklyComparison(context.Background(), 1, monday)
	assert.Error(t, err)

	// ProgressRepo error
	svc = NewViewService(
		&mockViewMainItemRepo{items: []model.MainItem{{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1"}}},
		&mockViewSubItemRepo{items: []model.SubItem{{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A"}}},
		&mockViewProgressRepo{listErr: errors.New("db error")},
	)
	_, err = svc.WeeklyComparison(context.Background(), 1, monday)
	assert.Error(t, err)
}

func TestWeeklyComparison_SubItemCreatedAfterWeek_NotActive(t *testing.T) {
	// Viewing week April 6-12, but sub-item was created April 14 (after that week)
	weekStart := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:  model.BaseModel{ID: 10, BizKey: 10, CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:      "Future Sub",
				ItemStatus: "progressing",
				Completion: 60,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	// Main item with no active sub-items in the viewed week should be omitted entirely
	assert.Empty(t, result.Groups)
}

func TestWeeklyComparison_SubItemCompletedBeforeWeek_NotActive(t *testing.T) {
	// Viewing week April 13-19, but sub-item was completed April 10 (before that week)
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	actualEnd := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:      model.BaseModel{ID: 10, BizKey: 10, CreateTime: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:          "Old Completed Sub",
				ItemStatus: "completed",
				Completion:     100,
				ActualEndDate:  &actualEnd,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	// Sub-item completed before the week should go to completedNoChange, not thisWeek
	assert.Empty(t, group.ThisWeek)
	assert.Empty(t, group.LastWeek)
	require.Len(t, group.CompletedNoChange, 1)
	assert.Equal(t, "10", group.CompletedNoChange[0].BizKey)
}

func TestWeeklyComparison_SubItemCreatedAfterWeek_Completed_NotShown(t *testing.T) {
	// Viewing week April 6-12, but sub-item was created and completed on April 14 (after that week)
	weekStart := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	actualEnd := time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:     model.BaseModel{ID: 10, BizKey: 10, CreateTime: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:         "Future Completed Sub",
				ItemStatus: "completed",
				Completion:    100,
				ActualEndDate: &actualEnd,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	// Main item should be omitted entirely — sub-item didn't exist during the viewed week
	assert.Empty(t, result.Groups)
}

func TestWeeklyComparison_SubItemCompletedAfterWeek_NotShown(t *testing.T) {
	// Viewing week Dec 1-7 2025, but sub-item was completed in April 2026 (after that week)
	weekStart := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	actualEnd := time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Test Item", Priority: "P2", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:     model.BaseModel{ID: 10, BizKey: 10, CreateTime: time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC)},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:         "Future Sub",
				ItemStatus: "completed",
				Completion:    100,
				ActualEndDate: &actualEnd,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	// Main item should not appear — sub-item was created and completed after the viewed week
	assert.Empty(t, result.Groups)
}

func TestWeeklyComparison_StatsPending(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Pending Sub", ItemStatus: "pending"},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 0, CreateTime: weekStart.AddDate(0, 0, 1)},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Stats.Pending)
	assert.Equal(t, 0, result.Stats.Pausing)
}

func TestWeeklyComparison_StatsPausing(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Pausing Sub", ItemStatus: "pausing"},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 20, CreateTime: weekStart.AddDate(0, 0, 1)},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	assert.Equal(t, 0, result.Stats.Pending)
	assert.Equal(t, 1, result.Stats.Pausing)
}

func TestWeeklyComparison_StatsOverdue(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)
	overdueDate := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC) // before weekEnd

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Overdue Sub", ItemStatus: "progressing", ExpectedEndDate: &overdueDate},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 30, CreateTime: weekStart.AddDate(0, 0, 1)},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Stats.Overdue)
	_ = weekEnd // used implicitly via weekStart+6d
}

func TestWeeklyComparison_StatsOverdue_NilExpectedEndDate_NotCounted(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			// ExpectedEndDate is nil — must NOT be counted as overdue
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "No Deadline Sub", ItemStatus: "progressing", ExpectedEndDate: nil},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 30, CreateTime: weekStart.AddDate(0, 0, 1)},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	assert.Equal(t, 0, result.Stats.Overdue)
}

func TestWeeklyComparison_StatsOverdue_CompletedNotCounted(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	overdueDate := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	completedDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			// completed — must NOT be counted as overdue even if past deadline
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Completed Sub", ItemStatus: "completed", ExpectedEndDate: &overdueDate, ActualEndDate: &completedDate, Completion: 100},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 100, CreateTime: completedDate},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	assert.Equal(t, 0, result.Stats.Overdue)
}

func TestWeeklyComparison_StatsOverdue_ClosedNotCounted(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	overdueDate := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			// closed — must NOT be counted as overdue
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Closed Sub", ItemStatus: "closed", ExpectedEndDate: &overdueDate},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 0, CreateTime: weekStart.AddDate(0, 0, 1)},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	assert.Equal(t, 0, result.Stats.Overdue)
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestGanttView_EmptyTeam_NoItems(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)
	assert.Empty(t, result.Items)
}

func TestGanttView_BasicStructure(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Title:           "Main 1",
				Priority:        "P1",
				PlanStartDate: &startDate,
				ExpectedEndDate: &endDate,
				Completion:      45.5,
				ItemStatus: "progressing",
			},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	item := result.Items[0]
	assert.Equal(t, "1", item.BizKey)
	assert.Equal(t, "Main 1", item.Title)
	assert.Equal(t, "P1", item.Priority)
	assert.Equal(t, "2026-04-01", item.StartDate)
	assert.Equal(t, "2026-04-30", item.ExpectedEndDate)
	assert.Equal(t, 45.5, item.Completion)
	assert.Equal(t, "progressing", item.Status)
	assert.False(t, item.IsOverdue)
}

func TestGanttView_OverdueItem(t *testing.T) {
	// expectedEndDate is in the past, status is not 已完成 or 已关闭
	pastDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Title:           "Overdue Main",
				Priority:        "P1",
				PlanStartDate: &startDate,
				ExpectedEndDate: &pastDate,
				ItemStatus: "progressing",
			},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.True(t, result.Items[0].IsOverdue)
}

func TestGanttView_Overdue_ExemptWhenCompleted(t *testing.T) {
	// expectedEndDate is in the past but status is 已完成
	pastDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	for _, status := range []string{"completed", "closed"} {
		t.Run(status, func(t *testing.T) {
			mainRepo := &mockViewMainItemRepo{
				items: []model.MainItem{
					{
						BaseModel:           model.BaseModel{ID: 1, BizKey: 1},
						TeamKey: 1,
						Title:           "Completed",
						Priority:        "P1",
						PlanStartDate: &startDate,
						ExpectedEndDate: &pastDate,
						ItemStatus: status,
					},
				},
			}
			svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

			result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
			require.NoError(t, err)

			require.Len(t, result.Items, 1)
			assert.False(t, result.Items[0].IsOverdue, "status=%s should not be overdue", status)
		})
	}
}

func TestGanttView_StatusFilter(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "In Progress", ItemStatus: "progressing", PlanStartDate: &startDate, ExpectedEndDate: &endDate},
		},
		listByTeamAndStatusResult: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "In Progress", ItemStatus: "progressing", PlanStartDate: &startDate, ExpectedEndDate: &endDate},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{Status: "progressing"})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.Equal(t, "1", result.Items[0].BizKey)
	assert.Equal(t, "In Progress", result.Items[0].Title)
}

func TestGanttView_StatusFilter_UsesSQLPushdown(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		listByTeamAndStatusResult: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "In Progress", ItemStatus: "progressing", PlanStartDate: &startDate, ExpectedEndDate: &endDate},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	_, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{Status: "progressing"})
	require.NoError(t, err)

	// Should call ListByTeamAndStatus (SQL pushdown)
	assert.True(t, mainRepo.listByTeamAndStatusCalled, "should call ListByTeamAndStatus when status filter is set")
	assert.Equal(t, int64(1), mainRepo.listByTeamAndStatusTeamID)
	assert.Equal(t, "progressing", mainRepo.listByTeamAndStatusStatus)
	// Should NOT call ListNonArchivedByTeam
	assert.False(t, mainRepo.listNonArchivedCalled, "should not call ListNonArchivedByTeam when status filter is set")
}

func TestGanttView_NoStatusFilter_UsesListNonArchived(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "A", ItemStatus: "progressing", PlanStartDate: &startDate, ExpectedEndDate: &endDate},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	_, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	// Should call ListNonArchivedByTeam (no filter)
	assert.True(t, mainRepo.listNonArchivedCalled, "should call ListNonArchivedByTeam when no status filter")
	// Should NOT call ListByTeamAndStatus
	assert.False(t, mainRepo.listByTeamAndStatusCalled, "should not call ListByTeamAndStatus when no status filter")
}

func TestGanttView_StatusFilter_SQLPushdownError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{
		listByTeamAndStatusErr: errors.New("db error"),
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	_, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{Status: "progressing"})
	assert.Error(t, err)
}

func TestGanttView_StatusFilterEmpty_ReturnsAll(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "A", ItemStatus: "progressing", PlanStartDate: &startDate, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Title: "B", ItemStatus: "completed", PlanStartDate: &startDate, ExpectedEndDate: &endDate},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 2)
}

func TestGanttView_SubItemsNested(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Title:           "Main 1",
				Priority:        "P1",
				PlanStartDate: &startDate,
				ExpectedEndDate: &endDate,
				ItemStatus: "progressing",
			},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:           model.BaseModel{ID: 10, BizKey: 10},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:           "Sub A",
				PlanStartDate: &startDate,
				ExpectedEndDate: &endDate,
				Completion:      80,
				ItemStatus: "reviewing",
			},
		},
	}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	require.Len(t, result.Items[0].SubItems, 1)
	sub := result.Items[0].SubItems[0]
	assert.Equal(t, "10", sub.BizKey)
	assert.Equal(t, "Sub A", sub.Title)
	assert.Equal(t, "2026-04-01", sub.StartDate)
	assert.Equal(t, "2026-04-30", sub.ExpectedEndDate)
	assert.Equal(t, 80.0, sub.Completion)
	assert.Equal(t, "reviewing", sub.Status)
}

func TestGanttView_SubItemsFromOtherMainItemNotIncluded(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", PlanStartDate: &startDate, ExpectedEndDate: &endDate, ItemStatus: "progressing"},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Title: "Main 2", PlanStartDate: &startDate, ExpectedEndDate: &endDate, ItemStatus: "progressing"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub for Main 1", PlanStartDate: &startDate, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 20, BizKey: 20}, TeamKey: 1, MainItemKey: int64(2), Title: "Sub for Main 2", PlanStartDate: &startDate, ExpectedEndDate: &endDate},
		},
	}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 2)
	require.Len(t, result.Items[0].SubItems, 1)
	assert.Equal(t, "10", result.Items[0].SubItems[0].BizKey)
	require.Len(t, result.Items[1].SubItems, 1)
	assert.Equal(t, "20", result.Items[1].SubItems[0].BizKey)
}

func TestGanttView_ArchivedItemsExcluded(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	// The repo's ListNonArchivedByTeam already filters out archived items,
	// so the mock only returns non-archived items.
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Title:           "Active",
				PlanStartDate: &startDate,
				ExpectedEndDate: &endDate,
				ItemStatus: "progressing",
			},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.Equal(t, "1", result.Items[0].BizKey)
	assert.Equal(t, "Active", result.Items[0].Title)
}

func TestGanttView_DatesFormattedAsISO8601(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Title:           "Main 1",
				PlanStartDate: &startDate,
				ExpectedEndDate: &endDate,
				ItemStatus: "progressing",
			},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:           model.BaseModel{ID: 10, BizKey: 10},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:           "Sub 1",
				PlanStartDate: &startDate,
				ExpectedEndDate: &endDate,
				ItemStatus: "progressing",
			},
		},
	}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	item := result.Items[0]
	assert.Equal(t, "2026-04-01", item.StartDate)
	assert.Equal(t, "2026-04-30", item.ExpectedEndDate)
	assert.Equal(t, "2026-04-01", item.SubItems[0].StartDate)
	assert.Equal(t, "2026-04-30", item.SubItems[0].ExpectedEndDate)
}

func TestGanttView_NilDates_FormatAsEmptyString(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:  model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Title:  "No dates",
				ItemStatus: "pending",
			},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.Equal(t, "", result.Items[0].StartDate)
	assert.Equal(t, "", result.Items[0].ExpectedEndDate)
}

func TestGanttView_MainItemRepoError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{listErr: errors.New("db error")}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{}, &mockViewProgressRepo{})

	_, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	assert.Error(t, err)
}

func TestGanttView_SubItemRepoError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1"},
		},
	}
	subRepo := &mockViewSubItemRepo{listErr: errors.New("db error")}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	_, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	assert.Error(t, err)
}

func TestGanttView_SubItemSummaryFieldsOnly(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", PlanStartDate: &startDate, ExpectedEndDate: &endDate, ItemStatus: "progressing"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:           model.BaseModel{ID: 10, BizKey: 10},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:           "Sub A",
				ItemDesc: "Should not appear in gantt sub-item DTO",
				Priority:        "P2",
				AssigneeKey: nil,
				PlanStartDate: &startDate,
				ExpectedEndDate: &endDate,
				ActualEndDate:   nil,
				ItemStatus: "progressing",
				Completion:      60,
			},
		},
	}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	require.Len(t, result.Items[0].SubItems, 1)
	sub := result.Items[0].SubItems[0]
	// Only these fields should be present
	assert.Equal(t, "10", sub.BizKey)
	assert.Equal(t, "Sub A", sub.Title)
	assert.Equal(t, "2026-04-01", sub.StartDate)
	assert.Equal(t, "2026-04-30", sub.ExpectedEndDate)
	assert.Equal(t, 60.0, sub.Completion)
	assert.Equal(t, "progressing", sub.Status)
	// GanttSubItemDTO should NOT have isOverdue (v1)
	// This is enforced by struct definition
}

func TestGanttView_Overdue_NilExpectedEndDate(t *testing.T) {
	// No expected end date set — should not be overdue
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:  model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Title:  "No end date",
				ItemStatus: "progressing",
			},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.False(t, result.Items[0].IsOverdue)
}

// ---------------------------------------------------------------------------
// Mock UserRepo for table view tests
// ---------------------------------------------------------------------------

type mockViewUserRepo struct {
	users         map[uint]*model.User
	findByIDCalls uint
	findByIDsCalls uint
	findByIDsArg  []uint
}

func (m *mockViewUserRepo) FindByID(_ context.Context, id uint) (*model.User, error) {
	m.findByIDCalls++
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}
func (m *mockViewUserRepo) FindByUsername(_ context.Context, _ string) (*model.User, error) {
	return nil, nil
}
func (m *mockViewUserRepo) List(_ context.Context) ([]*model.User, error) {
	return nil, nil
}
func (m *mockViewUserRepo) Update(_ context.Context, _ *model.User) error {
	return nil
}

func (m *mockViewUserRepo) Create(_ context.Context, _ *model.User) error {
	return nil
}
func (m *mockViewUserRepo) FindByIDs(_ context.Context, ids []uint) (map[uint]*model.User, error) {
	m.findByIDsCalls++
	m.findByIDsArg = ids
	result := make(map[uint]*model.User)
	for _, id := range ids {
		if u, ok := m.users[id]; ok {
			result[id] = u
		}
	}
	return result, nil
}
func (m *mockViewUserRepo) FindByBizKey(_ context.Context, _ int64) (*model.User, error) {
	return nil, nil
}
func (m *mockViewUserRepo) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
	return nil, 0, nil
}
func (m *mockViewUserRepo) SearchAvailable(_ context.Context, _ int64, _ string, _ int) ([]*model.User, error) {
	return nil, nil
}
func (m *mockViewUserRepo) SoftDelete(_ context.Context, _ *model.User) error { return nil }

// newViewServiceWithUsers creates a ViewService with a user repo for table view tests.
func newViewServiceWithUsers(mainRepo *mockViewMainItemRepo, subRepo *mockViewSubItemRepo, userRepo *mockViewUserRepo) ViewService {
	return &viewService{
		mainItemRepo: mainRepo,
		subItemRepo:  subRepo,
		progressRepo: &mockViewProgressRepo{},
		userRepo:     userRepo,
	}
}

// ---------------------------------------------------------------------------
// Tests: TableView
// ---------------------------------------------------------------------------

func TestTableView_EmptyTeam_ReturnsEmpty(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	svc := newViewServiceWithUsers(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewUserRepo{})

	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
}

func TestTableView_CombinesMainAndSubItems(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	assigneeID := uint(100)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Code:            "TEST-00001",
				Title:           "Main Item",
				Priority:        "P1",
				AssigneeKey: func() *int64 { v := int64(assigneeID); return &v }(),
				ItemStatus: "progressing",
				Completion:      50,
				ExpectedEndDate: &endDate,
			},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:           model.BaseModel{ID: 10, BizKey: 10},
				TeamKey: 1,
				MainItemKey: int64(1),
				Title:           "Sub Item",
				Priority:        "P2",
				AssigneeKey: func() *int64 { v := int64(assigneeID); return &v }(),
				ItemStatus: "pending",
				Completion:      0,
				ExpectedEndDate: &endDate,
			},
		},
	}
	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			100: {BaseModel: model.BaseModel{ID: 100, BizKey: 100}, DisplayName: "Alice"},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, userRepo)
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(2), result.Total)
	require.Len(t, result.Items, 2)

	// Items should be sorted by default: priority DESC (P1 before P2)
	assert.Equal(t, "main", result.Items[0].Type)
	assert.Equal(t, "TEST-00001", result.Items[0].Code)
	assert.Equal(t, "Main Item", result.Items[0].Title)
	assert.Equal(t, "P1", result.Items[0].Priority)
	assert.Equal(t, "Alice", result.Items[0].AssigneeName)

	assert.Equal(t, "sub", result.Items[1].Type)
	assert.Equal(t, "P2", result.Items[1].Priority)
}

func TestTableView_FilterByTypeMain(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Main", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub", Priority: "P2", ItemStatus: "pending", Completion: 0, ExpectedEndDate: &endDate},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{Type: "main"}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "main", result.Items[0].Type)
}

func TestTableView_FilterByTypeSub(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Main", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub", Priority: "P2", ItemStatus: "pending", Completion: 0, ExpectedEndDate: &endDate},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{Type: "sub"}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "sub", result.Items[0].Type)
}

func TestTableView_FilterByPriority(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "P1 Item", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Code: "TEST-00002", Title: "P3 Item", Priority: "P3", ItemStatus: "pending", Completion: 0, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{Priority: []string{"P1"}}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "P1", result.Items[0].Priority)
}

func TestTableView_FilterByStatus(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Active", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Code: "TEST-00002", Title: "Done", Priority: "P1", ItemStatus: "completed", Completion: 100, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{Status: []string{"completed"}}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "completed", result.Items[0].Status)
}

func TestTableView_FilterByAssigneeID(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	assignee1 := uint(100)
	assignee2 := uint(200)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Alice's", Priority: "P1", AssigneeKey: func() *int64 { v := int64(assignee1); return &v }(), ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Code: "TEST-00002", Title: "Bob's", Priority: "P1", AssigneeKey: func() *int64 { v := int64(assignee2); return &v }(), ItemStatus: "progressing", Completion: 30, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{AssigneeKey: func() *string { v := fmt.Sprintf("%d", assignee1); return &v }()}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "100", *result.Items[0].AssigneeID)
}

func TestTableView_Pagination(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "A", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Code: "TEST-00002", Title: "B", Priority: "P2", ItemStatus: "progressing", Completion: 30, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 3, BizKey: 3}, TeamKey: 1, Code: "TEST-00003", Title: "C", Priority: "P3", ItemStatus: "pending", Completion: 0, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 2})
	require.NoError(t, err)

	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 2, result.Size)
	require.Len(t, result.Items, 2)
	// Default sort: priority DESC — P1 first, then P2
	assert.Equal(t, "P1", result.Items[0].Priority)
	assert.Equal(t, "P2", result.Items[1].Priority)
}

func TestTableView_DefaultSort_PriorityDescThenExpectedEndDateAsc(t *testing.T) {
	endDate1 := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate2 := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate3 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "P2 Late", Priority: "P2", ItemStatus: "progressing", Completion: 30, ExpectedEndDate: &endDate3},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Code: "TEST-00002", Title: "P2 Early", Priority: "P2", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate2},
			{BaseModel: model.BaseModel{ID: 3, BizKey: 3}, TeamKey: 1, Code: "TEST-00003", Title: "P1", Priority: "P1", ItemStatus: "progressing", Completion: 80, ExpectedEndDate: &endDate1},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, result.Items, 3)
	// P1 first, then P2 items sorted by ExpectedEndDate ASC
	assert.Equal(t, "P1", result.Items[0].Priority)
	assert.Equal(t, "P2 Early", result.Items[1].Title)
	assert.Equal(t, "P2 Late", result.Items[2].Title)
}

func TestTableView_SortByCompletion_Asc(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "High", Priority: "P1", ItemStatus: "progressing", Completion: 80, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Code: "TEST-00002", Title: "Low", Priority: "P1", ItemStatus: "pending", Completion: 10, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{SortBy: "completion", SortOrder: "asc"}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, result.Items, 2)
	assert.Equal(t, 10.0, result.Items[0].Completion)
	assert.Equal(t, 80.0, result.Items[1].Completion)
}

func TestTableView_SortByCompletion_Desc(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Low", Priority: "P1", ItemStatus: "progressing", Completion: 10, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Code: "TEST-00002", Title: "High", Priority: "P1", ItemStatus: "pending", Completion: 80, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{SortBy: "completion", SortOrder: "desc"}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, result.Items, 2)
	assert.Equal(t, 80.0, result.Items[0].Completion)
	assert.Equal(t, 10.0, result.Items[1].Completion)
}

func TestTableView_SubItemCodeFormat(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Code: "TEST-00001-01", Title: "Sub", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{Type: "sub"}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.Equal(t, "TEST-00001-01", result.Items[0].Code)
}

func TestTableView_AssigneeNameResolved(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	assigneeID := uint(50)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Main", Priority: "P1", AssigneeKey: func() *int64 { v := int64(assigneeID); return &v }(), ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}
	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			50: {BaseModel: model.BaseModel{ID: 50, BizKey: 50}, DisplayName: "Bob"},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, userRepo)
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.Equal(t, "Bob", result.Items[0].AssigneeName)
}

func TestTableView_AssigneeNameEmpty_WhenNoAssignee(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "No Assignee", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.Equal(t, "", result.Items[0].AssigneeName)
	assert.Nil(t, result.Items[0].AssigneeID)
}

func TestTableView_ExpectedEndDateAndActualEndDateFormatted(t *testing.T) {
	expectedEnd := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	actualEnd := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Main", Priority: "P1", ItemStatus: "completed", Completion: 100, ExpectedEndDate: &expectedEnd, ActualEndDate: &actualEnd},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	require.NotNil(t, result.Items[0].ExpectedEndDate)
	assert.Equal(t, "2026-05-01", *result.Items[0].ExpectedEndDate)
	require.NotNil(t, result.Items[0].ActualEndDate)
	assert.Equal(t, "2026-04-20", *result.Items[0].ActualEndDate)
}

func TestTableView_NilDates_ReturnNil(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "No Dates", Priority: "P1", ItemStatus: "pending", Completion: 0},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.Nil(t, result.Items[0].ExpectedEndDate)
	assert.Nil(t, result.Items[0].ActualEndDate)
}

func TestTableView_Page2_ReturnsSecondPage(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "A", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Code: "TEST-00002", Title: "B", Priority: "P2", ItemStatus: "progressing", Completion: 30, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 3, BizKey: 3}, TeamKey: 1, Code: "TEST-00003", Title: "C", Priority: "P3", ItemStatus: "pending", Completion: 0, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 2, PageSize: 2})
	require.NoError(t, err)

	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, 2, result.Page)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "P3", result.Items[0].Priority)
}

func TestTableView_MainItemRepoError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{listErr: errors.New("db error")}
	svc := newViewServiceWithUsers(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewUserRepo{})

	_, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	assert.Error(t, err)
}

func TestTableView_SubItemRepoError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	subRepo := &mockViewSubItemRepo{listErr: errors.New("db error")}
	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})

	_, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: TableExportCSV
// ---------------------------------------------------------------------------

func TestTableExportCSV_EmptyResult_ReturnsNoDataError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	svc := newViewServiceWithUsers(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewUserRepo{})

	_, err := svc.TableExportCSV(context.Background(), 1, dto.TableFilter{})
	assert.ErrorIs(t, err, apperrors.ErrNoData)
}

func TestTableExportCSV_ReturnsValidCSV(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	actualEnd := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	assigneeID := uint(100)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1, BizKey: 1},
				TeamKey: 1,
				Code:            "TEST-00001",
				Title:           "Main Item",
				Priority:        "P1",
				AssigneeKey: func() *int64 { v := int64(assigneeID); return &v }(),
				ItemStatus: "completed",
				Completion:      100,
				ExpectedEndDate: &endDate,
				ActualEndDate:   &actualEnd,
			},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}
	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			100: {BaseModel: model.BaseModel{ID: 100, BizKey: 100}, DisplayName: "Alice"},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, userRepo)
	data, err := svc.TableExportCSV(context.Background(), 1, dto.TableFilter{})
	require.NoError(t, err)

	// Strip BOM before CSV parsing
	csvData := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Header + 1 data row
	require.Len(t, records, 2)

	// Check header
	assert.Equal(t, []string{"编号", "标题", "类型", "优先级", "负责人", "状态", "完成度", "预期完成时间", "结束时间"}, records[0])

	// Check data row
	assert.Equal(t, "TEST-00001", records[1][0])
	assert.Equal(t, "Main Item", records[1][1])
	assert.Equal(t, "main", records[1][2])
	assert.Equal(t, "P1", records[1][3])
	assert.Equal(t, "Alice", records[1][4])
	assert.Equal(t, "completed", records[1][5])
	assert.Equal(t, "100", records[1][6])
	assert.Equal(t, "2026-05-01", records[1][7])
	assert.Equal(t, "2026-04-20", records[1][8])
}

func TestTableExportCSV_MultipleRows(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Main", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub", Priority: "P2", ItemStatus: "pending", Completion: 0, ExpectedEndDate: &endDate},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	data, err := svc.TableExportCSV(context.Background(), 1, dto.TableFilter{})
	require.NoError(t, err)

	csvData := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Header + 2 data rows
	require.Len(t, records, 3)
	assert.Equal(t, "main", records[1][2])
	assert.Equal(t, "sub", records[2][2])
}

func TestTableExportCSV_ExportWithFilter(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "P1", Priority: "P1", ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2, BizKey: 2}, TeamKey: 1, Code: "TEST-00002", Title: "P3", Priority: "P3", ItemStatus: "pending", Completion: 0, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	data, err := svc.TableExportCSV(context.Background(), 1, dto.TableFilter{Priority: []string{"P1"}})
	require.NoError(t, err)

	csvData := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Header + 1 data row (only P1)
	require.Len(t, records, 2)
	assert.Equal(t, "P1", records[1][3])
}

func TestTableExportCSV_NilDates_AsEmpty(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "No Dates", Priority: "P1", ItemStatus: "pending", Completion: 0},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	data, err := svc.TableExportCSV(context.Background(), 1, dto.TableFilter{})
	require.NoError(t, err)

	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	require.NoError(t, err)

	require.Len(t, records, 2)
	assert.Equal(t, "", records[1][7]) // expectedEndDate
	assert.Equal(t, "", records[1][8]) // actualEndDate
}

func TestTableExportCSV_UTF8BOM(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Item", Priority: "P1", ItemStatus: "progressing", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	data, err := svc.TableExportCSV(context.Background(), 1, dto.TableFilter{})
	require.NoError(t, err)

	// UTF-8 BOM prefix
	assert.True(t, bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}), "CSV should have UTF-8 BOM")
}

// ---------------------------------------------------------------------------
// Tests: Batch resolution (N+1 fix verification)
// ---------------------------------------------------------------------------

func TestResolveAssigneeNames_UsesBatchFindByIDs(t *testing.T) {
	assignee1 := uint(100)
	assignee2 := uint(200)

	rows := []dto.TableRow{
		{BizKey: "1", Title: "A", AssigneeID: func() *string { v := fmt.Sprintf("%d", assignee1); return &v }()},
		{BizKey: "2", Title: "B", AssigneeID: func() *string { v := fmt.Sprintf("%d", assignee2); return &v }()},
		{BizKey: "3", Title: "C", AssigneeID: func() *string { v := fmt.Sprintf("%d", assignee1); return &v }()}, // duplicate assignee
	}

	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			100: {BaseModel: model.BaseModel{ID: 100, BizKey: 100}, DisplayName: "Alice"},
			200: {BaseModel: model.BaseModel{ID: 200, BizKey: 200}, DisplayName: "Bob"},
		},
	}

	resolveAssigneeNames(context.Background(), rows, userRepo)

	// Should call FindByIDs exactly once
	assert.Equal(t, uint(1), userRepo.findByIDsCalls, "resolveAssigneeNames should use a single FindByIDs call")
	// Should NOT call FindByID at all
	assert.Equal(t, uint(0), userRepo.findByIDCalls, "resolveAssigneeNames should not call FindByID")
	// Names resolved correctly
	assert.Equal(t, "Alice", rows[0].AssigneeName)
	assert.Equal(t, "Bob", rows[1].AssigneeName)
	assert.Equal(t, "Alice", rows[2].AssigneeName)
}

func TestResolveAssigneeNames_DeduplicatesIDs(t *testing.T) {
	id1 := uint(10)
	id2 := uint(20)
	rows := []dto.TableRow{
		{BizKey: "1", AssigneeID: func() *string { v := fmt.Sprintf("%d", id1); return &v }()},
		{BizKey: "2", AssigneeID: func() *string { v := fmt.Sprintf("%d", id1); return &v }()},
		{BizKey: "3", AssigneeID: func() *string { v := fmt.Sprintf("%d", id2); return &v }()},
	}

	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			10: {DisplayName: "A"},
			20: {DisplayName: "B"},
		},
	}

	resolveAssigneeNames(context.Background(), rows, userRepo)

	// FindByIDs should be called with deduplicated IDs
	require.Len(t, userRepo.findByIDsArg, 2)
	assert.Contains(t, userRepo.findByIDsArg, uint(10))
	assert.Contains(t, userRepo.findByIDsArg, uint(20))
}

func TestResolveAssigneeNames_NilUserRepo_NoPanic(t *testing.T) {
	rows := []dto.TableRow{{BizKey: "1"}}
	assert.NotPanics(t, func() {
		resolveAssigneeNames(context.Background(), rows, nil)
	})
}

func TestWeeklyComparison_UsesBatchFindByIDs(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	assignee1 := uint(100)
	assignee2 := uint(200)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "Main 1", Priority: "P1", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub A", ItemStatus: "progressing", Completion: 60, AssigneeKey: func() *int64 { v := int64(assignee1); return &v }()},
			{BaseModel: model.BaseModel{ID: 11, BizKey: 11}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub B", ItemStatus: "pending", Completion: 0, AssigneeKey: func() *int64 { v := int64(assignee2); return &v }()},
			{BaseModel: model.BaseModel{ID: 12, BizKey: 12}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub C", ItemStatus: "pending", Completion: 0, AssigneeKey: func() *int64 { v := int64(assignee1); return &v }()}, // same assignee
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemKey: 10, TeamKey: 1, Completion: 60, CreateTime: weekStart.AddDate(0, 0, 1)},
		},
	}
	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			100: {BaseModel: model.BaseModel{ID: 100, BizKey: 100}, DisplayName: "Alice"},
			200: {BaseModel: model.BaseModel{ID: 200, BizKey: 200}, DisplayName: "Bob"},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo, userRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	// Should call FindByIDs exactly once, never FindByID
	assert.Equal(t, uint(1), userRepo.findByIDsCalls, "WeeklyComparison should use a single FindByIDs call")
	assert.Equal(t, uint(0), userRepo.findByIDCalls, "WeeklyComparison should not call FindByID")

	// Verify names resolved correctly in output
	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	// Find Sub A (ID=10) in ThisWeek and check name
	for _, snap := range group.ThisWeek {
		if snap.BizKey == "10" {
			assert.Equal(t, "Alice", snap.AssigneeName)
		}
		if snap.BizKey == "11" {
			assert.Equal(t, "Bob", snap.AssigneeName)
		}
	}
}

func TestTableView_UsesBatchFindByIDs(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	assignee1 := uint(100)
	assignee2 := uint(200)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Main", Priority: "P1", AssigneeKey: func() *int64 { v := int64(assignee1); return &v }(), ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub", Priority: "P2", AssigneeKey: func() *int64 { v := int64(assignee2); return &v }(), ItemStatus: "pending", Completion: 0, ExpectedEndDate: &endDate},
		},
	}
	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			100: {BaseModel: model.BaseModel{ID: 100, BizKey: 100}, DisplayName: "Alice"},
			200: {BaseModel: model.BaseModel{ID: 200, BizKey: 200}, DisplayName: "Bob"},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, userRepo)
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	// Should call FindByIDs exactly once
	assert.Equal(t, uint(1), userRepo.findByIDsCalls, "TableView should use a single FindByIDs call")
	assert.Equal(t, uint(0), userRepo.findByIDCalls, "TableView should not call FindByID")

	// Names resolved correctly
	require.Len(t, result.Items, 2)
	assert.Equal(t, "Alice", result.Items[0].AssigneeName)
	assert.Equal(t, "Bob", result.Items[1].AssigneeName)
}

func TestTableExportCSV_UsesBatchFindByIDs(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	assignee1 := uint(100)
	assignee2 := uint(200)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Code: "TEST-00001", Title: "Main", Priority: "P1", AssigneeKey: func() *int64 { v := int64(assignee1); return &v }(), ItemStatus: "progressing", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, TeamKey: 1, MainItemKey: int64(1), Title: "Sub", Priority: "P2", AssigneeKey: func() *int64 { v := int64(assignee2); return &v }(), ItemStatus: "pending", Completion: 0, ExpectedEndDate: &endDate},
		},
	}
	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			100: {BaseModel: model.BaseModel{ID: 100, BizKey: 100}, DisplayName: "Alice"},
			200: {BaseModel: model.BaseModel{ID: 200, BizKey: 200}, DisplayName: "Bob"},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, userRepo)
	_, err := svc.TableExportCSV(context.Background(), 1, dto.TableFilter{})
	require.NoError(t, err)

	// Should call FindByIDs exactly once
	assert.Equal(t, uint(1), userRepo.findByIDsCalls, "TableExportCSV should use a single FindByIDs call")
	assert.Equal(t, uint(0), userRepo.findByIDCalls, "TableExportCSV should not call FindByID")
}

// ---------------------------------------------------------------------------
// Tests: buildWeeklyGroups stats counting (unit, no service layer)
// ---------------------------------------------------------------------------

func TestBuildWeeklyGroups_Stats(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 6)

	overdueDate := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)       // before weekEnd
	completedInWeek := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)   // within week
	beforeWeekStart := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)    // before weekStart

	mainItem := model.MainItem{BaseModel: model.BaseModel{ID: 1, BizKey: 1}, TeamKey: 1, Title: "M", Priority: "P1"}

	activeSet := func(bizKeys ...int64) map[int64]struct{} {
		m := make(map[int64]struct{})
		for _, k := range bizKeys {
			m[k] = struct{}{}
		}
		return m
	}
	progressFor := func(subBizKey int64) map[int64][]model.ProgressRecord {
		return map[int64][]model.ProgressRecord{
			subBizKey: {{ID: 1, SubItemKey: subBizKey, Completion: 50, CreateTime: weekStart.AddDate(0, 0, 1)}},
		}
	}
	emptyProgress := map[int64][]model.ProgressRecord{}
	emptyActive := map[int64]struct{}{}

	tests := []struct {
		name      string
		subs      []model.SubItem
		active    map[int64]struct{}
		progress  map[int64][]model.ProgressRecord
		wantStats dto.WeeklyStats
	}{
		{
			name: "progressing active → inProgress=1 activeSubItems=1",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "progressing"},
			},
			active:    activeSet(10),
			progress:  progressFor(10),
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, InProgress: 1},
		},
		{
			name: "blocking active → blocked=1 activeSubItems=1",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "blocking"},
			},
			active:    activeSet(10),
			progress:  progressFor(10),
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, Blocked: 1},
		},
		{
			name: "pending active → pending=1 activeSubItems=1",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "pending"},
			},
			active:    activeSet(10),
			progress:  progressFor(10),
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, Pending: 1},
		},
		{
			name: "pausing active → pausing=1 activeSubItems=1",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "pausing"},
			},
			active:    activeSet(10),
			progress:  progressFor(10),
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, Pausing: 1},
		},
		{
			name: "justCompleted → newlyCompleted=1 activeSubItems=1",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "completed", ActualEndDate: &completedInWeek, Completion: 100},
			},
			active:    emptyActive,
			progress:  emptyProgress,
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, NewlyCompleted: 1},
		},
		{
			name: "overdue progressing → overdue=1 inProgress=1 activeSubItems=1",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "progressing", ExpectedEndDate: &overdueDate},
			},
			active:    activeSet(10),
			progress:  progressFor(10),
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, InProgress: 1, Overdue: 1},
		},
		{
			name: "pending not active → pending=0",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "pending"},
			},
			active:    emptyActive,
			progress:  emptyProgress,
			wantStats: dto.WeeklyStats{},
		},
		{
			name: "nil expectedEndDate → overdue=0",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "progressing", ExpectedEndDate: nil},
			},
			active:    activeSet(10),
			progress:  progressFor(10),
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, InProgress: 1, Overdue: 0},
		},
		{
			name: "completed with past deadline → overdue=0",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "completed", ExpectedEndDate: &overdueDate, ActualEndDate: &completedInWeek, Completion: 100},
			},
			active:    activeSet(10),
			progress:  progressFor(10),
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, NewlyCompleted: 1, Overdue: 0},
		},
		{
			name: "closed with past deadline → overdue=0",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "closed", ExpectedEndDate: &overdueDate},
			},
			active:    activeSet(10),
			progress:  progressFor(10),
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, Overdue: 0},
		},
		{
			name: "progress record in week but actualEndDate before weekStart → activeSubItems=1 (progress wins)",
			subs: []model.SubItem{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, MainItemKey: int64(1), ItemStatus: "progressing", ActualEndDate: &beforeWeekStart},
			},
			active:    emptyActive,
			progress:  progressFor(10), // has progress record in this week
			wantStats: dto.WeeklyStats{ActiveSubItems: 1, InProgress: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subsByMain := map[uint][]model.SubItem{1: tt.subs}
			_, stats := buildWeeklyGroups(
				[]model.MainItem{mainItem},
				subsByMain,
				emptyActive,        // lastWeekActive
				tt.active,          // thisWeekActive
				emptyProgress,      // lastWeekProgress
				tt.progress,        // thisWeekProgress
				map[int64]float64{}, // lastWeekCompletion
				map[int64]string{},  // latestProgressDesc
				map[uint]string{},   // assigneeNames
				weekStart,
				weekEnd,
			)
			assert.Equal(t, tt.wantStats, stats)
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

// seedBenchmarkData creates a dataset of n main items with sub-items and progress.
func seedBenchmarkData(n int) (*mockViewMainItemRepo, *mockViewSubItemRepo, *mockViewProgressRepo, *mockViewUserRepo) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	mainItems := make([]model.MainItem, n)
	for i := range mainItems {
		id := uint(i + 1)
		assigneeID := uint((i % 50) + 1) // 50 unique assignees
		mainItems[i] = model.MainItem{
			BaseModel:      model.BaseModel{ID: id, BizKey: int64(id)},
			TeamKey: 1,
			Code:           fmt.Sprintf("BENCH-%05d", id),
			Title:          fmt.Sprintf("Main Item %d", id),
			Priority:       []string{"P1", "P2", "P3"}[i%3],
			ItemStatus: "progressing",
			Completion:     float64(i % 100),
			PlanStartDate: &startDate,
			ExpectedEndDate: &endDate,
			AssigneeKey: func() *int64 { v := int64(assigneeID); return &v }(),
		}
	}

	// Each main item has 2 sub-items
	subItems := make([]model.SubItem, n*2)
	for i := range subItems {
		mainID := uint(i/2 + 1)
		subItems[i] = model.SubItem{
			BaseModel:       model.BaseModel{ID: uint(i + 1), BizKey: int64(i + 1)},
			TeamKey: 1,
			MainItemKey: int64(mainID),
			Title:           fmt.Sprintf("Sub Item %d", i+1),
			Priority:        []string{"P1", "P2", "P3"}[(i/2)%3],
			ItemStatus: "progressing",
			Completion:      float64(i % 100),
			PlanStartDate: &startDate,
			ExpectedEndDate: &endDate,
			AssigneeKey: mainItems[i/2].AssigneeKey,
		}
	}

	// Progress records for the current week
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	records := make([]model.ProgressRecord, n)
	for i := range records {
		records[i] = model.ProgressRecord{
			ID:          uint(i + 1),
			SubItemKey: int64(i + 1),
			TeamKey: 1,
			Completion:  float64(i % 100),
			Achievement: fmt.Sprintf("Achievement %d", i),
			CreateTime:   weekStart.AddDate(0, 0, i%7),
		}
	}

	// Build user map for assignee resolution (50 users)
	users := make(map[uint]*model.User, 50)
	for i := 1; i <= 50; i++ {
		users[uint(i)] = &model.User{
			BaseModel:   model.BaseModel{ID: uint(i), BizKey: int64(i)},
			DisplayName: fmt.Sprintf("User %d", i),
		}
	}

	return &mockViewMainItemRepo{items: mainItems},
		&mockViewSubItemRepo{items: subItems},
		&mockViewProgressRepo{records: records},
		&mockViewUserRepo{users: users}
}

func BenchmarkTableView(b *testing.B) {
	b.StopTimer()
	mainRepo, subRepo, _, userRepo := seedBenchmarkData(200)
	svc := newViewServiceWithUsers(mainRepo, subRepo, userRepo)
	ctx := context.Background()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := svc.TableView(ctx, 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 20})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTableView_LargePage(b *testing.B) {
	b.StopTimer()
	mainRepo, subRepo, _, userRepo := seedBenchmarkData(200)
	svc := newViewServiceWithUsers(mainRepo, subRepo, userRepo)
	ctx := context.Background()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := svc.TableView(ctx, 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 100})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGanttView(b *testing.B) {
	b.StopTimer()
	mainRepo, subRepo, _, _ := seedBenchmarkData(200)
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})
	ctx := context.Background()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := svc.GanttView(ctx, 1, dto.GanttFilter{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGanttView_WithStatusFilter(b *testing.B) {
	b.StopTimer()
	mainRepo, subRepo, _, _ := seedBenchmarkData(200)
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})
	ctx := context.Background()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := svc.GanttView(ctx, 1, dto.GanttFilter{Status: "progressing"})
		if err != nil {
			b.Fatal(err)
		}
	}
}
