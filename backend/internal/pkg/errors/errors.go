package errors

import (
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	ErrCannotRemoveSelf  = &AppError{Code: "CANNOT_REMOVE_SELF", Status: 422, Message: "PM cannot remove themselves"}
	ErrCannotModifySelf  = &AppError{Code: "CANNOT_MODIFY_SELF", Status: 422, Message: "cannot modify your own permissions"}
	ErrItemAlreadyProcessed = &AppError{Code: "ITEM_ALREADY_PROCESSED", Status: 422, Message: "item already processed"}
	ErrNoData               = &AppError{Code: "NO_DATA", Status: 422, Message: "no data to export"}
	ErrUserExists           = &AppError{Code: "USER_EXISTS", Status: 422, Message: "username already exists"}
	ErrUserNotFound         = &AppError{Code: "USER_NOT_FOUND", Status: 404, Message: "user not found"}
	ErrUserDisabled         = &AppError{Code: "USER_DISABLED", Status: 403, Message: "账号已被禁用"}
	ErrCannotDisableSelf    = &AppError{Code: "CANNOT_DISABLE_SELF", Status: 422, Message: "cannot disable your own account"}
	ErrFutureWeekNotAllowed = &AppError{Code: "FUTURE_WEEK_NOT_ALLOWED", Status: 422, Message: "cannot create progress for future weeks"}
	ErrTerminalMainItem     = &AppError{Code: "TERMINAL_MAIN_ITEM", Status: 422, Message: "cannot modify a completed or closed main item"}
	ErrSubItemsNotTerminal  = &AppError{Code: "SUB_ITEMS_NOT_TERMINAL", Status: 422, Message: "all sub-items must be completed or closed before closing the main item"}
	ErrTeamCodeDuplicate    = &AppError{Code: "TEAM_CODE_DUPLICATE", Status: 400, Message: "该编码已被使用"}
)

// MapNotFound maps gorm.ErrRecordNotFound and ErrNotFound to the provided domain error.
// Returns the original error unchanged if it is not a not-found error.
func MapNotFound(err error, domainErr *AppError) error {
	if err == nil {
		return nil
	}
	if stderrors.Is(err, gorm.ErrRecordNotFound) || stderrors.Is(err, ErrNotFound) {
		return domainErr
	}
	return err
}

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
