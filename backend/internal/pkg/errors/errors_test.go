package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAppError_ImplementsErrorInterface(t *testing.T) {
	var err error = ErrUnauthorized
	assert.Contains(t, err.Error(), ErrUnauthorized.Message)
}

func TestAppError_ErrorMethod(t *testing.T) {
	appErr := &AppError{Code: "TEST_CODE", Message: "test message", Status: 400}
	assert.Equal(t, "test message", appErr.Error())
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		code     string
		status   int
		message  string
	}{
		{"Unauthorized", ErrUnauthorized, "UNAUTHORIZED", 401, "authentication required"},
		{"Forbidden", ErrForbidden, "FORBIDDEN", 403, "insufficient permissions"},
		{"TeamNotFound", ErrTeamNotFound, "TEAM_NOT_FOUND", 404, "team not found"},
		{"ItemNotFound", ErrItemNotFound, "ITEM_NOT_FOUND", 404, "item not found"},
		{"NotTeamMember", ErrNotTeamMember, "NOT_TEAM_MEMBER", 403, "not a member of this team"},
		{"InvalidStatus", ErrInvalidStatus, "INVALID_STATUS", 422, "invalid status transition"},
		{"ArchiveNotAllowed", ErrArchiveNotAllowed, "ARCHIVE_NOT_ALLOWED", 422, "only completed or closed items can be archived"},
		{"ProgressRegression", ErrProgressRegression, "PROGRESS_REGRESSION", 422, "completion cannot be lower than previous record"},
		{"Validation", ErrValidation, "VALIDATION_ERROR", 400, "request validation failed"},
		{"Internal", ErrInternal, "INTERNAL_ERROR", 500, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.Equal(t, tt.status, tt.err.Status)
			assert.Equal(t, tt.message, tt.err.Message)
		})
	}
}

func TestRespondError_WithAppError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondError(c, ErrForbidden)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "FORBIDDEN", body["code"])
	assert.Equal(t, "insufficient permissions", body["message"])
}

func TestRespondError_WithGenericError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondError(c, errors.New("something broke"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", body["code"])
	assert.Equal(t, "internal server error", body["message"])
}

func TestRespondOK(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondOK(c, map[string]string{"name": "test"})

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), body["code"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test", data["name"])
}

func TestRespondOK_WithNilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondOK(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), body["code"])
	assert.Nil(t, body["data"])
}
