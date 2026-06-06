---
iteration: 1
total_score: 560
pass: false
dimensions:
  completeness: 115
  semantic_purity: 130
  precondition_exclusivity: 95
  fact_alignment: 80
  surface_fitness: 80
  internal_consistency: 60
---

# Journey Eval Report: list-filtering-and-sorting (Iteration 1)

**Date**: 2026-06-04
**Journey**: `list-filtering-and-sorting`
**Surface Types**: web, api

---

## Dimension Scores

### 1. Completeness — 115/200 (min 120) -- FAIL

| Criterion | Score | Notes |
|-----------|-------|-------|
| Journey metadata | 40/50 | Name is kebab-case. Risk level "Medium" is justified. Surface types declared as `["web", "api"]`. Missing: no `fact_table` or `facts` section linking to verifiable source facts. |
| Steps complete | 45/80 | 5 happy-path steps + 4 edge-case steps present. Each step has action and expected outcome. However: (1) Steps lack formal Step numbering in a structured format; (2) Several steps conflate multiple behaviors into one step (e.g., Step 1 covers both assignee filtering AND sub-item penetration AND visual indicator). (3) No explicit API steps at all -- the journey is entirely written from a web UI perspective despite declaring `api` as a surface type. |
| Outcomes cover required derived scenarios | 30/70 | Happy path outcomes are present. **Critical gaps**: (1) No `validation-error` outcome for web (per surface-web rules). (2) No `session-expired` outcome for web (per surface-web rules). (3) No `unauthorized` outcome for api (per surface-api rules). (4) The edge case "Performance under large dataset" (Step 4b) is a non-functional performance concern, not a functional boundary outcome. Missing: not-found, network-error for web; not-found, validation-error for api. |

**Deductions**: No formal Outcome structure per step (multiple outcomes with distinct preconditions); API surface completely absent from step definitions.

### 2. Semantic Purity — 130/200 (min 120) -- PASS

| Criterion | Score | Notes |
|-----------|-------|-------|
| Natural language outcomes | 60/80 | Outcomes are written in natural language describing what the user observes. One minor issue: "500ms" in Step 4b is a technical performance metric rather than a user-observable state ("renders without noticeable lag" is better). No regex, CSS selectors, or framework assertions found. |
| Declarative preconditions | 40/60 | Preconditions in edge cases are mostly declarative ("Assignee C is only responsible for main items"). However, Setup section mixes setup instructions with state declarations. The Setup reads more like test data preparation than declarative preconditions. |
| No implementation coupling in steps | 30/60 | Steps are mostly user-level actions. However, Step 4b ("Filter response returns within 500ms") leaks implementation detail about response timing. The concept of "sub-item penetration" is a business rule, not implementation, so it passes. Overall clean but with minor leaks. |

### 3. Precondition Exclusivity — 95/150 (min 90) -- PASS

| Criterion | Score | Notes |
|-----------|-------|-------|
| Preconditions distinct across outcomes | 35/60 | Most outcomes within steps have a single outcome per step, so there is minimal ambiguity. However, the happy-path steps (Steps 1-5) each have only one outcome, meaning precondition exclusivity is trivially satisfied but not meaningfully tested. Edge cases have distinct preconditions. -5 for Step 3 and Step 3b: both operate on the progress page, and Step 3's "first time" vs Step 3b's "progress page is loaded" overlap semantically (loading the page IS opening it for the first time or subsequently). |
| Preconditions sufficient to uniquely select outcome | 35/50 | The structure generally works because each step has a single outcome. However, several error/boundary scenarios lack explicit preconditions stating what triggers them. Step 5 (Empty state) lacks a precondition -- it should state "Filters are applied that match no items in the system." Step 4b lacks a proper boundary trigger precondition. |
| No missing preconditions for error/boundary | 25/40 | Step 5's precondition is implicit (filters match nothing). Step 4b has a precondition ("1000 main items and 5000 sub-items exist") but no trigger condition explaining what makes this a boundary case vs normal operation. Missing preconditions for mandatory derived outcomes that are absent entirely (validation-error, session-expired, unauthorized). |

**Deductions**: -20 for overlapping precondition between Step 3 and Step 3b.

### 4. Fact Alignment — 80/150 (min 90) -- FAIL

| Criterion | Score | Notes |
|-----------|-------|-------|
| Factual claims traceable to fact_id or marked UNKNOWN | 20/60 | The journey has a `sources` section referencing prd-user-stories.md (Story 8, Story 9) and prd-spec.md. However, NO individual outcome or claim is traced to a specific fact_id. There is no fact table in the journey. Claims like "matched via sub-item" visual indicator, terminal status sorting, and "In progress" default checkbox are grounded in the PRD stories, but lack formal traceability. -10 for each of 4 unverified claims without UNKNOWN marking. |
| Inferred claims have required_outcomes support and source: inferred | 25/50 | The edge cases (Steps 1b-4b) appear to be inferred boundary outcomes, but none are annotated with `source: inferred` or cite which `required_outcomes` rule triggered their generation. The mandatory derived outcomes (validation-error, session-expired, unauthorized) are entirely absent -- zero support for the surface type's required_outcomes rules. |
| No hallucinated claims without classification | 35/40 | Most claims are reasonable and align with the PRD stories. One potential hallucination: the empty state message "No items match the criteria" does not appear in the PRD. The PRD says "没有符合条件的事项" (Chinese). The journey uses an English translation that differs from the PRD's actual text. This is not marked as UNKNOWN or inferred. -5 for unclassified deviation from source material. |

