---
created: 2026-04-27
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: Decision Log

## Overview

Add a DecisionLog feature to Main Items with draft→published lifecycle management. The feature follows the ProgressRecord pattern: own BizKey (no BaseModel embedding), append-only after publishing, and team-scoped access.

**Key decisions:**
- Reuse `main_item:update` permission for write operations (per PRD)
- Routes nested under main-items resource group
- Separate API query for decision logs (not bundled in main item response)
- Tags stored as JSON array string in TEXT column, parsed to `[]string` in VO layer
- Add `TeamKey` field for query consistency (ProgressRecord has it)

## Architecture

### Layer Placement

Full-stack feature touching all layers: Model → Repository → Service → Handler → VO/DTO → Router (backend) + Types → API → Components → Page (frontend).

### Component Diagram

```
Frontend                              Backend
┌─────────────────────────┐          ┌──────────────────────────────┐
│ MainItemDetailPage      │          │ Router                       │
│  ├─ DecisionLogCard     │──GET───> │  ├─ DecisionLogHandler.List  │
│  └─ DecisionLogDialog   │──POST──> │  ├─ DecisionLogHandler.Create│
│                         │──PUT───> │  ├─ DecisionLogHandler.Update│
│                         │─PATCH──> │  └─ DecisionLogHandler.Publish│
│  api/decisionLogs.ts    │          │      └─ DecisionLogService   │
│  types/index.ts         │          │          └─ DecisionLogRepo  │
└─────────────────────────┘          │              └─ DecisionLog  │
                                     └──────────────────────────────┘
```

### Dependencies

| Dependency | Purpose | New? |
|-----------|---------|------|
| `model.DecisionLog` | GORM model for `pmw_decision_logs` | Yes |
| `encoding/json` | Tags serialization (stdlib) | No |
| `pkg/snowflake` | BizKey generation | No |
| `pkg/handler.ResolveBizKey` | URL param → internal ID | No |
| `middleware.GetTeamID/GetUserID` | Auth context extraction | No |
| `@radix-ui/react-dialog` | Dialog primitive (already installed) | No |

## Interfaces

### Repository: DecisionLogRepo

```go
type DecisionLogRepo interface {
    Create(ctx context.Context, log *model.DecisionLog) error
    FindByID(ctx context.Context, id uint) (*model.DecisionLog, error)
    FindByBizKey(ctx context.Context, bizKey int64) (*model.DecisionLog, error)
    ListByItem(ctx context.Context, mainItemID uint, userID uint) ([]model.DecisionLog, error)
    Update(ctx context.Context, log *model.DecisionLog) error
}
```

- `ListByItem` applies draft visibility filter: published decisions for all + drafts for `userID` only.
- `Update` is used for draft editing (service layer enforces status/ownership checks).

### Service: DecisionLogService

```go
type DecisionLogService interface {
    Create(ctx context.Context, mainItemID uint, userID uint, req DecisionLogCreateReq) (*model.DecisionLog, error)
    Update(ctx context.Context, bizKey int64, userID uint, req DecisionLogUpdateReq) (*model.DecisionLog, error)
    Publish(ctx context.Context, bizKey int64, userID uint) (*model.DecisionLog, error)
    List(ctx context.Context, mainItemID uint, userID uint) ([]model.DecisionLog, error)
}
```

- `Create` — validates main item exists, generates BizKey via snowflake, sets `CreatedBy` from `userID`.
- `Update` — enforces draft-only + owner-only; returns `ErrForbidden` if either check fails.
- `Publish` — enforces draft-only + owner-only; transitions status to "published".
- `List` — returns published decisions + current user's drafts, ordered by `CreateTime` DESC.

### Handler: DecisionLogHandler

```go
type DecisionLogHandler struct {
    svc          service.DecisionLogService
    userRepo     repository.UserRepo
    mainItemRepo repository.MainItemRepo
}

func NewDecisionLogHandler(svc service.DecisionLogService, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo) *DecisionLogHandler
```

