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
- 页面导航前若有未提交卡片（UF-3/UF-5），弹出离开确认提示。

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
| 空（首次展开/新会话） | 欢迎语 + 首次引导卡片（UF-7） | 新会话 |
| 思考中 | 输入区禁用 + 三点跳动 | 指令已发送，等 AI |
| 流式返回 | Agent 过程追踪（UF-8）流式追加 | AI 流式返回（不支持流式则退化单条"思考中"） |
| 正常对话 | 消息历史 + 输入区可用 | AI 返回完成 |
| 选项组模式 | 输入区为键盘选项组（↑↓/Enter/←/Esc） | 存在待确认决策（提交/取消/撤回/diff/歧义/降级） |
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
- 多卡片可共存，但 AI 请求串行（思考态全局唯一）；切换会话时前一会话待确认卡片与撤回窗口保留。

---

## UI Function 3: 预填表单卡片（写操作）

### Placement

- **Mode**: existing-page
- **Target Page**: global — 渲染于 UF-2 聊天面板的消息区内
- **Position**: 作为消息流中的一条卡片消息

### Description

写操作（创建/修改/分配）的确认卡片。AI 推送预填字段，必填且无法推导的字段留空并高亮。用户可直接编辑卡片字段，或继续对话补充——两种方式均写入同一份卡片状态（卡片为唯一数据源）。提交前经状态机/权限预校验。

### User Interaction Flow

AI 推送卡片 → 用户直接编辑字段（onChange 更新卡片 state）或对话补充（后端解析增量变更后更新卡片 state）→ 用户点击提交：
- 若为状态变更类操作且目标实体支持 available-transitions 端点（MainItem/SubItem/MilestoneMap/Milestone）→ 提交前调用 available-transitions 预校验 → 通过则调用现有 API；不通过则卡片内显示错误（含 `validTransitions` 合法目标状态）。
- 若为创建操作、ProgressRecord、ItemPool、或非状态变更类修改/分配 → 不做预校验，直接调用现有 API，依赖后端 RBAC + 状态机校验；失败由后端返回错误信息在卡片内展示。
- 高影响写操作（分配/状态变更）提交前，卡片必须显示目标实体的 title + bizCode 供用户二次确认（防止错实体）。
结果反馈到聊天界面。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 卡片类型 | enum(create/update/assign) | AI 意图 | 决定字段集 |
| 目标实体 | enum(MainItem/SubItem/Milestone/MilestoneMap/ProgressRecord/ItemPool) | AI 实体抽取 | 决定 schema |
| 字段集 | Map<fieldName, {value, required, derived}> | AI 抽取 + 实体 schema | required 且无值字段高亮 |
| 父实体引用 | bizKey | 页面上下文 / 模糊匹配 | SubItem 需 parent MainItem；Milestone 需 parent MilestoneMap |
| 校验错误 | string | available-transitions 预校验 | 不合法时卡片内展示 |
| undoAvailable | boolean | 操作类型判定 | 仅可逆操作（分配 / 非 terminal 状态变更）为 true；terminal 转移为 false |
| undoDeadline | timestamp | 提交成功时间 + 5min | 到期后撤回按钮隐藏，卡片显示"撤回窗口已过期" |
| previousValue | {assignee?, status?} | 提交前快照 | 撤回时恢复的原值；分配存 assignee，状态变更存 status |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 预填展示 | 字段值 + 必填空字段高亮 + 高影响操作的目标实体 title/bizCode 二次确认区 | AI 推送完成 |
| 编辑中 | 字段可编辑 | 用户聚焦字段 |
| 校验失败 | 错误说明 + `validTransitions` 合法目标状态列表 | available-transitions 返回不合法（仅状态变更类） |
| 提交中 | 提交按钮禁用 + 加载 | 用户点击提交 |
| 成功 | 成功反馈 + 实体跳转入口 + （分配/状态变更类）5 分钟撤回按钮 | API 返回成功 |
| 权限不足 | 不渲染字段区与提交控件，卡片体替换为权限提示（lock 图标 + 文字） | canSubmit=false（RBAC 校验） |
| 流式填充中断 | 丢弃半填充骨架卡，替换为系统消息 + 重试（沿用上一条指令） | 流式连接断开（仅 UF-3 增量流式） |
| 失败 | 错误信息 + 重试（字段保留可编辑） | API 返回失败 |

