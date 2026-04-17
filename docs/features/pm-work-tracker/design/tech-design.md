---
created: 2026-04-17
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: PM Work Tracker

## Overview

PM Work Tracker is a web-based work item tracking system built as a single deployable product with a React SPA frontend and a Go/Gin REST API backend. The two are deployed as separate artifacts but live in the same monorepo under `frontend/` and `backend/`.

The core data hierarchy is **MainItem → SubItem → ProgressRecord**. A parallel **ItemPool** flow lets members propose work that PM reviews before it becomes a SubItem. Four read-only views (Item, Weekly, Gantt, Table) are computed at query time — no denormalized view tables.

Deployment topology:

```
Browser (React SPA)
      │  HTTPS
      ▼
  Nginx / CDN  ──── static assets
      │  /api/*  reverse proxy
      ▼
  Go/Gin API Server  (single binary)
      │
      ▼
  SQLite (dev) / MySQL (prod)   ← GORM dual-driver
```

No message queue, no cache layer, no background workers in v1. Progress rollup is computed synchronously on every SubItem progress append.

---

## Architecture

### Layer Placement

```
┌─────────────────────────────────────────────────────┐
│  Frontend  (React 18 + TypeScript + Vite)           │
│  Pages / Components / Hooks / API Client            │
└────────────────────┬────────────────────────────────┘
                     │ HTTP/JSON  (JWT in Authorization header)
┌────────────────────▼────────────────────────────────┐
│  Transport Layer  (Gin router + middleware)          │
│  AuthMiddleware · TeamScopeMiddleware · RBAC         │
└────────────────────┬────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────┐
│  Service Layer  (domain logic, business rules)      │
│  AuthService · TeamService · MainItemService        │
│  SubItemService · ProgressService · ItemPoolService │
│  ReportService · ViewService                        │
└────────────────────┬────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────┐
│  Repository Layer  (GORM, DB-agnostic queries)      │
│  UserRepo · TeamRepo · MainItemRepo · SubItemRepo   │
│  ProgressRepo · ItemPoolRepo                        │
└────────────────────┬────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────┐
│  Database  SQLite (dev) / MySQL (prod)              │
└─────────────────────────────────────────────────────┘
```

### Component Diagram

```
frontend/
├── src/
│   ├── pages/          # route-level components
│   │   ├── Login/
│   │   ├── ItemView/
│   │   ├── WeeklyView/
│   │   ├── GanttView/
│   │   ├── TableView/
│   │   ├── ItemPool/
│   │   ├── Report/
│   │   └── Admin/
│   ├── components/     # shared UI components
│   ├── hooks/          # data-fetching hooks (React Query)
│   ├── api/            # typed API client (axios)
│   ├── store/          # auth state (Zustand)
│   └── types/          # shared TypeScript types

backend/
├── cmd/server/         # main entry point
├── internal/
│   ├── handler/        # Gin handlers (HTTP layer)
│   ├── service/        # business logic interfaces + implementations
│   ├── repository/     # GORM repository interfaces + implementations
│   ├── model/          # GORM model structs
│   ├── middleware/      # JWT auth, RBAC, team scope
│   ├── dto/            # request/response DTOs
│   └── pkg/
│       ├── jwt/        # token sign/verify
│       ├── errors/     # domain error types
│       └── report/     # Markdown report generator
├── config/             # env-based config loader
└── migrations/         # SQL migration files
```

### Dependencies

Internal:
- All cross-domain calls go through service interfaces, never directly between handlers or repositories of different domains.

External:
- `github.com/gin-gonic/gin` — HTTP router
- `gorm.io/gorm` + `gorm.io/driver/sqlite` + `gorm.io/driver/mysql` — ORM
- `github.com/golang-jwt/jwt/v5` — JWT
- `golang.org/x/crypto/bcrypt` — password hashing
- Frontend: `react-query` (data fetching), `zustand` (auth state), `axios` (HTTP), `frappe-gantt` (Gantt chart — chosen over dhtmlx-gantt: MIT license, no server-side dependency, lighter bundle)

