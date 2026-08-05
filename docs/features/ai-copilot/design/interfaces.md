---
created: "2026-06-26"
updated: "2026-06-26"
parent: tech-design.md
---

# Interfaces

> 返回 [`tech-design.md`](./tech-design.md)

## 0. 与 AB-001 的偏离说明（bizKey 类型）

项目规范 AB-001 要求"Service boundary uses int64 BizKey"。本 Copilot 模块**有意偏离**：所有 bizKey 在 Copilot 边界（Service / Repository / Schema / SSE payload）均为 **`string` (VARCHAR(36))**。

**理由**：

1. **来源不同**：现有模块的 bizKey 由 snowflake int64 生成；Copilot 模块的 bizKey 由 `copilot_messages`、`copilot_turns`、`copilot_sessions` 的独立 snowflake string 生成（与现有实体 bizKey 空间隔离，避免冲突）。
2. **跨实体引用**：Copilot 引用的目标实体 bizKey 来自 6 个不同实体表（MainItem / SubItem / Milestone / ...），这些实体的 bizKey 在 Service 层虽为 int64，但在 SSE payload / 卡片 JSON 中需要以字符串形式呈现（前端无 int64 类型）。统一在 Copilot 边界用 string，避免重复转换。
3. **既有边界不破坏**：Copilot 在**调用现有 entity service 时**遵守 AB-001——Dispatcher 内部把 Copilot 的 `targetBizKey string` 经 `ParseBizKeyParam` 转 `int64`，再传给 `MainItemService.FindByBizKey(ctx, bizKey int64)` 等。Copilot Repository / Handler 不直接调现有 entity service。

**实现契约**：

| 层 | bizKey 类型 | 说明 |
|----|-----------|------|
| Copilot HTTP Handler / Request URL | `string` (URL param) | `c.Param("id")`，不调 `ParseBizKeyParam` |
| Copilot Service interface | `string` | `MessageRepository.GetByBizKey(ctx, bizKey string)` |
| Copilot Repository / GORM | `string` (`gorm:"type:varchar(36)"`) | 与 schema.sql 对齐 |
| Copilot SSE payload JSON | `string` | 前端 TS 一致 |
| Dispatcher → 现有 entity service 边界 | `int64` | Dispatcher 内部 `id, _ := strconv.ParseInt(targetBizKey, 10, 64); mainItemSvc.FindByBizKey(ctx, id)` |

**结论**：AB-001 在现有 entity service 边界仍 100% 遵守；Copilot 模块作为独立子树使用 string bizKey，转换发生在 Dispatcher 边界。这与 AB-001 的精神（"typed at boundary"）一致，只是类型选择不同。

---

## 1. LLMProvider 接口（可插拔核心）

```go
// internal/copilot/provider/provider.go

type Provider interface {
    // 流式聊天，返回事件 channel
    StreamChat(ctx context.Context, p ProviderParams) (<-chan ProviderEvent, error)
    
    // Token 计数（用于预算管理）
    CountTokens(text string) int
    
    // 模型信息
    Models() []ModelInfo
}

type ProviderParams struct {
    Model        string         // 如 "glm-4-plus"
    SystemPrompt string
    Messages     []ProviderMsg  // 对话历史
    Tools        []ToolDef      // Function Calling 定义
    Temperature  float64        // 0.0-1.0
    MaxTokens    int            // 输出上限
    Stream       bool           // 必为 true（Copilot 强制流式）
}

type ProviderMsg struct {
    Role      string             `json:"role"`      // system / user / assistant / tool
    Content   string             `json:"content"`
    ToolCalls []ProviderToolCall `json:"toolCalls,omitempty"`
}

type ProviderEvent struct {
    Type       ProviderEventType
    Delta      string             // 文本增量（thinking/content）
    ToolCall   *ProviderToolCall
    ToolResult *ProviderToolResult
    Usage      *ProviderUsage     // 仅 done 事件
    Error      error
}

type ProviderEventType string
const (
    ProviderEventDelta      ProviderEventType = "delta"
    ProviderEventToolCall   ProviderEventType = "tool_call"
    ProviderEventToolResult ProviderEventType = "tool_result"
    ProviderEventDone       ProviderEventType = "done"
    ProviderEventError      ProviderEventType = "error"
)

type ModelInfo struct {
    Name            string  // "glm-4-plus"
    ContextWindow   int     // 128000
    InputCostPer1K  float64
    OutputCostPer1K float64
}

type ProviderUsage struct {
    InputTokens  int
    OutputTokens int
}
```

