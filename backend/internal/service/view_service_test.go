package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
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
func (m *mockViewMainItemRepo) List(_ context.Context, _ uint, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockViewMainItemRepo) NextCode(_ context.Context, _ uint) (string, error) {
	return "", nil
}
func (m *mockViewMainItemRepo) CountByTeam(_ context.Context, _ uint) (int64, error) {
	return 0, nil
}
func (m *mockViewMainItemRepo) ListNonArchivedByTeam(_ context.Context, _ uint) ([]model.MainItem, error) {
	if m.listErr != nil {
		return nil, m.listErr
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
func (m *mockViewSubItemRepo) List(_ context.Context, _ uint, _ uint, _ dto.SubItemFilter, _ dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}
func (m *mockViewSubItemRepo) ListByMainItem(_ context.Context, _ uint) ([]*model.SubItem, error) {
	return nil, nil
}
func (m *mockViewSubItemRepo) ListByTeam(_ context.Context, _ uint) ([]model.SubItem, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
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
func (m *mockViewProgressRepo) ListBySubItem(_ context.Context, _ uint, _ uint) ([]model.ProgressRecord, error) {
	return nil, nil
}
func (m *mockViewProgressRepo) LatestBySubItem(_ context.Context, _ uint) (*model.ProgressRecord, error) {
	return nil, nil
}
func (m *mockViewProgressRepo) UpdateCompletion(_ context.Context, _ uint, _ float64) error {
	return nil
}
func (m *mockViewProgressRepo) ListByTeamInRange(_ context.Context, _ uint, _, _ time.Time) ([]model.ProgressRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.records, nil
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView validation
// ---------------------------------------------------------------------------

func TestWeeklyView_RejectsNonMonday(t *testing.T) {
	svc := NewViewService(&mockViewMainItemRepo{}, &mockViewSubItemRepo{}, &mockViewProgressRepo{})

	// Tuesday
	tuesday := time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC)
	_, err := svc.WeeklyView(context.Background(), 1, tuesday)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestWeeklyView_RejectsSunday(t *testing.T) {
	svc := NewViewService(&mockViewMainItemRepo{}, &mockViewSubItemRepo{}, &mockViewProgressRepo{})

	sunday := time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)
	_, err := svc.WeeklyView(context.Background(), 1, sunday)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestWeeklyView_RejectsWednesday(t *testing.T) {
	svc := NewViewService(&mockViewMainItemRepo{}, &mockViewSubItemRepo{}, &mockViewProgressRepo{})

	wed := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	_, err := svc.WeeklyView(context.Background(), 1, wed)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestWeeklyView_AcceptsMonday(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{}, &mockViewProgressRepo{})

	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	result, err := svc.WeeklyView(context.Background(), 1, monday)
	require.NoError(t, err)
	assert.Equal(t, "2026-04-13", result.WeekStart)
	assert.Equal(t, "2026-04-19", result.WeekEnd)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView empty team
// ---------------------------------------------------------------------------

func TestWeeklyView_EmptyTeam_NoGroups(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	result, err := svc.WeeklyView(context.Background(), 1, monday)
	require.NoError(t, err)
	assert.Empty(t, result.Groups)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView with data — newly completed
// ---------------------------------------------------------------------------

func TestWeeklyView_NewlyCompleted(t *testing.T) {
	// Monday of the week
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	completedDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:        model.BaseModel{ID: 10},
				TeamID:       1,
				MainItemID:   1,
				Title:        "Sub A",
				Status:       "已完成",
				Completion:   100,
				ActualEndDate: &completedDate,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{
				ID:         100,
				SubItemID:  10,
				TeamID:     1,
				Completion: 100,
				CreatedAt:  time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyView(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	assert.Equal(t, uint(1), group.MainItem.ID)
	assert.Equal(t, "Main 1", group.MainItem.Title)
	require.Len(t, group.NewlyCompleted, 1)
	assert.Equal(t, uint(10), group.NewlyCompleted[0].ID)
	assert.Empty(t, group.HasProgress)
	assert.Empty(t, group.NoChangeFromLastWeek)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView with data — has progress (not completed)
// ---------------------------------------------------------------------------

func TestWeeklyView_HasProgress(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:      model.BaseModel{ID: 10},
				TeamID:     1,
				MainItemID: 1,
				Title:      "Sub A",
				Status:     "进行中",
				Completion: 60,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{
				ID:         100,
				SubItemID:  10,
				TeamID:     1,
				Completion: 60,
				CreatedAt:  time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyView(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	assert.Empty(t, group.NewlyCompleted)
	require.Len(t, group.HasProgress, 1)
	assert.Equal(t, uint(10), group.HasProgress[0].ID)
	require.Len(t, group.HasProgress[0].ProgressThisWeek, 1)
	assert.Equal(t, uint(100), group.HasProgress[0].ProgressThisWeek[0].ID)
	assert.Empty(t, group.NoChangeFromLastWeek)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView with data — no change from last week
// ---------------------------------------------------------------------------

func TestWeeklyView_NoChangeFromLastWeek(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Completion: 30},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:      model.BaseModel{ID: 10},
				TeamID:     1,
				MainItemID: 1,
				Title:      "Sub A",
				Status:     "待开始",
				Completion: 0,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyView(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	assert.Empty(t, group.NewlyCompleted)
	assert.Empty(t, group.HasProgress)
	require.Len(t, group.NoChangeFromLastWeek, 1)
	assert.Equal(t, uint(10), group.NoChangeFromLastWeek[0].ID)
	assert.Equal(t, "Sub A", group.NoChangeFromLastWeek[0].Title)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView — main item with no sub-items is omitted
// ---------------------------------------------------------------------------

func TestWeeklyView_MainItemWithNoSubItems_Omitted(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Empty Main", Completion: 0},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyView(context.Background(), 1, weekStart)
	require.NoError(t, err)
	assert.Empty(t, result.Groups)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView — mixed scenario across multiple main items
// ---------------------------------------------------------------------------

func TestWeeklyView_MixedScenario(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	completedDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Completion: 80},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Title: "Main 2", Completion: 20},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			// Main 1: sub-item completed this week
			{
				BaseModel:         model.BaseModel{ID: 10},
				TeamID:        1,
				MainItemID:    1,
				Title:         "Sub Completed",
				Status:        "已完成",
				Completion:    100,
				ActualEndDate: &completedDate,
			},
			// Main 1: sub-item with progress
			{
				BaseModel:      model.BaseModel{ID: 11},
				TeamID:     1,
				MainItemID: 1,
				Title:      "Sub In Progress",
				Status:     "进行中",
				Completion: 60,
			},
			// Main 2: sub-item with no change
			{
				BaseModel:      model.BaseModel{ID: 20},
				TeamID:     1,
				MainItemID: 2,
				Title:      "Sub Dormant",
				Status:     "待开始",
				Completion: 0,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			// Progress for sub 10 (completed this week)
			{
				ID:         100,
				SubItemID:  10,
				TeamID:     1,
				Completion: 100,
				CreatedAt:  time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
			},
			// Progress for sub 11 (in progress)
			{
				ID:         101,
				SubItemID:  11,
				TeamID:     1,
				Completion: 60,
				CreatedAt:  time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyView(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 2)

	// Main 1 group
	g1 := result.Groups[0]
	assert.Equal(t, uint(1), g1.MainItem.ID)
	require.Len(t, g1.NewlyCompleted, 1)
	assert.Equal(t, uint(10), g1.NewlyCompleted[0].ID)
	require.Len(t, g1.HasProgress, 1)
	assert.Equal(t, uint(11), g1.HasProgress[0].ID)
	assert.Empty(t, g1.NoChangeFromLastWeek)

	// Main 2 group
	g2 := result.Groups[1]
	assert.Equal(t, uint(2), g2.MainItem.ID)
	assert.Empty(t, g2.NewlyCompleted)
	assert.Empty(t, g2.HasProgress)
	require.Len(t, g2.NoChangeFromLastWeek, 1)
	assert.Equal(t, uint(20), g2.NoChangeFromLastWeek[0].ID)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView — completed before this week → no change
// ---------------------------------------------------------------------------

func TestWeeklyView_CompletedBeforeThisWeek_NoChange(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	// Completed last week
	lastWeekEnd := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:         model.BaseModel{ID: 10},
				TeamID:        1,
				MainItemID:    1,
				Title:         "Sub Old Completed",
				Status:        "已完成",
				Completion:    100,
				ActualEndDate: &lastWeekEnd,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyView(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	assert.Empty(t, group.NewlyCompleted)
	assert.Empty(t, group.HasProgress)
	require.Len(t, group.NoChangeFromLastWeek, 1)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView — repo errors
// ---------------------------------------------------------------------------

func TestWeeklyView_MainItemRepoError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{listErr: errors.New("db error")}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{}, &mockViewProgressRepo{})

	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	_, err := svc.WeeklyView(context.Background(), 1, monday)
	assert.Error(t, err)
}

func TestWeeklyView_SubItemRepoError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1"},
		},
	}
	subRepo := &mockViewSubItemRepo{listErr: errors.New("db error")}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	_, err := svc.WeeklyView(context.Background(), 1, monday)
	assert.Error(t, err)
}

func TestWeeklyView_ProgressRepoError(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub A"},
		},
	}
	progressRepo := &mockViewProgressRepo{listErr: errors.New("db error")}
	svc := NewViewService(mainRepo, subRepo, progressRepo)

	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	_, err := svc.WeeklyView(context.Background(), 1, monday)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView — weekEnd computation
// ---------------------------------------------------------------------------

func TestWeeklyView_WeekEndIsSunday(t *testing.T) {
	mainRepo := &mockViewMainItemRepo{items: []model.MainItem{}}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	result, err := svc.WeeklyView(context.Background(), 1, monday)
	require.NoError(t, err)
	assert.Equal(t, "2026-04-19", result.WeekEnd)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView — newly completed requires ActualEndDate within week
// ---------------------------------------------------------------------------

func TestWeeklyView_NewlyCompleted_ActualEndDateOutsideWeek(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	// Completed before this week
	oldDate := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:         model.BaseModel{ID: 10},
				TeamID:        1,
				MainItemID:    1,
				Title:         "Sub A",
				Status:        "已完成",
				Completion:    100,
				ActualEndDate: &oldDate,
			},
		},
	}
	// Has progress this week — should make it "hasProgress" not "newlyCompleted"
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{
				ID:         100,
				SubItemID:  10,
				TeamID:     1,
				Completion: 100,
				CreatedAt:  time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyView(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	// Not newly completed because ActualEndDate is outside the week range
	assert.Empty(t, group.NewlyCompleted)
	// Has progress because there's a progress record this week
	require.Len(t, group.HasProgress, 1)
	assert.Equal(t, uint(10), group.HasProgress[0].ID)
}

// ---------------------------------------------------------------------------
// Tests: WeeklyView — sub-item with no progress records ever → noChange
// ---------------------------------------------------------------------------

func TestWeeklyView_SubItemNoProgressEver_NoChange(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Completion: 0},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:      model.BaseModel{ID: 10},
				TeamID:     1,
				MainItemID: 1,
				Title:      "Sub A",
				Status:     "待开始",
				Completion: 0,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyView(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	group := result.Groups[0]
	assert.Empty(t, group.NewlyCompleted)
	assert.Empty(t, group.HasProgress)
	require.Len(t, group.NoChangeFromLastWeek, 1)
	assert.Equal(t, uint(10), group.NoChangeFromLastWeek[0].ID)
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "In Progress", Status: "进行中", Completion: 60},
			{BaseModel: model.BaseModel{ID: 11}, TeamID: 1, MainItemID: 1, Title: "Blocked", Status: "阻塞中", Completion: 30},
			{BaseModel: model.BaseModel{ID: 12}, TeamID: 1, MainItemID: 1, Title: "Completed", Status: "已完成", Completion: 100, ActualEndDate: &completedDate},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemID: 10, TeamID: 1, Completion: 60, CreatedAt: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
			{ID: 101, SubItemID: 12, TeamID: 1, Completion: 100, CreatedAt: time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub A", Status: "进行中", Completion: 70},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			// Last week: completion was 40
			{ID: 90, SubItemID: 10, TeamID: 1, Completion: 40, CreatedAt: lastWeekStart.AddDate(0, 0, 2)},
			// This week: completion is 70
			{ID: 100, SubItemID: 10, TeamID: 1, Completion: 70, CreatedAt: weekStart.AddDate(0, 0, 1)},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "New Sub", Status: "待开始", Completion: 0},
		},
	}
	// Only this week progress, no last week progress
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemID: 10, TeamID: 1, Completion: 0, CreatedAt: weekStart.AddDate(0, 0, 1)},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Priority: "P1", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:         model.BaseModel{ID: 10},
				TeamID:        1,
				MainItemID:    1,
				Title:         "Just Done",
				Status:        "已完成",
				Completion:    100,
				ActualEndDate: &completedDate,
			},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{ID: 100, SubItemID: 10, TeamID: 1, Completion: 100, CreatedAt: completedDate},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Priority: "P1", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:         model.BaseModel{ID: 10},
				TeamID:        1,
				MainItemID:    1,
				Title:         "Old Completed",
				Status:        "已完成",
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
	assert.Equal(t, uint(10), group.CompletedNoChange[0].ID)
}

func TestWeeklyComparison_GroupsSortedByPriority(t *testing.T) {
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "P3 Item", Priority: "P3", Completion: 10},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Title: "P1 Item", Priority: "P1", Completion: 50},
			{BaseModel: model.BaseModel{ID: 3}, TeamID: 1, Title: "P2 Item", Priority: "P2", Completion: 30},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub 1", Status: "进行中", Completion: 10},
			{BaseModel: model.BaseModel{ID: 20}, TeamID: 1, MainItemID: 2, Title: "Sub 2", Status: "进行中", Completion: 50},
			{BaseModel: model.BaseModel{ID: 30}, TeamID: 1, MainItemID: 3, Title: "Sub 3", Status: "进行中", Completion: 30},
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
				BaseModel:           model.BaseModel{ID: 1},
				TeamID:          1,
				Title:           "Main 1",
				Priority:        "P1",
				StartDate:       &startDate,
				ExpectedEndDate: &endDate,
				Completion:      58,
			},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub A", Status: "进行中", Completion: 60},
			{BaseModel: model.BaseModel{ID: 11}, TeamID: 1, MainItemID: 1, Title: "Sub B", Status: "待开始", Completion: 0},
		},
	}
	progressRepo := &mockViewProgressRepo{records: []model.ProgressRecord{}}

	svc := NewViewService(mainRepo, subRepo, progressRepo)
	result, err := svc.WeeklyComparison(context.Background(), 1, weekStart)
	require.NoError(t, err)

	require.Len(t, result.Groups, 1)
	mi := result.Groups[0].MainItem
	assert.Equal(t, uint(1), mi.ID)
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Empty Main", Completion: 0},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub A", Status: "进行中", Completion: 60},
		},
	}
	progressRepo := &mockViewProgressRepo{
		records: []model.ProgressRecord{
			{
				ID:         100,
				SubItemID:  10,
				TeamID:     1,
				Completion: 60,
				Achievement: "Token sign done",
				Blocker:     "Blacklist WIP",
				CreatedAt:   weekStart.AddDate(0, 0, 1),
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

func TestWeeklyComparison_RepoErrors(t *testing.T) {
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	// MainItemRepo error
	svc := NewViewService(&mockViewMainItemRepo{listErr: errors.New("db error")}, &mockViewSubItemRepo{}, &mockViewProgressRepo{})
	_, err := svc.WeeklyComparison(context.Background(), 1, monday)
	assert.Error(t, err)

	// SubItemRepo error
	svc = NewViewService(
		&mockViewMainItemRepo{items: []model.MainItem{{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1"}}},
		&mockViewSubItemRepo{listErr: errors.New("db error")},
		&mockViewProgressRepo{},
	)
	_, err = svc.WeeklyComparison(context.Background(), 1, monday)
	assert.Error(t, err)

	// ProgressRepo error
	svc = NewViewService(
		&mockViewMainItemRepo{items: []model.MainItem{{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1"}}},
		&mockViewSubItemRepo{items: []model.SubItem{{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub A"}}},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Priority: "P1", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:  model.BaseModel{ID: 10, CreatedAt: time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)},
				TeamID:     1,
				MainItemID: 1,
				Title:      "Future Sub",
				Status:     "进行中",
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", Priority: "P1", Completion: 100},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:      model.BaseModel{ID: 10, CreatedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
				TeamID:         1,
				MainItemID:     1,
				Title:          "Old Completed Sub",
				Status:         "已完成",
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
	assert.Equal(t, uint(10), group.CompletedNoChange[0].ID)
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
				BaseModel:           model.BaseModel{ID: 1},
				TeamID:          1,
				Title:           "Main 1",
				Priority:        "P1",
				StartDate:       &startDate,
				ExpectedEndDate: &endDate,
				Completion:      45.5,
				Status:          "进行中",
			},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	item := result.Items[0]
	assert.Equal(t, uint(1), item.ID)
	assert.Equal(t, "Main 1", item.Title)
	assert.Equal(t, "P1", item.Priority)
	assert.Equal(t, "2026-04-01", item.StartDate)
	assert.Equal(t, "2026-04-30", item.ExpectedEndDate)
	assert.Equal(t, 45.5, item.Completion)
	assert.Equal(t, "进行中", item.Status)
	assert.False(t, item.IsOverdue)
}

func TestGanttView_OverdueItem(t *testing.T) {
	// expectedEndDate is in the past, status is not 已完成 or 已关闭
	pastDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1},
				TeamID:          1,
				Title:           "Overdue Main",
				Priority:        "P1",
				StartDate:       &startDate,
				ExpectedEndDate: &pastDate,
				Status:          "进行中",
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

	for _, status := range []string{"已完成", "已关闭"} {
		t.Run(status, func(t *testing.T) {
			mainRepo := &mockViewMainItemRepo{
				items: []model.MainItem{
					{
						BaseModel:           model.BaseModel{ID: 1},
						TeamID:          1,
						Title:           "Completed",
						Priority:        "P1",
						StartDate:       &startDate,
						ExpectedEndDate: &pastDate,
						Status:          status,
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "In Progress", Status: "进行中", StartDate: &startDate, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Title: "Completed", Status: "已完成", StartDate: &startDate, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 3}, TeamID: 1, Title: "Pending", Status: "待开始", StartDate: &startDate, ExpectedEndDate: &endDate},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{Status: "进行中"})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.Equal(t, uint(1), result.Items[0].ID)
	assert.Equal(t, "In Progress", result.Items[0].Title)
}

func TestGanttView_StatusFilterEmpty_ReturnsAll(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "A", Status: "进行中", StartDate: &startDate, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Title: "B", Status: "已完成", StartDate: &startDate, ExpectedEndDate: &endDate},
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
				BaseModel:           model.BaseModel{ID: 1},
				TeamID:          1,
				Title:           "Main 1",
				Priority:        "P1",
				StartDate:       &startDate,
				ExpectedEndDate: &endDate,
				Status:          "进行中",
			},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:           model.BaseModel{ID: 10},
				TeamID:          1,
				MainItemID:      1,
				Title:           "Sub A",
				StartDate:       &startDate,
				ExpectedEndDate: &endDate,
				Completion:      80,
				Status:          "待验收",
			},
		},
	}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	require.Len(t, result.Items[0].SubItems, 1)
	sub := result.Items[0].SubItems[0]
	assert.Equal(t, uint(10), sub.ID)
	assert.Equal(t, "Sub A", sub.Title)
	assert.Equal(t, "2026-04-01", sub.StartDate)
	assert.Equal(t, "2026-04-30", sub.ExpectedEndDate)
	assert.Equal(t, 80.0, sub.Completion)
	assert.Equal(t, "待验收", sub.Status)
}

func TestGanttView_SubItemsFromOtherMainItemNotIncluded(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", StartDate: &startDate, ExpectedEndDate: &endDate, Status: "进行中"},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Title: "Main 2", StartDate: &startDate, ExpectedEndDate: &endDate, Status: "进行中"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub for Main 1", StartDate: &startDate, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 20}, TeamID: 1, MainItemID: 2, Title: "Sub for Main 2", StartDate: &startDate, ExpectedEndDate: &endDate},
		},
	}
	svc := NewViewService(mainRepo, subRepo, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 2)
	require.Len(t, result.Items[0].SubItems, 1)
	assert.Equal(t, uint(10), result.Items[0].SubItems[0].ID)
	require.Len(t, result.Items[1].SubItems, 1)
	assert.Equal(t, uint(20), result.Items[1].SubItems[0].ID)
}

func TestGanttView_ArchivedItemsExcluded(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	// The repo's ListNonArchivedByTeam already filters out archived items,
	// so the mock only returns non-archived items.
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1},
				TeamID:          1,
				Title:           "Active",
				StartDate:       &startDate,
				ExpectedEndDate: &endDate,
				Status:          "进行中",
			},
		},
	}
	svc := NewViewService(mainRepo, &mockViewSubItemRepo{items: []model.SubItem{}}, &mockViewProgressRepo{})

	result, err := svc.GanttView(context.Background(), 1, dto.GanttFilter{})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	assert.Equal(t, uint(1), result.Items[0].ID)
	assert.Equal(t, "Active", result.Items[0].Title)
}

func TestGanttView_DatesFormattedAsISO8601(t *testing.T) {
	startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:           model.BaseModel{ID: 1},
				TeamID:          1,
				Title:           "Main 1",
				StartDate:       &startDate,
				ExpectedEndDate: &endDate,
				Status:          "进行中",
			},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:           model.BaseModel{ID: 10},
				TeamID:          1,
				MainItemID:      1,
				Title:           "Sub 1",
				StartDate:       &startDate,
				ExpectedEndDate: &endDate,
				Status:          "进行中",
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
				BaseModel:  model.BaseModel{ID: 1},
				TeamID: 1,
				Title:  "No dates",
				Status: "待开始",
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1"},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Title: "Main 1", StartDate: &startDate, ExpectedEndDate: &endDate, Status: "进行中"},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:           model.BaseModel{ID: 10},
				TeamID:          1,
				MainItemID:      1,
				Title:           "Sub A",
				Description:     "Should not appear in gantt sub-item DTO",
				Priority:        "P2",
				AssigneeID:      nil,
				StartDate:       &startDate,
				ExpectedEndDate: &endDate,
				ActualEndDate:   nil,
				Status:          "进行中",
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
	assert.Equal(t, uint(10), sub.ID)
	assert.Equal(t, "Sub A", sub.Title)
	assert.Equal(t, "2026-04-01", sub.StartDate)
	assert.Equal(t, "2026-04-30", sub.ExpectedEndDate)
	assert.Equal(t, 60.0, sub.Completion)
	assert.Equal(t, "进行中", sub.Status)
	// GanttSubItemDTO should NOT have isOverdue (v1)
	// This is enforced by struct definition
}

