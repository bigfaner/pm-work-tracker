---
created: 2026-04-27
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: 决策日志

## Design System

Based on the project's existing design system (DESIGN.md). Core tokens:

- **Primary**: `#2563eb` / hover `#1d4ed8`
- **Surface**: `#ffffff` on `#f8fafc` background
- **Border**: `#e2e8f0` / dark `#cbd5e1`
- **Text**: primary `#0f172a`, secondary `#475569`, tertiary `#94a3b8`
- **Warning**: `#d97706` bg `#fffbeb` text `#92400e`
- **Font**: 13px body, 14px labels, 12px badges
- **Radius**: rounded-lg (buttons), rounded-xl (cards/dialogs)
- **All colors via theme tokens**, never hardcoded Tailwind classes

---

## Component 1: DecisionTimeline

Embedded section within the main item detail page, displaying decision records in reverse chronological order.

### Layout Structure

```
┌─ Card (rounded-xl, border, bg-surface) ─────────────────────┐
│ Header (flex between, p-4 px-5, border-b)                   │
│   ├── "决策记录" (section heading, 14px font-medium)          │
│   └── [+ 添加决策] button (primary, sm size, hidden if       │
│       no main_item:update permission or main item is final)  │
│                                                              │
│ Content (p-5)                                                │
│   ┌─ Timeline Item ──────────────────────────────────────┐  │
│   │ flex gap-3                                           │  │
│   │  ┌─ Left Rail ─┐  ┌─ Right Content ──────────────┐  │  │
│   │  │  ● dot       │  │ Row 1: flex between          │  │  │
│   │  │  │ line      │  │   Category Badge + [草稿]    │  │  │
│   │  │              │  │   Badge (if draft)            │  │  │
│   │  │              │  │   Time (tertiary, 12px)       │  │  │
│   │  │              │  │                               │  │  │
│   │  │              │  │ Row 2: Content text           │  │  │
│   │  │              │  │   (13px, secondary, max 2     │  │  │
│   │  │              │  │    lines with line-clamp-2,   │  │  │
│   │  │              │  │    click to expand/collapse)  │  │  │
│   │  │              │  │                               │  │  │
│   │  │              │  │ Row 3: Tags + Creator + Edit  │  │  │
│   │  │              │  │   Tags: badge list (12px)     │  │  │
│   │  │              │  │   Creator: tertiary text      │  │  │
│   │  │              │  │   [编辑] ghost btn (draft only)│  │  │
│   │  └──────────────┘  └──────────────────────────────┘  │  │
│   └──────────────────────────────────────────────────────┘  │
│                                                              │
│   ── divider (border-b) between items ──                    │
│                                                              │
│   ── Sentinel for infinite scroll ──                         │
│                                                              │
│ Footer (p-3 px-5, border-t, bg-bg-alt)                      │
│   └── "已加载 N 条" (tertiary, 12px) or loading spinner      │
└──────────────────────────────────────────────────────────────┘
```

**Timeline rail**: 24px wide left column with a 6px dot and 1px vertical line connecting items. Line color: `border-border` (`#e2e8f0`). Dot color: `bg-primary` (`#2563eb`) for published, `bg-warning` (`#d97706`) for draft.

**Category badge colors** (mapped to existing badge system):

| Category | Label | Badge Variant |
|----------|-------|---------------|
| technical | 技术 | default (slate) |
| resource | 资源 | default (slate) |
| requirement | 需求 | default (slate) |
| schedule | 进度 | warning variant |
| risk | 风险 | error variant |
| other | 其他 | default (slate) |

**Draft badge**: inline-flex, rounded-full, px-2 py-0.5, text-xs, warning colors (`bg-warning-bg`, `text-warning-text`).

**Content truncation**: 2 lines with `line-clamp-2` (CSS `-webkit-line-clamp`). Expanded state removes clamp, shows full content. Cursor pointer, hover text-primary.

