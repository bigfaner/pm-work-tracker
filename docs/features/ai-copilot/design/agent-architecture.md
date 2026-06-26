---
created: "2026-06-26"
updated: "2026-06-26"
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

## 2. Planner Agent

### 2.1 职责边界

| Planner **要做** | Planner **不做** |
|-----------------|-----------------|
| 识别所有意图（含多意图） | 调用 commit_create / commit_update |
| 必要时发意图回执 + 澄清问题 | 调用 query_entities（让 reader 做） |
| 拆解为 step 计划 | 调用 validate_transition（让 updater 做） |
| 标注每 step 用哪个 Executor | 渲染卡片（让 Executor 做） |
| input_rewrite（代词消解、省略补全） | 持久化（由 Orchestrator 负责） |

### 2.2 Planner System Prompt 结构

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

## 输入预处理（必做，在 thinking 之前）
<message type="input_rewrite">
{
  "original": "<用户原文>",
  "rewritten": "<语义完整版>",
  "context_used": [...],
  "ambiguity_notes": [...]
}
</message>

## 改写规则
1. 代词（"它"、"那个"）→ 用 Environment.pageContext 或 DraftState 解析
2. 省略主语 → 从最近 turn 推断
3. 多意图拆分："X 然后 Y" → rewritten 列出多个完整意图
4. 相对时间锚定："下周五" → 用 Environment.currentTime 计算依据（不展开绝对日期）
5. 简称保留："pm" 等歧义简称不强行展开，列入 ambiguity_notes

## 输出格式
按顺序输出：
<message type="thinking">...</message>
<message type="intent">                       // 意图消息（一条，含文本 + 结构化字段）
{
  "text": "好的，我帮你创建一个 P1 主事项「认证模块」。",
  "intents": [
    {
      "id": "intent_1",
      "label": "创建 MainItem",
      "opType": "create",
      "entityType": "main_item",
      "fields": [
        {"name": "title", "value": "认证模块", "required": true},
        {"name": "priority", "value": "P1", "required": true}
      ],
      "executor": "writer"
    }
  ],
  "missingInfo": [],                          // 如有缺失：[{"intentIndex":0,"field":"title","question":"标题？"}]
  "state": "awaiting_confirm"                 // 或 "info_complete"（澄清收齐后）
}
</message>

## 重要规则
- **意图消息是 source of truth**——持久化到 messages 表，用户确认后从中重建 plan
- 改写版本仅用于本次调用，不会持久化到对话历史
- 用户对话历史永远保留 original
- 一次指令含多意图时，全部识别并列在 intents[]
- 任意意图必填缺失时，在 missingInfo 列出所有缺失字段（一轮问全）
- 用户回答后重新进入 Planner（注入历史 + 回答），推送新意图消息
- **不直接执行意图**——推送后等用户点"✓ 理解正确"再执行
```

### 2.3 Planner 工具集（只读 + 决策类）

```go
plannerTools := []ToolDef{
    {Name: "query_team_schema", Description: "查当前 Team 的实体 schema", Params: {"entity_type": "string"}},
    {Name: "query_team_members", Description: "查 Team 成员列表", Params: {"filter": "string"}},
    {Name: "query_team_milestones", Description: "查 Team 里程碑列表", Params: {"filter": "string"}},
    {Name: "fuzzy_match_member", Description: "姓名/工号模糊匹配成员", Params: {"name": "string"}},
    {Name: "fuzzy_match_milestone", Description: "名称模糊匹配里程碑", Params: {"name": "string"}},
}
```

→ Planner **没有写工具**（commit_create 等），物理隔离职责。

## 3. Executor Agents（按 op_type 抽象，4 个）

| Executor | 处理范围 | System Prompt 核心 | 工具集 |
|----------|---------|------------------|--------|
| **writer** | 6 实体的 create | "你是创建专家。Planner 已准备完整字段，你按 schema 抽取、调 commit_create、推送 form card。" | `commit_create`, `fuzzy_match_*`, `query_team_schema` |
| **reader** | 6 实体的 query | "你是查询专家。按 filter 调 query_entities、整理结果、推送 query_result card。" | `query_entities`, `query_team_schema` |
| **updater** | 6 实体的 update / assignee / 状态变更 | "你是更新专家。先 validate_transition，通过后 commit_update。" | `commit_update`, `validate_transition`, `fuzzy_match_*`, `query_team_schema` |
| **mover** | SubItem.move | "你是 SubItem.move 专家。校验源/目标、保留 item_status 和 assignee。" | `move_sub_item`, `validate_source_target`, `query_entities` |

**为什么不按实体拆 6 个**：
- 复用度高（字段抽取、模糊匹配、必填校验逻辑共享）
- 维护成本低（4 套 system prompt vs 6 套）
- system prompt 内按 entity_type 动态注入 schema（按需，token 不浪费）

## 4. Writer Executor System Prompt 示例

```
你是 PM Work Tracker 的创建专家 Executor。

