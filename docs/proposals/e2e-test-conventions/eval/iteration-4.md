# Eval Report: E2E Test Conventions — Iteration 4

**Score: 764/1000**

---

## Iteration 3 Issues — Addressed Status

| Iter 3 Issue | Status | Assessment |
|---|---|---|
| No Web before/after example in Developer Observable Behavior | **Not addressed** | Section still reads "Web Before/After: waitForTimeout(2000) + CSS -> toBeVisible() + getByTestId() + login(page) + .serial" — a single-line shorthand, not a concrete before/after example comparable to the API one. |
| Industry benchmarking shallow (name-dropping) | **Not addressed** | Identical single-sentence name-drop of Playwright projects/Vitest workspace/Django/Next.js with zero elaboration since iteration 1. |
| Test count discrepancy (52 files vs ~22 in target) | **Not addressed** | Proposal line 144 still says "52 spec 文件中 ~41 活跃（.graduated/ 含 12 个已完成 feature 的归档标记目录，不含 spec 文件），全部迁移到新结构" — the parenthetical clarifies .graduated/ has no specs, but the 52->22 gap in the target directory listing remains unexplained. The target shows ~22 files; 30 files are listed as "全部迁移" but have no target location. |
| tests/e2e/tests/ subdirectory unaccounted | **Not addressed** | `tests/e2e/tests/e2e/results/cli-startup-output.txt` exists. No mention in proposal. |
| No test parallelism discussion | **Not addressed** | Proposal specifies `sequence: { sequential: true }` in vitest.config.ts (line 101) but never explains the rationale. Is this a deliberate choice based on dependency analysis, or cargo-culting from `workers: 1`? |
| NFR section vague | **Not addressed** | Performance claim on line 271: "预估 ~20s（消除浏览器加载 + 纯 Node 启动）。正式数据待迁移后 benchmark" — the "30-50%" claim was removed but replaced with an equally vague "~20s" estimate with no methodology. |
| .serial failure propagation documented | **Addressed** | Compatibility matrix covers this. Stable. |
| vitest.config.ts contents specified | **Addressed** | Line 101-103 specifies config. Stable. |
| Shared helpers Playwright-free enforcement | **Not addressed — new wrinkle** | SC says `grep '@playwright/test' tests/shared/` should output empty. But `helpers.ts` has `import type { Page, Locator } from '@playwright/test'`. Functions using `Page`/`Locator` (login, screenshot, navTo, createTestMainItem) must be extracted to `tests/web/helpers/`. The proposal mentions this at line 116 ("Playwright-specific helpers 放在 tests/web/helpers/") but never discusses the extraction mechanics: which functions exactly, what the resulting `shared/helpers.ts` surface looks like, and whether the type import removal breaks anything. |
| File misclassification risk (20 no-suffix files) | **Addressed** | Listed as risk with M/M. Stable. |
| Vitest config independence from frontend | **Not addressed** | `frontend/package.json` has vitest ^1.6.0. `tests/e2e/package.json` has no vitest dependency. Proposal shows `vitest.config.ts` in `tests/api/` but never states whether vitest is added as a new dependency to `tests/api/package.json` or relies on the frontend's installation. |
| CI-specific SC | **Addressed** | SC line 349 covers CI jobs. Stable. |
| tests/e2e/.graduated/ handling | **Not addressed** | `.graduated/` has 13 empty subdirectories and one `.md` archive. Proposal never mentions whether these are deleted, archived, or left in place during migration. |
| Retries difference | **Not addressed** | Playwright config has `retries: 0`. Not a functional concern but never confirmed as analyzed. |
| Per-surface package.json splitting | **Not addressed** | Current monolithic `tests/e2e/package.json` has 3 deps: @playwright/test, typescript, yaml. Proposal shows separate package.json in each surface dir but never discusses how dependencies are split. |

**Net assessment**: Of 14 carry-forward issues from iteration 3, 4 addressed, 10 not addressed. The proposal document has not been materially revised since its initial commit. Issues are confirmed persistent.

