# Eval Report: E2E Test Conventions — Iteration 1

**Score: 735/1000**

---

## DIMENSIONS

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Problem Definition | 92 | 110 |
| 2. Solution Clarity | 90 | 120 |
| 3. Industry Benchmarking | 55 | 120 |
| 4. Requirements Completeness | 70 | 110 |
| 5. Solution Creativity | 65 | 100 |
| 6. Feasibility | 75 | 100 |
| 7. Scope Definition | 72 | 80 |
| 8. Risk Assessment | 78 | 90 |
| 9. Success Criteria | 68 | 80 |
| 10. Logical Consistency | 70 | 90 |

**Attack density: 5 attacks on pre-revised regions, 15 attacks on unannotated regions.**

---

## ATTACKS

### Dimension 1: Problem Definition (92/110)

**Attack 1.1** [Problem stated clearly — 32/40]: Two distinct problems are conflated without establishing whether they share a root cause or are coincidental. The first (missing practical conventions) is a documentation gap; the second (surface mixing) is an architectural gap. The proposal never argues why these *must* be solved together. If surface reorganization were done first, would the conventions gap persist identically? This causal chain is absent.

**Attack 1.2** [Evidence provided — 35/40]: The evidence table is strong but contains an inaccuracy. It states "3 种方式混用（curl/fetch/request.newContext）" under HTTP Client, but code inspection shows only 2 files use `playwright.request.newContext` directly (`item-list.spec.ts`, `item-list-fixes.spec.ts`, `progress-auto-status.spec.ts` — 3 files, confirmed). The "大量 `waitForTimeout(2000)` 硬编码延迟" claim is verifiable but the word "大量" is vague — actual count is ~177 occurrences across 52 files, but concentrated in ~10 files (item-list.spec.ts alone has 54). The evidence is directionally correct but not precisely quantified.

**Attack 1.3** [Urgency justified — 25/30]: The urgency argument rests entirely on AI Agent workflow (`/gen-test-scripts`). The cost of delay is described as "review cost increase" which is qualitative. Missing: how many agent-generated tests have been reviewed to date? What was the defect rate? Without this, urgency is plausible but not quantified.

---

### Dimension 2: Solution Clarity (90/120)

**Attack 2.1** [User-facing behavior — 30/45]: The proposal describes an internal restructuring (directory layout, framework migration) but never describes what the *developer or agent experiences differently*. The "Key Scenarios" section lists scenarios but describes them as bullet points, not as observable behaviors. What does the developer type differently? What does the AI agent's output look like before vs after? There is no "before/after" concrete example of a generated test.

**Attack 2.2** [Approach is concrete — 35/40]: The directory structure is detailed and the migration table is clear. However, the proposed `tests/api/` structure in the target directory shows file names like `rbac.spec.ts` that don't match current naming (e.g., `rbac-api.spec.ts`). The proposal mentions "去掉 -api 前缀" but the current files use `-api` suffix, not prefix. This naming confusion suggests the author hasn't fully mapped existing files to target locations.

**Attack 2.3** [Technical direction clear — 25/35]: The Playwright-to-Vitest migration direction is clear, but critical technical details are missing. The proposal says "API 测试文件基本只需改 import 和文件名后缀" but code inspection shows at least 3 test files use `playwright.request.newContext` directly (not through `curl()`), meaning these need full HTTP client rewriting, not just import changes. The compatibility matrix addresses `.serial` but ignores Playwright's `test.use()` fixtures, which may exist in some files.

---

### Dimension 3: Industry Benchmarking (55/120)

**Attack 3.1** [Industry solutions referenced — 10/40]: No external tools, published patterns, or open-source projects are cited. The "Alternatives" section only compares internal directory organization strategies. Missing: how do mature projects (Next.js, Django, Rails) organize API vs browser tests? What does the Playwright team itself recommend? What patterns exist in the Vitest ecosystem? This is entirely self-referential.

**Attack 3.2** [At least 3 meaningful alternatives — 20/30]: The directory structure alternatives include 3 options plus "do nothing" (implicit in "维持现状"). The framework alternatives include 4 options. However, some are weak straw men: "Playwright request API" is presented as "改进不大" without evidence, and "Go httptest" is rejected with "团队以 TS 为主" — but this team constraint was never stated as a requirement. The alternatives are genuine but the evaluation is superficial.

**Attack 3.3** [Honest trade-off comparison — 15/25]: The "Cons" column for the selected Surface-first approach says "跨 surface 共享 helpers 需相对路径" which understates the real cost: every shared import is a coupling point that makes independent evolution harder. No alternative was evaluated for this coupling cost.

