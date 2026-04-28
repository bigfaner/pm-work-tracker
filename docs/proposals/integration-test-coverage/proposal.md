# Proposal: Comprehensive Integration Test Coverage

## Problem

The system has 54 API endpoints but only 18 (33%) have integration tests. The 36 untested endpoints span every core domain: item management, team operations, admin functions, views, and reports. Several unit test gaps also exist (`permission_handler.go` completely untested, `ConvertToMain`/`UpdateTeam`/`GetByBizKey` methods lack coverage).

**Evidence:**

| Domain | Total Endpoints | Tested | Gap |
|--------|----------------|--------|-----|
| MainItem CRUD | 6 | 0 | 100% |
| SubItem lifecycle | 7 | 0 | 100% |
| ItemPool | 6 | 1 | 83% |
| Progress | 3 | 1 | 67% |
| Views | 4 | 0 | 100% |
| Team management | 7 | 0 | 100% |
| Admin users | 5 | 1 | 80% |
| Reports | 2 | 1 | 50% |

**Unit test gaps:** `permission_handler.go` (0 tests), `ItemPoolService.ConvertToMain`, `TeamService.UpdateTeam`, 3x `GetByBizKey` methods.

**Urgency:** Without integration tests, regressions in API behavior (auth, validation, status transitions, cascading effects) are only caught manually. The existing unit tests cover service logic but miss handler→middleware→service→repository wiring errors. A recent example: the Views domain had zero integration tests when commit `1883499` (weekly view refactor) introduced a timezone bug in `view_handler.go` that rejected valid week-start dates as "future weeks." The bug escaped to manual testing because no integration test exercised the `GET /teams/:id/views/weekly` endpoint with local-timezone dates (recorded in `docs/lessons/weekly-view-bug-fixes.md`, Bug 3). The same refactor also broke weekly filtering logic in `view_service.go` (Bug 2: all weeks returned identical data), again undetected by any automated test.

## Solution

Organize integration tests by **user flow** — each test file exercises a complete end-to-end workflow, covering multiple endpoints in sequence. This catches inter-endpoint bugs (e.g., create→status-change→archive cascading) that isolated endpoint tests miss.

### Test Files (5 flow files + 1 unit gap file)

#### F1: Item Lifecycle (`item_lifecycle_test.go`)
Full item management flow: create main item → create sub items → append progress → status transitions → archive.

| Endpoint | Tests |
|----------|-------|
| `POST /teams/:id/main-items` | Happy path (201), validation errors (missing title, invalid priority, invalid dates), duplicate code, permission denied (member role) |
| `GET /teams/:id/main-items` | List with pagination, filter by status/priority/assignee, empty list, search by title |
| `GET /teams/:id/main-items/:itemId` | Get existing item (200), not found (404), wrong team (403) |
| `PUT /teams/:id/main-items/:itemId` | Update title/priority/dates (200), validation errors, not found (404), permission denied |
| `PUT /teams/:id/main-items/:itemId/status` | All valid transitions (in-progress→completed, etc.), invalid transition (422), terminal status side effects (sub-items auto-complete), permission denied |
| `GET /teams/:id/main-items/:itemId/available-transitions` | Returns correct transitions for current status, empty for terminal states |
| `POST /teams/:id/main-items/:itemId/archive` | Archive completed item (200), archive in-progress item (fails), not found (404) |
| `POST /teams/:id/main-items/:itemId/sub-items` | Create sub-item (201), weight validation, duplicate code, main-item-not-found (404) |
| `GET /teams/:id/main-items/:itemId/sub-items` | List sub-items for main item, empty list, pagination |
| `GET /teams/:id/sub-items/:subId` | Get sub-item detail (200), not found (404), wrong team (403) |
| `PUT /teams/:id/sub-items/:subId` | Update sub-item fields (200), validation, not found |
| `PUT /teams/:id/sub-items/:subId/status` | All valid transitions, invalid transition, terminal cascade to main-item recalculation |
| `GET /teams/:id/sub-items/:subId/available-transitions` | Correct transitions by status, empty for terminal |
| `PUT /teams/:id/sub-items/:subId/assignee` | Assign to team member (200), assign to non-member (403), clear assignee |
| `POST /teams/:id/sub-items/:subId/progress` | Append progress (200), regression blocked (422), auto status transition on 100%, completion rollup to main item |
| `GET /teams/:id/sub-items/:subId/progress` | List progress records (reverse chronological), empty list |
| `PATCH /teams/:id/progress/:recordId/completion` | Correct latest record (syncs sub-item), correct non-latest (no cascade), not found (404) |

