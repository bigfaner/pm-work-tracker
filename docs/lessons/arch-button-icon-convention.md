---
created: "2026-05-04"
tags: [interface, dependencies]
---

# Action Buttons Must Use Icon + Text Pattern

## Problem

DecisionTimeline's "添加决策" and "编辑" buttons used plain text without icons, while every other action button in the codebase pairs text with a lucide-react icon:

- Every "编辑" button elsewhere uses `<Pencil>` + "编辑"
- Every "添加/新增" button elsewhere uses `<Plus>` + text or inline SVG plus + text
- DecisionTimeline only imported `ChevronDown` for the collapse toggle, ignoring action button icons

## Root Cause

Causal chain (5 levels):

1. **Symptom**: DecisionTimeline buttons are plain text, visually inconsistent with the rest of the app
2. **Direct cause**: The component only imported `ChevronDown` from lucide-react, never added `<Plus>` or `<Pencil>` for action buttons
3. **Code cause**: ui-design.md specified `[+ 添加决策]` (line 34) but the agent dropped the `+` prefix entirely; `[编辑]` and `[重试]` had no icon markers so the agent used plain text
4. **Spec interpretation failure**: ui-design.md uses wireframe notation `[+ text]` which is ambiguous — the agent treated `[+]` as a decorative bracket symbol rather than a requirement to show a plus indicator. For `[编辑]` and `[重试]`, the spec has no icon notation at all
5. **Root cause**: **The agent did not cross-reference existing similar components when interpreting the spec.** The ui-design doc is a wireframe-level description — it doesn't (and shouldn't) repeat every project-wide styling convention. The implementer must check existing code in the same context (SubItemsTable, ItemPoolPage) for unwritten conventions. In this case, every action button in the codebase uses icon + text, but the agent built DecisionTimeline in isolation without looking at sibling components

**Why the spec's `[+ 添加决策]` was dropped**: The `[brackets]` in the wireframe denote a button element. The `+` inside the brackets is part of the button label, but the agent interpreted only the Chinese text as the actual label and discarded the `+` prefix as wireframe notation.

## Solution

Added lucide-react icons to DecisionTimeline action buttons:
- `<Plus size={14} />` before "添加决策"
- `<Pencil size={12} />` before "编辑"
- `<RefreshCw size={12} />` before "重试"

## Reusable Pattern

### Button icon convention (must follow):

| Action | Icon | Import | Example |
|--------|------|--------|---------|
| 添加 / 新增 | Plus | `import { Plus } from "lucide-react"` | `<Plus size={14} /> 添加决策` |
| 编辑 | Pencil | `import { Pencil } from "lucide-react"` | `<Pencil size={12} /> 编辑` |
| 刷新 / 重试 | RefreshCw | `import { RefreshCw } from "lucide-react"` | `<RefreshCw size={12} /> 重试` |
| 删除 | Trash2 | `import { Trash2 } from "lucide-react"` | `<Trash2 size={14} /> 删除` |
| 查看 / 预览 | Eye | `import { Eye } from "lucide-react"` | `<Eye className="w-4 h-4" /> 预览` |

### Rules for action buttons:

1. **All action buttons must use icon + text**, not text alone. This includes header actions, row actions, and error retry buttons.
2. **Use lucide-react icons**, not inline `<svg>`. The inline SVG pattern in MainItemDetailPage/SubItemDetailPage is legacy — new code uses lucide-react.
3. **Icon size conventions**: `size={14}` for header/default buttons, `size={12}` for small/ghost row actions.

### When building new components with buttons:

- **Cross-reference existing sibling components** before writing any button. Look at components in the same context (e.g., SubItemsTable for anything in main-item-detail page) and match their button style exactly
- **ui-design.md wireframe notation is a minimum spec, not a complete spec.** If the wireframe says `[+ 添加决策]`, implement `+` (as icon or text). If the wireframe says `[编辑]` but existing similar buttons use `<Pencil>`, add the icon to match
- Import the appropriate lucide-react icon alongside the Button component
- The icon goes inside the `<Button>` tag, before the text

## Example

```tsx
// ❌ Wrong — plain text button
<Button variant="primary" size="sm" onClick={onAdd}>
  添加决策
</Button>

// ✅ Correct — icon + text
import { Plus, Pencil, RefreshCw } from "lucide-react";

<Button variant="primary" size="sm" onClick={onAdd}>
  <Plus size={14} />
  添加决策
</Button>

<Button variant="ghost" size="sm" onClick={() => onEdit(log.bizKey)}>
  <Pencil size={12} />
  编辑
</Button>
```

## Related Files

- `frontend/src/pages/main-item-detail/SubItemsTable.tsx` — correct pattern: `<Pencil>` + `<Plus>` with text
- `frontend/src/pages/main-item-detail/DecisionTimeline.tsx` — was missing icons on action buttons
- `frontend/src/pages/item-view/ItemSummaryView.tsx` — correct pattern: icon + text for edit/add
- `frontend/.claude/rules/frontend.md` — frontend conventions (should add button icon rule)
