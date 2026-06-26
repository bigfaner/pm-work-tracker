---
created: "2026-06-26"
updated: "2026-06-26"
parent: tech-design.md
---

# Request Model: 单一端点 + 意图消息驱动

> 返回 [`tech-design.md`](./tech-design.md)

## 1. 核心原则

### 1.1 三个核心设计

1. **单一端点**：所有用户输入通过 `POST /sessions/:id/messages`，按请求体 `type` 字段分派
2. **意图作为 message 持久化**：意图 ≠ plan，意图是 LLM 对用户想法的推测（文本 + 结构化确认体），作为 `type=intent` 消息持久化
3. **Plan 从意图消息重建**：用户确认意图后，从持久化的意图消息提取结构化部分重建 plan，无需前端缓存

### 1.2 意图 vs 计划

| 概念 | 含义 | 存储方式 |
|------|------|---------|
| **Intent（意图）** | LLM 对用户想法的推测，含文本回执 + 结构化字段 | 作为 `type=intent` 消息持久化到 messages 表 |
| **Plan（计划）** | 从已确认意图重建的执行步骤 | **不入库**，goroutine 内临时构造 |

**关键洞察**：意图消息是 source of truth。Plan 是从意图消息的结构化部分派生的临时调度状态。

### 1.3 单一端点设计

```
POST /api/v1/copilot/sessions/:id/messages
```

请求体含 `type` 字段决定处理逻辑：

```typescript
type RequestBody =
  | { type: "free_text"; content: string }                              // 自由文本指令
  | { type: "answer_clarify"; answer: string; intentMessageId: string } // 回答澄清
  | { type: "confirm_intent"; intentMessageId: string }                 // 确认意图
  | { type: "adjust_intent"; intentMessageId: string; newContent: string } // 调整意图
  | { type: "select_candidate"; messageId: string; candidateBizKey: string } // 选候选
  | { type: "commit_card"; messageId: string }                          // 提交表单
  | { type: "cancel"; messageId?: string }                              // 取消
```

响应：
- 自由文本 / answer_clarify / confirm_intent / select_candidate → **SSE 流**（无前缀 JSON 内容）
- commit_card → **JSON**（不走 LLM，直接 entity service）
- cancel → **JSON**（标记状态）

### 1.4 PATCH 端点（保留 RESTful）

卡片字段编辑、展开/折叠等就地更新通过独立端点：

```
PATCH /api/v1/copilot/messages/:id
```

不属于"用户向对话添加内容"，所以不合并到 POST /messages。

## 2. TurnContext（与 goroutine 同生命周期）

### 2.1 结构定义

```go
// internal/copilot/orchestrator/turn_context.go

type TurnContext struct {
    TurnID      string
    SessionID   string

    // Turn 状态（从 copilot_turns 表读取，加速重建）
    Status          TurnStatus         // planning / awaiting_confirm_intent / executing / ...
    ConfirmedIntent *IntentPayload     // 从 turn.confirmed_intent 读取
    IntentMessageID *string            // 关联的意图消息

    // 用户输入（自由文本场景）
    UserMessage string

    // 上下文
    Environment Environment         // user / team / page / currentTime
    History     []MessageSnapshot   // 从 messages 表读取的历史
    DraftState  *DraftState         // clarify 累积的草稿

    // Agent 调用历史（goroutine 内累积）
    AgentCalls []AgentCallSnapshot

    // 同步
    mu sync.Mutex

    StartedAt time.Time
}
```

**关键设计点**：
- **Status 字段**：从 turn 表读取，Orchestrator 据此决定路由分支
- **ConfirmedIntent**：从 turn.confirmed_intent 读取（用户确认后冗余存储）
- **不再有 `Plan` 字段**——plan 是从 `ConfirmedIntent` 派生的临时状态

### 2.2 Handler 创建 TurnContext

