---
created: "2026-06-24"
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: AI Copilot 对话助手

> 设计层：定义界面"长什么样、怎么表现"，与 PRD 的"做什么"分离。
> 设计系统沿用项目既有 `DESIGN.md`（PM Work Tracker），保证 Copilot 与现有 15 个页面视觉一致。

## Design System

> 来源：项目根 `DESIGN.md`（Custom / Tailwind CSS v4 + Radix UI + CVA）。Copilot 作为全局 overlay，必须复用下列 tokens 与组件规格，不得引入新色板或新字号。

### 视觉基调

紧凑、信息密度优先的项目管理界面。蓝色主色 + slate 中性色；白卡片浮于冷灰页面背景上，以边框（非阴影）分隔。13px 正文，密集表格。整体接近开发者工具而非消费应用。

### 色板（必须使用 token，禁止硬编码 Tailwind 色类）

| 角色 | 值 | 用途 |
|------|----|------|
| Background | `#f8fafc` | 页面背景、表头、hover |
| Surface | `#ffffff` | 卡片、侧边栏、输入框、弹窗 |
| Border | `#e2e8f0` | 卡片边框、分隔线 |
| Border Dark | `#cbd5e1` | 输入框边框、次级按钮边框 |
| Text Primary | `#0f172a` | 标题、正文 |
| Text Secondary | `#475569` | 描述、表格单元格 |
| Text Tertiary | `#94a3b8` | 占位符、次级标签 |
| Accent | `#2563eb` | 主按钮、激活态、品牌 |
| Accent Light | `#3b82f6` | 进度填充、链接 |
| Accent Hover | `#1d4ed8` | 主按钮 hover |
| Accent BG | `#eff6ff` | 激活背景、淡填充 |
| Accent Ring | `#bfdbfe` | 焦点环 |
| Success | `#3b82f6`（蓝，非绿） | 完成/成功态 |
| Warning | `#d97706` | 警告/挂起（图标、大字标题，≥3:1） |
| Error | `#dc2626` | 错误/阻断 |

#### Status variants（语义化背景/文字配对）

| 角色 | 值 | 用途 |
|------|----|------|
| Success BG | `#eff6ff` | 成功背景、淡填充 |
| Success Text | `#1d4ed8` | 成功正文（深蓝，≥4.5:1 on success-bg） |
| Warning BG | `#fffbeb` | 警告背景（超时/降级提示卡） |
| Warning Text | `#92400e` | 警告正文（深棕，≥4.5:1 on warning-bg；正文用此） |
| Warning Title | `#d97706` | 警告图标/大字标题（≥3:1 large-text only） |
| Error BG | `#fef2f2` | 错误背景（校验错误条、失败提示） |
| Error Text | `#991b1b` | 错误正文（深红，≥4.5:1 on error-bg） |

### 字体（系统字体栈）

`-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif`

| 角色 | 字号 | 字重 | 用途 |
|------|------|------|------|
| 组件文本 | 13px | 500 | 按钮、输入框、表格单元格、聊天气泡正文 |
| 正文/表单标签 | 14px | 500 | 卡片字段标签、通用文本 |
| 区块标题 | 14px | 500 | 卡片小节标题 |
| 徽章/说明 | 12px | 500 | 状态徽章、上下文提示、时间戳 |
| 表头 | 12px | 500 | 大写 tracking-wider |

### 组件规格（直接复用）

- **按钮**：Primary bg `#2563eb`/白字/rounded-lg(8px)/h-10 px-4/hover `#1d4ed8`；Secondary 白底/border `#cbd5e1`；Ghost 透明/text-secondary/hover bg-bg-alt；Danger 透明/border error-text/40。sm=h-8 px-3 text-xs。过渡 150ms。
- **卡片**：bg `#fff`，1px solid `#e2e8f0`，rounded-xl(12px)，shadow level-1 `0 1px 2px 0 rgb(0 0 0 / 0.05)`。
- **输入框**：bg `#fff`，1px solid `#cbd5e1`，rounded-md(6px)，h-10 px-3，13px；focus border `#3b82f6` + ring-2 `#bfdbfe`；disabled opacity-50。
- **徽章**：inline-flex，rounded-full，px-2.5 py-0.5，text-xs 500；按 success/warning/error 语义配色。
- **弹窗/抽屉**：overlay black/50；内容 rounded-xl，shadow level-3 `0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)`。

### 深度层级

| 层 | 阴影 | Copilot 用途 |
|----|------|--------------|
| 1 | `0 1px 2px 0 rgb(0 0 0 / 0.05)` | 卡片消息、输入框 |
| 2 | `0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)` | 浮动气泡、tooltip、下拉 |
| 3 | `0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)` | 聊天面板（抽屉） |

### 签名规则（必须遵守）

1. **蓝=成功**：完成/成功态用蓝 `#3b82f6`，不用绿。
2. **13px 密度**：聊天气泡正文 13px，与表格/表单一致。
3. **边框分隔**：卡片用边框分隔，不靠重阴影。
4. **token 化**：所有颜色走 token，ESLint 禁止硬编码 Tailwind 色类。
5. **桌面优先**：不优化 <768px。

---

## Schema 驱动渲染（UF-3 表单卡片 + UF-4 结果卡片共用）

> 6 个实体（MainItem / SubItem / Milestone / MilestoneMap / ProgressRecord / ItemPool）的字段渲染统一由 [`../prd/entity-schemas.md`](../prd/entity-schemas.md) 中各实体的 schema 驱动。**同一份 schema 同时驱动 form 与 result 两种渲染模式**，避免双份字段定义导致错配。

### Role → 槽位映射（result 模式）

| Role | 槽位 | 控件（form） | 视觉 |
|------|------|-------------|------|
| `title` | head 行左 | Input | 13px 500 text-primary |
| `code` | head 行右 | readonly | 11px text-tertiary |
| `priority` | fields 行 | Select P1/P2/P3 | badge（error/warning/neutral） |
| `status` | fields 行 | Select（实体特定状态集） | badge（success/neutral/warning） |
| `assignee` | fields 行 | Select teamMembers | 👤 {name} |
| `submitter` | fields 行 | Select currentUser（默认） | 👤 {name} |
| `date` | meta 行 | DatePicker | 📅 {label} {yyyy/MM/dd} |
| `parent` | meta 行 | Select parent entities | 📁 {label} {parent title} |
| `team` | meta 行 | Select userTeams | 👥 {teamName} |
| `progress` | progress 条 | Number 0–100 | progress bar + label |
| `text` | text 块 | Textarea | 12px text-secondary（截断 1–2 行） |

### 渲染器 API（前端）

```ts
renderEntityCard(entityType: EntityType, mode: 'form' | 'result', data: EntityData): string
```

- `entityType`：6 实体之一，查 `ENTITY_SCHEMAS` 表得 schema
- `mode='form'`：遍历 `schema.fields` → 按 `control` 渲染输入控件 → required 加 `*` 高亮、derived 加"AI 推断"标记
- `mode='result'`：遍历 `schema.result_slots` → 按 role 把字段塞进 head/fields/meta/progress/text 五个槽位 → 拼成 result-card

### 新增实体的扩展流程

1. 在 [`entity-schemas.md`](../prd/entity-schemas.md) 新增实体 schema 段
2. 前端 `ENTITY_SCHEMAS` 加 entry（与文档一致）
3. **渲染器零改动**——form/result 自动支持

### 与业务逻辑的边界

