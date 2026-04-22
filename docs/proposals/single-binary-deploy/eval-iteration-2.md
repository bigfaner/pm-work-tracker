# Proposal Evaluation Report

**Proposal:** 前后端单体部署（Go embed.FS）
**Iteration:** 2
**Date:** 2026-04-21
**Evaluator:** adversarial-scorer

---

## Overall Score: 72/100

| Dimension               | Score | Max |
|-------------------------|-------|-----|
| Problem Definition      | 15    | 20  |
| Solution Clarity        | 14    | 20  |
| Alternatives Analysis   | 11    | 15  |
| Scope Definition        | 13    | 15  |
| Risk Assessment         | 10    | 15  |
| Success Criteria        | 9     | 15  |

---

## Previous Iteration Addressed

The three major attack points from iteration 1 were addressed as follows:

1. **Success Criteria (was 0/15, now 9/15):** Fully addressed. A dedicated section with 7 numbered acceptance criteria now exists, each tied to a specific in-scope deliverable and verifiable by inspection or test.

2. **Risk Assessment (was 2/15, now 10/15):** Fully addressed. Three risks are now identified (SPA routing, cache control, config misdeployment) with likelihood/impact ratings and concrete mitigations. The SPA routing risk (R1) is the single strongest item in the entire proposal — it identifies a specific technical failure mode and prescribes an exact code-level fix.

3. **Problem Definition — Urgency (was 1/6, now 5/6):** Fully addressed. The proposal now states a concrete urgency trigger: "项目进入功能验收阶段，远程团队成员（产品经理、测试）需要访问 dev 环境" with the specific operational pain of requiring two local processes for remote validation.

---

## Dimension Details

### 1. Problem Definition: 15/20

**Problem stated clearly (6/7):**
The four bullet points unambiguously describe the deployment gap. The problem is that the project has no production deployment mechanism and the frontend build output has no hosting strategy. Two readers would arrive at the same interpretation. Minor deduction: the "本地开发体验良好" line at the end of the Problem section is a non-problem statement that dilutes focus. It belongs in context/background, not in the problem definition.

**Evidence provided (4/7):**
The four bullets are self-reported observations ("no Dockerfile exists," "no deployment scripts exist"), not external evidence. No incident report, no stakeholder complaint, no failed deployment log is cited. However, for a greenfield deployment problem where nothing exists, "these things do not exist" is a form of evidence — just not a strong one. The urgency section partially compensates by citing the specific pain of cross-team validation.

**Urgency justified (5/6):**
Strong urgency case: the project is in acceptance phase, remote team members need dev environment access, and current validation requires starting two local processes. This is concrete and specific. Minor deduction: no deadline or target date is given — "when does this need to be ready?" remains unanswered.

### 2. Solution Clarity: 14/20

**Approach is concrete (6/7):**
The build flow diagram (3-step), runtime routing table, and explicit file-level change list (router.go, static.go, build.sh) make the approach implementable. A developer could pick this up and code it. Minor gap: the routing section says `/static/*` serves embedded files, but Vite's default output uses subdirectories like `assets/` not `static/`. The exact mapping between the Vite build output structure and the Go embed path is left implicit.

**User-facing behavior described (5/7):**
The deployer experience is well-described (scp binary + config, single command). The end-user experience is partially covered — success criteria #1 and #2 describe what HTTP responses the application user would see. However, operational behaviors like startup time, page load latency, or error states the user sees on misconfiguration are not described.

**Distinguishes from alternatives (3/6):**
The Recommended Approach section lists 4 reasons but they are high-level assertions rather than discriminating analysis:
- "与'单体部署'目标完全吻合" is circular — the goal was defined to match this solution.
- "部署只需二进制 + config" is a restatement, not a distinction.
- "Go embed 是标准库特性" is the one genuinely discriminating reason.
The proposal does not explain why the standard-library advantage outweighs Docker's portability or Nginx's caching maturity. The differentiator section is the weakest part of Solution Clarity.

### 3. Alternatives Analysis: 11/15

**At least 2 alternatives listed (4/5):**
Three alternatives are presented (embed, Nginx, Docker). The "do nothing" alternative is still not formally named and analyzed, though the Problem section implicitly describes the current state. For a proposal about adding deployment capability, "remain without deployment" is a legitimate alternative that should be explicitly evaluated.

**Pros/cons for each (4/5):**
Each alternative now has a business-value/costs structure. Alternative B's costs section ("需维护 Nginx 配置，部署两个进程") is a genuine trade-off, not circular reasoning as in iteration 1. Alternative C's costs ("Docker 抽象层收益不大") still lacks a threshold — at what team size or server count would Docker become worthwhile? But the reasoning is defensible for the stated context (single developer, single server).

**Rationale for chosen approach (3/5):**
Four numbered reasons are provided. They are logical but remain a list rather than a systematic comparison. A comparison matrix (criteria x alternatives) would make the verdict more convincing. The rationale also does not address the strongest argument *against* Approach A — that any frontend change requires a full binary rebuild and redeploy.

### 4. Scope Definition: 13/15

**In-scope items are concrete (5/5):**
Five specific deliverables with file names and behaviors. Each is a concrete artifact: router routes, embed declaration file, build script, config files, .gitignore entry. The Build Script Design subsection adds specificity about the script's interface and behavior.

**Out-of-scope explicitly listed (4/5):**
Four items are explicitly deferred: Docker/CI/CD, local dev changes, Nginx, process management. These are well-chosen. Minor gap: monitoring/logging/observability is not mentioned as in or out of scope — the current deployment presumably has no logging strategy, and the proposal does not address whether the single-binary approach adds any.

