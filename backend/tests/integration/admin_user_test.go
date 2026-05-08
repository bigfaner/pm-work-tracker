package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Test Admin User: List Users ==========

func TestAdminUser_ListUsers_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/users", "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	items := d["items"].([]interface{})
	assert.GreaterOrEqual(t, len(items), 4) // userA, userB, memberA, superadmin
	assert.NotNil(t, d["total"])
	assert.NotNil(t, d["page"])
	assert.NotNil(t, d["pageSize"])
}

func TestAdminUser_ListUsers_WithSearchReturns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/users?search=userA", "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	items := d["items"].([]interface{})
	assert.GreaterOrEqual(t, len(items), 1)
}

func TestAdminUser_ListUsers_NonSuperAdminReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	// memberA does not have user:read permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/users", "", token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Test Admin User: Create User ==========

func TestAdminUser_CreateUser_Returns201(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	body := `{"username":"newuser","displayName":"New User","email":"new@test.com"}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/users", body, token)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, "newuser", d["username"])
	assert.Equal(t, "New User", d["displayName"])
	assert.NotEmpty(t, d["initialPassword"])
}

func TestAdminUser_CreateUser_DuplicateUsernameReturns422(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	body := `{"username":"userA","displayName":"Duplicate","email":"dup@test.com"}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/users", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestAdminUser_CreateUser_NonSuperAdminReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	// memberA does not have user:manage_role
	token := loginAs(t, r, "memberA", "passwordMemberA")

	body := `{"username":"newuser","displayName":"New User"}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/users", body, token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Test Admin User: Get User ==========

func TestAdminUser_GetUser_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	// List users to get a valid bizKey
	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/users", "", token)
	require.Equal(t, http.StatusOK, w.Code)
	var listResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	items := listResp["data"].(map[string]interface{})["items"].([]interface{})
	require.GreaterOrEqual(t, len(items), 1)
	firstUser := items[0].(map[string]interface{})
	bizKey := firstUser["bizKey"].(string)

	w = makeRequest(t, r, http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%s", bizKey), "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, bizKey, d["bizKey"])
}

func TestAdminUser_GetUser_NotFoundReturns404(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/users/99999999", "", token)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Test Admin User: Update User ==========

func TestAdminUser_UpdateUser_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	// List users to get userA's bizKey
	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/users?search=userA", "", token)
	require.Equal(t, http.StatusOK, w.Code)
	var listResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	items := listResp["data"].(map[string]interface{})["items"].([]interface{})
	require.GreaterOrEqual(t, len(items), 1)
	userBizKey := items[0].(map[string]interface{})["bizKey"].(string)

	body := `{"displayName":"Updated Name","email":"updated@test.com"}`
	w = makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%s", userBizKey), body, token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Name", d["displayName"])
	assert.Equal(t, "updated@test.com", d["email"])
}

// fetchUserBizKey searches for a user and returns their bizKey.
func fetchUserBizKey(t *testing.T, r *gin.Engine, token, search string) string {
	t.Helper()
	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/users?search="+search, "", token)
	require.Equal(t, http.StatusOK, w.Code)
	var listResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	items := listResp["data"].(map[string]interface{})["items"].([]interface{})
	require.GreaterOrEqual(t, len(items), 1)
	return items[0].(map[string]interface{})["bizKey"].(string)
}

func TestAdminUser_UpdateUser_ValidationErrors(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")
	userBizKey := fetchUserBizKey(t, r, token, "userA")

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"EmptyDisplayName", `{"displayName":""}`, http.StatusBadRequest},
		{"DisplayNameTooLong", fmt.Sprintf(`{"displayName":"%s"}`, strings.Repeat("a", 65)), http.StatusBadRequest},
		{"EmailTooLong", fmt.Sprintf(`{"email":"%s"}`, strings.Repeat("a", 101)), http.StatusBadRequest},
		{"InvalidTeamKey", `{"teamKey":"99999999"}`, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%s", userBizKey), tt.body, token)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAdminUser_UpdateUser_NotFoundReturns404(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	body := `{"displayName":"No One"}`
	w := makeRequest(t, r, http.MethodPut, "/api/v1/admin/users/99999999", body, token)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Test Admin User: Toggle User Status ==========

func TestAdminUser_ToggleStatus_DisableReturns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	// Get userA's bizKey
	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/users?search=userA", "", token)
	require.Equal(t, http.StatusOK, w.Code)
	var listResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	items := listResp["data"].(map[string]interface{})["items"].([]interface{})
	require.GreaterOrEqual(t, len(items), 1)
	userBizKey := items[0].(map[string]interface{})["bizKey"].(string)

	body := `{"status":"disabled"}`
	w = makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%s/status", userBizKey), body, token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, "disabled", d["userStatus"])
}

func TestAdminUser_ToggleStatus_EnableReturns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")
	userBizKey := fetchUserBizKey(t, r, token, "userA")

	// Disable first
	body := `{"status":"disabled"}`
	w := makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%s/status", userBizKey), body, token)
	require.Equal(t, http.StatusOK, w.Code)

	// Re-enable
	body = `{"status":"enabled"}`
	w = makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%s/status", userBizKey), body, token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, "enabled", d["userStatus"])
}

func TestAdminUser_ToggleStatus_DisableSelfReturns422(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")
	adminBizKey := fetchUserBizKey(t, r, token, "superadmin")

	body := `{"status":"disabled"}`
	w := makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%s/status", adminBizKey), body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestAdminUser_ToggleStatus_NotFoundReturns404(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	body := `{"status":"disabled"}`
	w := makeRequest(t, r, http.MethodPut, "/api/v1/admin/users/99999999/status", body, token)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Test Admin User: List Teams ==========

func TestAdminUser_ListTeams_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/teams", "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	items := d["items"].([]interface{})
	assert.GreaterOrEqual(t, len(items), 2) // Team A, Team B
}
