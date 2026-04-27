---
date: "2026-04-27"
doc_dir: "docs/features/user-management-reset-delete/design/"
iteration: "2"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 92/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  19      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  7/7     │          │            │
│    Dependencies listed       │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  19      │  20      │ ✅         │
│    Interface signatures typed│  7/7     │          │            │
│    Models concrete           │  7/7     │          │            │
│    Directly implementable    │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  14      │  15      │ ✅         │
│    Error types defined       │  4/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  13      │  15      │ ✅         │
│    Per-layer test plan       │  4/5     │          │            │
│    Coverage target numeric   │  5/5     │          │            │
│    Test tooling named        │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  19      │  20      │ ✅         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  7/7     │          │            │
│    PRD AC coverage           │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  8       │  10      │ ✅         │
│    Threat model present      │  4/5     │          │            │
│    Mitigations concrete      │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  92      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 12/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:84-89 | Dependencies table lists 4 external items but omits `gin-contrib/limiter` which is referenced in the deferred rate-limiting section (line 281). Internal packages (`pkg/errors`, `pkg/jwt`) referenced in error propagation and auth sections are also absent from the table. | -1 pts (Dependencies) |
| tech-design.md:117-121 | Handler method signatures do not show the `ParseBizKeyParam` call or how `:userId` path param is extracted and converted. A developer must look at existing handlers to replicate the pattern. | -1 pts (Directly implementable) |
| tech-design.md:198-204 | Error table gives symbolic names (`ErrCannotDeleteSelf`, etc.) but not the Go declarations (e.g., `var ErrCannotDeleteSelf = &AppError{Code: "CANNOT_DELETE_SELF", ...}`). The existing codebase uses `&AppError{...}` pattern -- the exact var declarations a developer would add to `errors.go` are unspecified. | -1 pts (Error types) |
| tech-design.md:225-230 | Per-layer test table has no integration/E2E row. PRD Story 1 has network timeout scenarios and Story 3 has stale-state scenarios that benefit from integration-level testing. Auth middleware testing (line 241) is mentioned but not given its own layer or tooling entry. | -1 pts (Per-layer test plan) |
| tech-design.md:241 | Auth middleware test "deleted user's JWT rejected → 401" is listed as a key scenario but no test tooling or infrastructure is specified for middleware testing in the per-layer table. | -1 pts (Test tooling) |
| tech-design.md:301 | PRD Story 4 AC requires "该行的「删除」按钮为禁用状态，鼠标悬停显示'不可删除自身账号'" but the PRD coverage map only addresses the backend guard (`Service checks callerID == user.ID`). The frontend disabled-button + tooltip behavior is not mapped to any design component. | -1 pts (PRD AC coverage) |
| tech-design.md:260-266 | Threat model does not include "compromised admin session mass-resets passwords." The deferred rate-limiting section acknowledges this risk, but the threat model table itself is incomplete. | -1 pts (Threat model) |
| tech-design.md:272 | "No password logging" is stated as prose ("Password field is transient, never persisted in logs or response DTOs") but no code-level enforcement is shown (e.g., a `log:"-"` struct tag, a custom `MarshalJSON` that redacts the field, or a log sanitizer). A developer using structured logging could accidentally log the `ResetPasswordReq` including the plaintext password. | -1 pts (Mitigations) |

---

## Attack Points

### Attack 1: Breakdown-Readiness — Story 4 frontend AC unmapped in PRD coverage

**Where**: tech-design.md line 301 — PRD coverage map row for Story 4 reads "Service checks `callerID == user.ID`" with interface `ErrCannotDeleteSelf`. The frontend dialog states section (lines 186-192) covers the delete confirmation dialog but not the disabled-button rendering on the user list row.
**Why it's weak**: PRD Story 4 AC requires two things: (1) backend self-delete guard, and (2) frontend disabled button with hover tooltip "不可删除自身账号". The coverage map addresses only (1). The Frontend Dialog States section defines the delete *confirmation* dialog states but not the disabled-row rendering logic. A developer breaking down tasks would create a task for the backend guard but might miss the frontend disabled-button + tooltip task because it has no coverage row. This is exactly the kind of gap that PRD coverage maps are meant to prevent.
**What must improve**: Add a row to the PRD coverage map: "Story 4: Delete button disabled for self + hover tooltip | Frontend conditional rendering in UserManagementPage | `isSuperAdmin` check + `disabled` prop + tooltip text". Optionally add a row to the Frontend Dialog States section or a separate "List Row States" mini-table covering the disabled/active button states.

