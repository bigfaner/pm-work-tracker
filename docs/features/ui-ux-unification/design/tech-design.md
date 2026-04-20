---
created: 2026-04-20
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: UI/UX 统一与优化

## Overview

纯前端改造，无后端接口变更。8 个独立改造点，可并行实施。核心新增产物为 `WeekPicker` 共享组件和 `weekUtils.ts` 工具函数。其余改造均为现有文件的局部修改。

## Architecture

### Layer Placement

所有改动均在前端展示层：

```
pages/          ← F1(Sidebar), F2, F4, F5, F6, F7, F8 的修改点
components/
  shared/       ← F3 新增 WeekPicker.tsx
  layout/       ← F1 Sidebar.tsx 修改
src/utils/      ← F3 新增 weekUtils.ts（从 WeeklyViewPage 提取）
```

### Component Diagram

```
WeeklyViewPage ──uses──► WeekPicker
ReportPage     ──uses──► WeekPicker
                              │
                         weekUtils.ts
                    (getWeekNumber, toLocalDateString,
                     getCurrentWeekStart, addWeeks)
```

### Dependencies

- `lucide-react`（已有）：F2 图标、F3 WeekPicker 箭头图标（`ChevronLeft`, `ChevronRight`）
- 无新增 npm 依赖

## Interfaces

### WeekPicker Props

```typescript
// src/components/shared/WeekPicker.tsx
interface WeekPickerProps {
  /** 当前周起始日期，格式 YYYY-MM-DD（周一） */
  weekStart: string
  /** 切换周次时的回调，传入新的 weekStart */
  onChange: (weekStart: string) => void
  /** 最大可选周起始日期，默认为当前周周一 */
  maxWeek?: string
  className?: string
}
```

### weekUtils.ts 导出接口

```typescript
// src/utils/weekUtils.ts

/** 返回今天所在周的周一日期，格式 YYYY-MM-DD */
export function getCurrentWeekStart(): string

/** 返回给定日期所在的 ISO 周次（1-53） */
export function getWeekNumber(dateStr: string): number

/** 返回给定日期所在年份（ISO 周年，可能与自然年不同） */
export function getISOWeekYear(dateStr: string): number

/** 将 Date 对象格式化为 YYYY-MM-DD 字符串（本地时区） */
export function toLocalDateString(d: Date): string

/** 在 weekStart 基础上加减 n 周，返回新的 weekStart */
export function addWeeks(weekStart: string, n: number): string

/** 将 weekStart 格式化为显示文字，如 "2026年第16周  04/13 ~ 04/19" */
export function formatWeekLabel(weekStart: string): string
```

## Data Models

### WeekPicker 内部状态

WeekPicker 为受控组件，无内部状态。`weekStart` 由父组件持有，通过 `onChange` 回调更新。

```typescript
// 父组件使用示例
const [weekStart, setWeekStart] = useState(getCurrentWeekStart)
const maxWeek = useMemo(() => getCurrentWeekStart(), [])

<WeekPicker weekStart={weekStart} onChange={setWeekStart} maxWeek={maxWeek} />
```

## Implementation Plan

### F1 — Sidebar 菜单重排（`Sidebar.tsx`）

**改动范围**：`frontend/src/components/layout/Sidebar.tsx`

将 `navItems`、`adminItems`、`teamItem` 三个数组合并重组为两个数组：

```typescript
const businessItems = [
  { key: '/items',     label: '事项清单', icon: LayoutGrid },
  { key: '/item-pool', label: '待办事项', icon: Inbox },
  { key: '/weekly',    label: '每周进展', icon: Calendar },
  { key: '/gantt',     label: '整体进度', icon: AlignLeft, permission: 'view:gantt' },
  { key: '/report',    label: '周报导出', icon: FileDown },
]

const adminItems = [
  { key: '/teams', label: '团队管理', icon: Users },
  { key: '/users', label: '用户管理', icon: UserCog, permission: 'user:read' },
  { key: '/roles', label: '角色管理', icon: Shield, permission: 'user:manage_role' },
]
```

渲染时，`adminItems` 第一项上方加 `<div className="mt-2 pt-2 border-t border-border" />` 分隔线。

---

### F2 — 表格操作按钮加图标

**改动范围**：4 个页面文件，仅修改按钮内容，不改变 onClick 逻辑。

