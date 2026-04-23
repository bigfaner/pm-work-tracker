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

func (m *mockTeamRepo) Delete(ctx context.Context, teamID uint) error {
	args := m.Called(ctx, teamID)
	return args.Error(0)
}

func (m *mockTeamRepo) AddMember(ctx context.Context, member *model.TeamMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *mockTeamRepo) RemoveMember(ctx context.Context, teamID, userID uint) error {
	args := m.Called(ctx, teamID, userID)
	return args.Error(0)
}

func (m *mockTeamRepo) FindMember(ctx context.Context, teamID, userID uint) (*model.TeamMember, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TeamMember), args.Error(1)
}

func (m *mockTeamRepo) ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.TeamMemberDTO), args.Error(1)
}

func (m *mockTeamRepo) CountMembers(ctx context.Context, teamID uint) (int64, error) {
	args := m.Called(ctx, teamID)
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


func (m *mockTeamRepo) FindPMMembers(_ context.Context, _ []uint) (map[uint]string, error) {
	return map[uint]string{}, nil
}

func (m *mockTeamRepo) FindTeamsByUserIDs(ctx context.Context, userIDs []uint) (map[uint][]dto.TeamSummary, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uint][]dto.TeamSummary), args.Error(1)
}

// compile-time check that mockTeamRepo satisfies TeamRepo
var _ repository.TeamRepo = (*mockTeamRepo)(nil)

// mockRoleRepo is a testify mock for RoleRepo.
type mockRoleRepo struct {
	mock.Mock
}

func (m *mockRoleRepo) List(ctx context.Context) ([]model.Role, error) {
	args := m.Called(ctx)
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

func (m *mockRoleRepo) ListPermissions(ctx context.Context, roleID uint) ([]string, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockRoleRepo) SetPermissions(ctx context.Context, roleID uint, codes []string) error {
	args := m.Called(ctx, roleID, codes)
	return args.Error(0)
}

func (m *mockRoleRepo) CountMembersByRoleID(ctx context.Context, roleID uint) (int64, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRoleRepo) HasPermission(ctx context.Context, userID uint, code string) (bool, error) {
	args := m.Called(ctx, userID, code)
	return args.Bool(0), args.Error(1)
}

func (m *mockRoleRepo) GetUserTeamPermissions(ctx context.Context, userID uint) (map[uint][]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uint][]string), args.Error(1)
}

var _ repository.RoleRepo = (*mockRoleRepo)(nil)

// capturedTeamContext captures teamID and permCodes from the Gin context.
type capturedTeamContext struct {
	teamID    uint
	permCodes []string
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
		}
		if isAdmin == "true" {
			c.Set("isSuperAdmin", true)
		}
		c.Next()
	})

	r.Use(TeamScopeMiddleware(teamRepo, roleRepo))

	r.GET("/api/v1/teams/:teamId/items", func(c *gin.Context) {
		cc.teamID = GetTeamID(c)
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
	teamRepo.On("FindMember", mock.Anything, uint(1), uint(2)).Return(nil, fmt.Errorf("not found"))
	r, _ := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1/items?userID=2", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_TEAM_MEMBER")
}

func TestTeamScopeMiddleware_Member_SetsContext(t *testing.T) {
	roleID := uint(3)
	teamRepo := new(mockTeamRepo)
	roleRepo := new(mockRoleRepo)
	teamRepo.On("FindMember", mock.Anything, uint(5), uint(10)).Return(&model.TeamMember{
		TeamID: 5,
		UserID: 10,
		Role:   "pm",
		RoleID: &roleID,
	}, nil)
	roleRepo.On("ListPermissions", mock.Anything, uint(3)).Return([]string{"team:update", "team:invite"}, nil)
	r, cc := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/5/items?userID=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(5), cc.teamID)
	assert.Equal(t, []string{"team:update", "team:invite"}, cc.permCodes)
}

func TestTeamScopeMiddleware_SuperAdmin_BypassesMembership(t *testing.T) {
	teamRepo := new(mockTeamRepo)
	roleRepo := new(mockRoleRepo)
	r, cc := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/99/items?userID=1&isSuperAdmin=true", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(99), cc.teamID)
	assert.Equal(t, []string{}, cc.permCodes)
	teamRepo.AssertNotCalled(t, "FindMember", mock.Anything, mock.Anything, mock.Anything)
}

func TestTeamScopeMiddleware_MemberNoRoleID_SetsEmptyPermCodes(t *testing.T) {
	teamRepo := new(mockTeamRepo)
	roleRepo := new(mockRoleRepo)
	teamRepo.On("FindMember", mock.Anything, uint(3), uint(7)).Return(&model.TeamMember{
		TeamID: 3,
		UserID: 7,
		Role:   "member",
		RoleID: nil,
	}, nil)
	r, cc := setupTeamScopeRouter(teamRepo, roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/3/items?userID=7", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(3), cc.teamID)
	assert.Nil(t, cc.permCodes)
}

func TestGetTeamID_NoValue(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, uint(0), GetTeamID(c))
}
