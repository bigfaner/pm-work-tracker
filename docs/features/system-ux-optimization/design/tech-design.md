---
created: "2026-06-02"
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: System UX Optimization Batch

## Overview

16 项 UX 优化分两阶段实施，按架构层分类。无数据库表结构变更 — 所有变更通过新增 API 端点、现有 Service/View 逻辑增强、前端 UI 增强和 seed 数据更新（2 个新权限码）实现。

**按层分类**：

| Category | Items | Layer |
|----------|-------|-------|
| 后端 Bug 修复 | #8（member 权限）, #15（团队选择器） | Backend middleware + API |
| 后端 Service 变更 | #5（子事项排序）, #11（终态排序）, #16（每周过滤） | Backend service + repo |
| 新增 API 端点 | #3（删除）, #9（移动） | Backend full-stack |
| 后端 View 增强 | #10（过滤穿透）, #12（甘特图过滤） | Backend ViewService |
| 前端 UI 修复 | #1（错误提示）, #2（开始时间）, #4（描述置灰）, #6（表单清空）, #7（必填校验）, #13（甘特图时间范围）, #14（macOS 滚动条） | Frontend only |
| 全栈功能 | #3（删除）, #9（移动）, #10（过滤穿透） | Backend + Frontend |

**关键架构决策**：

1. **无 Schema 迁移** — 所有 16 项在现有 schema 上工作。`main_item:delete` 和 `sub_item:delete` 作为 seed 数据添加到 `pmw_role_permissions` 表。

2. **过滤穿透** — 在 ViewService 层内存过滤。响应 DTO 扩展 `matchType` 和 `matchedSubItemIds` 字段供前端区分。复用现有 List 端点，通过增强的 filter 参数传递。

3. **移动子事项** — 新增 `PUT /sub-items/:subId/move` 端点。事务内原子自增目标主事项 `sub_item_seq`，重新生成编号，保留状态和负责人。并发安全性由原子 UPDATE + 唯一索引兜底保证。

4. **删除** — 新增 `DELETE /main-items/:itemId` 和 `DELETE /sub-items/:subId` 端点。主事项删除在单个事务内级联软删除所有子事项。仅设置 `deleted_flag=1` 和 `deleted_time`，不修改 `item_status`。权限由新增的 `main_item:delete` / `sub_item:delete` 权限码控制。

5. **状态错误提示** — 纯前端变更。`StatusTransitionDropdown` 将 2 秒 tooltip 替换为持久 Alert，显示后端 422 响应体中的 `message` 字段。

6. **团队选择器过滤** — `GET /v1/teams` 端点当前返回系统中所有团队（无成员关系过滤）。修复方案：Handler 提取用户身份 → Service 接收 userBizKey → Repository join `pmw_team_members` 过滤。

## Architecture

### Layer Placement

所有 16 项变更位于现有四层架构（Router→Handler→Service→Repository→Model）内。无新层。

```
┌─────────────────────────────────────────────────────────┐
│                      Frontend                            │
│  React Pages / Components / Hooks / API modules          │
├─────────────────────────────────────────────────────────┤
│                   Backend Router                         │
│  Gin routes + permission middleware                      │
├─────────────────────────────────────────────────────────┤
│                  Backend Handlers                        │
│  MainItemHandler, SubItemHandler, ViewHandler            │
├─────────────────────────────────────────────────────────┤
│                  Backend Services                        │
│  MainItemService, SubItemService, ViewService            │
├─────────────────────────────────────────────────────────┤
│                Backend Repositories                      │
│  MainItemRepo, SubItemRepo, StatusHistoryRepo            │
├─────────────────────────────────────────────────────────┤
│                    Database (SQLite/MySQL)                │
│  Existing tables — no schema changes                     │
└─────────────────────────────────────────────────────────┘
```

### Component Diagram

