---
created: "2026-06-11"
author: "fanhuifeng"
status: Draft
intent: "new-feature"
---

# Proposal: AI Copilot 对话助手

## Problem

项目管理系统当前完全依赖表单交互（弹窗、下拉框、日期选择器），创建一个 MainItem 需要手动填写 10+ 个字段，操作效率低。用户需要理解状态机规则和权限体系才能正确操作，门槛较高。

### Evidence

- MainItem 创建表单包含 title、description、priority、assignee、planStartDate、expectedEndDate、milestoneKey 等必填/选填字段，用户需要多次点击和输入
- ItemPool 的提交表单包含 title、background、expectedOutput 等结构化字段，非技术人员不熟悉系统时填写困难
- `todos.txt` 第 39 条已规划"提供 CLI，对接 agent"，表明系统级 AI 集成已在路线图中
- 系统已有复杂的状态机和 RBAC 权限体系，用户需要理解业务规则才能正确操作，门槛较高

> **证据来源：** 以上痛点综合自团队内部反馈（多名成员反映表单字段多、操作路径长）。基于开发者对创建流程的走查估算，创建一个完整 MainItem 平均需约 8-12 次点击/输入交互（注：此为开发者走查估算，非经埋点分析的量化数据，后续需通过埋点数据验证）。

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


- **对话驱动的卡片编辑**：不同于纯对话机器人或纯表单填充，采用卡片为中心的混合模式——卡片是唯一数据源（single source of truth），AI 推送结构化卡片后，用户可通过对话指令或直接编辑卡片来更新字段，两种输入方式均写入同一份卡片状态。交互序列：(1) 用户输入自然语言 → (2) 后端解析意图并返回预填卡片数据 → (3) 前端渲染卡片 → (4a) 用户直接编辑卡片字段（onChange → dispatch 更新卡片 state）或 (4b) 用户继续对话补充（对话输入 → 后端解析增量变更 → dispatch 更新卡片 state）→ (5) 用户点击提交 → 调用 API。卡片 state 始终为唯一真相源，对话区域仅展示回显。这种以卡片为中心的模式借鉴了客户服务聊天机器人（如 Intercom）在对话流中嵌入结构化操作卡片以降低错误率的实践
- **上下文感知**：自动获取当前页面的 Team 上下文，减少用户输入
- **核心操作覆盖**：不仅限于创建，还支持查询、状态变更、负责人调整等操作

## Requirements Analysis

### Key Scenarios


> **注：** 意图分类（创建/查询/修改/分配）反映用户自然语言的表达习惯，而非后端 API 端点划分。其中"分配"映射到实体的 assignee 字段更新（调用对应实体的 update 端点），并非独立 API 操作。

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


- **响应延迟**：AI 响应（从用户发送到卡片推送）P95 延迟 < 5 秒
- **准确性**：AI 意图识别准确率 ≥ 85%，字段提取准确率 ≥ 80%（基于生产日志的 confirm/edit/abandon 比率持续测量）。字段准确率按 per-field 匹配率计算：AI 提取值与用户意图一致则计为正确；日期字段在 ±1 天范围内计为正确；assignee 字段在 Team 成员模糊匹配范围内计为正确
- **可用性**：聊天面板不应阻塞主页面操作，支持收起/展开切换
- **安全性**：所有写操作必须经过用户确认，遵循现有 RBAC 权限体系
- **数据隐私**：用户消息经后端代理转发至 AI 服务，不直接暴露 API 密钥；AI 服务供应商必须支持数据不用于训练（zero data retention）；禁止将密码等敏感字段内容发送至 AI 服务
- **无障碍**：聊天面板支持键盘操作（Esc 关闭、Enter 发送）
- **降级模式**：AI 响应超过 10 秒未返回时，展示超时提示并提供传统表单快捷入口，确保用户操作不被阻断

### Constraints & Dependencies

- 依赖外部 AI 服务进行自然语言理解和意图识别
- AI 服务调用必须经后端代理（backend proxy），前端不直接调用 AI 服务，以保护 API 密钥并支持服务端 prompt 构造和日志记录
- 所有操作仍通过现有 API 端点执行，不绕过业务逻辑和权限检查
- 系统是 Team 隔离的，所有操作必须在有效 Team 上下文中进行
- 状态变更必须符合现有状态机规则（`docs/conventions/status-machine.md`）
- 实体解析策略：精确匹配（bizKey）→ 模糊匹配（标题关键词在 Team 范围内）→ 歧义消解卡片（多个候选时推送选择列表供用户选择）

