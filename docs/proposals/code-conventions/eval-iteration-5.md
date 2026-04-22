---
proposal: docs/proposals/code-conventions/proposal.md
iteration: 5
score: 80/100
date: 2026-04-22
---

# Proposal Evaluation Report — Code Quality: Performance, Duplication & Readability

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 16 | 20 |
| Solution Clarity | 17 | 20 |
| Alternatives Analysis | 12 | 15 |
| Scope Definition | 12 | 15 |
| Risk Assessment | 12 | 15 |
| Success Criteria | 13 | 15 |
| Vague language penalty | -2 | — |
| **Total** | **80** | **100** |

---

## Changes Since Iteration 4

All three attacks from iteration 4 were addressed:

| Attack | Status | Notes |
|--------|--------|-------|
| ESLint has no measurable completion gate | ✅ Fixed | Success criteria now includes: "ESLint 对 API 层文件命名违规（非 camelCase.ts）和组件导出命名违规（非 PascalCase）返回非零退出码；验证命令：`npx eslint src/api/BadName.ts` 退出码非零，`npx eslint src/components/badComponent.tsx` 退出码非零" |
| linkageMuMap criterion is unverifiable | ✅ Fixed | Criterion now specifies: "验证方式：PR review checklist 中逐行确认删除路径存在" with concrete conditions (deletion call after EvaluateLinkage, or sync.Map with explicit Delete) |
| Evidence is approximate and misplaced | ✅ Fixed | "约60行" replaced with exact "62行" and specific line references (TableView 576–608, TableExportCSV 638–649, 651–662, 668–669); 21-duplicate breakdown moved into Problem section |

Score improvement: 78 → 80. Remaining structural weaknesses are in Phase 3 deferred specificity (affects scope, solution clarity, and success criteria), risk impact underrating, and the single-incident basis for the AI recurrence claim.

---

## Dimension Breakdown

### 1. Problem Definition — 16/20

**Problem stated clearly (6/7)**

Performance problems remain well-specified with file locations and line numbers. Duplication now has exact counts with line references. "规范缺失" is still underspecified — "AI每次会话无法获得一致的风格指引" names the symptom but not the scope: how many rule categories are inconsistent? Which ones? One historical snake_case incident is cited; it is unclear whether this is isolated or representative of a pattern. -1.

**Evidence provided (5/7)**

Genuine improvement: "约60行" is now "62行" with specific line ranges, and the 21-duplicate breakdown is now in the Problem section where it belongs. The N+1 example is concrete with file and line. GORM log observation is a real data point. Remaining gaps: no query latency measurements or response time data; the current item count is never stated — the 6-month projection ("500+ main_items") has no anchor. A reader cannot evaluate whether the projection is plausible without knowing today's count and growth rate. -2.

**Urgency justified (5/6)**

The "Why now" section provides a no-delete-path argument, a GORM log observation, a 6-month projection, and a historical AI session incident. The projection is concrete in its consequences. -1 because the projection is unanchored: "按当前迭代速度 main_items 预计增长至 500+" does not state the current count or the iteration velocity, making the timeline unverifiable.

---

### 2. Solution Clarity — 17/20

**Approach is concrete (6/7)**

Phase 0 is strong: specific method signatures, specific rename targets. Phase 1 lists exact file paths. Phase 2 names `tagliatelle` and the ESLint rules. Phase 3 "识别重复UI模式，抽取可复用组件（具体组件名在 Phase 3 开始前确定）" remains discovery work, not a deliverable. A scope item whose content is unknown at proposal time cannot be evaluated for concreteness. -1.

**User-facing behavior described (6/7)**

Phase 0: GORM query log verification — concrete and verifiable. Phase 1: behavioral verification via `@rules/naming.md` in a new session — genuine fix from iteration 3. Phase 2: non-zero exit codes for both golangci-lint and ESLint — concrete. Phase 3: grep count for backend — concrete; component extraction observable state is still deferred ("具体组件名在 Phase 3 开始前确定"), so the observable state for that deliverable cannot be evaluated now. -1.

**Distinguishes from alternatives (5/6)**

