---
date: "2026-04-28"
doc_dir: "docs/features/api-permission-test-coverage/prd/"
iteration: "4"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 4

**Score: 95/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  19      │  20      │ ✅          │
│    Background three elements │   6/7    │          │            │
│    Goals quantified          │   7/7    │          │            │
│    Logical consistency       │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  20      │  20      │ ✅          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   7/7    │          │            │
│    Decision + error branches │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  18      │  20      │ ✅          │
│    Tables complete           │   7/7    │          │            │
│    Field descriptions clear  │   6/7    │          │            │
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
│ TOTAL                        │  95      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| 人员 section | "开发者" and "代码审查者" stated without team context or current workflow pain — persists from iter 1 | -1 pt (vague, no context) |
| I-D 断言逻辑 step 2 | "从测试文件中提取所有出现的权限码字符串" — "出现" is underspecified; a string in a comment satisfies the check but provides no test coverage | -1 pt (field description unclear) |
| Story 3 AC | "无需重新登录" is ambiguous and inconsistent with spec's "无缓存" framing — same JWT? same session? does the permission check bypass cache? | -1 pt (AC not independently executable) |
| Out of Scope | Items listed without "deferred" vs "permanently excluded" distinction — persists from iter 2 | -1 pt (out-of-scope not fully explicit) |
| Story 3 AC vs I-B spec | Spec says "权限变更即时生效，无缓存"; AC says "无需重新登录" — two different test mechanisms implied for the same scenario | -1 pt (inconsistency between sections) |

---

## Attack Points

### Attack 1: Functional Specs — I-D assertion "出现" is underspecified and exploitable

**Where**: I-D 断言逻辑 step 2: `从测试文件中提取所有出现的权限码字符串`

**Why it's weak**: "出现" (appear/occur) means any string match — a permission code in a comment (`// main_item:archive is tested elsewhere`), a variable name, or a log message would satisfy the check. The assertion is supposed to enforce that every permission code has actual test coverage, but a naive string search cannot distinguish between a code appearing as a test parameter and a code appearing in dead code or documentation. A developer implementing I-D from this spec could write a `grep`-based check that passes trivially without providing any real coverage guarantee. The spec should state the extraction context: "出现在 `permCodes` 参数传入或测试矩阵表格中" or "作为 `c.Set(\"permCodes\", ...)` 的参数值出现".

**What must improve**: Replace step 2 with: `从测试文件中提取所有作为权限码参数传入的字符串（即出现在 permCodes 注入或集成测试矩阵中的权限码值）`. This closes the loophole and makes the assertion implementable without ambiguity.

---

### Attack 2: User Stories — Story 3 AC "无需重新登录" inconsistent with spec's "无缓存"

**Where**: Story 3 AC, third Given/When/Then block: `When 用户立即调用 POST /main-items（无需重新登录）` / `Then 返回 200（权限变更即时生效）`. Spec I-B step 5: `验证：POST /main-items → 200（权限变更即时生效，无缓存）`.

**Why it's weak**: The spec was correctly updated in a prior iteration to say "无缓存", which specifies the mechanism (no permission caching layer). The AC was not updated and still says "无需重新登录", which specifies a different constraint (same session/token). These are not equivalent: a system could satisfy "无需重新登录" by reusing the same JWT while still serving stale cached permissions, and would fail the spec's "无缓存" requirement. A developer writing the integration test from the AC alone would construct a test that passes even if a permission cache bug exists. The AC must be independently executable — it cannot rely on the reader cross-referencing the spec to understand what is actually being tested.

**What must improve**: Update Story 3 AC third block to: `When 用户使用同一 token 立即调用 POST /main-items（不重新登录，不重新获取 token）` / `Then 返回 200（权限从 DB 实时读取，无缓存层介入）`. This aligns with the spec's "无缓存" framing and makes the test mechanism unambiguous.

---

### Attack 3: Background — 人员 section lacks team context for the fourth consecutive iteration

**Where**: 人员 section: `开发者：编写和运行测试，在 CI 中获得权限回归保护` / `代码审查者：通过 PR 审查权限测试覆盖是否完整`

**Why it's weak**: This has been flagged in every iteration and remains unchanged. "开发者" and "代码审查者" are roles, not users — they carry no information about team size, current workflow, or why the existing process is insufficient. The background section already establishes the concrete pain (commit `3200bdc` introduced 2 permission codes with no test coverage; bizkey-unification will add more), but the 人员 section does not connect these users to that pain. A reader cannot tell whether "开发者" means one person or a team of ten, whether "代码审查者" is a separate role or the same developer, or what their current review process looks like. This matters because the scope of I-D (CI assertion) is justified by the claim that manual review is insufficient — but the 人员 section never establishes that.

**What must improve**: Add one sentence of context per user: e.g., `开发者：后端工程师，负责 RBAC 相关功能迭代；当前依赖人工检查权限码覆盖，commit 3200bdc 已证明该方式不可靠` and `代码审查者：PR 审查者，当前无工具辅助验证权限测试完整性，依赖作者自述`.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Scope I-A not updated to match fixed Goal 2 | ✅ | Scope now reads: `I-A: 预设角色矩阵集成测试 — superadmin/pm/member × 5 个代表性端点（main_item/team/progress/item_pool/report 各一个）` — count and domain breakdown now present |
| Attack 2: CI location ambiguous (".github/workflows 或 Makefile") | ✅ | 关联性需求改动 row 3 now reads: `CI \| .github/workflows/test.yml \| 在 go test 步骤后新增 permission-coverage 步骤` — single concrete location |
| Attack 3: Story 2 AC uses abbreviated endpoint paths | ✅ | Story 2 AC now uses full paths: `POST /teams/:teamId/main-items/:id/archive` and `POST /teams/:teamId/members` |

---

## Verdict

- **Score**: 95/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: All three iter-3 attacks resolved. Four residual issues remain: (1) I-D "出现" underspecified — highest priority, functional correctness risk; (2) Story 3 AC "无需重新登录" inconsistent with spec's "无缓存"; (3) Out-of-scope lacks deferred/permanent distinction; (4) 人员 section lacks team context (persists from iter 1).
