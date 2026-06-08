// Package service provides business logic for all domain entities.
//
//nolint:misspell // "cancelled" is a domain status value per PRD/API contract, not a misspelling.
package service

import (
	"context"
	"slices"
	"time"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/dates"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/pkg/status"
	"pm-work-tracker/backend/internal/repository"
)

// MilestoneMapService defines business operations for MilestoneMap entities.
type MilestoneMapService interface {
	Create(ctx context.Context, teamBizKey int64, creatorBizKey int64, req dto.MilestoneMapCreateReq) (*model.MilestoneMap, error)
	Update(ctx context.Context, teamBizKey int64, mapID uint, req dto.MilestoneMapUpdateReq) error
	Get(ctx context.Context, mapID uint) (*model.MilestoneMap, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error)
	List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error)
	Delete(ctx context.Context, teamBizKey int64, mapID uint) error
	ChangeStatus(ctx context.Context, teamBizKey int64, mapID uint, newStatus string) (*model.MilestoneMap, error)
	AvailableTransitions(ctx context.Context, teamBizKey int64, mapID uint) ([]string, error)
}

type milestoneMapService struct {
	mapRepo       repository.MilestoneMapRepo
	milestoneRepo repository.MilestoneRepo
	mainItemRepo  repository.MainItemRepo
	db            repo.DBTransactor
}

// NewMilestoneMapService creates a new MilestoneMapService.
func NewMilestoneMapService(
	mapRepo repository.MilestoneMapRepo,
	milestoneRepo repository.MilestoneRepo,
	mainItemRepo repository.MainItemRepo,
	db repo.DBTransactor,
) MilestoneMapService {
	return &milestoneMapService{
		mapRepo:       mapRepo,
		milestoneRepo: milestoneRepo,
		mainItemRepo:  mainItemRepo,
		db:            db,
	}
}

func (s *milestoneMapService) Create(ctx context.Context, teamBizKey, creatorBizKey int64, req dto.MilestoneMapCreateReq) (*model.MilestoneMap, error) {
	assigneeKey, err := pkg.ParseID(req.AssigneeBizKey)
	if err != nil {
		return nil, apperrors.ErrValidation
	}

	m := &model.MilestoneMap{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:     teamBizKey,
		CreatorKey:  creatorBizKey,
		AssigneeKey: assigneeKey,
		MapName:     req.MapName,
		MapDesc:     req.MapDesc,
		MapStatus:   "planning",
	}

	if req.PlanStartDate != nil {
		if t, e := dates.ParseDate(*req.PlanStartDate); e == nil {
			m.PlanStartDate = &t
		}
	}
	if req.ExpectedEndDate != nil {
		if t, e := dates.ParseDate(*req.ExpectedEndDate); e == nil {
			m.ExpectedEndDate = &t
		}
	}

	if err := s.mapRepo.Create(ctx, m); err != nil {
		if apperrors.IsMySQLDuplicateError(err) {
			return nil, apperrors.ErrDuplicateBizKey
		}
		return nil, err
	}
	return m, nil
}

func (s *milestoneMapService) Update(ctx context.Context, teamBizKey int64, mapID uint, req dto.MilestoneMapUpdateReq) error {
	m, err := s.mapRepo.FindByID(ctx, mapID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotFound)
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
	if req.AssigneeBizKey != nil {
		assigneeKey, e := pkg.ParseID(*req.AssigneeBizKey)
		if e != nil {
			return apperrors.ErrValidation
		}
		fields["assignee_key"] = assigneeKey
	}

	var planStart *time.Time
	var expectedEnd *time.Time

	if req.PlanStartDate != nil {
		t, e := dates.ParseDate(*req.PlanStartDate)
		if e != nil {
			return apperrors.ErrValidation
		}
		planStart = &t
		fields["plan_start_date"] = planStart
	}
	if req.ExpectedEndDate != nil {
		t, e := dates.ParseDate(*req.ExpectedEndDate)
		if e != nil {
			return apperrors.ErrValidation
		}
		expectedEnd = &t
		fields["expected_end_date"] = expectedEnd
	}

	// Validate date range if both dates are being set (or one is being set and the other already exists)
	effectivePlanStart := m.PlanStartDate
	effectiveExpectedEnd := m.ExpectedEndDate
	if planStart != nil {
		effectivePlanStart = planStart
	}
	if expectedEnd != nil {
		effectiveExpectedEnd = expectedEnd
	}
	if effectivePlanStart != nil && effectiveExpectedEnd != nil && effectivePlanStart.After(*effectiveExpectedEnd) {
		return apperrors.ErrValidation
	}

	if len(fields) == 0 {
		return nil
	}

	return s.mapRepo.Update(ctx, m, fields)
}

func (s *milestoneMapService) Get(ctx context.Context, mapID uint) (*model.MilestoneMap, error) {
	m, err := s.mapRepo.FindByID(ctx, mapID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	return m, nil
}

func (s *milestoneMapService) GetByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error) {
	m, err := s.mapRepo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	return m, nil
}

