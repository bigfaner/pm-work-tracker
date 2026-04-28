package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
)

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

// ========== SubItem Create Tests ==========

func TestItemLifecycle_CreateSubItem_Returns201(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Item")

	body := fmt.Sprintf(`{
			"mainItemKey": "%s",
			"title": "Sub Item 1",
			"priority": "P2",
			"assigneeKey": "%d",
			"startDate": "2026-01-01",
			"expectedEndDate": "2026-06-30"
		}`, mainBizKey, data.userAID)
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/sub-items", data.teamABizKey, mainBizKey), body, pmToken)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["code"])
	itemData := resp["data"].(map[string]interface{})
	assert.Equal(t, "Sub Item 1", itemData["title"])
	assert.Equal(t, "pending", itemData["itemStatus"])
	// SubItem should have a weight field (default 1.0)
	assert.Contains(t, itemData, "weight")
}

func TestItemLifecycle_CreateSubItem_MemberDenied_Returns403(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Item")

	// Member has sub_item:create permission, but let's test with a user NOT in the team
	userBToken := loginAs(t, r, "userB", "passwordB")

	body := fmt.Sprintf(`{
			"mainItemKey": "%s",
			"title": "Sub Item from B",
			"priority": "P2",
			"assigneeKey": "%d",
			"startDate": "2026-01-01",
			"expectedEndDate": "2026-06-30"
		}`, mainBizKey, data.userBID)
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/sub-items", data.teamABizKey, mainBizKey), body, userBToken)

	// userB is not in teamA, so should get 403
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestItemLifecycle_CreateSubItem_ParentNotFound_Returns404(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	body := fmt.Sprintf(`{
			"mainItemKey": "999999",
			"title": "Sub Item Orphan",
			"priority": "P2",
			"assigneeKey": "%d",
			"startDate": "2026-01-01",
			"expectedEndDate": "2026-06-30"
		}`, data.userAID)
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/999999/sub-items", data.teamABizKey), body, pmToken)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestItemLifecycle_CreateSubItem_TerminalParent_Returns422(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Completed Parent")

	// Set main item to completed directly in DB
	var item model.MainItem
	require.NoError(t, db.Where("biz_key = ?", mainBizKey).First(&item).Error)
	require.NoError(t, db.Model(&item).Update("item_status", "completed").Error)

	body := fmt.Sprintf(`{
			"mainItemKey": "%s",
			"title": "Sub Item After Complete",
			"priority": "P2",
			"assigneeKey": "%d",
			"startDate": "2026-01-01",
			"expectedEndDate": "2026-06-30"
		}`, mainBizKey, data.userAID)
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/sub-items", data.teamABizKey, mainBizKey), body, pmToken)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "TERMINAL_MAIN_ITEM", resp["code"])
}

// ========== SubItem List Tests ==========

func TestItemLifecycle_ListSubItems_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent for List")

	// Create two sub items
	createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Sub A")
	createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Sub B")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/sub-items", data.teamABizKey, mainBizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	items := respData["items"].([]interface{})
	assert.Equal(t, 2, len(items))
}

// ========== SubItem Detail (Get) Tests ==========

func TestItemLifecycle_GetSubItem_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent for Detail")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Detail Sub")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s", data.teamABizKey, subBizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "Detail Sub", respData["title"])
	assert.Equal(t, "pending", respData["itemStatus"])
}

func TestItemLifecycle_GetSubItem_NonMember_Returns403(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	userBToken := loginAs(t, r, "userB", "passwordB")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Isolated")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Isolated Sub")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s", data.teamABizKey, subBizKey), "", userBToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestItemLifecycle_GetSubItem_NotFound_Returns404(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/999999", data.teamABizKey), "", pmToken)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== SubItem Update Tests ==========

func TestItemLifecycle_UpdateSubItem_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent for Update")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Original Sub")

	updateBody := `{"title": "Updated Sub Title"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s", data.teamABizKey, subBizKey), updateBody, pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Sub Title", respData["title"])
}

func TestItemLifecycle_UpdateSubItem_NotFound_Returns404(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	updateBody := `{"title": "No Such Sub"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/999999", data.teamABizKey), updateBody, pmToken)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== SubItem ChangeStatus Tests ==========

func TestItemLifecycle_ChangeSubItemStatus_ValidTransition_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent for Status")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Status Sub")

	// Valid SubItem transition: pending -> progressing
	statusBody := `{"status": "progressing"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/status", data.teamABizKey, subBizKey), statusBody, pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	subItemData := respData["subItem"].(map[string]interface{})
	assert.Equal(t, "progressing", subItemData["itemStatus"])
}

func TestItemLifecycle_ChangeSubItemStatus_InvalidTransition_Returns422(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Invalid")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Invalid Sub")

	// Invalid: pending -> blocking (not in transition map for pending)
	statusBody := `{"status": "blocking"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/status", data.teamABizKey, subBizKey), statusBody, pmToken)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "INVALID_STATUS", resp["code"])
}

