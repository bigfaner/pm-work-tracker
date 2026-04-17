---
created: 2026-04-17
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: PM 工作事项追踪系统 (pm-work-tracker)

## Design System

### Color Tokens

| Token | Hex | Usage |
|-------|-----|-------|
| color-overdue | #ff4d4f | Overdue date text, overdue Gantt bar |
| color-p1 | #fa8c16 | P1 priority badge background |
| color-p2 | #1677ff | P2 priority badge background |
| color-p3 | #8c8c8c | P3 priority badge background |
| color-success | #52c41a | Completed status tag, progress bar (100%) |
| color-warning | #faad14 | At-risk status tag |
| color-processing | #1677ff | In-progress status tag |
| color-default | #d9d9d9 | Not-started, archived tag |

### Status Tag Color Mapping

SubItem states (8):

| Status | antd Tag color |
|--------|---------------|
| 未开始 | default (gray) |
| 进行中 | processing (blue) |
| 待评审 | warning (yellow) |
| 已完成 | success (green) |
| 已关闭 | default (gray) |
| 阻塞中 | error (red) |
| 延期 | orange |
| 归档 | default (gray) |

ItemPool states (3):

| Status | antd Tag color |
|--------|---------------|
| 待分配 | blue |
| 已分配 | green |
| 已拒绝 | red |

### Typography & Spacing
- Base font: system-ui / -apple-system, 14px
- Page content max-width: 1440px, min-width: 1280px
- Sidebar width: 220px (collapsed: 64px)
- Content area padding: 24px
- Card/section gap: 16px

### Navigation (Left Sidebar)
- Fixed left sidebar, 220px wide
- Top: product logo + team switcher (Select dropdown, 180px wide)
- Nav items with icons (antd Menu component, mode="inline"):
  - 事项视图 (AppstoreOutlined) — default
  - 周视图 (CalendarOutlined)
  - 甘特图 (BarChartOutlined)
  - 表格视图 (TableOutlined)
  - 事项池 (InboxOutlined)
  - 周报导出 (ExportOutlined)
- Bottom: user avatar + name + logout (Dropdown)
- SuperAdmin only: 管理后台 link appears at bottom of nav

### Global Interaction Feedback
- Success operations: antd `message.success()` toast, 2s auto-dismiss, top-center
- Error operations: antd `message.error()` toast, 3s auto-dismiss
- Destructive confirmations: antd `Modal.confirm()` with warning icon
- Form validation errors: inline below each field (antd Form `validateStatus`)
- Loading states: antd `Skeleton` for page-level, `Spin` for inline/button


---

## Page 1: 主事项列表（事项视图）

### Layout Structure

```
AppLayout (flex row)
├── SidebarNav (220px fixed)
└── MainContent (flex-1, padding 24px)
    ├── PageHeader (flex row, space-between)
    │   ├── Title "事项视图"
    │   └── Button "新建主事项" (type=primary, PlusOutlined)
    ├── FilterBar (flex row, gap 12px, margin-bottom 16px)
    │   ├── Select placeholder="优先级" (P1/P2/P3, width 120px)
    │   ├── Select placeholder="状态" (8 options, width 140px)
    │   ├── Select placeholder="负责人" (team members, width 140px)
    │   └── Button "重置" (type=link)
    └── MainItemList
        └── MainItemRow (antd Collapse, each panel = one MainItem)
            ├── CollapsePanel header: [编号 Tag] [标题] [优先级 Badge] [负责人 Avatar] [Progress bar 120px] [状态 Tag] [预期完成时间] [操作 Dropdown ...]
            └── CollapsePanel content: SubItemList
                └── SubItemRow (each row): [标题 link] [负责人] [完成度 Progress] [状态 Tag] [预期完成时间]
```

Key Ant Design components: `Layout`, `Menu`, `Collapse`, `Progress`, `Tag`, `Badge`, `Select`, `Button`, `Dropdown`, `Avatar`, `Tooltip`

MainItemRow header layout (horizontal, align center):
- 编号: `Tag` (gray, monospace, width ~80px)
- 标题: text, flex-1, font-weight 500, clickable → MainItem detail page
- 优先级: `Tag` with color (P1=orange, P2=blue, P3=gray), width 40px
- 负责人: `Avatar` (24px) + name text
- 完成度: `Progress` type="line" percent={n} size="small" style={{width:120px}}
- 状态: `Tag` with status color
- 预期完成时间: date text; if overdue + not completed → color #ff4d4f + `Tooltip` "已超期 N 天"
- 操作: `Dropdown` menu (编辑 / 归档), trigger="click", icon=EllipsisOutlined

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | `Skeleton` (3 rows, each with avatar + paragraph lines) | On initial page load and filter change |
| Empty (no items) | Centered `Empty` illustration + text "暂无事项" + Button "新建主事项" | Team has no MainItems |
| Populated | Collapse list, all panels collapsed by default | Normal state |
| Expanded panel | SubItemList visible below header | User clicked panel |
| Overdue item | 预期完成时间 text in #ff4d4f; row header has subtle red-tinted left border (3px solid #ff4d4f) | current date > 预期完成时间 AND status ≠ 已完成/已关闭 |
| P1 item | 优先级 Tag orange + `Badge` dot (orange) on left of title | priority = P1 |
| Filter empty | `Empty` with text "没有符合条件的事项" + Button "重置筛选" | Filters applied, no results |
| Archived (hidden) | Not rendered in list | status = 归档 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click panel header | Toggle expand/collapse SubItemList | Smooth antd Collapse animation |
| Click 标题 text | Navigate to MainItem detail page | Router push |
| Click SubItem 标题 | Navigate to SubItem detail page | Router push |
| Click "新建主事项" | Open `Modal` (title="新建主事项", width=520px) with Form | Modal appears |
| Submit 新建 form | POST create MainItem | `message.success("创建成功")`, modal closes, list refreshes |
| Click 编辑 (Dropdown) | Open same Modal pre-filled with item data | Modal appears with data |
| Click 归档 (Dropdown) | If status valid: `Modal.confirm()` "确认归档？" | On confirm: `message.success("已归档")`, item removed from list |
| Click 归档 (invalid status) | Show `message.warning("请先完成或关闭事项再归档")` | No modal |
| Change any filter Select | Re-fetch list with new params | Loading skeleton briefly, then updated list |
| Click "重置" | Clear all filter Selects | List resets to unfiltered |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 编号 Tag | `mainItem.code` | e.g. "MI-001" |
| 标题 | `mainItem.title` | Max 100 chars |
| 优先级 Tag | `mainItem.priority` | P1/P2/P3 → color map |
| 负责人 Avatar + name | `mainItem.assignee.name`, `mainItem.assignee.avatar` | |
| Progress bar | `mainItem.completionRate` | 0–100, computed from SubItems |
| 状态 Tag | `mainItem.status` | 8-value enum |
| 预期完成时间 | `mainItem.expectedEndDate` | ISO date string |
| Overdue highlight | `mainItem.expectedEndDate` vs today | Client-side computed |
| SubItem row 标题 | `subItem.title` | |
| SubItem 完成度 | `subItem.completionRate` | |
| SubItem 状态 | `subItem.status` | |

