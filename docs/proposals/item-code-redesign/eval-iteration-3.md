# Proposal Evaluation Report

**Proposal**: Item Code Redesign (`item-code-redesign`)
**Iteration**: 3
**Date**: 2026-04-22
**Overall Score**: 91/100

---

## Dimension Scores

### 1. Problem Definition: 15/20

| Criterion | Score | Notes |
|-----------|-------|-------|
| Problem stated clearly | 6/7 | Three distinct problems with code citations and line references. Problem 1 (team-agnostic MI-prefix) and Problem 2 (SubItem has no Code field, frontend uses runtime workaround) are well-framed. Problem 3 ("Team model lacks abbreviation field") is more of a prerequisite/consequence of Problem 1 than an independent problem -- it inflates the problem count without adding independent weight. |
| Evidence provided | 4/7 | Code citations are strong: `main_item_repo.go:102` for hardcoded prefix, `sub_item.go` for missing field, `MainItemDetailPage.tsx:407` for runtime workaround. Problem 2's citation of `SI-${itemId}-${subId}` is excellent evidence of an unmet need. However, there is still zero user feedback, no support tickets, no quantitative data on how many teams exist, how many items share the MI- prefix, or whether anyone has actually been confused. The "Do Nothing" alternative asserts "已造成辨识困难" without substantiating it. The internal testing context partially explains the absence, but claiming difficulty without evidence is a persistent gap. |
| Urgency justified | 5/6 | Well-justified. The "Do Nothing" verdict names downstream dependencies: "子事项编码是后续功能（周报关联、进度追踪）的前置依赖." The internal-testing-about-to-support-multi-team framing makes timing concrete. |

**Deductions**:
- -1: Problem 3 is not an independent problem; it is a consequence of Problem 1. Presenting it as a separate item overstates the problem landscape.
- -2: No user feedback or quantitative evidence for any claimed problem. "已造成辨识困难" is asserted, not demonstrated.
- -2: SubItem code need is evidenced only by the runtime workaround, not by a user workflow or feature request that requires stable sub-item codes. Who needs to reference sub-items, in what scenario?

### 2. Solution Clarity: 19/20

| Criterion | Score | Notes |
|-----------|-------|-------|
| Approach is concrete | 7/7 | Encoding format table with format, example, composition per level. Model changes specify exact field names, varchar sizes, index types. NextCode/NextSubCode functions named. Data migration procedure described step-by-step. A developer could implement directly from this. |
| User-facing behavior described | 6/7 | Major improvement from iteration 2. The "用户可见行为" section now provides: (1) Team Code management UX with input field location, validation rules, specific error messages for format violation and duplication, display in team list; (2) per-page table for 5 pages showing current behavior vs. changed behavior with component names and rendering details; (3) explicit search behavior description with examples. Minor gap: no description of Team Code editing after creation -- can an admin change a team code? The "核心规则" say codes are snapshot, but the admin UX for editing (or disabling edit) is not described. |
| Distinguishes from alternatives | 6/6 | Design decisions table differentiates on four axes: snapshot vs. real-time, per-team vs. global, 5-digit vs. 4-digit, direct switch vs. dual format. Each has explicit rationale. The "选中方案的代价" paragraph now honestly lists costs (more complex code generation, larger varchar, new column + index, data migration). |

**Deductions**:
- -1: Team Code editing UX after creation is not described. Can admins change a team code? If so, what happens in the UI? If not, is the field disabled/readonly after creation?

### 3. Alternatives Analysis: 15/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| At least 2 alternatives listed | 5/5 | Four alternatives: Do Nothing, UUID, Global Sequence, chosen approach. "Do Nothing" is included. |
| Pros/cons for each | 5/5 | Each alternative has honest, substantive pros and cons. UUID acknowledges genuine distributed advantages. Global sequence correctly identifies that global contention is worse than per-team. Do Nothing has concrete cons tied to downstream dependencies. |
| Rationale for chosen approach | 5/5 | Design decisions table provides per-decision rationale. The "选中方案的代价" paragraph explicitly lists costs, addressing the iteration-2 critique about not acknowledging trade-offs. |

**Deductions**: None.

### 4. Scope Definition: 13/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| In-scope items are concrete | 5/5 | Backend changes specific (field names, types, sizes, function names). Frontend changes describe 5 pages with component-level rendering details. Migration procedure described. Team Code CRUD scoped with UX description. |
| Out-of-scope explicitly listed | 5/5 | Four explicit items: custom prefixes, batch rename, deleted code recycling, cross-team uniqueness. Meaningful exclusions that prevent scope creep. |
| Scope is bounded | 3/5 | Still no timeline or effort estimate. The proposal does not state whether this is a 2-day or 1-week change. Migration has a procedural description but no estimate of affected row counts or migration duration. The "内测阶段" context partially bounds scope (small data), but without an effort estimate, scope remains open-ended. |

