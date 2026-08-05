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
  | { type: "free_text"; content: string; pageContext?: PageContext }                                  // 自由文本指令（携页面上下文）
  | { type: "answer_clarify"; answer: string; intentMessageId: string; pageContext?: PageContext }     // 回答澄清
  | { type: "confirm_intent"; intentMessageId: string }                 // 确认意图
  | { type: "adjust_intent"; intentMessageId: string; newContent: string; pageContext?: PageContext }  // 调整意图
  | { type: "select_intent"; messageId: string; candidateIntentId: string } // 选候选意图（decision=show_candidates 后）
  | { type: "select_candidate"; messageId: string; candidateBizKey: string } // 选候选实体（disambig 后）
  | { type: "commit_card"; messageId: string }                          // 提交表单
  | { type: "cancel"; messageId?: string }                              // 取消
```

响应：
- 自由文本 / answer_clarify / adjust_intent / confirm_intent / select_candidate → **SSE 流**（无前缀 JSON 内容）
- select_intent → **JSON**（候选升级为 IntentSpec，不走 LLM）
- commit_card → **JSON**（不走 LLM，直接 entity service）
- cancel → **JSON**（标记状态）

### 1.4 PATCH 端点（保留 RESTful）

卡片字段编辑、展开/折叠等就地更新通过独立端点：

```
PATCH /api/v1/copilot/messages/:id
```

不属于"用户向对话添加内容"，所以不合并到 POST /messages。

### 1.5 响应模式决策

**核心规则**：触发 LLM 调用 → **SSE 流**；纯 DB 操作或状态变更 → **JSON**。

**为什么 LLM 必须流式**：
- LLM 输出是 token-by-token（秒级延迟），用户需要 thinking / tool_call 进度反馈
- 支持中途 cancel（`ctx.Done()` 贯穿 channel → Provider → GLM）
- 不流式 → 用户盯着 spinner 10s，体验差

**为什么无 LLM 用 JSON**：
- 纯 DB 操作 <500ms，单一离散结果
- 无渐进式披露价值
- 客户端代码简单（无需流解析器）

#### POST /sessions/:id/messages（按 type 分派）

| Request Type | LLM 调用 | 响应模式 | 响应内容 |
|--------------|---------|---------|---------|
| `free_text` | Planner | **SSE** | thinking 流 → input_rewrite → card_message(intent) → turn_phase_done |
| `answer_clarify` | Planner（重跑） | **SSE** | 同上（注入回答 + 历史） |
| `adjust_intent` | Planner（重跑） | **SSE** | PATCH 旧意图 → thinking → 新 intent card_message |
| `confirm_intent` | Executor（按 plan） | **SSE** | step_started × N → tool_call/tool_result → form/query_result card_message |
| `select_candidate` | Executor（续跑） | **SSE** | 注入 bizKey → 后续 tool_call → card_message |
| `commit_card` | ✗ | **JSON** | `{messageId, newState, createdEntity, followupMessageId, followupContent}` |
| `select_intent` | ✗ | **JSON** | `{messageId, intentCardData, turnStatus: "awaiting_confirm_intent"}` |
| `cancel` | ✗ | **JSON** | `{cancelled: true, turnStatus: "cancelled"}` |

#### 其他端点（全部 JSON）

| 端点 | 用途 | LLM |
|------|------|-----|
| `POST /sessions` | 创建会话 | ✗ |
| `GET /sessions` | 列出会话（分页） | ✗ |
| `GET /sessions/:id` | 会话详情 | ✗ |
| `DELETE /sessions/:id` | 软删除 | ✗ |
| `GET /sessions/:id/messages` | 历史消息（分页） | ✗ |
| `PATCH /messages/:id` | 卡片字段编辑 / 展开 / 应用 diff | ✗ |
| `GET /health` | 健康探针 | ✗ |

#### 边界情况

1. **`commit_card` 的 followupContent**：当前为规则提取（"已为你创建 X（bizKey）"），非 LLM 生成 → JSON。若未来改 LLM 生成需切 SSE。
2. **`select_candidate`**：给已停顿的 Executor 注入 bizKey 续跑——前一个 SSE 流已关闭，必须新起一个流。
3. **多 intent 执行**：单个 `confirm_intent` 触发多个 Executor，全部串行在**一个 SSE 流**内（前端看到一个流，内部多段 step 序列）。
4. **error 事件**：SSE 流中任何环节失败（Planner / Executor / Provider）→ emit `error` 事件后关闭流，turn 标记 `failed`。
5. **客户端断开**：`c.Request.Context()` 取消时，所有下游 channel 关闭、goroutine 退出。已 persist 的 trace 和意图消息保留，用户刷新后仍可继续未完成的 turn。

## 2. TurnContext（与 goroutine 同生命周期）

> **权威定义**：TurnContext 即工作状态 `RequestState`（`type TurnContext = RequestState`），完整定义含 Environment / DraftState / MessageSnapshot / AgentCallAccumulator 的字段、provenance、StateLoader 重建见 [`state-model.md`](./state-model.md)。本节保留结构概览供请求流阅读。
>
> `h.buildEnv(c)` 已改为 `h.buildEnv(c, req)`（读 `req.PageContext` 注入 Environment），函数体见 [`state-model.md`](./state-model.md) §4.3。

### 2.1 结构定义

```go
// internal/copilot/orchestrator/turn_context.go
// TurnContext 是 RequestState 的别名（权威定义见 state-model.md §3.1，勿在此重复声明以免漂移）。
type TurnContext = RequestState
```

**字段概览**（权威定义见 [`state-model.md`](./state-model.md) §3.1 `RequestState`）：
- **标识**：TurnID / SessionID
- **turn 状态**（Load 读入，Accumulate 可改，Flush 写回）：Status / ConfirmedIntent / IntentMessageID / Summary / IntentsDone
- **用户输入**：UserMessage（free_text 场景）
- **上下文**（只读）：Environment / History []MessageSnapshot / DraftState
- **累加器**（Accumulate 填充，Flush 写回）：PendingMessages（含 trace 消息）/ Errors / Calls（*AgentCallAccumulator）

**关键设计点**：
- **Status 字段**：Load 从 turn 表读入，Orchestrator 据此决定路由分支；Accumulate 阶段可改（如 → executing），Flush 写回。
- **ConfirmedIntent**：confirm_intent 时由 handler 写入 turnCtx，Flush 写回 turn.confirmed_intent。
- **累加器是 Accumulate 阶段的产出收集处**——orchestrator/agent 把新消息/错误/调用元数据累积进 turnCtx，请求结束 Flush 单事务持久化（见 §5 intro）。
- **不再有 `Plan` 字段**——plan 是从 `ConfirmedIntent` 派生的临时状态。

### 2.2 Handler 创建 TurnContext

```go
// internal/copilot/handler/copilot_handler.go

