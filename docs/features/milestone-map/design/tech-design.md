---
created: 2026-05-12
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: 里程碑图

## Overview

新增 MilestoneMap 和 Milestone 两个层级实体，复用现有的 5 层架构（Model → Repository → Service → Handler → Router），遵循 BaseModel、状态机、权限码等既有模式。前端新增 `/milestones` 页面（两级视图），并在 3 个现有页面集成里程碑维度。

技术方案的核心约束：
- 完成度实时计算（GET 时不持久化）
- 软删除里程碑时事务内解绑关联 MI
- 状态机转换校验复用 `status/transition.go` 模式
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
  └── milestone_dto.go          [NEW]

Status (backend/internal/pkg/status/)
  └── milestone_status.go       [NEW] status defs + transitions

Permission (backend/internal/pkg/permissions/)
  └── codes.go                  [MODIFIED] add 4 milestone codes

Router (backend/internal/handler/router.go) [MODIFIED]

Frontend:
  src/api/milestones.ts         [NEW]
  src/types/index.ts            [MODIFIED]
  src/pages/MilestonesPage/     [NEW]
  src/components/               [MODIFIED] 3 integrations
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
| `pkg/status` | Internal | 状态机注册 |
| `pkg/permissions` | Internal | 权限码注册 |
| `middleware` | Internal | AuthMiddleware + TeamScopeMiddleware + RequirePermission |
| `vo` | Internal | VO 转换（FormatID, FormatTimePtr） |
| `dto` | Internal | 分页、筛选 DTO |

## Interfaces

### MilestoneMapRepo

```go
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
type MilestoneRepo interface {
    Create(ctx context.Context, m *model.Milestone) error
    FindByID(ctx context.Context, id uint) (*model.Milestone, error)
    FindByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error)
    FindByBizKeys(ctx context.Context, bizKeys []int64) (map[int64]*model.Milestone, error)
    Update(ctx context.Context, m *model.Milestone, fields map[string]interface{}) error
    ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error)
    ListByTeam(ctx context.Context, teamBizKey int64, excludeCancelled bool) ([]model.Milestone, error)
    SoftDelete(ctx context.Context, id uint) error
    DeleteByMap(ctx context.Context, milestoneMapID uint) error
}
```

### MilestoneMapService

```go
type MilestoneMapService interface {
    Create(ctx context.Context, teamBizKey int64, req dto.MilestoneMapCreateReq) (*model.MilestoneMap, error)
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
type MilestoneService interface {
    Create(ctx context.Context, teamBizKey int64, milestoneMapBizKey int64, req dto.MilestoneCreateReq) (*model.Milestone, error)
    Update(ctx context.Context, teamBizKey int64, milestoneID uint, req dto.MilestoneUpdateReq) error
    Get(ctx context.Context, milestoneID uint) (*model.Milestone, error)
    GetByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error)
    ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error)
    ListByTeam(ctx context.Context, teamBizKey int64, excludeCancelled bool) ([]model.Milestone, error)
    Delete(ctx context.Context, teamBizKey int64, milestoneID uint) error
    ChangeStatus(ctx context.Context, teamBizKey int64, milestoneID uint, newStatus string) (*model.Milestone, error)
    AvailableTransitions(ctx context.Context, milestoneID uint) ([]string, error)
}
```

### Computed Fields (Service Layer)

完成度计算在 GET 时实时查询，不持久化：

```go
// Milestone 完成度：关联 MI 的 completion 平均值
func (s *milestoneService) calcCompletion(ctx context.Context, milestoneBizKey int64) float64

// MilestoneMap 整体进度：所有关联 MI 的 completion 平均值
func (s *milestoneMapService) calcOverallProgress(ctx context.Context, milestoneMapBizKey int64) float64

// 关联 MI 计数
func (s *milestoneService) countRelatedMIs(ctx context.Context, milestoneBizKey int64) int64
```

### MainItemService (Modified)

MainItemCreateReq and MainItemUpdateReq gain a MilestoneKey field; List responses are enriched with milestoneName via batch lookup.

```go
// --- New Milestone DTO structs (dto/milestone_dto.go) ---

type MilestoneMapCreateReq struct {
    MapName string `json:"mapName" binding:"required,max=100"`
    MapDesc string `json:"mapDesc"`
}

type MilestoneMapUpdateReq struct {
    MapName *string `json:"mapName"`
    MapDesc *string `json:"mapDesc"`
}

type MilestoneCreateReq struct {
    MilestoneName  string  `json:"milestoneName" binding:"required,max=100"`
    ExpectedEndDate *string `json:"expectedEndDate" binding:"required"`
}

type MilestoneUpdateReq struct {
    MilestoneName   *string  `json:"milestoneName"`
    ExpectedEndDate *string  `json:"expectedEndDate"`
}

type MilestoneMapFilter struct {
    Status *string `form:"status"`
}
```

