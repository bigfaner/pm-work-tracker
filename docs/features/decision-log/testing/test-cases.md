---
feature: "decision-log"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
  - prd/prd-ui-functions.md
generated: "2026-05-04"
---

# Test Cases: decision-log

## Summary

| Type | Count |
|------|-------|
| UI   | 15   |
| API  | 13  |
| CLI  | 0  |
| **Total** | **28** |

---

## UI Test Cases

## TC-001: Publish decision from form
- **Source**: Story 1 / AC-1
- **Type**: UI
- **Target**: ui/decision-form
- **Test ID**: ui/decision-form/publish-decision-from-form
- **Pre-conditions**: User is logged in with main_item:update permission; a main item exists and is not in final state
- **Route**: /items/:mainItemId
- **Steps**:
  1. Navigate to main item detail page
  2. Click the "添加决策" button in the decision timeline section
  3. Select a category from the dropdown
  4. Enter decision content in the textarea
  5. Click the "发布" button
- **Expected**: Decision appears in the timeline as a published record visible to all team members; the item cannot be edited afterward
- **Priority**: P0

## TC-002: Save decision as draft
- **Source**: Story 2 / AC-1
- **Type**: UI
- **Target**: ui/decision-form
- **Test ID**: ui/decision-form/save-decision-as-draft
- **Pre-conditions**: User is logged in with main_item:update permission; a main item exists
- **Route**: /items/:mainItemId
- **Steps**:
  1. Navigate to main item detail page
  2. Click "添加决策" button
  3. Select a category
  4. Enter content
  5. Click "保存草稿" button
- **Expected**: Decision is saved as draft; timeline shows the item with a "草稿" badge; draft is only visible to the creator
- **Priority**: P0

## TC-003: Edit draft and save again
- **Source**: Story 2 / AC-2
- **Type**: UI
- **Target**: ui/decision-form
- **Test ID**: ui/decision-form/edit-draft-and-save
- **Pre-conditions**: User has a draft decision in the timeline
- **Route**: /items/:mainItemId
- **Steps**:
  1. Navigate to main item detail page
  2. Click the "编辑" button on a draft item
  3. Modify the content
  4. Click "保存草稿"
- **Expected**: Draft content is updated; item remains in draft status with "草稿" badge
- **Priority**: P0

## TC-004: Publish from draft via edit
- **Source**: Story 2 / AC-3
- **Type**: UI
- **Target**: ui/decision-form
- **Test ID**: ui/decision-form/publish-from-draft-via-edit
- **Pre-conditions**: User has a draft decision in the timeline
- **Route**: /items/:mainItemId
- **Steps**:
  1. Navigate to main item detail page
  2. Click "编辑" on the draft item
  3. Click "发布" in the dialog
- **Expected**: Decision changes to published status; visible to all team members; edit button no longer shown on the item
- **Priority**: P0

## TC-005: View published decisions in timeline
- **Source**: Story 3 / AC-1
- **Type**: UI
- **Target**: ui/decision-timeline
- **Test ID**: ui/decision-timeline/view-published-decisions
- **Pre-conditions**: Multiple published decisions exist for the main item; user is logged in
- **Route**: /items/:mainItemId
- **Steps**:
  1. Navigate to main item detail page
  2. Scroll to the decision timeline section
- **Expected**: All published decisions are displayed in reverse chronological order; each item shows category badge, free tags, content summary (first 80 chars + "..."), creator name, and time
- **Priority**: P0

## TC-006: Expand decision content beyond 80 chars
- **Source**: Story 3 / AC-2
- **Type**: UI
- **Target**: ui/decision-timeline
- **Test ID**: ui/decision-timeline/expand-decision-content
- **Pre-conditions**: A published decision exists with content longer than 80 characters
- **Route**: /items/:mainItemId
- **Steps**:
  1. Navigate to main item detail page
  2. Observe the decision item shows truncated content (80 chars + "...")
  3. Click on the content text