func (h *CopilotHandler) PostMessage(c *gin.Context) {
    var req MessageRequest
    // 用 ShouldBindBodyWith：上游 UserStreamGuard/TurnInFlightGuard 已用同方法缓存 body，
    // 这里重读缓存即可（若中间件未跑过，ShouldBindBodyWith 也能首次绑定）
    if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
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
    case "select_intent":
        h.handleSelectIntent(c, req)   // 返回 JSON（候选升级为 IntentSpec，不走 LLM）
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

    // 事务：创建 Turn + persist user msg + UPDATE session.current_turn_id
    // 三者同事务（state-machines.md §6 TurnInFlightGuard 依赖 current_turn_id 非空）
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

        // 3. UPDATE session.current_turn_id（TurnInFlightGuard 与 supersession 逻辑依赖此字段）
        // 若省略此调用，sess.CurrentTurnID 永远是 ''，guard 形同虚设、superseded 永不触发
        h.sessionRepo.UpdateCurrentTurn(c, sessionID, turnID)

        return nil
    })

    // 3. 构建 TurnContext
    turnCtx := &TurnContext{
        TurnID:      turnID,
        SessionID:   sessionID,
        Status:      TurnStatusPlanning,
        UserMessage: req.Content,
        Environment: h.buildEnv(c, req),
        History:     h.msgRepo.ListBySession(c, sessionID, 50, 0),
        Calls:       &AgentCallAccumulator{}, // 累加器初始化（Accumulate 阶段填充，Flush 写回）
        StartedAt:   time.Now(),
    }

    // 4. 启动 SSE 流 + goroutine
    h.startSSEStream(c)
    eventCh := make(chan sse.Event, 64)
    go func() {
        defer close(eventCh)
        if err := h.orchestrator.HandleUserMessage(c.Request.Context(), turnCtx, eventCh); err != nil {
            turnCtx.appendError(ErrorContext{Phase: "planner", Message: err.Error()})
        }
        if err := h.stateApplier.Flush(c.Request.Context(), turnCtx); err != nil { // Load→Accumulate→Flush
            eventCh <- sse.ErrorEvent(turnCtx.TurnID, fmt.Errorf("persist failed: %w", err)) // Flush 失败也通知前端
        }
    }()
    h.streamEvents(c, eventCh)
}

