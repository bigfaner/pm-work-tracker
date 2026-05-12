---
date: "2026-05-12"
doc_dir: "docs/proposals/milestone-map/"
iteration: 3
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 3

**Score: 90/100** (target: 90)

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 15 | 20 |
| Solution Clarity | 17 | 20 |
| Alternatives Analysis | 14 | 15 |
| Scope Definition | 15 | 15 |
| Risk Assessment | 14 | 15 |
| Success Criteria | 15 | 15 |

---

## Deductions

| Rule | Instance | Deduction |
|------|----------|-----------|
| Missing detail | Evidence is inferred, not primary: "以上痛点基于开发者对 PM 日常工作流的观察推断...尚未进行正式用户访谈或问卷调查". The proposal acknowledges this gap honestly, but building a 5-7 day feature without any direct user validation remains a weakness. Cost quantification (8 hrs/month) is a calculated estimate, not a measured observation. | -2 (Problem Definition: Evidence) |
| Vague language | Problem statement uses "一组事项共同达成某个目标" without defining what a "目标" is. Is it a release? A sprint? A project phase? Two readers could interpret this differently. | -1 (Problem Definition: Clarity) |
| Missing detail | Urgency is framed as efficiency gain (saving 8 hrs/month), not a blocking issue. The proposal does not acknowledge that this is a "nice to have" efficiency improvement rather than a critical capability gap. What happens if we don't build this? PMs spend 25 min/week -- not ideal, but not catastrophic. | -2 (Problem Definition: Urgency) |
| Missing detail | User-facing visual description is thin: "横向时间轴，里程碑为节点，关联 MI 按时间排列并连线到对应里程碑". Does the user see a Gantt-like bar chart? A node graph? A swim-lane diagram? Card-based layout? A developer or designer cannot implement this from the description alone. | -1 (Solution Clarity: User-facing behavior) |
| Missing detail | Solution does not articulate why timeline visualization is the differentiator vs. a simple grouped list with progress bars. The proposal asserts "时间线可视化是里程碑概念的核心价值" but provides no evidence or reasoning for why a grouped list wouldn't serve 80% of the need. | -2 (Solution Clarity: Distinguishes) |
| Logical gap | Chosen approach rationale uses circular reasoning: defines core value as timeline visualization, then argues staged delivery fails because it lacks timeline visualization. But this assumption is untested (acknowledged in evidence section). If PMs find grouped lists sufficient, the staged approach would be superior at lower cost and risk. | -1 (Alternatives Analysis: Rationale) |
| Missing detail | Risk mitigation for vis-timeline evaluation still lacks explicit owner and deadline: "优先评估 vis-timeline" -- who evaluates? By when? This was flagged in iteration 2 and remains unaddressed. | -1 (Risk Assessment: Mitigations) |

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 from iter 2 (Alternatives - cost quantification + staged delivery alternative) | Yes | New staged-delivery alternative added: "分阶段交付：v1 数据层+列表视图，v2 时间线可视化" with v1 estimated at ~3 days. Chosen approach now quantifies effort: "开发量约 5–7 天（数据模型+API 2 天，时间线页面 2–3 天，权限+测试 1–2 天）". Verdict includes cost tradeoff: "完整方案总工时仅比分阶段 v1 多 2–4 天，但避免了两轮部署和用户适应成本". |
| Attack 2 from iter 2 (Problem - primary evidence) | Partially | Evidence section now includes honest disclaimer: "以上痛点基于开发者对 PM 日常工作流的观察推断...尚未进行正式用户访谈或问卷调查。方案开发前将在 PRD 阶段安排 PM 验证环节". The honesty is commendable and the PRD-stage validation commitment is a good process fix. However, no primary evidence (user quotes, support tickets, analytics) has been added. The problem remains inferred, not reported. |
| Attack 3 from iter 2 (Solution - scale targets) | Yes | Solution now includes explicit performance targets: "性能目标：支持单个团队最多 20 个里程碑、200 个 MI（~10 个 MI/里程碑）在视口内渲染，初始绘制 < 500ms、交互帧率 ≥ 30fps。超出 200 个 MI 时启用分页加载（按里程碑分组，默认展开前 5 个里程碑，其余折叠）". Degradation strategy is clearly stated. |

