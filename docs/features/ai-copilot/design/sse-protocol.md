---
created: "2026-06-26"
updated: "2026-06-26"
parent: tech-design.md
---

# SSE Protocol

> 返回 [`tech-design.md`](./tech-design.md)
>
> 注意：本协议使用 **SSE 的标准 MIME**（`text/event-stream`）+ **无前缀 JSON 内容**（不是标准 SSE 的 `data:` 前缀格式）。前端用 `fetch + ReadableStream` 解析。
>
> **哪些请求走 SSE、哪些走 JSON**：见 [`request-model.md`](./request-model.md) §1.5（核心规则：触发 LLM 调用 → SSE；纯 DB 操作 → JSON）。

## 1. 协议格式

### 1.1 HTTP 头

**响应头**：
```
Content-Type: text/event-stream; charset=utf-8
Cache-Control: no-cache, no-transform
Connection: keep-alive
X-Accel-Buffering: no
```

**为什么用 `text/event-stream` MIME**：
- 兼容标准 SSE 语义（中间代理、nginx、CDN 识别为流式响应）
- 不被反向代理缓冲（保证流式实时性）
- 业界广泛认知

**为什么内容用无前缀 JSON**：
- 简洁（每行纯 JSON，无 `data:` / `event:` 冗余前缀）
- 调试友好（单行 grep / jq）
- 前端解析简单（split('\n') + JSON.parse）

> **协议澄清**：SSE 规范要求 `data:` 前缀，但浏览器 EventSource API 才会强制解析这个格式。如果用 `fetch + ReadableStream` 手动解析，则可以解析任意格式的内容。本设计选择 `text/event-stream` MIME + 无前缀 JSON 内容，兼顾标准 MIME 标识与简洁格式，前端用 fetch 解析。

### 1.2 内容格式

每行一个独立 JSON 对象，用 `\n` 分隔。

```
{"event":"input_rewrite","turnId":"turn_001","timestamp":1734900000000,"data":{"kind":"input_rewrite","original":"...","rewritten":"..."}}
{"event":"thinking","turnId":"turn_001","timestamp":1734900000200,"data":{"kind":"thinking","content":"..."}}
{"event":"text_message","turnId":"turn_001","timestamp":1734900000500,"messageId":"msg_002","data":{"kind":"text","content":"..."}}
```

## 2. 顶层泛型结构

```go
// internal/copilot/sse/event.go

type EventType string
const (
    EventInputRewrite   EventType = "input_rewrite"
    EventThinking       EventType = "thinking"
    EventToolCall       EventType = "tool_call"
    EventToolResult     EventType = "tool_result"
    EventTextMessage    EventType = "text_message"
    EventCardMessage    EventType = "card_message"
    EventStepStarted    EventType = "step_started"
    EventStepCompleted  EventType = "step_completed"
    EventStepPhaseDone  EventType = "step_phase_done"
    EventTurnPhaseDone  EventType = "turn_phase_done"
    EventTurnDone       EventType = "turn_done"
    EventError          EventType = "error"
)

type Event[T any] struct {
    Event      EventType   `json:"event"`
    TurnID     string      `json:"turnId"`
    StepID     string      `json:"stepId,omitempty"`
    IntentMeta *IntentMeta `json:"intentMeta,omitempty"`
    Timestamp  int64       `json:"timestamp"`
    MessageID  string      `json:"messageId,omitempty"`
    Data       T           `json:"data"`
}
```

**注意**：不再有 `EventMissingInfo` 和 `EventPlanReady`——这些信息内嵌在 `intent` 类型的 card_message 中（见 §3）。

## 3. Payload 类型族（kind 判别）

