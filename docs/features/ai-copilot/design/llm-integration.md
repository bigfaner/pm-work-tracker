---
created: "2026-06-30"
updated: "2026-06-30"
parent: tech-design.md
---

# LLM Integration: 接入与执行流程

> 返回 [`tech-design.md`](./tech-design.md)
>
> 本文档聚焦"LLM 真正怎么接进来跑"。接口契约见 [`interfaces.md`](./interfaces.md)，职责划分与 system prompt 见 [`agent-architecture.md`](./agent-architecture.md)，请求级时序见 [`request-model.md`](./request-model.md)。本文档串联所有层，讲清端到端执行链。

## 1. 端到端架构

### 1.1 分层调用链

```
┌─────────────────────────────────────────────────────────────────┐
│ Frontend (React)                                                 │
│   fetch(POST /messages) → ReadableStream 解析 SSE 行            │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP POST + SSE response
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ Handler (internal/copilot/handler)                              │
│   PostMessage → 按 req.Type 分派                                │
│   创建 Turn + persist user msg → 启动 SSE 流 + goroutine        │
└────────────────────────────┬────────────────────────────────────┘
                             │ TurnContext + eventCh chan
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ Orchestrator (internal/copilot/orchestrator)                    │
│   HandleUserMessage / ExecuteFromIntent / ...                   │
│   按 turnCtx.Status 路由到 Planner 或 Executor                  │
└────────────────────────────┬────────────────────────────────────┘
                             │ AgentRunParams
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ Agent (internal/copilot/agent)                                  │
│   Planner / Writer / Reader / Updater / Mover                   │
│   运行 LLM ↔ Tool 循环（§2）                                    │
└─────────┬───────────────────────────────────┬───────────────────┘
          │                                   │
          ▼                                   ▼
┌─────────────────────┐         ┌──────────────────────────────┐
│ ContextBuilder      │         │ ToolRegistry                 │
│ 组装 system+history │         │ 执行 tool_call               │
│ +schema+user msg    │         │ (复用现有 entity service)    │
└─────────┬───────────┘         └──────────────────────────────┘
          │ PromptContext
          ▼
┌─────────────────────────────────────────────────────────────────┐
│ Provider (internal/copilot/provider)                            │
│   StreamChat(ctx, params) (<-chan ProviderEvent, error)         │
│   唯一接触外部 LLM API 的层                                     │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTPS POST + SSE response
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ External LLM API (GLM / DeepSeek / OpenAI)                      │
│   POST {baseURL}/chat/completions  (stream=true)                │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 职责一览

| 层 | 输入 | 输出 | 关键职责 |
|----|------|------|---------|
| Handler | HTTP req | SSE 流 | 创建 Turn、persist user msg、管理 HTTP 生命周期 |
| Orchestrator | TurnContext | sse.Event 流 | 按 turn.status 路由、串行执行多 intent、注入 step 事件 |
| Agent | AgentRunParams | sse.Event 流 | 运行 LLM ↔ Tool 循环、解析 LLM 输出 markers |
| ContextBuilder | BuildParams | PromptContext | 组装 system+history+schema、FIFO token 裁剪 |
| ToolRegistry | tool name + args | ToolResult | 路由到现有 entity service（创建/查询/更新/移动） |
| Provider | ProviderParams | `<-chan ProviderEvent` | HTTP 调用 LLM、解析 SSE、累积 tool_call 分片 |

### 1.3 关键设计原则

1. **Provider 是唯一外部边界**——Agent 不直接发 HTTP，所有 LLM 调用经 Provider 抽象
2. **Agent 拥有循环**——LLM ↔ Tool 循环由 Agent 控制，Provider 只负责单次调用
3. **流式穿透**——Provider 解析 SSE → ProviderEvent → Agent 转 sse.Event → Handler 写 HTTP，全程 channel 串联，无缓冲落地
4. **Tool 错误不中断**——tool 执行失败回灌给 LLM 作为 `role=tool, status=error`，LLM 可自我修正

## 2. Agent 执行循环（核心）

### 2.1 循环流程

```
Agent.StreamRun(ctx, params)
    │
    ▼
ctxBuilder.Build()  ← 仅一次（组装 system + history + user msg）
    │
    ▼ PromptContext.Messages = [system, history..., user]
┌─── loop (max N iterations) ───────────────────────────────────┐
│                                                                │
│  1. provider.StreamChat(messages) → <-chan ProviderEvent      │
│                                                                │
│  2. 消费 stream：                                              │
│     - Delta → 累积 content，同时 emit thinking 事件            │
│     - ToolCall（已累积完整）→ 暂存                              │
│     - Done → 拿到 usage，退出内层 for                          │
│     - Error → emit error 事件，return                          │
│                                                                │
│  3. 判定：                                                     │
│     ┌─ 无 tool_calls → 自然结束（LLM 未调 emission 工具）     │
│     │   - 先尝试按本 Agent output schema 解析累积 content     │
│     │     成功 → 当作 emission 处理（emit 对应 card 事件 +    │
│     │            persist），写 log(warning=no_emission_tool,  │
│     │            recovered=parse_content)                     │
│     │     失败 → 累积 content 作为 text_message 兜底 emit,    │
│     │            写 log(warning=no_emission_tool)             │
│     │   - return（结束 StreamRun）                             │
│     │                                                          │
│     └─ 有 tool_calls → 逐个执行                                │
│         - append assistant msg (含 tool_calls) 到 messages    │
│         - for each tool_call:                                 │
│             emit tool_call 事件                                │
│             result = toolRegistry.Execute(name, args, outCh)  │
│             按 result.Status 分流：                            │
│               success / error:                                │
│                 emit tool_result 事件                          │
│                 append tool msg (role=tool)                   │
│                 → 继续下一个 tool                              │
│               terminal（emission 工具）:                      │
│                 emission 工具已直接写 SSE 事件到 outCh        │
│                 不 emit tool_result、不 append                │
│                 → 跳出循环，return（结束 StreamRun）          │
│         - 若所有 tool 均 success/error → continue loop        │
│           （带着 tool 结果再次调 Provider）                    │
│                                                                │
└────────────────────────────────────────────────────────────────┘
超过 max iterations → emit error 事件，return
```

**关键差异（与文本 marker 方式对比）**：
- LLM **不再**输出 `<message type="X">{json}</message>` 文本标记
- 结构化输出**只**通过 emission 工具的 args 传递，强类型、无文本解析
- 循环终止信号从"LLM 不再调 tool"变为"调到 emission tool"

### 2.2 实现骨架

```go
// internal/copilot/agent/base_agent.go

