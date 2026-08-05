---
created: "2026-08-05"
updated: "2026-08-05"
parent: tech-design.md
---

# State Model: 会话状态的重建、累积与持久化（Load → Accumulate → Flush）

> 返回 [`tech-design.md`](./tech-design.md)
>
> 本文档是 Copilot "会话涉及的信息"（用户信息、会话历史、自身定义上下文）的**权威定义**。核心模型：
>
> **请求开始** → `StateLoader` 从 DB 重建工作状态 `RequestState` → **请求过程中** orchestrator/agent 把新信息（消息、trace、工具调用、错误、草稿）**累积**进 `RequestState`（不直接写库） → **请求结束** `StateApplier.Flush` 把累积的新信息**单事务持久化**回 DB。
>
> 本文形式化定义此前散落且未定义的 4 个类型（`Environment` / `DraftState` / `MessageSnapshot` / `AgentCallAccumulator`）+ `RequestState` 的累加器字段 + Load/Accumulate/Flush 三阶段，并消除 `TurnContext` 与 `AgentRunParams` 的字段漂移。

## 0. 设计来源：借鉴 LangGraph state 思路（及边界）

LangGraph 的 state 模型有几个核心原语。本设计**选择性借鉴**——受本项目既有架构决策约束（[`tech-design.md`](./tech-design.md) 附录 A 明确拒绝跨 turn 持久化 plan/state），**不引入 per-super-step checkpoint**，而是在**请求边界**做 Load → Accumulate → Flush（= LangGraph checkpoint 粒度放到 HTTP 请求级，以 turnID 为 key）。

| LangGraph 原语 | 借鉴? | 映射到本项目 | 说明 |
|---|---|---|---|
| Typed State（共享、强类型） | ✅ | `RequestState`（工作状态）+ `SessionState`/`TurnState`（DB 行镜像） | RequestState 即请求周期内的 transient "信息中心" |
| Reducer（`add_messages` by-id / `operator.add`） | ✅ 轻量 | `DraftState.MergeAnswer`、`AgentCallAccumulator.Append`、RequestState 累加器、`cropHistory` group-aware（§5） | 累积型字段用显式合并函数 |
| Checkpointer（per super-step，`thread_id` 索引） | ✅ 粗粒度 | **请求边界** Load（开始）/ Flush（结束），以 turnID 为 key | 不做 per-step；附录 A 拒绝跨 turn 持久化。请求级粒度够用且保持 stateless |
| Partial update（节点返回 partial） | ✅ | 累加器只收集变更，Flush 单事务写回 | 取代散落 mid-request persist |
| State 承载业务累加器（不止 messages） | ✅ | `PendingMessages`（含 trace 消息）/ `Errors` / `Calls` / `DraftState` | 请求过程中的产出全部累积进 RequestState |

**核心立场**：借鉴的是"**typed State + Load/Accumulate/Flush + 轻量 reducer**"，不是"图执行框架 + per-step checkpoint"。`RequestState` 是**请求周期内的 transient 信息中心**（对标 LangGraph 的运行时 State 对象），但只在请求边界 Load/Flush，不跨请求存活——这符合 stateless HTTP 后端 + turns 独立（附录 A）的约束。

---

## 1. 核心模型：持久化状态 ↔ 工作状态（Load → Accumulate → Flush）

只有两个东西：**持久化状态**（DB 表行）和**工作状态**（`RequestState`，请求周期内的 transient 信息中心）。两者靠 Load / Flush 连接。

```
┌────────────────────────────────────────────────────────────────┐
│ 持久化状态（DB，跨请求存活）                                     │
│   copilot_sessions → SessionState    copilot_turns → TurnState  │
│   copilot_messages / copilot_agent_call_logs                    │
└──────────────┬─────────────────────────────────▲───────────────┘
               │ Load（请求开始）                   │ Flush（请求结束，单事务）
               ▼                                    │
┌────────────────────────────────────────────────────────────────┐
│ 工作状态 · RequestState（= TurnContext，goroutine 生命周期）     │
│                                                                  │
│  ── Load 进来的（只读快照）──                                    │
│    SessionID/TurnID · Status · ConfirmedIntent · Environment     │
│    History []MessageSnapshot · DraftState                        │
│                                                                  │
│  ── Accumulate 累积的新信息（请求过程中产出）──                  │
│    PendingMessages []Message    ← emission 产出的 intent/form…   │
│                                   + consume*Stream 聚合的 trace 消息 │
│    Errors          []ErrorContext ← 执行中的错误                 │
│    Calls           *AgentCallAccumulator ← LLM 调用元数据        │
│    DraftState（可变）            ← clarify 回答合并              │
└────────────────────────────────────────────────────────────────┘
```

