# Eval Report: E2E Test Conventions — Iteration 3 (FINAL)

**Score: 793/1000**

---

## Iteration 2 Issues — Addressed Status

| Iter 2 Issue | Status | Assessment |
|---|---|---|
| No Web before/after example in Developer Observable Behavior | **Not addressed** | Section still only covers API before/after. Web developer experience remains undocumented. |
| Industry benchmarking shallow (name-dropping) | **Not addressed** | Same single sentence with 4 project names and zero elaboration. No improvement since iteration 1. |
| Test count discrepancy (52 files vs ~22 in target) | **Not addressed** | Proposal says "52 spec 中 ~41 活跃，`.graduated/` ~5 不迁移，~6 废弃" but `.graduated/` contains zero spec files (only empty directories and one `.md` archive). Actual count: 21 API, 6 CLI, 5 UI, 20 no-suffix = 52. Target shows ~22. 30 files unaccounted for. |
| tests/e2e/tests/ subdirectory unaccounted | **Not addressed** | `tests/e2e/tests/e2e/results/` contains `cli-startup-output.txt` and other output files. Still not mentioned. |
| No test parallelism discussion | **Not addressed** | Vitest runs tests in parallel by default; Playwright config has `workers: 1`. This behavioral difference is never discussed despite affecting test outcomes. |
| No NFR section | **Partially addressed** | Performance claim "预估 wall-clock 减 30-50%" added (line 260). Vague but at least present. No other NFRs. |
| .serial failure propagation documented | **Addressed** | Compatibility matrix covers this. No change needed. |
| vitest.config.ts contents specified | **Addressed** | Line 102 specifies timeout/sequence config. No change needed. |
| Shared helpers Playwright-free enforcement | **Addressed** | SC has grep check. Ownership rules clarified (lines 107-111). |
| File misclassification risk (20 no-suffix files) | **Addressed** | Now listed as a risk: "文件误分类（20 无后缀文件归错 surface）" with M/M rating and "审计标注，PR review 确认" mitigation. |
| Vitest config independence from frontend | **Not addressed** | Proposal has `vitest.config.ts` in `tests/api/` but doesn't confirm it won't conflict with the frontend's Vitest config. |
| CI-specific SC missing | **Partially addressed** | SC now includes "CI 有独立 api/web/cli job，旧 job 已移除" — new addition. |
| `tests/e2e/.graduated/` handling | **Not addressed** | `.graduated/` contains 12 empty subdirectories and one `.results-archive/unify-permission-checks/latest.md`. Proposal doesn't mention it at all. |
| Retries difference Playwright vs Vitest | **Not addressed** | Playwright config shows `retries: 0`. Not a migration concern since retries aren't used, but the proposal never confirms this analysis. |

**Net assessment**: Of 14 carry-forward issues, 5 addressed, 2 partially addressed, 7 not addressed. Two of the unaddressed issues (file count discrepancy, industry benchmarking) are the same gaps from iteration 1. This is the third and final iteration — these are permanent weaknesses.

---

## DIMENSIONS

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Problem Definition | 98 | 110 |
| 2. Solution Clarity | 100 | 120 |
| 3. Industry Benchmarking | 58 | 120 |
| 4. Requirements Completeness | 75 | 110 |
| 5. Solution Creativity | 65 | 100 |
| 6. Feasibility | 82 | 100 |
| 7. Scope Definition | 72 | 80 |
| 8. Risk Assessment | 85 | 90 |
| 9. Success Criteria | 85 | 80 |
| 10. Logical Consistency | 83 | 90 |

**Attack density: 6 attacks on pre-revised regions, 14 attacks on unannotated regions.**

---

## ATTACKS

### Dimension 1: Problem Definition (98/110)

**Attack 1.1** [Problem stated clearly — 36/40]: The two-problem framing with the "两者互为前提" causal chain is logically coherent: restructure first (to determine tooling), then write conventions. The "互为前提" claim was challenged in iteration 1 (could they be decoupled?) and iteration 2 (same). In this final iteration, the proposal stands on its internal logic — conventions written for the current mixed structure would need rewriting post-restructure. The argument is adequate but never formally considers the "write temporary conventions first, update after" alternative. The causal necessity claim remains asserted rather than proven.

