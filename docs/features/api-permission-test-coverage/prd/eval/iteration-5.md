---
date: "2026-04-28"
doc_dir: "docs/features/api-permission-test-coverage/prd/"
iteration: "5"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 5

**Score: 96/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  19      │  20      │ ✅          │
│    Background three elements │   7/7    │          │            │
│    Goals quantified          │   7/7    │          │            │
│    Logical consistency       │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  20      │  20      │ ✅          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   7/7    │          │            │
│    Decision + error branches │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  19      │  20      │ ✅          │
│    Tables complete           │   7/7    │          │            │
│    Field descriptions clear  │   7/7    │          │            │
│    Validation rules explicit │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  19      │  20      │ ✅          │
│    Coverage per user type    │   7/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  19      │  20      │ ✅          │
│    In-scope concrete         │   7/7    │          │            │
│    Out-of-scope explicit     │   6/7    │          │            │
│    Consistent with specs     │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  96      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Background logical consistency | 背景说"53 处 perm() 绑定，其中任何一处权限码误写均无法被现有测试发现"，但目标仅覆盖 12 个端点，41 个端点的覆盖缺口未被承认 | -1 pt (logical gap) |
| I-D 断言逻辑 step 1 | "从 codes.go 提取所有 const 权限码" — 隐含假设所有权限码均以 Go const 定义，若存在 var 或字符串字面量定义则静默漏检 | -1 pt (validation rule implicit assumption) |
| Story 5 AC vs I-D spec | Story 5 AC 说"测试文件中未出现该字符串"（字符串完全不存在），但 spec I-D step 2 已修正为"不包括注释或日志中的字符串"（字符串存在但仅在注释中仍应失败）——两者测试的是不同条件 | -1 pt (inconsistency between sections) |
| Out of Scope | 四项 out-of-scope 条目均未区分"延期"与"永久排除"，读者无法判断哪些可能出现在后续 PRD 中 | -1 pt (out-of-scope not fully explicit) |

---

## Attack Points

### Attack 1: User Stories — Story 5 AC tests a weaker condition than the spec requires

**Where**: Story 5 AC, first block: `Given codes.go 中定义了权限码 foo:bar，但测试文件中未出现该字符串` / `Then 构建失败，输出 missing test coverage for: foo:bar`. Spec I-D step 2: `从测试文件中提取所有作为权限码参数传入的字符串（即出现在 permCodes 注入或集成测试矩阵中的权限码值，不包括注释或日志中的字符串）`.

**Why it's weak**: The spec was correctly tightened in a prior iteration to exclude strings appearing only in comments or logs. The AC was not updated to match. "未出现该字符串" means the string is entirely absent from the test file — a completely different condition from "present only in a comment". A developer writing the CI assertion from the AC alone would implement a simple `grep foo:bar test_file.go` check. That check would pass if `foo:bar` appears in a comment (`// foo:bar is tested elsewhere`), satisfying the AC while violating the spec's intent. The AC must be independently executable — it cannot rely on the reader cross-referencing the spec to understand the extraction context.

**What must improve**: Update Story 5 AC first block to: `Given codes.go 中定义了权限码 foo:bar，但该字符串未作为 permCodes 参数或测试矩阵值出现在测试文件中（注释或日志中的出现不计）` / `Then 构建失败，输出 missing test coverage for: foo:bar`. This aligns the AC with the spec's extraction rule.

---

### Attack 2: Background — 12/53 coverage gap is unacknowledged

**Where**: 需求背景: `路由层共有 53 处 perm() 绑定，其中任何一处权限码误写均无法被现有测试发现`. 需求目标: `路由层权限绑定覆盖 | 12 个代表性端点 100% 有单元测试 | 覆盖 pm/member 权限差异最大的端点`.

**Why it's weak**: The background asserts that all 53 perm() bindings are at risk of undetected misconfiguration. The goal then commits to covering only 12. The selection rationale ("pm/member 权限差异最大的端点") is stated, but the document never acknowledges that 41 endpoints remain untested at the route layer after this feature ships. A reader cannot tell whether the remaining 41 are considered low-risk, covered by other means, or simply deferred. This is a logical consistency failure: the problem statement implies a scope of 53, the solution delivers 12, and the gap is invisible. If a reviewer asks "what about the other 41?", the PRD has no answer.

**What must improve**: Add a sentence to the 需求目标 table's 说明 column or a footnote: e.g., `其余 41 个端点权限差异较小（pm/member 均有或均无），由 I-A 集成矩阵间接覆盖，不单独补充单元测试`. This closes the logical gap and makes the coverage decision explicit.

---

### Attack 3: Functional Specs — I-D step 1 silently excludes non-const permission codes

**Where**: I-D 断言逻辑 step 1: `从 codes.go 提取所有 const 权限码（如 main_item:archive）`.

**Why it's weak**: The extraction rule is scoped to `const` declarations only. In Go, a permission code could be defined as a `var`, a struct field, or a string literal returned by a function — none of which would be captured by a const-only extractor. The spec provides no fallback and does not state that all permission codes in this codebase are guaranteed to be constants. A developer implementing I-D from this spec would write a `const`-only parser that silently misses any non-const codes, producing a false-green CI result. Even if the current codebase happens to use only constants, the spec should either (a) assert this as a constraint ("所有权限码必须以 Go const 定义，否则 CI 断言不保证完整性") or (b) broaden the extraction rule to cover all string assignments in codes.go.

**What must improve**: Either add an explicit constraint: `注：codes.go 中所有权限码必须以 const 定义；使用 var 或字符串字面量定义的权限码不在断言覆盖范围内，视为规范违反` — or broaden step 1 to: `从 codes.go 提取所有权限码字符串值（包括 const 和 var 声明）`.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: I-D "出现" underspecified | ✅ | I-D step 2 now reads: `从测试文件中提取所有作为权限码参数传入的字符串（即出现在 permCodes 注入或集成测试矩阵中的权限码值，不包括注释或日志中的字符串）` — loophole closed in spec |
| Attack 2: Story 3 AC "无需重新登录" inconsistent with spec's "无缓存" | ✅ | Story 3 AC third block now reads: `When 用户使用同一 token 立即调用 POST /main-items（不重新登录，不重新获取 token）` / `Then 返回 200（权限从 DB 实时读取，无缓存层介入）` — aligned with spec |
| Attack 3: 人员 section lacks team context | ✅ | 人员 section now reads: `开发者：后端工程师，负责 RBAC 相关功能迭代；当前依赖人工检查权限码覆盖，commit 3200bdc 已证明该方式不可靠` and `代码审查者：PR 审查者，当前无工具辅助验证权限测试完整性，依赖作者自述；commit 3200bdc 的漏测正是在无工具支撑的审查中被遗漏` — exact context added |

---

## Verdict

- **Score**: 96/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: All three iter-4 attacks resolved. Three residual issues remain: (1) Story 5 AC tests a weaker condition than the spec's I-D step 2 — highest priority, creates a false-green CI risk; (2) 12/53 coverage gap unacknowledged — logical consistency hole; (3) I-D step 1 const-only assumption is implicit and unguarded.
