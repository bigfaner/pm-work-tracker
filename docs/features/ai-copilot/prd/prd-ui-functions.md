---
feature: "AI Copilot 对话助手"
---

# AI Copilot 对话助手 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

AI Copilot 以**全局浮动 overlay** 形式嵌入现有 Web 应用，覆盖所有已认证页面（不含 `/login`）。包含一个常驻浮动气泡（入口）和展开后的聊天面板（对话区 + 多种卡片）。本功能**不新建任何页面**，所有 UI 均为挂载在现有页面之上的 overlay 组件；查询结果卡片点击后跳转到**已有的**详情页。

## Navigation Architecture

- **Platform**: web

### Primary Navigation (shared across pages)

AI Copilot 的浮动气泡**不属于**现有侧边栏主导航，而是独立的全局 overlay 入口，出现在所有已认证页面的固定屏幕位置（视口右下角，距右边缘 24px、距下边缘 24px）。它不改变现有页面间的导航结构。

| # | Label | Target Page | Icon Keyword |
|---|-------|-------------|-------------|
| — | （无新增主导航项；Copilot 为 overlay 入口，非路由跳转） | — | — |

### Secondary Pages (navigated from a parent page)

Copilot 不创建新页面。唯一的跨页面跳转来自**查询结果卡片**（UF-4）点击后跳转到既有详情页：

