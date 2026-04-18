package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock TeamService for handler tests
// ---------------------------------------------------------------------------

type mockTeamService struct {
	createTeamResult struct {
		team *model.Team
		err  error
	}
	getTeamResult struct {
		team *model.Team
		err  error
	}
	getTeamDetailResult struct {
		detail *dto.TeamDetailResp
		err    error
	}
	listTeamsResult struct {
		teams []*model.Team
		err   error
	}
	updateTeamResult struct {
		team *model.Team
		err  error
	}
	inviteMemberErr error
	removeMemberErr error
	transferPMErr   error
	disbandTeamErr  error
	listMembersResult struct {
		members []*dto.TeamMemberDTO
		err     error
	}

	// capture calls
	createCalled   bool
	lastCreateReq  dto.CreateTeamReq
	lastCreatorID  uint

	listCalled bool

	getCalled  bool
	lastTeamID uint

	getDetailCalled bool

	updateCalled  bool
	lastUpdateReq dto.UpdateTeamReq
	lastPmID      uint

	inviteCalled   bool
	lastInviteReq  dto.InviteMemberReq
	inviteTeamID   uint

	removeCalled   bool
	removeTeamID   uint
	removeTargetID uint

	transferCalled bool
	transferTeamID uint
	newPmID        uint

	disbandCalled  bool
	disbandTeamID  uint
	lastConfirmName string

	listMembersCalled bool
	listMembersTeamID uint
}

func (m *mockTeamService) CreateTeam(_ context.Context, creatorID uint, req dto.CreateTeamReq) (*model.Team, error) {
	m.createCalled = true
	m.lastCreatorID = creatorID
	m.lastCreateReq = req
	return m.createTeamResult.team, m.createTeamResult.err
}

func (m *mockTeamService) GetTeam(_ context.Context, teamID uint) (*model.Team, error) {
	m.getCalled = true
	m.lastTeamID = teamID
	return m.getTeamResult.team, m.getTeamResult.err
}

func (m *mockTeamService) GetTeamDetail(_ context.Context, teamID uint) (*dto.TeamDetailResp, error) {
	m.getDetailCalled = true
	m.lastTeamID = teamID
	return m.getTeamDetailResult.detail, m.getTeamDetailResult.err
}

func (m *mockTeamService) ListTeams(_ context.Context, callerID uint, isSuperAdmin bool) ([]*model.Team, error) {
	m.listCalled = true
	return m.listTeamsResult.teams, m.listTeamsResult.err
}

func (m *mockTeamService) UpdateTeam(_ context.Context, pmID, teamID uint, req dto.UpdateTeamReq) (*model.Team, error) {
	m.updateCalled = true
	m.lastPmID = pmID
	m.lastTeamID = teamID
	m.lastUpdateReq = req
	return m.updateTeamResult.team, m.updateTeamResult.err
}

func (m *mockTeamService) InviteMember(_ context.Context, pmID, teamID uint, req dto.InviteMemberReq) error {
	m.inviteCalled = true
	m.lastPmID = pmID
	m.inviteTeamID = teamID
	m.lastInviteReq = req
	return m.inviteMemberErr
}

func (m *mockTeamService) RemoveMember(_ context.Context, pmID, teamID, targetUserID uint) error {
	m.removeCalled = true
	m.lastPmID = pmID
	m.removeTeamID = teamID
	m.removeTargetID = targetUserID
	return m.removeMemberErr
}

func (m *mockTeamService) TransferPM(_ context.Context, currentPMID, teamID, newPMID uint) error {
	m.transferCalled = true
	m.lastPmID = currentPMID
	m.transferTeamID = teamID
	m.newPmID = newPMID
	return m.transferPMErr
}

func (m *mockTeamService) DisbandTeam(_ context.Context, callerID uint, teamID uint, confirmName string) error {
	m.disbandCalled = true
	m.lastPmID = callerID
	m.disbandTeamID = teamID
	m.lastConfirmName = confirmName
	return m.disbandTeamErr
}

func (m *mockTeamService) UpdateMemberRole(_ context.Context, pmID, teamID, targetUserID uint, role string) error {
	return nil
}