---

## Data Models

All GORM models embed `gorm.Model` (ID uint, CreatedAt, UpdatedAt, DeletedAt) unless noted. Soft-delete is used for most entities.

### User

```go
type User struct {
    gorm.Model
    Username        string `gorm:"uniqueIndex;size:64;not null"`  // 账号
    DisplayName     string `gorm:"size:64;not null"`              // 姓名
    PasswordHash    string `gorm:"size:255;not null"`
    IsSuperAdmin    bool   `gorm:"default:false"`
    CanCreateTeam   bool   `gorm:"default:false"`                 // 创建团队权限
}
// Indexes: username (unique)
```

### Team

```go
type Team struct {
    gorm.Model
    Name        string `gorm:"size:100;not null"`   // 团队名称
    Description string `gorm:"size:500"`
    PMID        uint   `gorm:"not null;index"`       // 当前 PM 的 user_id
}
// Indexes: pm_id
```

### TeamMember

```go
// Join table: user ↔ team with role
type TeamMember struct {
    ID        uint      `gorm:"primaryKey;autoIncrement"`
    TeamID    uint      `gorm:"not null;index:idx_team_user,unique"`
    UserID    uint      `gorm:"not null;index:idx_team_user,unique"`
    Role      string    `gorm:"size:20;not null;default:'member'"` // "pm" | "member"
    JoinedAt  time.Time `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
