package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/handler"
	"pm-work-tracker/backend/internal/model"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
	"pm-work-tracker/backend/internal/service"
)

// ========== Progress Append → MainItem Completion Tests ==========

// seedProgressData creates a MainItem with two SubItems (weight=1 each) for progress tests.
// Returns the main item ID and the two sub item IDs.
func seedProgressData(t *testing.T, db *gorm.DB, teamID, userID uint) (mainItemID, subItem1ID, subItem2ID uint) {
	t.Helper()

	mainItem := &model.MainItem{
		TeamID:     teamID,
		Code:       "MI-001",
		Title:      "Test Main Item",
		Priority:   "P1",
		ProposerID: userID,
		Status:     "待开始",
	}
	require.NoError(t, db.Create(mainItem).Error)

	sub1 := &model.SubItem{
		TeamID:     teamID,
		MainItemID: mainItem.ID,
		Title:      "Sub Item 1",
		Priority:   "P2",
		Status:     "待开始",
		Weight:     1.0,
	}
	require.NoError(t, db.Create(sub1).Error)

	sub2 := &model.SubItem{
		TeamID:     teamID,
		MainItemID: mainItem.ID,
		Title:      "Sub Item 2",
		Priority:   "P2",
		Status:     "待开始",
		Weight:     1.0,
	}
	require.NoError(t, db.Create(sub2).Error)

	return mainItem.ID, sub1.ID, sub2.ID
}

// appendProgress sends a progress append request via the router.
func appendProgress(t *testing.T, r *gin.Engine, token string, teamID, subID uint, completion float64) *httptest.ResponseRecorder {
	t.Helper()

	body := fmt.Sprintf(`{"completion":%.0f,"achievement":"some progress"}`, completion)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%d/progress", teamID, subID),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	return w
}

// getMainItem fetches a MainItem directly from the DB to verify completion.
func getMainItem(t *testing.T, db *gorm.DB, id uint) *model.MainItem {
	t.Helper()
	var item model.MainItem
	require.NoError(t, db.First(&item, id).Error)
	return &item
}

// getSubItem fetches a SubItem directly from the DB to verify completion.
func getSubItem(t *testing.T, db *gorm.DB, id uint) *model.SubItem {
	t.Helper()
	var item model.SubItem
	require.NoError(t, db.First(&item, id).Error)
	return &item
}

func TestProgress_AppendToSubItem1_UpdatesMainItemCompletion(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	mainID, sub1ID, _ := seedProgressData(t, db, data.teamAID, data.userAID)

	// Append progress (completion=60) to SubItem1
	w := appendProgress(t, r, token, data.teamAID, sub1ID, 60)
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

	mainID, sub1ID, sub2ID := seedProgressData(t, db, data.teamAID, data.userAID)

	// First append to SubItem1 (completion=60)
	w := appendProgress(t, r, token, data.teamAID, sub1ID, 60)
	require.Equal(t, http.StatusCreated, w.Code)

	// Then append to SubItem2 (completion=80)
	w = appendProgress(t, r, token, data.teamAID, sub2ID, 80)
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

	mainID, sub1ID, _ := seedProgressData(t, db, data.teamAID, data.userAID)

	// First append completion=60 to SubItem1
	w := appendProgress(t, r, token, data.teamAID, sub1ID, 60)
	require.Equal(t, http.StatusCreated, w.Code)

	// Capture MainItem completion after first append
	mainAfterFirst := getMainItem(t, db, mainID)
	completionBefore := mainAfterFirst.Completion

	// Try to append completion=50 (regression from 60) — should fail
	w = appendProgress(t, r, token, data.teamAID, sub1ID, 50)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "PROGRESS_REGRESSION", resp["code"])

	// Verify MainItem.Completion unchanged
	mainAfterRegression := getMainItem(t, db, mainID)
	assert.Equal(t, completionBefore, mainAfterRegression.Completion)
}

// ========== ItemPool Assign Transaction Tests ==========

// seedPoolData creates a pool item and a main item for assign tests.
func seedPoolData(t *testing.T, db *gorm.DB, teamID, userID uint) (poolID, mainItemID uint) {
	t.Helper()

	poolItem := &model.ItemPool{
		TeamID:      teamID,
		Title:       "Pool Item Title",
		Background:  "Some background",
		SubmitterID: userID,
		Status:      "待分配",
	}
	require.NoError(t, db.Create(poolItem).Error)

	mainItem := &model.MainItem{
		TeamID:     teamID,
		Code:       "MI-POOL-001",
		Title:      "Main Item for Pool",
		Priority:   "P1",
		ProposerID: userID,
		Status:     "待开始",
	}
	require.NoError(t, db.Create(mainItem).Error)

	return poolItem.ID, mainItem.ID
}

func TestItemPool_Assign_Success(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	poolID, mainItemID := seedPoolData(t, db, data.teamAID, data.userAID)

	// Assign the pool item
	body := fmt.Sprintf(`{"mainItemId":%d,"assigneeId":%d,"priority":"P2","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`, mainItemID, data.userAID)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/assign", data.teamAID, poolID),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify ItemPool.Status = "已分配"
	var pool model.ItemPool
	require.NoError(t, db.First(&pool, poolID).Error)
	assert.Equal(t, "已分配", pool.Status)
	assert.NotNil(t, pool.AssignedSubID)

	// Verify a new SubItem exists under the main item
	var subItems []model.SubItem
	require.NoError(t, db.Where("main_item_id = ?", mainItemID).Find(&subItems).Error)
	require.Len(t, subItems, 1)
	assert.Equal(t, "Pool Item Title", subItems[0].Title)
	assert.Equal(t, data.userAID, *subItems[0].AssigneeID)
}

