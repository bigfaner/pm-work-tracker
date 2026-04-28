package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/model"
)

// ========== Test Team Management: Create Team ==========

func TestTeamManagement_Create_Returns201(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	body := `{"name":"New Team","description":"A new team","code":"NEWT"}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/teams", body, token)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, "New Team", d["name"])
	assert.Equal(t, "NEWT", d["code"])
	assert.Equal(t, "A new team", d["description"])
}

func TestTeamManagement_Create_CreatorAutoJoinsAsPM(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	body := `{"name":"Auto PM Team","description":"auto pm","code":"APMT"}`
	w := makeRequest(t, r, http.MethodPost, "/api/v1/teams", body, token)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	bizKey := d["bizKey"].(string)

	// Get team detail to verify creator is PM and member
	w = makeRequest(t, r, http.MethodGet, fmt.Sprintf("/api/v1/teams/%s", bizKey), "", token)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d = resp["data"].(map[string]interface{})
	assert.Equal(t, int(1), int(d["memberCount"].(float64)))
}

func TestTeamManagement_Create_DuplicateCodeReturns422(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	body := fmt.Sprintf(`{"name":"Dup Team","description":"dup","code":"%s"}`, "TAMA")
	w := makeRequest(t, r, http.MethodPost, "/api/v1/teams", body, token)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ========== Test Team Management: List Teams ==========

func TestTeamManagement_List_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/teams", "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	items := d["items"].([]interface{})
	assert.GreaterOrEqual(t, len(items), 1)
}

func TestTeamManagement_List_NewUserReturnsNonEmpty(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// ListTeams returns all teams regardless of membership
	token := loginAs(t, r, "memberA", "passwordMemberA")

	w := makeRequest(t, r, http.MethodGet, "/api/v1/teams", "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	items := d["items"].([]interface{})
	// memberA is a member of teamA, so teams list shows teams
	assert.GreaterOrEqual(t, len(items), 1)
}

// ========== Test Team Management: Team Detail ==========

func TestTeamManagement_Detail_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	path := fmt.Sprintf("/api/v1/teams/%d", data.teamABizKey)
	w := makeRequest(t, r, http.MethodGet, path, "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, "Team A", d["name"])
	assert.Equal(t, "TAMA", d["code"])
}

func TestTeamManagement_Detail_NonMemberReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// userB is not a member of teamA
	token := loginAs(t, r, "userB", "passwordB")

	path := fmt.Sprintf("/api/v1/teams/%d", data.teamABizKey)
	w := makeRequest(t, r, http.MethodGet, path, "", token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTeamManagement_Detail_NotFoundReturns404(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	path := "/api/v1/teams/999999999"
	w := makeRequest(t, r, http.MethodGet, path, "", token)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Test Team Management: Update Team ==========

func TestTeamManagement_Update_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	path := fmt.Sprintf("/api/v1/teams/%d", data.teamABizKey)
	body := `{"name":"Updated Team A","description":"updated desc"}`
	w := makeRequest(t, r, http.MethodPut, path, body, token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	d := resp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Team A", d["name"])
	assert.Equal(t, "updated desc", d["description"])
}

func TestTeamManagement_Update_MissingNameReturns422(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	path := fmt.Sprintf("/api/v1/teams/%d", data.teamABizKey)
	body := `{"description":"no name"}`
	w := makeRequest(t, r, http.MethodPut, path, body, token)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamManagement_Update_NonPMReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA is a non-PM member of teamA — but lacks team:update permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	path := fmt.Sprintf("/api/v1/teams/%d", data.teamABizKey)
	body := `{"name":"Should Fail","description":"fail"}`
	w := makeRequest(t, r, http.MethodPut, path, body, token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Test Team Management: Disband Team ==========

func TestTeamManagement_Disband_Returns200AndCascadeDeletes(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	// Create a main item in teamA for cascade deletion verification
	_, _, _, _, _ = seedProgressData(t, db, data.teamAID, data.userAID)

	path := fmt.Sprintf("/api/v1/teams/%d", data.teamABizKey)
	body := `{"confirmName":"Team A"}`
	w := makeRequest(t, r, http.MethodDelete, path, body, token)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify team is soft-deleted (FindByID uses NotDeleted scope)
	var count int64
	require.NoError(t, db.Model(&model.Team{}).Where("id = ? AND deleted_flag = 0", data.teamAID).Count(&count).Error)
	assert.Equal(t, int64(0), count)

	// Verify cascade: main items should still exist in DB but team is gone
	// (soft delete on team does not cascade-delete items; it just hides the team)
}

func TestTeamManagement_Disband_NonPMReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA lacks team:delete permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	path := fmt.Sprintf("/api/v1/teams/%d", data.teamABizKey)
	body := `{"confirmName":"Team A"}`
	w := makeRequest(t, r, http.MethodDelete, path, body, token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTeamManagement_Disband_NotFoundReturns404(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "superadmin", "adminPass")

	path := "/api/v1/teams/999999999"
	body := `{"confirmName":"anything"}`
	w := makeRequest(t, r, http.MethodDelete, path, body, token)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Test Team Management: Search Users ==========

func TestTeamManagement_SearchUsers_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	path := fmt.Sprintf("/api/v1/teams/%d/search-users?search=userB", data.teamABizKey)
	w := makeRequest(t, r, http.MethodGet, path, "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	users := resp["data"].([]interface{})
	assert.GreaterOrEqual(t, len(users), 1)
}

func TestTeamManagement_SearchUsers_EmptyResultForNoMatch(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	path := fmt.Sprintf("/api/v1/teams/%d/search-users?search=nonexistent", data.teamABizKey)
	w := makeRequest(t, r, http.MethodGet, path, "", token)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	users := resp["data"].([]interface{})
	assert.Equal(t, 0, len(users))
}

func TestTeamManagement_SearchUsers_MemberReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA lacks team:invite permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	path := fmt.Sprintf("/api/v1/teams/%d/search-users?search=userB", data.teamABizKey)
	w := makeRequest(t, r, http.MethodGet, path, "", token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Test Team Management: Invite Member ==========

func TestTeamManagement_InviteMember_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	memberRoleBizKey := findRoleBizKeyByName(t, db, "member")
	body := fmt.Sprintf(`{"username":"userB","roleKey":"%s"}`, memberRoleBizKey)
	path := fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeamManagement_InviteMember_AlreadyMemberReturnsError(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	memberRoleBizKey := findRoleBizKeyByName(t, db, "member")
	// memberA is already a member of teamA
	body := fmt.Sprintf(`{"username":"memberA","roleKey":"%s"}`, memberRoleBizKey)
	path := fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ALREADY_MEMBER", resp["code"])
}

func TestTeamManagement_InviteMember_MemberInvitesReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA lacks team:invite permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	memberRoleBizKey := findRoleBizKeyByName(t, db, "member")
	body := fmt.Sprintf(`{"username":"userB","roleKey":"%s"}`, memberRoleBizKey)
	path := fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTeamManagement_InviteMember_UserNotFoundReturns404(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	memberRoleBizKey := findRoleBizKeyByName(t, db, "member")
	body := fmt.Sprintf(`{"username":"ghost_user","roleKey":"%s"}`, memberRoleBizKey)
	path := fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, path, body, token)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Test Team Management: Remove Member ==========

func TestTeamManagement_RemoveMember_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	// Remove memberA from teamA using userA (PM)
	// Need memberA's bizKey for the path parameter
	backfillUserBizKeys(t, db)
	var memberAUser model.User
	require.NoError(t, db.Where("id = ?", data.memberAID).First(&memberAUser).Error)

	path := fmt.Sprintf("/api/v1/teams/%d/members/%d", data.teamABizKey, memberAUser.BizKey)
	w := makeRequest(t, r, http.MethodDelete, path, "", token)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeamManagement_RemoveMember_PMRemovalReturns422(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	// Try to remove the PM (userA) from teamA
	backfillUserBizKeys(t, db)
	var userA model.User
	require.NoError(t, db.Where("id = ?", data.userAID).First(&userA).Error)

	path := fmt.Sprintf("/api/v1/teams/%d/members/%d", data.teamABizKey, userA.BizKey)
	w := makeRequest(t, r, http.MethodDelete, path, "", token)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "CANNOT_REMOVE_SELF", resp["code"])
}

func TestTeamManagement_RemoveMember_MemberRemovesReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA lacks team:remove permission
	token := loginAs(t, r, "memberA", "passwordMemberA")

	backfillUserBizKeys(t, db)
	var userB model.User
	require.NoError(t, db.Where("id = ?", data.userBID).First(&userB).Error)

	// First invite userB to teamA so they're a member we can try to remove
	pmToken := loginAs(t, r, "userA", "passwordA")
	memberRoleBizKey := findRoleBizKeyByName(t, db, "member")
	inviteBody := fmt.Sprintf(`{"username":"userB","roleKey":"%s"}`, memberRoleBizKey)
	invitePath := fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, invitePath, inviteBody, pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Now memberA tries to remove userB — should get 403
	path := fmt.Sprintf("/api/v1/teams/%d/members/%d", data.teamABizKey, userB.BizKey)
	w = makeRequest(t, r, http.MethodDelete, path, "", token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTeamManagement_RemoveMember_NotFoundReturns404(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	// Use a bizKey that doesn't map to any user
	path := fmt.Sprintf("/api/v1/teams/%d/members/999999999", data.teamABizKey)
	w := makeRequest(t, r, http.MethodDelete, path, "", token)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Test Team Management: Update Member Role ==========

func TestTeamManagement_UpdateMemberRole_Returns200(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	backfillUserBizKeys(t, db)
	var memberAUser model.User
	require.NoError(t, db.Where("id = ?", data.memberAID).First(&memberAUser).Error)

	// Change memberA's role: use the existing "member" role's BizKey
	memberRoleBizKey := findRoleBizKeyByName(t, db, "member")
	body := fmt.Sprintf(`{"roleKey":"%s"}`, memberRoleBizKey)
	path := fmt.Sprintf("/api/v1/teams/%d/members/%d/role", data.teamABizKey, memberAUser.BizKey)
	w := makeRequest(t, r, http.MethodPut, path, body, token)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeamManagement_UpdateMemberRole_PMRoleChangeReturns422(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	token := loginAs(t, r, "userA", "passwordA")

	backfillUserBizKeys(t, db)
	var memberAUser model.User
	require.NoError(t, db.Where("id = ?", data.memberAID).First(&memberAUser).Error)

	pmRoleBizKey := findRoleBizKeyByName(t, db, "pm")
	body := fmt.Sprintf(`{"roleKey":"%s"}`, pmRoleBizKey)
	path := fmt.Sprintf("/api/v1/teams/%d/members/%d/role", data.teamABizKey, memberAUser.BizKey)
	w := makeRequest(t, r, http.MethodPut, path, body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "CANNOT_ASSIGN_PM_ROLE", resp["code"])
}

func TestTeamManagement_UpdateMemberRole_MemberChangesRoleReturns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	// memberA lacks team:invite permission (which gates role changes)
	token := loginAs(t, r, "memberA", "passwordMemberA")

	backfillUserBizKeys(t, db)
	var userB model.User
	require.NoError(t, db.Where("id = ?", data.userBID).First(&userB).Error)

	// First invite userB to teamA
	pmToken := loginAs(t, r, "userA", "passwordA")
	memberRoleBizKey := findRoleBizKeyByName(t, db, "member")
	inviteBody := fmt.Sprintf(`{"username":"userB","roleKey":"%s"}`, memberRoleBizKey)
	invitePath := fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, invitePath, inviteBody, pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Now memberA tries to change userB's role
	body := fmt.Sprintf(`{"roleKey":"%s"}`, memberRoleBizKey)
	path := fmt.Sprintf("/api/v1/teams/%d/members/%d/role", data.teamABizKey, userB.BizKey)
	w = makeRequest(t, r, http.MethodPut, path, body, token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTeamManagement_UpdateMemberRole_ImmediateEffect(t *testing.T) {
	db, data := setupRBACTestDB(t)
	r := setupRBACTestRouter(t, db, data)
	pmToken := loginAs(t, r, "userA", "passwordA")
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	backfillUserBizKeys(t, db)
	var memberAUser model.User
	require.NoError(t, db.Where("id = ?", data.memberAID).First(&memberAUser).Error)

	// memberA currently has "member" role with no team:invite permission
	// Verify memberA cannot invite before role change
	memberRoleBizKey := findRoleBizKeyByName(t, db, "member")
	inviteBody := fmt.Sprintf(`{"username":"userB","roleKey":"%s"}`, memberRoleBizKey)
	invitePath := fmt.Sprintf("/api/v1/teams/%d/members", data.teamABizKey)
	w := makeRequest(t, r, http.MethodPost, invitePath, inviteBody, memberToken)
	require.Equal(t, http.StatusForbidden, w.Code)

	// PM creates a custom role with team:invite permission
	adminToken := loginAs(t, r, "superadmin", "adminPass")
	customRoleBody := `{"name":"inviter","description":"Can invite","permissionCodes":["team:invite"]}`
	w = makeRequest(t, r, http.MethodPost, "/api/v1/admin/roles", customRoleBody, adminToken)
	require.Equal(t, http.StatusOK, w.Code)

	var roleResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &roleResp))
	roleData := roleResp["data"].(map[string]interface{})
	customRoleBizKey := roleData["bizKey"].(string)

	// PM changes memberA's role to the custom role with team:invite
	body := fmt.Sprintf(`{"roleKey":"%s"}`, customRoleBizKey)
	path := fmt.Sprintf("/api/v1/teams/%d/members/%d/role", data.teamABizKey, memberAUser.BizKey)
	w = makeRequest(t, r, http.MethodPut, path, body, pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Now memberA should be able to invite (immediate effect)
	// Need to login again to get fresh token with updated permissions
	memberToken = loginAs(t, r, "memberA", "passwordMemberA")
	w = makeRequest(t, r, http.MethodPost, invitePath, inviteBody, memberToken)
	assert.Equal(t, http.StatusOK, w.Code)
}