// Indexes: (team_id, user_id) unique composite
```

### MainItem

```go
type MainItem struct {
    gorm.Model
    TeamID          uint       `gorm:"not null;index"`
    Code            string     `gorm:"size:10;not null;uniqueIndex"` // 编号, e.g. "MI-0001"
    Title           string     `gorm:"size:100;not null"`            // 标题
    Priority        string     `gorm:"size:5;not null"`              // "P1"|"P2"|"P3"
    ProposerID      uint       `gorm:"not null"`                     // 提出人
    AssigneeID      *uint      `gorm:"index"`                        // 负责人
    StartDate       *time.Time                                        // 开始时间
    ExpectedEndDate *time.Time `gorm:"index"`                        // 预期完成时间
    ActualEndDate   *time.Time                                        // 实际完成时间
    Status          string     `gorm:"size:20;not null;default:'待开始'"` // 状态
    Completion      float64    `gorm:"default:0"`                    // 完成度 0-100, auto-computed
    IsKeyItem       bool       `gorm:"default:false"`                // 重点事项
    DelayCount      int        `gorm:"default:0"`                    // 延期次数
    ArchivedAt      *time.Time                                        // 归档时间
}
// Indexes: team_id, assignee_id, expected_end_date, (team_id, status), (team_id, priority)
```

### SubItem

```go
type SubItem struct {
    gorm.Model
    TeamID          uint       `gorm:"not null;index"`
    MainItemID      uint       `gorm:"not null;index"`
    Title           string     `gorm:"size:100;not null"`            // 标题
    Description     string     `gorm:"type:text"`                    // 描述
    Priority        string     `gorm:"size:5;not null"`              // "P1"|"P2"|"P3"
    AssigneeID      *uint      `gorm:"index"`                        // 负责人
    StartDate       *time.Time
    ExpectedEndDate *time.Time `gorm:"index"`                        // 预期完成时间
    ActualEndDate   *time.Time
    Status          string     `gorm:"size:20;not null;default:'待开始'"`
    Completion      float64    `gorm:"default:0"`                    // 最新进度记录的完成度
    IsKeyItem       bool       `gorm:"default:false"`
    DelayCount      int        `gorm:"default:0"`                    // 延期次数，>=2 自动升级 P1
    Weight          float64    `gorm:"default:1"`                    // 加权系数，用于主事项汇总
}
// Indexes: main_item_id, team_id, assignee_id, (team_id, status)
// Status values: 待开始|进行中|阻塞中|挂起|已延期|待验收|已完成|已关闭
```

### ProgressRecord

```go
type ProgressRecord struct {
    ID          uint      `gorm:"primaryKey;autoIncrement"`
    SubItemID   uint      `gorm:"not null;index"`
    TeamID      uint      `gorm:"not null;index"`               // denormalized for team isolation
    AuthorID    uint      `gorm:"not null"`                     // 创建人
    Completion  float64   `gorm:"not null"`                     // 完成度 0-100
    Achievement string    `gorm:"type:text"`                    // 成果
    Blocker     string    `gorm:"type:text"`                    // 卡点
    Lesson      string    `gorm:"type:text"`                    // 经验
    IsPMCorrect bool      `gorm:"default:false"`                // PM 修正标记
    CreatedAt   time.Time `gorm:"not null;index"`
    // No UpdatedAt, no DeletedAt — append-only, never modified except PM completion correction
}
// Indexes: sub_item_id, (sub_item_id, created_at), team_id
// Constraint: append-only; only Completion field may be updated by PM (IsPMCorrect=true)
```

### ItemPool

```go
type ItemPool struct {
    gorm.Model
    TeamID          uint       `gorm:"not null;index"`
    Title           string     `gorm:"size:100;not null"`   // 标题
    Background      string     `gorm:"type:text"`           // 背景
    ExpectedOutput  string     `gorm:"type:text"`           // 预期产出
    SubmitterID     uint       `gorm:"not null"`            // 提交人
    Status          string     `gorm:"size:20;not null;default:'待分配'"` // 待分配|已分配|已拒绝
    AssignedMainID  *uint                                   // 挂载的主事项 ID
    AssignedSubID   *uint                                   // 创建的子事项 ID
    AssigneeID      *uint                                   // 指定负责人
    RejectReason    string     `gorm:"size:200"`            // 拒绝原因
    ReviewedAt      *time.Time
    ReviewerID      *uint
}
// Indexes: team_id, (team_id, status), submitter_id
```

### Key Constraints Summary

| Rule | Enforcement |
|------|-------------|
| ProgressRecord append-only | No DELETE/UPDATE in ProgressRepo except PM completion correction |
| MainItem.Completion = weighted avg of SubItem.Completion | Recomputed in SubItemService after every progress append |
| SubItem.DelayCount >= 2 → IsKeyItem=true, Priority="P1" | Checked in SubItemService.ChangeStatus when status→已延期 |
| Team data isolation | TeamScopeMiddleware injects team_id; all repo queries filter by team_id |
| ItemPool assign → creates SubItem atomically | DB transaction in ItemPoolService.Assign |

---

## Interfaces

### AuthService

```go
type AuthService interface {
    Login(ctx context.Context, username, password string) (token string, user *model.User, err error)
    Logout(ctx context.Context, token string) error  // stateless JWT: client-side only; server-side blacklist optional
    ParseToken(ctx context.Context, token string) (*Claims, error)
}
```

### TeamService

```go
type TeamService interface {
    CreateTeam(ctx context.Context, creatorID uint, req dto.CreateTeamReq) (*model.Team, error)
    GetTeam(ctx context.Context, teamID uint) (*model.Team, error)
    ListTeams(ctx context.Context, callerID uint, isSuperAdmin bool) ([]*model.Team, error)
    InviteMember(ctx context.Context, pmID, teamID uint, req dto.InviteMemberReq) error
    RemoveMember(ctx context.Context, pmID, teamID, targetUserID uint) error
    TransferPM(ctx context.Context, currentPMID, teamID, newPMID uint) error
    DisbandTeam(ctx context.Context, callerID uint, teamID uint, confirmName string) error
    UpdateMemberRole(ctx context.Context, pmID, teamID, targetUserID uint, role string) error
    ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error)
}
```

### MainItemService

```go
type MainItemService interface {
    Create(ctx context.Context, teamID, pmID uint, req dto.CreateMainItemReq) (*model.MainItem, error)
    Update(ctx context.Context, teamID, callerID, itemID uint, req dto.UpdateMainItemReq) (*model.MainItem, error)
    Archive(ctx context.Context, teamID, pmID, itemID uint) error
    Get(ctx context.Context, teamID, itemID uint) (*model.MainItem, error)
    List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)
    RecalcCompletion(ctx context.Context, mainItemID uint) error  // called internally after SubItem progress update
}
```

### SubItemService

```go
type SubItemService interface {
    Create(ctx context.Context, teamID, callerID uint, req dto.CreateSubItemReq) (*model.SubItem, error)
    Update(ctx context.Context, teamID, callerID, itemID uint, req dto.UpdateSubItemReq) (*model.SubItem, error)
    ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) error
    Get(ctx context.Context, teamID, itemID uint) (*model.SubItem, error)
    List(ctx context.Context, teamID uint, mainItemID *uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error)
    Assign(ctx context.Context, teamID, pmID, itemID, assigneeID uint) error
}
```

### ProgressService

```go
type ProgressService interface {
    Append(ctx context.Context, teamID, callerID, subItemID uint, req dto.AppendProgressReq) (*model.ProgressRecord, error)
    CorrectCompletion(ctx context.Context, teamID, pmID, recordID uint, completion float64) error
    List(ctx context.Context, teamID, subItemID uint) ([]*model.ProgressRecord, error)
}
```

### ItemPoolService

```go
type ItemPoolService interface {
    Submit(ctx context.Context, teamID, submitterID uint, req dto.SubmitItemPoolReq) (*model.ItemPool, error)
    Assign(ctx context.Context, teamID, pmID, poolItemID uint, req dto.AssignItemPoolReq) error  // tx: update pool + create SubItem
    Reject(ctx context.Context, teamID, pmID, poolItemID uint, reason string) error
    List(ctx context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error)
    Get(ctx context.Context, teamID, poolItemID uint) (*model.ItemPool, error)
}
```

### ViewService

```go
type ViewService interface {
    WeeklyView(ctx context.Context, teamID uint, weekStart time.Time) (*dto.WeeklyViewResult, error)
    GanttView(ctx context.Context, teamID uint, filter dto.GanttFilter) (*dto.GanttResult, error)
    TableView(ctx context.Context, teamID uint, filter dto.TableFilter, page dto.Pagination) (*dto.PageResult[dto.TableRow], error)
    TableExportCSV(ctx context.Context, teamID uint, filter dto.TableFilter) ([]byte, error)
}
```

### ReportService

```go
type ReportService interface {
    Preview(ctx context.Context, teamID uint, weekStart time.Time) (*dto.ReportPreview, error)
    ExportMarkdown(ctx context.Context, teamID uint, weekStart time.Time) ([]byte, error)  // must complete < 5s
}
```

### AdminService

```go
type AdminService interface {
    ListUsers(ctx context.Context) ([]*model.User, error)
    SetCanCreateTeam(ctx context.Context, superAdminID, targetUserID uint, canCreate bool) error
    ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error)
}
```

---

## Core DTO Definitions

Key shared types referenced by service interfaces. CRUD request DTOs are derived directly from the API Handbook field tables.

```go
// Pagination input
type Pagination struct {
    Page     int `form:"page" binding:"min=1"`
    PageSize int `form:"pageSize" binding:"min=1,max=100"`
}