**Deductions**: -30 for hallucinated empty state message text that deviates from PRD without classification.

### 5. Surface Fitness — 80/150 (min 90) -- FAIL

| Criterion | Score | Notes |
|-----------|-------|-------|
| Mandatory derived outcomes present | 0/60 | **Critical failure**: None of the mandatory derived outcomes are present. Web requires `validation-error` and `session-expired` -- neither exists. API requires `unauthorized` for authenticated endpoints -- does not exist. The journey does not include ANY API-specific steps despite declaring `api` as a surface type. Score 0 per rubric rule: "Score 0 if mandatory Outcomes are completely absent." |
| Test strategy proportions match surface guidance | 40/50 | Web and API both call for balanced 50/50 Contract/Journey. The journey covers user workflows adequately for web. However, the API surface is entirely missing, making the effective proportion 100% web / 0% api. -10 for imbalance. |
| Surface-specific environment assumptions realistic | 40/40 | Web assumptions (filter selection, checkbox, page load, visual indicators) are realistic for browser interaction. No CLI/TUI or mobile assumptions leaked in. The progress page filter interaction is well-described for web. |

**Deductions**: -25 per surface type violation (api surface declared but zero api content). Single instance.

### 6. Internal Consistency — 60/150 (min 90) -- FAIL

| Criterion | Score | Notes |
|-----------|-------|-------|
| Invariants hold in every step | 30/60 | 5 invariants declared. Checking: (1) "Assignee filter always penetrates" -- holds in Steps 1, 2b, 4b. (2) "Terminal items sort to bottom" -- holds in Step 2, 3b. (3) "No filters = all items" -- holds in Step 1b. (4) "Empty filter results show clear empty state" -- holds in Step 5. (5) "Sub-item match indicator displayed" -- holds in Step 1. **However**: Invariant 2 ("terminal items always sort to bottom regardless of active filters") is potentially violated by Step 4 ("deselect all status checkboxes shows all items including terminal"). If terminal items are shown among all items without sorting to the bottom, the invariant is violated. The step says "All items are displayed" but does not confirm terminal sorting still applies. |
| Cross-step references consistent | 15/50 | Steps are mostly self-contained with minimal cross-references. However: Step 1 references "assignee A" defined in Setup, and Step 2b references "Assignee C" defined in its own precondition -- these are fine. The issue is that the Setup section defines data relationships (assignees, statuses) but Steps 2-5 do not reference the setup data explicitly, creating ambiguity about which items are being acted upon. Step 3 references "first time" which has no prior state establishment. |
| Risk level consistent with content | 15/40 | Risk level "Medium" is justified ("multi-step interaction without irreversible side effects"). This is appropriate for filter/sort operations. However, declaring two surface types (web + api) but only covering web makes the scope assessment inconsistent with the content. |

**Deductions**: -40 for potential invariant violation (terminal sorting not confirmed in Step 4 where all items are shown).

---

## Summary of Critical Issues

1. **Missing mandatory derived outcomes** (Surface Fitness, 0/60): Web requires `validation-error` and `session-expired`. API requires `unauthorized`. None are present. This is an automatic failure for Surface Fitness.

2. **API surface completely absent** (Completeness, Surface Fitness): The journey declares `surface_types: ["web", "api"]` but contains zero API-specific steps (no HTTP method descriptions, no status code assertions, no authentication scenarios).

3. **No fact traceability** (Fact Alignment): Claims are grounded in PRD content but lack formal `fact_id` references. No fact table exists. Inferred outcomes lack `source: inferred` annotations.

4. **Invariant violation risk** (Internal Consistency): Step 4 (clear all filters) says "All items are displayed" but does not confirm that terminal sorting invariant still applies when all filters are cleared.

5. **Empty state message mismatch** (Fact Alignment): Journey says "No items match the criteria" but PRD spec says "没有符合条件的事项" -- the English translation differs and is not marked as inferred or UNKNOWN.

---

## Pass/Fail Status

| Dimension | Score | Min | Status |
|-----------|-------|-----|--------|
| Completeness | 115/200 | 120 | **FAIL** |
| Semantic Purity | 130/200 | 120 | PASS |
| Precondition Exclusivity | 95/150 | 90 | PASS |
| Fact Alignment | 80/150 | 90 | **FAIL** |
| Surface Fitness | 80/150 | 90 | **FAIL** |
| Internal Consistency | 60/150 | 90 | **FAIL** |
| **Total** | **560/1000** | **850** | **FAIL** |

**Result**: FAIL -- 4 of 6 dimensions below threshold, total 290 points below target.
