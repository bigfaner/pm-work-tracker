---
created: "2026-06-26"
updated: "2026-06-30"
parent: tech-design.md
---

# Agent Architecture: Planner & Executors

> 返回 [`tech-design.md`](./tech-design.md)

## 1. 设计原则

**职责分离**：
- **Planner**：理解用户意图、必要时澄清、生成可执行计划。**不做实际写操作**。
- **Executor**：执行某一类操作（create/query/update/move）。每个 Executor 有独立 system prompt + 工具集。

**线性消息流 + IntentID 分组**：
- 一个 turn 内可能含多个 intent（用户一句话多意图）
- 每个 intent 的产出（trace + text + card）通过 `IntentID` 聚合
- 不采用树形（避免嵌套渲染复杂度），仍是扁平消息列表 + intent 头部视觉分组

**结构化输出优先走 function calling（emission 工具），带自然结束兜底**：
- 所有结构化输出（intent / form card / query result / disambig / input_rewrite）**首选**通过 emission 工具调用产生——强类型、可程序调度、无需文本解析
- 文本通道仅用于 thinking（推理过程），system prompt 禁止输出 `<message type="X">` 标签或裸 JSON
- **兜底（safety net）**：若 LLM 未遵守、一轮结束未调任何 emission 工具，Agent 会尝试按该 Agent 的 output schema 解析累积 content（见 [llm-integration.md](./llm-integration.md) §2.4）；解析失败才降级为纯文本回复。这样把"LLM 偶尔忘记调 emission 工具"的失败面从"一定丢结构化"降为"仅当 content 也不可解析时才丢"

## 2. 工具三分类（Read / Action / Emission）

每个 Agent 持有的工具分三类，决定 Agent 循环如何处理：

| 类别 | 行为 | 流后处理 | 示例 |
|------|------|---------|------|
| **Read** | 纯查询，返回数据给 LLM | append `role=tool` 到 messages，继续循环 | `query_team_schema`, `fuzzy_match_member` |
| **Emission** | 结构化输出，写 SSE 事件到 outCh（含预填表单 card）；**不触 DB 写** | **不 append，终止流**（StreamRun 结束） | `submit_rewrite`, `submit_intent`, `emit_form_card`, `emit_query_result`, `emit_disambig` |

**关键：写操作不发任何"Action"工具**。Executor 只组装字段 + 调 `emit_form_card` 推送"预填表单卡片"（in-memory only）。真正的 DB 写发生在 `commit_card` Handler 路径，调用 Dispatcher → 现有 entity service（见 [`request-model.md`](./request-model.md) §6.1 与 [`security.md`](./security.md) §7.3）。这样保证"AI 不直接写库"在物理层面成立，且用户放弃提交时无孤儿实体。

### 2.1 Emission 工具语义

- Emission 工具是 Agent 产出的**首选通道**，调用即**终止**——一旦调用，Agent 不再继续 LLM 循环
- Agent 有**三条结束路径**：(a) 调到 terminal emission 工具（首选，强类型产出）；(b) 一轮 LLM 结束未调任何工具（自然结束），此时 Agent 尝试按该 Agent 的 output schema 解析累积 content 作为兜底产出，解析失败才降级为纯文本（见 [llm-integration.md](./llm-integration.md) §2.4）；(c) 超过 max_iterations 兜底失败
- Emission 工具直接把 args 渲染为对应 SSE 事件，**所有 emit_* 工具仅写 SSE 事件 + persist 卡片消息到 copilot_messages**，绝不调用 entity service 的 Create/Update/Move：
  - `submit_rewrite` → `input_rewrite` 事件（仅 Planner，作为第 1 步）
  - `submit_intent` → `card_message(cardType=intent)` 事件（仅 Planner，终止）
  - `emit_form_card` → `card_message(cardType=form, status=prefilled)` 事件（Writer/Updater/Mover，终止）；**targetEntity.bizKey 留空**（待 commit_card 阶段由 entity service 生成后回填）
  - `emit_query_result` → `card_message(cardType=query_result)` 事件（Reader，终止）
  - `emit_disambig` → `card_message(cardType=disambig)` 事件（任一 Executor，终止）