- **Expected**: Content expands to show full text; clicking again collapses back to 2-line preview
- **Priority**: P1

## TC-007: Draft not visible to other users
- **Source**: Story 4 / AC-1
- **Type**: UI
- **Target**: ui/decision-timeline
- **Test ID**: ui/decision-timeline/draft-not-visible-to-other-users
- **Pre-conditions**: User A has a draft decision; User B is a different team member
- **Route**: /items/:mainItemId
- **Steps**:
  1. Log in as User B
  2. Navigate to the same main item detail page
  3. Check the decision timeline
- **Expected**: User A's draft does not appear in User B's timeline view
- **Priority**: P0

## TC-008: Add decision button hidden without permission
- **Source**: Story 4 / AC-3
- **Type**: UI
- **Target**: ui/decision-timeline
- **Test ID**: ui/decision-timeline/add-button-hidden-without-permission
- **Pre-conditions**: User does not have main_item:update permission; published decisions exist
- **Route**: /items/:mainItemId
- **Steps**:
  1. Log in as a user without main_item:update permission
  2. Navigate to main item detail page
  3. Check the decision timeline section header
- **Expected**: "添加决策" button is not visible; published decisions are still viewable
- **Priority**: P0

## TC-009: Edit button only shown on own drafts
- **Source**: Spec Section 5.2 — 校验规则 #3, #4
- **Type**: UI
- **Target**: ui/decision-timeline
- **Test ID**: ui/decision-timeline/edit-button-only-on-own-drafts
- **Pre-conditions**: User A has a draft; User A and User B both have main_item:update permission
- **Route**: /items/:mainItemId
- **Steps**:
  1. Log in as User A; verify "编辑" button appears on own draft
  2. Log in as User B; view the same main item
  3. Check that User A's draft is not visible (per TC-007) and no edit button appears for others' drafts
- **Expected**: "编辑" button only appears on draft items belonging to the current user
- **Priority**: P1

## TC-010: Form validation — empty category
- **Source**: Spec Section 5.3 — 校验规则 #1; UI Function 2 — Validation Rules (Required: category)
- **Type**: UI
- **Target**: ui/decision-form
- **Test ID**: ui/decision-form/validation-empty-category
- **Pre-conditions**: User is on the decision form dialog
- **Route**: /items/:mainItemId
- **Steps**:
  1. Open the "添加决策" dialog
  2. Leave category unselected
  3. Enter content
  4. Click "保存草稿" or "发布"
- **Expected**: Error message "请选择分类" appears below the category field; form is not submitted
- **Priority**: P0

## TC-011: Form validation — empty content
- **Source**: Spec Section 5.3 — 校验规则 #2; UI Function 2 — Validation Rules (Required: content)
- **Type**: UI
- **Target**: ui/decision-form
- **Test ID**: ui/decision-form/validation-empty-content
- **Pre-conditions**: User is on the decision form dialog
- **Route**: /items/:mainItemId
- **Steps**:
  1. Open the "添加决策" dialog
  2. Select a category
  3. Leave content empty
  4. Click "保存草稿" or "发布"
- **Expected**: Error message "请输入决策内容" appears below the content field; form is not submitted
- **Priority**: P0

## TC-012: Form validation — content exceeds 2000 chars
- **Source**: Spec Section 5.3 — 校验规则 #3; UI Function 2 — Validation Rules (Max length: content)
- **Type**: UI
- **Target**: ui/decision-form
- **Test ID**: ui/decision-form/validation-content-max-length
- **Pre-conditions**: User is on the decision form dialog
- **Route**: /items/:mainItemId
- **Steps**:
  1. Open the "添加决策" dialog
  2. Select a category
  3. Enter content exceeding 2000 characters
- **Expected**: Error message "内容不能超过 2000 字符" appears below the content field; character counter shows count approaching/exceeding 2000
- **Priority**: P1