// Generic paginated result
type PageResult[T any] struct {
    Items    []T   `json:"items"`
    Total    int64 `json:"total"`
    Page     int   `json:"page"`
    PageSize int   `json:"pageSize"`
}

// MainItem list filter
type MainItemFilter struct {
    Priority   []string `form:"priority"`   // ["P1","P2","P3"]
    Status     []string `form:"status"`
    AssigneeID *uint    `form:"assigneeId"`
    Archived   bool     `form:"archived"`   // default false
}

// SubItem list filter
type SubItemFilter struct {
    Priority   []string `form:"priority"`
    Status     []string `form:"status"`
    AssigneeID *uint    `form:"assigneeId"`
}

// ItemPool list filter
type ItemPoolFilter struct {
    Status []string `form:"status"` // 待分配|已分配|已拒绝
}

// Table view filter + sort
type TableFilter struct {
    Type       string   `form:"type"`       // "main"|"sub"|""
    Priority   []string `form:"priority"`
    Status     []string `form:"status"`
    AssigneeID *uint    `form:"assigneeId"`
    SortBy     string   `form:"sortBy"`
    SortOrder  string   `form:"sortOrder"`  // "asc"|"desc"
}

// Weekly view result
type WeeklyViewResult struct {
    WeekStart string             `json:"weekStart"` // ISO8601 date
    WeekEnd   string             `json:"weekEnd"`
    Groups    []WeeklyGroupDTO   `json:"groups"`
}

