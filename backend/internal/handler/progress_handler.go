package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	pkgHandler "pm-work-tracker/backend/internal/pkg/handler"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// ProgressHandler handles progress record endpoints.
type ProgressHandler struct {
	svc        service.ProgressService
	userRepo   repository.UserRepo
	subItemSvc service.SubItemService
}

// NewProgressHandler creates a new ProgressHandler with service and repo dependencies.
func NewProgressHandler(svc service.ProgressService, userRepo repository.UserRepo, subItemSvc service.SubItemService) *ProgressHandler {
	if svc == nil {
		panic("progress_handler: progressService must not be nil")
	}
	if userRepo == nil {
		panic("progress_handler: userRepo must not be nil")
	}
	if subItemSvc == nil {
		panic("progress_handler: subItemService must not be nil")
	}
	return &ProgressHandler{svc: svc, userRepo: userRepo, subItemSvc: subItemSvc}
}

// appendProgressReq is the request DTO for appending progress.
// Completion uses a pointer so JSON "0" is treated as present (not the zero value).
type appendProgressReq struct {
	Completion  *float64 `json:"completion" binding:"required"`
	Achievement string   `json:"achievement"`
	Blocker     string   `json:"blocker"`
	Lesson      string   `json:"lesson"`
}

// correctCompletionReq is the request DTO for PM completion correction.
type correctCompletionReq struct {
	Completion *float64 `json:"completion" binding:"required"`
}

// validateCompletion checks that a completion value is within [0, 100].
func validateCompletion(val float64) bool {
	return val >= 0 && val <= 100
}

// Append handles POST /api/v1/teams/:teamId/sub-items/:subId/progress
func (h *ProgressHandler) Append(c *gin.Context) {
	subID, ok := pkgHandler.ResolveBizKey(c, "subId", func(ctx context.Context, bizKey int64) (uint, error) {
		item, err := h.subItemSvc.GetByBizKey(ctx, bizKey)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	})
	if !ok {
		return
	}

	var req appendProgressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	completion := *req.Completion
	if !validateCompletion(completion) {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)
	callerBizKey := middleware.GetUserBizKey(c)
	pmFlag := isPMOrSuperAdmin(c)

	record, err := h.svc.Append(c.Request.Context(), teamBizKey, callerBizKey, subID, completion, req.Achievement, req.Blocker, req.Lesson, pmFlag)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": buildProgressRecordVOs([]model.ProgressRecord{*record}, h.userRepo, c)[0]})
}

// List handles GET /api/v1/teams/:teamId/sub-items/:subId/progress
func (h *ProgressHandler) List(c *gin.Context) {
	subID, ok := pkgHandler.ResolveBizKey(c, "subId", func(ctx context.Context, bizKey int64) (uint, error) {
		item, err := h.subItemSvc.GetByBizKey(ctx, bizKey)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	})
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	records, err := h.svc.List(c.Request.Context(), teamBizKey, subID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildProgressRecordVOs(records, h.userRepo, c))
}

// CorrectCompletion handles PATCH /api/v1/teams/:teamId/progress/:recordId/completion
func (h *ProgressHandler) CorrectCompletion(c *gin.Context) {
	recordBizKeyStr := c.Param("recordId")
	recordBizKey, err := strconv.ParseInt(recordBizKeyStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	record, err := h.svc.GetByBizKey(c.Request.Context(), recordBizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	var req correctCompletionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	completion := *req.Completion
	if !validateCompletion(completion) {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	err = h.svc.CorrectCompletion(c.Request.Context(), teamBizKey, record.ID, completion)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// buildProgressRecordVOs converts a slice of ProgressRecord to ProgressRecordVO using batch lookups (fixes N+1).
// Single-item callers pass a 1-element slice; the batch path has no N+1 overhead.
func buildProgressRecordVOs(records []model.ProgressRecord, userRepo repository.UserRepo, c *gin.Context) []vo.ProgressRecordVO {
	if len(records) == 0 {
		return []vo.ProgressRecordVO{}
	}

	ctx := c.Request.Context()

	// Collect unique author BizKeys
	authorBizKeys := make(map[int64]struct{})
	for i := range records {
		authorBizKeys[records[i].AuthorKey] = struct{}{}
	}

	// Batch lookup
	userMap := make(map[int64]*model.User)
	if len(authorBizKeys) > 0 {
		bizKeys := make([]int64, 0, len(authorBizKeys))
		for k := range authorBizKeys {
			bizKeys = append(bizKeys, k)
		}
		if m, err := userRepo.FindByBizKeys(ctx, bizKeys); err == nil {
			userMap = m
		}
	}

	// Build VOs from map
	result := make([]vo.ProgressRecordVO, 0, len(records))
	for i := range records {
		authorName := ""
		if u, ok := userMap[records[i].AuthorKey]; ok {
			authorName = u.DisplayName
		}
		result = append(result, vo.NewProgressRecordVO(&records[i], authorName))
	}
	return result
}
