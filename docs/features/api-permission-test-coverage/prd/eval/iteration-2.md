---
date: "2026-04-28"
doc_dir: "docs/features/api-permission-test-coverage/prd/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 2

**Score: 86/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  17      │  20      │ ✅          │
│    Background three elements │   6/7    │          │            │
│    Goals quantified          │   5/7    │          │            │
│    Logical consistency       │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  16      │  20      │ ✅          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   5/7    │          │            │
│    Decision + error branches │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  17      │  20      │ ✅          │
│    Tables complete           │   6/7    │          │            │
│    Field descriptions clear  │   6/7    │          │            │
│    Validation rules explicit │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  19      │  20      │ ✅          │
│    Coverage per user type    │   7/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  17      │  20      │ ✅          │
│    In-scope concrete         │   6/7    │          │            │
│    Out-of-scope explicit     │   6/7    │          │            │
│    Consistent with specs     │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  86      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| 需求目标 Goal 2 | "superadmin/pm/member × 代表性端点全覆盖" — "代表性端点" still unquantified; no count, no list | -2 pts (vague language, persists from iter 1) |
| 人员 section | "开发者" and "代码审查者" stated without team context or current workflow pain | -1 pt (vague, persists from iter 1) |
| 流程图 | I-D scope item added but has zero representation in the Mermaid diagram | -2 pts (main path incomplete) |
| 流程图 I-C path | Linear sequence with no decision node — no branch for "token validation fails vs passes" | -2 pts (missing decision branch) |
| I-A table | Covers 5 of 12 U1 endpoints with no stated selection criteria for why these 5 are "representative" | -1 pt (tables incomplete) |
| I-D 关联性需求改动 | CI location stated as ".github/workflows 或 Makefile" — ambiguous; developer cannot determine where to implement | -1 pt (field description unclear) |
| I-B step 5 | "无需重新登录" is technically ambiguous — does this mean same JWT, same session, or same HTTP connection? | -1 pt (validation rule not explicit) |
| User Stories AC | Stories 1, 2 use abbreviated endpoint paths ("POST /archive", "POST /members") instead of full paths | -1 pt (AC precision) |
| Scope I-A | "superadmin/pm/member × 代表性端点" in scope does not clarify that "代表性端点" is a 5-endpoint subset of U1's 12 | -1 pt (scope vs spec inconsistency) |
| Out of Scope | Items listed without "deferred" vs "permanently excluded" distinction — reader cannot tell if these are future iterations or never | -1 pt (out-of-scope not fully explicit) |

---

## Attack Points

### Attack 1: Flow Diagrams — I-D is a named scope item with zero diagram representation

**Where**: Scope lists `I-D: 权限码覆盖率 CI 断言` as an in-scope deliverable. The Mermaid diagram has four paths (U1, I-A, I-B, I-C) and stops there. I-D does not appear anywhere in the diagram.

**Why it's weak**: The diagram was not updated when I-D was promoted from 监控需求 to a first-class scope item. A reader using the diagram to understand the full test execution flow will have no mental model of how the CI assertion fits in — when it runs, what it reads, what it outputs on failure. The I-C path also has no decision node: it is a flat linear sequence (`IC3 → ID3 → IE3 → IF3`) with no branch for "what happens if the assertion fails." Every other path (U1, I-A, I-B) has a diamond decision node with a failure branch; I-C and I-D do not.

**What must improve**: Add an I-D path to the diagram: `Start → ID[I-D: 读取 codes.go 权限码] → IE{测试文件中均出现?} →|是| IF[CI 断言通过] →|否| IG[CI 失败，输出缺失列表]`. Add a decision node to the I-C path for each of the three boundary scenarios.

---

### Attack 2: Functional Specs — I-A table endpoint selection is unjustified

**Where**: I-A table covers `main_item:archive | team:invite | progress:update | item_pool:review | report:export` — 5 endpoints. The U1 table lists 12 endpoints. The scope says "superadmin/pm/member × 代表性端点全覆盖" without defining what makes an endpoint "representative."

**Why it's weak**: A developer implementing I-A cannot tell whether the 5-endpoint selection is exhaustive for integration purposes or an arbitrary sample. There is no stated selection criterion (e.g., "one endpoint per permission domain" or "endpoints where pm/member diverge"). The 7 omitted endpoints — including `POST /main-items` (both pm and member have access) and `GET /views/weekly` (both have access) — are precisely the ones where a superadmin-bypass test would be most valuable. Without a rationale, a code reviewer cannot verify the selection is correct.

**What must improve**: Add a footnote or column to the I-A table stating the selection criterion. For example: "选取 pm/member 权限差异最大的端点各一个，覆盖每个权限域（main_item, team, progress, item_pool, report）." Alternatively, expand the matrix to all 12 U1 endpoints.

---

### Attack 3: Background & Goals — "代表性端点" persists as unquantified language in Goal 2

**Where**: 需求目标 table, Goal 2: `预设角色矩阵验证 | superadmin/pm/member × 代表性端点全覆盖 | 集成测试断言响应码符合权限矩阵`

**Why it's weak**: This was flagged in iteration 1 and remains unchanged. "代表性端点" appears in both the goal and the scope without ever being defined. Goal 1 is precise ("12 个代表性端点 100% 有单元测试"), but Goal 2 uses the same phrase without anchoring it to a count. The I-A table implicitly defines it as 5 endpoints, but the goal text does not say "5 个代表性端点." This means the goal is unverifiable as written — a CI run covering 3 endpoints could claim to satisfy "代表性端点全覆盖."

**What must improve**: Replace "代表性端点" in Goal 2 with a concrete count: "superadmin/pm/member × 5 个代表性端点（覆盖 main_item/team/progress/item_pool/report 各一个）全覆盖." This makes the goal independently verifiable without cross-referencing the I-A table.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: "非403" is not a valid test assertion | ✅ | I-A table now has concrete 200/403 values per role. I-C table has explicit 403/200/401. Note added: "若 superadmin/pm 返回 404 或 500，视为测试 fixture 数据缺失，不算通过" |
| Attack 2: Tables missing test-execution columns | ✅ | U1 table now has 测试文件, 测试函数, Mock 依赖 columns with specific values for all 12 endpoints |
| Attack 3: CI assertion deliverable orphaned in 监控需求 | ✅ | I-D promoted to named scope item, full functional spec added (§ I-D with 3-step assertion logic), Story 5 added with 2 Given/When/Then ACs |

---

## Verdict

- **Score**: 86/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: All three iteration-1 attacks resolved. Three new issues remain: (1) I-D and I-C missing from flow diagram, (2) I-A endpoint selection criteria unstated, (3) "代表性端点" still unquantified in Goal 2.
