---
proposal: docs/proposals/code-conventions/proposal.md
iteration: 6
score: 83/100
date: 2026-04-22
---

# Proposal Evaluation Report — Code Quality: Performance, Duplication & Readability

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 14 | 20 |
| Solution Clarity | 19 | 20 |
| Alternatives Analysis | 12 | 15 |
| Scope Definition | 13 | 15 |
| Risk Assessment | 13 | 15 |
| Success Criteria | 14 | 15 |
| Vague language penalty | -2 | — |
| **Total** | **83** | **100** |

---

## Changes Since Iteration 5

| Attack | Status | Notes |
|--------|--------|-------|
| Phase 3 component work is unbounded across three dimensions | ✅ Fixed | Components are now named with specific locations: `StatusChangeConfirmDialog` (MainItemDetailPage ×2, SubItemDetailPage ×1), `ProgressAppendDialog` (3 pages), `SubItemFormDialog` (merge create/edit). Scope, solution, and success criteria all now reference concrete names. |
| "规范文档过于理想化" impact underrated | ✅ Fixed | Impact rerated to High. Mitigation now includes adoption verification: "Phase 1 完成后，在 3 个连续 AI 会话中验证 @rules/naming.md 被正确引用；若任一会话未引用，调整规则文件结构或 CLAUDE.md 引用方式" |
| Projection unanchored; recurrence claim rests on one incident | ⚠️ Partially addressed | Recurrence claim is reframed well: "规范缺失的风险不依赖违规频率——即使每季度仅有一次未被发现的违规，随迭代累积也会造成风格漂移" — this is a valid argument that does not require frequency data. However, the projection placeholders [N] and [M] remain literally unfilled in the submitted text. |

Score improvement: 80 → 83. The Phase 3 component naming is a genuine fix that resolves the largest structural gap. The remaining score ceiling is held down by the unfilled projection placeholders, the unquantified Alternative B pros, and the missing test-update scope boundary.

---

## Dimension Breakdown

### 1. Problem Definition — 14/20

**Problem stated clearly (6/7)**

Performance problems remain well-specified with file locations and line numbers. Duplication has exact counts with line references. "规范缺失" is still underspecified — "AI每次会话无法获得一致的风格指引" names the symptom but not the scope: how many rule categories are inconsistent? The recurrence reframing ("即使每季度仅有一次未被发现的违规") is a valid argument that no longer requires frequency data. -1.

**Evidence provided (4/7)**

The N+1 example is concrete. The 62-line duplication has specific line ranges. The 21-duplicate breakdown is in the Problem section. GORM log observation is a real data point. Critical failure: the 6-month projection reads "当前 main_items 约 [N] 条，月增约 [M] 条（提交前以实际 DB 数据填入）" — these placeholders were explicitly called out in iteration 5 and remain unfilled in the submitted proposal. The note "(提交前以实际 DB 数据填入)" shows awareness but not action. A proposal with literal placeholder text in its evidence section is not ready for evaluation. -3.

**Urgency justified (4/6)**

The no-delete-path argument is concrete. The GORM log observation is a real data point. The recurrence reframing is now valid. -2 because the 6-month projection ("按此速度 6 个月后将超过 500 条") is still unanchored — [N] and [M] are unfilled, so the timeline remains unverifiable.

---

### 2. Solution Clarity — 19/20

**Approach is concrete (7/7)**

Phase 0: specific method signatures and rename targets. Phase 1: exact file paths. Phase 2: names `tagliatelle` and ESLint rules. Phase 3: now names specific components with page-level occurrence counts — this is a genuine fix. Full marks.

**User-facing behavior described (7/7)**

Phase 0: GORM query log verification — concrete. Phase 1: behavioral verification via `@rules/naming.md` in a new session — concrete. Phase 2: non-zero exit codes for both golangci-lint and ESLint — concrete. Phase 3: "各组件在原页面中的内联 Dialog 替换为组件引用" — observable. Full marks.

**Distinguishes from alternatives (5/6)**

The bundle rationale is substantive: "Phase 0 修复的是实现 bug，Phases 1–3 修复的是产生 bug 的过程缺陷" is a real argument. The recurrence prevention logic is now reframed to not require frequency data. -1 because the claim that AI sessions will reintroduce N+1 queries specifically (not just naming violations) is still asserted without evidence — the only cited incident is a snake_case tag, not a query pattern.

---

### 3. Alternatives Analysis — 12/15

**At least 2 alternatives listed (5/5)**

Four alternatives including Do Nothing. Full marks.

**Pros/cons for each (3/5)**

Alternative D's cons are genuinely honest and specific. Alternative B's pros ("后续改动减少重复文件触碰") remain unquantified — vague language penalty applied separately. Alternative A's cons ("总工作量比单做 Phase 0 大；Phase 3 工期较长") do not acknowledge the risks introduced by the Phase 0 performance fixes themselves: behavior changes in `TableView`, potential test gaps, interface changes to `UserRepo` that require caller and mock updates. Alternative C's cons are real but not elaborated. -2.

**Rationale for chosen approach (4/5)**

The "implementation bug vs. process defect" framing is a legitimate and clear argument for bundling. The dependency chain (Phase 1 → Phase 2 → Phase 3) is well-argued. -1 because the claim that AI sessions will reintroduce N+1 queries is asserted without evidence of that specific pattern recurring.

