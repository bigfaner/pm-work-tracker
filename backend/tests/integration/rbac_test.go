package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	"pm-work-tracker/backend/internal/pkg/dbutil"
	"pm-work-tracker/backend/internal/pkg/snowflake"
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
	body := `{"name":"custom-role","description":"A custom role","permissionCodes":["team:read","main_item:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", body, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	createData := createResp["data"].(map[string]interface{})
	roleBizKey := createData["bizKey"].(string)
	assert.Equal(t, "custom-role", createData["roleName"])
	assert.Equal(t, false, createData["isPreset"])

	// READ (list)
	w = makeRequest(t, r, http.MethodGet, "/api/v1/admin/roles", "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
	var listResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	items := listResp["data"].(map[string]interface{})["items"].([]interface{})
	assert.GreaterOrEqual(t, len(items), 4) // 3 preset + 1 custom

	// READ (single)
	w = makeRequest(t, r, http.MethodGet, fmt.Sprintf("/api/v1/admin/roles/%s", roleBizKey), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &getResp))
	getData := getResp["data"].(map[string]interface{})
	assert.Equal(t, "custom-role", getData["roleName"])
	perms := getData["permissions"].([]interface{})
	assert.Len(t, perms, 2)

	// UPDATE
	updateBody := `{"description":"Updated description","permissionCodes":["team:read","main_item:read","main_item:create"]}`
	w = makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%s", roleBizKey), updateBody, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
	var updateResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updateResp))
	updateData := updateResp["data"].(map[string]interface{})
	assert.Equal(t, "Updated description", updateData["roleDesc"])
	updatedPerms := updateData["permissions"].([]interface{})
	assert.Len(t, updatedPerms, 3)

	// DELETE
	w = makeRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/v1/admin/roles/%s", roleBizKey), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deleted
	w = makeRequest(t, r, http.MethodGet, fmt.Sprintf("/api/v1/admin/roles/%s", roleBizKey), "", adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Preset Role Protection Tests ==========

func TestPresetRole_SuperadminCannotBeUpdated(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// Try to update superadmin role
	superadminBizKey := findRoleBizKeyByName(t, db, "superadmin")
	body := `{"description":"hacked"}`
	w := makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%s", superadminBizKey), body, adminToken)
	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ERR_PRESET_ROLE_IMMUTABLE", resp["code"])
}

func TestPresetRole_SuperadminCannotBeDeleted(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	superadminBizKey := findRoleBizKeyByName(t, db, "superadmin")
	w := makeRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/v1/admin/roles/%s", superadminBizKey), "", adminToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPresetRole_PMNameLocked(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// Find PM role bizKey
	roleBizKey := findRoleBizKeyByName(t, db, "pm")

	// Try to rename PM role
	body := fmt.Sprintf(`{"name":"renamed-pm"}`)
	w := makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%s", roleBizKey), body, adminToken)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// But description change should succeed
	body = fmt.Sprintf(`{"description":"new PM desc"}`)
	w = makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%s", roleBizKey), body, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPresetRole_MemberNameLocked(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	roleBizKey := findRoleBizKeyByName(t, db, "member")

	body := fmt.Sprintf(`{"name":"renamed-member"}`)
	w := makeRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%s", roleBizKey), body, adminToken)
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
		fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey), body, memberToken)
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
		fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey), body, memberToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPermissionDenied_MemberCannotArchiveMainItem(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/999/archive", data.teamABizKey), "", memberToken)
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
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// SuperAdmin can also list members
	w = makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey), "", adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSuperAdmin_CanManageRoles(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// Create a role
	body := `{"name":"test-role","description":"test","permissionCodes":["team:read"]}`
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
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", userAToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// But UserA should NOT be able to access TeamB (not a member)
	w = makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamBBizKey), "", userAToken)
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
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", memberToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// Create a new custom role with NO team:read or main_item:read
	customRoleBody := `{"name":"minimal-role","description":"minimal","permissionCodes":["sub_item:read"]}`
	w = makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", customRoleBody, adminToken)
	require.Equal(t, http.StatusOK, w.Code)
	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	customRoleBizKey := createResp["data"].(map[string]interface{})["bizKey"].(string)

	// Look up the internal numeric ID for the DB update
	customRoleID := findRoleIDByBizKey(t, db, customRoleBizKey)

	// Update memberA's role in teamA to the new custom role
	var member model.TeamMember
	require.NoError(t, db.Where("team_key = ? AND user_key = ?", data.teamAID, data.memberAID).First(&member).Error)
	roleKey := int64(customRoleID); member.RoleKey = &roleKey
	require.NoError(t, db.Save(&member).Error)

	// Now member should NOT be able to read main items (no main_item:read permission)
	w = makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", memberToken)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// But sub_item:read should still work
	w = makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/999", data.teamABizKey), "", memberToken)
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
	customRoleBody := `{"name":"deleteme-role","description":"to be deleted","permissionCodes":["team:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", customRoleBody, adminToken)
	require.Equal(t, http.StatusOK, w.Code)
	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	customRoleBizKey := createResp["data"].(map[string]interface{})["bizKey"].(string)
	customRoleID := findRoleIDByBizKey(t, db, customRoleBizKey)

	// Assign custom role to memberA in teamA
	var member model.TeamMember
	require.NoError(t, db.Where("team_key = ? AND user_key = ?", data.teamAID, data.memberAID).First(&member).Error)
	roleKey := int64(customRoleID); member.RoleKey = &roleKey
	require.NoError(t, db.Save(&member).Error)

	// Try to delete the role — should be rejected because it's in use
	w = makeRequest(t, r, http.MethodDelete,
	fmt.Sprintf("/api/v1/admin/roles/%s", customRoleBizKey), "", adminToken)
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
	body := `{"name":"unused-role","description":"no users","permissionCodes":["team:read"]}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", body, adminToken)
	require.Equal(t, http.StatusOK, w.Code)
	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	roleBizKey := createResp["data"].(map[string]interface{})["bizKey"].(string)

	// Delete should succeed
	w = makeRequest(t, r, http.MethodDelete,
		fmt.Sprintf("/api/v1/admin/roles/%s", roleBizKey), "", adminToken)
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
		TeamKey: int64(data.teamAID), UserKey: int64(newUser.ID), 
		RoleKey: func() *int64 { v := int64(memberRoleID); return &v }(), JoinedAt: time.Now(),
	}).Error)

	// Login as newuser and verify member-level permissions
	newUserToken := loginAs(t, r, "newuser", "newUserPass")

	// Should be able to read main items (member has main_item:read)
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), "", newUserToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// Should NOT be able to invite others (member does NOT have team:invite)
	w = makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey),
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
	assert.Equal(t, false, permData["isSuperAdmin"])
	teamPerms := permData["teamPermissions"].(map[string]interface{})
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
	assert.Equal(t, true, permData["isSuperAdmin"])
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
	snowflake.Init(1)


	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	// Create all tables (including Role, RolePermission for RBAC)
	err = db.AutoMigrate(
		&model.User{}, &model.Team{}, &model.TeamMember{},
		&model.MainItem{}, &model.SubItem{},
		&model.ProgressRecord{}, &model.ItemPool{},
		&model.Role{}, &model.RolePermission{},
		&model.StatusHistory{},
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
	superadminRole := model.Role{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Name: "superadmin", Description: "系统超级管理员", IsPreset: true}
	require.NoError(t, db.Create(&superadminRole).Error)

	pmRole := model.Role{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Name: "pm", Description: "Project Manager", IsPreset: true}
	require.NoError(t, db.Create(&pmRole).Error)
	pmRoleID := pmRole.ID

	memberRole := model.Role{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Name: "member", Description: "Team Member", IsPreset: true}
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
	_ = superadminRole

	// Seed teams (with BizKey so middleware can resolve bizKey to internal ID)
	teamA := &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team A", PmKey: int64(userA.ID), Code: "TAMA"}
	teamB := &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team B", PmKey: int64(userB.ID), Code: "TAMB"}
	require.NoError(t, db.Create(teamA).Error)
	require.NoError(t, db.Create(teamB).Error)

	// Seed team members
	now := time.Now()
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey: int64(teamA.ID), UserKey: int64(userA.ID),  RoleKey: func() *int64 { v := int64(pmRoleID); return &v }(), JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey: int64(teamA.ID), UserKey: int64(memberA.ID),  RoleKey: func() *int64 { v := int64(memberRoleID); return &v }(), JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey: int64(teamB.ID), UserKey: int64(userB.ID),  RoleKey: func() *int64 { v := int64(pmRoleID); return &v }(), JoinedAt: now,
	}).Error)

	return db, &seedData{
		userAID:      userA.ID,
		userBID:      userB.ID,
		memberAID:    memberA.ID,
		superAdminID: superAdmin.ID,
		teamAID:      teamA.ID,
		teamBID:      teamB.ID,
		teamABizKey:  teamA.BizKey,
		teamBBizKey:  teamB.BizKey,
	}
}

// setupRBACTestRouter wires the full router with RBAC-aware services.
func setupRBACTestRouter(t *testing.T, db *gorm.DB, data *seedData) *gin.Engine {
	t.Helper()

	userRepo := gormrepo.NewGormUserRepo(db)
	teamRepo := gormrepo.NewGormTeamRepo(db)
	dialect := dbutil.NewDialect(db)
	mainItemRepo := gormrepo.NewGormMainItemRepo(db, dialect)
	subItemRepo := gormrepo.NewGormSubItemRepo(db, dialect)
	progressRepo := gormrepo.NewGormProgressRepo(db)
	itemPoolRepo := gormrepo.NewGormItemPoolRepo(db)
	roleRepo := gormrepo.NewGormRoleRepo(db)

	authSvc := service.NewAuthService(userRepo, testJWTSecret)
	statusHistoryRepo := gormrepo.NewGormStatusHistoryRepo(db)
	statusHistorySvc := service.NewStatusHistoryService(statusHistoryRepo)
	mainItemSvc := service.NewMainItemService(mainItemRepo, subItemRepo, statusHistorySvc)
	subItemSvc := service.NewSubItemService(subItemRepo, mainItemSvc, statusHistorySvc)
	progressSvc := service.NewProgressService(progressRepo, subItemRepo, mainItemSvc, statusHistorySvc)
	itemPoolSvc := service.NewItemPoolService(itemPoolRepo, subItemRepo, mainItemRepo, transactor{db: db})
	teamSvc := service.NewTeamService(teamRepo, userRepo, mainItemRepo, roleRepo, transactor{db: db})
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
			GinMode:  "test",
			BasePath: "/api",
		},
	}

	deps := &handler.Dependencies{
		Config:     cfg,
		TeamRepo:   teamRepo,
		UserRepo:   userRepo,
		RoleRepo:   roleRepo,
		Auth:       handler.NewAuthHandler(authSvc),
		Team:       handler.NewTeamHandler(teamSvc, userRepo),
		MainItem:   handler.NewMainItemHandler(mainItemSvc, userRepo, subItemRepo),
		SubItem:    handler.NewSubItemHandler(subItemSvc, mainItemSvc),
		Progress:   handler.NewProgressHandler(progressSvc, userRepo, subItemSvc),
		ItemPool:   handler.NewItemPoolHandler(itemPoolSvc, userRepo, mainItemRepo),
		View:       handler.NewViewHandler(viewSvc),
		Report:     handler.NewReportHandler(reportSvc),
		Admin:      handler.NewAdminHandler(adminSvc),
		Role:       handler.NewRoleHandler(roleSvc),
		Permission: handler.NewPermissionHandler(roleSvc),
	}

	return handler.SetupRouter(deps, nil)
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
	require.NoError(t, db.Where("role_name = ?", name).First(&role).Error)
	return role.ID
}

// findRoleBizKeyByName looks up a role's BizKey as string by name from the database.
func findRoleBizKeyByName(t *testing.T, db *gorm.DB, name string) string {
	t.Helper()
	var role model.Role
	require.NoError(t, db.Where("role_name = ?", name).First(&role).Error)
	return fmt.Sprintf("%d", role.BizKey)
}

// findRoleIDByBizKey looks up a role's numeric ID by its BizKey string.
func findRoleIDByBizKey(t *testing.T, db *gorm.DB, bizKey string) uint {
	t.Helper()
	var role model.Role
	bk, err := strconv.ParseInt(bizKey, 10, 64)
	require.NoError(t, err)
	require.NoError(t, db.Where("biz_key = ?", bk).First(&role).Error)
	return role.ID
}