### 2.2 工具分类总表

| 工具 | 类别 | Planner | Writer | Reader | Updater | Mover |
|------|------|---------|--------|--------|---------|-------|
| `query_team_schema` | Read | ✓ | ✓ | ✓ | ✓ | |
| `query_team_members` | Read | ✓ | | | | |
| `query_team_milestones` | Read | ✓ | | | | |
| `fuzzy_match_member` | Read | ✓ | ✓ | | ✓ | |
| `fuzzy_match_milestone` | Read | ✓ | ✓ | | ✓ | |
| `query_entities` | Read | | | ✓ | | ✓ |
| `validate_transition` | Read | | | | ✓ | |
| `validate_source_target` | Read | | | | | ✓ |
| `submit_rewrite` | Emission | ✓ | | | | |
| `submit_intent` | Emission | ✓ | | | | |
| `emit_form_card` | Emission | | ✓ | | ✓ | ✓ |
| `emit_query_result` | Emission | | | ✓ | | |
| `emit_disambig` | Emission | | ✓ | | ✓ | ✓ |

**已删除 `commit_create` / `commit_update` / `move_sub_item` 工具**。真实 DB 写由 `commit_card` Handler 路径触发 Dispatcher（见 [`request-model.md`](./request-model.md) §6.1）。Updater/Mover 在推送 form card 前可调 Read 工具 `validate_transition` / `validate_source_target` 做预校验，结果写入 cardData.errors 但不阻断流（最终由用户在 form card 上调整后提交时再次校验）。

## 3. Planner Agent

### 3.1 职责边界

| Planner **要做** | Planner **不做** |
|-----------------|-----------------|
| 识别所有意图（含多意图） | 调用任何写工具（写操作无 Action 工具，见 §2） |
| 必要时发意图回执 + 澄清问题 | 调用 query_entities（让 reader 做） |
| 拆解为 step 计划 | 调用 validate_transition（让 updater 做） |
| 标注每 step 用哪个 Executor | 渲染 form / query_result 卡片（让 Executor 做） |
| input_rewrite（代词消解、省略补全） | 持久化（由 Orchestrator 负责） |

### 3.2 Planner System Prompt 结构