The bundle rationale is substantive: "Phase 0 修复的是实现 bug，Phases 1–3 修复的是产生 bug 的过程缺陷" is a real argument. The recurrence prevention logic ("Phase 0 的修复会被逐步侵蚀") is concrete and logical. -1 because the "AI will reintroduce bugs" claim rests on a single historical incident (one snake_case tag in model layer). One data point does not establish a pattern; the proposal does not state how frequently AI sessions have introduced violations or whether the rate is increasing. If violations are rare, the urgency of bundling is weaker than stated.

---

### 3. Alternatives Analysis — 12/15

**At least 2 alternatives listed (5/5)**

Four alternatives including Do Nothing. Full marks.

**Pros/cons for each (3/5)**

Alternative D's cons are genuinely honest and specific — the strongest section. Alternative B's pros ("后续改动减少重复文件触碰") are still unquantified — vague language penalty applied separately. Alternative A's cons ("总工作量比单做 Phase 0 大；Phase 3 工期较长") are real but do not acknowledge the risks introduced by the performance fixes themselves (behavior changes, test gaps in Phase 0). Alternative C's cons are real but not elaborated. -2.

**Rationale for chosen approach (4/5)**

The "implementation bug vs. process defect" framing is a legitimate and clear argument for bundling. The dependency chain (Phase 1 → Phase 2 → Phase 3) is well-argued. -1 because the claim that "下一个 AI 会话...仍会引入新的 N+1 查询" is asserted without evidence of frequency — if AI sessions introduce violations rarely, the urgency of bundling is weaker than stated.

---

### 4. Scope Definition — 12/15

**In-scope items are concrete (4/5)**

Phase 0 items are concrete. Phase 1 file paths are concrete. Phase 2 names specific lint rules for both Go and TypeScript. Phase 3 "前端可复用组件抽取（≥3 个，名称在 Phase 3 开始前确定）" remains open-ended discovery work — a deliverable whose content is unknown at proposal time is not a bounded scope item. -1.

**Out-of-scope explicitly listed (4/5)**

Includes CI/CD lint integration and bulk migration of existing violations. Still missing: whether existing tests are in scope for updates when Phase 0 changes repo interfaces (e.g., `UserRepo` gains `FindByIDs` — callers and test mocks may need updating). This is a concrete consequence of Phase 0 that is neither in-scope nor explicitly out-of-scope. -1.

**Scope is bounded (4/5)**

T-shirt sizing per phase is present. Sequencing constraints are stated. Total target "3 个 sprint（约 3 週）" is a real boundary. -1 because the "约" estimates have no basis stated — Phase 0 is "约3–5天" but the estimate methodology is not given. A reader cannot validate whether M = 3–5 days is realistic without knowing team size or velocity.

---

### 5. Risk Assessment — 12/15

**Risks identified (5/5)**

Five risks listed, all meaningful. The lint adoption failure mode is present with a specific mitigation. Full marks.

**Likelihood + impact rated (3/5)**

Ratings are provided but remain conservative. "规范文档过于理想化" is rated Medium/Medium — if the docs are never adopted by AI sessions, the entire Phase 1 investment is wasted, which is High impact, not Medium. The linkageMuMap risk is Low/High — a global mutex map that grows unboundedly in a production process is a memory leak with potential OOM consequences; "High" is correct but the likelihood of "Low" is questionable given the code already has no delete path. -2.

**Mitigations are actionable (4/5)**

The lint adoption mitigation is very specific: a concrete command with a threshold (>20 violations → add cleanup subtask). The linkageMuMap mitigation is actionable. The "规范文档过于理想化" mitigation ("规范中提供具体代码示例和反例") is actionable. Remaining gap: "改动前后对比测试；现有 repo 测试覆盖" for the TableView risk — what tests? Against what baseline? The existing test suite is not described, so "现有 repo 测试覆盖" is not actionable without knowing what it covers. -1.

---

### 6. Success Criteria — 13/15

**Criteria are measurable (5/5)**

Phase 0 criteria are binary and measurable. linkageMuMap now has a specified verification method: "PR review checklist 中逐行确认删除路径存在" with concrete conditions (deletion call after EvaluateLinkage, or sync.Map with explicit Delete). ESLint criterion added with specific test commands. Full marks.

