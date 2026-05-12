---
date: "2026-05-12"
doc_dir: "docs/proposals/milestone-map/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 1

**Score: 63/100** (target: 90)

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 15 | 20 |
| Solution Clarity | 13 | 20 |
| Alternatives Analysis | 10 | 15 |
| Scope Definition | 11 | 15 |
| Risk Assessment | 7 | 15 |
| Success Criteria | 7 | 15 |

---

## Deductions

| Rule | Instance | Deduction |
|------|----------|-----------|
| Vague language | "开发量较大" — no quantification of effort | -2 (Solution Clarity) |
| Vague language | "项目规模增长后" — no data on current or projected scale | -2 (Problem Definition) |
| Vague language | "使用成熟的前端时间线库" — no specific library named | -2 (Risk Assessment) |
| Vague language | "开发量小" — no quantification | -2 (Alternatives Analysis) |
| Missing detail | No concrete user scenario or persona described | -2 (Problem Definition) |
| Missing detail | Success criteria lack acceptance thresholds or quantitative measures | -3 (Success Criteria) |
| Missing detail | Risk mitigations are high-level, not actionable steps | -3 (Risk Assessment) |
| Missing detail | "拖拽+缩放实现困难" risk mitigation names no specific library or approach | -2 (Risk Assessment) |
| Weak justification | "完整体验" vs other alternatives — verdict driven by assertion, not tradeoff analysis | -2 (Alternatives Analysis) |
| Incomplete coverage | Success criteria do not cover RBAC permission granularity or soft-delete behavior edge cases | -2 (Success Criteria) |

---

## Attack Points

### Attack 1: Risk Assessment — Mitigations are hand-waving, not actionable
**Where**: "使用成熟的前端时间线库，降低自研成本"
**Why it's weak**: The highest-impact risk (visual timeline complexity) has a mitigation that names no specific library, no evaluation criteria, and no fallback plan if no suitable library exists. The other two mitigations merely describe the design decision itself (nullable FK, a selector UI) rather than addressing the risk. A developer reading this cannot act on any of these mitigations — they are restatements of the approach, not risk responses.
**What must improve**: Name at least one candidate library (e.g., vis-timeline, react-calendar-timeline). State the evaluation criteria. For each risk, the mitigation must be a concrete action: "Spike on library X by [date]; if unsuitable, fall back to simpler static layout in v1." The second and third mitigations should address what goes wrong (FK migration risks, UX conflicts), not just describe the feature.

### Attack 2: Success Criteria — Not measurable or testable
**Where**: "PM 能创建/编辑/删除里程碑，关联/解绑 MainItem" and "里程碑权限独立控制，无权限用户无法进行未授权操作"
**Why it's weak**: Every success criterion is a qualitative statement with "能" (can) — no acceptance threshold, no edge case, no quantifiable bar. "正确展示所有里程碑" — what is "correct"? What about milestones with zero MIs? Milestones with cancelled status? The permission criterion says "无权限用户无法进行未授权操作" but does not define which roles have which permissions by default. Without specificity, these criteria cannot be turned into test cases with pass/fail outcomes.
**What must improve**: Each criterion must have a verifiable condition. Example: "Given a milestone with 3 associated MIs (1 completed, 1 in_progress, 1 not_started), the milestone completion percentage displays as 33%." Add edge-case criteria: empty milestones, cancelled milestones, MI moved between milestones. Permission criteria should specify which operations map to which permission codes and what the denied-state UX is.

### Attack 3: Problem Definition — Urgency is asserted, not justified
**Where**: "项目规模增长后，事项数量增多，PM 需要更高层级的跟踪视角。缺少里程碑视图会导致进度把控困难，无法快速识别哪些阶段有风险。"
**Why it's weak**: The urgency section contains zero data. How many MIs does the average project have now? At what scale does the problem become acute? Is there user feedback requesting this? Is there a specific project or incident where the lack of milestones caused a missed deadline or miscommunication? Without concrete evidence, the urgency is an assumption, not a justification. The "evidence" section lists only the absence of a feature, not proof that the absence causes harm.
**What must improve**: Add quantitative evidence: current average MI count per project, number of projects where PMs reported difficulty, or at minimum a specific scenario ("Project X had 47 MIs across 3 phases; PM could not identify which phase was delayed without manual tracking"). State the cost of inaction concretely: "Without this, PMs spend approximately N hours/week on manual milestone tracking via spreadsheets."

---

## Verdict

- **Score**: 63/100
- **Target**: 90/100
- **Gap**: 27 points
- **Action**: Continue to iteration 2
