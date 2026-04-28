package integration

import (
	"database/sql"
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
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/handler"
	"pm-work-tracker/backend/internal/model"
	appjwt "pm-work-tracker/backend/internal/pkg/jwt"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/pkg/dbutil"
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
	teamABizKey  int64
	teamBBizKey  int64
}

// transactor wraps *gorm.DB to satisfy repo.DBTransactor.
type transactor struct{ db *gorm.DB }

func (t transactor) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return t.db.Transaction(fc, opts...)
}

// ========== Unified DI Wiring ==========

// wireHandlers creates all repos+services+handlers and returns *handler.Dependencies.
// Called by all setup* functions to eliminate duplicated DI wiring.
// includeRBAC adds Role and Permission handlers (needed for RBAC-aware routes).
func wireHandlers(t *testing.T, db *gorm.DB, data *seedData, includeRBAC bool) *handler.Dependencies {
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
		Config:   cfg,
		TeamRepo: teamRepo,
		UserRepo: userRepo,
		RoleRepo: roleRepo,
		Auth:     handler.NewAuthHandler(authSvc),
		Team:     handler.NewTeamHandler(teamSvc, userRepo),
		MainItem: handler.NewMainItemHandler(mainItemSvc, userRepo, subItemRepo),
		SubItem:  handler.NewSubItemHandler(subItemSvc, mainItemSvc),
		Progress: handler.NewProgressHandler(progressSvc, userRepo, subItemSvc),
		ItemPool: handler.NewItemPoolHandler(itemPoolSvc, userRepo, mainItemRepo),
		View:     handler.NewViewHandler(viewSvc),
		Report:   handler.NewReportHandler(reportSvc),
		Admin:    handler.NewAdminHandler(adminSvc),
	}

	if includeRBAC {
		roleSvc := service.NewRoleService(roleRepo, userRepo)
		deps.Role = handler.NewRoleHandler(roleSvc)
		deps.Permission = handler.NewPermissionHandler(roleSvc)
	}

	return deps
}

// ========== Setup Functions ==========

// setupTestDB creates an in-memory SQLite DB, runs migrations, and seeds test data.
// Each call gets a unique database to avoid cross-test state leakage.
func setupTestDB(t *testing.T) (*gorm.DB, *seedData) {
	t.Helper()

	snowflake.Init(1)

	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(
		&model.User{}, &model.Team{}, &model.TeamMember{},
		&model.MainItem{}, &model.SubItem{},
		&model.ProgressRecord{}, &model.ItemPool{},
		&model.Role{}, &model.RolePermission{},
		&model.StatusHistory{},
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

	userA := &model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Username: "userA", DisplayName: "User A", PasswordHash: string(hashA)}
	userB := &model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Username: "userB", DisplayName: "User B", PasswordHash: string(hashB)}
	memberA := &model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Username: "memberA", DisplayName: "Member A", PasswordHash: string(hashMemberA)}
	superAdmin := &model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		Username: "superadmin", DisplayName: "Super Admin",
		PasswordHash: string(hashAdmin), IsSuperAdmin: true,
	}

	require.NoError(t, db.Create(userA).Error)
	require.NoError(t, db.Create(userB).Error)
	require.NoError(t, db.Create(memberA).Error)
	require.NoError(t, db.Create(superAdmin).Error)

	// Seed roles and permissions for RBAC
	pmRole := model.Role{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Name: "pm", Description: "Project Manager", IsPreset: true}
	require.NoError(t, db.Create(&pmRole).Error)
	memberRole := model.Role{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Name: "member", Description: "Team Member", IsPreset: true}
	require.NoError(t, db.Create(&memberRole).Error)

	// PM gets all team-scoped permissions
	pmPermCodes := []string{
		"team:read", "team:update", "team:delete", "team:invite", "team:remove", "team:transfer",
		"main_item:create", "main_item:read", "main_item:update", "main_item:archive",
		"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:change_status", "sub_item:assign",
		"progress:create", "progress:read", "progress:update",
		"item_pool:submit", "item_pool:review",
		"view:weekly", "view:gantt", "view:table", "report:export",
	}
	for _, code := range pmPermCodes {
		require.NoError(t, db.Create(&model.RolePermission{RoleKey: pmRole.BizKey, PermissionCode: code}).Error)
	}
	// Member gets limited permissions
	memberPermCodes := []string{
		"main_item:read", "sub_item:create", "sub_item:read", "sub_item:update",
		"sub_item:change_status", "progress:create", "progress:read",
		"item_pool:submit", "view:weekly", "view:table", "report:export",
	}
	for _, code := range memberPermCodes {
		require.NoError(t, db.Create(&model.RolePermission{RoleKey: memberRole.BizKey, PermissionCode: code}).Error)
	}

	// Seed teams (with BizKey so middleware can resolve bizKey to internal ID)
	teamA := &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team A", PmKey: userA.BizKey, Code: "TAMA"}
	teamB := &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team B", PmKey: userB.BizKey, Code: "TAMB"}
	require.NoError(t, db.Create(teamA).Error)
	require.NoError(t, db.Create(teamB).Error)

	// Seed team members (with RoleID pointing to seeded roles)
	now := time.Now()
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey: teamA.BizKey, UserKey: userA.BizKey, RoleKey: &pmRole.BizKey, JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey: teamA.BizKey, UserKey: memberA.BizKey, RoleKey: &memberRole.BizKey, JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey: teamB.BizKey, UserKey: userB.BizKey, RoleKey: &pmRole.BizKey, JoinedAt: now,
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