const maxAgentIterations = 8  // 单次 Agent 调用最多 8 轮 LLM↔Tool

func (a *baseAgent) StreamRun(ctx context.Context, p AgentRunParams) (<-chan sse.Event, error) {
    // 1. 组装上下文（仅一次）
    promptCtx, err := a.ctxBuilder.Build(ctx, prompt.BuildParams{
        Role:       a.role,
        UserMsg:    p.UserMsg,
        StepParams: p.StepParams,
        Env:        p.Env,
        History:    p.History,
        DraftState: p.DraftState,
    })
    if err != nil {
        return nil, fmt.Errorf("build context: %w", err)
    }

    outCh := make(chan sse.Event, 64)
    go func() {
        defer close(outCh)
        a.runLoop(ctx, promptCtx, p, outCh)
    }()
    return outCh, nil
}

func (a *baseAgent) runLoop(
    ctx context.Context, promptCtx prompt.PromptContext,
    p AgentRunParams, outCh chan<- sse.Event,
) {
    messages := promptCtx.Messages  // [system, history..., user]

    // Emission 工具通过此参数拿到 outCh 写事件 + 暂存回调。
    // p.Persist / p.OnAgentCall 由 RequestState 投影器（ForPlanner/ForExecutor）注入闭包，
    // 把产出【累积进 RequestState】（Persist → rs.PendingMessages 返回预生成 bizKey；
    // logAgentCall → p.OnAgentCall → rs.Calls.Append）。Agent 内零 DB 写——
    // 持久化在请求结束 StateApplier.Flush 单事务完成（见 state-model.md §3.3）。
    toolParams := tool.ToolExecParams{
        OutCh:   outCh,
        TurnID:  p.TurnID,
        StepID:  p.StepID,
        Persist: p.Persist,   // stage Message → rs.PendingMessages，返回 bizKey（不写库）
    }

    for iter := 0; iter < maxAgentIterations; iter++ {
        // 2. 调 Provider
        stream, err := a.provider.StreamChat(ctx, provider.ProviderParams{
            Model:        a.model,
            SystemPrompt: promptCtx.SystemPrompt,
            Messages:     messages,
            Tools:        promptCtx.Tools,
            Temperature:  a.temperature,
            Stream:       true,
        })
        if err != nil {
            outCh <- sse.ErrorEvent(p.TurnID, err)
            a.logAgentCall(p, nil, "failed", err.Error())
            return
        }

        // 3. 消费 stream，累积 content 与 tool_calls
        var contentBuf strings.Builder
        var toolCalls []*provider.ProviderToolCall
        var usage *provider.ProviderUsage

        for ev := range stream {
            switch ev.Type {
            case provider.EventDelta:
                contentBuf.WriteString(ev.Delta)
                outCh <- sse.ThinkingEvent(p.TurnID, p.StepID, ev.Delta)

            case provider.EventToolCall:
                toolCalls = append(toolCalls, ev.ToolCall)

            case provider.EventDone:
                usage = ev.Usage

            case provider.EventError:
                outCh <- sse.ErrorEvent(p.TurnID, ev.Error)
                a.logAgentCall(p, nil, "failed", ev.Error.Error())
                return
            }
        }

        // 4. 无 tool_call = 自然结束（LLM 未调 emission 工具）
        //    先尝试按本 Agent 的 output schema 解析累积 content 作为兜底产出；
        //    解析成功则当作 emission 处理，失败才降级为纯文本（见 §2.4）
        if len(toolCalls) == 0 {
            if recovered := a.tryRecoverOutput(ctx, contentBuf.String(), p, outCh); recovered {
                a.logAgentCall(p, usage, "success", "no_emission_tool,recovered=parse_content")
            } else {
                if contentBuf.Len() > 0 {
                    outCh <- sse.TextMessageEvent(p.TurnID, p.StepID, contentBuf.String())
                }
                a.logAgentCall(p, usage, "success", "no_emission_tool")
            }
            return
        }

        // 5. 有 tool_call = 逐个执行
        messages = append(messages, provider.ProviderMsg{
            Role:      "assistant",
            ToolCalls: toolCalls,
        })

        var terminal bool
        for _, tc := range toolCalls {
            // Emission 工具不 emit tool_call/tool_result——避免与它自己写的专用事件重复
            t := a.tools.Lookup(tc.Name)
            isEmission := t != nil && t.Kind() == tool.ToolKindEmission
            if !isEmission {
                outCh <- sse.ToolCallEvent(p.TurnID, p.StepID, tc)
            }

            result, err := a.tools.Execute(ctx, tc.Name, tc.Arguments, toolParams)
            if err != nil {
                result = tool.ToolResult{Status: tool.ToolStatusError, Error: err.Error()}
            }

            // emission 工具已通过 toolParams.OutCh 写完专用事件
            if result.Status == tool.ToolStatusTerminal {
                terminal = true
                break // emission 后续 tool_call 不再执行（不该出现）
            }

            // Read / Action 工具：emit tool_result + append 给下一轮 LLM
            if !isEmission {
                outCh <- sse.ToolResultEvent(p.TurnID, p.StepID, tc.ID, result)
            }
            messages = append(messages, provider.ProviderMsg{
                Role:       "tool",
                Content:    serializeToolResult(result),
                ToolCallID: tc.ID,
            })
        }

        if terminal {
            a.logAgentCall(p, usage, "success", "")
            return
        }
        // 否则 continue loop，messages 已追加 assistant(tool_calls) + tool(results)
    }

    outCh <- sse.ErrorEvent(p.TurnID, errors.New("agent exceeded max iterations"))
    a.logAgentCall(p, nil, "failed", "max iterations exceeded")
}
```

**`logAgentCall` 实现**——构建 `AgentCallLog` 并通过 `p.OnAgentCall` **累积进 `rs.Calls`**（不直接写库，Flush 时 Drain 落 `agent_call_logs`）：

```go
// logAgentCall —— Agent 调用元数据累加器（不写库）
func (a *baseAgent) logAgentCall(p AgentRunParams, usage *provider.ProviderUsage, status, warning string) {
    log := AgentCallLog{
        SessionID: p.SessionID, TurnID: p.TurnID, StepID: p.StepID,
        AgentRole: string(a.role), Provider: a.providerName(), Model: a.model,
        Status: status, Warning: warning,
    }
    if usage != nil {
        log.InputTokens, log.OutputTokens = usage.InputTokens, usage.OutputTokens
    }
    p.OnAgentCall(log) // → rs.Calls.Append（见 state-model.md §2.4）
}
```

### 2.3 关键不变量

1. **Provider 调用次数 = 循环迭代次数**——每次循环恰好一次 `StreamChat`
2. **结构化输出首选 emission 工具**——system prompt 要求 LLM 不输出文本 marker / 裸 JSON，结构化数据走工具 args；但 Agent 对"LLM 未遵守、自然结束无 tool call"的情况有兜底解析（见不变量 8 与 §2.4）
3. **两条结束路径**——(a) emission 工具返回 `Status=terminal` 立即结束 StreamRun（首选，除 `submit_rewrite` 它是中间步骤不终止）；(b) 一轮无 tool_call 自然结束，按 output schema 兜底解析
4. **Read / Action 工具先 emit 后执行**——前端能看到"调了什么"再看到"结果是什么"
5. **Tool 错误不 break**——以 `role=tool, status=error` 回灌给 LLM，让 LLM 决定重试或放弃
6. **迭代上限保护**——`maxAgentIterations=8`，超过即 fail（防止 LLM 无限调 tool 死循环）
7. **ContextBuilder 只跑一次**——历史裁剪在循环外完成；循环内只在 messages 末尾追加 assistant + tool 消息
8. **自然结束有兜底**——LLM 一轮未调任何工具时，`tryRecoverOutput` 按本 Agent output schema（Planner→IntentPayload，Writer/Mover→FormCardData，Reader→QueryResultCardData）解析累积 content；解析成功复用 emission 的 emit+stage 路径，失败才 text 兜底。此路径仅作 safety net，system prompt 仍优先要求走 emission 工具
9. **Agent 循环不写库**——所有产出（emission 消息、trace、agent 调用元数据）经 `p.Persist` / `logAgentCall`（→`p.OnAgentCall`）**累积进 RequestState**，Agent 内零 DB 写；持久化集中在请求结束 `StateApplier.Flush` 单事务（见 state-model.md §1、§3.3）

### 2.4 各 Agent 的 emission 调用顺序

每个 Agent 必有 emission 工具作为完成方式。**无写工具**——所有 emit_form_card 仅生成预填表单（targetEntity.bizKey 留空），真实 DB 写由 `commit_card` Handler 触发：

| Agent | 必调 emission 工具（按典型顺序） | 终止信号 | 备注 |
|-------|------------------------------|---------|------|
| Planner | 1. `submit_rewrite`（不终止） → 2. `submit_intent` | `submit_intent` | submit_rewrite 是第 1 步，之后还要 thinking + submit_intent |
| Writer | （可选 Read: `fuzzy_match_*`） → `emit_form_card` 或 `emit_disambig` | `emit_form_card` / `emit_disambig` | 多候选时先 emit_disambig 终止；form card 仅含预填字段，bizKey 留空 |
| Reader | `query_entities` → `emit_query_result` | `emit_query_result` | — |
| Updater | （可选 Read: `validate_transition`） → `emit_form_card` 或 `emit_disambig` | `emit_form_card` / `emit_disambig` | 预校验失败时把 errors 填入 cardData，但仍 emit_form_card 终止 |
| Mover | （可选 Read: `validate_source_target`） → `emit_form_card` 或 `emit_disambig` | `emit_form_card` / `emit_disambig` | 同上 |

**特殊：`submit_rewrite` 是唯一不终止的 emission 工具**。返回 `Status=success`，Agent 继续 LLM 循环。这样 Planner 能在单次 StreamRun 内串行完成 input_rewrite + thinking + submit_intent。

**自然结束兜底（safety net）**：若 LLM 一轮结束未调任何 emission 工具，Agent 并非直接降级为纯文本，而是先调 `tryRecoverOutput` 按 output schema 解析累积 content：
- **解析成功** → 复用对应 emission 工具的 emit + stage 路径（等同 LLM 调了该工具，消息进 `rs.PendingMessages`），`rs.Calls` 记 `warning=no_emission_tool, recovered=parse_content`。用户仍看到结构化卡片。
- **解析失败** → 累积 content 作为 `text_message` 兜底 emit + stage，`rs.Calls` 记 `warning=no_emission_tool`。此时用户看到纯文本回复。

各 Agent 的 output schema 对照：Planner→`IntentPayload`、Writer/Mover→`FormCardData`（或 `DisambigCardData`）、Reader→`QueryResultCardData`。此兜底仅针对 LLM 偶发不守 prompt 的情况；system prompt 仍明确要求结构化输出走 emission 工具，兜底不作为常规路径。

## 3. GLM Provider 实现

### 3.1 GLM API 契约（OpenAI 兼容）

- **Endpoint**: `POST {baseURL}/chat/completions`
- **Auth**: `Authorization: Bearer <api_key>`
- **Request body**（标准 OpenAI 格式）:

```json
{
  "model": "glm-4-plus",
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."}
  ],
  "tools": [
    {"type": "function", "function": {"name": "commit_create", "description": "...", "parameters": {...}}}
  ],
  "tool_choice": "auto",
  "temperature": 0.3,
  "stream": true,
  "stream_options": {"include_usage": true}
}
```

- **Response**: SSE，每行 `data: {chunk_json}\n\n`，结束 `data: [DONE]\n\n`
- **Chunk 结构**:

```json
{
  "id": "chatcmpl-xxx",
  "choices": [{
    "index": 0,
    "delta": {
      "content": "文本增量",
      "tool_calls": [
        {"index": 0, "id": "call_001", "type": "function",
         "function": {"name": "commit_create", "arguments": "{\"entity"}}
      ]
    },
    "finish_reason": null
  }],
  "usage": null
}
```

**关键点：tool_call 分片流式到达**——`function.arguments` 是 JSON 字符串的分片（如 `{"entity` → `_type":"main` → `_item",...}`），必须按 `index` 累积拼装。（注：示例中 `commit_create` 仅为 GLM SSE 分片格式演示；实际 Writer Executor 只持有 `emit_form_card` 等 Read/Emission 工具，见 [`agent-architecture.md`](./agent-architecture.md) §2.2。）

### 3.2 StreamChat 实现

```go
// internal/copilot/provider/glm_provider.go