图标统一规格：`<Icon className="w-3.5 h-3.5" />`，Button 内容改为：
```tsx
<Button variant="ghost" size="sm">
  <Pencil className="w-3.5 h-3.5" />
  编辑
</Button>
```

Button 组件已有 `gap-1.5` flex 布局，图标与文字自动对齐，无需额外样式。

| 文件 | 新增 import |
|------|------------|
| `UserManagementPage.tsx` | `Pencil, ToggleLeft, ToggleRight` |
| `RoleManagementPage.tsx` | `Pencil, Trash2` |
| `TeamDetailPage.tsx` | `Crown, UserMinus, RefreshCw` |
| `ItemPoolPage.tsx` | `ArrowUpCircle, ArrowDownCircle, XCircle` |

`修改状态`按钮图标根据 `user.status` 动态选择：
```tsx
{user.status === 'enabled'
  ? <ToggleRight className="w-3.5 h-3.5" />
  : <ToggleLeft className="w-3.5 h-3.5" />
}
```

---

### F3 — WeekPicker 组件（新建）

**新建文件**：
- `frontend/src/utils/weekUtils.ts`
- `frontend/src/components/shared/WeekPicker.tsx`
- `frontend/src/utils/weekUtils.test.ts`（单元测试）
- `frontend/src/components/shared/WeekPicker.test.tsx`（组件测试）

**weekUtils.ts 实现要点**：
- 从 `WeeklyViewPage.tsx` 提取 `getCurrentWeekStart`、`getWeekNumber`、`toLocalDateString` 函数
- 新增 `addWeeks(weekStart, n)`：使用 `setDate(date.getDate() + n * 7)` 而非毫秒计算，规避夏令时（DST）切换日的 ±1 小时偏差
- 新增 `formatWeekLabel(weekStart)` 返回 `"2026年第16周  04/13 ~ 04/19"` 格式字符串
- 新增 `getISOWeekYear(dateStr)` 用于处理跨年周次（如 12 月底的第 1 周属于下一年）

**WeekPicker.tsx 实现要点**：
```tsx
export function WeekPicker({ weekStart, onChange, maxWeek, className }: WeekPickerProps) {
  const effectiveMax = maxWeek ?? getCurrentWeekStart()
  const isAtMax = weekStart >= effectiveMax

  const weekEnd = toLocalDateString(new Date(new Date(weekStart).getTime() + 6 * 86400000))
  const label = formatWeekLabel(weekStart)  // "2026年第16周  04/13 ~ 04/19"

  return (
    <div className={cn("inline-flex items-center gap-1 h-8 rounded-md border border-border bg-white px-2", className)}>
      <button onClick={() => onChange(addWeeks(weekStart, -1))} ...>
        <ChevronLeft className="w-4 h-4" />
      </button>
      <span className="text-sm text-secondary px-2 whitespace-nowrap">{label}</span>
      <button onClick={() => !isAtMax && onChange(addWeeks(weekStart, 1))}
              disabled={isAtMax} ...>
        <ChevronRight className="w-4 h-4" />
      </button>
    </div>
  )
}
```

**WeeklyViewPage.tsx 改动**：
- 删除本地 `getCurrentWeekStart`、`getWeekNumber`、`toLocalDateString`、`toWeekInputValue`、`handleWeekChange` 函数
- 用 `weekUtils` 中的函数替代
- 将 `<input type="week" ...>` 替换为 `<WeekPicker weekStart={weekStart} onChange={setWeekStart} maxWeek={maxWeek} />`
- 删除 `maxWeek` 的 `toWeekInputValue` 转换（WeekPicker 直接接受 `YYYY-MM-DD`）

**ReportPage.tsx 改动**：
- 将 `weekValue` state 从 `"YYYY-Www"` 格式改为 `"YYYY-MM-DD"` 格式（周一日期）
- 删除 `formatWeekLabel` 本地函数，改用 `weekUtils.formatWeekLabel`
- 将 `<Input type="week" ...>` 替换为 `<WeekPicker weekStart={weekValue} onChange={setWeekValue} />`
- `handlePreview` 和 `handleExport` 中的周次转日期逻辑简化：`weekStart` 已是 `YYYY-MM-DD`，直接传给 API

---

### F4 — 详情页信息布局统一

**改动范围**：`SubItemDetailPage.tsx`

