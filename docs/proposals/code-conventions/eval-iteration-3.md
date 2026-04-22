---
proposal: docs/proposals/code-conventions/proposal.md
iteration: 3
score: 72/100
date: 2026-04-22
---

# Proposal Evaluation Report — Code Quality: Performance, Duplication & Readability

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Problem Definition | 15 | 20 |
| Solution Clarity | 15 | 20 |
| Alternatives Analysis | 11 | 15 |
| Scope Definition | 12 | 15 |
| Risk Assessment | 10 | 15 |
| Success Criteria | 11 | 15 |
| Vague language penalty | -2 | — |
| **Total** | **72** | **100** |

---

## Changes Since Iteration 2

Three attacks from iteration 2 were addressed:

| Attack | Status | Notes |
|--------|--------|-------|
| Developer experience after completion never described | ✅ Fixed | "完成后的可观测状态" added to each phase with concrete developer-facing outcomes |
| Four-phase bundle never justified | ✅ Partial | "四阶段捆绑的依赖关系" paragraph added; dependency chain between phases 1–3 is now argued |
| No timeline and no sequencing constraints | ✅ Fixed | T-shirt sizing per phase added; sequencing constraints stated; total target "3 个 sprint（约 3 周）" |

Score improvement: 64 → 72. Remaining structural weaknesses are in risk assessment, alternatives rationale, and success criteria testability.

---

## Dimension Breakdown

### 1. Problem Definition — 15/20

**Problem stated clearly (6/7)**

Performance problems remain well-specified with file locations and concrete descriptions. "规范缺失" is still underspecified — "AI每次会话无法获得一致的风格指引" names the symptom but not the scope of inconsistency. One historical snake_case incident is cited; it is unclear whether this is isolated or representative. -1.

**Evidence provided (4/7)**

The N+1 example is concrete. The GORM log observation is a real data point. However: "约60行相同的fetch+filter代码" still uses "约" (approximately) — the exact line count is verifiable and should be stated. The 21 duplicate count appears only in the success criteria section, not in the Problem section where it would serve as evidence. No query latency measurements, no response time data. -3.

**Urgency justified (5/6)**

The "Why now" section provides a no-delete-path argument, a GORM log observation, a 6-month projection (500+ items → 500+ mutex entries), and a historical AI session incident. The projection is concrete. -1 because the current data volume is never stated — "500+" is a projection without anchoring to today's item count or growth rate, making the timeline unverifiable.

---

### 2. Solution Clarity — 15/20

**Approach is concrete (6/7)**

Phase 0 is strong: specific method signatures, specific rename targets. Phase 1 lists exact file paths. Phase 2 now names `tagliatelle` and the ESLint rules ("API 层文件命名（camelCase.ts）和组件导出命名（PascalCase）") — improvement over iteration 2. Phase 3 "识别重复UI模式，抽取可复用组件（具体组件名在 Phase 3 开始前确定）" remains discovery work, not a deliverable. -1.

**User-facing behavior described (5/7)**

Significantly improved. Each phase now has a "完成后的可观测状态" section. Phase 0: "GORM 查询日志中 resolveAssigneeNames 从 N 条串行查询变为 1 条 WHERE id IN (...) 查询" — concrete and verifiable. Phase 2: "golangci-lint run ./... 在检测到 snake_case JSON tag 时以非零退出码退出，阻断 CI" — concrete. Phase 3: "grep -r 'ErrRecordNotFound' backend/internal 返回 ≤10 处" — concrete. Remaining gap: Phase 1's AI session loading verification ("通过检查文件存在且 CLAUDE.md Key Documents 章节已列出路径来验证") tests file existence, not actual AI session behavior — the observable state claimed ("Claude Code 在读取 CLAUDE.md 时自动加载这些规则文件") is not verified by the proposed check. -2.

**Distinguishes from alternatives (4/6)**

