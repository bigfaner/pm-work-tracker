# Proposal Evaluation Report

**Proposal:** 前后端单体部署（Go embed.FS）
**Iteration:** 3
**Date:** 2026-04-21
**Evaluator:** adversarial-scorer

---

## Overall Score: 78/100

| Dimension               | Score | Max |
|-------------------------|-------|-----|
| Problem Definition      | 15    | 20  |
| Solution Clarity        | 16    | 20  |
| Alternatives Analysis   | 12    | 15  |
| Scope Definition        | 13    | 15  |
| Risk Assessment         | 11    | 15  |
| Success Criteria        | 11    | 15  |

---

## Previous Iteration Addressed

The three major attack points from iteration 2 were addressed as follows:

1. **Success Criteria -- Coverage gaps (fully addressed):** The proposal now has 9 criteria (up from 7). Criterion #5 covers prod branch validation ("scripts/build.sh prod 在 main 分支执行成功；在非 main 分支执行时脚本以非零退出码退出"). Criterion #6 covers Cache-Control headers ("GET / 响应包含 Cache-Control: no-cache；GET /assets/<hash>.js 响应包含 Cache-Control: max-age=31536000, immutable"). Both gaps are closed.

2. **Solution Clarity -- Vite output path ambiguity (fully addressed):** The routing table now correctly says `/assets/*` instead of `/static/*`. The text explicitly states: "Vite 默认产出 dist/assets/<hash>.js，base 为 '/'" and "Go embed 声明 `//go:embed dist/* dist/assets/*`，Gin 注册 /assets/* 路由指向 embed.FS 的 assets/ 子目录。无需重命名 Vite 产出文件。" The ambiguity is eliminated.

3. **Alternatives Analysis -- Missing rebuild coupling trade-off (fully addressed):** The "代价" section now explicitly acknowledges the unmitigated coupling cost: "前端改动需重新 npm run build && go build 并重启服务。CI 不在范围内，部署需手动执行 build.sh 并 scp。" It provides an acceptance rationale with a concrete threshold: "当前仅一名开发者，发布频率约每周 1-2 次，构建耗时约 2 分钟。若频率增加应引入 CI 或迁移至方案 B。" This is honest and actionable.

---

## Dimension Details

### 1. Problem Definition: 15/20

**Problem stated clearly (6/7):**
The four bullet points unambiguously describe the deployment gap. Two readers would arrive at the same interpretation. Minor deduction: the "本地开发体验良好" trailing line is a non-problem statement that dilutes the problem definition section. It belongs in context or background, not as the closing statement of the problem.

**Evidence provided (4/7):**
The four bullets are self-reported observations ("no Dockerfile exists," "no deployment scripts exist"), not external evidence. No incident report, no stakeholder complaint, no failed deployment log. For a greenfield deployment gap where nothing exists, "these things do not exist" is a form of evidence -- just not a strong one. The urgency section partially compensates by citing the specific pain of cross-team validation.

**Urgency justified (5/6):**
Strong urgency case: the project is in acceptance phase, remote team members (PM, testers) need dev environment access, and current validation requires starting two local processes. Concrete and specific. Minor deduction: no target date -- "when does this need to be ready?" remains unanswered.

### 2. Solution Clarity: 16/20

**Approach is concrete (7/7):**
The build flow (3-step), runtime routing table with correct paths (`/api/*`, `/`, `/assets/*`), explicit embed directive (`//go:embed dist/* dist/assets/*`), and file-level change list make the approach fully implementable. The Vite output path is now explicitly clarified with no ambiguity. A developer can pick this up and code it without guessing.

**User-facing behavior described (5/7):**
The deployer experience is well-described (scp binary + config, single command). The end-user experience is partially covered by success criteria #1 and #2 (HTTP responses). Gaps: no description of error states the end-user sees when things go wrong (e.g., empty embed.FS, stale binary serving outdated frontend). The operational observability story (what does the operator see in logs?) is absent.

**Distinguishes from alternatives (4/6):**
Four reasons in the Recommended Approach section. "Go embed 是标准库特性，无额外依赖" is genuinely discriminating. The "代价" section now provides honest contrast by acknowledging rebuild coupling. Remaining weakness: "与'单体部署'目标完全吻合" is still circular -- it says the solution matches the goal that was defined to match this solution. The section does not explain why standard-library advantage outweighs Docker's portability or Nginx's caching maturity at any analytical depth.

