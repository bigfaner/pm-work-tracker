---
created: 2026-04-17
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: PM Work Tracker

## Design System

- **Component Library**: Ant Design (antd) v5
- **Framework**: React 18 + TypeScript
- **Color Tokens**:
  - Overdue: `#ff4d4f` (antd red-5)
  - P1 priority: `#fa8c16` (antd orange-6) badge
  - P2 priority: `#1677ff` (antd blue-6) badge
  - P3 priority: `#8c8c8c` (antd gray-7) badge
  - Completed: `#52c41a` (antd green-6)
  - Key item: `#ff4d4f` star icon
- **Status Tag Colors**:

| Status | Tag Color |
|--------|-----------|
| 待开始 | default (gray) |
| 进行中 | processing (blue) |
| 阻塞中 | error (red) |
| 挂起 | warning (orange) |
| 已延期 | warning (orange) |
| 待验收 | cyan |
| 已完成 | success (green) |
| 已关闭 | default (gray) |
| 待分配 (pool) | default |
| 已分配 (pool) | success |
| 已拒绝 (pool) | error |

- **Layout**: Left sidebar (240px) + main content area. Min-width 1280px. PC-first.
- **Navigation Sidebar**:
  - Logo / team name at top
  - Menu items: 事项视图 / 周视图 / 甘特图 / 表格视图 / 事项池 / 周报导出
  - Bottom: team switcher dropdown + user avatar + logout
- **Filter Bar Pattern**: Horizontal inline row above every list. Uses antd `Select` (multiple), not sidebar filters.
- **Feedback**: `antd message.success/error` for mutations; `antd Modal.confirm` for destructive actions.

---

## Page 1: 登录页 (Login)

### Layout Structure

Full-page centered card. No sidebar.

```
AppLayout (full viewport, bg: #f0f2f5)
└── Card (480px wide, centered vertically)
    ├── Logo + "PM 工作事项追踪"
    ├── Form
    │   ├── Form.Item: 账号 (Input)
    │   ├── Form.Item: 密码 (Input.Password)
    │   └── Button (primary, full-width): 登录
    └── error Alert (conditional)
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default | Empty form, button enabled | — |
| Submitting | Button shows `loading` spinner, disabled | Prevent double submit |
| Error | Red `Alert` below form: "账号或密码错误" | Form values preserved |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 账号或密码为空 | 登录按钮 disabled | Inline Form validation |
| 点击登录 | POST /auth/login | Button loading state |
| 登录成功 | Redirect to 事项视图 | — |
| 登录失败 | Show error Alert | Alert auto-dismisses on next submit |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 账号 Input | username | |
| 密码 Input | password | masked |
| Error Alert | API error message | 统一显示"账号或密码错误" |

---

## Page 2: 主事项列表（事项视图）

### Layout Structure

```
AppLayout
├── Sidebar (240px)
└── MainContent
    ├── PageHeader: "事项视图" + Button "新建主事项" (PM only)
    ├── FilterBar (horizontal)
    │   ├── Select (multiple): 优先级
    │   ├── Select (multiple): 状态
    │   └── Select: 负责人
    └── MainItemList
        └── MainItemRow (Collapse.Panel per item)
            ├── Header: [编号] [标题] [优先级Tag] [负责人Avatar] [进度Progress] [状态Tag] [预期完成时间] [操作Menu]
            └── SubItemTable (expanded, antd Table, compact)
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | Skeleton rows (3 items) | — |
| Empty | Empty illustration + "暂无事项，点击新建主事项" | Button visible to PM |
| Populated | Collapse list | Default: all collapsed |
| Overdue | 预期完成时间 text color `#ff4d4f` | When today > expectedEndDate and not completed |
| Key item | Star icon (red) before title | isKeyItem=true |
| Filter empty | "没有符合条件的事项" in list area | — |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击主事项行 | Expand/collapse SubItem table | Smooth antd Collapse animation |
| 点击标题 | Navigate to 主事项详情页 | — |
| 点击"新建主事项" | Open CreateMainItemDrawer | Drawer slides in from right |
| 操作菜单 → 编辑 | Open EditMainItemDrawer | — |
| 操作菜单 → 归档 | Modal.confirm → POST archive | Success: item disappears from list; Error: message.error |
| 筛选变更 | Re-fetch with filter params | List updates, URL query params sync |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 编号 | code | |
| 标题 | title | |
| 优先级 Tag | priority | Color per design system |
| 负责人 | assigneeName | Avatar with initials |
| 进度条 | completion | antd Progress percent |
| 状态 Tag | status | Color per design system |
| 预期完成时间 | expectedEndDate | Red if overdue |
| SubItem table | subItems[] | Compact table: 标题/负责人/完成度/状态/预期完成时间 |

