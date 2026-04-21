# Proposal Evaluation Report

**Proposal**: Item Code Redesign (`item-code-redesign`)
**Iteration**: 2
**Date**: 2026-04-22
**Overall Score**: 75/100

---

## Dimension Scores

### 1. Problem Definition: 12/20

| Criterion | Score | Notes |
|-----------|-------|-------|
| Problem stated clearly | 5/7 | Three distinct problems listed with concrete descriptions. However, "SubItem 没有任何人类可读的标识符" still lacks a user-workflow framing -- who needs to reference a sub-item by code, in what scenario, and what breaks without it? This is a gap, not a self-evident truth. |
| Evidence provided | 2/7 | No user feedback, no support tickets, no quantitative data. The "Do Nothing" alternative adds a scenario ("需要额外说明是哪个团队") but this is asserted, not evidenced. How many teams exist? How many items share the MI- prefix? Has anyone actually complained? None of this is provided. |
| Urgency justified | 5/6 | Significantly improved from iteration 1. The "Do Nothing" verdict explicitly names downstream dependencies: "子事项编码是后续功能（周报关联、进度追踪）的前置依赖". This makes the urgency concrete and justifies the timing. |

**Deductions**:
- -2: "无法从编码识别所属团队" is still asserted without evidence that users find this confusing in practice.
- -3: Zero quantitative or qualitative evidence. No data on team count, item count, user complaints, or support tickets related to the current code format.
- -3: SubItem lacking a code remains unmoored from a specific user workflow. What feature request or user action requires sub-item codes? Name it.

### 2. Solution Clarity: 15/20

| Criterion | Score | Notes |
|-----------|-------|-------|
| Approach is concrete | 7/7 | Encoding format table is excellent with format, example, and composition for each level. Model changes specify exact field names, varchar sizes, and index types. NextCode/NextSubCode functions are named. A developer could implement directly from this. |
| User-facing behavior described | 3/7 | Internal model changes are well-specified, but the end-user experience is largely omitted. "更新各页面中编码的展示和搜索" is a scope list, not a behavior description. What does a user see differently on ItemViewPage? How does search behave with the new format? What happens when an admin creates or edits a Team Code -- what validation feedback do they get? None of this is described. |
| Distinguishes from alternatives | 5/6 | The design decisions table effectively differentiates the chosen approach: snapshot vs. real-time, per-team vs. global, 5-digit vs. 4-digit, direct switch vs. dual format. Each decision has an explicit rationale. |

**Deductions**:
- -2: "更新各页面（ItemViewPage、TableViewPage、WeeklyViewPage、MainItemDetailPage）中编码的展示和搜索" -- this names pages but describes no user-facing behavior. What renders differently? Is the code a link? Clickable? Copyable?
- -2: Team Code management UX is entirely absent. This is a core new user-facing feature (admins creating/editing team codes) with zero description of the interaction flow.

### 3. Alternatives Analysis: 14/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| At least 2 alternatives listed | 5/5 | Four alternatives: Do Nothing, UUID, Global Sequence, and the chosen approach. "Do nothing" is included. Comprehensive coverage of the design space. |
| Pros/cons for each | 5/5 | Each alternative has honest, substantive pros and cons. The UUID alternative acknowledges genuine advantages (distributed, no race condition) alongside its fatal disadvantage. The global sequence analysis is particularly insightful, correctly identifying that global contention is worse than per-team. |
| Rationale for chosen approach | 4/5 | The design decisions table provides explicit rationale for each key decision. However, the final verdict is somewhat tautological: "唯一同时满足可读性、团队辨识、子事项编码三个需求的方案" -- this asserts the chosen approach meets all requirements without analyzing its costs or trade-offs. |

**Deductions**:
- -1: The chosen approach's verdict frames it as meeting all requirements without acknowledging its own costs: more complex code generation logic, larger varchar columns, snapshot complexity, no backward compatibility. A honest alternatives analysis should acknowledge what the chosen approach sacrifices.

### 4. Scope Definition: 11/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| In-scope items are concrete | 4/5 | Backend changes are specific (field names, types, sizes, function names). Frontend changes name 5 pages. Migration is listed. |
| Out-of-scope explicitly listed | 4/5 | Four explicit items: custom prefixes, batch rename, deleted code recycling, cross-team uniqueness. |
| Scope is bounded | 3/5 | No timeline or effort estimate. Migration is a single line item despite being the highest-risk scope element. What SQL runs? On how many rows? What is the rollback procedure? |

**Deductions**:
- -1: Data migration scope is still underspecified. "直接切换，不需要兼容旧 MI-XXXX 格式" is a strategy, not a scope description. How many existing items need codes rewritten? How many SubItems need codes generated from scratch?
- -1: No effort estimation or timeline. Is this a 2-day change or a 1-week change? Without effort bounding, scope is open-ended.
- -2: Missing scope item: Team Code CRUD operations. The proposal adds a `Code` field to Team but never scopes the admin interface for managing it.