func (m *mockTeamService) ListMembers(_ context.Context, teamID uint) ([]*dto.TeamMemberDTO, error) {
	m.listMembersCalled = true
	m.listMembersTeamID = teamID
	return m.listMembersResult.members, m.listMembersResult.err
}

// ---------------------------------------------------------------------------
// Mock UserRepo for CanCreateTeam check
// ---------------------------------------------------------------------------

type mockUserRepoForHandler struct {
	user *model.User
	err  error
}

func (m *mockUserRepoForHandler) FindByID(_ context.Context, _ uint) (*model.User, error) {
	return m.user, m.err
}

func (m *mockUserRepoForHandler) FindByUsername(_ context.Context, _ string) (*model.User, error) {
	return m.user, m.err
}

func (m *mockUserRepoForHandler) List(_ context.Context) ([]*model.User, error) {
	return nil, nil
}

func (m *mockUserRepoForHandler) Update(_ context.Context, _ *model.User) error {
	return nil
}

func (m *mockUserRepoForHandler) Create(_ context.Context, _ *model.User) error {
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// depsWithTeamSvc creates test deps with a mock TeamService wired in.
func depsWithTeamSvc(t *testing.T, svc *mockTeamService, userRepo *mockUserRepoForHandler) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.Team = NewTeamHandlerWithDeps(svc, userRepo)
	return deps
}

// ---------------------------------------------------------------------------
// Tests: POST /teams (Create)
// ---------------------------------------------------------------------------

func TestCreateTeam_Success(t *testing.T) {
	svc := &mockTeamService{}
	svc.createTeamResult.team = &model.Team{Name: "Alpha", Description: "desc", PmID: 1}
	svc.createTeamResult.team.ID = 10

	userRepo := &mockUserRepoForHandler{
		user: &model.User{CanCreateTeam: true},
	}
	userRepo.user.ID = 1

	deps := depsWithTeamSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"name":"Alpha","description":"desc"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Alpha", data["name"])
	assert.True(t, svc.createCalled)
	assert.Equal(t, uint(1), svc.lastCreatorID)
}

func TestCreateTeam_UserCannotCreateTeam(t *testing.T) {
	svc := &mockTeamService{}
	userRepo := &mockUserRepoForHandler{
		user: &model.User{CanCreateTeam: false},
	}
	userRepo.user.ID = 1

	deps := depsWithTeamSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"name":"Alpha"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.createCalled)
}

