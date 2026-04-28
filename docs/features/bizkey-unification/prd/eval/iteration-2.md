---
date: "2026-04-28"
doc_dir: "docs/features/bizkey-unification/prd/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 2

**Score: 92/100** (target: N/A)

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
│ 2. Flow Diagrams             │  17      │  20      │ ⚠️          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   6/7    │          │            │
│    Decision + error branches │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  19      │  20      │ ✅          │
│    Tables complete           │   6/7    │          │            │
│    Field descriptions clear  │   7/7    │          │            │
│    Validation rules explicit │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  18      │  20      │ ⚠️          │
│    Coverage per user type    │   5/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  18      │  20      │ ⚠️          │
│    In-scope concrete         │   7/7    │          │            │
│    Out-of-scope explicit     │   7/7    │          │            │
│    Consistent with specs     │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  92      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md §流程说明 | Node J (`Service 内部调用 FindByBizKey`) has no error branch — what happens when the record is not found at the service level? | -1 pt (Main path complete) |
| prd-spec.md §流程说明 | No diagram for the role permission check path (Story 2 / `isPMRole`); the flow only covers the middleware→handler→service→repo happy path | -2 pts (Decision + error branches) |
| prd-spec.md §功能描述 | Scope lists 7 handler files explicitly (`item_pool_handler.go`, `main_item_handler.go`, `progress_handler.go`, `report_handler.go`, `sub_item_handler.go`, `team_handler.go`, `view_handler.go`) but the functional spec table has zero rows for any handler file | -1 pt (Tables complete) |
| prd-spec.md §Scope vs §功能描述 | 7 handler files declared in-scope; functional spec documents none of them individually — row 1 says "所有 handler 调用方需同步更新" but does not enumerate what changes per handler | -2 pts (Scope consistent with specs) |
| prd-user-stories.md Story 2 | "团队 PM" introduced as user type but §需求背景 人员 defines only "后端开发者" and "系统用户（间接）" — "团队 PM" is still absent from the background section | -2 pts (Coverage per user type, inconsistency between sections) |

---

## Attack Points

### Attack 1: Flow Diagrams — Service-level FindByBizKey has no error branch, isPMRole flow absent

**Where**: Node J in the flowchart: `I --> J[Service 内部调用 FindByBizKey]` → `J --> K[Repository 内部使用 uint ID 做 FK 关联]` → `K --> L([返回结果])`. No branch exits J on failure. Also, Story 2 describes a role permission check (`isPMRole`) that rejects PM role assignment — this path has no diagram at all.

**Why it's weak**: The middleware-level `FindByBizKey` (node D) correctly shows a 404 branch, but the service-level `FindByBizKey` (node J) is a dead end — no error path. If the team record disappears between middleware and service execution (race condition, soft-delete), the diagram gives no guidance. The `isPMRole` path is the only user-visible business rule in the entire PRD (Story 2 AC: "系统正确识别该角色为 PM 角色并返回 `ErrCannotAssignPMRole` 错误") yet it has no flow representation whatsoever.

**What must improve**: Add an error branch from node J (e.g., `J -->|记录不存在| M[返回 500 / ErrTeamNotFound]`). Add a second flowchart covering the `InviteMember` → `isPMRole` decision path, showing the `ErrCannotAssignPMRole` branch.

---

### Attack 2: User Stories — "团队 PM" user type undefined in background section

**Where**: prd-user-stories.md Story 2: `**As a** 团队 PM`. §需求背景 人员 section: "**后端开发者**：修改后的接口签名由编译器强制约束..." and "**系统用户（间接）**：进度记录的 `team_key` 字段将存储正确的雪花 ID...". "团队 PM" appears nowhere in the background section.

**Why it's weak**: The rubric requires "one story per target user" where target users are defined in the background section. "团队 PM" is a third user type introduced only in the user stories file with no corresponding definition in §需求背景 人员. A reader of the spec alone would not know that team PMs are affected. This is a direct inconsistency between sections (-3 pts per the deduction rules), softened here because the story itself is coherent — the inconsistency is a documentation gap, not a logical contradiction.

**What must improve**: Add "团队 PM" to §需求背景 人员 with a one-line description of how they are affected (e.g., "团队 PM：角色权限判断依赖 bizKey 正确性，修复后 isPMRole 判断不再因 ID 类型混用而误判").

---

### Attack 3: Scope Clarity — 7 handler files in scope but functionally undocumented

**Where**: §Scope In Scope: "Handler 调用点（7 个文件）：`item_pool_handler.go`、`main_item_handler.go`、`progress_handler.go`、`report_handler.go`、`sub_item_handler.go`、`team_handler.go`、`view_handler.go` — 从 context/请求中传递 bizKey，而非解析后的 uint ID". §功能描述 table: 14 rows, zero rows for any handler file. Row 1 says "所有 handler 调用方需同步更新" but names no specific handler or method.

**Why it's weak**: The scope makes a concrete promise (7 named files will change) that the functional spec does not fulfill. A developer reading only §功能描述 cannot determine which methods in `report_handler.go` change or what the before/after looks like. "所有 handler 调用方需同步更新" is a blanket statement, not a functional spec. If a handler has both a `teamID uint` path and a `teamBizKey int64` path, the spec gives no guidance on which to change.

**What must improve**: Add one row per handler file to the functional spec table, or add a note explicitly stating "handler changes are mechanical call-site updates — see service rows 2–14 for the corresponding signature changes, no handler-specific logic changes". Either approach resolves the scope-spec gap.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Functional Specs table covers only 2 of 8 service files | ✅ | Table expanded to 14 rows; all 8 service files now have explicit rows (rows 2–14); `role_service` explicitly noted as no-change-needed (row 12) |
| Attack 2: Zero validation rules | ✅ | New §5.5 added with 3 explicit rules: decimal integer string, positive int64, must match existing team; failure conditions, HTTP codes, and error descriptions all stated |
| Attack 3: "约 20 个文件" vague language + scope-spec inconsistency | ⚠️ | "约" removed — fixed. Service-level scope-spec inconsistency resolved. Handler-level inconsistency (7 files in scope, 0 rows in spec) persists — partial fix only |

---

## Verdict

- **Score**: 92/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Significant improvement from iteration 1 (80 → 92). Remaining gaps are narrow: add "团队 PM" to background section, add error branch from node J in the flow diagram, and resolve the handler-level scope-spec gap. These are all small edits, not structural rewrites.