## TC-013: Form validation — tag exceeds 20 chars
- **Source**: Spec Section 5.3 — 校验规则 #4; UI Function 2 — Validation Rules (Max length: tags)
- **Type**: UI
- **Target**: ui/decision-form
- **Test ID**: ui/decision-form/validation-tag-max-length
- **Pre-conditions**: User is on the decision form dialog
- **Route**: /items/:mainItemId
- **Steps**:
  1. Open the "添加决策" dialog
  2. Type a tag longer than 20 characters in the tag input
  3. Press Enter or comma to add the tag
- **Expected**: Error message "标签不能超过 20 字符" appears below the tag field; the overlong tag is not added
- **Priority**: P1

## TC-014: Timeline loading and empty states
- **Source**: UI Function 1 — States (Loading, Empty, Error)
- **Type**: UI
- **Target**: ui/decision-timeline
- **Test ID**: ui/decision-timeline/loading-and-empty-states
- **Pre-conditions**: Two test scenarios: (a) main item has no decisions; (b) main item has decisions
- **Route**: /items/:mainItemId
- **Steps**:
  1. Navigate to a main item with no decisions
  2. Verify "暂无决策记录" text and "添加决策" button (if permission) are shown
  3. Navigate to a main item with decisions; verify skeleton appears briefly then data loads
- **Expected**: Empty state shows correct placeholder; loading state shows skeleton; populated state shows timeline items
- **Priority**: P1

## TC-015: Timeline infinite scroll pagination
- **Source**: Spec Section 5.1 — 翻页设置 (每页 20 条，滚动加载)
- **Type**: UI
- **Target**: ui/decision-timeline
- **Test ID**: ui/decision-timeline/infinite-scroll-pagination
- **Pre-conditions**: More than 20 decision records exist for the main item
- **Route**: /items/:mainItemId
- **Steps**:
  1. Navigate to main item detail page
  2. Scroll to the bottom of the decision timeline
  3. Verify the next page of 20 items loads automatically
  4. Continue scrolling until all items are loaded
- **Expected**: Each scroll-to-bottom loads the next 20 items; footer shows "已加载 N 条" when all items are loaded; spinner appears during loading
- **Priority**: P2

---

## API Test Cases

## TC-016: Create draft decision
- **Source**: Story 2 / AC-1; Spec Section 5.2 — 保存草稿
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/create-draft
- **Pre-conditions**: User is authenticated with main_item:update permission; a main item exists
- **Steps**:
  1. `POST /api/v1/teams/:teamId/main-items/:itemId/decision-logs` with body `{ category: "technical", content: "Decision text", status: "draft" }`
- **Expected**: Returns 200 with the created decision log (status: "draft", createdBy: current user); response includes bizKey
- **Priority**: P0

## TC-017: Create and publish decision in one step
- **Source**: Story 1 / AC-1
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/create-and-publish
- **Pre-conditions**: User is authenticated with main_item:update permission; a main item exists
- **Steps**:
  1. `POST /api/v1/teams/:teamId/main-items/:itemId/decision-logs` with body `{ category: "technical", content: "Decision text", status: "published" }`
- **Expected**: Returns 200 with the created decision log (status: "published"); decision is immutable after creation
- **Priority**: P0

## TC-018: List decisions returns published and own drafts only
- **Source**: Story 3 / AC-1; Story 4 / AC-1; Spec Section 5.1 — 显示范围
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/list-filtered-by-visibility
- **Pre-conditions**: User A has 1 published and 1 draft; User B has 1 published and 1 draft for the same main item
- **Steps**:
  1. As User B, call `GET /api/v1/teams/:teamId/main-items/:itemId/decision-logs`
- **Expected**: Response contains 2 published decisions (User A's + User B's) + User B's draft only; User A's draft is excluded
- **Priority**: P0

## TC-019: List decisions sorted by creation time descending
- **Source**: Spec Section 5.1 — 排序方式
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/list-sorted-desc
- **Pre-conditions**: Multiple decisions exist with different creation times
- **Steps**:
  1. `GET /api/v1/teams/:teamId/main-items/:itemId/decision-logs`
