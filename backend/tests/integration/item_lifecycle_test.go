package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
)

// setupLifecycleTest creates a fresh DB and router for item lifecycle tests.
// It extends setupRBACTestDB by adding main_item:change_status to the PM role,
// which is required by the ChangeStatus endpoint but missing from the seed data.
func setupLifecycleTest(t *testing.T) (*gin.Engine, *seedData, *gorm.DB) {
	t.Helper()

	db, data := setupRBACTestDB(t)

	// Add main_item:change_status permission for PM role (required by router but missing from seed)
	pmRoleID := findRoleIDByName(t, db, "pm")
	require.NoError(t, db.Create(&model.RolePermission{
		RoleID:          pmRoleID,
		PermissionCode:  "main_item:change_status",
	}).Error)

	r := setupRBACTestRouter(t, db, data)
	return r, data, db
}

// createTestMainItem is a helper that creates a MainItem via the API and returns the bizKey string.
func createTestMainItem(t *testing.T, r *gin.Engine, token string, teamBizKey int64, title string) string {
	t.Helper()
	body := fmt.Sprintf(`{
		"title": "%s",
		"priority": "P1",
		"assigneeKey": "1",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30",
		"isKeyItem": false
	}`, title)
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items", teamBizKey), body, token)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	bizKey, ok := data["bizKey"].(string)
	require.True(t, ok, "expected bizKey string in response")
	return bizKey
}

// ========== Create Tests ==========

func TestItemLifecycle_CreateMainItem_Returns201(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	body := `{
		"title": "Test Item",
		"priority": "P1",
		"assigneeKey": "1",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30",
		"isKeyItem": false
	}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), body, pmToken)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["code"])

	itemData := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, itemData["bizKey"])
	assert.Equal(t, "Test Item", itemData["title"])
	assert.Equal(t, "P1", itemData["priority"])
	assert.Equal(t, "pending", itemData["itemStatus"])
}

func TestItemLifecycle_CreateMainItem_MissingTitle_Returns422(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	// Missing required field: title
	body := `{"priority": "P1", "assigneeKey": "1"}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), body, pmToken)

	// Binding validation error returns 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemLifecycle_CreateMainItem_Member_Returns403(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	body := `{
		"title": "Member Item",
		"priority": "P1",
		"assigneeKey": "1",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30"
	}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), body, memberToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== List Tests ==========

func TestItemLifecycle_ListMainItems_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	// Create a couple items first
	createTestMainItem(t, r, pmToken, data.teamABizKey, "Item A")
	createTestMainItem(t, r, pmToken, data.teamABizKey, "Item B")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	items := respData["items"].([]interface{})
	assert.Equal(t, 2, len(items))
	assert.Equal(t, float64(2), respData["total"])
}

func TestItemLifecycle_ListMainItems_Paginated_ReturnsCorrectPage(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	// Create 3 items
	for i := 0; i < 3; i++ {
		createTestMainItem(t, r, pmToken, data.teamABizKey, fmt.Sprintf("Item %d", i))
	}

	// Request page 1, pageSize 2
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items?page=1&pageSize=2", data.teamABizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	items := respData["items"].([]interface{})
	assert.Equal(t, 2, len(items))
	assert.Equal(t, float64(3), respData["total"])
	assert.Equal(t, float64(1), respData["page"])
	assert.Equal(t, float64(2), respData["size"])
}

// ========== Detail (Get) Tests ==========

func TestItemLifecycle_GetMainItem_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Detail Item")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s", data.teamABizKey, bizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "Detail Item", respData["title"])
	// subItems field is present (may be empty array for item with no sub-items)
	assert.Contains(t, respData, "subItems")
}

func TestItemLifecycle_GetMainItem_NonMember_Returns403(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	userBToken := loginAs(t, r, "userB", "passwordB")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Team A Item")

	// UserB is not in TeamA
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s", data.teamABizKey, bizKey), "", userBToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestItemLifecycle_GetMainItem_NotFound_Returns404(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items/999999", data.teamABizKey), "", pmToken)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Update Tests ==========