type WeeklyGroupDTO struct {
    MainItem            MainItemSummaryDTO      `json:"mainItem"`
    NewlyCompleted      []SubItemWeekDTO        `json:"newlyCompleted"`
    HasProgress         []SubItemWeekDTO        `json:"hasProgress"`
    NoChangeFromLastWeek []SubItemSummaryDTO    `json:"noChangeFromLastWeek"`
}

type SubItemWeekDTO struct {
    SubItem           model.SubItem          `json:"subItem"`
    ProgressThisWeek  []model.ProgressRecord `json:"progressThisWeek"`
}

// Gantt view result
type GanttResult struct {
    Items []GanttMainItemDTO `json:"items"`
}

type GanttMainItemDTO struct {
    ID              uint             `json:"id"`
    Title           string           `json:"title"`
    Priority        string           `json:"priority"`
    StartDate       *string          `json:"startDate"`
    ExpectedEndDate *string          `json:"expectedEndDate"`
    Completion      float64          `json:"completion"`
    Status          string           `json:"status"`
    IsOverdue       bool             `json:"isOverdue"`
    SubItems        []GanttSubItemDTO `json:"subItems"`
}

type GanttSubItemDTO struct {
    ID              uint    `json:"id"`
    Title           string  `json:"title"`
    StartDate       *string `json:"startDate"`
    ExpectedEndDate *string `json:"expectedEndDate"`
    Completion      float64 `json:"completion"`
    Status          string  `json:"status"`
}

// Table view row
type TableRow struct {
    ID              uint    `json:"id"`
    Type            string  `json:"type"` // "main"|"sub"
    Code            string  `json:"code"`
    Title           string  `json:"title"`
    Priority        string  `json:"priority"`
    AssigneeID      *uint   `json:"assigneeId"`
    AssigneeName    string  `json:"assigneeName"`
    Status          string  `json:"status"`
    Completion      float64 `json:"completion"`
    ExpectedEndDate *string `json:"expectedEndDate"`
    ActualEndDate   *string `json:"actualEndDate"`
}

// Report preview
type ReportPreview struct {
    WeekStart string              `json:"weekStart"`
    WeekEnd   string              `json:"weekEnd"`
    Sections  []ReportSectionDTO  `json:"sections"`
}

type ReportSectionDTO struct {
    MainItem MainItemSummaryDTO   `json:"mainItem"`
    SubItems []ReportSubItemDTO   `json:"subItems"`
}

type ReportSubItemDTO struct {
    ID           uint     `json:"id"`
    Title        string   `json:"title"`
    Completion   float64  `json:"completion"`
    Achievements []string `json:"achievements"`
    Blockers     []string `json:"blockers"`
}

// Shared summary types
type MainItemSummaryDTO struct {
    ID         uint    `json:"id"`
    Title      string  `json:"title"`
    Completion float64 `json:"completion"`
}

type SubItemSummaryDTO struct {
    ID         uint    `json:"id"`
    Title      string  `json:"title"`
    Status     string  `json:"status"`
    Completion float64 `json:"completion"`
}

