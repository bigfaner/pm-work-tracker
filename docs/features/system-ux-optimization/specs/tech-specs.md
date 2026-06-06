---
feature: "system-ux-optimization"
generated: "2026-06-05"
status: draft
---

# Technical Specifications: System UX Optimization Batch

## API Endpoints

### TECH-001: Delete Main Item Endpoint

**Requirement**: `DELETE /teams/:teamId/main-items/:itemId` with `main_item:delete` permission. Returns 200 with `{"message": "ok"}` on success. Error responses: 404 (NOT_FOUND), 403 (FORBIDDEN).
**Scope**: [CROSS]
**Source**: Tech Design Interface 1

### TECH-002: Delete Sub Item Endpoint

**Requirement**: `DELETE /teams/:teamId/sub-items/:subId` with `sub_item:delete` permission. Returns 200 with `{"message": "ok"}` on success. Error responses: 404 (NOT_FOUND), 403 (FORBIDDEN).
**Scope**: [CROSS]
**Source**: Tech Design Interface 2

### TECH-003: Move Sub Item Endpoint

**Requirement**: `PUT /teams/:teamId/sub-items/:subId/move` with `sub_item:update` permission. Request body: `{ "targetMainItemBizKey": "729384756" }`. Response 200: `{ "newSubCode": "T001-007", "mainItemBizKey": "729384756" }`. Error responses: 400 (BAD_REQUEST for closed target or same-main-item), 404 (NOT_FOUND).
**Scope**: [CROSS]
**Source**: Tech Design Interface 3

## Error Handling

### TECH-004: New Business Error Codes

**Requirement**: Two new business errors added to the error code registry:
- `ErrTargetClosed`: Code `BAD_REQUEST`, Message "目标主事项已关闭", HTTP 400
- `ErrSameMainItem`: Code `BAD_REQUEST`, Message "不能移动到同一主事项", HTTP 400

These follow the existing `AppError` pattern in `backend/internal/pkg/errors/errors.go` and are mapped via `apperrors.RespondError`.
**Scope**: [CROSS]
**Source**: Tech Design "Error Handling"

## Permission Codes

### TECH-005: New Permission Codes for Delete Operations

**Requirement**: Two new permission codes added as seed data to `pmw_role_permissions`:
- `main_item:delete` -- controls main item deletion
- `sub_item:delete` -- controls sub item deletion

Both codes follow the `resource:action` format convention. Added to the PM preset role during seed sync. Registered in Go code `internal/pkg/permissions/codes.go`.
**Scope**: [CROSS]
**Source**: Tech Design "Architecture", PRD #3

## Filter Patterns

### TECH-006: Multi-Select Status Filter

**Requirement**: Status filter parameters accept comma-separated values (`?status=progressing,blocking`) parsed as `[]string` via `form:"status"` tag. Applied to MainItemFilter and GanttFilter. Single-value queries continue to work (one-element slice).
**Scope**: [CROSS]
**Source**: Tech Design Interface 4, Interface 7

### TECH-007: Filter Penetration Response DTO

**Requirement**: Main item response DTOs extended with optional fields:
- `matchType`: `"direct" | "indirect"` (omitted when no filter applied)
- `matchedSubItemIds`: `string[]` (only present when `matchType=indirect`)

Backend DTO: `MainItemMatchInfo` struct. Frontend type: optional fields on MainItem interface.
**Scope**: [LOCAL]
**Source**: Tech Design Interface 4

## Data Patterns

### TECH-008: Completion Recalculation After Structural Changes

**Requirement**: After sub-item deletion or move, the parent main item's `completion_pct` is recalculated synchronously within the same transaction. For move, both source and target main items are recalculated.
**Scope**: [CROSS]
**Source**: Tech Design Interface 1, Interface 2, Interface 3

### TECH-009: Sub-Item List Reverse Order

**Requirement**: `SubItemRepo.ListByMainItem` returns results ordered by `id DESC` (newest first). Previously was `id ASC`.
**Scope**: [LOCAL]
**Source**: Tech Design Interface 5

## Security

### TECH-010: Permission-Guarded UI Buttons

**Requirement**: Destructive action buttons (delete, move) are wrapped in `<PermissionGuard code="...">` component. This is defensive UI only -- backend middleware enforces actual authorization.
**Scope**: [CROSS]
**Source**: Tech Design "Security Considerations"

## Frontend

### TECH-011: Status Error Alert Display

**Requirement**: Status transition errors display as persistent Alert component (not auto-dismissing tooltip). Alert shows the `message` field from the backend 422 response. Alert dismisses on user action or successful retry.
**Scope**: [LOCAL]
**Source**: PRD #1, Tech Design Interface 11

### TECH-012: Form Reset on Close/Success

**Requirement**: All create/convert forms reset all fields on explicit close/cancel and on successful submission. Fields are preserved on submission failure to allow correction.
**Scope**: [LOCAL]
**Source**: PRD #6, #7, Tech Design Interface 11

## Performance

### TECH-013: Filter Penetration Response Time

**Requirement**: Filter penetration queries must respond in ≤500ms for 1000 main items + 5000 sub items. Current implementation uses in-memory filtering at ViewService layer (no pagination).
**Scope**: [CROSS]
**Source**: PRD "Performance Requirements"

## ViewService

### TECH-014: Weekly View Inactive Item Filtering

**Requirement**: ViewService.WeeklyComparison applies post-filter to remove terminal main items with no activity in the current or previous week. Activity detection uses: (a) status_history records in time range, (b) sub-item create/update times, (c) main item update time. Requires new `StatusHistoryRepo.ListByItemKeysInRange` method.
**Scope**: [LOCAL]
**Source**: Tech Design Interface 8