**Attack 3.4** [Chosen approach justified against benchmarks — 10/25]: No industry benchmark was used as a reference point, so there is nothing to justify against. The rationale is entirely internal logic ("符合 Forge surface 生命周期").

---

### Dimension 4: Requirements Completeness (70/110)

**Attack 4.1** [Scenario coverage — 25/40]: The Key Scenarios section lists 4 scenarios but misses critical edge cases:
- What happens when a test needs both API and Web assertions? (e.g., "create via API, verify in UI")
- What happens to the `smoke/full-e2e.spec.ts` test that crosses all surfaces?
- How are shared test data fixtures handled across surfaces?
- How does the `tests/e2e/tests/` subdirectory (visible in directory listing) factor in?

**Attack 4.2** [Non-functional requirements — 20/40]: The proposal claims Vitest is "启动快 5-10x" compared to Playwright but provides no benchmark data. This is a vague performance claim penalized under deduction rules. No other NFRs are addressed: test execution order guarantees, parallel execution strategy, test timeout configuration differences between frameworks, or memory usage.

**Attack 4.3** [Constraints & dependencies — 25/30]: The constraints section is adequate — it identifies the auto-generated marker, frontmatter domains, and `.serial` API difference. However, it misses the constraint that `playwright.request.newContext` is used in some API tests, which means not all API tests can be migrated with a simple import change.

---

### Dimension 5: Solution Creativity (65/100)

**Attack 5.1** [Novelty over industry baseline — 25/40]: The "Innovation Highlights" section frames "practice-first" as novel, but extracting conventions from existing code is standard practice, not an innovation. The surface-first directory organization follows the Forge model rather than introducing a new pattern. There is no novel insight beyond applying an existing framework's prescribed structure.

**Attack 5.2** [Cross-domain inspiration — 20/35]: No cross-domain ideas are cited. The proposal stays entirely within testing convention patterns. No mention of how other domains handle multi-tool test organization (e.g., monorepo tooling, polyglot testing).

**Attack 5.3** [Simplicity of insight — 20/25]: The core insight — "use Vitest for API tests since they don't need a browser" — is genuinely simple and elegant. This is the strongest creative element.

---

### Dimension 6: Feasibility (75/100)

**Attack 6.1** [Technical feasibility — 30/40]: The proposal claims "API 测试文件基本只需改 import 和文件名后缀" but this is inaccurate. At least 3 files use `playwright.request.newContext` and would need HTTP client rewriting. The migration is feasible but the stated effort is underestimated. Additionally, 20 test files have no surface suffix in their names — classifying these correctly requires manual inspection, adding to effort.

**Attack 6.2** [Resource & timeline feasibility — 22/30]: The estimate of "约 3-5 小时" for ~52 test files (not "~49" as stated — actual count is 52) is tight. With 3+ files needing full HTTP client rewriting, plus the ~20 ambiguous-surface files needing manual classification, the rework estimate should be higher. The "迁移前审计 ~30 min" is a good addition.

**Attack 6.3** [Dependency readiness — 23/30]: The project already uses Vitest for frontend tests, so the dependency exists. However, the proposal doesn't confirm whether Vitest's API test capabilities (e.g., HTTP test helpers, environment configuration) are already set up or need additional configuration. The `vitest.config.ts` for the API surface is shown in the target structure but its contents aren't specified.

---

### Dimension 7: Scope Definition (72/80)

**Attack 7.1** [In-scope items are concrete — 27/30]: The in-scope items are mostly concrete deliverables. However, "更新各文件的 `domains` frontmatter" is vague — which files, and what domains?