The "四阶段捆绑的依赖关系" paragraph now argues the dependency chain: Phase 1 → Phase 2 → Phase 3. This is a real improvement. However, the argument for bundling Phase 0 with Phases 1–3 is still asserted rather than demonstrated — "同根同源（AI 辅助开发缺乏项目级约束）" is a claim, not an argument. Phase 0 fixes N+1 queries; Phases 1–3 fix documentation and lint. These have different root causes (implementation bug vs. missing process). The proposal does not explain why fixing the N+1 query requires also writing ARCHITECTURE.md in the same proposal. -2.

---

### 3. Alternatives Analysis — 11/15

**At least 2 alternatives listed (5/5)**

Four alternatives including Do Nothing. Full marks.

**Pros/cons for each (3/5)**

Alternative D's cons are genuinely honest and specific — the strongest section. Alternative B's pros ("后续改动减少重复文件触碰") are still unquantified — vague language penalty applied separately. Alternative A's cons ("总工作量比单做 Phase 0 大；Phase 3 工期较长") are real but do not acknowledge the risks introduced by the performance fixes themselves (behavior changes, test gaps). Alternative C's cons are real but not elaborated. -2.

**Rationale for chosen approach (3/5)**

Improved. The dependency chain argument ("lint 规则必须与文档中的决策对齐；清理后需要 lint 全量验证无回归") is a legitimate reason to sequence Phases 1–3 together. The "avoid touching files twice" argument ("Phase 3 清理在 lint 就位后执行，避免二次触碰同一文件") is concrete. However, the verdict still does not address why Phase 0 belongs in the same proposal as Phases 1–3 — the dependency chain runs Phase 1 → Phase 2 → Phase 3, with Phase 0 explicitly noted as independent ("Phase 0 独立可先行"). If Phase 0 is independent, the rationale for bundling it here is not argued. -2.

---

### 4. Scope Definition — 12/15

**In-scope items are concrete (4/5)**

Phase 0 items are concrete. Phase 1 file paths are concrete. Phase 2 now names specific lint rules for both Go and TypeScript — improvement. Phase 3 "前端可复用组件抽取（≥3 个，名称在 Phase 3 开始前确定）" remains open-ended discovery work in the scope section. -1.

**Out-of-scope explicitly listed (4/5)**

Now includes "CI/CD 流水线集成 lint" — improvement. Still missing: whether existing tests are in scope for updates when Phase 0 changes repo interfaces; no mention of a migration guide or timeline for existing code that violates the new lint rules (the "warn 模式" decision is in the out-of-scope list but the implications for existing violations are not addressed). -1.

**Scope is bounded (4/5)**

T-shirt sizing per phase is now present (M, S, S, L). Sequencing constraints are stated. Total target "3 个 sprint（约 3 周）" is a real boundary. -1 because the "约" estimates have no basis stated — Phase 0 is "约3–5天" but the estimate methodology is not given. A reader cannot validate whether M = 3–5 days is realistic without knowing team size or velocity. Minor deduction only.

---

### 5. Risk Assessment — 10/15

**Risks identified (4/5)**

Four risks listed, all meaningful. Still missing: lint rules (Phase 2) may conflict with a large volume of existing code, requiring a bulk fix pass before enforcement can be enabled — this is the most common failure mode for lint adoption and would block Phase 2 completion. The proposal notes "warn 模式启用，不强制修复存量违规" in the out-of-scope section but does not treat this as a risk. -1.

**Likelihood + impact rated (3/5)**

Ratings are provided but remain conservative. "规范文档过于理想化" is rated Medium/Medium — if the docs are never adopted by AI sessions (a plausible outcome given that "可在AI会话中通过@rules引用加载" is not actually verified), the entire Phase 1 investment is wasted, which is High impact. The linkageMuMap risk is Low/High — concurrent bugs in production are typically catastrophic, not merely "High." -2.

**Mitigations are actionable (3/5)**

"改动前后对比测试；现有 repo 测试覆盖" — what tests? Against what baseline? The existing test suite is not described, so "现有 repo 测试覆盖" is not actionable without knowing what it covers. The other three mitigations are actionable. -2.

---

### 6. Success Criteria — 11/15

**Criteria are measurable (4/5)**

Phase 0 criteria are mostly binary and measurable. "linkageMuMap 不再无限增长" still has no verification method — code review? load test? memory profiling? The criterion is stated but not operationalized. -1.

