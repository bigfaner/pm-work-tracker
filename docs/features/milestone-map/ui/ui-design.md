---
created: 2026-05-12
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: 里程碑图

## Design System

基于项目现有设计系统（Tailwind CSS v4 + Radix UI + CVA）。核心特征：
- 蓝色主色调（Accent `#2563eb`），紧凑信息密度（13px body text）
- 白色卡片 + `#f8fafc` 页面背景，border 分隔（非 shadow）
- 状态徽章系统：success=蓝, warning=橙, error=红
- 240px 固定侧边栏 + 流式主内容区
- 对话框 rounded-xl, 按钮 rounded-lg, 输入框 rounded-md

### Token Reference

| Token | Value | Usage |
|-------|-------|-------|
| accent-light | `#3b82f6` | 进度条背景、高亮填充 |
| accent-bg | `#eff6ff` | 里程碑 Badge 背景 |
| accent-ring | `#bfdbfe` | 选中节点 ring-2 颜色 |
| elevation-2 | `0 4px 6px -1px rgb(0 0 0/0.1), 0 2px 4px -2px rgb(0 0 0/0.1)` | 节点悬停阴影 |
| elevation-3 | `0 10px 15px -3px rgb(0 0 0/0.1), 0 4px 6px -4px rgb(0 0 0/0.1)` | 详情面板阴影 |

## Component: 里程碑图列表视图（UF-1 第一级）

### Placement

- **Mode**: new-page
- **Target**: /milestones
- **Position**: 独立页面，侧边栏导航"事项清单"和"甘特图"之间。默认视图（第一级），点击卡片进入时间线视图（第二级）

### Layout Structure

```
┌─────────────────────────────────────────────────────────────┐
│ 页面标题"里程碑图"            [+ 创建里程碑图] [刷新]         │
│ 筛选：状态(全部▾)                                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ 产品 MVP     │  │ 二期迭代     │  │ 技术债务清理  │       │
│  │ [实施中]     │  │ [待实施]     │  │ [已完成]     │       │
│  │ 4里程碑 8事项│  │ 3里程碑 0事项│  │ 2里程碑 5事项│       │
│  │ ██████░░ 60% │  │ ░░░░░░░░  0% │  │ ██████████100%│      │
│  │ ●─○─○─○     │  │ ○─○─○       │  │ ●─●         │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│                                                             │
│  ┌ ─ ─ ─ ─ ─ ─ ┐                                           │
│  │  +           │                                           │
│  │ 创建里程碑图  │                                           │
│  └ ─ ─ ─ ─ ─ ─ ┘                                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

- 页面标题区：h1 "里程碑图"（18px font-semibold） + 右侧操作按钮
- 筛选栏：状态筛选下拉框（全部/规划中/已评审/待实施/实施中/已完成）
- 卡片网格：`grid-template-columns: repeat(auto-fill, minmax(340px, 1fr))`，gap 16px
- 每张卡片：名称 + 状态 Badge + 里程碑数量/事项数量 + 进度条 + 节点缩略图（dot+line 序列）
- 虚线创建卡片：dashed border，居中显示"+"图标和文字

### 里程碑图状态 Badge 颜色映射

| Status | Badge Variant | Visual |
|--------|--------------|--------|
| 规划中 | neutral | 灰色背景，secondary text |
| 已评审 | warning | 黄色背景，warning text |
| 待实施 | warning | 橙色背景，warning text |
| 实施中 | accent（蓝） | 蓝色背景，accent text |
| 已完成 | success | 绿色背景，success text |

### States

| State | Visual | Behavior |
|--------|--------|----------|
| Loading | 3 个骨架屏卡片 | 数据加载中 |
| Populated | 卡片网格 | — |
| Empty | 居中空状态图标 + "暂无里程碑图" + 创建按钮 | 团队无里程碑图 |
| Error | 错误图标 + "加载失败" + 重试按钮 | API 失败 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击卡片 | 进入该里程碑图的时间线视图（UF-1 第二级） | 视图切换动画 |
| 点击"+ 创建里程碑图"按钮 | 弹出创建里程碑图弹窗（UF-7） | — |
| 状态筛选 | 过滤卡片列表 | 即时筛选 |
| 点击刷新 | 重新加载列表 | 刷新按钮 loading |

---

## Component: 里程碑时间线视图（UF-1 第二级）

### Placement

- **Mode**: existing-page:/milestones（第二级视图，从列表视图切换进入）
- **Target**: /milestones
- **Position**: 列表视图的替换内容，面包屑导航 + 时间线

### Layout Structure

```
┌─────────────────────────────────────────────────────────────┐
│ 里程碑图 > 产品 MVP                                          │
│ 产品 MVP           [← 返回列表] [+ 创建里程碑] [刷新] [缩放] │
│ 筛选：状态(全部▾)  搜索里程碑...                               │
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

