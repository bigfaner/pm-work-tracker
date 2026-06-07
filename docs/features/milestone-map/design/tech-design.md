---
created: 2026-05-12
updated: 2026-06-06
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: 里程碑图

## Overview

新增 MilestoneMap 和 Milestone 两个层级实体，复用现有 5 层架构（Model → Repository → Service → Handler → Router），遵循 BaseModel、状态机、权限码等既有模式。前端新增 `/milestones` 页面（两级视图），并在 3 个现有页面集成里程碑维度。

核心约束：
- 完成度实时计算（GET 时不持久化）
- 软删除里程碑时事务内解绑关联 MI
- 状态机转换校验复用 `status/transition.go` 的 `IsValidTransition` / `GetAvailableTransitions`
- MainItem 新增 `milestone_key` 列（无 DDL 外键）

## Architecture

### Layer Placement

```
Model (backend/internal/model/)
  ├── milestone_map.go          [NEW]
  └── milestone.go              [NEW]

Repository (backend/internal/repository/)
  ├── milestone_map_repo.go     [NEW] interface
  ├── milestone_repo.go         [NEW] interface
  └── gorm/
      ├── milestone_map_repo.go [NEW] implementation
      └── milestone_repo.go     [NEW] implementation

Service (backend/internal/service/)
  ├── milestone_map_service.go  [NEW] interface + implementation
  └── milestone_service.go      [NEW] interface + implementation

Handler (backend/internal/handler/)
  ├── milestone_map_handler.go  [NEW]
  └── milestone_handler.go      [NEW]

VO (backend/internal/vo/)
  └── milestone_vo.go           [NEW]

DTO (backend/internal/dto/)
  ├── milestone_dto.go          [NEW]
  └── item_dto.go               [MODIFIED] MainItemFilter/CreateReq/UpdateReq

Status (backend/internal/pkg/status/)
  └── milestone_status.go       [NEW]

Permission (backend/internal/pkg/permissions/)
  └── codes.go                  [MODIFIED] add milestone resource

Model (backend/internal/model/)
  └── main_item.go              [MODIFIED] add MilestoneKey field

VO (backend/internal/vo/)
  └── item_vo.go                [MODIFIED] add MilestoneName field

Router (backend/internal/handler/)
  └── router.go                 [MODIFIED] add routes + deps

Frontend:
  src/api/milestones.ts         [NEW]
  src/types/index.ts            [MODIFIED]
  src/pages/MilestonesPage.tsx  [NEW]
  src/pages/milestones/         [NEW] sub-components
  src/components/layout/Sidebar.tsx  [MODIFIED] add nav item
  src/App.tsx                   [MODIFIED] add route
```

### Component Diagram

```
┌──────────────────────────────────────────────────────────┐
│ Frontend                                                  │
│  ┌──────────────┐  ┌────────────────┐  ┌──────────────┐ │
│  │MilestonesPage│  │ItemsPage+filter│  │TableView+col │ │
│  └──────┬───────┘  └───────┬────────┘  └──────┬───────┘ │
│         │                  │                   │         │
│  ┌──────┴──────────────────┴───────────────────┴───────┐ │
│  │              api/milestones.ts + mainItems.ts       │ │
│  └──────────────────────┬──────────────────────────────┘ │
└─────────────────────────┼────────────────────────────────┘
                          │ HTTP
┌─────────────────────────┼────────────────────────────────┐
│ Backend                  │                                │
│  ┌───────────────────────┴──────────────────────────────┐│
│  │ Router (teams group)                                  ││
│  │   milestone-maps/*  milestones/*                      ││
│  └──────┬─────────────────────────┬──────────────────────┘│
│  ┌──────┴──────────┐  ┌──────────┴───────────┐           │
│  │MilestoneMap     │  │Milestone             │           │
│  │Handler          │  │Handler               │           │
│  └──────┬──────────┘  └──────────┬───────────┘           │
│  ┌──────┴──────────┐  ┌──────────┴───────────┐           │
│  │MilestoneMap     │  │Milestone             │           │
│  │Service          │  │Service               │           │
│  └──────┬──────────┘  └──────────┬───────────┘           │
│  ┌──────┴──────────┐  ┌──────────┴───────────┐           │
│  │MilestoneMap     │  │Milestone             │           │
│  │Repo (GORM)      │  │Repo (GORM)           │           │
│  └─────────────────┘  └──────────────────────┘           │
│         │                      │                          │
│  ┌──────┴──────────────────────┴──────────────────────┐  │
│  │  SQLite / MySQL                                     │  │
│  │  pmw_milestone_maps  pmw_milestones  pmw_main_items │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

### Dependencies

| Dependency | Type | Notes |
|-----------|------|-------|
| `pkg/snowflake` | Internal | BizKey 生成 |
| `pkg/status` | Internal | 状态机注册（复用 `IsValidTransition`, `GetAvailableTransitions`） |
| `pkg/permissions` | Internal | 权限码注册（添加到 `Registry`） |
| `middleware` | Internal | `AuthMiddleware` + `TeamScopeMiddleware` + `RequirePermission` |
| `pkg` | Internal | VO 转换（`FormatID`, `FormatIDPtr`） |
| `pkg/dates` | Internal | 日期格式化（`FormatTimePtr`） |
| `pkg/errors` | Internal | 错误处理（`MapNotFound`, `ErrValidation`, `RespondOK` 等） |
| `pkg/handler` | Internal | `ResolveBizKey` 路由参数解析 |
| `pkg/repo` | Internal | `FindByID[T]`, `UpdateFields[T]`, `DBTransactor` |
| `dto` | Internal | 分页（`ApplyPaginationDefaults`）、`PageResult[T]` |
| `repository/gorm` | Internal | `NotDeleted` scope |

## Data Models

> Full SQL DDL in `design/schema.sql`，ER 图在 `design/er-diagram.md`。

### MilestoneMap Model

```go
// model/milestone_map.go
package model

