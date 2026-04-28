# Tech Design Evaluation Report — Integration Test Coverage

**Document:** `docs/features/integration-test-coverage/design/tech-design.md`
**Iteration:** 2
**Date:** 2026-04-27

---

## Dimension 1: Architecture Clarity — 20/20

### Layer Placement (7/7)
Explicitly states: "Single-layer feature: test code only. Changes confined to `backend/tests/integration/` (integration tests) and `backend/internal/*/` (unit test gap fixes). Zero production code changes." Clear and unambiguous.

### Component Diagram (7/7)
ASCII file tree showing all 11 files under `backend/tests/integration/` plus F6 unit test files. Each file labeled with Feature ID (F1-F7) and marked NEW vs existing. Helper extraction relationships are annotated with inline comments (e.g., "helper definitions REMOVED"). Complete and actionable.

### Dependencies (6/6)
Three categories cleanly separated: **Existing** (6 external packages with correct import paths including `net/http/httptest`), **Internal packages referenced** (5 packages with specific structs/functions named: `handler.Dependencies`, `handler.SetupRouter`, `appjwt.Claims`, `dto.MainItemCreateReq`, etc.), **New**: None. The iteration-1 garbled import path has been fixed. All internal modules referenced in interface signatures are now listed. Full credit.

---

## Dimension 2: Interface & Model Definitions — 19/20

### Interface Signatures (7/7)
All 13 helper functions have full Go signatures with typed params and return values in code blocks. Multi-return signatures like `seedProgressData` and `seedPoolData` have named return groups. `makeRequest` has parameter semantics documented. `wireHandlers` has a typed signature with clear behavioral contract. Complete.

### Models Concrete (7/7)
`seedData` struct given with all 10 fields named, typed (`uint`, `int64`), and annotated. Status machine transitions for both MainItem and SubItem are spelled out with terminal states and the key difference noted. `dto.MainItemCreateReq` is referenced by name. Sufficient for implementation.

### Directly Implementable (5/6)
A developer can code all helpers from these signatures. The `wireHandlers` function returns `*handler.Dependencies` but the struct's fields are not shown — the doc references `backend/internal/handler/router.go` in the Appendix but does not inline the struct definition. For a test-only feature where this struct already exists and is imported, this is a minor gap. A developer must look up one existing struct to implement `wireHandlers`. Deduct 1.

---

## Dimension 3: Error Handling — 14/15

### Error Types Defined (5/5)
The canonical error response shape is now defined: `{"error": "ERR_CODE", "message": "Human-readable description"}`. Specific error codes are named (`PROGRESS_REGRESSION`, `ERR_ITEM_NOT_FOUND`, `ERR_ROLE_IN_USE`) with reference to `backend/internal/pkg/errors/codes.go`. The iteration-1 gap is resolved.

### Propagation Strategy Clear (4/5)
The bug severity table provides clear triage: Critical (data loss, wrong status transition, auth bypass) → "Fix immediately in same PR"; Non-critical → "File bug, continue." The error assertion pattern shows a concrete code example. However, there is still no stated strategy for how test failures interact with CI — is the suite fail-fast? Does any single test failure block the PR? The doc says "Suite execution < 150s" as a timing target but does not address failure propagation in CI context. For a test-only feature this is partially excusable, but a test design should specify whether all tests run to completion or fail-fast on first error. Deduct 1.

### HTTP Status Codes Mapped (5/5)
The PRD Coverage Map explicitly maps every test to its expected status code: 201, 200, 403, 404, 409, 422. The test naming convention embeds the expected code (e.g., `Returns422`, `Returns403`). The code example shows `http.StatusUnprocessableEntity` and `http.StatusCreated`. Combined with the per-flow test count estimates that list key scenarios, the status codes are effectively mapped across all 160 test cases.

---

## Dimension 4: Testing Strategy — 14/15

### Per-Layer Test Plan (5/5)
Integration tests: F1-F5 each have a dedicated file with estimated test counts (50, 20, 30, 20, 25) and key scenario descriptions. Unit test gaps (F6): 4 gap categories with specific test files, test patterns (httptest+mock, mock repos), and test approaches. F7: extraction with pass-through verification ("All existing tests pass after extraction"). Comprehensive.

### Coverage Target Numeric (4/5)
"54/54 API endpoints with integration tests (100%)", "150+ new test cases", "6 unit test gaps closed", "Suite execution < 150s." Numeric and verifiable. However, there is still no Go coverage percentage target (e.g., `go test -cover` percentage). The PRD's "端点集成测试覆盖率从 33% 提升到 100%" is an endpoint-level metric, but the design does not define how to measure this programmatically or whether a Go `coverage` percentage is expected. For a test-only feature, endpoint counting is the right metric, but the absence of any mention of `go test -cover` or a tool-based verification approach is a gap. Deduct 1.

