package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Test Role: List (role:read) ==========

func TestRole_ListRoles_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/roles", "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	items := d["items"].([]interface{})
	assert.GreaterOrEqual(t, len(items), 2) // pm + member at minimum
}

func TestRole_ListRoles_NonSuperAdminWithoutPermissionReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA has no role:read permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/roles", "", token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Test Role: Get (role:read) ==========

func TestRole_GetRole_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	// List to get a valid bizKey
	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/roles", "", token)
	require.Equal(t, http.StatusOK, w.Code)
	var listResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	items := listResp["data"].(map[string]interface{})["items"].([]interface{})
	require.GreaterOrEqual(t, len(items), 1)
	bizKey := items[0].(map[string]interface{})["bizKey"].(string)

	w = makeRequest(t, r, http.MethodGet, fmt.Sprintf("/api/v1/admin/roles/%s", bizKey), "", token)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Test Role: Create (role:create) ==========

func TestRole_CreateRole_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	body := `{"name":"test-role","description":"Test role","permissionCodes":["main_item:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", body, token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, "test-role", d["roleName"])
}

func TestRole_CreateRole_NonSuperAdminWithoutPermissionReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA has no role:create permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	body := `{"name":"test-role","description":"Test role","permissionCodes":["main_item:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", body, token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Test Role: Update (role:update) ==========

func TestRole_UpdateRole_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	// Create a role first
	body := `{"name":"updatable-role","description":"Before update","permissionCodes":["main_item:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", body, token)
	require.Equal(t, http.StatusOK, w.Code)
	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	bizKey := createResp["data"].(map[string]interface{})["bizKey"].(string)

	// Update it
	updateBody := `{"description":"After update","permissionCodes":["main_item:read","main_item:create"]}`
	w = makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%s", bizKey), updateBody, token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, "After update", d["roleDesc"])
}

func TestRole_UpdateRole_NonSuperAdminWithoutPermissionReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA has no role:update permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	body := `{"description":"Unauthorized update"}`
	w := makeRequest(t, r, http.MethodPut, "/api/v1/admin/roles/99999999", body, token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Test Role: Delete (role:delete) ==========

func TestRole_DeleteRole_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	// Create a role to delete
	body := `{"name":"deletable-role","description":"To be deleted","permissionCodes":["main_item:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", body, token)
	require.Equal(t, http.StatusOK, w.Code)
	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	bizKey := createResp["data"].(map[string]interface{})["bizKey"].(string)

	// Delete it
	w = makeRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/v1/admin/roles/%s", bizKey), "", token)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRole_DeleteRole_NonSuperAdminWithoutPermissionReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA has no role:delete permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	w := makeRequest(t, r, http.MethodDelete, "/api/v1/admin/roles/99999999", "", token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