func (h *CopilotHandler) handleConfirmIntent(c *gin.Context, req MessageRequest) {
    // 1. 读取意图消息
    intentMsg, _ := h.msgRepo.GetByBizKey(c, req.IntentMessageID)
    var intentPayload IntentPayload
    json.Unmarshal(intentMsg.Card, &intentPayload)

    // 2. 从 turn 表读取 turn
    turn, _ := h.turnRepo.Get(c, intentMsg.TurnID)

    // 3. PATCH 意图消息 status=confirmed（existing-row 更新，eager 例外——
    //    Flush 只 INSERT PendingMessages + UPDATE turn，不更新既有消息 status，见 tech-design.md §4.4）
    h.msgRepo.UpdateStatus(c, intentMsg.BizKey, MsgStatusIntentConfirmed)

    // 4. Load 工作状态 + mutate（Status=executing + ConfirmedIntent；由 Flush 写回 turn 表）
    turnCtx, _ := h.stateLoader.Load(c.Request.Context(), turn.SessionID, turn.BizKey, h.buildEnv(c, req))
    turnCtx.Status = TurnStatusExecuting
    turnCtx.ConfirmedIntent = &intentPayload

    // 5. 启动 SSE 流 + 执行
    h.startSSEStream(c)
    eventCh := make(chan sse.Event, 64)
    go func() {
        defer close(eventCh)
        if err := h.orchestrator.ExecuteFromIntent(c.Request.Context(), turnCtx, eventCh); err != nil {
            turnCtx.appendError(ErrorContext{Phase: "execute", Message: err.Error()})
        }
        if err := h.stateApplier.Flush(c.Request.Context(), turnCtx); err != nil { // Load→Accumulate→Flush
            eventCh <- sse.ErrorEvent(turnCtx.TurnID, fmt.Errorf("persist failed: %w", err))
        }
    }()
    h.streamEvents(c, eventCh)
}
```

> **其余 handler**（`handleAnswerClarify` / `handleAdjustIntent` / `handleSelectCandidate`）遵循与上述相同的 **SSE 模式**：`StateLoader.Load` → 对应 Orchestrator 方法（累积进 `turnCtx`）→ `stateApplier.Flush`（goroutine 内、流关闭前；捕获 orchestrator 返回 err 进 `appendError`，捕获 Flush err 发 final error 事件）。
> **JSON 模式 handler**（`handleSelectIntent` / `handleCommitCard` / `handleCancel`）无 SSE 流：用 `StateApplier.ApplyTurnUpdate`（commit_card 另加 Dispatcher）在自己的事务内同步落库，不经 Flush。

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

    // 路由决策（见 agent-architecture.md §3.4）：confirm / show_candidates / cannot_understand
    Decision string `json:"decision,omitempty"`
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
| 调整意图 | `adjust_intent` | PATCH 旧意图 state=adjusted + 重新调 Planner | Planner | SSE 流 |
| 确认意图 | `confirm_intent` | 从意图消息重建 plan + 调 executor | Executor（按 plan） | SSE 流 |
| 选候选意图 | `select_intent` | 候选升级为 IntentSpec + persist | ❌ | JSON |
| 选候选 | `select_candidate` | 注入 bizKey + 调 executor | Executor | SSE 流 |
| 提交表单 | `commit_card` | 直接调 entity service | ❌ | JSON |
| 取消 | `cancel` | PATCH 状态 | ❌ | JSON |

> 响应模式决策规则与完整矩阵见 §1.5。

## 5. 路由实现

> **Load → Accumulate → Flush**（见 [state-model.md](./state-model.md) §1）：Handler 先 `StateLoader.Load` 重建 `RequestState`（=`TurnContext`），传给下列 Orchestrator 方法；Orchestrator **只累积**新产出进 `turnCtx`（PendingMessages（含 trace 消息）/ Errors / Calls / DraftState / turn 字段），**不直接写库**；Handle* 方法返回后，**Handler 调 `StateApplier.Flush(turnCtx)` 单事务持久化**（在 SSE 流关闭前）。messageId 预生成（snowflake），SSE 事件先携带、Flush 时落库。

### 5.1 Orchestrator 接口

```go
// internal/copilot/orchestrator/orchestrator.go