### New MainItem Modal Form

Fields (antd Form, layout="vertical"):
- 标题 (Input, required, maxLength=100, showCount)
- 优先级 (Select, required, options P1/P2/P3)
- 负责人 (Select, showSearch, options = team members)
- 预期完成时间 (DatePicker, disabledDate = past dates optional)
- 描述 (TextArea, optional, rows=3)

Footer: Button "取消" (default) + Button "确认" (primary, htmlType=submit)


---

## Page 2: 子事项详情与进度记录

### Layout Structure

```
AppLayout
├── SidebarNav
└── MainContent (padding 24px)
    ├── Breadcrumb: 事项视图 > [MainItem title] > [SubItem title]
    ├── PageHeader
    │   ├── Title: SubItem title (h2)
    │   └── Button "追加进度" (type=primary, PlusOutlined) — visible to assignee + PM only
    ├── InfoCard (antd Card, marginBottom 16px)
    │   └── Descriptions (layout="horizontal", column=3, bordered)
    │       Fields: 编号 / 所属主事项 / 优先级 / 负责人 / 状态 / 预期完成时间 / 当前完成度
    ├── ProgressSummary (antd Card, title="当前完成度")
    │   └── Progress type="line" percent={n} strokeColor={colorByStatus} showInfo
    └── ProgressTimeline (antd Card, title="进度记录")
        ├── Empty state OR
        └── Timeline (antd Timeline, mode="left")
            └── Timeline.Item (per record, sorted ascending by createdAt)
                ├── Time label: "YYYY-MM-DD HH:mm  创建人姓名"
                ├── 完成度: Progress bar (small, inline, width 200px) + percent text
                │   └── [PM only] percent text is clickable → inline InputNumber edit
                ├── 成果: text block (label + content, hidden if empty)
                ├── 卡点: text block (label + content, hidden if empty)
                └── 经验: text block (label + content, hidden if empty)
```

Key components: `Breadcrumb`, `Descriptions`, `Progress`, `Timeline`, `Modal`, `Form`, `InputNumber`, `Tag`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | `Skeleton` (Descriptions skeleton + Timeline skeleton 3 items) | On page load |
| No progress records | `Empty` in timeline card, text "暂无进度记录，点击「追加进度」开始记录" | SubItem has no records |
| Has records | Timeline with items ascending | Normal |
| Overdue | 预期完成时间 in Descriptions shown in #ff4d4f | date exceeded + not completed |
| PM editing completion | Inline `InputNumber` replaces percent text, with confirm/cancel icons | PM clicked percent value |
| Append form open | `Modal` (width=480px) with form | "追加进度" clicked |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click "追加进度" | Open append Modal | Modal slides in |
| Submit append form | POST new progress record | `message.success("进度已记录")`, modal closes, timeline appends new item at bottom |
| Submit with completion < previous | Block submit, show inline error "完成度不能低于上一条记录（当前最高：N%）" | Form stays open |
| PM clicks percent value | Replace with `InputNumber` (0–100) + CheckOutlined / CloseOutlined icons | Inline edit mode |
| PM confirms inline edit | PATCH progress record completion | `message.success("已修正")`, value updates |
| PM cancels inline edit | Revert to display value | No change |
| Click 所属主事项 link | Navigate to MainItem detail page | Router push |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| Breadcrumb MainItem | `subItem.mainItem.title` | Link to MainItem detail |
| Title | `subItem.title` | |
| Descriptions fields | `subItem.code`, `subItem.mainItem.title`, `subItem.priority`, `subItem.assignee.name`, `subItem.status`, `subItem.expectedEndDate`, `subItem.completionRate` | |
| Progress bar | `subItem.completionRate` | |
| Timeline items | `subItem.progressRecords[]` | Sorted by `createdAt` ASC |
| Record time + author | `record.createdAt`, `record.createdBy.name` | |
| Record completion | `record.completionRate` | PM-editable |
| Record 成果/卡点/经验 | `record.achievement`, `record.blocker`, `record.lesson` | Hidden if null/empty |

### Append Progress Modal Form