```go
// internal/copilot/handler/copilot_handler.go

func (h *CopilotHandler) PostMessage(c *gin.Context) {
    var req MessageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        apperrors.RespondError(c, apperrors.ErrValidation)
        return
    }
    
    // 按类型分派
    switch req.Type {
    case "free_text":
        h.handleFreeText(c, req)
    case "answer_clarify":
        h.handleAnswerClarify(c, req)
    case "confirm_intent":
        h.handleConfirmIntent(c, req)
    case "adjust_intent":
        h.handleAdjustIntent(c, req)
    case "select_candidate":
        h.handleSelectCandidate(c, req)
    case "commit_card":
        h.handleCommitCard(c, req)  // 返回 JSON
    case "cancel":
        h.handleCancel(c, req)       // 返回 JSON
    default:
        apperrors.RespondError(c, apperrors.ErrValidation)
    }
}

func (h *CopilotHandler) handleFreeText(c *gin.Context, req MessageRequest) {
    turnID := snowflake.Next()
    sessionID := c.Param("id")
    userBizKey := middleware.GetUserBizKey(c)

    // 事务：创建 Turn + persist user msg
    err := h.txManager.WithinTx(c, func(tx *gorm.DB) error {
        // 1. 创建 Turn 行
        h.turnRepo.Create(c, Turn{
            BizKey:       turnID,
            SessionID:    sessionID,
            UserBizKey:   userBizKey,
            Status:       TurnStatusPlanning,
            StartedAt:    time.Now(),
            LastActiveAt: time.Now(),
        })

        // 2. persist user msg（user_message 不在 turns 表，统一在 messages）
        h.msgRepo.Append(c, Message{
            Role: RoleUser, Type: TypeText, Status: MsgStatusSent,
            TurnID: turnID, SessionID: sessionID,
            Content: req.Content,
            Seq: h.msgRepo.NextSeq(c, sessionID),
        })

        return nil
    })

    // 3. 构建 TurnContext
    turnCtx := &TurnContext{
        TurnID:      turnID,
        SessionID:   sessionID,
        Status:      TurnStatusPlanning,
        UserMessage: req.Content,
        Environment: h.buildEnv(c),
        History:     h.msgRepo.ListBySession(c, sessionID, 50, 0),
        StartedAt:   time.Now(),
    }

    // 4. 启动 SSE 流 + goroutine
    h.startSSEStream(c)
    eventCh := make(chan sse.Event, 64)
    go func() {
        defer close(eventCh)
        h.orchestrator.HandleUserMessage(c.Request.Context(), turnCtx, eventCh)
    }()
    h.streamEvents(c, eventCh)
}

func (h *CopilotHandler) handleConfirmIntent(c *gin.Context, req MessageRequest) {
    // 1. 读取意图消息
    intentMsg, _ := h.msgRepo.GetByBizKey(c, req.IntentMessageID)
    var intentPayload IntentPayload
    json.Unmarshal(intentMsg.Card, &intentPayload)

    // 2. 从 turn 表读取 turn 状态
    turn, _ := h.turnRepo.Get(c, intentMsg.TurnID)

    // 3. 事务：UPDATE turn + PATCH 意图消息
    h.txManager.WithinTx(c, func(tx *gorm.DB) error {
        // 更新 turn: status=executing + confirmed_intent 冗余存储
        h.turnRepo.UpdateStatus(c, turn.BizKey, TurnStatusExecuting)
        h.turnRepo.UpdateConfirmedIntent(c, turn.BizKey, &intentPayload)
        // PATCH 意图消息 status=confirmed
        h.msgRepo.UpdateStatus(c, intentMsg.BizKey, MsgStatusIntentConfirmed)
        return nil
    })

    // 4. 构建 TurnContext（从 turn 表读，无需扫描所有 messages）
    turnCtx := &TurnContext{
        TurnID:          turn.BizKey,
        SessionID:       turn.SessionID,
        Status:          TurnStatusExecuting,
        ConfirmedIntent: &intentPayload,
        IntentMessageID: &intentMsg.BizKey,
        Environment:     h.buildEnv(c),
        History:         h.msgRepo.ListByTurn(c, turn.BizKey, OrderBySeq),
        StartedAt:       time.Now(),
    }

    // 5. 启动 SSE 流 + 执行
    h.startSSEStream(c)
    eventCh := make(chan sse.Event, 64)
    go func() {
        defer close(eventCh)
        h.orchestrator.ExecuteFromIntent(c.Request.Context(), turnCtx, eventCh)
    }()
    h.streamEvents(c, eventCh)
}
```

