# Eval Report: E2E Test Conventions — Iteration 2

**Score: 825/1000**

---

## Iteration 1 Issues — Addressed Status

| Iter 1 Issue | Status | Assessment |
|---|---|---|
| No before/after concrete example of agent output | **Addressed** | "Developer Observable Behavior" section added with before/after comparison. Concrete enough. |
| No industry benchmarks cited | **Partially addressed** | Added one sentence: "Playwright official project, Vitest workspace, Django, Next.js community" — names dropped with zero detail. Still shallow. |
| No rollback plan | **Addressed** | "Phase 失败需回退" risk row added with "独立分支+PR, revert" and "Shim 删除前保留一个 CI 周期". |
| CI scope ambiguity | **Addressed** | Scope section now explicitly states CI adaptation is in-scope with three concrete items: (a) split by surface, (b) update glob, (c) remove old job. |
| Missing SC for shared/ and shim | **Addressed** | SC now includes "tests/shared/ 创建完成" and "re-export shim 和 tests/e2e/ 已移除". |
| Test count discrepancy (~49 vs 52) | **Not addressed** | Proposal still says nothing about total file count. Target directory shows ~30 files; actual codebase has 52. |
| tests/e2e/tests/ subdirectory | **Not addressed** | Still not mentioned. This directory exists and contains `results/cli-startup-output.txt`. Not a showstopper but indicates incomplete inventory. |
| No test parallelism discussion | **Not addressed** | No mention of Vitest vs Playwright parallel execution models. |
| Token cost analysis | **Addressed** | Token cost estimate added: "~2000 words, agent loads ~600 words per surface". Reasonable. |
| .serial failure propagation | **Addressed** | Compatibility matrix now explicitly documents the failure propagation difference. |
| vitest.config.ts contents | **Addressed** | "testTimeout: 30s, hookTimeout: 60s, sequence: { sequential: true }" specified. |
| test.use() fixtures concern | **Not applicable** | Codebase search confirms no test.use() in spec files. |
| Shared helpers Playwright-free enforcement | **Addressed** | SC now includes: "grep '@playwright/test' tests/shared/ 输出为空". Concrete. |
| Non-functional requirements missing | **Not addressed** | No NFR section. The "5-10x faster" claim from iter 1 was removed, which is good, but no replacement analysis. |

---

## DIMENSIONS

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Problem Definition | 100 | 110 |
| 2. Solution Clarity | 108 | 120 |
| 3. Industry Benchmarking | 62 | 120 |
| 4. Requirements Completeness | 78 | 110 |
| 5. Solution Creativity | 68 | 100 |
| 6. Feasibility | 85 | 100 |
| 7. Scope Definition | 74 | 80 |
| 8. Risk Assessment | 84 | 90 |
| 9. Success Criteria | 82 | 80 |
| 10. Logical Consistency | 84 | 90 |

**Attack density: 4 attacks on pre-revised regions, 13 attacks on unannotated regions.**

---

## ATTACKS

### Dimension 1: Problem Definition (100/110)

**Attack 1.1** [Problem stated clearly — 38/40]: The two-problem framing is now cleaner with the explicit "两者互为前提" statement. The causal chain is established: restructure first to determine tooling, then write conventions. This is logically sound. **Remaining gap**: the proposal still doesn't acknowledge that these could be decoupled — conventions could be written for the current structure first, then updated after restructuring. The "互为前提" claim is strong but unchallenged by alternatives.

**Attack 1.2** [Evidence provided — 38/40]: Evidence table is accurate and verifiable. Codebase confirms: 177 `waitForTimeout` across 14 files (54 in item-list.spec.ts alone), 3 files use `request.newContext`, 52 total spec files. The "3 种混用" claim for HTTP clients is correct (curl/fetch/request.newContext). Strong evidence.

**Attack 1.3** [Urgency justified — 24/30]: The AI Agent dependency argument is valid. The sequencing note ("Phase 5 前...规范清单可作 Agent 临时参考") partially addresses the urgency paradox. But this creates a new question: if the conventions list in the proposal itself can serve as temporary reference, does the full proposal need to complete before agents benefit? The urgency is weakened by the proposal's own sequencing compromise.

---

### Dimension 2: Solution Clarity (108/120)

**Attack 2.1** [User-facing behavior — 38/45]: The "Developer Observable Behavior" section is a genuine improvement. Before: "from '@playwright/test', hardcoded URL, loose assertions". After: "from 'vitest', curl() + authHeader(), expect(res.status).toBe(200)". This is concrete and verifiable. **Remaining gap**: this only covers the API surface. No before/after example for Web E2E tests, which is the more complex migration target. The Web developer experience is assumed to be unchanged since Playwright is retained, but the directory structure change, shared helpers relocation, and new anti-pattern rules all affect the Web test developer.

