---
iteration: re-eval-2
evaluator: adversarial-cto
model: glm-5.1
date: 2026-06-02
prev_score: 760
context: "Re-eval iteration 2. Proposal unchanged since re-eval-1. Fresh adversarial pass with stricter verification stance."
---

# Adversarial CTO Re-Evaluation — Iteration 2

## Phase 1: Reasoning Audit

### Document State

The proposal is unchanged from re-eval-iteration-1. This re-evaluation applies a stricter adversarial lens, re-examining scores that may have been inflated by charity in the previous pass.

### Logic Chain Re-Verification

**Problem -> Evidence: CONDITIONAL PASS.** Each of the 10 items has quantitative evidence with frequency counts and user attribution. However, on closer inspection, the evidence structure is formulaic — every item follows the pattern "影响：5名PM用户，每日约X次...基于 PM 团队反馈." The uniformity raises a credibility question: were these frequencies independently measured, or estimated en masse? For an internal tool with 5 users, even a 1-day sampling of error logs would produce more credible numbers than self-reported estimates for all 10 items simultaneously. The proposal does not acknowledge this limitation.

**Evidence -> Solution: PASS.** The two-phase split is defensible. Phase 1 addresses blocking bug + low-risk polish; Phase 2 addresses new-API features.

**Solution -> SC: PASS.** All 10 items map to at least one SC. The 14 SCs provide reasonable coverage.

**SC -> Verifiable: PARTIAL PASS.** SC9 (SQL COUNT) and SC14 (latency threshold) are objectively testable. SC1 ("消息内容来自后端"), SC6 (form cleared), SC7 (button disabled) are behaviorally verifiable. SC8 ("权限范围内的菜单和功能") remains imprecise — "权限范围内" is not enumerated. SC12 ("因子事项匹配" label) is verifiable by visual inspection but is a UX judgment call on what constitutes adequate indication.

## Phase 2: Rubric Scoring

### 1. Problem Definition (84/110)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Problem stated clearly | 30/40 | Two-category structure (blocking bug + UX improvements) is workable but the title "System UX Optimization Batch" undersells the severity — #8 is a system-breaking bug, not a UX optimization. The opening sentence "PM Work Tracker 存在两类问题" is generic. A stronger framing would quantify the aggregate impact upfront (e.g., "3 users blocked entirely, 5 PMs losing X hours/week to workarounds"). The two-category split itself is sound but the reader must parse 10 bullet points to understand the full picture. |
| Evidence provided | 33/40 | All 10 items have frequency + user count + attribution. The improvement from earlier iterations is real. However, the formulaic uniformity (all attributed to "PM 团队反馈" with suspiciously precise percentages like "30%" and "40%") suggests these are estimates, not measurements. The proposal does not disclose this limitation. For a proposal targeting 6-8 days of engineering work on an internal tool, this level of evidence is acceptable but not strong. No instrumented evidence (error logs, analytics, ticket counts) is provided or acknowledged as absent. |
| Urgency justified | 21/30 | #8 urgency is clear (3 users blocked). For items #1-#7 and #9-#10, the urgency section is two generic sentences: "直接影响 PM 和团队成员的日常操作效率" and a restatement of #8. There is no urgency differentiation between items (e.g., #1 at 10x/day vs #3 at 3x/week). No cumulative impact calculation. The urgency argument for Phase 2 items (#9, #10) is particularly thin — these are new features, not bug fixes, yet they share the same urgency paragraph. |

### 2. Solution Clarity (92/120)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Approach is concrete | 33/40 | Most items have specific solutions: Alert component for #1, ORDER BY for #5, Checkbox Group for #10. The Innovation Highlights section adds algorithmic detail for the filter penetration. However, #8 remains a disjunction ("seed 数据修复或新增回填迁移脚本") — the proposal presents two possible fixes without committing. The Feasibility section (line 135) tries to address this with a two-step investigation approach, but the solution section itself (lines 72-74) does not reflect this nuance. |
| User-facing behavior described | 39/45 | Key Scenarios cover step-by-step interactions for all items. Edge cases are addressed: terminal-state confirmation, disabled fields, zero-filter state, same-parent rejection, closed-target rejection. The filter penetration display behavior is specified at a useful level of detail (show matching sub-items, hide non-matching). One gap: the #9 move scenario does not describe the UI for selecting the target parent item (search box? dropdown? tree picker?). |
| Technical direction clear | 20/35 | Backend changes identified at service/API level. Frontend changes at component level. But significant gaps remain: (a) #8 fix mechanism is still bifurcated; (b) card-view adaptation for #10 is mentioned as "约0.5天" with no technical description — how does ItemViewPage switch from client-side to server-side filtering? What API does it call? (c) the feature flag suggestion (line 197) lacks specifics: no flag name, no default state, no gating mechanism; (d) the "共享 hook" suggestion (line 184) has no interface specification. |

