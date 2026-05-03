---
date: "2026-04-30"
doc_dir: "docs/features/e2e-test-scripts-rebuild/prd/"
iteration: "1"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

**Score: 88/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  20      │  20      │ ✅          │
│    Background three elements │  7/7     │          │            │
│    Goals quantified          │  7/7     │          │            │
│    Logical consistency       │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  17      │  20      │ ⚠️          │
│    Mermaid diagram exists    │  7/7     │          │            │
│    Main path complete        │  6/7     │          │            │
│    Decision + error branches │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  13      │  20      │ ⚠️          │
│    Tables complete           │  5/7     │          │            │
│    Field descriptions clear  │  5/7     │          │            │
│    Validation rules explicit │  3/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  19      │  20      │ ✅          │
│    Coverage per user type    │  7/7     │          │            │
│    Format correct            │  7/7     │          │            │
│    AC per story              │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  19      │  20      │ ✅          │
│    In-scope concrete         │  7/7     │          │            │
│    Out-of-scope explicit     │  7/7     │          │            │
│    Consistent with specs     │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  88      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| 流程图: VerifyMarker → Error([错误：重试]) | 错误节点无出边，"重试"语义未定义，流程图在此处断路 | -1 pt (Main path) |
| 流程图: Start([开始：遍历 11 个 feature]) | 图中无循环结构，Done 节点后无箭头指向下一个 feature | -1 pt (Main path) |
| 流程图: AuthorConfirm / Error 分支 | 两个终止节点（Block、Error）均无后续路径，异常处理语义不完整 | -2 pts (Decision + error branches) |
| 5.1 脚本生成规范 | 无 /gen-test-scripts 失败时的处理规则（如 sitemap.json 缺失、脚本语法错误） | -2 pts (Validation rules) |
| 5.3 已知失败处理 | KNOWN_FAILURES.md 仅列字段名，无格式规范（Markdown 表格？YAML？字段是否必填？） | -1 pt (Validation rules) |
| 5.1/5.2 表格 | 表格行数偏少（5.1 共 5 行，5.2 共 4 行），未覆盖脚本幂等性、版本冲突等边界情况 | -2 pts (Tables) |
| 5.1/5.2 字段描述 | "去重：/graduate-tests 自动处理与已有脚本的重复覆盖" 未说明重复判定逻辑（按文件名？按内容 hash？） | -2 pts (Field descriptions) |
| prd-user-stories.md: Story 3 AC | AC 中 "graduate 可继续进行" 跳过了流程图中的 "作者确认" 门控步骤，与 prd-spec.md 不一致 | -1 pt (AC per story) |
| Scope vs 流程图 | In-scope 写 "graduate 前更新 sitemap.json（如已过期）"，流程图条件是 "UI feature?"，两者判定条件不同 | -1 pt (Consistent with specs) |

---

## Attack Points

### Attack 1: [Flow Diagrams — 错误重试路径断路，迭代结构缺失]

**Where**: `VerifyMarker -->|否| Error([错误：重试])` 以及 `Done([该 feature 完成])`

**Why it's weak**: `Error` 节点标注"重试"但无出边——读者无法判断重试是回到 `Graduate` 还是回到 `GenScripts`，甚至是整个流程终止。同样，`Done` 节点后没有箭头指向"处理下一个 feature"，而图的起点明确写了"遍历 11 个 feature"，循环结构完全缺失。这两处让流程图在关键路径上失去可执行性。

**What must improve**: 为 `Error` 节点添加出边（例如 `Error --> Graduate` 表示重试 graduate 步骤，或 `Error --> Block` 表示人工介入）；在 `Done` 后添加循环箭头回到 `CheckTC` 或新增 `NextFeature` 节点，明确遍历逻辑。

---

### Attack 2: [Functional Specs — 验证规则缺失，关键失败场景无处理定义]

**Where**: 5.1 脚本生成规范表格（5 行）和 5.3 已知失败处理表格（3 行）

**Why it's weak**: 5.1 只列出了"正常生成"的规范，完全没有定义失败场景的处理规则：`/gen-test-scripts` 执行失败怎么办？`sitemap.json` 不存在或格式错误时是阻塞还是跳过？生成的脚本有语法错误时是否算入通过率计算？5.3 的 KNOWN_FAILURES.md 只列了字段名（"feature slug、失败断言描述、失败原因、负责人"），没有格式规范——是 Markdown 表格、YAML 还是自由文本？字段是否全部必填？这让"作者确认"这个门控步骤无法被机械验证。

**What must improve**: 在 5.1 增加一行"脚本生成失败处理"，明确阻塞条件；在 5.3 给出 KNOWN_FAILURES.md 的具体条目格式（至少一个示例行），并标注哪些字段必填。

---

### Attack 3: [User Stories — Story 3 AC 与流程图存在门控步骤不一致]

**Where**: prd-user-stories.md Story 3 AC: `Then graduate 可继续进行，KNOWN_FAILURES.md 包含 feature slug、失败描述和负责人，且无未记录的失败断言进入回归套件`

**Why it's weak**: prd-spec.md 流程图中，通过率不足时的路径是 `RecordFailures → AuthorConfirm{作者确认?} →|否| Block`，即"作者确认"是 graduate 的前置门控。但 Story 3 的 AC 完全省略了这个确认步骤——只要记录到 KNOWN_FAILURES.md 就可以 graduate。这不是措辞问题，而是业务逻辑矛盾：按 AC 执行，维护者可以绕过作者确认直接 graduate；按流程图执行，则需要等待确认。两个文档对同一场景的定义不同。

**What must improve**: Story 3 AC 的 When 步骤需要加入"作者确认"动作，或在 Then 中明确"graduate 需等待作者确认后方可进行"，与流程图保持一致。

---

## Previous Issues Check

<!-- Only for iteration > 1 — N/A for iteration 1 -->

---

## Verdict

- **Score**: 88/100
- **Target**: 90/100
- **Gap**: -2 points
- **Action**: Continue to iteration 2 — fix Attack 2 (validation rules) and Attack 3 (Story 3 AC inconsistency) to close the gap