import "time"

type MilestoneMap struct {
	BaseModel
	TeamKey          int64      `gorm:"not null;index:idx_milestone_maps_team_status" json:"teamKey"`
	CreatorKey       int64      `gorm:"not null" json:"creatorKey"`
	AssigneeKey      int64      `gorm:"not null" json:"assigneeKey"`
	MapName          string     `gorm:"type:varchar(100);not null" json:"mapName"`
	MapDesc          string     `gorm:"type:varchar(2000);not null;default:''" json:"mapDesc"`
	MapStatus        string     `gorm:"type:varchar(20);not null;default:'planning';index:idx_milestone_maps_team_status" json:"mapStatus"`
	PlanStartDate *time.Time `json:"planStartDate"`
	ExpectedEndDate   *time.Time `json:"expectedEndDate"`
}

func (MilestoneMap) TableName() string { return "pmw_milestone_maps" }
```

### Milestone Model

```go
// model/milestone.go
package model

import "time"

type Milestone struct {
	BaseModel
	TeamKey          int64      `gorm:"not null;index:idx_milestones_team_status" json:"teamKey"`
	MilestoneMapKey  int64      `gorm:"not null" json:"milestoneMapKey"`
	MilestoneName    string     `gorm:"type:varchar(100);not null" json:"milestoneName"`
	MilestoneDesc    string     `gorm:"type:varchar(2000);not null;default:''" json:"milestoneDesc"`
	ExpectedEndDate  *time.Time `json:"expectedEndDate"`
	MilestoneStatus  string     `gorm:"type:varchar(20);not null;default:'not_started';index:idx_milestones_team_status" json:"milestoneStatus"`
}

func (Milestone) TableName() string { return "pmw_milestones" }
```

### MainItem Model (Modified)

```go
// model/main_item.go — 新增字段
MilestoneKey     *int64     `gorm:"index" json:"milestoneKey"`  // NEW
```

### Field Quick Reference

| Model | Key Fields | Notes |
|-------|------------|-------|
| MilestoneMap | map_name, map_desc, map_status, team_key, creator_key, assignee_key, plan_start_date, expected_end_date | 5 态：planning→reviewed→ready→executing→completed |
| Milestone | milestone_name, milestone_desc, expected_end_date, milestone_status, milestone_map_key, team_key | 4 态：not_started→in_progress→completed/cancelled |
| MainItem (modified) | +milestone_key | 可空，引用 pmw_milestones.biz_key |

## Status System

### Status Definitions

```go
// pkg/status/milestone_status.go
package status

var MilestoneMapStatuses = map[string]StatusDef{
	"planning":  {Code: "planning", Name: "规划中", Terminal: false},
	"reviewed":  {Code: "reviewed", Name: "已评审", Terminal: false},
	"ready":     {Code: "ready", Name: "待实施", Terminal: false},
	"executing": {Code: "executing", Name: "实施中", Terminal: false},
	"completed": {Code: "completed", Name: "已完成", Terminal: true},
}

var MilestoneStatuses = map[string]StatusDef{
	"not_started": {Code: "not_started", Name: "未开始", Terminal: false},
	"in_progress": {Code: "in_progress", Name: "进行中", Terminal: false},
	"completed":   {Code: "completed", Name: "已完成", Terminal: false},
	"cancelled":   {Code: "cancelled", Name: "已取消", Terminal: true},
}

func GetMilestoneMapStatus(code string) (StatusDef, bool) {
	def, ok := MilestoneMapStatuses[code]
	return def, ok
}

func GetMilestoneStatus(code string) (StatusDef, bool) {
	def, ok := MilestoneStatuses[code]
	return def, ok
}
```

### Transitions

```go
// pkg/status/milestone_status.go (continued)

var MilestoneMapTransitions = map[string][]string{
	"planning":  {"reviewed"},
	"reviewed":  {"ready", "planning"},
	"ready":     {"executing", "reviewed"},
	"executing": {"completed", "ready"},
	// "completed" is terminal — no entries
}