---

### 4. Scope Definition — 13/15

**In-scope items are concrete (5/5)**

Phase 0 items are concrete. Phase 1 file paths are concrete. Phase 2 names specific lint rules for both Go and TypeScript. Phase 3 now names specific components with occurrence counts — fixed from iteration 5. Full marks.

**Out-of-scope explicitly listed (4/5)**

Includes CI/CD lint integration, bulk migration, business logic changes, new features, zcode plugin. Still missing: whether existing tests are in scope for updates when Phase 0 changes the `UserRepo` interface (`FindByIDs` is a new method — callers and test mocks may need updating). This is a concrete consequence of Phase 0 that is neither in-scope nor explicitly out-of-scope. -1.

**Scope is bounded (4/5)**

T-shirt sizing per phase is present. Sequencing constraints are stated. Total target "3 个 sprint（约 3 週）" is a real boundary. -1 because the "约" estimates have no basis stated — Phase 0 is "约3–5天" but the estimate methodology is not given. A reader cannot validate whether M = 3–5 days is realistic without knowing team size or velocity.

---

### 5. Risk Assessment — 13/15

**Risks identified (5/5)**

Five risks listed, all meaningful. The lint adoption failure mode is present with a specific mitigation. Full marks.

**Likelihood + impact rated (4/5)**

"规范文档过于理想化" is now rated Medium/High — fixed. `linkageMuMap` is rated Low/High — the likelihood of "Low" remains questionable: the code already has no delete path and is in active production use; "Medium" would be more honest. -1.

**Mitigations are actionable (4/5)**

The lint adoption mitigation is very specific: a concrete command with a threshold (>20 violations → add cleanup subtask). The `linkageMuMap` mitigation is actionable. "规范文档过于理想化" now has an adoption verification mechanism — fixed. Remaining gap: "改动前后对比测试；现有 repo 测试覆盖" for the `TableView` risk — what tests? Against what baseline? The existing test suite is not described, so "现有 repo 测试覆盖" is not actionable without knowing what it covers. -1.

---

### 6. Success Criteria — 14/15

**Criteria are measurable (5/5)**

Phase 0 criteria are binary and measurable. `linkageMuMap` has a specified verification method with concrete conditions. ESLint criterion has specific test commands. Full marks.

**Coverage is complete (4/5)**

Phase 0: 1:1 mapping with scope. Phase 1: section checklists for ARCHITECTURE.md and DECISIONS.md. Phase 2: both golangci-lint and ESLint covered. Phase 3: now names specific components — fixed. Remaining gap: no criterion for test updates when Phase 0 changes `UserRepo` interface. -1.

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

### Attack 1 — Problem Definition: [N] and [M] placeholders are still unfilled after explicit callout

The 6-month projection was identified as unanchored in iteration 5. The author's response was to add a parenthetical note "(提交前以实际 DB 数据填入)" but never fill in the actual values. The proposal now contains literal placeholder text in its evidence section.

Quote: "当前 main_items 约 [N] 条，月增约 [M] 条（提交前以实际 DB 数据填入），按此速度 6 个月后将超过 500 条"

What must improve: Fill in the actual current item count and monthly growth rate from the database before submitting. A proposal with unfilled placeholders in its evidence section cannot be approved — the projection is the primary urgency argument for the performance fixes, and it has no anchor.

---

### Attack 2 — Alternatives Analysis: Alternative A's cons omit the risks introduced by Phase 0 itself

The cons for the recommended approach list only "总工作量比单做 Phase 0 大；Phase 3 工期较长". This ignores the risks that Phase 0 introduces: `TableView` behavior changes (full-table load → DB pagination is a semantic change, not just a performance change), `UserRepo` interface expansion (callers and test mocks need updating), and the `linkageMuMap` concurrency change. An honest alternatives analysis should acknowledge that the recommended approach carries execution risk, not just schedule risk.

Quote: "总工作量比单做 Phase 0 大；Phase 3 工期较长" (Alternative A, Cons)

What must improve: Add at least one execution risk to Alternative A's cons: e.g., "Phase 0 的 TableView 改动涉及行为变更，需要对比测试验证结果一致性；UserRepo 接口扩展需同步更新调用方和测试 mock"

---

### Attack 3 — Scope / Risk: Test update scope is unresolved; TableView mitigation is not actionable

Two related gaps persist. First, Phase 0 adds `UserRepo.FindByIDs` — this is an interface change that requires updating callers and test mocks, but the scope table neither includes nor excludes this work. Second, the `TableView` risk mitigation says "现有 repo 测试覆盖" without specifying what those tests cover or what baseline they establish. If the existing tests do not cover the full-table-load behavior, the mitigation provides no actual safety net.

Quote: "改动前后对比测试；现有 repo 测试覆盖" (Risk: TableView 改 DB 分页后行为差异, Mitigation)

What must improve: Add to Out-of-Scope or In-Scope: "Phase 0 接口变更（UserRepo.FindByIDs）所需的调用方和 mock 更新". For the TableView mitigation, specify what the comparison test covers: e.g., "编写集成测试，对同一数据集分别运行改动前后的 TableView，对比返回的 item IDs 和分页元数据一致"
