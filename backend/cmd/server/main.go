package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/handler"
	"pm-work-tracker/backend/internal/migration"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/web"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	dev := flag.Bool("dev", false, "dev mode: skip embedded asset validation")
	flag.Parse()

	if err := run(*configPath, *dev); err != nil {
		log.Fatalf("%v", err)
	}
}

// run wires the full application: config, DB, seed, repos, services, handlers,
// router, and HTTP server with graceful shutdown.
// devMode: when true, skips embedded asset validation (for local development).
func run(configPath string, devMode bool) error {
	if !devMode {
		if err := web.ValidateAssets(web.FS); err != nil {
			return fmt.Errorf("startup: %w", err)
		}
	}
	// 1. Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// 2. Init DB
	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// 3. Run migrations
	runner := migration.NewRunner(db, "migrations")
	if err := runner.Run(); err != nil {
		return fmt.Errorf("migration error: %w", err)
	}

	// 3b. RBAC migration (roles, role_permissions, preset roles)
	if err := migration.MigrateToRBAC(db); err != nil {
		return fmt.Errorf("rbac migration error: %w", err)
	}

	// 4. Seed admin user
	if err := config.SeedAdmin(db, &cfg.Auth); err != nil {
		log.Printf("warning: seed admin: %v", err)
	}

	// 4. Init repositories
	teamRepo := gormrepo.NewGormTeamRepo(db)
	userRepo := gormrepo.NewGormUserRepo(db)
	mainItemRepo := gormrepo.NewGormMainItemRepo(db)
	subItemRepo := gormrepo.NewGormSubItemRepo(db)
	progressRepo := gormrepo.NewGormProgressRepo(db)
	itemPoolRepo := gormrepo.NewGormItemPoolRepo(db)

	// 5. Init services
	authSvc := service.NewAuthService(userRepo, cfg.Auth.JWTSecret)
	roleRepo := gormrepo.NewGormRoleRepo(db)
	teamSvc := service.NewTeamService(teamRepo, userRepo, mainItemRepo, roleRepo, db)
	statusHistoryRepo := gormrepo.NewGormStatusHistoryRepo(db)
	statusHistorySvc := service.NewStatusHistoryService(statusHistoryRepo)
	mainItemSvc := service.NewMainItemService(mainItemRepo, subItemRepo, statusHistorySvc)
	subItemSvc := service.NewSubItemService(subItemRepo, mainItemSvc, statusHistorySvc)
	progressSvc := service.NewProgressService(progressRepo, subItemRepo, mainItemSvc, statusHistorySvc)
	itemPoolSvc := service.NewItemPoolService(itemPoolRepo, subItemRepo, mainItemRepo, db)
	viewSvc := service.NewViewService(mainItemRepo, subItemRepo, progressRepo)
	reportSvc := service.NewReportService(mainItemRepo, subItemRepo, progressRepo)
	adminSvc := service.NewAdminService(userRepo, teamRepo)
	roleSvc := service.NewRoleService(roleRepo, userRepo)

	// 6. Init handlers
	deps := &handler.Dependencies{
		Config:     cfg,
		TeamRepo:   teamRepo,
		UserRepo:   userRepo,
		RoleRepo:   roleRepo,
		Auth:       handler.NewAuthHandler(authSvc),
		Team:       handler.NewTeamHandler(teamSvc, userRepo),
		MainItem:   handler.NewMainItemHandler(mainItemSvc, userRepo, subItemRepo),
		SubItem:    handler.NewSubItemHandler(subItemSvc),
		Progress:   handler.NewProgressHandler(progressSvc, userRepo),
		ItemPool:   handler.NewItemPoolHandler(itemPoolSvc, userRepo, mainItemRepo),
		View:       handler.NewViewHandler(viewSvc),
		Report:     handler.NewReportHandler(reportSvc),
		Admin:      handler.NewAdminHandler(adminSvc),
		Role:       handler.NewRoleHandler(roleSvc),
		Permission: handler.NewPermissionHandler(roleSvc),
	}

	// 7. Setup router
	r := handler.SetupRouter(deps, nil)

	// 8. Start server with timeouts from config
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout.Value(),
		WriteTimeout: cfg.Server.WriteTimeout.Value(),
	}

	go func() {
		log.Printf("starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}
	log.Println("server stopped")
	return nil
}
