---
feature: "pm-work-tracker"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
  - prd/prd-ui-functions.md
  - ui/ui-design.md
generated: "2026-04-19"
---

# Test Cases: pm-work-tracker

## Summary

| Type | Count |
|------|-------|
| UI   | 52    |
| API  | 20    |
| CLI  | 0     |
| **Total** | **72** |

---

## UI Test Cases

### Login (UI Function 7)

## TC-001: Login button disabled when both fields empty
- **Source**: UI Function 7 — Validation Rules
- **Type**: UI
- **Pre-conditions**: User is on login page, both account and password fields are empty
- **Steps**:
  1. Navigate to login page
  2. Verify both account and password fields are empty
  3. Observe login button state
- **Expected**: Login button is disabled and cannot be clicked
- **Priority**: P0

## TC-002: Login button disabled when only one field filled
- **Source**: UI Function 7 — Validation Rules
- **Type**: UI
- **Pre-conditions**: User is on login page
- **Steps**:
  1. Fill in account field with valid text
  2. Leave password field empty
  3. Observe login button state
  4. Clear account, fill in password
  5. Observe login button state
- **Expected**: Login button remains disabled when only one field has value
- **Priority**: P0

## TC-003: Successful login redirects to item view
- **Source**: UI Function 7 — Interactions / Story 6 context
- **Type**: UI
- **Pre-conditions**: User has a valid account with team membership
- **Steps**:
  1. Fill in valid account and password
  2. Click login button
  3. Wait for response
- **Expected**: Page redirects to item view (事项视图), the default homepage
- **Priority**: P0

## TC-004: Failed login shows generic error, preserves account
- **Source**: UI Function 7 — Validation Rules / States
- **Type**: UI
- **Pre-conditions**: User is on login page
- **Steps**:
  1. Fill in account "testuser"
  2. Fill in incorrect password
  3. Click login button
  4. Observe error display
  5. Check account field value
  6. Check password field value
- **Expected**: Error alert shows "账号或密码错误" above button; account field retains "testuser"; password field is cleared; inputs are re-enabled
- **Priority**: P0

## TC-005: Login submitting state shows loading
- **Source**: UI Function 7 — States
- **Type**: UI
- **Pre-conditions**: User has filled both fields with valid data
- **Steps**:
  1. Fill in account and password
  2. Click login button
  3. Observe button and input states during request
- **Expected**: Button shows loading spinner and is disabled; inputs are disabled; no duplicate submissions possible
- **Priority**: P1

### Team Management (UI Function 8, Stories 7-8)

## TC-006: Create team and auto-assign creator as PM
- **Source**: Story 7 / AC-1
- **Type**: UI
- **Pre-conditions**: User has "create team" permission and is logged in
- **Steps**:
  1. Navigate to team creation page
  2. Fill in team name and description
  3. Click create button
- **Expected**: Team is created; user is automatically assigned as PM; success toast appears
- **Priority**: P0

## TC-007: Invite member via account search
- **Source**: Story 7 / AC-1, UI Function 8
- **Type**: UI
- **Pre-conditions**: User is PM of a team
- **Steps**:
  1. Open team management page
  2. Click "邀请成员" button
  3. Type account/name in search field
  4. Select user from dropdown
  5. Confirm invite
- **Expected**: Invite modal opens; search returns matching users; after confirmation, success toast appears and member list refreshes
- **Priority**: P0

## TC-008: Transfer PM role with confirmation
- **Source**: Story 8 / AC-1, UI Function 8
- **Type**: UI
- **Pre-conditions**: User is PM, team has other members
- **Steps**:
  1. Open team management page
  2. Click action menu on a member row
  3. Click "设为 PM"
  4. Observe confirmation dialog
  5. Click confirm
- **Expected**: Confirmation dialog shows warning that operation is irreversible; after confirm, new PM gets management permissions; old PM demoted to regular member; success toast appears
- **Priority**: P0

## TC-009: After PM transfer, new PM can perform PM operations
- **Source**: Story 8 / AC-1
- **Type**: UI
- **Pre-conditions**: PM role has been transferred
- **Steps**:
  1. Log in as new PM
  2. Navigate to team management
  3. Attempt to create a main item
  4. Attempt to edit a sub-item
- **Expected**: New PM can perform all PM operations immediately; old PM cannot perform PM operations
- **Priority**: P0

## TC-010: Dissolve team requires exact team name input
- **Source**: UI Function 8 — Validation Rules
- **Type**: UI
- **Pre-conditions**: User is PM of a team
- **Steps**:
  1. Open team management page
  2. Click "解散团队" button
  3. Observe confirmation modal
  4. Type partial team name
  5. Observe confirm button state
  6. Type exact team name
  7. Click confirm
- **Expected**: Confirm button disabled until exact team name is typed; after confirm with correct name, team is dissolved and user redirected to team selection
- **Priority**: P1

## TC-011: Cannot remove self from team
- **Source**: UI Function 8 — Validation Rules
- **Type**: UI
- **Pre-conditions**: User is PM viewing team management page
- **Steps**:
  1. Open team management page
  2. Locate own row in member table
  3. Observe action column for own row
- **Expected**: Own row has no action dropdown/menu; cannot remove self
- **Priority**: P1

### Super Admin (UI Function 11, Story 6)

## TC-012: Grant create team permission with confirmation
- **Source**: Story 6 / AC-1, UI Function 11
- **Type**: UI
- **Pre-conditions**: Super admin is logged in
- **Steps**:
  1. Navigate to admin backend
  2. Find a user without create team permission
  3. Toggle Switch to ON
  4. Observe confirmation dialog
  5. Click confirm