---

## Attack Points

### Attack 1: Problem Definition — Evidence remains circumstantial, not primary
**Where**: "以上痛点基于开发者对 PM 日常工作流的观察推断（命名约定"v1.1-需求阶段"等实际存在于数据中），尚未进行正式用户访谈或问卷调查"
**Why it's weak**: The proposal is asking for 5-7 days of engineering investment based on a problem the developer inferred from observing data patterns, not from any user reporting the pain. The cost quantification (8 hours/month waste) is a back-of-envelope calculation, not a measured observation. The PRD-stage validation commitment is a good process safeguard, but the proposal itself would be stronger with even one piece of primary evidence: a PM saying "I struggle to track which phase has delays", an analytics data point showing excessive time in list views, or a support ticket requesting milestone grouping. The honesty about the evidence gap is appreciated but does not fill it.
**What must improve**: Conduct a 15-minute informal PM interview before finalizing the proposal, or add analytics evidence (e.g., "PMs average N Gantt chart interactions per session, suggesting difficulty finding phase-level information"). At minimum, document one concrete PM workflow observation with specifics (which PM, which project, what they did to work around the gap).

### Attack 2: Solution Clarity — Timeline visual description is under-specified
**Where**: "时间线图：横向时间轴，里程碑为节点，关联 MI 按时间排列并连线到对应里程碑"
**Why it's weak**: This single sentence is the entire description of what the user will see. A designer or frontend developer reading this cannot determine: Is the timeline a horizontal bar chart? A node-and-edge graph? A swim-lane layout with milestones as column headers? Are MIs shown as cards, bars, or dots? How are dates displayed? What does the zoom interaction look like visually (wider/narrower time bands, or completely different layout)? The performance targets are excellent (500ms render, 30fps), but they describe constraints, not the actual visual design. For a proposal committing to a specific visualization approach (vs. alternatives), the visual description should enable a designer to create a wireframe.
**What must improve**: Add 2-3 sentences of visual layout description: "The timeline is a horizontal scrollable axis with milestone nodes positioned at their planned dates. Each milestone node expands to show its associated MIs as horizontal bars spanning their start/end dates, connected to the milestone by vertical lines. Zoom controls switch between week/month/quarter granularity, adjusting the axis scale and label format." Or include a simple ASCII wireframe or reference to an existing tool's similar view.

### Attack 3: Risk Assessment — Fallback plan creates unscoped schedule risk
**Where**: "若两库均不满足，v1 回退到静态卡片布局（无拖拽、无缩放），拖拽作为 v1.1 增强"
**Why it's weak**: The fallback to "static card layout" introduces a materially different user experience (no drag, no zoom) that was not listed as an alternative or scoped as a possible outcome. The proposal estimates 5-7 days for the full timeline with drag and zoom, but does not estimate the effort for the static card fallback or acknowledge that the fallback would deliver a significantly degraded experience. If the fallback triggers, the project might still fit within 5-7 days, but the deliverable would be closer to the rejected "只做数据层+列表视图" alternative than to the chosen "完整里程碑图" approach. This is an unmanaged scope/schedule risk. Additionally, the vis-timeline evaluation mitigation still has no owner or deadline (flagged in iteration 2, unaddressed).
**What must improve**: Either (a) add the static card fallback to the alternatives table as a conditional outcome with its own pros/cons, or (b) budget the evaluation spike as an explicit pre-development task with owner and deadline (e.g., "Frontend lead evaluates vis-timeline by [date]; if unsuitable, scope reverts to static cards and proposal is re-evaluated"). Acknowledge that the fallback scenario partially converges with the rejected staged-delivery alternative.

---

## Verdict

- **Score**: 90/100
- **Target**: 90/100
- **Gap**: 0 points
- **Action**: Target reached