| Page | Entry Point (UF# or action) | Return Target |
|------|-----------------------------|---------------|
| `/items/:mainItemId` 主事项详情 | UF-4 查询结果卡片点击 | 来源页面（浏览器返回） |
| `/items/:mainItemId/sub/:subItemId` 子事项详情 | UF-4 查询结果卡片点击 | 来源页面（浏览器返回） |
| `/milestones/:mapId` 里程碑图详情 | UF-4 查询结果卡片点击 | 来源页面（浏览器返回） |
| `/item-pool` 待办事项 | UF-6 降级提示中的"去手动操作"快捷入口 | 来源页面 |

### Navigation Rules

- Primary navigation is shared across pages（现有侧边栏不变）。
- Copilot 浮动气泡在所有已认证页面可见，feature flag 关闭时整体隐藏。
- 查询结果卡片跳转目标必须是 sitemap 中已存在的路由，不产生新路由。
- 页面导航不弹离开确认 —— 未提交卡片草稿会丢失，但用户主动重新发起即可（不强制挽留）。

---

## UI Function 1: 浮动气泡（全局入口）

### Placement

- **Mode**: existing-page
- **Target Page**: global — 所有已认证路由（`/items`、`/items/:mainItemId`、`/items/:mainItemId/sub/:subItemId`、`/weekly`、`/gantt`、`/table`、`/item-pool`、`/milestones`、`/milestones/:mapId`、`/report`、`/teams`、`/teams/:teamId`、`/users`、`/roles`），`/login` 不显示
- **Position**: 屏幕固定位置（视口右下角，距右边缘 24px、距下边缘 24px），`z-index: 50`，悬浮于页面内容之上，不遮挡主操作区

### Description

常驻浮动气泡，是 Copilot 的唯一入口。点击展开聊天面板（UF-2），展开时气泡隐藏。首次出现带红点徽章引导发现。面板收起后会话后台继续，agent 返回消息或需确认时气泡显示脉冲活动徽章（activity-badge）提示用户回来。

### User Interaction Flow

用户点击气泡 → 展开聊天面板（UF-2），气泡隐藏、首次红点消除、焦点入输入区。会话内支持拖拽改变气泡位置（位移 <4px 视为点击）。面板展开期间无"点击气泡收起"（气泡已隐藏）；收起由 Esc 或点面板外区域触发（见 UF-2）。收起后若 agent 有新活动，气泡显示活动徽章，点击再次展开。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 首次引导徽章 | boolean | 本地存储（首次标记） | 首次展示红点，用户首次展开后消除 |
| 活动徽章 | number（未读数） | 后端推送 | 面板收起期间累计的未读消息/待确认数；展开后清零 |
| 气泡位置 | {x, y} | 会话内状态 | 拖拽后更新，会话内保持 |
| feature flag 可见性 | boolean | 后端配置 | 关闭时气泡不渲染 |
| AI 可用性 | boolean | 后端代理健康 | 不可用时降为禁用态 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 默认 | accent 圆形气泡 + 消息图标 | 页面加载、feature flag 开启 |
| 首次引导 | 气泡 + 右上红点 | 用户首次访问、未展开过 |
| 后台活动（面板收起态） | 气泡右上活动徽章（error 红、带未读数、脉冲动画） | 面板收起期间 agent 返回消息或需确认；展开即消除 |
| AI 不可用 | 灰色 icon、border-dark | 后端代理报告 AI 服务不可用 |
| 隐藏 | 不渲染（display:none） | feature flag 关闭 |

### Validation Rules

- feature flag 关闭时气泡完全不渲染（非仅隐藏）。
- 气泡不得遮挡页面主操作按钮（需与现有页面元素做避让）。
- 活动徽章在面板展开后立即清零。

---

## UI Function 2: 聊天面板（全局 overlay）

### Placement

- **Mode**: existing-page
- **Target Page**: global — 同 UF-1 所有已认证路由
- **Position**: 右侧抽屉，`position: fixed`，right:0/top:0/bottom:0，**默认宽 630px**，从视口右边缘滑入（translateX 200ms），shadow level-3，`z-index: 50`。**左边缘 6px 拖拽手柄可调节宽度（区间 420–960px），松手写入 sessionStorage，同会话恢复**。不阻塞主页面（主页可滚动/点击）。**无关闭按钮**——由 Esc 或点面板外区域收起。

### Description

展开后的对话主界面，含三段式头部 + 双视图内容区 + 双模式输入区。**头部**：Team 徽章（左）· 当前会话标题（中）· 开启新会话 + 历史会话按钮（右）。**双视图**：对话视图（消息历史 + Agent 过程追踪 UF-8 + 卡片 UF-3~UF-7 + 输入区）↔ 会话列表视图（历史会话，单活动会话，不并行多会话）。**双模式输入区**：文本模式（textarea）↔ 选项组模式（键盘选项组，承载所有确认操作，无二级弹窗）。单会话最大 50 轮。

### User Interaction Flow

用户键入指令 → Enter/发送 → 消息追加 → 流式展示 Agent 过程追踪（UF-8）→ 渲染回复（文字/卡片）。需确认时输入区切换为选项组（↑↓/Enter/←/Esc）。点头部「历史」切到会话列表视图（标题变"历史会话"），选中会话切换活动会话；点「+」新建会话（已是新会话则不重复创建）。Esc 或点面板外区域收起（会话后台继续）。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 当前 Team 上下文 | Team {bizKey,name} | 页面路由/全局状态 | 头部 Team 徽章 |
| 当前会话标题 | string | 首条指令摘要或"新会话" | 头部居中；列表视图时显示"历史会话" |
| 会话列表 | Session[]{id,title,preview,time,msgCount,pending} | 本地/后端 | 会话列表视图 |
| 消息历史 | List<{role,content,cardRef,ts}> | 会话内状态 | 单会话 ≤50 轮，不跨会话持久化 |
| 输入文本 | string | 用户输入 | 限 1000 字符，超出截断 |
| 输入模式 | enum(text/options) | 当前是否有待确认 | options 模式渲染键盘选项组 |
| 待确认选项 | Option[]{label,action,hint} | 当前卡片决策 | 选项组内容 |
| 发送状态 | enum(idle/thinking/streaming/error/timeout) | 后端代理 | 驱动加载/错误/超时态 |
| 面板宽度 | number(px) | sessionStorage（copilotPanelW） | 420–960 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 空（首次展开 / 新建会话） | 欢迎语 + 能力引导卡片（UF-7） | 面板首次展开、或用户新建会话使消息区为空 |
| 思考中 | 输入框 + 发送按钮禁用 + 三点跳动；输入区上方 inline 提示"AI 处理中，请稍候…" | 指令已发送，等 AI；用户在此态不能发送新指令 |
| 流式返回 | Agent 过程追踪（UF-8）流式追加 | AI 流式返回（不支持流式则退化单条"思考中"） |
| 正常对话 | 消息历史 + 输入区可用 | AI 返回完成 |
| 选项组模式 | 输入区为键盘选项组（↑↓/Enter/←/Esc） | 存在待确认决策（提交/取消/diff/歧义/降级） |
| 会话列表视图 | 历史会话列表，标题"历史会话" | 点头部「历史」 |
| 发送阻断（Team 缺失） | 发送禁用 + inline notice | teamCtx 缺失 |
| 发送阻断（轮次上限） | 发送禁用 + inline notice + 新建会话 | sessionRoundCount ≥50 |
| 流式中断 | 半填充骨架丢弃 + 系统消息 + 重试 | 流式连接断开（仅 UF-3） |
| Team 切换（in-flight） | in-flight 卡片冻结 + warning notice | 导航切 Team 且有 pending 卡片 |
| 错误 | 系统消息红色 + 重试 | 后端代理错误 |
| 超时 | 超时提示 + 降级入口（UF-6） | AI >10s 未返回 |
| 收起 | translateX(100%) 隐藏 | Esc / 点面板外（会话后台继续） |

### Validation Rules

- 单次输入超过 1000 字符截断并提示"已超出单次最大长度 1000 字符，已截断"；textarea min-h 120px、max-h 280px、按内容自动增高、到顶滚动。
- 面板宽度可拖拽调节，区间 420–960px，会话内持久化。
- Team 上下文缺失时阻止发送并提示先进入 Team 页面；会话达 50 轮阻止发送并提示新建。
- 无关闭按钮；Esc 或点面板外区域收起（不中断会话）。
- 选项组模式下 ← 返回文本输入，Esc 收起面板；点击选项直接选中并确认（不收起面板）。
- 多卡片可共存，但 AI 请求串行（思考态全局唯一）；切换会话时前一会话待确认卡片保留。
- **用户串行发送**：用户在 AI 处理上一条指令期间（思考态/流式态）**不能发送新指令**——输入框与发送按钮同时禁用，输入区上方提示"AI 处理中，请稍候…"；待上一轮 AI 返回（成功/失败/超时）后恢复可发送。**单次只能发送一条用户消息**，避免并发 prompt 组装冲突与上下文错乱。
- **Agent 多消息返回**：单次用户指令触发 AI 一轮响应，但该轮响应可由**多条 AI 消息**组成（按顺序追加到消息流）：Agent 过程追踪（UF-8，思考/计划/操作步骤）→ 最终卡片或文字响应（UF-3/4/5/6）→ AI 跟进消息（如确认完成、引导下一轮）。用户看到的"AI 在回答我"是这一串消息的整体呈现，而非单条消息。

---

## UI Function 3: 预填表单卡片（写操作）

### Placement

- **Mode**: existing-page
- **Target Page**: global — 渲染于 UF-2 聊天面板的消息区内
- **Position**: 作为消息流中的一条卡片消息

### Description

写操作（创建/修改/分配）的确认卡片。**前置依赖：意图回执 + 主动澄清（UF-9）通过后渲染**——agent 先用文本回执意图、主动澄清收齐所有必填信息，本卡片（UF-3）才 append 到消息流（此时所有必填字段已预填、无 required-highlight 高亮）。用户可直接编辑卡片字段（如改主意向 P0、调整截止），或继续对话补充——两种方式均写入同一份卡片状态（卡片为唯一数据源）。提交前经状态机/权限预校验。

**表单生命周期**：表单卡片是消息流中的一条消息，自带状态机 —— `预填展示 → 编辑中 → 提交中 → 已提交（折叠态）/ 提交失败 / 已丢弃`。**已提交、已丢弃 两种终态默认折叠为单行摘要**（"✓ 已提交 · {title}" / "⊘ 已丢弃 · {title}"），点击可展开只读字段。**提交成功后 AI 跟进一条自然语言消息**衔接下一轮对话，避免表单成为对话末端的哑终端。表单本身的数据不二次抽取为结构化摘要——表单消息不可变即天然快照，**上下文完整性靠 transcript 自身维持**，不依赖额外的结构化 envelope。

### User Interaction Flow

AI 推送卡片 → 用户直接编辑字段（onChange 更新卡片 state）或对话补充（后端解析增量变更后更新卡片 state）→ 用户点击提交：
- 若为状态变更类操作且目标实体支持 available-transitions 端点（MainItem/SubItem/MilestoneMap/Milestone）→ 提交前调用 available-transitions 预校验 → 通过则调用现有 API；不通过则卡片内显示错误（含 `validTransitions` 合法目标状态）。
- 若为创建操作、ProgressRecord、ItemPool、或非状态变更类修改/分配 → 不做预校验，直接调用现有 API，依赖后端 RBAC + 状态机校验；失败由后端返回错误信息在卡片内展示。
- 高影响写操作（分配/状态变更）提交前，卡片必须显示目标实体的 title + bizCode 供用户二次确认（防止错实体）。

**提交后的生命周期**：
- **提交中**：字段锁定、提交按钮转 spinner；trace（UF-8）末尾追加一行"⏳ 调用 {Entity} {Op} API…"步回显，成功后该步变 ✓ —— 让用户明确"现在在写库了"。
- **提交成功**：表单折叠为单行"✓ 已提交 · {title}"（默认折叠态，点击展开只读字段）。**AI 跟进一条自然语言消息**（如"好，已为你创建 P1 事项「完成用户认证模块」，分配给了张三，下周五截止。还需要加里程碑吗？"）——衔接下一轮对话。
- **提交失败（后端校验）**：表单**保持展开**，错误信息以**字段级就近展示**（出错字段红框 + 字段下方错误说明）+ 顶部错误条概述；用户已编辑的字段值全部保留（不回退到 AI 预填值）；当前 turn 顶部小字标注"重试 N 次"，历史会话回看此 turn 时不显示重试次数；提交按钮恢复可点用于重试。
- **丢弃**：表单折叠为"⊘ 已丢弃 · {title}"单行，保留在 transcript 中避免出现孤儿用户输入。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 卡片类型 | enum(create/update/assign) | AI 意图 | 决定字段集 |
| 目标实体 | enum(MainItem/SubItem/Milestone/MilestoneMap/ProgressRecord/ItemPool) | AI 实体抽取 | 决定 schema |
| 字段集 | Map<fieldName, {value, required, derived}> | AI 抽取 + 实体 schema（见 [`entity-schemas.md`](./entity-schemas.md)） | required 且无值字段高亮；schema 同时驱动表单控件类型与结果展示槽位 |
| 父实体引用 | bizKey | 页面上下文 / 模糊匹配 | SubItem 需 parent MainItem；Milestone 需 parent MilestoneMap |
| 校验错误 | string | available-transitions 预校验 | 不合法时卡片内展示 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 预填展示 | 字段值 + 必填空字段高亮 + 高影响操作的目标实体 title/bizCode 二次确认区 | AI 推送完成 |
| 编辑中 | 字段可编辑；onChange 更新卡片 state | 用户聚焦字段 |
| 校验失败（提交前） | 错误说明 + `validTransitions` 合法目标状态列表，表单保持可编辑 | available-transitions 返回不合法（仅状态变更类，预校验阶段） |
| 提交中 | 提交按钮禁用 + 字段锁定 + trace（UF-8）末尾追加"⏳ 调用 {Entity} {Op} API…"步回显 | 用户点击提交 |
| 提交失败（后端校验） | 表单保持展开 + 字段级错误就近展示（红框 + 字段下方说明）+ 顶部错误条概述；已编辑字段值全保留；当前 turn 顶部小字"重试 N 次"（历史 turn 回看时不显示） | 后端返回参数错误 |
| 已提交（折叠态） | 表单折叠为单行"✓ 已提交 · {title}"（点击展开只读字段）+ AI 跟进自然语言消息 | API 成功 |
| 已丢弃（折叠态） | 表单折叠为单行"⊘ 已丢弃 · {title}"，保留在 transcript 中 | 用户点取消/丢弃 |
| 权限不足 | 不渲染字段区与提交控件，卡片体替换为权限提示（lock 图标 + 文字） | canSubmit=false（RBAC 校验） |
| 流式填充中断 | 丢弃半填充骨架卡，替换为系统消息 + 重试（沿用上一条指令） | 流式连接断开（仅 UF-3 增量流式） |

### Validation Rules

- 必填字段为空时（含 `data-required-highlight` 属性的字段无值）提交按钮禁用并提示。
- 仅状态变更类操作且目标实体支持 available-transitions 端点（MainItem/SubItem/MilestoneMap/Milestone）时，提交前必须先通过预校验；创建操作、ProgressRecord、ItemPool、非状态变更类修改/分配不做预校验。
- 权限不足时不渲染可提交卡片，改为权限提示文字。
- **错实体防护**：高影响写操作（分配/状态变更）提交前卡片必须显示目标实体的 title + bizCode 供用户二次确认（防止错实体）。提交即生效，**不提供撤回**——误操作需由用户重新发起反向操作（如再次分配、再次状态变更）纠正。
- **并发写入合并语义**：用户直接编辑字段（onChange）与对话补充（异步 AI 增量变更）写入同一卡片 state 时，以**时间戳晚者胜出**（last-write-wins）合并；对话补充产生的增量变更在应用到卡片前先展示 diff 供用户确认，不静默覆盖正在被用户编辑的字段。
- **失败重试语义**：失败后重试必须重新运行 available-transitions 预校验（若适用）；卡片保持可编辑态，所有字段值保留；后端写操作应为幂等或事务性，失败不产生半成品副作用。**后端返回的参数校验错误按字段级就近展示**（出错字段红框 + 字段下方错误说明），不仅依赖顶部错误条；多次重试只保留最后一次错误的字段级提示，避免错误堆积；当前 turn 顶部小字标注"重试 N 次"，**历史会话回看此 turn 时不显示重试次数**（保留 transcript 简洁）。
- **字段控件类型规则**（强制，禁止裸文本输入时间或关联值）：所有字段控件类型由 [`entity-schemas.md`](./entity-schemas.md) 中各实体 schema 的 `control` 声明决定（同一份 schema 同时驱动 form 与 result 渲染）。**核心约束**：
  - `role: date` 字段 → DatePicker，禁止手敲日期字符串
  - `role: parent` / `assignee` / `submitter` / `team` / `priority` / `status` 字段 → Select 下拉，选项来自后端预加载的 Team 范围实体列表
  - `role: title` / `text` 字段 → Input / Textarea
  - `role: progress` 字段 → Number 输入（0–100）
  - `role: code` 字段 → readonly（自动生成）
- **diff 内联（无二级浮窗）**：对话补充产生的增量变更在**卡片体内**内联展示 diff 区（`.diff-inline`，accent-bg），确认动作由输入区选项组承载（应用/丢弃）；不弹二级浮窗，保持卡片交互扁平。
- **提交中 API 步回显**：用户点击提交后，trace（UF-8）末尾追加一行"⏳ 调用 {Entity} {Op} API…"步，API 返回后该步状态图标迁移为 ✓（成功）或 ✗（失败）；该步计入 trace 的步数与耗时统计，与流式期间的 plan/tool_call 步同等展示。
- **已提交/已丢弃 form 默认折叠**：表单进入这两种终态后默认折叠为单行 summary（含状态图标 + title），点击 summary 行可展开只读字段区（含原 form 的全部字段值）；展开/折叠态记忆于会话内（同一用户再次回到该 turn 时保持上次选择）。**不抽取 keyFields/diff 等结构化摘要**——表单消息本身就是当时提交的完整快照（消息不可变），上下文完整性由 transcript 自身维持。
- **AI 跟进消息触发**：表单提交成功（committed）后，AI 必须跟发一条普通的 `ai_text` 消息（自然语言），用一两句话陈述"做了什么"+ 引导下一轮（如"还需要加里程碑吗？"）；该消息由 AI 从自己已有的上下文（form payload、用户原指令）生成，不依赖额外结构化字段或快照。失败、丢弃态不发跟进消息（错误条/折叠态 summary 已是反馈）。

---

## UI Function 4: 查询结果卡片（查询操作）

### Placement

- **Mode**: existing-page
- **Target Page**: global — 渲染于 UF-2 聊天面板的消息区内
- **Position**: 作为消息流中的摘要 + 卡片列表

### Description

查询操作的返回结果：先展示摘要文字（如"你有 3 个 P0 事项"），再展示**内容丰富的实体卡片列表**——每张卡片直接展示该实体的核心字段（不只标题与编号），让用户**在聊天面板内就能看到必要信息**，无需跳转到详情页。

**核心原则：内容自包含**。用户在面板里就能回答"这个事项是什么状态、谁负责、什么时候截止、进度如何"等基本问题。点击卡片仍可跳详情页，但跳转不再是看基本信息的必要操作。

### User Interaction Flow

AI 返回查询结果 → 渲染摘要 + 内容丰富的卡片列表 → 用户在面板内浏览核心字段；如需更多细节或执行操作，点击卡片跳转到对应详情页（既有路由）。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 摘要文字 | string | AI/后端生成 | 如"你有 3 个 P0 事项" |
| 实体卡片列表 | List<EntityCard> | 查询 API 结果 | 每张含核心字段 + 跳转路由 |
| 进度统计（可选） | {completed, total, percent} | 查询 API 结果 | 里程碑/MainItem/SubItem 进度场景 |

**EntityCard 按实体类型携带的核心字段**（由 [`entity-schemas.md`](./entity-schemas.md) 中各实体 `result_slots` 决定渲染槽位）：

| 实体类型 | head | fields | meta | progress | text |
|---------|------|--------|------|----------|------|
| MainItem | title, code | priority, status, assignee | expectedEndDate, milestoneKey | 子任务完成率 | description |
| SubItem | title, code | status, assignee | parent, expectedEndDate | completion | achievement |
| Milestone | title, code | status | parent, expectedEndDate | 子事项完成率 | description |
| MilestoneMap | title, code | status, team | expectedEndDate, milestoneCount | 里程碑完成率 | description |
| ProgressRecord | subItem.title, code | — | createdAt | completion | achievement |
| ItemPool | title, code | priority, status, submitter | createdAt | — | background, expectedOutput |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 有结果 | 摘要 + 内容丰富的卡片列表（每张展示核心字段） | 查询命中 |
| 无结果 | "未找到匹配事项" + 建议 | 查询未命中 |

### Validation Rules

- 卡片跳转路由必须为 sitemap 已有路由（`/items/:mainItemId`、`/items/:mainItemId/sub/:subItemId`、`/milestones/:mapId`、`/item-pool`），不生成新路由。
- 单次查询最多展示 20 张卡片；超过 20 张时仅展示前 20 张并提示"结果过多，请缩小查询范围（如指定标题关键词或负责人）"。
- **字段展示规则**：每张卡片按 [`entity-schemas.md`](./entity-schemas.md) 中实体 `result_slots` 渲染——所有槽位（head/fields/meta/progress/text）的字段映射统一由 schema 驱动，渲染器与实体类型解耦。过长的文本字段（如 ItemPool 的 background/expectedOutput）在卡片内截断为 1–2 行，完整内容仍可点击跳详情页查看。卡片之间用边框分隔，hover 高亮整张卡片以提示可点击跳转。

---

## UI Function 5: 歧义消解卡片

### Placement

- **Mode**: existing-page
- **Target Page**: global — 渲染于 UF-2 聊天面板的消息区内
- **Position**: 作为消息流中的选择卡片

### Description

当实体模糊匹配命中多个候选时，列出所有候选实体（标题 + 编号）供用户选择，选定后继续后续流程。避免 AI 猜测导致操作错误对象。

### User Interaction Flow

AI 检测到多候选 → 推送歧义消解卡片列出候选 → 用户选择其一 → 系统以所选实体继续（推送对应的写操作卡片 UF-3 或查询结果 UF-4）。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 候选列表 | List<{bizKey, title, code, meta}> | 模糊匹配结果 | Team 范围内 |
| 原始意图 | {op, fields} | 暂存的 AI 解析结果 | 选定实体后复用 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 待选择 | 候选列表可选 | 多候选命中 |
| 已选择 | 标记所选 + 继续后续 | 用户选定 |

### Validation Rules

- 仅当候选 ≥ 2 时触发；候选唯一时直接精确匹配不展示此卡片。
- 候选列表限定在当前 Team 范围内。

---

## UI Function 6: 降级与超时提示

### Placement

- **Mode**: existing-page
- **Target Page**: global — 渲染于 UF-2 聊天面板内
- **Position**: 替代或追加到 AI 回复消息位置

### Description

当 AI 响应超时（>10s）或服务不可用时，展示提示并提供传统表单快捷入口，确保用户操作不被阻断。

### User Interaction Flow

AI 超 10 秒未返回 → 展示超时提示 + "去手动操作"快捷入口；AI 服务不可用时面板展示降级提示。用户点击快捷入口跳转到对应传统表单页面（如 `/item-pool`、`/items` 新增）。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 降级类型 | enum(timeout/unavailable) | 后端代理状态 | |
| 快捷入口路由 | route | 降级类型映射 | 如创建→`/items`、ItemPool→`/item-pool` |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 超时 | "AI 响应超时" + 手动入口 | 单次调用 >10s |
| 不可用 | "AI 暂不可用" + 手动入口 | AI 服务整体不可用 |

### Validation Rules

- 快捷入口路由必须为已有页面路由。
- 降级态不得阻塞用户使用传统表单。

---

## UI Function 7: 首次引导卡片

### Placement

- **Mode**: existing-page
- **Target Page**: global — 渲染于 UF-2 聊天面板首次展开时
- **Position**: 面板消息区首位（空会话时）

### Description

空会话时展示的能力引导卡片，说明 Copilot 能做什么（创建/查询/修改/分配四类操作）并给出示例指令。**首次展开面板与每次新建会话共用同一张完整能力卡片**——不做首次/再次区分，确保用户每次进入空会话都能看到 AI 能力边界与示例指令。用户开始输入后该卡片可收起。

### User Interaction Flow

用户首次展开面板 → 展示引导卡片（能力说明 + 示例指令）→ 用户点击示例指令填入输入框 或 自行输入 → 引导卡片收起。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 能力说明 | string | 静态文案 | 列出四类操作 |
| 示例指令 | List<string> | 静态文案 | 可点击填入输入框 |
| 本会话内已收起 | boolean | 会话内状态 | 用户点 X 或开始输入后置 true；本会话内不再展示，新建会话或重新展开面板时重置 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 空会话展示 | 完整能力卡片 + 示例 chips | 面板首次展开、或用户新建会话使消息区为空 |
| 本会话内已收起 | 不展示 | 用户点 X 或开始输入；本会话内有效，新建会话或重新展开时重置 |

### Validation Rules

- **空会话即展示**：面板首次展开 + 用户新建会话使消息区为空时，都展示同一张完整能力卡片（能力列表 + 示例 chips），不做首次/再次区分；用户开始输入或点 X 后本会话内收起。
- 示例指令点击后填入输入框但不自动发送，由用户确认。

---

## UI Function 8: Agent 过程追踪

### Placement

- **Mode**: existing-page
- **Target Page**: global — 渲染于 UF-2 聊天面板消息区内
- **Position**: 用户指令与最终卡片（UF-3/4）之间，按后端流式事件增量渲染

### Description

借鉴 Claude Code 等成熟 agent，把 AI 的思考、计划、工具调用（tool_call）步骤流式展示给用户，降低"黑盒"感、建立信任、便于调试。最终卡片在过程结束后渲染于其下方。

### User Interaction Flow

用户发送指令 → 面板流式追加：思考 → 计划步骤 → 操作步骤（每步含状态 ✓/✗ 与耗时）→ 过程结束在其下方渲染最终卡片 → 用户可点击追踪头部折叠/展开（折叠态记忆于会话内）。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 思考文本 | string | 流式事件 `thinking` | 斜体展示 |
| 计划步骤 | string[] | 流式事件 `plan_step` | 编号列表 |
| 操作步骤 | `{name,status,durationMs,detail}[]` | 流式事件 `tool_call` / `tool_result` | status: streaming/done/error |
| 摘要 | `{tokens,durationMs,costUsd}` | 后端汇总 | 折叠态仍可见 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 流式中 | 步骤逐条追加，末步闪烁光标 | 后端流式事件 |
| 已完成（成功） | 全部 ✓，**下方渲染下一步内容**（写操作 → UF-9 意图确认卡片；查询 → UF-4 结果卡片）；**自动折叠**为头部摘要（思考+计划+操作执行完成即折叠，约 600ms 延迟，让下一步内容成为视觉焦点）。UF-3 表单卡片需用户在 UF-9 确认后才渲染 | 流式结束且全部成功 |
| 步骤失败 | 该步 ✗ error；**不自动折叠**（保持展开让错误可见），中断后续步骤，不渲染下一步内容改为引导文字 | 工具调用失败 |
| 折叠 | 仅头部（步数 · 耗时 / 或失败图标 + 错误摘要） | 用户点击头部切换；**历史会话回看过往 turn 的 trace 默认折叠**，避免长 transcript 一打开就刷屏 |

### Validation Rules

- 首字节（思考出现）≤ 1s；计划可见 ≤ 2s；最终卡片 P95 < 5s。
- AI 服务不支持流式时，退化为单条"AI 思考中…"系统消息 + 最终卡片。
- 任一步骤失败 → 不渲染最终卡片，改为引导文字（如"匹配失败，请补充信息"）。
- **折叠规则（不对称）**：流式成功完成 → 约 600ms 后自动折叠为头部摘要；流式出现失败步骤 → 保持展开不自动折叠（错误可见性优先于简洁）；用户手动展开/折叠的状态记忆于会话内；**历史会话打开时，过往 turn 的 trace 一律默认折叠**，用户想看哪轮点哪轮。
- **与 UF-3 的衔接**：表单卡片（UF-3）进入"提交中"态时，trace 末尾追加一行"⏳ 调用 {Entity} {Op} API…"步，与流式期间的 plan/tool_call 步同等展示；API 返回后状态图标迁移为 ✓/✗。该步计入 trace 步数与耗时统计。

---

## UI Function 9: 意图回执 + 主动澄清（写操作前置 · 文本模式）

### Placement

- **Mode**: existing-page
- **Target Page**: global — 渲染于 UF-2 聊天面板的消息区内
- **Position**: 作为用户指令与 UF-3 表单卡片之间的若干条 AI 文本消息（位于 Agent 过程追踪 UF-8 之后、表单卡片 UF-3 之前）。**采用普通 AI 文本消息（`ai_text`）实现，不渲染卡片**——意图回执与主动澄清都是自然语言，文本最贴切；卡片结构化反而把意图拆碎成字段列表，违反"简约"原则。

### Description

写操作（创建/修改/分配）流程的**意图回执 + 主动澄清两阶段前置**——agent 解析用户指令后**先用自然语言简单回执意图**（不做字段列表），用户确认"理解正确"后再**主动询问模糊不清的内容**，循环问答直到所有必填信息收齐，**才渲染 UF-3 表单卡片**。**查询操作（UF-4）不经此流程**，trace 完成后直接渲染 result-card。

**两阶段**：

1. **意图回执**（单条 AI 文本消息）：用一两句话复述用户要做什么。
   - 例：「好的，我帮你创建一个 P1 主事项「完成用户认证模块」，分配给张三。」
   - 不罗列所有字段，只表达"我懂你要做什么"。
2. **主动澄清**（用户确认意图后，可选；仅当必填信息有缺失时触发）：列出还需补充的字段。
   - 例：「还需要补充以下信息：\n• 预期截止日期是？\n• 归到哪个里程碑下？（第一阶段 / 第二阶段）」
   - 一次列全部缺失项（不一次问一个，减少轮次）。
   - 用户用自然语言回答（如「下周五，第二阶段」）；agent 解析后填入 draft state，若仍有缺失继续追问；若全部收齐，发"信息已收集完整，请核对表单："+ 渲染 UF-3 表单卡片。

**核心价值**：在用户看到表单之前，先把意图对齐 + 必填信息收齐，让表单出现时就是"全部填好、可直接提交"的状态，减少用户在表单上来回编辑的成本。

### User Interaction Flow

**阶段 1 · 意图回执**：
agent 解析完成 → 渲染 Agent 过程追踪 UF-8（思考/计划/操作 ✓）→ 渲染意图回执文本消息 → 输入区切换为选项组（↑↓/Enter/←/Esc）：
- **「✓ 理解正确」** → 进入阶段 2（主动澄清，若需）或直接渲染 UF-3 表单（若用户原话已涵盖所有必填字段，跳过 Q&A）。
- **「✎ 我要调整」** → 输入区切回文本模式、textarea 聚焦并预填用户上一条指令**原文** → 用户编辑后重发 → agent 重新解析 → 新意图回执 append 到消息流。
- **「✗ 取消」** → 终止本次写操作。

**阶段 2 · 主动澄清**（仅当阶段 1 确认后仍存在必填缺失时触发）：
输入区切回**文本模式**（用户用自然语言回答）→ agent 解析回答、更新 draft state → 若仍有缺失，发新 AI 文本消息继续追问；若全部收齐，发"信息已收集完整，请核对表单："+ 渲染 UF-3 表单卡片。

**阶段 3 · 表单渲染**：
UF-3 表单卡片渲染时，所有必填字段已通过阶段 1-2 收齐，**无 `data-required-highlight` 高亮**；用户只需核对、可选编辑、提交。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 意图回执文本 | string | AI 生成（基于 opType + entityType + 已知字段） | 一两句自然语言 |
| 主动澄清问题列表 | string[] | AI 抽取结果与 schema required 对比 | 仅列缺失项；附 hint（如可选范围） |
| Draft 字段状态 | Map<fieldName, value> | 跨轮累积（用户回答持续补充） | 阶段 3 渲染表单的预填来源 |
| 用户原指令 | string | 上一条 user msg textContent | "我要调整"时预填输入框 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 意图回执 · 待确认 | 意图回执文本 + 输入区选项组（理解正确 / 我要调整 / 取消） | agent 解析完成，渲染于 UF-8 之后 |
| 主动澄清 · 待回答 | （前面的意图回执保留 transcript）+ 主动澄清文本消息（列缺失项）+ 输入区文本模式 | 用户选"理解正确"且必填缺失 |
| 主动澄清 · 继续追问 | 上一轮回答 + 新追问文本消息 + 输入区文本模式 | 用户回答后仍有缺失 |
| 信息收齐 · 渲染表单 | "信息已收集完整，请核对表单："文本消息 + UF-3 表单卡片（无 required-highlight）| 阶段 2 收齐所有必填 |
| 已调整 | 输入区聚焦预填用户原指令 | 用户选"我要调整"（任意阶段可触发） |
| 已取消 | 输入区切回文本；不渲染表单 | 用户选"取消" |

### Validation Rules

- **仅写操作触发**：UF-4 查询结果不经此流程，trace 完成后直接渲染 result-card。
- **意图回执必须简约**：一两句自然语言复述意图，**不罗列字段**；标题/字段列表/属性表等结构化形式禁止（违反"重点是用户意图"原则）。
- **主动澄清仅问缺失必填项**：可选空字段（如 `description`、`planStartDate`）不问，避免噪声；问题附 hint（如枚举值范围、示例格式）。
- **主动澄清一轮问全**：一次列出所有缺失项，用户可一条消息全部回答；不在缺失项多于 1 个时拆成多轮单问。
- **回答解析失败的处理**：用户回答模糊或不可解析时，agent 在文本消息中指出具体哪一项未明（如「"下周五"我理解了，但"第二阶段"是指标题含第二阶段的里程碑图吗？请明确 bizCode 或全名」），不静默猜测。
- **表单渲染条件**：UF-3 表单卡片仅在所有必填字段已收齐时渲染；表单出现即"全字段已填、无 required-highlight"，让用户专注于核对而非编辑。
- **"我要调整"语义**：点击后输入框预填用户上一条指令**原文**；用户编辑重发触发新一轮解析，原意图回执保留 transcript 作历史。
- **意图回执不触发 AI 跟进消息**：跟进消息仅由 UF-3 表单提交成功后触发（保持现有规则）。

---

## Page Composition

| Page | Type | UI Functions | Position Notes |
|------|------|-------------|----------------|
| 所有已认证路由（global overlay） | existing | UF-1, UF-2, UF-3, UF-4, UF-5, UF-6, UF-7, UF-8, UF-9 | 浮动气泡 UF-1 常驻；UF-2~UF-9 在气泡展开后于面板内渲染；UF-9（文本消息）位于 UF-8 之后、UF-3 表单卡片之前，由 1-3 条 AI 文本消息组成 |
| `/login` | existing | （无） | 未认证页面不显示 Copilot |

> 本功能不创建任何新页面（new-page 计数为 0）。所有 UI 均为挂载在现有页面之上的 overlay 组件。查询结果卡片（UF-4）与降级快捷入口（UF-6）跳转到的均为 sitemap 中已有路由。

---

## Accessibility Requirements

鉴于本功能目标用户包含非技术人员，必须满足以下可访问性要求（参照 WCAG 2.1 AA）：

### 焦点管理
- 用户点击浮动气泡展开聊天面板（UF-2）时，焦点自动移入输入框。
- 按 `Esc` 键焦点返回浮动气泡（UF-1），面板收起。
- 卡片（UF-3/UF-4/UF-5/UF-7）出现时，焦点不抢断（不自动移入卡片），用户可通过 `Tab` 键顺序到达。

### 键盘可操作性
- 所有卡片操作（编辑字段、选择候选、提交、重试）均可用键盘完成，无键盘陷阱（用户随时可 `Esc` 返回）。
- `Enter` 发送指令，`Shift+Enter` 换行；`Esc` 收起面板。

### ARIA 与语义
- 聊天消息历史区设置 `role="log"` 与 `aria-live="polite"`，新消息（用户/AI/系统）追加时由屏幕阅读器播报。
- 浮动气泡（UF-1）设置 `aria-label="AI 助手（展开/收起）"`；展开后面板设置 `role="dialog"` 与 `aria-label`。
- 卡片设置 `aria-label` 描述卡片类型（如"创建主事项卡片"）；错误态使用 `role="alert"`。
- Team 上下文指示器使用 `aria-live="polite"`，Team 切换时播报。

### 屏幕阅读器支持
- 所有状态变化（思考中/流式返回/错误/超时/成功）必须有对应的 `aria-live` 区域播报，不仅依赖视觉指示。
- 加载态需有文字标签（如 `aria-label="AI 思考中"`），不仅依赖动画。

### 对比度与可读性
- 文字与背景对比度 ≥ 4.5:1（正文）/ 3:1（大字）；高亮必填空字段使用对比度达标的边框色，不仅依赖颜色。

