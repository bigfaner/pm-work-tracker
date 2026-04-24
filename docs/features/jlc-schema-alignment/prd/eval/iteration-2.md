---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/prd/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 2

**Score: 98/100** (target: N/A)

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
│ 3. Functional Specs          │  18      │  20      │ ⚠️          │
│    Tables complete           │   7/7    │          │            │
│    Field descriptions clear  │   6/7    │          │            │
│    Validation rules explicit │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  20      │  20      │ ✅          │
│    Coverage per user type    │   7/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  20      │  20      │ ✅          │
│    In-scope concrete         │   7/7    │          │            │
│    Out-of-scope explicit     │   7/7    │          │            │
│    Consistent with specs     │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  98      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| 功能描述 §5.1 — TEXT→VARCHAR 表 | `description` 和 `background` 取 2000，其余 5 个字段取 1000，长度选取依据完全缺失；超限策略仍是条件性兜底（"若...超出上限"）而非明确约束 | -1 pt (field descriptions) |
| 功能描述 §5.1 — 全局变更表 | `completion DECIMAL(5,2)` 无有效范围约束：Go model 保持 `float64`，service 层是否校验 0.00–100.00 未说明；`deleted_flag TINYINT(1)` 允许值（0/1）未显式声明为约束 | -1 pt (validation rules) |

---

## Attack Points

### Attack 1: Functional Specs — VARCHAR 长度选取依据仍然缺失

**Where**: `§5.1 TEXT → VARCHAR 表` — "`description` → `VARCHAR(2000)`"、"`background` → `VARCHAR(2000)`"、"`expected_output` → `VARCHAR(1000)`"、"`achievement`/`blocker`/`lesson` → `VARCHAR(1000)`"

**Why it's weak**: `description` 和 `background` 取 2000，其余 5 个字段取 1000，但文档对这一差异只字未提。DBA 无法判断 2000 是基于实测数据、业务经验还是随意估算。超限策略"若测试环境导入真实数据后任一字段超出上限，将该字段升级为 TEXT 并移入独立 detail 表"是一个条件性兜底，不是约束——它意味着当前长度可能是错的，只是暂时凑合。这在 schema 评审中会被直接打回。

**What must improve**: 在 TEXT→VARCHAR 表中增加"选取依据"列，说明每个长度的来源（如"现有数据最大值 N 字符，取整至 1000"或"业务规格：描述字段不超过 2000 字"）。超限策略应改为明确约束（如"超出上限时 API 返回 400，不静默截断"），而非事后升级方案。

---

### Attack 2: Functional Specs — `completion` 字段无范围校验规则

**Where**: `§5.1 全局变更表` — "`completion REAL` → `completion DECIMAL(5,2) NOT NULL DEFAULT 0.00`"；`§5.2 model 层` — "`Completion` 类型：保持 `float64`，JSON tag 不变`"

**Why it's weak**: `DECIMAL(5,2)` 在数据库层允许 -999.99 到 999.99，但完成度的业务语义是 0–100。文档没有说明 service 层是否校验范围，也没有说明超出范围时的行为（返回 400？截断？panic？）。Go model 保持 `float64` 意味着类型本身不提供任何约束。这是一个 breaking change 中新增的精度字段，却没有任何验收标准能验证"完成度不会写入 150.00"。

**What must improve**: 在 §5.1 或独立"字段约束"小节中，为 `completion` 声明有效范围（`0.00 ≤ completion ≤ 100.00`），并在 Story 1 或 Story 2 的 AC 中增加一条：`Given completion=150.00 的写入请求，When 调用 service 层，Then 返回校验错误`。

---

### Attack 3: User Stories — Story 5 AC 是人工核查清单，不可自动化验收

**Where**: `Story 5 AC` — "Given 新的 `schema.sql`，When 逐条对照 JLC 规范检查，Then 所有表有 `COMMENT`，所有索引符合 `idx_`/`uk_` 命名，无 `status` 关键字直接用作字段名，无 TEXT 字段，无 FLOAT/DOUBLE 用于数值字段"

**Why it's weak**: "逐条对照 JLC 规范检查"是一个人工操作，不是可触发的测试条件。Given/When/Then 格式要求 When 是一个可重复执行的动作，但"人工检查"无法在 CI 中验证，也无法在代码评审中自动断言。这意味着 Story 5 的验收完全依赖人的记忆和注意力，而不是可执行的证据。

**What must improve**: 将 AC 改为可执行的 SQL 查询或脚本断言，例如：`When 对新 schema.sql 执行 grep/SQL 检查脚本，Then 输出 0 个违规项`，并在 §5.1 或附录中提供具体的检查命令（如 `grep -c ' TEXT' schema.sql` 应返回 0）。

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: 状态字段枚举值完全缺失 | ✅ | §5.1 新增完整枚举表，10 行，含 `user_status`/`item_status`/`pool_status` 全部允许值及含义，并说明校验层（service 层）和列类型（`VARCHAR(20) NOT NULL`） |
| Attack 2: `biz_key` 字段描述不完整 | ✅ | §5.1 全局变更表 `biz_key` 行补充：生成时机（service 层 Create 时）、JSON tag（`-`，不对外暴露）、唯一索引（`UNIQUE KEY uk_biz_key(biz_key)`）；§5.2 model 层表同步补充 `BizKey JSON tag: json:"-"` |
| Attack 3: Out-of-Scope 列表过于单薄 | ✅ | Out of Scope 从 3 项扩展至 7 项，新增回滚方案、CI/CD 流水线适配、监控查询适配、性能基准测试，每项均附排除原因 |

---

## Verdict

- **Score**: 98/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Document is near-complete. Two remaining deductions are both in Functional Specs: VARCHAR length rationale and completion range validation. Fixing these closes the final 2 points.
