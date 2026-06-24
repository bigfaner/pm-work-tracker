---
feature: "AI Copilot 对话助手"
---

# User Stories: AI Copilot 对话助手

> 覆盖四类角色（PM / 研发成员 / TL / ItemPool 提交者）与六个实体（MainItem / SubItem / Milestone / MilestoneMap / ProgressRecord / ItemPool），以及两个跨角色场景（歧义消解、超时降级）。

## Story 1: PM 通过自然语言创建主事项与里程碑

**As a** 项目经理 / PM
**I want to** 用一句话描述意图（如"创建一个 P1 事项，标题完成用户认证模块，分配给张三，下周五截止"），系统自动推送预填卡片让我确认
**So that** 我无需在 10+ 字段的表单中逐项填写，大幅降低创建操作成本

**Acceptance Criteria:**
- Given PM 在任意已认证页面，浮动气泡可见（首次有徽章提示）
- When PM 输入"创建一个 P1 事项，标题完成用户认证模块，分配给张三，下周五截止"并点击发送
- Then 单次调用从发送到卡片推送响应时间 ≤ 5 秒；推送的预填卡片字段值精确等于：title=`完成用户认证模块`、priority=`P1`、assignee=张三（Team 成员模糊匹配命中）、expectedEndDate=下周五日期、milestoneKey 字段渲染为空且带 `data-required-highlight` 属性
- Given PM 已发送创建指令，后端正在处理
- When AI 流式返回中间过程事件（思考/计划/工具调用）
- Then 聊天面板在用户指令与最终卡片之间流式展示 Agent 过程追踪（UF-8）：思考首条 ≤1s 出现、计划 ≤2s 可见、操作步骤随真实工具调用实时追加（每步含状态 ✓/✗ 与耗时），过程结束后在其下方渲染预填卡片；用户可点击追踪头部折叠/展开
- Given 预填卡片已推送
- When PM 直接编辑卡片修正 assignee 字段，或继续对话输入"优先级改成 P0"
- Then 两种方式均更新同一份卡片状态，卡片始终为唯一数据源，对话区域仅回显
- Given 卡片字段完整（无 `data-required-highlight` 字段为空）
- When PM 点击提交
- Then 系统调用现有 MainItem 创建 API 执行（创建操作不做 available-transitions 预校验），成功后聊天界面反馈创建结果
- Given PM 输入"加个里程碑，叫第一阶段验收"
- When 系统识别为 Milestone 创建意图
- Then 推送 Milestone 创建卡片，关联当前 Team 的 MilestoneMap（若 Team 下仅一个 MilestoneMap 自动关联，多个则 milestoneKey 字段渲染为空且带 `data-required-highlight` 属性）
- Given Team 下唯一的 MilestoneMap 状态为 `completed` 或 `cancelled`（terminal）
- When PM 输入"加个里程碑，叫第一阶段验收"
- Then 系统不推送 Milestone 创建卡片（依据 BIZ-milestone-005：Map terminal 时禁止子 Milestone 创建/更新），返回文字提示"目标里程碑图已完成/取消，无法新增里程碑，请选择其他里程碑图或先重启里程碑图"
- Given Milestone 创建卡片字段完整（title + parent MilestoneMap 已填）
- When PM 点击提交
- Then 系统调用现有 Milestone 创建 API 执行（Milestone 创建操作不做 available-transitions 预校验），成功后聊天界面反馈创建结果

---

## Story 2: 研发成员创建子任务并更新进度

**As a** 研发成员
**I want to** 通过自然语言在指定主事项下创建子任务，并更新子任务完成度
**So that** 我执行任务时无需离开工作上下文去翻找父事项和填写进度表单