#### F2: Item Pool Flow (`item_pool_test.go`)
Submit → review → assign/convert/reject lifecycle.

| Endpoint | Tests |
|----------|-------|
| `POST /teams/:id/item-pool` | Submit item (201), validation errors, duplicate submission, permission denied |
| `GET /teams/:id/item-pool` | List with status filter, pagination, empty pool |
| `GET /teams/:id/item-pool/:poolId` | Get detail (200), not found (404) |
| `POST /teams/:id/item-pool/:poolId/assign` | Assign to main item (creates sub-item, updates pool status), invalid main item (rollback), already assigned (409), permission denied |
| `POST /teams/:id/item-pool/:poolId/convert-to-main` | Convert to new main item (200), already processed (409), permission denied |
| `POST /teams/:id/item-pool/:poolId/reject` | Reject with reason (200), already processed (409), missing reason (422), permission denied |

#### F3: Team Management (`team_management_test.go`)
Team CRUD, membership, role management.

| Endpoint | Tests |
|----------|-------|
| `POST /teams` | Create team (201), auto-join as PM, duplicate code validation |
| `GET /teams` | List user's teams, empty list for new user |
| `GET /teams/:id` | Team detail (200), not member (403), not found (404) |
| `PUT /teams/:id` | Update name/description (200), not PM (403), validation errors |
| `DELETE /teams/:id` | Disband team (200), cascade deletes items, not PM (403), team not found |
| `GET /teams/:id/search-users` | Search by username (200), empty results, permission denied |
| `POST /teams/:id/members` | Invite member with role (200), already member (409), user not found (404), permission denied |
| `DELETE /teams/:id/members/:userId` | Remove member (200), PM cannot be removed (422), not member (404), permission denied |
| `PUT /teams/:id/members/:userId/role` | Update role (200), cannot change PM role (403), invalid role (422), permission denied |

#### F4: Admin User Management (`admin_user_test.go`)
User CRUD and status management, admin team listing.

| Endpoint | Tests |
|----------|-------|
| `GET /admin/users` | List with pagination, search by username/displayName |
| `POST /admin/users` | Create user (201), duplicate username (409), validation errors, permission denied |
| `GET /admin/users/:userId` | Get user detail (200), not found (404) |
| `PUT /admin/users/:userId` | Update display name/role (200), validation errors, not found (404) |
| `PUT /admin/users/:userId/status` | Enable/disable user (200), cannot disable self (422), not found (404) |
| `GET /admin/teams` | List all teams with member counts |

#### F5: Views & Reports (`views_reports_test.go`)
Data aggregation, export, and reporting.

