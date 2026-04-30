package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/repository"
)

// mockTeamRepo is a testify mock for TeamRepo.
type mockTeamRepo struct {
	mock.Mock
}

func (m *mockTeamRepo) Create(ctx context.Context, team *model.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *mockTeamRepo) FindByID(ctx context.Context, teamID uint) (*model.Team, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Team), args.Error(1)
}

func (m *mockTeamRepo) List(ctx context.Context) ([]*model.Team, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Team), args.Error(1)
}

func (m *mockTeamRepo) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.Team, int64, error) {
	return nil, 0, nil
}

func (m *mockTeamRepo) Update(ctx context.Context, team *model.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *mockTeamRepo) SoftDelete(ctx context.Context, teamID uint) error {
	args := m.Called(ctx, teamID)
	return args.Error(0)
}

func (m *mockTeamRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.Team, error) {
	args := m.Called(ctx, bizKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Team), args.Error(1)
}

func (m *mockTeamRepo) AddMember(ctx context.Context, member *model.TeamMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *mockTeamRepo) RemoveMember(ctx context.Context, teamBizKey, userBizKey int64) error {
	args := m.Called(ctx, teamBizKey, userBizKey)
	return args.Error(0)
}

func (m *mockTeamRepo) FindMember(ctx context.Context, teamBizKey, userBizKey int64) (*model.TeamMember, error) {
	args := m.Called(ctx, teamBizKey, userBizKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TeamMember), args.Error(1)
}

func (m *mockTeamRepo) ListMembers(ctx context.Context, teamBizKey int64) ([]*dto.TeamMemberDTO, error) {
	args := m.Called(ctx, teamBizKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.TeamMemberDTO), args.Error(1)
}

func (m *mockTeamRepo) CountMembers(ctx context.Context, teamBizKey int64) (int64, error) {
	args := m.Called(ctx, teamBizKey)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTeamRepo) UpdateMember(ctx context.Context, member *model.TeamMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *mockTeamRepo) ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.AdminTeamDTO), args.Error(1)
}


func (m *mockTeamRepo) FindPMMembers(_ context.Context, _ []int64) (map[int64]string, error) {
	return map[int64]string{}, nil
}

func (m *mockTeamRepo) FindTeamsByUserIDs(ctx context.Context, userIDs []uint) (map[uint][]dto.TeamSummary, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uint][]dto.TeamSummary), args.Error(1)
}

func (m *mockTeamRepo) FindTeamsByUserBizKeys(ctx context.Context, userBizKeys []int64) (map[int64][]dto.TeamSummary, error) {
	args := m.Called(ctx, userBizKeys)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64][]dto.TeamSummary), args.Error(1)
}

// compile-time check that mockTeamRepo satisfies TeamRepo
var _ repository.TeamRepo = (*mockTeamRepo)(nil)

// mockRoleRepo is a testify mock for RoleRepo.
type mockRoleRepo struct {
	mock.Mock
}