// setupTestRouter creates an in-memory DB, wires the full router with real services,
// and returns the gin.Engine + seed data.
func setupTestRouter(t *testing.T) (*gin.Engine, *seedData) {
	t.Helper()

	db, data := setupTestDB(t)
	deps := wireHandlers(t, db, data, false)
	r := handler.SetupRouter(deps, nil)
	return r, data
}

// setupTestRouterWithDB creates a router that reuses an existing DB instance.
// Needed for tests that seed data and then verify via the same DB.
func setupTestRouterWithDB(t *testing.T, db *gorm.DB, data *seedData) (*gin.Engine, *seedData) {
	t.Helper()
	deps := wireHandlers(t, db, data, false)
	r := handler.SetupRouter(deps, nil)
	return r, data
}

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

	userA := &model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Username: "userA", DisplayName: "User A", PasswordHash: string(hashA)}
	userB := &model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Username: "userB", DisplayName: "User B", PasswordHash: string(hashB)}
	memberA := &model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Username: "memberA", DisplayName: "Member A", PasswordHash: string(hashMemberA)}
	superAdmin := &model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
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

	memberRole := model.Role{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Name: "member", Description: "Team Member", IsPreset: true}
	require.NoError(t, db.Create(&memberRole).Error)

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
		require.NoError(t, db.Create(&model.RolePermission{RoleKey: pmRole.BizKey, PermissionCode: code}).Error)
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
		require.NoError(t, db.Create(&model.RolePermission{RoleKey: memberRole.BizKey, PermissionCode: code}).Error)
	}

	// Superadmin has no permission codes (bypasses all checks)
	_ = superadminRole

	// Seed teams (with BizKey so middleware can resolve bizKey to internal ID)
	teamA := &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team A", PmKey: userA.BizKey, Code: "TAMA"}
	teamB := &model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team B", PmKey: userB.BizKey, Code: "TAMB"}
	require.NoError(t, db.Create(teamA).Error)
	require.NoError(t, db.Create(teamB).Error)

	// Seed team members
	now := time.Now()
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey: teamA.BizKey, UserKey: userA.BizKey, RoleKey: &pmRole.BizKey, JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey: teamA.BizKey, UserKey: memberA.BizKey, RoleKey: &memberRole.BizKey, JoinedAt: now,
	}).Error)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey: teamB.BizKey, UserKey: userB.BizKey, RoleKey: &pmRole.BizKey, JoinedAt: now,
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
	deps := wireHandlers(t, db, data, true)
	return handler.SetupRouter(deps, nil)
}

