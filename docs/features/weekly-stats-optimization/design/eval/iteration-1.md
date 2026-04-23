---
date: "2026-04-23"
doc_dir: "docs/features/weekly-stats-optimization/design/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

**Score: 80/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  14      │  20      │ ⚠️          │
│    Layer placement explicit  │   6/7    │          │            │
│    Component diagram present │   4/7    │          │            │
│    Dependencies listed       │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  18      │  20      │ ✅          │
│    Interface signatures typed│   7/7    │          │            │
│    Models concrete           │   6/7    │          │            │
│    Directly implementable    │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  11      │  15      │ ⚠️          │
│    Error types defined       │   3/5    │          │            │
│    Propagation strategy clear│   3/5    │          │            │
│    HTTP status codes mapped  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │   7      │  15      │ ❌          │
│    Per-layer test plan       │   4/5    │          │            │
│    Coverage target numeric   │   0/5    │          │            │
│    Test tooling named        │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  20      │  20      │ ✅          │
│    Components enumerable     │   7/7    │          │            │
│    Tasks derivable           │   7/7    │          │            │
│    PRD AC coverage           │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  10      │  10      │ N/A        │
│    Threat model present      │   5/5    │          │            │
│    Mitigations concrete      │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  80      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 20/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md § Architecture Fit | ASCII tree shows file paths only; no data-flow arrows between layers or between backend and frontend | -3 pts (Architecture: Component diagram) |
| tech-design.md § Architecture Fit | External package dependencies not named (`@radix-ui/react-tooltip`, `@tanstack/react-query` version, etc.) | -2 pts (Architecture: Dependencies) |
| tech-design.md § Architecture Fit | Frontend layer placement listed as prose below the diagram, not integrated into it; handler layer says "无变更" but no explanation of what it does pass through | -1 pt (Architecture: Layer placement) |
| tech-design.md § Interface 2 | `si.Status`, `si.ExpectedEndDate` fields used in service snippet but SubItem model definition not included or referenced — developer must hunt for it | -1 pt (Interface: Models concrete) |
| tech-design.md § Interface 6 / Error Handling | `isError` prop integration described as "需在 StatsBar 调用处传入 isError prop 或在父组件处理" — two options left open, no decision made | -1 pt (Interface: Directly implementable) |
| tech-design.md § Error Handling | No Go error constants defined; error codes are bare strings (`"VALIDATION_ERROR"`) with no reference to `apperrors` package or equivalent | -2 pts (Error Handling: Error types defined) |
| tech-design.md § Error Handling | Backend service→handler error propagation not described: how does a validation failure in `buildWeeklyGroups` surface as a 400? | -2 pts (Error Handling: Propagation strategy) |
| tech-design.md § Testing Strategy | Zero numeric coverage target anywhere in the document | -5 pts (Testing: Coverage target numeric) |
| tech-design.md § Testing Strategy | Test frameworks not named in the design doc itself: "table-driven" is a pattern, not a tool; "RTL" is named but `vitest` / `testify` are absent | -2 pts (Testing: Test tooling named) |
| tech-design.md § Testing Strategy | No E2E test plan for the tooltip interaction or the 7-column layout across breakpoints | -1 pt (Testing: Per-layer test plan) |

---

## Attack Points

### Attack 1: Testing Strategy — zero coverage target and unnamed tooling

**Where**: `tech-design.md § Testing Strategy` table — the entire "Coverage target numeric" row is absent from the document. The tooling column says "RTL" for the component layer but names no framework for Go tests.

**Why it's weak**: The rubric requires a numeric coverage target. Without one, there is no pass/fail gate for the test phase. "Table-driven" describes a pattern, not a tool — a new contributor cannot derive which runner, assertion library, or coverage reporter to use from this doc alone. The project conventions (CLAUDE.md) fill the gap, but a design doc should be self-contained enough to drive implementation without cross-referencing project meta-files.

**What must improve**: Add a coverage target (e.g., "backend service: 90% line coverage on `buildWeeklyGroups`; frontend component: all 7 stat cards and tooltip states covered"). Name the tooling explicitly: `go test` + `testify/assert` for Go, `vitest` + `@testing-library/react` + `userEvent` for frontend.

---

### Attack 2: Architecture Clarity — diagram shows file tree, not component interactions

**Where**: `tech-design.md § Architecture Fit` — the diagram is:
```
Handler (view_handler.go)          — 无变更
  └─ Service (view_service.go)     — buildWeeklyGroups 新增 3 个计数分支
       └─ DTO (item_dto.go)        — WeeklyStats 新增 3 字段

Frontend
  types/index.ts                   — WeeklyStats 接口新增 3 字段
  ...
```

**Why it's weak**: This is an annotated file list, not an architecture diagram. It shows no data flow: how does `WeeklyStats` travel from the DTO through the handler to the HTTP response and into the frontend `useQuery` hook? The backend and frontend sections are separated by a blank line with no connection shown. External dependencies (`@radix-ui/react-tooltip`, React Query) are entirely absent from the dependency listing.

**What must improve**: Replace or supplement the tree with a flow diagram (ASCII arrows are fine) showing: `buildWeeklyGroups → WeeklyStats DTO → JSON response → useQuery → WeeklyViewResponse → StatsBar → StatCard`. Add an explicit dependency block listing external packages with versions.

---

### Attack 3: Error Handling — no Go error types and silent backend propagation gap

**Where**: `tech-design.md § Error Handling` table and `api-handbook.md § Error Responses`. The api-handbook maps HTTP codes correctly, but the design doc never defines how the backend produces those codes. The Error Handling table entry for `weekEnd undefined` says "调用方须抛出错误" without specifying what error or where.

**Why it's weak**: The api-handbook says `400 VALIDATION_ERROR` for "weekStart 格式错误或非周一", but `tech-design.md` never shows where this validation lives (handler? service?), what Go error type is returned, or how the handler converts it to a 400. A developer implementing the handler has to guess the error-to-status mapping. The `"VALIDATION_ERROR"` string is used as a code but never defined as a constant — it could be misspelled in implementation.

**What must improve**: Define the validation error path explicitly: "handler calls `dto.ParseWeekStart()`; on error, returns `apperrors.ErrValidation` which the error middleware maps to 400 + `VALIDATION_ERROR`." Define `VALIDATION_ERROR` as a named constant or reference the existing error package. For the frontend `weekEnd undefined` case, specify the exact throw: `throw new Error("weekEnd is required")` or equivalent.

---

## Previous Issues Check

N/A — iteration 1.

---

## Verdict

- **Score**: 80/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 20/20 — can proceed to `/breakdown-tasks`
- **Action**: Design is implementable as-is. Address Attack 1 (testing) and Attack 3 (error propagation) before the next iteration if a higher score is required.
