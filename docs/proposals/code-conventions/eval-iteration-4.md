---
proposal: docs/proposals/code-conventions/proposal.md
iteration: 4
score: 78/100
date: 2026-04-22
---

# Proposal Evaluation Report — Code Quality: Performance, Duplication & Readability

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 15 | 20 |
| Solution Clarity | 17 | 20 |
| Alternatives Analysis | 12 | 15 |
| Scope Definition | 12 | 15 |
| Risk Assessment | 12 | 15 |
| Success Criteria | 12 | 15 |
| Vague language penalty | -2 | — |
| **Total** | **78** | **100** |

---

## Changes Since Iteration 3

Three attacks from iteration 3 were addressed:

| Attack | Status | Notes |
|--------|--------|-------|
| AI session rule loading is claimed but not verifiable | ✅ Fixed | Phase 1 observable state now specifies behavioral verification: execute `@rules/naming.md` in a new session and verify rule content appears in context; test with a snake_case snippet and verify Claude Code flags the violation |
| Lint adoption failure mode is absent | ✅ Fixed | New risk row added: "Phase 2 lint 规则因存量违规过多而长期停留在 warn 模式" rated Medium/High with a specific mitigation command and threshold (>20 violations → add cleanup subtask) |
| Phase 0 independence undermines the bundle rationale | ✅ Fixed | Alternative A now explicitly argues: "Phase 0 修复的是实现 bug，Phases 1–3 修复的是产生 bug 的过程缺陷；若只做 Phase 0，下一个 AI 会话在没有规范和 lint 约束的情况下仍会引入新的 N+1 查询或 snake_case tag，Phase 0 的修复会被逐步侵蚀" |

Score improvement: 72 → 78. Remaining structural weaknesses are in success criteria coverage (ESLint gap, linkageMuMap verification), evidence precision, and Phase 3 deferred specificity.

---

## Dimension Breakdown

### 1. Problem Definition — 15/20

**Problem stated clearly (6/7)**

Performance problems remain well-specified with file locations and concrete descriptions. "规范缺失" is still underspecified — "AI每次会话无法获得一致的风格指引" names the symptom but not the scope of inconsistency. One historical snake_case incident is cited; it is unclear whether this is isolated or representative. -1.

**Evidence provided (4/7)**

The N+1 example is concrete. The GORM log observation is a real data point. However: "约60行相同的fetch+filter代码" still uses "约" (approximately) — the exact line count is verifiable and should be stated. The 21 duplicate count appears only in the success criteria section, not in the Problem section where it would serve as evidence. No query latency measurements, no response time data. -3.

**Urgency justified (5/6)**

The "Why now" section provides a no-delete-path argument, a GORM log observation, a 6-month projection (500+ items → 500+ mutex entries), and a historical AI session incident. The projection is concrete. -1 because the current data volume is never stated — "500+" is a projection without anchoring to today's item count or growth rate, making the timeline unverifiable.

---

### 2. Solution Clarity — 17/20

**Approach is concrete (6/7)**

Phase 0 is strong: specific method signatures, specific rename targets. Phase 1 lists exact file paths. Phase 2 names `tagliatelle` and the ESLint rules. Phase 3 "识别重复UI模式，抽取可复用组件（具体组件名在 Phase 3 开始前确定）" remains discovery work, not a deliverable. -1.

**User-facing behavior described (6/7)**

Significantly improved across all phases. Phase 0: GORM query log verification — concrete and verifiable. Phase 1: now specifies behavioral verification — "在新 Claude Code 会话中执行 `@rules/naming.md`，规则内容出现在上下文中（行为验证，而非仅检查文件存在）；用一段含 snake_case JSON tag 的代码片段提问，Claude Code 应指出违规并引用 naming.md 中的规则" — this is a genuine fix from iteration 3. Phase 2: golangci-lint non-zero exit code — concrete. Phase 3: grep count — concrete. Remaining gap: Phase 3's component extraction observable state is still deferred ("具体组件名在 Phase 3 开始前确定"), so the observable state for that deliverable cannot be evaluated now. -1.