Methods: `Create(c *gin.Context)`, `Update(c *gin.Context)`, `Publish(c *gin.Context)`, `List(c *gin.Context)`.

### Frontend API: decisionLogs

```ts
function listDecisionLogsApi(teamBizKey: string, mainBizKey: string): Promise<DecisionLog[]>
function createDecisionLogApi(teamBizKey: string, mainBizKey: string, req: CreateDecisionLogReq): Promise<DecisionLog>
function updateDecisionLogApi(teamBizKey: string, mainBizKey: string, bizKey: string, req: UpdateDecisionLogReq): Promise<DecisionLog>
function publishDecisionLogApi(teamBizKey: string, mainBizKey: string, bizKey: string): Promise<DecisionLog>
```

### Frontend Components

- `DecisionLogCard` — Card component rendering timeline list with lazy loading
- `DecisionLogDialog` — Dialog for creating/editing decisions (dual-mode: new + edit draft)

## Data Models

### Model: DecisionLog

```go
type DecisionLog struct {
    ID          uint      `gorm:"primarykey;autoIncrement" json:"-"`
    BizKey      int64     `gorm:"not null" json:"bizKey"`
    MainItemKey int64     `gorm:"not null;index" json:"mainItemKey"`
    TeamKey     int64     `gorm:"not null" json:"teamKey"`
    Category    string    `gorm:"type:varchar(20);not null" json:"category"`
    Tags        string    `gorm:"type:text;not null;default:''" json:"tags"`
    Content     string    `gorm:"type:text;not null" json:"content"`
    LogStatus   string    `gorm:"type:varchar(10);not null;default:'draft'" json:"logStatus"`
    CreatedBy   int64     `gorm:"not null" json:"createdBy"`
    CreateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
    UpdateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updateTime"`
}

func (DecisionLog) TableName() string { return "pmw_decision_logs" }
```

Field notes:
- No BaseModel embedding — append-only model, no soft-delete needed
- `Tags` — JSON array stored as TEXT (e.g., `'["缓存策略","性能优化"]'`). Parsed to `[]string` in VO.
- `LogStatus` — enum: `"draft"` | `"published"`. Column `log_status` (avoids MySQL reserved word `status`). Follows `item_status`/`pool_status` convention.
- `MainItemKey` — FK to main item's BizKey (named explicitly, not `ItemKey`, to avoid ambiguity with StatusHistory).
- `TeamKey` — added for team-scoped queries and consistency with ProgressRecord.

### VO: DecisionLogVO

```go
type DecisionLogVO struct {
    BizKey      string   `json:"bizKey"`
    MainItemKey string   `json:"mainItemKey"`
    Category    string   `json:"category"`
    Tags        []string `json:"tags"`
    Content     string   `json:"content"`
    LogStatus   string   `json:"logStatus"`
    CreatedBy   string   `json:"createdBy"`
    CreatorName string   `json:"creatorName"`
    CreateTime  string   `json:"createTime"`
    UpdateTime  string   `json:"updateTime"`
}
```

- `Tags` converted from JSON string → `[]string` in `NewDecisionLogVO()`.
- All int64 BizKeys formatted as strings via `pkg.FormatID()`.
- `CreatorName` resolved via batch user lookup (same pattern as `buildProgressRecordVOs`).
- Time fields formatted as RFC3339.

### DTO: Request Types

```go
type DecisionLogCreateReq struct {
    Category string `json:"category" binding:"required,oneof=technical resource requirement schedule risk other"`
    Tags     string `json:"tags"`
    Content  string `json:"content" binding:"required,max=2000"`
    Status   string `json:"logStatus" binding:"required,oneof=draft published"`
}