### 3. Alternatives Analysis: 12/15

**At least 2 alternatives listed (4/5):**
Three alternatives (embed, Nginx, Docker). Each has a structure. "Do nothing" (remain without deployment capability) is still not formally named and analyzed as a standalone alternative, though the Problem section describes the current state. For a proposal about adding deployment, "do nothing" is a legitimate alternative that deserves explicit evaluation with its own costs.

**Pros/cons for each (4/5):**
Each alternative now has an honest business-value/costs structure. Nginx's costs ("需维护 Nginx 配置，部署两个进程") are genuine trade-offs. Docker's costs are contextualized to the single-developer scenario. The embed approach's costs section now honestly acknowledges rebuild coupling with an acceptance threshold. Minor gap: Alternative C (Docker) lists a build time cost of "2-3 minutes" but does not compare this to the embed approach's build time, making the comparison asymmetric.

**Rationale for chosen approach (4/5):**
Four numbered reasons, logical and more balanced than iteration 2. The rebuild coupling cost is now acknowledged in the "代价" section with a clear acceptance rationale ("仅一名开发者，发布频率约每周 1-2 次") and an explicit revisit threshold ("若频率增加应引入 CI 或迁移至方案 B"). This is the right pattern. Remaining weakness: no systematic comparison matrix. The rationale is a list of assertions rather than a structured evaluation against weighted criteria.

### 4. Scope Definition: 13/15

**In-scope items are concrete (5/5):**
Five specific deliverables with file names and behaviors: router routes, embed declaration file, build script with ENV parameter, config files with placeholder strategy, .gitignore entry. Each is a concrete artifact. The Build Script Design subsection adds specificity about the script's interface (`scripts/build.sh [dev|prod]`).

**Out-of-scope explicitly listed (4/5):**
Four items explicitly deferred: Docker/CI/CD, local dev changes, Nginx, process management. Well-chosen boundaries. Minor gap: monitoring/logging/observability is not mentioned as in or out of scope. The current deployment presumably has some logging (the existing Go server), but the proposal does not address whether the single-binary approach changes the logging story or adds requirements.

**Scope is bounded (4/5):**
The scope is achievable by a single developer in a bounded timeframe. The Build Script Design section sits between Scope and Risk Assessment -- it is solution detail, not scope definition. This placement creates confusion about whether the build script design is a hard requirement (part of scope) or an implementation suggestion. Moving it into the Proposed Solution section would resolve this.

### 5. Risk Assessment: 11/15

**Risks identified (4/5):**
Three meaningful risks: R1 (SPA routing conflict) is excellent -- it identifies a specific technical failure mode inherent to embed.FS + SPA. R2 (static asset cache) and R3 (config misdeployment) are real operational risks. The rebuild coupling risk from iteration 2's attack is now acknowledged in the "代价" section rather than as a formal risk. This is acceptable but not ideal -- it means the risk table does not capture the "what if release frequency increases beyond 1-2/week" scenario that could invalidate the acceptance rationale.

**Likelihood + impact rated (3/5):**
Ratings are provided (R1: high/high, R2: medium/medium, R3: medium/high). R1's "高" likelihood is honest -- this *will* happen without mitigation. R2's "中/中" is defensible. R3's "中" likelihood is reasonable pre-mitigation. The ratings are present but not justified -- the reader cannot evaluate whether "中" is the right call without understanding the assessor's reasoning. Why is R2's impact only "中" when a stale-asset white screen blocks all users? Why is R3's likelihood "中" when placeholder validation reduces it?

**Mitigations are actionable (4/5):**
R1: "最后注册 NoRoute handler 返回 embed.FS 中的 index.html" -- code-level, highly actionable. R2: specific Cache-Control header values for different path patterns -- actionable. R3: "服务启动时校验必填字段非空非占位值，若为空则拒绝启动并打印明确错误信息" -- actionable, and now confirmed by success criterion #8. The information is split across Risk Assessment and Success Criteria but is complete. Minor gap: R2's mitigation does not specify whether Cache-Control headers are set in the Gin route handler or via middleware.

### 6. Success Criteria: 11/15