func TestGanttView_Overdue_NilExpectedEndDate(t *testing.T) {
	// No expected end date set — should not be overdue
	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{
				BaseModel:  model.BaseModel{ID: 1},
				TeamID: 1,
				Title:  "No end date",
				Status: "进行中",
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
	users map[uint]*model.User
}

func (m *mockViewUserRepo) FindByID(_ context.Context, id uint) (*model.User, error) {
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
				BaseModel:           model.BaseModel{ID: 1},
				TeamID:          1,
				Code:            "MI-0001",
				Title:           "Main Item",
				Priority:        "P1",
				AssigneeID:      &assigneeID,
				Status:          "进行中",
				Completion:      50,
				ExpectedEndDate: &endDate,
			},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{
				BaseModel:           model.BaseModel{ID: 10},
				TeamID:          1,
				MainItemID:      1,
				Title:           "Sub Item",
				Priority:        "P2",
				AssigneeID:      &assigneeID,
				Status:          "待开始",
				Completion:      0,
				ExpectedEndDate: &endDate,
			},
		},
	}
	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			100: {BaseModel: model.BaseModel{ID: 100}, DisplayName: "Alice"},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, userRepo)
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(2), result.Total)
	require.Len(t, result.Items, 2)

	// Items should be sorted by default: priority DESC (P1 before P2)
	assert.Equal(t, "main", result.Items[0].Type)
	assert.Equal(t, "MI-0001", result.Items[0].Code)
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "Main", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub", Priority: "P2", Status: "待开始", Completion: 0, ExpectedEndDate: &endDate},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "Main", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub", Priority: "P2", Status: "待开始", Completion: 0, ExpectedEndDate: &endDate},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "P1 Item", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Code: "MI-0002", Title: "P3 Item", Priority: "P3", Status: "待开始", Completion: 0, ExpectedEndDate: &endDate},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "Active", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Code: "MI-0002", Title: "Done", Priority: "P1", Status: "已完成", Completion: 100, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{Status: []string{"已完成"}}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "已完成", result.Items[0].Status)
}