## 2. Provider Factory（配置驱动）

```go
// internal/copilot/provider/factory.go

func NewProvider(cfg ProviderConfig) (Provider, error) {
    switch cfg.Type {
    case "glm":
        return NewGLMProvider(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout), nil
    case "deepseek":
        return NewDeepSeekProvider(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout), nil
    case "openai":
        return NewOpenAIProvider(cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Timeout), nil
    case "mock":
        return NewMockProvider(cfg.MockResponses), nil
    default:
        return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
    }
}

type ProviderConfig struct {
    Type    string        `yaml:"type"`     // glm / deepseek / openai / mock
    APIKey  string        `yaml:"api_key"`  // 从 env 注入
    BaseURL string        `yaml:"base_url"` // https://open.bigmodel.cn/api/paas/v4
    Model   string        `yaml:"model"`    // glm-4-plus
    Timeout time.Duration `yaml:"timeout"`  // 30s
}
```

**配置示例**（`backend/config.yaml`）：
```yaml
copilot:
  provider:
    type: glm
    api_key: ${GLM_API_KEY}
    base_url: https://open.bigmodel.cn/api/paas/v4
    model: glm-4-plus
    timeout: 30s
```

## 3. GLM Provider 实现要点（首版）

```go
// internal/copilot/provider/glm_provider.go

type GLMProvider struct {
    apiKey  string
    baseURL string
    model   string
    timeout time.Duration
    client  *http.Client
}

func NewGLMProvider(apiKey, baseURL, model string, timeout time.Duration) *GLMProvider {
    return &GLMProvider{
        apiKey:  apiKey,
        baseURL: baseURL,
        model:   model,
        timeout: timeout,
        client:  &http.Client{Timeout: timeout},
    }
}

func (p *GLMProvider) StreamChat(ctx context.Context, params ProviderParams) (<-chan ProviderEvent, error) {
    // 构造 GLM API 请求（chat/completions stream=true）
    reqBody := p.buildRequest(params)
    
    // HTTPS POST with stream
    resp, err := p.doStreamRequest(ctx, reqBody)
    if err != nil {
        return nil, err
    }
    
    ch := make(chan ProviderEvent, 64)
    go p.parseSSEStream(resp.Body, ch)  // 解析 GLM 的 SSE 响应转为 ProviderEvent
    return ch, nil
}

func (p *GLMProvider) CountTokens(text string) int {
    // GLM 没有公开 tokenizer，用粗略估算：1 token ≈ 1.5 中文字符 / 4 英文字符
    return estimateTokens(text)
}
```

**GLM API 关键端点**：
- `POST {baseURL}/chat/completions` with `stream: true`
- SSE 格式：`data: {...}\n\n`
- 完成标记：`data: [DONE]`

## 4. ContextBuilder 接口