var MilestoneTransitions = map[string][]string{
	"not_started": {"in_progress", "cancelled"},
	"in_progress": {"completed", "cancelled"},
	"completed":   {"cancelled"},
	// "cancelled" is terminal — no entries
}
```

复用现有 `IsValidTransition(transitions, from, to)` 和 `GetAvailableTransitions(transitions, current)` 函数。

## Permission System

### Registry Addition

在 `permissions/codes.go` 的 `Registry` 变量中，`item_pool` 资源之后添加：

```go
{
	Resource: "milestone",
	Permissions: []Permission{
		{Code: "milestone:create", Description: "创建里程碑图和里程碑"},
		{Code: "milestone:read", Description: "查看里程碑图和里程碑"},
		{Code: "milestone:update", Description: "编辑里程碑图/里程碑信息及状态切换"},
		{Code: "milestone:delete", Description: "删除里程碑图和里程碑"},
	},
},
```

### Permission-to-Operation Mapping

| Permission Code | Guarded Operations |
|-----------------|-------------------|
| `milestone:create` | POST `/teams/:teamId/milestone-maps` |
|                 | POST `/teams/:teamId/milestone-maps/:mapId/milestones` |
| `milestone:read`  | GET `/teams/:teamId/milestone-maps` |
|                 | GET `/teams/:teamId/milestone-maps/:mapId` |
|                 | GET `/teams/:teamId/milestone-maps/:mapId/available-transitions` |
|                 | GET `/teams/:teamId/milestone-maps/:mapId/milestones` |
|                 | GET `/teams/:teamId/milestones` |
|                 | GET `/teams/:teamId/milestones/:milestoneId` |
|                 | GET `/teams/:teamId/milestones/:milestoneId/available-transitions` |
| `milestone:update` | PUT `/teams/:teamId/milestone-maps/:mapId` |
|                 | PUT `/teams/:teamId/milestone-maps/:mapId/status` |
|                 | PUT `/teams/:teamId/milestones/:milestoneId` |
|                 | PUT `/teams/:teamId/milestones/:milestoneId/status` |
| `milestone:delete` | DELETE `/teams/:teamId/milestone-maps/:mapId` |
|                 | DELETE `/teams/:teamId/milestones/:milestoneId` |

> `ChangeStatus` 使用 `milestone:update`。`Delete` 使用 `milestone:delete`。MainItem 绑定/解绑通过 `main_item:update` 权限控制。

## Interfaces

### MilestoneMapRepo

```go
// repository/milestone_map_repo.go
type MilestoneMapRepo interface {
	Create(ctx context.Context, m *model.MilestoneMap) error
	FindByID(ctx context.Context, id uint) (*model.MilestoneMap, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error)
	Update(ctx context.Context, m *model.MilestoneMap, fields map[string]interface{}) error
	List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error)
	SoftDelete(ctx context.Context, id uint) error
}
```

### MilestoneRepo

```go
// repository/milestone_repo.go
type MilestoneRepo interface {
	Create(ctx context.Context, m *model.Milestone) error
	FindByID(ctx context.Context, id uint) (*model.Milestone, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error)
	FindBatchByBizKeys(ctx context.Context, bizKeys []int64) (map[int64]*model.Milestone, error)
	Update(ctx context.Context, m *model.Milestone, fields map[string]interface{}) error
	ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error)
	ListByTeam(ctx context.Context, teamBizKey int64, filter dto.MilestoneTeamFilter) ([]model.Milestone, error)
	SoftDelete(ctx context.Context, id uint) error
	SoftDeleteByMap(ctx context.Context, milestoneMapID uint) error
}
```

### MilestoneMapService

```go
// service/milestone_map_service.go
type MilestoneMapService interface {
	Create(ctx context.Context, teamBizKey int64, creatorBizKey int64, req dto.MilestoneMapCreateReq) (*model.MilestoneMap, error)
	Update(ctx context.Context, teamBizKey int64, mapID uint, req dto.MilestoneMapUpdateReq) error
	Get(ctx context.Context, mapID uint) (*model.MilestoneMap, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error)
	List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error)
	Delete(ctx context.Context, teamBizKey int64, mapID uint) error
	ChangeStatus(ctx context.Context, teamBizKey int64, mapID uint, newStatus string) (*model.MilestoneMap, error)
	AvailableTransitions(ctx context.Context, mapID uint) ([]string, error)
}
```

### MilestoneService

```go
// service/milestone_service.go
type MilestoneService interface {
	Create(ctx context.Context, teamBizKey int64, milestoneMapBizKey int64, req dto.MilestoneCreateReq) (*model.Milestone, error)
	Update(ctx context.Context, teamBizKey int64, milestoneID uint, req dto.MilestoneUpdateReq) error
	Get(ctx context.Context, milestoneID uint) (*model.Milestone, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error)
	ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error)
	ListByTeam(ctx context.Context, teamBizKey int64, filter dto.MilestoneTeamFilter) ([]model.Milestone, error)
	Delete(ctx context.Context, teamBizKey int64, milestoneID uint) error
	ChangeStatus(ctx context.Context, teamBizKey int64, milestoneID uint, newStatus string) (*model.Milestone, error)
	AvailableTransitions(ctx context.Context, milestoneID uint) ([]string, error)
}
```

### Computed Fields (Service Layer)

完成度在 GET 时实时查询，不持久化：

```go
// Milestone 完成度：关联 MI 的 completion 平均值
func (s *milestoneService) calcCompletion(ctx context.Context, milestoneBizKey int64) float64