- 页面标题区：h1 "里程碑图"（18px font-semibold） + 右侧操作按钮
- 工具栏：状态筛选下拉框 + 搜索框 + 缩放控件（周/月/季三个按钮组）
- 时间线区域：卡片容器，overflow-x-auto，min-height 400px
  - 时间轴：横向时间刻度线，刻度标签按缩放级别（周=每周, 月=每月, 季=每季）
  - 里程碑节点：垂直定位在时间轴上方，按 planned_completion_date 横向排列，相邻节点之间用单向箭头连接（从右侧节点卡片右边缘 → 指向下一节点卡片左边缘，2px 实线 + 三角箭头，颜色 text-tertiary）
  - MI 条目：每个节点下方展开的竖向列表

**Timeline Layout Algorithm**（节点定位逻辑）：

以时间线容器左边缘为原点，`originDate` 为视口最左端对应的日期（由初始加载/缩放/滚动决定）。

每个里程碑节点的 x 坐标：
```
x = (milestone.planned_completion_date - originDate) / (1 day) * pxPerDay
```

各缩放级别像素比：
| Zoom | pxPerDay | 刻度间隔 | 适用场景 |
|------|----------|---------|---------|
| 周   | 80       | 7 天    | 里程碑 ≤ 10 |
| 月   | 20       | 30 天   | 里程碑 10–30 |
| 季   | 5        | 90 天   | 里程碑 > 30 |

重叠处理：若两节点 x 坐标差 < 40px（w-40 = 160px 节点宽度的 25%），将右侧节点向右推至最小间距 40px。MI 连接线从节点底部中心引至对应 MI 行，MI 行垂直堆叠无重叠。

**里程碑节点卡片**（核心视觉单元）：
```
┌──────────────────┐
│ ● MVP 发布  80%  │  ← 状态色点 + 名称 + 完成度
│   2026-06-30     │  ← 计划完成时间
│   3 个事项       │  ← 关联数量
└──────────────────┘
```
- 尺寸：w-40，border `#e2e8f0`，rounded-xl，bg white
- 状态色点：not_started=#64748b（WCAG AA 4.6:1 on white）, in_progress=#3b82f6, completed=#1d4ed8, cancelled=#cbd5e1
- 完成度：13px text-secondary，进度条 bg accent-light，高度 4px
- 悬停：bg `#f8fafc`，shadow elevation-2，cursor pointer
- 选中：border accent，ring-2 accent-ring

**MI 条目行**（里程碑下方）：
```
┌─ 连线 ─ [MI-0001] 需求分析     进度 60%  状态:进行中  ──┐
```
- 紧凑单行：h-8，px-3，rounded-md，text 13px
- 拖拽手柄：左侧 6px grip icon，cursor grab
- 连线：1px border-tertiary 虚线从里程碑节点中心引出
- 悬停：bg `#f8fafc`，显示拖拽提示

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 骨架屏：3 个 w-40 h-20 矩形 + 6 条 w-32 h-8 行条，animate-pulse；屏幕阅读器通过 `aria-busy="true"` + 隐藏文本 `sr-only` "加载里程碑数据中" 通报加载状态 | 数据加载中 |
| Empty | 居中提示图标 + "暂无里程碑"（14px text-secondary）+ 创建按钮（如无权限则隐藏） | 团队无里程碑 |
| Populated | 完整时间线渲染 | 正常态 |
| No Permission | 居中 403 图标 + "无权限访问" + 返回按钮 | 缺少 milestone:read |
| Error | 居中错误图标 + "加载失败" + 重试按钮 | API 失败 |