### 5. Risk Assessment: 12/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| Risks identified | 4/5 | Four meaningful risks: data migration failure, NextCode race condition, NextSubCode concurrency, old code reference invalidation. All are substantive and specific to this change. |
| Likelihood + impact rated | 4/5 | Ratings are honest and varied: not everything is "low likelihood, high impact." Race conditions are rated medium likelihood, which is realistic for concurrent code generation. Old reference invalidation is rated low/low, which matches the internal testing context. |
| Mitigations are actionable | 4/5 | Mitigations are specific: "迁移脚本包裹在事务内," "迁移前全量备份数据库," "编写迁移回滚脚本," "利用唯一索引作为最终防线, service 层重试最多3次." Both short-term and long-term mitigations are provided for race conditions. |

**Deductions**:
- -1: Missing risk: existing SubItems have no codes. The migration section does not address how existing SubItems get codes assigned. This is a migration gap that could cause runtime errors or data inconsistency.
- -1: Missing risk: Team Code edge cases. What happens if a team is created without a code? What if the code is set to empty string or a single character? The 2-6 character constraint needs enforcement, and the risk of invalid team codes propagating into item codes is unaddressed.
- -1: The NextSubCode risk mitigation says "同一主事项并发创建子事项的场景较少" -- this is an assumption, not a mitigation. Relying on low concurrency is not a defense; the unique index + retry is the real mitigation.

### 6. Success Criteria: 11/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| Criteria are measurable | 4/5 | Seven criteria with "可测试条件" column. The migration criterion includes a verifiable SQL query. The immutability criterion specifies unit test verification. |
| Coverage is complete | 3/5 | Covers team code field, main item format, sub-item format, immutability, migration, frontend display, and existing tests. Gaps: no criterion for SubItem migration (existing SubItems need codes), no criterion for Team Code creation validation flow, no criterion for NextCode retry behavior under concurrent writes. |
| Criteria are testable | 4/5 | Most criteria are directly testable. The per-team strictly increasing criterion and the SQL verification query for migration are well-designed. |

**Deductions**:
- -1: Missing success criterion for SubItem migration. Existing SubItems have no codes; a criterion should verify that all existing SubItems receive valid codes after migration.
- -1: Missing criterion for Team Code validation. The proposal specifies 2-6 character letters, globally unique -- but no criterion verifies that invalid inputs (too short, too long, non-alpha, duplicate) are rejected with appropriate errors.
- -2: The "前端展示" criterion is vague: "编码显示为新格式；搜索框输入新格式编码可正确匹配." Which new format? The exact expected rendering should be specified (e.g., "displays as `FEAT-00001` for main items, `FEAT-00001-01` for sub-items"). "正确匹配" is not measurable -- what does "correct" mean? Exact match? Partial match? Fuzzy match?

---

## Summary of Attacks

1. **Problem Definition**: Zero evidence provided for any claimed problem -- "无法从编码识别所属团队" and "SubItem 没有任何人类可读的标识符" are asserted as self-evident without user feedback, support tickets, or quantitative data. A proposal to redesign a core identifier system should demonstrate that the current system causes measurable pain.

2. **Solution Clarity**: User-facing behavior is almost entirely omitted. The proposal describes database schema changes in precise detail but never describes what the user sees, does, or experiences differently. "更新各页面中编码的展示和搜索" names pages but describes no interaction. Team Code admin UX is a core new feature with zero user-flow description.

3. **Success Criteria**: Coverage gaps for SubItem migration and Team Code validation. The proposal adds a `Code` field to SubItem but has no criterion verifying that existing SubItems receive codes during migration. The frontend criterion ("编码显示为新格式") is too vague to be objectively verified.

## Improvement Summary vs. Iteration 1

| Dimension | Iter 1 | Iter 2 | Delta |
|-----------|--------|--------|-------|
| Problem Definition | 12 | 12 | 0 |
| Solution Clarity | 14 | 15 | +1 |
| Alternatives Analysis | 2 | 14 | +12 |
| Scope Definition | 11 | 11 | 0 |
| Risk Assessment | 6 | 12 | +6 |
| Success Criteria | 6 | 11 | +5 |
| **Total** | **51** | **75** | **+24** |

Major improvements: Alternatives Analysis (+12) with four well-analyzed alternatives and a design decisions table; Risk Assessment (+6) with four concrete risks and actionable mitigations; Success Criteria (+5) with seven testable criteria including a verifiable SQL query.

Persistent weaknesses: Problem evidence (still zero user data), user-facing behavior description (still absent), scope bounding (still no effort estimate).
