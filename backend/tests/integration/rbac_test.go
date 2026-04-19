package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/handler"
	"pm-work-tracker/backend/internal/migration"
	"pm-work-tracker/backend/internal/model"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
	"pm-work-tracker/backend/internal/service"
)

// ========== RBAC Migration Tests ==========

func TestRBACMigration_RunsSuccessfully(t *testing.T) {
	db := createFreshDB(t)

	err := migration.MigrateToRBAC(db)
	require.NoError(t, err)

	// Verify roles table exists with preset roles
	var roles []model.Role
	require.NoError(t, db.Find(&roles).Error)
	assert.Len(t, roles, 3) // superadmin, pm, member

	// Verify role_permissions has entries
	var count int64
	require.NoError(t, db.Model(&model.RolePermission{}).Count(&count).Error)
	assert.Greater(t, count, int64(0))
}

func TestRBACMigration_IsIdempotent(t *testing.T) {
	db := createFreshDB(t)

	// Run migration twice
	err := migration.MigrateToRBAC(db)
	require.NoError(t, err)

	var rolesAfterFirst []model.Role
	require.NoError(t, db.Find(&rolesAfterFirst).Error)
	firstCount := len(rolesAfterFirst)

	var permsAfterFirst int64
	require.NoError(t, db.Model(&model.RolePermission{}).Count(&permsAfterFirst).Error)

	// Run again
	err = migration.MigrateToRBAC(db)
	require.NoError(t, err)

	var rolesAfterSecond []model.Role
	require.NoError(t, db.Find(&rolesAfterSecond).Error)
	assert.Equal(t, firstCount, len(rolesAfterSecond))

	var permsAfterSecond int64
	require.NoError(t, db.Model(&model.RolePermission{}).Count(&permsAfterSecond).Error)
	assert.Equal(t, permsAfterFirst, permsAfterSecond)
}

func TestRBACMigration_PresetRoleCodesAreCorrect(t *testing.T) {
	db := createFreshDB(t)
	require.NoError(t, migration.MigrateToRBAC(db))

	// Verify that preset role codes match the expected set
	err := migration.VerifyPresetRoleCodes(db)
	assert.NoError(t, err)
}

// ========== Role CRUD Full Flow Tests ==========

func TestRoleCRUD_FullFlow(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// CREATE
	body := `{"name":"custom-role","description":"A custom role","permission_codes":["team:read","main_item:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", body, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	createData := createResp["data"].(map[string]interface{})
	roleID := uint(createData["id"].(float64))
	assert.Equal(t, "custom-role", createData["name"])
	assert.Equal(t, false, createData["is_preset"])

	// READ (list)
	w = makeRequest(t, r, http.MethodGet, "/api/v1/admin/roles", "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
	var listResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	items := listResp["data"].(map[string]interface{})["items"].([]interface{})
	assert.GreaterOrEqual(t, len(items), 4) // 3 preset + 1 custom

	// READ (single)
	w = makeRequest(t, r, http.MethodGet, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &getResp))
	getData := getResp["data"].(map[string]interface{})
	assert.Equal(t, "custom-role", getData["name"])
	perms := getData["permissions"].([]interface{})
	assert.Len(t, perms, 2)

	// UPDATE
	updateBody := `{"description":"Updated description","permission_codes":["team:read","main_item:read","main_item:create"]}`
	w = makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), updateBody, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
	var updateResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updateResp))
	updateData := updateResp["data"].(map[string]interface{})
	assert.Equal(t, "Updated description", updateData["description"])
	updatedPerms := updateData["permissions"].([]interface{})
	assert.Len(t, updatedPerms, 3)

	// DELETE
	w = makeRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deleted
	w = makeRequest(t, r, http.MethodGet, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), "", adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Preset Role Protection Tests ==========

func TestPresetRole_SuperadminCannotBeUpdated(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// Try to update superadmin role (id=1)
	body := `{"description":"hacked"}`
	w := makeRequest(t, r, http.MethodPut, "/api/v1/admin/roles/1", body, adminToken)
	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ERR_PRESET_ROLE_IMMUTABLE", resp["code"])
}

func TestPresetRole_SuperadminCannotBeDeleted(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	w := makeRequest(t, r, http.MethodDelete, "/api/v1/admin/roles/1", "", adminToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPresetRole_PMNameLocked(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// Find PM role id
	roleID := findRoleIDByName(t, db, "pm")

	// Try to rename PM role
	body := fmt.Sprintf(`{"name":"renamed-pm"}`)
	w := makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), body, adminToken)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// But description change should succeed
	body = fmt.Sprintf(`{"description":"new PM desc"}`)
	w = makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), body, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPresetRole_MemberNameLocked(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	roleID := findRoleIDByName(t, db, "member")

	body := fmt.Sprintf(`{"name":"renamed-member"}`)
	w := makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), body, adminToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Permission-Denied Scenarios ==========

func TestPermissionDenied_MemberCannotAccessPMEndpoints(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	// Member should NOT have team:invite permission
	// POST /api/v1/teams/:teamId/members (invite) requires "team:invite"
	body := `{"username":"userB"}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/members", data.teamAID), body, memberToken)
	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ERR_FORBIDDEN", resp["code"])
}