### Validation Rules

- 必填字段为空时（含 `data-required-highlight` 属性的字段无值）提交按钮禁用并提示。
- 仅状态变更类操作且目标实体支持 available-transitions 端点（MainItem/SubItem/MilestoneMap/Milestone）时，提交前必须先通过预校验；创建操作、ProgressRecord、ItemPool、非状态变更类修改/分配不做预校验。
- 权限不足时不渲染可提交卡片，改为权限提示文字。
- **错实体防护 + 5 分钟撤回**：高影响写操作（分配/状态变更）提交前卡片必须显示目标实体的 title + bizCode 供用户二次确认。提交成功后，对**可逆操作**（分配、非 terminal 状态变更）提供 5 分钟撤回窗口；对**不可逆操作**（转入 terminal `completed`/`cancelled` 的状态变更）不提供撤回，卡片明确标注"该操作不可撤回"。
  - 撤回语义：分配类 → 将 assignee 恢复为 `previousValue.assignee`；状态变更类 → 将状态恢复为 `previousValue.status`，**恢复前重新调用 available-transitions 校验**（若当前状态已不允许回到原状态，撤回失败并提示当前合法目标状态）。
  - 撤回窗口边界：`undoDeadline` = 提交成功时间 + 5 分钟；到期后撤回按钮隐藏，卡片显示"撤回窗口已过期"。
  - 撤回窗口范围：窗口在**当前会话内有效**；面板收起 / 同会话页面导航不中断窗口；会话结束（关闭浏览器/登出）后撤回失效（跨会话持久化不在 v1 范围）。
  - 撤回不依赖 AI：撤回调用现有实体 API 的反向操作，AI 服务不可用时不影响撤回执行。
  - 撤回唯一性：每次操作仅允许一次撤回；撤回后不可经 Copilot 重做，需重新发起正向操作。
- **并发写入合并语义**：用户直接编辑字段（onChange）与对话补充（异步 AI 增量变更）写入同一卡片 state 时，以**时间戳晚者胜出**（last-write-wins）合并；对话补充产生的增量变更在应用到卡片前先展示 diff 供用户确认，不静默覆盖正在被用户编辑的字段。
- **失败重试语义**：失败后重试必须重新运行 available-transitions 预校验（若适用）；卡片保持可编辑态，所有字段值保留；后端写操作应为幂等或事务性，失败不产生半成品副作用。
- **字段控件类型规则**（强制，禁止裸文本输入时间或关联值）：
  - 日期/时间字段（planStartDate、expectedEndDate 等）→ **日期选择组件**（`<input type="date">` 或 DatePicker），禁止手敲日期字符串。
  - 关联/引用字段（milestoneKey 里程碑、assignee 负责人、parent MainItem、teamKey 等）→ **Select 下拉组件**，选项来自当前 Team 范围内实体列表（后端预加载），禁止自由文本。
  - 纯文本字段（title、description、achievement 等）→ Input/Textarea。
- **diff 内联（无二级浮窗）**：对话补充产生的增量变更在**卡片体内**内联展示 diff 区（`.diff-inline`，accent-bg），确认动作由输入区选项组承载（应用/丢弃）；不弹二级浮窗，避免与离开确认 Dialog 的 z-index 叠压。

---

## UI Function 4: 查询结果卡片（查询操作）

### Placement

- **Mode**: existing-page
- **Target Page**: global — 渲染于 UF-2 聊天面板的消息区内
- **Position**: 作为消息流中的摘要 + 卡片列表

### Description

查询操作的返回结果：先展示摘要文字（如"你有 3 个 P0 事项"），再展示可点击的实体卡片列表，点击跳转到对应详情页。

### User Interaction Flow

