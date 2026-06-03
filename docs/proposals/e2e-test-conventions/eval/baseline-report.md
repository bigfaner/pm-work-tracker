# Baseline Evaluation Report

**Document:** `docs/proposals/e2e-test-conventions/proposal.md`
**Date:** 2026-06-03
**Scorer:** CTO Adversarial Evaluation

---

## Phase 1: Reasoning Audit

**Argument chain:** Problem (conventions missing + surface mixing) -> Solution (restructure directories + supplement conventions) -> Evidence (framework doc vs actual code gap table) -> Success Criteria (10 checkboxes).

The reasoning is internally coherent: the two problems are related but separable, and the solution addresses both. However, the argument chain has a notable gap: the proposal claims "规范缺失" (missing conventions) is a problem, then acknowledges that framework docs already exist. The real problem is not "missing" but "auto-generated and not grounded in practice." This framing shift matters because it affects scope -- if the docs are wrong, should they be rewritten or appended to?

**Self-contradiction check:** The proposal states in Scope Out: "不修改测试逻辑" (don't modify test logic) and "规范与迁移后代码不一致" is listed as a risk. But the proposed directory structure in the target shows significantly fewer files than exist today (e.g., `items/` has 17 spec files, but the target `tests/api/items/` shows only 3 files and `tests/web/items/` shows only 2). This is a major gap -- many existing test files have no defined destination in the target structure.

---

## Phase 2: Rubric Scoring

### D1. Problem Definition (110 pts)

**Problem stated clearly: 32/40**

Two problems stated concretely. The evidence table with framework vs actual code comparison is strong. Deduction: The first problem ("规范缺失") is slightly misleading -- the conventions exist, they are just auto-generated and poor quality. "Missing" implies absence; "inadequate" would be more precise.

**Evidence provided: 35/40**

The 8-row evidence table is concrete and directly observable in the codebase. Spot-checked against actual code:
- `waitForTimeout` usage: confirmed (15+ instances across web tests)
- 3 HTTP client methods: confirmed (`curl()`, raw `fetch()`, `request.newContext`)
- 4 auth methods: plausible but not exhaustively verified
- All 57 test files use Playwright imports: confirmed

Deduction: The evidence is strong but presented as a summary without raw counts or file-level attribution. "4 种登录方式" is stated without listing which files use which method.

**Urgency justified: 22/30**

The AI agent dependency on conventions is a valid urgency driver. However:
- No mention of how often `/gen-test-scripts` is actually run or how many tests have been generated this way
- "增加 review 成本" is vague -- how much review cost? Measured how?
- No cost-of-delay quantification

**D1 Total: 89/110**

---

### D2. Solution Clarity (120 pts)

**Approach is concrete: 35/40**

The two-part approach (restructure + supplement docs) is clear. The target directory structure with file-level detail is excellent. However, the target structure is incomplete -- it shows only ~20 files in the target but 52 exist today. Major gaps:
- `tests/api/items/` shows 3 files; actual items directory has 8 API-named files plus ambiguous ones (improve-ui-items.spec.ts, untested-endpoints.spec.ts)
- `tests/cli/infra/` shows 4 files; actual infra directory has 9 spec files
- `auth/` directory has 2 files, none shown in target
- `smoke/` has files in both infra and top-level

**User-facing behavior described: 35/45**

The "user" here is the developer/agent writing tests. The convention rules (E-100 through E-123) with correct/error examples are strong. Deduction: No description of the developer experience change after migration. What commands do they run differently? How does CI invocation change? The proposal explicitly defers CI changes to out-of-scope, leaving a gap in observable behavior.

**Technical direction clear: 30/35**

The migration table (Playwright -> Vitest) is clear. The Vitest + native fetch direction is specific. Deduction: The proposal mentions `test.describe.serial` -> `describe.sequential` but there are 4 files using `.serial` and the Vitest equivalent is `describe.sequential` (not `test.sequential` as stated in Constraints section). This imprecision suggests the author has not verified the exact Vitest API.

**D2 Total: 100/120**

---

### D3. Industry Benchmarking (120 pts)

**Industry solutions referenced: 20/40**

No industry solutions, patterns, or external references are cited. The comparison is entirely internal (3 directory layout options + 4 framework options). No mention of how established projects (Next.js, Rails, Django, etc.) organize E2E vs API tests. No reference to testing best practices literature or established patterns.

**At least 3 meaningful alternatives: 25/30**

Four framework alternatives and three directory layout alternatives are presented. "Do nothing" is effectively the first option in each table. The alternatives are genuinely different approaches. Deduction: The "Go httptest" alternative is weakly presented and easily dismissed -- it functions as a borderline straw man ("丢失共享 helpers，不同工具链" is obvious to anyone considering it).

**Honest trade-off comparison: 20/25**

Trade-offs are reasonably honest. The selected surface-first approach acknowledges "跨 surface 共享 helpers 需相对路径" as a con. Deduction: The cons are understated. For example, the journey-first approach con ("结构深") is real but its pro ("同功能测试相邻") is underrated -- many teams prefer co-located tests for discoverability.

**Chosen approach justified against benchmarks: 15/25**

Justification is "符合 Forge surface 生命周期" and "最适合 API surface." These are circular -- the project invented the Forge model, so citing it as external validation is not benchmarking. No external validation of the surface-first approach vs industry practice.

**D3 Total: 80/120**

---

### D4. Requirements Completeness (110 pts)

**Scenario coverage: 30/40**

Four key scenarios are identified. The AI agent scenario is the most novel and well-argued. Deduction: Missing edge case scenarios:
- What happens to existing test files that don't fit neatly into api/web/cli (e.g., `items/improve-ui-items.spec.ts` -- is this api or web?)
- What about files with mixed API and web tests in the same file?
- What about the `KNOWN_FAILURES.md` file?
- The `infra/` directory contains API, CLI, and UI tests mixed -- how are they categorized?

**Non-functional requirements: 20/40**

Performance is mentioned (Vitest "启动快 5-10x" in anti-pattern #6), but:
- "5-10x" is stated without measurement or source
- No security considerations (test data cleanup, credential handling in shared config)
- No compatibility considerations (Node version requirements for native fetch)
- No accessibility requirements for web tests
- The proposal mentions "无 afterAll 清理，数据累积" as a current problem but does not address it in scope (it is out of scope to fix tests)

**Constraints & dependencies: 25/30**

Well-identified constraints: auto-generated markers, frontmatter domains, Vitest serial API difference. Deduction: Missing constraint: the helpers.ts imports `type { Page, Locator } from '@playwright/test'` -- if API tests move to Vitest, they cannot import Playwright types. The shared helpers need to handle this type dependency.

**D4 Total: 75/110**

---

### D5. Solution Creativity (100 pts)

**Novelty over industry baseline: 20/40**

The "practice-first induction" approach (extracting conventions from code rather than prescribing them) is a reasonable but not novel idea. The surface-first directory reorganization is a direct application of the project's own Forge model. No genuine innovation beyond applying an existing internal convention.

**Cross-domain inspiration: 15/35**

No cross-domain inspiration is evident. The proposal stays entirely within the testing domain. No borrowing from compiler design (convention extraction from AST), from linting (pattern enforcement), or from documentation-as-code practices.

**Simplicity of insight: 20/25**

The core insight -- "use the right tool for the right test type" (Vitest for API, Playwright for web) -- is clean and obvious in retrospect. The convention extraction from existing patterns is pragmatic. Not overengineered.

**D5 Total: 55/100**

---

### D6. Feasibility (100 pts)

**Technical feasibility: 30/40**

The migration is mostly mechanical (import changes, file moves). Verified against codebase: 57 files with Playwright imports, 0 with Vitest -- all need conversion. The claim "API 几乎一致" between Playwright and Vitest test APIs is mostly true for basic cases but:
- `test.describe.serial` has different semantics than `describe.sequential` in Vitest (serial forces sequential execution even in parallel mode; sequential is the default behavior in Vitest)
- Playwright's `test.beforeAll` passes a fixture object; Vitest's does not
- Playwright's `test.afterAll` has different cleanup semantics
- The `request.newContext` usage in 3 files needs refactoring, not just import changes

**Resource & timeline feasibility: 20/30**

"约 2-3 小时（~49 个测试文件）" -- I count 52 spec files, and each needs more than an import change (the proposal acknowledges serial/sequential, but there are other API differences). The 2-3 hour estimate seems optimistic. Each file needs: import change, verify API compatibility, test that tests pass. That is more like 5-10 minutes per file = 4-8 hours minimum.

**Dependency readiness: 25/30**

Vitest is already in the project. Native fetch is available in Node 18+. No external API dependencies. Deduction: Does the project's Node version support native fetch? Not stated. The helpers.ts already uses native fetch in `curl()`, so this is likely fine.

**D6 Total: 75/100**

---

### D7. Scope Definition (80 pts)

**In-scope items are concrete: 22/30**

Six in-scope items are listed, each referencing specific deliverables. Deduction: "Surface 优先目录重组" is a broad area -- the proposal should enumerate exact file mappings (existing path -> new path) as a deliverable, since many files lack a clear destination.

**Out-of-scope explicitly listed: 22/25**

Four items explicitly out of scope. Good boundary setting. Deduction: The "不修复现有测试中的反模式代码" out-of-scope item creates a tension: the conventions will document correct patterns, but the existing codebase will violate them. New developers or AI agents will see inconsistency.

**Scope is bounded: 15/25**

The scope is bounded in intent but not in execution detail. The target directory structure shows ~20 files but 52 exist. The scope does not account for:
- Files not shown in the target structure (where do they go?)
- The infra/ directory's mixed-type files
- Files like `items/improve-ui-items.spec.ts` that don't cleanly map to a single surface
- The `e2e/tests/` subdirectory (what is it?)
- `config-setup.ts` in infra (not a spec file, but exists)

**D7 Total: 59/80**

---

### D8. Risk Assessment (90 pts)

**Risks identified: 25/30**

Four risks identified. The regression risk and spec drift risk are the most meaningful. Deduction: Missing risks:
- Shared helpers.ts importing Playwright types breaks API-only Vitest runs
- The `KNOWN_FAILURES.md` file may reference paths that no longer exist
- Risk of losing test history/traceability if files are renamed/moved
- Risk that CI pipelines break if directory structure changes but CI is out of scope

**Likelihood + impact rated: 22/30**

Ratings are reasonable (not everything is high). Deduction: The "Vitest 和 Playwright 的 test API 有细微差异" risk is rated L likelihood but M impact. From codebase inspection, there are 4 files using `.serial` and 3 using `request.newContext` -- this is more than "细微" (minor) and affects 13% of files. Likelihood should be M.

**Mitigations are actionable: 22/30**

Mitigations are somewhat actionable but could be more specific:
- "逐个 surface 迁移并验证" -- good, but what is the verification command?
- "先完成迁移，再从最终代码提炼规范" -- this changes the execution order from what the scope implies (both are in scope simultaneously)
- "逐个文件检查" -- for `test.describe.serial` -> `describe.sequential` -- this is specific and actionable

**D8 Total: 69/90**

---

### D9. Success Criteria (80 pts)

**Criteria are measurable and testable: 25/30**

Most criteria are checkable: directory exists, no Playwright in API tests, tests pass, convention rules present. Deduction:
- "所有 API 测试迁移后通过" -- does this include currently-failing tests (KNOWN_FAILURES.md)?
- "两个 surface 文件各包含反模式清单（5 条以上）" -- countable, good
- "与现有框架文档的反模式不重复" -- how is this verified? Manual review?

**Coverage is complete: 18/25**

Gaps in SC coverage:
- No SC for CLI tests passing after migration (only API and Web are listed)
- No SC for shared helpers working from both api/ and web/ relative paths
- No SC for the domains frontmatter update (listed in scope but not in SC -- wait, it is: last SC item)
- No SC for import path resolution (tests can find shared/ correctly)

**SC internal consistency: 18/25**

Consistency analysis by cluster:

*Cluster: Directory restructure*
- SC: "tests/ 目录按 surface 优先结构组织" and "所有 API/Web 测试迁移后通过" -- consistent
- SC: "API 测试全部使用 Vitest，无 Playwright 依赖" -- but shared helpers.ts imports `Page` and `Locator` from Playwright. If API tests import from shared helpers that transitively import Playwright, this SC may be unsatisfiable.

*Cluster: Convention docs*
- SC: "api/core.md 包含 7 条 API 实践规范（E-100 ~ E-106）" -- E-100 through E-106 is 7 rules. Consistent.
- SC: "web/core.md 包含 6 条 Web 实践规范（E-110 ~ E-115）" -- E-110 through E-115 is 6 rules. Consistent.
- SC: "index.md 包含 4 条 helpers 使用规范（E-120 ~ E-123）" -- E-120 through E-123 is 4 rules. Consistent.

*Ambiguity: The SC says "所有 API 测试迁移后通过" but the scope says infra/ API tests exist and are not listed in the target structure.*

**D9 Total: 61/80**

---

### D10. Logical Consistency (90 pts)

**Solution addresses the stated problem: 30/35**

The two-part solution maps directly to the two stated problems. Directory restructure addresses surface mixing; conventions supplement addresses framework doc gaps. Deduction: The conventions will be written from pre-migration code patterns, but the risk mitigation says "先完成迁移，再从最终代码提炼规范" -- if conventions are post-migration, they should describe Vitest patterns, but all evidence in the proposal is from Playwright-era code. There is a temporal inconsistency.

**Scope <-> Solution <-> SC aligned: 22/30**

Most items have corresponding scope entries and success criteria. Gaps:
- Scope: "更新各文件的 domains frontmatter" -- SC covers this (last item)
- Scope: "共享 helpers 提取到 tests/shared/" -- no SC for helpers being functional after move
- Target structure shows `tests/api/smoke/deploy.spec.ts` but existing `infra/deploy-smoke.spec.ts` is in infra/ -- is this an API test or a CLI test? Inconsistent classification.

**Requirements <-> Solution coherent: 20/25**

The four key scenarios map to the solution. The AI agent scenario maps to conventions. The developer scenario maps to conventions. The code review scenario maps to anti-pattern lists. The CI scenario maps to surface separation. Deduction: The CI scenario is listed as a requirement but CI changes are out of scope. This means the requirement is acknowledged but not addressed.

**D10 Total: 72/90**

---

## Phase 3: Blindspot Hunt

**[blindspot-1] File classification is underspecified.** The target directory structure lists ~20 files but 52 exist. Files like `items/improve-ui-items.spec.ts`, `items/untested-endpoints.spec.ts`, `items/sub-item-edit.spec.ts`, `items/refresh-button.spec.ts`, `items/improve-ui-items.spec.ts`, `items/view-pages.spec.ts` have no defined destination. The `infra/` directory has 9 spec files that mix API/CLI/UI concerns but the target shows only 4 CLI files. This is the single biggest execution risk.

**[blindspot-2] Shared helpers Playwright type coupling.** The existing `helpers.ts` imports `type { Page, Locator } from '@playwright/test'`. If API tests are moved to Vitest and they import from `tests/shared/helpers.ts`, the transitive Playwright dependency remains. The proposal does not address this type dependency -- either helpers.ts needs conditional imports, or API tests will still pull in Playwright at the type level, contradicting SC "API 测试全部使用 Vitest，无 Playwright 依赖."

**[blindspot-3] No rollback plan.** The proposal involves moving 52 files, changing imports, and splitting configuration. If the migration is half-complete and a blocking issue is discovered (e.g., Vitest API incompatibility), there is no rollback strategy. The "逐个 surface 迁移" mitigation is directional but not a rollback plan.

**[blindspot-4] Temporal ordering contradiction.** The Scope says both directory restructure AND convention docs are in scope. The Risk section says "先完成迁移，再从最终代码提炼规范" (migration first, then conventions). But the proposed conventions are written in present tense with Playwright-era code examples. If conventions are written from migrated code, all examples should use Vitest patterns -- but no Vitest example code is shown.

**[blindspot-5] KNOWN_FAILURES.md ignored.** There is a `KNOWN_FAILURES.md` in `tests/e2e/` that presumably references test files by path. The proposal does not address what happens to this file during reorganization. Paths will break.

**[blindspot-6] Estimate is likely underestimated.** "约 2-3 小时（~49 个测试文件）" for file moves, import changes, API compatibility verification, and test execution. With 52 files (not 49), API differences beyond simple import swaps (serial, request.newContext, fixture patterns), and verification needed per file, this is likely 6-10 hours of careful work.

**[blindspot-7] test-results and results/ directories.** The existing structure has `test-results/` and `results/screenshots/` directories that are configured relative to the current test location. Moving tests without updating these paths will break screenshot output and test artifacts.

---

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| D1. Problem Definition | 89 | 110 |
| D2. Solution Clarity | 100 | 120 |
| D3. Industry Benchmarking | 80 | 120 |
| D4. Requirements Completeness | 75 | 110 |
| D5. Solution Creativity | 55 | 100 |
| D6. Feasibility | 75 | 100 |
| D7. Scope Definition | 59 | 80 |
| D8. Risk Assessment | 69 | 90 |
| D9. Success Criteria | 61 | 80 |
| D10. Logical Consistency | 72 | 90 |
| **Total** | **735** | **1000** |

---

## ATTACKS

1. **[D7] Target directory structure is incomplete** -- 52 files exist, ~20 shown in target. Files like `items/improve-ui-items.spec.ts`, `items/untested-endpoints.spec.ts`, `items/sub-item-edit.spec.ts` have no defined destination. Must provide a complete file mapping table (every existing file -> new location).

2. **[D6] Timeline estimate is optimistic** -- "~49 个测试文件" but actual count is 52. More importantly, the claim "基本只需改 import" understates the work: 4 files use `.serial`, 3 use `request.newContext`, helpers.ts has Playwright type imports. Must revise estimate to account for non-trivial refactoring.

3. **[D2] Shared helpers Playwright coupling ignored** -- `helpers.ts` line 6: `import type { Page, Locator } from '@playwright/test'`. API tests moving to Vitest cannot transitively import Playwright. Must propose how shared helpers will handle dual-framework type dependencies.

4. **[D3] Zero external benchmarking** -- No industry references, no published patterns cited, no comparison to how established frameworks (Next.js, Remix, Rails) organize test types. Must include at least 2-3 external references for test directory organization and API testing frameworks.

5. **[D10] Temporal ordering contradiction** -- Risk mitigation says "先完成迁移，再从最终代码提炼规范" but conventions in the proposal are written from pre-migration code. If conventions are post-migration, all code examples should be Vitest-based. Must clarify: are conventions written before or after migration? If after, remove the convention examples from the proposal and define the extraction process.

6. **[D9] SC "无 Playwright 依赖" may be unsatisfiable** -- If shared helpers.ts imports Playwright types, API tests importing from shared/ have a transitive Playwright dependency. Must either: (a) split helpers into Playwright-dependent and framework-agnostic modules, or (b) change SC to "API tests do not directly import Playwright."

7. **[D4] Infra directory classification is ambiguous** -- `infra/` contains API tests (config-yaml-api, jlc-schema-api, schema-alignment-api), CLI tests (bizkey-cli, config-yaml-cli, jlc-schema-cli, lint-keywords-cli, e2e-rebuild-cli), and UI tests (schema-alignment-ui, unify-permission-checks-build). The target shows only 4 CLI files under `tests/cli/infra/`. Must specify where each infra file goes.

8. **[blindspot] No rollback plan** -- Moving 52 files with no documented rollback procedure. Must include a rollback strategy (e.g., git branch per surface, or complete migration in one branch with revert capability).

9. **[blindspot] KNOWN_FAILURES.md path references will break** -- File exists in tests/e2e/ and likely contains paths to test files. Must address this file in scope or explicitly document how path references will be updated.

10. **[D8] Missing risk: artifact path breakage** -- `results/screenshots/` and `test-results/` are configured relative to current test locations. Moving tests changes relative paths. Must identify all path-dependent configuration and include in migration scope.
