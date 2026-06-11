---
created: "2026-06-11"
author: "fanhuifeng"
status: Draft
intent: "new-feature"
---

# Proposal: AI Copilot 对话助手

## Problem

项目管理系统当前完全依赖表单交互（弹窗、下拉框、日期选择器），创建一个 MainItem 需要手动填写 10+ 个字段，操作效率低且缺乏智能辅助（如优先级建议、负责人推荐）。

### Evidence

- MainItem 创建表单包含 title、description、priority、assignee、planStartDate、expectedEndDate、milestoneKey 等必填/选填字段，用户需要多次点击和输入
- ItemPool 的提交表单包含 title、background、expectedOutput 等结构化字段，非技术人员不熟悉系统时填写困难
- `todos.txt` 第 39 条已规划"提供 CLI，对接 agent"，表明系统级 AI 集成已在路线图中
- 系统已有复杂的状态机和 RBAC 权限体系，用户需要理解业务规则才能正确操作，门槛较高

### Urgency

随着团队和事项数量增长，表单交互的效率瓶颈会加剧。AI 对话助手可以在不改变现有业务逻辑的前提下大幅降低操作成本，且越早引入，用户越早受益。延迟实施意味着团队持续承担低效操作的时间成本。

## Proposed Solution

构建一个全局浮动气泡式的 AI Copilot 对话助手，嵌入现有 Web 应用中。用户通过自然语言描述意图（如"帮我创建一个 P1 事项：完成用户认证模块"），AI 解析意图后推送预填表单卡片到聊天窗口。必填且无法推导的字段留空，用户可通过直接编辑卡片或继续对话补充字段，确认后提交执行。

核心交互流程：
1. 用户点击浮动气泡展开聊天面板
2. 输入自然语言指令（创建/查询/修改/分配）
3. AI 识别意图和实体，从上下文推断字段值
4. 推送预填表单卡片（写操作）或摘要 + 可点击卡片（查询操作）
5. 用户确认/修改后提交，结果反馈到聊天界面

### Innovation Highlights

- **对话 + 卡片双向编辑**：不同于纯对话机器人或纯表单填充，采用混合模式——AI 推送结构化卡片，用户可以对话补充字段也可以直接编辑卡片，卡片实时同步更新
- **上下文感知**：自动获取当前页面的 Team 上下文，减少用户输入
- **全操作覆盖**：不仅限于创建，还支持查询、状态变更、负责人调整等完整 CRUD 操作

## Requirements Analysis

### Key Scenarios

**创建场景：**
- 用户："创建一个 P1 事项，标题是完成用户认证模块，分配给张三，下周五截止" → AI 推送预填卡片，字段：title=完成用户认证模块, priority=P1, assignee=张三, expectedEndDate=下周五对应日期, milestoneKey=留空
- 用户："加个里程碑，叫第一阶段验收" → AI 推送 Milestone 创建卡片
- 用户："在XX事项下加一个子任务" → AI 需识别 parent MainItem

**查询场景：**
- 用户："我的 P0 事项有哪些？" → 返回摘要文字 + 事项卡片列表，点击跳转详情
- 用户："里程碑A的进度怎么样？" → 返回进度摘要 + 子事项完成率统计

**修改场景：**
- 用户："把用户认证模块的状态改为进行中" → 推送状态变更确认卡片
- 用户："更新子任务完成度为 60%" → 推送进度更新卡片

**分配/调整场景：**
- 用户："把XX事项分配给李四" → 推送负责人变更确认卡片
- 用户："这个里程碑延期到下个月" → 推送日期调整确认卡片

**边界情况：**
- 用户输入模糊（如"创建一个事项"缺少关键字段）→ AI 推送部分预填卡片，高亮必填空字段
- AI 无法理解意图 → 返回引导性文字，列出可支持的操作类型
- 操作涉及权限不足 → 返回权限提示，不推送卡片
- 状态变更不符合状态机规则 → 返回错误说明，提示合法的目标状态

### Non-Functional Requirements

- **响应延迟**：AI 响应（从用户发送到卡片推送）应在 3 秒内完成
- **准确性**：AI 意图识别准确率 ≥ 85%，字段提取准确率 ≥ 80%
- **可用性**：聊天面板不应阻塞主页面操作，支持收起/展开切换
- **安全性**：所有写操作必须经过用户确认，遵循现有 RBAC 权限体系
- **无障碍**：聊天面板支持键盘操作（Esc 关闭、Enter 发送）

### Constraints & Dependencies

- 依赖外部 AI 服务进行自然语言理解和意图识别
- 所有操作仍通过现有 API 端点执行，不绕过业务逻辑和权限检查
- 系统是 Team 隔离的，所有操作必须在有效 Team 上下文中进行
- 状态变更必须符合现有状态机规则（`docs/conventions/status-machine.md`）
- 前端技术栈为 React + TypeScript + Tailwind CSS + Radix UI

## Alternatives & Industry Benchmarking

### Industry Solutions