// Admin DTOs
type AdminTeamDTO struct {
    ID            uint   `json:"id"`
    Name          string `json:"name"`
    PMDisplayName string `json:"pmDisplayName"`
    MemberCount   int    `json:"memberCount"`
    MainItemCount int    `json:"mainItemCount"`
    CreatedAt     string `json:"createdAt"`
}
```

### Frontend Error Handling Convention

Axios response interceptor handles errors globally:

| HTTP Status | Interceptor Behavior |
|-------------|----------------------|
| 401 | Clear auth state → redirect to `/login` |
| 403 | `message.error("权限不足")` |
| 404 | `message.error("资源不存在")` |
| 422 | Pass error to calling component for inline display |
| 500 | `message.error("服务器错误，请稍后重试")` |

Business-logic errors (422) are surfaced inline (form errors, modal alerts) rather than via global toast, so the interceptor re-throws them for the component to handle.

---

## API Design Summary

All endpoints are prefixed with `/api/v1`. JWT token is passed as `Authorization: Bearer <token>`.

### Auth
| Method | Path | Role |
|--------|------|------|
| POST | /api/v1/auth/login | public |
| POST | /api/v1/auth/logout | any authenticated |

### Teams
| Method | Path | Role |
|--------|------|------|
| POST | /api/v1/teams | user with CanCreateTeam |
| GET | /api/v1/teams | SuperAdmin (all) / user (own) |
| GET | /api/v1/teams/:teamId | team member |
| PUT | /api/v1/teams/:teamId | PM |
| DELETE | /api/v1/teams/:teamId | PM / SuperAdmin |
| GET | /api/v1/teams/:teamId/members | team member |
| POST | /api/v1/teams/:teamId/members | PM |
| DELETE | /api/v1/teams/:teamId/members/:userId | PM |
| PUT | /api/v1/teams/:teamId/pm | PM |

### MainItems
| Method | Path | Role |
|--------|------|------|
| POST | /api/v1/teams/:teamId/main-items | PM |
| GET | /api/v1/teams/:teamId/main-items | team member |
| GET | /api/v1/teams/:teamId/main-items/:itemId | team member |
| PUT | /api/v1/teams/:teamId/main-items/:itemId | PM |
| POST | /api/v1/teams/:teamId/main-items/:itemId/archive | PM |

### SubItems
| Method | Path | Role |
|--------|------|------|
| POST | /api/v1/teams/:teamId/main-items/:mainId/sub-items | PM / member |
| GET | /api/v1/teams/:teamId/main-items/:mainId/sub-items | team member |
| GET | /api/v1/teams/:teamId/sub-items/:itemId | team member |
| PUT | /api/v1/teams/:teamId/sub-items/:itemId | PM / assignee |
| PUT | /api/v1/teams/:teamId/sub-items/:itemId/status | PM / assignee |
| PUT | /api/v1/teams/:teamId/sub-items/:itemId/assignee | PM |

### ProgressRecords
| Method | Path | Role |
|--------|------|------|
| POST | /api/v1/teams/:teamId/sub-items/:itemId/progress | member / assignee |
| GET | /api/v1/teams/:teamId/sub-items/:itemId/progress | team member |
| PATCH | /api/v1/teams/:teamId/progress/:recordId/completion | PM |

### ItemPool
| Method | Path | Role |
|--------|------|------|
| POST | /api/v1/teams/:teamId/item-pool | any team member |
| GET | /api/v1/teams/:teamId/item-pool | PM / member |
| GET | /api/v1/teams/:teamId/item-pool/:poolId | PM / member |
| POST | /api/v1/teams/:teamId/item-pool/:poolId/assign | PM |
| POST | /api/v1/teams/:teamId/item-pool/:poolId/reject | PM |

### Views
| Method | Path | Role |
|--------|------|------|
| GET | /api/v1/teams/:teamId/views/weekly | team member |
| GET | /api/v1/teams/:teamId/views/gantt | team member |
| GET | /api/v1/teams/:teamId/views/table | team member |
| GET | /api/v1/teams/:teamId/views/table/export | team member |

### Reports
| Method | Path | Role |
|--------|------|------|
| GET | /api/v1/teams/:teamId/reports/weekly/preview | PM / member |
| GET | /api/v1/teams/:teamId/reports/weekly/export | PM / member |

### Admin
| Method | Path | Role |
|--------|------|------|
| GET | /api/v1/admin/users | SuperAdmin |
| PUT | /api/v1/admin/users/:userId/can-create-team | SuperAdmin |
| GET | /api/v1/admin/teams | SuperAdmin |

Full request/response schemas are in `api-handbook.md`.

---

## Error Handling

### Error Types

```go
// internal/pkg/errors/errors.go