```
你是 PM Work Tracker Copilot 的 Planner。

## 你的职责
理解用户意图、必要时澄清、生成可执行计划。

## 能力范围
- create：MainItem / SubItem / Milestone / MilestoneMap / ProgressRecord / ItemPool
- query：上述 6 实体
- update：上述 6 实体的字段（含 assignee）
- move：SubItem.move 特殊操作

## Executor 路由表
| op_type | executor |
|---------|----------|
| create  | writer   |
| query   | reader   |
| update  | updater  |
| move    | mover    |

## 输出方式：必须通过 function calling

你的所有结构化输出**只能**通过调用工具产生。文本通道只用于 thinking（推理过程），
禁止在文本里输出 <message type="X"> 标签或裸 JSON。

### 第 1 步：调用 submit_rewrite（必做，最早）

对用户输入做代词消解、省略补全、多意图拆分。

改写规则：
1. 代词（"它"、"那个"）→ 用 Environment.pageContext 或 DraftState 解析
2. 省略主语 → 从最近 turn 推断
3. 多意图拆分："X 然后 Y" → rewritten 列出多个完整意图
4. 相对时间锚定："下周五" → 用 Environment.currentTime 计算依据（不展开绝对日期）
5. 简称保留："pm" 等歧义简称不强行展开，列入 ambiguity_notes

submit_rewrite 参数：
- original: 用户原文（回显）
- rewritten: 语义完整版
- context_used: 引用的上下文项
- ambiguity_notes: 歧义提示

### 第 2 步：thinking（可选）

如需查询 Team 信息（schema / 成员 / 里程碑），调用对应的只读工具。
不要把工具结果原样复述——thinking 通道只用于内部推理。

### 第 3 步：调用 submit_intent（必做，最后）

推送意图消息（含文本回执 + 结构化 intents）。这是终止信号——调用后 Planner 流结束。

submit_intent 参数 schema：
- text: 自然语言意图回执（如"好的，我帮你创建一个 P1 主事项「认证模块」。"）
- intents[]: 意图规格数组
  - id: intent_1, intent_2, ...
  - label: "创建 MainItem"
  - opType: create / query / update / move
  - entityType: main_item / sub_item / ...
  - fields[]: 字段列表（每个含 name / value / required）
  - executor: writer / reader / updater / mover
- missingInfo[]: 缺失字段（如有，每个含 intentIndex / field / question / hint）
- state: awaiting_confirm 或 info_complete（澄清收齐后）
- decision: 路由决策枚举 confirm / show_candidates / cannot_understand（见 §3.4，Orchestrator 据此分流而非用浮点阈值）

## 重要规则

- **结构化输出只走工具**——submit_rewrite 和 submit_intent 是必调工具
- **抽取字段前必查 schema**——对每个识别到的 entityType，submit_intent 前必须先调 `query_team_schema(entity_type)` 获取字段名/类型/必填，据此抽取 `intents[].fields` 与 `missingInfo`。禁止在不知道 schema 的情况下盲抽字段（否则字段名与下游 Executor / entity-schemas 不一致）。ContextBuilder 不为 Planner 自动注入 schema（见 [interfaces.md](./interfaces.md) §5.3），故必须显式查询
- **意图消息是 source of truth**——持久化到 messages 表，用户确认后从中重建 plan
- 改写版本仅用于本次调用，不会持久化到对话历史；用户对话历史永远保留 original
- 一次指令含多意图时，全部识别并列在 intents[]
- 任意意图必填缺失时，在 missingInfo 列出所有缺失字段（一轮问全）
- 用户回答后重新进入 Planner（注入历史 + 回答），推送新意图消息
- **不直接执行意图**——submit_intent 之后等用户点"✓ 理解正确"再执行
```

### 3.3 Planner 工具集

```go
// Read tools（查询用）
plannerReadTools := []ToolDef{
    {Name: "query_team_schema",     Description: "查当前 Team 的实体 schema（每个识别到的 entityType 必调，抽字段前先查）", Params: {"entity_type": "string"}},
    {Name: "query_team_members",    Description: "查 Team 成员列表",          Params: {"filter": "string"}},
    {Name: "query_team_milestones", Description: "查 Team 里程碑列表",        Params: {"filter": "string"}},
    {Name: "fuzzy_match_member",    Description: "姓名/工号模糊匹配成员",     Params: {"name": "string"}},
    {Name: "fuzzy_match_milestone", Description: "名称模糊匹配里程碑",        Params: {"name": "string"}},
}

// Emission tools（结构化输出，强制必调）
plannerEmissionTools := []ToolDef{
    {Name: "submit_rewrite", Description: "改写用户输入（必做最早调用）",       Params: submitRewriteSchema},
    {Name: "submit_intent",  Description: "推送意图消息（必做最后调用，终止流）", Params: submitIntentSchema},
}
```

→ Planner **没有写工具**（`commit_create` 等），也**没有 form/query_result emission**——物理隔离职责。

### 3.4 路由决策（PRD Story 7，离散 decision 驱动）

Planner 在 `submit_intent` 的 `decision` 字段输出一个**离散路由决策**（枚举），Orchestrator 在 persist 意图消息**之前**按 decision 分流：

