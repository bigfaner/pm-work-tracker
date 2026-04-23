---
feature: "weekly-stats-optimization"
status: draft
---

# Tech Design: 每周进展统计优化

## Overview

在现有 `WeeklyComparison` 接口内扩展统计字段，不新增接口调用。后端 `WeeklyStats` DTO 新增 3 个字段，`buildWeeklyGroups` 补充计数逻辑；前端 `StatsBar` 从 4 列扩展为 7 列并添加 Tooltip；`isOverdue` 函数签名改为必传 `referenceDate`。

## Architecture Fit

变更完全在现有层次内。数据流：

```
buildWeeklyGroups (view_service.go)
  → WeeklyStats DTO (item_dto.go)
  → view_handler.go  →  HTTP 200 JSON
                              ↓
                    useQuery (WeeklyViewPage.tsx)
                              ↓
                    WeeklyViewResponse.stats
                              ↓
                    StatsBar (WeeklyViewPage.tsx)
                              ↓
                    StatCard × 7
```

**外部依赖**

| 包 | 用途 |
|----|------|
| `@radix-ui/react-tooltip` | StatCard Tooltip（已安装） |
| `@tanstack/react-query` v5 | `useQuery` 数据获取（已安装） |
| `github.com/stretchr/testify` | Go 断言库（已安装） |

## Interface 1: Backend DTO — WeeklyStats

**文件**: `backend/internal/dto/item_dto.go`

```go
// 现有
type WeeklyStats struct {
    ActiveSubItems int `json:"activeSubItems"`
    NewlyCompleted int `json:"newlyCompleted"`
    InProgress     int `json:"inProgress"`
    Blocked        int `json:"blocked"`
}

// 变更后（新增 3 字段）
type WeeklyStats struct {
    ActiveSubItems int `json:"activeSubItems"`
    NewlyCompleted int `json:"newlyCompleted"`
    InProgress     int `json:"inProgress"`
    Blocked        int `json:"blocked"`
    Pending        int `json:"pending"`
    Pausing        int `json:"pausing"`
    Overdue        int `json:"overdue"`
}
```

Go struct 新增字段默认零值，JSON 序列化向后兼容（旧客户端忽略新字段）。

## Interface 2: Backend Service — buildWeeklyGroups

**文件**: `backend/internal/service/view_service.go`

在现有 `inThisWeek || justCompleted` 分支内，新增 3 个计数分支：

```go
if inThisWeek || justCompleted {
    group.ThisWeek = append(group.ThisWeek, snapshot)
    stats.ActiveSubItems++
    if si.Status == "progressing" {
        stats.InProgress++
    }
    if si.Status == "blocking" {
        stats.Blocked++
    }
    // 新增
    if si.Status == "pending" {
        stats.Pending++
    }
    if si.Status == "pausing" {
        stats.Pausing++
    }
    // 逾期：expectedEndDate < weekEnd AND status ∉ {completed, closed}
    if si.ExpectedEndDate != nil && si.ExpectedEndDate.Before(weekEnd) &&
        si.Status != "completed" && si.Status != "closed" {
        stats.Overdue++
    }
}
```

**逾期判定说明**：
- `si.ExpectedEndDate` 是 `*time.Time`，nil 时不计入逾期
- 比较使用 `weekEnd`（已由调用方传入），不使用 `time.Now()`
- 逾期与状态卡片（3-6）可重叠（PRD 数学不变式允许）

## Interface 3: Backend Unit Tests

**文件**: `backend/internal/service/view_service_test.go`（新增测试函数）

测试函数：`TestBuildWeeklyGroups_Stats`，使用 table-driven 结构，覆盖：

| Fixture | 验证点 |
|---------|--------|
| 7 种状态各 1 个活跃子事项 | 7 个卡片计数均为 1 |
| pending 子事项不活跃 | pending 计数为 0 |
| overdue + progressing 同一事项 | overdue=1, inProgress=1（重叠） |
| justCompleted 事项 | newlyCompleted=1, activeSubItems=1 |
| expectedEndDate=nil | overdue=0 |
| status=completed 且 expectedEndDate < weekEnd | overdue=0（terminal 状态不计入） |
| status=closed 且 expectedEndDate < weekEnd | overdue=0 |

## Interface 4: Frontend Type — WeeklyStats

**文件**: `frontend/src/types/index.ts`

```typescript
// 现有
export interface WeeklyViewResponse {
  weekStart: string
  weekEnd: string
  stats: {
    activeSubItems: number
    newlyCompleted: number
    inProgress: number
    blocked: number
  }
  groups: WeeklyComparisonGroup[]
}

// 变更后（stats 内联类型新增 3 字段）
export interface WeeklyViewResponse {
  weekStart: string
  weekEnd: string
  stats: {
    activeSubItems: number
    newlyCompleted: number
    inProgress: number
    blocked: number
    pending: number
    pausing: number
    overdue: number
  }
  groups: WeeklyComparisonGroup[]
}
```

