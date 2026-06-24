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
   - 时间戳：消息下方 text-tertiary 11px
3. **输入区**（border-t `#e2e8f0`，bg surface，p-3）：
   - textarea（文本模式）：`min-height: 120px`、`max-height: 280px`、**按内容自动增高**（`resize: none`，JS 按scrollHeight 在 120–280px 区间自适应）、达到 280px 后 `overflow-y: auto` 出现滚动条、rounded-md、border border-dark、px-3 py-2、13px、placeholder text-tertiary"描述你想做的事…"；focus border accent + ring accent-ring。**单次输入上限 1000 字符**，超出截断并提示。
   > **与全局约定一致**：`min-h 120px` 对齐 `frontend-components.md` TECH-frontend-001 的 `min-h-[120px]`（不再偏离）；上限 280px 为 Copilot 面板内的增高上限，超出滚动。
   - 字符计数：右下角 text-tertiary 11px `{n}/1000`，超 1000 红色
   - 发送按钮（文本模式）：Primary sm，h-8 px-3，右下；Enter 发送，Shift+Enter 换行；thinking 态禁用 + 加载点

#### 输入区双模式（文本 ↔ 选项组，键盘优先）

输入区是 Copilot 的**唯一确认入口**——卡片本身不弹二级浮窗，所有"需要用户确认/选择"的操作都把输入区从「文本模式」切换为「选项组模式」，用户用键盘完成，无需鼠标：

- **文本模式（默认）**：即上述 textarea，用户输入自然语言指令。
- **选项组模式（有待确认时触发）**：textarea 隐藏，输入区渲染为选项列表（`.option-list`）。触发场景：写卡片待提交（提交/编辑字段/取消）、校验失败（修改/取消）、撤回窗口内（撤回/完成）、权限不足（换目标/取消）、对话补充 diff（应用/丢弃）、歧义消解（候选实体即选项）、超时降级（手动入口/重试）。
  - **键盘**：`↑`/`↓` 在选项间移动高亮（循环），`Enter` 确认当前高亮项，`←`（左方向键）返回文本输入模式，`Esc` 收起面板；选项过多时 `scrollIntoView({block:'nearest'})` 跟随。
  - **鼠标**：点击选项 = **直接选中并确认**该选项（事件 `stopPropagation`，不触发"面板外点击收起"，面板保持展开）。
  - **视觉**：每项 `.option` 14px 行高、border；高亮项 border-2 accent + accent-bg + text accent-hover；右侧 `kbd` 提示键位。
  - **提示行**：选项组顶部固定 `.options-hint`："↑↓ 选择 · Enter 确认 · ← 返回输入 · Esc 收起"。
- **阻塞语义**：选项组模式期间不接受新自由文本（textarea 隐藏），迫使用户先解决当前确认，避免上下文错乱；`Esc` 可退回文本模式重新描述。

#### 面板双视图（对话视图 ↔ 会话列表视图）

面板内容区有两个互斥视图，由头部「历史会话」按钮切换：

- **对话视图（chat-view，默认）**：即上述消息区 + 输入区，展示当前活动会话。
- **会话列表视图（session-list-view）**：展示历史会话列表，每项含会话标题、首条预览、时间、消息数、待确认标记；当前活动会话 accent-bg 高亮。**列表视图无独立标题栏**——打开列表视图时，**面板头中间标题切换为"历史会话"**（替代会话标题）；新建会话由头部右侧「+」图标按钮承担（不在列表内重复放按钮）。
  - 点某会话 → 切换为活动会话（加载其消息、恢复待确认状态与撤回倒计时、更新头部会话标题），返回对话视图，标题恢复为该会话标题。
  - 头部「+」新建会话：**若当前已是新空会话（无消息），不重复创建**，直接返回对话视图；否则创建空会话为活动会话，展示首次引导（UF-7）。
- **多会话决策（不并行）**：面板同一时刻只有**一个活动会话**；**不支持同时打开多个会话**——并行多会话需多窗口/标签，与「单全局 overlay」模型冲突，且增加上下文切换负担。用户通过会话列表快速切换活动会话；切换时前一会话状态保留（待确认卡片、5 分钟撤回窗口继续计时）。
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
| 主页面导航（hasUncommittedCards=true） | 拦截路由变更，弹离开确认 Dialog | 见「离开确认 Dialog」规格 |

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
| 未提交卡片存在 | hasUncommittedCards | 任一 UF-3 处于 预填/编辑中/校验失败/提交中，或 UF-5 处于 待选择/已选择未确认 时为 true |
| 流式中断 | streamInterrupted | 流式连接断开时 true，触发骨架卡丢弃 |
| 会话内 Team 切换 | teamChangedMidFlight {oldName,newName} | 检测 teamCtx.bizKey 变化且存在 in-flight/pending 卡片时填充 |