| decision | 处理 | 前端表现 |
|----------|------|---------|
| `confirm` | persist 意图消息（state=awaiting_confirm） + emit `card_message(intent)` + emit `turn_phase_done(nextAction=await_confirm_intent)` | 展示意图确认卡片（PRD：直接推送预填候选意图，不返回候选列表） |
| `show_candidates` | Orchestrator **不 persist 意图消息**；emit `card_message(cardType=candidate_list)` + 引导文字 + emit `turn_phase_done(nextAction=await_select_intent)` | 展示候选意图列表（每项含 label + 字段摘要），用户点选 → `POST /messages {type:select_intent, candidateIntentId}` → 转 `confirm_intent` 流程 |
| `cannot_understand` | Orchestrator emit `text_message("无法理解，请重新描述")` + `turn_phase_done(nextAction=retry)`，列出支持的操作类型清单 | 展示失败提示 + 操作类型清单 |

**为什么用离散 decision 而非浮点 confidence**：LLM 自报的 confidence 数值校准极差——同一模型内 confidence 与正确性常常弱相关甚至负相关，跨模型更不可比。用 0.7/0.4 等魔法阈值驱动路由既难写测试验证、换模型又要重调。让 LLM 直接判"我该确认 / 该列候选 / 没听懂"这种离散语义判断，比让它给一个校准度不可信的概率分数更可靠，也消除了阈值魔法数。这与 LangGraph 的 `Literal[...]` 条件边路由、Pydantic-AI 的 union 输出类型同源——路由决策应是离散字段，不是带阈值的浮点数。

> `decision` 取代了早期设计的 `confidence: float64`（0.0–1.0 + 0.7/0.4 阈值）。PRD Story 7 原 AC 以 confidence 边界（0.7/0.69/0.4/0.39）表述，本设计相应改为 decision 枚举边界——见 [`tech-design.md`](./tech-design.md) §1.5 PRD 偏离说明。confidence 作为可选观测元信息保留在 `agent_call_logs`（调试用），不再驱动路由。

**边界 AC 对齐（PRD Story 7，从 confidence 边界改为 decision 边界）**：

| AC | decision | 路径 |
|----|----------|------|
| 清晰意图 | `confirm` | intent card |
| 模糊但可列候选 | `show_candidates` | candidate_list |
| 无意义 / 超出能力 | `cannot_understand` | "无法理解" + 操作清单 |

决策枚举在 `internal/copilot/orchestrator/decision.go`：

```go
type PlannerDecision string
const (
    DecisionConfirm          PlannerDecision = "confirm"
    DecisionShowCandidates   PlannerDecision = "show_candidates"
    DecisionCannotUnderstand PlannerDecision = "cannot_understand"
)
```

**candidate_list 卡片**（cardType）：CardPayload 含 `candidates: [{id, label, opType, entityType, fieldsSummary}]`。用户点选后，Orchestrator 把该候选升级为正式 IntentSpec（生成 intent_message_id）并进入 `confirm_intent` 流程。state-machines.md §3 新增状态 `awaiting_select_intent`（介于 planning 与 awaiting_confirm_intent 之间）。

**为什么 show_candidates 不直接进 confirm_intent**：此决策下用户尚未确认"是这个意图"，直接展示确认卡片会让用户误以为 AI 已理解。candidate_list 强制用户从候选中选一个，避免误执行。

**单测场景**（testing-strategy.md §3.1 补充）：

| 输入 | decision | 期望路径 |
|------|----------|---------|
| "创建 P1 事项叫认证模块"（清晰） | `confirm` | intent card |
| "改一下那个东西的状态"（含代词+省略） | `show_candidates` | candidate_list |
| "asdfgh"（无意义） | `cannot_understand` | "无法理解" |

## 4. Executor Agents（按 op_type 抽象，4 个）

