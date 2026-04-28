package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

// ========== Submit Tests ==========

func TestItemPool_Submit_Returns201(t *testing.T) {
	r, data := setupTestRouter(t)
	token := loginAs(t, r, "userA", "passwordA")

	body := `{"title":"New Pool Item","background":"some context","expectedOutput":"a feature"}`
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "pending", respData["poolStatus"])
	assert.Equal(t, "New Pool Item", respData["title"])
}

func TestItemPool_Submit_MissingTitle_Returns400(t *testing.T) {
	r, data := setupTestRouter(t)
	token := loginAs(t, r, "userA", "passwordA")

	body := `{"background":"no title"}`
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemPool_Submit_MemberDenied_Returns403(t *testing.T) {
	r, data := setupTestRouter(t)
	// userB is PM of teamB, not a member of teamA — should be denied
	otherPMToken := loginAs(t, r, "userB", "passwordB")

	body := `{"title":"Pool Item from outsider"}`
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, otherPMToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== List Tests ==========

func TestItemPool_List_WithStatusFilter_Returns200(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	// Seed a pending pool item
	poolItem := &model.ItemPool{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:      data.teamABizKey,
		Title:        "Pending Item",
		SubmitterKey: int64(data.userAID),
		PoolStatus:   "pending",
	}
	require.NoError(t, db.Create(poolItem).Error)

	// Seed a rejected pool item
	rejectedItem := &model.ItemPool{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:      data.teamABizKey,
		Title:        "Rejected Item",
		SubmitterKey: int64(data.userAID),
		PoolStatus:   "rejected",
	}
	require.NoError(t, db.Create(rejectedItem).Error)

	// Filter by status=pending
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool?status=pending", data.teamABizKey)
	w := makeRequest(t, r, http.MethodGet, path, "", token)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	items := respData["items"].([]interface{})
	assert.Len(t, items, 1)

	item := items[0].(map[string]interface{})
	assert.Equal(t, "pending", item["poolStatus"])
	assert.Equal(t, "Pending Item", item["title"])
}

// ========== Detail Tests ==========

func TestItemPool_Detail_Returns200(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	_, _, poolBizKey, _ := seedPoolData(t, db, data.teamABizKey, data.userAID)

	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodGet, path, "", token)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "Pool Item Title", respData["title"])
	assert.Equal(t, "pending", respData["poolStatus"])
}

func TestItemPool_Detail_NotFound_Returns404(t *testing.T) {
	r, data := setupTestRouter(t)
	token := loginAs(t, r, "userA", "passwordA")

	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/9999999", data.teamABizKey)
	w := makeRequest(t, r, http.MethodGet, path, "", token)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Assign Tests ==========

func TestItemPool_Assign_CreatesSubItemAndUpdatesPoolStatus(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	_, _, poolBizKey, mainItemBizKey := seedPoolData(t, db, data.teamABizKey, data.userAID)

	body := fmt.Sprintf(`{
		"mainItemKey": "%d",
		"assigneeKey": "%d",
		"priority": "P2",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30"
	}`, mainItemBizKey, data.userAID)
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/assign", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	// Use json.Number-aware decoder to avoid float64 precision loss on large int64 BizKeys
	decoder := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&resp))
	respData := resp["data"].(map[string]interface{})
	// Compare mainItemBizKey as string to avoid float64 precision issues
	assert.Equal(t, fmt.Sprintf("%d", mainItemBizKey), respData["mainItemBizKey"].(json.Number).String())
	subKey, ok := respData["subItemBizKey"]
	assert.True(t, ok)
	assert.NotEmpty(t, subKey)

	// Verify pool status in DB
	var pool model.ItemPool
	require.NoError(t, db.First(&pool, "biz_key = ?", poolBizKey).Error)
	assert.Equal(t, "assigned", pool.PoolStatus)
}

