---
date: "2026-04-28"
doc_dir: "docs/features/api-permission-test-coverage/prd/"
iteration: "3"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 3

**Score: 93/100** (target: N/A)

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
│ 5. Scope Clarity             │  17      │  20      │ ✅          │
│    In-scope concrete         │   6/7    │          │            │
│    Out-of-scope explicit     │   6/7    │          │            │
│    Consistent with specs     │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  93      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| 人员 section | "开发者" and "代码审查者" stated without team context or current workflow pain — persists from iter 1 | -1 pt (vague, no context) |
| 关联性需求改动 row 3 | CI location stated as ".github/workflows 或 Makefile" — ambiguous; developer cannot determine where to implement — persists from iter 2 | -1 pt (field description unclear) |
| I-B step 5 | "无需重新登录" is technically ambiguous — same JWT? same session? same HTTP connection? — persists from iter 2 | -1 pt (validation rule not explicit) |
| Story 2 AC | Uses abbreviated endpoint paths "POST /archive" and "POST /members" instead of full paths — persists from iter 2 | -1 pt (AC precision) |
| Scope I-A | "superadmin/pm/member × 代表性端点" — scope item was not updated when Goal 2 was fixed; goal now says "5 个代表性端点" but scope still omits the count | -1 pt (in-scope not concrete) |
| Out of Scope | Items listed without "deferred" vs "permanently excluded" distinction — persists from iter 2 | -1 pt (out-of-scope not fully explicit) |
| Scope I-A vs Goal 2 | Scope says "代表性端点" (no count); Goal 2 says "5 个代表性端点"; functional spec footnote says "5 个" — three sections use different levels of specificity for the same concept | -1 pt (scope vs spec inconsistency) |

---

## Attack Points

### Attack 1: Scope Clarity — I-A scope item not updated to match the fixed Goal 2

**Where**: Scope section: `I-A: 预设角色矩阵集成测试 — superadmin/pm/member × 代表性端点`. Goal 2 (now fixed): `superadmin/pm/member × 5 个代表性端点（覆盖 main_item/team/progress/item_pool/report 各一个）全覆盖`. Functional spec footnote: `选取标准：从 U1 的 12 个端点中，选取 pm/member 权限差异最大的端点各一个，覆盖每个权限域（main_item、team、progress、item_pool、report），共 5 个`.

**Why it's weak**: Goal 2 was correctly updated in iter 3 to include the count and domain breakdown. The functional spec footnote was added in iter 2. But the Scope section — the first place a reader looks to understand deliverables — was never updated. A reader scanning only the Scope section still sees "代表性端点" with no count, and must cross-reference Goal 2 and the functional spec footnote to understand what is actually being built. The three sections now use three different levels of specificity for the same concept, creating a consistency gap that a code reviewer would flag.

**What must improve**: Update the I-A scope item to: `I-A: 预设角色矩阵集成测试 — superadmin/pm/member × 5 个代表性端点（main_item/team/progress/item_pool/report 各一个）`. This makes the scope self-contained and consistent with Goal 2.

---

### Attack 2: Functional Specs — CI location ambiguity persists through three iterations

**Where**: 关联性需求改动 table, row 3: `CI | .github/workflows 或 Makefile | 新增权限码覆盖率断言步骤 | 解析 codes.go 与测试文件，缺失覆盖则 CI 失败`.

**Why it's weak**: This was flagged in iter 2 and remains unchanged. ".github/workflows 或 Makefile" are two fundamentally different implementation locations with different semantics: a GitHub Actions workflow step runs in CI only and is invisible locally; a Makefile target can be run locally and invoked from CI. The choice affects how developers run the check during development, whether it appears in `make test`, and how the CI failure is surfaced. A developer assigned I-D cannot make this decision from the PRD — they must either guess or interrupt the PM for clarification. The "或" signals the author has not decided, not that either is acceptable.

**What must improve**: Pick one and state it. For example: `CI | .github/workflows/test.yml | 在 go test 步骤后新增 permission-coverage 步骤` or `CI | Makefile (perm-coverage target) + .github/workflows 调用该 target`. If the decision is genuinely deferred, add a note explaining what criteria will drive the choice.

---

### Attack 3: User Stories — Story 2 AC uses abbreviated endpoint paths for the third consecutive iteration

**Where**: Story 2, AC 1: `When 各用户调用 POST /archive（需要 main_item:archive，member 无此权限）`. AC 2: `When 各用户调用 POST /members（需要 team:invite，member 无此权限）`.

**Why it's weak**: The full paths are `POST /teams/:teamId/main-items/:id/archive` and `POST /teams/:teamId/members`. The abbreviated forms "/archive" and "/members" are ambiguous — "/members" could refer to multiple endpoints in a REST API. The U1 functional spec table uses full paths throughout; the AC should match. More critically, a developer writing the integration test from this AC cannot construct the HTTP request without looking up the full path in the U1 table. ACs should be independently executable without cross-referencing other sections.

**What must improve**: Replace abbreviated paths with full paths in both ACs: `POST /teams/:teamId/main-items/:id/archive` and `POST /teams/:teamId/members`. This is a one-line fix that has been deferred across three iterations.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: I-D missing from flow diagram; I-C has no decision node | ✅ | I-D path added: `ID_start → ID_check{测试文件中均出现?} → ID_pass/ID_fail`. I-C now has three decision nodes: `ID3{响应码?}`, `IF3{响应码?}`, `IH3{响应码?}` with failure branches for each boundary scenario |
| Attack 2: I-A endpoint selection criteria unstated | ✅ | Footnote added to I-A table: "选取标准：从 U1 的 12 个端点中，选取 pm/member 权限差异最大的端点各一个，覆盖每个权限域（main_item、team、progress、item_pool、report），共 5 个" |
| Attack 3: "代表性端点" unquantified in Goal 2 | ✅ | Goal 2 now reads: "superadmin/pm/member × 5 个代表性端点（覆盖 main_item/team/progress/item_pool/report 各一个）全覆盖" |

---

## Verdict

- **Score**: 93/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: All three iteration-2 attacks resolved. Three residual issues remain, all persisting from prior iterations: (1) Scope I-A not updated to match the now-fixed Goal 2, (2) CI location ambiguous in 关联性需求改动, (3) Story 2 AC uses abbreviated endpoint paths.