// ========== Auth Helpers ==========

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

// ========== HTTP Request Helpers ==========

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

// ========== Data Seeding Helpers ==========

// seedProgressData creates a MainItem with two SubItems (weight=1 each) for progress tests.
// Returns the main item ID, the two sub item IDs, and their bizKey values.
func seedProgressData(t *testing.T, db *gorm.DB, teamBizKey int64, userID uint) (mainItemID, subItem1ID, subItem2ID uint, subItem1BizKey, subItem2BizKey int64) {
	t.Helper()

	mainItem := &model.MainItem{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:      teamBizKey,
		Code:         "TAMA-00001",
		Title:        "Test Main Item",
		Priority:     "P1",
		ProposerKey:  int64(userID),
		ItemStatus:   "pending",
	}
	require.NoError(t, db.Create(mainItem).Error)

	sub1 := &model.SubItem{
		BaseModel:     model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:       teamBizKey,
		MainItemKey:   int64(mainItem.ID),
		Title:         "Sub Item 1",
		Priority:      "P2",
		ItemStatus:    "pending",
		Weight:        1.0,
	}
	require.NoError(t, db.Create(sub1).Error)

	sub2 := &model.SubItem{
		BaseModel:     model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:       teamBizKey,
		MainItemKey:   int64(mainItem.ID),
		Title:         "Sub Item 2",
		Priority:      "P2",
		ItemStatus:    "pending",
		Weight:        1.0,
	}
	require.NoError(t, db.Create(sub2).Error)

	return mainItem.ID, sub1.ID, sub2.ID, sub1.BizKey, sub2.BizKey
}

// appendProgress sends a progress append request via the router.
func appendProgress(t *testing.T, r *gin.Engine, token string, teamBizKey, subBizKey int64, completion float64) *httptest.ResponseRecorder {
	t.Helper()

	body := fmt.Sprintf(`{"completion":%.0f,"achievement":"some progress"}`, completion)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/sub-items/%d/progress", teamBizKey, subBizKey),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	return w
}

// seedPoolData creates a pool item and a main item for assign tests.
func seedPoolData(t *testing.T, db *gorm.DB, teamBizKey int64, userID uint) (poolID, mainItemID uint, poolBizKey, mainItemBizKey int64) {
	t.Helper()

	poolItem := &model.ItemPool{
		BaseModel:     model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:       teamBizKey,
		Title:         "Pool Item Title",
		Background:    "Some background",
		SubmitterKey:  int64(userID),
		PoolStatus:    "pending",
	}
	require.NoError(t, db.Create(poolItem).Error)

	mainItem := &model.MainItem{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:      teamBizKey,
		Code:         "TAMA-00002",
		Title:        "Main Item for Pool",
		Priority:     "P1",
		ProposerKey:  int64(userID),
		ItemStatus:   "pending",
	}
	require.NoError(t, db.Create(mainItem).Error)

	return poolItem.ID, mainItem.ID, poolItem.BizKey, mainItem.BizKey
}

