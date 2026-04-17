package main

import (
	"fmt"
	"log"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/handler"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

func main() {
	// 1. Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2. Init DB
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("database error: %v", err)
	}

	// 3. Init repositories
	teamRepo := gormrepo.NewGormTeamRepo(db)

	// 4. Init handler stubs
	// Services will be wired into handlers in a later task.
	deps := &handler.Dependencies{
		Config:   cfg,
		TeamRepo: teamRepo,
		Auth:     handler.NewAuthHandler(),
		Team:     handler.NewTeamHandler(),
		MainItem: handler.NewMainItemHandler(),
		SubItem:  handler.NewSubItemHandler(),
		Progress: handler.NewProgressHandler(),
		ItemPool: handler.NewItemPoolHandler(),
		View:     handler.NewViewHandler(),
		Report:   handler.NewReportHandler(),
		Admin:    handler.NewAdminHandler(),
	}

	// 5. Setup router
	r := handler.SetupRouter(deps)

	// 6. Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