**Acceptance Criteria:**
- Given 研发成员在主事项详情页，当前 Team 上下文已自动检测
- When 研发成员输入"在这个事项下加一个子任务，标题写接口联调"
- Then 系统识别父 MainItem（来自页面上下文），推送 SubItem 创建卡片，parent 字段已预填
- Given 父事项页面上下文不可用
- When 研发成员输入"在用户认证模块事项下加一个子任务"
- Then 系统按标题模糊匹配父 MainItem，命中唯一则预填 parent，命中多个则推送歧义消解卡片
- Given 子任务已存在且状态为进行中
- When 研发成员输入"更新子任务完成度为 60%"
- Then 系统推送 ProgressRecord 进度更新卡片，含 subItemKey、completion=60、achievement 字段
- Given ProgressRecord 卡片字段完整且用户有 SubItem 写权限
- When 研发成员点击提交
- Then 系统调用现有 ProgressRecord 创建 API 执行（ProgressRecord 无 available-transitions 端点，不做预校验），成功后聊天界面反馈"进度已更新至 60%"
- Given 研发成员输入"把用户认证模块的状态改为已完成"，该 MainItem 当前状态为 `in_progress`（不允许直接转 `completed`）
- When 系统对目标状态做 available-transitions 预校验
- Then 预校验返回失败，错误响应 payload 含 `validTransitions: ["paused", "cancelled"]` 数组列出当前合法目标状态；卡片内显示错误说明"当前状态不支持直接转为已完成，合法目标状态为：paused、cancelled"；不调用后端写 API
- Given 研发成员输入"把子任务接口联调移到用户认证模块事项下"，源子任务状态非 terminal、源与目标 MainItem 同 Team 且为不同实体、目标 MainItem 非 terminal
- When 研发成员确认提交
- Then 系统调用现有 SubItem move API 执行（依据 BIZ-lifecycle-003/004：保留 item_status 与 assignee_key，仅更新 main_item_key 与 item_code）；成功后聊天界面反馈"子任务已移动"

---

## Story 3: TL 查询事项、创建里程碑图并调整分配

**As a** 团队负责人 / TL
**I want to** 通过自然语言查询事项进度、创建里程碑图、调整负责人分配
**So that** 我监控与调配工作时无需在多个页面间切换

**Acceptance Criteria:**
- Given TL 在任意已认证页面
- When TL 输入"我的 P0 事项有哪些？"
- Then 系统按 BIZ-filter-001 规则返回**直接 + 间接**匹配：包含 assignee 字段为该 TL 的 MainItem（直接），以及 assignee 为该 TL 的 SubItem 所属的 MainItem（间接）；状态筛选（P0）与 assignee 筛选 AND 组合；返回摘要文字（含总数）+ 可点击的事项卡片列表
- Given 查询结果卡片已展示
- When TL 点击某张事项卡片
- Then 跳转到该事项的详情页 `/items/:mainItemId`
- Given TL 输入"里程碑图第二阶段的进度怎么样？"
- When 系统识别为 Milestone 查询（经统一查询处理器，与 MainItem 查询同机制）
- Then 返回摘要文字（含子事项完成率：completed/total/percent）+ 可点击的 Milestone 卡片，点击跳转 `/milestones/:mapId`
- Given TL 输入"我提交的 ItemPool 申请有哪些？"
- When 系统识别为 ItemPool 查询（经统一查询处理器，与 MainItem 查询同机制）
- Then 返回摘要文字 + 可点击的 ItemPool 卡片列表，点击跳转 `/item-pool`；查询支持全部 6 实体，MainItem 查询为本故事的参考 AC，其余实体走同一查询处理器
- Given TL 输入"创建一个里程碑图，叫第二阶段"
- When 系统识别为 MilestoneMap 创建意图
- Then 推送 MilestoneMap 创建卡片，Team 字段自动预填当前 Team 上下文
- Given MilestoneMap 创建卡片字段完整（title 已填）
- When TL 点击提交
- Then 系统调用现有 MilestoneMap 创建 API 执行（MilestoneMap 创建操作不做 available-transitions 预校验），成功后聊天界面反馈创建结果
- Given TL 输入"把用户认证模块分配给李四"
- When 系统解析为 assignee 字段更新
- Then 推送负责人变更确认卡片，TL 确认后调用现有 MainItem 更新 API 执行

---

## Story 4: ItemPool 提交者（非技术人员）申请事项

**As a** ItemPool 提交者（含非技术人员）
**I want to** 用自然语言描述我想做的事项背景和预期产出，系统帮我填好结构化申请卡片
**So that** 我不熟悉结构化字段也能独立提交申请，无需他人协助

**Acceptance Criteria:**
- Given ItemPool 提交者在 `/item-pool` 页面或任意页面打开 Copilot
- When 提交者输入"我想申请做一个客户导出功能，背景是销售每周手动导出很耗时，预期产出是一个一键导出按钮"
- Then 系统解析为 ItemPool 创建意图，推送预填卡片：title=客户导出功能、background=销售每周手动导出很耗时、expectedOutput=一键导出按钮
- Given 预填卡片已推送，部分结构化字段 AI 无法推导
- When 提交者继续对话补充"优先级 P2"或直接编辑卡片
- Then 两种方式均写入同一卡片状态
- Given ItemPool 卡片字段完整且提交者有 `item_pool:create` 权限
- When 提交者点击提交
- Then 系统调用现有 ItemPool 创建 API 执行（ItemPool 无 available-transitions 端点，不做预校验），成功后聊天界面反馈"申请已提交"
- Given 提交者无 ItemPool 创建权限
- When 提交者尝试提交
- Then 系统返回明确的权限不足提示，不执行创建

