# Eval-PRD Iteration 2 — Merged Report (pm + qa)

**Merged Score**: 888/1000 (target 900)
- PM scorer: 903/1000
- QA scorer: 872/1000
- Gate score: average = 887.5 → 888/1000

**Outcome**: Target NOT reached — 2/2 iterations exhausted (12 points short).

## Score Progression

| Iteration | Score | Delta |
|-----------|-------|-------|
| 1 | 710/1000 | — |
| 2 | 888/1000 | +178 |

## Per-Dimension Breakdown (iteration 2, averaged)

| Dimension | Max | PM | QA | Merged |
|-----------|-----|----|----|--------|
| Background & Goals | 100 | 88 | 88 | 88 |
| Flow Diagrams | 150 | 140 | 138 | 139 |
| Functional Specs | 200 | 172 | 172 | 172 |
| User Stories | 200 | 184 | 160 | 172 |
| Scenario Completeness | 150 | 132 | 118 | 125 |
| Edge Case Coverage | 100 | 88 | 78 | 83 |
| Scope Clarity | 100 | 89 | 88 | 89 |

> Note: averaged dimension sub-scores are illustrative; the authoritative merged total is the average of the two scorer totals (888).

## Iteration-1 → Iteration-2 Improvement

All 23 iteration-1 attacks were addressed by the reviser. The +178-point jump reflects: fixed available-transitions logic error, expanded story coverage with full submit→success lifecycles, added Story 7 (8 boundary/error ACs), added BIZ-milestone-005 + BIZ-filter-001 + sub-item move ACs, decomposed latency budget, reconciled data retention, added wrong-entity guard + undo, added accessibility, fixed vague language.

## Residual Attack Points (merged, both scorers)

Tier 1 — flagged by BOTH scorers (highest-impact):

1. **5-min undo window under-specified (HIGHEST OPERATIONAL RISK)** — asserted in 3 places (UF-3 成功 state, UF-3 Validation, spec Security) but has ZERO ACs. Reverse-op semantics for non-idempotent state transitions undefined; panel-close/session-loss/concurrent-undo/AI-down-during-window undefined. (PM-5, PM-12, QA-1, QA-4)

2. **Confidence-threshold bands not in diagram** — ≥0.7/0.4-0.7/<0.4 defined in spec but M1 diamond collapses "not recognized" with "low confidence"; boundary ACs at 0.7/0.4/0.69 missing. (PM-2, PM-13, QA-10)

3. **Scope 24-combo claim vs ~10 evidenced** — Goal/In-Scope says "4 意图 × 6 实体" but query stories exist only for MainItem; query coverage for other 5 entities absent. Reconcile by adding query ACs or narrowing the scope claim. (PM-6, PM-11)

Tier 2 — single-scorer but concrete:

4. [Functional Specs]: ±1-day rule misplaced in UF-3 Validation Rules — it's a measurement rule; relocate to Goals/per-field metrics. (PM-3, QA-6)
5. [Functional Specs]: UF-3 undo has no Data fields — add undoDeadline/undoAvailable to Data Requirements. (PM-4)
6. [Scenario Completeness]: MainItem milestoneKey modify missing despite BIZ-milestone-003 (terminal MainItem/target terminal Milestone rejected). (PM-7)
7. [Edge Case Coverage]: AI semantic-invalid output not covered — impossible field values (priority=P5, past date, non-team assignee). (PM-8)
8. [Edge Case Coverage]: Quota counting semantics undefined — increments on request/response/success? timeouts/malformed count? day boundary? atomicity under multi-tab. Boundary ACs at 49/50/51 missing. (QA-3)
9. [Scenario Completeness]: Team-context contradictions — UF-1 lists /users,/roles as bubble routes but UF-2 blanket send-blocks and Story 7 blocks writes on /users; queries on Team-less pages? Empty-Team edge cases missing. (QA-5, QA-8)
10. [Edge Case Coverage]: Multi-tab concurrent Copilot undefined. (PM-9)
11. [Functional Specs]: available-transitions fallback UX undefined — silent/spinner/banner? (QA-7)
12. [User Stories]: Sub-item move rejection cases (BIZ-lifecycle-004: terminal/cross-team/same-source target) have no Copilot ACs. (QA-2)
13. [Background & Goals]: description-field "实词" undefined — replace with deterministic tokenization rule. (PM-1)
14. [blindspot]: Late-arriving AI response after timeout fallback — dedup/idempotency for the card that arrives after user used the fallback form. (QA-12)
15. [blindspot]: Multi-intent utterances undefined — create+assign bundles; two write intents across two entities; define reject/serialize/pick-one. (QA-13)
16. [blindspot]: Sensitive-field regex false-positives — legit input REDACTED breaks intent recognition with no user-visible notice. (QA-14)
17. [blindspot]: 50-round cap ∩ unsubmitted card — can user submit round-49 card after cap hit? (PM-14, QA-15)
18. [blindspot]: Team-context staleness during in-flight AI call — navigate to different Team while thinking; pin Team-at-send vs Team-at-render. (QA-16)
19. [Scope Clarity]: In-Scope still mixes deliverables with implementation language ("意图识别服务", "后端 AI 代理层（prompt 构造...）"). (PM-10)
20. [User Stories]: Concurrent-edit clock conflated — two near-simultaneous direct edits, sub-second direct-edit-vs-AI boundaries. (QA-9)