### 1.1 三阶段

| 阶段 | 时机 | 动作 | 谁做 |
|---|---|---|---|
| **Load** | 请求开始 | 从 DB 读 session+turn+messages，重建 Draft + 裁剪 History，**初始化空累加器** | `StateLoader.Load`（§3.1） |
| **Accumulate** | 请求过程 | orchestrator/agent 把产出（消息/trace/工具调用/错误/草稿）**累积进 RequestState**，**不直接写库** | orchestrator + agent |
| **Flush** | 请求结束 | **单事务**写回所有累加器：PendingMessages→messages（含 trace 消息）、Calls.Drain()→agent_call_logs、UPDATE turn | `StateApplier.Flush`（§3.3） |

**为什么不再分"三层"**：之前把 SessionState/TurnState/RequestState 并称"三层 State"是过度包装——SessionState/TurnState 只是 DB 行的 Go 镜像（GORM model），RequestState 是请求级工作状态。真正有架构意义的是**持久化边界**（Load/Flush），不是"层数"。

**DraftState 不入库**：[`tech-design.md`](./tech-design.md):599 明确"DraftState 不跨 turn"。clarify 回答本身已作为 user text 消息持久化（[request-model.md](./request-model.md) §6.2），DraftState 是其**派生投影**——Load 时从 messages 重建，Accumulate 时可变（MergeAnswer），但 Flush 时不单独写（其变化已体现在新 persist 的 user/intent 消息里）。

---

## 2. 4 个类型的形式化定义

下列 struct 是**唯一 normative 定义**。其他文档（[interfaces.md](./interfaces.md) §11、[request-model.md](./request-model.md) §2.1、[agent-architecture.md](./agent-architecture.md) §6）以指针引用本文。

### 2.1 Environment（请求上下文，每请求新建，不入库）

```go
// internal/copilot/model/environment.go

type Environment struct {
    UserBizKey  string        // 来源：middleware.GetUserBizKey(c) — 当前唯一接线的用户标识
    UserID      uint          // 来源：session 行 / auth context
    TeamID      *uint         // 来源：TeamScopeMiddleware 注入（AB-003）；全局页（如 /users）为 nil
    TeamName    string        // 冗余存储，prompt 渲染用
    CurrentTime time.Time     // 服务端时间，相对时间改写锚点（如"下周五"→ 计算依据）
    PageContext *PageContext  // 来源：POST body req.PageContext（见 §4）；nil = 无页面上下文
}

type PageContext struct {
    Route        string // 如 "/teams/12/items/main" — 代词消解来源（"它/那个"）
    EntityType   string // 页面主实体类型，如 "main_item"；空 = 无聚焦实体
    EntityBizKey string // 页面聚焦实体 bizKey（如打开的 MainItem 详情）；空 = 无聚焦实体
}
```

**字段 provenance**：

| 字段 | 来源 | 备注 |
|---|---|---|
| UserBizKey | `middleware.GetUserBizKey(c)` | 唯一已接线来源（[request-model.md](./request-model.md):185） |
| UserID / TeamID / TeamName | session 行 + `TeamScopeMiddleware` 注入的 context | 复用现有 RBAC（[tech-design.md](./tech-design.md):203 `internal/middleware/permission.go` / `team_scope.go`） |
| CurrentTime | 服务端 `time.Now()` | 相对时间改写锚点（[agent-architecture.md](./agent-architecture.md) §3.2 input_rewrite 规则 4） |
| PageContext | POST body `req.PageContext` | 见 §4 pageContext 契约 |