---

## Page 3: 主事项详情页

### Layout Structure

```
AppLayout
└── MainContent
    ├── Breadcrumb: 事项视图 > [主事项标题]
    ├── DetailHeader
    │   ├── Left: 编号 + 标题 + 优先级Tag + 状态Tag
    │   ├── Center: Progress bar (completion%) + 数值
    │   └── Right: 操作按钮 (PM: 编辑 / 归档)
    ├── InfoGrid (2-col): 提出人 / 负责人 / 开始时间 / 预期完成时间 / 实际完成时间
    ├── AggregateSection (Collapse, default collapsed)
    │   ├── 成果摘要 (list of achievement texts)
    │   └── 卡点摘要 (list of blocker texts)
    └── SubItemSection
        ├── SectionHeader: "子事项" + Button "新增子事项" (PM only)
        ├── FilterBar: 优先级 / 状态 / 负责人
        └── SubItemTable (antd Table)
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | Skeleton for header + table | — |
| No sub-items | Table empty state: "暂无子事项，点击新增" | — |
| Overdue | expectedEndDate red | — |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击子事项标题 | Navigate to 子事项详情页 | — |
| 新增子事项 | Open CreateSubItemDrawer | — |
| 编辑主事项 | Open EditMainItemDrawer | — |
| 归档 | Modal.confirm → archive | Redirect back to list on success |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| Progress bar | completion | antd Progress |
| 成果摘要 | subItems[].progressRecords[].achievement | Flattened list |
| 卡点摘要 | subItems[].progressRecords[].blocker | Flattened list |
| SubItem table | subItems[] | 标题/负责人/完成度/状态/预期完成时间 |

---

## Page 4: 子事项详情页

### Layout Structure

```
AppLayout
└── MainContent
    ├── Breadcrumb: 事项视图 > [主事项] > [子事项标题]
    ├── DetailHeader: 标题 + 优先级Tag + 状态Tag + 操作(编辑/状态变更)
    ├── InfoGrid: 负责人 / 开始时间 / 预期完成时间 / 完成度
    ├── ProgressSection
    │   ├── SectionHeader: "进度记录" + Button "追加进度" (assignee/member)
    │   └── Timeline (antd Timeline, ascending)
    │       └── TimelineItem: [时间] [创建人] [完成度%] [成果] [卡点] [经验]
    └── AppendProgressDrawer (conditional)
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| No progress | Timeline empty: "暂无进度记录，点击追加" | — |
| Has records | Timeline list ascending | Latest at bottom |
| PM correction | Completion value shows edit icon on hover | PM only |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 追加进度 | Open AppendProgressDrawer | Drawer from right |
| 提交进度 | POST /progress | Drawer closes, timeline appends new item |
| PM 点击完成度数值 | Inline edit (InputNumber) | PATCH completion; value updates in place |
| 状态变更 | Dropdown select → PUT status | Tag updates immediately |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| Timeline items | progressRecords[] | Ordered by createdAt ASC |
| 完成度 | progressRecords[n].completion | PM can edit |
| 成果 | progressRecords[n].achievement | |
| 卡点 | progressRecords[n].blocker | |
| 经验 | progressRecords[n].lesson | |

### AppendProgressDrawer Fields

| Field | Component | Validation |
|-------|-----------|------------|
| 完成度 | InputNumber (0-100) | Required; ≥ last record value |
| 成果 | TextArea | Optional |
| 卡点 | TextArea | Optional |
| 经验 | TextArea | Optional |

---

## Page 5: 事项池

### Layout Structure

