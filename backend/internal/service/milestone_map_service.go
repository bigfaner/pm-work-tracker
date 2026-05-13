package service

import (
	"context"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/pkg/status"
	"pm-work-tracker/backend/internal/repository"
)

// MilestoneMapService defines business operations for MilestoneMap.
type MilestoneMapService interface {
	Create(ctx context.Context, teamBizKey int64, req dto.MilestoneMapCreateReq) (*model.MilestoneMap, error)
	Get(ctx context.Context, mapID uint) (*model.MilestoneMap, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error)
	List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error)
	Update(ctx context.Context, teamBizKey int64, mapID uint, req dto.MilestoneMapUpdateReq) error
	Delete(ctx context.Context, mapID uint) error
	ChangeStatus(ctx context.Context, mapID uint, newStatus string) (*model.MilestoneMap, error)
	AvailableTransitions(ctx context.Context, mapID uint) ([]string, error)
	CalcOverallProgress(ctx context.Context, milestoneMapBizKey int64) float64
	CalcMilestoneCount(ctx context.Context, milestoneMapBizKey int64) int64
	CalcItemCount(ctx context.Context, milestoneMapBizKey int64) int64
}

type milestoneMapService struct {
	repo          repository.MilestoneMapRepo
	milestoneRepo repository.MilestoneRepo
	mainItemRepo  repository.MainItemRepo
	db            repo.DBTransactor
}

// NewMilestoneMapService creates a new MilestoneMapService.
func NewMilestoneMapService(mapRepo repository.MilestoneMapRepo, milestoneRepo repository.MilestoneRepo, mainItemRepo repository.MainItemRepo, db repo.DBTransactor) MilestoneMapService {
	return &milestoneMapService{repo: mapRepo, milestoneRepo: milestoneRepo, mainItemRepo: mainItemRepo, db: db}
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

func (s *milestoneMapService) Delete(ctx context.Context, mapID uint) error {
	m, err := s.repo.FindByID(ctx, mapID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrMilestoneMapNotFound)
	}

	return s.db.Transaction(func(_ *gorm.DB) error {
		// Unbind all MIs pointing to milestones of this map
		if err := s.mainItemRepo.UnbindByMap(ctx, m.BizKey); err != nil {
			return err
		}
		// Soft-delete all milestones of this map
		if err := s.milestoneRepo.DeleteByMap(ctx, m.BizKey); err != nil {
			return err
		}
		// Soft-delete the map itself
		return s.repo.SoftDelete(ctx, mapID)
	})
}

func (s *milestoneMapService) ChangeStatus(ctx context.Context, mapID uint, newStatus string) (*model.MilestoneMap, error) {
	m, err := s.repo.FindByID(ctx, mapID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrMilestoneMapNotFound)
	}

	// Validate the new status is a known milestone map status
	if _, ok := status.GetMilestoneMapStatus(newStatus); !ok {
		return nil, apperrors.ErrInvalidStatus
	}

	// Validate the transition is legal
	if !status.IsValidTransition(status.MilestoneMapTransitions, m.MapStatus, newStatus) {
		return nil, apperrors.ErrInvalidStatus
	}

	fields := map[string]interface{}{
		"map_status": newStatus,
	}

	if err := s.repo.Update(ctx, m, fields); err != nil {
		return nil, err
	}

	// Fetch updated record
	updated, findErr := s.repo.FindByID(ctx, mapID)
	if findErr != nil {
		return nil, findErr
	}
	return updated, nil
}

func (s *milestoneMapService) AvailableTransitions(ctx context.Context, mapID uint) ([]string, error) {
	m, err := s.repo.FindByID(ctx, mapID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrMilestoneMapNotFound)
	}
	return status.GetAvailableTransitions(status.MilestoneMapTransitions, m.MapStatus), nil
}

func (s *milestoneMapService) CalcOverallProgress(ctx context.Context, milestoneMapBizKey int64) float64 {
	avg, err := s.mainItemRepo.CalcCompletionByMap(ctx, milestoneMapBizKey)
	if err != nil {
		return 0
	}
	return avg
}

func (s *milestoneMapService) CalcMilestoneCount(ctx context.Context, milestoneMapBizKey int64) int64 {
	milestones, err := s.milestoneRepo.ListByMap(ctx, milestoneMapBizKey)
	if err != nil {
		return 0
	}
	return int64(len(milestones))
}

func (s *milestoneMapService) CalcItemCount(ctx context.Context, milestoneMapBizKey int64) int64 {
	count, err := s.mainItemRepo.CountByMap(ctx, milestoneMapBizKey)
	if err != nil {
		return 0
	}
	return count
}