### 2.2 DraftState（clarify 累加器，派生，不入库）

```go
// internal/copilot/model/draft_state.go

type DraftState struct {
    // 累积的 clarify 回答：intentIndex → field → 用户回答
    Answers map[int]map[string]string
    // 上一轮 Planner 输出的 missingInfo（判断还需问什么）
    PendingMissing []MissingItem
    // 当前活跃意图消息 id（turn.intent_message_id 的镜像，便于 Agent 取用）
    ActiveIntentMessageID string
}

// MergeAnswer —— reducer：用户每次 answer_clarify 合并进 Answers
func (d *DraftState) MergeAnswer(intentIndex int, field, answer string) {
    if d.Answers == nil {
        d.Answers = map[int]map[string]string{}
    }
    if d.Answers[intentIndex] == nil {
        d.Answers[intentIndex] = map[string]string{}
    }
    d.Answers[intentIndex][field] = answer
}
```

`MissingItem` 复用既有定义（[interfaces.md](./interfaces.md) §11 / [request-model.md](./request-model.md) §3.1）。

**重建**：`StateLoader` 扫描 turn 内所有 `role=user` 且关联 clarify 的消息（`answer_clarify` 请求产生的 user text msg），按时间序 `MergeAnswer` 重放，并取最新 intent 消息的 `missingInfo` 填 `PendingMissing`。详见 §3.2。

### 2.3 MessageSnapshot（Message 行的可序列化投影，供 history）

```go
// internal/copilot/model/message_snapshot.go

type MessageSnapshot struct {
    BizKey    string
    Seq       int
    Role      MsgRole
    Type      MsgType
    Status    MsgStatus
    Content   string          // 文本部分（user 原文 / ai 文本回执）
    CardType  *CardType        // card 消息类型（intent/form/query_result/disambig/fallback）
    Card      json.RawMessage  // card payload
    IntentID  *string          // 按 intent 聚组的 key（group-aware 裁剪用）
    CreatedAt time.Time
}

// snapshotFromMessage —— Message → MessageSnapshot 投影（裁剪掉 trace 等重组装时不需的字段）
func snapshotFromMessage(m Message) MessageSnapshot
```

**与 `Message` 的关系**：投影（非别名）。`Message`（[tech-design.md](./tech-design.md) §3 完整定义）含 `Trace` 等仅供 UI 调试的 bulky 字段；`MessageSnapshot` 只保留注入 prompt 与裁剪所需的最小字段集。

### 2.4 AgentCallAccumulator（原 AgentCallSnapshot 重命名）

```go
// internal/copilot/orchestrator/agent_call_accumulator.go

type AgentCallAccumulator struct {
    calls []AgentCallLog
    mu    sync.Mutex
}

// Append —— reducer：goroutine 内每次 LLM 调用累积
func (a *AgentCallAccumulator) Append(log AgentCallLog) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.calls = append(a.calls, log)
}

// Drain —— flush 到 copilot_agent_call_logs（请求结束前调用）
func (a *AgentCallAccumulator) Drain() []AgentCallLog {
    a.mu.Lock()
    defer a.mu.Unlock()
    out := a.calls
    a.calls = nil
    return out
}
```

`AgentCallLog` 复用既有定义（[er-diagram.md](./er-diagram.md) / schema.sql `copilot_agent_call_logs`）。

**重命名动机**：原名 `AgentCallSnapshot` 暗示"快照"，但实际是请求内**累积**并最终 Drain 落库的累加器。`AgentCallAccumulator` 名副其实，也对应 LangGraph 的累加型 reducer。

---

## 3. StateLoader / StateApplier（替代未定义的接缝）

此前 `buildEnv` / `serializeEnv` / `toProviderMsg` 三个函数只有调用点、无函数体（详见探索结论）。本文形式化为下列接口与函数。

### 3.1 SessionState / TurnState / RequestState 定义

**SessionState**（持久化状态，`copilot_sessions` 行的 Go 镜像，权威列见 [er-diagram.md](./er-diagram.md)）：