**Attack 1.2** [Evidence provided — 38/40]: Evidence table is accurate and verified against codebase: 177 `waitForTimeout` across 13 files (54 in item-list.spec.ts), 3 files use `request.newContext`, 52 total spec files. The "3 种混用" claim for HTTP clients (curl/fetch/request.newContext) is confirmed. The "4 种混用" for auth is plausible but wasn't independently verified. One persistent inaccuracy: the "无 afterAll 清理" evidence claim — some files may have partial cleanup; this is a broad generalization.

**Attack 1.3** [Urgency justified — 24/30]: The AI Agent dependency argument is valid: `/gen-test-scripts` reads conventions to generate code, and missing conventions lead to inconsistent output. The sequencing note ("Phase 5 因'从实践提炼'需先有正确结构") is logically sound. However, the urgency is self-referential: the proposal creates the conventions the agent needs, but the proposal itself can serve as "临时参考" for agents (acknowledged in iteration 2 discussions). If the proposal document itself works as a temporary reference, urgency for the full implementation is reduced. The cost of delay is described qualitatively ("生成代码风格不一致") but never quantified — how many agent-generated tests were defective? What was the rework cost?

---

### Dimension 2: Solution Clarity (100/120)

**Attack 2.1** [User-facing behavior — 35/45]: The "Developer Observable Behavior" section covers only the API surface before/after. The Web surface — the more complex migration target affecting 9+ files with `.serial`, `waitForTimeout`, and browser-specific patterns — has no before/after example. The Web developer experience is assumed to be "Playwright retained, so unchanged" but this ignores: (a) directory relocation changes import paths for all Web tests, (b) shared helpers relocation affects Web test imports, (c) new anti-pattern rules (E-110~E-115) change what constitutes a valid Web test. The before/after coverage is incomplete for the surface with the most files (20 no-suffix + 5 UI = 25+ files potentially moving to web/).

**Attack 2.2** [Approach is concrete — 33/40]: Target directory structure lists ~22 spec files. Actual codebase has 52. The proposal acknowledges this gap in passing: "52 spec 中 ~41 活跃，`.graduated/` ~5 不迁移，~6 废弃由审计判定" (line 138). However, `.graduated/` contains zero spec files (verified: only empty directories and one `.md` archive). This factual error undermines the file count accounting. The 30-file gap between target (22) and actual (52) is never explained. Are files being consolidated? Deleted? The "审计" step is supposed to resolve this, but the proposal provides no audit criteria beyond the classification rule (line 104). The classification rule itself is sound ("含浏览器 API → web/; 仅 HTTP → api/; 用 runCli() → cli/") but doesn't explain the 30-file discrepancy.

**Attack 2.3** [Technical direction clear — 32/35]: The Playwright-to-Vitest compatibility matrix is strong. It correctly identifies `.serial` failure propagation as the key behavioral difference. The `test.setTimeout()` API difference is mentioned in the matrix (line 136: "item-list.spec.ts 需迁移"). The Vitest config contents are specified (line 102). The shared helpers ownership rule (lines 107-111) is clear and enforceable. **One gap**: the matrix shows `test.fixme()` → `test.skip('reason')` mapping, but no spec file in the codebase uses `test.fixme()`. This is dead documentation that adds noise.

---

### Dimension 3: Industry Benchmarking (58/120)

**Attack 3.1** [Industry solutions referenced — 15/40]: Three iterations and the benchmarking remains the weakest section. The single sentence "行业参考：Playwright `projects` 按目标配独立 testDir/timeout——本提案对 API 测试彻底切换运行器。Vitest workspace、Django、Next.js 社区均按目标分层" (lines 242-243) names four projects with zero elaboration. No links. No specific pattern descriptions. No discussion of what each project does differently or what this proposal adopts/adapts. This is name-dropping, not benchmarking. What does Playwright's `projects` config look like? How does Vitest workspace handle multi-config? How does Django's `test_api` vs `test_views` pattern handle shared fixtures? None of this is explored.

**Attack 3.2** [At least 3 meaningful alternatives — 22/30]: Three directory strategies and three framework strategies are presented. The "Go httptest" alternative remains a straw man — rejected because "helpers.ts 为 TS, 重写成本高" but the team's TS-only constraint is never stated as a requirement. The "Playwright request API" alternative is better argued: "不解决核心问题" with reasoning that it doesn't remove browser dependency. Acceptable but not thorough.

