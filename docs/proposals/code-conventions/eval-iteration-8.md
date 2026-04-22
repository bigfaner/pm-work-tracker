---
proposal: docs/proposals/code-conventions/proposal.md
iteration: 8
score: 93/100
date: 2026-04-22
---

# Proposal Evaluation Report — Code Quality: Performance, Duplication & Readability

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 17 | 20 |
| Solution Clarity | 20 | 20 |
| Alternatives Analysis | 14 | 15 |
| Scope Definition | 15 | 15 |
| Risk Assessment | 14 | 15 |
| Success Criteria | 15 | 15 |
| Vague language penalty | -2 | — |
| **Total** | **93** | **100** |

---

## Changes Since Iteration 7

| Attack | Status | Notes |
|--------|--------|-------|
| Projection numbers are self-admitted estimates, not measurements | ⚠️ Partially addressed | Changed from "(估算值，未查询生产 DB)" to "(基于已完成 feature 数 × 每 feature 平均 item 数的保守估算，反映观测到的开发速度)". The methodology is now described, which is a genuine improvement. However, the inputs are still not stated: how many completed features? What is the average item count per feature? The estimate is now methodologically grounded but not independently verifiable. |
| N+1 recurrence claim asserted without evidence of that specific pattern | ✅ Fixed | The claim "仍会引入新的 N+1 查询或 snake_case tag" has been narrowed. Alternative A now reads "同类命名违规会在 Phase 0 修复后的新会话中重新出现", and Alternative A' similarly scopes the recurrence risk to naming violations only. The N+1 recurrence assertion has been dropped. Genuine fix. |
| Per-phase time estimates have no stated basis | ✅ Fixed | A new line added: "工期估算基准：单人开发，有效编码时间约 4h/天." The basis is now stated. Genuine fix. |

Score improvement: 89 → 93. Two of three attacks fully resolved. The remaining ceiling is held by the projection methodology (inputs unstated), the unquantified Alternative B pro, and the `linkageMuMap` likelihood rating.

---

## Dimension Breakdown

### 1. Problem Definition — 17/20

**Problem stated clearly (6/7)**

Performance problems remain well-specified with file locations and line numbers. Duplication has exact counts with line references. "规范缺失" is still underspecified — "AI每次会话无法获得一致的风格指引" names the symptom but not the scope: how many rule categories are currently inconsistent? What is the current observable state of style drift? -1.

**Evidence provided (6/7)**

The N+1 example is concrete. The 62-line duplication has specific line ranges. The 21-duplicate breakdown is in the Problem section. GORM log observation is a real data point. The projection methodology has improved: "(基于已完成 feature 数 × 每 feature 平均 item 数的保守估算，反映观测到的开发速度)" describes the estimation approach rather than just admitting it is an estimate. This is a genuine improvement over iteration 7. However, the inputs are not stated — the reader cannot verify "约 200 条" without knowing how many completed features and what the assumed average item count is. The estimate is methodologically described but not independently verifiable. -1.

**Urgency justified (5/6)**

The no-delete-path argument is concrete. The GORM log observation is a real data point. The recurrence reframing ("规范缺失的风险不依赖违规频率") is a valid argument. The 6-month projection is now anchored to a described methodology — improvement from iteration 7. -1 because the methodology inputs are still not stated, leaving the projection unverifiable in practice.

---

### 2. Solution Clarity — 20/20

**Approach is concrete (7/7)**

Phase 0: specific method signatures and rename targets. Phase 1: exact file paths. Phase 2: names `tagliatelle` and ESLint rules. Phase 3: names specific components with page-level occurrence counts. Full marks.

**User-facing behavior described (7/7)**

Phase 0: GORM query log verification — concrete. Phase 1: behavioral verification via `@rules/naming.md` in a new session — concrete. Phase 2: non-zero exit codes for both golangci-lint and ESLint — concrete. Phase 3: "各组件在原页面中的内联 Dialog 替换为组件引用" — observable. Full marks.

**Distinguishes from alternatives (6/6)**

The bundle rationale is substantive. The recurrence prevention logic is now properly scoped to naming violations, which are supported by the cited incident (snake_case tag in model layer). The "implementation bug vs. process defect" framing is a real argument. The N+1 recurrence claim has been dropped from the rationale. Full marks.

---

### 3. Alternatives Analysis — 14/15

**At least 2 alternatives listed (5/5)**

Four alternatives including Do Nothing. Full marks.

**Pros/cons for each (4/5)**

Alternative A's cons include execution risks — genuine. Alternative D's cons are honest and specific. Alternative A' cons are now properly scoped to naming violation recurrence. Alternative B's pros ("后续改动减少重复文件触碰") remain unquantified — vague language penalty applied separately. Alternative C's cons ("单 PR 改动范围横跨性能、文档、lint、清理，难以 review，回归风险高") are real but not elaborated: what is the estimated PR size? How many files? -1.

**Rationale for chosen approach (5/5)**

The "implementation bug vs. process defect" framing is a legitimate and clear argument for bundling. The dependency chain (Phase 1 → Phase 2 → Phase 3) is well-argued. The recurrence argument is now properly supported by the cited naming violation incident. Full marks.