```go
type SessionState struct {
    BizKey        string
    UserID        uint
    TeamID        *uint
    TeamName      string
    Title         string
    CurrentTurnID string
    Status        SessionStatus
    LastActiveAt  time.Time
    ExpiresAt     time.Time
}
```

**TurnState**（持久化状态，`copilot_turns` 行的 Go 镜像）：

```go
type TurnState struct {
    BizKey          string
    SessionID       string
    Status          TurnStatus
    UserQueryShort  string
    Summary         string
    IntentsTotal    int
    IntentsDone     int
    IntentMessageID *string
    ConfirmedIntent *IntentPayload
    StartedAt       time.Time
    LastActiveAt    time.Time
    CompletedAt     *time.Time
}
```

**RequestState**（工作状态，goroutine 内的**扁平**结构；Load 时从 SessionState + TurnState + Environment + History 填充，Accumulate 时累加新产出）：

```go
// internal/copilot/orchestrator/state_loader.go

type RequestState struct {
    // 标识
    TurnID    string
    SessionID string

    // ── turn 状态（Load 读入；Accumulate 阶段可改；Flush 写回 turn 表）──
    Status          TurnStatus
    ConfirmedIntent *IntentPayload
    IntentMessageID *string
    Summary         string
    IntentsDone     int

    // 用户输入（free_text 场景）
    UserMessage string

    // ── 上下文（Load 读入，只读）──
    Environment Environment
    History     []MessageSnapshot   // 裁剪后
    DraftState  *DraftState         // 重建（§3.2）；Accumulate 阶段可 MergeAnswer

    // ── Accumulate 累加器（请求过程中产出，Flush 时写回）──
    PendingMessages []Message              // 新消息：emission 产出（intent/form/query_result/disambig）+ consume*Stream 聚合的 trace 消息
    Errors          []ErrorContext        // 执行中的错误上下文（供重试反思 + Flush 时落 turn.status=failed）
    Calls           *AgentCallAccumulator // LLM 调用元数据（Flush 时 Drain 落 agent_call_logs）

    StartedAt time.Time
    mu        sync.Mutex                  // 并发保护（trace 聚合 / 累加器）
}

// ErrorContext —— 错误累加器条目（rs.Errors）
type ErrorContext struct {
    Phase   string // planner / executor / commit
    Code    string // ERR_COPILOT_*
    Message string
    StepID  string // intent id 或空
}

// TurnContext 是 RequestState 的别名（保留旧名减少调用点 churn；权威定义在本文）
type TurnContext = RequestState

type StateLoader interface {
    // Load 重建完整 RequestState：读 session + turn + messages，重建 DraftState，裁剪 History
    Load(ctx context.Context, sessionID, turnID string, env Environment) (*RequestState, error)
}
```

**为什么 RequestState 扁平而非嵌套 Session/Turn**：handler 与 orchestrator 大量直接读 `turnCtx.Status` / `turnCtx.ConfirmedIntent`（见 [request-model.md](./request-model.md) §2、§5）。扁平字段避免 `turnCtx.Turn.Status` 穿透，调用点零改动。SessionState/TurnState 仍是 DB 行的独立类型，StateLoader 加载后把所需字段**拷贝**进 RequestState 的扁平字段。

加载委托既有 `SessionRepository` / `TurnRepository` / `MessageRepository`（[interfaces.md](./interfaces.md) §6），**不新造 repo**。

### 3.2 DraftState 重建（StateLoader 内部）

