---
created: 2026-05-12
updated: 2026-06-07
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: 里程碑图

## Design System

基于项目现有设计系统。核心特征：
- Tailwind CSS v4 主题 token（`text-primary`, `text-secondary`, `text-tertiary`, `bg-bg-alt`, `border-border` 等）
- 组件库：`src/components/ui/` 下的 Radix UI + CVA 组件（Dialog, Select, Input, Button, Badge, Popover, Table 等）
- 域组件：`src/components/shared/` 下的 StatusBadge, PriorityBadge, ProgressBar, MemberSelect, StatusTransitionDropdown, StatusTagFilter
- 状态色系统：通过 `StatusBadge` variant 映射（success/warning/error/neutral）
- 240px 固定侧边栏（`Sidebar.tsx`）+ 流式主内容区（`AppLayout.tsx`）
- 13px body text，紧凑信息密度
- 白色卡片 + `bg-bg-alt` 页面背景，`border-border` 分隔

### Token Reference

| Token | Usage |
|-------|-------|
| `text-primary` | 主要文本 |
| `text-secondary` | 次要文本 |
| `text-tertiary` | 辅助文本 |
| `bg-primary-50` / `text-primary-700` | 选中/激活态 |
| `bg-bg-alt` | 悬停背景 |
| `border-border` | 边框分隔 |
| `accent-*` (自定义) | 里程碑专属强调色，用于 Badge 和进度条 |

### Reusable Components

| 组件 | 位置 | 用途 |
|------|------|------|
| `Dialog` | `components/ui/dialog.tsx` | 创建/编辑弹窗 |
| `Select` | `components/ui/select.tsx` | 下拉选择 |
| `Button` | `components/ui/button.tsx` | 操作按钮 |
| `Badge` | `components/ui/badge.tsx` | 状态标签 |
| `Input` | `components/ui/input.tsx` | 文本输入 |
| `Popover` | `components/ui/popover.tsx` | 日期选择器容器 |
| `StatusBadge` | `components/shared/StatusBadge.tsx` | 状态徽章（复用 variant 映射） |
| `StatusTransitionDropdown` | `components/shared/StatusTransitionDropdown.tsx` | 状态切换下拉（复用） |
| `ProgressBar` | `components/shared/ProgressBar.tsx` | 进度条 |
| `MemberSelect` | `components/shared/MemberSelect.tsx` | 成员选择器 |
| `ConfirmDialog` | `components/shared/ConfirmDialog.tsx` | 删除确认弹窗（复用） |

## Navigation Integration

### Sidebar (MODIFIED)

**文件**: `frontend/src/components/layout/Sidebar.tsx`

在 `businessItems` 数组中，`/item-pool` 之后、`/weekly` 之前新增：

```tsx
import { Milestone, Calendar, /* ... */ } from 'lucide-react'

const businessItems = [
  { key: '/items', label: '事项清单', icon: LayoutGrid },
  { key: '/item-pool', label: '待办事项', icon: Inbox },
  { key: '/milestones', label: '里程碑图', icon: Milestone },  // NEW
  { key: '/weekly', label: '每周进展', icon: Calendar },
  { key: '/gantt', label: '整体进度', icon: AlignLeft, permission: 'view:gantt' },
  { key: '/report', label: '周报导出', icon: FileDown },
]
```

### Route (MODIFIED)

**文件**: `frontend/src/App.tsx`

```tsx
<Route path="/milestones" element={<MilestonesPage />} />
<Route path="/milestones/:mapId" element={<MilestoneDetailPage />} />
```

### API Module (NEW)

**文件**: `frontend/src/api/milestones.ts`

遵循现有 API 模块模式（参考 `mainItems.ts`）：
- 使用共享 `client` 实例
- `getTeamId()` 从 Zustand store 获取当前团队 ID
- 每个函数返回 `client.get/post/put/delete(...)` 的 Promise