```go
// internal/copilot/prompt/context_builder.go

type ContextBuilder interface {
    Build(ctx context.Context, p BuildParams) (PromptContext, error)
}

type BuildParams struct {
    SessionID  string
    UserMsg    string                 // Planner 时填
    StepParams map[string]any         // Executor 时填
    DraftState *DraftState
    Env        Environment
    History    []MessageSnapshot       // 滑动窗口裁剪前的历史（Build 内裁剪）
    Role       AgentRole              // 决定 schema 注入策略
}

type PromptContext struct {
    SystemPrompt string
    Messages     []ProviderMsg        // 含 system + history + current
    Tools        []ToolDef
    TotalTokens  int                  // 估算（写入 agent_call_logs）
    Schema       string               // 按需注入的实体 schema
    History      []MessageSnapshot    // 滑动窗口裁剪后
}

// 实现要点：
// - 按 Role 决定 system prompt 模板
// - 按 entity_type 按需加载 schema（仅 Executor）
// - 滑动窗口裁剪历史（FIFO 丢弃最早）
// - Token 计数委托 Provider
```

## 5. Context Builder 实现

```go
// internal/copilot/prompt/context_builder.go

type contextBuilder struct {
    provider      llm.Provider
    schemaLoader  SchemaLoader
    flagRepo      FeatureFlagRepository
    budget        int  // 12000
}

func (b *contextBuilder) Build(ctx context.Context, p BuildParams) (PromptContext, error) {
    // 1. 估算固定部分 token（system prompt + env + tools）
    sysPrompt := b.systemPromptFor(p.Role)
    env := SerializeEnv(p.Env)  // 定义见 state-model.md §3.4
    tools := b.toolsFor(p.Role)
    
    fixedTokens := b.provider.CountTokens(sysPrompt + env + serializeTools(tools))
    
    // 2. Schema 按需加载（仅 Executor）
    var schema string
    if p.Role != RolePlanner && p.StepParams["entity_type"] != nil {
        entityType := p.StepParams["entity_type"].(string)
        schema = b.schemaLoader.Load(entityType)
        fixedTokens += b.provider.CountTokens(schema)
    }
    
    // 3. 历史滑动窗口裁剪
    budget := b.budget - fixedTokens - 1500 /* output budget */
    cropped := b.cropHistory(p.History, budget)
    
    // 4. 组装 ProviderMsg 数组
    var msgs []ProviderMsg
    msgs = append(msgs, ProviderMsg{Role: "system", Content: sysPrompt + env + schema})
    for _, m := range cropped {
        msgs = append(msgs, ToProviderMsg(m))  // 定义见 state-model.md §3.4
    }
    msgs = append(msgs, ProviderMsg{Role: "user", Content: p.UserMsg})
    
    return PromptContext{
        SystemPrompt: sysPrompt,
        Messages:     msgs,
        Tools:        tools,
        TotalTokens:  fixedTokens + b.provider.CountTokens(p.UserMsg),
        Schema:       schema,
        History:      cropped,
    }, nil
}

// cropHistory 实现见 llm-integration.md §5.4（group-aware 裁剪 + FIFO 兜底）。
// 此处不重复实现，避免两处漂移；接口契约：按 token 预算裁剪 history，
// 优先丢最旧完整 intent 组、保护当前 turn 最新 1-2 组。
```

## 6. Message Repository 接口