func (p *GLMProvider) StreamChat(
    ctx context.Context, params ProviderParams,
) (<-chan ProviderEvent, error) {
    reqBody, err := p.buildRequest(params)
    if err != nil {
        return nil, fmt.Errorf("build request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx,
        http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(reqBody))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+p.apiKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "text/event-stream")

    resp, err := p.client.Do(req)
    if err != nil {
        return nil, classifyHTTPError(err)  // 区分 timeout / network / canceled
    }
    if resp.StatusCode != http.StatusOK {
        defer resp.Body.Close()
        body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
        return nil, fmt.Errorf("glm api %d: %s", resp.StatusCode, body)
    }

    ch := make(chan ProviderEvent, 64)
    go p.parseSSEStream(ctx, resp.Body, ch)
    return ch, nil
}
```

### 3.3 SSE 解析与 tool_call 累积

```go
// internal/copilot/provider/glm_provider.go

func (p *GLMProvider) parseSSEStream(
    ctx context.Context, body io.ReadCloser, ch chan<- ProviderEvent,
) {
    defer close(ch)
    defer body.Close()

    scanner := bufio.NewScanner(body)
    scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 单行最大 1MB

    // tool_call 按 index 累积（GLM 可能并行多个 tool_call）
    toolAcc := map[int]*ProviderToolCall{}
    var usage *ProviderUsage

    for scanner.Scan() {
        // ctx 取消立即终止
        if err := ctx.Err(); err != nil {
            ch <- ProviderEvent{Type: EventError, Error: err}
            return
        }

        line := scanner.Text()
        if !strings.HasPrefix(line, "data: ") {
            continue
        }
        data := strings.TrimPrefix(line, "data: ")

        if data == "[DONE]" {
            ch <- ProviderEvent{Type: EventDone, Usage: usage}
            return
        }

        var chunk glmChunk
        if err := json.Unmarshal([]byte(data), &chunk); err != nil {
            continue // 跳过畸形行（不中断流）
        }

        if chunk.Usage != nil {
            usage = &ProviderUsage{
                InputTokens:  chunk.Usage.PromptTokens,
                OutputTokens: chunk.Usage.CompletionTokens,
            }
        }
        if len(chunk.Choices) == 0 {
            continue
        }
        delta := chunk.Choices[0].Delta

        // 文本增量
        if delta.Content != "" {
            ch <- ProviderEvent{Type: EventDelta, Delta: delta.Content}
        }

        // tool_call 分片累积
        for _, tc := range delta.ToolCalls {
            acc, ok := toolAcc[tc.Index]
            if !ok {
                acc = &ProviderToolCall{
                    ID:   tc.ID,
                    Name: tc.Function.Name,
                }
                toolAcc[tc.Index] = acc
            }
            acc.Arguments += tc.Function.Arguments // 流式拼接 JSON
        }
    }

    if err := scanner.Err(); err != nil {
        ch <- ProviderEvent{Type: EventError, Error: err}
        return
    }

    // 流自然结束但没收到 [DONE]——按完成处理，但 tool_call 仍要 emit
    // （边界：某些代理可能截断 [DONE]）
    emitAccumulatedToolCalls(ch, toolAcc)
    ch <- ProviderEvent{Type: EventDone, Usage: usage}
}