---

## DIMENSIONS

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Problem Definition | 95 | 110 |
| 2. Solution Clarity | 96 | 120 |
| 3. Industry Benchmarking | 55 | 120 |
| 4. Requirements Completeness | 72 | 110 |
| 5. Solution Creativity | 62 | 100 |
| 6. Feasibility | 78 | 100 |
| 7. Scope Definition | 70 | 80 |
| 8. Risk Assessment | 83 | 90 |
| 9. Success Criteria | 78 | 80 |
| 10. Logical Consistency | 80 | 90 |

---

## ATTACKS

### Dimension 1: Problem Definition (95/110)

**Attack 1.1** [Problem stated clearly — 35/40]: The two-problem framing is coherent. The "互为前提" causal chain (restructure first -> determine tooling -> write conventions) is logically valid. However, the causal necessity claim was challenged in iterations 1-3 and the response remains the same assertion without proof. An alternative exists: write temporary conventions for the current structure (usable by agents immediately), then update post-restructure. The proposal dismisses this implicitly via the sequencing argument but never evaluates it. The claim that conventions for the current mixed structure would "need rewriting" is plausible but unproven — if conventions are written as surface-agnostic (e.g., "use curl() for HTTP"), they'd survive the directory move.

**Attack 1.2** [Evidence provided — 38/40]: Evidence table is substantially accurate. Verified against codebase:
- 52 spec files total — confirmed.
- 166 `waitForTimeout` across 13 files — confirmed (proposal says 177 in 10 files; actual count is 166 across 13 files). This is a minor inaccuracy: the proposal overstates occurrences by 11 and understates file count by 3.
- 3 files use `request.newContext` — confirmed.
- 4 files use `.serial` — confirmed.
- "无 afterAll 清理" — only 6 of 52 files have `afterAll`/`afterEach` hooks. The claim is directionally correct.
- "3 种混用 HTTP 客户端" — confirmed (curl, fetch, request.newContext).
- "4 种混用认证方式" — plausible but not independently verified.

One persistent inaccuracy: waitForTimeout count (177 vs 166) and file count (10 vs 13) have been wrong since iteration 1 and never corrected despite being flagged.

**Attack 1.3** [Urgency justified — 22/30]: The AI Agent dependency argument (`/gen-test-scripts` reads conventions) is valid. The cost of delay is described as "生成代码风格不一致" but never quantified. How many agent-generated tests were defective? How much rework resulted? Without this data, the urgency is asserted rather than demonstrated. The sequencing argument ("Phase 5 因从实践提炼需先有正确结构") is reasonable but creates a paradox: if urgency is high (agents generating bad code now), waiting until Phase 5 for conventions is the slowest path. A lightweight interim convention (1 page, surface-agnostic) would address urgency faster.

---

### Dimension 2: Solution Clarity (96/120)

**Attack 2.1** [User-facing behavior — 32/45]: The "Developer Observable Behavior" section has a one-line shorthand for Web ("waitForTimeout(2000) + CSS -> toBeVisible() + getByTestId() + login(page) + .serial") that is not a concrete before/after example. Compare with the API section which shows the full observable change. The Web surface has 25+ files being migrated (20 no-suffix + 5 UI), and the developer experience is reduced to a shorthand. Import path changes (`../helpers` -> `../shared/helpers` or `../web/helpers`), the extraction of Playwright-typed functions from helpers.ts to web/helpers/, and the new anti-pattern rules (E-110~E-115) are all Web developer experience changes that are not illustrated.