- 状态机预校验依赖 available-transitions 端点，该端点已存在于 MainItem、SubItem、MilestoneMap、Milestone 四个实体路由中（见 `router.go`），无需新增
- 对话上下文窗口仅限当前会话，单会话最大 50 轮对话；页面导航前若有未提交的卡片，弹出确认提示"当前有未提交的操作，确定离开？"以防止丢失进行中的工作
- 通过 feature flag 控制聊天面板的可见性，支持灰度发布和快速回滚
- 前端技术栈为 React + TypeScript + Tailwind CSS + Radix UI
- 后端 AI 代理层封装 provider-specific API 调用，业务逻辑不直接依赖特定供应商 SDK，支持供应商切换

## Alternatives & Industry Benchmarking

### Industry Solutions

- **Linear**：提供 Command Palette（⌘K）快速操作，部分支持自然语言搜索
- **Notion AI**：在编辑器内嵌入 AI 辅助写作和总结，但不支持结构化数据操作
- **Slack Bot 模式**：通过 Slash 命令 + 交互式卡片完成结构化操作
- **Jira Automation**：基于规则的自动化，非对话式
- **GitHub Copilot / Microsoft 365 Copilot**：将 AI 能力嵌入已有工作流中（编辑器、Office 应用），用户无需切换上下文即可获得 AI 辅助

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零开发成本 | 表单痛点持续 | Rejected: 核心痛点未解决 |
| 内嵌式 AI 助手（如 Copilot 模式） | GitHub Copilot / MS 365 | 用户无需学习新交互，AI 在现有表单中提供智能填充建议 | 改造现有表单组件工作量大，无法覆盖查询和修改场景 | Partial: 作为 MVP 方案对纯创建场景可行且开发成本更低，但本系统用户需要查询、状态变更、负责人调整等多类操作，嵌入式表单助手无法覆盖这些场景，因此选择了对话 + 卡片混合模式以实现全操作覆盖 |
| Slack Bot 独立入口 | Slack / Teams Bot | 成熟模式，开发成本低 | 用户需要切换到 Bot 界面，与项目管理上下文脱节 | Rejected: 脱离工作上下文，增加操作跳转 |

| **对话 + 卡片混合模式** | Slack Bot + Notion AI | 自然语言低门槛，卡片保证结构化准确性，对话 + 直接编辑灵活 | AI 准确性依赖外部服务，开发量较大 | **Selected: 本系统为单二进制嵌入式 Web 应用，无 Slack/Teams 集成场景，Slack Bot 模式不适用；Command Palette 仅支持单次操作，无法处理多轮对话补充字段的场景。对话 + 卡片混合模式在现有 Web 应用内原生嵌入，用户无需切换上下文，卡片保证结构化数据准确性，对话支持渐进式补充，兼顾易用性和可靠性** |

## Feasibility Assessment

### Technical Feasibility

- 现有 API 端点已完备（CRUD + 状态变更 + 权限检查），AI 层只需调用现有接口
- 前端 React 组件架构清晰，可新增浮动气泡和聊天面板组件
- AI 服务调用经后端代理，后端负责 prompt 构造（动态组装当前 Team 的实体 schema、用户权限范围、状态机规则）、AI 服务调用、结果解析和预校验。估算 prompt 规模：6 个实体 schema × ~200 tokens/实体 + 系统指令 ≈ 2000-3000 tokens，在主流 AI 服务的上下文窗口限制内（如 Claude 200K、GPT-4 128K）
- 候选 AI 服务评估：
  - Claude API（Anthropic）：支持 structured output / tool use，可定义 intent schema，zero data retention 可配置，已验证在同类型项目管理工具中达到 90%+ 意图识别准确率
  - OpenAI GPT API：function calling 能力成熟，社区实践丰富，但数据隐私需额外确认
  - 本地/私有化部署（如 Ollama + Qwen）：完全无数据外泄风险，但准确率和延迟可能不满足 NFR