```go
// internal/copilot/sse/payload.go

// ─── 文本类 ──────────────────────────────────────
type TextPayload struct {
    Kind    string `json:"kind"`              // 永远 "text"
    Content string `json:"content"`
    Variant string `json:"variant,omitempty"` // intent_echo / clarify / followup / info
}

// ─── 卡片类（discriminated union） ────────────────
type CardPayload struct {
    Kind      string          `json:"kind"`      // 永远 "card"
    CardType  string          `json:"cardType"`  // intent / form / query_result / disambig / fallback
    Status    string          `json:"status"`    // 多态状态（按 cardType 解释，对齐 messages.status）
    CardData  json.RawMessage `json:"cardData"`
}

// CardType 枚举：
// - intent: 意图消息（Planner 推送，含文本 + 结构化字段，等用户确认）
// - form: 表单卡片（Executor 推送，等用户提交）
// - query_result: 查询结果（Executor 推送）
// - disambig: 歧义消解（Executor 推送，等用户选候选）
// - fallback: 降级提示（超时/不可用）

// ─── 意图卡片数据（CardType=intent） ─────────────────────────
type IntentCardData struct {
    Text       string        `json:"text"`         // 自然语言意图回执
    Intents    []IntentSpec  `json:"intents"`      // 结构化意图规格
    MissingInfo []MissingItem `json:"missingInfo,omitempty"` // 主动澄清
}

type IntentSpec struct {
    ID           string       `json:"id"`
    Label        string       `json:"label"`        // "创建 MainItem"
    OpType       string       `json:"opType"`       // create / query / update / move
    EntityType   string       `json:"entityType"`
    Fields       []FieldState `json:"fields"`
    TargetEntity *EntityRef   `json:"targetEntity,omitempty"`
    Executor     string       `json:"executor"`     // writer / reader / updater / mover
    Filter       map[string]any `json:"filter,omitempty"` // query 类型的过滤器
}

type MissingItem struct {
    IntentIndex int    `json:"intentIndex"` // 对应 intents[int] 的索引
    Field       string `json:"field"`
    Question    string `json:"question"`
    Hint        string `json:"hint,omitempty"`
}

// 意图卡片状态（IntentPayload.State）
const (
    IntentStateAwaitingConfirm = "awaiting_confirm"
    IntentStateInfoComplete    = "info_complete"    // clarify 收齐后
    IntentStateConfirmed       = "confirmed"
    IntentStateAdjusted        = "adjusted"
    IntentStateCancelled       = "cancelled"
)

// ─── 表单卡片数据（CardType=form） ──────────────────────────
type FormCardData struct {
    OpType       string       `json:"opType"`
    EntityType   string       `json:"entityType"`
    TargetEntity *EntityRef   `json:"targetEntity"`
    Fields       []FieldState `json:"fields"`
    Errors       *FormErrors  `json:"errors,omitempty"`
    RetryCount   int          `json:"retryCount"`
    DiffOverlay  *DiffOverlay `json:"diffOverlay,omitempty"`
}

// ─── 查询结果卡片数据（CardType=query_result） ────────────────
type QueryResultCardData struct {
    Summary    string         `json:"summary"`
    Records    []EntityRecord `json:"records"`
    Truncated  bool           `json:"truncated"`
    EntityType string         `json:"entityType"`
}

// ─── 歧义消解卡片数据（CardType=disambig） ───────────────────
type DisambigCardData struct {
    OriginalRef string      `json:"originalRef"`
    Candidates  []EntityRef `json:"candidates"`
    IntentID    string      `json:"intentId"`  // 关联的 intent
}

// ─── 降级卡片数据（CardType=fallback） ──────────────────────
type FallbackCardData struct {
    FallbackType string `json:"fallbackType"` // timeout / unavailable / quota_exceeded / parse_failed
    Route        string `json:"route"`
}

// ─── Planner 阶段专用 ─────────────────────────────
type InputRewritePayload struct {
    Kind           string   `json:"kind"`      // "input_rewrite"
    Original       string   `json:"original"`
    Rewritten      string   `json:"rewritten"`
    ContextUsed    []string `json:"contextUsed"`
    AmbiguityNotes []string `json:"ambiguityNotes"`
}

type ThinkingPayload struct {
    Kind    string `json:"kind"`  // "thinking"
    Content string `json:"content"`
}

// ─── 工具调用 ─────────────────────────────────────
type ToolCallPayload struct {
    Kind      string         `json:"kind"`  // "tool_call"
    CallID    string         `json:"callId"`
    ToolName  string         `json:"toolName"`
    Arguments map[string]any `json:"arguments"`
}

type ToolResultPayload struct {
    Kind       string         `json:"kind"`  // "tool_result"
    CallID     string         `json:"callId"`
    Status     string         `json:"status"`  // success / error
    Result     map[string]any `json:"result,omitempty"`
    Error      string         `json:"error,omitempty"`
    DurationMs int            `json:"durationMs"`
}

// ─── 控制事件 ─────────────────────────────────────
type StepStartedPayload struct {
    Kind   string     `json:"kind"`  // "step_started"
    Intent IntentMeta `json:"intent"`
}

type StepCompletedPayload struct {
    Kind     string `json:"kind"`  // "step_completed"
    IntentID string `json:"intentId"`
    Outcome  string `json:"outcome"`  // committed / cancelled / failed / awaiting_user_commit
}

type StepPhaseDonePayload struct {
    Kind       string     `json:"kind"`  // "step_phase_done"
    IntentID   string     `json:"intentId"`
    Outcome    string     `json:"outcome"`
    NextAction NextAction `json:"nextAction"`
}

type TurnPhaseDonePayload struct {
    Kind         string     `json:"kind"`  // "turn_phase_done"
    PhaseType    string     `json:"phaseType"`    // planner / clarify / execute
    Outcome      string     `json:"outcome"`      // success / failed / cancelled / awaiting_user_action
    NextAction   NextAction `json:"nextAction"`
    IntentsTotal int        `json:"intentsTotal,omitempty"`
    IntentsDone  int        `json:"intentsDone,omitempty"`
}

type TurnDonePayload struct {
    Kind         string `json:"kind"`  // "turn_done"
    Summary      string `json:"summary"`
    IntentsTotal int    `json:"intentsTotal"`
    IntentsDone  int    `json:"intentsDone"`
}

type ErrorPayload struct {
    Kind           string `json:"kind"`  // "error"
    Code           string `json:"code"`
    Message        string `json:"message"`
    Recoverable    bool   `json:"recoverable"`
    FallbackAction string `json:"fallbackAction,omitempty"`  // retry / use_form / restart
}

// NextAction 提示前端下一步
type NextAction struct {
    Type      string `json:"type"`
    // await_confirm_intent: 等用户点"理解正确"（意图消息已展示）
    // await_clarify: 等用户回答 missing_info（切文本模式）
    // await_commit: 等用户点提交（form card 已展示）
    // await_select: 等用户选歧义候选（disambig card 已展示）
    // turn_complete: turn 完成
    // error: 错误终止
    MessageID string `json:"messageId,omitempty"`
    IntentID  string `json:"intentId,omitempty"`
}
```

