---
id: "T-test-gen-scripts-backend"
title: "Generate API Functional Test Scripts"
priority: "P1"
estimated_time: "1-2h"
dependencies: ["T-eval-contract"]
type: "test.gen-scripts"
surface-key: "backend"
surface-type: "api"
---

## Acceptance Criteria
- [ ] API functional test scripts generated for all backend journeys
- [ ] Scripts follow project testing conventions (Vitest, tests/backend/)

Generate executable test scripts for the milestone-map feature.
Test type: api.

## Feature Paths

Discover the feature's testing directory layout before starting:
```bash
ls docs/features/milestone-map/testing/                                 # journeys
ls docs/features/milestone-map/testing/<journey>/contracts/              # contracts
```

Read the approved test cases and generate scripts using the framework from the surface.

Type: **api**
