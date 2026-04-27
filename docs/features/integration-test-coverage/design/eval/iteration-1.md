# Tech Design Evaluation Report — Integration Test Coverage

**Document:** `docs/features/integration-test-coverage/design/tech-design.md`
**Iteration:** 1
**Date:** 2026-04-27

---

## Dimension 1: Architecture Clarity — 18/20

### Layer Placement (7/7)
Explicitly states: "Single-layer feature: test code only. Changes confined to `backend/tests/integration/` (integration tests) and `backend/internal/*/` (unit test gap fixes). Zero production code changes." Clear, unambiguous, single sentence that any developer can act on.

### Component Diagram (7/7)
ASCII file tree diagram showing all 11 files under `backend/tests/integration/` plus F6 unit test files. Each file is labeled with its Feature ID (F1-F7) and marked NEW vs existing. Relationships between helpers.go and consumer files are clear. Well done.

### Dependencies (4/6)
Lists existing dependencies by name (`testify/assert`, `testify/require`, `gin`, `gorm`, `httptest`) and correctly states "New: None." However, the `testify` packages are listed in a single comma-separated line that includes a formatting error: `"github.com/gin-gonic/gin/test httptest"` — this is not a valid Go import path (should be `net/http/httptest` or `github.com/gin-gonic/gin` separately). The import for `appjwt` (used in `signTokenWithClaims`) and `dto` (used in `createMainItem`) are not listed as dependencies despite being referenced in interface signatures. Deduct 2 for missing internal module dependencies and the garbled import path.

---

## Dimension 2: Interface & Model Definitions — 19/20

### Interface Signatures (7/7)
All 13 helper functions have full Go signatures with typed parameters and return values in a code block. `seedProgressData` and `seedPoolData` have multi-return signatures with named return values. `makeRequest` is described with parameter semantics. `wireHandlers` has a typed signature. Thorough and directly implementable.

### Models Concrete (7/7)
`seedData` struct is given with all fields named, typed (`uint`, `int64`), and annotated with inline comments. Status machine reference is provided with explicit transition rules for both MainItem and SubItem. The `dto.MainItemCreateReq` type is referenced by name for `createMainItem`'s parameter.

### Directly Implementable (5/6)
A developer can code all helpers from these signatures. However, the `wireHandlers` function references `*handler.Dependencies` without showing what that struct contains or where it is defined. The doc says "Called by all setup* functions" but does not provide the struct definition — a developer would need to look up `backend/internal/handler/router.go` to understand the return type. Deduct 1.

---

## Dimension 3: Error Handling — 12/15

### Error Types Defined (4/5)
The doc defines an error assertion pattern with code examples showing assertion on `http.StatusUnprocessableEntity` and structured response with `"error"` key. However, no custom error types or error codes are defined. The doc describes what tests assert (HTTP status codes + response structure) but does not enumerate the specific error codes or error response shapes that the API returns. For a test-only feature this is partially acceptable, but the doc should specify the canonical error response shape (e.g., `{"error": "CODE", "message": "..."}`) to avoid each test writer guessing at the structure. Deduct 1.

### Propagation Strategy Clear (4/5)
The bug severity table provides a clear triage strategy: "Fix immediately in same PR" vs "File bug, continue." The error assertion pattern shows a concrete code example. However, there is no stated strategy for how test failures propagate to CI — is the suite fail-fast? Do individual test failures block the PR? The doc mentions "Suite execution < 150s" as a target but does not say what happens when a test finds a real bug in production code. Deduct 1.

### HTTP Status Codes Mapped (4/5)
The test naming convention and PRD coverage map implicitly reference status codes (201, 200, 403, 404, 409, 422). The example code shows `http.StatusUnprocessableEntity` and `http.StatusCreated`. However, there is no explicit status code mapping table in the design doc itself — the codes are scattered across the PRD coverage map and test names. A consolidated table mapping error scenarios to expected status codes would improve implementability. Deduct 1.

---

## Dimension 4: Testing Strategy — 14/15

### Per-Layer Test Plan (5/5)
Integration tests: each of F1-F5 has a dedicated file with estimated test counts and key scenarios. Unit test gaps (F6): each gap has a named test file and test pattern (httptest+mock, mock repo, etc.). F7 helpers: extraction with pass-through verification. All layers covered.

### Coverage Target Numeric (4/5)
"54/54 API endpoints with integration tests (100%)", "150+ new test cases", "6 unit test gaps closed", "Suite execution < 150s." These are numeric and verifiable. However, there is no Go coverage percentage target (e.g., `go test -cover` target). The PRD says "端点集成测试覆盖率从 33% 提升到 100%" but the design does not translate this into a `go test -cover` percentage or a way to measure endpoint coverage programmatically. Deduct 1.