// emitAccumulatedToolCalls 在 finish_reason="tool_calls" 后调用
// 实际实现：在 scanner 循环里检测 finish_reason，先 emit tool_calls，
// 再继续读直到 [DONE] 拿到 usage（GLM 的 usage 在最后单独一帧）
```

### 3.4 Token 计数

GLM 没有公开 tokenizer，采用字符粗估：

```go
func (p *GLMProvider) CountTokens(text string) int {
    // 经验估计：1 token ≈ 1.5 中文字符 / 4 英文字符
    cn := countCJK(text)
    en := len(text) - cn
    return cn*2/3 + en/4
}
```

**用途**：仅用于 ContextBuilder 的预算裁剪（不需要精确，±20% 可接受）。实际计费以 GLM 返回的 `usage` 为准（写入 `agent_call_logs`）。

### 3.5 错误分类

| 触发 | 分类 | 处理 |
|------|------|------|
| `context.Canceled` | 客户端断开 | 静默终止，不写 error 事件 |
| `context.DeadlineExceeded` | 请求超时 | emit error 事件，标记 turn failed |
| 网络错误（connection refused/DNS） | 暂时性 | 由 Provider 内部重试 §6.2 |
| HTTP 429 | 限流 | 按 `Retry-After` 重试 |
| HTTP 5xx | 服务端错误 | 重试 1 次 |
| HTTP 4xx（除 429） | 配置/请求错误 | 不重试，立即失败 |
| SSE 中途畸形 | 容错 | 跳过当前行继续 |

## 4. Provider 生命周期 & 配置注入

### 4.1 应用启动时初始化

Provider 在 app 启动时由 Factory 创建一次，全应用共享一个实例。

```go
// cmd/server/main.go (或 internal/app/wire.go)