## 4. 前端 TypeScript 契约

```typescript
interface SSEEvent<T extends PayloadBase> {
  event: EventType;
  turnId: string;
  stepId?: string;
  intentMeta?: IntentMeta;
  timestamp: number;
  messageId?: string;
  data: T;
}

interface PayloadBase {
  kind: string;
}

// 类型判别
type TextPayload = { kind: "text"; content: string; variant?: string };
type CardPayload = {
  kind: "card";
  cardType: "intent" | "form" | "query_result" | "disambig" | "fallback";
  status: string;  // 与 DB messages.status 字段对齐
  cardData: unknown;
};

function isCard(p: PayloadBase): p is CardPayload {
  return p.kind === "card";
}

// 事件分发
function handleEvent(evt: SSEEvent<PayloadBase>) {
  switch (evt.data.kind) {
    case "text":            /* render text */ break;
    case "card":
      switch ((evt.data as CardPayload).cardType) {
        case "intent":      /* render intent card + 切选项组 */ break;
        case "form":        /* render form card */ break;
        case "query_result": /* render query result */ break;
        case "disambig":    /* render disambig */ break;
        case "fallback":    /* render fallback */ break;
      }
      break;
    case "input_rewrite":   /* render in trace */ break;
    case "thinking":        /* append to trace */ break;
    case "tool_call":       /* append tool call */ break;
    case "tool_result":     /* update tool result */ break;
    case "step_started":    /* mark step started */ break;
    case "step_completed":  /* mark step completed */ break;
    case "step_phase_done": /* prepare for next user action */ break;
    case "turn_phase_done": /* end of phase */ break;
    case "turn_done":       /* turn complete */ break;
    case "error":           /* show error */ break;
  }
}

// SSE 流解析（fetch + ReadableStream）
async function postMessage(url: string, body: unknown): Promise<void> {
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Accept': 'text/event-stream' },
    body: JSON.stringify(body),
  });
  
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  if (!response.body) throw new Error('No body');
  
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    
    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    buffer = lines.pop() ?? '';
    
    for (const line of lines) {
      if (line.trim()) {
        try {
          const evt = JSON.parse(line) as SSEEvent<PayloadBase>;
          handleEvent(evt);
        } catch (err) {
          console.error('SSE parse error:', err, 'line:', line);
        }
      }
    }
  }
  
  if (buffer.trim()) {
    const evt = JSON.parse(buffer.trim()) as SSEEvent<PayloadBase>;
    handleEvent(evt);
  }
}

// 统一调用入口
async function sendUserMessage(sessionId: string, req: UserRequest): Promise<void> {
  await postMessage(`/api/v1/copilot/sessions/${sessionId}/messages`, req);
}

// 不同请求类型示例
await sendUserMessage(sid, { type: "free_text", content: "创建 P1 事项" });
await sendUserMessage(sid, { type: "confirm_intent", intentMessageId: "msg_002" });
await sendUserMessage(sid, { type: "answer_clarify", intentMessageId: "msg_010", answer: "认证模块" });
await sendUserMessage(sid, { type: "select_candidate", messageId: "msg_005", candidateBizKey: "MI-0023" });
await sendUserMessage(sid, { type: "commit_card", messageId: "msg_003" });
await sendUserMessage(sid, { type: "cancel", messageId: "msg_002" });
```

