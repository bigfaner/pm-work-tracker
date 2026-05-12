---
date: "2026-05-12"
doc_dir: "docs/proposals/milestone-map/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 2

**Score: 88/100** (target: 90)

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 17 | 20 |
| Solution Clarity | 16 | 20 |
| Alternatives Analysis | 12 | 15 |
| Scope Definition | 14 | 15 |
| Risk Assessment | 14 | 15 |
| Success Criteria | 15 | 15 |

---

## Deductions

| Rule | Instance | Deduction |
|------|----------|-----------|
| Vague language | "开发量较大" in alternatives table — still no quantification of effort (hours, sprints, story points) | -2 (Alternatives Analysis) |
| Weak justification | Chosen approach verdict "一次性交付完整体验" — does not quantify cost tradeoff against staged delivery. Why is a single large delivery superior to incremental rollout? | -2 (Alternatives Analysis) |
| Missing detail | Solution description does not specify the timeline visualization's data granularity — how many MIs per milestone before performance degrades? No perf budget mentioned. | -1 (Solution Clarity) |
| Missing detail | Problem evidence relies on assumed PM workflows (naming conventions, memory-based tracking) without citing specific user interviews, support tickets, or direct PM feedback quoting the pain. | -2 (Problem Definition) |
| Missing detail | Risk mitigations, while much improved, still lack an owner or deadline. "优先评估 vis-timeline" — by when? Who does the evaluation? | -1 (Risk Assessment) |

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Risk mitigations are hand-waving, not actionable | Partially | Specific libraries named (vis-timeline, react-calendar-timeline), 3 evaluation criteria listed, fallback plan to static cards specified. Still missing: spike owner and deadline. Mitigation quality went from 2/10 to 8/10. |
| Attack 2: Success criteria not measurable or testable | Fully | Every criterion now has quantified acceptance conditions: "completion=0", "DECIMAL(5,2)", "3 MI (100/50/0) → 50.00", specific role-permission mappings (PM=all 4, member=read only), denied UX state (disabled + tooltip). All edge cases covered (empty milestone, cancelled status, drag to blank = unbind). |
| Attack 3: Problem urgency asserted, not justified | Fully | Urgency now quantified: 30-50 MIs per team, 3-4 phases per project, 20-30 min/week manual tracking per project, 8 hours/month PM waste across 5 projects. Specific workflow described (naming conventions + memory). |

---

## Attack Points

### Attack 1: Alternatives Analysis — Chosen approach lacks cost-weighted tradeoff
**Where**: "完整里程碑图（本方案） | 一次性交付完整体验 | 开发量较大 | Selected"
**Why it's weak**: The verdict for the chosen approach relies on the appeal of "完整体验" without quantifying what "开发量较大" means in concrete terms. There is no staged-delivery alternative (e.g., "data model + list view in v1, timeline visualization in v1.1"). The proposal jumps from "do nothing" to "full experience" with no middle-ground alternative that would let a team validate the milestone concept before committing to the full timeline visualization. The "只做数据层，不做可视化" alternative is dismissed as "PM 体验差" without evidence — has any PM been asked whether a simple grouped list view would be acceptable as an interim step?
**What must improve**: Add a staged-delivery alternative (data model + simple list/grouped view first, timeline in phase 2). Quantify "开发量较大" (e.g., estimated sprint count or story points). Explain why a single delivery is preferred over incremental — is there a deadline? A stakeholder commitment?

### Attack 2: Problem Definition — Evidence is circumstantial, not primary
**Where**: "PM 目前通过命名约定（如"v1.1-需求阶段"）和记忆管理阶段关系"
**Why it's weak**: The evidence section lists the absence of features (4 views are MI-dimension, no grouping mechanism) and describes inferred PM behavior (naming conventions, memory-based tracking). However, there is no direct user evidence — no PM quote saying "I struggle to track milestones", no support ticket, no user interview data, no analytics showing PMs spending excessive time in list views. The urgency section's cost quantification (8 hours/month) appears to be a calculated estimate, not a measured observation. The proposal is building a solution based on a problem the author inferred rather than one users reported.
**What must improve**: Add at least one piece of primary evidence: a PM quote, a support request, a user interview finding, or analytics data (e.g., "average PM views the Gantt chart N times per session, suggesting difficulty finding phase-level information"). If no direct evidence exists, acknowledge this gap honestly and state that the proposal includes a user validation step before full development.

### Attack 3: Solution Clarity — No performance or scale considerations for the timeline view
**Where**: "时间线图：横向时间轴，里程碑为节点，关联 MI 按时间排列并连线到对应里程碑"
**Why it's weak**: The solution describes the timeline visualization without any performance parameters. How many milestones and MIs should the timeline handle before it becomes unusable? The proposal states projects have 30-50 MIs across 3-4 phases. At what scale does the timeline break down? 100 MIs? 200? Is there pagination or virtual scrolling? The risk table addresses visualization complexity from an implementation perspective (library choice), but the solution description itself does not state the expected scale boundaries or degradation strategy. A developer reading this cannot make informed decisions about rendering approach without knowing the target data volume.
**What must improve**: Add a scale target to the solution description: "Timeline must render up to N milestones with up to M associated MIs per milestone (total ~X nodes) at 60fps." State the degradation strategy: pagination, virtual scrolling, or clustering beyond a threshold. This feeds directly into the library evaluation criteria.

---

## Verdict

- **Score**: 88/100
- **Target**: 90/100
- **Gap**: 2 points
- **Action**: Continue to iteration 3