**State Transitions**：Loading → Populated（API 成功返回数据）；Loading → Empty（API 返回空数组）；Loading → Error（API 失败）；Error → Loading（点击重试）；Empty → Loading（创建第一个里程碑后自动刷新）。No Permission 为独立入口态，不与其他状态转换。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击里程碑节点 | 打开详情面板（UF-3） | 节点高亮（border accent），右侧面板滑入 |
| 悬停里程碑节点 | 显示 tooltip | tooltip 显示"X 个事项，Y 已完成"（12px，dark bg） |
| 点击"创建里程碑" | 打开创建弹窗（UF-2） | — |
| 拖拽 MI 条目到另一里程碑 | 调用 API 更新 milestone_key | 拖拽中：MI 条目 opacity-50 + 移动跟随鼠标；目标节点 border-dashed accent；完成后两个里程碑完成度动画更新 + 底部显示撤销 toast "MI-XXXX 已移至 [里程碑名]"，含"撤销"按钮（5s 自动消失），点击撤销调用 API 恢复原 milestone_key |
| 在详情面板点击 MI 行 × 按钮 | 调用 API 置空 milestone_key | × 按钮 hover 时显示（opacity 0→1，color text-tertiary），hover × 时变为 error-text 颜色；点击后行移除，原里程碑完成度更新 + 底部显示撤销 toast（5s 超时） |
| 点击缩放控件（周/月/季） | 重新计算时间刻度 | 时间轴刻度标签更新，里程碑位置重排，transition 200ms |
| 点击 MI 条目 | 跳转主事项详情 | — |
| 点击"刷新"按钮 | 重新调用里程碑列表 API | 时间线区域显示 Loading 状态，完成后恢复 Populated |
| 输入搜索关键词 | 过滤里程碑节点 | 实时过滤（debounce 300ms），不匹配的节点隐藏，匹配节点高亮名称 |

### Keyboard Navigation

| Key | Context | Action |
|-----|---------|--------|
| Tab | 页面全局 | 标题 → 工具栏控件（创建、刷新、缩放）→ 筛选下拉 → 搜索框 → 时间线节点 |
| ArrowRight / ArrowLeft | 焦点在时间线节点上 | 移动焦点到相邻里程碑节点 |
| Enter | 焦点在里程碑节点上 | 打开详情面板（UF-3） |
| Enter | 焦点在 MI 条目上 | 跳转主事项详情 |
| Escape | 详情面板（UF-3）打开 | 关闭面板，焦点回到触发的里程碑节点 |
| Escape | 弹窗（UF-2）打开 | 关闭弹窗，焦点回到触发按钮 |
| Space | 拖拽手柄获得焦点时 | 进入键盘拖拽模式（见下方） |

**Keyboard drag-and-drop**：Space 进入拖拽模式后，ArrowRight/Left 切换目标里程碑节点（目标节点显示 border-dashed accent 高亮），Enter 确认放置，Escape 取消。拖拽过程中 `aria-live="polite"` 区域播报 "正在将 [MI名称] 移至 [里程碑名称]"。完成后播报 "[MI名称] 已分配至 [里程碑名称]"。

### ARIA Roles