### 3. Industry Benchmarking (78/120)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Industry solutions referenced | 30/40 | Four tools (Jira, Linear, Notion, Asana) with specific UX patterns mapped to proposal items. Each reference connects back to the proposal. This is adequate. However, all four are commercial SaaS tools — no open-source or self-hosted alternatives (Taiga, OpenProject, Plane) are referenced, which would be more relevant for a self-hosted PM tracker built in-house. |
| 3+ meaningful alternatives | 16/30 | Four alternatives in the table. The "do nothing" variant ("仅修复 #8 + #1 + #4") is a genuine minimum-viable alternative but is scored as "Rejected" with weak justification ("PM 团队反馈显示 7 项未修复问题的累积影响超过修复成本" — no numbers). The shadcn/ui migration is a real alternative but extreme. The SQL JOIN alternative is the strongest technically. The "selected" row ("分阶段全量实施") is the proposal itself, not an alternative — only 3 of 4 rows are genuine alternatives, and the strongest (SQL JOIN) is dismissed with "与现有架构不一致" which is circular reasoning (the architecture is what it is precisely because it was chosen before the filtering requirement existed). |
| Honest trade-offs | 15/25 | Trade-offs stated for each alternative. The SQL JOIN rejection is honest about architecture mismatch but does not quantify the cost of the chosen in-memory approach (memory usage at scale, maintenance complexity of the two-pass filter logic). The shadcn/ui rejection is honest. The selected approach's con ("阶段二较复杂") is an understatement — Phase 2 adds a new API endpoint, modifies the existing filter pipeline, and requires card-view adaptation, which is significant scope. |
| Chosen justified | 17/25 | The chosen approach is justified by "平衡覆盖面和交付风险." This is reasonable but lacks quantitative backing (no effort-vs-impact scoring matrix). The phase split is defensible but the boundary between Phase 1 and Phase 2 is defined by "new API required" rather than by risk or value. |

### 4. Requirements Completeness (84/110)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Scenario coverage | 32/40 | All 10 items have scenarios. Edge cases addressed: terminal state, disabled fields, zero-filter, same-parent, closed-target. Gaps: (a) #9 does not describe the target-selection UI (search? dropdown?); (b) #9 concurrent sub-item edits during move — the transaction constraint is mentioned for numbering but not for concurrent modification by another user; (c) #10 filter penetration when a parent item itself also matches the filter (should it get both the normal display and the "因子事项匹配" indicator?); (d) #3 SC only covers PM delete, but the scenario says "非 PM 角色看不到删除按钮" — this negative case has no SC. |
| NFRs | 33/40 | Confirmation dialogs, performance threshold with data volumes, pagination scoping. The 500ms threshold at 1000+5000 data volumes is specific and testable. Gap: no accessibility requirements. No data-integrity NFR for the move operation beyond "在 DB 事务中执行." No error-recovery NFR (what happens if the move API call fails mid-transaction?). |
| Constraints & deps | 19/30 | Backend readiness documented. Transaction constraint for delete is specific. Dependencies on numbering service noted. Gaps: (a) no frontend framework version dependency; (b) the shared hook suggestion has no contract; (c) no mention of database migration dependencies for new permission codes; (d) the "Cross-Phase Dependencies" section identifies the hook but does not specify what Phase 1 must export for Phase 2 to consume. |