Fields (Form layout="vertical"):
- 完成度 (InputNumber, required, min=0, max=100, suffix="%", placeholder="0–100")
  - Helper text: "不能低于上一条记录的完成度（当前：N%）"
- 成果 (TextArea, optional, rows=3, placeholder="本阶段完成了什么")
- 卡点 (TextArea, optional, rows=3, placeholder="遇到了哪些阻碍")
- 经验 (TextArea, optional, rows=2, placeholder="有哪些可复用的经验")

Footer: Button "取消" + Button "提交" (primary)


---

## Page 3: 事项池

### Layout Structure

```
AppLayout
├── SidebarNav
└── MainContent (padding 24px)
    ├── PageHeader (flex row, space-between)
    │   ├── Title "事项池"
    │   └── Button "提交到事项池" (type=primary, PlusOutlined) — visible to all members
    ├── FilterBar (flex row, gap 12px, marginBottom 16px)
    │   └── Radio.Group (buttonStyle="solid"): 全部 / 待分配 / 已分配 / 已拒绝
    └── PoolItemList
        └── PoolItemCard (antd Card, hoverable, marginBottom 12px)
            ├── Card header: [标题 text] [状态 Tag] [提交人 + 时间 (secondary text)]
            ├── Card body: 背景 / 预期产出 (collapsed to 2 lines, expandable)
            ├── [待分配] Footer: Button "分配" (primary, small) + Button "拒绝" (danger, small)
            ├── [已分配] Footer: text "已挂载至：[MainItem title link]  负责人：[name]"
            └── [已拒绝] Footer: text "拒绝原因：[reason]" (color #8c8c8c)
```

Key components: `Radio.Group`, `Card`, `Tag`, `Button`, `Modal`, `Form`, `Select`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | `Skeleton` (3 card skeletons) | On page load / filter switch |
| Empty (all) | `Empty` + text "事项池暂无内容" + Button "提交到事项池" | No items at all |
| Empty (filtered) | `Empty` + text "该状态下暂无事项" | Filter yields no results |
| 待分配 | Card with blue border-left (3px), action buttons visible | status = 待分配 |
| 已分配 | Card grayed out (opacity 0.7), no action buttons | status = 已分配 |
| 已拒绝 | Card grayed out (opacity 0.7), rejection reason shown | status = 已拒绝 |
| Submit form open | `Modal` (width=520px) | "提交到事项池" clicked |
| Assign modal open | `Modal` (width=480px) with MainItem selector | "分配" clicked |
| Reject modal open | `Modal` (width=400px) with reason textarea | "拒绝" clicked |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click "提交到事项池" | Open submit Modal | Modal appears |
| Submit pool item form | POST new pool item | `message.success("已提交到事项池")`, modal closes, list refreshes |
| Click "分配" | Open assign Modal | Modal appears |
| Select MainItem in assign modal | Populate assignee Select with MainItem's team members | Assignee Select updates |
| Submit assign modal | PATCH item status → 已分配, create SubItem under MainItem | `message.success("已分配")`, card updates to 已分配 state |
| Submit assign without MainItem | Block submit, inline error "请选择挂载的主事项" | Form stays open |
| Click "拒绝" | Open reject Modal | Modal appears |
| Submit reject modal | PATCH item status → 已拒绝 with reason | `message.success("已拒绝")`, card updates to 已拒绝 state |
| Submit reject without reason | Block submit, inline error "请填写拒绝原因" | Form stays open |
| Switch Radio filter | Re-fetch with status filter | Loading briefly, list updates |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 标题 | `poolItem.title` | |
| 状态 Tag | `poolItem.status` | 待分配/已分配/已拒绝 → color map |
| 提交人 | `poolItem.submittedBy.name` | |
| 提交时间 | `poolItem.submittedAt` | Relative time (e.g. "3小时前") via dayjs |
| 背景 | `poolItem.background` | |
| 预期产出 | `poolItem.expectedOutput` | |
| 已挂载主事项 | `poolItem.mainItem.title` | Link to MainItem detail |
| 负责人 | `poolItem.assignee.name` | After assignment |
| 拒绝原因 | `poolItem.rejectReason` | |

### Submit Pool Item Modal Form

Fields (Form layout="vertical"):
- 标题 (Input, required, maxLength=100, showCount)
- 背景 (TextArea, optional, rows=3)
- 预期产出 (TextArea, optional, rows=3)

### Assign Modal Form

Fields (Form layout="vertical"):
- 挂载主事项 (Select, required, showSearch, options = active MainItems with code+title)
- 负责人 (Select, required, options = team members, updates based on MainItem selection)


---

## Page 4: 周视图

### Layout Structure

```
AppLayout
├── SidebarNav
└── MainContent (padding 24px)
    ├── PageHeader (flex row, space-between, align center)
    │   ├── Title "周视图"
    │   └── WeekPicker (antd DatePicker.WeekPicker, defaultValue=current week, disabledDate=future weeks)
    ├── WeekSummaryBar (flex row, gap 16px, marginBottom 16px)
    │   ├── Statistic "本周更新事项" (number)
    │   ├── Statistic "本周新完成" (number, color #52c41a)
    │   └── Statistic "阻塞中" (number, color #ff4d4f)
    └── WeekGroupList
        ├── SectionHeader "本周新完成" (antd Divider with text, color #52c41a)
        │   └── WeekItemCard[] (see below)
        ├── SectionHeader "本周有进度更新"
        │   └── WeekItemCard[]
        └── SectionHeader "上周完成，本周无变化" (collapsed by default, antd Collapse)
            └── WeekItemCard[]

WeekItemCard (antd Card, size="small", marginBottom 8px):
├── Card title: [SubItem title link] — [MainItem title] (secondary, smaller)
├── Right of title: 负责人 Avatar + name, 状态 Tag
├── Left column "计划": SubItem description / expected output (gray background, padding 8px)
└── Right column "实际": list of this-week progress records
    └── ProgressRecordItem: [完成度 Progress small] [成果 text] [卡点 text] [时间 secondary]
```