- **Expected**: Results are ordered by createTime descending (newest first)
- **Priority**: P1

## TC-020: List decisions paginated at 20 items
- **Source**: Spec Section 5.1 — 翻页设置
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/list-pagination
- **Pre-conditions**: More than 20 decision records exist for the main item
- **Steps**:
  1. `GET /api/v1/teams/:teamId/main-items/:itemId/decision-logs?page=1&pageSize=20`
  2. `GET /api/v1/teams/:teamId/main-items/:itemId/decision-logs?page=2&pageSize=20`
- **Expected**: First call returns 20 items; second call returns the next batch; total count is accurate
- **Priority**: P1

## TC-021: Update draft decision
- **Source**: Story 2 / AC-2; Spec Section 5.2 — 编辑
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/update-draft
- **Pre-conditions**: A draft decision exists belonging to the current user
- **Steps**:
  1. `PUT /api/v1/teams/:teamId/main-items/:itemId/decision-logs/:logId` with body `{ category: "resource", content: "Updated content" }`
- **Expected**: Returns 200 with updated draft; status remains "draft"; only the draft creator can update
- **Priority**: P0

## TC-022: Publish draft decision
- **Source**: Story 2 / AC-3; Spec Section 5.2 — 发布
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/publish-draft
- **Pre-conditions**: A draft decision exists belonging to the current user
- **Steps**:
  1. `PATCH /api/v1/teams/:teamId/main-items/:itemId/decision-logs/:logId/publish`
- **Expected**: Returns 200; decision status changes to "published"; subsequent edit attempts return 403
- **Priority**: P0

## TC-023: Edit published decision returns 403
- **Source**: Story 4 / AC-2
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/edit-published-returns-403
- **Pre-conditions**: A published decision exists
- **Steps**:
  1. `PUT /api/v1/teams/:teamId/main-items/:itemId/decision-logs/:logId` on a published decision
- **Expected**: Returns 403 Forbidden; published decision content is unchanged
- **Priority**: P0

## TC-024: Create decision requires main_item:update permission
- **Source**: Spec Section 5.2 — 权限控制; Story 4 / AC-3
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/create-requires-permission
- **Pre-conditions**: User is authenticated but lacks main_item:update permission
- **Steps**:
  1. `POST /api/v1/teams/:teamId/main-items/:itemId/decision-logs` with valid body
- **Expected**: Returns 403 Forbidden
- **Priority**: P0

## TC-025: List decisions requires authentication
- **Source**: Spec Section 5.1 — 数据来源 (API 查询)
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/list-requires-auth
- **Pre-conditions**: No authentication token provided
- **Steps**:
  1. `GET /api/v1/teams/:teamId/main-items/:itemId/decision-logs` without Authorization header
- **Expected**: Returns 401 Unauthorized
- **Priority**: P1

## TC-026: Create decision with tags
- **Source**: Spec Section 5.3 — 表单字段 (自由标签)
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/create-with-tags
- **Pre-conditions**: User is authenticated with main_item:update permission
- **Steps**:
  1. `POST /api/v1/teams/:teamId/main-items/:itemId/decision-logs` with body `{ category: "technical", tags: ["缓存策略", "性能优化"], content: "Decision text", status: "draft" }`
- **Expected**: Returns 200; response includes the tags array as provided
- **Priority**: P1

## TC-027: Create decision with all six categories
- **Source**: Spec Section 5.3 — 分类 (6 个预定义选项: technical/resource/requirement/schedule/risk/other)
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/create-all-categories
- **Pre-conditions**: User is authenticated with main_item:update permission
- **Steps**:
  1. Create 6 decisions, one per category: technical, resource, requirement, schedule, risk, other
- **Expected**: Each creation returns 200 with the correct category value; list endpoint returns all 6 with correct category labels
- **Priority**: P1