```
┌─ Frontend ────────────────────────────────────────────────┐
│                                                            │
│  ItemViewPage ─────────┬─ StatusTransitionDropdown (#1)    │
│  (卡片/表格视图)         ├─ FilterBar (#7,#10,#11)          │
│                        ├─ CreateMainItemDialog (#6,#7)     │
│                        ├─ EditSubItemDialog (#2,#6)        │
│                        └─ CreateSubItemDialog (#4,#6,#7)   │
│                                                            │
│  MainItemDetailPage ───┬─ SubItemList (#5)                 │
│  (主事项详情)           ├─ DeleteButton (#3)               │
│                        └─ MoveSubItemButton (#9)           │
│                                                            │
│  SubItemDetailPage ────┬─ DeleteButton (#3)                │
│  (子事项详情)           └─ MoveSubItemButton (#9)          │
│                                                            │
│  GanttViewPage (#12,#13,#14)                               │
│  WeeklyProgressPage (#16)                                  │
│  Sidebar TeamSelector (#15)                                │
│  ItemPoolConvertForm (#4,#6,#7)                            │
│                                                            │
└────────────────────────────────────────────────────────────┘

┌─ Backend ─────────────────────────────────────────────────┐
│                                                            │
│  Router                                                    │
│  ├─ DELETE /main-items/:itemId (#3) ← main_item:delete    │
│  ├─ DELETE /sub-items/:subId (#3) ← sub_item:delete       │
│  ├─ PUT /sub-items/:subId/move (#9)                       │
│  └─ Enhanced query params on existing endpoints (#10,#11) │
│                                                            │
│  Handlers                                                  │
│  ├─ MainItemHandler.Delete (#3)                            │
│  ├─ SubItemHandler.Delete (#3)                             │
│  ├─ SubItemHandler.Move (#9)                               │
│  └─ ViewHandler (enhanced filters #10,#12,#16)            │
│                                                            │
│  Services                                                  │
│  ├─ MainItemService.Delete (#3) — tx: cascade soft-delete │
│  ├─ SubItemService.Delete (#3)                             │
│  ├─ SubItemService.Move (#9) — tx: reassign + renumber    │
│  ├─ ViewService (enhanced #10,#11,#12,#16)                │
│  └─ MainItemService.List (sort #11)                       │
│                                                            │
│  Middleware                                                │
│  └─ TeamScopeMiddleware (#8 fix: nil RoleKey)             │
│                                                            │
│  Repositories (no new repos, minor query changes)         │
│  ├─ MainItemRepo.ListNonArchivedByTeam (#11 ORDER BY)     │
│  ├─ SubItemRepo.ListByMainItem (#5 ORDER BY id DESC)      │
│  └─ TeamRepo.ListByUserMembership (#15)                   │
│                                                            │
└────────────────────────────────────────────────────────────┘

┌─ Database ────────────────────────────────────────────────┐
│  Seed data only: main_item:delete, sub_item:delete         │
│  No DDL changes                                            │
└────────────────────────────────────────────────────────────┘
```

### Dependencies