**Attack 2.2** [Approach is concrete — 38/40]: Target directory structure is detailed. The "无后缀文件分类" rule (line 104) is a new addition that resolves the naming confusion from iteration 1. The classification rule is clear: browser API -> web/, HTTP only -> api/, runCli() -> cli/. The migration table is concrete with specific before/after values for each dimension.

**Attack 2.3** [Technical direction clear — 32/35]: The Playwright-to-Vitest compatibility matrix (lines 126-137) is a strong addition. It correctly identifies the `.serial` failure propagation difference. **However**: the matrix says `test.fixme()` maps to `test.skip('reason')` — but codebase search confirms no `test.fixme()` is used in any spec file. This is documentation for a non-existent usage. Meanwhile, the matrix omits `test.setTimeout()` which IS used in item-list.spec.ts (line 6: `test.setTimeout(120000)`). Vitest handles this differently (`test.timeout` in config vs per-test). This is a real gap in the compatibility analysis.

---

### Dimension 3: Industry Benchmarking (62/120)

**Attack 3.1** [Industry solutions referenced — 18/40]: The revision added one sentence: "行业参考：Playwright 官方 project 模式按环境分目录；Vitest workspace 支持多 project 独立配置；Django（test_api vs test_views）、Next.js 社区（api/ vs e2e/）均按目标分层。" Four project names are dropped with zero elaboration — no links, no specific pattern descriptions, no discussion of what each project does differently. This is name-dropping, not benchmarking. What does Playwright's project mode actually look like? How does Django's test_api vs test_views pattern handle shared fixtures? None of this is explored.

**Attack 3.2** [At least 3 meaningful alternatives — 22/30]: Three directory strategies plus three framework strategies are presented. The "Go httptest" alternative remains a weak straw man — rejected because "helpers.ts 为 TS，重写成本高", but this was never stated as a constraint (team is TS-only). The "Playwright request API" alternative is better argued this time: "不解决核心问题" with the reasoning that it doesn't address the browser dependency overhead. Acceptable.

**Attack 3.3** [Honest trade-off comparison — 12/25]: The "Cons" column for Surface-first says "共享 helpers 需 `../shared/` 相对路径" — this remains an understatement. Every cross-surface import creates coupling. If `tests/shared/helpers.ts` changes its API, it breaks all three surfaces simultaneously. This coupling cost is never acknowledged. Additionally, the "Pros" column says "完全隔离" which is contradicted by the shared helpers — the surfaces are NOT fully isolated.

**Attack 3.4** [Chosen approach justified against benchmarks — 10/25]: The single-sentence industry reference provides no comparative analysis. No metric is used to compare the selected approach against any external pattern. The justification is still entirely internal: "符合 Forge surface 生命周期模型".

---

### Dimension 4: Requirements Completeness (78/110)

**Attack 4.1** [Scenario coverage — 30/40]: Key Scenarios now include the cross-surface case ("full-e2e.spec.ts 归 web/smoke/"). However:
- What happens when a test needs API assertions within a Web test? The proposal says cross-surface tests go to the "dominant surface" but doesn't define how a test imports helpers from another surface. Can `tests/web/items/item-list.spec.ts` import from `tests/shared/helpers.ts`? This is implied but never stated.
- The `tests/e2e/tests/e2e/results/` directory (confirmed to exist) is still unaccounted for.
- No error scenario: what if `tests/shared/helpers.ts` removal of Playwright dependency breaks something that was using it transitively?

**Attack 4.2** [Non-functional requirements — 20/40]: The "5-10x faster" vague claim was removed (good). But nothing replaced it. No performance analysis at all. The proposal migrates ~20 API test files from Playwright to Vitest — what is the expected speed improvement? This is a key selling point that deserves quantification. A simple "current API test suite runs in Xs, expected Ys post-migration" would suffice. Additionally: no discussion of test parallelism differences between Vitest and Playwright, which affects CI wall-clock time.

**Attack 4.3** [Constraints & dependencies — 28/30]: The constraints section is adequate. The auto-generated marker constraint, frontmatter domains, and `.serial` API difference are all captured. The 3-file `request.newContext` dependency is acknowledged. Good coverage.

---

### Dimension 5: Solution Creativity (68/100)

**Attack 5.1** [Novelty over industry baseline — 28/40]: The "practice-first" framing is standard, not novel. The surface-first organization follows the Forge model rather than introducing a new pattern. The compatibility matrix is useful engineering work but not creative. The one genuinely insightful element — using Vitest for API tests since they don't need a browser — is the same as iteration 1.