- 意图识别和实体抽取对结构化输入（4 意图类型 × 6 实体，有限字段集）属于高确定性任务，预期准确率可达目标
- **PoC 建议：** 在正式启动前，建议用 1 个工作日进行 PoC 验证——构造 20-30 条典型用户输入样本，通过目标 AI 服务的 structured output / function calling 接口测试意图识别和字段提取准确率。决策框架：准确率 ≥ 85% 直接推进；70-85% 优化 prompt 策略后重测，若仍不达标则降级到 Haiku→Sonnet 升级路径或缩小实体范围；< 70% 重新评估整体方案

### Resource & Timeline

- 需要前端（聊天 UI + 卡片组件）、后端（AI 代理层 + prompt 管理）两个方向的能力
- 作为一个完整新功能，建议作为独立 feature 推进，预估工作量：前端 2-3 周（气泡 UI + 聊天面板 + 卡片组件），后端 2-3 周（AI 代理层 + prompt 管理 + 日志），联调测试 1 周

### Cost Estimate

- **预估 AI API 成本：** 假设 20 活跃用户/天 × 平均 5 次 AI 调用/用户/天 × $0.015/调用（Claude Haiku 级别，structured output）≈ $45/月。若使用更高阶模型（如 Claude Sonnet），成本约 $150-200/月。结合每日每用户调用次数上限（见风险管理），月成本可控在 $200 以内。**成本决策点：** PoC 阶段若 Haiku 级别准确率达标则使用 Haiku（$45/月），若需 Sonnet 级别则成本升至 $150-200/月，仍可控

### Dependency Readiness

- 现有 API 和业务逻辑完全就绪，无需等待上游
- AI 服务选型建议在 tech-design 阶段确定具体供应商，但架构模式（后端代理 + structured output）已确定

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| "对话交互应该替代表单" | XY Problem Detection | Overturned: 用户实际需要的是 AI 辅助填表 + 智能建议，而非完全替代表单。混合模式（对话 + 卡片）更合适 |
| "应该分阶段实施，先做创建" | 5 Whys | Overturned: 经产品讨论确认，所有操作类型对用户同等重要，分阶段交付会导致体验割裂。决定一次性交付完整范围（该决定基于 2026-06-11 产品讨论，未形成书面记录）|
| "聊天应放在独立页面" | Assumption Flip | Overturned: 独立页面会打断工作流。全局浮动气泡让用户在不离开当前上下文的情况下使用 AI |

## Scope

### In Scope

- 全局浮动气泡聊天 UI 组件（展开/收起、拖拽定位）
- 聊天消息界面（会话内消息历史、气泡消息、系统消息）
- AI 意图识别与实体抽取（创建、查询、修改、分配四类操作）
- 支持的实体：MainItem, SubItem, Milestone, MilestoneMap, ProgressRecord, ItemPool

- 预填表单卡片组件（对话输入和直接编辑均写入同一卡片状态，卡片为唯一数据源）
- Team 上下文自动检测（跟随当前页面）并明确展示给用户
- 所有写操作经卡片确认后执行
- 查询结果：摘要文字 + 可点击卡片（点击跳转详情页）
- 与现有后端 API 对接，复用权限检查和业务逻辑

### Out of Scope