**Coverage is complete (4/5)**

Phase 0: 1:1 mapping with scope — good. Phase 1: section checklists for ARCHITECTURE.md and DECISIONS.md — good. Phase 2: both golangci-lint and ESLint now covered — fixed from iteration 4. Phase 3: "至少抽取3个可复用前端UI组件（需在Phase 3开始前列出具体组件名）" defers specificity — the criterion is incomplete as written. -1.

**Criteria are testable (4/5)**

Phase 0: testable. Phase 1: behavioral test present — "在新 Claude Code 会话中执行 `@rules/naming.md`，规则内容出现在上下文中" — genuine improvement. Phase 2: both golangci-lint and ESLint testable with specific commands — fixed from iteration 4. Phase 3: "至少抽取3个可复用前端UI组件" cannot be tested until component names are specified. -1.

---

## Vague Language Penalties

| Instance | Location | Penalty |
|----------|----------|---------|
| "后续改动减少重复文件触碰" (reduced file touching, unquantified) | Alternatives B, Pros | -2 |

Total penalty: **-2**

---

## Top Attacks

### Attack 1 — Scope / Solution Clarity / Success Criteria: Phase 3 component work is unbounded across three dimensions

Phase 3 frontend component extraction appears in scope, solution, and success criteria — and in all three places it defers specificity to "Phase 3 开始前确定". This is not a minor gap: it means the largest phase (L, 5–7 days) has an unknown deliverable. A scope item, solution description, and success criterion that all say "TBD" are not scope, solution, or criteria — they are placeholders.

Quote: "前端可复用组件抽取（≥3 个，名称在 Phase 3 开始前确定）" (Scope); "识别重复UI模式，抽取可复用组件（具体组件名在 Phase 3 开始前确定）" (Solution); "至少抽取3个可复用前端UI组件（需在Phase 3开始前列出具体组件名）" (Success Criteria)

What must improve: Either name the components now (they can be identified by reading the existing frontend code) or explicitly scope Phase 3 frontend work as a separate proposal. A proposal that defers its largest deliverable's definition to a future date is not ready for approval.

---

### Attack 2 — Risk Assessment: "规范文档过于理想化" impact is underrated

"规范文档过于理想化" is rated Medium/Medium. If the convention docs are written but never loaded or followed in AI sessions, Phase 1 produces zero value — the entire 1–2 day investment is wasted and Phases 2–3 lose their rationale (lint rules without a backing convention doc are arbitrary). This is not Medium impact; it is High. The mitigation ("规范中提供具体代码示例和反例") addresses doc quality but not adoption — there is no mechanism to verify that AI sessions actually load and apply the rules.

Quote: "规范文档过于理想化 | Medium | Medium | 规范中提供具体代码示例和反例"

What must improve: Rerate impact to High. Add an adoption verification mechanism to the mitigation — for example: "Phase 1 完成后，在 3 个连续 AI 会话中验证 @rules/naming.md 被正确引用；若任一会话未引用，调整规则文件结构或 CLAUDE.md 引用方式。"

---

### Attack 3 — Problem Definition: Projection is unanchored; recurrence claim rests on one incident

Two related evidence gaps persist. First, "按当前迭代速度 main_items 预计增长至 500+" never states the current item count or the growth rate — without these, the 6-month timeline is unverifiable. Second, the entire bundle rationale depends on "AI 会话已出现过 snake_case JSON tag（model 层历史遗留）" — one incident. One incident does not establish a recurrence pattern; it could be an outlier. The proposal does not state how many AI sessions have run, how many introduced violations, or whether the rate is increasing.

Quote: "按当前迭代速度 main_items 预计增长至 500+" and "AI 会话已出现过 snake_case JSON tag（model 层历史遗留），同类风格漂移将持续累积"

What must improve: State the current item count and monthly growth rate to anchor the projection. For the recurrence claim, either cite additional incidents or reframe the argument: "even one undetected violation per quarter compounds over time" is a valid argument that does not require a high frequency claim.