### Test Tooling Named (5/5)
`stretchr/testify/assert`, `stretchr/testify/require`, `gin`, `gorm`, `httptest`, in-memory SQLite, `encoding/json`. Existing patterns referenced by file path. F6 specifies `httptest + mock RoleService` and mock repos. All already in `go.mod`. Complete.

---

## Dimension 5: Breakdown-Readiness — 19/20

### Components Enumerable (7/7)
All components listed and countable: 5 new integration test files (F1-F5), 1 helpers.go (F7), 1 new unit test file (`permission_handler_test.go`), and edits to existing `*_test.go` files (F6). Total: 7 feature groups, 12+ deliverable files. Each labeled with Feature ID.

### Tasks Derivable (7/7)
Each Feature maps to files with estimated test counts. The F6 unit gaps table maps each gap to a specific test file and pattern. F7 maps to extraction + cleanup. The doc provides enough granularity for task breakdown: F1 (50 tests) → could be split into "CRUD tests", "status transition tests", "cascade tests" etc., but the per-flow file + test count estimates are sufficient for design-level task derivation. The PRD's recommended execution order (F1 → F7 → F2-F5 → F6) provides sequencing guidance. Full credit.

### PRD AC Coverage (5/6)
The PRD Coverage Map addresses all 8 user stories with specific test function names mapped to each AC.

**S1 (7 ACs):** All 7 addressed. PASS.
**S2 (6 ACs):** All 6 addressed. PASS.
**S3 (6 ACs):** All 6 addressed. PASS.
**S4 (5 ACs):** All 5 addressed. PASS.
**S5 (5 ACs):** All 5 addressed. PASS.
**S6 (3 ACs):** All 3 addressed. PASS.
**S7 (5 ACs):** All 5 addressed. PASS.
**S8 (4 ACs):** All 4 addressed. PASS.

The PRD deviation on data isolation is now explicitly acknowledged with a dedicated "DB Isolation Strategy (PRD Deviation)" section, rationale, and reference to Alternatives Considered. This iteration-1 gap is resolved.

However, S7 AC5 states: "Given 审查全部 PR 合并后的代码库, When 检查测试覆盖, Then 各 Feature 测试文件独立且覆盖对应业务域". The PRD Coverage Map addresses this with "all *_test.go" entries, but the design does not define a verification step or acceptance criteria for confirming post-merge coverage completeness. The design assumes the test counts will cover the domains but provides no mechanism to verify that 100% endpoint coverage was actually achieved after all PRs land. Deduct 1.

---

## Dimension 6: Security Considerations — 9/10

### Threat Model Present (4/5)
States "No new production attack surface. Test-only changes." Three concrete risks identified: (1) test-to-production DB connection, (2) JWT test secret leakage, (3) test credential persistence. The threat model is adequate for a test-only feature. However, the analysis is still brief — it does not consider: what happens if a developer runs integration tests against a real database by accident (misconfigured `DATABASE_URL`), or whether the test helpers could be imported by non-test code. For test-only changes this is a minor concern. Deduct 1.

### Mitigations Concrete (5/5)
"In-memory SQLite only — never connect to production databases", "JWT test secret is a hardcoded constant ... only used in test context", "Test users are ephemeral — no credentials persist." Each threat has a paired mitigation. The hardcoded JWT secret is documented as test-only. The in-memory DB pattern is inherent to the architecture. Adequate for the threat surface.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Architecture Clarity | 20 | 20 |
| Interface & Model Definitions | 19 | 20 |
| Error Handling | 14 | 15 |
| Testing Strategy | 14 | 15 |
| Breakdown-Readiness | 19 | 20 |
| Security Considerations | 9 | 10 |
| **Total** | **95** | **100** |

---

## ATTACKS

1. **Interface & Model Definitions — `wireHandlers` return type not inlined**: The doc defines `func wireHandlers(t *testing.T, db *gorm.DB, data *seedData) *handler.Dependencies` but does not show what `*handler.Dependencies` contains. A developer implementing `wireHandlers` must look up `backend/internal/handler/router.go` to understand the struct fields. Must either inline the struct definition or add a comment with the key fields.

2. **Error Handling — CI failure propagation undefined**: The bug severity table defines in-PR triage but does not specify how test failures propagate in CI. Questions left unanswered: Does the suite fail-fast or run to completion? Does any single test failure block the PR merge? Must add a sentence like "CI runs all tests to completion; any failure blocks PR merge" or "Suite fails fast on critical-tier failures only."

3. **Testing Strategy — no programmatic coverage verification**: The design targets "54/54 API endpoints with integration tests (100%)" but provides no mechanism to verify this programmatically (e.g., `go test -cover`, a coverage script, or an endpoint checklist audit step). Must either add a `go test -cover` target or define a manual verification step in the task breakdown.

4. **Breakdown-Readiness — no post-merge coverage verification for S7 AC5**: S7 AC5 requires confirming that "各 Feature 测试文件独立且覆盖对应业务域" after all PRs merge. The design maps test files to features but does not define how to verify post-merge that coverage is complete. Must add a verification step or acceptance test for the coverage claim.