```go
// internal/copilot/repository/interfaces.go

type MessageRepository interface {
    Append(ctx context.Context, msg Message) (*Message, error)
    GetByBizKey(ctx context.Context, bizKey string) (*Message, error)
    ListBySession(ctx context.Context, sessionID string, limit, offset int) ([]Message, error)
    ListByTurn(ctx context.Context, turnID string, order OrderBy) ([]Message, error)
    UpdateStatus(ctx context.Context, bizKey string, status MsgStatus) error  // 统一状态更新
    UpdateCardField(ctx context.Context, bizKey, fieldName string, value any) error
    UpdateCardExpanded(ctx context.Context, bizKey, recordBizKey string, expanded bool) error
    NextSeq(ctx context.Context, sessionID string) (int, error)
}

type SessionRepository interface {
    Create(ctx context.Context, s Session) (*Session, error)
    Get(ctx context.Context, bizKey string) (*Session, error)
    ListByUser(ctx context.Context, userID uint, limit, offset int) ([]Session, error)
    UpdateCurrentTurn(ctx context.Context, bizKey, turnID string) error
    UpdateTitle(ctx context.Context, bizKey, title string) error
    UpdateStatus(ctx context.Context, bizKey string, status SessionStatus) error
    Touch(ctx context.Context, bizKey string) error  // 更新 last_active_at
}

// TurnRepository（新增）—— 3 级数据模型核心
type TurnRepository interface {
    Create(ctx context.Context, t Turn) (*Turn, error)
    Get(ctx context.Context, bizKey string) (*Turn, error)
    ListBySession(ctx context.Context, sessionID string, limit, offset int) ([]Turn, error)
    ListByUser(ctx context.Context, userBizKey string, limit, offset int) ([]Turn, error)
    UpdateStatus(ctx context.Context, bizKey string, status TurnStatus) error
    UpdateSummary(ctx context.Context, bizKey string, summary string, intentsDone int) error
    UpdateConfirmedIntent(ctx context.Context, bizKey string, intent *IntentPayload) error
    UpdateIntentMessageID(ctx context.Context, bizKey, intentMsgID string) error
}

// 注意：SummaryRepository 已移除（合并到 TurnRepository.UpdateSummary）

type AgentCallLogRepository interface {
    Append(ctx context.Context, log AgentCallLog) error
    CountTodayByUser(ctx context.Context, userBizKey string) (int, error)  // 配额检查
    MonthlyCost(ctx context.Context, year, month int) (float64, error)
}

type FeatureFlagRepository interface {
    Get(ctx context.Context, key string, scopeType, scopeID string) (bool, error)
    Set(ctx context.Context, key string, enabled bool, scopeType, scopeID string, reason string) error
}

type IdempotencyKeyRepository interface {
    // GetByRequestID 命中返回上次结果（含 ResultBizKey + Status）；未命中返回 nil
    GetByRequestID(ctx context.Context, requestID string) (*IdempotencyKey, error)
    // InsertTx 事务内插入 pending 行（见 §7.1）
    InsertTx(tx *gorm.DB, k IdempotencyKey) error
    // UpdateCommittedTx 标记 committed + 填 ResultBizKey
    UpdateCommittedTx(tx *gorm.DB, requestID, resultBizKey string) error
}
```

## 6.5 State 层接口（StateLoader / StateApplier）

State 层把上述 Repository 组装成请求级 `RequestState`（工作状态）。完整定义（`RequestState` / `SessionState` / `TurnState` / `Environment` / `DraftState` / `MessageSnapshot` / `AgentCallAccumulator` / `TurnDiff`）见 [`state-model.md`](./state-model.md)。

```go
// internal/copilot/orchestrator/state_loader.go

type StateLoader interface {
    // Load 重建完整 RequestState：读 session + turn + messages，重建 DraftState，裁剪 History
    Load(ctx context.Context, sessionID, turnID string, env Environment) (*RequestState, error)
}

// internal/copilot/orchestrator/state_applier.go

type StateApplier interface {
    // Flush —— 请求结束时单事务写回所有累加器（核心持久化路径，见 state-model.md §3.3）
    Flush(ctx context.Context, rs *RequestState) error
    // ApplyTurnUpdate 只写 diff 中非 nil 的字段（partial update；用于 commit_card 等同步路径）
    ApplyTurnUpdate(ctx context.Context, turnID string, diff TurnDiff) error
}
```

加载与更新均委托 §6 各 Repository（`SessionRepository` / `TurnRepository` / `MessageRepository`），**不新造数据访问层**。

## 7. Dispatcher 接口（复用现有 entity service，由 commit_card Handler 调用）

Dispatcher 是 LLM 与 entity service 之间的边界，**仅由 `commit_card` Handler 路径同步调用**（不经 LLM 工具，见 [`request-model.md`](./request-model.md) §6.1 请求 3 + [`security.md`](./security.md) §7.3）。Executor 的 `emit_form_card` 仅组装字段、不调 Dispatcher。

