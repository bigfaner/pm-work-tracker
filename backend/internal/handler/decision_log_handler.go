package handler

import (
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
)

// DecisionLogHandler handles decision log endpoints.
type DecisionLogHandler struct {
	svc          service.DecisionLogService
	userRepo     repository.UserRepo
	mainItemRepo repository.MainItemRepo
}

// NewDecisionLogHandler creates a new DecisionLogHandler with service and repo dependencies.
func NewDecisionLogHandler(svc service.DecisionLogService, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo) *DecisionLogHandler {
	if svc == nil {
		panic("decision_log_handler: decisionLogService must not be nil")
	}
	if userRepo == nil {
		panic("decision_log_handler: userRepo must not be nil")
	}
	if mainItemRepo == nil {
		panic("decision_log_handler: mainItemRepo must not be nil")
	}
	return &DecisionLogHandler{svc: svc, userRepo: userRepo, mainItemRepo: mainItemRepo}
}