type Orchestrator struct {
    planner      agent.Agent
    registry     *agent.Registry
    ctxBuilder   prompt.ContextBuilder
    stateApplier StateApplier            // Flush 由 handler 在方法返回后调（见 §2.2 goroutine）；同包，无需前缀
    dispatcher   service.Dispatcher
    provider     llm.Provider
}
// （Orchestrator 不再持有 msgRepo/turnRepo —— 写库集中在 StateApplier.Flush，取代散落 mid-request persist）

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
        
        // 投影自 RequestState（ForExecutor）；Executor emission 累积进 turnCtx.PendingMessages
        stream, err := executor.StreamRun(ctx, turnCtx.ForExecutor(spec, ""))
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

> **Accumulate，不 persist**：本方法把 intent 消息 + turn 字段变更累积进 `turnCtx`（RequestState），**不写库**；持久化由 Handler 在方法返回后调 `StateApplier.Flush(turnCtx)` 单事务完成（见 §5 intro 与 §2.2 goroutine）。messageId 预生成，SSE 事件先用。

```go
func (o *Orchestrator) HandleUserMessage(
    ctx context.Context, turnCtx *TurnContext, eventCh chan<- sse.Event,
) error {
    // 调 Planner（投影自 RequestState；thinking/tool 事件由 consumePlannerStream 聚合为 trace 消息进 PendingMessages，见 §5.4）
    plannerStream, err := o.planner.StreamRun(ctx, turnCtx.ForPlanner(turnCtx.UserMessage))
    if err != nil {
        turnCtx.appendError(ErrorContext{Phase: "planner", Message: err.Error()})
        eventCh <- errorEvent(turnCtx, err)
        return err
    }

    // 消费 Planner 流：事件实时 emit + 累积进 turnCtx（不写库）
    intentPayload := o.consumePlannerStream(ctx, plannerStream, turnCtx, eventCh)

    // ── Accumulate：intent 消息 + turn 字段变更进 turnCtx（不 persist）──
    intentMsgID := snowflake.Next() // 预生成：SSE 事件先用，Flush 时落库
    turnCtx.PendingMessages = append(turnCtx.PendingMessages, Message{
        BizKey:        intentMsgID,
        Role:          RoleAI, Type: TypeIntent,
        TurnID:        turnCtx.TurnID, SessionID: turnCtx.SessionID,
        Content:       intentPayload.Text,
        IntentPayload: marshal(intentPayload),
    })
    turnCtx.IntentMessageID = &intentMsgID
    if len(intentPayload.MissingInfo) > 0 {
        turnCtx.Status = TurnStatusAwaitingClarify
    } else {
        turnCtx.Status = TurnStatusAwaitingConfirmIntent
    }

    // emit card_message（用预生成 messageId；Flush 在 Handler 端、流关闭前执行）
    eventCh <- cardMessageEvent(turnCtx, "intent", intentPayload.State, intentPayload, intentMsgID)

    // 推送 turn_phase_done
    var nextAction NextAction
    if len(intentPayload.MissingInfo) > 0 {
        nextAction = NextAction{Type: "await_clarify"}
    } else {
        nextAction = NextAction{Type: "await_confirm_intent", MessageID: intentMsgID}
    }

    eventCh <- turnPhaseDoneEvent(turnCtx, "planner", "success", nextAction)
    return nil
}
```

### 5.4 consume*Stream：trace 聚合与 PendingMessages 累积