```go
// internal/copilot/service/dispatcher.go

type Dispatcher interface {
    // 按 form card payload 路由到现有 entity service，并在 Create 路径上做幂等保护（§7.1）。
    // 调用方：handleCommitCard Handler（POST /messages type=commit_card）
    Dispatch(ctx context.Context, req DispatchRequest) (*DispatchResult, error)
}

// DispatchRequest —— commit_card 的派发入参（含幂等上下文）
type DispatchRequest struct {
    FormCard  FormCard // form card payload
    RequestID string   // 客户端生成的幂等键（UUID v4），来自 MessageRequest.RequestID
    MessageID string   // form card 消息 id（关联 copilot_idempotency_keys.message_id）
    TurnID    string
    SessionID string
}

// FormCard = 持久化的 form card 消息反序列化结构
type FormCard struct {
    OpType       string       // create / update / move
    EntityType   string       // main_item / sub_item / ...
    TargetBizKey string       // update/move 必填（已存在实体）；create 留空
    Fields       []FieldState // 用户编辑后的最终值
}

type DispatchResult struct {
    BizKey  string  // 创建/更新的实体 bizKey（commit 后回填到 followup 消息）
    Title   string
    Raw     any     // 原始返回值
}

// 实现路由表（均复用现有 service，零改动）：
// - opType=create, entityType=main_item   → MainItemService.Create
// - opType=create, entityType=sub_item    → SubItemService.Create
// - opType=create, entity_type=milestone  → MilestoneService.Create
// - opType=update, entityType=main_item   → MainItemService.Update（assignee 是普通字段）
// - opType=move,   entityType=sub_item    → SubItemService.Move
// - ...
//
// Dispatcher 内部 idempotency：见 §7.1（commit_card 幂等）
```

### 7.1 commit_card 幂等性（防 LLM 重试 / 网络重试导致重复创建）

`commit_card` 请求携带客户端生成的 `requestId`（UUID v4，前端在用户每次点提交时生成一次）。Dispatcher 在 entity service Create 路径上：

1. 查 `copilot_idempotency_keys` 表（`request_id` UNIQUE）；命中则直接返回上次结果（含 bizKey），不再调 entity service。
2. 未命中 → 事务内：INSERT idempotency row + 调 entity service Create + UPDATE form card status=submitted。
3. 重复请求（同 requestId）在 step 1 即返回，保证不重复创建。

> 此机制同时覆盖 LLM 失败重试（若未来 Executor 重引入 Action 工具）与前端网络抖动重试。schema 见 [`schema.sql`](./schema.sql) `copilot_idempotency_keys` 表；Go model `IdempotencyKey` + `IdempotencyKeyRepository`（`GetByRequestID` / `InsertTx` / `UpdateCommittedTx`，见 §6 / §11）。Dispatcher 实现注入 `IdempotencyKeyRepository`，在 `Dispatch` 内完成查/插/更新；`handleCommitCard` 把 `MessageRequest.RequestID` 连同 messageId/turnID/sessionID 组装成 `DispatchRequest` 传入。
```

## 8. ToolRegistry 接口

工具两分类（Read / Emission）与具体工具清单见 [`agent-architecture.md`](./agent-architecture.md) §2。Emission 工具在 Agent 循环中的终止语义见 [`llm-integration.md`](./llm-integration.md) §2。**无 Action 类**——写操作不经 LLM 工具，由 `commit_card` Handler 路径触发（见 [`request-model.md`](./request-model.md) §6.1 + [`security.md`](./security.md) §7.3）。

```go
// internal/copilot/tools/tool_registry.go

type ToolKind string

const (
    ToolKindRead     ToolKind = "read"
    ToolKindEmission ToolKind = "emission"
)

type Tool interface {
    Name() string
    Description() string
    ParametersSchema() map[string]any
    Kind() ToolKind
    Execute(ctx context.Context, args map[string]any, p ToolExecParams) (ToolResult, error)
}