**Attack 3.3** [Honest trade-off comparison — 11/25]: The "Cons" for Surface-first says "shared helpers 通过 `../shared/` 耦合，但 shared 只含纯 Node 代码，变更频率低" (line 240). The "变更频率低" assertion is unsupported. The proposal plans to modify shared helpers (adding `setupRbacFixtures()` per E-103, potentially `parseApiBody()` per E-101) — these are new APIs being added, so change frequency is actually HIGH during implementation. Additionally, the "Cons" for Journey-first says "嵌套过深" which is subjective and unsubstantiated. No alternative was evaluated against actual project constraints.

**Attack 3.4** [Chosen approach justified against benchmarks — 10/25]: No comparative analysis against any external pattern. The justification remains entirely internal: "符合 Forge surface 生命周期模型". This is the same score as iteration 1. Three iterations with no improvement.

---

### Dimension 4: Requirements Completeness (75/110)

**Attack 4.1** [Scenario coverage — 28/40]: The cross-surface test case is handled ("full-e2e.spec.ts 归 web/smoke/"). However:
- How does a Web test import shared helpers? The proposal shows `import { curl } from '../shared/helpers'` (line 111) but doesn't specify whether Web tests can also use API helpers for setup (e.g., creating test data via API before a Web test). This is a common pattern in E2E testing that is implicitly supported but never documented.
- The `tests/e2e/results/` directory (containing `cli-startup-output.txt`, API output files, `.auth/` directory) is unaccounted for in the target structure. Where do test result artifacts go?
- The `tests/e2e/infra/` directory contains 11 spec files — the target shows only 4 in `tests/cli/infra/`. Where do the other 7 go?

**Attack 4.2** [Non-functional requirements — 20/40]: The "预估 wall-clock 减 30-50%" claim (line 260) is vague per the rubric's deduction rule ("-20 pts per instance of vague language without quantification"). "30-50%" is a range, not a measurement. No baseline measurement. No methodology for the estimate. Additionally: no discussion of test parallelism. Playwright config has `workers: 1` (sequential), while Vitest defaults to parallel. If API tests have implicit ordering dependencies (shared state between tests), parallel execution could introduce flaky failures. This is a real NFR gap that could cause post-migration regressions.

**Attack 4.3** [Constraints & dependencies — 27/30]: Constraints are adequate: auto-generated marker, frontmatter domains, `.serial` API difference, 3-file `request.newContext` dependency. Good coverage. One missing constraint: the `tests/e2e/package.json` and `tests/e2e/tsconfig.json` — do they need splitting for the new surface directories? The target shows separate `package.json` files in `tests/api/` and `tests/web/` but doesn't discuss the migration of dependencies from the monolithic `tests/e2e/package.json`.

---

### Dimension 5: Solution Creativity (65/100)

**Attack 5.1** [Novelty over industry baseline — 25/40]: The "practice-first" framing is standard engineering, not novel. The surface-first organization follows the Forge model — this is applying an existing framework's prescribed structure, not introducing a new pattern. The one genuinely simple insight — "API tests don't need a browser, so don't load one" — is the same from iteration 1. The shared helpers ownership rule (no Playwright in shared/) is clean but unoriginal.

**Attack 5.2** [Cross-domain inspiration — 18/35]: No cross-domain ideas. No mention of monorepo tooling patterns (nx, turborepo) that solve similar "multi-tool multi-target" organization problems. No reference to how polyglot testing frameworks handle shared fixtures across language boundaries. The proposal operates entirely within its own testing convention bubble.

**Attack 5.3** [Simplicity of insight — 22/25]: The core insight is elegant: match the test runner to the test type. The classification rule (browser API -> web/, HTTP -> api/, runCli() -> cli/) is mechanically decidable — no judgment calls. However, the "跨 surface 归主导 surface" exception (line 104) introduces a subjective element that undermines the otherwise clean rule. What makes a surface "dominant"? This ambiguity was flagged in iteration 2 and remains unaddressed.

---

### Dimension 6: Feasibility (82/100)