**Tag badges**: inline-flex, rounded-full, px-2 py-0.5, text-xs, bg-alt (`#f8fafc`), text-secondary, border `border-border`. Single tag max-width 120px with `overflow: hidden; text-overflow: ellipsis; white-space: nowrap`. If more than 3 tags, show first 3 + "+N" overflow badge (same style, text "+N").

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | Card header visible, content area shows 3 skeleton rows (h-16, rounded-lg, bg-bg-alt animate-pulse) | Appears on first load |
| Empty | Card header + "暂无决策记录" centered text (tertiary) + [+ 添加决策] button (if permission) | No timeline items |
| Populated | Timeline list with items as described above | Normal state |
| Loading More | Bottom spinner (h-6 w-6, border-2 border-t-transparent, animate-spin) below last item | Scrolling to bottom triggers next page load |
| Error | Red alert strip: "加载失败" text + [重试] ghost button, bg-error-bg, text-error-text | API failure on first page (full-state replacement) |
| Pagination Error | Inline error below last loaded item: "加载更多失败" text + [重试] ghost button (text-error-text, bg-error-bg), above footer | API failure on page 2+ (items already visible above) |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Scroll to bottom sentinel | Fetch next page (20 items) | Bottom spinner appears until data loads |
| Click content text (collapsed) | Expand to full content | Smooth height transition (200ms) |
| Click content text (expanded) | Collapse to 2-line preview | Smooth height transition (200ms) |
| Click [编辑] on draft item | Open DecisionFormDialog in Edit mode | Dialog opens with pre-filled data |
| Draft not found on edit | API returns 404 for draft | Show toast error: "该草稿已被删除或不存在", remove the item from timeline list, do not open dialog |
| Click [+ 添加决策] | Open DecisionFormDialog in New mode | Dialog opens with blank form |
| No more items to load | Hide sentinel, show "已加载 N 条" in footer | Text replaces spinner |
| Click [重试] on pagination error | Retry failed page request | Spinner replaces inline error until success or another failure |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| Category badge | `category` | API — mapped to label via enum lookup |
| Draft badge | `status === 'draft'` | API — conditional render |
| Content text | `content` | API — truncated/expanded |
| Tag badges | `tags[]` | API — iterated |
| Creator text | `createdBy` → user name | API — server embeds `createdByName` in decision list response; no client-side lookup |
| Time text | `createTime` | API — formatted relative time (e.g. "2026-04-27 14:30") |
| Edit button | `status === 'draft' && createdBy === currentUser` | API + auth context |

### Accessibility

**Keyboard navigation**:

| Context | Key | Action |
|---------|-----|--------|
| Timeline item (focused) | Tab | Move to next timeline item |
| Timeline item (focused) | Shift+Tab | Move to previous timeline item or header |
| Collapsed content (focused) | Enter / Space | Expand content |
| Expanded content (focused) | Enter / Space | Collapse content |
| Draft edit button (focused) | Enter | Open DecisionFormDialog in Edit mode |
| Add button (focused) | Enter | Open DecisionFormDialog in New mode |

**ARIA attributes**:

- Timeline container: `role="feed"` with `aria-label="决策记录"`
- Each timeline item: `role="article"`, `aria-label="{category} — {relative time}"`
- Collapsible content: `aria-expanded="false"` (collapsed) / `aria-expanded="true"` (expanded); `aria-controls` pointing to the content region's `id`
- Content region: `id` matching the `aria-controls` value
- Draft badge: `aria-label="草稿"`
- Category badge: `aria-label="{category label}"` (e.g. "技术")

**Live regions**:

- Loading/error region: `aria-live="polite"`, `aria-atomic="true"` on the footer area. Announces "加载中" during loading, "加载失败" on error, "已加载 N 条" on completion.
- Pagination error: same live region announces "加载更多失败" when page-2+ request fails.

**Focus management**:

- First timeline item receives `tabindex="0"`; subsequent items also `tabindex="0"` to be tab-navigable.
- Edit button and Add button are native `<button>` elements (focusable by default).

---

## Component 2: DecisionFormDialog

Dialog form for creating a new decision or editing an existing draft.

### Layout Structure

