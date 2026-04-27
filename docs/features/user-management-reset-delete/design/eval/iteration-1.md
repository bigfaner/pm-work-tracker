---
date: "2026-04-27"
doc_dir: "docs/features/user-management-reset-delete/design/"
iteration: "1"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

**Score: 74/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  18      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  7/7     │          │            │
│    Dependencies listed       │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  15      │  20      │ ⚠️         │
│    Interface signatures typed│  6/7     │          │            │
│    Models concrete           │  4/7     │          │            │
│    Directly implementable    │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  13      │  15      │ ✅         │
│    Error types defined       │  4/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  12      │  15      │ ⚠️         │
│    Per-layer test plan       │  4/5     │          │            │
│    Coverage target numeric   │  4/5     │          │            │
│    Test tooling named        │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  13      │  20      │ ⚠️         │
│    Components enumerable     │  6/7     │          │            │
│    Tasks derivable           │  4/7     │          │            │
│    PRD AC coverage           │  3/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┤
│ 6. Security Considerations   │  3       │  10      │ ⚠️         │
│    Threat model present      │  2/5     │          │            │
│    Mitigations concrete      │  1/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┤
│ TOTAL                        │  74      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 12/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:88 | Dependencies table lists only 4 items; missing internal package dependencies (GORM scopes, `pkg/errors`, `pkg/jwt`) that are directly referenced in the design | -1 pts (Dependencies) |
| tech-design.md:96-105 | AdminService interface is shown as prose method signatures, not as a proper Go interface block with full type names — `AdminUserDTO` package path ambiguous | -1 pts (Interface signatures) |
| tech-design.md:134-157 | Only one new DTO (`ResetPasswordReq`) is defined. No DTO/model shown for the delete response, and `AdminUserDTO` is referenced but never defined — developer must hunt existing code for its shape | -3 pts (Models concrete) |
| tech-design.md:164-169 | Error table defines new codes `CANNOT_DELETE_SELF` and `USER_DELETED` but does not show their Go declarations. Existing codebase uses `&AppError{...}` pattern — the exact declarations must be specified so a developer can add them to `errors.go` | -1 pts (Error types) |
| tech-design.md:96-99 | `ResetPassword` returns `*dto.AdminUserDTO` — but the API handbook shows `ResetPasswordResp` with different fields (bizKey, username, displayName). The design contradicts itself on what the response shape is | -1 pts (HTTP status codes) |
| tech-design.md:191-196 | Per-layer test table lists coverage targets but conflates unit tests for handler/service into "85%" with no explanation of why it differs from the "80%" overall target stated later | -1 pts (Coverage target) |
| tech-design.md:191-196 | No integration/E2E test layer listed. PRD Story 1 has a network timeout scenario and Story 3 has stale-state — these require integration-level testing not just unit mocks | -1 pts (Per-layer test plan) |
| tech-design.md:213-216 | Frontend test tooling lists `vitest` and `@testing-library/react` but does not mention MSW or any HTTP mocking library for API module tests, which the design implies with `vi.mock` | -1 pts (Test tooling) |
| tech-design.md:243-254 | PRD Story 2 AC: "两次输入不一致" (passwords don't match) — the design only validates single-password strength. No mention of a confirm-password field or its validation logic in the DTO or frontend type | -1 pts (PRD AC coverage) |
| tech-design.md:243-254 | PRD Story 3 AC: "该用户的历史工作记录仍完整保留" — design mentions "Does NOT remove team memberships" but does not address whether progress records, sub-items, or other foreign-key references are preserved or become orphan-visible | -1 pts (PRD AC coverage) |
| tech-design.md:243-254 | PRD Story 4 AC: "鼠标悬停显示'不可删除自身账号'" — no design for tooltip/hover behavior, only service-level self-delete guard | -1 pts (PRD AC coverage) |
| tech-design.md:226-233 | Threat model lists "Password transmitted in plaintext" but mitigation only says "HTTPS only" without specifying whether the API will enforce HSTS or redirect HTTP→HTTPS at the application level | -1 pts (Threat model) |
| tech-design.md:235-239 | Mitigations section is mostly prose restatements of the threat table. No concrete code-level mitigation shown (e.g., the exact GORM scope call, the exact middleware code diff) | -2 pts (Mitigations) |
| tech-design.md:235-239 | No mention of rate limiting on the reset-password endpoint — a super admin account could be used to brute-force password resets on other users | -2 pts (Mitigations) |
| api-handbook.md vs tech-design.md | API handbook `ResetPasswordReq` type differs from tech-design: handbook uses `ResetPasswordReq`/`ResetPasswordResp` struct names, tech-design says DTO lives in `dto/auth.go`. Inconsistent file placement | -1 pts (Directly implementable) |

---

## Attack Points

### Attack 1: Breakdown-Readiness — Response shape contradiction blocks task derivation

**Where**: tech-design.md line 99 shows `ResetPassword(...) (*dto.AdminUserDTO, error)` but api-handbook.md lines 131-137 show `ResetPasswordResp` with fields `bizKey`, `username`, `displayName`. The existing `AdminUserDTO` likely has more fields (role, status, teams, etc.).
**Why it's weak**: A developer implementing the handler cannot determine what the response should be. Should the service return `AdminUserDTO` (full user) or a trimmed `ResetPasswordResp` (3 fields)? The VO conversion step is undefined. This directly blocks task breakdown because the handler implementation task is ambiguous.
**What must improve**: Pick one response shape. If `ResetPasswordResp` is the intended response, update the service signature to return it (or define the VO conversion). If `AdminUserDTO`, update the API handbook. Show the exact Go struct for whatever is returned.

### Attack 2: Security — Password reset has no rate limiting or audit logging

**Where**: tech-design.md lines 226-239 list threats and mitigations, but neither mentions rate limiting on `PUT /admin/users/:userId/password` or audit logging for password changes.
**Why it's weak**: A compromised admin session could silently reset every user's password with no rate limit and no trace. The threat model identifies "Non-admin calls reset/delete API" but not "Compromised admin session mass-resets passwords." The existing codebase already has a `DeletedTime` audit trail pattern — yet the design does not propose logging password reset events with timestamp and caller identity. This is a significant security gap for an admin credential-change operation.
**What must improve**: Add rate limiting (e.g., max N resets per minute per admin) or at minimum an audit log entry per reset. Define the log structure or database record. If the project has no audit infrastructure, state this as a known gap with a remediation plan.

### Attack 3: Breakdown-Readiness — Missing confirm-password validation and frontend dialog state machine

**Where**: tech-design.md line 247 says "Frontend form validation + Gin binding tags" covers Story 2, but PRD Story 2 AC requires: "两次输入不一致" (passwords don't match). The `ResetPasswordReq` DTO has only `newPassword` — no confirm field. The frontend types also show only `newPassword: string`.
**Why it's weak**: Two concrete gaps: (1) No `confirmPassword` field in any DTO or type definition, so the task-breakdown cannot derive a form-validation task. (2) The design describes dialogs as "opens, validates, submits, closes" but never defines the dialog state machine: what props trigger open/close, what loading state exists during mutation, what error state keeps the dialog open. PRD Story 1 AC explicitly requires "弹窗保持打开，显示后端返回的错误信息" — this implies a form-state design (idle → submitting → error → idle) that is completely absent.
**What must improve**: Add `confirmPassword` to the frontend type (not the API DTO — it is frontend-only validation). Define the dialog component's props interface, open/close trigger, loading state, and error-display mechanism. At minimum, a typed React component signature for the reset-password dialog.

---

## Previous Issues Check

<!-- Only for iteration > 1 -->

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|

---

## Verdict

- **Score**: 74/100
- **Target**: 90/100
- **Gap**: 16 points
- **Breakdown-Readiness**: 13/20 — can proceed to /breakdown-tasks but with significant ambiguity
- **Action**: Continue to iteration 2 to address the 3 attack points (response shape contradiction, security gaps, missing frontend dialog design)