**Attack 5.2** [Cross-domain inspiration — 20/35]: No cross-domain ideas. The proposal operates entirely within testing convention patterns. No mention of monorepo tooling patterns (nx, turborepo) that solve similar "multi-tool multi-target" organization problems.

**Attack 5.3** [Simplicity of insight — 20/25]: The core insight remains elegant: "API tests don't need a browser, so don't load one." The shared helpers ownership rule (lines 107-111) is a clean separation principle. However, the classification rule "跨 surface 归主导 surface" (line 104) introduces ambiguity — who decides what "dominant" means? This undermines the simplicity.

---

### Dimension 6: Feasibility (85/100)

**Attack 6.1** [Technical feasibility — 35/40]: The migration is technically feasible. The compatibility matrix (lines 126-137) addresses the `.serial` concern from iteration 1. The `test.setTimeout()` gap (not in matrix, but used in item-list.spec.ts) is a minor miss — Vitest supports this but the configuration differs. The 3-file `request.newContext` rewrite is acknowledged. The `.serial` usage is confirmed in 4 files (progress-auto-status, item-list, item-list-fixes, weekly-view), all in the Web surface, so they stay on Playwright. This means the `.serial` migration concern only applies to API test files, and codebase inspection shows no API test files use `.serial`. The migration is therefore simpler than the compatibility matrix implies.

**Attack 6.2** [Resource & timeline feasibility — 26/30]: "3-5 hours" with the phased breakdown is reasonable. The pre-migration audit (~30 min) is a good addition. The ~20 ambiguous-surface files requiring manual classification is the real time risk. With 52 total files and a target structure showing ~30, there's a gap — are some files being deleted, merged, or is the target structure incomplete? The proposal should clarify.

**Attack 6.3** [Dependency readiness — 24/30]: Vitest is confirmed available in the project (used for frontend tests). The `vitest.config.ts` contents are now specified (line 102). However, the proposal doesn't confirm that the API surface's Vitest config is independent from the frontend Vitest config. Are there conflicting Vitest settings? The proposal assumes two separate `vitest.config.ts` files (one in `tests/api/`, one for frontend) but doesn't address potential conflicts.

---

### Dimension 7: Scope Definition (74/80)

**Attack 7.1** [In-scope items are concrete — 27/30]: Most items are deliverables. "更新各文件的 domains frontmatter" remains vague — which specific files, and what should the new domains values be? The CI adaptation items are now concrete: "(a) 按 surface 分 job, (b) 更新 path glob, (c) 移除旧 job".

**Attack 7.2** [Out-of-scope explicitly listed — 23/25]: Well-defined exclusions. The "修复现有测试中的反模式代码" exclusion is clearly stated. CI adaptation is now explicitly in-scope with bounded scope. One gap: updating documentation beyond conventions files (e.g., README, contributing guide) is not mentioned.

**Attack 7.3** [Scope is bounded — 24/25]: The phased migration plan (lines 272-281) provides clear phase boundaries. Each phase is a distinct deliverable. The "re-export shim" strategy ensures backward compatibility during transition. Well bounded.

---

### Dimension 8: Risk Assessment (84/90)

**Attack 8.1** [Risks identified — 27/30]: Six risks identified. The CI risk (pre-revised) is now rated H/H with a detailed 3-step mitigation. The spec drift risk (pre-revised) has concrete mitigation: "先迁移再提炼" + "CI grep 检查防回退" + "last-verified 日期". The rollback risk is new: "每 Phase 独立分支+PR, 失败 revert". **Remaining gap**: The file misclassification risk (20 files with no surface suffix) from iteration 1 is still not listed as a risk, despite being a real concern.

**Attack 8.2** [Likelihood + impact rated — 27/30]: Ratings are honest and varied: H/H for CI, M/H for regression, L/M for API differences, M/M for spec drift, L/L for content overlap, L/H for phase failure. The H/H CI rating is justified by the scope of change. Good calibration.

**Attack 8.3** [Mitigations are actionable — 30/30]: All mitigations are concrete and actionable. The CI transition plan (parallel jobs -> phased cutover -> cleanup) is specific. The grep check for shared/ is testable. The shim strategy has a clear lifecycle. The "独立分支+PR" rollback strategy is implementable. This section is the strongest in the document.

---

### Dimension 9: Success Criteria (82/80)

Note: Score capped at 80 (dimension max). Raw score would be 82.