// seedReportData creates a MainItem with a SubItem that has progress during the given week.
func seedReportData(t *testing.T, db *gorm.DB, teamBizKey int64, userID uint, weekStart time.Time) (mainItemTitle string) {
	t.Helper()

	mainItem := &model.MainItem{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:      teamBizKey,
		Code:         "TAMA-00003",
		Title:        "Report Test Main Item",
		Priority:     "P1",
		ProposerKey:  int64(userID),
		ItemStatus:   "progressing",
	}
	require.NoError(t, db.Create(mainItem).Error)

	subItem := &model.SubItem{
		BaseModel:     model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:       teamBizKey,
		MainItemKey:   int64(mainItem.ID),
		Title:         "Report Test Sub Item",
		Priority:      "P2",
		ItemStatus:    "progressing",
		Completion:    50,
		Weight:        1.0,
	}
	require.NoError(t, db.Create(subItem).Error)

	// Create a progress record within the week
	record := &model.ProgressRecord{
		BizKey:      snowflake.Generate(),
		SubItemKey:  subItem.BizKey,
		TeamKey:     teamBizKey,
		AuthorKey:   int64(userID),
		Completion:  50,
		Achievement: "Completed half the work",
		CreateTime:  weekStart.Add(24 * time.Hour), // Tuesday of the week
	}
	require.NoError(t, db.Create(record).Error)

	// Update MainItem completion to match
	require.NoError(t, db.Model(mainItem).Update("completion_pct", 50).Error)

	return mainItem.Title
}

// ========== DB Query Helpers ==========

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

// ========== Role Lookup Helpers ==========

// findRoleByName looks up a role by name from the database.
func findRoleByName(t *testing.T, db *gorm.DB, name string) *model.Role {
	t.Helper()
	var role model.Role
	require.NoError(t, db.Where("role_name = ?", name).First(&role).Error)
	return &role
}

// findRoleBizKeyByName looks up a role's BizKey as string by name from the database.
func findRoleBizKeyByName(t *testing.T, db *gorm.DB, name string) string {
	t.Helper()
	var role model.Role
	require.NoError(t, db.Where("role_name = ?", name).First(&role).Error)
	return fmt.Sprintf("%d", role.BizKey)
}

// findRoleBizKeyInt64ByName looks up a role's BizKey as int64 by name.
func findRoleBizKeyInt64ByName(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()
	var role model.Role
	require.NoError(t, db.Where("role_name = ?", name).First(&role).Error)
	return role.BizKey
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

// ========== Item Lifecycle Setup Helpers ==========

// setupLifecycleTest creates a fresh DB and router for item lifecycle tests.
// It extends setupRBACTestDB by adding main_item:change_status to the PM role,
// which is required by the ChangeStatus endpoint but missing from the seed data.
func setupLifecycleTest(t *testing.T) (*gin.Engine, *seedData, *gorm.DB) {
	t.Helper()

	db, data := setupRBACTestDB(t)

	// Add main_item:change_status permission for PM role (required by router but missing from seed)
	pmRole := findRoleByName(t, db, "pm")
	require.NoError(t, db.Create(&model.RolePermission{
		RoleKey:        pmRole.BizKey,
		PermissionCode: "main_item:change_status",
	}).Error)

	r := setupRBACTestRouter(t, db, data)
	return r, data, db
}

// setupRouterFromDB creates a router that reuses an existing DB (needed for seedProgressData-based tests).
func setupRouterFromDB(t *testing.T, db *gorm.DB, data *seedData) *gin.Engine {
	t.Helper()
	return setupRBACTestRouter(t, db, data)
}

// ========== Test Item Creation Helpers ==========

// createTestMainItem is a helper that creates a MainItem via the API and returns the bizKey string.
func createTestMainItem(t *testing.T, r *gin.Engine, token string, teamBizKey int64, title string) string {
	t.Helper()
	body := fmt.Sprintf(`{
			"title": "%s",
			"priority": "P1",
			"assigneeKey": "1",
			"startDate": "2026-01-01",
			"expectedEndDate": "2026-06-30",
			"isKeyItem": false
		}`, title)
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items", teamBizKey), body, token)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	bizKey, ok := data["bizKey"].(string)
	require.True(t, ok, "expected bizKey string in response")
	return bizKey
}

// createTestSubItem creates a SubItem via the API under the given main item.
// Returns the SubItem bizKey string.
func createTestSubItem(t *testing.T, r *gin.Engine, token string, teamBizKey int64, mainItemBizKey string, title string) string {
	t.Helper()
	body := fmt.Sprintf(`{
			"mainItemKey": "%s",
			"title": "%s",
			"priority": "P2",
			"assigneeKey": "1",
			"startDate": "2026-01-01",
			"expectedEndDate": "2026-06-30"
		}`, mainItemBizKey, title)
	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items/%s/sub-items", teamBizKey, mainItemBizKey), body, token)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	bizKey, ok := data["bizKey"].(string)
	require.True(t, ok, "expected bizKey string in response")
	return bizKey
}

// ========== New Helpers (F7 spec) ==========

// createTeamWithMembers creates a team with the given PM and additional members.
// It seeds the team and team_member records directly into the database.
// Returns the team's internal ID.
//
// This is the F7 shared helper extracted from F1 patterns for use by
// future integration test files that need multi-member team setups.
func createTeamWithMembers(t *testing.T, db *gorm.DB, pmID uint, memberCount int) uint {
	t.Helper()

	team := &model.Team{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamName:  fmt.Sprintf("Team-PM%d-M%d", pmID, memberCount),
		PmKey:     int64(pmID),
		Code:      fmt.Sprintf("TPM%d", pmID),
	}
	require.NoError(t, db.Create(team).Error)

	// Add PM as team member with PM role
	pmRoleBizKey := findRoleBizKeyInt64ByName(t, db, "pm")
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey:  int64(team.ID),
		UserKey:  int64(pmID),
		RoleKey:  &pmRoleBizKey,
		JoinedAt: time.Now(),
	}).Error)

	// Create additional member users if needed
	memberRoleBizKey := findRoleBizKeyInt64ByName(t, db, "member")
	for i := 0; i < memberCount; i++ {
		hash, err := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("member%dpass", i)), 4)
		require.NoError(t, err)
		member := &model.User{
			Username:     fmt.Sprintf("teammember-pm%d-%d", pmID, i),
			DisplayName:  fmt.Sprintf("Member %d", i),
			PasswordHash: string(hash),
		}
		require.NoError(t, db.Create(member).Error)
		require.NoError(t, db.Create(&model.TeamMember{
			TeamKey:  int64(team.ID),
			UserKey:  int64(member.ID),
			RoleKey:  &memberRoleBizKey,
			JoinedAt: time.Now(),
		}).Error)
	}

	return team.ID
}