| Executor | 处理范围 | Read 工具 | Emission 工具（终止） |
|----------|---------|----------|----------------------|
| **writer** | 6 实体的 create | `query_team_schema`, `fuzzy_match_member`, `fuzzy_match_milestone` | `emit_form_card`, `emit_disambig` |
| **reader** | 6 实体的 query | `query_team_schema`, `query_entities` | `emit_query_result` |
| **updater** | 6 实体的 update / assignee / 状态变更 | `query_team_schema`, `fuzzy_match_*`, `validate_transition` | `emit_form_card`, `emit_disambig` |
| **mover** | SubItem.move | `query_entities`, `validate_source_target` | `emit_form_card`, `emit_disambig` |

**Executor 无任何 Action/写工具**——所有 emit_form_card 仅生成预填表单（`status=prefilled, targetEntity.bizKey=""`）。真实 DB 写由用户点提交后 `commit_card` Handler 调 Dispatcher 触发现有 entity service（[`request-model.md`](./request-model.md) §6.1 请求 3）。`validate_transition` / `validate_source_target` 是 Read 工具，预校验结果填入 `cardData.errors` 供前端展示但不阻塞流。

**为什么不按实体拆 6 个**：
- 复用度高（字段抽取、模糊匹配、必填校验逻辑共享）
- 维护成本低（4 套 system prompt vs 6 套）
- system prompt 内按 entity_type 动态注入 schema（按需，token 不浪费）

## 5. Writer Executor System Prompt 示例

```
你是 PM Work Tracker 的创建专家 Executor。

## 当前任务
- op_type: {{op_type}}
- entity_type: {{entity_type}}
- 完整字段：{{fields}}

## 实体 Schema（按需注入）
{{entity_schema}}

## 输出方式：function calling

文本通道只用于 thinking。所有结构化输出走工具。

## 你的工具

### Read
- query_team_schema(entity_type) → 查 schema 补充知识
- fuzzy_match_member(name) → 姓名转 user_bizkey（多候选时返回候选列表）
- fuzzy_match_milestone(name) → 名称转 milestone_bizkey

### Emission（终止流）
- emit_form_card(cardData) → 推送**预填表单卡片**（targetEntity.bizKey 留空，等用户提交后由 entity service 生成）
- emit_disambig(candidates[]) → 多候选时让用户选

**注意：你没有任何写工具。** 真正的 MainItemService.Create 由用户在 form card 上点"提交"后由后端 commit_card 路径触发，不在这条 LLM 流里。

## 流程

1. 字段中如有姓名/里程碑引用，先调 fuzzy_match_* 转换为 bizKey
2. 若 fuzzy_match_* 返回多候选 → 调 emit_disambig，终止流（等用户 select_candidate）
3. 字段齐备后调 emit_form_card 推送**预填表单卡片**（cardData 含 opType / entityType / fields，但 targetEntity.bizKey 为空字符串），流结束
4. 用户在 UI 上确认/编辑后点提交 → 后端 commit_card Handler 调用 MainItemService.Create → bizKey 回填到 followup 消息

## 重要规则

- 卡片 derived 字段标 derived=true
- 创建操作不调 validate_transition（无源状态）
- 卡片状态默认 prefilled，等待用户在 UI 上提交
- **targetEntity.bizKey 必须留空**——此卡片是预填草稿，不是已创建实体的回显
- **禁止在文本里输出 JSON 或 <message> 标签**
```

## 6. Agent 接口定义