func TestPermissionDenied_MemberCannotInvite(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	body := `{"username":"userB"}`
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/members", data.teamAID), body, memberToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPermissionDenied_MemberCannotArchiveMainItem(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/999/archive", data.teamAID), "", memberToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== SuperAdmin Bypass Tests ==========

func TestSuperAdmin_BypassesTeamLevelPermissions(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// SuperAdmin should be able to access team-scoped routes even though they
	// are not a member of the team
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamAID), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// SuperAdmin can also list members
	w = makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/members", data.teamAID), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSuperAdmin_CanManageRoles(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// Create a role
	body := `{"name":"test-role","description":"test","permission_codes":["team:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", body, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// List permission codes
	w = makeRequest(t, r, http.MethodGet, "/api/v1/admin/permissions", "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Cross-Team Isolation Tests ==========

func TestCrossTeamIsolation_PMInTeamACannotUsePMPermissionsInTeamB(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	userAToken := loginAs(t, r, "userA", "passwordA") // PM of Team A

	// UserA is PM of TeamA, should be able to read main items in TeamA
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamAID), "", userAToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// But UserA should NOT be able to access TeamB (not a member)
	w = makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamBID), "", userAToken)
	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_TEAM_MEMBER", resp["code"])
}

// ========== Role Edit Immediate Effect Tests ==========

func TestRoleEdit_ImmediateEffectOnNextRequest(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	// Member should be able to read main items (member has main_item:read)
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamAID), "", memberToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// Create a new custom role with NO team:read or main_item:read
	customRoleBody := `{"name":"minimal-role","description":"minimal","permission_codes":["sub_item:read"]}`
	w = makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", customRoleBody, adminToken)
	require.Equal(t, http.StatusOK, w.Code)
	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	customRoleID := uint(createResp["data"].(map[string]interface{})["id"].(float64))

	// Update memberA's role in teamA to the new custom role
	var member model.TeamMember
	require.NoError(t, db.Where("team_id = ? AND user_id = ?", data.teamAID, data.memberAID).First(&member).Error)
	member.RoleID = &customRoleID
	require.NoError(t, db.Save(&member).Error)

	// Now member should NOT be able to read main items (no main_item:read permission)
	w = makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamAID), "", memberToken)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// But sub_item:read should still work
	w = makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/999", data.teamAID), "", memberToken)
	// 404 is expected (sub-item doesn't exist), not 403 (permission denied)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Delete Role with Users Tests ==========

func TestDeleteRole_WithUsers_Rejected(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// PM role (id=pmRoleID) is in use by members, but it's preset so cannot be deleted.
	// Create a custom role and assign it to a member.
	customRoleBody := `{"name":"deleteme-role","description":"to be deleted","permission_codes":["team:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", customRoleBody, adminToken)
	require.Equal(t, http.StatusOK, w.Code)
	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	customRoleID := uint(createResp["data"].(map[string]interface{})["id"].(float64))

	// Assign custom role to memberA in teamA
	var member model.TeamMember
	require.NoError(t, db.Where("team_id = ? AND user_id = ?", data.teamAID, data.memberAID).First(&member).Error)
	member.RoleID = &customRoleID
	require.NoError(t, db.Save(&member).Error)

	// Try to delete the role — should be rejected because it's in use
	w = makeRequest(t, r, http.MethodDelete,
		fmt.Sprintf("/api/v1/admin/roles/%d", customRoleID), "", adminToken)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ERR_ROLE_IN_USE", resp["code"])
}

