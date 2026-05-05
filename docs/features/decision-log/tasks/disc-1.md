---
id: "disc-1"
title: "Fix: Wire DecisionLog handler in backend main.go"
priority: "P0"
dependencies: []
status: 
breaking: true
---

# disc-1: Fix: Wire DecisionLog handler in backend main.go

DecisionLog handler, service, and repository code exists but is never wired in backend/cmd/server/main.go deps struct (lines 113-129). This causes nil pointer dereference at decision_log_handler.go:42 on all decision-log API endpoints, returning HTTP 500. Fix: add decisionLogRepo, decisionLogSvc, and handler.NewDecisionLogHandler(...) to deps.
