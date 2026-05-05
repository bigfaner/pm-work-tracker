---
created: "2026-05-04"
tags: [architecture, testing]
---

# Task Breakdown Must Include Explicit "Wire into Page" Task for New UI Components

## Problem

DecisionTimeline component was fully built (24 unit tests, all passing) but never integrated into MainItemDetailPage. This caused:
1. TC-005 e2e failure — published decision content not visible on the detail page
2. 9 skipped UI e2e tests that depend on DecisionTimeline being present in the page
3. The feature appears unimplemented from the user's perspective despite all backend + component code being complete

## Root Cause

Causal chain (4 levels):

1. **Symptom**: DecisionTimeline exists as a standalone component file but is never imported or rendered in MainItemDetailPage
2. **Direct cause**: No task in the 18-task breakdown covered "integrate DecisionTimeline into MainItemDetailPage"
3. **Code cause**: Task 3.2 ("Build DecisionTimeline component") description said "embedded in the main item detail page" but acceptance criteria only listed component-level checks (layout, infinite scroll, permissions, empty states). No acceptance criterion required the component to actually appear in the page
4. **Root cause**: The task breakdown process decomposed PRD requirements into component-creation tasks but treated "component works" as equivalent to "component is integrated". The PRD line 208 explicitly states "在子事项表格上方插入 DecisionTimeline 组件" but this integration step was absorbed into task 3.2's description text without a verifiable acceptance criterion

**Why the breakdown missed this**: Component creation and page wiring are distinct activities requiring different files and different test strategies. The breakdown treated them as one task because the PRD describes them together. But "build a component" and "insert it into a page" are separate concerns — one tests in isolation, the other tests in context.

## Solution

The fix requires a dedicated integration task:
1. Import DecisionTimeline in MainItemDetailPage
2. Add it above the sub-items table
3. Wire data fetching (list decision logs for the current mainItem)
4. Verify with e2e tests that decisions appear on the detail page

## Reusable Pattern

### When breaking down tasks for a new UI component, ALWAYS include two separate tasks:

1. **Component Implementation Task** — build the component in isolation with unit tests
   - Acceptance criteria: rendering, interactions, edge cases (all testable in isolation)
   - Deliverable: `<Component>.tsx` + `<Component>.test.tsx`

2. **Page Integration Task** — wire the component into its parent page
   - Acceptance criteria: component appears at correct position, data flows correctly, e2e test passes
   - Deliverable: modified parent page + integration e2e test verification

### Why separate tasks:

| Aspect | Component Task | Integration Task |
|--------|---------------|-----------------|
| Files changed | New component + test | Existing page file |
| Test strategy | Unit test (render in isolation) | E2E test (verify in page context) |
| Failure mode | Component logic bugs | Wiring/data-flow bugs |
| Verifiable by | Unit test pass | E2E test pass |

### Quick checklist for task breakdown review:

- [ ] For each new UI component in the PRD, is there both a "build" and an "integrate" task?
- [ ] Does the integration task have acceptance criteria that reference the parent page?
- [ ] Does any task description say "embedded in" or "shown on" without a corresponding integration task?

## Example

**PRD requirement**: "在子事项表格上方插入 DecisionTimeline 组件"

**Correct task breakdown**:
- Task 3.2: Build DecisionTimeline component (unit tests: layout, scroll, permissions, empty state)
- Task 3.3: Integrate DecisionTimeline into MainItemDetailPage (verify: appears above sub-items table, e2e TC-005 passes)

**Incorrect task breakdown** (what happened):
- Task 3.2: Build DecisionTimeline component "embedded in the main item detail page" (unit tests only — no integration verification)

## Related Files

- `docs/features/decision-log/prd/prd-spec.md` — line 208: integration requirement
- `docs/features/decision-log/tasks/index.json` — 18 tasks, no integration task
- `frontend/src/pages/main-item-detail/DecisionTimeline.tsx` — standalone component, never imported
- `frontend/src/pages/main-item-detail/MainItemDetailPage.tsx` — missing DecisionTimeline import