**Attack 2.2** [Approach is concrete — 33/40]: Target directory lists ~22 spec files. Codebase has 52. Line 144 says "全部迁移到新结构" but the target only shows ~22 positions. The 30-file gap has been flagged in every iteration and remains unexplained. Additionally, the `infra/` subdirectory has 11 spec files but the target shows:
- `tests/api/smoke/`: 4 files (deploy-smoke, jlc-schema-api, schema-alignment-api, config-yaml-api)
- `tests/web/smoke/`: 1 file (schema-alignment-ui)
- `tests/cli/infra/`: 5 files (config-yaml-cli, bizkey-cli, jlc-schema-cli, lint-keywords-cli, e2e-rebuild-cli)
- That accounts for 10 of 11 infra files. `unify-permission-checks-build.spec.ts` is listed under `tests/cli/infra/` in the mapping (line 111) — that's 11. But the target tree (lines 91-98) only shows 4 files in cli/infra/. The mapping (lines 109-112) lists 6 files for cli/infra/. These are inconsistent.

**Attack 2.3** [Technical direction clear — 31/35]: The Playwright-to-Vitest compatibility matrix is strong. The vitest.config.ts contents are specified. The shared helpers ownership rule is clear. However: the proposal references `test.sequential` (line 232: "用 test.sequential") but the Vitest API uses `describe.sequential()` not `test.sequential`. The compatibility matrix (line 135) correctly says `describe.sequential()` — this inconsistency within the proposal suggests the author conflates `test.serial` (Playwright) with `describe.sequential` (Vitest). Additionally, `test.setTimeout()` is used in 7 files (not just `item-list.spec.ts` as line 142 implies): full-e2e, progress-auto-status, item-pool, item-list, status-flow-dynamic, weekly-view, item-list-fixes. All are Web surface, so they stay on Playwright — but the claim that only one file needs migration is wrong.

---

### Dimension 3: Industry Benchmarking (55/120)

**Attack 3.1** [Industry solutions referenced — 12/40]: Four iterations. The benchmarking section remains a single-sentence name-drop. The expanded alternatives table (lines 243-255) provides marginally more detail than before but still lacks:
- No links to documentation or source code for any referenced pattern
- No description of what Playwright's `projects` config actually does (testDir, timeout, retries per target)
- No description of Vitest workspace's actual behavior
- No code example from any external project
- No discussion of what this proposal adopts vs. diverges from any benchmark

This is the weakest section of the document and has not improved since iteration 1.

**Attack 3.2** [At least 3 meaningful alternatives — 22/30]: Three directory strategies and three framework strategies are presented. "Go httptest" remains a straw man — the team's TS-only constraint is stated in Constraints but is a project constraint, not a design constraint for test tooling. The constraint should be challenged: if API tests validate HTTP contracts, language is irrelevant. "Playwright request API" is better argued. The alternatives are genuine but not deeply explored.

**Attack 3.3** [Honest trade-off comparison — 11/25]: The Surface-first "Cons" column (line 245) says "shared helpers 通过 ../shared/ 耦合——shared/helpers.ts 变更可能级联影响多 surface，但其纯 Node 无状态特性使变更频率低". The "变更频率低" assertion is contradicted by the proposal itself: it plans to add `setupRbacFixtures()` (E-103) and potentially modify `parseApiBody()` (E-101) — these are new/modifed shared APIs. During implementation, shared helpers will have HIGH change frequency. The Cons for Journey-first ("嵌套过深") is subjective. No alternative was evaluated with concrete measurements (directory depth, import path length, CI job granularity).

**Attack 3.4** [Chosen approach justified against benchmarks — 10/25]: No comparative analysis against any external pattern. Justification is entirely internal ("符合 Forge surface 生命周期模型"). Four iterations, no improvement.

---

### Dimension 4: Requirements Completeness (72/110)