func TestItemLifecycle_UpdateMainItem_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Original Title")

	updateBody := `{"title": "Updated Title"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s", data.teamABizKey, bizKey), updateBody, pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Title", respData["title"])
}

func TestItemLifecycle_UpdateMainItem_TerminalState_Returns422(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Terminal Item")

	// Directly set the item to a terminal status in DB to simulate completed state
	var item model.MainItem
	require.NoError(t, db.Where("biz_key = ?", bizKey).First(&item).Error)
	require.NoError(t, db.Model(&item).Update("item_status", "completed").Error)

	updateBody := `{"title": "Try Update"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s", data.teamABizKey, bizKey), updateBody, pmToken)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "TERMINAL_MAIN_ITEM", resp["code"])
}

func TestItemLifecycle_UpdateMainItem_Member_Returns403(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "PM Item")

	updateBody := `{"title": "Member Update Attempt"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s", data.teamABizKey, bizKey), updateBody, memberToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestItemLifecycle_UpdateMainItem_NotFound_Returns404(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	updateBody := `{"title": "No Such Item"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/main-items/999999", data.teamABizKey), updateBody, pmToken)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== ChangeStatus Tests ==========

func TestItemLifecycle_ChangeStatus_ValidTransition_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Status Item")

	// Valid transition: pending -> progressing
	statusBody := `{"status": "progressing"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/status", data.teamABizKey, bizKey), statusBody, pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "progressing", respData["itemStatus"])
}

func TestItemLifecycle_ChangeStatus_InvalidTransition_Returns422(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Invalid Transition")

	// Invalid transition: pending -> completed (must go through reviewing first)
	statusBody := `{"status": "completed"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/status", data.teamABizKey, bizKey), statusBody, pmToken)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "INVALID_STATUS", resp["code"])
}

func TestItemLifecycle_ChangeStatus_Member_Returns403(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Member Status Test")

	statusBody := `{"status": "progressing"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/status", data.teamABizKey, bizKey), statusBody, memberToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== AvailableTransitions Tests ==========

func TestItemLifecycle_AvailableTransitions_ReturnsTransitionList(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Transitions Item")

	// Item is in pending state: should see ["progressing", "closed"]
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/available-transitions", data.teamABizKey, bizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	transitions := respData["transitions"].([]interface{})
	assert.Equal(t, 2, len(transitions))

	// Verify expected transitions for pending state
	transitionSet := make(map[string]bool)
	for _, tr := range transitions {
		transitionSet[tr.(string)] = true
	}
	assert.True(t, transitionSet["progressing"])
	assert.True(t, transitionSet["closed"])
}

func TestItemLifecycle_AvailableTransitions_TerminalState_ReturnsEmpty(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Terminal Transitions")

	// Directly set the item to a terminal status in DB
	var item model.MainItem
	require.NoError(t, db.Where("biz_key = ?", bizKey).First(&item).Error)
	require.NoError(t, db.Model(&item).Update("item_status", "completed").Error)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/available-transitions", data.teamABizKey, bizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	transitions := respData["transitions"].([]interface{})
	assert.Equal(t, 0, len(transitions))
}

// ========== Archive Tests ==========

func TestItemLifecycle_Archive_CompletedItem_Returns200(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Archive Completed")

	// Set item to completed state directly
	var item model.MainItem
	require.NoError(t, db.Where("biz_key = ?", bizKey).First(&item).Error)
	require.NoError(t, db.Model(&item).Update("item_status", "completed").Error)

	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/archive", data.teamABizKey, bizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestItemLifecycle_Archive_InProgressItem_Returns422(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	bizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Archive InProgress")

	// Set item to progressing (non-terminal, non-completed/closed)
	var item model.MainItem
	require.NoError(t, db.Where("biz_key = ?", bizKey).First(&item).Error)
	require.NoError(t, db.Model(&item).Update("item_status", "progressing").Error)

	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/archive", data.teamABizKey, bizKey), "", pmToken)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ARCHIVE_NOT_ALLOWED", resp["code"])
}

func TestItemLifecycle_Archive_NotFound_Returns404(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/999999/archive", data.teamABizKey), "", pmToken)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