## 3. 意图消息（type=intent）

### 3.1 IntentPayload 结构

```go
// internal/copilot/model/message.go

type IntentPayload struct {
    // 文本部分（自然语言意图回执）
    Text string `json:"text"`
    // 例："好的，我帮你创建一个 P1 主事项「认证模块」，分配给张三。"
    
    // 结构化部分（一个或多个意图规格）
    Intents []IntentSpec `json:"intents"`
    
    // 主动澄清（如果有缺失字段）
    MissingInfo []MissingItem `json:"missingInfo,omitempty"`
    
    // 状态：awaiting_confirm / confirmed / adjusted / cancelled
    State string `json:"state"`
}

type IntentSpec struct {
    ID          string       `json:"id"`          // intent_1, intent_2
    Label       string       `json:"label"`       // "创建 MainItem"
    OpType      string       `json:"opType"`      // create / query / update / move
    EntityType  string       `json:"entityType"`  // main_item / sub_item / ...
    Fields      []FieldState `json:"fields"`
    TargetEntity *EntityRef  `json:"targetEntity,omitempty"` // 写操作的目标实体
    Executor    string       `json:"executor"`    // writer / reader / updater / mover
}

type MissingItem struct {
    IntentIndex int    `json:"intentIndex"` // 对应 intents[int] 的索引
    Field       string `json:"field"`
    Question    string `json:"question"`
    Hint        string `json:"hint,omitempty"`
}

// 意图消息状态
const (
    IntentStateAwaitingConfirm = "awaiting_confirm"
    IntentStateConfirmed       = "confirmed"
    IntentStateAdjusted        = "adjusted"
    IntentStateCancelled       = "cancelled"
    IntentStateInfoComplete    = "info_complete"  // clarify 收齐后
)
```

### 3.2 意图消息的存储

意图消息持久化到 messages 表：

```go
Message{
    Role:          RoleAI,
    Type:          TypeIntent,           // 新增消息类型
    TurnID:        turnID,
    IntentID:      nil,                  // 意图消息本身不属于某个 intent
    Content:       payload.Text,         // 文本部分冗余存储（便于检索）
    IntentPayload: payloadMarshaled,     // 完整结构化 payload
}

// 意图消息不与具体 intent_id 关联，因为它可能含多个 intents[]
```

### 3.3 意图消息的渲染

前端收到 `card_message` 事件且 `cardType="intent"` 时，渲染为：
- 文本部分：作为普通 AI 文本消息显示
- 结构化部分：渲染为"意图确认卡片"（含字段列表 + 确认/调整/取消按钮）
- 主动澄清：如果在 missingInfo 中有内容，切文本模式等用户回答

### 3.4 意图消息的生命周期

```
Planner 推送 → state=awaiting_confirm
    │
    ├─ 用户点"✓ 理解正确" → PATCH state=confirmed → 执行 plan
    │
    ├─ 用户点"✎ 我要调整" → PATCH state=adjusted → 切文本模式重新输入
    │                                                  → 新的意图消息推送
    │
    ├─ 用户点"✗ 取消" → PATCH state=cancelled → 终止
    │
    └─ 含 missingInfo → 用户回答 → 重新调 Planner → 新的意图消息推送
                       （原意图消息保留在历史，state=awaiting_confirm 但已被新消息替代）
```

## 4. 完整请求类型与路由表

| 操作 | Request Type | 处理 | LLM 调用 | 响应 |
|------|-------------|------|---------|------|
| 自由文本 | `free_text` | 调 Planner + 推送意图消息 | Planner | SSE 流 |
| 回答澄清 | `answer_clarify` | 重新调 Planner（注入回答） | Planner | SSE 流 |
| 确认意图 | `confirm_intent` | 从意图消息重建 plan + 调 executor | Executor（按 plan） | SSE 流 |
| 调整意图 | `adjust_intent` | 切文本模式重新输入（前端处理，可能不需后端） | — | JSON |
| 选候选 | `select_candidate` | 注入 bizKey + 调 executor | Executor | SSE 流 |
| 提交表单 | `commit_card` | 直接调 entity service | ❌ | JSON |
| 取消 | `cancel` | PATCH 状态 | ❌ | JSON |