func setupCopilotDeps(cfg Config, db *gorm.DB) (*orchestrator.Orchestrator, error) {
    // 1. 创建 Provider（单例）
    prov, err := provider.NewProvider(cfg.Copilot.Provider)
    if err != nil {
        return nil, fmt.Errorf("init llm provider: %w", err)
    }

    // 2. 创建共享依赖
    schemaLoader := prompt.NewSchemaLoader(/* entity schemas */)
    ctxBuilder := prompt.NewContextBuilder(prov, schemaLoader, cfg.Copilot.TokenBudget)
    toolRegistry := buildToolRegistry(db, /* existing services */)

    // 3. 创建 Agent Registry（所有 Agent 共享同一 Provider + ctxBuilder）
    registry := agent.NewRegistry(prov, ctxBuilder, toolRegistry, cfg.Copilot.Agent)

    // 4. 创建 Orchestrator
    msgRepo := repository.NewMessageRepository(db)
    turnRepo := repository.NewTurnRepository(db)
    dispatcher := service.NewDispatcher(/* existing entity services */)

    return orchestrator.New(registry, ctxBuilder, msgRepo, turnRepo, dispatcher, prov), nil
}
```

### 4.2 共享策略

| 资源 | 共享范围 | 并发安全 | 说明 |
|------|---------|---------|------|
| Provider 实例 | 全应用 | ✓ | 无状态，仅持 cfg + `*http.Client` |
| `*http.Client` | 全应用 | ✓ | goroutine-safe，连接池复用 |
| ContextBuilder | 全应用 | ✓ | 无状态（SchemaLoader 内部只读） |
| Agent 实例 | 全应用 | ✓ | 无状态，每次调用传 AgentRunParams |
| ToolRegistry | 全应用 | ✓ | 工具实现无状态或自带锁 |

**结论**：N 个并发请求 = N 个 goroutine 共享同一组 Provider/Agent。无 per-request 实例化开销。

### 4.3 配置注入

```yaml
# backend/config.yaml
copilot:
  provider:
    type: glm                          # glm / deepseek / openai / mock
    api_key: ${GLM_API_KEY}            # 从 env 注入，不硬编码
    base_url: https://open.bigmodel.cn/api/paas/v4
    model: glm-4-plus
    timeout: 30s
  token_budget: 12000                  # ContextBuilder 裁剪预算
  agent:
    max_iterations: 8
    temperature: 0.3
```

**关键约束**：
- `api_key` 永远从环境变量注入（`${VAR}` 语法），**不入 git、不入日志**
- `base_url` 可配置（方便切 GLM 私有部署 / 测试代理）
- 切换 Provider 只改 `type`，不动代码（Factory 模式的收益）

### 4.4 Agent 与 Provider 的关系

```
                    ┌──────────────────────┐
                    │  Provider (singleton)│
                    │  - glm-4-plus        │
                    └──────────┬───────────┘
                               │ 共享
            ┌──────────────────┼──────────────────┐
            │                  │                  │
       ┌────▼─────┐      ┌────▼─────┐       ┌────▼─────┐
       │ Planner  │      │  Writer  │  ...  │  Mover   │
       │ Agent    │      │  Agent   │       │  Agent   │
       └──────────┘      └──────────┘       └──────────┘
```

所有 Agent 调用同一个 Provider，区别仅在：
- **System prompt**（Agent 子类提供）
- **Tools**（Agent 子类提供，Planner 只读 / Writer 写 / ...）
- **Temperature**（Planner 0.5 偏发散，Executor 0.1 偏严谨）

模型选择（`model: glm-4-plus`）由 Provider 持有——所有 Agent 默认用同一模型。如未来需要"Planner 用更大模型、Executor 用更小模型"，扩展为 per-Agent model 配置即可（YAGNI，暂不实现）。

## 5. 上下文组装流程

### 5.1 ContextBuilder 在循环中的位置

```
┌─ Agent.StreamRun ──────────────────────────────────────┐
│                                                         │
│   ctxBuilder.Build()  ← 循环外，仅一次                  │
│         │                                               │
│         ▼                                               │
│   messages = [system, history..., user]                 │
│                                                         │
│   ┌─ loop ───────────────────────────────────────────┐ │
│   │   provider.StreamChat(messages)                  │ │
│   │   ... 消费 stream ...                             │ │
│   │   if tool_calls:                                 │ │
│   │       messages.append(assistant + tool_calls)    │ │
│   │       messages.append(tool results)              │ │
│   │       continue  ← 循环内不再调 ctxBuilder         │ │
│   └──────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

**关键**：ContextBuilder 不参与循环内迭代。多轮 tool 调用的消息累积由 Agent 自己维护。

### 5.2 组装内容（一次性）

```go
func (b *contextBuilder) Build(ctx context.Context, p BuildParams) (PromptContext, error) {
    // 1. 固定部分（按 Role 选模板）
    sysPrompt := b.systemPromptFor(p.Role)
    env := SerializeEnv(p.Env)  // 渲染 user/team/page/time 进 prompt（定义见 state-model.md §3.4）
    tools := b.toolsFor(p.Role)
    schema := b.maybeLoadSchema(p)  // 仅 Executor 加载 entity schema

    fixed := sysPrompt + env + schema
    fixedTokens := b.provider.CountTokens(fixed + serializeTools(tools))

    // 2. 历史裁剪预算
    outputBudget := 1500
    historyBudget := b.budget - fixedTokens - outputBudget
    if historyBudget <= 0 {
        return PromptContext{}, errors.New("system prompt exceeds token budget")
    }
    history := b.cropHistory(p.History, historyBudget)  // group-aware 裁剪（§5.4），FIFO 兜底

    // 3. 组装 messages
    var msgs []ProviderMsg
    msgs = append(msgs, ProviderMsg{Role: "system", Content: fixed})
    for _, m := range history {
        msgs = append(msgs, ToProviderMsg(m))  // MessageSnapshot → ProviderMsg（定义见 state-model.md §3.4）
    }
    if p.UserMsg != "" {
        msgs = append(msgs, ProviderMsg{Role: "user", Content: p.UserMsg})
    }

    return PromptContext{
        SystemPrompt: sysPrompt,
        Messages:     msgs,
        Tools:        tools,
        TotalTokens:  fixedTokens + b.provider.CountTokens(p.UserMsg),
        Schema:       schema,
        History:      history,
    }, nil
}
```

### 5.3 Schema 按需注入