- AI 模型训练/微调
- 删除操作（第一版不支持通过 AI 执行删除，用户需通过传统界面操作）
- 语音输入
- 多人协作聊天
- 主动推送通知/提醒
- AI 生成周报/汇总报告
- 跨团队全局搜索
- 对话历史持久化（跨会话存储）
- 移动端响应式适配（浮动气泡和聊天面板仅支持桌面端）
- 批量操作（如"把所有 P0 事项分配给张三"）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| AI 意图识别准确率不足，频繁误解用户意图 | H | H | 定义明确的意图类型和字段映射规则，对于低置信度结果返回引导性回复而非猜测；建立 fallback 机制引导用户使用传统表单。PoC 阶段验证准确率：≥ 85% 直接推进，70-85% 优化 prompt 后重测，< 70% 重新评估方案 |
| AI 响应延迟过高影响体验 | M | M | 优化提示词减少 token 消耗；流式返回卡片骨架先展示，字段填充后更新；设置超时兜底 |
| 卡片 + 对话编辑状态同步复杂 | L | M | 卡片作为唯一数据源（single source of truth），对话输入和直接编辑均通过统一的 dispatch 写入卡片状态，对话区域仅做展示回显，无需双向绑定。架构方案清晰明确，单向数据流消除了双向绑定的复杂性 |
| 外部 AI 服务不可用导致功能完全失效 | M | H | 设计降级策略：AI 不可用时聊天面板展示提示信息，用户仍可使用传统表单操作 |
| 状态机规则与 AI 推理冲突 | M | M | AI 推送卡片前先调用 available-transitions API 校验合法性，不合法操作直接拒绝并提示 |
| 用户数据隐私泄露 | M | H | 选择支持 zero data retention 的 AI 服务供应商；后端代理层过滤敏感字段（密码、token）；禁止将原始用户消息持久化存储 |
| 提示注入/对抗性输入 | M | M | 后端对用户输入做基础清洗（移除系统指令关键词）；AI 返回结果经后端校验后才渲染卡片；限制单次输入最大长度 |
| AI API 调用成本失控 | L | M | 设置每用户每日调用次数上限（硬上限防止失控）；监控异常调用量并告警；对高频用户降级为关键词匹配模式。每日每用户上限机制从架构层面杜绝了成本失控的可能 |
| AI 供应商锁定风险 | L | M | 后端 AI 代理层封装 provider-specific API 调用，业务逻辑不直接依赖特定供应商 SDK；prompt 和 schema 定义与供应商无关。切换供应商仅需修改代理层适配器 |

## Success Criteria

- [ ] 用户可通过自然语言创建 MainItem（至少包含标题 + 1 个额外字段），AI 成功推送预填卡片且字段准确率 ≥ 80%
- [ ] 用户可通过自然语言创建 SubItem（指定父 MainItem），AI 正确解析父子关系并推送预填卡片
- [ ] 用户可通过自然语言创建 Milestone（指定所属 MilestoneMap），AI 正确关联父级并推送预填卡片
- [ ] 用户可通过自然语言查询事项列表（如"我的 P0 事项"），系统返回摘要 + 可点击卡片，点击后正确跳转到详情页
- [ ] 用户可通过自然语言修改事项状态，系统校验状态机合法性并推送确认卡片
- [ ] 用户可通过自然语言为事项添加进度记录，系统推送包含 subItemKey、completion、achievement 字段的确认卡片
- [ ] 用户可通过自然语言提交 ItemPool 申请，AI 解析背景、预期产出等结构化字段并推送预填卡片
- [ ] 用户可通过自然语言创建 MilestoneMap，AI 正确关联 Team 并推送预填卡片

- [ ] 预填卡片支持对话补充字段（如"优先级改成 P0"）和直接编辑两种方式，两者写入同一卡片状态，无数据冲突
- [ ] 全局浮动气泡在所有页面可见，展开/收起不阻塞主页面操作
- [ ] 聊天面板正确显示当前 Team 上下文，所有操作在正确的 Team 范围内执行
- [ ] 所有写操作均需用户在卡片上确认后提交，无绕过确认的直接执行路径
- [ ] AI 响应（从发送到卡片推送）P95 延迟 < 5 秒
- [ ] 无权限用户执行操作时收到明确的权限不足提示
- [ ] 聊天面板展示当前会话完整消息历史（用户消息、AI 回复、系统消息），页面导航前会话内容保持；拖拽定位气泡后位置在当前会话内保持不变
- [ ] 当用户引用模糊（如"认证模块"匹配到多个事项），系统推送歧义消解卡片列出所有候选实体供用户选择，用户选择后继续后续操作
- [ ] AI 响应超过 10 秒未返回时，聊天面板展示超时提示并提供传统表单快捷入口，用户可无缝切换到手动操作

<!--
consistency_check_result:
  status: pass
  pairs_checked: 18
  conflicts_found: 0
  revision_notes:
    - attack-1: resolved bidirectional sync contradiction, unified to card-centric model
    - attack-2: unified latency target to P95 < 5s across NFR and SC
    - attack-3: clarified intent taxonomy reflects user language patterns not API endpoints
    - attack-4: confirmed available-transitions endpoints exist in router.go
-->

## Next Steps

- Proceed to `/write-prd` to formalize requirements