## 5. 路由实现

### 5.1 Orchestrator 接口

```go
// internal/copilot/orchestrator/orchestrator.go

type Orchestrator struct {
    planner    agent.Agent
    registry   *agent.Registry
    ctxBuilder prompt.ContextBuilder
    msgRepo    repository.MessageRepository
    dispatcher service.Dispatcher
    provider   llm.Provider
}

// HandleUserMessage 处理自由文本（Planner 流 + 推送意图消息）
func (o *Orchestrator) HandleUserMessage(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error

// HandleAnswerClarify 处理 clarify 回答（重新调 Planner）
func (o *Orchestrator) HandleAnswerClarify(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error

// ExecuteFromIntent 从已确认意图重建 plan + 执行
func (o *Orchestrator) ExecuteFromIntent(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error

// ExecuteSelectCandidate 处理选歧义候选（注入 bizKey + 执行）
func (o *Orchestrator) ExecuteSelectCandidate(
    ctx context.Context, turnCtx *TurnContext, candidateBizKey string,
    intent IntentSpec, eventCh chan<- sse.Event,
) error
```

### 5.2 ExecuteFromIntent 实现

```go
// internal/copilot/orchestrator/orchestrator.go

func (o *Orchestrator) ExecuteFromIntent(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error {
    intent := turnCtx.ConfirmedIntent
    
    // 检查是否仍有 missing_info（不应该有，但兜底）
    if len(intent.MissingInfo) > 0 {
        eventCh <- errorEvent(turnCtx, errors.New("intent has missing info"))
        return nil
    }
    
    // 按 intents 顺序执行
    for idx, spec := range intent.Intents {
        eventCh <- stepStartedEvent(turnCtx, spec, idx)
        
        executor, err := o.registry.Executor(spec.Executor)
        if err != nil {
            eventCh <- errorEvent(turnCtx, err)
            return err
        }
        
        stream, err := executor.StreamRun(ctx, agent.AgentRunParams{
            StepParams: map[string]any{
                "op_type": spec.OpType,
                "entity_type": spec.EntityType,
                "fields": spec.Fields,
                "target_entity": spec.TargetEntity,
            },
            Env: turnCtx.Environment,
            History: turnCtx.History,
        })
        if err != nil {
            eventCh <- errorEvent(turnCtx, err)
            return err
        }
        
        outcome := o.consumeExecutorStream(ctx, stream, turnCtx, spec, eventCh)
        
        if outcome.NeedsUserInteraction {
            // form card / disambig card 中断
            // 剩余 intents 丢弃
            eventCh <- stepPhaseDoneEvent(turnCtx, spec, outcome)
            eventCh <- turnPhaseDoneEvent(turnCtx, "execute", "awaiting_user_action",
                NextAction{Type: outcome.NextActionType, MessageID: outcome.CardMessageID})
            return nil
        }
        
        eventCh <- stepCompletedEvent(turnCtx, spec, outcome)
    }
    
    eventCh <- turnPhaseDoneEvent(turnCtx, "execute", "success",
        NextAction{Type: "turn_complete"})
    return nil
}
```

### 5.3 HandleUserMessage 实现（Planner 推送意图消息）

