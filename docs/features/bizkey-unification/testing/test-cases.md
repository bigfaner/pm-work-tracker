---
feature: "bizkey-unification"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
generated: "2026-04-28"
---

# Test Cases: bizkey-unification

## Summary

| Type | Count |
|------|-------|
| UI   | 0     |
| API  | 5     |
| CLI  | 2     |
| **Total** | **7** |

---

## API Test Cases

## TC-001: Progress record stores correct snowflake team_key

- **Source**: Story 1 / AC-1
- **Type**: API
- **Target**: api/progress
- **Test ID**: api/progress/progress-record-stores-correct-snowflake-team-key
- **Pre-conditions**: Team exists with bizKey `123456789012345678`; authenticated user is a member of that team
- **Steps**:
  1. POST `/api/v1/teams/123456789012345678/progress` with valid progress payload
  2. Query the database for the newly created progress record
- **Expected**: The `team_key` column of the inserted record equals `123456789012345678` (the snowflake bizKey), not the team's internal uint auto-increment ID (e.g. `5`)
- **Priority**: P0

---

## TC-002: Invite member with PM role bizKey returns ErrCannotAssignPMRole

- **Source**: Story 2 / AC-1
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/invite-member-with-pm-role-bizkey-returns-errcannot-assign-pm-role
- **Pre-conditions**: PM role exists with bizKey `987654321098765432` and internal ID `2`; caller is authenticated as team PM
- **Steps**:
  1. POST `/api/v1/teams/{teamBizKey}/members/invite` with body `{ "roleBizKey": 987654321098765432 }`
- **Expected**: HTTP 400 (or appropriate error status); response body contains error code / message corresponding to `ErrCannotAssignPMRole`
- **Priority**: P0

---

## TC-003: Non-numeric teamId in URL returns 400

- **Source**: Spec Section 5.5 — bizKey 校验规则 (rule 1)
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/non-numeric-team-id-returns-400
- **Pre-conditions**: Server is running; any authenticated user
- **Steps**:
  1. Send any request to a team-scoped endpoint with a non-numeric teamId, e.g. GET `/api/v1/teams/abc/items`
- **Expected**: HTTP 400; response indicates validation error (invalid teamId format)
- **Priority**: P1

---

## TC-004: Non-positive teamId in URL returns 400

- **Source**: Spec Section 5.5 — bizKey 校验规则 (rule 2)
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/non-positive-team-id-returns-400
- **Pre-conditions**: Server is running; any authenticated user
- **Steps**:
  1. Send any request to a team-scoped endpoint with teamId `0`, e.g. GET `/api/v1/teams/0/items`
  2. Repeat with a negative value, e.g. GET `/api/v1/teams/-1/items`
- **Expected**: HTTP 400 for both cases; response indicates validation error
- **Priority**: P1

---

## TC-005: Non-existent teamId returns 404

- **Source**: Spec Section 5.5 — bizKey 校验规则 (rule 3)
- **Type**: API
- **Target**: api/teams
- **Test ID**: api/teams/non-existent-team-id-returns-404
- **Pre-conditions**: Server is running; the teamId used does not correspond to any team in the database
- **Steps**:
  1. Send any request to a team-scoped endpoint with a valid positive int64 that has no matching team, e.g. GET `/api/v1/teams/999999999999999999/items`
- **Expected**: HTTP 404; response indicates team not found
- **Priority**: P1

---

## CLI Test Cases

## TC-006: Build fails when uint internal ID passed to int64 bizKey parameter

- **Source**: Story 3 / AC-1
- **Type**: CLI
- **Target**: cli/build
- **Test ID**: cli/build/build-fails-when-uint-passed-to-int64-bizkey-param
- **Pre-conditions**: Go toolchain installed; repository is at the post-migration state where all Service interface methods accept `int64` for external ID params
- **Steps**:
  1. In any service call site, temporarily change a `teamBizKey int64` argument to a `uint` variable
  2. Run `go build ./...` from the `backend/` directory
- **Expected**: Compiler reports a type mismatch error; build exits non-zero; the offending line is identified in the error output
- **Priority**: P0

---

## TC-007: No uint(bizKey) or int64(teamID) casts exist in service and handler layers

- **Source**: Spec 需求目标 — 消除 uint/int64 混用导致的数据错误
- **Type**: CLI
- **Target**: cli/grep
- **Test ID**: cli/grep/no-uint-bizkey-or-int64-teamid-casts-in-service-handler
- **Pre-conditions**: Repository is at the post-migration state
- **Steps**:
  1. Run `grep -rn "uint(bizKey)\|uint(teamBizKey)\|int64(teamID)" backend/internal/service/ backend/internal/handler/`
- **Expected**: Command returns no matches (exit code 1 from grep, meaning zero matches found); zero occurrences of forced uint/int64 casts at the service/handler boundary
- **Priority**: P1

---

## Traceability

| TC ID  | Source                                    | Type | Target       | Priority |
|--------|-------------------------------------------|------|--------------|----------|
| TC-001 | Story 1 / AC-1                            | API  | api/progress | P0       |
| TC-002 | Story 2 / AC-1                            | API  | api/teams    | P0       |
| TC-003 | Spec Section 5.5 — bizKey 校验规则 rule 1 | API  | api/teams    | P1       |
| TC-004 | Spec Section 5.5 — bizKey 校验规则 rule 2 | API  | api/teams    | P1       |
| TC-005 | Spec Section 5.5 — bizKey 校验规则 rule 3 | API  | api/teams    | P1       |
| TC-006 | Story 3 / AC-1                            | CLI  | cli/build    | P0       |
| TC-007 | Spec 需求目标 — 消除 uint/int64 混用      | CLI  | cli/grep     | P1       |