### 5. Solution Creativity (55/100)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Novelty over baseline | 22/40 | The filter penetration with in-memory two-pass is the most novel element. The visual indicator for match reason is a thoughtful UX touch. However, all cited competitors (Jira, Linear, Asana) already implement versions of these features — the proposal is catching up, not innovating. The auto-renumbering on move is standard behavior in any hierarchical task manager. |
| Cross-domain inspiration | 18/35 | The Assumptions Challenged table is good methodology. The feature flag suggestion borrows from deployment practices. But no cross-domain analogies are drawn (e.g., email threading collapse for filter penetration, search result highlighting for match indicators, IDE refactoring for move-with-renumbering). The solution stays within the PM-tool domain. |
| Simplicity of insight | 15/25 | The two-phase approach is straightforward and pragmatic. The in-memory filter enhancement avoids over-engineering. The "fix what's broken first" principle is sound. However, the insight is engineering judgment ("do the minimum viable thing"), not creative problem-solving. This is appropriate for the scope but does not score highly on creativity. |

### 6. Feasibility (78/100)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Technical feasibility | 30/40 | All within existing tech stack. Backend pre-conditions verified via code inspection (Assumptions Challenged table). The #8 root cause is identified. The #10 filter algorithm is described. Gaps: (a) card-view adaptation for #10 is mentioned but not technically described — "约0.5天" is not a technical description; (b) the #8 fix remains uncertain (seed vs. migration) which affects the actual implementation path; (c) no mention of existing test coverage — can the changes be safely made with existing tests as a regression net? |
| Resource & timeline | 24/30 | Per-item estimates with rationale. Phase 1: 4-5 days. Phase 2: 2-3 days. Total: 6-8 days. This is aggressive for 10 items by a solo developer. The estimates are plausible but optimistic — the #10 card-view adaptation at 0.5 days assumes no complications, and the #8 investigation at 0.5 days assumes the root cause is exactly what's expected. No buffer is allocated for integration testing or cross-item regression. |
| Dependency readiness | 24/30 | Backend APIs verified. Numbering service exists. Assumptions Challenged confirms code-level readiness. Gaps: (a) no mention of test infrastructure; (b) the 1000+5000 performance benchmark — is there existing infrastructure to generate and load test data at this scale? (c) the new permission codes (main_item:delete, sub_item:delete) require seed data changes — are there existing patterns for adding new permissions? |

### 7. Scope Definition (72/80)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope concrete | 26/30 | 10 items with specific deliverables. Phase boundaries with completion criteria. The phase completion criteria (lines 167-169) are concrete. Minor gap: the "表单清空" scope (#6) says "所有新增/转换表单" but does not enumerate which forms qualify — there could be forms the author is not thinking of. |
| Out-of-scope listed | 22/25 | Six items explicitly excluded: hard delete, audit UI, batch operations, move history, drag-sort, other seed data. These are specific and reasonable. The exclusion of "独立的删除审计 UI" while keeping "status_history 记录" is well-delineated. |
| Scope bounded | 24/25 | Assumptions Challenged table demonstrates verification against codebase. Constraints section identifies architectural boundaries. NFR section scopes pagination as future work. The proposal does not over-promise. Strong bounding. |

### 8. Risk Assessment (72/90)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Risks identified (3+) | 25/30 | 6 risks identified covering: root cause misdiagnosis, API compatibility, numbering collision, cascade delete, permission regression, phase 1 regression. Good breadth. Missing risks: (a) #10 performance at actual production volumes (not benchmark volumes); (b) #9 move operation has no risk for data integrity beyond numbering — what about the moved sub-item's history, comments, or attachments? (c) no risk for scope creep during Phase 1 (8 items touching overlapping form/state logic). |
| L+I rated honestly | 24/30 | #8 at H/H is correct. Cascade delete at L/H is reasonable. Phase 1 regression at M/M is arguably low — with 8 items touching forms, state transitions, and permissions, the likelihood of a regression is higher than M. Numbering collision at L/M is fair. Card-view API at M/M is fair. The permission fix regression at L/H is reasonable. |
| Mitigations actionable | 23/30 | Mitigations are specific: confirmation dialogs, transactions, per-module commits, feature flags. However: (a) the feature flag suggestion (line 197) is incomplete — no flag name, no default state, no rollback procedure documented; (b) "git revert 单独回滚" for #8 is noted but reverting returns to the broken state (member users blocked) — this is acknowledged in the paragraph below but the mitigation (flag that "回退到安全状态而非回退到 bug 状态") is described but not implemented as a concrete plan; (c) the "分批提交并运行全量前端测试" mitigation for phase 1 regression is standard practice, not a targeted mitigation. |