| Role | 加载 schema？ | 来源 |
|------|--------------|------|
| Planner | ✗（不自动注入） | ContextBuilder 不为 Planner 预注入 schema；但 Planner 抽取 `intents[].fields` / `missingInfo` 前**必须**显式调 `query_team_schema(entity_type)` 工具获取字段定义（system prompt 硬规则，见 [agent-architecture.md](./agent-architecture.md) §3.2），否则字段名与下游 Executor 不一致 |
| Writer / Updater | ✓ | `stepParams.entity_type` → schema_loader.Load |
| Reader | ✓ | 同上（filter 字段需要对齐 schema） |
| Mover | ✗ | SubItem.move 是固定结构，schema 内嵌在 prompt |

**目的**：避免每次都把 6 个实体的 schema 全塞进 prompt（省 ~3000 tokens）。Planner 不自动注入是因为它按需只为识别到的 entityType 查询、而非全量；但"不自动注入"不等于"可以不查"——见上表 Planner 行。

### 5.4 历史裁剪策略

```go
// cropHistory —— group-aware 裁剪（对标 LangGraph add_messages 的 by-id 整体性：
// 不拆散同一 intent 的消息组）。先丢最旧的完整组，保留当前 turn 的部分组；
// 单组过大时退回组内 FIFO 兜底，保证总能收敛到预算内。
func (b *contextBuilder) cropHistory(history []MessageSnapshot, budget int) []MessageSnapshot {
    if b.provider.CountTokens(serialize(history)) <= budget {
        return history
    }
    // 1. 按 IntentID 分组（IntentID=nil 的消息各自独立成组），保留首次出现顺序
    groups := groupByIntent(history) // [][]MessageSnapshot

    // 2. 从最旧的完整组开始丢，直到达标或只剩最新 1-2 组（保护当前 turn 上下文）
    for len(groups) > 2 && b.provider.CountTokens(serialize(flatten(groups))) > budget {
        groups = groups[1:]
    }

    // 3. 仍超预算 → 组内 FIFO 兜底（逐条丢最早，极端情况收敛到空）
    flat := flatten(groups)
    for len(flat) > 0 && b.provider.CountTokens(serialize(flat)) > budget {
        flat = flat[1:]
    }
    return flat
}
```

**不变量**：
- 当前 turn 的 user msg 永远保留（不在 history 内，由 Build 单独 append）
- 默认丢弃粒度是**完整 intent 组**（组内语义不拆散）；兜底阶段才逐条丢
- 当前 turn 的最新 1-2 组受保护（最后才丢）
- 极端情况：history 全被丢弃，仅剩 system + user msg（短上下文也能跑）

## 6. 错误处理 & 超时 & 重试

### 6.1 错误分级与传播

```
GLM API 错误          Provider 错误分类        Agent 处理            用户感知
─────────────────────────────────────────────────────────────────────────────
4xx (除 429)    →    ConfigError         →   立即失败        →   error 事件 + turn failed
429              →    RateLimitError      →   按 Retry-After  →   最多重试 2 次
5xx              →    ServerError        →   指数退避重试    →   最多重试 2 次
network          →    TransientError     →   指数退避重试    →   最多重试 2 次
timeout          →    TimeoutError       →   不重试          →   error 事件
context cancel   →    Canceled           →   静默退出        →   流关闭（用户主动断开）
tool 执行失败    →    —                  →   回灌给 LLM      →   LLM 可能自我修正
```

### 6.2 Provider 重试（仅 StreamChat 调用前）

```go
// internal/copilot/provider/retry.go

func (p *GLMProvider) StreamChat(ctx context.Context, params ProviderParams) (<-chan ProviderEvent, error) {
    var lastErr error
    backoff := time.Second
    for attempt := 0; attempt < 3; attempt++ {
        stream, err := p.doStreamChat(ctx, params)
        if err == nil {
            return stream, nil
        }
        lastErr = err

        if !isRetryable(err) {
            return nil, err  // 4xx 直接放弃
        }

        // 429: 优先尊重 Retry-After
        if raErr, ok := err.(*RateLimitError); ok {
            backoff = raErr.RetryAfter
        }

        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(backoff):
            backoff *= 2 // 指数退避
        }
    }
    return nil, fmt.Errorf("after 3 retries: %w", lastErr)
}
```

**重试边界**：
- 仅在建立连接阶段重试（`doStreamChat` 返回 error 时）
- 流建立后中途出错**不重试**——已经 emit 过部分事件，重试会导致重复输出
- 中途出错由 Agent 转 error 事件处理

### 6.3 Tool 错误回收

```go
// tool 执行失败 → 包装成 tool_result 回灌给 LLM

result, err := a.tools.Execute(ctx, tc.Name, tc.Arguments, toolParams)
if err != nil {
    result = tool.ToolResult{
        Status: tool.ToolStatusError,
        Error:  err.Error(),
    }
    // 不 break，继续 append tool message
}

// LLM 在下一轮看到 role=tool, content={status:error,...}
// 可能：① 重试（修正 args）  ② 改用其他工具  ③ 放弃并 emit 失败答复
```

**为什么 tool 错误不直接失败**：业务实践中大部分 tool 错误是参数问题（如 `assignee` 拼写错），LLM 看到 `member not found` 会主动调 `fuzzy_match_member` 重试。

### 6.4 超时配置

| 层 | 超时 | 配置 |
|----|------|------|
| HTTP 请求（Provider→GLM） | 30s | `cfg.copilot.provider.timeout` |
| 单次 Agent 迭代（一次 LLM 调用） | 30s | 同上（共用） |
| 单次 Agent StreamRun 总时长 | 60s | **显式 wall-clock 上限**（不再是隐式 8×30s） |
| Turn 总执行 | 120s | 由客户端断开 / wall-clock 双重限制 |

**为什么 StreamRun 改为显式 60s**：原"8 iter × 30s = 4 分钟"是隐式上限，单个用户可钉住 goroutine + GLM 上游连接长达 4 分钟。20 用户 × 4 min = goroutine 爆炸风险。改为 60s wall-clock：足够覆盖典型 turn（Planner 2 调 + Executor 1–2 调，每调 ~5s），同时把资源占用上限降到 1/4。