### Attack 2: Security — "No password logging" mitigation is unenforced prose

**Where**: tech-design.md line 272 — "No password logging: Password field is transient, never persisted in logs or response DTOs."
**Why it's weak**: This is a policy statement, not a concrete mitigation. In a Go service with structured logging, a developer could easily write `log.Info().Interface("request", req).Msg("reset password")` and leak the plaintext password into logs. Without a code-level safeguard (e.g., making `ResetPasswordReq` implement `fmt.Stringer` with redaction, adding a `logging:"redact"` struct tag, or adding a linter rule), this mitigation depends entirely on developer discipline. For a security-critical field like a password, a concrete enforcement mechanism should be specified in the design.
**What must improve**: Either (1) add a concrete code-level safeguard (e.g., a custom `String()` or `MarshalZerologObject()` method on `ResetPasswordReq` that redacts `NewPassword`), or (2) explicitly state that the existing codebase has a request-logging middleware that sanitizes sensitive fields and reference that middleware. If neither exists, state this as a known gap in the mitigations section.

### Attack 3: Error Handling — Error type declarations are not shown

**Where**: tech-design.md lines 198-204 — the error table lists codes and names (`ErrCannotDeleteSelf`, `ErrUserDeleted`, `ErrUserNotFound`, `ErrValidation`) but the actual Go variable declarations are absent.
**Why it's weak**: The existing codebase uses `&AppError{...}` pattern (referenced in the propagation strategy at line 209). A developer must infer the exact struct initialization: does it use `Code` field? `Message` field? `HTTPStatus` field? What's the pattern for a new error vs an existing one? `ErrUserNotFound` is listed as a "new" error code but the codebase likely already has `ErrUserNotFound` for general 404s. The design does not clarify whether these are new declarations, aliases of existing ones, or reuse of existing sentinel errors with different HTTP mappings. This ambiguity could lead to duplicate error declarations or incorrect HTTP status mapping during implementation.
**What must improve**: Show the exact Go declarations for at least one error (e.g., `var ErrCannotDeleteSelf = apperrors.New("CANNOT_DELETE_SELF", 422, "cannot delete self")`). Clarify which errors are genuinely new vs reuses of existing sentinels. This takes 3-4 lines and removes all ambiguity.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Response shape contradiction between service interface and API handbook | ✅ Yes | Service interface (line 99) now returns `*dto.ResetPasswordResp`. API handbook (lines 131-139) and tech design (lines 146-152) both define `ResetPasswordResp` with identical fields: `bizKey`, `username`, `displayName`. All three sources are consistent. |
| Attack 2: No rate limiting or audit logging for password reset | ✅ Yes (deferred with rationale) | Lines 276-289 add two explicit "Deferred" sections with rationale: admin routes are already gated by permission middleware, and no audit infrastructure exists in the current system. Remediation paths are specified (`gin-contrib/limiter` pattern for rate limiting, `audit_log` table for audit logging). The threat is acknowledged rather than silently omitted. |
| Attack 3: Missing confirmPassword field and frontend dialog states | ✅ Yes | Line 171 adds an explicit note explaining `confirmPassword` is frontend-only and never sent to the API. Lines 173-192 define complete dialog state tables for both Reset Password (5 states: idle, validationError, submitting, success, error) and Delete Confirmation (4 states: confirming, submitting, success, error) with triggers and UI display descriptions. |

---

## Verdict

- **Score**: 92/100
- **Target**: 90/100
- **Gap**: 0 points (target exceeded by 2)
- **Breakdown-Readiness**: 19/20 — can proceed to /breakdown-tasks
- **Action**: Target reached. All three iteration-1 attacks have been addressed. Minor issues remain (Story 4 frontend unmapped, error declarations not shown, password logging unenforced) but none block task breakdown. Proceed to `/breakdown-tasks`.