- **Linear**：提供 Command Palette（⌘K）快速操作，部分支持自然语言搜索
- **Notion AI**：在编辑器内嵌入 AI 辅助写作和总结，但不支持结构化数据操作
- **Slack Bot 模式**：通过 Slash 命令 + 交互式卡片完成结构化操作
- **Jira Automation**：基于规则的自动化，非对话式

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零开发成本 | 表单痛点持续，无智能辅助 | Rejected: 核心痛点未解决 |
| 仅优化表单 | 行业常规 | 开发量小，风险低 | 不解决智能辅助需求，仍是手动操作 | Rejected: 未满足智能辅助决策需求 |
| Command Palette (⌘K) | Linear | 键盘友好，快速 | 不支持自然语言，学习成本高 | Rejected: 不符合 AI Native 交互目标 |
| **对话 + 卡片混合模式** | Slack Bot + Notion AI | 自然语言低门槛，卡片保证结构化准确性，双向编辑灵活 | AI 准确性依赖外部服务，开发量较大 | **Selected: 兼顾易用性和可靠性，与用户需求最匹配** |

## Feasibility Assessment

### Technical Feasibility

- 现有 API 端点已完备（CRUD + 状态变更 + 权限检查），AI 层只需调用现有接口
- 前端 React 组件架构清晰，可新增浮动气泡和聊天面板组件
- 意图识别和实体抽取是成熟 AI 能力，风险可控

### Resource & Timeline

- 需要前端（聊天 UI + 卡片组件）、后端（AI 意图解析层）、AI 集成三个方向的能力
- 作为一个完整新功能，工作量较大，建议作为独立 feature 推进

### Dependency Readiness

- 现有 API 和业务逻辑完全就绪，无需等待上游
- AI 服务选型需要在 tech-design 阶段确定

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| "对话交互应该替代表单" | XY Problem Detection | Overturned: 用户实际需要的是 AI 辅助填表 + 智能建议，而非完全替代表单。混合模式（对话 + 卡片）更合适 |
| "应该分阶段实施，先做创建" | 5 Whys | Overturned: 用户认为所有操作同等重要，希望一次到位。Challenge Override: user chose to proceed with full scope. Reason: 希望一次到位 |
| "聊天应放在独立页面" | Assumption Flip | Overturned: 独立页面会打断工作流。全局浮动气泡让用户在不离开当前上下文的情况下使用 AI |

## Scope

### In Scope

- 全局浮动气泡聊天 UI 组件（展开/收起、拖拽定位）
- 聊天消息界面（会话内消息历史、气泡消息、系统消息）
- AI 意图识别与实体抽取（创建、查询、修改、分配四类操作）
- 支持的实体：MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool
- 预填表单卡片组件（支持对话 + 直接编辑双向同步）
- Team 上下文自动检测（跟随当前页面）并明确展示给用户
- 所有写操作经卡片确认后执行
- 查询结果：摘要文字 + 可点击卡片（点击跳转详情页）
- 与现有后端 API 对接，复用权限检查和业务逻辑

### Out of Scope

- AI 模型训练/微调
- 语音输入
- 多人协作聊天
- 主动推送通知/提醒
- AI 生成周报/汇总报告
- 跨团队全局搜索
- 对话历史持久化（跨会话存储）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| AI 意图识别准确率不足，频繁误解用户意图 | M | H | 定义明确的意图类型和字段映射规则，对于低置信度结果返回引导性回复而非猜测；建立 fallback 机制引导用户使用传统表单 |
| AI 响应延迟过高影响体验 | M | M | 优化提示词减少 token 消耗；流式返回卡片骨架先展示，字段填充后更新；设置超时兜底 |
| 卡片 + 对话双向编辑状态同步复杂 | M | M | 设计明确的单向数据流：对话修改 → 更新卡片状态；卡片编辑 → 同步到对话显示。避免双向绑定的复杂性 |
| 外部 AI 服务不可用导致功能完全失效 | L | H | 设计降级策略：AI 不可用时聊天面板展示提示信息，用户仍可使用传统表单操作 |
| 状态机规则与 AI 推理冲突 | M | M | AI 推送卡片前先调用 available-transitions API 校验合法性，不合法操作直接拒绝并提示 |

## Success Criteria

- [ ] 用户可通过自然语言创建 MainItem（至少包含标题 + 1 个额外字段），AI 成功推送预填卡片且字段准确率 ≥ 80%
- [ ] 用户可通过自然语言查询事项列表（如"我的 P0 事项"），系统返回摘要 + 可点击卡片，点击后正确跳转到详情页
- [ ] 用户可通过自然语言修改事项状态，系统校验状态机合法性并推送确认卡片
- [ ] 预填卡片支持对话补充字段（如"优先级改成 P0"）和直接编辑两种方式，双向同步无冲突
- [ ] 全局浮动气泡在所有页面可见，展开/收起不阻塞主页面操作
- [ ] 聊天面板正确显示当前 Team 上下文，所有操作在正确的 Team 范围内执行
- [ ] 所有写操作均需用户在卡片上确认后提交，无绕过确认的直接执行路径
- [ ] AI 响应（从发送到卡片推送）P95 延迟 < 5 秒
- [ ] 无权限用户执行操作时收到明确的权限不足提示

<!--
consistency_check_result:
  status: pass
  pairs_checked: 18
  conflicts_found: 0
-->

## Next Steps

- Proceed to `/write-prd` to formalize requirements