```
AppLayout
└── MainContent
    ├── PageHeader: "事项池" + Button "提交事项" (all members)
    ├── FilterBar: 状态 Select (待分配/已分配/已拒绝)
    └── ItemPoolTable (antd Table)
        Columns: 标题 / 提交人 / 提交时间 / 状态Tag / 操作
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | Table skeleton | — |
| Empty | "暂无待分配事项" | — |
| 待分配 row | 操作列: "分配" + "拒绝" buttons (PM only) | — |
| 已分配 row | 操作列: 显示挂载的主事项链接 | — |
| 已拒绝 row | Row text gray; 操作列: 显示拒绝原因 Tooltip | — |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 提交事项 | Open SubmitItemDrawer | — |
| 分配 | Open AssignItemModal | — |
| 拒绝 | Open RejectModal (reason input) | — |
| 筛选状态 | Re-fetch | — |

### AssignItemModal Fields

| Field | Component | Validation |
|-------|-----------|------------|
| 挂载主事项 | Select (searchable, required) | Required: "请选择挂载的主事项" |
| 负责人 | Select (team members) | Required |

### RejectModal Fields

| Field | Component | Validation |
|-------|-----------|------------|
| 拒绝原因 | TextArea | Required, max 200 |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 标题 | title | |
| 提交人 | submitterName | |
| 提交时间 | createdAt | Format: YYYY-MM-DD HH:mm |
| 状态 Tag | status | Color per design system |

---

## Page 6: 周视图

### Layout Structure

```
AppLayout
└── MainContent
    ├── PageHeader: "周视图"
    ├── WeekPicker (antd DatePicker.WeekPicker, max = current week)
    └── WeeklyContent
        └── MainItemGroup (per main item, Card)
            ├── GroupHeader: 主事项标题 + 完成度
            ├── SubSection: "本周新完成" (green header)
            ├── SubSection: "本周有进度更新" (blue header)
            └── SubSection: "上周完成本周无变化" (gray header)
                └── SubItemWeekRow
                    ├── Left (plan): 子事项标题 + 描述
                    └── Right (actual): Timeline of this-week progress records
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | Skeleton cards | — |
| Empty week | "本周暂无事项更新" centered | — |
| Has data | Grouped cards | — |
| Switching week | Skeleton overlay on content | — |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 切换周 | Re-fetch weekly data | Skeleton loading |
| 选择未来周 | Disabled in picker | antd DatePicker disabledDate |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| WeekPicker | weekStart | Monday of selected week |
| 主事项标题 | groups[].mainItem.title | |
| 子事项标题 (plan) | subItem.title + description | |
| 进度记录 (actual) | subItem.progressThisWeek[] | Records in week range |

---

## Page 7: 甘特图视图

### Layout Structure

```
AppLayout
└── MainContent
    ├── PageHeader: "甘特图"
    ├── FilterBar: 状态 Select
    └── GanttContainer (horizontally scrollable)
        ├── LeftPanel (fixed 280px): item titles list
        │   └── GanttRow: [展开icon] [优先级Tag] [标题]
        └── RightPanel (scrollable): timeline grid
            └── GanttBar: CSS positioned bar (startDate → expectedEndDate)
                - Normal: blue fill with completion% overlay
                - Overdue: red border
                - Completed: green fill
                - No dates: gray dashed bar + "未设置时间"
```

Implementation: Custom CSS Grid timeline. No heavy Gantt library dependency.

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Collapsed (default) | Only MainItem rows visible | — |
| Expanded | MainItem row + SubItem rows below it | SubItem rows indented 24px |
| Overdue bar | Red border on bar | today > expectedEndDate and not completed |
| No date | Gray dashed bar | startDate or expectedEndDate is null |
| Completed | Green fill | status = 已完成 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击主事项行 | Toggle expand/collapse SubItem rows | Smooth height animation |
| 点击标题 | Navigate to detail page | — |
| Hover bar | Tooltip: 标题 / 完成度 / 开始-结束时间 | antd Tooltip |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| Bar left position | startDate | CSS calc from timeline origin |
| Bar width | expectedEndDate - startDate | Min 1 day width |
| Bar fill % | completion | Overlay div width |
| SubItem rows | subItems[] | Shown when parent expanded |

---

## Page 8: 表格视图

### Layout Structure

```
AppLayout
└── MainContent
    ├── PageHeader: "表格视图" + Button "导出 CSV"
    ├── FilterBar
    │   ├── Select: 类型 (主事项/子事项)
    │   ├── Select (multiple): 优先级
    │   ├── Select: 负责人
    │   └── Select (multiple): 状态
    └── antd Table
        Columns: 编号 / 标题 / 类型 / 优先级 / 负责人 / 状态 / 完成度 / 预期完成时间 / 实际完成时间
        (sortable: 编号/优先级/完成度/预期完成时间/实际完成时间)
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | Table skeleton | — |
| Empty | "没有符合条件的事项" | — |
| Exporting | 导出按钮 loading | Disabled during export |
| Export empty | message.warning: "当前筛选条件下无数据可导出" | — |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击列标题 | Toggle sort asc/desc | antd Table built-in sorter |
| 筛选变更 | Re-fetch | — |
| 点击行 | Navigate to item detail | — |
| 导出 CSV | GET table/export with current filters | Browser file download |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 类型 Tag | type | "主事项" blue / "子事项" default |
| 优先级 Tag | priority | Color per design system |
| 完成度 | completion | Text "60%" |
| 预期完成时间 | expectedEndDate | Red if overdue |

---

## Page 9: 周报导出

### Layout Structure

```
AppLayout
└── MainContent
    ├── PageHeader: "周报导出"
    ├── ControlBar
    │   ├── WeekPicker (max = current week)
    │   └── Button "导出 Markdown" (primary)
    └── PreviewPanel (Card, scrollable)
        └── Markdown preview (read-only, rendered)
            ├── # 周报标题
            ├── ## 主事项 1
            │   ├── 完成度: 60%
            │   └── ### 子事项列表
            │       └── - 子事项标题: 成果 / 卡点
            └── ...
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default | WeekPicker + empty preview | — |
| Loading preview | Skeleton in preview panel | On week change |
| Preview loaded | Rendered Markdown | — |
| No data | message.warning: "所选周暂无数据" | Preview panel empty |
| Exporting | 导出按钮 loading (max 5s) | Disabled during export |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 切换周 | Fetch preview | Skeleton |
| 导出 Markdown | GET export → file download | Button loading → success message |