- **Expected**: Confirmation dialog appears; after confirm, Switch toggles ON; success toast "已授权"
- **Priority**: P0

## TC-013: Revoke create team permission with confirmation
- **Source**: Story 6 / AC-1, UI Function 11
- **Type**: UI
- **Pre-conditions**: Super admin is logged in, target user has create team permission
- **Steps**:
  1. Navigate to admin backend
  2. Find user with create team permission
  3. Toggle Switch to OFF
  4. Confirm in dialog
- **Expected**: Confirmation dialog appears; after confirm, Switch toggles OFF; success toast "已撤销"
- **Priority**: P0

## TC-014: Cannot modify own super admin permissions
- **Source**: UI Function 11 — Validation Rules
- **Type**: UI
- **Pre-conditions**: Super admin is logged in
- **Steps**:
  1. Navigate to admin backend user management
  2. Locate own row in user table
  3. Observe Switch state for own row
  4. Hover over Switch
- **Expected**: Switch is disabled; tooltip shows "不能修改自身权限"
- **Priority**: P1

## TC-015: Super admin views all teams globally
- **Source**: Story 6 / AC-1, UI Function 11
- **Type**: UI
- **Pre-conditions**: Super admin is logged in, multiple teams exist
- **Steps**:
  1. Navigate to admin backend
  2. Switch to "团队列表" tab
  3. Observe team list
- **Expected**: All teams are visible with name, PM, member count, main item count, creation time
- **Priority**: P0

## TC-016: Super admin readonly team view
- **Source**: UI Function 11
- **Type**: UI
- **Pre-conditions**: Super admin is logged in
- **Steps**:
  1. In admin team list, click a team name
  2. Observe team detail view
- **Expected**: Page shows info banner "只读模式"; all create/edit/archive buttons are hidden; only viewing is allowed
- **Priority**: P1

## TC-017: Admin entry visible only to super admin
- **Source**: UI Function 11 — Access Control
- **Type**: UI
- **Pre-conditions**: Two users: one super admin, one regular user
- **Steps**:
  1. Log in as super admin, check sidebar navigation
  2. Log in as regular user, check sidebar navigation
- **Expected**: Super admin sees "管理后台" entry in sidebar; regular user does not see it
- **Priority**: P1

### Main Item List / Item View (UI Function 1, Spec 5.3)

## TC-018: Display main items with all required fields
- **Source**: UI Function 1 — Data Requirements, Spec 5.3 列表字段
- **Type**: UI
- **Pre-conditions**: Team has multiple active main items
- **Steps**:
  1. Navigate to item view (default page after login)
  2. Observe main item rows
- **Expected**: Each row displays: code (Tag), title, priority (colored Tag), assignee (Avatar+name), completion (Progress bar), status (colored Tag), expected end date
- **Priority**: P0

## TC-019: Expand/collapse main item to show sub-items
- **Source**: UI Function 1 — Interactions
- **Type**: UI
- **Pre-conditions**: Main item has sub-items
- **Steps**:
  1. Click main item panel header
  2. Observe sub-item list
  3. Click header again
- **Expected**: Sub-items expand below main item with smooth animation; each sub-item shows title, assignee, completion, status, expected end date; clicking again collapses
- **Priority**: P0

## TC-020: Create main item via modal form
- **Source**: UI Function 1 — Interactions
- **Type**: UI
- **Pre-conditions**: User is PM
- **Steps**:
  1. Click "新建主事项" button
  2. Observe modal with form
  3. Fill in title, priority, assignee, expected end date
  4. Click confirm
- **Expected**: Modal opens with form; after submit, success toast "创建成功"; modal closes; list refreshes with new item
- **Priority**: P0

## TC-021: Main item title validation — required, max 100 chars
- **Source**: Spec 5.3 表单校验规则, UI Function 1 — Validation Rules
- **Type**: UI
- **Pre-conditions**: PM opens create/edit main item modal
- **Steps**:
  1. Open create modal
  2. Leave title empty, click submit — observe error
  3. Fill title with 101 characters, click submit — observe error
  4. Fill title with 100 characters, click submit
- **Expected**: Empty title shows "请填写标题（最多 100 字）"; 101 chars shows validation error; 100 chars is accepted
- **Priority**: P0

## TC-022: Main item priority validation — required
- **Source**: Spec 5.3 表单校验规则, UI Function 1 — Validation Rules
- **Type**: UI
- **Pre-conditions**: PM opens create main item modal
- **Steps**:
  1. Open create modal
  2. Fill title but leave priority unselected
  3. Click submit
- **Expected**: Validation error "请选择优先级" shown
- **Priority**: P0

## TC-023: Expected end date not earlier than start date
- **Source**: Spec 5.3 表单校验规则
- **Type**: UI
- **Pre-conditions**: PM opens create/edit main item modal
- **Steps**:
  1. Set start date to 2026-04-20
  2. Set expected end date to 2026-04-15
  3. Observe validation
- **Expected**: Validation error "预期完成时间不能早于开始时间" shown on blur or submit
- **Priority**: P1

## TC-024: Edit main item via modal
- **Source**: UI Function 1 — Interactions
- **Type**: UI
- **Pre-conditions**: PM has an existing main item
- **Steps**:
  1. Click action dropdown on main item row
  2. Click "编辑"
  3. Observe modal pre-filled with item data
  4. Modify title
  5. Click confirm
- **Expected**: Modal opens pre-filled; after submit, success toast; list refreshes with updated data
- **Priority**: P0

## TC-025: Archive main item only when completed or closed
- **Source**: UI Function 1 — Validation Rules, Spec 5.3 功能说明
- **Type**: UI
- **Pre-conditions**: Main item exists in non-terminal status
- **Steps**:
  1. Click action dropdown on a main item with status "进行中"
  2. Click "归档"
  3. Observe feedback
  4. Repeat with a main item with status "已完成"
  5. Confirm archive in dialog