```go
func (o *Orchestrator) HandleUserMessage(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error {
    // 调 Planner
    plannerStream, err := o.planner.StreamRun(ctx, agent.AgentRunParams{
        UserMsg: turnCtx.UserMessage,
        Env: turnCtx.Environment,
        History: turnCtx.History,
        DraftState: turnCtx.DraftState,
    })
    if err != nil {
        eventCh <- errorEvent(turnCtx, err)
        return err
    }
    
    // 消费 Planner 流，聚合到 IntentPayload
    intentPayload := o.consumePlannerStream(ctx, plannerStream, turnCtx, eventCh)
    
    // persist 意图消息
    intentMsg := o.msgRepo.Append(Message{
        Role: RoleAI, Type: TypeIntent,
        TurnID: turnCtx.TurnID,
        Content: intentPayload.Text,
        IntentPayload: marshal(intentPayload),
    })
    
    // 推送 card_message 事件（cardType=intent）
    eventCh <- cardMessageEvent(turnCtx, "intent", intentPayload.State, intentPayload, intentMsg.BizKey)
    
    // 推送 turn_phase_done
    var nextAction NextAction
    if len(intentPayload.MissingInfo) > 0 {
        nextAction = NextAction{Type: "await_clarify"}
    } else {
        nextAction = NextAction{Type: "await_confirm_intent", MessageID: intentMsg.BizKey}
    }
    
    eventCh <- turnPhaseDoneEvent(turnCtx, "planner", "success", nextAction)
    return nil
}
```

## 6. 完整时序示例

### 6.1 单意图写操作（完整流程）

```
─── 请求 1（turn_001，发指令）────────────────────────────────────────
POST /sessions/sess_xxx/messages
{ "type": "free_text", "content": "创建一个 P1 事项叫认证模块" }

→ 后端：创建 TurnContext + persist user msg + 启动 SSE 流

SSE 流（Content-Type: text/event-stream，无前缀 JSON）：
{"event":"input_rewrite","turnId":"turn_001","data":{"kind":"input_rewrite","original":"创建一个 P1 事项叫认证模块","rewritten":"创建一个 MainItem，title=认证模块，priority=P1","contextUsed":[],"ambiguityNotes":[]}}

{"event":"thinking","turnId":"turn_001","data":{"kind":"thinking","content":"用户要创建 MainItem，字段已完整。"}}

{"event":"card_message","turnId":"turn_001","messageId":"msg_002","data":{"kind":"card","cardType":"intent","status":"awaiting_confirm","cardData":{"text":"好的，我帮你创建一个 P1 主事项「认证模块」。","intents":[{"id":"intent_1","label":"创建 MainItem","opType":"create","entityType":"main_item","fields":[{"name":"title","value":"认证模块","required":true},{"name":"priority","value":"P1","required":true}],"executor":"writer"}],"missingInfo":[]}}}

{"event":"turn_phase_done","turnId":"turn_001","data":{"kind":"turn_phase_done","phaseType":"planner","outcome":"success","nextAction":{"type":"await_confirm_intent","messageId":"msg_002"}}}

→ 流关闭
→ 前端：展示意图消息（文本 + 字段列表 + 确认/调整/取消按钮）

─── 请求 2（turn_001，确认意图）─────────────────────────────────────
POST /sessions/sess_xxx/messages
{ "type": "confirm_intent", "intentMessageId": "msg_002" }

→ 后端：
  - 读取 msg_002 的 IntentPayload
  - PATCH msg_002 state=confirmed
  - 构建 TurnContext（ConfirmedIntent = payload）
  - 启动 SSE 流

SSE 流：
{"event":"step_started","turnId":"turn_001","stepId":"intent_1","intentMeta":{"id":"intent_1","label":"创建 MainItem","seq":1,"total":1},"data":{"kind":"step_started","intent":{...}}}

{"event":"thinking","turnId":"turn_001","stepId":"intent_1","data":{"kind":"thinking","content":"开始执行 writer..."}}

{"event":"tool_call","turnId":"turn_001","stepId":"intent_1","data":{"kind":"tool_call","callId":"call_001","toolName":"commit_create","arguments":{"entity_type":"main_item","fields":{...}}}}

{"event":"tool_result","turnId":"turn_001","stepId":"intent_1","data":{"kind":"tool_result","callId":"call_001","status":"success","result":{"bizKey":"MI-0023"},"durationMs":300}}

{"event":"card_message","turnId":"turn_001","stepId":"intent_1","messageId":"msg_003","data":{"kind":"card","cardType":"form","status":"prefilled","cardData":{"opType":"create","entityType":"main_item","targetEntity":{"bizKey":"MI-0023","title":"认证模块","bizCode":"MI-0023"},"fields":[...]}}

{"event":"step_phase_done","turnId":"turn_001","stepId":"intent_1","data":{"kind":"step_phase_done","intentId":"intent_1","outcome":"awaiting_user_commit","nextAction":{"type":"await_commit","messageId":"msg_003"}}}

{"event":"turn_phase_done","turnId":"turn_001","data":{"kind":"turn_phase_done","phaseType":"execute","outcome":"awaiting_user_action","nextAction":{"type":"await_commit","messageId":"msg_003"},"intentsTotal":1,"intentsDone":0}}

→ 流关闭
→ 前端：展示 form card，等用户提交

─── 请求 3（turn_001，提交表单）─────────────────────────────────────
POST /sessions/sess_xxx/messages
{ "type": "commit_card", "messageId": "msg_003" }

→ 后端：
  - 读取 msg_003 的 form card payload
  - 调 entity service（不走 LLM）
  - PATCH msg_003 state=submitted
  - persist 跟进 text 消息

JSON 响应：
{
  "messageId": "msg_003",
  "newState": "submitted",
  "createdEntity": { "bizKey": "MI-0023", "title": "认证模块" },
  "followupMessageId": "msg_004",
  "followupContent": "已为你创建 P1 事项「认证模块」（MI-0023）。"
}
```