func TestItemPool_Assign_Rollback_OnInvalidMainItem(t *testing.T) {
	db, data := setupTestDB(t)
	r, _ := setupTestRouterWithDB(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	poolID, _ := seedPoolData(t, db, data.teamAID, data.userAID)

	// Assign with a non-existent mainItemID to trigger failure
	invalidMainID := uint(99999)
	body := fmt.Sprintf(`{"mainItemId":%d,"assigneeId":%d,"priority":"P2","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`, invalidMainID, data.userAID)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/assign", data.teamAID, poolID),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// Should fail with 404 (item not found)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify ItemPool.Status remains "待分配"
	var pool model.ItemPool
	require.NoError(t, db.First(&pool, poolID).Error)
	assert.Equal(t, "待分配", pool.Status)

	// Verify no SubItem was created
	var count int64
	db.Model(&model.SubItem{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

// ========== Weekly Report Integration Tests ==========

// seedReportData creates a MainItem with a SubItem that has progress during the given week.
func seedReportData(t *testing.T, db *gorm.DB, teamID, userID uint, weekStart time.Time) (mainItemTitle string) {
	t.Helper()

	mainItem := &model.MainItem{
		TeamID:     teamID,
		Code:       "MI-RPT-001",
		Title:      "Report Test Main Item",
		Priority:   "P1",
		ProposerID: userID,
		Status:     "进行中",
	}
	require.NoError(t, db.Create(mainItem).Error)

	subItem := &model.SubItem{
		TeamID:     teamID,
		MainItemID: mainItem.ID,
		Title:      "Report Test Sub Item",
		Priority:   "P2",
		Status:     "进行中",
		Completion: 50,
		Weight:     1.0,
	}
	require.NoError(t, db.Create(subItem).Error)

	// Create a progress record within the week
	record := &model.ProgressRecord{
		SubItemID:   subItem.ID,
		TeamID:      teamID,
		AuthorID:    userID,
		Completion:  50,
		Achievement: "Completed half the work",
		CreatedAt:   weekStart.Add(24 * time.Hour), // Tuesday of the week
	}
	require.NoError(t, db.Create(record).Error)

	// Update MainItem completion to match
	require.NoError(t, db.Model(mainItem).Update("completion", 50).Error)

	return mainItem.Title
}

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
			data.teamAID, weekStart.Format("2006-01-02")), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/markdown", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, mainItemTitle)
	assert.Contains(t, body, "Completed half the work")
}

// ========== Helper: setupTestRouterWithDB ==========
// This is needed because the existing setupTestRouter creates its own DB internally,
// but we need to share the DB instance so we can seed data and then verify via the same DB.

func setupTestRouterWithDB(t *testing.T, db *gorm.DB, data *seedData) (*gin.Engine, *seedData) {
	t.Helper()

	userRepo := gormrepo.NewGormUserRepo(db)
	teamRepo := gormrepo.NewGormTeamRepo(db)
	mainItemRepo := gormrepo.NewGormMainItemRepo(db)
	subItemRepo := gormrepo.NewGormSubItemRepo(db)
	progressRepo := gormrepo.NewGormProgressRepo(db)
	itemPoolRepo := gormrepo.NewGormItemPoolRepo(db)

	authSvc := service.NewAuthService(userRepo, testJWTSecret)
	statusHistoryRepo := gormrepo.NewGormStatusHistoryRepo(db)
	statusHistorySvc := service.NewStatusHistoryService(statusHistoryRepo)
	mainItemSvc := service.NewMainItemService(mainItemRepo, subItemRepo, statusHistorySvc)
	subItemSvc := service.NewSubItemService(subItemRepo, mainItemSvc, statusHistorySvc)
	progressSvc := service.NewProgressService(progressRepo, subItemRepo, mainItemSvc)
	itemPoolSvc := service.NewItemPoolService(itemPoolRepo, subItemRepo, mainItemRepo, poolTransactor{db: db})
	teamSvc := service.NewTeamService(teamRepo, userRepo, mainItemRepo, teamTransactor{db: db})
	adminSvc := service.NewAdminService(userRepo, teamRepo)
	viewSvc := service.NewViewService(mainItemRepo, subItemRepo, progressRepo)
	reportSvc := service.NewReportService(mainItemRepo, subItemRepo, progressRepo)

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: testJWTSecret,
		},
		CORS: config.CORSConfig{
			Origins: []string{"http://localhost:3000"},
		},
		Server: config.ServerConfig{
			GinMode: "test",
		},
	}

	deps := &handler.Dependencies{
		Config:   cfg,
		TeamRepo: teamRepo,
		UserRepo: userRepo,
		RoleRepo: gormrepo.NewGormRoleRepo(db),
		Auth:     handler.NewAuthHandler(authSvc),
		Team:     handler.NewTeamHandler(teamSvc, userRepo),
		MainItem: handler.NewMainItemHandler(mainItemSvc, userRepo, subItemRepo),
		SubItem:  handler.NewSubItemHandler(subItemSvc),
		Progress: handler.NewProgressHandlerWithDeps(progressSvc, userRepo),
		ItemPool: handler.NewItemPoolHandler(itemPoolSvc, userRepo, mainItemRepo),
		View:     handler.NewViewHandlerWithDeps(viewSvc),
		Report:   handler.NewReportHandlerWithDeps(reportSvc),
		Admin:    handler.NewAdminHandler(adminSvc),
	}

	r := handler.SetupRouter(deps)
	return r, data
}