| Dependency | Type | Purpose | Version |
|------------|------|---------|---------|
| `@radix-ui/react-alert-dialog` | existing | Delete/Move confirmation dialog | current |
| `@radix-ui/react-checkbox` | **new** | Multi-select status filter (#10,#12) | latest stable |
| GORM `Expr("sub_item_seq + 1")` | existing | Atomic counter increment in sub-item move | current |

仅新增 1 个前端依赖（checkbox primitive）。Radix 已在项目中使用。

## Interfaces

### Interface 1: 删除主事项

**API**:
```
DELETE /teams/:teamId/main-items/:itemId
Auth: main_item:delete
```

**Request**: 无 body

**Response 200**:
```json
{ "message": "ok" }
```

**Error Responses**:

| Status | Body | Trigger |
|--------|------|---------|
| 404 | `{ "code": "NOT_FOUND", "message": "主事项不存在" }` | 主事项已删除或不存在 |
| 403 | `{ "code": "FORBIDDEN", "message": "无权限" }` | 无 delete 权限 |

**Service 签名**:
```go
// MainItemService
Delete(ctx context.Context, teamBizKey, itemBizKey int64, operatorBizKey int64) error
```

**事务内操作序列**:
1. `FindByBizKey` 获取主事项
2. `subItemRepo.ListByMainItem` 获取子事项列表
3. 批量 `subItemRepo.SoftDelete` 所有子事项
4. `mainItemRepo.SoftDelete` 主事项
5. 为每个已删除事项插入 `status_histories`（`from_status=当前状态, to_status="deleted"`），仅作为审计记录，不修改 `item_status`
6. 重新计算相关主事项 `completion_pct`（如需）

注：仅设置 `deleted_flag=1` 和 `deleted_time`，不修改 `item_status`。`status_histories` 插入用于审计追踪。

---

### Interface 2: 删除子事项

**API**:
```
DELETE /teams/:teamId/sub-items/:subId
Auth: sub_item:delete
```

**Response**: 同 Interface 1 格式

**Service 签名**:
```go
// SubItemService
Delete(ctx context.Context, teamBizKey, subItemBizKey int64, operatorBizKey int64) error
```

**事务内操作序列**:
1. `FindByBizKey` 获取子事项
2. `subItemRepo.SoftDelete`
3. 插入 `status_histories`（`from_status=当前状态, to_status="deleted"`），仅审计记录
4. 重新计算父主事项 `completion_pct`

注：仅设置 `deleted_flag=1` 和 `deleted_time`，不修改 `item_status`。

---

### Interface 3: 移动子事项

**API**:
```
PUT /teams/:teamId/sub-items/:subId/move
Auth: sub_item:update
Body: { "targetMainItemBizKey": "729384756" }
```

**Response 200**:
```json
{
  "newSubCode": "T001-007",
  "mainItemBizKey": "729384756"
}
```

**Error Responses**:

| Status | Code | Message | Trigger |
|--------|------|---------|---------|
| 400 | BAD_REQUEST | 目标主事项已关闭 | 目标为终态 |
| 400 | BAD_REQUEST | 不能移动到同一主事项 | 源 = 目标 |
| 404 | NOT_FOUND | 子事项不存在 | — |
| 404 | NOT_FOUND | 目标主事项不存在 | — |

**Service 签名**:
```go
// SubItemService
Move(ctx context.Context, teamBizKey, subItemBizKey int64, targetMainItemBizKey int64, operatorBizKey int64) (*MoveResult, error)

type MoveResult struct {
    NewSubCode     string
    MainItemBizKey int64
}
```

**事务内操作序列**:
1. `FindByBizKey` 获取子事项（校验未删除）
2. `FindByBizKey` 获取目标主事项（校验未删除、`team_key == teamBizKey`、非终态、非同一主事项）
3. 原子自增目标主事项 `sub_item_seq`：`UPDATE ... SET sub_item_seq = sub_item_seq + 1`
4. 读回新的 `sub_item_seq` 值
5. 更新子事项：`main_item_key`、`item_code`（新编号）
6. 重算源主事项和目标主事项的 `completion_pct`

---

### Interface 4: 过滤穿透查询（增强现有 List 接口）

**API**:
```
GET /teams/:teamId/main-items?status=progressing,blocking&assigneeKey=12345
```

**变更点**: `MainItemFilter` 扩展

```go
type MainItemFilter struct {
    Statuses    []string `form:"status" json:"status"`       // 改为多选
    Priority    string   `form:"priority" json:"priority"`
    AssigneeKey *string  `form:"assigneeKey" json:"assigneeKey"`
    IsKeyItem   *bool    `form:"isKeyItem"`
    Archived    bool     `form:"archived"`
}
```

**Response 扩展** — 每个 `MainItem` 增加：

```typescript
// Frontend type
interface MainItem {
  // ...existing fields
  matchType?: "direct" | "indirect"
  matchedSubItemIds?: string[]
}
```

```go
// Backend DTO
type MainItemMatchInfo struct {
    MatchType         string   `json:"matchType,omitempty"`
    MatchedSubItemIds []string `json:"matchedSubItemIds,omitempty"`
}
```

**穿透逻辑**（ViewService 层内存过滤）:
1. 仅选状态 → 返回状态匹配的主事项，`matchType` 全部为 `direct`
2. 仅选负责人 → 返回直接匹配的主事项 + 含该负责人子事项的主事项（`matchType=indirect`）
3. 同时选 → AND 逻辑：主事项状态须匹配 AND（负责人直接匹配 OR 子事项负责人匹配）
4. 未选任何 → 全量返回，无 `matchType` 字段

---

### Interface 5: 子事项列表倒序（#5）

**变更点**: `SubItemRepo.ListByMainItem` 查询排序

```go
// Before: ORDER BY id ASC
// After:  ORDER BY id DESC
```

纯 Repository 层变更，Service 和 Handler 接口不变。

---

### Interface 6: 终态排序（#11）

**变更点**: `MainItemService.List` 和 `ViewService.GanttView` 排序逻辑

在现有排序后追加终态下沉：

```go
sort.SliceStable(items, func(i, j int) bool {
    iTerminal := status.IsMainTerminal(items[i].Status)
    jTerminal := status.IsMainTerminal(items[j].Status)
    if iTerminal != jTerminal {
        return !iTerminal // 终态排后面
    }
    // 保持原有排序
    return /* existing comparison */
})
```

---

### Interface 7: 甘特图多状态过滤（#12）

**变更点**: `GanttFilter` 从单选改为多选

```go
// Before
type GanttFilter struct {
    Status string `form:"status"`
}

// After
type GanttFilter struct {
    Statuses []string `form:"status" json:"status"`
}
```

`ViewService.GanttView` 改为多状态过滤（复用现有 `ListNonArchivedByTeam` + 内存过滤）。前端首次加载默认传 `["progressing"]`。

---

### Interface 8: 每周进展过滤（#16）

**变更点**: `ViewService.WeeklyComparison` — 在 `buildWeeklyGroups` 后追加过滤

```go
// StatusHistoryRepo 新增方法
ListByItemKeysInRange(ctx context.Context, itemType string, itemKeys []int64, start, end time.Time) ([]model.StatusHistory, error)
```

```go
func filterInactiveTerminal(groups []WeeklyComparisonGroup, weekStart, weekEnd time.Time, statusHistories []model.StatusHistory) []WeeklyComparisonGroup {
    // 隐藏条件: 主事项终态 AND 本周/上周都无活跃子事项
    // 活跃定义:
    //   (a) status_history 中存在该主事项的状态变更记录（create_time 在时间范围内）
    //   (b) 该主事项下存在 create_time 或 db_update_time 在时间范围内的子事项
    //   (c) 该主事项的 db_update_time 在时间范围内（进度更新）
}
```

需要新增 `StatusHistoryRepo.ListByTeamInRange` 方法获取时间范围内的状态变更记录。

---

### Interface 9: 团队选择器权限过滤（#15）

**根因**: `GET /v1/teams` 当前返回系统中所有团队（无成员关系过滤）。Handler 不传用户身份，Service 不接收 userBizKey，Repository 直接查 `pmw_teams WHERE deleted_flag=0`。

**变更点**:
- **Handler**: `TeamHandler.List` 提取 `GetUserBizKey(c)` 传入 Service
- **Service**: `ListTeams(ctx, userBizKey, search, page, pageSize)` 新增 `userBizKey` 参数
- **Repository**: 新增 `ListByUserMembership` 方法，join `pmw_team_members` 过滤（复用现有 `FindTeamsByUserBizKeys` 模式）

```go
// Handler
func (h *TeamHandler) List(c *gin.Context) {
    userBizKey := middleware.GetUserBizKey(c)
    // ...
    teams, total, err := h.teamSvc.ListTeams(ctx, userBizKey, search, page, pageSize)
}

// Service
ListTeams(ctx context.Context, userBizKey int64, search string, page, pageSize int) (*dto.PageResult[dto.TeamDTO], error)

// Repository
ListByUserMembership(ctx context.Context, userBizKey int64, search string, offset, limit int) ([]*model.Team, int64, error)
```

前端无需改动。

---

### Interface 10: Member 权限修复（#8）

**根因**: `TeamScopeMiddleware` 中当 `pmw_team_members.role_key` 为 NULL 时，权限查询返回空集，导致 member 角色用户获取不到任何权限。

**变更点**: 中间件中处理 nil RoleKey — 当 role_key 为 NULL 时，查询 member 预设角色的默认权限集。

```go
// middleware/team_scope.go — 权限查询逻辑修正
// Before: 直接用 member.RoleKey 查询权限
// After: if member.RoleKey == nil → 查询 "member" preset role 的权限码
```

纯中间件层变更，Service 和 Handler 接口不变。

---

### Interface 11: 前端组件接口变更（#1,#2,#4,#6,#7,#13,#14）

**StatusTransitionDropdown (#1)**:

```typescript
const [errorMessage, setErrorMessage] = useState<string | null>(null)
// mutation onError: setErrorMessage(err.response.data.message)
// mutation onSuccess: setErrorMessage(null)
// 渲染: errorMessage && <Alert variant="destructive" onClose={() => setErrorMessage(null)}>{errorMessage}</Alert>
```

**EditSubItemDialog (#2)**:

```typescript
interface EditSubItemFormState {
  title: string
  priority: string
  assigneeKey: string
  startDate: string        // 新增
  expectedEndDate: string
  description: string
}
```

**转换表单 (#4,#6,#7)**:

```typescript
// AssignItemPoolReq: description 置灰禁用
// 提交按钮 disabled 条件: !assigneeKey || !priority
// 表单关闭/提交成功时: resetForm() 清空所有字段
```

## Data Models

数据库无结构变更。仅 seed 数据新增 2 个权限码。

> **ER Diagram**: design/er-diagram.md
> **SQL Schema**: design/schema.sql

### Field Quick Reference

| Model | Key Fields | Notes |
|-------|------------|-------|
| `MainItem` | `biz_key BIGINT, team_key BIGINT, item_code VARCHAR(12), sub_item_seq INT, item_status VARCHAR(20), assignee_key BIGINT, completion_pct REAL` | 移动/删除/排序涉及字段 |
| `SubItem` | `biz_key BIGINT, main_item_key BIGINT, item_code VARCHAR(15), plan_start_date DATETIME, item_status VARCHAR(20), assignee_key BIGINT` | 移动/删除/排序/开始时间涉及字段 |
| `StatusHistory` | `item_type VARCHAR(20), item_key BIGINT, from_status VARCHAR(20), to_status VARCHAR(20), changed_by BIGINT` | 每周活跃判断 |
| `pmw_role_permissions` | `role_key BIGINT, permission_code VARCHAR(50)` | 新增 `main_item:delete`, `sub_item:delete` |

## Error Handling

### Error Types & Codes

| Error Code | Name | Description | HTTP Status |
|------------|------|-------------|-------------|
| `NOT_FOUND` | ErrNotFound | 事项/团队不存在（含已软删除） | 404 |
| `FORBIDDEN` | ErrForbidden | 无权限执行操作 | 403 |
| `BAD_REQUEST` | ErrBadRequest | 业务校验失败（目标已关闭/同一主事项） | 400 |
| `INVALID_STATUS` | ErrInvalidStatus | 状态流转不合法 | 422 |
| `CONFLICT` | ErrConflict | 并发冲突（乐观锁） | 409 |
| `INTERNAL` | ErrInternal | 事务执行失败 | 500 |

### Propagation Strategy

**后端**：沿用现有 `apperrors` 包 + `pkgerrors.MapNotFound` 映射。新增两个业务错误：

```go
var (
    ErrTargetClosed = &apperrors.AppError{Code: "BAD_REQUEST", Message: "目标主事项已关闭", Status: 400}
    ErrSameMainItem = &apperrors.AppError{Code: "BAD_REQUEST", Message: "不能移动到同一主事项", Status: 400}
)
```

Handler 层统一通过 `apperrors.RespondError(c, err)` 映射到 HTTP 响应（现有模式）。

**前端**：`StatusTransitionDropdown` 的 mutation `onError` 从 `err.response.data.message` 提取错误消息显示为 Alert。删除和移动操作同理。

**事务失败**：删除和移动均在事务内执行。GORM 回滚后返回 `ErrInternal`，`RespondError` 自动映射为 500，前端展示通用错误提示，数据不变。

## Cross-Layer Data Map

| Field Name | Storage Layer | Backend Model | API/DTO | Frontend Type | Validation Rule |
|------------|---------------|---------------|---------|---------------|-----------------|
| `status` (多选) | — | `MainItemFilter.Statuses []string` | `form:"status"` | `string[]` | 每个值须为合法状态码 |
| `assigneeKey` | `pmw_main_items.assignee_key BIGINT` | `*int64` | `json:"assigneeKey"` (query: `form:"assigneeKey"` as `*string`) | `string \| null` | 穿透到子事项匹配；query param 为 string，Service 层用 `pkg.ParseID` 转为 int64 |
| `matchType` | — (计算值) | — | `json:"matchType,omitempty"` | `"direct" \| "indirect" \| undefined` | 仅过滤穿透时返回 |
| `matchedSubItemIds` | — (计算值) | — | `json:"matchedSubItemIds,omitempty"` | `string[] \| undefined` | 仅 matchType=indirect 时返回 |
| `startDate` | `pmw_sub_items.plan_start_date DATETIME` | `*time.Time` | `json:"startDate"` | `string` | 日期格式，不晚于 endDate |
| `targetMainItemBizKey` | `pmw_main_items.biz_key BIGINT` | `int64` | `json:"targetMainItemBizKey"` | `string` | 必填，正整数，非终态 |
| `newSubCode` | — (计算值) | — | `json:"newSubCode"` | `string` | 由后端 sub_item_seq 生成 |
| `permission_code` | `pmw_role_permissions.permission_code VARCHAR(50)` | `string` | — | `"main_item:delete" \| "sub_item:delete"` | seed 数据 |

## Integration Specs

### Integration: StatusTransitionDropdown → ItemViewPage / MainItemDetailPage

- **Target File**: `frontend/src/components/shared/StatusTransitionDropdown.tsx`
- **Insertion Point**: DropdownMenu 下方，组件 return JSX 中追加 Alert 组件
- **Data Source**: mutation `onError` 回调中的 `err.response.data.message`

### Integration: EditSubItemDialog → ItemViewPage / MainItemDetailPage

- **Target File**: `frontend/src/pages/item-view/EditSubItemDialog.tsx`, `frontend/src/pages/main-item-detail/EditSubItemDialog.tsx`
- **Insertion Point**: `expectedEndDate` 字段之前，插入 `startDate` 日期选择器
- **Data Source**: `SubItemUpdateReq.startDate`，form state 扩展 `startDate` 字段

### Integration: DeleteButton → MainItemDetailPage

- **Target File**: `frontend/src/pages/main-item-detail/MainItemDetailPage.tsx`
- **Insertion Point**: 操作按钮区域（状态流转按钮旁），用 `PermissionGuard code="main_item:delete"` 包裹
- **Data Source**: `DELETE /teams/:teamId/main-items/:itemId`，确认对话框需 `subItemCount`

### Integration: DeleteButton → SubItemDetailPage

- **Target File**: 子事项详情页组件
- **Insertion Point**: 操作按钮区域
- **Data Source**: `DELETE /teams/:teamId/sub-items/:subId`

### Integration: MoveSubItemButton → SubItemDetailPage

- **Target File**: 子事项详情页组件
- **Insertion Point**: 操作按钮区域（删除按钮旁）
- **Data Source**: `PUT /teams/:teamId/sub-items/:subId/move`，目标选择器数据来自 `listMainItemsApi`

### Integration: FilterBar → ItemViewPage

- **Target File**: `frontend/src/pages/item-view/useItemViewPage.ts` + `ItemViewPage.tsx`
- **Insertion Point**: 状态过滤器改为 Checkbox Group（多选），负责人过滤器增加穿透逻辑
- **Data Source**: `listMainItemsApi` 返回值中的 `matchType` 和 `matchedSubItemIds`

### Integration: Member Permission Fix → TeamScopeMiddleware

- **Target File**: `backend/internal/middleware/team_scope.go`
- **Insertion Point**: 权限查询逻辑中处理 `role_key IS NULL` 分支
- **Data Source**: 查询 `pmw_roles` 中 `role_name='member'` 的预设权限集

### Integration: StatusCheckboxGroup → GanttViewPage

- **Target File**: `frontend/src/pages/GanttViewPage.tsx`
- **Insertion Point**: 页面顶部过滤器区域，新增状态 Checkbox Group
- **Data Source**: `getGanttViewApi(teamId, { statuses })` — 默认选中 `"progressing"`

### Integration: GanttTimeRange → GanttViewPage

- **Target File**: `frontend/src/pages/GanttViewPage.tsx` 中日期范围计算逻辑
- **Insertion Point**: 替换现有的 14 天 padding，改为 `minDate = earliestStartDate - 1 day`，`maxDate = latestEndDate + 1 day`
- **Data Source**: 可见主事项的 `startDate` / `expectedEndDate`

### Integration: MacOSScrollbar → gantt-overrides.css

- **Target File**: `frontend/src/pages/gantt-overrides.css`
- **Insertion Point**: 滚动容器样式，追加 `overflow-x: auto` + `-webkit-scrollbar` hover 显示规则
- **Data Source**: 无

### Integration: TeamSelector → Sidebar

- **Target File**: `frontend/src/api/teams.ts`（无需改动）, `frontend/src/components/layout/AppLayout.tsx`（无需改动）
- **Insertion Point**: 后端 `GET /v1/teams` 已按成员关系过滤后，前端自动仅展示用户所属团队
- **Data Source**: `GET /v1/teams` 响应

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| Backend Service | Unit | Go testing + testify | Delete 事务、Move 编号生成、过滤穿透逻辑、终态排序、weekly 过滤 | 85% |
| Backend Handler | Integration | Go testing + httptest | Delete/Move API 端点契约、权限码拦截、参数校验 | 80% |
| Backend Middleware | Unit | Go testing | nil RoleKey 权限查询修复 | 100%（该函数） |
| Backend Repository | Unit | Go testing + testify | SubItemRepo ORDER BY, TeamRepo membership filter, StatusHistoryRepo range query | 80% |
| Frontend Component | Unit | Vitest + @testing-library/react | Alert 生命周期、表单必填校验、表单清空、多选过滤器 | 80% |
| Frontend Hook | Unit | Vitest + @testing-library/react | useItemViewPage 过滤穿透逻辑、matchType 展示 | 80% |

### Key Test Scenarios

**后端 — 删除（#3）**:
- 主事项删除 → 所有子事项 `deleted_flag=1`，主事项 `deleted_flag=1`，`item_status` 不变
- 子事项删除 → 父主事项 `completion_pct` 重算
- 无权限用户 → 403
- 不存在事项 → 404
- 空子事项主事项删除 → 正常，无子事项报错

**后端 — 移动（#9）**:
- 正常移动 → `item_code` 使用目标主事项 `sub_item_seq` 生成，`item_status` 和 `assignee_key` 不变
- 移到已关闭主事项 → 400 `ErrTargetClosed`
- 移到同一主事项 → 400 `ErrSameMainItem`
- 两个请求同时移到同一目标 → 第二个得到不同编号（原子自增保证）
- 移动后 → 源和目标主事项 `completion_pct` 均重算

**后端 — 过滤穿透（#10）**:
- 仅选状态 → 只返回状态匹配的主事项，无 `matchType`
- 仅选负责人 A → 返回 A 负责的主事项 + 含 A 子事项的主事项（`matchType=indirect`）
- 同时选状态和负责人 → AND 逻辑：状态不匹配的主事项即使子事项匹配负责人也不展示
- `matchedSubItemIds` 仅包含匹配的子事项

**后端 — 终态排序（#11）**:
- 混合状态列表 → 终态（completed/closed）排在最后
- 全部终态 → 保持原有相对顺序
- 无终态 → 结果不变

**后端 — Weekly 过滤（#16）**:
- 终态主事项 + 本周有活跃子事项 → 展示
- 终态主事项 + 本周/上周均无活跃 → 隐藏
- 非终态主事项 → 始终展示

**后端 — Member 权限修复（#8）**:
- `role_key=NULL` 的 member → 权限查询返回 member 角色默认权限集

**后端 — Repository 层变更**:
- `SubItemRepo.ListByMainItem` → 验证返回结果按 `id DESC` 排序
- `MainItemRepo.ListNonArchivedByTeam` → 验证终态排序生效
- `TeamRepo.ListByUserMembership` → 验证仅返回用户所属团队
- `StatusHistoryRepo.ListByItemKeysInRange` → 验证时间范围过滤和 item_type 过滤

**后端 — 团队选择器过滤（#15）**:
- 用户属于 2 个团队 → 仅返回这 2 个团队
- 用户不属于任何团队 → 返回空列表
- SuperAdmin → 返回所有团队（bypass）

**前端 — 状态流转错误提示（#1）**:
- 流转失败 → Alert 展示后端消息
- 再次点击流转 → Alert 更新或清除
- 手动关闭 Alert → Alert 隐藏
- 流转成功 → Alert 消失

**前端 — 表单校验和清空（#4,#6,#7）**:
- 待办→子事项：描述字段 disabled
- 未选负责人或优先级 → 提交按钮禁用
- 提交成功后再次打开 → 字段为空
- 手动关闭后再次打开 → 字段为空
- 提交失败 → 字段保留

**前端 — 甘特图时间范围（#13）**:
- 最早开始时间 4/12，最晚结束 6/30 → 时间轴从 4/11 到 7/1，无空白

### Overall Coverage Target

80%

## Security Considerations

### Threat Model

| Threat | Risk Level | Description |
|--------|-----------|-------------|
| 未授权删除 | 中 | 非 PM 用户尝试删除事项 |
| 越权移动 | 中 | 用户将子事项移到无权限团队的主事项 |
| 权限码绕过 | 低 | 新增权限码被跳过 |

### Mitigations

- **删除**：`main_item:delete` / `sub_item:delete` 权限码通过 `RequirePermission` 中间件在 Router 层拦截。前端用 `PermissionGuard` 隐藏按钮（防御性 UI，非安全边界）
- **移动**：复用现有 `sub_item:update` 权限码。Service 层显式校验目标主事项 `team_key == 当前 teamBizKey`，防止跨团队移动。`FindByBizKey` 仅验证记录存在性，不校验团队归属
- **团队隔离**：所有操作在 `TeamScopeMiddleware` 上下文内执行，`teamBizKey` 从 JWT + URL param 绑定

## PRD Coverage Map

| PRD Item | Design Component | Interface / Model |
|----------|------------------|-------------------|
| #1 状态流转错误提示 | StatusTransitionDropdown Alert | Interface 11 (前端) |
| #2 子事项开始时间 | EditSubItemDialog + SubItemUpdateReq | Interface 11 (前端) |
| #3 删除事项 | MainItemService.Delete, SubItemService.Delete | Interface 1, 2 |
| #4 描述置灰 | ItemPool assign form disabled | Interface 11 (前端) |
| #5 子事项倒序 | SubItemRepo.ListByMainItem ORDER BY | Interface 5 |
| #6 表单清空 | Form reset on close/success | Interface 11 (前端) |
| #7 必填校验 | Assign form submit disabled | Interface 11 (前端) |
| #8 Member 权限修复 | TeamScopeMiddleware nil RoleKey fix | Interface 10 (后端中间件) |
| #9 移动子事项 | SubItemService.Move | Interface 3 |
| #10 过滤穿透 | ViewService + MainItemFilter 增强 | Interface 4 |
| #11 终态排序 | Service 层排序逻辑 | Interface 6 |
| #12 甘特图状态过滤 | GanttFilter 多选 + 前端 Checkbox Group | Interface 7 |
| #13 甘特图时间范围 | GanttViewPage 日期计算 | Interface 11 (前端) |
| #14 macOS 滚动条 | gantt-overrides.css | Interface 11 (前端) |
| #15 团队选择器过滤 | TeamHandler.List + TeamRepo.ListByUserMembership | Interface 9 |
| #16 每周进展过滤 | ViewService.WeeklyComparison 过滤 | Interface 8 |

## Open Questions

(All resolved)

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| 前端过滤穿透 | 无后端改动 | 需全量数据到前端，1000+5K 规模下性能差 | PRD 要求 ≤500ms，前端内存过滤不可控 |
| 独立穿透 API 端点 | 职责清晰 | 新增端点 + 前端双数据源管理 | 复用现有 List 端点更简洁，通过 DTO 扩展即可 |
| FOR UPDATE 行锁 | 强一致性 | SQLite 不支持，MySQL 下锁争用 | 原子自增 `sub_item_seq` 足够，唯一索引兜底 |
| 删除时修改 item_status | 状态一致 | 软删除已通过 deleted_flag 标记，改状态冗余 | 仅设 deleted_flag，查询统一用 NotDeleted scope |

### References

- `backend/internal/pkg/status/status.go` — 终态定义
- `backend/internal/middleware/team_scope.go` — 权限中间件
- `backend/internal/service/view_service.go` — ViewService 内存过滤
- `docs/conventions/soft-delete.md` — SD-003 SoftDelete 方法模式
- `docs/conventions/api-boundary.md` — AB-002 Handler 两步删除模式
- `docs/conventions/permission-codes.md` — 权限码注册和迁移模式