type DecisionLogUpdateReq struct {
    Category string `json:"category" binding:"required,oneof=technical resource requirement schedule risk other"`
    Tags     string `json:"tags"`
    Content  string `json:"content" binding:"required,max=2000"`
}
```

- `Tags` arrives as JSON string from frontend (already serialized). No backend validation on individual tag length — frontend handles this.

### Frontend Type: DecisionLog

```ts
interface DecisionLog {
  bizKey: string
  mainItemKey: string
  category: string
  tags: string[]
  content: string
  logStatus: 'draft' | 'published'
  createdBy: string
  creatorName: string
  createTime: string
  updateTime: string
}

interface CreateDecisionLogReq {
  category: string
  tags: string[]
  content: string
  logStatus: 'draft' | 'published'
}

interface UpdateDecisionLogReq {
  category: string
  tags: string[]
  content: string
}
```

## Error Handling

### Error Types & Codes

| Error Code | Condition | HTTP Status |
|-----------|-----------|-------------|
| `VALIDATION_ERROR` | Missing/invalid category, content, or status | 400 |
| `DECISION_LOG_NOT_FOUND` | Decision log BizKey doesn't exist | 404 |
| `FORBIDDEN` | Edit/publish a published decision; edit another user's draft | 403 |
| `ITEM_NOT_FOUND` | Parent main item doesn't exist | 404 |

### Propagation Strategy

- **Handler**: catches all errors, calls `apperrors.RespondError(c, err)`.
- **Service**: returns domain errors (`ErrForbidden`, `ErrItemNotFound`). Uses `apperrors.MapNotFound()` for repo not-found errors.
- **Repository**: returns `ErrNotFound` for missing records (via `repo.FindByID` generic helper).

No new sentinel errors needed — reuse `ErrForbidden`, `ErrItemNotFound`, `ErrNotFound`, `ErrValidation`. Add `ErrDecisionLogNotFound` as a domain-specific alias for clarity.

## Cross-Layer Data Map

| Field | Storage (SQL) | Backend Model | API (JSON) | Frontend Type | Validation |
|-------|--------------|---------------|------------|---------------|------------|
| bizKey | INTEGER, NOT NULL, UNIQUE | int64 | string (FormatID) | string | system-generated |
| mainItemKey | INTEGER, NOT NULL, INDEX | int64 | string (FormatID) | string | URL path param |
| teamKey | INTEGER, NOT NULL | int64 | omitted | — | middleware injection |
| category | VARCHAR(20), NOT NULL | string | string | string | required, enum of 6 |
| tags | TEXT, DEFAULT '' | string (JSON) | []string (parsed) | string[] | optional, each ≤20 chars |
| content | TEXT, NOT NULL | string | string | string | required, ≤2000 chars |
| logStatus | VARCHAR(10), NOT NULL, DEFAULT 'draft' (column: log_status) | string | string | 'draft' \| 'published' | required on create |
| createdBy | INTEGER, NOT NULL | int64 | string (FormatID) | string | system (from auth) |
| creatorName | — (joined) | — | string | string | — |
| createTime | DATETIME, DEFAULT NOW | time.Time | string (RFC3339) | string | system-generated |
| updateTime | DATETIME, AUTO UPDATE | time.Time | string (RFC3339) | string | system-auto |

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| Repository | Unit | go test + mock DB | CRUD operations, ListByItem draft filter | 80% |
| Service | Unit | go test + mock repo | Business rules: draft-only edit, owner-only, status transitions | 90% |
| Handler | Unit | httptest + mock service | Request binding, permission checks, response format | 80% |
| Frontend API | Unit | vitest + mock client | API call patterns (endpoint, method, payload) | 80% |
| Frontend Components | Unit | vitest + testing-library | Render states, form interactions, expand/collapse | 80% |

### Key Test Scenarios

1. **Create draft** — valid request → 201, status=draft
2. **Create published** — valid request → 201, status=published
3. **Edit draft** — owner edits → 200, content updated
4. **Edit draft (not owner)** — different user → 403
5. **Edit published** — attempt to edit → 403
6. **Publish draft** — owner publishes → 200, status=published
7. **List — visibility** — user A sees own drafts + all published; user B doesn't see user A's drafts
8. **List — ordering** — results sorted by createTime DESC
9. **Invalid category** — out-of-enum → 400
10. **Empty content** — blank content → 400

### Overall Coverage Target

80%

## Security Considerations

### Threat Model

| Threat | Risk | Mitigation |
|--------|------|------------|
| Unauthorized draft access | User sees another's draft | Service filters by `CreatedBy` for drafts |
| Published decision tampering | User modifies published record | Service enforces draft-only edit; route middleware checks `main_item:update` |
| Cross-team access | User accesses another team's decisions | Team scope middleware + `TeamKey` in model |
| Content injection | Malicious content in tags/content | No HTML rendering; content displayed as text |

### Mitigations

- Write operations gated by `main_item:update` permission middleware at route level
- Draft ownership verified in service layer (not just frontend hiding)
- Team scope verified via middleware + `TeamKey` field on model
- Content stored and served as plain text (no HTML parsing)

## PRD Coverage Map

| PRD Requirement / AC | Design Component | Interface / Model |
|----------------------|------------------|-------------------|
| Story 1: Record & publish decision | DecisionLogHandler.Create | POST /main-items/:mainId/decision-logs |
| Story 1 AC: Published decision visible to all | DecisionLogService.List | ListByItem (no CreatedBy filter for published) |
| Story 1 AC: Published decision immutable | DecisionLogService.Update | draft-only check |
| Story 2: Save as draft | DecisionLogHandler.Create (status=draft) | DecisionLogCreateReq.Status |
| Story 2 AC: Draft visible to creator only | DecisionLogService.List | ListByItem with userID filter for drafts |
| Story 2 AC: Edit draft | DecisionLogHandler.Update | PUT /main-items/:mainId/decision-logs/:id |
| Story 2 AC: Publish from draft | DecisionLogHandler.Publish | PATCH /main-items/:mainId/decision-logs/:id/publish |
| Story 3: View timeline (reverse chronological) | DecisionLogCard + DecisionLogService.List | ORDER BY create_time DESC |
| Story 3 AC: 80-char summary + expand | DecisionLogCard | Frontend truncation logic |
| Story 4: Draft invisible to other users | DecisionLogService.List | CreatedBy filter on drafts |
| Story 4 AC: Published edit → 403 | DecisionLogService.Update/Publish | Status check + ErrForbidden |
| Story 4 AC: No permission → no add button | PermissionGuard | `main_item:update` check |
| PRD 5.1: Pagination (20/page) | DecisionLogCard | Frontend lazy loading |
| PRD 5.2: Permission = main_item:update | Route middleware | `deps.perm("main_item:update")` |
| PRD 5.3: Category dropdown (6 options) | DecisionLogDialog | Frontend enum binding |
| PRD 5.3: Tags input with suggestions | DecisionLogDialog | Frontend extracts tags from existing logs |
| PRD 5.4: New DB table | DecisionLog model | pmw_decision_logs |
| PRD 5.4: Nested routes | Router | main-items/:mainId/decision-logs |

## Open Questions

None. All PRD requirements mapped to specific design components.

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Bundle decision logs in main item response | Fewer API calls | Larger response, can't lazy-load, pagination harder | Separate endpoint enables lazy loading and cleaner separation |
| Separate permission codes (decision_log:create, etc.) | Fine-grained access control | Overhead for admin setup, PRD explicitly says reuse | PRD specifies `main_item:update` reuse |
| Tags as separate relation table | Normalized storage, easier tag queries | Join overhead, over-engineering for append-only data | Tags are display-only, not a query dimension |

### References

- PRD: `docs/features/decision-log/prd/prd-spec.md`
- Proposal: `docs/proposals/decision-log/proposal.md`
- ProgressRecord pattern: `backend/internal/model/progress_record.go`