### 6.2 主动澄清场景

```
─── 请求 1（turn_002，发指令，含缺失字段）─────────────────────────
POST /sessions/sess_xxx/messages
{ "type": "free_text", "content": "创建一个 P1 事项" }

SSE 流：
{"event":"input_rewrite","turnId":"turn_002","data":{"kind":"input_rewrite","original":"创建一个 P1 事项","rewritten":"创建一个 P1 MainItem（缺失：title）","contextUsed":[],"ambiguityNotes":[]}}

{"event":"thinking","turnId":"turn_002","data":{"kind":"thinking","content":"title 缺失，需要澄清。"}}

{"event":"card_message","turnId":"turn_002","messageId":"msg_010","data":{"kind":"card","cardType":"intent","status":"awaiting_confirm","cardData":{"text":"好的，我帮你创建一个 P1 主事项。","intents":[{"id":"intent_1","label":"创建 MainItem","opType":"create","entityType":"main_item","fields":[{"name":"priority","value":"P1","required":true}],"executor":"writer"}],"missingInfo":[{"intentIndex":0,"field":"title","question":"标题是什么？"}]}}}

{"event":"turn_phase_done","turnId":"turn_002","data":{"kind":"turn_phase_done","phaseType":"planner","outcome":"success","nextAction":{"type":"await_clarify"}}}

→ 流关闭
→ 前端：展示意图消息 + 切文本模式（等用户回答 missing_info）

─── 请求 2（turn_002，回答澄清）────────────────────────────────────
POST /sessions/sess_xxx/messages
{ "type": "answer_clarify", "answer": "认证模块", "intentMessageId": "msg_010" }

→ 后端：注入历史 + 回答，重新调 Planner

SSE 流（新意图消息，含完整字段）：
{"event":"input_rewrite","turnId":"turn_002","data":{"kind":"input_rewrite","original":"认证模块","rewritten":"补充：title=认证模块，组合上文意图：创建 P1 MainItem title=认证模块","contextUsed":["prev_turn: 创建 P1 事项"],"ambiguityNotes":[]}}

{"event":"thinking","turnId":"turn_002","data":{"kind":"thinking","content":"信息收齐。"}}

{"event":"text_message","turnId":"turn_002","messageId":"msg_011","data":{"kind":"text","content":"信息已收集完整，请核对：","variant":"info"}}

{"event":"card_message","turnId":"turn_002","messageId":"msg_012","data":{"kind":"card","cardType":"intent","status":"info_complete","cardData":{"text":"","intents":[{"id":"intent_1","label":"创建 MainItem","opType":"create","entityType":"main_item","fields":[{"name":"title","value":"认证模块","required":true},{"name":"priority","value":"P1","required":true}],"executor":"writer"}],"missingInfo":[]}}}

{"event":"turn_phase_done","turnId":"turn_002","data":{"kind":"turn_phase_done","phaseType":"clarify","outcome":"success","nextAction":{"type":"await_confirm_intent","messageId":"msg_012"}}}

→ 流关闭
→ 用户点"理解正确"触发 confirm_intent（请求 3）
```

