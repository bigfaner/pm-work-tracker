package main

import (
	"fmt"
	"log"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/handler"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
	"pm-work-tracker/backend/internal/service"
)

func main() {
	// 1. Load config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2. Init DB
	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}

	// 3. Init repositories
	teamRepo := gormrepo.NewGormTeamRepo(db)
	userRepo := gormrepo.NewGormUserRepo(db)

	// 4. Init services
	authSvc := service.NewAuthService(userRepo, cfg.Auth.JWTSecret)

	// 5. Init handlers
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

	// 6. Setup router
	r := handler.SetupRouter(deps)

	// 7. Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
