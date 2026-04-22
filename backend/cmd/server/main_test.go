package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/handler"
	"pm-work-tracker/backend/internal/model"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
	"pm-work-tracker/backend/internal/service"
)

// writeTestConfig writes a minimal valid config.yaml to a temp file and returns its path.
func writeTestConfig(t *testing.T, jwtSecret string) string {
	t.Helper()
	content := []byte(`server:
  port: "18080"
  gin_mode: test
  read_timeout: 5s
  write_timeout: 5s
  max_body_size: 1048576

database:
  driver: sqlite
  path: ":memory:"

auth:
  jwt_secret: "` + jwtSecret + `"
  jwt_expiry: 1h
  initial_admin:
    username: ""
    password: ""

cors:
  origins: []

logging:
  level: info
  format: json
`)
	f, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	_, err = f.Write(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestRun_FailsWhenAssetsInvalid(t *testing.T) {
	// In the test environment dist/index.html is absent, so ValidateAssets must
	// cause run() to return a "startup: ..." error before touching the DB.
	path := writeTestConfig(t, "test-secret-that-is-at-least-32-bytes!!")
	err := run(path, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "startup:")
}

func TestRun_LoadsConfigFromFile(t *testing.T) {
	path := writeTestConfig(t, "test-secret-that-is-at-least-32-bytes!!")
	cfg, err := config.LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, "18080", cfg.Server.Port)
	assert.Equal(t, "test-secret-that-is-at-least-32-bytes!!", cfg.Auth.JWTSecret)
}

func TestRun_InitDBWithConfig(t *testing.T) {
	path := writeTestConfig(t, "test-secret-that-is-at-least-32-bytes!!")
	cfg, err := config.LoadConfig(path)
	require.NoError(t, err)

	db, err := config.InitDB(&cfg.Database)
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func TestRun_SeedAdminSkippedWhenNoUsername(t *testing.T) {
	path := writeTestConfig(t, "test-secret-that-is-at-least-32-bytes!!")
	cfg, err := config.LoadConfig(path)
	require.NoError(t, err)

	db, err := config.InitDB(&cfg.Database)
	require.NoError(t, err)

	// Config has empty initial_admin.username, so SeedAdmin should be a no-op
	err = config.SeedAdmin(db, &cfg.Auth)
	assert.NoError(t, err)
}

func TestRun_SeedAdminCreatesUser(t *testing.T) {
	content := []byte(`server:
  port: "18080"
  gin_mode: test
  read_timeout: 5s
  write_timeout: 5s
  max_body_size: 1048576

database:
  driver: sqlite
  path: ":memory:"

auth:
  jwt_secret: "test-secret-that-is-at-least-32-bytes!!"
  jwt_expiry: 1h
  initial_admin:
    username: "admin"
    password: "admin123"

cors:
  origins: []

logging:
  level: info
  format: json
`)
	f, err := os.CreateTemp("", "config-seed-*.yaml")
	require.NoError(t, err)
	_, err = f.Write(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	t.Cleanup(func() { os.Remove(f.Name()) })

	cfg, err := config.LoadConfig(f.Name())
	require.NoError(t, err)

	db, err := config.InitDB(&cfg.Database)
	require.NoError(t, err)

	// Run migrations so the users table exists
	err = db.AutoMigrate(&model.User{})
	require.NoError(t, err)

	err = config.SeedAdmin(db, &cfg.Auth)
	assert.NoError(t, err)

	// Verify SeedAdmin is idempotent
	err = config.SeedAdmin(db, &cfg.Auth)
	assert.NoError(t, err)
}

func TestRun_ServerAppliesTimeouts(t *testing.T) {
	path := writeTestConfig(t, "test-secret-that-is-at-least-32-bytes!!")
	cfg, err := config.LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, cfg.Server.ReadTimeout.Value())
	assert.Equal(t, 5*time.Second, cfg.Server.WriteTimeout.Value())
	assert.Equal(t, int64(1048576), cfg.Server.MaxBodySize)
}

func TestRun_WiredRouterHealthCheck(t *testing.T) {
	path := writeTestConfig(t, "test-secret-that-is-at-least-32-bytes!!")
	cfg, err := config.LoadConfig(path)
	require.NoError(t, err)

	db, err := config.InitDB(&cfg.Database)
	require.NoError(t, err)

	teamRepo := gormrepo.NewGormTeamRepo(db)
	userRepo := gormrepo.NewGormUserRepo(db)
	authSvc := service.NewAuthService(userRepo, cfg.Auth.JWTSecret)

	deps := &handler.Dependencies{
		Config:   cfg,
		TeamRepo: teamRepo,
		UserRepo: userRepo,
		RoleRepo: gormrepo.NewGormRoleRepo(db),
		Auth:     handler.NewAuthHandler(authSvc),
		Team:     handler.NewTeamHandler(&handler.StubTeamSvc{}, &handler.StubRouterRepoUser{}),
		MainItem: handler.NewMainItemHandler(&handler.StubMainItemSvc{}, &handler.StubRouterRepoUser{}, &handler.StubRouterRepoSubItem{}),
		SubItem:  handler.NewSubItemHandler(&handler.StubSubItemSvc{}),
		Progress: handler.NewProgressHandler(),
		ItemPool: handler.NewItemPoolHandler(&handler.StubItemPoolSvc{}, &handler.StubRouterRepoUser{}, &handler.StubRouterRepoMainItem{}),
		View:     handler.NewViewHandler(),
		Report:   handler.NewReportHandler(),
		Admin:    handler.NewAdminHandler(&handler.StubAdminSvc{}),
	}

	r := handler.SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
