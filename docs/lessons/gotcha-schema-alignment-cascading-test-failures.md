# Schema Alignment: Type Changes Cascade Across All Tests

## Problem

Task 2.2 was scoped to Team + TeamMember table alignment. But when running `go test ./...` to verify, tests in `main_item_service_test.go`, `progress_service_test.go`, and `view_service_test.go` also failed — files completely unrelated to Team.

## Root Cause

Schema alignment changes field types globally (e.g., `uint` → `int64` for all FK/BizKey fields). These type changes affect:
- `ProposerKey`, `SubmitterKey`, `AssigneeKey`, `ChangedBy`, `RoleKey`, `AssigneeID` — all changed from `uint` to `int64`
- Field rename: `status` → `item_status` in GORM update maps

Test assertions written before the alignment still use `uint(10)` or `updatedFields["status"]`. They compile fine (Go allows `uint` vs `int64` comparison via `assert.Equal`) but fail at runtime.

## Solution

When a schema alignment task changes field types or field names, **scan all test files for the old type/key before marking the task done**:

```bash
# Find all uint assertions on fields that changed to int64
grep -rn 'uint([0-9]*)' backend/internal/service/ --include='*_test.go'

# Find old field key references
grep -rn 'updatedFields\["status"\]' backend/internal/service/ --include='*_test.go'
```

Fix all occurrences in one pass, not just the files in the current task scope.

## Key Takeaway

Schema alignment tasks have **blast radius beyond their stated scope**. A "Team table" task that changes `uint→int64` will break any test that asserts on an `int64` field using `uint(N)`. Always run the full test suite and fix all type-mismatch failures before recording the task as complete — even if the failures are in unrelated packages.