```ts
// 主要函数
createMilestoneMap(data: { mapName: string; mapDesc?: string; assigneeBizKey: string; plannedStartDate?: string; plannedEndDate?: string })
listMilestoneMaps(params?: { name?: string; assigneeKey?: string; status?: string; page?: number; pageSize?: number })
getMilestoneMap(mapId: string)
updateMilestoneMap(mapId: string, data: { mapName?: string; mapDesc?: string; assigneeBizKey: string; plannedStartDate?: string; plannedEndDate?: string })
deleteMilestoneMap(mapId: string)
changeMilestoneMapStatus(mapId: string, status: string)
getMilestoneMapAvailableTransitions(mapId: string)

createMilestone(mapId: string, data: { milestoneName: string; expectedEndDate: string; milestoneDesc?: string })
listMilestonesByMap(mapId: string)
listMilestonesByTeam(params?: { name?: string; status?: string; excludeCancelled?: boolean })
getMilestone(milestoneId: string)
updateMilestone(milestoneId: string, data: { milestoneName?: string; expectedEndDate?: string; milestoneDesc?: string })
deleteMilestone(milestoneId: string)
changeMilestoneStatus(milestoneId: string, status: string)
getMilestoneAvailableTransitions(milestoneId: string)
```

### Types (MODIFIED)

**文件**: `frontend/src/types/index.ts`

新增类型定义：

```ts
// MilestoneMap
interface MilestoneMap {
  bizKey: string
  teamKey: string
  creatorKey: string
  creatorName: string
  assigneeKey: string
  assigneeName: string
  mapName: string
  mapDesc: string
  mapStatus: string
  statusName: string
  plannedStartDate: string | null
  plannedEndDate: string | null
  milestoneCount: number
  itemCount: number
  overallProgress: number
  createTime: string
  dbUpdateTime: string
}

// Milestone
interface Milestone {
  bizKey: string
  teamKey: string
  milestoneMapKey: string
  milestoneName: string
  milestoneDesc: string
  expectedEndDate: string | null
  milestoneStatus: string
  statusName: string
  completion: number
  relatedMICount: number
  createTime: string
  dbUpdateTime: string
}
```

## Component: 里程碑图列表视图（UF-1 第一级）

### Placement

- **Mode**: new-page
- **Target**: `/milestones`
- **Position**: 独立页面，侧边栏"待办事项"和"每周进展"之间。默认视图（第一级），点击卡片进入时间线视图（第二级）

### Page Structure

```
frontend/src/pages/
  MilestonesPage.tsx              ← 列表页入口（milestones.html）
  MilestoneDetailPage.tsx         ← 详情/时间线页入口（milestone-detail.html）
  milestones/                     ← 子组件目录
    MilestoneMapList.tsx          ← 列表视图
    MilestoneTimeline.tsx         ← 时间线视图
    MilestoneMapCard.tsx          ← 卡片组件
    MilestoneNode.tsx             ← 时间线节点
    MilestoneDetailPanel.tsx      ← 详情面板
    CreateMilestoneDialog.tsx     ← 创建/编辑里程碑弹窗
    CreateMilestoneMapDialog.tsx  ← 创建/编辑里程碑图弹窗
    QuickAddMainItemDialog.tsx    ← 快速添加事项弹窗
```

### Layout Structure

