package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

type permMatrixFixtures struct {
	mainItemBizKey       int64  // completed main item for archive
	progressBizKey       int64  // progress record for progress:update
	poolItemBizKey       int64  // pool item for item_pool:review
	assignMainItemBizKey int64  // separate main item for pool assign (avoids sub-item conflict with archive)
	weekStart            string // past Monday for weekly export
}

// seedPermMatrixFixtures creates the minimum DB fixtures needed for the permission matrix tests.
func seedPermMatrixFixtures(t *testing.T, db *gorm.DB, data *seedData) permMatrixFixtures {
	t.Helper()

	// Main item for archive — must be in completed status
	archiveItem := &model.MainItem{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:     data.teamABizKey,
		Code:        "TAMA-PM001",
		Title:       "Perm Matrix Archive Item",
		Priority:    "P1",
		ProposerKey: int64(data.userAID),
		ItemStatus:  "completed",
	}
	require.NoError(t, db.Create(archiveItem).Error)

	// Sub item + progress record for progress:update
	subItem := &model.SubItem{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:     data.teamABizKey,
		MainItemKey: int64(archiveItem.ID),
		Title:       "Perm Matrix Sub Item",
		Priority:    "P2",
		ItemStatus:  "pending",
		Weight:      1.0,
	}
	require.NoError(t, db.Create(subItem).Error)

	progressRecord := &model.ProgressRecord{
		BizKey:      snowflake.Generate(),
		SubItemKey:  subItem.BizKey,
		TeamKey:     data.teamABizKey,
		AuthorKey:   int64(data.userAID),
		Completion:  50,
		Achievement: "Initial progress",
		CreateTime:  time.Now(),
	}
	require.NoError(t, db.Create(progressRecord).Error)

	// Pool item for item_pool:review
	poolItem := &model.ItemPool{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:      data.teamABizKey,
		Title:        "Perm Matrix Pool Item",
		Background:   "Background info",
		SubmitterKey: int64(data.userAID),
		PoolStatus:   "pending",
	}
	require.NoError(t, db.Create(poolItem).Error)

	// Separate main item for pool assign (so archive item stays sub-item-free)
	assignItem := &model.MainItem{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:     data.teamABizKey,
		Code:        "TAMA-PM002",
		Title:       "Perm Matrix Assign Target",
		Priority:    "P1",
		ProposerKey: int64(data.userAID),
		ItemStatus:  "pending",
	}
	require.NoError(t, db.Create(assignItem).Error)

	// Seed report data for weekly export — use last week's Monday
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	lastMonday := now.AddDate(0, 0, -(weekday-1)-7)
	lastMonday = time.Date(lastMonday.Year(), lastMonday.Month(), lastMonday.Day(), 0, 0, 0, 0, time.UTC)
	seedReportData(t, db, data.teamABizKey, data.userAID, lastMonday)

	return permMatrixFixtures{
		mainItemBizKey:       archiveItem.BizKey,
		progressBizKey:       progressRecord.BizKey,
		poolItemBizKey:       poolItem.BizKey,
		assignMainItemBizKey: assignItem.BizKey,
		weekStart:            lastMonday.Format("2006-01-02"),
	}
}

// TestRBACPermMatrix_PresetRoles verifies the permission matrix for superadmin/pm/member
// across 5 representative endpoints (15 assertions total).
func TestRBACPermMatrix_PresetRoles(t *testing.T) {
	type roleCase struct {
		username   string
		password   string
		wantStatus int
	}

	roles := []roleCase{
		{"superadmin", "adminPass", http.StatusOK},
		{"userA", "passwordA", http.StatusOK},              // pm role
		{"memberA", "passwordMemberA", http.StatusForbidden},
	}

	for _, role := range roles {
		role := role
		t.Run(role.username, func(t *testing.T) {
			db, data := setupRBACTestDB(t)
			fixtures := seedPermMatrixFixtures(t, db, data)
			r := setupRBACTestRouter(t, db, data)
			token := loginAs(t, r, role.username, role.password)
			memberRoleBizKey := findRoleBizKeyByName(t, db, "member")

			// 1. GET /views/gantt — view:gantt (member lacks this permission)
			w := makeRequest(t, r, http.MethodGet,
				fmt.Sprintf("/api/v1/teams/%d/views/gantt", data.teamABizKey),
				"", token)
			assert.Equal(t, role.wantStatus, w.Code, "gantt view")

			// 2. POST /members — team:invite
			inviteBody := fmt.Sprintf(`{"username":"userB","roleKey":"%s"}`, memberRoleBizKey)
			w = makeRequest(t, r, http.MethodPost,
				fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey),
				inviteBody, token)
			assert.Equal(t, role.wantStatus, w.Code, "invite member")

			// 3. PATCH /progress/:recordId/completion — progress:update
			w = makeRequest(t, r, http.MethodPatch,
				fmt.Sprintf("/api/v1/teams/%d/progress/%d/completion", data.teamABizKey, fixtures.progressBizKey),
				`{"completion":75}`, token)
			assert.Equal(t, role.wantStatus, w.Code, "correct completion")

			// 4. POST /item-pool/:poolId/assign — item_pool:review
			assignBody := fmt.Sprintf(`{"mainItemKey":"%d","assigneeKey":"%d","priority":"P2","startDate":"2026-01-01","expectedEndDate":"2026-12-31"}`, fixtures.assignMainItemBizKey, data.userAID)
			w = makeRequest(t, r, http.MethodPost,
				fmt.Sprintf("/api/v1/teams/%d/item-pool/%d/assign", data.teamABizKey, fixtures.poolItemBizKey),
				assignBody, token)
			assert.Equal(t, role.wantStatus, w.Code, "assign pool item")

			// 5. POST /main-items/:itemId/archive — main_item:archive
			w = makeRequest(t, r, http.MethodPost,
				fmt.Sprintf("/api/v1/teams/%d/main-items/%d/archive", data.teamABizKey, fixtures.mainItemBizKey),
				"", token)
			assert.Equal(t, role.wantStatus, w.Code, "archive main item")
		})
	}
}

