# Architecture

PM Work Tracker is a project management tool with a Go backend (Gin + GORM) and a React/TypeScript frontend (Vite + Tailwind v4).

## Backend Layers

Data flows bottom-up through five layers. Each layer has a single responsibility and may only depend on layers below it.

```
Router -> Handler -> Service -> Repository -> Model
                    |-> VO (response shaping)
                    |-> DTO (request/response types)
```

### Model (`backend/internal/model/`)

GORM-annotated structs that map 1:1 to database tables. No business logic.

- Embeds `BaseModel` for `id`, `createdAt`, `updatedAt`, `DeletedAt`
- JSON tags are camelCase (`json:"teamId"`) — GORM maps these to snake_case columns automatically
- Defines `TableName()` for explicit table mapping

```go
type MainItem struct {
    BaseModel
    TeamID   uint   `gorm:"not null" json:"teamId"`
    Title    string `gorm:"type:varchar(100);not null" json:"title"`
    Priority string `gorm:"type:varchar(5);not null" json:"priority"`
}
```

### DTO (`backend/internal/dto/`)

Request/response types for the API boundary. No GORM tags.

- Request DTOs use `binding` tags for Gin validation
- Response DTOs carry computed or aggregated data (e.g., `WeeklyViewResponse`, `TableRow`)
- Generic `PageResult[T]` for pagination responses
- Filter structs use `form` tags for query parameter binding

```go
type MainItemCreateReq struct {
    Title     string  `json:"title" binding:"required,max=100"`
    Priority  string  `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
}
```

### VO (`backend/internal/vo/`)

View Objects that shape model data for the frontend. Converts internal representations to API-friendly formats.

- Converts `time.Time` to `"2006-01-02"` strings or RFC3339
- Omits sensitive or unnecessary fields
- Constructor functions: `NewMainItemVO(model) -> VO`
- Imported by handlers, not by services

```go
func NewMainItemVO(m *model.MainItem) MainItemVO {
    return MainItemVO{
        ID:    m.ID,
        Title: m.Title,
        CreatedAt: m.CreatedAt.Format(time.RFC3339),
    }
}
```

### Service (`backend/internal/service/`)

Business logic layer. Defines interfaces that handlers depend on.

- Each domain entity has one service file and one interface
- Receives DTOs from handlers, returns models or DTOs
- Calls repository for data access
- Uses `pkg/errors` for domain error mapping
- Constructor: `NewXxxService(repoA, repoB) XxxService`

```go
type MainItemService interface {
    Create(ctx context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error)
    List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)
}
```

### Handler (`backend/internal/handler/`)

HTTP layer. Binds requests, calls services, returns responses.

- Each handler struct holds service and repo dependencies
- Parses path/query params, binds JSON to DTOs
- Calls service methods, then converts results via VO constructors
- Uses `apperrors.RespondError` / `RespondOK` for consistent error/success responses
- Constructor pattern: `NewXxxHandlerWithDeps(svc, repoA, repoB)`

```go
func (h *MainItemHandler) Create(c *gin.Context) {
    teamID := middleware.GetTeamID(c)
    var req dto.MainItemCreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        apperrors.RespondError(c, apperrors.ErrValidation)
        return
    }
    item, err := h.svc.Create(c.Request.Context(), teamID, pmID, req)
    if err != nil {
        apperrors.RespondError(c, err)
        return
    }
    apperrors.RespondOK(c, vo.NewMainItemVO(item))
}
```

### Repository (`backend/internal/repository/` + `repository/gorm/`)

Data access layer. Defines interfaces in `repository/`, GORM implementations in `repository/gorm/`.

- Interface-driven: handlers/services depend on interfaces, not concrete types
- GORM implementations wrap `*gorm.DB` operations
- Returns `errors.ErrNotFound` (not raw GORM errors) for missing records
- Constructor: `NewGormXxxRepo(db *gorm.DB) repository.XxxRepo`

## Cross-Cutting Concerns

### Middleware (`backend/internal/middleware/`)

- **Auth**: JWT token extraction and user loading (`auth.go`)
- **Team Scope**: Validates team membership and injects `teamID` into context (`team_scope.go`)
- **Permission**: Checks permission codes via role-based access (`permission.go`)
- **RBAC**: Role-based access control helpers (`rbac.go`)

### Errors (`backend/internal/pkg/errors/`)

- `AppError` struct with `Code`, `Message`, `Status` (HTTP code)
- Sentinel errors: `ErrItemNotFound`, `ErrUnauthorized`, etc.
- `RespondError(c, err)` maps any error to a JSON envelope
- `RespondOK(c, data)` wraps success responses in `{code: 0, data: ...}`

### Migrations (`backend/internal/migration/`)

- `runner.go` manages auto-migration for all models
- `rbac.go` seeds default roles and permissions

### Other Packages (`backend/internal/pkg/`)

- `jwt/`: Token generation and validation
- `permissions/`: Permission code constants
- `report/`: HTML report rendering

## Frontend Structure

```
frontend/src/
  api/          API client modules (one per domain: teams.ts, mainItems.ts)
  components/
    ui/         Generic UI primitives (Button, Input, Badge, Dialog, Toast...)
    shared/     Domain-aware shared components (PriorityBadge, ProgressBar, StatusBadge)
    layout/     App layout shell (AppLayout, Sidebar)
  hooks/        Custom React hooks (usePermission)
  pages/        Route-level page components (one file per route)
  store/        Zustand stores (auth.ts, team.ts)
  types/        Shared TypeScript interfaces
  lib/          Utilities (permissions, utils)
  mocks/        MSW handlers for integration tests
```

### API Layer (`frontend/src/api/`)

- One module per domain entity (e.g., `teams.ts`, `mainItems.ts`)
- Uses shared axios `client` instance with auth interceptor
- Client auto-unwraps `{code: 0, data: ...}` envelope
- 401 responses trigger redirect to login

### UI Components (`frontend/src/components/ui/`)

Generic primitives with no domain knowledge. Each has a co-located test file.

- Styled with Tailwind using theme tokens (not hardcoded colors)
- Built on Radix UI primitives where applicable
- Pattern: `forwardRef` + `cn()` for className merging

### Pages (`frontend/src/pages/`)

Route-level components. Each page corresponds to one route. Pages compose shared and UI components. Each has a co-located test file.

### State (`frontend/src/store/`)

Zustand stores with flat structure. Currently: `auth.ts` (token + user), `team.ts` (selected team).

## Router Wiring (`backend/internal/handler/router.go`)

All routes are registered in `SetupRouter(deps *Dependencies)`. The `Dependencies` struct holds all handlers and key repos. Route groups:

- `GET /health` -- no auth
- `/api/v1/auth/*` -- login/logout (rate-limited)
- `/api/v1/teams/:teamId/*` -- auth + team membership required
- `/api/v1/admin/*` -- auth + permission-gated
- `/api/v1/me/permissions` -- auth only