**客户端断开处理**：`c.Request.Context()` 取消时，所有下游 channel 关闭、goroutine 退出。已 persist 的 trace 和意图消息保留。

### 6.5 并发限制（防 goroutine / 连接耗尽）

单个 StreamRun 可钉住一个 goroutine + 一条 GLM 上游 HTTP 连接最长 60s。必须显式限制并发，否则恶意/异常用户可耗尽资源。

| 限制 | 配置键 | 默认 | 实现位置 | 说明 |
|------|-------|------|---------|------|
| 单用户并发 SSE 流 | `cfg.copilot.concurrency.max_streams_per_user` | 1 | Handler middleware `UserStreamGuard` | 同一 user 同时只允许 1 个活跃 SSE 流；第 2 个返回 HTTP 429 + `ERR_COPILOT_USER_STREAM_BUSY` |
| 全局 GLM 出站并发 | `cfg.copilot.concurrency.max_outbound_llm` | 10 | Provider 内 `chan struct{}` semaphore（capacity=10） | 全应用共享；超过则 StreamChat 阻塞排队（带 5s 超时 → 503） |
| 全局 SSE 流并发 | `cfg.copilot.concurrency.max_global_streams` | 50 | Handler middleware `GlobalStreamGuard` | 全应用同时活跃 SSE 流上限；超限返回 503 |

```go
// internal/copilot/provider/glm_provider.go

type GLMProvider struct {
    // ... 其他字段
    outboundSem chan struct{}  // capacity = max_outbound_llm（默认 10）
}

func (p *GLMProvider) StreamChat(ctx context.Context, params ProviderParams) (<-chan ProviderEvent, error) {
    // 获取信号量（5s 超时）
    select {
    case p.outboundSem <- struct{}{}:
        defer func() { <-p.outboundSem }()
    case <-time.After(5 * time.Second):
        return nil, ErrLLMOutboundBusy  // → HTTP 503
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    // ... 原有逻辑
}
```

```go
// internal/copilot/handler/middleware.go

// 注意：中间件用 ShouldBindBodyWith（非 ShouldBindJSON）绑定——后者会消费 body，
// 下游 handler（PostMessage / TurnInFlightGuard）再次绑定时 EOF。ShouldBindBodyWith
// 把 body bytes 缓存到 context，后续无论 ShouldBindJSON 还是 ShouldBindBodyWith 都能重读。
// 测试：TestUserStreamGuard_BodyReReadable 验证中间件绑定后 handler 仍能拿到完整 body。
func (h *CopilotHandler) UserStreamGuard(c *gin.Context) {
    var req MessageRequest
    if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil { c.Next(); return }
    if !triggersSSEStream(req.Type) { c.Next(); return }

    userBizKey := middleware.GetUserBizKey(c)
    if !h.streamLimiter.TryAcquire(userBizKey) {
        c.JSON(http.StatusTooManyRequests, gin.H{
            "code": "ERR_COPILOT_USER_STREAM_BUSY",
            "message": "已有 AI 请求处理中，请等待完成或取消后再试",
            "recoverable": true,
        })
        c.Abort()
        return
    }
    // 流关闭时 Release（在 streamEvents defer 内）
    c.Set("releaseStreamSlot", func() { h.streamLimiter.Release(userBizKey) })
    c.Next()
}
```

**StreamLimiter 实现**：`map[userBizKey]int`（计数）+ `sync.Mutex`；TryAcquire 在已有计数 ≥ 1 时返回 false。流结束（正常 / 异常 / ctx 取消）时 Release。

**容量计算依据**：
- 全局 10 并发 GLM 出站：GLM API 单 key QPS 限制 ~ 10（保守估算），且每流式调用平均占用 ~5s，10 并发 = 2 QPS 长期可持续。
- 单用户 1 并发：避免一个用户开多个会话并行触发，把全局限额耗光。
- 全局 50 SSE 流：预留 goroutine / 文件描述符余量；可通过压测调整。

**新增错误码**（tech-design.md §4.1 补充）：

| HTTP | code | 场景 |
|------|------|------|
| 429 | `ERR_COPILOT_USER_STREAM_BUSY` | 单用户已有活跃 SSE 流 |
| 503 | `ERR_COPILOT_LLM_OUTBOUND_BUSY` | 全局 GLM 出站信号量耗尽（5s 排队超时） |

## 7. 端到端时序示例

### 7.1 单意图写操作（完整 LLM 调用链）

以 "创建一个 P1 事项叫认证模块" 为例，标注 LLM 调用次数与 Provider 事件流：