```
┌─────────────────────────────────────────────────────────────┐
│ 页面标题"里程碑图"                    [+ 创建里程碑图] [刷新] │
│ 筛选：[搜索名称...] [负责人▾] [状态▾]                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────────────┐ ┌─────────────────┐ ┌──────────┐│
│  │ 产品 MVP     [实施中] │ │ 二期迭代 [待实施]│ │技术债清理││
│  │ 4里程碑···张三        │ │ 3里程碑···李四  │ │2项···张三││
│  │ 05~12 整体进度  60%   │ │ 08~12 整体进度 0%│ │03~05 100%││
│  │ ●───○───○───○         │ │ ○───○───○       │ │ ●───●    ││
│  └───────────────────────┘ └─────────────────┘ └──────────┘│
│                                                             │
│  ┌ ─ ─ ─ ─ ─ ─ ┐                                           │
│  │  +           │                                           │
│  │ 创建里程碑图  │                                           │
│  └ ─ ─ ─ ─ ─ ─ ┘                                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

- 页面标题区：h1 "里程碑图"（18px font-semibold）+ 右侧操作按钮
- 筛选栏：按名称搜索（`Input`，客户端模糊匹配）+ 负责人下拉（`Select`，团队成员列表）+ 状态下拉（`Select`，全部/规划中/已评审/待实施/实施中/已完成），从左到右依次排列
- 卡片网格：`grid grid-cols-[repeat(auto-fill,minmax(340px,1fr))]` gap-4
- 每张卡片四行布局：第一行名称+`StatusBadge`；第二行里程碑数量+事项数量+负责人（`justify-content: space-between` 左右对齐）；第三行计划时间跨度（左）+"整体进度"+`ProgressBar`+百分比（右）；底部节点缩略图（状态色点+连线）
- 虚线创建卡片：`border-dashed border-border`，居中显示"+"图标和文字

### 状态 Badge 颜色映射

需要在 `frontend/src/utils/status.ts` 中为里程碑状态新增 variant 条目。

**MilestoneMap 状态：**

| Status | StatusBadge variant | Display |
|--------|---------------------|---------|
| planning (规划中) | neutral | 灰色 |
| reviewed (已评审) | warning | 黄色 |
| ready (待实施) | warning | 橙色 |
| executing (实施中) | info | 蓝色 |
| completed (已完成) | success | 绿色 |

**Milestone 状态：**

| Status | StatusBadge variant | Display |
|--------|---------------------|---------|
| not_started (未开始) | neutral | 灰色 |
| in_progress (进行中) | info | 蓝色 |
| completed (已完成) | success | 绿色 |
| cancelled (已取消) | error | 灰色删除线 |

### States

| State | Visual | Behavior |
|--------|--------|----------|
| Loading | 3 个骨架屏卡片 | 数据加载中 |
| Populated | 卡片网格 | — |
| Empty | 居中空状态 + "暂无里程碑图" + 创建按钮 | 团队无里程碑图 |
| Error | 错误提示 + "加载失败" + 重试按钮 | API 失败 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击卡片 | 跳转至该里程碑图的详情页 | 路由跳转（href 至 `/milestones/:mapId`） |
| 点击"+ 创建里程碑图"按钮 | 弹出 `CreateMilestoneMapDialog` | — |
| 名称搜索 | 客户端模糊匹配卡片列表 | debounce 300ms |
| 负责人筛选 | 按负责人过滤卡片列表 | 即时筛选 |
| 状态筛选 | 按状态过滤卡片列表 | 即时筛选 |
| 点击刷新 | 重新加载列表 | 刷新按钮 loading |

---

## Component: 里程碑时间线视图（UF-1 第二级）

### Placement

- **Mode**: separate-page（独立页面，路由跳转，非视图切换）
- **Target**: `/milestones/:mapId`（独立路由，列表页卡片通过 href 跳转至此页）
- **Position**: 列表页点击卡片进入，独立 HTML 页面（`milestone-detail.html`）

### Layout Structure

```
┌─────────────────────────────────────────────────────────────┐
│ 里程碑图 > 产品 MVP                                         │
│                                                             │
│ 产品 MVP  [实施中]                          [编辑] [删除]    │ ← 标题区（卡片外部）
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 负责人      计划开始      计划完成      整体进度         │ │ ← 元数据行（左右对齐）
│ │ 张三        2026-05-01    2026-12-31    60%              │ │
│ │ ─────────────────────────────────────────────────────── │ │
│ │ 产品MVP版本，包含核心功能模块...                          │ │ ← 描述（分隔线下方）
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ 筛选：[搜索名称...] [状态▾]          [+ 创建里程碑] [缩放]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ──────── 6月 ──── 7月 ──── 8月 ──── 9月 ──── 10月 ──────  │
│       │           │           │           │                 │
│    [M1 node]   [M2 node]   [M3 node]   [M4 node]           │
│       │           │           │           │                 │
│    ┌ MI-001    ┌ MI-004    ┌ MI-007    ┌ MI-009            │
│    ├ MI-002    ├ MI-005    ├ MI-008    │                   │
│    └ MI-003    └ MI-006    │           │                   │
│                                                             │
│  ─── 时间轴底线 ────────────────────────────────────────   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

- 面包屑：`text-secondary`，可点击"里程碑图"返回列表页
- **详情标题区**（卡片外部）：名称 + 可点击的 `StatusBadge`（左侧，点击弹出状态切换下拉，复用 `StatusTransitionDropdown`），编辑/删除按钮（右上角，位于卡片外部）
  - 删除按钮仅当 mapStatus === `planning` 时显示