`consumePlannerStream` / `consumeExecutorStream` 消费 Agent 事件流，做三件事：① 转发事件给 `eventCh`（SSE）；② 聚合 thinking / tool_call / tool_result 进 per-call trace；③ 流结束时把聚合的 trace 作为 `Message{type=trace}` 累积进 `turnCtx.PendingMessages`（经 `stageMessage` 预生成 bizKey）。

```go
// internal/copilot/orchestrator/stream_consumer.go

// consumePlannerStream —— 消费 Planner 流，返回解析出的 IntentPayload。
// 流结束时聚合 trace 为一条 trace 消息累积进 PendingMessages（planner trace 无 intent_id）。
func (o *Orchestrator) consumePlannerStream(
    ctx context.Context, stream <-chan sse.Event,
    turnCtx *TurnContext, eventCh chan<- sse.Event,
) IntentPayload {
    o.consumeAgentStream(ctx, stream, turnCtx, eventCh, nil /*planner trace 无 intent_id*/)
    // ... 从 card_message(intent) 事件解析 IntentPayload
    return intent
}

// consumeExecutorStream —— 消费 Executor 流，返回执行 outcome。trace 消息带 intent_id。
func (o *Orchestrator) consumeExecutorStream(
    ctx context.Context, stream <-chan sse.Event,
    turnCtx *TurnContext, spec IntentSpec, eventCh chan<- sse.Event,
) ExecOutcome {
    intentID := spec.ID
    o.consumeAgentStream(ctx, stream, turnCtx, eventCh, &intentID)
    // ... 解析 outcome（NeedsUserInteraction / CardMessageID / NextActionType）
    return outcome
}

// consumeAgentStream —— 共用聚合器：转发 SSE + 聚合 trace + 流结束追加 trace 消息
func (o *Orchestrator) consumeAgentStream(
    ctx context.Context, stream <-chan sse.Event,
    turnCtx *TurnContext, eventCh chan<- sse.Event, intentID *string,
) {
    var thinking strings.Builder
    var actions []TraceAction
    startedAt := time.Now()
    traceMsgID := snowflake.Next() // 预生成 trace messageId
    failed := false

    for ev := range stream {
        eventCh <- ev // ① 转发给 SSE（thinking/tool_call/tool_result/card_message/error 等）
        switch ev.Kind {
        case "thinking":
            thinking.WriteString(ev.Content)
        case "tool_call":
            actions = append(actions, TraceAction{
                CallID: ev.CallID, ToolName: ev.ToolName,
                Arguments: ev.Arguments, StartedAt: time.Now(),
            })
        case "tool_result":
            pairToolResult(actions, ev) // 配对填入最后一条 action 的 Result/Error/Status/DurationMs
        case "error":
            failed = true
            turnCtx.appendError(ErrorContext{Phase: "agent", Message: ev.Error.Error(), StepID: deref(intentID)})
        }
    }

    // ③ 流结束：聚合 trace 为 Message{type=trace} 累积进 PendingMessages
    turnCtx.stageMessage(Message{
        BizKey:   traceMsgID,
        Role:     RoleAI, Type: TypeTrace,
        TurnID:   turnCtx.TurnID, SessionID: turnCtx.SessionID,
        IntentID: intentID,
        Trace: &TracePayload{
            Thinking:  thinking.String(),
            Actions:   actions,
            StartedAt: startedAt, EndedAt: time.Now(),
            Status: ternary(failed, "failed", "done"),
        },
    })
}
```

**关键点**：
- **trace 是 per-agent-call**：每个 Planner/Executor 调用产生一条 trace 消息（planner 的 `intent_id=nil`，executor 的 `intent_id=spec.ID`），全部累积进 `PendingMessages`，Flush 时统一落 `copilot_messages`。
- **trace messageId 预生成**（`traceMsgID := snowflake.Next()`），未来若需 SSE 携带 trace freeze 事件可复用此 id（当前 trace 随其他 PendingMessages 在 Flush 落库，不在 SSE 单独发 freeze 事件）。
- **错误路径**：Agent 流以 error 事件结束时，`consumeAgentStream` 仍聚合已收到的 trace（`Status="failed"`）追加进 PendingMessages，同时 `turnCtx.appendError` 记错误（Flush 据 `rs.Errors` 设 `turn.status=failed`，见 §5.3 与 llm-integration.md §2）。
- **TracePayload / TraceAction** 复用既有定义（[interfaces.md](./interfaces.md) §11）。
- 此设计消除了原先的 `RequestState.TraceBuffer` 字段（声明却无人写的幽灵累加器）——trace 不再是 turn 级 buffer，而是 per-call 消息进 PendingMessages。