```go
// internal/copilot/agent/agent.go

type Agent interface {
    Role() AgentRole
    SystemPrompt(p AgentRunParams) string
    Tools() []ToolDef
    StreamRun(ctx context.Context, p AgentRunParams) (<-chan sse.Event, error)
}

type AgentRole string
const (
    RolePlanner AgentRole = "planner"
    RoleWriter  AgentRole = "writer"
    RoleReader  AgentRole = "reader"
    RoleUpdater AgentRole = "updater"
    RoleMover   AgentRole = "mover"
)

type AgentRunParams struct {
    SessionID  string
    TurnID     string
    StepID     string                  // planner 时 = turnID；executor 时 = intent.id

    // Planner 输入
    UserMsg    string                  // 自由文本（仅 planner）

    // Executor 输入
    StepParams map[string]any          // plan.intent.params

    // 共享上下文
    History    []MessageSnapshot       // 滑动窗口裁剪后的历史
    DraftState *DraftState             // 当前 turn 草稿
    Env        Environment             // user / team / page / currentTime

    // Executor 续跑（select_candidate 场景）
    InjectedBizKey string              // 选定候选后注入到 StepParams

    // ── 累加器回调（由 RequestState 投影器 ForPlanner/ForExecutor 注入；Agent 不直接写库）──
    Persist     func(msg Message) (string, error) // stage 消息 → rs.PendingMessages，返回预生成 bizKey
    OnAgentCall func(log AgentCallLog)            // 累积 LLM 调用元数据 → rs.Calls
}
```

> **AgentRunParams 是 `RequestState` 的投影**（消除与 TurnContext 的字段漂移）。不要在调用点手写 `AgentRunParams{...}` 字面量，改用 `RequestState` 的构造器（定义见 [state-model.md](./state-model.md) §6）：
>
> ```go
> func (rs *RequestState) ForPlanner(userMsg string) AgentRunParams                              // UserMsg + Env + History + Draft
> func (rs *RequestState) ForExecutor(intent IntentSpec, injectedBizKey string) AgentRunParams  // StepParams + Env + History
> ```
>
> [request-model.md](./request-model.md) §5.2/§5.3 的内联 `AgentRunParams{...}` 构造改为调这两个方法。

## 7. Agent Registry

```go
// internal/copilot/agent/registry.go

type Registry struct {
    planner   Agent
    executors map[string]Agent
}

func NewRegistry(provider llm.Provider, deps Deps) *Registry {
    return &Registry{
        planner: NewPlanner(provider, deps),
        executors: map[string]Agent{
            "writer":  NewWriterExecutor(provider, deps),
            "reader":  NewReaderExecutor(provider, deps),
            "updater": NewUpdaterExecutor(provider, deps),
            "mover":   NewMoverExecutor(provider, deps),
        },
    }
}

func (r *Registry) Executor(name string) (Agent, error) {
    e, ok := r.executors[name]
    if !ok {
        return nil, fmt.Errorf("unknown executor: %s", name)
    }
    return e, nil
}
```

## 8. Tool 实现

### 8.1 Read 工具

| 工具 | 实现 | 说明 |
|------|------|------|
| `query_team_schema` | schema_loader.Load(entity_type) | 从 entity-schemas 常量加载 |
| `query_team_members` | TeamMemberRepo.List(team_id, filter) | 复用现有 repository |
| `query_team_milestones` | MilestoneRepo.ListByTeam(team_id, filter) | 复用现有 repository |
| `query_entities` | Dispatcher.DispatchQuery(entity_type, filter) | 调用现有 list API |
| `fuzzy_match_member` | FuzzyMatcher.MatchMember(name) | 简单字符串匹配 + 候选列表 |
| `fuzzy_match_milestone` | FuzzyMatcher.MatchMilestone(name) | 同上 |
| `validate_transition` | GET available-transitions endpoint | 复用现有端点 |
| `validate_source_target` | 综合校验（源/目标非终态、同 Team、不同实体） | 新增 helper |

### 8.2 写操作（无 LLM 工具）

写操作（create / update / move）**不暴露给 LLM**。真实 DB 写路径：

1. Executor 调 `emit_form_card`（Emission 工具）→ 推送预填表单 card（`status=prefilled`, `targetEntity.bizKey=""`）→ persist 到 `copilot_messages`
2. 用户在 form card 上编辑/确认后点提交 → 前端发 `POST /messages {type: commit_card, messageId}`
3. Handler `handleCommitCard` 读 form card payload → 调 Dispatcher → 调现有 entity service（`MainItemService.Create` 等）→ 事务内 UPDATE form card `status=submitted` + persist followup text msg
4. entity service 返回 bizKey → 写入 followup 消息内容