| Element | Role / Attributes |
|---------|-------------------|
| 时间线容器 | `role="application" aria-label="里程碑时间线"` |
| 里程碑节点 | `role="button" aria-label="[名称]，[状态]，完成度 [X]%，[N] 个事项" tabindex="0"` |
| MI 条目行 | `role="listitem"`，外层列表 `role="list" aria-label="[里程碑名称] 关联事项"` |
| 筛选下拉 | `role="listbox" aria-label="按里程碑筛选"` |
| 详情面板 | `role="dialog" aria-label="里程碑详情" aria-modal="true"` |
| 创建/编辑弹窗 | `role="dialog" aria-label="创建里程碑" / "编辑里程碑" aria-modal="true"` |
| 拖拽状态 | `aria-live="polite" aria-atomic="true"` 区域，播报拖拽来源、目标、结果 |
| 搜索框 | `role="searchbox" aria-label="搜索里程碑"` |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 节点名称 | milestone.name | API |
| 节点日期 | milestone.planned_completion_date | API |
| 状态色点 | milestone.status | API → 颜色映射 |
| 完成度进度条+数字 | completion（计算值） | API → AVG(main_items.completion) |
| 关联数量 | 关联 MI count | API |
| MI 编号 | main_item.code | API |
| MI 标题 | main_item.title | API |
| MI 完成度 | main_item.completion | API |
| MI 状态徽章 | main_item.status | API → Badge 组件 |
| 刷新按钮 | — | 触发里程碑列表 API 重新请求 |
| 搜索框 | milestone.name 过滤 | 客户端过滤已加载的里程碑列表 |

---

## Component: 创建/编辑里程碑弹窗（UF-2）

### Placement

