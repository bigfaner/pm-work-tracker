# Eval-PRD Iteration 1 — Merged Report (pm + qa)

**Merged Score**: 710/1000 (target 900)
- PM scorer: 722/1000
- QA scorer: 698/1000
- Gate score: average = 710/1000

**Outcome**: Below target. Revising (iteration 2).

## Per-Dimension Breakdown (averaged)

| Dimension | Max | PM | QA | Merged |
|-----------|-----|----|----|--------|
| Background & Goals | 100 | 70 | 80 | 75 |
| Flow Diagrams | 150 | 122 | 115 | 119 |
| Functional Specs | 200 | 152 | 145 | 149 |
| User Stories | 200 | 138 | 135 | 137 |
| Scenario Completeness | 150 | 96 | 95 | 96 |
| Edge Case Coverage | 100 | 64 | 50 | 57 |
| Scope Clarity | 100 | 80 | 78 | 79 |
| **Total** | **1000** | **722** | **698** | **710** |

## Merged Attack Points

1. [Flow Diagrams]: Mermaid omits 3 of 4 documented exception branches — prose lists timeout/unavailable/permission/max-length but diagram only renders the timeout dashed edge. Add unavailable, permission-denied, and input-overflow branches to the diagram. (PM-3 + QA-6)

2. [Functional Specs]: available-transitions logic error AND coverage gap — (a) "状态变更类操作必须先通过 available-transitions 预校验" misapplies a state-transition endpoint to create operations (create has no source state); error repeats in spec, UF-3, Story 2 AC4. (b) The endpoint exists only for MainItem/SubItem/MilestoneMap/Milestone but In-Scope promises 6-entity write coverage — ProgressRecord/ItemPool lack it. Reconcile: skip pre-check for create ops and for entities without the endpoint, with explicit UX. (PM-4 + QA-1)

3. [Background & Goals]: Field-accuracy goal unmeasurable — "≥80% 字段提取准确率" but per-field rule only defines date (±1d) and assignee (fuzzy); title/priority/description/milestoneKey undefined. Define per-field matching rules for all fields. (PM-2)

4. [Background & Goals]: Goal baseline unverified — "8-12 次... 待埋点验证基线" while the goal claims "从 ~8-12 降至 ≤3". Either run baseline pre-launch or hedge the ≤3 goal to a relative reduction. (PM-1)

5. [User Stories]: Story coverage insufficient — In-Scope promises "创建/查询/修改/分配 × 6 实体" (24 combos) but stories evidence ~10; query only covered for MainItem; Milestone/MilestoneMap/ProgressRecord/ItemPool lack full submit→confirm lifecycle ACs (stories stop at card push). Add full-lifecycle submit→success ACs for Milestone, MilestoneMap, ProgressRecord, ItemPool. (PM-5 + QA-4)

6. [User Stories]: ACs not objectively verifiable — "系统在 P95 < 5 秒内推送" uses a population statistic as a single-call assertion; "milestoneKey=留空高亮"/"返回错误说明并列出合法目标状态" lack objectively assertable structure. Rewrite as per-call thresholds and structured payloads. (QA-2)

7. [User Stories]: Boundary/error AC coverage near zero — no AC for max input length, daily quota hit, confidence threshold, malformed AI output, concurrent card edits, abandoned cards, Team-context-missing, or the 50-round session cap. Add explicit boundary + negative ACs. (QA-3)

8. [Scenario Completeness]: BIZ-milestone-005 unhandled — Story 1 says "若 Team 下仅一个 MilestoneMap 自动关联" but no rule rejects Milestone creation when the MilestoneMap is terminal (Map terminal → no Milestone create). Add scenario + AC respecting the parent-terminal block. (PM-6)

9. [Scenario Completeness]: Business-rule consistency gaps — (a) BIZ-filter-001 direct+indirect assignee query semantics not stated for Copilot queries ("我的 P0 事项" — direct only or direct+indirect?), risking inconsistency with existing UI. (b) BIZ-lifecycle-004 sub-item move has no Copilot story though "修改" is supported. Add a query-matching rule + a move story or explicit exclusion. (PM-7 + QA-7)