```
StateLoader.Load(sessionID, turnID, env)
  ├─ session = sessionRepo.Get(sessionID)            → SessionState
  ├─ turn    = turnRepo.Get(turnID)                  → TurnState（持久化部分）
  ├─ msgs    = msgRepo.ListByTurn(turnID, OrderBySeq)
  │
  ├─ draft = reconstructDraft(msgs, turn):
  │     - 扫描 role=user 且源自 answer_clarify 的 text msg（按 seq 升序）
  │     - 对每个回答调 draft.MergeAnswer(intentIndex, field, answer)
  │     - 取最新 type=intent 消息的 missingInfo → draft.PendingMissing
  │     - draft.ActiveIntentMessageID = turn.IntentMessageID
  │
  ├─ history = [snapshotFromMessage(m) for m in msgs]
  ├─ history = cropHistory(history, tokenBudget)     → §5.3 group-aware
  │
  └─ return RequestState{
       TurnID: turn.BizKey, SessionID: session.BizKey,
       Status: turn.Status, ConfirmedIntent: turn.ConfirmedIntent,
       IntentMessageID: turn.IntentMessageID, Summary: turn.Summary, IntentsDone: turn.IntentsDone,
       Environment: env, History: history, DraftState: draft,
       // 累加器初始化为空（Accumulate 阶段填充，Flush 阶段写回）
       PendingMessages: nil, Errors: nil,
       Calls: &AgentCallAccumulator{}, StartedAt: now,
     }
```

### 3.3 StateApplier：Flush（请求结束单事务持久化）

```go
// internal/copilot/orchestrator/state_applier.go

type StateApplier interface {
    // Flush —— 请求结束时单事务写回所有累加器（核心持久化路径，对标 LangGraph checkpointer.put）
    Flush(ctx context.Context, rs *RequestState) error

    // ApplyTurnUpdate —— eager 写 turn 字段（仅用于 commit_card 等非 LLM 同步路径，
    // 或需提前落库的 superseded 标记；正常 LLM 流走 Flush）
    ApplyTurnUpdate(ctx context.Context, turnID string, diff TurnDiff) error
}
```

**Flush 单事务内做的事**：

1. `INSERT copilot_messages`：`rs.PendingMessages`（intent / form / query_result / disambig 消息 **+ trace 消息**——trace 由 `consume*Stream` 在每个 agent 调用结束时聚合为 `Message{type=trace}` 累积进此切片，见 [request-model.md](./request-model.md) §5.4）
2. `INSERT copilot_agent_call_logs`：`rs.Calls.Drain()`
3. `UPDATE copilot_turns`：`Status` / `ConfirmedIntent` / `IntentMessageID` / `Summary` / `IntentsDone` / `CompletedAt`（Accumulate 阶段改过的字段；若 `rs.Errors` 非空则 `Status=failed`）
4. `UPDATE copilot_sessions`：`LastActiveAt`（touch）

**单事务的收益**：要么全部 persist 要么全部回滚——消除了 [`tech-design.md`](./tech-design.md) §4.4 担心的"intent emit 了但 persist 失败 → messageId='' → 卡死 turn"问题（不再有"emit 了但没落库"的中间态）。

**TurnDiff（给 ApplyTurnUpdate 的 partial update，用于非 Flush 同步路径）**：

```go
type TurnDiff struct {
    Status          *TurnStatus
    Summary         *string
    IntentsDone     *int
    IntentMessageID *string
    ConfirmedIntent *IntentPayload
    CompletedAt     *time.Time
}
```

`ApplyTurnUpdate` 映射到既有 `TurnRepository.UpdateStatus` / `UpdateSummary` / `UpdateConfirmedIntent` / `UpdateIntentMessageID`（[interfaces.md](./interfaces.md) §6），在同一事务内执行。`commit_card` handler 用它在自己的事务里同步落 turn 变更（因为 JSON 响应依赖 entity service 写结果，不能等 Flush）。

### 3.4 序列化接缝（替代 serializeEnv / toProviderMsg）

```go
// internal/copilot/prompt/serialize.go

// SerializeEnv 把 Environment 渲染为 system prompt 片段（user/team/page/time 锚点）
func SerializeEnv(env Environment) string

// ToProviderMsg 把 MessageSnapshot 投影为 Provider 消息格式
func ToProviderMsg(s MessageSnapshot) ProviderMsg
```

`ProviderMsg` 见 [interfaces.md](./interfaces.md) §1。这两个函数在 [llm-integration.md](./llm-integration.md) §5 ContextBuilder.Build 中被调用。

---

## 4. pageContext 契约（POST body 显式携带）

### 4.1 请求字段

`MessageRequest`（[interfaces.md](./interfaces.md) §10 / [request-model.md](./request-model.md) §1.3）增加可选字段：