```
┌─ Dialog Overlay (bg-black/50, z-50) ────────────────────────┐
│                                                              │
│   ┌─ Dialog Content (rounded-xl, bg-surface, w-md 480px) ─┐│
│   │ Header (p-5 pb-0)                                      ││
│   │   ├── Title: "添加决策" / "编辑决策" (18px font-semibold)││
│   │   └── [×] close icon button (top-right, ghost)         ││
│   │                                                         ││
│   │ Body (p-5, space-y-4)                                   ││
│   │                                                         ││
│   │   ┌─ 分类 Field ──────────────────────────────────┐    ││
│   │   │  Label: "分类" (14px font-medium, required *) │    ││
│   │   │  Select: native select, h-10, w-full           │    ││
│   │   │  Placeholder: "请选择分类"                      │    ││
│   │   │  Error: "请选择分类" (error-text, 12px)        │    ││
│   │   └────────────────────────────────────────────────┘    ││
│   │                                                         ││
│   │   ┌─ 标签 Field ──────────────────────────────────┐    ││
│   │   │  Label: "标签" (14px font-medium)              │    ││
│   │   │  Input wrapper: flex, flex-wrap, gap-1,        │    ││
│   │   │    border, rounded-md, min-h-10, p-1           │    ││
│   │   │    ├── Badge: entered tags (× to remove)       │    ││
│   │   │    └── Text input: flex-1, min-w-20, no border │    ││
│   │   │  Recent tags dropdown: absolute, w-full,       │    ││
│   │   │    bg-surface, shadow-lg, rounded-md, z-10     │    ││
│   │   │    └── Tag items: px-3 py-2, hover bg-bg-alt   │    ││
│   │   │  Error: "标签不能超过 20 字符" (error-text)     │    ││
│   │   └────────────────────────────────────────────────┘    ││
│   │                                                         ││
│   │   ┌─ 决策内容 Field ──────────────────────────────┐    ││
│   │   │  Label: "决策内容" (14px font-medium, required*)│   ││
│   │   │  Textarea: min-h-[120px], resize-y, w-full     │    ││
│   │   │  Placeholder: "请输入决策内容"                   │    ││
│   │   │  Footer: flex justify-between                   │    ││
│   │   │    ├── Error: "请输入决策内容" (error-text)     │    ││
│   │   │    └── Counter: "0/2000" (tertiary, 12px)      │    ││
│   │   └────────────────────────────────────────────────┘    ││
│   │                                                         ││
│   │ Footer (flex justify-end, gap-2, px-5 pb-5)            ││
│   │   ├── [取消] secondary button                           ││
│   │   ├── [保存草稿] secondary button (with loading state)  ││
│   │   └── [发布] primary button (with loading state)        ││
│   └─────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
```

**Category select**: styled to match existing input pattern (h-10, rounded-md, border-border-dark). Options:

| Value | Label |
|-------|-------|
| technical | 技术 |
| resource | 资源 |
| requirement | 需求 |
| schedule | 进度 |
| risk | 风险 |
| other | 其他 |

**Tags input (tag-input pattern)**:
- Wrapper: flex, flex-wrap, align-center, gap-1, border `#cbd5e1`, rounded-md, min-h-10, px-1 py-1.5, focus-within ring-2 ring `#bfdbfe`
- Each entered tag rendered as a small badge (rounded-md, px-2 py-0.5, bg-bg-alt, text-secondary, 12px) with a × remove button
- Text input: borderless, flex-1, min-w-20, placeholder "输入标签，回车添加"
- Separators: Enter key or comma → creates tag, clears input
- Duplicate prevention: ignore if tag already exists
- Recent tags dropdown: shows when input is focused and recent tags exist, filtered by current input text. Max 10 suggestions.

**Textarea**: matches input styling, min-h-[120px], resize-y. Character counter in bottom-right corner updates live.

### States

