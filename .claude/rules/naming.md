# Naming Conventions

## Go Struct Tags

All JSON tags must be camelCase. GORM maps camelCase JSON tags to snake_case columns automatically.

```go
// ✅ Correct
type MainItem struct {
    BaseModel
    TeamID   uint   `gorm:"not null" json:"teamId"`
    Title    string `gorm:"type:varchar(100);not null" json:"title"`
    Priority string `gorm:"type:varchar(5);not null" json:"priority"`
}

// ❌ Wrong — snake_case JSON tags
type MainItem struct {
    TeamID   uint   `json:"team_id"`
    Title    string `json:"title"`
}
```

Enforced by `tagliatelle` linter in `backend/.golangci.yml`.

## Go Exported Names

Functions, types, and interfaces use PascalCase. File names use snake_case.

```go
// ✅ Correct
func NewMainItemHandler(svc MainItemService) *MainItemHandler { ... }
type MainItemService interface { ... }
type PageResult[T any] struct { ... }

// ❌ Wrong
func newMainItemHandler(...) { ... }  // unexported constructor
type mainItemService interface { ... }  // unexported interface
```

File naming:
- `backend/internal/handler/main_item_handler.go`
- `backend/internal/repository/gorm/main_item_repo.go`
- `backend/internal/service/main_item_service.go`

## Go Package Names

Lowercase, single word, no underscores.

```go
// ✅ Correct
package handler
package service
package repository
package model
package dto
package vo
package errors  // pkg/errors
package dates   // pkg/dates

// ❌ Wrong
package http_handler
package repo
package util
```

## TypeScript Interfaces and Types

PascalCase for interfaces and types. Files use camelCase.

```go
// ✅ Correct
interface MainItem { id: number; teamId: number; title: string }
interface WeeklyViewResponse { rows: TableRow[] }
type PageResult<T> = { items: T[]; total: number }

// ❌ Wrong
interface main_item { ... }
interface IMainItem { ... }  // no Hungarian prefix
type pageResult<T> = { ... }
```

## TypeScript Files and Modules

One API module per domain entity, matching the backend service name.

```
frontend/src/api/mainItems.ts    ← matches MainItemService
frontend/src/api/itemPool.ts     ← matches ItemPoolService
frontend/src/api/progress.ts     ← matches ProgressService
frontend/src/api/permissions.ts
frontend/src/api/auth.ts
```

## Database Columns

Always snake_case, managed by GORM's naming strategy. Never reference column names directly in Go code — let GORM handle the mapping.

```go
// ✅ Correct — let GORM map json:"teamId" → column "team_id"
TeamID uint `gorm:"not null" json:"teamId"`

// ❌ Wrong — manual column override
TeamID uint `gorm:"column:team_id" json:"teamId"`
```