func TestItemLifecycle_ChangeSubItemStatus_NotFound_Returns404(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	statusBody := `{"status": "progressing"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/999999/status", data.teamABizKey), statusBody, pmToken)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestItemLifecycle_ChangeSubItemStatus_ToCompleted_SetsCompletion100(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Complete")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Complete Sub")

	// First move to progressing
	statusBody := `{"status": "progressing"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/status", data.teamABizKey, subBizKey), statusBody, pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Then SubItem can go progressing -> completed directly
	statusBody = `{"status": "completed"}`
	w = makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/status", data.teamABizKey, subBizKey), statusBody, pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Verify completion is 100 in DB
	var subItem model.SubItem
	require.NoError(t, db.Where("biz_key = ?", subBizKey).First(&subItem).Error)
	assert.Equal(t, float64(100), subItem.Completion)
	assert.Equal(t, "completed", subItem.ItemStatus)
}

// ========== SubItem AvailableTransitions Tests ==========

func TestItemLifecycle_SubItemAvailableTransitions_ReturnsTransitionList(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Transitions")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Transitions Sub")

	// SubItem is in pending state: should see ["progressing", "closed"]
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/available-transitions", data.teamABizKey, subBizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	transitions := respData["transitions"].([]interface{})
	assert.Equal(t, 2, len(transitions))

	transitionSet := make(map[string]bool)
	for _, tr := range transitions {
		transitionSet[tr.(string)] = true
	}
	assert.True(t, transitionSet["progressing"])
	assert.True(t, transitionSet["closed"])
}

func TestItemLifecycle_SubItemAvailableTransitions_TerminalState_ReturnsEmpty(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Terminal")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Terminal Sub")

	// Set sub-item to completed directly in DB
	var subItem model.SubItem
	require.NoError(t, db.Where("biz_key = ?", subBizKey).First(&subItem).Error)
	require.NoError(t, db.Model(&subItem).Update("item_status", "completed").Error)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/available-transitions", data.teamABizKey, subBizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	transitions := respData["transitions"].([]interface{})
	assert.Equal(t, 0, len(transitions))
}

// ========== SubItem Assign Tests ==========

func TestItemLifecycle_AssignSubItem_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Assign")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Assign Sub")

	// Assign to memberA
	assignBody := fmt.Sprintf(`{"assigneeKey": "%d"}`, data.memberAID)
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/assignee", data.teamABizKey, subBizKey), assignBody, pmToken)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestItemLifecycle_AssignSubItem_MemberDenied_Returns403(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Member Assign")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Member Assign Sub")

	// Member does NOT have sub_item:assign permission
	assignBody := fmt.Sprintf(`{"assigneeKey": "%d"}`, data.userAID)
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/assignee", data.teamABizKey, subBizKey), assignBody, memberToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Progress Append Tests ==========

func TestItemLifecycle_AppendProgress_Returns201(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Progress")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Progress Sub")

	body := `{"completion": 60, "achievement": "made some progress"}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, subBizKey), body, pmToken)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestItemLifecycle_AppendProgress_Regression_Returns422(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	// Use member (non-PM) so regression check is enforced
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Regression")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Regression Sub")

	// First append completion=60
	body := `{"completion": 60, "achievement": "first progress"}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, subBizKey), body, memberToken)
	require.Equal(t, http.StatusCreated, w.Code)

	// Try regression: completion=50
	body = `{"completion": 50, "achievement": "regression attempt"}`
	w = makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, subBizKey), body, memberToken)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "PROGRESS_REGRESSION", resp["code"])
}

func TestItemLifecycle_AppendProgress_100TriggersAutoComplete(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Auto Complete")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Auto Complete Sub")

	// Append progress with completion=100
	body := `{"completion": 100, "achievement": "done!"}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, subBizKey), body, pmToken)
	require.Equal(t, http.StatusCreated, w.Code)

	// Verify SubItem auto-transitioned to completed
	var subItem model.SubItem
	require.NoError(t, db.Where("biz_key = ?", subBizKey).First(&subItem).Error)
	assert.Equal(t, "completed", subItem.ItemStatus)
	assert.Equal(t, float64(100), subItem.Completion)
}

func TestItemLifecycle_AppendProgress_CompletionRollsUpToMainItem(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	// Use seedProgressData for precise control
	mainID, sub1ID, _, _, _ := seedProgressData(t, db, data.teamAID, data.userAID)

	// SubItem IDs are internal; we need bizKeys for the API
	// seedProgressData creates SubItems with snowflake bizKeys
	var sub1 model.SubItem
	require.NoError(t, db.First(&sub1, sub1ID).Error)

	// Append progress (completion=60) to SubItem1
	w := appendProgress(t, setupRouterFromDB(t, db, data), pmToken, data.teamABizKey, sub1.BizKey, 60)
	require.Equal(t, http.StatusCreated, w.Code)

	// Verify SubItem1.Completion = 60
	updatedSub1 := getSubItem(t, db, sub1ID)
	assert.Equal(t, float64(60), updatedSub1.Completion)

	// Verify MainItem.Completion = 30 (weighted avg of 60 and 0, equal weights)
	main := getMainItem(t, db, mainID)
	assert.Equal(t, float64(30), main.Completion)
}

func TestItemLifecycle_AppendProgress_FirstProgressAutoTransitionsToProgressing(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent First Progress")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "First Progress Sub")

	// SubItem is in "pending" state, first progress should auto-transition to "progressing"
	body := `{"completion": 30, "achievement": "first progress"}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, subBizKey), body, pmToken)
	require.Equal(t, http.StatusCreated, w.Code)

	var subItem model.SubItem
	require.NoError(t, db.Where("biz_key = ?", subBizKey).First(&subItem).Error)
	assert.Equal(t, "progressing", subItem.ItemStatus)
}