Schema 只管"长什么样、有哪些字段"；状态机转移、RBAC、跨字段校验仍由后端服务层负责，前端通过现有 API 复用。详见 [`entity-schemas.md` 与业务逻辑的边界](../prd/entity-schemas.md#与业务逻辑的边界)。

---

## Component 1: 浮动气泡（Copilot 入口）

### Placement

- **Mode**: existing-page
- **Target**: global — 所有已认证路由 overlay（不含 `/login`）
- **Position**: `position: fixed`，视口右下角，距右边缘 24px、距下边缘 24px，`z-index: 50`，悬浮于主内容之上，不遮挡现有右下角主操作（与现有页面元素避让）

### Layout Structure

- 圆形按钮，56×56px，rounded-full，bg accent `#2563eb`，shadow level-2，hover bg `#1d4ed8`（150ms 过渡）
- 内含图标：消息/气泡 SVG，20px，白色
- 首次引导徽章：右上角红点，10×10px，bg error `#dc2626`，border-2 white，绝对定位 top-0 right-0
- AI 不可用态：气泡降为 border-dark、icon 灰色（text-tertiary）、cursor not-allowed
- 可拖拽：`cursor: grab`（拖拽中 `cursor: grabbing`），拖拽后位置存会话内 state，限制在视口范围内（不超出边缘）
- **面板展开时气泡隐藏**（`display:none`，非 opacity），避免与面板（right:0、宽 630px、z-50）重叠；面板收起后气泡按上次记录位置恢复显示。收起由 Esc 或点面板外区域触发（面板无关闭按钮）；面板展开期间无需"点击气泡收起"。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 默认 | accent 圆形气泡 + 白色消息图标 | 点击展开面板（UF-2） |
| 首次引导 | 气泡 + 右上红点徽章 | 首次访问未展开过；用户首次展开后徽章消除（本地标记） |
| AI 不可用 | 灰色 icon、border-dark 描边 | 点击仍可展开面板，面板内展示降级提示（UF-6） |
| 后台活动（面板收起态） | 气泡右上 activity-badge（error 红、带未读数或圆点、`bubble-pulse` 1.6s 脉冲动画） | 面板收起期间 agent 返回消息或需要用户确认时显示；用户展开面板即消除 |
| 隐藏 | 不渲染 | feature flag 关闭时整体不渲染（非仅隐藏） |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击气泡 | 展开聊天面板 UF-2 | 面板从右滑入；焦点移入输入框；气泡徽章（若有）消除；气泡隐藏 |
| 点击气泡（AI 不可用态） | 展开面板 | 面板内消息区首条渲染降级提示卡片（UF-6，fallbackType=unavailable）；输入框禁用并 placeholder 改为"AI 暂不可用，请稍后再试或使用下方入口"；快捷入口可见 |
| 面板收起时收到后台活动 | 显示 activity-badge | badge pulse 提示有新消息/待确认；点击气泡展开面板后消除 badge，焦点定位到新消息或待确认卡片 |
| 拖拽气泡 | 改变位置 | 会话内保持新位置；松手吸附到最近边缘 24px；拖拽位移 <4px 视为点击（避免误触发） |
| hover | — | bg `#1d4ed8`，150ms |
| focus（键盘） | — | ring-2 accent-ring `#bfdbfe` |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 徽章红点 | hasUnseenBadge | 本地存储（首次标记） |
| 气泡位置 | bubblePos {x,y} | 会话内状态 |
| 可见性 | featureFlagVisible | 后端配置 |
| AI 状态 | aiAvailable | 后端代理健康状态 |
| 活动徽章 | activityCount | 后端推送（面板收起期间累计的未读消息/待确认数）；展开面板后清零 |

---

## Component 2: 聊天面板（对话主界面）

### Placement

- **Mode**: existing-page
- **Target**: global overlay（同 UF-1 所有已认证路由）
- **Position**: 右侧抽屉，`position: fixed`，right: 0，top: 0，bottom: 0，**默认宽 630px**（`--panel-w`），`z-index: 50`，从右边缘滑入（translateX 动画 200ms），shadow level-3。不阻塞主页面（主页可滚动/点击）。展开时不添加全屏遮罩（保持主页面可操作）。
- **可拖拽调节宽度**：面板左边缘有 6px 拖拽手柄（`.panel-resizer`，`cursor: col-resize`），用户向左拖动增宽、向右拖动缩窄，宽度限制区间 **420px–960px**（`--panel-min-w` / `--panel-max-w`）；松手时宽度写入 `sessionStorage`（key `copilotPanelW`），同会话内 remount 后恢复。拖拽位移 <4px 不触发宽度变更（避免误触）。

### Layout Structure

垂直三段式（flex column，全高）：

1. **面板头**（h-auto，border-b `#e2e8f0`，px-4 py-3，flex row、align-center、justify-between，三段式）：
   - **左（header-left）**：Team 上下文徽章（accent-bg，text accent-hover，rounded-full，px-2.5 py-0.5，text-xs：`Team: {teamName}`）。
   - **中（header-center，flex:1 居中）**：当前会话标题（session-title，14px 500 text-primary，取自会话首条指令摘要或"新会话"，超长省略号截断、`title` 属性补全）。
   - **右（header-actions，gap 4px）**：**开启新会话按钮**（Ghost icon，`+` 图标，新建空会话）+ **历史会话按钮**（Ghost icon，时钟/列表图标，切换到「会话列表视图」）。
   - **无关闭按钮**：面板通过 `Esc` 或点面板外区域收起（见 Interactions）。
   - **收起 ≠ 中断会话**：收起仅隐藏面板，会话后台继续；agent 返回消息或需用户确认时，浮动气泡（UF-1）显示活动徽章（见 Component 1）。
2. **消息区**（flex-1，overflow-y-auto，p-4，gap-3 垂直堆叠）：
   - 用户消息：右对齐，max-w 85%，气泡 bg accent `#2563eb`、白字、rounded-xl（右下角直角）、px-3 py-2、13px
   - AI 消息：左对齐，max-w 90%，卡片式（surface bg、border `#e2e8f0`、rounded-xl、px-3 py-2、13px text-primary）；AI 回复含卡片时渲染 UF-3/4/5/7
   - 系统消息：居中，text-tertiary 12px（如"AI 思考中…"、超时提示、降级提示）
   - 时间戳：**每条消息（用户气泡、AI 文字、AI 卡片）上方**统一展示 `yyyy/MM/dd HH:mm` 格式时间戳，text-tertiary 11px、margin-bottom 4px；用户消息 ts 右对齐、AI 消息 ts 左对齐（与各自气泡对齐方向一致）
3. **输入区**（border-t `#e2e8f0`，bg surface，p-3）：
   - textarea（文本模式）：`min-height: 120px`、`max-height: 280px`、**按内容自动增高**（`resize: none`，JS 按scrollHeight 在 120–280px 区间自适应）、达到 280px 后 `overflow-y: auto` 出现滚动条、rounded-md、border border-dark、px-3 py-2、13px、placeholder text-tertiary"描述你想做的事…"；focus border accent + ring accent-ring。**单次输入上限 1000 字符**，超出截断并提示。
   > **与全局约定一致**：`min-h 120px` 对齐 `frontend-components.md` TECH-frontend-001 的 `min-h-[120px]`（不再偏离）；上限 280px 为 Copilot 面板内的增高上限，超出滚动。
   - 字符计数：右下角 text-tertiary 11px `{n}/1000`，超 1000 红色
   - 发送按钮（文本模式）：Primary sm，h-8 px-3，右下；Enter 发送，Shift+Enter 换行；thinking 态禁用 + 加载点

#### 输入区双模式（文本 ↔ 选项组，键盘优先）

输入区是 Copilot 的**唯一确认入口**——卡片本身不弹二级浮窗，所有"需要用户确认/选择"的操作都把输入区从「文本模式」切换为「选项组模式」，用户用键盘完成，无需鼠标：

- **文本模式（默认）**：即上述 textarea，用户输入自然语言指令。
- **选项组模式（有待确认时触发）**：textarea 隐藏，输入区渲染为选项列表（`.option-list`）。触发场景：写卡片待提交（提交/编辑字段/取消）、校验失败（修改/取消）、权限不足（换目标/取消）、对话补充 diff（应用/丢弃）、歧义消解（候选实体即选项）、超时降级（手动入口/重试）。
  - **键盘**：`↑`/`↓` 在选项间移动高亮（循环），`Enter` 确认当前高亮项，`←`（左方向键）返回文本输入模式，`Esc` 收起面板；选项过多时 `scrollIntoView({block:'nearest'})` 跟随。
  - **鼠标**：点击选项 = **直接选中并确认**该选项（事件 `stopPropagation`，不触发"面板外点击收起"，面板保持展开）。
  - **视觉**：每项 `.option` 14px 行高、border；高亮项 border-2 accent + accent-bg + text accent-hover；右侧 `kbd` 提示键位。
  - **提示行**：选项组顶部固定 `.options-hint`："↑↓ 选择 · Enter 确认 · ← 返回输入 · Esc 收起"。
- **阻塞语义**：选项组模式期间不接受新自由文本（textarea 隐藏），迫使用户先解决当前确认，避免上下文错乱；`Esc` 可退回文本模式重新描述。

#### 面板双视图（对话视图 ↔ 会话列表视图）

面板内容区有两个互斥视图，由头部「历史会话」按钮切换：

- **对话视图（chat-view，默认）**：即上述消息区 + 输入区，展示当前活动会话。
- **会话列表视图（session-list-view）**：展示历史会话列表，每项含会话标题、首条预览、时间、消息数、待确认标记；当前活动会话 accent-bg 高亮。**列表视图无独立标题栏**——打开列表视图时，**面板头中间标题切换为"历史会话"**（替代会话标题）；新建会话由头部右侧「+」图标按钮承担（不在列表内重复放按钮）。
  - 点某会话 → 切换为活动会话（加载其消息、恢复待确认状态、更新头部会话标题），返回对话视图，标题恢复为该会话标题。
  - 头部「+」新建会话：**若当前已是新空会话（无消息），不重复创建**，直接返回对话视图；否则创建空会话为活动会话，展示首次引导（UF-7）。
- **多会话决策（不并行）**：面板同一时刻只有**一个活动会话**；**不支持同时打开多个会话**——并行多会话需多窗口/标签，与「单全局 overlay」模型冲突，且增加上下文切换负担。用户通过会话列表快速切换活动会话；切换时前一会话状态保留（待确认卡片）。
- **会话标题生成**：新会话默认"新会话"；用户发送首条指令后，由首条指令摘要自动生成（截断 16 字，如"创建用户认证模块…"）；列表视图打开时头部中间临时显示"历史会话"，关闭后恢复会话标题；重命名不在 v1 范围。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 空（首次展开） | 欢迎语 + 首次引导卡片 UF-7 | 无消息历史 |
| 思考中 | 输入区禁用；消息区末尾三点跳动"AI 思考中…" | 指令已发送，等 AI；可转 流式返回 / 错误 / 超时 |
| 流式返回 | 卡片骨架（border-dashed 占位）先显示，字段增量填充 | AI 流式返回；可转 正常对话 / 流式中断 |
| 正常对话 | 完整消息历史 + 输入框可用 | AI 返回完成 |
| 错误 | 系统消息红色文字 + "重试"Ghost 按钮 | 后端代理返回错误 |
| 超时 | 系统消息"AI 响应超时"+ 降级提示 UF-6 | AI >10s 未返回 |
| 发送阻断（Team 缺失） | 发送按钮 disabled；输入区上方 inline notice warning-text `#92400e`"未识别到 Team 上下文，请先进入某个 Team 页面" | teamCtxMissing=true |
| 发送阻断（轮次上限） | 发送按钮 disabled；inline notice"会话已达 50 轮上限，请新建会话"+「新建会话」Ghost 按钮 | sessionRoundCount ≥50 |
| 流式中断 | 半填充骨架卡丢弃；系统消息"AI 响应中断，点击重试"+ 重试 Ghost 按钮（沿用上一条指令） | 流式连接中断 |
| Team 切换（in-flight） | in-flight/pending 卡片冻结（opacity-60 + 锁图标，不可提交）；inline warning notice（warning-text）"Team 已切换至 {newName}，当前卡片基于原 Team {oldName}" | teamChangedMidFlight 非空（teamCtx.bizKey 变化且存在 pending/in-flight 卡片） |
| 收起 | translateX(100%) 隐藏 | Esc / 点面板外区域（气泡恢复显示，后台会话继续） |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 输入 + Enter / 点发送 | 发送指令 | 用户气泡追加；切思考态；输入清空。前置校验：teamCtxMissing 或 sessionRoundCount≥50 时阻断发送 |
| 输入超 1000 字符 | 截断至 1000 | 计数红色 + inline toast"已截断至 1000 字符"（truncationNotice=true，3s 后自消） |
| 发送阻断（Team 缺失） | 不发送；展示 inline notice | 按钮保持 disabled，notice 内「前往 Team 页」Ghost 跳转 |
| 发送阻断（轮次上限） | 不发送；展示 inline notice | 「新建会话」Ghost 重置会话与 sessionRoundCount |
| Team 切换（teamChangedMidFlight） | 冻结 in-flight 卡片（不可提交）；弹 inline warning notice"Team 已切换至 {newName}，当前卡片基于原 Team {oldName}" | 用户可选「按原 Team 继续」（解冻卡片）/「丢弃卡片」（移除）；新指令按新 TeamCtx 处理 |
| Esc | 收起面板 | 焦点回气泡（无关闭按钮；Esc 或点面板外为收起入口） |
| 点击面板外区域（主页面） | 收起面板（不中断会话） | 与 Esc 等价；会话后台继续，有新活动时气泡提示 |
| 消息区滚动 | 自动滚到底（新消息追加时） | — |
| 主页面导航（含未提交卡片） | 直接放行，不弹离开确认 | 未提交卡片草稿会丢失，用户主动重新发起即可（不强制挽留） |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| Team 徽章 | teamCtx {bizKey,name} | 当前页面路由/全局状态 |
| 消息列表 | messages[{role,content,cardRef,ts}] | 会话内状态（≤50 轮） |
| 输入文本 | inputText | 用户输入 |
| 发送状态 | sendState enum | 后端代理响应 |
| 字符计数 | inputLen | inputText 长度 |
| Team 缺失阻断 | teamCtxMissing | 路由解析 teamBizKey 为空时 true |
| 轮次计数 | sessionRoundCount | 当前会话已发送指令数（上限 50） |
| 截断提示 | truncationNotice | inputText 被 500 截断时置 true，3s 自动复位 |
| 流式中断 | streamInterrupted | 流式连接断开时 true，触发骨架卡丢弃 |
| 会话内 Team 切换 | teamChangedMidFlight {oldName,newName} | 检测 teamCtx.bizKey 变化且存在 in-flight/pending 卡片时填充 |

#### 并发卡片策略

- **多卡片共存**：消息流中可同时存在多张 UF-3/UF-5 卡片（用户可连续发起多个操作），每张卡片 state 独立、互不覆盖。
- **AI 请求串行 + 用户串行发送**：同一会话同一时刻仅允许一个 in-flight AI 请求（思考态全局唯一）；用户在 AI 处理上一条指令期间（思考态/流式态）**不能发送新指令**——输入框与发送按钮同时禁用（`applyStateInput` 中 `cfg.disabled=true` 时两者都 disable），输入区上方 inline 提示"AI 处理中，请稍候…"；待上一轮 AI 返回后恢复可发送。**单次只能发送一条用户消息**，避免并发 prompt 组装冲突与上下文错乱（替代原"入队等待"模型）。
- **Agent 多消息返回**：单次用户指令触发 AI 一轮响应，但该轮响应由**多条 AI 消息**组成（按顺序追加到消息流）：Agent 过程追踪（UF-8 思考/计划/操作）→ 最终卡片或文字响应（UF-3/4/5/6）→ AI 跟进消息（确认完成、引导下一轮）。多条 AI 消息共享同一 turn 上下文，构成"AI 在回答我"的整体呈现。
- **写操作提交**：多张卡片的提交互不阻塞（用户可依次确认提交）；同一实体的并发写由后端乐观锁/版本号兜底（非 UI 层职责）。

---

## Component 3: 预填表单卡片（写操作）

### Placement

- **Mode**: existing-page
- **Target**: 渲染于 UF-2 聊天面板消息区内
- **Position**: 作为 AI 消息流中的一条卡片消息（max-w 90%，左对齐）。卡片自身 width 100% of 消息容器。**前置依赖：意图确认卡片（Component 9 / UF-9）通过后渲染**——agent 先推送 UF-9 列出提取字段，用户点"理解正确"后本卡片 append 到消息流。

### Layout Structure

卡片有两种**主形态**：

- **展开态**（预填/编辑/校验失败/提交失败）：完整字段区可见，可编辑或带错误。
- **折叠态**（已提交/已丢弃）：折叠为单行 summary（含状态图标 + title），点击可展开只读字段区。**已提交/已丢弃 两种终态默认折叠**，避免长 transcript 堆满已完成的 form。

#### 展开态布局

卡片（surface bg、border `#e2e8f0`、rounded-xl、shadow level-1、p-4，flex column gap-3）：

> **渲染规则**：`canSubmit=false` 时仅渲染卡片头（操作徽章+实体确认行）与权限提示体，跳过字段区、校验错误条、卡片底（不渲染可提交控件）。`canSubmit` 由前端 RBAC 依据 targetEntity.bizKey + cardType 在卡片渲染前同步校验。

1. **卡片头**：flex between
   - 左：操作类型徽章（rounded-full px-2.5 py-0.5 text-xs）—— create=accent-bg/text accent-hover；update=warning-bg/text warning-text；assign=accent-bg/text accent-hover + 实体类型文字（如"创建主事项"）
   - 右：目标实体 title + bizCode 二次确认区（见下方「高影响操作确认行」）
2. **字段区**：垂直堆叠字段行——字段集、控件类型、必填规则统一由 [`entity-schemas.md`](../prd/entity-schemas.md) 各实体 `fields` 驱动（参见本文档开头「Schema 驱动渲染」）
   - 每字段：label（14px 500 text-primary）+ 控件（h-10 rounded-md）+ 必填星号
   - **字段控件类型规则**（强制，禁止裸文本输入时间或关联值，由 schema 的 `control` 字段决定）：
     - `role: date` 字段 → DatePicker，禁止手敲日期字符串
     - `role: parent` / `assignee` / `submitter` / `team` / `priority` / `status` 字段 → Select 下拉，选项来自后端预加载的 Team 范围实体列表
     - `role: title` / `text` 字段 → Input / Textarea
     - `role: progress` 字段 → Number（0–100）
     - `role: code` 字段 → readonly
   - 必填且无值字段：`data-required-highlight` → border-2 warning `#d97706` + 左侧 3px warning 竖条（`border-left: 3px solid #d97706`，紧贴字段控件左边沿，用 `::before` 伪元素绝对定位 `left:-3px; top:0; bottom:0`）+ label 后红色 `*`
   - 字段值：13px text-primary；derived（AI 推导）值带浅灰底 accent-bg 标注"AI 推断"
   - **字段级错误就近展示**（提交失败时）：出错字段 border-2 error `#dc2626` + 字段下方 12px error-text 错误说明（`.field-error`）；与顶部错误条并存（顶部概述、字段级具体）
   - **高影响操作确认行**（update/assign/create 写操作，位于卡片头与字段区之间，独立成行）：accent-bg `#eff6ff`、rounded-md、px-3 py-2、flex items-center gap-2；实体标题 13px 500 text-primary（强权重），bizCode 12px text-tertiary（次权重，视觉分离）。例："`认证模块 · MI-0023`" → 标题"认证模块" 13px 500 text-primary + bizCode"MI-0023" 12px text-tertiary。提升二次确认可读性，避免误操作到错误实体。
3. **顶部错误条**（条件渲染，提交失败/校验失败）：bg error-bg `#fef2f2`、text error-text `#991b1b`、rounded-md、p-2、12px；提交失败时含当前 turn 的"重试 N 次"小字标注（历史会话回看此 turn 时不渲染该标注）；校验失败时含 `validTransitions` 合法目标状态列表（chips）
4. **卡片底**（flex justify-end gap-2，border-t `#e2e8f0`，p-3，bg `#f8fafc`）：
   - 取消（Ghost）、提交（Primary，必填空时 disabled）
   - 提交中态：提交按钮转 spinner + 文字"提交中…"，字段全部锁定

#### 折叠态布局（已提交 / 已丢弃）

折叠为单行 summary 行（`.form-folded`，surface bg、border `#e2e8f0`、rounded-lg、px-3 py-2、flex items-center gap-2、cursor pointer、hover bg-bg-alt）：

- **状态图标**（左）：✓ 已提交（accent `#3b82f6`）/ ⊘ 已丢弃（text-tertiary）
- **title 文本**：13px 500 text-primary（取自 form payload 的 title 字段，消息不可变即天然快照）
- **辅助标注**（右）：无标注（已提交与已丢弃都仅靠图标与 title 表达终态语义）
- **展开箭头**（右）：`›` 图标，点击 summary 行展开只读字段区（含原 form 的全部字段值，13px text-secondary 只读）
- **展开/折叠态**：记忆于会话内（同一用户回到该 turn 时保持上次选择）；历史会话回看时默认折叠

> **数据策略**：折叠态 summary 不二次抽取 keyFields/diff 等结构化字段。表单消息本身不可变，就是当时提交的完整快照；上下文完整性由 transcript 自身维持。用户要核对完整提交值 → 点击展开只读字段区；要看当前实体最新状态 → 点击 bizCode 跳详情页。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 预填展示 | 展开态：字段值 + 必填空字段 warning 高亮 + 目标实体确认区 | AI 推送完成 |
| 编辑中 | 展开态：字段可编辑；onChange 更新卡片 state | 用户聚焦字段 |
| 校验失败（提交前） | 展开态：顶部错误条 + validTransitions chips；表单保持可编辑 | available-transitions 不合法（仅状态变更类预校验阶段） |
| 提交中 | 展开态：字段锁定 + 提交按钮转 spinner + trace（Component 8）末尾追加"⏳ 调用 X API…"步 | 用户点提交 |
| 提交失败（后端校验） | 展开态：顶部错误条（含"重试 N 次"，历史回看不显示）+ 字段级错误就近展示（出错字段 border error + 字段下方说明）；已编辑字段值全保留 | API 返回参数错误 |
| 已提交 | **折叠态**："✓ 已提交 · {title}"单行；**AI 跟进一条自然语言消息**（作为独立 ai_text 消息，不在卡片内） | API 成功 |
| 已丢弃 | **折叠态**："⊘ 已丢弃 · {title}"单行（保留在 transcript 中避免孤儿用户输入） | 用户点取消/丢弃 |
| 权限不足 | 卡片体替换为权限提示：13px text-secondary + lock 图标"你对目标实体没有{op}权限，请联系管理员或更换目标" | canSubmit=false（RBAC 校验） |
| 流式填充中断 | 丢弃半填充骨架卡（不留半填字段），替换为系统消息"AI 响应中断，点击重试" + 重试 Ghost（沿用上一条指令重新生成卡片） | 流式连接中断（streamInterrupted=true） |

> **折叠规则**：已提交 / 已丢弃 两种终态默认折叠为单行 summary，点击 summary 行展开只读字段区。展开/折叠态记忆于会话内；历史会话回看时默认折叠。**不抽取 keyFields/diff 等结构化摘要**——表单消息本身不可变即天然快照。
> **AI 跟进消息**：仅"已提交"态触发（成功后），AI 跟发一条普通 `ai_text` 消息衔接下一轮；失败、丢弃态不发跟进消息（错误条/折叠态 summary 已是反馈）。跟进消息由 AI 从自己已有的上下文（form payload、用户原指令）生成，不依赖额外结构化字段或快照。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 编辑字段 | onChange 更新卡片 state | 必填空字段补值后高亮消除 |
| 继续对话补充 | 后端解析增量 → diff 内联区确认 → 更新卡片 state | 展示 diff 内联区（卡片体内）供确认，不静默覆盖，不弹浮层 |
| 点提交（必填全满） | 字段锁定 + trace 末尾追加"⏳ 调用 X API…"步 + 状态变更类先 available-transitions 预校验 → 通过则调 API | 切提交中态 |
| 提交成功 | 表单折叠为 summary 单行；AI 跟进一条 ai_text 消息 | 切已提交态（折叠） |
| 提交失败 | 表单保持展开；字段级错误就近展示 + 顶部错误条概述；保留已编辑值；当前 turn 标"重试 N 次"（历史回看不显示） | 提交按钮恢复可点 |
| 点击折叠态 summary 行 | 展开/折叠只读字段区 | 箭头旋转 150ms；展开/折叠态记忆于会话内 |
| 点重试（失败态） | 状态变更类先重跑 available-transitions 预校验 → 通过则重新调提交 API | 切提交中态；预校验失败回校验失败态并刷新 validTransitions |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 操作徽章 | cardType {create/update/assign} | AI 意图 |
| 实体确认区 | targetEntity {title,bizCode} | AI 实体抽取 |
| 字段控件 | fieldSet Map<name,{value,required,derived}> | AI 抽取 + schema |
| 顶部错误条 | validationError + validTransitions[] | available-transitions（预校验）/ 后端返回错误（提交失败） |
| 字段级错误 | fieldErrors Map<name, message> | 后端返回的参数校验错误 |
| 折叠态 summary | foldState {folded/expanded} + statusIcon + title | 表单状态机 + form payload 的 title 字段 |
| 重试计数 | retryCount number | 当前 turn 的失败重试次数（仅当前 turn 显示，历史回看不显示） |
| 提交权限 | canSubmit boolean | 前端 RBAC（targetEntity.bizKey + cardType） |

#### diff 内联区（对话补充增量变更，无二级浮窗）

当用户在卡片编辑中继续对话补充，后端解析出对当前卡片的增量变更时，在**卡片体内**内联展示 diff 区（不弹浮窗），确认动作由输入区选项组承载（见 Component 2「输入区双模式」）。

- **定位**：内联于卡片字段区下方（非 absolute 浮层），随卡片正常排版、随消息区滚动，不会脱离锚点。
- **外观**：`.diff-inline`，accent-bg `#eff6ff`、border accent-ring `#bfdbfe`、rounded-md、p-2.5，flex column gap-1.5。
- **内容**：标题 13px 500 text accent-hover"对话补充将更新以下字段："+ 字段 diff 行（字段名 text-secondary · 旧值 text-tertiary 删除线 → 新值 text-primary 500）；底部 `.diff-hint` 12px text-tertiary"用下方输入区确认（↑↓ 选择 · Enter 确认）"。
- **确认动作**：卡片内不放按钮（避免二级浮窗/按钮冗余）；输入区切到选项组模式，选项 =「应用变更」「丢弃」（`Esc`=丢弃）。应用则按 last-write-wins 合并入卡片 state，丢弃则保留卡片当前值。
- **无浮层收益**：不引入额外浮层；diff 随卡片滚动不脱离锚点；卡片交互保持扁平。

---

## Component 4: 查询结果卡片（查询操作）

### Placement

- **Mode**: existing-page
- **Target**: 渲染于 UF-2 聊天面板消息区内
- **Position**: AI 消息流中的一条卡片消息

### Layout Structure

卡片容器（surface、border、rounded-xl、shadow-1、p-4、flex column gap-2）：

1. **摘要行**：14px 500 text-primary（如"你有 3 个 P0 事项"）
2. **结果列表**：垂直堆叠，每项为**双态 result-card**（surface bg、border `#e2e8f0`、rounded-md、cursor pointer、hover border-accent + bg-bg-alt、flex column gap-6px）。两个态：
   - **折叠态（默认 · 多记录）**：展示核心 8 字段
   - **展开态（默认 · 单记录 / 用户点击触发 · 多记录）**：折叠态 8 字段 + 追加约 4 字段（共 ~12 字段）
   
   **字段按 [`entity-schemas.md`](../prd/entity-schemas.md) 各实体 `result_slots`（折叠）+ `result_slots.expanded`（展开追加）渲染**，渲染器与实体类型解耦。折叠态槽位：
   - **头行（rc-head，slot=head，可点击 toggle）**：`title` role（13px 500 text-primary，超长省略号）+ `code` role（11px text-tertiary）+ chevron 图标（text-tertiary，hover 转 accent；展开态旋转 90°）
   - **字段行（rc-fields，slot=fields）**：横排核心字段，含徽章 + 文本——`priority`（badge-error P0 / badge-warning P1 / badge-neutral P2/P3）、`status`（按状态语义映射 badge-success/neutral/warning）、`assignee` / `submitter`（`👤 {name}`）
   - **元信息行（rc-meta，slot=meta）**：12px text-tertiary 横排——`date` role（`📅 {label} {date}`）、`parent` role（`📁 {label} {parent title}`）、`team` role（`👥 {teamName}`），按实体 schema 的 meta 列表选填
   - **进度条（rc-progress，slot=progress，可选）**：`progress` role 字段渲染为 4px 细进度条 + 进度文字（如"75% · 3/4 子任务"）；source 为 direct / subitems / subitems_of_milestone / milestones，按 schema 声明
   
   **展开态追加块（rc-extra，border-top 分隔，p-top 8px，flex column gap-4px）**——仅 `.result-card.expanded` 时可见，追加字段按实体类型差异化（MainItem 示例 ~4 字段）：
   - **描述全文**（description）：12px text-secondary，line-height 1.5，超长截断 2 行（ellipsis）
   - **创建人 + 创建时间**：12px text-tertiary 横排"由 {createdBy} 创建于 {createdAt}"
   - **子任务/里程碑统计**：12px text-secondary，如"3/5 已完成 · 最近 1 条 ProgressRecord: 60% 接口对接中"
   - **最近状态变更**：12px text-tertiary，如"2026-06-20: in_progress → review"
3. **截断提示**（条件）：超过 20 张时底部 text-tertiary 12px"结果过多，显示前 20 条，请缩小范围（指定标题/负责人）"
4. **空态**：text-secondary 13px"未找到匹配事项"+ 建议（如"试试'我的 P1 事项'"）

> **核心原则：内容自包含 + 上下文不中断**。用户在面板内就能回答"这个事项是什么状态、谁负责、什么时候截止、进度如何、谁创建的、最近一次状态变更"等问题。**不再有跨页跳转**——单记录自动展开、多记录点击就地展开（独立 toggle，不互斥）。对话上下文与多轮查询始终保持在同一面板内。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 单记录 · 自动展开 | 摘要 + 单张 `.result-card.expanded`（核心 8 字段 + 追加 ~4 字段）；chevron 旋转 90° | 查询命中且结果唯一，渲染时直接 `expanded=true` |
| 多记录 · 全部折叠 | 摘要 + 折叠态 result-card 列表（每张核心 8 字段）；chevron 朝右 | 查询命中且结果 ≥ 2 |
| 多记录 · 部分展开 | 被点击的卡片切到 `.result-card.expanded`（多张可同时展开，独立 toggle）；chevron 旋转 90° | 用户点击卡片头行 |
| 无结果 | "未找到匹配事项"+ 建议 | 查询未命中 |
| 截断 | 前 20 张折叠卡 + 截断提示 | 结果 >20 |

> **流式说明**：查询结果（UF-4）与歧义消解（UF-5）均**原子返回**（后端一次性返回完整结果，非增量流式），因此不适用 UF-3 的「流式填充中断」清理规则；UF-2 的「流式中断」态仅针对 UF-3 写卡片的骨架增量填充场景。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击 result-card 头行（多记录折叠态） | 切换该卡片 `expanded` class（**独立 toggle，不互斥**，多张可同时展开） | chevron 旋转 90°（150ms）；追加块 `rc-extra` 显示/隐藏 |
| 再次点击展开态卡片头行 | 折叠回核心字段 | chevron 旋转回 0°；追加块隐藏 |
| 单记录卡片 | 无 toggle 行为（渲染即展开，chevron 仍旋转 90° 作为视觉一致性提示） | — |
| hover 卡片 | border 转 accent、bg 转 bg-alt | chevron 转 accent |
| 键盘聚焦头行（`tabindex=0`） + Enter/Space | 切换 expanded | 同点击 |

> **不跳转、不弹窗**：UF-4 卡片不产生跨页导航、不弹二级浮窗。所有交互在面板内完成，保持对话上下文连续。

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 摘要 | summaryText | AI/后端 |
| result-card 头行 | entity.title + entity.bizCode | 查询 API |
| result-card 字段行 | entity.priority + entity.status + entity.assignee | 查询 API |
| result-card 元信息行 | entity.{expectedEndDate,milestone,parent,...} | 查询 API（按实体类型 schema 选填） |
| result-card 进度条 | progress{completed,total,percent} | 查询 API（MainItem/SubItem/Milestone/MilestoneMap） |
| result-card 展开追加（描述） | entity.description / background / expectedOutput / achievement | 查询 API（schema `result_slots.expanded`） |
| result-card 展开追加（创建人/时间） | entity.{createdBy,createdAt} | 查询 API |
| result-card 展开追加（子任务/里程碑统计） | entity._childrenStats 或 _milestoneStats | 查询 API（按实体类型） |
| result-card 展开追加（最近状态变更） | entity._lastStatusChange {time,from,to} | 查询 API |
| 单记录标志 | isSingleResult boolean | 查询 API（结果数 = 1 时 true） |
| 展开态 class | expanded boolean | 用户点击 toggle / 单记录时自动 true |

---

## Component 5: 歧义消解卡片

### Placement

- **Mode**: existing-page
- **Target**: 渲染于 UF-2 聊天面板消息区内
- **Position**: AI 消息流中的一条卡片消息（位于写操作卡片之前）

### Layout Structure

卡片（surface、border、rounded-xl、shadow-1、p-4、flex column gap-2）：

1. **提示行**：14px text-primary"找到多个匹配，请选择目标："+ 原始引用（text-secondary 12px）
2. **候选列表**（radio-group 样式）：每项（rounded-lg、border `#e2e8f0`、p-2、flex、gap-2）：
   - 左：radio 圆点（未选 border-dark；选中 accent 实心）
   - 中：标题（13px text-primary）+ bizCode（12px text-tertiary）+ 状态徽章
   - 选中态：border-2 accent + accent-bg
3. **确认按钮**：Primary，未选时 disabled

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 待选择 | 候选列表可选，确认 disabled | 多候选命中 |
| 已选择 | 选中项 accent 高亮，确认可用 | 用户选定 |
| 已确认 | 卡片折叠为"已选择：{title}" | 继续后续流程（推送 UF-3/4） |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点候选 | 选中该项 | accent 高亮，确认可用 |
| 点确认 | 以所选实体继续 | 本卡片折叠为"已选择：{title}"摘要行（保留在原消息位）；后续写/查卡片（UF-3/UF-4）作为新的 AI 消息 **append 到消息列表末尾**（与普通 AI 回复同一消息流，非替换本卡片），`pendingIntent` + 所选 `candidate.bizKey` 注入新卡片作为预填来源 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 候选列表 | candidates[{bizKey,title,code,meta}] | 模糊匹配 |
| 原始意图 | pendingIntent {op,fields} | 暂存 AI 解析 |

---

## Component 6: 降级与超时提示

### Placement

- **Mode**: existing-page
- **Target**: 渲染于 UF-2 聊天面板消息区内（替代/追加到 AI 回复位）
- **Position**: 消息流中的系统/卡片消息

### Layout Structure

提示卡片（warning-bg `#fffbeb`、border warning `#d97706`/30、rounded-lg、p-3、flex column gap-2）：

1. **图标+标题**：warning 图标（warning-title `#d97706`，图形 3:1 即可）+ 14px 500 text warning-text `#92400e`"AI 响应超时"/"AI 暂不可用"（14px/500 非大字，需 ≥4.5:1，故用 warning-text `#92400e` 而非 warning-title `#d97706`）
2. **说明**：13px 500 text warning-text `#92400e`"已超过 10 秒未响应，可使用传统表单继续操作"（≥4.5:1 on warning-bg）
3. **快捷入口**（Ghost 按钮，border border-dark）：按降级类型映射路由——创建→"去 `/items` 手动新增"、ItemPool→"去 `/item-pool` 申请"、查询→"去 `/items` 筛选"

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 超时 | warning 提示 + 手动入口 | 单次 >10s |
| 不可用 | 同上，标题"AI 暂不可用" | AI 整体不可用 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点快捷入口 | 跳转既有路由 | 导航 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 降级类型 | fallbackType {timeout/unavailable} | 后端代理状态 |
| 快捷入口 | fallbackRoute | 类型→路由映射 |

---

## Component 7: 首次引导卡片

### Placement

- **Mode**: existing-page
- **Target**: 渲染于 UF-2 聊天面板首次展开时
- **Position**: 面板消息区首位（空会话）

### Layout Structure

卡片（accent-bg `#eff6ff`、border accent-ring `#bfdbfe`、rounded-xl、p-4、flex column gap-3）：

> **居中规则**：onboarding 卡片所在 `.msg.ai` 用 `align-self: center`、`max-width: 80%`、`align-items: stretch` 水平居中呈 hero 形态（与普通左对齐 AI 消息视觉区分，强调"欢迎/引导"语义）。**卡片内部内容（标题、能力列表、示例 chips）一律左对齐**，与普通 AI 消息内部排版一致。

1. **图标+标题**：气泡图标（accent）+ 14px 500 text accent-hover"我是 AI 助手，可以帮你："
2. **能力列表**：四行（flex、gap-2、13px text-primary），每行图标+文字：
   - 创建（事项/子任务/里程碑/里程碑图/ItemPool）
   - 查询（"我的 P0 事项"）
   - 修改（状态/进度）
   - 分配（负责人）
3. **示例指令**（chips，可点击填入输入框）：rounded-full border border-dark px-3 py-1 13px，如"创建一个 P1 事项"、"我的 P0 事项有哪些？"、"更新子任务完成度为 60%"
4. **关闭**：右上 Ghost X（收起引导卡片，本次会话内不再展示；新建会话或重新展开面板时再次出现）

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 空会话展示 | 完整能力卡片（intro + 能力列表 + 示例 chips） | 面板首次展开、或用户新建会话使消息区为空；首次展开与新会话**同一张卡片**，不做区分 |
| 本会话内已收起 | 不展示 | 用户点 X 或开始输入；本会话内有效，新建会话或重新展开时再次展示 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点示例 chip | 填入输入框（不自动发送） | 焦点留输入框 |
| 点 X / 开始输入 | 收起引导卡片 | 会话内状态置 dismissed=true（本会话内不再展示；新建会话或重新展开面板时重置） |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 能力/示例 | 静态文案 | — |
| 本会话内已收起 | dismissed | 会话内状态（不跨会话持久化） |

---

## Component 8: Agent 过程追踪（透明化中间步骤）

> 借鉴 Claude Code 等成熟 agent：把 AI 的**思考、计划、工具调用**等中间过程流式展示给用户，建立信任、可调试、降低"黑盒"感。最终卡片（UF-3/4）在过程结束后渲染于其下方。

### Placement

- **Mode**: existing-page
- **Target**: 渲染于 UF-2 聊天面板消息区内
- **Position**: 作为 AI 消息流中的一条消息，位于用户指令与最终卡片之间；按后端流式事件增量渲染。

### Layout Structure

可折叠卡片（`.agent-trace`，surface、border、rounded-xl、shadow-1、max-w 90%）：

1. **头部**（`.at-header`，flex、cursor pointer、点击折叠/展开）：折叠箭头（`.at-caret`，展开↓/折叠▶旋转 150ms）+ "Agent 思考过程" + 右侧元信息（步数 · 累计耗时）。
2. **主体**（`.at-body`，分节，每节 `.at-section` + `.at-section-label` 大写小标签）：
   - **思考**（thinking）：斜体、text-secondary，流式追加（agent 对意图的理解与字段推导理由）。
   - **计划**（plan）：编号步骤列表（如 解析意图→加载 Team schema→模糊匹配负责人→解析日期→推送卡片）。
   - **操作**（action）：每个工具/后端调用一行——动作名 + 状态图标（streaming 闪烁光标 / done ✓ / error ✗）+ 耗时；如「加载 Team「平台组」MainItem schema ✓ 0.3s」「模糊匹配「张三」→ 命中 张三 (user_017) ✓ 0.4s」。
3. **摘要条**（`.at-summary`，bg `#f8fafc`、border-t）：完成提示 + token 用量/成本。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 流式中 | 步骤逐条追加，末步带闪烁光标（`.at-cursor`） | 后端流式事件驱动；首字节（思考出现）< 1s |
| 已完成（成功） | 全部步骤 done ✓，光标消失，**下方渲染下一步内容**（写操作 → UF-9 意图确认卡片；查询 → UF-4 结果卡片） | **自动折叠**为头部摘要（约 600ms 延迟，让下一步内容成为视觉焦点），用户可点击重新展开。UF-3 表单卡片需用户在 UF-9 确认后才渲染 |
| 步骤失败 | 该步 error ✗ + error-text | **不自动折叠**（保持展开让错误可见，错误可见性优先于简洁）；中断后续步骤，引导用户（如"匹配失败，请补充信息"），不渲染下一步内容 |
| 折叠 | 仅显示头部（步数·耗时 / 或失败图标 + 错误摘要） | 用户点击头部切换；折叠态记忆于会话内；**历史会话打开时，过往 turn 的 trace 一律默认折叠**，避免长 transcript 一打开就刷屏 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击头部 | 折叠/展开 | 箭头旋转 150ms |
| 点击操作行（可选） | 展开该步详情（请求/响应摘要） | 内联展开 |
| 流式成功完成 | 末步 done，光标消失 → 约 600ms 后**自动折叠**为头部摘要 | 下方渲染下一步内容（写操作 → UF-9 意图卡片 / 查询 → UF-4 结果卡片）；UF-3 表单需用户在 UF-9 确认后才渲染。表单提交中态时同步追加 API 步并最终折叠 |
| 流式出现失败步骤 | 失败步 ✗ + error-text；**保持展开**不自动折叠 | 中断后续，不渲染下一步内容改为引导文字 |
| 历史会话打开 | 过往 turn 的 trace **默认折叠**为头部摘要 | 用户想看哪轮点哪轮 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 思考文本 | trace.thinking | 流式事件 `thinking` |
| 计划步骤 | trace.plan[] | 流式事件 `plan_step` |
| 操作行 | trace.actions[{name,status,durationMs,detail}] | 流式事件 `tool_call` / `tool_result` |
| 摘要 | trace.summary{tokens, durationMs, costUsd} | 后端汇总 |

> **流式协议**：后端 AI 代理以 SSE/流式返回结构化事件序列 `thinking → plan_step* → tool_call/tool_result* → final_card`；前端按事件增量渲染 trace，最终卡片在 `final_card` 事件后渲染。**延迟目标**：首字节（思考出现）< 1s，计划可见 < 2s，最终卡片 P95 < 5s（见 PRD NFR）。
> **透明 vs 噪声（不对称折叠规则）**：trace 流式期间展开（建立信任、可调试）；**成功完成 → 约 600ms 后自动折叠**为头部摘要，让最终卡片成为视觉焦点；**出现失败步骤 → 保持展开不自动折叠**，错误可见性优先于简洁；用户手动展开/折叠态记忆于会话内；**历史会话打开时，过往 turn 的 trace 一律默认折叠**，用户想看哪轮点哪轮。失败步骤高亮，便于用户理解"为何没出卡片"。
> **与表单卡片（Component 3）的衔接**：表单进入"提交中"态时，trace 末尾追加一行"⏳ 调用 {Entity} {Op} API…"步，与流式期间的 plan/tool_call 步同等展示；API 返回后状态图标迁移为 ✓（成功）/ ✗（失败）。该步计入 trace 的步数与耗时统计。表单提交成功后折叠时，trace 也按上述规则同步折叠（成功即折叠、失败保持展开）。
> **"trace 失败"语义边界**：trace 中任一步骤（包括末尾的 API 调用步）失败都算"trace 出现失败步骤"→ 保持展开。两种典型场景：(a) **AI 推理失败**（如实体匹配失败、解析失败）—— trace 末步前就 ✗，下方不渲染卡片改为引导文字；(b) **AI 推理成功 + 后端写拒绝**（如字段校验失败、状态机不允许）—— AI 推导步骤全 ✓、末尾 API 步 ✗，下方仍渲染表单卡片并展开字段级错误。两种场景的 trace 都保持展开（不对称折叠规则一致），但 summary 文案需区分：(a) "执行失败 · 不渲染卡片"；(b) "AI 推导成功 · 后端写拒绝"。
> **降级**：AI 不支持流式时，trace 退化为单条"AI 思考中…"系统消息 + 下一步内容（UF-9 意图卡片或 UF-4 结果卡片）。

---

## Component 9: 意图回执 + 主动澄清（UF-9 写操作前置 · 文本模式）

### Placement

- **Mode**: existing-page
- **Target**: 渲染于 UF-2 聊天面板消息区内
- **Position**: 作为 AI 消息流中的若干条**普通 AI 文本消息**（`.msg.ai .msg-bubble`），位于 Component 8 trace 之后、Component 3 表单卡片之前。**不渲染任何卡片**——意图与澄清都是自然语言，文本最贴切（卡片结构化反而把意图拆碎成字段列表）。仅写操作（创建/修改/分配）触发；查询（UF-4）不经此流程。

### Layout Structure

由 1–3 条普通 AI 文本消息组成（沿用 Component 2 `.msg.ai .msg-bubble` 样式，无独立卡片类）：

1. **意图回执消息**（必发，单条）：一两句自然语言复述用户意图，不罗列字段。
   - 例：`好的，我帮你创建一个 P1 主事项「完成用户认证模块」，分配给张三。`
   - 视觉：与其他 AI 文本消息完全一致（surface bg、border `#e2e8f0`、rounded-xl、px-3 py-2、13px text-primary）。
   - 输入区切选项组（理解正确 / 我要调整 / 取消）。
2. **主动澄清消息**（条件，用户选"理解正确"后且必填有缺失时发）：列出还需补充的字段。
   - 例：`还需要补充以下信息：\n• 预期截止日期是？\n• 归到哪个里程碑下？（第一阶段 / 第二阶段）`
   - 一轮列全部缺失项（不拆成多轮单问）；附 hint（枚举范围、示例格式）。
   - 输入区切回文本模式，用户用自然语言回答。
3. **信息收齐确认消息**（条件，主动澄清收齐后发）：宣告信息已收集完整，作为表单卡片的引子。
   - 例：`收到，所有信息已收集完整。请核对表单：`
   - 紧随其后渲染 Component 3 表单卡片（所有必填字段已预填、无 required-highlight）。

> **不渲染卡片**：意图与澄清是对话行为，沿用普通 `.msg.ai .msg-bubble` 即可；卡片结构化（border-accent / padding / shadow）只用于承载结构化数据（表单/查询结果/候选），单句文本不需要这层包装。对齐 ChatGPT/Claude 等成熟对话助手的"意图回执 = 文本"模式。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 意图回执 · 待确认 | 意图回执文本消息 + 输入区选项组（理解正确 / 我要调整 / 取消） | agent 解析完成，渲染于 trace 之后 |
| 主动澄清 · 待回答 | 意图回执 + 主动澄清文本消息 + 输入区文本模式 | 用户选"理解正确"且必填缺失 |
| 主动澄清 · 继续追问 | 前述消息 + 新追问文本消息 + 输入区文本模式 | 用户回答后仍有缺失 |
| 信息收齐 · 渲染表单 | "信息已收集完整"文本消息 + Component 3 表单卡片（无 required-highlight） | 阶段 2 收齐所有必填 |
| 已调整 | 输入区聚焦预填用户原指令 | 用户选"我要调整"（任意阶段可触发） |
| 已取消 | 输入区切回文本；不渲染表单 | 用户选"取消" |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 输入区选"✓ 理解正确"（意图回执阶段） | 若必填有缺失 → 发主动澄清消息 + 切文本模式；若已收齐 → 发"信息收齐"消息 + 渲染 Component 3 表单卡片 | 选项组释放，输入区切文本或保持选项（取决于是否还有 Q&A） |
| 用户在主动澄清阶段发送回答 | agent 解析回答、更新 draft state；若仍有缺失继续追问；若全部收齐发"信息收齐"消息 + 渲染 Component 3 表单卡片 | 用户 msg append + AI 文本响应 |
| 输入区选"✎ 我要调整" | 输入区切回文本模式、textarea 聚焦、预填用户上一条指令**原文** | 用户编辑重发 → agent 重新解析 → 新意图回执 append |
| 输入区选"✗ 取消" | 输入区切回文本模式 | 本次写操作终止 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 意图回执文本 | intent echo text | AI 生成（基于 opType + entityType + 已知字段） |
| 主动澄清问题列表 | missing required fields + hints | AI 抽取 vs schema required 对比 |
| Draft 字段状态 | fieldMap | 跨轮累积（用户回答持续补充） |
| 用户原指令（调整时预填） | 上一条 user msg textContent | transcript |

---

## Accessibility（跨组件，参照 WCAG 2.1 AA）

- **焦点管理**：展开面板焦点入输入框；Esc 焦点回气泡；卡片出现不抢断焦点，Tab 可达。
- **键盘**：所有卡片操作（编辑、选择、提交、重试）键盘可完成，无键盘陷阱；Enter 发送、Shift+Enter 换行、Esc 收起。
- **ARIA**：消息历史 `role="log"` `aria-live="polite"`；气泡 `aria-label="AI 助手（展开/收起）"`；面板 `role="dialog"` `aria-label`；卡片 `aria-label` 描述类型；错误态 `role="alert"`；Team 徽章 `aria-live="polite"`。
- **屏幕阅读器**：所有状态（思考/流式/错误/超时/成功）有 aria-live 播报；加载态有文字标签。
- **对比度**：正文 ≥4.5:1；大字（≥18.66px/500 常规，或 ≥14px/700 加粗，或 ≥24px）≥3:1。**14px/500 不属大字**。Component 6 警告标题与正文统一用 warning-text `#92400e`（≥4.5:1 on warning-bg `#fffbeb`），warning-title `#d97706` 仅用于图标（图形非文字，3:1 即可）。必填高亮用对比达标边框色，不仅靠颜色。
- **动效兜底（reduced-motion）**：所有动画（面板滑入 translateX、三点跳动思考态、骨架流式填充、进度条 width 过渡、卡片状态切换）在 `@media (prefers-reduced-motion: reduce)` 下降级为瞬时切换（无位移/无跳动，仅颜色或可见性变化）；状态信息不依赖动效传达（思考态附"AI 思考中…"文字，不仅靠跳动点）。

---

## Prototype 映射

原型文件结构（位于 `ui/prototype/`）：

| 文件 | 内容 | 对应组件 |
|------|------|----------|
| `index.html` | 评审 hub：列出全部演示状态与设计系统说明，链接到 demo.html | 导航入口 |
| `demo.html` | 模拟宿主页（侧边栏 + 内容区）+ 浮动气泡 + 聊天面板 + 全部卡片状态 + 左下角状态切换器（16 个 state） | UF-1/2/3/4/5/6/7/8/9 全部 |
| `styles.css` | 共享 token + 组件类（复用 DESIGN.md） | 全部 |
| `app.js` | 交互（展开/收起、发送、卡片状态切换、Agent trace 流式、form 折叠、result-card 折叠/展开 toggle、`ENTITY_SCHEMAS` + `renderEntityCard` 渲染器） | 全部 |

`demo.html` 左下角的状态切换器覆盖 16 个 state：onboarding / **intent-echo（UF-9 意图回执文本，等待用户确认）** / write-prefilled（UF-9 收齐信息后的完整对话流 + UF-3 表单）/ validation / submit-failed / committed / discarded / permission / diff-overlay / **query（UF-4 多记录折叠态，点击独立 toggle 展开）** / **query-single（UF-4 单记录自动展开紧凑详情）** / query-entities / disambiguation / fallback / thinking / agent-trace / trace-failed。intent-echo 演示单条文本意图回执 + 选项组；write-prefilled 演示完整对话流；query 演示多记录折叠卡片 + 点击就地展开（独立 toggle，不互斥）；query-single 演示单记录自动展开（追加描述/创建人/子任务统计/最近状态变更等约 4 字段）；其余 state 对应 UF-3/5/6/7/8 的状态分支。