func (m *mockRoleRepo) List(ctx context.Context, search string) ([]model.Role, error) {
	args := m.Called(ctx, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *mockRoleRepo) FindByID(ctx context.Context, id uint) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *mockRoleRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.Role, error) {
	args := m.Called(ctx, bizKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *mockRoleRepo) FindByName(ctx context.Context, name string) (*model.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *mockRoleRepo) Create(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *mockRoleRepo) Update(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *mockRoleRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRoleRepo) ListPermissions(ctx context.Context, roleKey int64) ([]string, error) {
	args := m.Called(ctx, roleKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockRoleRepo) SetPermissions(ctx context.Context, roleKey int64, codes []string) error {
	args := m.Called(ctx, roleKey, codes)
	return args.Error(0)
}

func (m *mockRoleRepo) CountMembersByRoleKey(ctx context.Context, roleKey int64) (int64, error) {
	args := m.Called(ctx, roleKey)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRoleRepo) HasPermission(ctx context.Context, userBizKey int64, code string) (bool, error) {
	args := m.Called(ctx, userBizKey, code)
	return args.Bool(0), args.Error(1)
}

func (m *mockRoleRepo) GetUserTeamPermissions(ctx context.Context, userBizKey int64) (map[int64][]string, error) {
	args := m.Called(ctx, userBizKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64][]string), args.Error(1)
}

var _ repository.RoleRepo = (*mockRoleRepo)(nil)

// capturedTeamContext captures teamBizKey and permCodes from the Gin context.
type capturedTeamContext struct {
	teamBizKey int64
	permCodes  []string
}

// setupTeamScopeRouter creates a test router with a simulated AuthMiddleware + TeamScopeMiddleware
// and a dummy handler that captures teamID, callerTeamRole, and permCodes.
func setupTeamScopeRouter(teamRepo repository.TeamRepo, roleRepo repository.RoleRepo) (*gin.Engine, *capturedTeamContext) {
	r := gin.New()
	cc := &capturedTeamContext{}

	// Simulate AuthMiddleware by setting userID, isSuperAdmin from query params
	r.Use(func(c *gin.Context) {
		uid := c.Query("userID")
		isAdmin := c.Query("isSuperAdmin")
		if uid != "" {
			var id uint
			fmt.Sscanf(uid, "%d", &id)
			c.Set("userID", id)
			c.Set("userBizKey", int64(id))
		}
		if isAdmin == "true" {
			c.Set("isSuperAdmin", true)
		}
		c.Next()
	})

	r.Use(TeamScopeMiddleware(teamRepo, roleRepo))

	r.GET("/api/v1/teams/:teamId/items", func(c *gin.Context) {
		cc.teamBizKey = GetTeamBizKey(c)
		cc.permCodes = GetPermCodes(c)
		c.Status(http.StatusOK)
	})

	return r, cc
}

func TestTeamScopeMiddleware_InvalidTeamID(t *testing.T) {
	teamRepo := new(mockTeamRepo)
	roleRepo := new(mockRoleRepo)
	r, _ := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/abc/items?userID=1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestTeamScopeMiddleware_NonMember_Returns403(t *testing.T) {
	teamRepo := new(mockTeamRepo)
	roleRepo := new(mockRoleRepo)
	teamRepo.On("FindByBizKey", mock.Anything, int64(1)).Return(&model.Team{BaseModel: model.BaseModel{ID: 1, BizKey: 1}}, nil)
	teamRepo.On("FindMember", mock.Anything, int64(1), int64(2)).Return(nil, fmt.Errorf("not found"))
	r, _ := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1/items?userID=2", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_TEAM_MEMBER")
}

func TestTeamScopeMiddleware_Member_SetsContext(t *testing.T) {
	roleBizKey := int64(3003)
	teamRepo := new(mockTeamRepo)
	roleRepo := new(mockRoleRepo)
	teamRepo.On("FindByBizKey", mock.Anything, int64(5)).Return(&model.Team{BaseModel: model.BaseModel{ID: 5, BizKey: 5}}, nil)
	teamRepo.On("FindMember", mock.Anything, int64(5), int64(10)).Return(&model.TeamMember{
		TeamKey: int64(5),
		UserKey: 10,
		RoleKey: &roleBizKey,
	}, nil)
	roleRepo.On("FindByBizKey", mock.Anything, roleBizKey).Return(&model.Role{BaseModel: model.BaseModel{ID: 3, BizKey: roleBizKey}}, nil)
	roleRepo.On("ListPermissions", mock.Anything, roleBizKey).Return([]string{"team:update", "team:invite"}, nil)
	r, cc := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/5/items?userID=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(5), cc.teamBizKey)
	assert.Equal(t, []string{"team:update", "team:invite"}, cc.permCodes)
}

func TestTeamScopeMiddleware_SuperAdmin_BypassesMembership(t *testing.T) {
	teamRepo := new(mockTeamRepo)
	roleRepo := new(mockRoleRepo)
	teamRepo.On("FindByBizKey", mock.Anything, int64(99)).Return(&model.Team{BaseModel: model.BaseModel{ID: 99}}, nil)
	r, cc := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/99/items?userID=1&isSuperAdmin=true", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(99), cc.teamBizKey)
	assert.Equal(t, []string{}, cc.permCodes)
	teamRepo.AssertNotCalled(t, "FindMember", mock.Anything, mock.Anything, mock.Anything)
}

func TestTeamScopeMiddleware_MemberNoRoleID_SetsEmptyPermCodes(t *testing.T) {
	teamRepo := new(mockTeamRepo)
	roleRepo := new(mockRoleRepo)
	teamRepo.On("FindByBizKey", mock.Anything, int64(3)).Return(&model.Team{BaseModel: model.BaseModel{ID: 3, BizKey: 3}}, nil)
	teamRepo.On("FindMember", mock.Anything, int64(3), int64(7)).Return(&model.TeamMember{
		TeamKey: int64(3),
		UserKey: 7,
		RoleKey: nil,
	}, nil)
	r, cc := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/3/items?userID=7", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(3), cc.teamBizKey)
	assert.Nil(t, cc.permCodes)
}

func TestTeamScopeMiddleware_RoleBizKeyNotFound_Returns500(t *testing.T) {
	teamRepo := new(mockTeamRepo)
	roleRepo := new(mockRoleRepo)
	teamRepo.On("FindByBizKey", mock.Anything, int64(4)).Return(&model.Team{BaseModel: model.BaseModel{ID: 4, BizKey: 4}}, nil)
	teamRepo.On("FindMember", mock.Anything, int64(4), int64(8)).Return(&model.TeamMember{
		TeamKey: int64(4),
		UserKey: 8,
		RoleKey: func() *int64 { v := int64(99); return &v }(),
	}, nil)
	roleRepo.On("FindByBizKey", mock.Anything, int64(99)).Return(nil, fmt.Errorf("not found"))
	r, _ := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/4/items?userID=8", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetTeamBizKey_NoValue(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, int64(0), GetTeamBizKey(c))
}