Two-column plan/actual layout inside WeekItemCard uses CSS grid: `grid-template-columns: 1fr 1fr`, gap 16px, with column headers "计划" / "实际" in secondary text.

Key components: `DatePicker.WeekPicker`, `Statistic`, `Divider`, `Card`, `Progress`, `Avatar`, `Tag`, `Collapse`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | `Skeleton` (WeekPicker disabled + 3 card skeletons) | On load / week switch |
| Empty week | `Empty` illustration + text "本周暂无事项更新" | Selected week has no data |
| Has data | Grouped sections with cards | Normal |
| No actual records | "实际" column shows "本周暂无进度更新" in gray italic | SubItem in plan but no records this week |
| Future week selected | DatePicker blocks selection (disabledDate) | User tries to pick future week |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Change WeekPicker | Re-fetch data for selected week | Loading skeleton, then updated content |
| Click SubItem title | Navigate to SubItem detail page | Router push |
| Click MainItem title | Navigate to MainItem detail page | Router push |
| Click "上周完成，本周无变化" Collapse | Expand/collapse that section | Smooth animation |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| WeekPicker | `selectedWeek` (state) | ISO week string |
| Section grouping | `subItem.weekStatus` | Computed: new-complete / updated / stale |
| SubItem title | `subItem.title` | |
| MainItem title | `subItem.mainItem.title` | |
| 负责人 | `subItem.assignee.name` | |
| 状态 Tag | `subItem.status` | |
| 计划 column | `subItem.description` | Static plan content |
| 实际 records | `subItem.progressRecords[]` filtered by week range | `record.createdAt` within week |
| Record 完成度 | `record.completionRate` | |
| Record 成果/卡点 | `record.achievement`, `record.blocker` | |
| Summary statistics | Computed from filtered data | Client-side |


---

## Page 5: 甘特图视图

### Layout Structure

```
AppLayout
├── SidebarNav
└── MainContent (padding 24px, overflow hidden)
    ├── PageHeader (flex row, space-between)
    │   ├── Title "甘特图"
    │   └── RangePicker (antd DatePicker.RangePicker, default = current month ± 2 weeks)
    └── GanttContainer (flex row, height calc(100vh - 160px))
        ├── LabelPanel (width 280px, flex-shrink 0, overflow-y auto, border-right 1px solid #f0f0f0)
        │   └── LabelRow (height 40px each, align center, padding 0 12px)
        │       ├── [MainItem row] CollapseToggle icon + 编号 Tag + 标题 (font-weight 500)
        │       └── [SubItem row, indented 24px] 标题 (font-weight 400, color #595959)
        └── TimelinePanel (flex-1, overflow-x auto, overflow-y auto, sync scroll with LabelPanel)
            ├── TimelineHeader (sticky top, height 40px)
            │   └── DateColumns (CSS grid, column width = 28px per day)
            │       ├── MonthLabel row (spans days, background #fafafa)
            │       └── DayLabel row (day number, weekend columns in #fafafa)
            └── TimelineRows (CSS grid rows, height 40px each, matching LabelPanel rows)
                ├── [MainItem row] GanttBar
                │   ├── Bar: rounded rect, height 20px, vertically centered
                │   ├── Color: blue (#1677ff) normal / red (#ff4d4f) overdue / green (#52c41a) completed
                │   ├── Inner fill: darker shade showing completionRate %
                │   └── No time data: dashed gray border, no fill, label "未设置时间"
                └── [SubItem row] GanttBar (same structure, height 16px, slightly lighter shade)
```

Implementation note: GanttContainer uses CSS Grid for the timeline columns. Column count = number of days in visible range. LabelPanel and TimelinePanel share a synchronized vertical scroll via `ref` + `onScroll` handler. No heavy Gantt library — built with CSS grid + absolute-positioned bars.