**Attack 7.2** [Out-of-scope explicitly listed — 22/25]: Out-of-scope items are well-defined. The exclusion of "修复现有测试中的反模式代码" is clearly stated, which prevents scope creep. One gap: no mention of whether updating CI configuration for the new directory structure is in or out of scope (it's mentioned in risk mitigation but not in scope).

**Attack 7.3** [Scope is bounded — 23/25]: The phased migration plan provides clear boundaries. The "无需 PRD/tech-design" justification in Next Steps is reasonable for a doc+refactor task.

---

### Dimension 8: Risk Assessment (78/90)

**Attack 8.1** [Risks identified — 24/30]: Four risks are identified. The CI risk (pre-revised, high severity) is a strong addition. Missing risks:
- Risk of test file misclassification (20 files with no surface suffix)
- Risk of `playwright.request.newContext` usage in API tests requiring more than import changes
- Risk that `tests/shared/helpers.ts` being "Playwright-free" isn't enforced by any tooling — the proposed grep check in CI is for `tests/api/` but not for `tests/shared/`

**Attack 8.2** [Likelihood + impact rated — 26/30]: Ratings are honest — not all "low likelihood, high impact." The CI risk is correctly rated H/H. The spec drift risk is M/M, reasonable.

**Attack 8.3** [Mitigations are actionable — 28/30]: The mitigations are the strongest part of this section. The CI transition plan is concrete (parallel jobs, phased cutover). The drift detection grep check is specific. The "re-export shim" strategy for backward compatibility is actionable. This is well done.

---

### Dimension 9: Success Criteria (68/80)

**Attack 9.1** [Criteria are measurable and testable — 22/30]: Most criteria are testable ("API 测试全部使用 Vitest, 无 Playwright 依赖" can be verified with `grep`). However:
- "api/core.md 包含 7 条 API 实践规范（E-100 ~ E-106）" — the count is now 8 anti-patterns (items 1-8) but only 7 rules (E-100 to E-106). The SC says "7 条" but the anti-pattern list has 8 entries. Inconsistent.
- "与现有框架文档的反模式不重复" — this is subjective and hard to verify mechanically.

**Attack 9.2** [Coverage is complete — 20/25]: Missing SC for:
- `tests/shared/` directory creation and helpers extraction
- Re-export shim creation and removal
- CLI test migration
- `config.yaml` relocation
- The "迁移前审计" deliverable mentioned in timeline

**Attack 9.3** [SC internal consistency — 26/25]: SC item "所有 API 测试迁移后通过" is consistent with "API 测试全部使用 Vitest". However, the SC "两个 surface 文件各包含反模式清单（5 条以上）" conflicts slightly with the actual anti-pattern counts: API has 8 items, Web has 7 items. The "5 条以上" floor is satisfied but the mismatch with SC item "7 条 API 实践规范" is confusing — are "rules" and "anti-patterns" counted separately or together?

---

### Dimension 10: Logical Consistency (70/90)

**Attack 10.1** [Solution addresses stated problem — 25/35]: The proposal states two problems: (1) missing practical conventions, (2) surface mixing. The solution addresses both, but there's a sequencing issue: Phase 5 (conventions) comes *after* Phases 1-4 (migration). This means the conventions will be written from the migrated code, which is correct for "从实践提炼规范" but contradicts the stated urgency — if AI agents need conventions *now* to generate tests, they won't get them until after the full migration completes.

**Attack 10.2** [Scope ↔ Solution ↔ Success Criteria aligned — 22/30]: The scope includes "API 测试从 Playwright 迁移到 Vitest" and the SC includes "API 测试全部使用 Vitest" — aligned. But the scope says "不修改 helpers.ts 的实现（仅移动位置）" while the Evidence table shows helpers have anti-patterns (like `waitForTimeout`). The proposal plans to codify these as anti-patterns in conventions but won't fix them, which is consistent but worth flagging as a known gap.

**Attack 10.3** [Requirements ↔ Solution coherent — 23/25]: The requirements (Key Scenarios) map to the solution: agent reads conventions → conventions are updated ✓; developer writes tests → helpers documented ✓; CI runs by surface → directory restructured ✓. One orphan: the Code Review scenario ("以 conventions 中的反模式清单为检查标准") has no corresponding success criterion about review integration.

---

## BLINDSPOT HUNT: What the rubric missed

1. **No rollback plan**: The proposal describes a phased migration but never discusses what happens if Phase 2 (API migration to Vitest) fails. Do you revert to `tests/e2e/`? The re-export shim helps during transition but there's no rollback after shim removal.

2. **`tests/e2e/tests/` subdirectory**: The current directory listing shows a `tests/` subdirectory inside `tests/e2e/` (line from `ls` output). The proposal's target structure doesn't account for this. What's in it? Is it relevant?

3. **Test count discrepancy**: The proposal says "~49 个测试文件" but actual count is 52. This is a minor factual error that could indicate the author hasn't done a complete file inventory.

4. **No mention of test parallelism**: Vitest and Playwright have different parallel execution models. The proposal doesn't discuss whether this affects test outcomes (e.g., test isolation, shared state).

5. **The `beforeAll`/`afterAll` gap in Web tests**: The evidence says "无 afterAll 清理，数据累积" but the Web anti-pattern list only mentions it as item #7 (pre-revised). The API anti-pattern list also has it as item #7 (pre-revised). Since "修复现有测试中的反模式代码" is out of scope, this means data accumulation will continue post-migration — the proposal documents the problem but deliberately doesn't solve it.

6. **Token cost of conventions**: The CLAUDE.md says conventions are "按需加载" to save tokens. The proposal adds substantial content to api/core.md, web/core.md, and index.md. No analysis of whether this bloats the token cost for agents that load these files.