当前子事项详情页基本信息卡片已使用 `grid grid-cols-4 gap-4`，标签/值结构与主事项详情页一致（`text-xs text-tertiary mb-1` + `text-[13px]`）。

对比后发现主要差异：
- 子事项详情页「编号」字段用 `font-mono text-xs bg-bg-alt px-1.5 py-0.5 rounded` 内联样式，与主事项详情页的 `<Badge>` 不同
- 子事项详情页「当前完成度」单独一个字段，主事项详情页无此字段

**改动**：将子事项详情页「编号」字段改用 `<Badge variant="default" className="font-mono text-[11px]">` 与主事项详情页保持一致。其余字段结构已对齐，无需改动。

---

### F5 — 可跳转文字链接高亮

**统一样式**：`text-primary-600 hover:text-primary-700 hover:underline`

**改动范围**（全量排查结果）：

| 文件 | 位置 | 当前样式 | 改为 |
|------|------|----------|------|
| `ItemViewPage.tsx` | SummaryView 主事项标题 | `text-primary hover:text-primary-600` | 统一样式 |
| `ItemViewPage.tsx` | DetailView 主事项标题 | `font-medium text-primary hover:text-primary-600` | 统一样式（保留 font-medium） |
| `ItemViewPage.tsx` | DetailView 子事项标题 | `font-medium text-primary hover:text-primary-600 ml-4` | 统一样式（保留 font-medium ml-4） |
| `MainItemDetailPage.tsx` | 子事项标题 | `text-[13px] font-medium text-primary hover:text-primary-600` | 统一样式（保留 text-[13px] font-medium） |
| `SubItemDetailPage.tsx` | 所属主事项标题 | `text-[13px] font-medium text-primary hover:text-primary-600` | 统一样式（保留 text-[13px] font-medium） |
| `WeeklyViewPage.tsx` | 主事项标题 | `text-[15px] font-semibold text-primary hover:text-primary-600` | 统一样式（保留 text-[15px] font-semibold） |
| `WeeklyViewPage.tsx` | SubItemRow 子事项标题 | `text-[13px] text-text hover:text-primary-600` | 统一样式（保留 text-[13px]） |
| `WeeklyViewPage.tsx` | completedNoChange 子事项 | `text-[13px] text-text hover:text-primary-600` | 统一样式（保留 text-[13px]） |
| `ItemPoolPage.tsx` | 已分配主事项链接 | `font-medium text-primary-600 hover:text-primary-700` | 补充 `hover:underline` |
| `TeamManagementPage.tsx` | 团队名称 | `font-medium text-primary hover:underline` | 统一样式（保留 font-medium） |
| `TableViewPage.tsx` | 事项标题 | `font-medium text-primary hover:text-primary-600` | 统一样式（保留 font-medium） |

**原则**：只改颜色和 hover 样式，保留原有的 `font-*`、`text-[size]`、`ml-*`、`truncate` 等布局类。

---

### F6 — 「追加进度」按钮位置调整

**改动范围**：`SubItemDetailPage.tsx`

1. 删除标题栏中的 `<PermissionGuard code="progress:update"><Button onClick={() => setAppendOpen(true)}>追加进度</Button></PermissionGuard>`
2. 在「进度记录」`<CardHeader>` 中添加该按钮：

```tsx
<CardHeader>
  <h3 className="text-sm font-semibold text-primary m-0">进度记录</h3>
  <PermissionGuard code="progress:update">
    <Button size="sm" onClick={() => setAppendOpen(true)}>追加进度</Button>
  </PermissionGuard>
</CardHeader>
```

`CardHeader` 已有 `flex items-center justify-between` 布局，按钮自动对齐右侧。

---

### F7 — 每周进展子事项进度显示

**改动范围**：`WeeklyViewPage.tsx` 中的 `SubItemRow` 组件

在 `<Link>` 标题之后、`assigneeName` 之前插入：

```tsx
<span className={cn(
  "text-[11px] font-semibold whitespace-nowrap",
  item.completion === 100 ? "text-success-text" : "text-secondary"
)}>
  {item.completion}%
</span>
```

---

### F8 — 按钮文案统一

**改动范围**：`ItemViewPage.tsx`

将第 397 行 `创建主事项` 改为 `新增主事项`。

## Error Handling

本需求无新增错误处理逻辑。WeekPicker 为纯展示组件，不涉及异步操作。

