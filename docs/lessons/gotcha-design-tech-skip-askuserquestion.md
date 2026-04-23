# /design-tech 跳过 AskUserQuestion 直接输出文档

## Problem

执行 `/design-tech` 时，AI 读完 PRD 和代码后直接写出完整的 `tech-design.md`，没有使用 `AskUserQuestion` 确认设计决策，也没有逐节等待用户审批。用户无法在文档生成前介入，只能事后修改。

## Root Cause

SKILL.md 的 Step 4 和 Step 6 是**建议性描述**，不是强制约束：

- Step 4: "Use `AskUserQuestion` for ALL uncertain areas" — 依赖 AI 判断"是否有不确定项"。PRD 详细时 AI 判断为零，直接跳过
- Step 6: "For each section, wait for user approval" — 没有任何机制阻止 AI 一次性写完所有 section

对比 `eval-prd` 的 `<HARD-GATE>` 模式：它把"主会话控制循环"写成铁律，AI 无法绕过。`design-tech` 缺少这种结构。

根本原因：**PRD 详细 ≠ 设计决策已确认**。`AskUserQuestion` 的目的不是补充需求，而是让用户在文档生成前对技术方向有控制权。跳过它，用户变成被动审阅者而非主动参与者。

## Solution

### 执行层面（下次遇到时）

Step 3（Identify Decisions）完成后必须停下来，用 `AskUserQuestion` 提问，再进入 Step 5。

最少问两个问题：
1. 识别出的最关键架构决策（给出 2-3 个选项 + tradeoff）
2. 开放问题："有没有我应该注意的技术约束或偏好？"

Step 5 每输出一个 section 后停止，等用户明确回应后再继续。

### Skill 优化层面（修改 SKILL.md）

在 Step 3 后加 HARD-GATE：

```markdown
<HARD-GATE>
After Step 3, you MUST call AskUserQuestion before writing any design content.
This is non-negotiable regardless of PRD completeness.
Minimum: 1 question about the key architectural tradeoff + 1 open constraint question.
</HARD-GATE>
```

Step 5 改为逐节强制停止：

```markdown
Output ONE section, then STOP and wait for user response before proceeding.
```

Step 4 问题模板具体化（避免"没有不确定项"的误判）：

```markdown
Always ask at minimum:
1. The top architectural decision (present 2-3 options with tradeoffs)
2. Any constraint not covered by PRD (performance, team preference, timeline)
The goal is user control, not information gathering — ask even if PRD seems complete.
```

Step 2 代码探索委托给 Explore subagent，保护主会话 context，确保探索深度一致。

## Key Takeaway

**探索代码后先问，再写。** 即使 PRD 再详细，`/design-tech` 的交互步骤也不是可选的。

Skill 层面的根本修复：用 `<HARD-GATE>` 把"必须 AskUserQuestion"变成铁律，而不是依赖 AI 自判断是否有不确定项。