10. [Edge Case Coverage]: Missing critical error paths — AI returns malformed/unparseable output; AI returns entity outside user's Team/permission scope; AI returns stale bizKey that no longer exists; available-transitions endpoint itself errors. Define handling for each. (PM-8)

11. [Edge Case Coverage]: Boundary values unspecified — "限制单次最大长度" (no number); "每用户每日调用次数达上限" (no cap value); "结果过多时分页或限制展示数量" (no limit). Add concrete numbers. (PM-9)

12. [Edge Case Coverage]: Post-failure recovery undefined — UF-3 has 失败+重试 states but no scenario/AC describes whether retry re-runs pre-check, whether the card stays editable, or how partial server-side effects are handled. Define post-failure recovery explicitly. (QA-5)

13. [Scope Clarity]: In-Scope items mix deliverables with implementation language — "AI 意图识别与实体抽取" and "复用现有 API 权限检查与业务逻辑" are not concrete deliverables. Rewrite as specific deliverables. (PM-10)

14. [blindspot]: Latency budget undecomposed — "P95 < 5 秒" is one end-to-end number with no backend-proxy/AI-roundtrip/schema-load/fuzzy-match sub-budgets; P95 misses will be undiagnosable. Decompose. (PM-11)

15. [blindspot]: Daily quota value + UX undefined — "每用户每日调用次数硬上限" and "达上限时降级为关键词匹配模式" give no number and no user-facing story. Specify the numeric quota, a quota-exceeded Story/AC, and a global kill-switch threshold. (PM-12 + QA-10)

16. [blindspot]: Data retention self-contradiction — "不持久化原始消息" vs "记录每次 AI 调用的意图识别结果、字段提取结果" vs DF005 "AI 调用记录". Extracted fields may be PII; no retention/TTL/access policy for compliance stakeholders. Reconcile and define a retention policy. (PM-13)

17. [blindspot]: Wrong-entity safety net absent — "把张三的 P0 事项分配给李四" can reassign the wrong MainItem if AI picks wrong yet passes the assignee check; no undo window, no stronger target confirmation. Add a wrong-entity guard (highest operational risk). (PM-14)

18. [blindspot]: Accessibility entirely absent — no keyboard-trap, focus-management, ARIA, or screen-reader rules across UF-1..UF-7 despite the goal of lowering barriers for non-technical users; the NFR mentions Esc/Enter only. Add accessibility requirements. (PM-15)

19. [blindspot]: Concurrent card-state write race undefined — "用户可直接编辑卡片字段（onChange）或对话补充（后端解析增量变更后更新卡片 state）" — merge order between keystrokes and async AI updates undefined; define merge semantics + add an AC. (QA-8)

20. [blindspot]: ItemPool "分配" semantic ambiguity — "其中'分配'映射到实体的 assignee 字段更新" may conflict with the existing POST /item-pool/:poolId/assign review action (item_pool:review permission). Distinguish field-level assignee update from ItemPool assign workflow; state Copilot coverage per case. (QA-9)

21. [blindspot]: Intent-recognition confidence threshold absent — "决策点：意图是否被识别？" has no numeric cutoff, making the M1 decision branch untestable at the boundary. Define a threshold or explicit fallback rule + boundary ACs. (QA-11)

22. [blindspot]: Sensitive-field filter asserted but not enumerated — "敏感字段（密码、token）已在代理层过滤" gives only two examples and no mechanism. Provide an enumerated list and filter strategy so the security test plan is falsifiable. (QA-12)

23. [blindspot]: Quality Checklist self-certifies "no vague wording [x]" while the doc contains "等", "如右下角", "结果过多时". Either remove the rubber-stamp marks or verify and fix the vague instances. (PM-16)
