---
id: "T-test-gen-scripts-web"
title: "Generate Web E2E Test Scripts"
priority: "P1"
estimated_time: "1-2h"
dependencies: ["T-eval-contract"]
type: "test.gen-scripts"
surface-key: ""
surface-type: "web"
---

Generate executable test scripts for the system-ux-optimization feature.
Test type: web.

## Feature Paths

Discover the feature's testing directory layout before starting:
```bash
ls docs/features/system-ux-optimization/testing/                                 # journeys
ls docs/features/system-ux-optimization/testing/<journey>/contracts/              # contracts
```

Read the approved test cases and generate scripts using the framework from the surface.

Type: **web**

## Acceptance Criteria
- [ ] 1. Task completes without error