| Endpoint | Tests |
|----------|-------|
| `GET /teams/:id/views/weekly` | Given 3 items created and 1 completed this week, 2 in-progress: response contains `stats: {NEW: 0, completed: 1, inProgress: 2, overdue: 0}`; previous-week delta shows `+3` for created; empty team returns all-zero stats |
| `GET /teams/:id/views/gantt` | Given 2 items spanning Jan 1-31 and Jan 15-Feb 15: response contains 2 entries with correct `startDate`/`endDate`, `status` field maps to expected color key; sub-items nested under parent; empty team returns empty array |
| `GET /teams/:id/views/table` | Given items with mixed statuses: `?status=completed` returns only completed items; `?overdue=true` returns items past dueDate with status != completed; pagination `?page=1&pageSize=2` returns `total` matching filtered count; empty result returns `{items: [], total: 0}` |
| `GET /teams/:id/views/table/export` | CSV response starts with UTF-8 BOM (`\xEF\xBB\xBF`); first row is headers matching field names; data rows match table endpoint output for same filters; empty team returns BOM + headers only (no data rows) |
| `GET /teams/:id/reports/weekly/preview` | Given items with activity this week: markdown contains `## Summary` section with correct stat counts, `## Items` section listing item titles; empty team returns markdown with "no activity this week" placeholder |
| `GET /teams/:id/reports/weekly/export` | Full markdown includes all sections (summary, item details, progress highlights); `Content-Type: text/markdown`; response body is valid markdown parseable by any standard renderer |

#### F6: Unit Test Gaps
Fill missing unit tests alongside integration work.

| File | Gap |
|------|-----|
| `permission_handler.go` | Entire file — add handler tests for `GetPermissions`, `GetPermissionCodes` |
| `item_pool_service_test.go` | Add `ConvertToMain` test (transactional: creates main item + updates pool status) |
| `team_service_test.go` | Add `UpdateTeam` test (PM auth check + field update) |
| `item_pool_service_test.go` | Add `GetByBizKey` test |
| `progress_service_test.go` | Add `GetByBizKey` test |
| `sub_item_service_test.go` | Add `GetByBizKey` test |

## Alternatives

### A. Endpoint-isolated test files (rejected)
One file per handler (e.g., `main_item_test.go`, `sub_item_test.go`).

- **Pro:** Each test is self-contained, easier to debug when it fails. Faster to write per-endpoint (~15 min/test case) because no setup chaining is needed. Lower test brittleness — one endpoint changing doesn't break tests for other endpoints.
- **Con:** Misses inter-endpoint bugs. Example: a `ConvertToMain` flow spans pool submission, sub-item creation, and pool status update — isolated tests for each endpoint would all pass even if the handoff between them breaks.
- **Estimated effort:** ~120 test cases across 36 endpoints (3.3 avg per endpoint), ~30 developer-hours. CI adds ~20s (tests are short and independent).

### B. Do nothing (status quo)
Maintain current 33% endpoint coverage.

- **Pro:** Zero immediate investment. Team continues shipping features without test-maintenance overhead. CI time stays the same. No risk of flaky integration tests blocking PRs.
- **Con:** Each new feature or refactor in untested domains (Views, Reports, ItemPool, Team management) risks the same class of escape that hit the weekly view (see urgency section). The blast radius grows as the codebase ages — the schema alignment refactor required 12+ fix commits, many catching issues that integration tests would have caught on first break.
- **Estimated cost:** Near-zero upfront, but each escaped regression costs ~2–4 hours of manual diagnosis (based on the weekly-view incident timeline: 5 bugs over ~1 day).

### C. Contract tests instead of integration tests
Use API contract testing (e.g., Pact) to verify request/response shapes.

- **Pro:** Catches schema regressions (field renames, type changes) automatically — would have caught the `int64→string` snowflake ID migration and the `status→item_status` field rename that broke tests in the schema alignment refactor.
- **Con:** Contract tests verify shapes but not business logic. Specific bug classes they miss: (1) middleware chain bugs — e.g., a `RequireTeamMember()` middleware that fails to inject `teamID` into context causes a 500 on every team-scoped endpoint, but the response shape is still `{error: string}` so the contract passes; (2) GORM scope errors — e.g., a `db.Where("team_id = ?", teamID)` that silently returns all rows when `teamID` is zero (unscoped query), producing correct-shaped but wrong-data responses; (3) status machine violations — e.g., an "archive in-progress item" request that should return 422 but succeeds against the database, a bug that only surfaces when the test exercises the full handler→service→repository chain with real state. A layered approach (contract tests for shape + integration tests for logic) was considered but rejected: the integration tests already assert response structure in every test case, making a separate contract layer redundant for a single-service application with one consumer.
- **Estimated effort:** ~40 developer-hours to set up Pact + write contracts for 36 endpoints, plus ongoing broker maintenance.

