---
date: "2026-04-23"
doc_dir: "docs/features/weekly-stats-optimization/design/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 96/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  18      │  20      │ ⚠️          │
│    Layer placement explicit  │   7/7    │          │            │
│    Component diagram present │   6/7    │          │            │
│    Dependencies listed       │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  19      │  20      │ ⚠️          │
│    Interface signatures typed│   7/7    │          │            │
│    Models concrete           │   6/7    │          │            │
│    Directly implementable    │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  15      │  15      │ ✅          │
│    Error types defined       │   5/5    │          │            │
│    Propagation strategy clear│   5/5    │          │            │
│    HTTP status codes mapped  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  15      │  15      │ ✅          │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   5/5    │          │            │
│    Test tooling named        │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  19      │  20      │ ⚠️          │
│    Components enumerable     │   7/7    │          │            │
│    Tasks derivable           │   7/7    │          │            │
│    PRD AC coverage           │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  10      │  10      │ N/A        │
│    Threat model present      │   5/5    │          │            │
│    Mitigations concrete      │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  96      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 19/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md § Architecture Fit | Data flow diagram is a linear chain; does not show error path (isError → "-"), isOverdue function position, or Tooltip component relationships | -1 pt (Architecture: Component diagram) |
| tech-design.md § Architecture Fit | `pkg/apperrors` referenced in Error Handling but absent from the External Dependencies table; `@radix-ui/react-tooltip` and `testify` versions not specified | -1 pt (Architecture: Dependencies) |
| tech-design.md § Interface 2 | `si.Status`, `si.ExpectedEndDate` used in service snippet but SubItem struct definition not included or referenced — developer must locate the model independently | -1 pt (Interface: Models concrete) |
| tech-design.md § Interface 3 / PRD Coverage Map | PRD scope explicitly requires "含条件优先级 fixture" (condition 1 overrides condition 2); Interface 3 fixture table has no such case; PRD Coverage Map omits Interface 3 and Interface 7 from its traceability rows | -1 pt (Breakdown-Readiness: PRD AC coverage) |

---

## Attack Points

### Attack 1: Architecture — diagram is a linear flow, not a component diagram

**Where**: `tech-design.md § Architecture Fit` — the diagram reads:
```
buildWeeklyGroups (view_service.go)
  → WeeklyStats DTO (item_dto.go)
  → view_handler.go  →  HTTP 200 JSON
                              ↓
                    useQuery (WeeklyViewPage.tsx)
                              ...
                    StatCard × 7
```

**Why it's weak**: The diagram shows the happy-path data flow only. It omits: (1) the error path — when `isError=true`, how does the signal reach StatCard to render `"-"`? (2) the `isOverdue` function — it's a key changed component but appears nowhere in the architecture diagram; (3) the Tooltip component — it's a structural child of StatCard but the diagram ends at `StatCard × 7` with no further decomposition. A reviewer cannot tell from the diagram alone that StatCard wraps a Radix Tooltip. The External Dependencies table also omits `pkg/apperrors`, which is referenced in the Error Handling section.

**What must improve**: Extend the diagram with a branch for the error path (`isError=true → StatsBar renders "-"`). Add `isOverdue (lib/status.ts)` as a node connected to `WeeklyViewPage.tsx`. Add `pkg/apperrors` to the External Dependencies table with its role.

---

### Attack 2: Interface & Model Defs — SubItem model fields used but never defined

**Where**: `tech-design.md § Interface 2` — the service snippet uses `si.Status`, `si.ExpectedEndDate`, `si.ExpectedEndDate.Before(weekEnd)` without defining the SubItem struct. The doc states "si.ExpectedEndDate 是 `*time.Time`，nil 时不计入逾期" in a prose note, but the full SubItem field set is never shown.

**Why it's weak**: A developer implementing `buildWeeklyGroups` must hunt for the SubItem model definition elsewhere. The rubric criterion "Models concrete" requires all model fields to be named with types and constraints. The field `ExpectedEndDate *time.Time` is mentioned in prose but not in a struct definition. If the actual field name differs (e.g., `DueDate` or `EndDate`), the snippet silently breaks. The iteration 1 report flagged this exact issue and it remains unaddressed.

**What must improve**: Add a minimal SubItem field excerpt showing at minimum the fields used in the snippet: `Status string`, `ExpectedEndDate *time.Time`, `ActualEndDate *time.Time`. A one-line comment referencing the source file (`model/sub_item.go`) would also suffice.

---

### Attack 3: Breakdown-Readiness — PRD "条件优先级 fixture" not covered in Interface 3

**Where**: `prd-spec.md § In Scope` — "新增 `buildWeeklyGroups` 单元测试，覆盖 7 个卡片的计数规则（**含条件优先级 fixture**）". `tech-design.md § Interface 3` fixture table lists 7 cases but none tests the condition priority rule: "若子事项本周内有进展记录，则无论其 actualEndDate 是否早于 weekStart，该事项均计为本周活跃."

**Why it's weak**: The PRD business flow explicitly calls out condition 1 (progress record) overriding condition 2 (date overlap) as a non-obvious rule. Without a fixture that sets `actualEndDate < weekStart` AND `progressRecord ∈ [weekStart, weekEnd)` and asserts `activeSubItems=1`, this edge case will not be regression-tested. The PRD Coverage Map also omits Interface 3 and Interface 7 entirely from its traceability rows — two of the seven interfaces have no PRD AC mapped to them.

**What must improve**: Add a fixture to Interface 3: "有进展记录但 actualEndDate < weekStart → activeSubItems=1（条件 1 优先）". Update the PRD Coverage Map to include Interface 3 (→ PRD scope item: backend unit tests) and Interface 7 (→ PRD scope item: frontend tooltip unit tests).

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Testing — zero coverage target, unnamed tooling | ✅ | Testing Strategy table now has numeric targets (`buildWeeklyGroups` ≥ 90% line coverage; all tooltip states covered) and names `go test` + `testify/assert`, `vitest` + `@testing-library/react` + `userEvent`, `Playwright` |
| Attack 2: Architecture — diagram shows file tree, not data flow | ✅ | Replaced with ASCII flow diagram showing full backend→frontend data path; External Dependencies table added with 3 packages |
| Attack 3: Error Handling — no Go error types, silent propagation gap | ✅ | `apperrors.ErrValidation` and `apperrors.CodeValidation` now shown with package reference; handler→middleware→HTTP 400 propagation path explicitly described; frontend `throw new Error("weekEnd is required")` specified |

---

## Verdict

- **Score**: 96/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 19/20 — can proceed to `/breakdown-tasks`
- **Action**: All three iteration-1 attacks resolved. Three minor gaps remain (diagram completeness, SubItem model reference, PRD fixture coverage). Design is implementable as-is.