| State | Visual | Behavior |
|-------|--------|----------|
| New | Title "添加决策", all fields empty, category placeholder "请选择分类" | Creating new decision |
| Edit | Title "编辑决策", fields pre-filled with draft data | Editing existing draft |
| Loading Draft | Title "编辑决策", body shows centered spinner (h-8 w-8), fields hidden | Edit mode: fetching draft data by bizKey from API. Transitions to Edit state on success, or closes with toast error on 404 |
| Submitting (Draft) | [保存草稿] shows spinner + disabled, other buttons disabled | API call in progress |
| Submitting (Publish) | [发布] shows spinner + disabled, other buttons disabled | API call in progress |
| Validation Error | Error messages below affected fields, field border turns error color | On submit with invalid fields |
| Character limit | Counter text turns warning color when > 1800, error color at 2000 | Live on content input |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Open dialog | Focus category select | Keyboard-ready |
| Open dialog (Edit mode) | Parent passes `bizKey` as prop; dialog calls `GET /api/v1/teams/:teamId/main-items/:itemId/decisions/:bizKey` to fetch fresh draft data | Show "加载中..." spinner in dialog body until API responds; populate fields from API response |
| Select category | Set value, clear error | Field updates |
| Type in tag input | Filter recent tags dropdown | Dropdown shows matches |
| Press Enter/comma in tag input | Add tag badge, clear input, close dropdown | Tag appears in wrapper |
| Click × on tag badge | Remove tag | Tag removed |
| Click recent tag suggestion | Add tag to wrapper | Tag appears, dropdown stays open |
| Type in content | Update character counter | Counter updates live |
| Click [保存草稿] | Validate → call save draft API | Button loading → on success: close dialog, refresh timeline |
| Click [发布] | Validate → call publish API | Button loading → on success: close dialog, refresh timeline |
| Click [取消] / [×] / overlay | Close dialog (no confirmation if unchanged) | Dialog closes |
| Close with unsaved changes | Show confirmation: "未保存的内容将丢失，确认关闭？" | [确认] closes, [取消] returns to form |
| Escape key | Close dialog (same as [×]) | — |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| Dialog title | mode (new/edit) | Prop from parent |
| Category select | `category` | New mode: form state (empty). Edit mode: API response from `GET .../decisions/:bizKey` |
| Tag badges | `tags[]` | New mode: form state (empty). Edit mode: API response from `GET .../decisions/:bizKey` |
| Content textarea | `content` | New mode: form state (empty). Edit mode: API response from `GET .../decisions/:bizKey` |
| Character counter | `content.length` | Derived from form state |
| Submit buttons | loading state | API call state |

### Accessibility

**Dialog ARIA attributes**:

- Dialog element: `role="dialog"`, `aria-modal="true"`, `aria-labelledby` pointing to the dialog title's `id`
- Dialog title: `id` matching the `aria-labelledby` value
- Overlay: `aria-hidden="true"` (prevents screen readers from accessing background content)

**Focus trap**:

- When dialog opens: focus moves to the category select (first interactive element)
- Tab / Shift+Tab cycles only through focusable elements within the dialog (focus trap)
- When dialog closes: focus returns to the element that triggered the dialog (edit button or add button)
- Escape key closes the dialog (same as cancel/close button)

**Keyboard navigation**:

| Context | Key | Action |
|---------|-----|--------|
| Dialog open | Escape | Close dialog (with unsaved-changes confirmation if dirty) |
| Tag text input | Enter | Add tag (prevent form submission) |
| Tag text input | , (comma) | Add tag |
| Recent tags dropdown open | ArrowDown / ArrowUp | Navigate suggestions |
| Recent tags dropdown open | Enter | Select focused suggestion |
| Recent tags dropdown open | Escape | Close dropdown only (not dialog) |

**Screen reader announcements**:

- Validation errors: each error message uses `role="alert"` so screen readers announce it immediately when it appears
- Character counter: `aria-live="polite"` on the counter element so the count is announced when the user pauses typing
- Submitting state: submit button `aria-disabled="true"` with `aria-busy="true"` during API call; screen reader announces loading via button text change
- Confirmation dialog (unsaved changes): `role="alertdialog"` with `aria-labelledby` pointing to the confirmation text
