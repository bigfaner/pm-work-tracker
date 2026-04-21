# Proposal Evaluation Report

**Proposal**: Item Code Redesign (`item-code-redesign`)
**Iteration**: 1
**Date**: 2026-04-22
**Overall Score**: 51/100

---

## Dimension Scores

### 1. Problem Definition: 12/20

| Criterion | Score | Notes |
|-----------|-------|-------|
| Problem stated clearly | 5/7 | Three distinct problems listed; however, the second point ("SubItem has no human-readable identifier, can only rely on database ID") conflates a technical observation with a user-facing problem. Who is harmed by this and when? Unclear. |
| Evidence provided | 2/7 | No data, no user feedback, no concrete examples of where the current `MI-` format caused confusion or errors. Every claim is "we think X is a problem" without backing. The proposal asserts the prefix lacks business meaning, but never demonstrates that anyone actually found this confusing in practice. |
| Urgency justified | 5/6 | Implicitly justified by being a foundation for other features (sub-items need codes), but never states this explicitly. No "what happens if we don't" analysis. |

**Deductions**:
- -2: "无法从编码识别所属团队" (cannot identify team from code) is asserted without evidence that this is actually a pain point for users.
- -3: No quantitative or qualitative evidence anywhere in the problem section. Zero user feedback, zero data on support tickets, zero examples of confusion incidents.
- -3: SubItem lacking a code is presented as self-evident, but the impact is never articulated. In what workflow does a user need a sub-item code? What breaks without it?

### 2. Solution Clarity: 14/20

| Criterion | Score | Notes |
|-----------|-------|-------|
| Approach is concrete | 6/7 | The encoding format table is excellent: clear format, examples, and composition rules. A reader can explain this back. The column size changes (varchar(10) -> varchar(12), new varchar(15)) are precise. |
| User-facing behavior described | 3/7 | The proposal describes internal model changes well but almost entirely omits what the end user experiences. How does a user see the new code? Where do they search by code? What happens when they type `FEAT-` in a search box? What validation errors do they get if they enter a bad Team Code? None of this is described. |
| Distinguishes from alternatives | 5/6 | The format is clearly distinct from the current `MI-NNNN`, and the immutability + snapshot rules differentiate it from simpler approaches. |

**Deductions**:
- -2: "更新各页面中编码的展示和搜索" (update code display and search in pages) is vague hand-waving. Which specific UI elements change? What does the user see differently?
- -2: No description of the Team Code management UX. How does an admin create or edit a Team Code? What validation feedback do they get? This is a core user-facing feature with zero UX description.
- -2: The data migration section says "direct switch, no compatibility with old format needed" but does not describe the migration behavior at all. What happens to existing `MI-0001` codes? Are they rewritten? Left as-is? If left as-is, users will see two different code formats in the system simultaneously, which is a user-facing behavior that must be described.

### 3. Alternatives Analysis: 2/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| At least 2 alternatives listed | 1/5 | Zero alternatives are presented. The "do nothing" option is not discussed. No comparison of different encoding schemes (e.g., UUID-based, auto-increment without prefix, project-scoped codes). |
| Pros/cons for each | 0/5 | No alternatives means no pros/cons analysis. |
| Rationale for chosen approach | 1/5 | The proposal states its solution but never explains why this approach was chosen over others. Why team-based prefix vs. a global counter? Why 5-digit sequence vs. 4? Why immutable snapshot vs. updatable code? |

**Deductions**:
- -5: Entire section missing. This is a critical gap for a redesign proposal. There are many reasonable alternatives (keep MI- prefix but add team info elsewhere, use UUIDs, use project-scoped codes, make codes editable) that should be evaluated.
- -3: No "do nothing" analysis. The current system works; why is the disruption justified?
- -5: No rationale for specific design choices. Why snapshot team code at creation time vs. always using current team code? Why per-team sequence vs. global? These are non-obvious design decisions that need justification.

### 4. Scope Definition: 11/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| In-scope items are concrete | 4/5 | Model changes are specific (field names, types, varchar sizes). Frontend changes are listed by page name. The `NextCode()` and `NextSubCode()` functions are named. Good specificity on the backend side. |
| Out-of-scope explicitly listed | 4/5 | Four explicit out-of-scope items listed, which is good. Each names a specific capability being deferred. |
| Scope is bounded | 3/5 | The scope is moderately bounded, but "数据迁移" (data migration) is listed as a single line item with no detail. Migration of existing codes is arguably the highest-risk part of this change, and reducing it to one bullet point makes the scope deceptively small. The frontend changes list four pages by name but gives no indication of complexity per page. |

