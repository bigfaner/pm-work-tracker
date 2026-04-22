---
proposal: docs/proposals/code-conventions/proposal.md
iteration: 7
score: 89/100
date: 2026-04-22
---

# Proposal Evaluation Report — Code Quality: Performance, Duplication & Readability

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 16 | 20 |
| Solution Clarity | 19 | 20 |
| Alternatives Analysis | 13 | 15 |
| Scope Definition | 14 | 15 |
| Risk Assessment | 14 | 15 |
| Success Criteria | 15 | 15 |
| Vague language penalty | -2 | — |
| **Total** | **89** | **100** |

---

## Changes Since Iteration 6

| Attack | Status | Notes |
|--------|--------|-------|
| [N] and [M] placeholders still unfilled | ⚠️ Partially addressed | Placeholders replaced with "约 200 条（估算值，未查询生产 DB）" and "月增约 50 条". No longer literal placeholders, but the note "(估算值，未查询生产 DB)" is a self-admission that the numbers are not from actual DB data. The projection is now anchored but to unverified estimates. |
| Alternative A's cons omit Phase 0 execution risks | ✅ Fixed | Cons now include: "Phase 0 本身存在执行风险：TableView 行为变更（全表加载→DB 分页）需对比测试验证结果一致性；UserRepo 接口扩展需同步更新调用方和测试 mock". Genuine fix. |
| Test update scope unresolved; TableView mitigation not actionable | ✅ Fixed | Scope now explicitly includes "UserRepo.FindByIDs 接口变更所需的调用方更新和测试 mock 更新". TableView mitigation now specifies: "编写集成测试，对同一数据集分别运行改动前后的 TableView，对比返回的 item IDs 和分页元数据（total、page、pageSize）一致". Both gaps resolved. |

Score improvement: 83 → 89. The two concrete structural fixes (Alternative A cons, test update scope + TableView mitigation) are genuine improvements. The remaining ceiling is held by the estimate-based projection, the unsubstantiated N+1 recurrence claim, and unanchored time estimates.

---

## Dimension Breakdown

### 1. Problem Definition — 16/20

**Problem stated clearly (6/7)**

Performance problems remain well-specified with file locations and line numbers. Duplication has exact counts with line references. "规范缺失" is still underspecified — "AI每次会话无法获得一致的风格指引" names the symptom but not the scope: how many rule categories are currently inconsistent? What is the current observable state of style drift? -1.

**Evidence provided (5/7)**

The N+1 example is concrete. The 62-line duplication has specific line ranges. The 21-duplicate breakdown is in the Problem section. GORM log observation is a real data point. The [N] and [M] placeholders have been replaced with "约 200 条" and "月增约 50 条" — an improvement over literal placeholders. However, the parenthetical "(估算值，未查询生产 DB)" is a self-admission that these numbers are not from actual DB data. The projection "6 个月后将超过 500 条" is now arithmetically anchored but to unverified inputs. A proposal that explicitly acknowledges its key urgency numbers are estimates rather than measurements cannot receive full marks on evidence. -2.

**Urgency justified (5/6)**

The no-delete-path argument is concrete. The GORM log observation is a real data point. The recurrence reframing ("规范缺失的风险不依赖违规频率") is a valid argument. The 6-month projection is now anchored to numbers, making it verifiable in principle — improvement from iteration 6. -1 because the anchor numbers are self-admitted estimates, not measurements.

---

### 2. Solution Clarity — 19/20

**Approach is concrete (7/7)**

Phase 0: specific method signatures and rename targets. Phase 1: exact file paths. Phase 2: names `tagliatelle` and ESLint rules. Phase 3: names specific components with page-level occurrence counts. Full marks.

**User-facing behavior described (7/7)**

Phase 0: GORM query log verification — concrete. Phase 1: behavioral verification via `@rules/naming.md` in a new session — concrete. Phase 2: non-zero exit codes for both golangci-lint and ESLint — concrete. Phase 3: "各组件在原页面中的内联 Dialog 替换为组件引用" — observable. Full marks.