// TestCustomRole_PartialPermissions verifies that a custom role with partial permissions
// enforces access correctly, and that permission changes take effect immediately (no cache).
func TestCustomRole_PartialPermissions(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// Step 1: Create custom role with only main_item:read + progress:read
	var customRole model.Role
	customRole = model.Role{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		Name:        "custom-partial",
		Description: "Partial permissions role",
		IsPreset:    false,
	}
	require.NoError(t, db.Create(&customRole).Error)
	for _, code := range []string{"main_item:read", "progress:read"} {
		require.NoError(t, db.Create(&model.RolePermission{RoleKey: customRole.BizKey, PermissionCode: code}).Error)
	}

	// Step 2: Create user and assign custom role to teamA
	hash, err := bcrypt.GenerateFromPassword([]byte("customPass"), 4)
	require.NoError(t, err)
	customUser := &model.User{Username: "customuser", DisplayName: "Custom User", PasswordHash: string(hash)}
	require.NoError(t, db.Create(customUser).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey:  int64(data.teamAID),
		UserKey:  int64(customUser.ID),
		RoleKey:  &customRole.BizKey,
		JoinedAt: time.Now(),
	}).Error)

	token := loginAs(t, r, "customuser", "customPass")

	// Step 3: Verify partial permissions
	// GET /main-items → 200 (has main_item:read)
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", token)
	assert.Equal(t, http.StatusOK, w.Code, "read main items should be allowed")

	// POST /main-items → 403 (no main_item:create)
	createBody := `{"title":"New Item","priority":"P1","assigneeKey":"1","startDate":"2026-01-01","expectedEndDate":"2026-12-31","isKeyItem":false}`
	w = makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), createBody, token)
	assert.Equal(t, http.StatusForbidden, w.Code, "create main item should be denied")

	// POST /archive → 403 (no main_item:archive)
	w = makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/999999/archive", data.teamABizKey), "", token)
	assert.Equal(t, http.StatusForbidden, w.Code, "archive should be denied")

	// Step 4: Add main_item:create to the custom role via direct DB insert
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: customRole.BizKey, PermissionCode: "main_item:create"}).Error)

	// Step 5: Verify POST /main-items → 200 with same token (no re-login)
	_ = adminToken // not needed for permission update (used direct DB)
	w = makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), createBody, token)
	assert.Equal(t, http.StatusCreated, w.Code, "create main item should now be allowed after role update")
}

// TestPermBoundary_EmptyRole verifies that a user with a role that has no permissions
// is denied access to any protected endpoint.
func TestPermBoundary_EmptyRole(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)

	// Create role with zero permissions
	emptyRole := model.Role{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		Name:        "empty-role",
		Description: "No permissions",
		IsPreset:    false,
	}
	require.NoError(t, db.Create(&emptyRole).Error)

	hash, err := bcrypt.GenerateFromPassword([]byte("emptyPass"), 4)
	require.NoError(t, err)
	emptyUser := &model.User{Username: "emptyuser", DisplayName: "Empty User", PasswordHash: string(hash)}
	require.NoError(t, db.Create(emptyUser).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey:  int64(data.teamAID),
		UserKey:  int64(emptyUser.ID),
		RoleKey:  &emptyRole.BizKey,
		JoinedAt: time.Now(),
	}).Error)

	token := loginAs(t, r, "emptyuser", "emptyPass")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestPermBoundary_SuperAdminBypass verifies that superadmin bypasses all permission checks.
func TestPermBoundary_SuperAdminBypass(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestPermBoundary_InvalidToken401 verifies that an invalid JWT returns 401, not 403.
func TestPermBoundary_InvalidToken401(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", "invalid-token-string")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusForbidden, w.Code)
}
