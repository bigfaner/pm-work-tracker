package service

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/repository"
)

// MilestoneMapService defines business operations for MilestoneMap.
type MilestoneMapService interface {
	Create(ctx context.Context, teamBizKey int64, req dto.MilestoneMapCreateReq) (*model.MilestoneMap, error)
	Get(ctx context.Context, mapID uint) (*model.MilestoneMap, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error)
	List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error)
	Update(ctx context.Context, teamBizKey int64, mapID uint, req dto.MilestoneMapUpdateReq) error
}

type milestoneMapService struct {
	repo repository.MilestoneMapRepo
}

// NewMilestoneMapService creates a new MilestoneMapService.
func NewMilestoneMapService(repo repository.MilestoneMapRepo) MilestoneMapService {
	return &milestoneMapService{repo: repo}
}

func (s *milestoneMapService) Create(ctx context.Context, teamBizKey int64, req dto.MilestoneMapCreateReq) (*model.MilestoneMap, error) {
	m := &model.MilestoneMap{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:   teamBizKey,
		MapName:   req.MapName,
		MapDesc:   req.MapDesc,
		MapStatus: "planning",
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *milestoneMapService) Get(ctx context.Context, mapID uint) (*model.MilestoneMap, error) {
	m, err := s.repo.FindByID(ctx, mapID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrMilestoneMapNotFound)
	}
	return m, nil
}

func (s *milestoneMapService) GetByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error) {
	m, err := s.repo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrMilestoneMapNotFound)
	}
	return m, nil
}

func (s *milestoneMapService) List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error) {
	return s.repo.List(ctx, teamBizKey, filter, page)
}

func (s *milestoneMapService) Update(ctx context.Context, teamBizKey int64, mapID uint, req dto.MilestoneMapUpdateReq) error {
	m, err := s.repo.FindByID(ctx, mapID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrMilestoneMapNotFound)
	}
	if m.TeamKey != teamBizKey {
		return apperrors.ErrForbidden
	}

	fields := map[string]interface{}{}
	if req.MapName != nil {
		fields["map_name"] = *req.MapName
	}
	if req.MapDesc != nil {
		fields["map_desc"] = *req.MapDesc
	}

	if len(fields) == 0 {
		return nil
	}

	return s.repo.Update(ctx, m, fields)
}