**Scope is bounded (4/5):**
The scope is achievable by a single developer in a bounded timeframe. The k8s/ConfigMap discussion from iteration 1 has been removed, which eliminates the previous inconsistency. Minor deduction: the Build Script Design section sits between Scope and Risk Assessment, which is an awkward placement — it is solution detail, not scope definition. This creates confusion about whether the build script design is a hard requirement or a suggestion.

### 5. Risk Assessment: 10/15

**Risks identified (4/5):**
Three meaningful risks are now identified. R1 (SPA routing conflict) is excellent — it names a specific technical failure mode that anyone implementing embed.FS would encounter. R2 (cache control) and R3 (config misdeployment) are also real operational risks. Minor gap: no risk addresses the build-time coupling between frontend and backend (e.g., a frontend-only CSS fix still requires a full Go rebuild and redeploy).

**Likelihood + impact rated (3/5):**
Ratings are provided (R1: high/high, R2: medium/medium, R3: medium/high). R1 is rated "高" likelihood which is honest — this *will* happen without mitigation. R2's "中/中" rating is defensible. R3's "中" likelihood is slightly generous — with placeholder validation, the likelihood should drop, but the rating appears to be pre-mitigation. The ratings are present but not deeply justified — why is R2's impact only "中"? A user seeing a white screen due to stale assets could consider that "高" impact.

**Mitigations are actionable (3/5):**
R1's mitigation is highly actionable: "最后注册 NoRoute handler 返回 embed.FS 中的 index.html." This is a code-level instruction. R2's mitigation describes specific Cache-Control headers which is also actionable. R3's mitigation (startup validation of required fields) is actionable but incomplete — what happens after the error message? Does it exit? Does it log? The success criteria #6 clarifies this ("进程退出并打印包含 'PLACEHOLDER' 的错误信息"), so the information exists but is split across sections.

### 6. Success Criteria: 9/15

**Criteria are measurable (4/5):**
Most criteria are specific enough to verify:
- #1: GET / returns 200 + index.html content; GET /static/*.js returns JS with correct Content-Type
- #4: build.sh dev succeeds on dev branch, fails on non-dev branch with non-zero exit code
- #6: Service exits with error containing "PLACEHOLDER"

These are objective. Minor gap: #5 (config files "提交到 git") is binary (committed or not) but #7 ("make dev... 热更新正常") is subjective — what does "正常" mean? No hot-reload timeout is specified.

**Coverage is complete (3/5):**
The 7 criteria cover: embed routing (#1), SPA fallback (#2), API regression (#3), build script (#4), config files (#5), config validation (#6), local dev (#7). Missing coverage:
- No criterion for binary size or build time
- No criterion for the Cache-Control headers described in R2's mitigation
- No criterion for the `/static/*` route specifically serving files from the correct Vite output directory (assets/ vs static/)

**Criteria are testable (2/5):**
Criteria #1, #2, #3, #4, #6 are testable — they specify HTTP status codes, exit codes, or error messages. However:
- #5 ("提交到 git") is testable by `git ls-files` but trivial
- #7 ("热更新正常") is not testable without defining what "正常" means
- No criterion specifies how to verify cache headers — this is a gap given that R2's mitigation depends on them

**Vague language penalty (-0):**
No instances of unquantified "better/improved/enhanced" detected. The proposal uses specific technical language throughout.

---

## Top 3 Attacks

1. **Success Criteria — Coverage gaps:** Two in-scope items lack corresponding success criteria. The Cache-Control header strategy described in R2's mitigation has no verification criterion — a developer could implement the risk mitigation incorrectly or not at all and still pass all 7 criteria. The build script's `ENV=prod` behavior is specified only for `dev` (criterion #4) — what happens when `build.sh prod` runs on a non-main branch? Add criteria for cache header correctness and prod branch validation.

2. **Solution Clarity — Vite output path ambiguity:** The routing section states `/static/*` serves static files from embed.FS, but Vite's default build output places hashed assets in `dist/assets/`, not `dist/static/`. The proposal does not specify how this mapping works — does the Go server serve from `/assets/*` or `/static/*`? Does the embed path capture the Vite output directory as-is or is a rename step needed? This ambiguity will cause the first implementation attempt to fail. Quote: "`/static/*` → 从 embed.FS 提供静态文件（JS/CSS/图片等）"

3. **Alternatives Analysis — Missing rebuild coupling trade-off:** The strongest argument *against* Approach A is not analyzed: any frontend change (even a CSS color tweak) requires a full Go binary rebuild and server restart. This is a real operational cost that Alternatives B and C do not impose (they allow independent frontend deployment). The proposal's "代价" section mentions "每次前端改动都需要重新 go build" but dismisses it with "CI 中自动化后无感知" — yet CI is explicitly out of scope. If CI is out of scope, the rebuild cost is not mitigated. Either bring CI into scope or acknowledge the unmitigated coupling cost.

---

## Additional Observations

- The document structure has improved significantly from iteration 1. The removal of the k8s/ConfigMap discussion eliminates a major inconsistency.
- The Risk Assessment section is now one of the strongest parts of the proposal — R1 in particular shows deep technical understanding of the embed.FS + SPA interaction.
- The Build Script Design subsection placement between Scope and Risk Assessment is awkward. It is solution detail and should be part of the Proposed Solution section.
- Success criteria #7 ("本地开发不受影响") tests a negative — that nothing changed. This is a valid criterion but should be verified by running the existing test suite, not by manual "热更新正常" inspection.
- The environment configuration table (local/dev/prod with branch names and database types) is useful context but the branch naming convention (dev branch for dev environment) is presented without rationale. Why not use feature branches that merge to dev?