## 5. 完整事件流示例

### 5.1 单意图写操作

```
POST /api/v1/copilot/sessions/sess_xxx/messages
{ "type": "free_text", "content": "创建一个 P1 事项叫认证模块" }

HTTP 响应头：
Content-Type: text/event-stream; charset=utf-8
Cache-Control: no-cache, no-transform
X-Accel-Buffering: no

响应体（每行一个 JSON，\n 分隔）：
{"event":"input_rewrite","turnId":"turn_001","timestamp":1734900000000,"data":{"kind":"input_rewrite","original":"创建一个 P1 事项叫认证模块","rewritten":"创建一个 MainItem，title=认证模块，priority=P1","contextUsed":["team_context"],"ambiguityNotes":[]}}

{"event":"thinking","turnId":"turn_001","timestamp":1734900000200,"data":{"kind":"thinking","content":"用户要创建 MainItem，字段完整。无需澄清。"}}

{"event":"card_message","turnId":"turn_001","timestamp":1734900000500,"messageId":"msg_002","data":{"kind":"card","cardType":"intent","status":"awaiting_confirm","cardData":{"text":"好的，我帮你创建一个 P1 主事项「认证模块」。","intents":[{"id":"intent_1","label":"创建 MainItem","opType":"create","entityType":"main_item","fields":[{"name":"title","value":"认证模块","required":true},{"name":"priority","value":"P1","required":true}],"executor":"writer"}],"missingInfo":[]}}}

{"event":"turn_phase_done","turnId":"turn_001","timestamp":1734900000800,"data":{"kind":"turn_phase_done","phaseType":"planner","outcome":"success","nextAction":{"type":"await_confirm_intent","messageId":"msg_002"}}}
```

→ 流关闭。前端展示意图消息，切选项组。

### 5.2 用户确认意图

```
POST /api/v1/copilot/sessions/sess_xxx/messages
{ "type": "confirm_intent", "intentMessageId": "msg_002" }

响应（SSE 流）：
{"event":"step_started","turnId":"turn_001","timestamp":1734900009000,"stepId":"intent_1","intentMeta":{"id":"intent_1","label":"创建 MainItem","seq":1,"total":1},"data":{"kind":"step_started","intent":{"id":"intent_1","label":"创建 MainItem","seq":1,"total":1}}}

{"event":"thinking","turnId":"turn_001","stepId":"intent_1","timestamp":1734900009200,"data":{"kind":"thinking","content":"开始执行 writer。"}}

{"event":"tool_call","turnId":"turn_001","stepId":"intent_1","timestamp":1734900009500,"data":{"kind":"tool_call","callId":"call_001","toolName":"commit_create","arguments":{"entity_type":"main_item","fields":{"title":"认证模块","priority":"P1"}}}}

{"event":"tool_result","turnId":"turn_001","stepId":"intent_1","timestamp":1734900009800,"data":{"kind":"tool_result","callId":"call_001","status":"success","result":{"bizKey":"MI-0023"},"durationMs":300}}

{"event":"card_message","turnId":"turn_001","stepId":"intent_1","timestamp":1734900009900,"messageId":"msg_003","intentMeta":{"id":"intent_1","label":"创建 MainItem","seq":1,"total":1},"data":{"kind":"card","cardType":"form","status":"prefilled","cardData":{"opType":"create","entityType":"main_item","targetEntity":{"bizKey":"MI-0023","title":"认证模块","bizCode":"MI-0023"},"fields":[{"name":"title","value":"认证模块","required":true},{"name":"priority","value":"P1","required":true}]}}}

{"event":"step_phase_done","turnId":"turn_001","stepId":"intent_1","timestamp":1734900010000,"data":{"kind":"step_phase_done","intentId":"intent_1","outcome":"awaiting_user_commit","nextAction":{"type":"await_commit","messageId":"msg_003","intentId":"intent_1"}}}

{"event":"turn_phase_done","turnId":"turn_001","timestamp":1734900010100,"data":{"kind":"turn_phase_done","phaseType":"execute","outcome":"awaiting_user_action","nextAction":{"type":"await_commit","messageId":"msg_003"},"intentsTotal":1,"intentsDone":0}}
```