## Interface 5: Frontend — isOverdue 签名更新

**文件**: `frontend/src/lib/status.ts`

```typescript
// 现有（referenceDate 隐式使用 new Date()）
export function isOverdue(expectedEndDate?: string, status?: string): boolean {
  if (!expectedEndDate || !status) return false
  const isTerminal = ...
  if (isTerminal) return false
  return new Date(expectedEndDate) < new Date()
}

// 变更后（referenceDate 必传）
export function isOverdue(
  expectedEndDate: string | undefined,
  status: string | undefined,
  referenceDate: Date
): boolean {
  if (!expectedEndDate || !status) return false
  const isTerminal = (MAIN_ITEM_STATUSES as Record<string, { terminal: boolean }>)[status]?.terminal
    || (SUB_ITEM_STATUSES as Record<string, { terminal: boolean }>)[status]?.terminal
  if (isTerminal) return false
  return new Date(expectedEndDate) < referenceDate
}
```

**调用方更新**（共 7 处，均传入 `new Date()`，WeeklyViewPage 传入 `weekEnd`）：

| 文件 | 行 | 变更 |
|------|----|------|
| `pages/WeeklyViewPage.tsx:151` | mainItem 逾期标记 | `isOverdue(..., weekEnd ? new Date(weekEnd) : new Date())` |
| `pages/WeeklyViewPage.tsx:282` | subItem 逾期标记 | 同上 |
| `pages/SubItemDetailPage.tsx:268` | 子事项详情页 | `isOverdue(..., new Date())` |
| `pages/TableViewPage.tsx` | 表格视图 | `isOverdue(..., new Date())` |
| `pages/main-item-detail/ItemInfoCard.tsx:39` | 主事项信息卡 | `isOverdue(..., new Date())` |
| `pages/item-view/ItemDetailView.tsx:104` | 事项详情视图 | `isOverdue(..., new Date())` |
| `pages/item-view/ItemSummaryView.tsx:91,152` | 事项摘要视图 | `isOverdue(..., new Date())` |

`weekEnd` 来自 `WeeklyViewPage` 的 `data.weekEnd`（string），转换为 `new Date(data.weekEnd)`。

## Interface 6: Frontend — StatsBar 组件

**文件**: `frontend/src/pages/WeeklyViewPage.tsx`

### 响应式布局

```tsx
// 7 列响应式 grid
<div className="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-7 gap-3 mb-6">
```

| 断点 | Tailwind 类 | 效果 |
|------|-------------|------|
| <768px | `grid-cols-2` | 2 列多行 |
| 768–1279px | `md:grid-cols-4` | 4+3 两行 |
| ≥1280px | `xl:grid-cols-7` | 7 列单行 |

### Tooltip 实现

使用现有 Radix UI `Tooltip` 组件（`frontend/src/components/ui/tooltip.tsx`）。

**桌面端**：Radix `Tooltip` 默认 hover 触发，`delayDuration={300}`。

**移动端 fallback**：Radix `Tooltip` 在触摸设备上不触发 hover。通过 `open` + `onOpenChange` 受控模式，在 `onClick` 时切换 open 状态，实现点击展开/收起。

**键盘可访问性**：`TooltipTrigger` 渲染为 `<button>`（Radix 默认），Tab 可聚焦，聚焦时 tooltip 可见。`TooltipContent` 自动生成 `id`，通过 `aria-describedby` 关联到 trigger。

### StatCard 子组件

```tsx
interface StatCardProps {
  value: number | string
  label: string
  tooltip: string
  valueClassName?: string
  testId?: string
}

function StatCard({ value, label, tooltip, valueClassName, testId }: StatCardProps) {
  const [open, setOpen] = useState(false)
  return (
    <Tooltip open={open} onOpenChange={setOpen} delayDuration={300}>
      <TooltipTrigger asChild>
        <button
          className="rounded-xl border border-border bg-white p-4 text-center w-full"
          onClick={() => setOpen(v => !v)}
        >
          <div className={cn('text-2xl font-semibold', valueClassName)} data-testid={testId}>
            {value}
          </div>
          <div className="text-[13px] text-tertiary mt-0.5">{label}</div>
        </button>
      </TooltipTrigger>
      <TooltipContent side="top">
        <p className="max-w-[200px] text-center">{tooltip}</p>
      </TooltipContent>
    </Tooltip>
  )
}
```

### 7 个卡片定义