---

## Story 5: 实体引用歧义时消解（跨角色）

**As a** 任意角色用户
**I want to** 当我的自然语言引用匹配到多个候选实体时，系统列出候选让我选择，而不是猜错
**So that** 操作不会作用到错误的事项上

**Acceptance Criteria:**
- Given Team 下存在多个标题含"认证"的事项
- When 用户输入"把认证模块的状态改为进行中"
- Then 系统按标题模糊匹配命中多个候选，推送歧义消解卡片列出所有候选实体（标题 + 编号）
- Given 歧义消解卡片已展示
- When 用户选择其中一个候选
- Then 系统以所选实体继续后续流程（推送状态变更确认卡片）
- Given 实体可精确匹配（bizKey）
- When 用户输入含明确编号的引用
- Then 系统直接精确匹配，不推送歧义消解卡片

---

## Story 6: AI 响应超时或不可用时降级（跨角色）

**As a** 任意角色用户
**I want to** 当 AI 响应超时或服务不可用时，系统提示我并提供传统表单入口，而不是让我一直等
**So that** 我的操作不被阻断，仍能完成任务

**Acceptance Criteria:**
- Given 用户已发送自然语言指令
- When AI 服务超过 10 秒未返回
- Then 聊天面板展示超时提示，并提供传统表单快捷入口（如"去手动创建事项"）
- Given AI 服务整体不可用
- When 用户打开 Copilot 面板
- Then 面板展示降级提示，说明 AI 暂不可用，用户可使用传统表单操作
- Given 页面导航前存在未提交的卡片
- When 用户尝试离开当前页面
- Then 系统弹出确认提示"当前有未提交的操作，确定离开？"

---

## Story 7: 边界与异常路径（跨角色）

**As a** 任意角色用户
**I want to** 在输入超长、配额耗尽、AI 解析异常、并发编辑、会话轮数达上限等情况下获得明确反馈，而非系统卡死或静默失败
**So that** 我能理解系统状态并采取正确行动

**Acceptance Criteria:**
- Given 用户输入超过 1000 字符
- When 用户点击发送
- Then 输入被截断至 1000 字符，聊天界面提示"已超出单次最大长度 1000 字符，已截断"，截断后的内容正常发送
- Given 用户当日 AI 调用次数已达 50 次
- When 用户发送新指令
- Then 系统返回降级提示"今日 AI 调用已达上限（50 次），已切换为关键词匹配模式"，关键词匹配模式下不做意图识别，仅按关键词命中返回提示或拒绝
- Given AI 返回的 JSON 结构无法解析（malformed）或字段不符合 schema
- When 后端代理解析失败
- Then 系统不推送卡片，返回"AI 解析失败，请重新描述意图或使用传统表单"+ 传统表单快捷入口
- Given 用户正在直接编辑卡片字段，同时通过对话发送补充指令
- When 两路写入同一字段
- Then 系统以**时间戳晚者胜出**合并（最后写入胜出，last-write-wins），且对话补充产生的增量变更在应用到卡片前先展示 diff 供用户确认（不直接覆盖正在编辑的字段）
- Given 用户展开聊天面板后未输入任何指令，直接点击页面其他位置或导航
- When 卡片区无未提交卡片
- Then 面板正常收起/导航，不弹离开确认（仅当存在 UF-3/UF-5 未提交卡片时才弹确认）
- Given 用户当前不在任何 Team 上下文页面（如 `/users` 全局页面）
- When 用户尝试在 Copilot 发送写操作指令
- Then 发送被阻止，输入框上方提示"请先进入具体 Team 页面后再执行操作"
- Given 会话已达到 50 轮上限
- When 用户尝试发送第 51 轮指令
- Then 发送被阻止，提示"当前会话已达 50 轮上限，请点击'开启新会话'"
- Given AI 返回的 bizKey 对应实体已被删除（数据库查询返回不存在）
- When 后端校验该引用
- Then 返回提示"该实体已不存在，请重新描述或使用传统表单"，不推送可执行卡片
- Given 用户输入一条指令，AI 识别置信度 = 0.7（高置信下界）
- When 后端返回结果
- Then 系统按高置信处理，直接推送预填卡片（不返回候选意图列表）
- Given 用户输入一条指令，AI 识别置信度 = 0.69（中置信上界）
- When 后端返回结果
- Then 系统按中置信处理，返回引导文字 + 候选意图列表供用户选择/澄清，不直接推送预填卡片
- Given 用户输入一条指令，AI 识别置信度 = 0.4（中置信下界）
- When 后端返回结果
- Then 系统按中置信处理，返回引导文字 + 候选意图列表
- Given 用户输入一条指令，AI 识别置信度 = 0.39（低置信上界）
- When 后端返回结果
- Then 系统按低置信处理，返回"无法理解，请重新描述" + 支持的操作类型，不推送卡片

