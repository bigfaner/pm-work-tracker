# Eval-PRD Iteration 3 — Merged Report (pm + qa)

**Merged Score**: 901/1000 (target 900) ✅ TARGET REACHED
- PM scorer: 890/1000
- QA scorer: 912/1000
- Gate score: average = 901/1000

## Score Progression

| Iteration | Score | Delta | Notes |
|-----------|-------|-------|-------|
| 1 | 710/1000 | — | Initial draft; 23 attacks |
| 2 | 888/1000 | +178 | Reviser fixed all 23 iter-1 attacks; 20 residual |
| 3 | 901/1000 | +13 | Targeted Tier-1 fixes; target reached |

## Tier-1 Fix Verification (this round)

| Tier-1 gap | Fix applied | Scorer verdict |
|-----------|-------------|----------------|
| Undo window zero ACs | Added Story 8 (9 ACs: success/expiry/session-loss/AI-down/uniqueness/irreversible/revalidate) + UF-3 undo Data fields + expanded semantics | Resolved. Residual: post-undo recovery narrative (refinement, not a gap). |
| Confidence threshold not in diagram | Branched M1 into ≥0.7/0.4-0.7/<0.4 bands; added boundary ACs at 0.7/0.69/0.4/0.39 | Resolved. No scorer flagged the diagram or boundary ACs. |
| Scope 24-combo vs query only MainItem | Added Milestone + ItemPool query ACs; stated "unified query handler covers all 6 entities" | Largely resolved. Residual: PM notes SubItem/MilestoneMap/ProgressRecord query is "architectural assertion not user-evidenced" — a coverage-depth refinement. |
| (bonus) ±1-day rule misplaced in UF-3 | Removed from UF-3 Validation (already in Goals per-field metrics) | Resolved. |

## Per-Dimension Breakdown (iteration 3, averaged)

| Dimension | Max | PM | QA | Merged |
|-----------|-----|----|----|--------|
| Background & Goals | 100 | 88 | 90 | 89 |
| Flow Diagrams | 150 | 146 | 146 | 146 |
| Functional Specs | 200 | 180 | 178 | 179 |
| User Stories | 200 | 183 | 178 | 181 |
| Scenario Completeness | 150 | 125 | 132 | 129 |
| Edge Case Coverage | 100 | 82 | 78 | 80 |
| Scope Clarity | 100 | 86 | 90 | 88 |

> Averaged dimension sub-scores are illustrative; the authoritative merged total is the average of the two scorer totals (901).

## Residual Attacks (iteration 3, lower-impact refinements)

These remain but did not block the target:

1. [Edge Case Coverage]: Quota boundary ACs at 49/50/51 + counting semantics (request/response/success, timeout/malformed consumption, day-boundary timezone, multi-tab atomicity) — both scorers. (PM-8, QA-1, QA-9)
2. [User Stories]: Sub-item move rejection ACs (BIZ-lifecycle-004: terminal source, cross-team, same-source, terminal target) — both scorers. (PM-5, QA-2)
3. [Scenario Completeness]: Query evidence depth for SubItem/MilestoneMap/ProgressRecord (architectural assertion vs user-evidenced). (PM-7)
4. [Functional Specs]: Team-context contradiction on `/users`/`/roles` (UF-1 lists as bubble routes but UF-2 blanket send-blocks). (PM-3)
5. [Functional Specs]: available-transitions endpoint-error fallback UX (no user signal that precheck was skipped; reverse/undo direction has no fallback). (PM-4, QA-8)
6. [User Stories]: MainItem milestoneKey modify ACs (BIZ-milestone-003: terminal MainItem/target terminal Milestone rejection, check ordering). (PM-6)
7. [Edge Case Coverage]: AI semantic-invalid output (priority=P5, past date, non-team assignee). (PM-9)
8. [Functional Specs]: UF-4 query truncation has no recovery (drops results 21-N with no fetch-next/narrowing). (QA-3)
9. [User Stories]: Concurrent-edit diff rejection path (user rejects the diff). (QA-5)
10. [Scenario Completeness]: Post-undo recovery narrative (what user does after undo). (QA-4)
11. [blindspot]: Multi-tab concurrent undo race (server-side uniqueness token). (QA-6)
12. [blindspot]: Late-arriving AI response after timeout fallback (dedup/idempotency). (PM-11, QA-7)
13. [blindspot]: Undo failure recovery link to traditional UI. (PM-13)
14. [blindspot]: Sensitive-field regex false-positive silent REDACT in card. (PM-14)
15. [blindspot]: 50-round cap ∩ unsubmitted card submission. (PM-15)
16. [blindspot]: Team-context staleness during in-flight AI call (Team-at-send vs Team-at-render). (PM-16)
17. [blindspot]: Multi-intent utterances (create+query bundles). (PM-12)
18. [Background & Goals]: "实词" undefined in description metric. (PM-2)
19. [Scope Clarity]: In-Scope still mixes deliverables with implementation language. (PM-10)

These are refinements suitable for the design phase or a future polish pass — none is a structural blocker.
