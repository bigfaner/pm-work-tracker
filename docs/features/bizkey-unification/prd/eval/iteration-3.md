---
date: "2026-04-28"
doc_dir: "docs/features/bizkey-unification/prd/"
iteration: "3"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 3

**Score: 97/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  20      │  20      │ ✅          │
│    Background three elements │   7/7    │          │            │
│    Goals quantified          │   7/7    │          │            │
│    Logical consistency       │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  19      │  20      │ ⚠️          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   7/7    │          │            │
│    Decision + error branches │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  20      │  20      │ ✅          │
│    Tables complete           │   7/7    │          │            │
│    Field descriptions clear  │   7/7    │          │            │
│    Validation rules explicit │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  19      │  20      │ ⚠️          │
│    Coverage per user type    │   7/7    │          │            │
│    Format correct            │   6/7    │          │            │
│    AC per story              │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  19      │  20      │ ⚠️          │
│    In-scope concrete         │   7/7    │          │            │
│    Out-of-scope explicit     │   7/7    │          │            │
│    Consistent with specs     │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  97      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md §流程说明 第二张图 | `G2[继续执行 InviteMember]` → `H2([成员邀请成功])` 无错误分支 — 邀请操作本身失败时（如成员已存在、数据库错误）无路径覆盖 | -1 pt (Decision + error branches) |
| prd-user-stories.md Story 1 | 用户类型标签 "系统用户（团队成员）" 与 §需求背景 人员 中定义的 "系统用户（间接）" 不一致 | -1 pt (Format correct) |
| prd-spec.md §Scope vs §功能描述 | In Scope 列出 "单元测试和集成测试（20 个文件）"，但 §功能描述 表格中无任何测试文件行，无法从功能描述推断测试变更范围 | -1 pt (Consistent with specs) |

---

## Attack Points

### Attack 1: Flow Diagrams — InviteMember 第二张图缺少邀请操作本身的错误分支

**Where**: `G2[继续执行 InviteMember] --> H2([成员邀请成功])` — 第二张流程图在角色检查通过后直接跳到成功节点，无任何错误出口。

**Why it's weak**: 第一张图在每个关键节点（解析失败 → 400、团队不存在 → 404、Service 层 FindByBizKey 失败 → 500）都有错误分支，但第二张图在 `G2` 之后完全没有错误路径。`InviteMember` 实际执行时可能失败（成员已存在、数据库写入失败、并发冲突），这些场景在图中完全不可见。与第一张图的严谨性相比，第二张图的错误覆盖明显不完整。

**What must improve**: 在 `G2` 后添加错误分支，例如 `G2 -->|邀请失败| I2[返回对应错误]`，至少覆盖"成员已存在"这一最常见的业务错误场景。

---

### Attack 2: User Stories — Story 1 用户类型标签与背景章节不一致

**Where**: prd-user-stories.md Story 1: `**As a** 系统用户（团队成员）`。prd-spec.md §需求背景 人员: `**系统用户（间接）**：进度记录的 team_key 字段将存储正确的雪花 ID，查询结果不再静默错误`。

**Why it's weak**: 背景章节定义的用户类型是"系统用户（间接）"，强调该用户是间接受益者（不直接感知代码变更）。Story 1 将其改写为"系统用户（团队成员）"，引入了一个背景章节中未定义的标签。读者无法确认这两个标签是否指同一类用户，也无法判断"团队成员"是否是"系统用户（间接）"的子集。这是跨文件的术语不一致，违反了 PRD 内部一致性要求。

**What must improve**: 将 Story 1 的用户类型改为 `系统用户（间接）` 以与背景章节保持一致，或在背景章节中将"系统用户（间接）"更新为"系统用户（团队成员）"并说明两者关系。

---

### Attack 3: Scope Clarity — 20 个测试文件在 Scope 中但功能描述中无对应条目

**Where**: §Scope In Scope: `单元测试和集成测试（20 个文件：10 个 handler 测试、7 个 service 测试、team_scope_test.go、views_reports_test.go、helpers.go）`。§功能描述 表格：21 行，全部为 handler 和 service 源文件，无任何测试文件行。

**Why it's weak**: Scope 明确承诺 20 个测试文件将被修改，但功能描述对这些文件的变更内容只字未提。开发者无法从功能描述中得知：哪些测试需要新增用例、哪些只是机械性的签名更新、`helpers.go` 的变更是什么。与 handler 文件的处理方式对比（handler 文件在 iteration 2 后补充了 7 行），测试文件的处理方式明显不一致。

**What must improve**: 在功能描述中补充一行或一段说明测试文件的变更性质，例如："所有测试文件的变更为机械性签名更新（`uint teamID` → `int64 teamBizKey`），无新增测试用例，`helpers.go` 中的测试辅助函数同步更新参数类型"。

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Service-level FindByBizKey 无错误分支 + isPMRole 流程图缺失 | ✅ | 节点 J 新增 `J -->|记录不存在| M[返回 500 ErrTeamNotFound]`；新增第二张 InviteMember/isPMRole 流程图，覆盖 `ErrRoleNotFound` 和 `ErrCannotAssignPMRole` 分支 |
| Attack 2: "团队 PM" 未在背景章节定义 | ✅ | §需求背景 人员 新增第三条："**团队 PM**：角色权限判断依赖 bizKey 正确性，修复后 `isPMRole` 判断不再因 ID 类型混用而误判，邀请成员时 PM 角色限制得到正确执行" |
| Attack 3: 7 个 handler 文件在 Scope 中但功能描述无对应条目 | ✅ | §功能描述 表格新增第 15–21 行，覆盖全部 7 个 handler 文件，每行说明变更性质为"机械性调用点更新，无业务逻辑变更" |

---

## Verdict

- **Score**: 97/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: 从 iteration 2 的 92 分提升至 97 分。三处遗留问题全部修复。剩余 3 分差距来自三处细节：InviteMember 图缺少邀请失败分支、Story 1 用户标签与背景章节不一致、测试文件在 Scope 中但功能描述无对应说明。均为小幅修改，无需结构性重写。