#### 离开确认 Dialog（路由守卫）

`max-w-sm`（400px，DESIGN.md sm 规格）、居中、overlay black/50、surface bg、rounded-xl、shadow level-3、p-4。

- **标题**：14px 500 text-primary"当前有未提交的操作，确定离开？"
- **说明**：13px text-secondary"离开将丢失未提交的卡片草稿。"
- **按钮**（flex justify-end gap-2，mt-4）：取消（Ghost sm，恢复导航前状态）/ 确定离开（Danger sm，border error-text/40，放行导航并丢弃草稿）
- **未提交判定**：`hasUncommittedCards=true` 即触发（见 Data Binding）
- **焦点**：打开时焦点移入「取消」（防误点离开）；Esc 等同取消；Tab 顺序 取消→确定离开。

#### 并发卡片策略

- **多卡片共存**：消息流中可同时存在多张 UF-3/UF-5 卡片（用户可连续发起多个操作），每张卡片 state 独立、互不覆盖。
- **AI 请求串行**：同一会话同一时刻仅允许一个 in-flight AI 请求（思考态全局唯一）；用户在思考态发送新指令时，新指令入队等待（输入区 inline 提示"上一条处理中…"），前一条返回后再处理下一条，避免并发 prompt 组装冲突与限流。
- **写操作提交**：多张卡片的提交互不阻塞（用户可依次确认提交）；同一实体的并发写由后端乐观锁/版本号兜底（非 UI 层职责）。

---

## Component 3: 预填表单卡片（写操作）

### Placement

- **Mode**: existing-page
- **Target**: 渲染于 UF-2 聊天面板消息区内
- **Position**: 作为 AI 消息流中的一条卡片消息（max-w 90%，左对齐）。卡片自身 width 100% of 消息容器。

### Layout Structure

卡片（surface bg、border `#e2e8f0`、rounded-xl、shadow level-1、p-4，flex column gap-3）：

> **渲染规则**：`canSubmit=false` 时仅渲染卡片头（操作徽章+实体确认行）与权限提示体，跳过字段区、校验错误条、卡片底（不渲染可提交控件）。`canSubmit` 由前端 RBAC 依据 targetEntity.bizKey + cardType 在卡片渲染前同步校验。

1. **卡片头**：flex between
   - 左：操作类型徽章（rounded-full px-2.5 py-0.5 text-xs）—— create=accent-bg/text accent-hover；update=warning-bg/text warning-text；assign=accent-bg/text accent-hover + 实体类型文字（如"创建主事项"）
   - 右：目标实体 title + bizCode 二次确认区（见下方「高影响操作确认行」）
2. **字段区**：垂直堆叠字段行
   - 每字段：label（14px 500 text-primary）+ 控件（h-10 rounded-md）+ 必填星号
   - **字段控件类型规则**（强制，禁止裸文本输入时间或关联值）：
     - 日期/时间字段（planStartDate、expectedEndDate 等）→ **日期选择组件**（`<input type="date">` 或 DatePicker），禁止手敲日期字符串。
     - 关联/引用字段（milestoneKey 里程碑、assignee 负责人、parent MainItem、teamKey 等）→ **Select 下拉组件**，选项来自当前 Team 范围内的实体列表（后端预加载），禁止自由文本。
     - 纯文本字段（title、description、achievement 等）→ Input/Textarea。
   - 必填且无值字段：`data-required-highlight` → border-2 warning `#d97706` + 左侧 3px warning 竖条（`border-left: 3px solid #d97706`，紧贴字段控件左边沿，用 `::before` 伪元素绝对定位 `left:-3px; top:0; bottom:0`）+ label 后红色 `*`
   - 字段值：13px text-primary；derived（AI 推导）值带浅灰底 accent-bg 标注"AI 推断"
   - **高影响操作确认行**（update/assign/create 写操作，位于卡片头与字段区之间，独立成行）：accent-bg `#eff6ff`、rounded-md、px-3 py-2、flex items-center gap-2；实体标题 13px 500 text-primary（强权重），bizCode 12px text-tertiary（次权重，视觉分离）。例："`认证模块 · MI-0023`" → 标题"认证模块" 13px 500 text-primary + bizCode"MI-0023" 12px text-tertiary。提升二次确认可读性，避免误操作到错误实体。
