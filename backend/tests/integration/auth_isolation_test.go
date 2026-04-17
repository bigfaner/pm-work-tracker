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
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/handler"
	"pm-work-tracker/backend/internal/model"
	appjwt "pm-work-tracker/backend/internal/pkg/jwt"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
	"pm-work-tracker/backend/internal/service"
)

const testJWTSecret = "test-secret-that-is-at-least-32-bytes!!"

// seedData holds IDs created during test setup.
type seedData struct {
	userAID      uint
	userBID      uint
	memberAID    uint // regular member of teamA (non-PM)
	superAdminID uint
	teamAID      uint
	teamBID      uint
}

// setupTestDB creates an in-memory SQLite DB, runs migrations, and seeds test data.
// Each call gets a unique database to avoid cross-test state leakage.
func setupTestDB(t *testing.T) (*gorm.DB, *seedData) {
	t.Helper()

	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(
		&model.User{}, &model.Team{}, &model.TeamMember{},
		&model.MainItem{}, &model.SubItem{},
		&model.ProgressRecord{}, &model.ItemPool{},
	)
	require.NoError(t, err)

	// Seed users with bcrypt-hashed passwords at cost 4 (fast for tests)
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

	// Seed teams
	teamA := &model.Team{Name: "Team A", PmID: userA.ID}
	teamB := &model.Team{Name: "Team B", PmID: userB.ID}
	require.NoError(t, db.Create(teamA).Error)
	require.NoError(t, db.Create(teamB).Error)

	// Seed team members
	now := time.Now()
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: teamA.ID, UserID: userA.ID, Role: "pm", JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: teamA.ID, UserID: memberA.ID, Role: "member", JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamID: teamB.ID, UserID: userB.ID, Role: "pm", JoinedAt: now,
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

// setupTestRouter creates an in-memory DB, wires the full router with real services,
// and returns the gin.Engine + seed data.
func setupTestRouter(t *testing.T) (*gin.Engine, *seedData) {
	t.Helper()

	db, data := setupTestDB(t)

	userRepo := gormrepo.NewGormUserRepo(db)
	teamRepo := gormrepo.NewGormTeamRepo(db)
	mainItemRepo := gormrepo.NewGormMainItemRepo(db)
	subItemRepo := gormrepo.NewGormSubItemRepo(db)
	progressRepo := gormrepo.NewGormProgressRepo(db)
	itemPoolRepo := gormrepo.NewGormItemPoolRepo(db)

	authSvc := service.NewAuthService(userRepo, testJWTSecret)
	mainItemSvc := service.NewMainItemService(mainItemRepo, subItemRepo)
	subItemSvc := service.NewSubItemService(subItemRepo, mainItemSvc)
	progressSvc := service.NewProgressService(progressRepo, subItemRepo, mainItemSvc)
	itemPoolSvc := service.NewItemPoolService(itemPoolRepo, subItemRepo, mainItemRepo, poolTransactor{db: db})
	teamSvc := service.NewTeamService(teamRepo, userRepo, mainItemRepo, teamTransactor{db: db})
	adminSvc := service.NewAdminService(userRepo, teamRepo)
	viewSvc := service.NewViewService(mainItemRepo, subItemRepo, progressRepo)
	reportSvc := service.NewReportService(mainItemRepo, subItemRepo, progressRepo)

	cfg := &config.Config{
		JWTSecret:   testJWTSecret,
		CORSOrigins: []string{"http://localhost:3000"},
		GinMode:     "test",
	}

	deps := &handler.Dependencies{
		Config:   cfg,
		TeamRepo: teamRepo,
		Auth:     handler.NewAuthHandler(authSvc),
		Team:     handler.NewTeamHandlerWithDeps(teamSvc, userRepo),
		MainItem: handler.NewMainItemHandlerWithDeps(mainItemSvc, userRepo, subItemRepo),
		SubItem:  handler.NewSubItemHandlerWithDeps(subItemSvc),
		Progress: handler.NewProgressHandlerWithDeps(progressSvc, userRepo),
		ItemPool: handler.NewItemPoolHandlerWithDeps(itemPoolSvc, userRepo),
		View:     handler.NewViewHandlerWithDeps(viewSvc),
		Report:   handler.NewReportHandlerWithDeps(reportSvc),
		Admin:    handler.NewAdminHandlerWithDeps(adminSvc),
	}

	r := handler.SetupRouter(deps)
	return r, data
}

// ========== Auth Flow Tests ==========

func TestAuthFlow_LoginWithCorrectCredentials_Returns200(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := `{"username":"userA","password":"passwordA"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, data["token"])

	user, ok := data["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "userA", user["username"])
}

func TestAuthFlow_LoginWithWrongPassword_Returns401(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := `{"username":"userA","password":"wrongpassword"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "UNAUTHORIZED", resp["code"])
}

func TestAuthFlow_TokenOnProtectedRoute_Returns200(t *testing.T) {
	r, data := setupTestRouter(t)

	// Login to get a real token
	token := loginAs(t, r, "userA", "passwordA")

	// Use the token on a protected route: GET /api/v1/teams/:teamAId/main-items
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamAID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// Should be 200 (empty list) not 401
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthFlow_ExpiredToken_Returns401(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Sign a token that is already expired
	claims := &appjwt.Claims{
		UserID: 999,
		Role:   "member",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := signTokenWithClaims(t, claims)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1/main-items", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthFlow_NoToken_Returns401(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1/main-items", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ========== Team Isolation Tests ==========

func TestTeamIsolation_UserACannotAccessTeamB_Returns403(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "userA", "passwordA")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamBID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_TEAM_MEMBER", resp["code"])
}

func TestTeamIsolation_UserACanAccessTeamA_Returns200(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "userA", "passwordA")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamAID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeamIsolation_UserBCannotAccessTeamA_Returns403(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "userB", "passwordB")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamAID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_TEAM_MEMBER", resp["code"])
}

// ========== SuperAdmin Bypass Tests ==========

func TestSuperAdmin_CanAccessTeamA(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "superadmin", "adminPass")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamAID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSuperAdmin_CanAccessTeamB(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "superadmin", "adminPass")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamBID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSuperAdmin_CanAccessAdminRoutes(t *testing.T) {
	r, _ := setupTestRouter(t)

	token := loginAs(t, r, "superadmin", "adminPass")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegularUser_CannotAccessAdminRoutes(t *testing.T) {
	r, _ := setupTestRouter(t)

	token := loginAs(t, r, "userA", "passwordA")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Logout Test ==========

func TestLogout_Returns200(t *testing.T) {
	r, _ := setupTestRouter(t)

	token := loginAs(t, r, "userA", "passwordA")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["code"])
}

func TestLogout_WithoutAuth_Returns401(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ========== Helpers ==========

// loginAs performs a login request and returns the JWT token.
func loginAs(t *testing.T, r *gin.Engine, username, password string) string {
	t.Helper()

	body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	token, ok := data["token"].(string)
	require.True(t, ok)
	return token
}

// signTokenWithClaims signs a JWT with custom claims for testing.
func signTokenWithClaims(t *testing.T, claims *appjwt.Claims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return s
}

// teamTransactor wraps *gorm.DB to satisfy service.TransactionDB.
type teamTransactor struct{ db *gorm.DB }

func (t teamTransactor) Transaction(fc func(tx interface{}) error) error {
	return t.db.Transaction(func(db *gorm.DB) error {
		return fc(db)
	})
}

// poolTransactor wraps *gorm.DB to satisfy service.dbTransactor.
type poolTransactor struct{ db *gorm.DB }

func (p poolTransactor) Transaction(fc func(tx *gorm.DB) error) error {
	return p.db.Transaction(fc)
}