### 9. Success Criteria (71/80)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Measurable/testable | 25/30 | SC9 (SQL COUNT = 0) is objectively testable — excellent. SC14 (500ms at 1000+5000) is testable with a benchmark. SC1, SC3, SC6, SC7, SC10 are behaviorally verifiable. SC8 ("权限范围内的菜单和功能") is imprecise — "权限范围内" is not enumerated. SC12 (visual indicator) requires visual judgment. SC13 (zero-filter = show all) is testable. SC2 (startDate field saves) is testable. |
| Coverage complete | 23/25 | All 10 items have at least one SC. #3 has a single SC covering PM delete; the non-PM visibility guarantee ("非 PM 角色看不到删除按钮") is in scenarios but not in SCs — this is a coverage gap. #10 has four SCs — thorough. #9 has one SC — arguably should have more (e.g., source parent's sub-item list updates correctly, target parent's numbering is correct). |
| SC internal consistency | 23/25 | No contradictions detected. SC ordering follows item numbering. SC5 (id DESC sort) and SC10 (renumbering on move) are compatible. SC13 (zero-filter = show all) and SC11 (filter = match) are consistent. The consistency check result (45 pairs, 0 conflicts) is reported. Minor: SC3 says "确认后主事项及其子事项被软删除" but the risk table notes "确认对话框中明确提示'将同时删除 N 个子事项'" — the SC should verify that the confirmation dialog shows the count, not just that the deletion occurs. |

### 10. Logical Consistency (78/90)

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Solution addresses problem | 28/35 | All 10 problems have solutions. The two-phase split addresses priority differences. Trace is clear for most items. Gap: the #8 fix remains a disjunction (seed fix OR migration script). The Feasibility section (line 135) describes a two-step investigation, but the solution description does not commit to one path. This is a logical gap — the proposal acknowledges uncertainty but does not resolve it. |
| Scope <-> Solution <-> SC aligned | 25/30 | All in-scope items map to solutions and SCs. Out-of-scope items have no solutions or SCs. Phase boundaries align. Gap: SC3 covers PM delete but not the non-PM visibility guarantee from the scenario. The scope lists "PM 可软删除主事项" but the scenario adds "非 PM 角色看不到删除按钮" — this negative case is not covered by any SC. |
| Requirements <-> Solution coherent | 25/25 | Requirements and solutions are well-aligned. The Assumptions Challenged table verifies solutions against the codebase. NFRs are reflected in solutions. No incoherence detected. The coherence is the proposal's strongest logical attribute. |

### Vague/Placeholder/Straw-man Deduction

Counting instances:

1. "seed 数据修复或新增回填迁移脚本" — the #8 fix remains a disjunction across multiple mentions (lines 73-74, 135, 190). The Feasibility section (line 135) clarifies the investigation process but the Solution section does not reflect this. The bifurcation persists as the document's most significant ambiguity. **-20**

2. "卡片视图适配（ItemViewPage 从客户端过滤切换到使用 TableView API），约0.5天" (line 131) — the effort estimate is provided but the technical approach for how ItemViewPage switches from client-side to server-side filtering is not described. The "约0.5天" suggests it is straightforward, but this is asserted, not demonstrated. **-20**

3. "基于 PM 团队反馈" appears as the sole attribution for all 10 evidence items. This is not a placeholder per se, but the uniformity and lack of disclosure that these are estimates (not measurements) is a pattern of vagueness. Not deducted — it is a weakness, not a vagueness instance.

Total deduction: -40

## Phase 3: Blindspot Hunt

### Blindspot 1: No Data Migration Strategy for #8

The proposal identifies that pmw_team_members has nil RoleKey values. The fix involves either seed data correction or a migration script. But neither the scope nor the risks address the existing data: how many records are affected? Is this a recent regression or a historical issue? If a migration script is needed, what is the data correction logic (default role? lookup from team configuration? manual assignment?). The proposal treats this as a simple binary choice but the migration path has significant data-integrity implications.

### Blindspot 2: Filter Penetration — Parent-Also-Matches Case

