# Proposal Evaluation Report

**Proposal:** 前后端单体部署（Go embed.FS）
**Iteration:** 1
**Date:** 2026-04-21
**Evaluator:** adversarial-scorer

---

## Overall Score: 43/100

| Dimension               | Score | Max |
|-------------------------|-------|-----|
| Problem Definition      | 11    | 20  |
| Solution Clarity        | 10    | 20  |
| Alternatives Analysis   | 9     | 15  |
| Scope Definition        | 11    | 15  |
| Risk Assessment         | 2     | 15  |
| Success Criteria        | 0     | 15  |

---

## Dimension Details

### 1. Problem Definition: 11/20

**Problem stated clearly (6/7):**
The core problem is well-articulated: the project has no production deployment mechanism, and the frontend build output has no hosting strategy. Two readers would arrive at the same interpretation. Minor deduction: the problem section mixes "what is missing" with "local dev is fine" — the positive statement about local dev could be separated as context rather than problem definition.

**Evidence provided (4/7):**
The four bullet points describe observable symptoms (no static file serving, no Dockerfile, etc.) but these are self-reported observations by the author, not backed by data. No evidence of actual deployment failures, no user/stakeholder complaints cited, no incident reports referenced. The evidence is circumstantial — "these things don't exist" rather than "we tried to deploy and here is what broke."

**Urgency justified (1/6):**
No urgency case is made. The proposal does not answer: Why solve this now? Is there an upcoming launch? A customer waiting? A production incident? Without urgency, this reads like a nice-to-have rather than a priority. The single mention of operational complexity is an assertion, not a justification.

### 2. Solution Clarity: 10/20

**Approach is concrete (6/7):**
The build flow diagram, runtime routing table, and explicit file-level change list make the approach implementable. A developer could read this and know what to code. The build script pseudo-code in the Scope section adds further concreteness.

**User-facing behavior described (4/7):**
The proposal describes the deployer experience (scp a binary) but not the end-user experience. What does the application user observe differently? Will page load times change? Are there new error behaviors if the binary is misconfigured? The focus is entirely on the operations side, neglecting the consumer of the deployed application.

**Distinguishes from alternatives (4/6):**
Four reasons are given for choosing Approach A over B and C. The reasons are valid at a high level (standard library, zero dependencies, single artifact). However, the distinction is weakened by vague language penalties.

**Vague language penalties (-4):**
- "部署极简" — what does "extremely simple" mean quantitatively? How many steps? How many minutes?
- "无运维依赖" — contradicted later by k8s ConfigMap/Secret discussion in the environment configuration section
- "偏重" for Docker — what does "too heavy" mean? Image size? Build time? Learning curve?

### 3. Alternatives Analysis: 9/15

**At least 2 alternatives listed (4/5):**
Three alternatives are presented: Go embed (A), Nginx reverse proxy (B), and Docker multi-stage build (C). The "do nothing" (current state) alternative is implied but not formally included as a named alternative with its own analysis.

**Pros/cons for each (2/5):**
Each alternative has a one-sentence pro and a one-sentence con. This is insufficient for decision-making:
- Alternative B's con ("与单体部署目标不符") is circular reasoning — the proposal's goal is single-binary, so of course a multi-process solution doesn't fit. The actual tradeoffs (Nginx caching, gzip, CDN integration, graceful static asset updates) are not discussed.
- Alternative C's con ("当前项目规模下偏重") is vague and unsupported. What scale threshold would make Docker appropriate?
- Alternative A's costs section mentions rebuild and binary size but does not address SPA routing challenges (client-side routing with `/` catch-all).

**Rationale for chosen approach (3/5):**
Four numbered reasons are provided. They are logical but read as post-hoc justification rather than a systematic comparison. The reasoning would be stronger with a comparison matrix or weighted criteria.

### 4. Scope Definition: 11/15

**In-scope items are concrete (4/5):**
Five specific deliverables are listed with file names and behaviors. Each is a concrete artifact. Minor gap: the "环境与配置策略" section describes k8s deployment patterns but this is not reflected in scope items.

**Out-of-scope explicitly listed (4/5):**
Four items are explicitly out of scope: Docker/CI/CD, local dev changes, Nginx, process management. These are well-chosen boundaries.

**Scope is bounded (3/5):**
The scope is achievable, but there is an inconsistency: the solution section discusses k8s ConfigMap/Secret deployment in detail, yet k8s is not in scope or out of scope. The environment configuration table references dev/prod branches and databases, but no branch strategy or database migration work is listed in scope. This creates ambiguity about what the deliverable actually includes.

### 5. Risk Assessment: 2/15

**Risks identified (1/5):**
No dedicated risk section exists. The "代价" (costs) subsection under Approach A mentions three items (rebuild cost, binary size, build tag complexity) but these are costs, not risks. Missing risks include:
- SPA client-side routing may break if the `/` catch-all route conflicts with API routes
- No cache-control headers for embedded static assets (stale assets after redeployment)
- Config file with placeholder secrets could be accidentally deployed without override
- Binary size growth over time as frontend grows
- `embed.FS` does not support hot-reload — any frontend fix requires full rebuild and redeploy

**Likelihood + impact rated (0/5):**
No risk ratings exist.

**Mitigations are actionable (1/5):**
"CI 中自动化后无感知" is mentioned for rebuild cost, but this is an assertion, not a mitigation plan. No other mitigations are provided.

### 6. Success Criteria: 0/15

**Criteria are measurable (0/5):**
No success criteria section exists. There is no definition of "done" or acceptance criteria.

**Coverage is complete (0/5):**
Without criteria, coverage cannot be assessed.

**Criteria are testable (0/5):**
Without criteria, testability cannot be assessed.

---

## Top 3 Attacks

1. **Success Criteria:** The proposal has zero success criteria — there is no definition of "done." The reader cannot determine when the implementation is complete or verify it meets requirements. Add a dedicated Success Criteria section with measurable, testable acceptance criteria for each in-scope item.

2. **Risk Assessment:** No risk section exists. SPA routing conflicts, cache-control for embedded assets, and config misdeployment are unaddressed risks that could cause production failures. Add a Risk Assessment section with at least 3 risks, each rated for likelihood and impact, with actionable mitigations.

3. **Problem Definition — Urgency:** The proposal provides no justification for why this must be solved now. Without urgency, this is a nice-to-have that could be deprioritized indefinitely. Add a concrete urgency trigger: an upcoming launch date, a stakeholder request, or a specific operational pain point with evidence.

---

## Additional Observations

- The k8s deployment discussion (ConfigMap, Secret, env override) in the solution section is detailed but disconnected from the scope. Either bring it into scope or remove it.
- The environment configuration strategy (branch names, database types) is presented as fact without rationale for the choices (why SQLite locally but MySQL in dev/prod? why separate config files per environment instead of env vars everywhere?).
- The build script design section is placed after Scope, but it is really part of the solution. The document structure mixes solution details into the scope section.
- Alternative B's analysis is a straw man — "与单体部署目标不符" dismisses it for not being the chosen approach, rather than evaluating it on its own merits.