**Deductions**:
- -2: Data migration is a major scope item treated as an afterthought. "直接切换" (direct switch) is a strategy, not a scope description. What SQL runs? On how many rows? What is the rollback plan?
- -2: No indication of timeline or effort. Is this a one-day change or a one-week change? Without effort estimation, scope is unbounded in practice.

### 5. Risk Assessment: 6/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| Risks identified | 2/5 | Zero risks are explicitly identified. There is no Risk section at all. Implicitly, one could infer risks (data loss during migration, sequence gaps, code collision), but the proposal does not name them. |
| Likelihood + impact rated | 2/5 | No likelihood or impact ratings since no risks are listed. The only risk-adjacent content is the immutability rule and snapshot approach, which implicitly mitigate certain risks without naming them. |
| Mitigations are actionable | 2/5 | The snapshot rule is a mitigation for team-code-renaming, but it is not framed as a risk mitigation. The "direct switch, no compatibility" migration approach is actually a risk decision (choosing to accept format breakage) but is stated as if it were consequence-free. |

**Deductions**:
- -3: No risk section exists. For a schema redesign touching three models and a migration, this is a significant omission. Risks like concurrent code generation race conditions, existing data with malformed codes, and search/index performance with longer codes are all unaddressed.
- -3: The migration approach ("direct switch, no backward compatibility") is the single highest-risk decision in the proposal and is treated as a one-liner. What if migration fails halfway? What if external systems reference old codes? What if users have bookmarked or shared old-format codes?
- -3: No consideration of the race condition risk in NextCode. The current implementation uses `MAX(code)` which is known to be vulnerable under concurrent writes. The proposal adds per-team sequences but does not discuss whether this makes the race condition better or worse.

### 6. Success Criteria: 6/15

| Criterion | Score | Notes |
|-----------|-------|-------|
| Criteria are measurable | 2/5 | No success criteria section exists. There is no list of acceptance criteria or measurable outcomes. |
| Coverage is complete | 2/5 | Without explicit criteria, coverage cannot be evaluated. The scope items (team code field, main item format, sub-item codes, frontend updates, migration) have no corresponding verification conditions. |
| Criteria are testable | 2/5 | Not applicable without explicit criteria. However, the encoding format table is precise enough that tests could be written from it, which earns partial credit. |

**Deductions**:
- -4: No success criteria section at all. There is no way to determine when this feature is "done" other than developer judgment.
- -3: The Open Questions section says "all confirmed through conversation, no remaining questions" which is a missed opportunity to list acceptance criteria. If all questions were resolved, those resolutions should be documented as verifiable criteria.
- -2: The proposal does not define what "done" looks like for the data migration. All existing codes converted? New format only? Mixed format acceptable?

---

## Summary of Attacks

1. **Alternatives Analysis**: No alternatives evaluated at all -- the proposal jumps straight to a solution without considering "do nothing," UUID-based codes, global vs. per-team sequences, or any other encoding scheme. This is the weakest dimension.

2. **Risk Assessment**: Zero explicit risks identified for a change that modifies three database models, introduces a data migration, and changes the user-facing identifier format. The "direct switch" migration is the highest-risk decision and gets one sentence.

3. **Success Criteria**: No acceptance criteria, no testable outcomes, no definition of "done." The Open Questions section claims everything is resolved, but none of those resolutions are documented as verifiable criteria.

## Recommendations for Next Iteration

1. Add an Alternatives section with at least 2 alternatives (including "do nothing"), with honest pros/cons for each.
2. Add a Risk section identifying at least 3 risks (migration failure, race conditions in code generation, search performance with longer codes) with likelihood, impact, and mitigations.
3. Add a Success Criteria section with measurable, testable acceptance criteria for each in-scope item.
4. Flesh out the data migration: what SQL runs, how many rows are affected, what the rollback plan is, and what happens to existing MI-format codes.
5. Add evidence to the Problem section: user feedback, support tickets, or concrete examples of the current format causing problems.
