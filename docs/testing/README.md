# Testing Strategy: Multi-Feature Test Case Management

This project has 3 features that collectively define the PM Work Tracker system.
Each feature generates its own test cases from its PRD, but they overlap and
modify the same product. This document explains how to merge them into a
coherent test suite.

## Feature Inventory

| Feature | Scope | Test Cases | Authority |
|---------|-------|-----------|-----------|
| `pm-work-tracker` | Full product (original PRD, Ant Design) | 90 TCs (68 UI + 22 API) | API tests are current; UI tests superseded by improve-ui |
| `improve-ui` | UI rebuild (shadcn/ui, 5 structural changes) | 56 TCs (52 UI + 4 API) | UI tests are current; API deltas are current |
| `config-yaml` | Backend config refactor | 22 TCs (20 CLI + 2 API) | All current |

## Merge Strategy

```
Effective Test Suite =
    config-yaml  CLI tests          (all 22 TCs)
  + improve-ui   UI tests           (all 52 TCs, supersedes pm-work-tracker UI)
  + improve-ui   API tests          (4 TCs — new endpoints)
  + pm-work-tracker API tests       (22 TCs — core API behavior)
  - pm-work-tracker UI tests        (superseded by improve-ui)
```

### Why This Works

1. **pm-work-tracker API tests stay current** — the API layer was not
   fundamentally changed by improve-ui (only 4 new/modified endpoints).
   Core API behavior (auth, RBAC, team isolation, CRUD, status transitions,
   progress records, performance) remains valid.

2. **improve-ui UI tests supersede pm-work-tracker UI tests** — all 11 pages
   were rebuilt with shadcn/ui. The old Ant Design-based UI test cases describe
   a UI that no longer exists. improve-ui's test cases reflect the current UI.

3. **improve-ui API tests cover deltas** — 4 new endpoints (user CRUD, weekly
   comparison API) that pm-work-tracker's tests don't cover.

4. **config-yaml is orthogonal** — backend config only, no UI overlap.

### Overlap Notes

| Area | pm-work-tracker TCs | improve-ui TCs | Resolution |
|------|--------------------|-----------------|------------|
| Login page UI | TC-001~005 | (part of page rebuilds) | Use improve-ui |
| Team management UI | TC-006~011 | (page rebuild) | Use improve-ui |
| Super admin UI | TC-012~017 | (structural change: split tabs→independent pages) | Use improve-ui |
| Item list UI | TC-018~030 | (structural change: Summary/Detail toggle) | Use improve-ui |
| All API (auth, CRUD, status, isolation) | TC-069~090 | TC-053~056 (deltas only) | Use both (no conflict) |
| Config/startup | — | — | Use config-yaml |

## Files

```
docs/
├── testing/
│   └── README.md                          (this file)
└── features/
    ├── pm-work-tracker/testing/test-cases.md   (90 TCs)
    ├── improve-ui/testing/test-cases.md        (56 TCs)
    └── config-yaml/testing/test-cases.md       (22 TCs)
```

## When to Regenerate

Regenerate a feature's test cases when its PRD changes:

```bash
# After editing any PRD under docs/features/<slug>/prd/
/gen-test-cases docs/features/<slug>
```

The merge strategy documented here remains stable as long as the feature
hierarchy doesn't change (i.e., improve-ui remains a UI-layer delta on
pm-work-tracker, and config-yaml remains backend-only).

## For E2E Script Generation

When running `/gen-test-scripts`, use these test case files:

- **UI scripts**: `improve-ui/testing/test-cases.md` (52 TCs)
- **API scripts**: `pm-work-tracker/testing/test-cases.md` (API section, 22 TCs)
  + `improve-ui/testing/test-cases.md` (API section, 4 TCs)
- **CLI scripts**: `config-yaml/testing/test-cases.md` (22 TCs)
