//nolint:misspell // "cancelled" is a domain status value per PRD/API contract, not a misspelling.
package service

import (
	"context"
	"slices"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/dates"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/pkg/status"
	"pm-work-tracker/backend/internal/repository"
)

// MilestoneService defines business operations for Milestone entities.
type MilestoneService interface {
	Create(ctx context.Context, teamBizKey int64, milestoneMapBizKey int64, req dto.MilestoneCreateReq) (*model.Milestone, error)
	Update(ctx context.Context, teamBizKey int64, milestoneID uint, req dto.MilestoneUpdateReq) error
	Get(ctx context.Context, milestoneID uint) (*model.Milestone, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error)
	ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error)
	ListByTeam(ctx context.Context, teamBizKey int64, filter dto.MilestoneTeamFilter) ([]model.Milestone, error)
	Delete(ctx context.Context, teamBizKey int64, milestoneID uint) error
	ChangeStatus(ctx context.Context, teamBizKey int64, milestoneID uint, newStatus string) (*model.Milestone, error)
	AvailableTransitions(ctx context.Context, teamBizKey int64, milestoneID uint) ([]string, error)
}

type milestoneService struct {
	milestoneRepo repository.MilestoneRepo
	mapRepo       repository.MilestoneMapRepo
	mainItemRepo  repository.MainItemRepo
	db            repo.DBTransactor
}

// NewMilestoneService creates a new MilestoneService.
func NewMilestoneService(
	milestoneRepo repository.MilestoneRepo,
	mapRepo repository.MilestoneMapRepo,
	mainItemRepo repository.MainItemRepo,
	db repo.DBTransactor,
) MilestoneService {
	return &milestoneService{
		milestoneRepo: milestoneRepo,
		mapRepo:       mapRepo,
		mainItemRepo:  mainItemRepo,
		db:            db,
	}
}