3. **校验错误条**（条件渲染）：bg error-bg `#fef2f2`、text error-text `#991b1b`、rounded-md、p-2、12px；含 `validTransitions` 合法目标状态列表（chips）
4. **卡片底**（flex justify-end gap-2，border-t `#e2e8f0`，p-3，bg `#f8fafc`）：
   - 取消（Ghost）、提交（Primary，必填空时 disabled）
   - 成功态：提交区替换为"已创建 ✓"（蓝色对勾，蓝=成功）+ 实体跳转链接 + （可逆操作）撤回按钮（Ghost，text-error，带倒计时 mm:ss）

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 预填展示 | 字段值 + 必填空字段 warning 高亮 + 目标实体确认区 | AI 推送完成 |
| 编辑中 | 字段可编辑；onChange 更新卡片 state | 用户聚焦字段 |
| 校验失败 | 错误条 + validTransitions chips | available-transitions 不合法（仅状态变更类） |
| 提交中 | 提交按钮 disabled + 加载点 | 用户点提交 |
| 成功 | 蓝色对勾"已创建"+ 跳转链接 +（可逆）撤回按钮+倒计时 | API 成功；undoDeadline = now+5min |
| 不可逆成功 | 蓝色对勾 + 标注"该操作不可撤回"，无撤回按钮 | 转入 terminal 状态变更 |
| 失败 | 错误信息（error-bg）+ 重试（字段保留可编辑） | API 失败 |
| 撤回成功 | 卡片标注"已撤回"，撤回按钮消失 | 用户点撤回，反向操作成功 |
| 撤回窗口过期 | 撤回按钮消失，标注"撤回窗口已过期" | undoDeadline 到达 |
| 权限不足 | 不渲染字段区与提交按钮；卡片体替换为权限提示：13px text-secondary + lock 图标"你对目标实体没有{op}权限，请联系管理员或更换目标" | canSubmit=false（RBAC 校验） |
| 流式填充中断 | 丢弃半填充骨架卡（不留半填字段），替换为系统消息"AI 响应中断，点击重试" + 重试 Ghost（沿用上一条指令重新生成卡片） | 流式连接中断（streamInterrupted=true） |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 编辑字段 | onChange 更新卡片 state | 必填空字段补值后高亮消除 |
| 继续对话补充 | 后端解析增量 → diff 确认 → 更新卡片 state | 展示 diff 浮层供确认，不静默覆盖 |
| 点提交（必填全满） | 状态变更类先 available-transitions 预校验 → 通过则调 API | 切提交中态 |
| 点撤回（窗口内） | 调后端反向操作（状态变更类先重校验） | 切撤回成功 / 撤回失败（提示 validTransitions） |
| 点重试（失败态） | 状态变更类先重跑 available-transitions 预校验 → 通过则重新调提交 API | 切提交中态；预校验失败回校验失败态并刷新 validTransitions |
| 点跳转链接 | 跳转实体详情页 | 既有路由 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 操作徽章 | cardType {create/update/assign} | AI 意图 |
| 实体确认区 | targetEntity {title,bizCode} | AI 实体抽取 |
| 字段控件 | fieldSet Map<name,{value,required,derived}> | AI 抽取 + schema |
| 错误条 | validationError + validTransitions[] | available-transitions |
| 撤回按钮 | undoAvailable, undoDeadline, previousValue | 提交后状态 |
| 撤回倒计时显示 | undoCountdown "mm:ss" | 客户端 ticker：由 undoDeadline 倒推（每秒 tick），deadline 到达触发「撤回窗口过期」 |
| 提交权限 | canSubmit boolean | 前端 RBAC（targetEntity.bizKey + cardType） |

> **撤销态持久化**：`undoAvailable` / `undoDeadline` / `previousValue` 经 zustand `persist` middleware 写入 `sessionStorage`，key 为操作 id（`undo:{opId}`）。全局 overlay 在同会话页面导航时可能 unmount/remount，sessionStorage 保证 5 分钟撤回窗口在 remount 后仍可恢复（卡片重新挂载后由 opId 查表重建撤回按钮与倒计时）。会话结束（关闭标签页 / 登出）时 sessionStorage 随之清除，符合 PRD「同会话」语义。

#### diff 内联区（对话补充增量变更，无二级浮窗）

当用户在卡片编辑中继续对话补充，后端解析出对当前卡片的增量变更时，在**卡片体内**内联展示 diff 区（不弹浮窗），确认动作由输入区选项组承载（见 Component 2「输入区双模式」）。

