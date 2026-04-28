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

// ---------- Tests ----------

func TestTransferPM_SuperAdminNotTeamMember_Succeeds(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)

	adminToken := loginAs(t, r, "superadmin", "adminPass")

	var memberA model.User
	require.NoError(t, db.First(&memberA, data.memberAID).Error)

	body := fmt.Sprintf(`{"newPmUserKey":"%d"}`, memberA.BizKey)
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/pm", data.teamABizKey), body, adminToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var team model.Team
	require.NoError(t, db.First(&team, data.teamAID).Error)
	assert.Equal(t, int64(data.memberAID), team.PmKey)
}

func TestTransferPM_PMTransfersToMember(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)

	pmToken := loginAs(t, r, "userA", "passwordA")

	var memberA model.User
	require.NoError(t, db.First(&memberA, data.memberAID).Error)

	body := fmt.Sprintf(`{"newPmUserKey":"%d"}`, memberA.BizKey)
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/pm", data.teamABizKey), body, pmToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var team model.Team
	require.NoError(t, db.First(&team, data.teamAID).Error)
	assert.Equal(t, int64(data.memberAID), team.PmKey)
}

func TestTransferPM_RegularMemberWithoutPermission_Returns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)

	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	var userA model.User
	require.NoError(t, db.First(&userA, data.userAID).Error)

	body := fmt.Sprintf(`{"newPmUserKey":"%d"}`, userA.BizKey)
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/pm", data.teamABizKey), body, memberToken)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ERR_FORBIDDEN", resp["code"])
}

func TestTransferPM_SuperAdminTargetNotMember_Returns403(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)

	adminToken := loginAs(t, r, "superadmin", "adminPass")

	// userB is PM of Team B, NOT a member of Team A
	var userB model.User
	require.NoError(t, db.First(&userB, data.userBID).Error)

	body := fmt.Sprintf(`{"newPmUserKey":"%d"}`, userB.BizKey)
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/pm", data.teamABizKey), body, adminToken)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_TEAM_MEMBER", resp["code"])
}

func TestTransferPM_UserNotFound_Returns404(t *testing.T) {
	db, data := setupRBACTestDB(t)
	backfillUserBizKeys(t, db)
	r := setupRBACTestRouter(t, db, data)

	adminToken := loginAs(t, r, "superadmin", "adminPass")

	body := `{"newPmUserKey":"9999999999"}`
	w := makeRequest(t, r, http.MethodPut,
		fmt.Sprintf("/api/v1/teams/%d/pm", data.teamABizKey), body, adminToken)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