// ToolExecParams 仅 Emission 工具使用 OutCh / Persist；
// Read 工具可忽略这两个字段。
type ToolExecParams struct {
    OutCh   chan<- sse.Event  // Emission 工具直接写专用事件
    TurnID  string
    StepID  string
    Persist func(msg Message) (string, error)  // 持久化消息并返回 bizKey
}

type ToolResultStatus string

const (
    ToolStatusSuccess  ToolResultStatus = "success"  // append tool msg，继续循环
    ToolStatusError    ToolResultStatus = "error"    // append tool msg（含错误），继续循环（让 LLM 修正）
    ToolStatusTerminal ToolResultStatus = "terminal" // 不 append，终止 StreamRun
)

type ToolResult struct {
    Status ToolResultStatus
    Data   map[string]any
    Error  string
}

type ToolRegistry struct {
    tools map[string]Tool
}

func (r *ToolRegistry) Register(t Tool) {
    r.tools[t.Name()] = t
}

func (r *ToolRegistry) Lookup(name string) Tool {
    return r.tools[name]
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

## 9. Orchestrator 接口

```go
// internal/copilot/orchestrator/orchestrator.go

type Orchestrator struct {
    planner    agent.Agent
    registry   *agent.Registry
    ctxBuilder prompt.ContextBuilder
    turnRepo   repository.TurnRepository   // 新增
    msgRepo    repository.MessageRepository
    dispatcher service.Dispatcher
    provider   llm.Provider
}

// HandleUserMessage 处理自由文本（调 Planner + 推送意图消息，不执行）
func (o *Orchestrator) HandleUserMessage(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error

// HandleAnswerClarify 处理 clarify 回答（重新调 Planner，推送新意图消息）
func (o *Orchestrator) HandleAnswerClarify(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error

// ExecuteFromIntent 从已确认意图重建 plan + 执行（用户点"理解正确"触发）
func (o *Orchestrator) ExecuteFromIntent(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error

// ExecuteSelectCandidate 处理选歧义候选（注入 bizKey + 调对应 Executor）
func (o *Orchestrator) ExecuteSelectCandidate(
    ctx context.Context, turnCtx *TurnContext, candidateBizKey string,
    intent IntentSpec, eventCh chan<- sse.Event,
) error

// HandleCancel 处理取消（PATCH 意图消息 state=cancelled）
func (o *Orchestrator) HandleCancel(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error
```

## 10. Handler 接口（Gin 端点）

```go
// internal/copilot/handler/copilot_handler.go

type CopilotHandler struct {
    sessionRepo  repository.SessionRepository
    turnRepo     repository.TurnRepository
    msgRepo      repository.MessageRepository
    stateLoader  orchestrator.StateLoader    // Load 工作状态（见 state-model.md §3.1）
    stateApplier orchestrator.StateApplier   // Flush / ApplyTurnUpdate（见 state-model.md §3.3）
    orchestrator *orchestrator.Orchestrator
    dispatcher   service.Dispatcher
    flagCache    *service.FeatureFlagCache
    quotaSvc     *service.QuotaService
}

// 创建会话
func (h *CopilotHandler) CreateSession(c *gin.Context)

// 列出用户会话
func (h *CopilotHandler) ListSessions(c *gin.Context)

// 获取会话详情
func (h *CopilotHandler) GetSession(c *gin.Context)

// 删除会话（软删除）
func (h *CopilotHandler) DeleteSession(c *gin.Context)

// 单一消息端点（按请求体 type 字段分派，SSE 流或 JSON 响应）
// type: free_text / answer_clarify / confirm_intent / adjust_intent /
//       select_candidate / commit_card / cancel
func (h *CopilotHandler) PostMessage(c *gin.Context)

// 列出会话历史消息（分页 JSON）
func (h *CopilotHandler) ListMessages(c *gin.Context)

// 更新卡片状态/字段（PATCH JSON，独立端点）
// 用于：字段编辑、展开/折叠、应用/丢弃 diff
func (h *CopilotHandler) PatchMessage(c *gin.Context)

// AI 健康检查
func (h *CopilotHandler) Health(c *gin.Context)
```

### 单一端点请求类型

```go
// internal/copilot/handler/copilot_handler.go

type MessageRequest struct {
    Type            string `json:"type"`            // 8 种类型之一
    Content         string `json:"content,omitempty"`         // free_text / adjust_intent
    Answer          string `json:"answer,omitempty"`          // answer_clarify
    IntentMessageID string `json:"intentMessageId,omitempty"` // confirm_intent / adjust_intent / answer_clarify
    MessageID       string `json:"messageId,omitempty"`       // select_intent / select_candidate / commit_card / cancel
    CandidateBizKey string `json:"candidateBizKey,omitempty"` // select_candidate（实体候选 bizKey）
    CandidateIntentID string `json:"candidateIntentId,omitempty"` // select_intent（候选意图 ID，见 state-machines.md §3 规则 3b）
    NewContent      string `json:"newContent,omitempty"`      // adjust_intent
    RequestID       string `json:"requestId,omitempty"`       // commit_card only：客户端生成的幂等键（UUID v4）
    PageContext     *PageContext `json:"pageContext,omitempty"` // free_text/answer_clarify/adjust_intent：前端携当前页面上下文（见 state-model.md §4）
}
```

## 11. 内嵌 Struct 完整定义

下列 struct 在 tech-design.md §3、request-model.md §3、sse-protocol.md §3 中被引用，统一在此定义。

```go
// internal/copilot/model/types.go

// FieldState —— 单个字段的状态与值（IntentSpec.Fields / FormCardData.Fields 共用）
type FieldState struct {
    Name     string `json:"name"`               // schema 字段名（如 "title" / "assignee" / "priority"）
    Value    any    `json:"value"`               // 字段值；类型由 entity-schemas 定义（string / int / []string / *time.Time）
    Required bool   `json:"required"`            // 是否必填（来自 entity-schemas）
    Derived  bool   `json:"derived,omitempty"`   // AI 推断的字段（vs 用户原文提取）—— UI 标注
    Version  int64  `json:"version,omitempty"`   // 并发控制：每次 PATCH 递增（见 api-handbook.md §4）
    EditedAt int64  `json:"editedAt,omitempty"`  // 最后编辑时间戳（ms）—— last-write-wins 判定
}

// EntityRef —— 实体引用（IntentSpec.TargetEntity / FormCardData.TargetEntity / DisambigCardData.Candidates）
type EntityRef struct {
    BizKey   string `json:"bizKey"`              // 实体 bizKey（form card 预填阶段为 ""，commit 后回填）
    EntityType string `json:"entityType"`        // main_item / sub_item / milestone / ...
    Title    string `json:"title,omitempty"`     // 显示用标题
    BizCode  string `json:"bizCode,omitempty"`   // 业务编码（如 MI-0023）
}

// EntityRecord —— 查询结果单行（QueryResultCardData.Records）
type EntityRecord struct {
    BizKey    string         `json:"bizKey"`
    EntityType string        `json:"entityType"`
    Title     string         `json:"title"`
    BizCode   string         `json:"bizCode,omitempty"`
    Status    string         `json:"status,omitempty"`     // 实体当前状态（completed / in_progress / ...）
    Fields    map[string]any `json:"fields,omitempty"`     // 完整字段（按 entity-schemas）
    Expanded  bool           `json:"expanded"`             // UI 是否默认展开详情（reader 单记录=true，多记录=false）
}

// TracePayload —— type=trace 消息的 trace 字段结构
type TracePayload struct {
    Thinking string         `json:"thinking"`             // LLM thinking 文本聚合
    Actions  []TraceAction  `json:"actions"`              // tool_call + tool_result 配对序列
    StartedAt int64         `json:"startedAt"`
    EndedAt   int64         `json:"endedAt,omitempty"`
    Status   string         `json:"status"`               // streaming / done / failed
}

type TraceAction struct {
    CallID     string         `json:"callId"`
    ToolName   string         `json:"toolName"`
    Arguments  map[string]any `json:"arguments"`
    Result     map[string]any `json:"result,omitempty"`
    Error      string         `json:"error,omitempty"`
    Status     string         `json:"status"`              // success / error
    StartedAt  int64          `json:"startedAt"`
    DurationMs int            `json:"durationMs"`
}

// IntentMeta —— SSE 事件的 intent 上下文（Event.IntentMeta 字段，step 级事件携带）
type IntentMeta struct {
    ID    string `json:"id"`              // intent_1, intent_2
    Label string `json:"label"`           // "创建 MainItem"
    Seq   int    `json:"seq"`             // 在 turn 内的顺序（1-based）
    Total int    `json:"total"`           // turn 内 intent 总数
}

// DiffOverlay —— 并发编辑冲突时的 diff 预览（FormCardData.DiffOverlay）
type DiffOverlay struct {
    FieldName      string `json:"fieldName"`          // 冲突字段名
    YourValue      any    `json:"yourValue"`           // 用户本地编辑的值
    OtherValue     any    `json:"otherValue"`          // 对话路径写入的值
    OtherSource    string `json:"otherSource"`         // "dialog_adjust" / "intent_adjust"
    OtherEditedAt  int64  `json:"otherEditedAt"`       // 对方写入时间戳（ms）
    DiffPreview    string `json:"diffPreview"`         // 人类可读 diff（如 "-认证模块 +认证模块 v2"）
}

// FormErrors —— form card 校验错误（FormCardData.Errors）
type FormErrors struct {
    Validation        map[string]string `json:"validation,omitempty"`        // 字段级错误（fieldName → message）
    ValidTransitions  []string          `json:"validTransitions,omitempty"`  // 状态变更预校验：合法目标列表
    Permission        string            `json:"permission,omitempty"`        // RBAC 拒绝原因
}

// IntentPayload —— 意图消息的完整 payload（与 request-model.md §3.1 一致）
type IntentPayload struct {
    Text       string        `json:"text"`
    Intents    []IntentSpec  `json:"intents"`
    MissingInfo []MissingItem `json:"missingInfo,omitempty"`
    State      string        `json:"state"`              // awaiting_confirm / info_complete / confirmed / adjusted / cancelled
    Decision   string        `json:"decision,omitempty"` // Planner 路由决策：confirm / show_candidates / cannot_understand（见 agent-architecture.md §3.4）
}

// IdempotencyKey —— commit_card 幂等行（schema.sql copilot_idempotency_keys；见 §7.1）
type IdempotencyKey struct {
    RequestID    string     `gorm:"type:varchar(36);uniqueIndex;not null"`
    MessageID    string     `gorm:"type:varchar(36);index;not null"` // 关联的 form card 消息
    TurnID       string     `gorm:"type:varchar(36);not null"`
    SessionID    string     `gorm:"type:varchar(36);not null"`
    UserBizKey   string     `gorm:"type:varchar(36);not null"`
    ResultBizKey string     `gorm:"type:varchar(36)"`                  // entity service 返回的实体 bizKey（committed 后填）
    Status       string     `gorm:"type:varchar(16);not null"`         // pending / committed / failed
    CreatedAt    time.Time
    CommittedAt  *time.Time
}
```

> **State 相关类型**权威定义见 [`state-model.md`](./state-model.md)：`Environment` / `PageContext`（§2.1）、`DraftState`（§2.2）、`MessageSnapshot`（§2.3）、`AgentCallAccumulator`（§2.4）、`RequestState` / `SessionState` / `TurnState` / `TurnDiff`（§3）。本节只定义卡片/字段相关 struct；state 层 struct 统一在 state-model.md，避免重复定义。