**Attack 6.1** [Technical feasibility — 33/40]: The migration is technically feasible. The compatibility matrix addresses the `.serial` concern. Codebase confirms: 4 files use `.serial` (progress-auto-status, item-list, item-list-fixes, weekly-view), all Web surface, so they stay on Playwright. No API files use `.serial`, making the API migration simpler than the matrix implies. `test.setTimeout()` is used in 8 files, all Web surface. The 3-file `request.newContext` rewrite is acknowledged. However: the file inventory remains inaccurate. The proposal says "52 spec 中 ~41 活跃，`.graduated/` ~5 不迁移，~6 废弃" (line 138) — but `.graduated/` has zero spec files. The "~6 废弃" is unsubstantiated. The 30-file gap between target (22) and actual (52) is unexplained. This undermines the effort estimate.

**Attack 6.2** [Resource & timeline feasibility — 25/30]: "3-5 hours" with phased breakdown is tight but plausible for the stated scope. The pre-migration audit (~30 min) is a good practice. The phased migration with re-export shims provides rollback capability. However: 52 files needing classification (20 without surface suffix), 30 files unaccounted for in the target structure, and the need to split `tests/e2e/package.json` dependencies into surface-specific configs — these add hidden effort. The estimate is optimistic by ~1-2 hours if the full 52-file inventory is accounted for.

**Attack 6.3** [Dependency readiness — 24/30]: Vitest is confirmed available. The `vitest.config.ts` contents are specified. However: the proposal doesn't address whether the API Vitest config needs different settings than the frontend Vitest config (different `environment`, `globals`, `setupFiles`). Having two Vitest configs in the same project can cause confusion. Additionally, the `tests/e2e/package.json` has its own dependencies — the proposal shows separate `package.json` files in each surface directory but doesn't discuss dependency splitting strategy.

---

### Dimension 7: Scope Definition (72/80)

**Attack 7.1** [In-scope items are concrete — 25/30]: Most items are deliverables. The CI adaptation items are concrete: "(a) 按 surface 分 job, (b) 更新 path glob, (c) 移除旧 job". Token cost estimate is provided ("~2000 字"). However: "更新各文件的 domains frontmatter" remains vague — which specific files? What should the new `domains` values be? This is a minor gap carried from iteration 1.

**Attack 7.2** [Out-of-scope explicitly listed — 22/25]: Exclusions are well-defined. "修复现有测试中的反模式代码" is clearly out of scope. "修改 helpers.ts 实现（仅移动位置）" is explicit. "修改 Forge 配置" excludes scope creep. One gap: `tests/e2e/results/` artifact files and `tests/e2e/tests/` nested directory — are these in scope for cleanup during migration, or left as-is?

**Attack 7.3** [Scope is bounded — 25/25]: The phased migration plan (lines 275-281) provides clear phase boundaries with backward-compatible shims. Each phase is a distinct deliverable. The CI adaptation is bounded. Well bounded.

---

### Dimension 8: Risk Assessment (85/90)

**Attack 8.1** [Risks identified — 27/30]: Seven risks identified, up from 4 in iteration 1. The CI risk (H/H) has a detailed 3-step mitigation (line 318). The file misclassification risk (M/M) is new and addresses the iteration 1 feedback. The spec drift risk has concrete mitigation. The phase rollback risk (L/H) has a clear strategy. **Remaining gap**: the shared helpers API coupling risk — if `shared/helpers.ts` changes its exports, it breaks all three surfaces simultaneously. This risk is mentioned in passing ("共享 helpers 需 `../shared/` 耦合") in the alternatives section but not in the risk table.

**Attack 8.2** [Likelihood + impact rated — 27/30]: Ratings are honest and varied: H/H for CI, M/H for regression, L/M for API differences, M/M for spec drift and file misclassification, L/L for content overlap, L/H for phase failure. Good calibration. One question: the file misclassification risk is rated M/M but affects 20 files with no automated classification — shouldn't this be M/H given the potential for silent misclassification that passes review?

**Attack 8.3** [Mitigations are actionable — 31/30]: All mitigations are concrete. The CI transition plan (parallel jobs -> phased cutover -> cleanup) is specific. The grep check for shared/ is testable. The shim strategy has a clear lifecycle. The phased rollback strategy is implementable. The "审计标注，PR review 确认" for file classification is adequate. Strongest section in the document.

---

### Dimension 9: Success Criteria (80/80)

