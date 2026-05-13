package service

import (
	"context"
	"time"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/dates"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/repository"
)

// MilestoneService defines business operations for Milestone.
type MilestoneService interface {
	Create(ctx context.Context, teamBizKey, milestoneMapBizKey int64, req dto.MilestoneCreateReq) (*model.Milestone, error)
	Get(ctx context.Context, id uint) (*model.Milestone, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error)
	ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error)
	ListByTeam(ctx context.Context, teamBizKey int64, excludeCancelled bool) ([]model.Milestone, error)
	Update(ctx context.Context, milestoneID uint, req dto.MilestoneUpdateReq) (*model.Milestone, error)
}

type milestoneService struct {
	milestoneRepo    repository.MilestoneRepo
	milestoneMapRepo repository.MilestoneMapRepo
}

// NewMilestoneService creates a new MilestoneService.
func NewMilestoneService(milestoneRepo repository.MilestoneRepo, milestoneMapRepo repository.MilestoneMapRepo) MilestoneService {
	return &milestoneService{milestoneRepo: milestoneRepo, milestoneMapRepo: milestoneMapRepo}
}

func (s *milestoneService) Create(ctx context.Context, teamBizKey, milestoneMapBizKey int64, req dto.MilestoneCreateReq) (*model.Milestone, error) {
	// Validate parent map exists and is not deleted
	_, err := s.milestoneMapRepo.FindByBizKey(ctx, milestoneMapBizKey)
	if err != nil {
		return nil, err
	}

	var expectedEndDate *time.Time
	if req.ExpectedEndDate != nil {
		t, err := dates.ParseDate(*req.ExpectedEndDate)
		if err != nil {
			return nil, apperrors.ErrValidation
		}
		expectedEndDate = &t
	}

	m := &model.Milestone{
		BaseModel:       model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:         teamBizKey,
		MilestoneMapKey: milestoneMapBizKey,
		MilestoneName:   req.MilestoneName,
		ExpectedEndDate: expectedEndDate,
		MilestoneStatus: "not_started",
	}

	if err := s.milestoneRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *milestoneService) Get(ctx context.Context, id uint) (*model.Milestone, error) {
	m, err := s.milestoneRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrMilestoneNotFound)
	}
	return m, nil
}

func (s *milestoneService) GetByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error) {
	m, err := s.milestoneRepo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrMilestoneNotFound)
	}
	return m, nil
}

func (s *milestoneService) ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error) {
	return s.milestoneRepo.ListByMap(ctx, milestoneMapBizKey)
}

func (s *milestoneService) ListByTeam(ctx context.Context, teamBizKey int64, excludeCancelled bool) ([]model.Milestone, error) {
	return s.milestoneRepo.ListByTeam(ctx, teamBizKey, excludeCancelled)
}

func (s *milestoneService) Update(ctx context.Context, milestoneID uint, req dto.MilestoneUpdateReq) (*model.Milestone, error) {
	m, err := s.milestoneRepo.FindByID(ctx, milestoneID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrMilestoneNotFound)
	}

	// Optimistic locking: compare request dbUpdateTime with current record
	reqTime, err := time.Parse(time.RFC3339, req.DbUpdateTime)
	if err != nil {
		return nil, apperrors.ErrValidation
	}
	if !m.DbUpdateTime.Equal(reqTime) {
		return nil, apperrors.ErrConcurrentEdit
	}

	fields := map[string]interface{}{}
	if req.MilestoneName != nil {
		fields["milestone_name"] = *req.MilestoneName
	}
	if req.ExpectedEndDate != nil {
		fields["expected_end_date"] = *req.ExpectedEndDate
	}

	if len(fields) == 0 {
		return m, nil
	}

	if err := s.milestoneRepo.Update(ctx, m, fields); err != nil {
		return nil, err
	}

	// Fetch updated record
	updated, findErr := s.milestoneRepo.FindByID(ctx, milestoneID)
	if findErr != nil {
		return nil, findErr
	}
	return updated, nil
}
