---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/prd/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

**Score: 91/100** (target: N/A)

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
│ 3. Functional Specs          │  13      │  20      │ ⚠️          │
│    Tables complete           │   5/7    │          │            │
│    Field descriptions clear  │   5/7    │          │            │
│    Validation rules explicit │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  20      │  20      │ ✅          │
│    Coverage per user type    │   7/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  18      │  20      │ ⚠️          │
│    In-scope concrete         │   7/7    │          │            │
│    Out-of-scope explicit     │   5/7    │          │            │
│    Consistent with specs     │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  91      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| 功能描述 §5.1 — 字段重命名表 | `user_status`、`item_status`、`pool_status` 的允许值域（枚举值）完全缺失，无法验收 | -3 pts (validation rules) |
| 功能描述 §5.1 — 全局变更表 | `biz_key` 仅注明"雪花算法生成"，未说明字段用途（业务唯一标识）、是否对外暴露、是否参与唯一索引 | -2 pts (field descriptions) |
| 功能描述 §5.1 — TEXT→VARCHAR 表 | VARCHAR 长度选取依据未说明（为何 `description` 是 2000 而非 1000 或 4000），超限策略是条件性兜底而非明确约束 | -2 pts (tables/validation) |
| Scope — Out of Scope | 仅列 3 项，缺少明确排除：回滚方案、CI/CD 流水线变更、监控告警配置、性能测试 | -2 pts (out-of-scope explicit) |

---

## Attack Points

### Attack 1: Functional Specs — 状态字段枚举值完全缺失

**Where**: `§5.1 字段重命名` — "`status` (users) → `user_status`"、"`status` (main_items/sub_items) → `item_status`"、"`status` (item_pools) → `pool_status`"

**Why it's weak**: 字段被重命名了，但允许的值域从未定义。`user_status` 能取哪些值？`item_status` 的状态机是什么？没有枚举值，DBA 无法写 CHECK 约束，后端无法写枚举类型，前端无法做状态映射，AC 中也无法验证"状态正确"。这是一个 breaking change，却对最关键的语义信息只字未提。

**What must improve**: 在 §5.1 或独立的"字段约束"小节中，为每个 status 字段列出完整枚举值及其含义（如 `item_status: 0=待处理, 1=进行中, 2=已完成`），并说明是否需要 MySQL CHECK 约束或应用层枚举校验。

---

### Attack 2: Functional Specs — `biz_key` 字段描述不完整

**Where**: `§5.1 全局变更表` — "`biz_key BIGINT NOT NULL`（雪花算法生成）"

**Why it's weak**: 这是新增的全局字段，影响所有 6 张表，但描述只有 5 个字。以下问题全部未回答：`biz_key` 是否对外暴露（JSON 响应中是否有 `bizKey`）？是否作为唯一索引（`uk_biz_key`）？雪花算法由谁生成（数据库触发器？Go 服务层？）？如果生成失败，插入是否回滚？前端 Story 4 的 AC 中完全没有提到 `bizKey`，说明这个字段的影响范围在文档中是模糊的。

**What must improve**: 在 §5.2 model 层变更表中补充 `BizKey` 字段的 JSON tag（是否暴露）、生成时机（service 层 Create 前赋值）、唯一索引声明，以及 Story 4 的 AC 中明确 `bizKey` 是否出现在响应体。

---

### Attack 3: Scope Clarity — Out-of-Scope 列表过于单薄

**Where**: `Scope — Out of Scope` — "数据迁移脚本（SQLite → MySQL 数据导出，独立任务）/ API 合同文档（OpenAPI/Swagger）更新（当前项目无正式 API 文档）/ MySQL 服务器部署与配置"

**Why it's weak**: 仅 3 项，且第 2 项的排除理由是"当前项目无正式 API 文档"——这是一个事实陈述，不是范围决策。以下明显相关的项目被完全忽略：回滚方案（schema 变更失败如何恢复？）、CI/CD 流水线适配（测试环境是否需要切换数据库？）、监控告警（`deleted_flag` 替换后现有告警查询是否失效？）、性能测试（`DECIMAL` 替换 `REAL` 对大批量计算的影响）。这些缺失会在执行阶段引发"这算不算本次范围"的争议。

**What must improve**: 补充至少 4 项明确排除：回滚脚本、CI/CD 数据库切换、监控查询适配、性能基准测试，并对每项简述排除原因（如"独立任务"或"当前无此基础设施"）。

---

## Previous Issues Check

<!-- Only for iteration > 1 — N/A for iteration 1 -->

---

## Verdict

- **Score**: 91/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Document is strong overall. Three targeted fixes in Functional Specs (status enum values, biz_key description) and Scope (out-of-scope list) would close the remaining 9 points.