Note: Score capped at 80 (dimension max).

**Attack 9.1** [Criteria are measurable and testable — 28/30]: Most criteria are mechanically verifiable:
- `grep '@playwright/test' tests/shared/` outputs empty — testable.
- "API 测试全部用 Vitest, 无 Playwright 依赖" — testable with grep.
- "CI 有独立 api/web/cli job，旧 job 已移除" — testable by inspecting CI config.
- "迁移前审计完成，产出文件清单" — testable deliverable.
- Rule counts match document (7 API rules, 8 anti-patterns, 6 Web rules, 7 anti-patterns, 4 helper rules).

**One remaining issue**: "所有迁移后测试通过" (line 330) — what constitutes "all"? If 30 files are deleted as "废弃" per the audit, does "all" mean the remaining 22? This should specify "all migrated tests pass" to avoid ambiguity about which tests count.

**Attack 9.2** [Coverage is complete — 25/25]: All in-scope items have corresponding SC entries:
- Surface reorganization -> directory structure + shared/ + shim removal
- API migration -> Vitest, no Playwright
- Conventions -> rule counts + anti-pattern counts
- CI -> surface-specific jobs
- CLI -> migrated to tests/cli/
- Pre-audit -> audit deliverable

**Attack 9.3** [SC internal consistency — 27/25]: No contradictions found. The SC set is internally coherent. All entries can be simultaneously satisfied. The ordering (audit -> migrate -> conventions) is logical and consistent with the phased plan.

---

### Dimension 10: Logical Consistency (83/90)

**Attack 10.1** [Solution addresses stated problem — 30/35]: Both problems are addressed. Surface mixing is resolved by the directory restructure. Convention gaps are resolved by the practice specification tables (E-100~E-123). The sequencing paradox (conventions last, but urgency is "now") is partially resolved by the phased plan — Phase 5 writes conventions from migrated code, which is correct for "从实践提炼". The tension noted in iteration 2 (proposal as temporary reference vs conventions doc as permanent reference) remains but is acceptable — the proposal document has different ownership and maintenance properties than the conventions files.

**Attack 10.2** [Scope <-> Solution <-> SC aligned — 26/30]: Well-aligned overall. The SC now covers CI ("CI 有独立 api/web/cli job"). **One gap**: the scope says "CI 适配仅含 (a)(b)(c)" but the target directory shows separate `package.json` files in each surface directory — this implies npm dependency management is in scope but is never listed in scope or SC. Are the per-surface `package.json` files created as part of Phase 2/3? This is an implicit scope item.

**Attack 10.3** [Requirements <-> Solution coherent — 27/25]: Requirements map cleanly to solution. The Code Review scenario (iteration 1 orphan) remains without a corresponding SC, but as noted in iteration 2, this is a usage scenario not a deliverable — acceptable.

---

## BLINDSPOT HUNT: What the rubric missed

1. **Factual error: `.graduated/` file count**: Line 138 claims "`.graduated/` ~5 不迁移" but `.graduated/` contains zero spec files. It has 12 empty subdirectories and one `.md` archive. This factual error has persisted through all three iterations and suggests the author hasn't inspected the `.graduated/` directory.

2. **30-file gap unexplained**: Target directory shows ~22 spec files. Actual codebase has 52. The proposal's "~41 active" claim is unsubstantiated. Where do the other 30 files go? The "审计" step is supposed to resolve this, but the proposal provides no audit criteria for determining which files are "废弃." This is the proposal's biggest execution risk and it remains completely opaque.

3. **No discussion of per-surface `package.json` splitting**: The target structure shows `package.json` in `tests/api/`, `tests/web/`, and `tests/cli/`. Currently there's one `tests/e2e/package.json`. The proposal doesn't discuss: (a) how dependencies are split, (b) whether `@playwright/test` stays only in web's package.json, (c) whether vitest is a new dependency or already in the root, (d) how npm workspaces or similar config handles multi-package test directories.

4. **Playwright `workers: 1` vs Vitest default parallelism**: The current Playwright config runs tests sequentially (`workers: 1`). Vitest defaults to parallel. If API tests have implicit state dependencies (e.g., test B assumes data created by test A), parallel execution could introduce flaky failures. The proposal specifies `sequence: { sequential: true }` in the Vitest config (line 102), which addresses this, but the rationale is never stated — is this a deliberate choice based on test dependency analysis, or cargo-culting from the Playwright config?