**Deductions**:
- -2: No effort estimation or timeline. Is this a half-day change, a 2-day sprint, or a week-long effort? Without effort bounding, the scope cannot be planned against.

### 5. Risk Assessment: 15/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| Risks identified | 5/5 | Five meaningful risks: data migration failure, NextCode race condition, NextSubCode concurrency, old code reference invalidation, Team Code missing/invalid. The iteration-2 gaps (SubItem migration risk, Team Code edge cases) are now addressed. |
| Likelihood + impact rated | 5/5 | Ratings are honest and varied: race conditions are medium likelihood (realistic for concurrent code generation), migration failure is low/high, old reference is low/low, Team Code invalid is low/high. Not everything is "low likelihood, high impact." |
| Mitigations are actionable | 5/5 | Specific mitigations: transaction-wrapped migration with rollback script, full database backup, unique index + service-layer retry (max 3 times), regex `^[A-Za-z]{2,6}$` + CHECK constraint + unique index as triple enforcement. The NextSubCode mitigation now correctly states "唯一索引 + 重试是实际防线，不依赖并发量低的假设" -- directly addressing the iteration-2 critique. Long-term alternatives (SELECT FOR UPDATE, database sequences) are named for future improvement. |

**Deductions**: None.

### 6. Success Criteria: 14/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| Criteria are measurable | 5/5 | Eleven criteria with "可测试条件" column. Migration criteria include executable SQL queries. Validation criterion specifies exact invalid inputs and expected HTTP responses. Frontend criteria specify exact component rendering expectations (e.g., "Badge 渲染 `FEAT-00001`"). Search criteria specify exact input/output pairs. |
| Coverage is complete | 4/5 | Covers: Team Code field, Team Code validation (addressed iteration-2 gap), main item format, sub-item format, immutability, SubItem migration with SQL (addressed iteration-2 gap), MainItem migration, frontend display (main + sub), frontend search, existing tests. Minor gap: no criterion for NextCode retry behavior under concurrent writes -- the risk section identifies this but no criterion verifies that the retry mechanism works. |
| Criteria are testable | 5/5 | Each criterion can be verified with a test or checklist. SQL queries are directly executable. Frontend criteria specify exact UI components and expected rendering. Validation criteria specify exact inputs and expected HTTP status codes. |

**Deductions**:
- -1: No success criterion for concurrent NextCode retry behavior. The risk section identifies race conditions and prescribes retry, but no criterion verifies that "creating 2 items concurrently under the same team results in 2 distinct codes without error."

---

## Summary of Attacks

1. **Problem Definition**: Evidence remains code-only with zero user feedback or quantitative data. "已造成辨识困难" is asserted in the Do Nothing alternative without substantiation. The proposal would be stronger with even a single user quote, support ticket reference, or data point like "3 teams currently share the MI- prefix across 47 items." Additionally, SubItem code need is evidenced only by a runtime workaround in code, not by a user workflow that requires stable sub-item identifiers.

2. **Scope Definition**: No effort estimate or timeline anywhere in the proposal. The scope lists concrete deliverables but does not bound them in time or effort. "Is this a 2-day change or a 1-week change?" remains unanswered. Without this, the scope cannot be planned, resourced, or tracked.

3. **Success Criteria**: No criterion for verifying that the concurrent code generation retry mechanism works. The risk section identifies race conditions as medium likelihood with high impact and prescribes unique-index + retry, but no success criterion tests this defense. A criterion like "concurrent creation of 2 items under the same team produces 2 distinct codes without error" would close this gap.

## Improvement Summary vs. Iteration 2

| Dimension | Iter 2 | Iter 3 | Delta |
|-----------|--------|--------|-------|
| Problem Definition | 12 | 15 | +3 |
| Solution Clarity | 15 | 19 | +4 |
| Alternatives Analysis | 14 | 15 | +1 |
| Scope Definition | 11 | 13 | +2 |
| Risk Assessment | 12 | 15 | +3 |
| Success Criteria | 11 | 14 | +3 |
| **Total** | **75** | **91** | **+16** |

Major improvements: Solution Clarity (+4) with comprehensive user-facing behavior section covering Team Code management UX, per-page rendering table, and search behavior; Risk Assessment (+3) adding SubItem migration risk and Team Code validation risk, plus fixing the "low concurrency" assumption; Success Criteria (+3) adding SubItem migration criterion with SQL and Team Code validation criterion with specific invalid inputs.

Persistent weaknesses: Problem evidence still relies solely on code citations without user feedback; scope still lacks effort estimation.
