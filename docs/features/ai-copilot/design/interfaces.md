---
created: "2026-06-26"
updated: "2026-06-26"
parent: tech-design.md
---

# Interfaces

> 返回 [`tech-design.md`](./tech-design.md)

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
    env := serializeEnv(p.Env)
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
        msgs = append(msgs, toProviderMsg(m))
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

func (b *contextBuilder) cropHistory(history []MessageSnapshot, budget int) []MessageSnapshot {
    cropped := history
    for {
        tokens := b.provider.CountTokens(serialize(cropped))
        if tokens <= budget || len(cropped) == 0 {
            break
        }
        cropped = cropped[1:]  // FIFO 丢弃最早
    }
    return cropped
}
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
```

## 7. Dispatcher 接口（复用现有 entity service）

```go
// internal/copilot/service/dispatcher.go

type Dispatcher interface {
    // 按 form card payload 路由到现有 entity service
    Dispatch(ctx context.Context, fc FormCard) (*DispatchResult, error)
}

type DispatchResult struct {
    BizKey  string  // 创建/更新的实体 bizKey
    Title   string
    Raw     any     // 原始返回值
}

// 实现路由表：
// - opType=create, entityType=main_item → MainItemService.Create
// - opType=create, entityType=sub_item → SubItemService.Create
// - opType=update, entityType=main_item → MainItemService.Update
// - opType=update, entityType=main_item, fields={assignee} → 同上（assignee 是字段）
// - opType=move, entityType=sub_item → SubItemService.Move
// - ...
```

## 8. ToolRegistry 接口

```go
// internal/copilot/tools/tool_registry.go

type Tool interface {
    Name() string
    Description() string
    ParametersSchema() map[string]any
    Execute(ctx context.Context, args map[string]any) (ToolResult, error)
}

type ToolResult struct {
    Status string         // success / error
    Data   map[string]any
    Error  string
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
    turnRepo     repository.TurnRepository   // 新增
    msgRepo      repository.MessageRepository
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
    Type            string `json:"type"`            // 7 种类型之一
    Content         string `json:"content,omitempty"`         // free_text / adjust_intent
    Answer          string `json:"answer,omitempty"`          // answer_clarify
    IntentMessageID string `json:"intentMessageId,omitempty"` // confirm_intent / adjust_intent / answer_clarify
    MessageID       string `json:"messageId,omitempty"`       // select_candidate / commit_card / cancel
    CandidateBizKey string `json:"candidateBizKey,omitempty"` // select_candidate
    NewContent      string `json:"newContent,omitempty"`      // adjust_intent
}
```