## 当前任务
- op_type: {{op_type}}
- entity_type: {{entity_type}}
- 完整字段：{{fields}}

## 实体 Schema（按需注入）
{{entity_schema}}

## 你的工具
- commit_create(entity_type, fields) → 调后端 API 创建，返回 bizKey
- fuzzy_match_member(name) → 姓名转 user_bizkey
- fuzzy_match_milestone(name) → 名称转 milestone_bizkey
- query_team_schema(entity_type) → 查 schema 补充知识

## 输出格式
<message type="thinking">...</message>
<message type="tool_call">...</message>  // 多次
<message type="card">{form_card_payload}</message>

## 重要规则
1. 字段中如有姓名引用，先调 fuzzy_match_* 转换为 bizKey
2. 卡片 derived 字段标 derived=true
3. 创建操作不调 validate_transition（无源状态）
4. 卡片状态默认 prefilled，等待用户在 UI 上提交
```

## 5. Agent 接口定义

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
}
```

## 6. Agent Registry

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

## 7. Tool 实现

### Planner 工具（只读）

| 工具 | 实现 | 说明 |
|------|------|------|
| `query_team_schema` | schema_loader.Load(entity_type) | 从 entity-schemas 常量加载 |
| `query_team_members` | TeamMemberRepo.List(team_id, filter) | 复用现有 repository |
| `query_team_milestones` | MilestoneRepo.ListByTeam(team_id, filter) | 复用现有 repository |
| `fuzzy_match_member` | FuzzyMatcher.MatchMember(name) | 简单字符串匹配 + 候选列表 |
| `fuzzy_match_milestone` | FuzzyMatcher.MatchMilestone(name) | 同上 |

### Writer 工具（写）

| 工具 | 实现 | 说明 |
|------|------|------|
| `commit_create` | Dispatcher.DispatchCreate(entity_type, fields) | 调用现有 entity service |

### Reader 工具（读）

| 工具 | 实现 | 说明 |
|------|------|------|
| `query_entities` | Dispatcher.DispatchQuery(entity_type, filter) | 调用现有 list API |

### Updater 工具（更新 + 预校验）

| 工具 | 实现 | 说明 |
|------|------|------|
| `commit_update` | Dispatcher.DispatchUpdate(bizKey, fields) | 调用现有 update API |
| `validate_transition` | GET available-transitions endpoint | 复用现有端点（MainItem/SubItem/MilestoneMap/Milestone） |

### Mover 工具

| 工具 | 实现 | 说明 |
|------|------|------|
| `move_sub_item` | SubItemService.Move | 复用现有端点 |
| `validate_source_target` | 综合校验（源/目标非终态、同 Team、不同实体） | 新增 helper |

## 8. Tool 接口

```go
// internal/copilot/tools/tool_registry.go

type Tool interface {
    Name() string
    Description() string
    ParametersSchema() map[string]any
    Execute(ctx context.Context, args map[string]any) (ToolResult, error)
}

type ToolResult struct {
    Status   string  // success / error
    Data     map[string]any
    Error    string
}

type ToolRegistry struct {
    tools map[string]Tool
}

func (r *ToolRegistry) Register(t Tool) {
    r.tools[t.Name()] = t
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, args map[string]any) (ToolResult, error) {
    t, ok := r.tools[name]
    if !ok {
        return ToolResult{}, fmt.Errorf("unknown tool: %s", name)
    }
    return t.Execute(ctx, args)
}
```