**Attack 4.1** [Scenario coverage — 26/40]: Gaps identified in iterations 1-3 remain:
- How do Web tests use shared API helpers for setup? The proposal says `import { curl } from '../shared/helpers'` (line 118) but doesn't specify that Web tests can also use API helpers for data setup — a standard E2E pattern (create data via API, verify via UI).
- `tests/e2e/results/` (5 output files) and `tests/e2e/tests/e2e/results/` (1 file) are unaccounted for in the target structure. The `screenshot()` helper (E-123) writes to `results/screenshots/`. Where does `results/` live in the new structure?
- `tests/e2e/test-results/` (Playwright's output directory) is also unmentioned.
- The `.graduated/` directory (13 empty subdirectories + 1 archive file) is not addressed in the target structure.
- The `helpers.ts` file currently has Playwright type imports (`Page`, `Locator`) — the proposal mentions extracting browser-specific helpers to `tests/web/helpers/` but doesn't specify which functions: `login()`, `screenshot()`, `navTo()`, `createTestMainItem()`, `fillFormItem()` all take `Page` parameters.

**Attack 4.2** [Non-functional requirements — 19/40]: The performance estimate on line 271 says "预估 ~20s（消除浏览器加载 + 纯 Node 启动）" — this is still an unquantified estimate. No baseline measurement of current API test runtime. No methodology for the estimate. No measurement plan. The rubric deduction rule applies: "-20 pts per instance of vague language without quantification." Additionally: no discussion of Vitest parallelism (default parallel vs `sequence: { sequential: true }`). No discussion of memory usage, test isolation, or test execution order determinism. No security NFR (shared helpers contain admin credentials in token cache).

**Attack 4.3** [Constraints & dependencies — 27/30]: Constraints are adequate: auto-generated marker, frontmatter domains, `.serial` API difference, `request.newContext` dependency, TS ecosystem. One missing constraint: `helpers.ts` imports `type { Page, Locator } from '@playwright/test'` — this means `shared/helpers.ts` cannot be truly Playwright-free without extracting the functions that use these types. The proposal acknowledges this at line 116 but doesn't list it as a constraint (a code dependency that must be resolved).

---

### Dimension 5: Solution Creativity (62/100)

**Attack 5.1** [Novelty over industry baseline — 24/40]: The "practice-first" approach is standard engineering (measure-then-standardize), not novel. The surface-first organization applies an existing framework's (Forge) prescribed structure. The one genuine insight — "API tests don't need a browser, so don't load one" — is obvious to any experienced engineer. The shared helpers ownership rule (no Playwright in shared/) is a sound constraint but not creative.

**Attack 5.2** [Cross-domain inspiration — 16/35]: No cross-domain ideas. No reference to monorepo tooling patterns (nx, turborepo) that solve "multi-tool multi-target" organization. No reference to polyglot testing patterns for shared fixtures across language boundaries. No reference to microservice test taxonomy (unit/integration/contract/e2e) and how surface-based organization maps to it. The proposal operates entirely within its own convention system.

**Attack 5.3** [Simplicity of insight — 22/25]: The classification rule (browser API -> web/, HTTP -> api/, runCli() -> cli/) is mechanically decidable and elegant. The exception "跨 surface 归主导 surface" (line 105) introduces subjectivity ("dominant" is undefined) — this was flagged in iteration 2 and remains unaddressed. What makes `full-e2e.spec.ts` Web-dominant rather than having its own `tests/cross-surface/` category? The exception undermines the otherwise clean rule.

---

### Dimension 6: Feasibility (78/100)

**Attack 6.1** [Technical feasibility — 30/40]: The migration is technically feasible but has more hidden complexity than acknowledged:
- **helpers.ts type extraction**: Currently `helpers.ts` has `import type { Page, Locator } from '@playwright/test'` and exports functions using these types (login, screenshot, navTo, createTestMainItem). These must be extracted to `tests/web/helpers/`. The proposal mentions this (line 116) but doesn't count the effort for this extraction or discuss the impact on the ~40 spec files that import these functions.
- **File inventory**: The target directory lists ~22 files. Codebase has 52. The "全部迁移" claim on line 144 is correct (all files migrate) but the target tree only shows ~22, leaving 30 files unplaced. This makes the target tree misleading as a implementation guide.
- **`test.setTimeout()` scope**: Line 142 says "item-list.spec.ts 需迁移" but 7 files use `test.setTimeout()`. All happen to be Web surface (staying on Playwright), so no migration needed — but the claim is factually incomplete.

**Attack 6.2** [Resource & timeline feasibility — 24/30]: "3-5 hours" is optimistic given:
- 52 files to classify and migrate (not ~22 as the target tree implies)
- helpers.ts extraction of Playwright-typed functions (affecting ~40 import sites)
- Per-surface package.json creation with dependency splitting
- tsconfig.json handling (current `tests/e2e/tsconfig.json` has `"include": ["**/*.ts"]` — does each surface need its own?)
- CI job creation for 3 surfaces + cleanup of old job
- The pre-migration audit (~30 min) must classify 20 no-suffix files

Realistic estimate: 5-8 hours. The proposal's estimate is achievable only if the 30 unplaced files are trivially classified (which the "审计" step assumes).

**Attack 6.3** [Dependency readiness — 24/30]: Vitest is available in the frontend workspace (`vitest ^1.6.0`). But `tests/e2e/package.json` has no vitest dependency. The proposal needs to add vitest to `tests/api/package.json` and `tests/cli/package.json`. This is not discussed. Additionally: the current `tests/e2e/package.json` has only 3 dependencies (`@playwright/test`, `typescript`, `yaml`). After splitting, `tests/web/package.json` needs `@playwright/test`, `tests/api/package.json` needs `vitest` (new dep), `tests/shared/` needs `yaml`. TypeScript is needed by all. The dependency splitting strategy is unaddressed.

---

### Dimension 7: Scope Definition (70/80)

**Attack 7.1** [In-scope items are concrete — 23/30]: Most items are deliverables. However:
- "更新各文件的 domains frontmatter" — which files? What values? Vague.
- "各 surface 的 package.json 创建和依赖拆分（shared deps 通过 workspace 引用）" — "workspace 引用" implies npm workspaces config at root level, which is not listed as an in-scope item (root package.json modification). Is this in scope or out?
- Token cost estimate (~2000 字) is a good concrete metric.

**Attack 7.2** [Out-of-scope explicitly listed — 22/25]: Exclusions are clear. One gap: `tests/e2e/test-results/` (Playwright's output) and `tests/e2e/results/` (custom output) — are these directories cleaned up during migration or left as-is? The `tests/e2e/tests/` nested directory is also unaccounted for. Are orphaned directories in scope for cleanup?

**Attack 7.3** [Scope is bounded — 25/25]: The phased plan (Phase 1-5) provides clear boundaries. Each phase is a distinct deliverable with backward-compatible shims. CI adaptation is bounded. Well bounded.

---

### Dimension 8: Risk Assessment (83/90)

**Attack 8.1** [Risks identified — 27/30]: Eight risks identified. The CI risk (H/H) has detailed mitigation. The file misclassification risk (M/M) is included. However:
- **New risk missing**: helpers.ts extraction — the proposal requires splitting `helpers.ts` into `shared/helpers.ts` (pure Node) and `web/helpers/` (Playwright-typed). Any extraction error breaks imports across ~40 files. This is a non-trivial refactoring risk that should be in the risk table.
- **Shared helpers API coupling**: mentioned in Alternatives (line 245) but not in the risk table. If `shared/helpers.ts` changes its exports, it breaks all three surfaces simultaneously.

**Attack 8.2** [Likelihood + impact rated — 26/30]: Ratings are generally honest. One question: "文件误分类" is rated M/M but affects 20 files (38% of total) with no automated classification — a misclassification that passes review silently could cause test failures in CI. M/H might be more appropriate. The CI risk is correctly rated H/H.

**Attack 8.3** [Mitigations are actionable — 30/30]: All mitigations are concrete and implementable. The CI transition plan (parallel jobs -> phased cutover -> cleanup) is specific. The grep check for shared/ is testable. The shim strategy has a clear lifecycle. The phased rollback per-branch is implementable. Strongest section.

---

### Dimension 9: Success Criteria (78/80)

**Attack 9.1** [Criteria are measurable and testable — 27/30]: Most criteria are mechanically verifiable. However:
- "所有迁移后测试通过" (line 336) — "all" is ambiguous. Does this mean all 52 files pass? Or all files that survive the audit? If the audit deletes 30 files as "废弃", does "all" mean the remaining 22? Should specify "all migrated tests pass" with a definition of "migrated" (files moved to new directory structure).
- "API 测试全部用 Vitest，无 Playwright 依赖" (line 338) — this is testable with `grep -r '@playwright/test' tests/api/` but only if the grep includes `.ts` files. The SC doesn't specify the grep command as it does for shared/ (line 345).

**Attack 9.2** [Coverage is complete — 24/25]: Most in-scope items have SC entries. One gap: the in-scope item "各 surface 的 package.json 创建和依赖拆分" has no corresponding SC. There's no criterion like "each surface directory has a valid package.json with correct dependencies" or "npm install succeeds in each surface directory". This is a deliverable without a verification criterion.

**Attack 9.3** [SC internal consistency — 27/25]: No contradictions found. The SC set is internally coherent. All entries can be simultaneously satisfied. The ordering (audit -> migrate -> conventions) is logical.

---

### Dimension 10: Logical Consistency (80/90)

**Attack 10.1** [Solution addresses stated problem — 29/35]: Both problems are addressed. Surface mixing -> directory restructure. Convention gaps -> practice specification tables. The sequencing tension (conventions last, urgency now) remains but is acceptable for the "practice-first" approach. One gap: the proposal claims urgency because "/gen-test-scripts depends on conventions" but conventions are Phase 5 (last). During Phases 1-4, the agent has no conventions to read. If the proposal itself serves as a temporary reference (as discussed in iteration 2), this is fine — but the proposal never acknowledges this interim usage.

**Attack 10.2** [Scope <-> Solution <-> SC aligned — 25/30]: Well-aligned overall. Two gaps:
- In-scope item "各 surface 的 package.json 创建和依赖拆分（shared deps 通过 workspace 引用）" has no corresponding SC. The scope says "workspace 引用" but no SC verifies workspace configuration.
- In-scope item "更新各文件的 domains frontmatter" has no SC specifying what the new domains values should be.

**Attack 10.3** [Requirements <-> Solution coherent — 26/25]: Requirements map cleanly to solution. The Code Review scenario has no corresponding deliverable but is a usage scenario, not a deliverable — acceptable. The CI scenario now has SC coverage. No orphan requirements or solution features without requirements.

---

## BLINDSPOT HUNT: What the rubric missed

1. **helpers.ts Playwright type extraction is underspecified**: `helpers.ts` has `import type { Page, Locator } from '@playwright/test'`. Functions using these types: `login()`, `screenshot()`, `navTo()`, `createTestMainItem()`, `fillFormItem()`, and potentially others. The proposal says "Playwright-specific helpers 放在 tests/web/helpers/" (line 116) but doesn't enumerate which functions move, what the remaining `shared/helpers.ts` surface looks like, or how the ~40 spec files that import these functions will update their imports. This is a significant implementation detail hidden behind a single sentence.

2. **waitForTimeout evidence is inaccurate**: The proposal claims "177 处 waitForTimeout()（10 文件，item-list.spec.ts 54 处）" (line 27). Actual count: 166 occurrences across 13 files. This factual error has persisted through 4 iterations and 4 evaluation rounds, suggesting the author never re-verified the claim.

3. **No npm workspace configuration discussed**: The in-scope item says "shared deps 通过 workspace 引用" but the root `package.json` is never mentioned. Setting up npm workspaces requires modifying the root package.json to add `workspaces` entries. This is an implicit scope expansion that could affect other developers' workflows.

4. **tests/e2e/tsconfig.json splitting**: The current tsconfig has `"include": ["**/*.ts"]`. If each surface gets its own directory with its own package.json, does each need its own tsconfig? Or does the existing tsconfig with its glob pattern work for the new structure? This is never discussed.

5. **`tests/e2e/node_modules/` handling**: The current `tests/e2e/` has its own `node_modules/`. After splitting into 3 surfaces + shared, does each surface run its own `npm install`? How are shared dependencies (typescript, yaml) handled across surfaces?

6. **Vitest version compatibility**: Frontend uses `vitest ^1.6.0`. If API tests add vitest as a dependency, should they use the same version? If vitest is installed at root vs per-surface, different versions could cause behavioral differences. Not discussed.

7. **Test result artifact paths**: The `screenshot()` helper uses a hardcoded relative path `join(__dirname, '..', 'results', 'screenshots')`. After moving helpers, `__dirname` changes, and the `results/` directory location is undefined in the target structure. This is a concrete breaking change not discussed.

8. **The "无后缀文件分类" rule has edge cases**: Line 105 says "含浏览器 API -> web/; 仅 HTTP -> api/; 用 runCli() -> cli/; 跨 surface 归主导 surface". But some infra files (e.g., `deploy-smoke.spec.ts`) test both API availability and CLI commands. The "主导 surface" exception is applied without defining "dominant." Additionally, files that import `Page` type but only use it for type annotations (not browser automation) could be misclassified.

---

## PRE-REVISED ANNOTATION ANALYSIS

Not applicable — the proposal was not annotated with revision markers between iteration 3 and iteration 4. The proposal document is unchanged since its initial commit (commit 692125d).

---

## SCORE CALCULATION

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Problem Definition | 95 | 110 |
| 2. Solution Clarity | 96 | 120 |
| 3. Industry Benchmarking | 55 | 120 |
| 4. Requirements Completeness | 72 | 110 |
| 5. Solution Creativity | 62 | 100 |
| 6. Feasibility | 78 | 100 |
| 7. Scope Definition | 70 | 80 |
| 8. Risk Assessment | 83 | 90 |
| 9. Success Criteria | 78 | 80 |
| 10. Logical Consistency | 80 | 90 |
| **TOTAL** | **769** | **1000** |

**Applying deduction rules**:
- Vague language: "预估 ~20s" on line 271 (performance claim without baseline measurement or methodology) — **-5 pts** (partial deduction; already penalized in D4 NFR scoring to avoid double-counting)
- Factual errors ("177 处 waitForTimeout, 10 文件" vs actual 166/13; "item-list.spec.ts 需迁移" for setTimeout when 7 files use it) already reflected in D1 and D2 scores.

**Raw total: 769/1000. Adjusted total: 764/1000.**

---

## SUMMARY VERDICT

The proposal remains at its initial commit state across 4 evaluation iterations. Score trajectory: ~735 (iter 0) -> 825 (iter 1) -> 803 (iter 2) -> 793 (iter 3) -> 764 (iter 4). The decline reflects: (a) deeper codebase verification revealing new factual errors (waitForTimeout count, test.setTimeout file count, helpers.ts Playwright type imports), (b) cumulative penalty for issues flagged in iteration 1 that remain unaddressed after 4 rounds, (c) new blindspots discovered (screenshot path breaking change, npm workspace configuration, tsconfig splitting).

Core strengths (unchanged):
- Clear problem definition with mostly-accurate evidence
- Concrete success criteria with mechanical verifiability
- Strong risk assessment with actionable mitigations
- Sound technical direction with compatibility matrix

Persistent weaknesses (unchanged since iteration 1):
- **Industry benchmarking** (55/120) — 4 iterations, zero improvement, name-dropping only
- **File inventory accuracy** — 30-file gap unexplained, evidence counts factually wrong
- **NFR coverage** — replaced one vague claim with another, no methodology
- **helpers.ts extraction mechanics** — Playwright type imports require function extraction, effort uncounted

New findings in iteration 4:
- `test.setTimeout()` used in 7 files, not 1 as claimed
- `helpers.ts` has `import type { Page, Locator }` requiring extraction, not just "moving"
- `screenshot()` helper uses hardcoded relative path that breaks on directory move
- npm workspace configuration is an implicit scope expansion
- No tsconfig.json splitting strategy discussed

The proposal is executable but has accumulating blindspots that will cause implementation surprises, particularly around the helpers.ts extraction and the 30-file inventory gap.