## Testing Strategy

### Unit Tests

**工具**：`vitest`

**`weekUtils.test.ts`** — 覆盖所有导出函数：

| 测试用例 | 验证点 |
|---------|--------|
| `getCurrentWeekStart` 返回本周周一 | 格式 YYYY-MM-DD，星期一 |
| `getWeekNumber` 跨年边界（12月底/1月初） | ISO 周次计算正确 |
| `addWeeks(weekStart, 1)` | 加 7 天，结果为周一 |
| `addWeeks(weekStart, -1)` | 减 7 天，结果为周一 |
| `addWeeks` DST 边界（夏令时切换日） | 不因 ±1 小时偏差产生错误日期 |
| `formatWeekLabel` 格式验证 | 包含年份、周次、日期范围 |

**`WeekPicker.test.tsx`** — 工具：`vitest` + `@testing-library/react`

| 测试用例 | 验证点 |
|---------|--------|
| 渲染显示正确的周次文字 | label 包含年份和周次 |
| 点击「‹」触发 onChange（weekStart - 7天） | onChange 被调用，参数正确 |
| 点击「›」触发 onChange（weekStart + 7天） | onChange 被调用，参数正确 |
| 已到 maxWeek 时「›」按钮 disabled | button 有 disabled 属性 |
| 已到 maxWeek 时点击「›」不触发 onChange | onChange 未被调用 |

### Integration Tests

**工具**：`Playwright`

现有 `weekly-view.spec.ts` 需更新：将 `type="week"` 输入框的选择操作改为点击 WeekPicker 的箭头按钮。

### Coverage Target

`weekUtils.ts`：100%（纯函数，易于覆盖）
`WeekPicker.tsx`：≥ 80%

## Security Considerations

### Threat Model
无新增安全风险。本需求为纯 UI 改造，不涉及权限逻辑变更、数据传输或用户输入校验。

### Mitigations
N/A

## Open Questions

- [x] WeekPicker 是否需要支持键盘操作（左右方向键切换周次）？→ 本期不做，后续可扩展
- [x] `weekUtils.ts` 放在 `src/utils/` 还是 `src/lib/`？→ 放 `src/utils/`（`src/lib/` 目前只有 `utils.ts` 工具函数，语义不同）

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| 使用 `dayjs` 处理周次计算 | API 简洁，处理跨年边界更可靠 | 引入额外依赖（虽然 dayjs 已在 package.json 中） | 现有逻辑已经正确，提取复用即可，无需引入新依赖 |
| 将链接样式提取为 CSS class | 统一维护 | 需要修改 Tailwind 配置，影响范围更大 | 直接修改 className 更直观，改动更小 |

### File Change Summary

| 文件 | 操作 | 改动点 |
|------|------|--------|
| `src/utils/weekUtils.ts` | 新建 | 周次计算工具函数 |
| `src/utils/weekUtils.test.ts` | 新建 | 单元测试 |
| `src/components/shared/WeekPicker.tsx` | 新建 | WeekPicker 组件 |
| `src/components/shared/WeekPicker.test.tsx` | 新建 | 组件测试 |
| `src/components/layout/Sidebar.tsx` | 修改 | F1 菜单重排 |
| `src/pages/WeeklyViewPage.tsx` | 修改 | F3 替换 WeekPicker + F7 进度显示 |
| `src/pages/ReportPage.tsx` | 修改 | F3 替换 WeekPicker |
| `src/pages/ItemViewPage.tsx` | 修改 | F5 链接高亮 + F8 按钮文案 |
| `src/pages/MainItemDetailPage.tsx` | 修改 | F5 链接高亮 |
| `src/pages/SubItemDetailPage.tsx` | 修改 | F4 布局统一 + F5 链接高亮 + F6 按钮位置 |
| `src/pages/UserManagementPage.tsx` | 修改 | F2 按钮图标 |
| `src/pages/RoleManagementPage.tsx` | 修改 | F2 按钮图标 |
| `src/pages/TeamDetailPage.tsx` | 修改 | F2 按钮图标 |
| `src/pages/ItemPoolPage.tsx` | 修改 | F2 按钮图标 + F5 链接高亮 |
| `src/pages/TeamManagementPage.tsx` | 修改 | F5 链接高亮 |
| `src/pages/TableViewPage.tsx` | 修改 | F5 链接高亮 |
