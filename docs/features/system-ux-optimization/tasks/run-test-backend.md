---
id: "T-test-run-backend"
title: "Run API Functional Test"
priority: "P1"
estimated_time: "30min-1h"
dependencies: ["T-test-gen-scripts-api", "T-test-gen-scripts-web"]
type: "test.run"
surface-key: "backend"
surface-type: "api"
---

Execute staged test scripts for the system-ux-optimization feature.

## Feature Paths

Discover the feature's testing directory layout before starting:
```bash
ls docs/features/system-ux-optimization/testing/                                 # journeys
ls docs/features/system-ux-optimization/testing/<journey>/contracts/              # contracts
```

## Feature Context
- Scope: backend

Run all staged test scripts. If tests fail, identify root cause, apply minimal fix, and re-run.

Type: **api**

## Acceptance Criteria
- [ ] 1. Task completes without error
