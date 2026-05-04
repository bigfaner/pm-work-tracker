package handler

import (
	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	pkgHandler "pm-work-tracker/backend/internal/pkg/handler"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// DecisionLogHandler handles decision log endpoints nested under main-items.
type DecisionLogHandler struct {
	svc          service.DecisionLogService
	userRepo     repository.UserRepo
	mainItemRepo repository.MainItemRepo
}

// NewDecisionLogHandler creates a new DecisionLogHandler with service and repo dependencies.
func NewDecisionLogHandler(svc service.DecisionLogService, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo) *DecisionLogHandler {
	if svc == nil {
		panic("decision_log_handler: decisionLogService must not be nil")
	}
	if userRepo == nil {
		panic("decision_log_handler: userRepo must not be nil")
	}
	if mainItemRepo == nil {
		panic("decision_log_handler: mainItemRepo must not be nil")
	}
	return &DecisionLogHandler{svc: svc, userRepo: userRepo, mainItemRepo: mainItemRepo}
}

// Create handles POST /api/v1/teams/:teamId/main-items/:itemId/decision-logs
func (h *DecisionLogHandler) Create(c *gin.Context) {
	mainItemKey, ok := pkgHandler.ParseBizKeyParam(c, "itemId")
	if !ok {
		return
	}

	userBizKey := middleware.GetUserBizKey(c)

	var req dto.DecisionLogCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	log, err := h.svc.Create(c.Request.Context(), mainItemKey, userBizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondCreated(c, buildDecisionLogVO(log, h.userRepo, c))
}

// Update handles PUT /api/v1/teams/:teamId/main-items/:itemId/decision-logs/:logId
func (h *DecisionLogHandler) Update(c *gin.Context) {
	logBizKey, ok := pkgHandler.ParseBizKeyParam(c, "logId")
	if !ok {
		return
	}

	userBizKey := middleware.GetUserBizKey(c)

	var req dto.DecisionLogUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	log, err := h.svc.Update(c.Request.Context(), logBizKey, userBizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildDecisionLogVO(log, h.userRepo, c))
}

// Publish handles PATCH /api/v1/teams/:teamId/main-items/:itemId/decision-logs/:logId/publish
func (h *DecisionLogHandler) Publish(c *gin.Context) {
	logBizKey, ok := pkgHandler.ParseBizKeyParam(c, "logId")
	if !ok {
		return
	}

	userBizKey := middleware.GetUserBizKey(c)

	log, err := h.svc.Publish(c.Request.Context(), logBizKey, userBizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildDecisionLogVO(log, h.userRepo, c))
}

// List handles GET /api/v1/teams/:teamId/main-items/:itemId/decision-logs
func (h *DecisionLogHandler) List(c *gin.Context) {
	mainItemKey, ok := pkgHandler.ParseBizKeyParam(c, "itemId")
	if !ok {
		return
	}

	userBizKey := middleware.GetUserBizKey(c)

	var page dto.Pagination
	if err := c.ShouldBindQuery(&page); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	result, err := h.svc.List(c.Request.Context(), mainItemKey, userBizKey, page)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	voItems := buildDecisionLogVOs(result.Items, h.userRepo, c)
	apperrors.RespondOK(c, gin.H{
		"items": voItems,
		"total": result.Total,
		"page":  result.Page,
		"size":  result.Size,
	})
}

// buildDecisionLogVO converts a single DecisionLog model to VO with creator name resolution.
func buildDecisionLogVO(log *model.DecisionLog, userRepo repository.UserRepo, c *gin.Context) vo.DecisionLogVO {
	keys := map[int64]struct{}{log.CreatedBy: {}}
	nameMap := batchLookupCreatorNames(keys, userRepo, c)
	return vo.NewDecisionLogVO(log, nameMap[log.CreatedBy])
}

// buildDecisionLogVOs converts a slice of DecisionLog models to VOs using batch user lookup.
func buildDecisionLogVOs(logs []model.DecisionLog, userRepo repository.UserRepo, c *gin.Context) []vo.DecisionLogVO {
	if len(logs) == 0 {
		return []vo.DecisionLogVO{}
	}

	keys := make(map[int64]struct{}, len(logs))
	for i := range logs {
		keys[logs[i].CreatedBy] = struct{}{}
	}
	nameMap := batchLookupCreatorNames(keys, userRepo, c)

	result := make([]vo.DecisionLogVO, 0, len(logs))
	for i := range logs {
		result = append(result, vo.NewDecisionLogVO(&logs[i], nameMap[logs[i].CreatedBy]))
	}
	return result
}

// batchLookupCreatorNames performs a batch user lookup by BizKeys and returns a map of BizKey -> DisplayName.
func batchLookupCreatorNames(creatorBizKeys map[int64]struct{}, userRepo repository.UserRepo, c *gin.Context) map[int64]string {
	result := make(map[int64]string, len(creatorBizKeys))
	if len(creatorBizKeys) == 0 {
		return result
	}

	bizKeys := make([]int64, 0, len(creatorBizKeys))
	for k := range creatorBizKeys {
		bizKeys = append(bizKeys, k)
	}

	userMap, err := userRepo.FindByBizKeys(c.Request.Context(), bizKeys)
	if err != nil {
		return result
	}

	for k, u := range userMap {
		result[k] = u.DisplayName
	}
	return result
}