func (s *milestoneMapService) List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error) {
	return s.mapRepo.List(ctx, teamBizKey, filter, page)
}

func (s *milestoneMapService) Delete(ctx context.Context, teamBizKey int64, mapID uint) error {
	m, err := s.mapRepo.FindByID(ctx, mapID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if m.TeamKey != teamBizKey {
		return apperrors.ErrForbidden
	}

	// BR-4: only planning, reviewed, ready can be deleted
	if m.MapStatus != "planning" && m.MapStatus != "reviewed" && m.MapStatus != "ready" {
		return apperrors.ErrMapCannotDelete
	}

	// Cascade: soft-delete all milestones + clear all MI milestone_keys in one transaction (SD-007)
	return s.db.Transaction(func(_ *gorm.DB) error {
		milestones, err := s.milestoneRepo.ListByMap(ctx, m.BizKey)
		if err != nil {
			return err
		}

		// Collect all milestone bizKeys for batch MI clearing
		if len(milestones) > 0 {
			bizKeys := make([]int64, 0, len(milestones))
			for _, ms := range milestones {
				bizKeys = append(bizKeys, ms.BizKey)
			}

			// Clear milestone_key on all related MIs
			if err := s.mainItemRepo.ClearMilestoneKeyByMap(ctx, bizKeys); err != nil {
				return err
			}
		}

		// Soft-delete all milestones under this map
		if err := s.milestoneRepo.SoftDeleteByMap(ctx, m.BizKey); err != nil {
			return err
		}

		// Soft-delete the map itself
		return s.mapRepo.SoftDelete(ctx, mapID)
	})
}

func (s *milestoneMapService) ChangeStatus(ctx context.Context, teamBizKey int64, mapID uint, newStatus string) (*model.MilestoneMap, error) {
	m, err := s.mapRepo.FindByID(ctx, mapID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if m.TeamKey != teamBizKey {
		return nil, apperrors.ErrForbidden
	}

	// Validate transition via IsValidTransition
	if !status.IsValidTransition(status.MilestoneMapTransitions, m.MapStatus, newStatus) {
		return nil, apperrors.ErrInvalidStatus
	}

	// BR-2: completed requires all milestones terminal
	if newStatus == "completed" {
		milestones, err := s.milestoneRepo.ListByMap(ctx, m.BizKey)
		if err != nil {
			return nil, err
		}
		for _, ms := range milestones {
			if !status.MilestoneStatuses[ms.MilestoneStatus].Terminal {
				return nil, apperrors.ErrMapHasNonTerminalMilestones
			}
		}
	}

	// BR-6: cancelled cascades — cancel non-terminal milestones + unbind all MIs
	if newStatus == "cancelled" {
		err := s.db.Transaction(func(_ *gorm.DB) error {
			milestones, err := s.milestoneRepo.ListByMap(ctx, m.BizKey)
			if err != nil {
				return err
			}

			for _, ms := range milestones {
				if !status.MilestoneStatuses[ms.MilestoneStatus].Terminal {
					if err := s.milestoneRepo.Update(ctx, &ms, map[string]interface{}{
						"milestone_status": "cancelled",
					}); err != nil {
						return err
					}
				}
			}

			// Collect all milestone bizKeys and unbind all MIs
			if len(milestones) > 0 {
				bizKeys := make([]int64, 0, len(milestones))
				for _, ms := range milestones {
					bizKeys = append(bizKeys, ms.BizKey)
				}
				if err := s.mainItemRepo.ClearMilestoneKeyByMap(ctx, bizKeys); err != nil {
					return err
				}
			}

			return s.mapRepo.Update(ctx, m, map[string]interface{}{
				"map_status": newStatus,
			})
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.mapRepo.Update(ctx, m, map[string]interface{}{
			"map_status": newStatus,
		}); err != nil {
			return nil, err
		}
	}

	return s.mapRepo.FindByID(ctx, mapID)
}

func (s *milestoneMapService) AvailableTransitions(ctx context.Context, teamBizKey int64, mapID uint) ([]string, error) {
	m, err := s.mapRepo.FindByID(ctx, mapID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if m.TeamKey != teamBizKey {
		return nil, apperrors.ErrForbidden
	}

	transitions := status.GetAvailableTransitions(status.MilestoneMapTransitions, m.MapStatus)

	// BR-2 filter: remove completed if milestones not all terminal
	if slices.Contains(transitions, "completed") {
		milestones, err := s.milestoneRepo.ListByMap(ctx, m.BizKey)
		if err != nil {
			return nil, err
		}
		allTerminal := true
		for _, ms := range milestones {
			if !status.MilestoneStatuses[ms.MilestoneStatus].Terminal {
				allTerminal = false
				break
			}
		}
		if !allTerminal {
			transitions = slices.DeleteFunc(transitions, func(t string) bool { return t == "completed" })
		}
	}

	return transitions, nil
}