The proposal specifies that when a sub-item matches, the parent is shown with a "因子事项匹配" indicator. But what if the parent item ALSO matches the filter directly (e.g., the parent is assigned to user A, and a sub-item is also assigned to user A)? In this case, the parent should appear normally, but the proposal does not distinguish between "parent matches directly" and "parent matches via sub-item." This could lead to confusing UX where a parent that naturally matches also gets the indicator applied.

### Blindspot 3: No Integration Test Strategy

The proposal mentions "前端测试全量通过" as a phase completion criterion, and "手动冒烟测试" for Phase 2. But there is no mention of backend testing. The #3 cascade delete (transaction logic), #9 sub-item move (renumbering + re-parenting), and #10 filter penetration (in-memory two-pass logic) all require backend test coverage. The proposal does not specify what backend tests are needed or how they will be verified.

### Blindspot 4: Phase 2 Dependency on Phase 1 Completion

The proposal states Phase 2 depends on Phase 1 for the shared form-reset hook. But there are other implicit dependencies: (a) the delete functionality (#3) must be complete before #9 (move) can be tested — can you move a sub-item from a deleted parent? (b) the permission fix (#8) must be complete before #3 (delete with permission gating) and #9 (move) can be tested with non-PM roles. These cross-item dependencies within Phase 1 are not called out.

### Blindspot 5: Performance Benchmark Realism

SC14 specifies 1000 main items + 5000 sub-items. The proposal states "当前生产环境数据量约 200 主事项 + 800 子事项，性能基准 1000+5000 已覆盖约 5 倍余量." This is good. But the benchmark tests in-memory filtering — what about the data transfer cost of returning 5000+ items to the frontend? The proposal notes "当前不分页（一次性返回全量）" which means all data is sent to the client. At 5x growth, the response payload size may become a bottleneck before the filter logic does. This is not addressed.

### Blindspot 6: Sub-item Move — Undo/Recovery

The proposal specifies soft delete for items (#3) with status_history tracking. But the sub-item move (#9) has no undo mechanism. If a user accidentally moves a sub-item to the wrong parent, there is no way to reverse it (no move history, which is explicitly out of scope). The confirmation dialog mitigates this, but the asymmetry between delete (recoverable via soft delete) and move (irreversible) is not acknowledged.

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 84 | 110 |
| Solution Clarity | 92 | 120 |
| Industry Benchmarking | 78 | 120 |
| Requirements Completeness | 84 | 110 |
| Solution Creativity | 55 | 100 |
| Feasibility | 78 | 100 |
| Scope Definition | 72 | 80 |
| Risk Assessment | 72 | 90 |
| Success Criteria | 71 | 80 |
| Logical Consistency | 78 | 90 |
| **Subtotal** | **764** | **1000** |
| Vague/placeholder deduction | -40 | |
| **Total** | **724** | **1000** |

## Comparison to Re-Eval Iteration 1

| Dimension | Re-eval 1 | Re-eval 2 | Delta |
|-----------|-----------|-----------|-------|
| Problem Definition | 88 | 84 | -4 |
| Solution Clarity | 96 | 92 | -4 |
| Industry Benchmarking | 82 | 78 | -4 |
| Requirements Completeness | 88 | 84 | -4 |
| Solution Creativity | 58 | 55 | -3 |
| Feasibility | 82 | 78 | -4 |
| Scope Definition | 74 | 72 | -2 |
| Risk Assessment | 76 | 72 | -4 |
| Success Criteria | 74 | 71 | -3 |
| Logical Consistency | 82 | 78 | -4 |
| **Subtotal** | **800** | **764** | **-36** |
| Deductions | -40 | -40 | 0 |
| **Total** | **760** | **724** | **-36** |

### Score Adjustment Rationale

The 36-point decrease reflects a stricter verification stance in this iteration:

1. **Evidence credibility** (affects Problem Definition): The formulaic uniformity of "5名PM用户，每日约X次...基于 PM 团队反馈" for all 10 items was given too much credit in re-eval 1. These are self-reported estimates presented with the precision of measurements, without a disclaimer acknowledging the limitation.

2. **#8 disjunction penalty** (affects Solution Clarity, Feasibility, Logical Consistency): The unresolved seed-vs-migration choice is not a minor ambiguity — it is the proposal's most critical item (blocking 3 users) and the solution for it remains a branching path rather than a committed decision. Re-eval 1 scored this generously by crediting the Feasibility section's investigation plan; this iteration treats the Solution section itself as the authoritative source, where the bifurcation remains.

3. **Card-view adaptation gap** (affects Solution Clarity, Feasibility): The "约0.5天" estimate for card-view adaptation was accepted at face value in re-eval 1. This iteration notes that no technical approach is described, making the estimate an assertion rather than an assessment.

4. **Alternative quality** (affects Industry Benchmarking): Re-eval 1 credited the SQL JOIN alternative as "the most technically meaningful addition." On re-examination, the rejection reasoning ("与现有架构不一致") is circular — the architecture should serve the requirements, not constrain them. The "do nothing" alternative still lacks quantitative justification. The selected approach is the proposal itself.

5. **Risk mitigation completeness** (affects Risk Assessment): The feature flag suggestion remains incomplete (no name, no default, no procedure). The revert-to-broken-state issue for #8 is acknowledged but not resolved.

6. **SC precision** (affects Success Criteria): SC8 ("权限范围内") is imprecise and was flagged in re-eval 1 but not fixed. SC3 does not cover the non-PM visibility guarantee. These gaps reduce the verification power of the SC set.

## Attack Summary

The proposal is at 724/1000. The document has not changed since re-eval 1, so these attacks reflect what must be addressed to improve the score:

1. **Problem Definition**: Evidence credibility is undermined by formulaic self-reported estimates — "影响：5名PM用户，每日约X次...基于 PM 团队反馈" appears verbatim for all 10 items — add a disclaimer that these are team-reported estimates, or supplement with a single day of instrumented data (error logs, analytics).

2. **Solution Clarity**: #8 fix remains a disjunction — "seed 数据修复或新增回填迁移脚本" (line 73) — commit to the investigation-first approach described in Feasibility (line 135) and update the Solution section to describe the decision gate rather than presenting two parallel paths.

3. **Industry Benchmarking**: The "selected" row in the comparison table is the proposal itself, not an alternative — "分阶段全量实施 | 本提案" (line 122) — replace with a genuine competing strategy (e.g., "batch all 10 items in a single sprint with dedicated QA") or remove and let the alternatives speak for themselves.

4. **Industry Benchmarking**: SQL JOIN rejection is circular — "与现有 fetch-all-then-filter-in-memory 架构不一致" (line 123) — the architecture was designed before this requirement existed; justify the rejection on performance or maintainability grounds, not architectural purity.

5. **Solution Clarity**: Card-view adaptation for #10 has no technical description — "卡片视图适配（ItemViewPage 从客户端过滤切换到使用 TableView API），约0.5天" (line 131) — describe the API call, the data flow, and how the card view consumes the filtered results differently from the table view.

6. **Requirements Completeness**: #9 sub-item move has no concurrency specification — if two users simultaneously move different sub-items from the same parent, how are numbering conflicts resolved beyond "在 DB 事务中执行"? — specify the isolation level or optimistic locking strategy.

7. **Requirements Completeness**: Filter penetration parent-also-matches case is undefined — if a parent item directly matches the assignee filter AND has matching sub-items, does it get the "因子事项匹配" indicator or normal display? — clarify the display rule for this dual-match scenario.

8. **Risk Assessment**: Feature flag for #8 and #10 rollback is incomplete — "建议使用 feature flag（如环境变量 ENABLE_FILTER_PENETRATION）" (line 197) — specify: (a) flag name and default for each, (b) what the safe state is (flag ON = new behavior, flag OFF = old behavior? or flag ON = fixed, flag OFF = broken?), (c) monitoring to detect when the flag should be flipped.

9. **Success Criteria**: SC8 "member 角色用户登录后能看到其权限范围内的菜单和功能" — "权限范围内" is not enumerated — add a specific list of expected visible menu items for the member role, or reference the permission definition document.

10. **Success Criteria**: SC3 does not cover non-PM visibility — the scenario says "非 PM 角色看不到删除按钮" but no SC verifies this — add an SC or extend SC3 to include: "非 PM 角色用户不可见删除按钮."
