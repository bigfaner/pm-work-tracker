package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/model"
)

// ========== Progress Append → MainItem Completion Tests ==========

func TestProgress_AppendToSubItem1_UpdatesMainItemCompletion(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	mainID, sub1ID, _, sub1BizKey, _ := seedProgressData(t, db, data.teamAID, data.userAID)

	// Append progress (completion=60) to SubItem1
	w := appendProgress(t, r, token, data.teamABizKey, sub1BizKey, 60)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify SubItem1.Completion = 60
	sub1 := getSubItem(t, db, sub1ID)
	assert.Equal(t, float64(60), sub1.Completion)

	// Verify MainItem.Completion = 30 (avg of 60 and 0)
	main := getMainItem(t, db, mainID)
	assert.Equal(t, float64(30), main.Completion)
}

func TestProgress_AppendToSubItem2_UpdatesMainItemCompletion(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	mainID, _, _, sub1BizKey, sub2BizKey := seedProgressData(t, db, data.teamAID, data.userAID)

	// First append to SubItem1 (completion=60)
	w := appendProgress(t, r, token, data.teamABizKey, sub1BizKey, 60)
	require.Equal(t, http.StatusCreated, w.Code)

	// Then append to SubItem2 (completion=80)
	w = appendProgress(t, r, token, data.teamABizKey, sub2BizKey, 80)
	require.Equal(t, http.StatusCreated, w.Code)

	// Verify MainItem.Completion = 70 (avg of 60 and 80)
	main := getMainItem(t, db, mainID)
	assert.Equal(t, float64(70), main.Completion)
}

func TestProgress_RegressionBlocked_Returns422(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	// Use a regular member (not PM) so the regression check is enforced
	token := loginAs(t, r, "memberA", "passwordMemberA")

	mainID, _, _, sub1BizKey, _ := seedProgressData(t, db, data.teamAID, data.userAID)

	// First append completion=60 to SubItem1
	w := appendProgress(t, r, token, data.teamABizKey, sub1BizKey, 60)
	require.Equal(t, http.StatusCreated, w.Code)

	// Capture MainItem completion after first append
	mainAfterFirst := getMainItem(t, db, mainID)
	completionBefore := mainAfterFirst.Completion

	// Try to append completion=50 (regression from 60) — should fail
	w = appendProgress(t, r, token, data.teamABizKey, sub1BizKey, 50)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "PROGRESS_REGRESSION", resp["code"])

	// Verify MainItem.Completion unchanged
	mainAfterRegression := getMainItem(t, db, mainID)
	assert.Equal(t, completionBefore, mainAfterRegression.Completion)
}

// ========== ItemPool Assign Transaction Tests ==========

func TestItemPool_Assign_Success(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	poolID, mainItemID, poolBizKey, mainItemBizKey := seedPoolData(t, db, data.teamAID, data.userAID)

	// Assign the pool item
	body := fmt.Sprintf(`{"mainItemKey":"%d","assigneeKey":"%d","priority":"P2","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`, mainItemBizKey, data.userAID)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/assign", data.teamABizKey, poolBizKey),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify ItemPool.Status = "assigned"
	var pool model.ItemPool
	require.NoError(t, db.First(&pool, poolID).Error)
	assert.Equal(t, "assigned", pool.PoolStatus)
	assert.NotNil(t, pool.AssignedSubKey)

	// Verify a new SubItem exists under the main item
	var subItems []model.SubItem
	require.NoError(t, db.Where("main_item_key = ?", mainItemID).Find(&subItems).Error)
	require.Len(t, subItems, 1)
	assert.Equal(t, "Pool Item Title", subItems[0].Title)
	assert.Equal(t, int64(data.userAID), *subItems[0].AssigneeKey)
}

func TestItemPool_Assign_Rollback_OnInvalidMainItem(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	poolID, _, poolBizKey, _ := seedPoolData(t, db, data.teamAID, data.userAID)

	// Assign with a non-existent mainItemID to trigger failure
	invalidMainID := uint(99999)
	body := fmt.Sprintf(`{"mainItemKey":"%d","assigneeKey":"%d","priority":"P2","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`, invalidMainID, data.userAID)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/assign", data.teamABizKey, poolBizKey),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// Should fail with 404 (item not found)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify ItemPool.Status remains "pending"
	var pool model.ItemPool
	require.NoError(t, db.First(&pool, poolID).Error)
	assert.Equal(t, "pending", pool.PoolStatus)

	// Verify no SubItem was created
	var count int64
	db.Model(&model.SubItem{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

// ========== Weekly Report Integration Tests ==========

func TestWeeklyExport_ReturnsMarkdownWithMainItemTitle(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	// Use a known Monday
	weekStart := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC) // Monday
	mainItemTitle := seedReportData(t, db, data.teamAID, data.userAID, weekStart)

	// GET /api/v1/teams/:teamId/reports/weekly/export?weekStart=2026-04-13
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/reports/weekly/export?weekStart=%s",
			data.teamABizKey, weekStart.Format("2006-01-02")), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/markdown", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, mainItemTitle)
	assert.Contains(t, body, "Completed half the work")
}