- **Mode**: existing-page:/milestones
- **Target**: /milestones
- **Position**: 页面居中 modal overlay（z-50），从时间线页面触发

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
│  计划完成时间 *                  │
│  ┌─────────────────────┐   │
│  │ 选择日期              │   │  ← Date picker
│  └─────────────────────┘   │
│                             │
├─────────────────────────────┤
│              [取消] [确认]   │  ← Dialog footer
└─────────────────────────────┘
```

- Dialog 尺寸：sm (400px)
- 表单字段间距：mb-4
- 日期选择器：使用 Radix UI Popover + 日历组件

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Create Mode | 标题"创建里程碑"，表单为空 | — |
| Edit Mode | 标题"编辑里程碑"，表单预填当前值 | — |
| Submitting | 确认按钮 loading（spinner + "保存中..."） | 禁用所有输入 |
| Validation Error | 字段下方红色文本（error-text），输入框 border error | 显示具体错误信息 |

**State Transitions**：Create/Edit Mode → Submitting（点击确认且校验通过）；Create/Edit Mode → Validation Error（校验失败）；Submitting → 弹窗关闭（API 成功）；Submitting → Validation Error（API 返回字段错误）；Validation Error → Create/Edit Mode（用户修正输入）。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击确认 | 校验 → 调用 API | 成功：弹窗关闭 + 时间线刷新；失败：显示错误 |
| 点击取消 / [×] | 关闭弹窗 | 不保存 |
| 点击 overlay | 关闭弹窗 | 不保存 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 名称输入框 | milestone.name | Form state |
| 日期选择器 | milestone.planned_completion_date | Form state |

---

## Component: 里程碑详情面板（UF-3）

### Placement

- **Mode**: existing-page:/milestones
- **Target**: /milestones
- **Position**: 页面右侧 slide-over panel，宽 360px，从时间线点击里程碑节点触发

### Layout Structure

```
┌─────────────────────────────────┐
│ [×]                             │  ← 关闭按钮
│                                 │
│ MVP 发布                        │  ← 名称（面板标题）
│                                 │
│ 计划完成时间          [进行中▾] [编辑] │  ← 同行：日期左，状态Badge+编辑右
│ 2026-06-30                      │
│                                 │
│ 完成度  80%                     │
│ ████████████░░░░                │  ← 进度条
│                                 │
│ ── 关联事项 (3) ─────── [+ 添加] │  ← 标题行右侧内联添加按钮
│ ┌─ MI-0001 需求分析  60% 进行中 ×│  ← 悬停显示 × 解绑按钮
│ ├─ MI-0003 UI设计   100% 已完成 ×│
│ └─ MI-0005 API开发    80% 进行中 ×│
│                                 │
│ ── 危险操作 ────────────────── │
│ [删除里程碑]                     │  ← Danger button
└─────────────────────────────────┘
```

- Panel：fixed right-0 top-0 h-full w-[360px]，bg white，shadow-elevation-3，z-40
- 滑入动画：translate-x 300ms ease-out
- 元信息行：flex justify-between，左侧计划完成时间（label+value 纵向），右侧状态 Badge（可点击切换）+ 编辑按钮
- 状态切换：点击状态 Badge → Radix DropdownMenu 弹出可用转换选项（复用 StatusTransitionDropdown 组件）
- 关联事项标题行：flex justify-between，左侧"关联事项 (N)"，右侧"+ 添加" ghost 按钮
- MI 行解绑：每行右侧 × 按钮，悬停时显示（opacity 0→1），点击触发解绑 API + 撤销 toast

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 骨架屏：标题矩形 + 元信息行 + 3 行列表条 | 数据加载中 |
| Populated | 完整信息展示 | — |
| Cancelled | 全局灰色调（text-tertiary），关联列表为空，状态下拉无可用选项 | 里程碑已取消 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击状态 Badge | 打开状态下拉菜单，显示可用转换 | 复用 StatusTransitionDropdown 组件；终态无选项时显示"暂无可用流转" |
| 点击编辑 | 打开编辑弹窗（UF-2） | — |
| 点击删除 | 打开删除确认弹窗 | — |
| 点击关联 MI 行 | 跳转 /items/:mainItemId | — |
| 点击 MI 行 × 按钮 | 调用解绑 API | 成功：行移除 + 撤销 toast；失败：错误提示 |
| 点击 "+ 添加" 按钮 | 打开快速添加事项弹窗（UF-3a） | — |
| 点击 [×] / overlay | 关闭面板 | slide-out 动画 |

### Delete Confirmation Dialog

从 UF-3 删除按钮触发的内联组件，覆盖在详情面板之上。

**Layout**：
```
┌─────────────────────────────────┐
│        ⚠ 确定删除里程碑 [名称]？   │  ← 16px font-medium，⚠ warning icon
│                                 │
│  关联的 X 个事项将解除绑定，       │  ← 13px text-secondary
│  里程碑数据不可恢复。             │
│                                 │
│            [取消] [确认删除]      │  ← 确认按钮 danger variant（bg error text white）
└─────────────────────────────────┘
```

- Dialog 尺寸：sm (400px)，居中于详情面板
- 取消按钮：secondary variant；确认删除按钮：danger variant（bg `#dc2626` text white）
- 确认按钮 hover: bg `#b91c1c`

**States**：

| State | Visual | Behavior |
|-------|--------|----------|
| Default | 确认按钮可点击 | 弹窗打开 |
| Submitting | 确认按钮 loading（spinner + "删除中..."），取消按钮禁用 | API 调用中 |
| Delete Error | 确认按钮恢复可点击，弹窗底部显示红色错误文本"删除失败，请重试" | API 失败 |

**Interactions**：

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击确认删除 | 调用删除 API | 成功：弹窗关闭 + 详情面板关闭 + 时间线刷新；失败：显示错误 |
| 点击取消 | 关闭弹窗 | 不删除 |
| 点击 overlay / Escape | 关闭弹窗 | 不删除 |

**Data Binding**：

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 里程碑名称 | milestone.name | UF-3 当前数据 |
| 关联事项数 | 关联 MI count | UF-3 当前数据 |

### Quick Add MainItem Dialog（UF-3a）

从 UF-3 详情面板关联事项标题行右侧的"+ 添加"按钮触发，覆盖在详情面板之上。复用 CreateMainItemDialog 组件。