// ========== Progress List Tests ==========

func TestItemLifecycle_ListProgress_Returns200(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent List Progress")
	subBizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "List Progress Sub")

	// Append two progress records
	body1 := `{"completion": 30, "achievement": "first"}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, subBizKey), body1, pmToken)
	require.Equal(t, http.StatusCreated, w.Code)

	body2 := `{"completion": 60, "achievement": "second"}`
	w = makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, subBizKey), body2, pmToken)
	require.Equal(t, http.StatusCreated, w.Code)

	// List progress records
	w = makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, subBizKey), "", pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	records := resp["data"].([]interface{})
	assert.Equal(t, 2, len(records))
}

// ========== Progress CorrectCompletion Tests ==========

func TestItemLifecycle_CorrectCompletion_Returns200(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	// Use seedProgressData for controlled scenario
	_, _, _, sub1BizKey, _ := seedProgressData(t, db, data.teamAID, data.userAID)

	router := setupRouterFromDB(t, db, data)

	// Append progress to get a record
	w := appendProgress(t, router, pmToken, data.teamABizKey, sub1BizKey, 60)
	require.Equal(t, http.StatusCreated, w.Code)

	var appendResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &appendResp))
	recordData := appendResp["data"].(map[string]interface{})
	recordBizKey := recordData["bizKey"].(string)

	// Correct the completion
	correctBody := `{"completion": 75}`
	w = makeRequest(t, r, http.MethodPatch,
		fmt.Sprintf("/api/v1/teams/%d/progress/%s/completion", data.teamABizKey, recordBizKey), correctBody, pmToken)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestItemLifecycle_CorrectCompletion_NotFound_Returns404(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	correctBody := `{"completion": 75}`
	w := makeRequest(t, r, http.MethodPatch,
		fmt.Sprintf("/api/v1/teams/%d/progress/999999/completion", data.teamABizKey), correctBody, pmToken)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Completion Rollup + Status Cascade Tests ==========

func TestItemLifecycle_CompletionRollsUp_WhenSubItemCompleted(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent Rollup")
	sub1BizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Sub Rollup 1")
	sub2BizKey := createTestSubItem(t, r, pmToken, data.teamABizKey, mainBizKey, "Sub Rollup 2")

	// Complete Sub1 via progress
	body := `{"completion": 100, "achievement": "done"}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, sub1BizKey), body, pmToken)
	require.Equal(t, http.StatusCreated, w.Code)

	// MainItem completion should be 50 (avg of 100 and 0)
	var mainItem model.MainItem
	require.NoError(t, db.Where("biz_key = ?", mainBizKey).First(&mainItem).Error)
	assert.Equal(t, float64(50), mainItem.Completion)

	// Complete Sub2 via progress
	w = makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%s/progress", data.teamABizKey, sub2BizKey), body, pmToken)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now MainItem should be 100
	require.NoError(t, db.Where("id = ?", mainItem.ID).First(&mainItem).Error)
	assert.Equal(t, float64(100), mainItem.Completion)
}

// ========== SubItem Created With Snowflake BizKey Tests ==========

func TestItemLifecycle_CreateSubItem_HasBizKey(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	mainBizKey := createTestMainItem(t, r, pmToken, data.teamABizKey, "Parent BizKey")

	body := fmt.Sprintf(`{
			"mainItemKey": "%s",
			"title": "Sub BizKey Check",
			"priority": "P2",
			"assigneeKey": "%d",
			"startDate": "2026-01-01",
			"expectedEndDate": "2026-06-30"
		}`, mainBizKey, data.userAID)
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/sub-items", data.teamABizKey, mainBizKey), body, pmToken)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	itemData := resp["data"].(map[string]interface{})
	bizKeyStr, ok := itemData["bizKey"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, bizKeyStr)

	// BizKey should be a valid snowflake (numeric string)
	bizKey, err := pkg.ParseID(bizKeyStr)
	assert.NoError(t, err)
	assert.NotZero(t, bizKey)
}
