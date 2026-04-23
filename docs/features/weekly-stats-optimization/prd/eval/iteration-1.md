---
date: "2026-04-23"
doc_dir: "docs/features/weekly-stats-optimization/prd/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

**Score: 89/100** (target: N/A)

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
│    Main path complete        │   7/7    │          │            │
│    Decision + error branches │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  16      │  20      │ ⚠️          │
│    Tables complete           │   7/7    │          │            │
│    Field descriptions clear  │   6/7    │          │            │
│    Validation rules explicit │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  19      │  20      │ ✅          │
│    Coverage per user type    │   6/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  17      │  20      │ ⚠️          │
│    In-scope concrete         │   7/7    │          │            │
│    Out-of-scope explicit     │   7/7    │          │            │
│    Consistent with specs     │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  89      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md § 业务流程图 | FetchData 节点无失败分支，违反"至少一个错误/异常分支"要求 | -3 pts (Flow Diagrams: decision+error) |
| prd-ui-functions.md § UI Function 1 & 2 States 表 | 仅定义加载中/已加载/空数据，缺少接口错误状态（error state） | -2 pts (Functional Specs: validation rules) |
| prd-spec.md § 性能需求 | "响应时间无额外增加" — 无量化基准，属于无数字的模糊表述 | -1 pt (Functional Specs: validation rules) |
| prd-spec.md § 5.4 关联性需求改动 | `isOverdue` 函数签名变更描述于功能描述，但未出现在 In Scope 清单中 | -3 pts (Scope Clarity: consistent with specs) |
| prd-user-stories.md § Story 4 | "键盘用户" 作为 As a 角色，在需求背景的人员章节中未定义 | -1 pt (User Stories: coverage per user type) |
| prd-spec.md § 5.4 关联性需求改动 | `isOverdue(item, referenceDate)` 变更说明未指定新默认值或 weekEnd 为 undefined 时的行为 | -1 pt (Functional Specs: field descriptions clear) |

---

## Attack Points

### Attack 1: Flow Diagrams — 无 API 错误分支

**Where**: `prd-spec.md` 业务流程图，`FetchData[后端获取子事项 + 进展记录]` 节点后直接进入 `ComputeActive`，无任何失败路径。

**Why it's weak**: 流程图的判断节点（ComputeActive、ClassifyStatus、CheckOverdue、UserHover）全部是业务逻辑分支，没有一个是异常/错误分支。当接口请求失败时（网络超时、权限错误、服务端 500），流程图完全没有描述系统行为。rubric 明确要求"at least one error/exception branch"，当前图不满足。

**What must improve**: 在 FetchData 后增加失败分支，描述接口错误时的系统行为（例如：显示错误提示、保留上次数据、或清空统计栏），并在流程图中以菱形判断节点体现。

---

### Attack 2: Functional Specs — UI 错误状态缺失 + 性能指标模糊

**Where 1**: `prd-ui-functions.md` UI Function 1 States 表：`| 加载中 | ... | 已加载 | ... | 空数据 | ...`，三行，无 error 行。UI Function 2 States 表同样只有隐藏/显示两行。

**Why it's weak**: 接口失败是真实场景。当 `/views/weekly` 返回错误时，统计栏应显示什么？当前 States 表完全没有定义，开发者无法实现，测试者无法验收。这是一个可执行性漏洞。

**Where 2**: `prd-spec.md` 性能需求："响应时间无额外增加" — 没有基准值，没有 P99 目标，无法验收。

**What must improve**: UI Functions States 表各增加一行 error 状态，定义触发条件（接口返回非 2xx）和展示行为（如显示"-"或错误图标）。性能需求改为可量化表述，例如"统计计算不使新增接口 P99 响应时间超过现有基线 +50ms"。

---

### Attack 3: Scope Clarity — `isOverdue` 函数变更游离于 Scope 之外

**Where**: `prd-spec.md` § 5.4 关联性需求改动，第 4 行："`lib/status.ts` 中 `isOverdue(item, referenceDate)` 将 `referenceDate` 默认值从 `today` 改为调用方传入的 `weekEnd`"。但 In Scope 清单中无对应条目。

**Why it's weak**: 这是一个函数签名变更，影响所有调用方。它出现在功能描述中，却不在 Scope 清单里，造成 Scope 与功能描述不一致（rubric 扣 -3 pts/conflict）。此外，变更说明本身也不完整：未说明 weekEnd 为 undefined 时的行为，也未说明是否有其他调用方受影响。

**What must improve**: 在 In Scope 清单中增加一条：`前端：isOverdue 函数签名更新（referenceDate 改为必传参数）`。同时在 5.4 中补充：weekEnd 为 undefined 时的 fallback 行为，以及是否需要更新所有现有调用方。

---

## Previous Issues Check

<!-- Iteration 1 — no previous report -->

N/A

---

## Verdict

- **Score**: 89/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: First iteration complete. Three attack points identified for revision.
