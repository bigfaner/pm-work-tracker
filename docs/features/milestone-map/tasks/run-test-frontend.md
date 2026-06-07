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

## Acceptance Criteria
- [ ] All Web E2E tests pass

Execute staged test scripts for the milestone-map feature.

## Feature Paths

Discover the feature's testing directory layout before starting:
```bash
ls docs/features/milestone-map/testing/                                 # journeys
ls docs/features/milestone-map/testing/<journey>/contracts/              # contracts
```

## Feature Context
- Scope: frontend

Run all staged test scripts. If tests fail, identify root cause, apply minimal fix, and re-run.

Type: **web**