func TestTableView_FilterByAssigneeID(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	assignee1 := uint(100)
	assignee2 := uint(200)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "Alice's", Priority: "P1", AssigneeID: &assignee1, Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Code: "MI-0002", Title: "Bob's", Priority: "P1", AssigneeID: &assignee2, Status: "进行中", Completion: 30, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{AssigneeID: &assignee1}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, uint(100), *result.Items[0].AssigneeID)
}

func TestTableView_Pagination(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "A", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Code: "MI-0002", Title: "B", Priority: "P2", Status: "进行中", Completion: 30, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 3}, TeamID: 1, Code: "MI-0003", Title: "C", Priority: "P3", Status: "待开始", Completion: 0, ExpectedEndDate: &endDate},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "P2 Late", Priority: "P2", Status: "进行中", Completion: 30, ExpectedEndDate: &endDate3},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Code: "MI-0002", Title: "P2 Early", Priority: "P2", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate2},
			{BaseModel: model.BaseModel{ID: 3}, TeamID: 1, Code: "MI-0003", Title: "P1", Priority: "P1", Status: "进行中", Completion: 80, ExpectedEndDate: &endDate1},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "High", Priority: "P1", Status: "进行中", Completion: 80, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Code: "MI-0002", Title: "Low", Priority: "P1", Status: "待开始", Completion: 10, ExpectedEndDate: &endDate},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "Low", Priority: "P1", Status: "进行中", Completion: 10, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Code: "MI-0002", Title: "High", Priority: "P1", Status: "待开始", Completion: 80, ExpectedEndDate: &endDate},
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
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
		},
	}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	result, err := svc.TableView(context.Background(), 1, dto.TableFilter{Type: "sub"}, dto.Pagination{Page: 1, PageSize: 10})
	require.NoError(t, err)

	require.Len(t, result.Items, 1)
	// SubItem code should be SI-XXXX format based on its ID
	assert.Equal(t, "SI-0010", result.Items[0].Code)
}

