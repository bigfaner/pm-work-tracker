---
id: "T-test-gen-scripts-frontend"
title: "Generate Web E2E Test Scripts"
priority: "P1"
estimated_time: "1-2h"
dependencies: ["T-test-gen-scripts-backend"]
type: "test.gen-scripts"
surface-key: "frontend"
surface-type: "web"
---

## Acceptance Criteria
- [ ] Web E2E test scripts generated for all frontend journeys
- [ ] Scripts follow project testing conventions (Playwright, tests/frontend/)

Generate executable test scripts for the milestone-map feature.
Test type: web.

## Feature Paths

Discover the feature's testing directory layout before starting:
```bash
ls docs/features/milestone-map/testing/                                 # journeys
ls docs/features/milestone-map/testing/<journey>/contracts/              # contracts
```

Read the approved test cases and generate scripts using the framework from the surface.

Type: **web**