func TestDeleteRole_WithoutUsers_Succeeds(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// Create a custom role with no users
	body := `{"name":"unused-role","description":"no users","permission_codes":["team:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", body, adminToken)
	require.Equal(t, http.StatusOK, w.Code)
	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	roleID := uint(createResp["data"].(map[string]interface{})["id"].(float64))

	// Delete should succeed
	w = makeRequest(t, r, http.MethodDelete,
		fmt.Sprintf("/api/v1/admin/roles/%d", roleID), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Invite Member with RoleID Tests ==========

// TestInviteMember_WithRoleID_MemberHasCorrectPermissions tests that a user invited
// with a specific role_id gets the correct set of permissions.
// We directly insert the team member into the DB to bypass the invite endpoint's
// PM-only check, then verify the middleware chain enforces the role's permissions.
func TestInviteMember_WithRoleID_MemberHasCorrectPermissions(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)

	// Create a new user and add them to teamA with member role
	hash, err := bcrypt.GenerateFromPassword([]byte("newUserPass"), 4)
	require.NoError(t, err)
	newUser := &model.User{Username: "newuser", DisplayName: "New User", PasswordHash: string(hash)}
	require.NoError(t, db.Create(newUser).Error)

	memberRoleID := findRoleIDByName(t, db, "member")
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: data.teamAID, UserID: newUser.ID, Role: "member",
		RoleID: &memberRoleID, JoinedAt: time.Now(),
	}).Error)

	// Login as newuser and verify member-level permissions
	newUserToken := loginAs(t, r, "newuser", "newUserPass")

	// Should be able to read main items (member has main_item:read)
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamAID), "", newUserToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// Should NOT be able to invite others (member does NOT have team:invite)
	w = makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/members", data.teamAID),
		`{"username":"userB"}`, newUserToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== User Permissions Endpoint Tests ==========

func TestUserPermissions_ReturnsCorrectPermissions(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	userAToken := loginAs(t, r, "userA", "passwordA")

	// GET /api/v1/me/permissions
	w := makeRequest(t, r, http.MethodGet, "/api/v1/me/permissions", "", userAToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	permData := resp["data"].(map[string]interface{})
	assert.Equal(t, false, permData["is_superadmin"])
	teamPerms := permData["team_permissions"].(map[string]interface{})
	assert.NotEmpty(t, teamPerms)
}

func TestUserPermissions_SuperAdmin(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/me/permissions", "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	permData := resp["data"].(map[string]interface{})
	assert.Equal(t, true, permData["is_superadmin"])
}

// ========== Permission Codes Registry Tests ==========

func TestPermissionCodes_RegistryEndpoint(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/admin/permissions", "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	groups := resp["data"].([]interface{})
	assert.GreaterOrEqual(t, len(groups), 7) // 7+ resource groups
}

// ========== Helper Functions ==========

// createFreshDB creates a fresh in-memory SQLite database for migration tests.
func createFreshDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	// Create tables that the migration expects to exist
	err = db.AutoMigrate(
		&model.User{}, &model.Team{}, &model.TeamMember{},
		&model.MainItem{}, &model.SubItem{},
		&model.ProgressRecord{}, &model.ItemPool{},
	)
	require.NoError(t, err)
	return db
}

// setupRBACTestDB creates an in-memory DB with RBAC tables seeded via the migration path.
func setupRBACTestDB(t *testing.T) (*gorm.DB, *seedData) {
	t.Helper()

	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	// Create all tables (including Role, RolePermission for RBAC)
	err = db.AutoMigrate(
		&model.User{}, &model.Team{}, &model.TeamMember{},
		&model.MainItem{}, &model.SubItem{},
		&model.ProgressRecord{}, &model.ItemPool{},
		&model.Role{}, &model.RolePermission{},
	)
	require.NoError(t, err)

	// Seed users
	hashA, err := bcrypt.GenerateFromPassword([]byte("passwordA"), 4)
	require.NoError(t, err)
	hashB, err := bcrypt.GenerateFromPassword([]byte("passwordB"), 4)
	require.NoError(t, err)
	hashMemberA, err := bcrypt.GenerateFromPassword([]byte("passwordMemberA"), 4)
	require.NoError(t, err)
	hashAdmin, err := bcrypt.GenerateFromPassword([]byte("adminPass"), 4)
	require.NoError(t, err)

	userA := &model.User{Username: "userA", DisplayName: "User A", PasswordHash: string(hashA)}
	userB := &model.User{Username: "userB", DisplayName: "User B", PasswordHash: string(hashB)}
	memberA := &model.User{Username: "memberA", DisplayName: "Member A", PasswordHash: string(hashMemberA)}
	superAdmin := &model.User{
		Username: "superadmin", DisplayName: "Super Admin",
		PasswordHash: string(hashAdmin), IsSuperAdmin: true,
	}

	require.NoError(t, db.Create(userA).Error)
	require.NoError(t, db.Create(userB).Error)
	require.NoError(t, db.Create(memberA).Error)
	require.NoError(t, db.Create(superAdmin).Error)

	// Seed roles
	pmRole := model.Role{Name: "superadmin", Description: "系统超级管理员", IsPreset: true}
	require.NoError(t, db.Create(&pmRole).Error)
	superadminRoleID := pmRole.ID

	pmRole = model.Role{Name: "pm", Description: "Project Manager", IsPreset: true}
	require.NoError(t, db.Create(&pmRole).Error)
	pmRoleID := pmRole.ID

	memberRole := model.Role{Name: "member", Description: "Team Member", IsPreset: true}
	require.NoError(t, db.Create(&memberRole).Error)
	memberRoleID := memberRole.ID

	// PM permissions (matching migration)
	pmPermCodes := []string{
		"team:create", "team:read", "team:update", "team:delete",
		"team:invite", "team:remove", "team:transfer",
		"main_item:create", "main_item:read", "main_item:update", "main_item:archive",
		"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:assign", "sub_item:change_status",
		"progress:create", "progress:read", "progress:update",
		"item_pool:submit", "item_pool:review",
		"view:weekly", "view:gantt", "view:table",
		"report:export",
		"user:read",
	}
	for _, code := range pmPermCodes {
		require.NoError(t, db.Create(&model.RolePermission{RoleID: pmRoleID, PermissionCode: code}).Error)
	}

	// Member permissions (matching migration)
	memberPermCodes := []string{
		"main_item:read",
		"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:change_status",
		"progress:create", "progress:read",
		"item_pool:submit",
		"view:weekly", "view:table",
		"report:export",
	}
	for _, code := range memberPermCodes {
		require.NoError(t, db.Create(&model.RolePermission{RoleID: memberRoleID, PermissionCode: code}).Error)
	}

	// Superadmin has no permission codes (bypasses all checks)
	_ = superadminRoleID

	// Seed teams
	teamA := &model.Team{Name: "Team A", PmID: userA.ID}
	teamB := &model.Team{Name: "Team B", PmID: userB.ID}
	require.NoError(t, db.Create(teamA).Error)
	require.NoError(t, db.Create(teamB).Error)

	// Seed team members
	now := time.Now()
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: teamA.ID, UserID: userA.ID, Role: "pm", RoleID: &pmRoleID, JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: teamA.ID, UserID: memberA.ID, Role: "member", RoleID: &memberRoleID, JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: teamB.ID, UserID: userB.ID, Role: "pm", RoleID: &pmRoleID, JoinedAt: now,
	}).Error)

	return db, &seedData{
		userAID:      userA.ID,
		userBID:      userB.ID,
		memberAID:    memberA.ID,
		superAdminID: superAdmin.ID,
		teamAID:      teamA.ID,
		teamBID:      teamB.ID,
	}
}

// setupRBACTestRouter wires the full router with RBAC-aware services.
func setupRBACTestRouter(t *testing.T, db *gorm.DB, data *seedData) *gin.Engine {
	t.Helper()

	userRepo := gormrepo.NewGormUserRepo(db)
	teamRepo := gormrepo.NewGormTeamRepo(db)
	mainItemRepo := gormrepo.NewGormMainItemRepo(db)
	subItemRepo := gormrepo.NewGormSubItemRepo(db)
	progressRepo := gormrepo.NewGormProgressRepo(db)
	itemPoolRepo := gormrepo.NewGormItemPoolRepo(db)
	roleRepo := gormrepo.NewGormRoleRepo(db)

	authSvc := service.NewAuthService(userRepo, testJWTSecret)
	mainItemSvc := service.NewMainItemService(mainItemRepo, subItemRepo)
	subItemSvc := service.NewSubItemService(subItemRepo, mainItemSvc)
	progressSvc := service.NewProgressService(progressRepo, subItemRepo, mainItemSvc)
	itemPoolSvc := service.NewItemPoolService(itemPoolRepo, subItemRepo, mainItemRepo, poolTransactor{db: db})
	teamSvc := service.NewTeamService(teamRepo, userRepo, mainItemRepo, teamTransactor{db: db})
	adminSvc := service.NewAdminService(userRepo, teamRepo)
	viewSvc := service.NewViewService(mainItemRepo, subItemRepo, progressRepo)
	reportSvc := service.NewReportService(mainItemRepo, subItemRepo, progressRepo)
	roleSvc := service.NewRoleService(roleRepo, userRepo)

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
		Config:     cfg,
		TeamRepo:   teamRepo,
		UserRepo:   userRepo,
		RoleRepo:   roleRepo,
		Auth:       handler.NewAuthHandler(authSvc),
		Team:       handler.NewTeamHandlerWithDeps(teamSvc, userRepo),
		MainItem:   handler.NewMainItemHandlerWithDeps(mainItemSvc, userRepo, subItemRepo),
		SubItem:    handler.NewSubItemHandlerWithDeps(subItemSvc),
		Progress:   handler.NewProgressHandlerWithDeps(progressSvc, userRepo),
		ItemPool:   handler.NewItemPoolHandlerWithDeps(itemPoolSvc, userRepo, nil),
		View:       handler.NewViewHandlerWithDeps(viewSvc),
		Report:     handler.NewReportHandlerWithDeps(reportSvc),
		Admin:      handler.NewAdminHandlerWithDeps(adminSvc),
		Role:       handler.NewRoleHandlerWithDeps(roleSvc),
		Permission: handler.NewPermissionHandlerWithDeps(roleSvc),
	}

	return handler.SetupRouter(deps)
}

// makeRequest is a helper to make an authenticated HTTP request.
func makeRequest(t *testing.T, r *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()

	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	var req *http.Request
	if bodyReader != nil {
		req = httptest.NewRequest(method, path, bodyReader)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// findRoleIDByName looks up a role ID by name from the database.
func findRoleIDByName(t *testing.T, db *gorm.DB, name string) uint {
	t.Helper()
	var role model.Role
	require.NoError(t, db.Where("name = ?", name).First(&role).Error)
	return role.ID
}