```go
type MessageRequest struct {
    Type            string        `json:"type"`
    Content         string        `json:"content,omitempty"`
    Answer          string        `json:"answer,omitempty"`
    IntentMessageID string        `json:"intentMessageId,omitempty"`
    MessageID       string        `json:"messageId,omitempty"`
    CandidateBizKey string        `json:"candidateBizKey,omitempty"`
    NewContent      string        `json:"newContent,omitempty"`
    RequestID       string        `json:"requestId,omitempty"`
    PageContext     *PageContext  `json:"pageContext,omitempty"` // 新增：前端携带当前页面上下文
}
```

### 4.2 前端发送时机

仅在需要页面上下文的请求类型携带：`free_text` / `answer_clarify` / `adjust_intent`（用户发新指令或调整时）。`confirm_intent` / `select_candidate` / `commit_card` / `cancel` 不需（操作目标已由 messageId/intentMessageId 锁定）。

### 4.3 handler 组装

`buildEnv`（即 handler 内显式组装 `Environment`）：

```go
func (h *CopilotHandler) buildEnv(c *gin.Context, req MessageRequest) Environment {
    userBizKey := middleware.GetUserBizKey(c)        // 既有：唯一接线来源
    teamID, teamName := readTeamScope(c)             // 复用 TeamScopeMiddleware 注入值
    return Environment{
        UserBizKey:  userBizKey,
        TeamID:      teamID,
        TeamName:    teamName,
        CurrentTime: time.Now(),
        PageContext: req.PageContext,                // 来自 POST body，nil 时降级
    }
}
```

**降级**：`PageContext == nil` 时，代词消解（"它/那个"）退回纯历史消息推断（[agent-architecture.md](./agent-architecture.md) §3.2 input_rewrite 规则 1 的回退路径）。功能可用但精度下降。

---

## 5. 轻量 reducer

### 5.1 DraftState.MergeAnswer（§2.2 已定义）

clarify 多轮回答的累积合并。

### 5.2 AgentCallAccumulator.Append / Drain（§2.4 已定义）

请求内 LLM 调用元数据累积，请求结束 Drain 落 `copilot_agent_call_logs`。

### 5.3 cropHistory：group-aware 裁剪（实现见 llm-integration.md §5.4）

`cropHistory` 采用 group-aware 策略（对标 LangGraph `add_messages` 的 by-id 整体性——不拆散同一意图的消息组），单组过大时退回组内 FIFO 兜底。**唯一实现在 [llm-integration.md](./llm-integration.md) §5.4**，本文不重复（避免两处漂移）。当前 turn 的 user msg 永远保留（不在 history 内，由 Build 单独 append）。

---

## 6. TurnContext ↔ AgentRunParams 漂移消除

`AgentRunParams`（[agent-architecture.md](./agent-architecture.md) §6）定义为 `RequestState` 的**按 Agent 调用投影**，加显式构造器：