**Distinguishes from alternatives (5/6)**

The bundle rationale is substantive: "Phase 0 修复的是实现 bug，Phases 1–3 修复的是产生 bug 的过程缺陷" is a real argument. The recurrence prevention logic is well-framed. -1 because the claim that AI sessions will reintroduce N+1 queries specifically is still asserted without evidence — the only cited incident is a snake_case tag violation, not a query pattern. The recurrence argument for naming violations is supported; the recurrence argument for query patterns is not.

---

### 3. Alternatives Analysis — 13/15

**At least 2 alternatives listed (5/5)**

Four alternatives including Do Nothing. Full marks.

**Pros/cons for each (4/5)**

Alternative A's cons now include execution risks — genuine fix. Alternative D's cons are honest and specific. Alternative B's pros ("后续改动减少重复文件触碰") remain unquantified — vague language penalty applied separately. Alternative C's cons ("单 PR 改动范围横跨性能、文档、lint、清理，难以 review，回归风险高") are real but not elaborated: what is the estimated PR size? How many files? -1.

**Rationale for chosen approach (4/5)**

The "implementation bug vs. process defect" framing is a legitimate and clear argument for bundling. The dependency chain (Phase 1 → Phase 2 → Phase 3) is well-argued. -1 because the claim that AI sessions will reintroduce N+1 queries specifically is asserted without evidence of that pattern recurring — the only cited incident is a naming violation.

---

### 4. Scope Definition — 14/15

**In-scope items are concrete (5/5)**

Phase 0 items are concrete. Phase 1 file paths are concrete. Phase 2 names specific lint rules for both Go and TypeScript. Phase 3 names specific components with occurrence counts. Phase 0 now explicitly includes "UserRepo.FindByIDs 接口变更所需的调用方更新和测试 mock 更新" — the gap from iteration 6 is resolved. Full marks.

**Out-of-scope explicitly listed (5/5)**

Includes CI/CD lint integration, bulk migration, business logic changes, new features, zcode plugin. The test update scope is now addressed in-scope. Full marks.

**Scope is bounded (4/5)**

T-shirt sizing per phase is present. Sequencing constraints are stated. Total target "3 个 sprint（约 3 週）" is a real boundary. -1 because the per-phase estimates ("约3–5天", "约1–2天", "约1天", "约5–7天") have no stated basis — no team size, no velocity reference, no complexity breakdown. A reader cannot validate whether Phase 0 at M = 3–5 days is realistic without knowing who is doing the work and at what pace.

---

### 5. Risk Assessment — 14/15

**Risks identified (5/5)**

Five risks listed, all meaningful. The lint adoption failure mode is present with a specific mitigation. Full marks.

**Likelihood + impact rated (4/5)**

"规范文档过于理想化" is rated Medium/High — fixed from iteration 5. `linkageMuMap` is rated Low/High — the likelihood of "Low" remains questionable: the code already has no delete path and is in active production use with a growing dataset. "Medium" would be more honest for a bug that is already confirmed to exist and has no remediation path in the current code. -1.

**Mitigations are actionable (5/5)**

The lint adoption mitigation is very specific: a concrete command with a threshold (>20 violations → add cleanup subtask). The `linkageMuMap` mitigation is actionable. The `TableView` risk mitigation is now specific: "编写集成测试，对同一数据集分别运行改动前后的 TableView，对比返回的 item IDs 和分页元数据（total、page、pageSize）一致" — genuine fix from iteration 6. Full marks.

---

### 6. Success Criteria — 15/15

**Criteria are measurable (5/5)**

Phase 0 criteria are binary and measurable. `linkageMuMap` has a specified verification method with concrete conditions. ESLint criterion has specific test commands. Full marks.

**Coverage is complete (5/5)**