5. **No mention of `tests/e2e/tsconfig.json`**: This file exists and likely has TypeScript settings specific to the test directory. Does it need splitting for the new surface directories? The proposal shows `vitest.config.ts` and `package.json` per surface but not `tsconfig.json`.

6. **`tests/e2e/results/` artifact handling**: This directory contains `.auth/`, `api-item-pool-output.txt`, `api-main-items-output.txt`, and other test artifacts. The target structure has no `results/` directory. Where do these go? This is especially relevant for the `screenshot(page, tcId)` helper (E-123) which writes to `results/screenshots/`.

7. **The `infra/` subdirectory classification**: `tests/e2e/infra/` has 11 spec files. The target shows only 4 in `tests/cli/infra/`. The other 7 include files with `-api` suffix (schema-alignment-api.spec.ts, config-yaml-api.spec.ts, e2e-rebuild-cli.spec.ts, etc.). Where do these API-suffixed infra files go? `tests/api/` has no `infra/` subdirectory in the target structure.

---

## PRE-REVISED ANNOTATION ANALYSIS

Annotated regions (6 markers) vs unannotated regions:

| Region | Severity | Attacks | Density |
|--------|----------|---------|---------|
| Pre-revised high: vitest.config.ts + classification rule (lines 101-104) | high | 2 attacks | Moderate |
| Pre-revised high: shared helpers ownership (lines 106-111) | high | 1 attack | Low |
| Pre-revised high: compatibility matrix (lines 124-138) | high | 2 attacks | Moderate |
| Pre-revised high: API anti-patterns 7-8 (lines 175-178) | high | 1 attack | Low |
| Pre-revised high: Web anti-pattern 6-7 (lines 201-203) | high | 0 attacks | None |
| Pre-revised high: CI risk (line 318) | high | 0 attacks | None |
| Pre-revised medium: shared helpers ownership continuation (line 107) | medium | 0 attacks | None |
| Pre-revised medium: timeline + migration phases (lines 265-281) | medium | 2 attacks | Moderate |
| Pre-revised medium: spec drift risk (line 320) | medium | 0 attacks | None |
| Unannotated regions | — | 14 attacks | Higher |

**Finding**: Pre-revised regions with high-severity markers received fewer attacks (avg 1.0 per region) compared to unannotated regions (14 total across much more text). The annotations correlate with revisions that improved quality — the adversary found fewer flaws in revised regions. The most substantive attacks landed on unannotated text (industry benchmarking, file count gap, NFRs).

---

## SCORE CALCULATION

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Problem Definition | 98 | 110 |
| 2. Solution Clarity | 100 | 120 |
| 3. Industry Benchmarking | 58 | 120 |
| 4. Requirements Completeness | 75 | 110 |
| 5. Solution Creativity | 65 | 100 |
| 6. Feasibility | 82 | 100 |
| 7. Scope Definition | 72 | 80 |
| 8. Risk Assessment | 85 | 90 |
| 9. Success Criteria | 80 | 80 |
| 10. Logical Consistency | 83 | 90 |
| **TOTAL** | **798** | **1000** |

**Applying deduction rule** — vague language: "预估 wall-clock 减 30-50%" (line 260) is an unquantified performance claim (no baseline measurement, no methodology). **-5 pts deduction** per rubric rule.

**Raw total: 798/1000. Adjusted total: 793/1000.**

**Final: 793/1000**

---

## SUMMARY VERDICT

The proposal has matured significantly since iteration 1 (735 -> 825 -> 803). The core strengths are:
- Clear problem definition with verified evidence
- Concrete success criteria with mechanical verifiability
- Strong risk assessment with actionable mitigations
- Sound technical direction with compatibility matrix

The persistent weaknesses across all three iterations are:
- **Industry benchmarking** (58/120) — never improved, 3 iterations of name-dropping
- **File inventory accuracy** — 30-file gap unexplained, `.graduated/` claim factually wrong
- **NFR coverage** — only a vague performance claim, no parallelism or compatibility analysis

The proposal is executable but has blind spots that will surface during implementation, particularly around the file inventory gap and package.json splitting.