// createMainItem creates a MainItem via the API using a dto.MainItemCreateReq.
// Returns the new item's BizKey (int64).
//
// This is the F7 shared helper extracted from F1 patterns for use by
// future integration test files. Unlike createTestMainItem which takes
// a simple title string, this accepts the full DTO for flexible request bodies.
func createMainItem(t *testing.T, r *gin.Engine, token string, teamBizKey int64, title string, priority string) int64 {
	t.Helper()

	body := fmt.Sprintf(`{
		"title": %q,
		"priority": %q,
		"assigneeKey": "1",
		"startDate": "2026-01-01",
		"expectedEndDate": "2026-06-30",
		"isKeyItem": false
	}`, title, priority)

	w := makeRequest(t, r, http.MethodPost,
		fmt.Sprintf("/api/v1/teams/%d/main-items", teamBizKey), body, token)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	bizKeyStr, ok := data["bizKey"].(string)
	require.True(t, ok, "expected bizKey string in response")
	bizKey, err := strconv.ParseInt(bizKeyStr, 10, 64)
	require.NoError(t, err)
	return bizKey
}

// ========== User Data Helpers ==========

// backfillUserBizKeys sets unique bizKeys on seeded users that have biz_key = 0.
func backfillUserBizKeys(t *testing.T, db *gorm.DB) {
	t.Helper()
	snowflake.Init(1)
	var users []model.User
	require.NoError(t, db.Where("biz_key = 0").Find(&users).Error)
	for _, u := range users {
		require.NoError(t, db.Model(&model.User{}).Where("id = ?", u.ID).Update("biz_key", snowflake.Generate()).Error)
	}
}