此设计保证"AI 不直接写库"在物理层面成立（详见 [`request-model.md`](./request-model.md) §6.1 请求 3 + [`security.md`](./security.md) §7.3）。LLM 失败 / 用户放弃提交时仅留一条 `status=prefilled` 的 form card 消息，无孤儿实体。

### 8.3 Emission 工具

| 工具 | 实现行为 | SSE 事件 |
|------|---------|---------|
| `submit_rewrite` | 解析 args → emit `input_rewrite` 事件 → 返回 success（继续循环） | `input_rewrite` |
| `submit_intent` | 解析 args → Orchestrator 持久化意图消息 → emit `card_message(intent)` → 返回 terminal | `card_message(cardType=intent)` |
| `emit_form_card` | 解析 args → Orchestrator 持久化 form 消息（**targetEntity.bizKey 留空**）→ emit `card_message(form, status=prefilled)` → 返回 terminal | `card_message(cardType=form)` |
| `emit_query_result` | 解析 args → 持久化 query_result 消息 → emit → 返回 terminal | `card_message(cardType=query_result)` |
| `emit_disambig` | 解析 args → 持久化 disambig 消息 → emit → 返回 terminal | `card_message(cardType=disambig)` |

**关键差异**：
- `submit_rewrite` **不终止**流（Planner 还要继续做 thinking + submit_intent）
- 其他 emission 工具**终止**流（返回 `Status=terminal`）

**为什么 submit_rewrite 不终止**：input_rewrite 是 Planner 的第 1 步，之后还要继续 thinking 和 submit_intent。如果改成终止，Planner 流就拆成多次 LLM 调用了——浪费且不必要。

## 9. Tool 接口

```go
// internal/copilot/tools/tool_registry.go

type ToolKind string

const (
    ToolKindRead     ToolKind = "read"
    ToolKindAction   ToolKind = "action"
    ToolKindEmission ToolKind = "emission"
)

type Tool interface {
    Name() string
    Description() string
    ParametersSchema() map[string]any
    Kind() ToolKind
    Execute(ctx context.Context, args map[string]any, p ToolExecParams) (ToolResult, error)
}

// Emission 工具通过此参数拿到 outCh 写事件
type ToolExecParams struct {
    OutCh   chan<- sse.Event  // 仅 Emission 工具使用
    TurnID  string
    StepID  string
    // 持久化回调（emission 工具调此 persist 消息后拿 bizKey 用于事件）
    Persist func(msg Message) (string, error)
}

type ToolResult struct {
    Status   ToolResultStatus  // success / error / terminal
    Data     map[string]any
    Error    string
}

type ToolResultStatus string

const (
    ToolStatusSuccess  ToolResultStatus = "success"   // append tool msg，继续循环
    ToolStatusError    ToolResultStatus = "error"     // append tool msg（含错误），继续循环（让 LLM 自我修正）
    ToolStatusTerminal ToolResultStatus = "terminal"  // 不 append，终止流
)

type ToolRegistry struct {
    tools map[string]Tool
}

func (r *ToolRegistry) Register(t Tool) {
    r.tools[t.Name()] = t
}

func (r *ToolRegistry) Execute(
    ctx context.Context, name string, args map[string]any, p ToolExecParams,
) (ToolResult, error) {
    t, ok := r.tools[name]
    if !ok {
        return ToolResult{}, fmt.Errorf("unknown tool: %s", name)
    }
    return t.Execute(ctx, args, p)
}
```

**Agent 循环按 `Status` 分流**：
- `success` / `error` → append `role=tool` 到 messages，继续 LLM 循环
- `terminal` → 不 append，结束 StreamRun（emission 工具已经写完事件）

详见 [`llm-integration.md`](./llm-integration.md) §2。