**Verdict:** We chose flow-based integration tests because they catch both inter-endpoint wiring bugs and business logic errors against a real database, with no new tooling dependencies. A hybrid approach (flow-based for complex paths, endpoint-isolated for simple CRUD) was considered but rejected: mixing two test patterns creates inconsistent test structure that increases cognitive overhead for reviewers and maintainers, while the effort savings are marginal (endpoint-isolated saves ~15 min/test case on setup but each CRUD endpoint has only 3-4 cases, saving ~45 min per domain). At ~40 developer-hours for the entire suite, the consistency of one pattern outweighs the estimated 4-6 hours saved by a hybrid split.

## Scope

### In Scope
- 5 integration test files covering 36 untested endpoints, organized by user flow
- Edge cases per endpoint: happy path, validation errors, permission denied, not found, cascading effects (see test tables for specific cases per endpoint)
- 6 unit test gap fixes (permission_handler, ConvertToMain, UpdateTeam, 3x GetByBizKey)
- Shared test helpers extracted from existing integration tests for reuse

**Estimated effort:** ~40 developer-hours for the entire suite.

**Recommended execution order:** F1 -> F2 -> F3 -> F4 -> F5 -> F6. F1 (item lifecycle) is first because it covers the largest surface area (17 endpoints) and establishes shared helpers that F2-F5 reuse. F2-F4 follow as independent flows of similar complexity. F5 (views/reports) comes after F1-F4 because its test data depends on seeded items and sub-items. F6 (unit gaps) is last because it is independent of the integration test infrastructure and can be picked up in parallel or as filler between flow PRs.

### Out of Scope
- Frontend test changes (frontend test suite already covers component and E2E flows)
- Performance/load testing
- E2E browser-based testing (separate workflow)
- New features or bug fixes — this is purely test coverage

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Test execution time increases significantly | Medium | Medium | Each flow file is independent; can run in parallel. Target <30s per file. |
| Test data pollution between flows | Medium | High | Each flow creates isolated teams/users. Use transaction rollback (`tx.Begin()` at test start, `tx.Rollback()` in `t.Cleanup`) to guarantee no persisted state. This is preferred over manual cleanup because it is automatic and cannot be forgotten. |
| Edge cases expose existing bugs | High | Low (good) | This is a benefit. Severity threshold: if the bug causes data loss, incorrect business state (e.g., wrong status transition accepted), or auth bypass, fix immediately in the same PR. Otherwise, file a bug and continue — do not block the test coverage PR on non-critical fixes. |
| Large PR difficult to review | Medium | Medium | Submit one flow file per PR for incremental review. |
| Test brittleness from response-schema coupling | Medium | Medium | Assert on structural fields (status code, top-level keys, specific business fields like `itemStatus`) but avoid asserting on field ordering, error message text, or pagination metadata format. When a test fails due to a legitimate API change, update the test — do not suppress failures. |

## Success Criteria

- All 54 API endpoints have at least one integration test
- Each endpoint covers: happy path (200/201), validation errors (400/422), permission denied (403), not found (404)
- Status transition endpoints test all valid transitions and at least one invalid transition
- Cascading effects are verified (e.g., progress append → completion rollup, status change → sub-item cascade)
- All 6 unit test gaps listed in F6 resolved (each named file/method has at least one passing test)
- Total integration test count ≥ 150 new test cases
- `go test ./tests/integration/...` passes with 0 failures
- Total integration test suite completes in <150s (5 flow files x 30s target per file)