## TC-028: List response includes required fields
- **Source**: Spec Section 5.1 — 列表字段 (category, tags, content, createdBy, createTime, status)
- **Type**: API
- **Target**: api/decision-logs
- **Test ID**: api/decision-logs/list-response-fields
- **Pre-conditions**: At least one decision exists
- **Steps**:
  1. `GET /api/v1/teams/:teamId/main-items/:itemId/decision-logs`
- **Expected**: Each item in the response includes: bizKey, category, tags, content, createdBy/createdByName, createTime, status
- **Priority**: P1

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 1 / AC-1 | UI | ui/decision-form | P0 |
| TC-002 | Story 2 / AC-1 | UI | ui/decision-form | P0 |
| TC-003 | Story 2 / AC-2 | UI | ui/decision-form | P0 |
| TC-004 | Story 2 / AC-3 | UI | ui/decision-form | P0 |
| TC-005 | Story 3 / AC-1 | UI | ui/decision-timeline | P0 |
| TC-006 | Story 3 / AC-2 | UI | ui/decision-timeline | P1 |
| TC-007 | Story 4 / AC-1 | UI | ui/decision-timeline | P0 |
| TC-008 | Story 4 / AC-3 | UI | ui/decision-timeline | P0 |
| TC-009 | Spec 5.2 #3,#4 | UI | ui/decision-timeline | P1 |
| TC-010 | Spec 5.3 #1; UI Func 2 | UI | ui/decision-form | P0 |
| TC-011 | Spec 5.3 #2; UI Func 2 | UI | ui/decision-form | P0 |
| TC-012 | Spec 5.3 #3; UI Func 2 | UI | ui/decision-form | P1 |
| TC-013 | Spec 5.3 #4; UI Func 2 | UI | ui/decision-form | P1 |
| TC-014 | UI Func 1 States | UI | ui/decision-timeline | P1 |
| TC-015 | Spec 5.1 Pagination | UI | ui/decision-timeline | P2 |
| TC-016 | Story 2 / AC-1; Spec 5.2 | API | api/decision-logs | P0 |
| TC-017 | Story 1 / AC-1 | API | api/decision-logs | P0 |
| TC-018 | Story 3/4 AC; Spec 5.1 | API | api/decision-logs | P0 |
| TC-019 | Spec 5.1 Sort | API | api/decision-logs | P1 |
| TC-020 | Spec 5.1 Pagination | API | api/decision-logs | P1 |
| TC-021 | Story 2 / AC-2; Spec 5.2 | API | api/decision-logs | P0 |
| TC-022 | Story 2 / AC-3; Spec 5.2 | API | api/decision-logs | P0 |
| TC-023 | Story 4 / AC-2 | API | api/decision-logs | P0 |
| TC-024 | Spec 5.2 Perms; Story 4 AC-3 | API | api/decision-logs | P0 |
| TC-025 | Spec 5.1 Auth | API | api/decision-logs | P1 |
| TC-026 | Spec 5.3 Tags | API | api/decision-logs | P1 |
| TC-027 | Spec 5.3 Categories | API | api/decision-logs | P1 |
| TC-028 | Spec 5.1 Fields | API | api/decision-logs | P1 |

---

## Route Validation

| Route | Status | TC IDs | Matched Route |
|-------|--------|--------|---------------|
| GET /api/v1/teams/:teamId/main-items/:itemId/decision-logs | Matched | TC-018, TC-019, TC-020, TC-025, TC-028 | router.go:140 |
| POST /api/v1/teams/:teamId/main-items/:itemId/decision-logs | Matched | TC-016, TC-017, TC-024, TC-026, TC-027 | router.go:141 |
| PUT /api/v1/teams/:teamId/main-items/:itemId/decision-logs/:logId | Matched | TC-021, TC-023 | router.go:142 |
| PATCH /api/v1/teams/:teamId/main-items/:itemId/decision-logs/:logId/publish | Matched | TC-022 | router.go:143 |