AI 返回查询结果 → 渲染摘要 + 卡片列表 → 用户点击某张卡片 → 跳转到该实体详情页（既有路由）。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 摘要文字 | string | AI/后端生成 | 如"你有 3 个 P0 事项" |
| 实体卡片列表 | List<{bizKey, title, route, meta}> | 查询 API 结果 | 每张含跳转路由 |
| 进度统计（可选） | {completed, total, percent} | 查询 API 结果 | 里程碑进度场景 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 有结果 | 摘要 + 卡片列表 | 查询命中 |
| 无结果 | "未找到匹配事项" + 建议 | 查询未命中 |

### Validation Rules

- 卡片跳转路由必须为 sitemap 已有路由（`/items/:mainItemId`、`/items/:mainItemId/sub/:subItemId`、`/milestones/:mapId`、`/item-pool`），不生成新路由。
- 单次查询最多展示 20 张卡片；超过 20 张时仅展示前 20 张并提示"结果过多，请缩小查询范围（如指定标题关键词或负责人）"。

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

用户首次展开聊天面板时展示的引导卡片，说明 Copilot 能做什么（创建/查询/修改/分配四类操作）并给出示例指令。用户开始输入后该卡片可收起。

### User Interaction Flow

用户首次展开面板 → 展示引导卡片（能力说明 + 示例指令）→ 用户点击示例指令填入输入框 或 自行输入 → 引导卡片收起。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 能力说明 | string | 静态文案 | 列出四类操作 |
| 示例指令 | List<string> | 静态文案 | 可点击填入输入框 |
| 已展示标记 | boolean | 本地存储 | 首次展开后标记，不再展示 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 首次展示 | 引导卡片 + 示例 | 用户首次展开 |
| 已引导 | 不展示 | 本地标记已存在 |

### Validation Rules

- 仅首次展开展示；用户清除本地存储或换设备时重新展示。
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
| 已完成 | 全部 ✓，下方渲染最终卡片；**自动折叠**（思考+计划+操作执行完成即折叠为头部摘要，让最终卡片成为焦点） | 流式结束 |
| 折叠 | 仅头部（步数 · 耗时） | 用户点击头部 |
| 步骤失败 | 该步 ✗ error，中断后续，不渲染最终卡片改为引导 | 工具调用失败 |

### Validation Rules

- 首字节（思考出现）≤ 1s；计划可见 ≤ 2s；最终卡片 P95 < 5s。
- AI 服务不支持流式时，退化为单条"AI 思考中…"系统消息 + 最终卡片。
- 任一步骤失败 → 不渲染最终卡片，改为引导文字（如"匹配失败，请补充信息"）。

---

## Page Composition

| Page | Type | UI Functions | Position Notes |
|------|------|-------------|----------------|
| 所有已认证路由（global overlay） | existing | UF-1, UF-2, UF-3, UF-4, UF-5, UF-6, UF-7, UF-8 | 浮动气泡 UF-1 常驻；UF-2~UF-8 在气泡展开后于面板内渲染 |
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
- 所有卡片操作（编辑字段、选择候选、提交、重试、撤回）均可用键盘完成，无键盘陷阱（用户随时可 `Esc` 返回）。
- `Enter` 发送指令，`Shift+Enter` 换行；`Esc` 收起面板。

### ARIA 与语义
- 聊天消息历史区设置 `role="log"` 与 `aria-live="polite"`，新消息（用户/AI/系统）追加时由屏幕阅读器播报。
- 浮动气泡（UF-1）设置 `aria-label="AI 助手（展开/收起）"`；展开后面板设置 `role="dialog"` 与 `aria-label`。
- 卡片设置 `aria-label` 描述卡片类型（如"创建主事项卡片"）；错误态使用 `role="alert"`。
- Team 上下文指示器使用 `aria-live="polite"`，Team 切换时播报。

### 屏幕阅读器支持
- 所有状态变化（思考中/流式返回/错误/超时/成功/撤回）必须有对应的 `aria-live` 区域播报，不仅依赖视觉指示。
- 加载态需有文字标签（如 `aria-label="AI 思考中"`），不仅依赖动画。

### 对比度与可读性
- 文字与背景对比度 ≥ 4.5:1（正文）/ 3:1（大字）；高亮必填空字段使用对比度达标的边框色，不仅依赖颜色。