### Test Tooling Named (5/5)
`stretchr/testify/assert`, `stretchr/testify/require`, `gin`, `gorm`, `httptest`, in-memory SQLite. Existing patterns from `auth_isolation_test.go`, `progress_completion_test.go`, `rbac_test.go` are referenced by file path. F6 unit tests specify `httptest + mock RoleService` and mock repos. All tooling is named and already in `go.mod`.

---

## Dimension 5: Breakdown-Readiness — 18/20

### Components Enumerable (7/7)
All components are explicitly listed and countable: 5 new integration test files, 1 helpers.go file, 1 new unit test file (`permission_handler_test.go`), and 5 edits to existing `*_test.go` files. Each has a Feature ID (F1-F7). Total: 13 deliverables.

### Tasks Derivable (6/7)
Each Feature maps to at least one file → at least one task. Helpers (F7) maps to extraction + cleanup tasks. F1-F5 each map to a test file with estimated test counts. F6 maps to 6 gap fixes in specific files. However, the doc does not break F1 (50 tests) into subtasks — a developer implementing F1 would need to further decompose it into "set up CRUD tests", "set up status transition tests", "set up cascade tests", etc. The design stops at the file level for F1-F5, which is reasonable for a design doc but slightly insufficient for direct task creation without intermediate decomposition. Deduct 1.

### PRD AC Coverage (5/6)
The PRD Coverage Map addresses all 8 user stories. Cross-referencing acceptance criteria:

**S1 (7 ACs):** All 7 addressed in the coverage map. PASS.
**S2 (6 ACs):** All 6 addressed. PASS.
**S3 (6 ACs):** All 6 addressed. PASS.
**S4 (5 ACs):** All 5 addressed. PASS.
**S5 (5 ACs):** All 5 addressed. PASS.
**S6 (3 ACs):** All 3 addressed. PASS.
**S7 (5 ACs):** All 5 addressed. PASS.
**S8 (4 ACs):** All 4 addressed. PASS.

However, there is a discrepancy between the PRD data isolation strategy and the design. The PRD states: "使用事务回滚（`tx.Begin()` + `t.Cleanup(tx.Rollback)`）保证无持久化残留" (use transaction rollback). The design explicitly contradicts this: "Each test creates a unique in-memory SQLite DB via `fmt.Sprintf(...)` ... Tests do NOT share databases." The design says the "Alternatives Considered" table rejected transaction rollback because "SQLite shared-cache + GORM tx support is fragile." This is a legitimate design decision but it is not flagged as a PRD deviation — the design doc should explicitly note this divergence from the PRD's data isolation approach. Deduct 1 for unacknowledged PRD deviation.

---

## Dimension 6: Security Considerations — 8/10

### Threat Model Present (4/5)
States "No new production attack surface. Test-only changes." Identifies three concrete risks: (1) tests connecting to production databases, (2) JWT test secret leakage, (3) test credential persistence. However, the threat model is thin — it does not consider risks like: test code accidentally committed with real credentials, test helpers creating users with known passwords that could be exploited if tests run against staging, or the hardcoded JWT secret being copy-pasted into non-test code. Deduct 1 for shallow analysis.

### Mitigations Concrete (4/5)
"In-memory SQLite only — never connect to production databases" and "JWT test secret is a hardcoded constant ... only used in test context" and "Test users are ephemeral — no credentials persist." These are concrete and paired with threats. However, the hardcoded JWT secret `test-secret-that-is-at-least-32-bytes!!` is written in plaintext in the design doc — there is no mitigation for the risk of this being copied to production config. Deduct 1.

---

## Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Architecture Clarity | 18 | 20 |
| Interface & Model Definitions | 19 | 20 |
| Error Handling | 12 | 15 |
| Testing Strategy | 14 | 15 |
| Breakdown-Readiness | 18 | 20 |
| Security Considerations | 8 | 10 |
| **Total** | **89** | **100** |

---

## ATTACKS

1. **Architecture Clarity — garbled import path and missing internal module deps**: The dependencies section contains `"github.com/gin-gonic/gin/test httptest"` which is not a valid import path (should be `net/http/httptest`). Additionally, `appjwt` and `dto` packages are referenced in interface signatures but not listed as dependencies. Must fix the import path and add all referenced internal packages.

2. **Error Handling — no canonical error response shape defined**: The doc shows an example asserting `resp` contains `"error"` but never defines the full error response contract (e.g., `{"error": {"code": "string", "message": "string"}}`). Without this, each test writer will reverse-engineer the error shape independently. Must add a concrete error response model or reference to the existing error response struct.

3. **Breakdown-Readiness — unacknowledged PRD deviation on data isolation**: The PRD specifies "使用事务回滚（`tx.Begin()` + `t.Cleanup(tx.Rollback)`）保证无持久化残留" but the design uses per-test in-memory SQLite databases instead, and explicitly rejected transaction rollback in the alternatives table without acknowledging this as a PRD deviation. Must add an explicit note like "PRD Deviation: Data isolation uses per-test DB rather than transaction rollback — see Alternatives Considered for rationale."
