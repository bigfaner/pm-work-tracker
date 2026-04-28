# Tech Design 必须主动识别并记录关键决策

## Problem

设计 DecisionLog 模型时使用了 `status` 作为列名，违反了已有的 `<prefix>_status` 命名约定（`item_status`、`pool_status`）。用户发现后要求修正。

## Root Cause

探索代码库时看到了 `item_status`、`pool_status` 的命名模式，但**只复制了结构，没有追问原因**。tech-design Step 2 要求检查 `docs/decisions/` 目录以发现已有决策，当时目录为空就跳过了。实际上应该从代码模式中提炼出隐含的决策：为什么不用 `status`？

因果链：
- 表象：用了裸 `status` 列名
- 直接原因：没识别出前缀命名是刻意约定
- 根本原因：探索模型时只看结构不问 why；`docs/decisions/` 为空时没有反向从代码提炼隐含决策
- 触发条件：决策目录为空 + 缺乏主动模式识别意识

## Solution

修正为 `log_status`（Go: `LogStatus`，JSON: `logStatus`）。创建了 `docs/decisions/prefixed-status-columns.md` 记录此约定。

## Key Takeaway

**tech-design 流程中，`docs/decisions/` 的作用是双向的：**

1.  **读**：检查已有决策，避免重复犯错（如 MySQL 保留字约定）
2.  **写**：设计过程中主动识别关键决策并归档，不要等用户纠正

**实操规则：**
- 探索现有模型时，对每个非显而易见的命名/结构选择追问 "why"
- 如果 `docs/decisions/` 为空但代码中有明显模式，说明决策尚未文档化——应主动提炼并记录
- tech-design 完成后执行 Step 7 归档流程，列出候选决策供用户确认