```go
// --- Modifications to existing DTOs (dto/item_dto.go) ---

// MainItemCreateReq gains one field:
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

// MainItemUpdateReq gains one field:
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

```go
// --- List response enrichment (service layer wiring) ---

// After MainItemService.List returns []MainItem VO items, the service
// collects all non-nil MilestoneKey values, calls
//   MilestoneRepo.FindByBizKeys(ctx, milestoneBizKeys)
// to get a map[int64]*model.Milestone, then sets
//   voItem.MilestoneName = milestone.MilestoneName
// on each item. If the milestone has been soft-deleted (not found in map),
// MilestoneName is set to "--".
```

## Data Models

> Full database design in separate files.

**ER Diagram**: design/er-diagram.md
**SQL Schema**: design/schema.sql

### Field Quick Reference

| Model | Key Fields | Notes |
|-------|------------|-------|
| MilestoneMap | map_name, map_desc, map_status, team_key | 5 态：planning→reviewed→ready→executing→completed |
| Milestone | milestone_name, expected_end_date, milestone_status, milestone_map_key, team_key | 4 态：not_started→in_progress→completed/cancelled |
| MainItem (modified) | +milestone_key | 可空 FK，引用 pmw_milestones.biz_key |

## Error Handling

### Error Types & Codes

| Error Code | Name | Description | HTTP Status |
|------------|------|-------------|-------------|
| INVALID_PARAMS | ParamError | 名称/日期校验失败 | 400 |
| NOT_FOUND | NotFoundError | 里程碑图/里程碑不存在 | 404 |
| CONFLICT | ConflictError | 并发编辑冲突（里程碑） | 409 |
| INVALID_STATUS | StatusError | 状态转换不合法 | 422 |
| FORBIDDEN | ForbiddenError | 无权限 | 403 |

### Propagation Strategy

遵循现有模式：
- Handler 层 catch service 错误，转换为 HTTP 响应
- Service 层返回业务错误（通过 `pkg/errors` 或 sentinel errors）
- Repository 层返回 GORM 原始错误，由 service 层包装

## Cross-Layer Data Map

| Field Name | Storage Layer | Backend Model | API/DTO | Frontend Type | Validation Rule |
|------------|---------------|---------------|---------|---------------|-----------------|
| bizKey | BIGINT NOT NULL | int64 (BaseModel.BizKey) | string (FormatID) | string | 自动生成，只读 |
| teamKey | BIGINT NOT NULL | int64 | string (FormatID) | string | 从路由上下文获取 |
| mapName | VARCHAR(100) NOT NULL | string | string | string | 必填，1-100 字符 |
| mapDesc | TEXT | string | string | string | 可选 |
| mapStatus | VARCHAR(20) NOT NULL DEFAULT 'planning' | string | string | string | 必须是有效状态码 |
| milestoneMapKey | BIGINT NOT NULL | int64 | string (FormatID) | string | 必须指向存在的里程碑图 |
| milestoneName | VARCHAR(100) NOT NULL | string | string | string | 必填，1-100 字符 |
| expectedEndDate | DATETIME | *time.Time | *string (FormatTimePtr) | string \| null | 必填 |
| milestoneStatus | VARCHAR(20) NOT NULL DEFAULT 'not_started' | string | string | string | 必须是有效状态码 |
| milestoneKey (MainItem) | BIGINT DEFAULT NULL | *int64 | *string (FormatIDPtr) | string \| null | 可选，必须指向存在的里程碑 |
| completion (computed) | — | float64 | number | number | 0-100，GET 时计算 |
| overallProgress (computed) | — | float64 | number | number | 0-100，GET 时计算 |

## Integration Specs

### Integration: MilestoneFilter → /items page (UF-4)

- **Target File**: `frontend/src/pages/ItemViewPage/` (or equivalent items list page)
- **Insertion Point**: 筛选栏区域，现有"状态"和"负责人"筛选器右侧
- **Data Source**: `GET /teams/:teamId/milestones` API，下拉选项排除 cancelled 里程碑

### Integration: MilestoneSelector → /items/:mainItemId page (UF-5)

- **Target File**: `frontend/src/components/CreateMainItemDialog/` (或主事项编辑弹窗)
- **Insertion Point**: 编辑弹窗中"负责人"字段下方，新增"所属里程碑"下拉框
- **Data Source**: `GET /teams/:teamId/milestones` API，下拉选项排除 cancelled 里程碑

### Integration: MilestoneColumn → /table page (UF-6)

- **Target File**: `frontend/src/pages/TableViewPage/` (或表格视图组件)
- **Insertion Point**: 表格列定义中，"标题"列和"优先级"列之间
- **Data Source**: MainItem 的 milestoneKey 字段 + milestoneName（由 list API 附加）

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| Repository | Integration | Go testing + GORM | CRUD、筛选、软删除、事务内解绑 | 80% |
| Service | Unit (mock repo) | Go testing | 状态转换校验、完成度计算、权限检查 | 85% |
| Handler | Integration | httptest | API 契约、请求校验、错误响应 | 80% |
| Frontend API | Unit | Vitest | API 函数调用参数和返回值 | 80% |
| Frontend Components | Unit | Vitest + RTL | 组件渲染、交互、状态切换 | 75% |
| E2E | Integration | Playwright | 完整用户流程（创建图→创建里程碑→绑定MI→状态切换） | 关键路径 |

### Key Test Scenarios

1. **创建里程碑图**: 名称校验（空/超长/正常）→ 状态默认 planning → 返回正确 VO
2. **创建里程碑**: 名称校验 → 日期校验 → 关联到正确的里程碑图
3. **状态转换**: MilestoneMap 5 态合法/非法转换 → Milestone 4 态合法/非法转换
4. **完成度计算**: 空里程碑=0 → 1 个 MI=MI.completion → 多个 MI=平均值
5. **删除里程碑**: 软删除 + 事务内解绑关联 MI
6. **删除里程碑图**: 级联软删除所有里程碑 + 解绑所有 MI
7. **里程碑取消**: 自动解绑关联 MI
8. **MainItem 绑定/解绑**: milestone_key 正确更新，边界情况（指向软删除的里程碑）

### Overall Coverage Target

80%

## Security Considerations

### Threat Model

1. **越权操作**: 用户在无权限情况下操作里程碑
2. **跨团队访问**: 用户访问其他团队的里程碑
3. **非法状态转换**: 用户绕过前端直接调用 API 进行非法状态变更

### Mitigations

1. **RBAC**: 4 个独立权限码（milestone:create/read/update/delete），通过 RequirePermission 中间件强制
2. **TeamScopeMiddleware**: 所有里程碑 API 嵌套在 teams group 下，自动注入 teamBizKey 并验证团队成员身份
3. **状态机校验**: Service 层调用 `status.IsValidTransition()` 校验，不依赖前端校验
4. **BizKey 不暴露 ID**: Handler 层通过 BizKey 解析，内部 ID 不暴露

### Permission-to-Operation Mapping

| Permission Code | Guarded Operations (Endpoint) |
|-----------------|-------------------------------|
| `milestone:create` | POST `/teams/:teamId/milestone-maps` (Create MilestoneMap) |
|                 | POST `/teams/:teamId/milestone-maps/:mapId/milestones` (Create Milestone) |
| `milestone:read`  | GET `/teams/:teamId/milestone-maps` (List MilestoneMaps) |
|                 | GET `/teams/:teamId/milestone-maps/:mapId` (Get MilestoneMap) |
|                 | GET `/teams/:teamId/milestone-maps/:mapId/available-transitions` |
|                 | GET `/teams/:teamId/milestone-maps/:mapId/milestones` (List Milestones by Map) |
|                 | GET `/teams/:teamId/milestones` (List Milestones by Team) |
|                 | GET `/teams/:teamId/milestones/:milestoneId` (Get Milestone) |
|                 | GET `/teams/:teamId/milestones/:milestoneId/available-transitions` |
| `milestone:update` | PUT `/teams/:teamId/milestone-maps/:mapId` (Update MilestoneMap) |
|                 | PUT `/teams/:teamId/milestone-maps/:mapId/status` (Change MilestoneMap Status) |
|                 | PUT `/teams/:teamId/milestones/:milestoneId` (Update Milestone) |
|                 | PUT `/teams/:teamId/milestones/:milestoneId/status` (Change Milestone Status) |
| `milestone:delete` | DELETE `/teams/:teamId/milestone-maps/:mapId` (Delete MilestoneMap) |
|                 | DELETE `/teams/:teamId/milestones/:milestoneId` (Delete Milestone) |

> Note: `ChangeStatus` uses `milestone:update` (not a separate code). `Delete` requires `milestone:delete`. MainItem binding/unbinding via `milestoneKey` field uses the existing `main_item:update` permission, not milestone permissions.

## PRD Coverage Map

| PRD Requirement / AC | Design Component | Interface / Model |
|----------------------|------------------|-------------------|
| Story 1: PM 创建里程碑图 | MilestoneMapHandler.Create | POST /milestone-maps, MilestoneMapCreateReq |
| Story 2: PM 编辑里程碑图 | MilestoneMapHandler.Update | PUT /milestone-maps/:mapId, MilestoneMapUpdateReq |
| Story 3: PM 切换里程碑图状态 | MilestoneMapHandler.ChangeStatus + status.MilestoneMapTransitions | PUT /milestone-maps/:mapId/status |
| Story 4a: PM 创建里程碑 | MilestoneHandler.Create | POST /milestone-maps/:mapId/milestones, MilestoneCreateReq |
| Story 4b: PM 编辑里程碑 | MilestoneHandler.Update | PUT /milestones/:milestoneId, MilestoneUpdateReq |
| Story 4c: PM 删除里程碑 | MilestoneService.Delete + 事务解绑 | DELETE /milestones/:milestoneId |
| Delete MilestoneMap (隐含需求) | MilestoneMapService.Delete + 级联软删除 | DELETE /milestone-maps/:mapId，事务内级联软删除所有里程碑并解绑 MI |
| Story 5: PM 切换里程碑状态 | MilestoneHandler.ChangeStatus + status.MilestoneTransitions | PUT /milestones/:milestoneId/status |
| Story 6: PM 分配事项到里程碑 | MainItemUpdateReq.MilestoneKey | PUT /main-items/:itemId |
| Story 6: PM 解绑事项 | MainItemUpdateReq.MilestoneKey=null | PUT /main-items/:itemId |
| Story 7: 详情面板管理事项 | MilestoneService (completion, unbind via MainItem update) | GET /milestones/:id + MainItem update |
| Story 7: 快速添加事项 | CreateMainItemApi (milestoneKey pre-filled) | POST /main-items |
| Story 7: 行内解绑 | MainItemUpdateReq.MilestoneKey=null | PUT /main-items/:itemId |
| Story 8: 里程碑图两级视图 | MilestonesPage (frontend) | GET /milestone-maps, GET /milestone-maps/:id/milestones |
| Story 8: 空状态提示 (0 个里程碑图) | MilestonesPage empty state component (frontend, UF-1) | GET /milestone-maps 返回空列表 → 显示"暂无里程碑图"空状态提示 + 创建按钮 |
| Story 8: 时间线缩放 (周/月/季) | MilestonesPage TimelineView zoom controls (frontend) | 纯前端交互，无后端变更 |
| Story 8: 加载失败错误状态 | MilestonesPage error state component (frontend) | 显示 "加载失败，请重试" + 重试按钮；API 5xx 时触发 |
| Story 9: 团队成员只读 | Permission checks (milestone:read only) | 所有 milestone API |
| Story 9: 空状态+无权限 (0 个里程碑图且无 create 权限) | MilestonesPage conditional rendering (frontend, UF-1) | GET /milestone-maps 返回空列表 + 无 milestone:create 权限 → 显示"暂无里程碑图"空状态提示，不显示创建按钮 |
| Story 9: 禁用操作按钮+tooltip | MilestonesPage 按钮禁用逻辑 (frontend) | 无 milestone:update 权限时禁用创建/编辑按钮并显示 tooltip |
| Story 6: 里程碑选择器空选项 (团队无里程碑) | MilestoneSelector component (frontend, UF-5) | GET /milestones 返回空列表 → 下拉框仅显示"未分配"选项 |
| Story 10: 表格视图里程碑列 | TableViewPage integration (UF-6) | MainItem.milestoneKey + milestoneName |
| Story 10: 软删除里程碑显示 "--" | MainItem VO 转换逻辑 | milestone_key 非空但对应里程碑已软删除时，milestoneName 显示 "--" |
| Story 10: 里程碑列默认升序 | TableViewPage 列排序 (frontend) | 点击里程碑列头默认按 milestoneName 升序排列 |
| Story 11: 管理层只读 | Permission checks | milestone:read, no create/update/delete |
| NFR: 完成度实时计算 | calcCompletion/calcOverallProgress (service layer) | GET 时计算 |
| NFR: 软删除+事务解绑 | MilestoneService.Delete (transaction) | MilestoneRepo.SoftDelete + MainItemRepo batch update |
| NFR: 权限独立RBAC | permissions/codes.go + migration/rbac.go | milestone:create/read/update/delete |
| NFR: 里程碑图列表<300ms | Index on (team_key, map_status) | schema.sql |
| NFR: 时间线<500ms | ListByMap (no pagination) + batch MI lookup | MilestoneRepo.ListByMap |

## Open Questions

- [ ] None

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| 完成度持久化到里程碑表 | 查询更快 | 需要在每次 MI 变更时更新所有关联里程碑和图，增加写路径复杂度 | PRD 明确要求实时计算，且数据量小（<200 MI），查询性能足够 |
| 里程碑嵌套在 MilestoneMap API 内（无独立 Milestone API） | 路由更简洁 | 违反现有扁平路由模式（main-items 独立于 teams），且需要 team-level milestone 查询 | 保持与现有模式一致，Milestone 有独立 CRUD + 状态切换 |
| MainItem.milestone_key 使用 DDL 外键 | 数据库级引用完整性 | 影响数据迁移灵活性，现有项目统一不使用 DDL FK | 遵循 DM-003 约定：No FOREIGN KEY constraints in DDL |