type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

var (
    ErrUnauthorized       = &AppError{Code: "UNAUTHORIZED",        Status: 401, Message: "authentication required"}
    ErrForbidden          = &AppError{Code: "FORBIDDEN",           Status: 403, Message: "insufficient permissions"}
    ErrTeamNotFound       = &AppError{Code: "TEAM_NOT_FOUND",      Status: 404, Message: "team not found"}
    ErrItemNotFound       = &AppError{Code: "ITEM_NOT_FOUND",      Status: 404, Message: "item not found"}
    ErrNotTeamMember      = &AppError{Code: "NOT_TEAM_MEMBER",     Status: 403, Message: "not a member of this team"}
    ErrInvalidStatus      = &AppError{Code: "INVALID_STATUS",      Status: 422, Message: "invalid status transition"}
    ErrArchiveNotAllowed  = &AppError{Code: "ARCHIVE_NOT_ALLOWED", Status: 422, Message: "only completed or closed items can be archived"}
    ErrProgressRegression = &AppError{Code: "PROGRESS_REGRESSION", Status: 422, Message: "completion cannot be lower than previous record"}
    ErrValidation         = &AppError{Code: "VALIDATION_ERROR",    Status: 400, Message: "request validation failed"}
    ErrInternal           = &AppError{Code: "INTERNAL_ERROR",      Status: 500, Message: "internal server error"}
)
```

### HTTP Status Mapping

| Scenario | HTTP Status | Code |
|----------|-------------|------|
| Missing / invalid JWT | 401 | UNAUTHORIZED |
| Valid JWT but wrong role | 403 | FORBIDDEN |
| Accessing another team's data | 403 | NOT_TEAM_MEMBER |
| Resource not found | 404 | *_NOT_FOUND |
| Invalid state transition | 422 | INVALID_STATUS |
| Business rule violation | 422 | specific code |
| Request body validation | 400 | VALIDATION_ERROR |
| Unexpected server error | 500 | INTERNAL_ERROR |

### Response Envelope

All responses use a consistent envelope:

```json
// success
{ "code": 0, "data": { ... } }