**Criteria are measurable (4/5):**
Most criteria specify exact verifiable conditions: HTTP status codes (#1, #2), Content-Type (#1), Cache-Control values (#6), non-zero exit codes (#4, #5), error message substrings (#8). Criterion #9 "热更新正常" remains subjective -- what does "正常" mean operationally? No timeout or specific behavior is defined. The remaining 8 criteria are objective and well-quantified.

**Coverage is complete (4/5):**
Nine criteria cover: embed routing (#1), SPA fallback (#2), API regression (#3), build script dev (#4), build script prod (#5), cache headers (#6), config files (#7), config validation (#8), local dev (#9). The iteration-2 gaps (cache headers, prod branch validation) are fully closed. Remaining gap: no criterion for the embed directive itself -- that `//go:embed dist/* dist/assets/*` correctly captures the Vite output. This is partially covered by #1 (serving files) but does not verify the embed path matches the Vite output structure.

**Criteria are testable (3/5):**
Criteria #1-#6 and #8 are directly testable with HTTP assertions, exit code checks, and string matching. #7 is testable via `git ls-files`. #9 ("热更新正常") is the weakest -- it requires a human to judge whether hot-reload works "normally." This could be made testable by specifying: "existing test suite passes without running build.sh." No criterion specifies how to automate verification of the Cache-Control headers -- criterion #6 describes the expected values but does not prescribe the test method.

**Vague language penalty (0):**
No instances of unquantified "better/improved/enhanced" detected. The proposal uses specific technical language throughout.

---

## Top 3 Attacks

1. **Problem Definition -- Weak evidence base:** The problem is asserted through self-observation ("后端仅提供 /api/* 接口" / "没有 Dockerfile") without any external validation. No stakeholder quote, no failed deployment incident, no team chat screenshot asking "how do I share this with the PM?" For a deployment capability proposal, the absence of evidence is particularly notable because the problem is inherently about enabling others (remote team members) -- yet no evidence from those team members is cited. Quote: "前端 npm run build 产出的 dist/ 目录无人托管" -- who discovered this? Under what circumstances? Add at least one concrete triggering event or stakeholder request.

2. **Alternatives Analysis -- Missing "do nothing" alternative:** The three alternatives are evaluated but "do nothing" (remain without deployment capability) is not formally analyzed. For a greenfield capability, the status quo is the primary alternative to beat. The Problem section describes the current pain but the Alternatives section does not weigh "accept the pain and continue local-only validation" as a legitimate option with its own costs (blocked remote validation, continued dual-process workflow). This omission makes the comparison feel one-sided. Quote the Problem: "当前每次验证需本地启动前后端两个进程，阻塞跨地域验收" -- this is the cost of "do nothing," but it should appear in the alternatives analysis, not just the problem statement.

3. **Solution Clarity -- Error state behavior undefined:** The proposal describes the happy path thoroughly (build, deploy, serve, cache) but does not address what happens when things go wrong. What does the user see if the binary starts but embed.FS is empty (build.sh ran but npm build failed silently)? What does the operator see in logs when NoRoute handler catches a request? What HTTP status code does a request for a missing asset return (404 from embed.FS or 200 with index.html fallback)? These error states affect both user experience and operational debugging. Quote: "NoRoute handler 返回 embed.FS 中的 index.html" -- but NoRoute catches *everything* not matched, including requests for missing assets. Does `/assets/nonexistent.js` return 404 or index.html?

---

## Additional Observations

- The proposal has reached a solid state. All three iteration-2 attacks were substantively addressed with specific, verifiable improvements.
- The strongest elements are: R1 (SPA routing risk with code-level mitigation), the Vite output path clarification, and the rebuild coupling acceptance rationale with explicit revisit threshold.
- The "代价" section's acceptance rationale pattern ("current state X, threshold Y for revisiting") is excellent and should be a model for other proposals.
- The Build Script Design subsection placement between Scope and Risk Assessment remains awkward. It is solution architecture and belongs in the Proposed Solution section.
- The environment configuration table (local/dev/prod) is valuable context but the branch naming convention (dev branch for dev environment) is presented without rationale. Why not use feature branches that merge to dev? This could confuse contributors.
- Criterion #9 could be strengthened by replacing "热更新正常" with "existing test suite passes and Vite dev server responds to file changes within 2 seconds" -- making it objectively testable.
