---
feature: decision-log
source: design/tech-design.md, design/api-handbook.md
---

# Technical Specifications: Decision Log

## TS-1: Data Model

### Table: `pmw_decision_logs`

| Column | Type | Constraints | Notes |
|--------|------|------------|-------|
| id | INTEGER | PK, AUTOINCREMENT | Internal ID |
| biz_key | INTEGER | NOT NULL, UNIQUE | Snowflake ID, exposed as string via FormatID |
| main_item_key | INTEGER | NOT NULL, INDEX | FK to main item's BizKey |
| team_key | INTEGER | NOT NULL | Team scope for query-level enforcement |
| category | VARCHAR(20) | NOT NULL | Enum: technical/resource/requirement/schedule/risk/other |
| tags | TEXT | NOT NULL, DEFAULT '' | JSON array string (e.g., `'["cache","perf"]'`) |
| content | TEXT | NOT NULL | Decision content, max 2000 chars |
| log_status | VARCHAR(10) | NOT NULL, DEFAULT 'draft' | Enum: draft/published. Named `log_status` (not `status`) to avoid reserved word and follow prefixed-column convention |
| created_by | INTEGER | NOT NULL | Creator's user BizKey |
| create_time | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| update_time | DATETIME | NOT NULL, DEFAULT CURRENT_TIMESTAMP, AUTO UPDATE | |

**Design choices:**
- No `BaseModel` embedding -- append-only model has no soft-delete or generic ID fields
- No `deleted_at` -- published records are immutable, drafts are only editable by owner
- `TeamKey` added for team-scoped queries, consistent with ProgressRecord pattern
- `MainItemKey` (not `ItemKey`) to avoid ambiguity with StatusHistory

## TS-2: Route Structure

Routes are nested under the `teamsGroup` (which applies `AuthMiddleware` + `TeamScopeMiddleware`):

```
GET    /api/v1/teams/:teamId/main-items/:mainId/decision-logs           → List
POST   /api/v1/teams/:teamId/main-items/:mainId/decision-logs           → Create  (perm: main_item:update)
PUT    /api/v1/teams/:teamId/main-items/:mainId/decision-logs/:logId    → Update  (perm: main_item:update)
PATCH  /api/v1/teams/:teamId/main-items/:mainId/decision-logs/:logId/publish → Publish (perm: main_item:update)
```

Read (List) requires team membership only. Write operations are gated by `deps.perm("main_item:update")`.

## TS-3: Repository Interface

```go
type DecisionLogRepo interface {
    Create(ctx context.Context, log *model.DecisionLog) error
    FindByID(ctx context.Context, id uint) (*model.DecisionLog, error)
    FindByBizKey(ctx context.Context, bizKey int64) (*model.DecisionLog, error)
    ListByItem(ctx context.Context, mainItemID uint, userID uint, offset, limit int) ([]model.DecisionLog, int64, error)
    Update(ctx context.Context, log *model.DecisionLog) error
}
```

`ListByItem` applies draft visibility filter: `WHERE (log_status = 'published' OR created_by = ?)`.

## TS-4: Service Interface

```go
type DecisionLogService interface {
    Create(ctx context.Context, mainItemID uint, userID uint, req DecisionLogCreateReq) (*model.DecisionLog, error)
    Update(ctx context.Context, bizKey int64, userID uint, req DecisionLogUpdateReq) (*model.DecisionLog, error)
    Publish(ctx context.Context, bizKey int64, userID uint) (*model.DecisionLog, error)
    List(ctx context.Context, mainItemID uint, userID uint, page dto.Pagination) (*dto.PageResult[model.DecisionLog], error)
}
```

- `Create`: validates main item exists, generates BizKey via snowflake, sets `CreatedBy` from `userID`
- `Update`: enforces draft-only + owner-only; returns `ErrForbidden` if either check fails
- `Publish`: enforces draft-only + owner-only; transitions status to "published"
- `List`: returns published + current user's drafts, ordered by `CreateTime` DESC, paginated

## TS-5: Request DTOs

```go
type DecisionLogCreateReq struct {
    Category  string   `json:"category" binding:"required,oneof=technical resource requirement schedule risk other"`
    Tags      []string `json:"tags" binding:"dive,max=20"`
    Content   string   `json:"content" binding:"required,max=2000"`
    LogStatus string   `json:"logStatus" binding:"required,oneof=draft published"`
}

type DecisionLogUpdateReq struct {
    Category string   `json:"category" binding:"required,oneof=technical resource requirement schedule risk other"`
    Tags     []string `json:"tags" binding:"dive,max=20"`
    Content  string   `json:"content" binding:"required,max=2000"`
}
```

Tags: `[]string` in DTO, serialized to JSON string for DB storage, parsed back to `[]string` in VO response.

## TS-6: VO (Response Shape)

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

- All int64 BizKeys formatted as strings via `pkg.FormatID()`
- `CreatorName` resolved via batch user lookup (same pattern as `buildProgressRecordVOs`)
- Time fields formatted as RFC3339

## TS-7: Frontend Types

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

## TS-8: Frontend API Module

File: `frontend/src/api/decisionLogs.ts`

```ts
function listDecisionLogsApi(teamBizKey: string, mainBizKey: string, page?: number, pageSize?: number): Promise<PageResult<DecisionLog>>
function createDecisionLogApi(teamBizKey: string, mainBizKey: string, req: CreateDecisionLogReq): Promise<DecisionLog>
function updateDecisionLogApi(teamBizKey: string, mainBizKey: string, bizKey: string, req: UpdateDecisionLogReq): Promise<DecisionLog>
function publishDecisionLogApi(teamBizKey: string, mainBizKey: string, bizKey: string): Promise<DecisionLog>
```

## TS-9: Frontend Components

- **DecisionLogCard**: Card component rendering timeline list with lazy loading (20/page, loads next page on scroll)
- **DecisionLogDialog**: Dialog for creating/editing decisions (dual-mode: new + edit draft). Built on `@radix-ui/react-dialog`.

## TS-10: Error Codes

| Code | HTTP Status | Condition |
|------|-------------|-----------|
| `VALIDATION_ERROR` | 400 | Missing/invalid category, content, or status |
| `FORBIDDEN` | 403 | Edit/publish a published decision; edit another user's draft |
| `ITEM_NOT_FOUND` | 404 | Parent main item doesn't exist |
| `DECISION_LOG_NOT_FOUND` | 404 | Decision log BizKey doesn't exist |

Add `ErrDecisionLogNotFound` as a domain-specific alias. Reuse `ErrForbidden`, `ErrItemNotFound`, `ErrValidation`.

## TS-11: Security

- Write operations gated by `deps.perm("main_item:update")` middleware
- Draft ownership verified in service layer via `CreatedBy == userID` (not just frontend hiding)
- Team scope via `TeamScopeMiddleware` on `teamsGroup` + `TeamKey` field on model
- Content stored as plain text; React renders as text nodes (`dangerouslySetInnerHTML` never used)
- Tag length validated at backend DTO binding layer via `binding:"dive,max=20"`