func (s *milestoneService) Create(ctx context.Context, teamBizKey, milestoneMapBizKey int64, req dto.MilestoneCreateReq) (*model.Milestone, error) {
	// BR-5: parent map must not be terminal
	m, err := s.mapRepo.FindByBizKey(ctx, milestoneMapBizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if status.MilestoneMapStatuses[m.MapStatus].Terminal {
		return nil, apperrors.ErrMapIsTerminal
	}

	// Duplicate name check
	exists, err := s.milestoneRepo.ExistsByNameAndMap(ctx, milestoneMapBizKey, req.MilestoneName, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrDuplicateMilestoneName
	}

	ms := &model.Milestone{
		BaseModel:       model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:         teamBizKey,
		MilestoneMapKey: milestoneMapBizKey,
		MilestoneName:   req.MilestoneName,
		MilestoneDesc:   req.MilestoneDesc,
		MilestoneStatus: "not_started",
	}

	if req.ExpectedEndDate != "" {
		t, e := dates.ParseDate(req.ExpectedEndDate)
		if e != nil {
			return nil, apperrors.ErrValidation
		}
		ms.ExpectedEndDate = &t
	}

	if err := s.milestoneRepo.Create(ctx, ms); err != nil {
		if apperrors.IsMySQLDuplicateError(err) {
			return nil, apperrors.ErrDuplicateBizKey
		}
		return nil, err
	}
	return ms, nil
}

func (s *milestoneService) Update(ctx context.Context, teamBizKey int64, milestoneID uint, req dto.MilestoneUpdateReq) error {
	ms, err := s.milestoneRepo.FindByID(ctx, milestoneID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if ms.TeamKey != teamBizKey {
		return apperrors.ErrForbidden
	}

	// BR-5: if any status-relevant fields are being updated, parent map must not be terminal
	if req.MilestoneName != nil || req.ExpectedEndDate != nil {
		m, err := s.mapRepo.FindByBizKey(ctx, ms.MilestoneMapKey)
		if err != nil {
			return apperrors.MapNotFound(err, apperrors.ErrNotFound)
		}
		if status.MilestoneMapStatuses[m.MapStatus].Terminal {
			return apperrors.ErrMapIsTerminal
		}
	}

	fields := map[string]interface{}{}
	if req.MilestoneName != nil {
		// Check duplicate name if changing name
		exists, e := s.milestoneRepo.ExistsByNameAndMap(ctx, ms.MilestoneMapKey, *req.MilestoneName, &milestoneID)
		if e != nil {
			return e
		}
		if exists {
			return apperrors.ErrDuplicateMilestoneName
		}
		fields["milestone_name"] = *req.MilestoneName
	}
	if req.MilestoneDesc != nil {
		fields["milestone_desc"] = *req.MilestoneDesc
	}
	if req.ExpectedEndDate != nil {
		t, e := dates.ParseDate(*req.ExpectedEndDate)
		if e != nil {
			return apperrors.ErrValidation
		}
		fields["expected_end_date"] = &t
	}

	if len(fields) == 0 {
		return nil
	}

	return s.milestoneRepo.Update(ctx, ms, fields)
}

func (s *milestoneService) Get(ctx context.Context, milestoneID uint) (*model.Milestone, error) {
	ms, err := s.milestoneRepo.FindByID(ctx, milestoneID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	return ms, nil
}

func (s *milestoneService) GetByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error) {
	ms, err := s.milestoneRepo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	return ms, nil
}

func (s *milestoneService) ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error) {
	return s.milestoneRepo.ListByMap(ctx, milestoneMapBizKey)
}

func (s *milestoneService) ListByTeam(ctx context.Context, teamBizKey int64, filter dto.MilestoneTeamFilter) ([]model.Milestone, error) {
	return s.milestoneRepo.ListByTeam(ctx, teamBizKey, filter)
}

func (s *milestoneService) Delete(ctx context.Context, teamBizKey int64, milestoneID uint) error {
	ms, err := s.milestoneRepo.FindByID(ctx, milestoneID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if ms.TeamKey != teamBizKey {
		return apperrors.ErrForbidden
	}

	// BR-4: only not_started and cancelled can be deleted
	if ms.MilestoneStatus != "not_started" && ms.MilestoneStatus != "cancelled" {
		return apperrors.ErrMilestoneCannotDelete
	}

	// Unbind all associated MIs in transaction
	return s.db.Transaction(func(_ *gorm.DB) error {
		if err := s.mainItemRepo.ClearMilestoneKeyByMilestone(ctx, ms.BizKey); err != nil {
			return err
		}
		return s.milestoneRepo.SoftDelete(ctx, milestoneID)
	})
}

func (s *milestoneService) ChangeStatus(ctx context.Context, teamBizKey int64, milestoneID uint, newStatus string) (*model.Milestone, error) {
	ms, err := s.milestoneRepo.FindByID(ctx, milestoneID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if ms.TeamKey != teamBizKey {
		return nil, apperrors.ErrForbidden
	}

	// BR-5: parent map must not be terminal
	m, err := s.mapRepo.FindByBizKey(ctx, ms.MilestoneMapKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if status.MilestoneMapStatuses[m.MapStatus].Terminal {
		return nil, apperrors.ErrMapIsTerminal
	}

	// Validate transition
	if !status.IsValidTransition(status.MilestoneTransitions, ms.MilestoneStatus, newStatus) {
		return nil, apperrors.ErrInvalidStatus
	}

	// BR-1: completed requires all MIs terminal
	if newStatus == "completed" {
		items, err := s.mainItemRepo.FindByMilestoneKey(ctx, ms.BizKey)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if !status.IsMainTerminal(item.ItemStatus) {
				return nil, apperrors.ErrMilestoneHasNonTerminalItems
			}
		}
	}

	// Cancelled: auto-unbind all MIs
	if newStatus == "cancelled" {
		err := s.db.Transaction(func(_ *gorm.DB) error {
			if err := s.mainItemRepo.ClearMilestoneKeyByMilestone(ctx, ms.BizKey); err != nil {
				return err
			}
			return s.milestoneRepo.Update(ctx, ms, map[string]interface{}{
				"milestone_status": newStatus,
			})
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.milestoneRepo.Update(ctx, ms, map[string]interface{}{
			"milestone_status": newStatus,
		}); err != nil {
			return nil, err
		}
	}

	return s.milestoneRepo.FindByID(ctx, milestoneID)
}

func (s *milestoneService) AvailableTransitions(ctx context.Context, teamBizKey int64, milestoneID uint) ([]string, error) {
	ms, err := s.milestoneRepo.FindByID(ctx, milestoneID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if ms.TeamKey != teamBizKey {
		return nil, apperrors.ErrForbidden
	}

	// BR-5: parent map terminal → no transitions
	m, err := s.mapRepo.FindByBizKey(ctx, ms.MilestoneMapKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}
	if status.MilestoneMapStatuses[m.MapStatus].Terminal {
		return []string{}, nil
	}

	transitions := status.GetAvailableTransitions(status.MilestoneTransitions, ms.MilestoneStatus)

	// BR-1: filter completed if MIs not all terminal
	if slices.Contains(transitions, "completed") {
		items, err := s.mainItemRepo.FindByMilestoneKey(ctx, ms.BizKey)
		if err != nil {
			return nil, err
		}
		allTerminal := true
		for _, item := range items {
			if !status.IsMainTerminal(item.ItemStatus) {
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

// calcCompletion returns average MI completion (0 for empty milestone).
func (s *milestoneService) calcCompletion(ctx context.Context, milestoneBizKey int64) float64 {
	items, err := s.mainItemRepo.FindByMilestoneKey(ctx, milestoneBizKey)
	if err != nil || len(items) == 0 {
		return 0
	}

	var total float64
	for _, item := range items {
		total += item.Completion
	}
	return total / float64(len(items))
}

// countRelatedMIs returns the count of MIs associated with the milestone.
func (s *milestoneService) countRelatedMIs(ctx context.Context, milestoneBizKey int64) int64 {
	count, err := s.mainItemRepo.CountByMilestoneKey(ctx, milestoneBizKey)
	if err != nil {
		return 0
	}
	return count
}

// ensure milestoneService exposes calcCompletion and countRelatedMIs for test access.
var _ = (*milestoneService)(nil)
