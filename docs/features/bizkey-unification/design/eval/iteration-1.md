---
date: "2026-04-28"
doc_dir: "docs/features/bizkey-unification/design/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

**Score: 62/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  16      │  20      │ ⚠️          │
│    Layer placement explicit  │   7/7    │          │            │
│    Component diagram present │   5/7    │          │            │
│    Dependencies listed       │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  18      │  20      │ ✅          │
│    Interface signatures typed│   7/7    │          │            │
│    Models concrete           │   5/7    │          │            │
│    Directly implementable    │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │   4      │  15      │ ❌          │
│    Error types defined       │   2/5    │          │            │
│    Propagation strategy clear│   2/5    │          │            │
│    HTTP status codes mapped  │   0/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │   6      │  15      │ ❌          │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   0/5    │          │            │
│    Test tooling named        │   1/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  16      │  20      │ ✅          │
│    Components enumerable     │   6/7    │          │            │
│    Tasks derivable           │   5/7    │          │            │
│    PRD AC coverage           │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │   2      │  10      │ ❌          │
│    Threat model present      │   1/5    │          │            │
│    Mitigations concrete      │   1/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  62      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 16/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Error Handling section | "No new error types. Existing error propagation is unchanged." — zero analysis of what happens when `GetTeamBizKey` receives a missing/malformed context value | -3 pts |
| Error Handling section | No HTTP status codes mapped anywhere — middleware failure path (team not found, invalid bizKey) has no documented response shape | -5 pts |
| Testing Strategy section | "Existing coverage maintained. No coverage regression." — not a numeric target | -5 pts |
| Testing Strategy section | No test frameworks named — testify, httptest, go test are all absent from the section | -4 pts |
| Security section | "No security impact." — one sentence dismissal in a system with RBAC and team-scoped access control | -8 pts |
| Architecture section | Component diagram shows layer flow only; the 7 handler files and 8 service files are mentioned in prose but not shown as enumerated components in the diagram | -2 pts |
| Interface section (§6) | MainItemService, SubItemService, ItemPoolService signatures shown as comments (`// BEFORE: Create(ctx, teamID...)`), not actual typed Go signatures | -2 pts |

---

## Attack Points

### Attack 1: Error Handling — complete absence of HTTP error contract

**Where**: "No new error types. Existing error propagation is unchanged." and "The `isPMRole` helper returns `false` on `FindByBizKey` error (role not found) — this behavior is preserved."

**Why it's weak**: The middleware is being changed to inject `teamBizKey int64` instead of `teamID uint`. If `c.Get("teamBizKey")` returns nil (context key missing, wrong type assertion), the handler silently gets a zero value `int64`. There is no documented behavior for this failure path — no error type, no HTTP status code, no panic/abort strategy. The doc treats error handling as a non-topic for a refactor, but the middleware boundary change is exactly where a type mismatch would surface at runtime. "Existing error propagation is unchanged" is an assertion, not a design decision.

**What must improve**: Document what `GetTeamBizKey` does when the context key is absent or wrong type. Map the middleware failure path to an HTTP status (401? 500?). Even if it's "same as before," show the before behavior explicitly so a reviewer can verify the claim.

---

### Attack 2: Testing Strategy — no numeric coverage target, no tooling named

**Where**: "Overall Coverage Target: Existing coverage maintained. No coverage regression."

**Why it's weak**: "Existing coverage maintained" is unmeasurable. There is no baseline number stated, so "no regression" cannot be verified by CI or a reviewer. The rubric requires a numeric target (e.g., 80%). Additionally, the entire testing section names test files but never names a single framework — `testify`, `httptest`, `gin.CreateTestContext`, `go test` are all absent. A new contributor reading this section cannot determine what tooling to use or what assertion style is expected.

**What must improve**: State the current coverage percentage as the baseline target (e.g., "maintain ≥ current 72%"). Name the test frameworks: `github.com/stretchr/testify/assert`, `net/http/httptest`, `go test -race`.

---

### Attack 3: Security — one-sentence dismissal in an RBAC system

**Where**: "No security impact. `bizKey` is already the externally-exposed identifier. Internal `uint` IDs remain hidden — this change reinforces that boundary."

**Why it's weak**: The project has an RBAC system (evidenced by `backend/internal/migration/rbac.go` and `backend/tests/integration/rbac_test.go` in the git history). The middleware change directly affects how team authorization is enforced — `TeamScopeMiddleware` now injects `teamBizKey int64` instead of resolving to an internal `uint`. The claim "no security impact" is made without any threat analysis. Specifically: (1) if a caller crafts a request with a valid bizKey for team A but the JWT belongs to team B, does the middleware catch it? (2) the `isPMRole` fix changes how role authorization works — is there a privilege escalation risk if `FindByBizKey` returns a wrong role? These are not addressed.

**What must improve**: Add a minimal threat model: enumerate the authorization checks that `TeamScopeMiddleware` performs, confirm the bizKey-to-team validation is preserved post-refactor, and explicitly state that the RBAC permission checks downstream are unaffected by the type change.

---

## Previous Issues Check

N/A — iteration 1.

---

## Verdict

- **Score**: 62/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 16/20 — can proceed to `/breakdown-tasks`
- **Action**: Iteration 1 complete. Three dimensions need work before this design is production-ready: Error Handling (4/15), Testing Strategy (6/15), Security (2/10).