## 6. 完整时序示例

### 6.1 单意图写操作（完整流程）

> **关键：写操作不进 LLM 流**。Executor 只通过 `emit_form_card` 推送**预填表单卡片**（`targetEntity.bizKey` 留空）；真实 DB 写由用户点提交后 `commit_card` Handler 同步调 Dispatcher → 现有 entity service。详见 [`agent-architecture.md`](./agent-architecture.md) §2、§8.2 与 [`security.md`](./security.md) §7.3。

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

SSE 流（Writer Executor 仅组装预填字段，不调任何写工具）：
{"event":"step_started","turnId":"turn_001","stepId":"intent_1","intentMeta":{"id":"intent_1","label":"创建 MainItem","seq":1,"total":1},"data":{"kind":"step_started","intent":{...}}}

{"event":"thinking","turnId":"turn_001","stepId":"intent_1","data":{"kind":"thinking","content":"字段齐备，组装预填表单。"}}

{"event":"card_message","turnId":"turn_001","stepId":"intent_1","messageId":"msg_003","data":{"kind":"card","cardType":"form","status":"prefilled","cardData":{"opType":"create","entityType":"main_item","targetEntity":{"bizKey":"","title":""},"fields":[{"name":"title","value":"认证模块","required":true},{"name":"priority","value":"P1","required":true}]}}}

{"event":"step_phase_done","turnId":"turn_001","stepId":"intent_1","data":{"kind":"step_phase_done","intentId":"intent_1","outcome":"awaiting_user_commit","nextAction":{"type":"await_commit","messageId":"msg_003"}}}

{"event":"turn_phase_done","turnId":"turn_001","data":{"kind":"turn_phase_done","phaseType":"execute","outcome":"awaiting_user_action","nextAction":{"type":"await_commit","messageId":"msg_003"},"intentsTotal":1,"intentsDone":0}}

→ 流关闭
→ 前端：展示预填 form card（targetEntity.bizKey 为空，表示尚未创建实体），等用户提交

─── 请求 3（turn_001，提交表单）─────────────────────────────────────
POST /sessions/sess_xxx/messages
{ "type": "commit_card", "messageId": "msg_003", "requestId": "req_uuid_v4" }

→ 后端（handleCommitCard，事务内）：
  - 读取 msg_003 的 form card payload
  - 幂等检查：copilot_idempotency_keys WHERE request_id=req_uuid_v4
      - 命中 → 返回上次结果（防网络重试，见 §7）
      - 未命中 → 继续：
  - INSERT copilot_idempotency_keys(request_id, message_id)
  - 调 Dispatcher.Dispatch(FormCard{opType:create, entityType:main_item, Fields:[...]})
      - Dispatcher → MainItemService.Create（复用既有 RBAC + bizKey + snowflake）
      - 返回 bizKey=MI-0023
  - UPDATE msg_003 status=submitted + cardData.targetEntity.bizKey=MI-0023
  - persist 跟进 text 消息（msg_004）

JSON 响应：
{
  "messageId": "msg_003",
  "newState": "submitted",
  "createdEntity": { "bizKey": "MI-0023", "title": "认证模块", "bizCode": "MI-0023" },
  "followupMessageId": "msg_004",
  "followupContent": "已为你创建 P1 事项「认证模块」（MI-0023）。"
}
```

**为什么写不在 LLM 流里**：
1. PRD Flow step 7 要求"用户点击提交 → 调用现有 API 端点"——提交动作才触发写
2. 用户放弃 form card（superseded / 网络断 / 浏览器关）时无孤儿实体（只留一条 `status=prefilled` 的 form card 消息，无 DB 实体副作用）
3. LLM 流可中途失败，已 persist 的 intent/form 消息仍可恢复；但失败的 LLM 不该已写过半实体

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