Today marker: vertical line (2px solid #1677ff, dashed) spanning full height of TimelineRows, with "今天" label at top.

Key components: `DatePicker.RangePicker`, custom CSS Grid layout, `Tooltip` on bar hover, `Tag`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | `Skeleton` (label panel lines + timeline placeholder rows) | On page load / range change |
| Collapsed MainItem | Only MainItem row visible | Default for all items |
| Expanded MainItem | SubItem rows inserted below MainItem row | After click on MainItem row |
| Normal bar | Blue fill with completion overlay | status = 进行中/未开始 |
| Completed bar | Green (#52c41a) solid | status = 已完成 |
| Overdue bar | Red (#ff4d4f), bar extends past today line | expectedEndDate < today AND not completed |
| No time data | Dashed gray outline bar spanning full visible range, label "未设置时间" | mainItem missing start or end date |
| SubItem no time | Inherits MainItem date range, lighter dashed style | subItem missing dates |
| Empty | `Empty` centered in TimelinePanel | No MainItems in team |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click MainItem LabelRow | Toggle expand/collapse SubItem rows | Rows insert/remove with fade animation; toggle icon rotates |
| Hover GanttBar | `Tooltip` showing: 标题 / 开始时间 / 预期完成时间 / 完成度 / 状态 | Tooltip appears |
| Click GanttBar | Navigate to MainItem or SubItem detail page | Router push |
| Change RangePicker | Re-render timeline columns for new date range | Timeline redraws |
| Scroll TimelinePanel horizontally | Timeline header scrolls in sync | Sticky header stays visible |
| Scroll either panel vertically | Both LabelPanel and TimelinePanel scroll in sync | Synchronized scroll |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| LabelRow 编号 | `mainItem.code` | |
| LabelRow 标题 | `mainItem.title` / `subItem.title` | |
| Bar left position | `item.startDate` relative to range start | CSS: `grid-column-start` |
| Bar width | `item.expectedEndDate - item.startDate` in days | CSS: `grid-column-end` |
| Bar fill width | `item.completionRate` % of bar width | Inner div width % |
| Bar color | `item.status` + overdue check | Color token map |
| Today marker position | today relative to range start | CSS grid column |


---

## Page 6: 周报导出

### Layout Structure

```
AppLayout
├── SidebarNav
└── MainContent (padding 24px, max-width 900px)
    ├── PageHeader
    │   └── Title "周报导出"
    ├── ConfigCard (antd Card, marginBottom 16px)
    │   └── Form (layout="inline")
    │       ├── WeekPicker (DatePicker.WeekPicker, defaultValue=current week, disabledDate=future)
    │       └── Button "生成预览" (type=default)
    ├── PreviewCard (antd Card, title="预览", extra=Button "导出 Markdown" primary)
    │   └── PreviewContent
    │       ├── [Loading] Spin centered
    │       ├── [Empty] Empty + text "所选周暂无数据"
    │       └── [Has data] Markdown-rendered preview
    │           ├── H2: 周报标题 (e.g. "2026年第16周 工作周报")
    │           ├── Per MainItem section:
    │           │   ├── H3: [编号] 主事项标题 (完成度 N%)
    │           │   └── Per SubItem:
    │           │       ├── SubItem title + 负责人 + 状态
    │           │       ├── 成果: bullet list
    │           │       └── 卡点: bullet list
    │           └── Footer: 导出时间
    └── [Export in progress] Button shows loading spinner, disabled
```

Preview renders as styled HTML (using a simple markdown-to-HTML renderer or pre-formatted div with monospace font). The actual export downloads a `.md` file.

Key components: `DatePicker.WeekPicker`, `Card`, `Button`, `Spin`, `Empty`, `message`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Initial | ConfigCard visible, PreviewCard shows placeholder "请选择周并点击生成预览" | Page load |
| Generating preview | PreviewCard body shows `Spin` centered | After "生成预览" click |
| Preview ready | Markdown content rendered in PreviewCard | Data fetched successfully |
| No data for week | `Empty` in PreviewCard + text "所选周暂无数据，无法导出" | API returns empty |
| Exporting | "导出 Markdown" button shows loading + disabled | After export click |
| Export complete | File download triggered by browser | Within 5s of click |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Change WeekPicker | Clear current preview, show placeholder | Preview resets |
| Click "生成预览" | Fetch weekly report data, render preview | Spin → content or empty state |
| Click "导出 Markdown" (with data) | Generate .md blob, trigger `<a download>` | Button loading briefly, then file downloads; `message.success("周报已导出")` |
| Click "导出 Markdown" (no data) | Block action | `message.warning("所选周暂无数据")` |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| WeekPicker | `selectedWeek` (state) | |
| Preview title | Computed from week range | e.g. "2026年第16周" |
| MainItem sections | `weekReport.mainItems[]` | |
| MainItem title + completion | `mainItem.title`, `mainItem.completionRate` | |
| SubItem rows | `mainItem.subItems[]` | Only those with this-week records |
| 成果 bullets | `record.achievement` | Aggregated from week's records |
| 卡点 bullets | `record.blocker` | Aggregated from week's records |


---

## Page 7: 登录页

### Layout Structure

```
LoginPage (full viewport, background #f0f2f5, flex center)
└── LoginCard (antd Card, width 400px, padding 40px 32px, box-shadow 0 4px 24px rgba(0,0,0,0.08))
    ├── Logo + ProductName (centered, marginBottom 32px)
    │   ├── Logo image (48px × 48px)
    │   └── H2 "PM 工作事项追踪" (font-size 20px, font-weight 600)
    ├── Form (layout="vertical")
    │   ├── Form.Item label="账号"
    │   │   └── Input (prefix=UserOutlined, placeholder="请输入账号", size="large")
    │   ├── Form.Item label="密码"
    │   │   └── Input.Password (prefix=LockOutlined, placeholder="请输入密码", size="large")
    │   ├── ErrorAlert (antd Alert, type="error", showIcon, marginBottom 16px)
    │   │   └── "账号或密码错误" — only visible on login failure
    │   └── Button "登录" (type=primary, block, size="large", htmlType=submit)
    └── Footer text (centered, color #8c8c8c, font-size 12px): 版本号 / 版权信息
```

No sidebar. Full-page centered layout. No registration or forgot-password links (out of scope).

Key components: `Card`, `Form`, `Input`, `Input.Password`, `Button`, `Alert`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default | Empty form, button enabled | Page load |
| Fields empty | "登录" button disabled | Both inputs empty |
| One field filled | Button still disabled | Only one field has value |
| Both fields filled | "登录" button enabled | Both inputs have value |
| Submitting | Button shows `loading` spinner, disabled; inputs disabled | After click |
| Login failed | `Alert` appears above button with "账号或密码错误"; inputs re-enabled; values preserved | API returns 401 |
| Login success | Redirect to 事项视图 | API returns 200 + token |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Type in either input | Re-evaluate button disabled state | Button enables when both non-empty |
| Press Enter in password field | Submit form | Same as clicking 登录 |
| Click "登录" (enabled) | POST login credentials | Button loading state |
| Login success | Store token, redirect to `/` (事项视图) | Page navigation |
| Login failure | Show error Alert, clear password field, keep account value | Alert visible |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 账号 Input | `form.username` | |
| 密码 Input | `form.password` | Masked, cleared on failure |
| Error Alert visible | `loginError` (state boolean) | |
| Button disabled | `!form.username \|\| !form.password` | Client-side |


---

## Page 8: 团队管理页

### Layout Structure

```
AppLayout
├── SidebarNav
└── MainContent (padding 24px, max-width 960px)
    ├── PageHeader (flex row, space-between)
    │   ├── Title "团队管理"
    │   └── Button "邀请成员" (type=primary, UserAddOutlined) — PM only
    ├── TeamInfoCard (antd Card, marginBottom 16px)
    │   └── Descriptions (column=3, layout="horizontal")
    │       Fields: 团队名称 / PM / 成员数 / 创建时间
    ├── MemberTable (antd Table, size="middle")
    │   Columns:
    │   ├── 姓名 (with Avatar 24px + name)
    │   ├── 账号
    │   ├── 角色 (Tag: PM=blue / 成员=default)
    │   ├── 加入时间 (date)
    │   └── 操作 (Dropdown menu, PM only, hidden for self row)
    │       ├── 设为 PM (only if role = 成员)
    │       └── 移除成员
    └── DangerZone (antd Card, title="危险操作", borderColor=#ff4d4f, marginTop 32px)
        └── Button "解散团队" (danger, ghost) — PM only
```

Key components: `Table`, `Descriptions`, `Tag`, `Avatar`, `Dropdown`, `Modal`, `Form`, `Input`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | `Skeleton` (Descriptions + Table skeleton rows) | On page load |
| Member list | Table with rows | Normal |
| Self row | 操作 column cell empty (no Dropdown) | Current user's own row |
| Invite modal open | `Modal` (width=480px) with search + role select | "邀请成员" clicked |
| Transfer PM confirm modal | `Modal.confirm()` with warning icon | "设为 PM" clicked |
| Remove member confirm | `Modal.confirm()` | "移除成员" clicked |
| Dissolve team modal | `Modal` (width=400px) with team name input | "解散团队" clicked |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click "邀请成员" | Open invite Modal | Modal appears |
| Type in invite search | Debounced search users by name/account | Dropdown options update |
| Submit invite form | POST invite member | `message.success("邀请已发送")`, modal closes, table refreshes |
| Click "设为 PM" | `Modal.confirm()`: "确认将 [name] 设为 PM？此操作不可撤销，您将降为普通成员。" | On confirm: PATCH roles, `message.success("已转让 PM 身份")`, page refreshes |
| Click "移除成员" | `Modal.confirm()`: "确认移除 [name]？" | On confirm: DELETE member, `message.success("已移除")`, row removed |
| Click "解散团队" | Open dissolve Modal with Input for team name | Modal appears |
| Type team name in dissolve modal | Enable "确认解散" button only when input matches team name exactly | Button disabled/enabled |
| Confirm dissolve | DELETE team | `message.success("团队已解散")`, redirect to team selection page |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 团队名称 | `team.name` | |
| PM | `team.pm.name` | |
| 成员数 | `team.memberCount` | |
| 创建时间 | `team.createdAt` | |
| Table rows | `team.members[]` | |
| 姓名 + Avatar | `member.name`, `member.avatar` | |
| 账号 | `member.username` | |
| 角色 Tag | `member.role` | PM / 成员 |
| 加入时间 | `member.joinedAt` | |
| Self row detection | `member.id === currentUser.id` | Hide 操作 column |

### Invite Modal Form

Fields (Form layout="vertical"):
- 搜索用户 (Select, showSearch, mode="multiple" not needed — single invite, remote search with debounce 300ms)
- 角色 (Radio.Group: PM / 成员, default=成员)

Footer: Button "取消" + Button "邀请" (primary, disabled until user selected)


---

## Page 9: 主事项详情页

### Layout Structure

```
AppLayout
├── SidebarNav
└── MainContent (padding 24px)
    ├── Breadcrumb: 事项视图 > [MainItem title]
    ├── PageHeader (flex row, space-between)
    │   ├── Left: [编号 Tag] [标题 h2] [优先级 Tag] [状态 Tag]
    │   └── Right (PM only): Button "编辑" (default) + Dropdown "更多" (归档)
    ├── InfoSection (flex row, gap 16px, marginBottom 16px)
    │   ├── MetaCard (antd Card, flex 1)
    │   │   └── Descriptions (column=2, layout="horizontal", bordered)
    │   │       Fields: 负责人 / 预期完成时间 / 创建时间 / 最后更新
    │   └── ProgressCard (antd Card, width 280px)
    │       ├── Title "整体完成度"
    │       ├── Progress type="circle" percent={n} width=120 strokeColor={colorByStatus}
    │       └── Text "N 个子事项 / M 个已完成"
    ├── SummarySection (antd Card, title="成果 & 卡点聚合", marginBottom 16px, collapsible)
    │   ├── Row "成果汇总": bullet list of all achievement texts from SubItem records
    │   └── Row "卡点汇总": bullet list of all blocker texts from SubItem records
    └── SubItemSection (antd Card, title="子事项列表")
        ├── SubItemFilterBar (flex row, gap 12px, marginBottom 12px)
        │   ├── Select placeholder="优先级" (width 120px)
        │   ├── Select placeholder="状态" (width 140px)
        │   ├── Select placeholder="负责人" (width 140px)
        │   └── Button "新增子事项" (type=primary, small, PlusOutlined) — PM only
        └── SubItemTable (antd Table, size="small")
            Columns: 标题(link) / 负责人(Avatar+name) / 完成度(Progress small) / 状态(Tag) / 预期完成时间 / 操作(PM: 编辑)
```

Key components: `Breadcrumb`, `Descriptions`, `Progress`, `Card`, `Table`, `Tag`, `Select`, `Button`, `Modal`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | `Skeleton` (Descriptions + circle progress placeholder + table skeleton) | On page load |
| No SubItems | SubItemTable replaced by `Empty` + text "暂无子事项，点击新增" + Button | MainItem has no SubItems |
| Overdue | 预期完成时间 in Descriptions shown in #ff4d4f | date exceeded + not completed |
| Filter empty | Table shows `Empty` "没有符合条件的子事项" | Filters yield no results |
| Edit modal open | `Modal` (width=520px) pre-filled with MainItem data | "编辑" clicked |
| New SubItem modal open | `Modal` (width=520px) with SubItem form | "新增子事项" clicked |
| Archive confirm | `Modal.confirm()` | "归档" clicked (valid status) |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click SubItem 标题 | Navigate to SubItem detail page | Router push |
| Click "编辑" | Open edit Modal pre-filled | Modal appears |
| Submit edit form | PATCH MainItem | `message.success("已更新")`, modal closes, page data refreshes |
| Click "归档" (valid) | `Modal.confirm()` "确认归档此主事项？" | On confirm: PATCH status, `message.success("已归档")`, redirect to list |
| Click "归档" (invalid) | `message.warning("请先完成或关闭事项再归档")` | No modal |
| Click "新增子事项" | Open new SubItem Modal | Modal appears |
| Submit new SubItem | POST SubItem | `message.success("子事项已创建")`, modal closes, table row appended |
| Change SubItem filter | Filter table rows client-side | Table updates immediately |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 编号 Tag | `mainItem.code` | |
| 标题 | `mainItem.title` | |
| 优先级 Tag | `mainItem.priority` | |
| 状态 Tag | `mainItem.status` | |
| Descriptions fields | `mainItem.assignee`, `mainItem.expectedEndDate`, `mainItem.createdAt`, `mainItem.updatedAt` | |
| Progress circle | `mainItem.completionRate` | |
| SubItem count text | `mainItem.subItems.length`, completed count | |
| 成果汇总 | All `record.achievement` from all SubItems | Flattened list |
| 卡点汇总 | All `record.blocker` from all SubItems | Flattened list |
| SubItem table rows | `mainItem.subItems[]` | Filtered client-side |

### New SubItem Modal Form

Fields (Form layout="vertical"):
- 标题 (Input, required, maxLength=100, showCount)
- 优先级 (Select, required, P1/P2/P3)
- 负责人 (Select, required, team members)
- 预期完成时间 (DatePicker)
- 描述 (TextArea, optional, rows=3)


---

## Page 10: 表格视图

### Layout Structure

```
AppLayout
├── SidebarNav
└── MainContent (padding 24px)
    ├── PageHeader (flex row, space-between)
    │   ├── Title "表格视图"
    │   └── Button "导出 CSV" (default, DownloadOutlined)
    ├── FilterBar (flex row, gap 12px, marginBottom 16px)
    │   ├── Select placeholder="类型" (主事项/子事项, width 120px)
    │   ├── Select placeholder="优先级" (P1/P2/P3, width 120px)
    │   ├── Select placeholder="负责人" (team members, width 140px)
    │   ├── Select placeholder="状态" (8 options, width 140px)
    │   └── Button "重置" (type=link)
    └── ItemTable (antd Table)
        Columns:
        ├── 类型 (Tag: 主事项=blue / 子事项=default, width 80px)
        ├── 编号 (monospace Tag, width 100px, sorter)
        ├── 标题 (flex-1, clickable link, ellipsis with Tooltip)
        ├── 优先级 (Tag with color, width 80px, sorter + filterDropdown)
        ├── 负责人 (Avatar 20px + name, width 120px, filterDropdown)
        ├── 完成度 (Progress size="small" width 100px + percent text, width 140px, sorter)
        ├── 状态 (Tag with color, width 100px, filterDropdown)
        └── 预期完成时间 (date text, color #ff4d4f if overdue, width 130px, sorter)
        
        Table props: rowKey="id", pagination={{ pageSize:20, showTotal }}, 
                     onRow: onClick → navigate to detail, rowClassName for overdue rows
```

Key components: `Table`, `Tag`, `Progress`, `Avatar`, `Select`, `Button`, `message`

Column sorters use antd built-in `sorter` prop (client-side for current page, server-side for full dataset). Filter dropdowns use antd `filterDropdown` with checkbox list for multi-select.

Overdue rows: `rowClassName` returns `"row-overdue"` → CSS `background: #fff2f0` (light red tint).

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | `Table` with `loading={true}` (built-in spin overlay) | On page load / filter change |
| Default | All active items, sorted by priority desc | Page load |
| Filtered | Only matching rows shown | After filter applied |
| Filter empty | Table `locale.emptyText` = "没有符合条件的事项" with Button "重置筛选" | No results |
| Exporting | "导出 CSV" button shows loading + disabled | After export click |
| Export empty | Button blocked | Filtered result is empty |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click column header (sortable) | Toggle sort asc/desc/none | Column header shows sort arrow |
| Click filter dropdown | Open checkbox list for that column | Dropdown appears |
| Apply column filter | Filter table rows | Table updates |
| Change FilterBar Select | Re-fetch with params | Table loading briefly, then updates |
| Click "重置" | Clear all FilterBar Selects + column filters | Table resets to default |
| Click table row | Navigate to MainItem or SubItem detail page | Router push |
| Click "导出 CSV" (has data) | Generate CSV blob of ALL matching records (no pagination limit), trigger download | Button loading → file downloads; `message.success("CSV 已导出")` |
| Click "导出 CSV" (empty) | Block | `message.warning("当前筛选条件下无数据可导出")` |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 类型 Tag | `item.type` | "mainItem" / "subItem" |
| 编号 | `item.code` | |
| 标题 | `item.title` | Link to detail page |
| 优先级 Tag | `item.priority` | |
| 负责人 | `item.assignee.name`, `item.assignee.avatar` | |
| 完成度 Progress | `item.completionRate` | |
| 状态 Tag | `item.status` | |
| 预期完成时间 | `item.expectedEndDate` | Red if overdue |
| Overdue row class | `item.expectedEndDate` vs today + `item.status` | Client-side |
| CSV export data | All filtered items (full dataset, no pagination) | API call with `limit=all` |


---

## Page 11: 超级管理员后台

### Layout Structure

```
AppLayout (SuperAdmin variant — sidebar shows extra "管理后台" entry)
├── SidebarNav
└── MainContent (padding 24px)
    ├── PageHeader
    │   └── Title "超级管理员后台"
    ├── Tabs (antd Tabs, type="line")
    │   ├── Tab "用户管理" (UserOutlined)
    │   └── Tab "团队列表" (TeamOutlined)
    │
    ├── [Tab: 用户管理]
    │   ├── FilterBar (flex row, gap 12px, marginBottom 16px)
    │   │   ├── Input.Search placeholder="搜索用户名/账号" (width 240px)
    │   │   └── Select placeholder="创建团队权限" (全部/有权限/无权限, width 160px)
    │   └── UserTable (antd Table, size="middle")
    │       Columns:
    │       ├── 姓名 (Avatar 24px + name)
    │       ├── 账号 (monospace)
    │       ├── 所属团队 (Tag list, max 2 shown + "+N more" Tooltip)
    │       ├── 创建团队权限 (Switch, checked=hasCreateTeamPermission)
    │       └── 操作 (Button "查看详情" link style)
    │
    └── [Tab: 团队列表]
        ├── FilterBar
        │   └── Input.Search placeholder="搜索团队名称" (width 240px)
        └── TeamTable (antd Table, size="middle")
            Columns:
            ├── 团队名称 (clickable link → team readonly view)
            ├── PM (Avatar 20px + name)
            ├── 成员数 (number)
            ├── 主事项数 (number)
            ├── 创建时间 (date, sorter)
            └── 操作 (Button "查看详情" link style)
```

Team readonly view: clicking a team name opens a full-page view of that team's 事项视图 with a top banner "只读模式 — 正在查看团队：[team name]" (antd `Alert` type="info", banner=true). All create/edit/archive actions are hidden.

Key components: `Tabs`, `Table`, `Switch`, `Input.Search`, `Select`, `Alert`, `Tag`, `Avatar`, `Modal`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading (users tab) | Table `loading={true}` | On tab load / search |
| Loading (teams tab) | Table `loading={true}` | On tab load / search |
| User list | Table with Switch per row | Normal |
| Switch ON (has permission) | Switch checked (blue) | hasCreateTeamPermission = true |
| Switch OFF (no permission) | Switch unchecked | hasCreateTeamPermission = false |
| Self row (SuperAdmin) | Switch disabled, Tooltip "不能修改自身权限" | currentUser.isSuperAdmin |
| Permission change confirm | `Modal.confirm()` before toggling | Switch clicked |
| Team readonly view | Full page with info banner | Team name clicked |
| Empty search | Table `locale.emptyText` = "未找到匹配的用户/团队" | Search yields no results |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Switch toggle (grant) | `Modal.confirm()` "确认授予 [name] 创建团队权限？" | On confirm: PATCH permission, `message.success("已授权")`, Switch updates |
| Switch toggle (revoke) | `Modal.confirm()` "确认撤销 [name] 的创建团队权限？" | On confirm: PATCH permission, `message.success("已撤销")`, Switch updates |
| Switch on self row | Blocked (disabled) | Tooltip "不能修改自身权限" on hover |
| Type in user search | Debounced 300ms, re-fetch user list | Table updates |
| Type in team search | Debounced 300ms, re-fetch team list | Table updates |
| Click team name / "查看详情" | Navigate to readonly team view | Router push with `?readonly=true&teamId=xxx` |
| Click "查看详情" (user) | Open user detail Drawer (right, width=400px) | Drawer slides in |
| Switch tabs | Load respective tab data | Loading state briefly |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 姓名 + Avatar | `user.name`, `user.avatar` | |
| 账号 | `user.username` | |
| 所属团队 Tags | `user.teams[].name` | Max 2 visible, rest in Tooltip |
| 创建团队权限 Switch | `user.hasCreateTeamPermission` | |
| Self row detection | `user.id === currentUser.id` | Disable Switch |
| 团队名称 | `team.name` | |
| PM | `team.pm.name`, `team.pm.avatar` | |
| 成员数 | `team.memberCount` | |
| 主事项数 | `team.mainItemCount` | |
| 创建时间 | `team.createdAt` | |

### Access Control
- The entire 管理后台 route is protected: redirect to 403 page if `currentUser.isSuperAdmin !== true`
- The sidebar entry "管理后台" is conditionally rendered only when `currentUser.isSuperAdmin === true`
- Readonly team view hides all mutation buttons by checking `readonly` query param

