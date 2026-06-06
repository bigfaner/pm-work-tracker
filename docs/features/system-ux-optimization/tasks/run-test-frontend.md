---
id: "T-test-run-frontend"
title: "Run Web E2E Test"
priority: "P1"
estimated_time: "30min-1h"
dependencies: ["T-test-run-backend"]
type: "test.run"
surface-key: "frontend"
surface-type: "web"
---

Execute staged test scripts for the system-ux-optimization feature.

## Feature Paths

Discover the feature's testing directory layout before starting:
```bash
ls docs/features/system-ux-optimization/testing/                                 # journeys
ls docs/features/system-ux-optimization/testing/<journey>/contracts/              # contracts
```

## Feature Context
- Scope: frontend

Run all staged test scripts. If tests fail, identify root cause, apply minimal fix, and re-run.

Type: **web**

## Acceptance Criteria
- [ ] 1. Task completes without error