func TestItemPool_Assign_InvalidMainItem_RollsBack(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	_, _, poolBizKey, _ := seedPoolData(t, db, data.teamABizKey, data.userAID)

	// Use a non-existent main item BizKey
	body := fmt.Sprintf(`{
		"mainItemKey": "9999999",
		"assigneeKey": "%d",
		"priority": "P2",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30"
	}`, data.userAID)
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/assign", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify pool status is still pending (rollback)
	var pool model.ItemPool
	require.NoError(t, db.First(&pool, "biz_key = ?", poolBizKey).Error)
	assert.Equal(t, "pending", pool.PoolStatus)

	// Verify no SubItem was created
	var count int64
	require.NoError(t, db.Model(&model.SubItem{}).Where("team_key = ?", data.teamAID).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestItemPool_Assign_AlreadyProcessed_Returns422(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	_, _, poolBizKey, mainItemBizKey := seedPoolData(t, db, data.teamABizKey, data.userAID)

	// First assign succeeds
	body := fmt.Sprintf(`{
		"mainItemKey": "%d",
		"assigneeKey": "%d",
		"priority": "P2",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30"
	}`, mainItemBizKey, data.userAID)
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/assign", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)
	require.Equal(t, http.StatusOK, w.Code)

	// Second assign fails (already processed)
	w = makeRequest(t, r, http.MethodPost, path, body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ITEM_ALREADY_PROCESSED", resp["code"])
}

func TestItemPool_Assign_MemberDenied_Returns403(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	_, _, poolBizKey, mainItemBizKey := seedPoolData(t, db, data.teamABizKey, data.userAID)

	body := fmt.Sprintf(`{
		"mainItemKey": "%d",
		"assigneeKey": "%d",
		"priority": "P2",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30"
	}`, mainItemBizKey, data.userAID)
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/assign", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, memberToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== ConvertToMain Tests ==========

func TestItemPool_ConvertToMain_Returns200(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	_, _, poolBizKey, _ := seedPoolData(t, db, data.teamABizKey, data.userAID)

	body := fmt.Sprintf(`{
		"priority": "P1",
		"assigneeKey": "%d",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30"
	}`, data.userAID)
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/convert-to-main", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	respData := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, respData["mainItemBizKey"])

	// Verify pool status is assigned
	var pool model.ItemPool
	require.NoError(t, db.First(&pool, "biz_key = ?", poolBizKey).Error)
	assert.Equal(t, "assigned", pool.PoolStatus)
}

func TestItemPool_ConvertToMain_MemberDenied_Returns403(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	_, _, poolBizKey, _ := seedPoolData(t, db, data.teamABizKey, data.userAID)

	body := fmt.Sprintf(`{
		"priority": "P1",
		"assigneeKey": "%d",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30"
	}`, data.memberAID)
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/convert-to-main", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, memberToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestItemPool_ConvertToMain_AlreadyProcessed_Returns422(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	_, _, poolBizKey, _ := seedPoolData(t, db, data.teamABizKey, data.userAID)

	// First convert succeeds
	body := fmt.Sprintf(`{
		"priority": "P1",
		"assigneeKey": "%d",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30"
	}`, data.userAID)
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/convert-to-main", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)
	require.Equal(t, http.StatusOK, w.Code)

	// Second convert fails
	w = makeRequest(t, r, http.MethodPost, path, body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ITEM_ALREADY_PROCESSED", resp["code"])
}

// ========== Reject Tests ==========

func TestItemPool_Reject_WithReason_Returns200(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	_, _, poolBizKey, _ := seedPoolData(t, db, data.teamABizKey, data.userAID)

	body := `{"reason":"Not aligned with current roadmap"}`
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/reject", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify pool status in DB
	var pool model.ItemPool
	require.NoError(t, db.First(&pool, "biz_key = ?", poolBizKey).Error)
	assert.Equal(t, "rejected", pool.PoolStatus)
}

func TestItemPool_Reject_MissingReason_Returns400(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	_, _, poolBizKey, _ := seedPoolData(t, db, data.teamABizKey, data.userAID)

	body := `{}`
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/reject", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemPool_Reject_MemberDenied_Returns403(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	_, _, poolBizKey, _ := seedPoolData(t, db, data.teamABizKey, data.userAID)

	body := `{"reason":"bad"}`
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/reject", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, memberToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestItemPool_Reject_AlreadyProcessed_Returns422(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	_, _, poolBizKey, _ := seedPoolData(t, db, data.teamABizKey, data.userAID)

	// First reject succeeds
	body := `{"reason":"not needed anymore"}`
	path := fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/reject", data.teamABizKey, poolBizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)
	require.Equal(t, http.StatusOK, w.Code)

	// Second reject fails
	w = makeRequest(t, r, http.MethodPost, path, body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ITEM_ALREADY_PROCESSED", resp["code"])
}