- **Expected**: Non-terminal status shows warning "请先完成或关闭事项再归档"; completed/closed status shows confirmation dialog; after confirm, item removed from active list
- **Priority**: P0

## TC-026: Overdue items highlighted in red
- **Source**: Spec 5.3 功能说明 — 超期高亮, UI Function 1 — States
- **Type**: UI
- **Pre-conditions**: Main item expected end date is past, status is not completed/closed
- **Steps**:
  1. Navigate to item view
  2. Locate an overdue main item
  3. Observe styling
- **Expected**: Expected end date text is #ff4d4f; row has red-tinted left border (3px solid #ff4d4f); tooltip shows "已超期 N 天"
- **Priority**: P1

## TC-027: P1 items show orange priority badge
- **Source**: UI Function 1 — States
- **Type**: UI
- **Pre-conditions**: Main item with P1 priority exists
- **Steps**:
  1. Navigate to item view
  2. Locate P1 priority main item
- **Expected**: Priority Tag shows orange background; Badge dot (orange) appears on left of title
- **Priority**: P2

## TC-028: Empty state when no main items
- **Source**: UI Function 1 — States
- **Type**: UI
- **Pre-conditions**: Team has no main items
- **Steps**:
  1. Navigate to item view for a team with no items
- **Expected**: Centered empty illustration; text "暂无事项"; "新建主事项" button visible
- **Priority**: P1

## TC-029: Filter by priority, status, assignee
- **Source**: Spec 5.3 筛选
- **Type**: UI
- **Pre-conditions**: Team has main items with varied priorities, statuses, assignees
- **Steps**:
  1. Select P1 in priority filter
  2. Observe filtered results
  3. Add status filter "进行中"
  4. Observe combined filter
  5. Click "重置"
- **Expected**: List filters to matching items only; combined filters work correctly; reset clears all filters and restores full list
- **Priority**: P0

## TC-030: Default sort — priority desc, creation time desc
- **Source**: Spec 5.3 排序与分页
- **Type**: UI
- **Pre-conditions**: Team has main items with different priorities and creation times
- **Steps**:
  1. Navigate to item view
  2. Observe item order
- **Expected**: Items sorted by priority descending (P1 first, then P2, then P3); same priority sorted by creation time descending
- **Priority**: P1

### Main Item Detail (UI Function 9)

## TC-031: Display detail with info, progress, and sub-item list
- **Source**: UI Function 9 — Description
- **Type**: UI
- **Pre-conditions**: Main item exists with sub-items and progress records
- **Steps**:
  1. Click main item title to navigate to detail page
  2. Observe page sections
- **Expected**: Breadcrumb shows "事项视图 > [title]"; header shows code Tag, title, priority Tag, status Tag; info card shows assignee, dates; progress circle shows completion; sub-item table lists all sub-items
- **Priority**: P0

## TC-032: Completion circle shows weighted average of sub-items
- **Source**: Spec 5.3 功能说明 — 进度自动汇总
- **Type**: UI
- **Pre-conditions**: Main item has 3 sub-items with different completion rates
- **Steps**:
  1. Navigate to main item detail
  2. Observe progress circle value
- **Expected**: Progress circle shows weighted average of all sub-item completion rates
- **Priority**: P1

## TC-033: Achievements and blockers aggregation
- **Source**: UI Function 9 — Description
- **Type**: UI
- **Pre-conditions**: Sub-items have progress records with achievements and blockers
- **Steps**:
  1. Navigate to main item detail
  2. Observe "成果 & 卡点聚合" section
- **Expected**: Section shows bullet list of all achievements and bullet list of all blockers from all sub-item progress records
- **Priority**: P1

## TC-034: No sub-items empty state
- **Source**: UI Function 9 — States
- **Type**: UI
- **Pre-conditions**: Main item has no sub-items
- **Steps**:
  1. Navigate to main item detail page for item with no sub-items
- **Expected**: Sub-item table area shows empty state "暂无子事项，点击新增" with "新增子事项" button (PM only)
- **Priority**: P1

### Sub-item Management (Stories 2, 10, 11)

## TC-035: PM creates sub-item under main item
- **Source**: Story 2 / AC-1
- **Type**: UI
- **Pre-conditions**: PM has created a main item
- **Steps**:
  1. Navigate to main item detail
  2. Click "新增子事项"
  3. Fill in title, description, priority, assignee, expected completion time
  4. Click confirm
- **Expected**: Sub-item appears under main item; assignee can see it in "my items"; status is "待开始"
- **Priority**: P0

## TC-036: Member adds sub-item under main item
- **Source**: Story 10 / AC-1
- **Type**: UI
- **Pre-conditions**: Regular member is logged in, team has active main items
- **Steps**:
  1. Navigate to main item detail
  2. Click "添加子事项"
  3. Fill in title, description, expected completion time
  4. Click submit
- **Expected**: Sub-item appears under main item with status "待开始"; PM can see it; PM can adjust assignee or priority
- **Priority**: P0

## TC-037: PM adjusts sub-item assignee or priority
- **Source**: Story 10 / AC-1
- **Type**: UI
- **Pre-conditions**: Member has added a sub-item
- **Steps**:
  1. PM navigates to sub-item
  2. Edit assignee and priority fields
  3. Save changes
- **Expected**: Changes saved successfully; new assignee can see the sub-item
- **Priority**: P1

## TC-038: Assignee marks sub-item as "阻塞中"
- **Source**: Story 11 / AC-1
- **Type**: UI
- **Pre-conditions**: Assignee's sub-item status is "进行中"
- **Steps**:
  1. Navigate to sub-item
  2. Change status to "阻塞中"
  3. Add progress record with blocker description