---

## Page 10: 团队管理页

### Layout Structure

```
AppLayout
└── MainContent
    ├── PageHeader: "团队设置"
    ├── MemberTable (antd Table)
    │   Columns: 姓名 / 角色Tag / 加入时间 / 操作Menu
    │   操作Menu (PM only): 移除成员 / 设为 PM
    ├── Button "邀请成员" (PM only)
    └── DangerZone (Card, red border)
        └── Button "解散团队" (PM only)
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Default | Member list | — |
| Invite modal | Modal with user search | — |
| Transfer PM confirm | Modal.confirm with warning | Irreversible action warning |
| Disband confirm | Modal with team name input | Must type exact team name |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 邀请成员 | Open InviteModal | — |
| 移除成员 | Modal.confirm → DELETE member | message.success |
| 设为 PM | Modal.confirm (warning: 不可撤销) → PUT /pm | message.success; page reloads role |
| 解散团队 | Modal with Input (team name) → DELETE team | Redirect to team list on success |

### InviteModal Fields

| Field | Component | Validation |
|-------|-----------|------------|
| 账号搜索 | Select with search (debounced API) | Required |
| 角色 | Select: member (only option) | Required |

---

## Page 11: 超级管理员后台

### Layout Structure

```
AppLayout (SuperAdmin only, separate nav entry)
└── MainContent
    ├── Tabs: "用户管理" | "团队列表"
    │
    ├── Tab: 用户管理
    │   └── antd Table
    │       Columns: 姓名 / 账号 / 创建团队权限(Switch) / 操作
    │
    └── Tab: 团队列表
        └── antd Table
            Columns: 团队名称 / PM / 成员数 / 主事项数 / 创建时间 / 操作(查看)
            点击"查看" → 只读事项视图 (same as Page 2 but read-only, no mutations)
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | Table skeleton | — |
| Permission toggle | Switch loading state | During API call |
| Confirm toggle | Modal.confirm before change | "确认修改权限？" |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 切换创建团队权限 | Modal.confirm → PUT can-create-team | Switch updates on success |
| 点击查看团队 | Navigate to read-only team item view | Banner: "只读模式" |

### Data Binding

| UI Element | Data Field | Notes |
|------------|-----------|-------|
| 创建团队权限 Switch | canCreateTeam | Disabled for self |
| 成员数 | memberCount | |
| 主事项数 | mainItemCount | |

---

## Shared Components

### CreateMainItemDrawer / EditMainItemDrawer

| Field | Component | Validation |
|-------|-----------|------------|
| 标题 | Input | Required, max 100 |
| 优先级 | Select | Required |
| 负责人 | Select (team members) | Optional |
| 开始时间 | DatePicker | Optional |
| 预期完成时间 | DatePicker | Optional; must be ≥ startDate |

### CreateSubItemDrawer

| Field | Component | Validation |
|-------|-----------|------------|
| 标题 | Input | Required, max 100 |
| 描述 | TextArea | Optional |
| 优先级 | Select | Required |
| 负责人 | Select (team members) | Required |
| 开始时间 | DatePicker | Optional |
| 预期完成时间 | DatePicker | Optional |

### TeamSwitcher (Sidebar bottom)

- antd `Select` dropdown listing user's teams
- On change: navigate to same view in new team context
- SuperAdmin sees all teams

### GlobalFilterBar Pattern

All filter bars use the same layout:
```
<Space wrap>
  <Select placeholder="优先级" mode="multiple" allowClear style={{width: 140}} />
  <Select placeholder="状态" mode="multiple" allowClear style={{width: 160}} />
  <Select placeholder="负责人" allowClear style={{width: 140}} />
</Space>
```
Filter state is synced to URL query params for shareability.