// MilestoneMap 整体进度：所有关联 MI 的 completion 平均值
func (s *milestoneMapService) calcOverallProgress(ctx context.Context, milestoneMapBizKey int64) float64

// 关联 MI 计数
func (s *milestoneService) countRelatedMIs(ctx context.Context, milestoneBizKey int64) int64
```

## DTO Definitions

### Milestone DTOs (NEW)

```go
// dto/milestone_dto.go
package dto

type MilestoneMapCreateReq struct {
	MapName          string `json:"mapName" binding:"required,max=100"`
	MapDesc          string `json:"mapDesc"`
	AssigneeBizKey   int64  `json:"assigneeBizKey" binding:"required"`
	PlanStartDate string `json:"planStartDate"`
	ExpectedEndDate   string `json:"expectedEndDate"`
}

type MilestoneMapUpdateReq struct {
	MapName          *string `json:"mapName"`
	MapDesc          *string `json:"mapDesc"`
	AssigneeBizKey   *int64  `json:"assigneeBizKey"`
	PlanStartDate *string `json:"planStartDate"`
	ExpectedEndDate   *string `json:"expectedEndDate"`
}

type MilestoneMapFilter struct {
	Name        *string `form:"name" json:"name"`
	AssigneeKey *string `form:"assigneeKey" json:"assigneeKey"`
	Status      *string `form:"status" json:"status"`
}

type MilestoneCreateReq struct {
	MilestoneName   string  `json:"milestoneName" binding:"required,max=100"`
	ExpectedEndDate *string `json:"expectedEndDate" binding:"required"`
	MilestoneDesc   string  `json:"milestoneDesc"`
}

type MilestoneUpdateReq struct {
	MilestoneName   *string  `json:"milestoneName"`
	ExpectedEndDate *string  `json:"expectedEndDate"`
	MilestoneDesc   *string  `json:"milestoneDesc"`
}

type MilestoneTeamFilter struct {
	Name             *string `form:"name"`
	Status           *string `form:"status"`
	ExcludeCancelled *bool   `form:"excludeCancelled"`
}
```

### MainItem DTOs (MODIFIED)

```go
// dto/item_dto.go — MainItemFilter 新增字段
type MainItemFilter struct {
	Statuses    []string `form:"status" json:"status"`
	Priority    string   `form:"priority" json:"priority"`
	AssigneeKey *string  `form:"assigneeKey" json:"assigneeKey"`
	IsKeyItem   *bool    `form:"isKeyItem"`
	Archived    bool     `form:"archived"`
	MilestoneKey *string `form:"milestoneKey" json:"milestoneKey"` // NEW: 按里程碑筛选
}

// dto/item_dto.go — MainItemCreateReq 新增字段
type MainItemCreateReq struct {
	Title           string  `json:"title" binding:"required,max=100"`
	Description     string  `json:"description"`
	Priority        string  `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
	AssigneeKey     string  `json:"assigneeKey" binding:"required"`
	StartDate       *string `json:"startDate" binding:"required"`
	ExpectedEndDate *string `json:"expectedEndDate" binding:"required"`
	IsKeyItem       bool    `json:"isKeyItem"`
	MilestoneKey    *string `json:"milestoneKey"` // NEW: optional, bizKey of milestone
}

// dto/item_dto.go — MainItemUpdateReq 新增字段
type MainItemUpdateReq struct {
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	Priority        *string `json:"priority"`
	AssigneeKey     *string `json:"assigneeKey"`
	StartDate       *string `json:"startDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
	IsKeyItem       *bool   `json:"isKeyItem"`
	MilestoneKey    *string `json:"milestoneKey"` // NEW: nil=no change, non-nil=set (empty string=unbind)
}
```

## VO Definitions

### Milestone VOs (NEW)

```go
// vo/milestone_vo.go
package vo

type MilestoneMapVO struct {
	BizKey          string  `json:"bizKey"`
	TeamKey         string  `json:"teamKey"`
	CreatorKey      string  `json:"creatorKey"`
	CreatorName     string  `json:"creatorName"`
	AssigneeKey     string  `json:"assigneeKey"`
	AssigneeName    string  `json:"assigneeName"`
	MapName         string  `json:"mapName"`
	MapDesc         string  `json:"mapDesc"`
	MapStatus       string  `json:"mapStatus"`
	StatusName      string  `json:"statusName"`
	PlanStartDate *string `json:"planStartDate"`
	ExpectedEndDate   *string `json:"expectedEndDate"`
	MilestoneCount  int     `json:"milestoneCount"`
	ItemCount       int     `json:"itemCount"`
	OverallProgress float64 `json:"overallProgress"`
	CreateTime      string  `json:"createTime"`
	DbUpdateTime    string  `json:"dbUpdateTime"`
}