**Layout**：
```
┌─────────────────────────────────┐
│ 新建主事项                  [×] │  ← 标题 + 关闭按钮
│                                 │
│ 标题 *                          │
│ ┌─────────────────────────────┐ │
│ │ 请输入标题                   │ │
│ └─────────────────────────────┘ │
│                                 │
│ 优先级 *          负责人 *      │  ← 两列布局
│ ┌──────────┐    ┌──────────┐   │
│ │ P2    ▾  │    │ 请选择 ▾ │   │
│ └──────────┘    └──────────┘   │
│                                 │
│ 开始时间 *      预期完成时间 *  │  ← 两列布局
│ ┌──────────┐    ┌──────────┐   │
│ │2026-05-12│    │          │   │
│ └──────────┘    └──────────┘   │
│                                 │
│ 所属里程碑 *                    │
│ ┌─────────────────────────────┐ │
│ │ MVP 发布（disabled）         │ │
│ └─────────────────────────────┘ │
│                                 │
│ 描述                            │
│ ┌─────────────────────────────┐ │
│ │ 请输入描述（可选）           │ │
│ └─────────────────────────────┘ │
│                                 │
│              [取消] [确认]      │
└─────────────────────────────────┘
```

- Dialog 尺寸：440px
- 字段顺序与 CreateMainItemDialog 完全一致：标题 → 优先级+负责人（两列）→ 开始时间+预期完成时间（两列）→ 所属里程碑 → 描述
- 所属里程碑字段：disabled 状态，bg-bg-secondary，预填当前里程碑名称（从详情面板上下文传入）
- 所属里程碑 BizKey 通过隐藏字段传入，不显示在表单中

**Interactions**：

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击确认 | 校验必填字段（标题/负责人/开始时间/预期完成时间） → 创建 MI + 绑定当前里程碑 | 成功：弹窗关闭 + 详情面板 MI 列表刷新；校验失败：字段下方红色提示 |
| 点击取消 / [×] / Escape | 关闭弹窗 | 不添加 |
| 点击 overlay | 关闭弹窗 | 不添加 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 名称 | milestone.name | API |
| 计划完成时间 | milestone.planned_completion_date | API |
| 状态标签 | milestone.status | API → Badge |
| 完成度 | completion（计算值） | API |
| 状态切换按钮 | 状态机允许的转换 | status → allowed transitions |
| MI 列表 | main_items[] | API |
| MI 编号/标题/状态/完成度 | code/title/status/completion | API |

---

## Component: 事项清单页里程碑筛选（UF-4）

### Placement

- **Mode**: existing-page:/items
- **Target**: /items
- **Position**: 筛选栏，现有"状态"和"负责人"下拉框右侧

### Layout Structure

- 新增下拉框：与现有筛选器样式一致（h-10 rounded-md border-border-dark）
- 选项：全部 / 未分配 / [各里程碑名称列表]
- MI 列表每条记录：标题右侧增加里程碑名称标签（Badge 样式，text-xs，rounded-full，bg accent-bg text-accent）

