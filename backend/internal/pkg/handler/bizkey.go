package handler

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ParseBizKeyParam extracts and validates a path parameter as int64 bizKey.
// Responds with ErrValidation on failure.
// Returns (bizKey, true) on success, (0, false) on failure.
func ParseBizKeyParam(c *gin.Context, paramName string) (int64, bool) {
	idStr := c.Param(paramName)
	bizKey, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return 0, false
	}
	return bizKey, true
}

// ResolveBizKey parses the bizKey param, looks up the entity, returns internal uint ID.
// On parse failure: responds with ErrValidation (HTTP 400), returns (0, false).
// On lookup failure: if lookupFn returns apperrors.ErrNotFound, responds HTTP 404;
// any other error is propagated via RespondError (defaults to HTTP 500).
// Returns (0, false) on any failure.
func ResolveBizKey(c *gin.Context, paramName string, lookupFn func(ctx context.Context, bizKey int64) (uint, error)) (uint, bool) {
	bizKey, ok := ParseBizKeyParam(c, paramName)
	if !ok {
		return 0, false
	}
	id, err := lookupFn(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return 0, false
	}
	return id, true
}