- **基本信息卡片**：白色卡片区域（`detail-section`），包含：
  - 元数据行（`justify-content: space-between` 左右对齐）：负责人、计划开始时间、计划完成时间、整体进度
  - 分隔线（`border-top`）
  - 描述文本（mapDesc，`line-clamp: 3` 最多三行溢出截断，悬停显示 `Tooltip` 展示完整内容）
- 筛选行：搜索名称（`Input`，客户端模糊匹配）+ 状态筛选（`StatusTagFilter`，可多选 toggle）+ 重置按钮 + 刷新按钮，从左到右依次排列；同行右侧为"+ 创建里程碑"按钮和缩放按钮组（周/月/季）
- 时间线区域：`overflow-x-auto`，min-height 400px

**Timeline Layout Algorithm**：

节点位置始终按日期等比计算，缩放**只改变时间轴刻度标签的粒度**，不改变节点位置比例，**不隐藏任何节点**。

每个里程碑节点的 x 坐标：
```
x = (milestone.expectedEndDate - originDate) / totalDays * containerWidth
```

- `totalDays` = 最晚里程碑日期 - 最早里程碑日期（至少 30 天，避免过密）
- `containerWidth` = 时间线容器实际宽度（自适应页面宽度）
- 所有节点等比分布在容器内，节点间自然按日期间隔拉开

**缩放只影响刻度标签**：

| Zoom | 刻度标签间隔 | 效果 |
|------|------------|------|
| 周   | 每 7 天画一条竖线 + 日期标签 | 刻度线密集，适合短跨度（1-3个月），能看清每周位置 |
| 月   | 每 30 天画一条竖线 + 月份标签 | 刻度线适中，适合中等跨度（3-12个月） |
| 季   | 每 90 天画一条竖线 + 季度标签 | 刻度线稀疏，适合长跨度（1年+），减少刻度视觉噪音 |

缩放切换时节点位置不变，仅刻度标签重新渲染，视觉无跳动感。

**节点不重叠保证**：

容器宽度自适应，节点按日期等比分布。当里程碑数量多导致容器内节点过密时（相邻节点 x 坐标差 < 184px，即卡片宽 160px + 间距 24px），容器自动出现水平滚动条（`overflow-x-auto`），用户可横向滚动查看所有节点。**不会出现节点重叠，也不会隐藏任何节点。**

具体实现：计算所有节点等比排列所需的最小容器宽度 `minWidth = nodeCount * 184`，当 `minWidth > 页面实际宽度` 时，将时间线容器宽度设为 `minWidth`，触发水平滚动。

**里程碑节点卡片**：
```
┌──────────────────┐
│ ● MVP 发布  80%  │  ← 状态色点 + 名称 + 完成度
│   2026-06-30     │  ← 计划完成时间
│   3 个事项       │  ← 关联数量
└──────────────────┘
```
- 尺寸 w-40，`border border-border`，`rounded-xl`，bg white
- 状态色点：`not_started`=`text-tertiary`，`in_progress`=`text-primary`，`completed`=`text-success`，`cancelled`=`text-tertiary`
- 完成度：`ProgressBar` 组件
- 悬停：`bg-bg-alt`，cursor pointer
- 选中：`border-primary`，`ring-2 ring-primary-100`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 骨架屏 | 数据加载中 |
| Empty | 居中 "暂无里程碑" + 创建按钮 | 无里程碑 |
| Populated | 完整时间线 | 正常态 |
| Error | 错误提示 + 重试按钮 | API 失败 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击里程碑节点 | 打开详情面板（UF-3） | 节点高亮，右侧面板滑入 |
| 悬停里程碑节点 | 显示 `Tooltip` | "X 个事项，Y 已完成" |
| 点击"创建里程碑" | 弹出 `CreateMilestoneDialog`（UF-2） | — |
| 点击 MI 条目 | 跳转 `/items/:mainItemId` | 路由跳转 |
| 点击缩放控件 | 重新计算时间刻度 | transition 200ms |
| 名称搜索 | 客户端模糊匹配里程碑节点 | debounce 300ms，不匹配的节点隐藏 |
| 状态筛选 | 过滤里程碑节点 | 即时筛选 |
| 点击标题区状态 Badge | `StatusTransitionDropdown` 弹出可用转换 | 复用现有组件，不可用状态灰色不可点击 |
| 点击标题区"编辑" | 打开编辑里程碑图弹窗（UF-7） | — |
| 点击标题区"删除" | 打开 `ConfirmDialog` | 仅 mapStatus === `planning` 时可见 |
| 点击面包屑"里程碑图" | 返回列表页 | 路由跳转至 `/milestones` |
| 拖拽 MI 条目到另一里程碑 | 调用 API 更新 milestone_key | 拖拽中 opacity-50 + 目标高亮；完成时显示撤销 toast（5s） |