type MilestoneVO struct {
	BizKey          string  `json:"bizKey"`
	TeamKey         string  `json:"teamKey"`
	MilestoneMapKey string  `json:"milestoneMapKey"`
	MilestoneName   string  `json:"milestoneName"`
	MilestoneDesc   string  `json:"milestoneDesc"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	MilestoneStatus string  `json:"milestoneStatus"`
	StatusName      string  `json:"statusName"`
	Completion      float64 `json:"completion"`
	RelatedMICount  int     `json:"relatedMICount"`
	CreateTime      string  `json:"createTime"`
	DbUpdateTime    string  `json:"dbUpdateTime"`
}
```

### MainItemVO (MODIFIED)

```go
// vo/item_vo.go — MainItemVO 新增字段
type MainItemVO struct {
	// ... existing fields ...
	MilestoneKey  *string `json:"milestoneKey"`  // NEW
	MilestoneName string  `json:"milestoneName"` // NEW: enrichment from batch lookup
}
```

### MainItem List Enrichment

MainItemService.List 返回结果后，在 service 或 handler 层执行 enrichment：

1. 收集所有非空 `MilestoneKey` 对应的 bizKey
2. 调用 `MilestoneRepo.FindBatchByBizKeys(ctx, bizKeys)` 获取 `map[int64]*model.Milestone`
3. 设置 `voItem.MilestoneName = milestone.MilestoneName`
4. 里程碑已软删除时，`MilestoneName` 设为 `"--"`

## Router Wiring

### Dependencies Struct (MODIFIED)

```go
// handler/router.go — Dependencies 新增字段
type Dependencies struct {
	// ... existing fields ...
	MilestoneMap *MilestoneMapHandler // NEW
	Milestone    *MilestoneHandler    // NEW
}
```

### Route Registration (MODIFIED)

在 `SetupRouter` 的 `teamsGroup` 中，Item Pool 路由之后添加：

```go
// Milestone maps
teamsGroup.POST("/milestone-maps", deps.perm("milestone:create"), deps.MilestoneMap.Create)
teamsGroup.GET("/milestone-maps", deps.perm("milestone:read"), deps.MilestoneMap.List)
teamsGroup.GET("/milestone-maps/:mapId", deps.perm("milestone:read"), deps.MilestoneMap.Get)
teamsGroup.PUT("/milestone-maps/:mapId", deps.perm("milestone:update"), deps.MilestoneMap.Update)
teamsGroup.PUT("/milestone-maps/:mapId/status", deps.perm("milestone:update"), deps.MilestoneMap.ChangeStatus)
teamsGroup.GET("/milestone-maps/:mapId/available-transitions", deps.perm("milestone:read"), deps.MilestoneMap.AvailableTransitions)
teamsGroup.DELETE("/milestone-maps/:mapId", deps.perm("milestone:delete"), deps.MilestoneMap.Delete)

// Milestones (nested under milestone-map)
teamsGroup.POST("/milestone-maps/:mapId/milestones", deps.perm("milestone:create"), deps.Milestone.Create)
teamsGroup.GET("/milestone-maps/:mapId/milestones", deps.perm("milestone:read"), deps.Milestone.ListByMap)

// Milestones (direct access)
teamsGroup.GET("/milestones", deps.perm("milestone:read"), deps.Milestone.ListByTeam)
teamsGroup.GET("/milestones/:milestoneId", deps.perm("milestone:read"), deps.Milestone.Get)
teamsGroup.PUT("/milestones/:milestoneId", deps.perm("milestone:update"), deps.Milestone.Update)
teamsGroup.PUT("/milestones/:milestoneId/status", deps.perm("milestone:update"), deps.Milestone.ChangeStatus)
teamsGroup.GET("/milestones/:milestoneId/available-transitions", deps.perm("milestone:read"), deps.Milestone.AvailableTransitions)
teamsGroup.DELETE("/milestones/:milestoneId", deps.perm("milestone:delete"), deps.Milestone.Delete)
```

## Handler Pattern

Handler 遵循现有模式（参考 `item_pool_handler.go`）：

- 构造函数 panic-on-nil 校验
- 使用 `middleware.GetTeamBizKey(c)` / `middleware.GetUserBizKey(c)` 获取上下文
- 路由参数通过 `pkgHandler.ResolveBizKey(c, "mapId", resolver)` 解析
- 请求绑定通过 `c.ShouldBindJSON(&req)` / `c.ShouldBindQuery(&filter)`
- 分页通过 `dto.ApplyPaginationDefaults(page, pageSize)`
- 响应通过 `apperrors.RespondOK(c, data)` / `apperrors.RespondError(c, err)`
- VO 转换在 handler 层的 `buildXxxVOs` 函数中完成

## Cross-Layer Data Map

| Field Name | Storage Layer | Backend Model | API/DTO/VO | Frontend Type | Validation Rule |
|------------|---------------|---------------|------------|---------------|-----------------|
| bizKey | INTEGER NOT NULL | int64 (BaseModel.BizKey) | string (FormatID) | string | 自动生成，只读 |
| teamKey | INTEGER NOT NULL | int64 | string (FormatID) | string | 从路由上下文获取 |
| creatorKey | INTEGER NOT NULL | int64 | string (FormatID) | string | 从 Auth 中间件获取，创建后不变 |
| assigneeKey | INTEGER NOT NULL | int64 | string (FormatID) | string | 必填，必须是当前团队成员 |
| mapName | VARCHAR(100) NOT NULL | string | string | string | 必填，1-100 字符，支持模糊搜索 |
| mapDesc | VARCHAR(2000) | string | string | string | 可选 |
| mapStatus | VARCHAR(20) NOT NULL DEFAULT 'planning' | string | string | string | 必须是有效状态码 |
| planStartDate | DATE DEFAULT NULL | *time.Time | *string (FormatTimePtr) | string \| null | 可选，不得晚于 expectedEndDate |
| expectedEndDate | DATE DEFAULT NULL | *time.Time | *string (FormatTimePtr) | string \| null | 可选，不得早于 planStartDate |
| milestoneMapKey | INTEGER NOT NULL | int64 | string (FormatID) | string | 必须指向存在的里程碑图 |
| milestoneName | VARCHAR(100) NOT NULL | string | string | string | 必填，1-100 字符 |
| milestoneDesc | VARCHAR(2000) | string | string | string | 可选 |
| expectedEndDate | DATE | *time.Time | *string (FormatTimePtr) | string \| null | 必填 |
| milestoneStatus | VARCHAR(20) NOT NULL DEFAULT 'not_started' | string | string | string | 必须是有效状态码 |
| milestoneKey (MainItem) | INTEGER DEFAULT NULL | *int64 | *string (FormatIDPtr) | string \| null | 可选，必须指向存在的里程碑 |
| completion (computed) | — | float64 | number | number | 0-100，GET 时计算 |
| overallProgress (computed) | — | float64 | number | number | 0-100，GET 时计算 |

## Integration Specs

### Integration: MilestoneFilter → ItemViewPage (UF-4)

- **Target**: `frontend/src/pages/item-view/` 组件
- **API Module**: `frontend/src/api/milestones.ts` — `listByTeam(teamId, {excludeCancelled: true})`
- **Insertion**: 筛选栏 `StatusTagFilter` + `MemberSelect` 右侧，新增里程碑下拉框
- **Data**: MainItemFilter 增加 `milestoneKey` 参数，传递给 `GET /main-items` API

### Integration: MilestoneSelector → EditMainItemDialog (UF-5)

- **Target**: `frontend/src/pages/item-view/EditMainItemDialog.tsx` 或 `frontend/src/pages/main-item-detail/EditMainItemDialog.tsx`
- **API Module**: `frontend/src/api/milestones.ts` — `listByTeam(teamId, {excludeCancelled: true})`
- **Insertion**: 编辑弹窗"负责人"字段下方，新增"所属里程碑"下拉框
- **Data**: CreateReq/UpdateReq 增加 `milestoneKey` 字段

### Business Rules: 状态转换约束

三条规则覆盖三层实体（里程碑图 → 里程碑 → 主事项）之间的状态联动。

#### BR-1: 里程碑终态前置条件

里程碑只有在所有关联主事项均处于终态时，才允许被标记为 `completed`。切换至 `cancelled` 时自动解绑所有关联主事项，不受主事项状态限制。

1. **目标状态为 completed 时校验**：当 `UpdateReq.MilestoneStatus` 为 `completed` 时，查询该里程碑下所有关联 `MainItem`，若存在非终态主事项，返回 `400 Bad Request`
2. **目标状态为 cancelled 时自动解绑**：切换至 `cancelled` 时，在事务内将所有关联 `MainItem` 的 `milestone_key` 置空，无需校验主事项状态
3. **空里程碑可直接终态**：无关联主事项的里程碑可自由切换为 `completed`
4. **非终态→非终态不校验**：其他状态转换（如 `not_started` → `in_progress`）不触发此校验

```go
func (s *Service) UpdateMilestoneStatus(ctx context.Context, milestoneID uint, newStatus string) error {
    if newStatus == "completed" {
        items, _ := s.itemRepo.FindByMilestoneKey(ctx, milestoneID)
        for _, item := range items {
            if !MainItemStatuses[item.ItemStatus].Terminal {
                return ErrMilestoneHasNonTerminalItems
            }
        }
    }
    if newStatus == "cancelled" {
        // auto-unbind all related MIs in transaction
        _ = s.itemRepo.ClearMilestoneKeyByMilestone(ctx, milestoneID)
    }
    // proceed with status update
}
```

#### BR-2: 里程碑图终态前置条件

里程碑图只有在所有里程碑均处于终态时，才允许被标记为终态（`completed`）。

1. **目标状态为 completed 时校验**：查询该里程碑图下所有 `Milestone`，若存在非终态里程碑（`not_started`/`in_progress`/`completed` 三者中 `completed` 为非终态，`cancelled` 为终态），返回 `400 Bad Request`
2. **空里程碑图可直接终态**：无关联里程碑的里程碑图可自由切换为 `completed`
3. **非终态→非终态不校验**：其他状态转换（如 `planning` → `reviewed`）不触发此校验
4. **回退转换不校验**：如 `executing` → `ready` 不触发此校验

```go
func (s *Service) UpdateMilestoneMapStatus(ctx context.Context, mapID uint, newStatus string) error {
    if MilestoneMapStatuses[newStatus].Terminal {
        milestones, _ := s.milestoneRepo.FindByMapKey(ctx, mapID)
        for _, ms := range milestones {
            if !MilestoneStatuses[ms.MilestoneStatus].Terminal {
                return ErrMapHasNonTerminalMilestones
            }
        }
    }
    // proceed with status update
}
```

#### BR-3: 主事项移动约束

更新 `MainItemUpdateReq.MilestoneKey` 时（即分配/移动主事项到里程碑），后端 Service 层需校验：

1. **主事项不能处于终态**：`MainItem.ItemStatus` 为终态（`completed`/`closed`）时，拒绝修改 `milestone_key`，返回 `400 Bad Request`
2. **目标里程碑不能处于终态**：目标 `Milestone.MilestoneStatus` 为终态（`cancelled`）时，拒绝绑定，返回 `400 Bad Request`
3. 校验顺序：先检查主事项状态 → 再检查目标里程碑状态

```go
func (s *Service) UpdateMainItemMilestone(ctx context.Context, itemID uint, milestoneKey *int64) error {
    item, _ := s.itemRepo.FindByID(ctx, itemID)
    if MainItemStatuses[item.ItemStatus].Terminal {
        return ErrTerminalItemCannotMove
    }
    if milestoneKey != nil && *milestoneKey != 0 {
        ms, _ := s.milestoneRepo.FindByBizKey(ctx, *milestoneKey)
        if MilestoneStatuses[ms.MilestoneStatus].Terminal {
            return ErrTerminalMilestoneCannotReceive
        }
    }
    // proceed with update
}
```

#### 联动总结

```
主事项 (MI)                    里程碑 (MS)              里程碑图 (Map)
─────────────────────────────────────────────────────────────────────────
MI 终态 → 不可移动            MS 终态 → 不可接收 MI    Map 终态 → 不可添加 MS
MI 全部终态 ← MS 转 completed  MS 全部终态 ← Map 可转终态
MS 转 cancelled → 自动解绑 MI（不受 MI 状态限制）
```

层级关系：Map → MS → MI。MS 转 `completed` 要求子层全终态；MS 转 `cancelled` 自动解绑子层不受限制；Map 转终态要求子层全终态；父层终态后拒绝子层变动。

#### BR-4: 删除约束

**里程碑图**：仅 `planning`（规划中）状态的里程碑图允许删除。其他状态（`reviewed`/`ready`/`executing`/`completed`）拒绝删除，返回 `400 Bad Request`。

```go
func (s *Service) DeleteMilestoneMap(ctx context.Context, mapID uint) error {
    m, _ := s.mapRepo.FindByID(ctx, mapID)
    if m.MapStatus != "planning" {
        return ErrMapCannotDeleteNonPlanning
    }
    // proceed with delete
}
```

**里程碑**：仅 `not_started`（未开始）和 `cancelled`（已取消）状态的里程碑允许删除。其他状态（`in_progress`/`completed`）拒绝删除，返回 `400 Bad Request`。

```go
func (s *Service) DeleteMilestone(ctx context.Context, milestoneID uint) error {
    ms, _ := s.milestoneRepo.FindByID(ctx, milestoneID)
    if ms.MilestoneStatus != "not_started" && ms.MilestoneStatus != "cancelled" {
        return ErrMilestoneCannotDelete
    }
    // proceed with delete
}
```

### Integration: MilestoneColumn → TableViewPage (UF-6)

- **Target**: `frontend/src/pages/TableViewPage.tsx`
- **API Module**: `frontend/src/api/views.ts` — 表格数据已含 `milestoneName`
- **Insertion**: 表格列定义中，"标题"列和"优先级"列之间新增"里程碑"列
- **Data**: `TableRow` DTO 新增 `MilestoneName` 字段

## Error Handling

遵循现有模式：
- Handler 层 catch service 错误，通过 `apperrors.RespondError(c, err)` 转换为 HTTP 响应
- Service 层返回业务错误（`apperrors.MapNotFound(err, apperrors.ErrXxx)`）
- Repository 层返回 GORM 原始错误，由 service 层包装

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| INVALID_PARAMS | 400 | 名称/日期校验失败 |
| NOT_FOUND | 404 | 里程碑图/里程碑不存在 |
| INVALID_STATUS | 422 | 状态转换不合法 |
| FORBIDDEN | 403 | 无权限 |

## Testing Strategy

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| Repository | Integration | Go testing + GORM | CRUD、筛选、软删除、事务内解绑 | 80% |
| Service | Unit (mock repo) | Go testing | 状态转换校验、完成度计算、权限检查 | 85% |
| Handler | Integration | httptest | API 契约、请求校验、错误响应 | 80% |
| Frontend API | Unit | Vitest | API 函数调用参数和返回值 | 80% |
| Frontend Components | Unit | Vitest + RTL | 组件渲染、交互、状态切换 | 75% |
| E2E | Integration | Playwright | 完整用户流程 | 关键路径 |

### Key Test Scenarios

1. 创建里程碑图：名称校验（空/超长/正常）→ 状态默认 planning → 返回正确 VO
2. 创建里程碑：名称校验 → 日期校验 → 关联到正确的里程碑图
3. 状态转换：MilestoneMap 5 态合法/非法转换 → Milestone 4 态合法/非法转换
4. 完成度计算：空里程碑=0 → 1 个 MI=MI.completion → 多个 MI=平均值
5. 删除里程碑：软删除 + 事务内解绑关联 MI
6. 删除里程碑图：级联软删除所有里程碑 + 解绑所有 MI
7. 里程碑取消：自动解绑关联 MI
8. MainItem 绑定/解绑：milestone_key 正确更新，边界情况（指向软删除的里程碑）

## Security Considerations

1. **RBAC**: 4 个独立权限码（milestone:create/read/update/delete），通过 `RequirePermission` 中间件强制
2. **TeamScopeMiddleware**: 所有 API 嵌套在 teams group 下，自动注入 teamBizKey 并验证团队成员身份
3. **状态机校验**: Service 层调用 `IsValidTransition()` 校验，不依赖前端校验
4. **BizKey 不暴露 ID**: Handler 层通过 `ResolveBizKey` 解析，内部 ID 不暴露
5. **越权操作**: Service 层校验 entity.TeamKey == teamBizKey

## PRD Coverage Map

| PRD Requirement / AC | Design Component | Interface / Model |
|----------------------|------------------|-------------------|
| Story 1: PM 创建里程碑图 | MilestoneMapHandler.Create | POST /milestone-maps, MilestoneMapCreateReq |
| Story 2: PM 编辑里程碑图 | MilestoneMapHandler.Update | PUT /milestone-maps/:mapId, MilestoneMapUpdateReq |
| Story 3: PM 切换里程碑图状态 | MilestoneMapHandler.ChangeStatus + MilestoneMapTransitions | PUT /milestone-maps/:mapId/status |
| Story 4a: PM 创建里程碑 | MilestoneHandler.Create | POST /milestone-maps/:mapId/milestones, MilestoneCreateReq |
| Story 4b: PM 编辑里程碑 | MilestoneHandler.Update | PUT /milestones/:milestoneId, MilestoneUpdateReq |
| Story 4c: PM 删除里程碑 | MilestoneService.Delete + 事务解绑 | DELETE /milestones/:milestoneId |
| Delete MilestoneMap (隐含) | MilestoneMapService.Delete + 级联软删除 | DELETE /milestone-maps/:mapId |
| Story 5: PM 切换里程碑状态 | MilestoneHandler.ChangeStatus + MilestoneTransitions | PUT /milestones/:milestoneId/status |
| Story 6: PM 分配事项到里程碑 | MainItemUpdateReq.MilestoneKey | PUT /main-items/:itemId |
| Story 6: PM 解绑事项 | MainItemUpdateReq.MilestoneKey=null | PUT /main-items/:itemId |
| Story 7: 详情面板管理事项 | MilestoneService (completion, unbind) | GET /milestones/:id + MainItem update |
| Story 7: 快速添加事项 | CreateMainItemApi (milestoneKey pre-filled) | POST /main-items |
| Story 8: 里程碑图两级视图 | MilestonesPage (frontend) | GET /milestone-maps, GET /milestone-maps/:id/milestones |
| Story 8: 空状态提示 | MilestonesPage empty state | GET /milestone-maps 返回空列表 |
| Story 8: 时间线缩放 | MilestonesPage TimelineView | 纯前端交互 |
| Story 9: 权限控制 | Permission checks | 所有 milestone API |
| Story 10: 表格视图里程碑列 | TableViewPage integration | MainItem.milestoneKey + milestoneName |
| NFR: 完成度实时计算 | calcCompletion/calcOverallProgress | GET 时计算 |
| NFR: 软删除+事务解绑 | MilestoneService.Delete (transaction) | MilestoneRepo.SoftDelete + MainItemRepo batch update |
| NFR: 权限独立RBAC | permissions/codes.go Registry | milestone:create/read/update/delete |
| NFR: 列表<300ms | Index on (team_key, map_status) | schema.sql |
| NFR: 时间线<500ms | ListByMap (no pagination) + batch MI lookup | MilestoneRepo.ListByMap |

## Appendix: Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| 完成度持久化 | 查询更快 | 每次MI变更需更新关联里程碑和图，增加写路径复杂度 | PRD 要求实时计算，数据量小（<200 MI） |
| 里程碑嵌套在 MilestoneMap API 内 | 路由更简洁 | 违反现有扁平路由模式 | Milestone 有独立 CRUD + 状态切换 |
| DDL 外键 | 数据库级引用完整性 | 现有项目统一不使用 DDL FK | 遵循约定：No FOREIGN KEY in DDL |