- **Expected**: Sub-item status updates to "阻塞中"; PM can see status change in item view
- **Priority**: P0

## TC-039: Assignee resolves block and restores to "进行中"
- **Source**: Story 11 / AC-1
- **Type**: UI
- **Pre-conditions**: Sub-item status is "阻塞中"
- **Steps**:
  1. Navigate to sub-item
  2. Change status back to "进行中"
- **Expected**: Status updates to "进行中"; PM sees updated status
- **Priority**: P0

### Progress Records (UI Function 2, Story 3)

## TC-040: Assignee adds progress record
- **Source**: Story 3 / AC-1, UI Function 2
- **Type**: UI
- **Pre-conditions**: User is assignee of a sub-item with status "进行中"
- **Steps**:
  1. Navigate to sub-item detail
  2. Click "追加进度" button
  3. Fill in completion (60%), achievements, blockers
  4. Click submit
- **Expected**: New record appended to progress timeline at the end; success toast "进度已记录"
- **Priority**: P0

## TC-041: Main item completion auto-updates after progress
- **Source**: Story 3 / AC-1
- **Type**: UI
- **Pre-conditions**: Sub-item assignee adds a progress record with 60% completion
- **Steps**:
  1. Add progress record with completion 60%
  2. Navigate to parent main item
  3. Check main item completion value
- **Expected**: Main item completion rate is automatically recalculated based on updated sub-item completion
- **Priority**: P0

## TC-042: Progress records cannot be deleted
- **Source**: Story 3 / AC-1, Spec 5.4
- **Type**: UI
- **Pre-conditions**: Sub-item has progress records
- **Steps**:
  1. Navigate to sub-item detail with existing progress records
  2. Look for delete option on any record
- **Expected**: No delete button or option exists on any progress record; history is append-only
- **Priority**: P0

## TC-043: Completion cannot decrease below previous record
- **Source**: UI Function 2 — Validation Rules
- **Type**: UI
- **Pre-conditions**: Sub-item has a progress record with 60% completion
- **Steps**:
  1. Click "追加进度"
  2. Enter completion value of 50% (lower than 60%)
  3. Click submit
- **Expected**: Submit is blocked; inline error shows "完成度不能低于上一条记录（当前最高：60%）"
- **Priority**: P0

## TC-044: PM can edit completion inline
- **Source**: UI Function 2 — Interactions, Spec 5.4
- **Type**: UI
- **Pre-conditions**: PM views sub-item progress records
- **Steps**:
  1. Navigate to sub-item detail as PM
  2. Click on a progress record's completion percentage
  3. Enter new value (e.g., 75)
  4. Click confirm icon
- **Expected**: Inline InputNumber replaces percentage text; after confirm, value updates; success toast "已修正"
- **Priority**: P1

## TC-045: Progress record empty state
- **Source**: UI Function 2 — States
- **Type**: UI
- **Pre-conditions**: Sub-item has no progress records
- **Steps**:
  1. Navigate to sub-item detail with no records
  2. Observe progress timeline area
- **Expected**: Empty state with text "暂无进度记录，点击「追加进度」开始记录"
- **Priority**: P1

### Item Pool (UI Function 3, Stories 1, 9)

## TC-046: Member submits item to pool
- **Source**: Story 1 / AC-1, UI Function 3
- **Type**: UI
- **Pre-conditions**: User is logged in and belongs to a team
- **Steps**:
  1. Navigate to item pool page
  2. Click "提交到事项池"
  3. Fill in title, background, expected output
  4. Click submit
- **Expected**: Item appears in pool with status "待分配"; PM can see the item; success toast
- **Priority**: P0

## TC-047: PM assigns pool item to main item
- **Source**: Story 9 / AC-1, UI Function 3
- **Type**: UI
- **Pre-conditions**: Pool has "待分配" items, main items exist
- **Steps**:
  1. PM clicks "分配" on a pool item
  2. Select a main item from dropdown
  3. Select an assignee
  4. Click confirm
- **Expected**: Pool item status changes to "已分配"; sub-item auto-created under selected main item; card updates to show assigned main item and assignee
- **Priority**: P0

## TC-048: PM rejects pool item with reason
- **Source**: Story 9 / AC-1, UI Function 3
- **Type**: UI
- **Pre-conditions**: Pool has "待分配" items
- **Steps**:
  1. PM clicks "拒绝" on a pool item
  2. Observe reject modal
  3. Type rejection reason
  4. Click confirm
- **Expected**: Pool item status changes to "已拒绝"; card updates to show rejection reason; success toast
- **Priority**: P0

## TC-049: Pool submit requires title (max 100)
- **Source**: UI Function 3 — Validation Rules
- **Type**: UI
- **Pre-conditions**: User opens submit modal
- **Steps**:
  1. Open submit modal
  2. Leave title empty, click submit
  3. Fill title with 101 characters
- **Expected**: Empty title blocks submit with validation error; 101 chars shows validation error
- **Priority**: P0

## TC-050: Pool assign requires main item selection
- **Source**: UI Function 3 — Validation Rules
- **Type**: UI
- **Pre-conditions**: PM opens assign modal
- **Steps**:
  1. Open assign modal
  2. Do not select a main item
  3. Click confirm
- **Expected**: Submit blocked; inline error "请选择挂载的主事项"
- **Priority**: P0

## TC-051: Pool reject requires reason
- **Source**: UI Function 3 — Validation Rules
- **Type**: UI
- **Pre-conditions**: PM opens reject modal
- **Steps**:
  1. Open reject modal
  2. Leave reason empty
  3. Click confirm