- **定位**：内联于卡片字段区下方（非 absolute 浮层），随卡片正常排版、随消息区滚动，不会脱离锚点。
- **外观**：`.diff-inline`，accent-bg `#eff6ff`、border accent-ring `#bfdbfe`、rounded-md、p-2.5，flex column gap-1.5。
- **内容**：标题 13px 500 text accent-hover"对话补充将更新以下字段："+ 字段 diff 行（字段名 text-secondary · 旧值 text-tertiary 删除线 → 新值 text-primary 500）；底部 `.diff-hint` 12px text-tertiary"用下方输入区确认（↑↓ 选择 · Enter 确认）"。
- **确认动作**：卡片内不放按钮（避免二级浮窗/按钮冗余）；输入区切到选项组模式，选项 =「应用变更」「丢弃」（`Esc`=丢弃）。应用则按 last-write-wins 合并入卡片 state，丢弃则保留卡片当前值。
- **无浮层收益**：消除浮层与离开确认 Dialog 的 z-index 叠压问题；diff 随卡片滚动不脱离锚点；卡片交互保持扁平。

---

## Component 4: 查询结果卡片（查询操作）

### Placement

- **Mode**: existing-page
- **Target**: 渲染于 UF-2 聊天面板消息区内
- **Position**: AI 消息流中的一条卡片消息

### Layout Structure

卡片容器（surface、border、rounded-xl、shadow-1、p-4、flex column gap-2）：

1. **摘要行**：14px 500 text-primary（如"你有 3 个 P0 事项"）；进度查询附统计（completed/total + 蓝色进度条：track bg `#e2e8f0`、h-1.5（6px）、rounded-full、w-full；fill bg accent-light `#3b82f6`、width=`{percent}%`、transition-width 300ms）
2. **结果列表**：垂直堆叠，每项为"迷你卡片行"（hover bg-bg-alt、rounded-lg、p-2、flex between）：
   - 左：实体标题（13px text-primary）+ bizCode（12px text-tertiary）+ meta 徽章（priority/status badge）
   - 右：chevron-right 图标（text-tertiary）
3. **截断提示**（条件）：超过 20 张时底部 text-tertiary 12px"结果过多，显示前 20 条，请缩小范围（指定标题/负责人）"
4. **空态**：text-secondary 13px"未找到匹配事项"+ 建议（如"试试'我的 P1 事项'"）

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 有结果 | 摘要 + 迷你卡片列表 | 查询命中 |
| 无结果 | "未找到匹配事项"+ 建议 | 查询未命中 |
| 截断 | 前 20 条 + 截断提示 | 结果 >20 |

> **流式说明**：查询结果（UF-4）与歧义消解（UF-5）均**原子返回**（后端一次性返回完整结果，非增量流式），因此不适用 UF-3 的「流式填充中断」清理规则；UF-2 的「流式中断」态仅针对 UF-3 写卡片的骨架增量填充场景。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击迷你卡片行 | 跳转详情页（既有路由） | 导航 |
| hover 行 | bg-bg-alt | chevron 变 accent |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 摘要 | summaryText | AI/后端 |
| 迷你卡片行 | results[{bizKey,title,code,route,meta}] | 查询 API |
| 进度统计 | progress{completed,total,percent} | 查询 API（里程碑场景） |

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

1. **图标+标题**：气泡图标（accent）+ 14px 500 text accent-hover"我是 AI 助手，可以帮你："
2. **能力列表**：四行（flex、gap-2、13px text-primary），每行图标+文字：
   - 创建（事项/子任务/里程碑/里程碑图/ItemPool）
   - 查询（"我的 P0 事项"）
   - 修改（状态/进度）
   - 分配（负责人）
3. **示例指令**（chips，可点击填入输入框）：rounded-full border border-dark px-3 py-1 13px，如"创建一个 P1 事项"、"我的 P0 事项有哪些？"、"更新子任务完成度为 60%"
4. **关闭**：右上 Ghost X（收起引导卡片，本次会话不再展示）

### States

