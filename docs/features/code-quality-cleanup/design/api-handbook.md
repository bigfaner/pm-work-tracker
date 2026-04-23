---
created: 2026-04-22
related: design/tech-design.md
---

# API Handbook: Code Quality Cleanup

## API Overview

This cleanup modifies existing API response shapes to fix contract mismatches between frontend and backend. No new endpoints are added. All changes are backward-compatible additions or field removals that the frontend already doesn't consume.

## Contract Changes

### SubItemVO — Add `statusName`

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/main-items/:itemId/sub-items` (and related endpoints)
**Change**: Add `statusName` field to match MainItemVO parity

#### Response (200) — Changed Fields

| Field | Type | Before | After |
|-------|------|--------|-------|
| `statusName` | string | absent | `"In Progress"` — human-readable status name |

**Impact**: Frontend `SubItem` type gains `statusName?: string`. No breaking change — field is additive.

---

### TeamMemberResp — Shape Alignment

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/members`
**Change**: Align frontend type to match actual backend flat `role` string

#### Current Frontend Type (Incorrect)

```typescript
interface TeamMemberResp {
  role: { name: string }  // nested object — backend never sends this
}
```

#### Aligned Type

```typescript
interface TeamMemberResp {
  role: string  // flat string — matches backend
}
```

---

### AdminTeam — PM Shape Alignment

**Method**: `GET`
**Path**: `/api/v1/admin/teams`
**Change**: Align frontend type to match actual backend flat `pmDisplayName` string

#### Current Frontend Type (Incorrect)

```typescript
interface AdminTeam {
  pm: { displayName: string }  // nested object — backend never sends this
}
```

#### Aligned Type

```typescript
interface AdminTeam {
  pmDisplayName: string  // flat string — matches backend
}
```

---

### changeMainItemStatusApi — Return Type Fix

**Method**: `PUT`
**Path**: `/api/v1/teams/:teamId/main-items/:itemId/status`
**Change**: Frontend return type corrected to match actual backend response

#### Before (Frontend Type — Incorrect)

```typescript
function changeMainItemStatusApi(...): Promise<{ status: string }>
```

#### After (Aligned)

```typescript
function changeMainItemStatusApi(...): Promise<MainItemVO>
```

**Impact**: Callers that destructure `{ status }` must now use `item.status` from the returned `MainItemVO`.

---

## Removed API Functions (Dead Code)

These frontend API functions are defined but never called from production code:

| Function | File | Action |
|----------|------|--------|
| `archiveMainItemApi` | `api/mainItems.ts` | Delete |
| `assignSubItemApi` | `api/subItems.ts` | Delete |
| `getItemPoolApi` | `api/itemPool.ts` | Delete — superseded by paginated list endpoint |
| `correctCompletionApi` | `api/progress.ts` | Delete — never called from production code |

## Removed Types (Dead Code)

| Type | File | Action |
|------|------|--------|
| `WeeklyViewResp` | `types/index.ts` | Delete |
| `WeeklyGroup` | `types/index.ts` | Delete |
| `SubItemWithProgress` | `types/index.ts` | Delete |
| `TeamMember` | `types/index.ts` | Delete — never imported outside its own test file |

## Removed Exports (Dead Code)

| Export | File | Action |
|--------|------|--------|
| `ApiSuccessEnvelope<T>` | `api/client.ts` | Remove export — never imported anywhere |
| `ApiErrorEnvelope` | `api/client.ts` | Remove export — never imported anywhere |

## Removed Fields (Contract Fixes)

| Type | Field | Reason |
|------|-------|--------|
| `MainItem` | `delayCount` | Backend never sends this field |
| `SubItem` | `delayCount` | Backend never sends this field |
| `MainItem` | `archivedAt` | Backend never sends this field |
| `MainItem` | `isKeyItem` | Removed from frontend type only — backend retains `is_key_item` DB column and repository filter; frontend never sends or consumes this field |

## Error Codes

No changes to error codes or error handling behavior.