### Keyboard Navigation

| Key | Context | Action |
|-----|---------|--------|
| Tab | 页面全局 | 标题 → 工具栏 → 筛选 → 搜索 → 时间线节点 |
| ArrowRight / ArrowLeft | 焦点在时间线节点上 | 移动焦点到相邻里程碑节点 |
| Enter | 焦点在里程碑节点上 | 打开详情面板（UF-3） |
| Enter | 焦点在 MI 条目上 | 跳转主事项详情 |
| Escape | 详情面板打开 | 关闭面板，焦点回到触发的节点 |
| Escape | 弹窗打开 | 关闭弹窗 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 节点名称 | milestone.milestoneName | API |
| 节点日期 | milestone.expectedEndDate | API |
| 状态色点 | milestone.milestoneStatus | API → 颜色映射 |
| 完成度 | milestone.completion | API（计算值） |
| 关联数量 | milestone.relatedMICount | API |
| MI 编号/标题 | mainItem.code / mainItem.title | API |

---

## Component: 创建/编辑里程碑弹窗（UF-2）

### Placement

- **Mode**: Dialog overlay
- **Target**: `/milestones`
- **Component**: `CreateMilestoneDialog.tsx`

### Layout Structure

```
┌─────────────────────────────┐
│ 创建里程碑              [×] │  ← Dialog header
├─────────────────────────────┤
│                             │
│  名称 *                     │
│  ┌─────────────────────┐   │
│  │ 请输入里程碑名称      │   │  ← Input
│  └─────────────────────┘   │
│                             │
│  计划完成时间 *              │
│  ┌─────────────────────┐   │
│  │ 选择日期              │   │  ← 日期选择器
│  └─────────────────────┘   │
│                             │
│  描述                       │
│  ┌─────────────────────┐   │
│  │ 请输入描述（可选）    │   │  ← Textarea
│  │                     │   │
│  └─────────────────────┘   │
│                             │
├─────────────────────────────┤
│              [取消] [确认]   │
└─────────────────────────────┘
```

- 使用 `Dialog` 组件，sm 尺寸（400px）
- 日期选择器：使用 `Popover` + 项目现有日期选择模式
- 描述字段：`Textarea`，可选填

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Create Mode | 标题"创建里程碑"，表单为空 | — |
| Edit Mode | 标题"编辑里程碑"，表单预填当前值 | — |
| Submitting | 确认按钮 loading | 禁用所有输入 |
| Validation Error | 字段下方红色文本，输入框 error 样式 | 显示具体错误 |

---

## Component: 里程碑详情面板（UF-3）

### Placement

- **Mode**: Slide-over panel
- **Target**: `/milestones`
- **Component**: `MilestoneDetailPanel.tsx`

### Layout Structure

```
┌─────────────────────────────────┐
│ [×]                             │
│                                 │
│ MVP 发布                        │  ← 名称
│                                 │
│ 描述            [进行中▾] [编辑] │  ← Row 1: label + status + edit
│ 完成产品MVP版本的核心功能开...   │  ← Row 2: 描述文本 (line-clamp:6, tooltip)
│                                 │
│ 计划完成时间                     │
│ 2026-06-30                      │
│                                 │
│ 进度                     80%    │
│ ████████████░░░░                │  ← ProgressBar
│                                 │
│ ── 关联事项 (3) ─────── [+ 添加] │
│ ┌─ MI-0001 需求分析  60% 进行中 ×│
│ ├─ MI-0003 UI设计   100% 已完成 ×│
│ └─ MI-0005 API开发    80% 进行中 ×│
│                                 │
│ ── 危险操作 ────────────────── │
│ [删除里程碑]                     │
└─────────────────────────────────┘
```