Phase 0: 1:1 mapping with scope. Phase 1: section checklists for ARCHITECTURE.md and DECISIONS.md. Phase 2: both golangci-lint and ESLint covered. Phase 3: names specific components. The test update scope is now in-scope AND has a success criterion: "UserRepo.FindByIDs 所有调用方已更新，相关测试 mock 已同步扩展" — the gap from iteration 6 is resolved. Full marks.

**Criteria are testable (5/5)**

Phase 0: testable. Phase 1: behavioral test present — "在新 Claude Code 会话中执行 `@rules/naming.md`，规则内容出现在上下文中". Phase 2: testable with specific commands. Phase 3: testable with named components and "各组件在原页面中的内联 Dialog 替换为组件引用". Full marks.

---

## Vague Language Penalties

| Instance | Location | Penalty |
|----------|----------|---------|
| "后续改动减少重复文件触碰" (reduced file touching, unquantified) | Alternatives B, Pros | -2 |

Total penalty: **-2**

---

## Top Attacks

### Attack 1 — Problem Definition: Projection numbers are self-admitted estimates, not measurements

The [N] and [M] placeholders have been replaced, but the replacement values are explicitly flagged as estimates: "(估算值，未查询生产 DB)". The 6-month projection "按此速度 6 个月后将超过 500 条" is the primary urgency argument for the performance fixes. It is now arithmetically consistent but rests on unverified inputs. A proposal that acknowledges its key urgency numbers are not from actual data is asking reviewers to approve work based on a projection the author has not verified.

Quote: "当前 main_items 约 200 条（估算值，未查询生产 DB），月增约 50 条"

What must improve: Run `SELECT COUNT(*) FROM main_items` and `SELECT COUNT(*) FROM main_items WHERE created_at >= NOW() - INTERVAL '30 days'` against the actual database and replace the estimates with real numbers. If production DB access is unavailable, state that explicitly and use staging data — but do not present unverified estimates as evidence for urgency.

---

### Attack 2 — Alternatives Analysis / Solution Clarity: N+1 recurrence claim is asserted without evidence of that specific pattern

The bundle rationale depends on the claim that AI sessions without Phase 1–3 will reintroduce the same bugs Phase 0 fixes. The only cited incident of AI-introduced violations is a snake_case tag in the model layer — a naming violation. The proposal then extends this to claim AI sessions will also reintroduce N+1 queries and full-table memory pagination. These are structurally different failure modes: naming violations are syntactic and easy to introduce accidentally; N+1 queries and memory pagination require specific architectural decisions. The recurrence argument for naming violations is supported; the recurrence argument for query patterns is not.

Quote: "若只做 Phase 0，下一个 AI 会话在没有规范和 lint 约束的情况下仍会引入新的 N+1 查询或 snake_case tag"

What must improve: Either provide evidence that AI sessions have previously introduced N+1 queries or memory pagination patterns (e.g., a git log reference), or narrow the recurrence claim to naming violations only and reframe the Phase 0 justification as fixing existing bugs rather than preventing recurrence.

---

### Attack 3 — Scope Definition: Per-phase time estimates have no stated basis

The scope table assigns T-shirt sizes and day ranges to each phase (Phase 0: M, 约3–5天; Phase 1: S, 约1–2天; Phase 2: S, 约1天; Phase 3: L, 约5–7天) with no stated basis. No team size, no velocity reference, no complexity breakdown. The total "3 个 sprint（约 3 週）" target is a real boundary, but a reader cannot validate whether the per-phase estimates are realistic. Phase 0 in particular involves a semantic behavior change to `TableView` (full-table load → DB pagination), a new `UserRepo` interface method with caller and mock updates, and a concurrency change to `linkageMuMap` — "约3–5天" may be optimistic depending on test coverage gaps.

Quote: "Phase 0 | ... | M（约 3–5 天）"

What must improve: Add a one-line basis for each estimate: e.g., "based on single developer, ~4h/day coding time" or "based on similar refactor in [prior task]". Without a basis, the estimates are not falsifiable and cannot be used for planning.