```
Frontend            Handler          Orchestrator     Planner.Agent    Provider      GLM API
   │                   │                  │                 │              │             │
   │──POST free_text──>│                  │                 │              │             │
   │                   │──create turn─────│                 │              │             │
   │                   │──persist user────│                 │              │             │
   │                   │──start SSE──────│                  │              │             │
   │<──SSE 200─────────│                  │                 │              │             │
   │                   │──HandleUserMsg──>│                 │              │             │
   │                   │                  │──StreamRun──────>│              │             │
   │                   │                  │                 │──Build ctx── │             │
   │                   │                  │                 │  (sys+hist+user)            │
   │                   │                  │                 │              │             │
   │                   │                  │                 │  ===== iter 1 (Planner) =====│
   │                   │                  │                 │──StreamChat─>│             │
   │                   │                  │                 │              │──POST──────>│
   │                   │                  │                 │              │<──SSE stream│
   │                   │                  │                 │<──delta × N──│             │
   │<──thinking × N────│<──thinking × N──│<──thinking × N──│              │             │
   │                   │                  │                 │<──tool_call──│             │
   │                   │                  │                 │  (submit_rewrite)           │
   │                   │                  │                 │              │             │
   │                   │                  │                 │  Execute(submit_rewrite)    │
   │                   │                  │                 │  → writes input_rewrite evt │
   │<──input_rewrite───│<──input_rewrite─│<──input_rewrite─│              │             │
   │                   │                  │                 │  returns success（不终止）  │
   │                   │                  │                 │  append tool msg            │
   │                   │                  │                 │              │             │
   │                   │                  │                 │  ===== iter 2 (Planner) =====│
   │                   │                  │                 │──StreamChat─>│──POST──────>│
   │                   │                  │                 │<──delta × N──│<──SSE stream│
   │<──thinking × N────│<──thinking × N──│<──thinking × N──│              │             │
   │                   │                  │                 │<──tool_call──│             │
   │                   │                  │                 │  (submit_intent)            │
   │                   │                  │                 │              │             │
   │                   │                  │                 │  Execute(submit_intent)     │
   │                   │                  │                 │  → persist intent msg       │
   │                   │                  │                 │  → writes card_message evt  │
   │<──card_message────│<──card_message──│<──card_message──│  → returns terminal         │
   │   (intent)        │                  │                 │  break loop                 │
   │<──turn_phase_done─│<──turn_phase_done<──turn_phase_done│              │             │
   │                   │                  │                 │              │             │
   │──POST confirm────>│                  │                 │              │             │
   │                   │──update turn─────│                 │              │             │
   │                   │──patch intent────│                 │              │             │
   │                   │──ExecuteFromInt──>│                 │              │             │
   │                   │                  │──StreamRun──────>│ (Writer)     │             │
   │                   │                  │                 │              │             │
   │                   │                  │                 │  ===== iter 1 (Writer) =====│
   │                   │                  │                 │──StreamChat─>│──POST──────>│
   │                   │                  │                 │<──delta × N──│<──SSE stream│
   │<──thinking × N────│<──thinking × N──│<──thinking × N──│              │             │
   │                   │                  │                 │<──tool_call──│             │
   │                   │                  │                 │  (emit_form_card, Emission) │
   │                   │                  │                 │              │             │
   │                   │                  │                 │  Execute(emit_form_card)    │
   │                   │                  │                 │  → persist form msg         │
   │                   │                  │                 │    (targetEntity.bizKey="") │
   │                   │                  │                 │  → writes card_message evt  │
   │<──card_message────│<──card_message──│<──card_message──│  → returns terminal         │
   │   (form, 预填)    │                  │                 │  break loop                 │
   │<──step_phase_done─│<──step_phase_done<──step_phase_done│              │             │
   │<──turn_phase_done─│<──turn_phase_done<──turn_phase_done│              │             │
   │                   │                  │                 │              │             │
   │──POST commit─────>│ (commit_card Handler 调 Dispatcher → MainItemService.Create, bizKey 在此阶段才生成)
```

**事件流标注**：
- Emission 工具（`submit_rewrite` / `submit_intent` / `emit_form_card`）的 `tool_call` / `tool_result` 事件被 Agent 抑制（不重复 emit）
- **写工具不存在**——`commit_create` 已移除，DB 写由 commit_card Handler 同步路径触发（[`request-model.md`](./request-model.md) §6.1 请求 3）。Writer Executor 仅靠 `emit_form_card` 终止流，预填表单不写库

### 7.2 LLM 调用次数统计

| 场景 | Planner 调用 | Executor 调用 | 总 LLM 调用 |
|------|-------------|---------------|------------|
| 单意图写（含 form card 中断） | 2（submit_rewrite + submit_intent） | 1（emit_form_card；可能含 1 次 fuzzy_match Read） | 3–4 |
| 主动澄清（一轮问答） | 4（rewrite+intent × 2 轮） | 0（未到执行） | 4 |
| 单意图查询 | 2（rewrite + intent） | 2（query_entities + emit_query_result） | 4 |
| 多意图（2 个 intent） | 2 | writer 1 + reader 2 = 3 | 5 |
| Tool 错误重试（1 次，仅 Read 工具可重试） | — | +1（LLM 自我修正） | +1 |

**配额意义**：单 turn LLM 调用数有上限（`max_iterations × Agent 数`），便于配额管理与成本预估。

**关于 submit_rewrite 单独占一轮**：当前设计把 input_rewrite 和 intent 拆成两次工具调用（两次 LLM 迭代），好处是 input_rewrite 能在 intent 之前先流给前端、LLM 能基于 rewrite 结果再思考 intent。若未来发现成本敏感，可合并为单工具 `submit_planner_output(rewrite, intent)` 把 Planner 调用压到 1 次。

### 7.3 失败传播示例

**场景**：GLM API 在 Writer iter 1 返回 5xx

```
Writer.StreamRun iter 1:
  provider.StreamChat → HTTP 500
  Provider 内部重试 1（等 1s）→ 仍 500
  Provider 内部重试 2（等 2s）→ 仍 500
  Provider 返回 ServerError

Agent.runLoop:
  case provider.EventError:
    outCh <- sse.ErrorEvent(turnID, "LLM service unavailable")
    logAgentCall(usage=nil, status=failed, err)
    return

Orchestrator.ExecuteFromIntent:
  stream 关闭，break 出 intents 循环
  emit turn_phase_done(outcome=failed)
  UPDATE turn SET status='failed'

Handler:
  SSE 流关闭（最后一个事件是 error + turn_phase_done）

Frontend:
  收到 error 事件 → 显示"AI 服务暂时不可用，请稍后重试"
```

**用户体验**：意图消息仍为 `confirmed` 状态（已落库），用户刷新页面后可看到执行失败的提示，可手动重试 `confirm_intent` 触发新一轮执行。

## 8. 关键不变量汇总

1. **Provider 是唯一外部边界**——Agent / Orchestrator / Handler 都不直接发 HTTP
2. **Agent 拥有 LLM↔Tool 循环**——循环边界 = 单次 Agent.StreamRun
3. **流式穿透无缓冲**——Provider→Agent→Orchestrator→Handler 全 channel 串联
4. **ContextBuilder 仅跑一次**——多轮 tool 消息由 Agent 在 messages 数组追加
5. **迭代上限 = 8**——防止 LLM 死循环消耗 token
6. **Tool 错误不中断**——回灌给 LLM，让其决定重试或放弃
7. **Provider 重试仅在建连阶段**——流建立后出错不重试（避免重复输出）
8. **客户端断开立即终止**——ctx 取消贯穿所有 channel
9. **共享单例**——Provider / Agent / ToolRegistry 全应用共享，无 per-request 实例化
10. **API key 永远 env 注入**——不入 git、不入日志、不入错误消息