**Distinguishes from alternatives (5/6)**

The bundle rationale is now substantive: "Phase 0 修复的是实现 bug，Phases 1–3 修复的是产生 bug 的过程缺陷" is a real argument, not assertion. The recurrence prevention logic ("Phase 0 的修复会被逐步侵蚀") is concrete and logical. -1 because the "AI will reintroduce bugs" claim rests on a single historical incident (one snake_case tag in model layer). One data point does not establish a pattern; the proposal does not state how frequently AI sessions have introduced violations or whether the rate is increasing.

---

### 3. Alternatives Analysis — 12/15

**At least 2 alternatives listed (5/5)**

Four alternatives including Do Nothing. Full marks.

**Pros/cons for each (3/5)**

Alternative D's cons are genuinely honest and specific — the strongest section. Alternative B's pros ("后续改动减少重复文件触碰") are still unquantified — vague language penalty applied separately. Alternative A's cons ("总工作量比单做 Phase 0 大；Phase 3 工期较长") are real but do not acknowledge the risks introduced by the performance fixes themselves (behavior changes, test gaps). Alternative C's cons are real but not elaborated. -2.

**Rationale for chosen approach (4/5)**

Substantially improved. The "implementation bug vs. process defect" framing is a legitimate and clear argument for bundling. The dependency chain (Phase 1 → Phase 2 → Phase 3) is well-argued. The recurrence prevention argument is concrete. -1 because the claim that "下一个 AI 会话...仍会引入新的 N+1 查询" is asserted without evidence of frequency — if AI sessions introduce violations rarely, the urgency of bundling is weaker than stated.

---

### 4. Scope Definition — 12/15

**In-scope items are concrete (4/5)**

Phase 0 items are concrete. Phase 1 file paths are concrete. Phase 2 names specific lint rules for both Go and TypeScript. Phase 3 "前端可复用组件抽取（≥3 个，名称在 Phase 3 开始前确定）" remains open-ended discovery work in the scope section — a deliverable whose content is unknown at proposal time is not a bounded scope item. -1.

**Out-of-scope explicitly listed (4/5)**

Includes CI/CD lint integration and bulk migration of existing violations. Still missing: whether existing tests are in scope for updates when Phase 0 changes repo interfaces (e.g., `UserRepo` gains `FindByIDs` — callers and test mocks may need updating). -1.

**Scope is bounded (4/5)**

T-shirt sizing per phase is present. Sequencing constraints are stated. Total target "3 个 sprint（约 3 週）" is a real boundary. -1 because the "约" estimates have no basis stated — Phase 0 is "约3–5天" but the estimate methodology is not given. A reader cannot validate whether M = 3–5 days is realistic without knowing team size or velocity.

---

### 5. Risk Assessment — 12/15

**Risks identified (5/5)**

Five risks listed, all meaningful. The lint adoption failure mode is now present with a specific mitigation. Full marks.

**Likelihood + impact rated (3/5)**

Ratings are provided but remain conservative. "规范文档过于理想化" is rated Medium/Medium — if the docs are never adopted by AI sessions, the entire Phase 1 investment is wasted, which is High impact, not Medium. The linkageMuMap risk is Low/High — concurrent bugs in production are typically catastrophic; "High" understates the severity for a global mutex map. -2.

**Mitigations are actionable (4/5)**

The lint adoption mitigation is now very specific: a concrete command with a threshold (>20 violations → add cleanup subtask). The linkageMuMap mitigation is actionable. The "规范文档过于理想化" mitigation ("规范中提供具体代码示例和反例") is actionable. Remaining gap: "改动前后对比测试；现有 repo 测试覆盖" for the TableView risk — what tests? Against what baseline? The existing test suite is not described, so "现有 repo 测试覆盖" is not actionable without knowing what it covers. -1.

---

### 6. Success Criteria — 12/15