- Panel：fixed right-0 top-0 h-full w-[360px]，bg white，shadow，z-40
- 滑入动画：translate-x 300ms ease-out
- 描述区域：两行布局——第一行"描述"标签（左，`text-tertiary` 12px）+ `StatusTransitionDropdown` + 编辑按钮（右，`flex-shrink: 0`）；第二行为描述文本（全宽，`line-clamp: 6` 最多六行溢出截断，`display: -webkit-box`，`-webkit-box-orient: vertical`，鼠标悬浮 `Tooltip` 展示完整内容）
- 进度标签：`text-tertiary` 12px，显示"进度"而非"完成度"，与其它 label 样式一致
- 状态切换：复用 `StatusTransitionDropdown` 组件
- MI 行解绑：悬停显示 × 按钮，点击触发解绑 API + 撤销 toast

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 骨架屏 | 数据加载中 |
| Populated | 完整信息展示 | — |
| Cancelled | 全局灰色调（`text-tertiary`），关联列表为空 | 里程碑已取消 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击状态 Badge | `StatusTransitionDropdown` 弹出可用转换 | 复用现有组件 |
| 点击编辑 | 打开编辑弹窗（UF-2） | — |
| 点击删除 | 打开 `ConfirmDialog` | 仅 milestoneStatus === `not_started` 或 `cancelled` 时可见 |
| 点击关联 MI 行 | 跳转 `/items/:mainItemId` | — |
| 点击 MI 行 × 按钮 | 调用解绑 API | 行移除 + 撤销 toast |
| 点击"+ 添加" | 弹出 `QuickAddMainItemDialog`（UF-3a） | — |
| 点击 [×] / overlay | 关闭面板 | slide-out 动画 |

### Delete Confirmation

使用现有 `ConfirmDialog` 组件：
- 标题："确定删除里程碑 [名称]？"
- 描述："关联的 X 个事项将解除绑定，里程碑数据不可恢复。"
- 确认按钮：danger variant

### Quick Add MainItem Dialog（UF-3a）

复用 `CreateMainItemDialog` 组件（`src/pages/item-view/CreateMainItemDialog.tsx`）：
- 所属里程碑字段 disabled，预填当前里程碑名称
- BizKey 通过 props 传入

---

## Component: 事项清单页里程碑筛选（UF-4）

### Placement

- **Mode**: existing-page
- **Target**: `/items`
- **Position**: `ItemViewPage.tsx` 筛选栏，`StatusTagFilter` 和 `MemberSelect` 右侧

### Layout Structure

```
现有筛选栏：  [状态▾] [负责人▾] [里程碑▾]  [搜索...] [重置] [刷新]
                                            ↑ 新增

MI 条目：
┌ MI-0001  需求分析  ···  [MVP发布]  ← 新增的里程碑 Badge
```

- 新增 `Select` 组件：与现有筛选器样式一致
- 选项：全部 / 未分配 / [各里程碑名称]
- MI 条目新增里程碑名称标签：`Badge` 组件，`text-xs`，`rounded-full`
- 筛选逻辑：传递 `milestoneKey` 参数到 MainItemFilter

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default | 下拉框显示"里程碑：全部" | 页面加载 |
| Filtered | 下拉框显示选中的里程碑名称 | 列表过滤 |
| Error | 下拉框"加载失败"，禁用 | API 失败 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 下拉选项 | 里程碑列表 | `listMilestonesByTeam({excludeCancelled: true})` |
| MI 里程碑标签 | mainItem.milestoneName | MainItem VO enrichment |

---

## Component: 主事项编辑弹窗里程碑选择器（UF-5）

### Placement

- **Mode**: existing-page
- **Target**: `/items/:mainItemId`
- **Position**: `EditMainItemDialog.tsx` 中"负责人"下拉框下方

### Layout Structure

```
┌─────────────────────────────┐
│ 编辑主事项              [×]  │
├─────────────────────────────┤
│ 优先级   [P1 ▾]              │
│ 负责人   [张三 ▾]            │
│ 所属里程碑 [MVP发布 ▾]       │  ← 新增
│ 描述     [........]          │
├─────────────────────────────┤
│                  [取消] [保存]│
└─────────────────────────────┘
```