---

## Story 8: 高影响写操作的撤回（跨角色）

**As a** 任意角色用户
**I want to** 在误操作分配或状态变更后，能在 5 分钟内一键撤回
**So that** AI 选错实体或我误确认时，能快速回滚，避免错误固化

**Acceptance Criteria:**
- Given 用户成功提交了一次可逆操作（分配，或非 terminal 状态变更），`undoAvailable = true`
- When 系统返回成功反馈
- Then 成功卡片显示目标实体 title + bizCode 二次确认已通过，并展示"撤回"按钮 + `undoDeadline`（提交时间 + 5 分钟）倒计时
- Given 撤回按钮可见且未过期
- When 用户点击撤回
- Then 系统调用后端反向操作：分配 → assignee 恢复为 `previousValue.assignee`；非 terminal 状态变更 → status 恢复为 `previousValue.status`；撤回成功后卡片更新为"已撤回"，撤回按钮消失
- Given 撤回为状态变更类，恢复原状态前重新调用 available-transitions 校验，当前状态已不允许回到原状态
- When 用户点击撤回
- Then 撤回失败，卡片显示"当前状态无法回到原状态，合法目标状态为：{validTransitions}"，不执行反向操作
- Given 撤回窗口已过 `undoDeadline`（提交后 > 5 分钟）
- When 用户查看原成功卡片
- Then 撤回按钮已消失，卡片显示"撤回窗口已过期"
- Given 用户在撤回窗口内收起聊天面板或导航到同会话其他页面，随后重新展开/返回，且仍在 `undoDeadline` 之前
- When 用户查看原成功卡片
- Then 撤回按钮仍可用
- Given 用户关闭浏览器或登出（会话结束）后重新进入
- When 查看该操作
- Then 撤回已失效（跨会话持久化不在 v1 范围），卡片显示"撤回窗口已过期"
- Given 撤回窗口内 AI 服务不可用
- When 用户点击撤回
- Then 撤回仍成功执行（撤回调用现有实体 API 反向操作，不依赖 AI 服务）
- Given 同一操作已被撤回
- When 用户尝试对该操作再次撤回
- Then 不允许（每次操作仅一次撤回）；用户需重新发起正向操作
- Given 用户提交的是不可逆操作（状态转入 terminal `completed` 或 `cancelled`），`undoAvailable = false`
- When 系统返回成功反馈
- Then 卡片明确标注"该操作不可撤回"，不展示撤回按钮

---

<!-- Coverage matrix:
- Roles:   PM (S1) | Dev (S2) | TL (S3) | ItemPool submitter (S4) | cross-role (S5, S6, S7, S8)
- Ops:     create (S1,S2,S3,S4) | query (S3 — unified handler covers all 6 entities; MainItem/Milestone/ItemPool evidenced) | modify (S2) | assign (S3) | move (S2) | undo (S8)
- Entities: MainItem (S1,S3) | SubItem (S2) | Milestone (S1,S3) | MilestoneMap (S3) | ProgressRecord (S2) | ItemPool (S3,S4)
- Full lifecycle (submit->success): MainItem (S1) | SubItem (S2) | Milestone (S1) | MilestoneMap (S3) | ProgressRecord (S2) | ItemPool (S4)
- Edge:    disambiguation (S5) | timeout/fallback (S6) | state-machine guard (S2) | permission guard (S4) | input overflow (S7) | quota (S7) | malformed AI (S7) | concurrent edit (S7) | abandoned card (S6) | team-missing (S7) | 50-round cap (S7) | stale bizKey (S7) | Map-terminal block (S1) | sub-item move (S2) | confidence boundary 0.7/0.69/0.4/0.39 (S7) | undo success/expiry/session-loss/AI-down/uniqueness/irreversible (S8)
-->