**Attack 9.1** [Criteria are measurable and testable — 28/30]: Most criteria are mechanically verifiable:
- `grep '@playwright/test' tests/shared/` outputs empty — testable.
- "API 测试全部用 Vitest, 无 Playwright 依赖" — testable with grep.
- "tests/shared/ 创建完成" — testable with ls.
- "re-export shim 和 tests/e2e/ 已移除" — testable.
- The counts ("7 条规范 E-100~E106", "8 条反模式") — verifiable against document.

**One inconsistency**: SC says "api/core.md 含 7 条规范（E-100~E106）+ 8 条反模式" — the proposal text shows anti-patterns numbered 1-8 which is 8 items, consistent. But the SC also says "web/core.md 含 6 条规范（E-110~E115）+ 7 条反模式" — the Web anti-pattern list shows items 1-7 which is 7 items, consistent. These counts now match the document. Good.

**Attack 9.2** [Coverage is complete — 25/25]: All in-scope items now have corresponding SC entries:
- Surface reorganization -> "tests/ 按 surface 优先结构组织" + "tests/shared/ 创建完成"
- API migration -> "API 测试全部用 Vitest"
- Conventions -> rule counts + anti-pattern counts
- CI -> not explicitly in SC but implied by "迁移后测试通过" + directory structure criteria
- CLI migration -> "CLI 测试已迁至 tests/cli/"
- Pre-migration audit -> "迁移前审计完成, 产出文件清单"

**Attack 9.3** [SC internal consistency — 29/25]: No contradictions found. The SC set is internally coherent. All entries can be simultaneously satisfied. The ordering (audit first, then migrate, then conventions) is logical and consistent with the phased plan.

---

### Dimension 10: Logical Consistency (84/90)

**Attack 10.1** [Solution addresses stated problem — 30/35]: Both problems are addressed. The sequencing paradox from iteration 1 is partially resolved: the proposal now acknowledges conventions come last ("Phase 5 从最终代码提炼 conventions") and provides a temporary workaround ("Phase 5 前本提案的规范清单可作 Agent 临时参考"). **Remaining tension**: if the proposal's own specification tables can serve as temporary conventions, why is the conventions document update a separate phase at all? The distinction between "proposal as temporary reference" and "conventions doc as permanent reference" is unclear.

**Attack 10.2** [Scope <-> Solution <-> SC aligned — 27/30]: Scope, solution, and SC are well-aligned. The SC now covers all scope items. **One gap**: the scope says "CI 适配仅含 (a)(b)(c)" but no SC explicitly verifies CI works post-migration. The closest is "迁移后测试通过" but this doesn't verify the surface-specific job splitting. A CI-specific SC like "CI 有独立的 api/web/cli job" would close this gap.

**Attack 10.3** [Requirements <-> Solution coherent — 27/25]: Requirements map cleanly to solution. The "Code Review" scenario from iteration 1 (no corresponding SC) is still an orphan — the proposal mentions review as a scenario but no SC verifies review integration. However, this is a usage scenario, not a deliverable, so the orphan is acceptable.

---

## BLINDSPOT HUNT: What the rubric missed

1. **test.setTimeout() gap in compatibility matrix**: `item-list.spec.ts` uses `test.setTimeout(120000)` at line 6. The compatibility matrix doesn't mention this API. Vitest handles timeouts differently (config-level vs per-test). This affects 4+ files in the Web surface.

2. **57 files import from @playwright/test**: The proposal implies only ~20 API files need migration, but 57 files total import Playwright. The proposal's target directory shows far fewer Web files than exist. Are some files being deleted? Merged? The file inventory is incomplete.

3. **tests/e2e/.graduated/ directory**: The codebase has a `.graduated` directory inside `tests/e2e/`. The proposal doesn't mention it. What happens to it during migration?

4. **tests/e2e/tests/e2e/results/**: This nested directory contains `cli-startup-output.txt`. Still unaccounted for in the target structure. Iteration 1 flagged this; still not addressed.

5. **No mention of test retries**: Playwright has built-in retry logic (`retries` in config). Vitest has `retry` but with different semantics. If any tests rely on Playwright retries, this is an unacknowledged migration concern.

6. **Shared helpers coupling underestimated**: The proposal treats `tests/shared/helpers.ts` as a simple extraction, but it's imported by 50+ files. Any breaking change to its API (even a renamed export) would cascade across all three surfaces. The risk section mentions "regression from path/config changes" but not "regression from shared API changes."

7. **The "跨 surface 归主导 surface" rule is subjective**: Line 104 says cross-surface tests go to the "dominant surface." Who decides dominance? What if a test is 60% API and 40% Web? This classification rule needs sharper criteria.