| # | label | field | tooltip | valueClassName |
|---|-------|-------|---------|----------------|
| 1 | 本周活跃 | `activeSubItems` | 本周有进展记录，或计划周期与本周重叠的子事项总数（含本周新完成） | — |
| 2 | 本周新完成 | `newlyCompleted` | 本周内实际完成（actualEndDate 落在本周）的子事项数 | `text-success-text` |
| 3 | 进行中 | `inProgress` | 状态为"进行中"且本周活跃的子事项数 | `text-primary-600` |
| 4 | 阻塞中 | `blocked` | 状态为"阻塞中"且本周活跃的子事项数 | `text-error` |
| 5 | 未开始 | `pending` | 已创建但尚未启动（状态为 pending）且本周活跃的子事项数 | `text-secondary` |
| 6 | 暂停中 | `pausing` | 状态为"暂停中"且本周活跃的子事项数 | `text-warning` |
| 7 | 逾期中 | `overdue` | 计划截止日在本周结束前已过、尚未完成/关闭且本周活跃的子事项数 | `text-error` |

### 错误状态

当 API 请求失败时（`isError` 为 true），`StatsBar` 所有卡片显示 `"-"`。`isError` 由 `WeeklyViewPage` 的 `useQuery` 返回，直接在父组件内联处理：`stats` 为 undefined 时各卡片 value 传 `"-"`，不新增 prop。

## Interface 7: Frontend Unit Tests

**文件**: `frontend/src/pages/WeeklyViewPage.test.tsx`（新增测试）

覆盖：
- 7 个卡片数字正确渲染（`data-testid` 断言）
- hover tooltip 内容正确（`userEvent.hover`）
- 点击 tooltip 展开/收起（`userEvent.click`）
- `aria-describedby` 属性存在

**文件**: `frontend/src/lib/status.test.ts`（更新现有测试）

- 所有现有 `isOverdue` 测试用例补充 `referenceDate` 参数
- 新增：`referenceDate` 为过去时间，expectedEndDate 在其之后 → false

## Error Handling

### 后端验证路径

`view_handler.go` 调用 `dto.ParseWeekStart(weekStartStr)` 解析并校验 `weekStart` 参数（格式 + 必须为周一 + 不得为未来日期）。校验失败时返回 `apperrors.ErrValidation`；错误中间件将 `apperrors.ErrValidation` 映射为 HTTP 400，响应体 code 字段为常量 `apperrors.CodeValidation`（值 `"VALIDATION_ERROR"`）。

```go
// pkg/apperrors/errors.go（已有，本次不新增）
var ErrValidation = errors.New("validation error")
const CodeValidation = "VALIDATION_ERROR"
```

`buildWeeklyGroups` 本身不做参数校验，仅处理已验证的输入，不返回 error。

### 场景汇总

| 场景 | 处理方式 |
|------|----------|
| `weekStart` 格式错误或非周一 | `dto.ParseWeekStart` 返回 `apperrors.ErrValidation` → handler 返回 400 + `"VALIDATION_ERROR"` |
| `weekStart` 在未来 | 同上 |
| API 超时 / 500 / 权限拒绝 | `useQuery` `isError=true`，StatsBar 所有卡片显示 `"-"` |
| `weekEnd` 为 undefined（isOverdue 调用方） | 调用方抛出 `throw new Error("weekEnd is required")`，不允许静默回退 today（PRD 5.4 要求） |
| `si.ExpectedEndDate` 为 nil（后端逾期计数） | nil 检查，不计入逾期 |

## Testing Strategy

| 层 | 测试类型 | 工具 | 文件 | 覆盖率目标 |
|----|----------|------|------|------------|
| 后端 DTO | 无需测试（纯数据结构） | — | — | — |
| 后端 Service | 单元测试（table-driven） | `go test` + `testify/assert` | `view_service_test.go` | `buildWeeklyGroups` 行覆盖率 ≥ 90% |
| 前端 lib | 单元测试 | `vitest` + `@testing-library/react` | `status.test.ts` | `isOverdue` 所有分支覆盖 |
| 前端 Component | 单元测试 | `vitest` + `@testing-library/react` + `userEvent` | `WeeklyViewPage.test.tsx` | 7 个卡片 + 3 种 tooltip 状态（hover/click/focus）全覆盖 |
| E2E | 端到端 | Playwright | `weekly-view.spec.ts` | tooltip 交互 + 7 列布局（xl 断点）各 1 个场景 |

## PRD Coverage Map

| PRD AC | 设计元素 |
|--------|----------|
| Story 1: 统计栏显示 pending/pausing/overdue 数字 | Interface 1, 2, 4, 6 |
| Story 2: hover tooltip 300ms 后显示，内容与方案一致 | Interface 6 (StatCard, delayDuration=300) |
| Story 3: 移动端点击展开/收起 tooltip | Interface 6 (onClick toggle open) |
| Story 4: Tab 聚焦 tooltip 可见，aria-describedby | Interface 6 (TooltipTrigger as button, Radix aria) |
| 5.3 响应式布局 | Interface 6 (grid-cols-2/4/7) |
| 5.4 isOverdue 签名更新 | Interface 5 |
| 性能：P99 不超出基线 +50ms | 无新增接口，仅在现有循环内加 3 个 if 分支，O(n) 不变 |