**Coverage is complete (4/5)**

Phase 0: 1:1 mapping with scope — good. Phase 1: section checklists for ARCHITECTURE.md and DECISIONS.md — good. Phase 2: golangci-lint criterion is specific ("能检测出 snake_case JSON tag 违规"), but there is no ESLint success criterion — the frontend lint work has no measurable completion gate. Phase 3: "至少抽取3个可复用前端UI组件（需在Phase 3开始前列出具体组件名）" defers specificity — the criterion is incomplete as written. -1.

**Criteria are testable (3/5)**

Phase 0: testable. ARCHITECTURE.md and DECISIONS.md section checklists: testable. ".claude/rules/*.md已创建，可在AI会话中通过@rules引用加载" — the verification method proposed ("检查文件存在且 CLAUDE.md Key Documents 章节已列出路径") tests file existence, not AI session loading behavior. The criterion claims AI sessions will load the rules but the test does not verify this. "至少抽取3个可复用前端UI组件" cannot be tested until component names are specified. -2.

---

## Vague Language Penalties

| Instance | Location | Penalty |
|----------|----------|---------|
| "后续改动减少重复文件触碰" (reduced file touching, unquantified) | Alternatives B, Pros | -2 |

Total penalty: **-2**

---

## Top Attacks

### Attack 1 — Solution Clarity: AI session rule loading is claimed but not verifiable

The proposal claims that `.claude/rules/*.md` files will be automatically loaded by Claude Code in AI sessions, and proposes verifying this by checking file existence and CLAUDE.md path references. File existence does not verify session loading behavior. If the rules are never actually loaded, Phase 1's core value proposition fails silently.

Quote: "可在AI会话中通过@rules引用加载（无需额外配置，通过检查文件存在且 CLAUDE.md Key Documents 章节已列出路径来验证）"

What must improve: Replace the file-existence check with a behavioral test. For example: "In a new Claude Code session, reference @rules/naming.md and verify the rule content is visible in context" or "Run a known-bad snippet through Claude Code and verify it flags the snake_case tag." The verification must test the behavior, not the precondition.

---

### Attack 2 — Risk Assessment: lint adoption failure mode is absent

The most common failure mode for lint adoption — existing code has too many violations to enable enforcement — is not listed as a risk. The proposal notes "warn 模式启用，不强制修复存量违规" in the out-of-scope section, but this is a scope decision, not a risk mitigation. If the violation count is high enough that warn mode is permanent, Phase 2's success criterion ("以非零退出码退出，阻断 CI") is never achievable.

Quote: *(no risk entry for lint adoption failure)*

What must improve: Add a risk entry: "Phase 2 lint rules remain in warn mode indefinitely because existing violation count is too high to fix in scope." Rate it Medium/High. Mitigation: count existing violations before enabling rules; if count exceeds threshold X, add a Phase 2.5 cleanup task before enforcement.

---

### Attack 3 — Alternatives Analysis: Phase 0 independence undermines the bundle rationale

The proposal explicitly states "Phase 0 独立可先行" and "Phase 1 可与 Phase 0 并行" — meaning Phase 0 has no dependency on Phases 1–3 and vice versa. Yet the chosen approach bundles all four phases into one proposal. The dependency chain argument (Phase 1 → Phase 2 → Phase 3) justifies sequencing Phases 1–3 together, but does not justify including Phase 0 in the same proposal. Alternative A' (Phase 0 only) is dismissed as "治标不治本" without arguing why the root cause must be addressed in the same proposal rather than a follow-on.

Quote: "Phase 0 独立可先行" and "✅ Recommended — 性能、规范、lint、清理四类问题同根同源"

What must improve: Either argue explicitly why Phase 0 must be bundled with Phases 1–3 (e.g., "without lint in place, the Phase 0 fixes will be undone within N AI sessions"), or split into two proposals: one for Phase 0 (immediate performance fix) and one for Phases 1–3 (process improvement). The current structure creates a large, hard-to-review scope without a clear justification for the bundle.