- 使用 `Select` 组件：与现有"负责人"下拉框样式一致
- 选项：未分配 / [各非 cancelled 里程碑名称]

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default | 显示当前里程碑名称 | 弹窗打开 |
| Empty | 显示"未分配" | MI 未绑定里程碑 |
| No Milestones | 仅"未分配"选项 | 团队无里程碑 |
| Error | "加载失败"，禁用 | API 失败 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 下拉框值 | mainItem.milestoneKey | 主事项数据 |
| 选项列表 | 里程碑列表 | `listMilestonesByTeam({excludeCancelled: true})` |

---

## Component: 表格视图里程碑列（UF-6）

### Placement

- **Mode**: existing-page
- **Target**: `/table`
- **Position**: `TableViewPage.tsx` 表格列定义，"标题"列和"优先级"列之间

### Layout Structure

```
| ... | 标题 | 优先级 | 里程碑 | 负责人 | 进度 | 状态 | 预期完成 | ...
| ... | 需求分析 | P1 | MVP发布 | 张三 | 60% | 进行中 | 06/30  | ...
| ... | 数据设计 | P2 | -      | 张三 | 40% | 进行中 | 07/15  | ...
```

- 列宽 w-32
- 已分配：`text-secondary`
- 未分配：`text-tertiary`，显示"-"
- 表头可点击排序（asc/desc 切换图标）
- 列头筛选：下拉全部 / 未分配 / 各里程碑

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Assigned | 里程碑名称（`text-secondary`） | 正常态 |
| Unassigned | "-"（`text-tertiary`） | MI 未绑定里程碑 |
| Deleted Milestone | "—" | 里程碑已软删除 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击列头排序 | 按里程碑名称排序 | asc/desc 切换 |
| 点击筛选 | 过滤 | 只显示匹配行 |

---

## Component: 创建/编辑里程碑图弹窗（UF-7）

### Placement

- **Mode**: Dialog overlay
- **Target**: `/milestones`
- **Component**: `CreateMilestoneMapDialog.tsx`

### Layout Structure

```
┌─────────────────────────────┐
│ 创建里程碑图            [×] │
├─────────────────────────────┤
│                             │
│  名称 *                     │
│  ┌─────────────────────┐   │
│  │ 请输入里程碑图名称    │   │
│  └─────────────────────┘   │
│                             │
│  负责人 *                   │
│  ┌─────────────────────┐   │
│  │ 选择负责人 *          │   │  ← MemberSelect
│  └─────────────────────┘   │
│                             │
│  计划开始时间               │
│  ┌─────────────────────┐   │
│  │ 选择日期              │   │  ← 日期选择器
│  └─────────────────────┘   │
│                             │
│  计划完成时间               │
│  ┌─────────────────────┐   │
│  │ 选择日期              │   │  ← 日期选择器
│  └─────────────────────┘   │
│                             │
│  描述                       │
│  ┌─────────────────────┐   │
│  │ 请输入描述（可选）    │   │
│  └─────────────────────┘   │
│                             │
├─────────────────────────────┤
│              [取消] [确认]   │
└─────────────────────────────┘
```

- 使用 `Dialog` 组件，sm 尺寸（400px）
- 负责人：复用 `MemberSelect` 组件，必填
- 计划开始/完成时间：使用 `Popover` + 项目现有日期选择模式
- 描述字段：`Textarea`，高度 160px，`resize: vertical`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Create Mode | 标题"创建里程碑图"，表单为空 | — |
| Edit Mode | 标题"编辑里程碑图"，表单预填值 | — |
| Submitting | 确认按钮 loading | 禁用所有输入 |

## Page Composition

| Page | Type | UI Functions | Files |
|------|------|-------------|-------|
| /milestones | new | UF-1, UF-7 | `MilestonesPage.tsx` + `milestones/` 子组件（列表页） |
| /milestones/:mapId | new | UF-1 第二级, UF-2, UF-3, UF-3a | `MilestoneDetailPage.tsx` + `milestones/` 子组件（详情/时间线页） |
| /items | existing | UF-4 | `ItemViewPage.tsx` 筛选栏修改 |
| /items/:mainItemId | existing | UF-5 | `EditMainItemDialog.tsx` 新增字段 |
| /table | existing | UF-6 | `TableViewPage.tsx` 列定义修改 |