- **Expected**: Submit blocked; inline error "请填写拒绝原因"
- **Priority**: P0

## TC-052: Pool filter by status
- **Source**: UI Function 3 — Data Requirements
- **Type**: UI
- **Pre-conditions**: Pool has items in different statuses
- **Steps**:
  1. Click "待分配" radio button
  2. Observe list
  3. Click "已分配"
  4. Observe list
  5. Click "全部"
- **Expected**: Each filter shows only matching items; "全部" shows all items
- **Priority**: P1

### Week View (UI Function 4, Story 4)

## TC-053: Week view shows plan vs actual grouped by main items
- **Source**: Story 4 / AC-1, UI Function 4
- **Type**: UI
- **Pre-conditions**: Current week has sub-items with assignments and progress updates
- **Steps**:
  1. Switch to week view
  2. Verify default is current week
  3. Observe grouped sections
- **Expected**: Sub-items grouped by main item; each shows title (plan) in left column and this week's progress records (actual) in right column; two-column layout with "计划" and "实际" headers
- **Priority**: P0

## TC-054: Week view supports this-week vs last-week comparison
- **Source**: Story 4 / AC-1, UI Function 4
- **Type**: UI
- **Pre-conditions**: Sub-items have progress in both this week and last week
- **Steps**:
  1. View current week
  2. Observe section grouping
- **Expected**: Sections show "本周新完成", "本周有进度更新", "上周完成，本周无变化" (collapsed); main item level comparison visible
- **Priority**: P0

## TC-055: Week view cannot select future weeks
- **Source**: UI Function 4 — Validation Rules
- **Type**: UI
- **Pre-conditions**: User is on week view
- **Steps**:
  1. Open week picker
  2. Attempt to select a future week
- **Expected**: Future weeks are disabled and cannot be selected; current week is the maximum selectable value
- **Priority**: P1

## TC-056: Week view empty state
- **Source**: UI Function 4 — States
- **Type**: UI
- **Pre-conditions**: Selected week has no data
- **Steps**:
  1. Navigate to week view
  2. Select a historical week with no item updates
- **Expected**: Empty illustration with text "本周暂无事项更新"
- **Priority**: P1

### Gantt View (UI Function 5, Story 12)

## TC-057: Gantt shows main items as timeline rows
- **Source**: Story 12 / AC-1, UI Function 5
- **Type**: UI
- **Pre-conditions**: Team has multiple main items with start and expected end dates
- **Steps**:
  1. Switch to Gantt view
  2. Observe timeline
- **Expected**: Each main item occupies one row with a time bar; bar spans from start date to expected end date; bar color reflects status
- **Priority**: P0

## TC-058: Click main item expands/collapses sub-items
- **Source**: Story 12 / AC-1, UI Function 5
- **Type**: UI
- **Pre-conditions**: Main item has sub-items
- **Steps**:
  1. Click main item label row
  2. Observe sub-items appearing
  3. Click main item label row again
- **Expected**: Sub-item rows appear below main item, each occupying its own row; rows fade in/out; clicking again collapses; toggle icon rotates
- **Priority**: P0

## TC-059: Overdue items marked in red in Gantt
- **Source**: Story 12 / AC-1, UI Function 5
- **Type**: UI
- **Pre-conditions**: Main item expected end date is past, not completed
- **Steps**:
  1. Switch to Gantt view
  2. Locate overdue main item bar