### 6.3 多意图场景

```
─── 请求 1（turn_003，多意图）──────────────────────────────────────
POST /sessions/sess_xxx/messages
{ "type": "free_text", "content": "创建 P1 事项 + 查我的 P0" }

SSE 流：
{"event":"input_rewrite","turnId":"turn_003","data":{"kind":"input_rewrite","original":"...","rewritten":"两个意图：1.创建 MainItem 2.查询 P0 MainItem","contextUsed":[],"ambiguityNotes":[]}}

{"event":"thinking","turnId":"turn_003","data":{"kind":"thinking","content":"用户两个独立意图。"}}

{"event":"card_message","turnId":"turn_003","messageId":"msg_020","data":{"kind":"card","cardType":"intent","status":"awaiting_confirm","cardData":{"text":"我帮你处理 2 个任务：","intents":[{"id":"intent_1","label":"创建 MainItem","opType":"create","entityType":"main_item","fields":[{"name":"priority","value":"P1"}],"executor":"writer","missingInfo":[{"field":"title","question":"标题？"}]},{"id":"intent_2","label":"查询 P0","opType":"query","entityType":"main_item","fields":[],"executor":"reader","filter":{"assignee":"currentUser","priority":"P0"}}],"missingInfo":[{"intentIndex":0,"field":"title","question":"创建的事项标题是什么？"}]}}}

{"event":"turn_phase_done","turnId":"turn_003","data":{"kind":"turn_phase_done","phaseType":"planner","outcome":"success","nextAction":{"type":"await_clarify"}}}

→ 流关闭，等用户回答 title

→ 用户回答后，重新调 Planner，新意图消息含完整字段
→ 用户确认 → 执行第一个 intent（writer）→ form card 中断
→ intent_2 丢弃（用户需重新发指令触发查询）
```

## 7. 调整意图场景

```
用户点"✎ 我要调整"
→ 前端切文本模式 + 预填用户上一条原文
→ 用户编辑后发送：

POST /sessions/sess_xxx/messages
{ "type": "adjust_intent", "intentMessageId": "msg_002", "newContent": "创建一个 P0 事项叫认证模块" }

→ 后端：
  - PATCH msg_002 state=adjusted
  - persist user msg（newContent）
  - 重新调 Planner（基于 newContent）
  - 推送新意图消息

→ 原意图消息保留在历史中（state=adjusted），新消息为 awaiting_confirm
```

## 8. 边界情况

### 8.1 多意图中断后的恢复

```
用户："创建认证模块 + 查 P0"（2 intents）
→ 确认后执行 intent_1（writer）→ form card 中断
→ intent_2 丢失
→ 用户提交后，需重新发指令触发查询
```

**用户体验**：UI 在意图消息中标注"intent_2 未执行"，引导用户重新发起。

### 8.2 流式中断

```
Planner 流式中网络断开
→ 已 persist 的 trace 保留在 messages 表
→ 已 persist 的意图消息保留（state=awaiting_confirm）
→ 用户可重新确认（意图消息仍有效）或重新发指令
```

### 8.3 刷新页面

```
用户在意图消息等待确认时刷新
→ 前端 state 丢失
→ 从 messages 表拉取历史
→ 看到意图消息（state=awaiting_confirm）
→ 用户仍可点确认（消息 ID 仍在）
→ confirm_intent 请求触发执行
```

→ **意图消息持久化的最大收益**：刷新页面后用户仍能继续未完成的 turn（只要意图消息未被 confirmed/cancelled）。

## 9. 数据流不变量

1. **每个 HTTP 请求独立**——handler 创建 TurnContext，请求结束释放
2. **意图作为 message 持久化**——source of truth
3. **Plan 是临时调度状态**——从已确认意图消息重建，不入库、不前端缓存
4. **跨请求状态通过 messages 表**——意图消息、form card、trace 都在 DB
5. **后端完全无状态**——任何实例处理任何请求，水平扩展天然支持
