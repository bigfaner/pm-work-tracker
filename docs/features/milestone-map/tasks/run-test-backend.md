---
id: "T-test-run-backend"
title: "Run API Functional Test"
priority: "P1"
estimated_time: "30min-1h"
dependencies: ["T-test-gen-scripts-backend", "T-test-gen-scripts-frontend"]
type: "test.run"
surface-key: "backend"
surface-type: "api"
---

## Acceptance Criteria
- [ ] All API functional tests pass

Execute staged test scripts for the milestone-map feature.

## Feature Paths

Discover the feature's testing directory layout before starting:
```bash
ls docs/features/milestone-map/testing/                                 # journeys
ls docs/features/milestone-map/testing/<journey>/contracts/              # contracts
```

## Feature Context
- Scope: backend

Run all staged test scripts. If tests fail, identify root cause, apply minimal fix, and re-run.

Type: **api**