- **Expected**: Time bar is red (#ff4d4f); bar extends past today line
- **Priority**: P1

## TC-060: No time data items show gray dashed bar
- **Source**: UI Function 5 — States
- **Type**: UI
- **Pre-conditions**: Main item has no start date or expected end date
- **Steps**:
  1. Switch to Gantt view
  2. Locate main item without dates
- **Expected**: Gray dashed outline bar; label "未设置时间"; still visible in Gantt
- **Priority**: P1

### Table View (UI Function 10, Spec 5.6)

## TC-061: Table view with sortable columns
- **Source**: UI Function 10, Spec 5.6 表格视图字段
- **Type**: UI
- **Pre-conditions**: Team has mixed main items and sub-items
- **Steps**:
  1. Switch to table view
  2. Click "优先级" column header to sort
  3. Click again to reverse sort
  4. Click "完成度" column header
- **Expected**: Sort toggles between ascending/descending; column headers show sort arrows; table reorders accordingly
- **Priority**: P0

## TC-062: Table view filter by type, priority, assignee, status
- **Source**: UI Function 10, Spec 5.6
- **Type**: UI
- **Pre-conditions**: Team has varied items
- **Steps**:
  1. Select "子事项" in type filter
  2. Observe results
  3. Add priority filter P1
  4. Click "重置"
- **Expected**: Filtered results show only matching items; combined filters work; reset clears all
- **Priority**: P0

## TC-063: Table view CSV export
- **Source**: UI Function 10 — Interactions, Spec 5.6
- **Type**: UI
- **Pre-conditions**: Table has filtered results with data
- **Steps**:
  1. Apply filters
  2. Click "导出 CSV" button
  3. Wait for download
- **Expected**: Button shows loading briefly; CSV file downloads; exported data matches current filter results (not limited by pagination); success toast "CSV 已导出"
- **Priority**: P0

## TC-064: CSV export empty warning
- **Source**: UI Function 10 — Validation Rules
- **Type**: UI
- **Pre-conditions**: Filters produce no results
- **Steps**:
  1. Apply filters that yield zero results
  2. Click "导出 CSV"
- **Expected**: Action blocked; warning toast "当前筛选条件下无数据可导出"
- **Priority**: P1

## TC-065: Table view default sort and pagination
- **Source**: Spec 5.6 表格视图规则
- **Type**: UI
- **Pre-conditions**: Team has more than 50 items
- **Steps**:
  1. Switch to table view
  2. Observe sort order
  3. Observe pagination
- **Expected**: Default sort is priority descending, then expected end date ascending; 50 items per page; pagination controls visible
- **Priority**: P1

### Report Export (UI Function 6, Story 5)

## TC-066: Preview report before export
- **Source**: UI Function 6 — Interactions
- **Type**: UI
- **Pre-conditions**: Current week has item progress data
- **Steps**:
  1. Navigate to report export page
  2. Select current week (default)
  3. Click "生成预览"
  4. Observe preview content
- **Expected**: Loading spinner briefly; then markdown-rendered preview showing: week title, main item sections with completion, sub-item achievements and blockers
- **Priority**: P0

## TC-067: Export Markdown within 5 seconds
- **Source**: Story 5 / AC-1, UI Function 6
- **Type**: UI
- **Pre-conditions**: Preview is ready with data
- **Steps**:
  1. Click "导出 Markdown" button
  2. Wait for download
- **Expected**: Button shows loading, disabled; file download triggers within 5 seconds; success toast "周报已导出"
- **Priority**: P0

## TC-068: Report export no data warning
- **Source**: UI Function 6 — States
- **Type**: UI
- **Pre-conditions**: Selected week has no item data
- **Steps**:
  1. Select a week with no data
  2. Click "生成预览"
  3. Observe result
- **Expected**: Empty state in preview card "所选周暂无数据，无法导出"; export button is blocked
- **Priority**: P1

---

## API Test Cases

### Authentication & Authorization

## TC-069: Login with valid credentials returns token
- **Source**: Spec 5.1 用户认证与权限
- **Type**: API
- **Pre-conditions**: User account exists in system
- **Steps**:
  1. POST login request with valid account and password
  2. Verify response
- **Expected**: Response 200 with authentication token; token is valid for subsequent requests
- **Priority**: P0

## TC-070: Login with invalid credentials returns 401
- **Source**: Spec 5.1, UI Function 7
- **Type**: API
- **Pre-conditions**: User account exists
- **Steps**:
  1. POST login request with incorrect password
  2. Verify response
- **Expected**: Response 401; generic error message "账号或密码错误"; no token returned
- **Priority**: P0

## TC-071: Unauthenticated access blocked
- **Source**: Spec 安全性需求
- **Type**: API
- **Pre-conditions**: No valid token in request
- **Steps**:
  1. Send request to any protected endpoint without token
  2. Verify response
- **Expected**: Response 401; redirects to login page for UI routes
- **Priority**: P0

### Team Data Isolation

## TC-072: Cross-team data access blocked
- **Source**: Spec 安全性需求 — 多团队隔离
- **Type**: API
- **Pre-conditions**: User belongs to Team A only; Team B has data
- **Steps**:
  1. Request Team B's main items using Team A user's token
  2. Verify response
- **Expected**: Response 403; Team B data is not accessible to Team A members
- **Priority**: P0

## TC-073: Super admin bypasses team isolation
- **Source**: Spec 5.2 团队管理 — 数据权限
- **Type**: API
- **Pre-conditions**: Super admin does not belong to Team A
- **Steps**:
  1. Request Team A's data using super admin token
  2. Verify response
- **Expected**: Response 200; super admin can access all team data
- **Priority**: P0

### Main Item Operations

## TC-074: Create main item with validation
- **Source**: Spec 5.3 表单校验规则
- **Type**: API
- **Pre-conditions**: PM is authenticated
- **Steps**:
  1. POST main item with empty title — expect error
  2. POST main item with title > 100 chars — expect error
  3. POST main item without priority — expect error
  4. POST main item with expected end date before start date — expect error
  5. POST valid main item
- **Expected**: Invalid requests return 400 with validation errors; valid request returns 201 with created item including auto-generated code
- **Priority**: P0

## TC-075: Archive main item only when completed or closed
- **Source**: Spec 5.3 功能说明 — 归档主事项
- **Type**: API
- **Pre-conditions**: Main items exist in various statuses
- **Steps**:
  1. PATCH archive on main item with status "进行中" — expect error
  2. PATCH archive on main item with status "已完成" — expect success
- **Expected**: Non-terminal status returns 400; completed/closed status returns 200 and item is archived
- **Priority**: P0

### Sub-item Status Transitions

## TC-076: Valid status transitions succeed
- **Source**: Spec 事项状态流转
- **Type**: API
- **Pre-conditions**: Sub-item exists with status "待开始"
- **Steps**:
  1. Change status to "进行中" — expect success
  2. Change status to "待验收" — expect success
  3. Change status to "已完成" — expect success
- **Expected**: Each valid transition returns 200; status updates correctly
- **Priority**: P0

## TC-077: Invalid status transitions rejected
- **Source**: Spec 事项状态流转
- **Type**: API
- **Pre-conditions**: Sub-item exists with status "待开始"
- **Steps**:
  1. Attempt to change status directly to "已完成" (skip 进行中, 待验收) — expect error
  2. Attempt to change status to "阻塞中" (requires 进行中) — expect error
- **Expected**: Invalid transitions return 400 with error message
- **Priority**: P1

### Progress Record Operations

## TC-078: Add progress record with completion validation
- **Source**: Spec 5.4 进度记录字段, UI Function 2 — Validation Rules
- **Type**: API
- **Pre-conditions**: Sub-item status is "进行中", latest record has 60% completion
- **Steps**:
  1. POST progress record with completion 50% (lower than 60%) — expect error
  2. POST progress record with completion 150% — expect error
  3. POST progress record with completion 75% — expect success
- **Expected**: Decreasing completion returns 400; out of range returns 400; valid completion returns 201
- **Priority**: P0

## TC-079: Progress records cannot be deleted
- **Source**: Spec 5.4 — 进度记录为追加式
- **Type**: API
- **Pre-conditions**: Progress records exist
- **Steps**:
  1. Attempt DELETE on a progress record endpoint
  2. Verify response
- **Expected**: DELETE endpoint either does not exist (404) or returns 405; records are never deleted
- **Priority**: P0

## TC-080: Completion auto-recalculated after progress update
- **Source**: Spec 5.3 功能说明 — 进度自动汇总
- **Type**: API
- **Pre-conditions**: Main item has 2 sub-items with 50% and 80% completion
- **Steps**:
  1. Update first sub-item completion to 100%
  2. Fetch main item completion rate
- **Expected**: Main item completion rate is recalculated as weighted average (updated value reflects new sub-item data)
- **Priority**: P0

### Item Pool Operations

## TC-081: Submit item to pool
- **Source**: Spec 5.5 功能说明
- **Type**: API
- **Pre-conditions**: User is authenticated and belongs to a team
- **Steps**:
  1. POST item to pool with title, background, expected output
  2. Verify response
- **Expected**: Response 201; item has status "待分配"; submitter and timestamp auto-set
- **Priority**: P0

## TC-082: Assign pool item creates sub-item
- **Source**: Story 9 / AC-1, Spec 5.5
- **Type**: API
- **Pre-conditions**: Pool item exists with status "待分配"; main items exist
- **Steps**:
  1. PATCH pool item to assign with main item ID and assignee ID
  2. Verify pool item status
  3. Verify new sub-item created under main item
- **Expected**: Pool item status becomes "已分配"; new sub-item created under specified main item with given assignee and status "待开始"
- **Priority**: P0

## TC-083: Reject pool item requires reason
- **Source**: UI Function 3 — Validation Rules
- **Type**: API
- **Pre-conditions**: Pool item exists with status "待分配"
- **Steps**:
  1. PATCH pool item to reject without reason — expect error
  2. PATCH pool item to reject with reason — expect success
- **Expected**: Missing reason returns 400; with reason, status becomes "已拒绝"
- **Priority**: P0

### Role-Based Access Control

## TC-084: Member cannot create main item
- **Source**: Spec 5.3 功能说明 — 可操作角色
- **Type**: API
- **Pre-conditions**: Regular member (non-PM) is authenticated
- **Steps**:
  1. POST request to create main item
- **Expected**: Response 403; only PM can create main items
- **Priority**: P0

## TC-085: Member cannot archive main item
- **Source**: Spec 5.3 功能说明 — 可操作角色
- **Type**: API
- **Pre-conditions**: Regular member (non-PM) is authenticated
- **Steps**:
  1. PATCH request to archive main item
- **Expected**: Response 403; only PM can archive
- **Priority**: P1

## TC-086: User without permission cannot create team
- **Source**: Spec 5.2 — 创建团队
- **Type**: API
- **Pre-conditions**: User does not have "create team" permission
- **Steps**:
  1. POST request to create team
- **Expected**: Response 403; only users with granted permission can create teams
- **Priority**: P0

## TC-087: Non-super-admin cannot access admin endpoints
- **Source**: UI Function 11 — Access Control
- **Type**: API
- **Pre-conditions**: Regular user or PM is authenticated
- **Steps**:
  1. GET request to admin user list endpoint
  2. PATCH request to admin permission endpoint
- **Expected**: Response 403 for all admin endpoints
- **Priority**: P0

### Auto-Calculations

## TC-088: Key item auto-upgrade after 2+ delays
- **Source**: Spec 5.3 功能说明 — 重点事项标记
- **Type**: API
- **Pre-conditions**: Main item has been delayed twice (expected end date extended twice)
- **Steps**:
  1. Create main item with P2 priority
  2. Extend expected end date first time
  3. Extend expected end date second time
  4. Verify priority
- **Expected**: After second delay, priority automatically upgrades to P1; item marked as key item
- **Priority**: P1

### Performance

## TC-089: Report export responds within 5 seconds
- **Source**: Spec 性能需求, Story 5 / AC-1
- **Type**: API
- **Pre-conditions**: Team has progress data for the selected week
- **Steps**:
  1. Request report export for current week
  2. Measure response time
- **Expected**: Response completes within 5 seconds; Markdown content returned
- **Priority**: P1

## TC-090: List page loads within 2 seconds
- **Source**: Spec 性能需求
- **Type**: API
- **Pre-conditions**: Team has typical data volume (20-50 items)
- **Steps**:
  1. Request main item list endpoint
  2. Measure response time
- **Expected**: Response completes within 2 seconds
- **Priority**: P1

---

## CLI Test Cases

_No CLI test cases — this feature is a web application with no CLI interface._

---

## Traceability

| TC ID | Source | Type | Priority |
|-------|--------|------|----------|
| TC-001 | UI Function 7 — Validation Rules | UI | P0 |
| TC-002 | UI Function 7 — Validation Rules | UI | P0 |
| TC-003 | UI Function 7 — Interactions | UI | P0 |
| TC-004 | UI Function 7 — States / Validation Rules | UI | P0 |
| TC-005 | UI Function 7 — States | UI | P1 |
| TC-006 | Story 7 / AC-1 | UI | P0 |
| TC-007 | Story 7 / AC-1, UI Function 8 | UI | P0 |
| TC-008 | Story 8 / AC-1, UI Function 8 | UI | P0 |
| TC-009 | Story 8 / AC-1 | UI | P0 |
| TC-010 | UI Function 8 — Validation Rules | UI | P1 |
| TC-011 | UI Function 8 — Validation Rules | UI | P1 |
| TC-012 | Story 6 / AC-1, UI Function 11 | UI | P0 |
| TC-013 | Story 6 / AC-1, UI Function 11 | UI | P0 |
| TC-014 | UI Function 11 — Validation Rules | UI | P1 |
| TC-015 | Story 6 / AC-1, UI Function 11 | UI | P0 |
| TC-016 | UI Function 11 | UI | P1 |
| TC-017 | UI Function 11 — Access Control | UI | P1 |
| TC-018 | UI Function 1, Spec 5.3 | UI | P0 |
| TC-019 | UI Function 1 — Interactions | UI | P0 |
| TC-020 | UI Function 1 — Interactions | UI | P0 |
| TC-021 | Spec 5.3 表单校验, UI Function 1 | UI | P0 |
| TC-022 | Spec 5.3 表单校验, UI Function 1 | UI | P0 |
| TC-023 | Spec 5.3 表单校验 | UI | P1 |
| TC-024 | UI Function 1 — Interactions | UI | P0 |
| TC-025 | UI Function 1, Spec 5.3 | UI | P0 |
| TC-026 | Spec 5.3, UI Function 1 | UI | P1 |
| TC-027 | UI Function 1 — States | UI | P2 |
| TC-028 | UI Function 1 — States | UI | P1 |
| TC-029 | Spec 5.3 筛选 | UI | P0 |
| TC-030 | Spec 5.3 排序与分页 | UI | P1 |
| TC-031 | UI Function 9 | UI | P0 |
| TC-032 | Spec 5.3 进度自动汇总 | UI | P1 |
| TC-033 | UI Function 9 | UI | P1 |
| TC-034 | UI Function 9 — States | UI | P1 |
| TC-035 | Story 2 / AC-1 | UI | P0 |
| TC-036 | Story 10 / AC-1 | UI | P0 |
| TC-037 | Story 10 / AC-1 | UI | P1 |
| TC-038 | Story 11 / AC-1 | UI | P0 |
| TC-039 | Story 11 / AC-1 | UI | P0 |
| TC-040 | Story 3 / AC-1, UI Function 2 | UI | P0 |
| TC-041 | Story 3 / AC-1 | UI | P0 |
| TC-042 | Story 3 / AC-1, Spec 5.4 | UI | P0 |
| TC-043 | UI Function 2 — Validation Rules | UI | P0 |
| TC-044 | UI Function 2 — Interactions, Spec 5.4 | UI | P1 |
| TC-045 | UI Function 2 — States | UI | P1 |
| TC-046 | Story 1 / AC-1, UI Function 3 | UI | P0 |
| TC-047 | Story 9 / AC-1, UI Function 3 | UI | P0 |
| TC-048 | Story 9 / AC-1, UI Function 3 | UI | P0 |
| TC-049 | UI Function 3 — Validation Rules | UI | P0 |
| TC-050 | UI Function 3 — Validation Rules | UI | P0 |
| TC-051 | UI Function 3 — Validation Rules | UI | P0 |
| TC-052 | UI Function 3 | UI | P1 |
| TC-053 | Story 4 / AC-1, UI Function 4 | UI | P0 |
| TC-054 | Story 4 / AC-1, UI Function 4 | UI | P0 |
| TC-055 | UI Function 4 — Validation Rules | UI | P1 |
| TC-056 | UI Function 4 — States | UI | P1 |
| TC-057 | Story 12 / AC-1, UI Function 5 | UI | P0 |
| TC-058 | Story 12 / AC-1, UI Function 5 | UI | P0 |
| TC-059 | Story 12 / AC-1, UI Function 5 | UI | P1 |
| TC-060 | UI Function 5 — States | UI | P1 |
| TC-061 | UI Function 10, Spec 5.6 | UI | P0 |
| TC-062 | UI Function 10, Spec 5.6 | UI | P0 |
| TC-063 | UI Function 10, Spec 5.6 | UI | P0 |
| TC-064 | UI Function 10 — Validation Rules | UI | P1 |
| TC-065 | Spec 5.6 | UI | P1 |
| TC-066 | UI Function 6 — Interactions | UI | P0 |
| TC-067 | Story 5 / AC-1, UI Function 6 | UI | P0 |
| TC-068 | UI Function 6 — States | UI | P1 |
| TC-069 | Spec 5.1 | API | P0 |
| TC-070 | Spec 5.1, UI Function 7 | API | P0 |
| TC-071 | Spec 安全性需求 | API | P0 |
| TC-072 | Spec 安全性需求 | API | P0 |
| TC-073 | Spec 5.2 | API | P0 |
| TC-074 | Spec 5.3 表单校验 | API | P0 |
| TC-075 | Spec 5.3 | API | P0 |
| TC-076 | Spec 事项状态流转 | API | P0 |
| TC-077 | Spec 事项状态流转 | API | P1 |
| TC-078 | Spec 5.4, UI Function 2 | API | P0 |
| TC-079 | Spec 5.4 | API | P0 |
| TC-080 | Spec 5.3 进度自动汇总 | API | P0 |
| TC-081 | Spec 5.5 | API | P0 |
| TC-082 | Story 9 / AC-1, Spec 5.5 | API | P0 |
| TC-083 | UI Function 3 | API | P0 |
| TC-084 | Spec 5.3 | API | P0 |
| TC-085 | Spec 5.3 | API | P1 |
| TC-086 | Spec 5.2 | API | P0 |
| TC-087 | UI Function 11 | API | P0 |
| TC-088 | Spec 5.3 重点事项标记 | API | P1 |
| TC-089 | Spec 性能需求, Story 5 / AC-1 | API | P1 |
| TC-090 | Spec 性能需求 | API | P1 |