// error
{ "code": "FORBIDDEN", "message": "not a member of this team" }
```

### Team Isolation Enforcement

`TeamScopeMiddleware` runs on every `/api/v1/teams/:teamId/*` route:

1. Extract `teamId` from URL path.
2. Look up `TeamMember` record for `(teamId, callerUserID)`.
3. If not found and caller is not SuperAdmin → return 403 `NOT_TEAM_MEMBER`.
4. Inject `teamID` and `callerRole` into Gin context.
5. All repository queries receive `teamID` as a mandatory filter — never derived from request body.

---

## Testing Strategy

### Unit Tests

Target: **≥ 80% coverage** on the service layer.

Key scenarios to cover per service:

- `MainItemService.RecalcCompletion` — weighted average with zero/one/many sub-items, all-zero weights.
- `SubItemService.ChangeStatus` — all valid transitions, all invalid transitions (expect `ErrInvalidStatus`), delay-count increment, auto-upgrade to P1 on `DelayCount >= 2`.
- `ProgressService.Append` — completion regression check, PM correction bypass.
- `ItemPoolService.Assign` — transaction rollback if SubItem creation fails.
- `ReportService.ExportMarkdown` — output format, empty week, large dataset timing.

Use `testify/mock` for repository mocks. No real DB in unit tests.

### Integration Tests

Use `go test` with an in-memory SQLite database (same GORM driver as dev). Run migrations before each test suite.

Cover:
- Full auth flow: login → get token → access protected route → logout.
- Team isolation: user A cannot read team B's items (expect 403).
- Progress append → MainItem completion auto-update (end-to-end through service + repo).
- ItemPool assign transaction: verify both ItemPool status update and SubItem creation are committed atomically.
- Weekly report export: seed data for a week, call export, assert Markdown structure.

### Frontend Tests

- Unit tests with Vitest + React Testing Library for form validation logic and state machines.
- API mock with MSW (Mock Service Worker) for page-level integration tests.
- No E2E tests in v1 scope.

### Coverage Target

| Layer | Target |
|-------|--------|
| Service (Go) | ≥ 80% |
| Repository (Go) | ≥ 60% (integration tests) |
| Handler (Go) | ≥ 70% |
| Frontend components | ≥ 60% |

---

## Security

### JWT Flow

1. Client POSTs credentials to `/api/v1/auth/login`.
2. Server verifies password with `bcrypt.CompareHashAndPassword`.
3. On success, server signs a JWT with claims: `{ sub: userID, role: "pm"|"member"|"superadmin", exp: now+24h }`.
4. Client stores token in memory (not localStorage) and attaches it as `Authorization: Bearer <token>` on every request.
5. `AuthMiddleware` validates signature and expiry on every protected route. Expired tokens return 401.
6. Logout is client-side token discard. Server-side blacklist is out of scope for v1 (stateless design).

JWT secret is loaded from environment variable `JWT_SECRET` (min 32 bytes). Never hardcoded.

### RBAC Middleware

```
Route group: /api/v1/admin/*
  → RequireRole("superadmin")

Route group: /api/v1/teams/:teamId/*
  → TeamScopeMiddleware (verifies membership, injects role)
  → Per-handler: RequireTeamRole("pm") for write operations
```

Role hierarchy: `superadmin > pm > member`. SuperAdmin bypasses team membership checks.

### Team Isolation at Query Layer

Every repository method that queries team-scoped data accepts `teamID uint` as a required parameter and applies `.Where("team_id = ?", teamID)` unconditionally. The `teamID` value always comes from the middleware-injected context, never from user-supplied request body or query params.

### Additional Mitigations

| Threat | Mitigation |
|--------|------------|
| SQL injection | GORM parameterized queries only; no raw SQL with user input |
| Password exposure | bcrypt cost ≥ 12; password never logged or returned in responses |
| Mass assignment | Explicit DTO structs; never bind request body directly to model |
| CORS | Gin CORS middleware with explicit allowed origins (env-configured) |
| Rate limiting | Gin rate-limit middleware on `/api/v1/auth/login` (10 req/min per IP) |

---

## Open Questions

- [x] Should JWT expiry be 24h with refresh token, or short-lived (1h) + refresh? **Decision: 24h, no refresh token in v1.**
- [x] SubItem weight for completion rollup: fixed at 1.0 for all sub-items in v1, or PM-configurable? **Decision: fixed at 1.0 in v1 (simple average).**
- [x] Should archived MainItems be queryable via a separate `?archived=true` filter, or a dedicated endpoint? **Decision: `?archived=true` query param on the existing list endpoint.**
- [x] PM correction of ProgressRecord.Completion: update in-place (with audit flag) or insert a new correction record? **Decision: update in-place with `IsPMCorrect=true` flag. Rationale: keeps the timeline clean; the flag provides sufficient audit trail for v1.**

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| PostgreSQL instead of MySQL | Better JSON support, full-text search | Heavier ops, team unfamiliar | MySQL is team standard; GORM abstracts the difference |
| Redis for completion cache | Faster reads | Extra infra, cache invalidation complexity | Data volume is small; synchronous recalc is fast enough |
| GraphQL API | Flexible queries for views | Higher complexity, less familiar to team | REST is sufficient; views are well-defined |
| Server-Sent Events for live updates | Real-time completion updates | Adds stateful connections | Out of scope v1; polling on focus is acceptable |

### References

- PRD Spec: `docs/features/pm-work-tracker/prd/prd-spec.md`
- UI Functions: `docs/features/pm-work-tracker/prd/prd-ui-functions.md`
- API Handbook: `docs/features/pm-work-tracker/design/api-handbook.md`
