---
created: "2026-05-04"
tags: [architecture, data-model, interface]
---

# New Feature Must Follow Existing BizKey Conventions, Not Default to Internal IDs

## Problem

DecisionLog handler used `middleware.GetUserID(c)` (returns `uint` internal ID) to populate `CreatedBy`, while all other models in the codebase use `middleware.GetUserBizKey(c)` (returns `int64` snowflake bizKey) for user references. This caused:

1. `CreatedBy` stored as `1` (internal admin user ID) instead of the user's snowflake bizKey
2. `batchLookupCreatorNames` called `FindByBizKeys` with internal IDs → no matches → `creatorName` always empty
3. 2 e2e test failures (TC-016, TC-028) checking `creatorName` field

## Root Cause

Causal chain (4 levels):

1. **Symptom**: `creatorName` returns empty string in API responses
2. **Direct cause**: `batchLookupCreatorNames` receives internal ID `1` but calls `FindByBizKeys` which expects snowflake IDs
3. **Code cause**: Handler uses `middleware.GetUserID(c)` instead of `middleware.GetUserBizKey(c)`
4. **Root cause**: The task-executor agent that wrote the handler/service code defaulted to `GetUserID` (the more "obvious" function name) without checking how existing handlers reference users. The tech design doc said "sets `CreatedBy` from `userID`" — ambiguous wording that the agent interpreted as internal ID rather than bizKey

**Why the design doc was ambiguous**: The doc used `userID` generically, but the codebase has two distinct functions: `GetUserID()` → `uint` internal, `GetUserBizKey()` → `int64` external. The naming convention `AuthorKey`, `SubmitterKey`, `AssigneeKey` in other models makes the intent clear (these are bizKeys), but the decision-log design used `CreatedBy` which doesn't follow the `*Key` naming convention.

## Solution

Changed all layers from `uint` (internal ID) to `int64` (bizKey):
- Handler: `GetUserID()` → `GetUserBizKey()`
- Service interface: `userID uint` → `userBizKey int64`
- Service impl: `int64(userID)` → `userBizKey` (direct, no cast)
- Repo interface/impl: `userID uint` → `userBizKey int64`

## Reusable Pattern

### When building a new entity that references users, ALWAYS follow these rules:

1. **Use `GetUserBizKey(c)`, never `GetUserID(c)`** for storing user references in business data. Internal IDs should only be used for auth/middleware internals (session lookup, permission checks).

2. **Name user-reference fields with `*Key` suffix** (e.g., `AuthorKey`, `SubmitterKey`, `CreatorKey` — NOT `CreatedBy`). The `*Key` suffix makes it unambiguous that the field stores a snowflake bizKey, matching the project convention.

3. **Tech design docs should specify `userBizKey` not `userID`** when describing how user references are set. The word "userID" is ambiguous in this codebase.

4. **Check existing models before writing new ones**: `ProgressRecord.AuthorKey`, `ItemPool.SubmitterKey`, `MainItem.AssigneeKey` all use `int64` bizKey. Any new user-reference field should follow the same pattern.

### Quick reference for handler writers:

```go
// ❌ Wrong — stores internal uint ID, breaks bizKey lookups
userID := middleware.GetUserID(c)
log.CreatedBy = int64(userID)

// ✅ Correct — stores snowflake bizKey, works with FindByBizKeys
userBizKey := middleware.GetUserBizKey(c)
log.CreatorKey = userBizKey
```

## Example

Existing correct patterns in the codebase:

| Model | Field | Type | Set by |
|-------|-------|------|--------|
| ProgressRecord | AuthorKey | int64 | `GetUserBizKey(c)` |
| ItemPool | SubmitterKey | int64 | `GetUserBizKey(c)` |
| MainItem | AssigneeKey | *int64 | request body (bizKey) |
| SubItem | AssigneeKey | *int64 | request body (bizKey) |

DecisionLog should have followed: `CreatorKey int64` set by `GetUserBizKey(c)`.

## Related Files

- `backend/internal/model/progress_record.go` — correct pattern: `AuthorKey int64`
- `backend/internal/model/item_pool.go` — correct pattern: `SubmitterKey int64`
- `backend/internal/model/decision_log.go` — was wrong: `CreatedBy int64` storing internal IDs
- `backend/internal/middleware/auth.go` — `GetUserID()` vs `GetUserBizKey()` distinction
- `docs/features/decision-log/design/tech-design.md` — ambiguous "userID" wording