```go
// internal/copilot/orchestrator/request_state_projectors.go

// ForPlanner —— Planner 调用的投影（含 UserMsg + Draft）
func (rs *RequestState) ForPlanner(userMsg string) AgentRunParams {
    return AgentRunParams{
        SessionID:  rs.SessionID,
        TurnID:     rs.TurnID,
        StepID:     rs.TurnID, // planner 时 stepID = turnID
        UserMsg:    userMsg,
        History:    rs.History,
        DraftState: rs.DraftState,
        Env:        rs.Environment,
        Persist:     rs.stageMessage,  // 累加器回调（见本节末 stageMessage）
        OnAgentCall: rs.Calls.Append,
    }
}

// ForExecutor —— Executor 调用的投影（含 intent + injectedBizKey）
func (rs *RequestState) ForExecutor(intent IntentSpec, injectedBizKey string) AgentRunParams {
    return AgentRunParams{
        SessionID: rs.SessionID,
        TurnID:    rs.TurnID,
        StepID:    intent.ID, // executor 时 stepID = intent.id
        StepParams: map[string]any{
            "op_type":       intent.OpType,
            "entity_type":   intent.EntityType,
            "fields":        intent.Fields,
            "target_entity": intent.TargetEntity,
        },
        History:        rs.History,
        Env:            rs.Environment,
        InjectedBizKey: injectedBizKey, // select_candidate 续跑时注入
        Persist:        rs.stageMessage,
        OnAgentCall:    rs.Calls.Append,
    }
}

// stageMessage —— 累加器回调（注入 AgentRunParams.Persist）：追加进 PendingMessages，
// 返回预生成 bizKey 给 SSE 事件；不写库，Flush 时落 copilot_messages。
func (rs *RequestState) stageMessage(msg Message) (string, error) {
    if msg.BizKey == "" {
        msg.BizKey = snowflake.Next()
    }
    rs.mu.Lock()
    defer rs.mu.Unlock()
    rs.PendingMessages = append(rs.PendingMessages, msg)
    return msg.BizKey, nil
}

// appendError —— 错误累加器：追加进 Errors；Flush 时据此设 turn.status=failed（见 §3.3）。
// 调用方：consume*Stream 捕获 agent error 事件（request-model.md §5.4）、orchestrator 错误路径、
// handler goroutine 捕获 orchestrator 返回 err（request-model.md §2.2）。
func (rs *RequestState) appendError(e ErrorContext) {
    rs.mu.Lock()
    defer rs.mu.Unlock()
    rs.Errors = append(rs.Errors, e)
}
```

**漂移消除**：所有 `AgentRunParams` 字段（`UserMsg` / `StepParams` / `InjectedBizKey` / `History` / `DraftState` / `Env`）均由这两个投影器从 `RequestState` 派生，不再由调用点散落组装。[request-model.md](./request-model.md) §5.2 `ExecuteFromIntent` 与 §5.3 `HandleUserMessage` 的内联 `AgentRunParams{...}` 字面量改为调 `ForExecutor` / `ForPlanner`。

---

## 7. 关键不变量

1. **两个东西，不是三层**——持久化状态（DB）+ 工作状态（RequestState），靠 Load/Flush 连接。
2. **RequestState 是请求周期内的信息中心**——Load 进来只读快照 + 空累加器，Accumulate 累积新产出（消息/trace/工具调用/错误/草稿），Flush 写回。**不跨请求存活**。
3. **TurnContext 即 RequestState**——`type TurnContext = RequestState`（别名），保留旧名减少 churn。
4. **单点持久化**——LLM 流的所有写集中在请求结束的 `Flush` 单事务（取代散落 mid-request persist）；`commit_card` 等同步路径走 `ApplyTurnUpdate` 独立事务。
5. **DraftState 不入库**——Load 从 messages 重建，Accumulate 可变，Flush 不单独写（体现在新 persist 的 user/intent 消息里）。
6. **零 schema 改动**——不新增 `draft_state` 列、不引入 checkpoint 表（遵守 [`tech-design.md`](./tech-design.md) 附录 A）。
7. **pageContext 请求级**——POST body 携带，注入 Environment，不入库；nil 时降级。
8. **AgentRunParams 是投影**——由 `ForPlanner` / `ForExecutor` 从 RequestState 派生。
9. **复用既有 repo**——StateLoader/Applier 委托 `SessionRepository` / `TurnRepository` / `MessageRepository`（[interfaces.md](./interfaces.md) §6）。

---

## 8. 不改的（显式保留既有决策）

- **不引入 checkpoint 表、不跨 turn 持久化 DraftState/Plan**（[`tech-design.md`](./tech-design.md) 附录 A 评估 8 场景全部否决）。
- **status 状态机不动**（见 [state-machines.md](./state-machines.md)）。
- **DB schema 零改动**——DraftState 从既有 messages 重建；pageContext 请求级不入库。
- **既有 entity service / RBAC 不动**——`middleware.GetUserBizKey` + `TeamScopeMiddleware` 照旧。
- **crash-safety 权衡**：Flush-at-end 意味着请求中途 crash（panic/OOM/超时）→ 该轮累积丢失（turn 仍为 `planning`/`executing`，cron 1h 标 `failed`，用户重发）。换取单事务原子性 + 单持久化点。可接受（failed turn 本就可重试）。
