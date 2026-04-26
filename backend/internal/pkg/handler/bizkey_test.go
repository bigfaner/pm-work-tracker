package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- ParseBizKeyParam tests ---

func TestParseBizKeyParam_Valid(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.AddParam("id", "123")

	bizKey, ok := ParseBizKeyParam(c, "id")
	assert.True(t, ok)
	assert.Equal(t, int64(123), bizKey)
	assert.Equal(t, http.StatusOK, w.Code) // no error response written
}

func TestParseBizKeyParam_Invalid(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.AddParam("id", "abc")

	bizKey, ok := ParseBizKeyParam(c, "id")
	assert.False(t, ok)
	assert.Equal(t, int64(0), bizKey)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "VALIDATION_ERROR", body["code"])
}

func TestParseBizKeyParam_Empty(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.AddParam("id", "")

	bizKey, ok := ParseBizKeyParam(c, "id")
	assert.False(t, ok)
	assert.Equal(t, int64(0), bizKey)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParseBizKeyParam_Negative(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.AddParam("id", "-5")

	bizKey, ok := ParseBizKeyParam(c, "id")
	assert.True(t, ok)
	assert.Equal(t, int64(-5), bizKey)
}

// --- ResolveBizKey tests ---

func TestResolveBizKey_ParseFails(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.AddParam("subId", "notanumber")

	id, ok := ResolveBizKey(c, "subId", nil)
	assert.False(t, ok)
	assert.Equal(t, uint(0), id)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestResolveBizKey_LookupSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.AddParam("id", "42")

	id, ok := ResolveBizKey(c, "id", func(_ context.Context, bizKey int64) (uint, error) {
		assert.Equal(t, int64(42), bizKey)
		return 7, nil
	})
	assert.True(t, ok)
	assert.Equal(t, uint(7), id)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResolveBizKey_LookupNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.AddParam("id", "42")

	id, ok := ResolveBizKey(c, "id", func(_ context.Context, _ int64) (uint, error) {
		return 0, apperrors.ErrNotFound
	})
	assert.False(t, ok)
	assert.Equal(t, uint(0), id)
	assert.Equal(t, http.StatusNotFound, w.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "NOT_FOUND", body["code"])
}

func TestResolveBizKey_LookupOtherError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.AddParam("id", "42")

	dbErr := errors.New("connection refused")
	id, ok := ResolveBizKey(c, "id", func(_ context.Context, _ int64) (uint, error) {
		return 0, dbErr
	})
	assert.False(t, ok)
	assert.Equal(t, uint(0), id)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "INTERNAL_ERROR", body["code"])
}