| State | Visual | Behavior |
|-------|--------|----------|
| 首次展示 | 引导卡片 + 示例 chips | 用户首次展开、本地无标记 |
| 已引导 | 不展示 | 本地标记已存在 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点示例 chip | 填入输入框（不自动发送） | 焦点留输入框 |
| 点 X / 开始输入 | 收起引导卡片 | 本地标记 hasSeenOnboarding=true |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 能力/示例 | 静态文案 | — |
| 已展示标记 | hasSeenOnboarding | 本地存储 |

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
| 已完成 | 全部步骤 done ✓，光标消失，下方渲染最终卡片 | **自动折叠**（思考+计划+操作执行完成即折叠为头部摘要），用户可点击重新展开 |
| 折叠 | 仅显示头部（步数·耗时） | 用户点击头部切换；折叠态记忆于会话内 |
| 步骤失败 | 该步 error ✗ + error-text | 中断后续步骤，引导用户（如"匹配失败，请补充信息"），不渲染最终卡片 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击头部 | 折叠/展开 | 箭头旋转 150ms |
| 点击操作行（可选） | 展开该步详情（请求/响应摘要） | 内联展开 |
| 流式完成 | 末步 done，光标消失 | 下方自动渲染最终卡片（UF-3/4） |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 思考文本 | trace.thinking | 流式事件 `thinking` |
| 计划步骤 | trace.plan[] | 流式事件 `plan_step` |
| 操作行 | trace.actions[{name,status,durationMs,detail}] | 流式事件 `tool_call` / `tool_result` |
| 摘要 | trace.summary{tokens, durationMs, costUsd} | 后端汇总 |

> **流式协议**：后端 AI 代理以 SSE/流式返回结构化事件序列 `thinking → plan_step* → tool_call/tool_result* → final_card`；前端按事件增量渲染 trace，最终卡片在 `final_card` 事件后渲染。**延迟目标**：首字节（思考出现）< 1s，计划可见 < 2s，最终卡片 P95 < 5s（见 PRD NFR）。
> **透明 vs 噪声**：trace 流式期间展开（建立信任、可调试），**思考+计划+操作执行完成即自动折叠**为头部摘要（让最终卡片成为视觉焦点）；用户可点击重新展开；展开/折叠态记忆于会话内。失败步骤高亮，便于用户理解"为何没出卡片"。
> **降级**：AI 不支持流式时，trace 退化为单条"AI 思考中…"系统消息 + 最终卡片（即原 UF-2 思考态）。

---

## Accessibility（跨组件，参照 WCAG 2.1 AA）

- **焦点管理**：展开面板焦点入输入框；Esc 焦点回气泡；卡片出现不抢断焦点，Tab 可达。
- **键盘**：所有卡片操作（编辑、选择、提交、重试、撤回）键盘可完成，无键盘陷阱；Enter 发送、Shift+Enter 换行、Esc 收起。
- **ARIA**：消息历史 `role="log"` `aria-live="polite"`；气泡 `aria-label="AI 助手（展开/收起）"`；面板 `role="dialog"` `aria-label`；卡片 `aria-label` 描述类型；错误态 `role="alert"`；Team 徽章 `aria-live="polite"`。
- **屏幕阅读器**：所有状态（思考/流式/错误/超时/成功/撤回）有 aria-live 播报；加载态有文字标签。
- **对比度**：正文 ≥4.5:1；大字（≥18.66px/500 常规，或 ≥14px/700 加粗，或 ≥24px）≥3:1。**14px/500 不属大字**。Component 6 警告标题与正文统一用 warning-text `#92400e`（≥4.5:1 on warning-bg `#fffbeb`），warning-title `#d97706` 仅用于图标（图形非文字，3:1 即可）。必填高亮用对比达标边框色，不仅靠颜色。
- **动效兜底（reduced-motion）**：所有动画（面板滑入 translateX、三点跳动思考态、骨架流式填充、进度条 width 过渡、卡片状态切换）在 `@media (prefers-reduced-motion: reduce)` 下降级为瞬时切换（无位移/无跳动，仅颜色或可见性变化）；状态信息不依赖动效传达（思考态附"AI 思考中…"文字，不仅靠跳动点）。

---

## Prototype 映射

原型文件结构（Step 8 生成于 `ui/prototype/`）：

| 文件 | 内容 | 对应组件 |
|------|------|----------|
| `index.html` | 入口 + 模拟主页面（侧边栏 + 内容区）+ 挂载浮动气泡 | UF-1 + 宿主页 |
| `panel.html` / 同页 overlay | 聊天面板展开态（含消息流、各类卡片） | UF-2/3/4/5/6/7 |
| `styles.css` | 共享 token + 组件类（复用 DESIGN.md） | 全部 |
| `app.js` | 交互（展开/收起、发送、卡片状态切换、示例填充） | 全部 |

各卡片状态（预填/编辑中/校验失败/成功/撤回/超时/降级/空/截断）均在原型中以可切换视图实现。