### 5.3 提交卡片（JSON 响应，无 SSE）

```
POST /api/v1/copilot/sessions/sess_xxx/messages
{ "type": "commit_card", "messageId": "msg_003" }

HTTP 响应头：
Content-Type: application/json

响应体（JSON）：
{
  "messageId": "msg_003",
  "newState": "submitted",
  "createdEntity": { "bizKey": "MI-0023", "title": "认证模块", "bizCode": "MI-0023" },
  "followupMessageId": "msg_004",
  "followupContent": "已为你创建 P1 事项「认证模块」（MI-0023）。"
}
```

## 6. 持久化时机与事件对应

| 事件 | 是否 persist message | messageId 在事件中 | 写 agent_call_logs |
|------|---------------------|-------------------|-------------------|
| input_rewrite | ❌ | — | ✅（input_rewrite_payload） |
| thinking（流式） | ❌（聚合到 trace） | — | — |
| tool_call / tool_result | ❌（聚合到 trace.actions） | — | — |
| trace freeze（隐式） | ✅ persist trace message | 下一个事件附带 | — |
| text_message | ✅ 立即 persist | ✅ | — |
| card_message (intent) | ✅ persist intent message | ✅ | — |
| card_message (form/query_result/disambig) | ✅ persist card message | ✅ | — |
| step_completed / turn_done | ❌ | — | ✅（更新该 agent 调用的 usage） |
| error | ❌ | — | ✅（status=failed） |

## 7. SSE 协议变体对比

| 维度 | 标准 SSE | 本设计（SSE MIME + 无前缀 JSON） |
|------|---------|------------------------------|
| Content-Type | text/event-stream | **text/event-stream** ✅ |
| 内容格式 | `event: X\ndata: Y\n\n` | `{...}\n`（一行一 JSON） |
| 浏览器 API | EventSource（仅 GET） | fetch + ReadableStream（支持 POST） |
| 自动重连 | ✅ | ❌（应用层处理） |
| 格式简洁 | ❌（前缀冗余） | ✅（纯 JSON） |
| 调试友好 | 多行格式 | 单行 grep / jq |
| 代理识别 | ✅ 标准流式 | ✅ 标准流式（MIME 一致） |

**结论**：本设计是 SSE 标准的" pragmatic 变体"——保留标准 MIME（让中间代理正确处理），简化内容格式（避免前缀冗余）。

## 8. 后端 SSE Writer 实现

```go
// internal/copilot/handler/stream.go

func (h *CopilotHandler) startSSEStream(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream; charset=utf-8")
    c.Header("Cache-Control", "no-cache, no-transform")
    c.Header("Connection", "keep-alive")
    c.Header("X-Accel-Buffering", "no")  // nginx 不缓冲
    c.Status(http.StatusOK)
}

func (h *CopilotHandler) streamEvents(c *gin.Context, eventCh <-chan sse.Event) {
    flusher, _ := c.Writer.(http.Flusher)
    writer := bufio.NewWriter(c.Writer)
    encoder := json.NewEncoder(writer)
    
    for evt := range eventCh {
        if err := encoder.Encode(evt); err != nil {
            break
        }
        writer.WriteByte('\n')  // 行分隔
        writer.Flush()
        if flusher != nil {
            flusher.Flush()
        }
    }
}
```

## 9. 错误事件示例

```
{"event":"error","turnId":"turn_001","timestamp":1734900000000,"data":{"kind":"error","code":"ERR_COPILOT_QUOTA_EXCEEDED","message":"今日 AI 调用已达上限（50 次）","recoverable":true,"fallbackAction":"use_form"}}
```

错误事件后流立即关闭，前端展示降级卡片。
