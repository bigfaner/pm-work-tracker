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

**结构化输出强制 function calling**：
- 所有结构化输出（intent / form card / query result / disambig / input_rewrite）**只能**通过工具调用产生
- 文本通道仅用于 thinking（推理过程），禁止输出 `<message type="X">` 标签或裸 JSON
- 这样 LLM 输出强类型、可程序调度、无需文本解析

## 2. 工具三分类（Read / Action / Emission）

每个 Agent 持有的工具分三类，决定 Agent 循环如何处理：

| 类别 | 行为 | 流后处理 | 示例 |
|------|------|---------|------|
| **Read** | 纯查询，返回数据给 LLM | append `role=tool` 到 messages，继续循环 | `query_team_schema`, `fuzzy_match_member` |
| **Action** | 变更状态，返回 bizKey / 操作结果 | append `role=tool` 到 messages，继续循环 | `commit_create`, `commit_update`, `move_sub_item` |
| **Emission** | 结构化输出，写 SSE 事件到 outCh | **不 append，终止流**（StreamRun 结束） | `submit_rewrite`, `submit_intent`, `emit_form_card` |

### 2.1 Emission 工具语义

- Emission 工具是 Agent 流的**终止信号**——一旦调用，Agent 不再继续 LLM 循环
- 一个 Agent 必有且仅有"终止 emission"作为完成方式（除非超过 max_iterations 兜底失败）
- Emission 工具直接把 args 渲染为对应 SSE 事件：
  - `submit_rewrite` → `input_rewrite` 事件（仅 Planner，作为第 1 步）
  - `submit_intent` → `card_message(cardType=intent)` 事件（仅 Planner，终止）
  - `emit_form_card` → `card_message(cardType=form)` 事件（Writer/Updater/Mover，终止）
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
| `commit_create` | Action | | ✓ | | | |
| `commit_update` | Action | | | | ✓ | |
| `move_sub_item` | Action | | | | | ✓ |
| `submit_rewrite` | Emission | ✓ | | | | |
| `submit_intent` | Emission | ✓ | | | | |
| `emit_form_card` | Emission | | ✓ | | ✓ | ✓ |
| `emit_query_result` | Emission | | | ✓ | | |
| `emit_disambig` | Emission | | ✓ | | ✓ | ✓ |

## 3. Planner Agent

### 3.1 职责边界

| Planner **要做** | Planner **不做** |
|-----------------|-----------------|
| 识别所有意图（含多意图） | 调用 commit_create / commit_update / move_sub_item |
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

## 重要规则

- **结构化输出只走工具**——submit_rewrite 和 submit_intent 是必调工具
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
    {Name: "query_team_schema",     Description: "查当前 Team 的实体 schema", Params: {"entity_type": "string"}},
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

## 4. Executor Agents（按 op_type 抽象，4 个）

| Executor | 处理范围 | Read 工具 | Action 工具 | Emission 工具（终止） |
|----------|---------|----------|------------|----------------------|
| **writer** | 6 实体的 create | `query_team_schema`, `fuzzy_match_member`, `fuzzy_match_milestone` | `commit_create` | `emit_form_card`, `emit_disambig` |
| **reader** | 6 实体的 query | `query_team_schema` | `query_entities` | `emit_query_result` |
| **updater** | 6 实体的 update / assignee / 状态变更 | `query_team_schema`, `fuzzy_match_*` | `commit_update`, `validate_transition` | `emit_form_card`, `emit_disambig` |
| **mover** | SubItem.move | `query_entities` | `move_sub_item`, `validate_source_target` | `emit_form_card`, `emit_disambig` |

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

### Action
- commit_create(entity_type, fields) → 调后端 API 创建，返回 bizKey

### Emission（终止流）
- emit_form_card(cardData) → 推送表单卡片（含已创建实体的 bizKey）
- emit_disambig(candidates[]) → 多候选时让用户选

## 流程

1. 字段中如有姓名/里程碑引用，先调 fuzzy_match_* 转换为 bizKey
2. 若 fuzzy_match_* 返回多候选 → 调 emit_disambig，终止流（等用户 select_candidate）
3. bizKey 齐备后调 commit_create 创建实体，拿到新实体的 bizKey
4. 调 emit_form_card 推送表单（cardData 包含 bizKey、字段、操作类型），流结束

## 重要规则

- 卡片 derived 字段标 derived=true
- 创建操作不调 validate_transition（无源状态）
- 卡片状态默认 prefilled，等待用户在 UI 上提交
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
}
```

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

### 8.2 Action 工具

| 工具 | 实现 | 说明 |
|------|------|------|
| `commit_create` | Dispatcher.DispatchCreate(entity_type, fields) | 调用现有 entity service |
| `commit_update` | Dispatcher.DispatchUpdate(bizKey, fields) | 调用现有 update API |
| `move_sub_item` | SubItemService.Move | 复用现有端点 |

### 8.3 Emission 工具

| 工具 | 实现行为 | SSE 事件 |
|------|---------|---------|
| `submit_rewrite` | 解析 args → emit `input_rewrite` 事件 → 返回 success（继续循环） | `input_rewrite` |
| `submit_intent` | 解析 args → Orchestrator 持久化意图消息 → emit `card_message(intent)` → 返回 terminal | `card_message(cardType=intent)` |
| `emit_form_card` | 解析 args → Orchestrator 持久化 form 消息 → emit `card_message(form)` → 返回 terminal | `card_message(cardType=form)` |
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
