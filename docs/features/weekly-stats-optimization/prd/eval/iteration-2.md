---
date: "2026-04-23"
doc_dir: "docs/features/weekly-stats-optimization/prd/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 2

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
│ 2. Flow Diagrams             │  20      │  20      │ ✅          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   7/7    │          │            │
│    Decision + error branches │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  19      │  20      │ ⚠️          │
│    Tables complete           │   7/7    │          │            │
│    Field descriptions clear  │   6/7    │          │            │
│    Validation rules explicit │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  19      │  20      │ ⚠️          │
│    Coverage per user type    │   6/7    │          │            │
│    Format correct            │   7/7    │          │            │
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
| prd-ui-functions.md § UI Function 2 Data Requirements + prd-spec.md § 5.1 tooltip text | "本周活跃" tooltip 文本未明确说明"含本周新完成"，而需求背景明确指出这是核心混淆点 | -1 pt (Functional Specs: field descriptions clear) |
| prd-user-stories.md § Story 4 | "使用键盘导航的用户" 作为 As a 角色，在 prd-spec.md 需求背景人员章节中仍未定义 | -1 pt (User Stories: coverage per user type) |
| prd-spec.md § In Scope | 错误状态处理（卡片显示"-"并提示"加载失败，请刷新"）在流程图和 prd-ui-functions.md States 表中均有描述，但未出现在 In Scope 清单中 | -1 pt (Scope Clarity: consistent with specs) |

---

## Attack Points

### Attack 1: Functional Specs — 核心 tooltip 未解决已知混淆

**Where**: `prd-ui-functions.md` UI Function 2 Tooltip 文本定义表，"本周活跃"行：`本周有进展记录，或计划周期与本周重叠的子事项总数`；同文本出现于 `prd-spec.md` § 5.1 第 1 行 Tooltip 说明文本列。

**Why it's weak**: 需求背景明确点名了具体混淆："典型混淆：`justCompleted`（本周新完成）的事项同时被计入活跃总数，该行为未文档化"。这是整个 feature 存在的核心理由之一。然而当前 tooltip 文本只描述了"活跃"的判定条件（有进展记录 OR 日期重叠），完全没有提及"含本周新完成"。用户读完 tooltip 后仍然无法理解为什么新完成的事项会出现在活跃总数里——tooltip 解决了"活跃是什么"，但没有解决"为什么新完成也算活跃"这个已知痛点。

**What must improve**: 将"本周活跃"的 tooltip 文本改为明确包含新完成的说明，例如："本周有进展记录，或计划周期与本周重叠的子事项总数（含本周新完成）"。同时在 prd-spec.md § 5.1 和 prd-ui-functions.md 中保持一致。

---

### Attack 2: User Stories — 键盘用户角色未在人员章节定义

**Where**: `prd-user-stories.md` Story 4：`As a 使用键盘导航的用户`；`prd-spec.md` 需求背景 § 用户是谁：仅列出"PM / 项目负责人"和"团队成员"两类角色。

**Why it's weak**: rubric 要求"every user type from the background section has at least one story"，反向同样成立——story 中出现的角色应在背景人员章节中定义。"键盘用户"是一个独立的无障碍访问场景，与"团队成员"不完全重叠（键盘用户可以是 PM 也可以是团队成员）。当前文档在背景中没有为这类用户建立上下文，导致 Story 4 的 So that 理由（"满足无障碍访问需求"）缺乏背景支撑。此问题在 Iteration 1 中已被标记，本次未修复。

**What must improve**: 在 `prd-spec.md` 需求背景人员章节新增第三类用户："使用键盘或辅助技术导航的用户：需要通过 Tab 键和屏幕阅读器访问统计信息，满足无障碍访问要求"。

---

### Attack 3: Scope Clarity — 错误状态处理未列入 In Scope

**Where**: `prd-spec.md` § In Scope 清单（8 条）中无任何条目涉及错误状态处理；但 `prd-spec.md` 业务流程图有 `FetchError[统计栏所有卡片显示"-"并提示加载失败]` 节点，`prd-ui-functions.md` UI Function 1 States 表有 `| 错误 | 所有卡片数字显示 "-"，统计栏顶部显示"加载失败，请刷新" | 接口返回非 2xx 或请求超时 |` 一行。

**Why it's weak**: 错误状态处理是一个需要前端实现的具体功能点（错误检测逻辑、UI 展示、文案），它出现在功能描述和流程图中，却不在 In Scope 清单里。这造成 Scope 与功能描述不一致——开发者看 Scope 清单时不会意识到需要实现错误状态，但看功能描述时又会发现有明确要求。

**What must improve**: 在 In Scope 清单中新增一条：`前端：StatsBar 错误状态处理（接口失败时所有卡片显示"-"并展示"加载失败，请刷新"提示）`。

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Flow Diagrams — FetchData 节点无失败分支 | ✅ | 流程图新增 `FetchOK{请求成功?}` 判断节点，`否` 分支指向 `FetchError[统计栏所有卡片显示"-"并提示加载失败]` |
| Attack 2: Functional Specs — UI 错误状态缺失 + 性能指标模糊 | ✅ | prd-ui-functions.md States 表新增错误行；性能需求改为"不得使该接口 P99 响应时间超出现有基线 +50ms" |
| Attack 3: Scope Clarity — isOverdue 函数变更游离于 Scope 之外 | ✅ | In Scope 新增"前端：`isOverdue` 函数签名更新（`referenceDate` 改为必传参数，移除 `today` 默认值）及所有现有调用方同步更新"；5.4 补充了 weekEnd 为 undefined 时须抛出错误的行为说明 |

---

## Verdict

- **Score**: 97/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Iteration 2 complete. All three Iteration 1 attacks resolved. Three minor issues remain (tooltip text gap, missing persona definition, scope/spec inconsistency on error state). Document is production-ready; further iteration optional.