func TestCreateTeam_SuperAdminBypassesCanCreateTeam(t *testing.T) {
	svc := &mockTeamService{}
	svc.createTeamResult.team = &model.Team{Name: "Alpha", PmID: 1}
	svc.createTeamResult.team.ID = 10

	userRepo := &mockUserRepoForHandler{
		user: &model.User{CanCreateTeam: false},
	}
	userRepo.user.ID = 1

	deps := depsWithTeamSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "superadmin")
	body := `{"name":"Alpha"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, svc.createCalled)
}

func TestCreateTeam_InvalidBody(t *testing.T) {
	svc := &mockTeamService{}
	userRepo := &mockUserRepoForHandler{
		user: &model.User{CanCreateTeam: true},
	}
	userRepo.user.ID = 1

	deps := depsWithTeamSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams (List)
// ---------------------------------------------------------------------------

func TestListTeams_Success(t *testing.T) {
	svc := &mockTeamService{}
	svc.listTeamsResult.teams = []*model.Team{
		{Name: "Team A"},
		{Name: "Team B"},
	}
	svc.listTeamsResult.teams[0].ID = 1
	svc.listTeamsResult.teams[1].ID = 2

	deps := depsWithTeamSvc(t, svc, nil)
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
	assert.True(t, svc.listCalled)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId (Get)
// ---------------------------------------------------------------------------

func TestGetTeam_Success(t *testing.T) {
	svc := &mockTeamService{}
	svc.getTeamDetailResult.detail = &dto.TeamDetailResp{
		ID:            1,
		Name:          "Alpha",
		PmDisplayName: "Alice",
		MemberCount:   3,
		MainItemCount: 5,
	}

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Alpha", data["name"])
	assert.Equal(t, "Alice", data["pmDisplayName"])
	assert.Equal(t, float64(3), data["memberCount"])
	assert.Equal(t, float64(5), data["mainItemCount"])
}

func TestGetTeam_NotFound(t *testing.T) {
	svc := &mockTeamService{}
	svc.getTeamDetailResult.err = apperrors.ErrTeamNotFound

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId (Update)
// ---------------------------------------------------------------------------

func TestUpdateTeam_Success(t *testing.T) {
	svc := &mockTeamService{}
	svc.updateTeamResult.team = &model.Team{Name: "Updated", Description: "new desc", PmID: 1}
	svc.updateTeamResult.team.ID = 1

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"name":"Updated","description":"new desc"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.updateCalled)
	assert.Equal(t, "Updated", svc.lastUpdateReq.Name)
}

func TestUpdateTeam_NotPM(t *testing.T) {
	svc := &mockTeamService{}

	deps := depsWithTeamSvc(t, svc, nil)
	// mockTeamRepo returns a member with role "member" (not pm)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "member"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"name":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// TeamScopeMiddleware passes (member is in team), but RequireTeamRole("pm") fails
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DELETE /teams/:teamId (Disband)
// ---------------------------------------------------------------------------

func TestDisbandTeam_Success(t *testing.T) {
	svc := &mockTeamService{}

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"confirmName":"Alpha"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.disbandCalled)
	assert.Equal(t, "Alpha", svc.lastConfirmName)
}

func TestDisbandTeam_ConfirmNameMismatch(t *testing.T) {
	svc := &mockTeamService{}
	svc.disbandTeamErr = apperrors.ErrValidation

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"confirmName":"Wrong"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/members (ListMembers)
// ---------------------------------------------------------------------------

func TestListMembers_Success(t *testing.T) {
	svc := &mockTeamService{}
	svc.listMembersResult.members = []*dto.TeamMemberDTO{
		{UserID: 1, DisplayName: "Alice", Role: "pm"},
		{UserID: 2, DisplayName: "Bob", Role: "member"},
	}

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1/members", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
	assert.True(t, svc.listMembersCalled)
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/members (InviteMember)
// ---------------------------------------------------------------------------

func TestInviteMember_Success(t *testing.T) {
	svc := &mockTeamService{}

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"username":"bob","role":"member"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/1/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.inviteCalled)
	assert.Equal(t, "bob", svc.lastInviteReq.Username)
}

func TestInviteMember_AlreadyMember(t *testing.T) {
	svc := &mockTeamService{}
	svc.inviteMemberErr = apperrors.ErrAlreadyMember

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"username":"bob","role":"member"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/1/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ALREADY_MEMBER", resp["code"])
}

func TestInviteMember_UserNotFound(t *testing.T) {
	svc := &mockTeamService{}
	svc.inviteMemberErr = apperrors.ErrNotFound

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"username":"ghost","role":"member"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/1/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DELETE /teams/:teamId/members/:userId (RemoveMember)
// ---------------------------------------------------------------------------

func TestRemoveMember_Success(t *testing.T) {
	svc := &mockTeamService{}

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/1/members/5", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.removeCalled)
	assert.Equal(t, uint(5), svc.removeTargetID)
}

func TestRemoveMember_CannotRemoveSelf(t *testing.T) {
	svc := &mockTeamService{}
	svc.removeMemberErr = apperrors.ErrCannotRemoveSelf

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/1/members/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "CANNOT_REMOVE_SELF", resp["code"])
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/pm (TransferPM)
// ---------------------------------------------------------------------------

func TestTransferPM_Success(t *testing.T) {
	svc := &mockTeamService{}

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"newPmUserId":5}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/1/pm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.transferCalled)
	assert.Equal(t, uint(5), svc.newPmID)
}

func TestTransferPM_TargetNotMember(t *testing.T) {
	svc := &mockTeamService{}
	svc.transferPMErr = apperrors.ErrNotTeamMember

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"newPmUserId":99}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/1/pm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "NOT_TEAM_MEMBER", resp["code"])
}

func TestTransferPM_InvalidBody(t *testing.T) {
	svc := &mockTeamService{}

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/1/pm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransferPM_ServiceError(t *testing.T) {
	svc := &mockTeamService{}
	svc.transferPMErr = errors.New("unexpected error")

	deps := depsWithTeamSvc(t, svc, nil)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	body := `{"newPmUserId":5}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/1/pm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
