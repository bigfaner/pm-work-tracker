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
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
	"pm-work-tracker/backend/internal/service"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	if err := run(*configPath); err != nil {
		log.Fatalf("%v", err)
	}
}

// run wires the full application: config, DB, seed, repos, services, handlers,
// router, and HTTP server with graceful shutdown.
func run(configPath string) error {
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

	// 3. Seed admin user
	if err := config.SeedAdmin(db, &cfg.Auth); err != nil {
		log.Printf("warning: seed admin: %v", err)
	}

	// 4. Init repositories
	teamRepo := gormrepo.NewGormTeamRepo(db)
	userRepo := gormrepo.NewGormUserRepo(db)

	// 5. Init services
	authSvc := service.NewAuthService(userRepo, cfg.Auth.JWTSecret)

	// 6. Init handlers
	deps := &handler.Dependencies{
		Config:   cfg,
		TeamRepo: teamRepo,
		Auth:     handler.NewAuthHandler(authSvc),
		Team:     handler.NewTeamHandler(),
		MainItem: handler.NewMainItemHandler(),
		SubItem:  handler.NewSubItemHandler(),
		Progress: handler.NewProgressHandler(),
		ItemPool: handler.NewItemPoolHandler(),
		View:     handler.NewViewHandler(),
		Report:   handler.NewReportHandler(),
		Admin:    handler.NewAdminHandler(),
	}

	// 7. Setup router
	r := handler.SetupRouter(deps)

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