---

### 4. Scope Definition — 15/15

**In-scope items are concrete (5/5)**

Phase 0 items are concrete. Phase 1 file paths are concrete. Phase 2 names specific lint rules for both Go and TypeScript. Phase 3 names specific components with occurrence counts. `UserRepo.FindByIDs` caller and mock updates are explicitly included. Full marks.

**Out-of-scope explicitly listed (5/5)**

Includes CI/CD lint integration, bulk migration, business logic changes, new features, zcode plugin. Full marks.

**Scope is bounded (5/5)**

T-shirt sizing per phase is present. Sequencing constraints are stated. Total target "3 个 sprint（约 3 週）" is a real boundary. Per-phase estimates now have a stated basis: "单人开发，有效编码时间约 4h/天" — the gap from iteration 7 is resolved. Full marks.

---

### 5. Risk Assessment — 14/15

**Risks identified (5/5)**

Five risks listed, all meaningful. Full marks.

**Likelihood + impact rated (4/5)**

"规范文档过于理想化" is rated Medium/High. `TableView` is rated Medium/High. `linkageMuMap` is rated Low/High — the likelihood of "Low" remains questionable: the bug is already confirmed to exist (no delete path in active production code with a growing dataset). A confirmed, already-present bug with no remediation path in the current code is not "Low" likelihood of causing impact — "Medium" would be more honest. -1.

**Mitigations are actionable (5/5)**

The lint adoption mitigation is very specific: a concrete command with a threshold (>20 violations → add cleanup subtask). The `linkageMuMap` mitigation is actionable. The `TableView` risk mitigation is specific with concrete verification conditions. Full marks.

---

### 6. Success Criteria — 15/15

**Criteria are measurable (5/5)**

Phase 0 criteria are binary and measurable. `linkageMuMap` has a specified verification method with concrete conditions. ESLint criterion has specific test commands. Full marks.

**Coverage is complete (5/5)**

Phase 0: 1:1 mapping with scope. Phase 1: section checklists for ARCHITECTURE.md and DECISIONS.md. Phase 2: both golangci-lint and ESLint covered. Phase 3: names specific components. Test update scope is in-scope and has a success criterion. Full marks.

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

### Attack 1 — Problem Definition: Projection methodology described but inputs unstated

The parenthetical changed from "(估算值，未查询生产 DB)" to "(基于已完成 feature 数 × 每 feature 平均 item 数的保守估算，反映观测到的开发速度)". This is a genuine improvement — the estimation approach is now described. However, the inputs are not stated: how many completed features? What is the assumed average item count per feature? A reader cannot verify "约 200 条" or "月增约 50 条" without these numbers. The 6-month projection "按此速度 6 个月后将超过 500 条" is the primary urgency argument for the performance fixes, and it rests on inputs that are described but not disclosed.

Quote: "当前 main_items 约 200 条（基于已完成 feature 数 × 每 feature 平均 item 数的保守估算，反映观测到的开发速度），月增约 50 条"

What must improve: State the actual inputs: e.g., "基于 N 个已完成 feature，每 feature 平均 M 个 item，当前估算约 200 条". If production DB access is available, replace with `SELECT COUNT(*) FROM main_items`. If not, state the specific feature count and average used so the estimate is independently verifiable.

---

### Attack 2 — Alternatives Analysis: Alternative B pros remain unquantified

"后续改动减少重复文件触碰" is the sole pro listed for Alternative B. It is unquantified: how many fewer files would be touched per future change? By what percentage? Without a number, this pro cannot be weighed against the stated con (N+1 queries and full-table memory pagination persist). The vague language penalty is applied, but the underlying weakness is that the trade-off analysis for Alternative B is asymmetric — the con is specific ("响应时间线性恶化") while the pro is vague.

Quote: "先消除 21 个重复副本，后续改动减少重复文件触碰"

What must improve: Quantify the pro or replace it with a concrete claim. For example: "21 个重复副本分布在约 X 个文件中，清理后每次 CRUD 变更预计减少触碰 Y 个文件". If the number is not known, say so — an honest "pro is hard to quantify" is better than an unverifiable claim.

---

### Attack 3 — Risk Assessment: `linkageMuMap` likelihood rated Low when bug is already confirmed

`linkageMuMap` is rated Low/High. The likelihood rating of "Low" is inconsistent with the evidence: the Problem section states the code has no delete path and the map grows without bound in active production use. This is not a hypothetical risk — it is a confirmed, already-present defect. The mitigation ("改用 `sync.Map`，不改锁语义") is actionable, but the likelihood rating understates the actual exposure. A confirmed bug with no remediation path in production code should be rated Medium, not Low.

Quote: "`linkageMuMap` 改动引入并发 bug | Low | High"

What must improve: Change likelihood from Low to Medium to reflect that the underlying condition (unbounded growth) is already confirmed to exist in production, not merely possible. The risk being assessed is the mitigation introducing a concurrent bug — that is Low — but the current framing conflates the mitigation risk with the original defect risk.