```
现有筛选栏：  [状态▾] [负责人▾] [里程碑▾]  [搜索...] [重置] [刷新]
                                            ↑ 新增

MI 条目：
┌ MI-0001  需求分析  ···  [MVP发布]  ← 新增的里程碑标签
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default | 下拉框显示"里程碑：全部" | 页面加载 |
| Filtered | 下拉框显示选中的里程碑名称 | 列表过滤 |
| Error | 下拉框显示"里程碑：加载失败"（text-error），禁用选择 | API 返回失败，下拉不可操作 |

**State Transitions**：Default → Filtered（选择筛选值）；Filtered → Default（重置或切换团队）；Default/Filtered → Error（里程碑列表 API 失败）；Error → Default（页面刷新重新加载）。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 选择里程碑 | 过滤 MI 列表 | 只显示该里程碑下的 MI |
| 选择"未分配" | 过滤 | 只显示 milestone_key 为空的 MI |
| 切换团队 | 重置为"全部" | 下拉选项刷新 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 下拉选项 | 里程碑列表 | API（排除 cancelled） |
| MI 里程碑标签 | 关联里程碑 name | main_item → milestone |

---

## Component: 主事项编辑弹窗里程碑选择器（UF-5）

### Placement

- **Mode**: existing-page:/items/:mainItemId
- **Target**: /items/:mainItemId
- **Position**: 编辑主事项 modal 中，"负责人"下拉框下方

### Layout Structure

```
┌─────────────────────────────┐
│ 编辑主事项              [×]  │
├─────────────────────────────┤
│ 优先级   [P1 ▾]              │
│ 负责人   [张三 ▾]            │
│ 所属里程碑 [MVP发布 ▾]       │  ← 新增字段
│ 描述     [........]          │
├─────────────────────────────┤
│                  [取消] [保存]│
└─────────────────────────────┘
```

- 下拉框：与现有"负责人"下拉框样式一致
- 选项：未分配 / [各非 cancelled 里程碑名称]

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default | 显示当前里程碑名称 | 弹窗打开 |
| Empty | 显示"未分配" | MI 未绑定里程碑 |
| No Milestones | 下拉框仅"未分配" | 团队无里程碑 |
| Error | 下拉框显示"加载失败"（text-error），禁用选择；保留当前绑定值不变 | API 返回失败 |

**State Transitions**：Default/Empty → Error（里程碑列表 API 失败）；Error → Default（重新打开弹窗触发重新加载）；No Milestones 为 Default 的子集（列表为空）。选择操作仅更新 formState，不触发独立的状态转换。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 选择里程碑下拉选项 | 更新表单 formState.milestone_key（不触发 API） | 下拉框显示选中的里程碑名称 |
| 选择"未分配" | 更新 formState.milestone_key 为 null（不触发 API） | 下拉框显示"未分配" |
| 点击父弹窗"保存"按钮 | 校验全部字段 → 调用 MainItem.update API（含 formState.milestone_key）→ 成功后 MI milestone_key 更新 | 成功：父弹窗关闭 + 详情页刷新；失败：显示错误 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 下拉框 | milestone_key | main_item.milestone_key |
| 选项列表 | 里程碑列表 | API（排除 cancelled） |

---

## Component: 表格视图里程碑列（UF-6）

### Placement

- **Mode**: existing-page:/table
- **Target**: /table
- **Position**: 表格列，"标题"列和"优先级"列之间

### Layout Structure

```
| ... | 标题 | 优先级 | 里程碑 | 负责人 | 进度 | 状态 | 预期完成 | ...
| ... | 需求分析 | P1 | MVP发布 | 张三 | 60% | 进行中 | 06/30  | ...
| ... | 数据设计 | P2 | -      | 张三 | 40% | 进行中 | 07/15  | ...
```

- 列宽：w-32
- 已分配：text-secondary，里程碑名称
- 未分配：text-tertiary，"-"
- 表头可点击排序（升序/降序图标）
- 筛选：列头下拉筛选器（全部 / 未分配 / 各里程碑）

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Assigned | 里程碑名称（text-secondary） | 可点击筛选 |
| Unassigned | "-"（text-tertiary） | 可点击筛选 |
| Error | 单元格显示 "—" + tooltip "里程碑信息加载失败" | 里程碑名称查找失败，降级为空值显示 |

**State Transitions**：Assigned ↔ Unassigned（MI 绑定/解绑里程碑后刷新）；Assigned/Unassigned → Error（里程碑名称查找失败）；Error → Assigned/Unassigned（表格刷新后恢复）。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击列头排序 | 按里程碑名称排序 | asc/desc 切换 |
| 点击筛选 | 过滤 | 只显示匹配行 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 单元格 | 里程碑名称 | main_item → milestone.name |
| 排序方向 | sort 参数 | URL query |
| 筛选值 | filter 参数 | URL query |
