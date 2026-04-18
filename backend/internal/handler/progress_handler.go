package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
)

// ProgressHandler handles progress record endpoints.
type ProgressHandler struct {
	svc      service.ProgressService
	userRepo repository.UserRepo
}

// NewProgressHandler creates a new ProgressHandler (stub, for router setup before service is ready).
func NewProgressHandler() *ProgressHandler {
	return &ProgressHandler{}
}

// NewProgressHandlerWithDeps creates a new ProgressHandler with service and repo dependencies.
func NewProgressHandlerWithDeps(svc service.ProgressService, userRepo repository.UserRepo) *ProgressHandler {
	return &ProgressHandler{svc: svc, userRepo: userRepo}
}

// isPMOrSuperAdmin returns true if the caller has PM role in the team or is a superadmin.
func (h *ProgressHandler) isPMOrSuperAdmin(c *gin.Context) bool {
	if middleware.GetUserRole(c) == "superadmin" {
		return true
	}
	return middleware.GetCallerTeamRole(c) == "pm"
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
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	subID, ok := parseSubID(c)
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

	teamID := middleware.GetTeamID(c)
	callerID := middleware.GetUserID(c)
	isPM := h.isPMOrSuperAdmin(c)

	record, err := h.svc.Append(c.Request.Context(), teamID, callerID, subID, completion, req.Achievement, req.Blocker, req.Lesson, isPM)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": progressRecordToDTO(record, h.userRepo, c)})
}

// List handles GET /api/v1/teams/:teamId/sub-items/:subId/progress
func (h *ProgressHandler) List(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	subID, ok := parseSubID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)

	records, err := h.svc.List(c.Request.Context(), teamID, subID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, progressRecordsToDTOs(records, h.userRepo, c))
}

// CorrectCompletion handles PATCH /api/v1/teams/:teamId/progress/:recordId/completion
func (h *ProgressHandler) CorrectCompletion(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	// PM only
	if !h.isPMOrSuperAdmin(c) {
		apperrors.RespondError(c, apperrors.ErrForbidden)
		return
	}

	recordIDStr := c.Param("recordId")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
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

	teamID := middleware.GetTeamID(c)

	err = h.svc.CorrectCompletion(c.Request.Context(), teamID, uint(recordID), completion)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// progressRecordToDTO converts a model.ProgressRecord to a response map matching the Data Contract.
// It resolves authorName by looking up the user.
func progressRecordToDTO(record *model.ProgressRecord, userRepo repository.UserRepo, c *gin.Context) gin.H {
	authorName := ""
	if userRepo != nil {
		user, err := userRepo.FindByID(c.Request.Context(), record.AuthorID)
		if err == nil && user != nil {
			authorName = user.DisplayName
		}
	}

	return gin.H{
		"id":          record.ID,
		"subItemId":   record.SubItemID,
		"authorId":    record.AuthorID,
		"authorName":  authorName,
		"completion":  record.Completion,
		"achievement": record.Achievement,
		"blocker":     record.Blocker,
		"lesson":      record.Lesson,
		"isPMCorrect": record.IsPMCorrect,
		"createdAt":   record.CreatedAt,
	}
}

// progressRecordsToDTOs converts a slice of ProgressRecord to DTO maps.
func progressRecordsToDTOs(records []model.ProgressRecord, userRepo repository.UserRepo, c *gin.Context) []gin.H {
	result := make([]gin.H, 0, len(records))
	for i := range records {
		result = append(result, progressRecordToDTO(&records[i], userRepo, c))
	}
	return result
}