func TestTableView_AssigneeNameResolved(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	assigneeID := uint(50)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "Main", Priority: "P1", AssigneeID: &assigneeID, Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}
	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			50: {BaseModel: model.BaseModel{ID: 50}, DisplayName: "Bob"},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "No Assignee", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "Main", Priority: "P1", Status: "已完成", Completion: 100, ExpectedEndDate: &expectedEnd, ActualEndDate: &actualEnd},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "No Dates", Priority: "P1", Status: "待开始", Completion: 0},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "A", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Code: "MI-0002", Title: "B", Priority: "P2", Status: "进行中", Completion: 30, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 3}, TeamID: 1, Code: "MI-0003", Title: "C", Priority: "P3", Status: "待开始", Completion: 0, ExpectedEndDate: &endDate},
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
				BaseModel:           model.BaseModel{ID: 1},
				TeamID:          1,
				Code:            "MI-0001",
				Title:           "Main Item",
				Priority:        "P1",
				AssigneeID:      &assigneeID,
				Status:          "已完成",
				Completion:      100,
				ExpectedEndDate: &endDate,
				ActualEndDate:   &actualEnd,
			},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}
	userRepo := &mockViewUserRepo{
		users: map[uint]*model.User{
			100: {BaseModel: model.BaseModel{ID: 100}, DisplayName: "Alice"},
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
	assert.Equal(t, []string{"编号", "标题", "类型", "优先级", "负责人", "状态", "完成度", "预期完成时间", "实际完成时间"}, records[0])

	// Check data row
	assert.Equal(t, "MI-0001", records[1][0])
	assert.Equal(t, "Main Item", records[1][1])
	assert.Equal(t, "main", records[1][2])
	assert.Equal(t, "P1", records[1][3])
	assert.Equal(t, "Alice", records[1][4])
	assert.Equal(t, "已完成", records[1][5])
	assert.Equal(t, "100", records[1][6])
	assert.Equal(t, "2026-05-01", records[1][7])
	assert.Equal(t, "2026-04-20", records[1][8])
}

func TestTableExportCSV_MultipleRows(t *testing.T) {
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mainRepo := &mockViewMainItemRepo{
		items: []model.MainItem{
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "Main", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
		},
	}
	subRepo := &mockViewSubItemRepo{
		items: []model.SubItem{
			{BaseModel: model.BaseModel{ID: 10}, TeamID: 1, MainItemID: 1, Title: "Sub", Priority: "P2", Status: "待开始", Completion: 0, ExpectedEndDate: &endDate},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "P1", Priority: "P1", Status: "进行中", Completion: 50, ExpectedEndDate: &endDate},
			{BaseModel: model.BaseModel{ID: 2}, TeamID: 1, Code: "MI-0002", Title: "P3", Priority: "P3", Status: "待开始", Completion: 0, ExpectedEndDate: &endDate},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "No Dates", Priority: "P1", Status: "待开始", Completion: 0},
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
			{BaseModel: model.BaseModel{ID: 1}, TeamID: 1, Code: "MI-0001", Title: "Item", Priority: "P1", Status: "进行中", Completion: 50},
		},
	}
	subRepo := &mockViewSubItemRepo{items: []model.SubItem{}}

	svc := newViewServiceWithUsers(mainRepo, subRepo, &mockViewUserRepo{})
	data, err := svc.TableExportCSV(context.Background(), 1, dto.TableFilter{})
	require.NoError(t, err)

	// UTF-8 BOM prefix
	assert.True(t, bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}), "CSV should have UTF-8 BOM")
}
