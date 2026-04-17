package errors

import (
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a structured application error with HTTP mapping.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrUnauthorized       = &AppError{Code: "UNAUTHORIZED", Status: 401, Message: "authentication required"}
	ErrForbidden          = &AppError{Code: "FORBIDDEN", Status: 403, Message: "insufficient permissions"}
	ErrTeamNotFound       = &AppError{Code: "TEAM_NOT_FOUND", Status: 404, Message: "team not found"}
	ErrItemNotFound       = &AppError{Code: "ITEM_NOT_FOUND", Status: 404, Message: "item not found"}
	ErrNotTeamMember      = &AppError{Code: "NOT_TEAM_MEMBER", Status: 403, Message: "not a member of this team"}
	ErrInvalidStatus      = &AppError{Code: "INVALID_STATUS", Status: 422, Message: "invalid status transition"}
	ErrArchiveNotAllowed  = &AppError{Code: "ARCHIVE_NOT_ALLOWED", Status: 422, Message: "only completed or closed items can be archived"}
	ErrProgressRegression = &AppError{Code: "PROGRESS_REGRESSION", Status: 422, Message: "completion cannot be lower than previous record"}
	ErrValidation         = &AppError{Code: "VALIDATION_ERROR", Status: 400, Message: "request validation failed"}
	ErrInternal           = &AppError{Code: "INTERNAL_ERROR", Status: 500, Message: "internal server error"}
	ErrNotFound           = &AppError{Code: "NOT_FOUND", Status: 404, Message: "resource not found"}
	ErrAlreadyExists      = &AppError{Code: "ALREADY_EXISTS", Status: 409, Message: "resource already exists"}
	ErrAlreadyMember      = &AppError{Code: "ALREADY_MEMBER", Status: 422, Message: "user is already a team member"}
	ErrCannotRemoveSelf   = &AppError{Code: "CANNOT_REMOVE_SELF", Status: 422, Message: "PM cannot remove themselves"}
)

// RespondError writes a standard error envelope. Non-AppError values fall back to ErrInternal.
func RespondError(c *gin.Context, err error) {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		c.JSON(appErr.Status, gin.H{"code": appErr.Code, "message": appErr.Message})
		return
	}
	c.JSON(ErrInternal.Status, gin.H{"code": ErrInternal.Code, "message": ErrInternal.Message})
}

// RespondOK writes a standard success envelope with the given data.
func RespondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data})
}