**Criteria are measurable (4/5)**

Phase 0 criteria are mostly binary and measurable. "linkageMuMap 不再无限增长" still has no verification method — code review? load test? memory profiling? The criterion is stated but not operationalized. -1.

**Coverage is complete (4/5)**

Phase 0: 1:1 mapping with scope — good. Phase 1: section checklists for ARCHITECTURE.md and DECISIONS.md — good. Phase 2: golangci-lint criterion is specific, but there is no ESLint success criterion — the frontend lint work (API layer camelCase.ts naming, PascalCase component exports) has no measurable completion gate in the success criteria section. Phase 3: "至少抽取3个可复用前端UI组件（需在Phase 3开始前列出具体组件名）" defers specificity — the criterion is incomplete as written. -1.

**Criteria are testable (4/5)**

Phase 0: testable. Phase 1: now has a behavioral test — "在新 Claude Code 会话中执行 `@rules/naming.md`，规则内容出现在上下文中" — this is a genuine improvement from iteration 3. Phase 2: golangci-lint testable; ESLint criterion is absent, so the frontend lint work has no test. Phase 3: "至少抽取3个可复用前端UI组件" cannot be tested until component names are specified. -1.

---

## Vague Language Penalties

| Instance | Location | Penalty |
|----------|----------|---------|
| "后续改动减少重复文件触碰" (reduced file touching, unquantified) | Alternatives B, Pros | -2 |

Total penalty: **-2**

---

## Top Attacks

### Attack 1 — Success Criteria: ESLint has no measurable completion gate

Phase 2 adds ESLint rules for API layer file naming (camelCase.ts) and component export naming (PascalCase). The solution section describes this work. The success criteria section has no corresponding criterion — only golangci-lint is mentioned. The frontend lint work can be declared "done" with no verifiable gate.

Quote: *(success criteria section has no ESLint entry — only "golangci-lint 新增 `tagliatelle` 规则，`go lint ./...` 能检测出 snake_case JSON tag 违规")*

What must improve: Add a success criterion: "ESLint 对 API 层文件命名违规（非 camelCase.ts）和组件导出命名违规（非 PascalCase）返回非零退出码" with a specific test command (e.g., `npx eslint src/api/BadName.ts` exits non-zero). Without this, Phase 2 frontend lint has no completion gate.

---

### Attack 2 — Success Criteria: linkageMuMap criterion is unverifiable

"linkageMuMap 不再无限增长" is a success criterion with no specified verification method. The criterion describes a runtime property (memory growth over time) but proposes no way to check it at completion. Code review alone cannot verify absence of unbounded growth; a load test or memory profile is required.

Quote: "linkageMuMap 不再无限增长"

What must improve: Specify the verification method. For example: "After merging, run a load test that creates 100 MainItems and calls `EvaluateLinkage` for each; verify via pprof heap profile that `linkageMuMap` entry count does not exceed the number of active items" — or at minimum: "Code review confirms a deletion path exists in `EvaluateLinkage` or `sync.Map` is used with explicit cleanup." The criterion must be operationalized.

---

### Attack 3 — Problem Definition: Evidence is approximate and misplaced

Two evidence gaps persist. First, "约60行相同的fetch+filter代码" uses "约" when the exact line count is verifiable by reading the file — this is a precision failure, not an estimation problem. Second, the 21 duplicate count ("12 个 repo 层内联 `ErrRecordNotFound` 检查 + 1 个 `parsePagination` 重复函数 + 8 个 service 层日期解析样板") appears only in the success criteria section, not in the Problem section where it would serve as evidence for the duplication claim.

Quote: "约60行相同的fetch+filter代码" (Problem section) and "后端重复副本从当前 21 个...减少到 ≤10 个" (Success Criteria section only)

What must improve: Replace "约60行" with the exact line count. Move the 21-duplicate breakdown into the Problem section as evidence for the duplication claim. Evidence belongs in the Problem section, not scattered into success criteria.
