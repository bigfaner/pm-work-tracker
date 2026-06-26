---
created: "2026-06-26"
updated: "2026-06-26"
parent: tech-design.md
---

# Testing Strategy

> 返回 [`tech-design.md`](./tech-design.md)

## 1. 测试金字塔

```
       ┌─────────────────┐
       │   E2E (~5%)     │  Copilot 全流程：发指令 → 提交 → 验证创建
       ├─────────────────┤
       │ Integration (~25%)│  Handler → Orchestrator → Mock Provider
       ├─────────────────┤
       │   Unit (~70%)   │  Agent / Repository / Tool 各自独立测试
       └─────────────────┘
```

**项目约定**（见 [`docs/conventions/testing/`](../../../../conventions/testing/)）：
- Backend：Go 标准 `testing` + `testify`
- Frontend：Vitest + React Testing Library + MSW
- 测试覆盖率：核心模块 ≥ 80%

## 2. LLM Provider 测试（接口 mock + 录像回放）

### 2.1 Mock Provider（单元测试用）

```go
// internal/copilot/provider/mock_provider.go

type MockProvider struct {
    responses map[string][]ProviderEvent  // key = 测试场景标识
    calls     []ProviderParams            // 记录所有调用（断言用）
    mu        sync.Mutex
}

func NewMockProvider() *MockProvider {
    return &MockProvider{
        responses: make(map[string][]ProviderEvent),
    }
}

func (m *MockProvider) StreamChat(ctx context.Context, p ProviderParams) (<-chan ProviderEvent, error) {
    m.mu.Lock()
    m.calls = append(m.calls, p)
    m.mu.Unlock()
    
    ch := make(chan ProviderEvent, 10)
    go func() {
        defer close(ch)
        key := fmt.Sprintf("%s:%s", p.Model, extractScenario(p))
        for _, evt := range m.responses[key] {
            ch <- evt
        }
    }()
    return ch, nil
}

func (m *MockProvider) CountTokens(text string) int {
    return len(text) / 4  // 粗略估算
}

func (m *MockProvider) Models() []ModelInfo {
    return []ModelInfo{{Name: "mock-model", ContextWindow: 8000}}
}

// 测试辅助：预置响应
func (m *MockProvider) When(scenario string, events []ProviderEvent) {
    m.responses[scenario] = events
}

// 测试辅助：断言调用参数
func (m *MockProvider) AssertCalled(t *testing.T, want ProviderParams) {
    m.mu.Lock()
    defer m.mu.Unlock()
    for _, call := range m.calls {
        if reflect.DeepEqual(call, want) {
            return
        }
    }
    t.Errorf("provider was not called with expected params")
}
```

### 2.2 VCR 风格录像/回放（集成测试用）

```go
// internal/copilot/provider/recorder.go

type RecorderProvider struct {
    inner    Provider
    mode     RecordMode  // record / replay
    cassette string      // 录像文件路径
    tapes    map[string][]ProviderEvent
}

type RecordMode int
const (
    ModeRecord RecordMode = iota
    ModeReplay
)

func NewRecorderProvider(inner Provider, cassette string, mode RecordMode) *RecorderProvider {
    r := &RecorderProvider{inner: inner, cassette: cassette, mode: mode}
    if mode == ModeReplay {
        r.loadCassette()  // 从 JSON 文件加载预录响应
    }
    return r
}

func (r *RecorderProvider) StreamChat(ctx context.Context, p ProviderParams) (<-chan ProviderEvent, error) {
    key := r.fingerprint(p)  // 基于参数生成指纹
    
    if r.mode == ModeReplay {
        // 回放模式：从 cassette 读取
        if events, ok := r.tapes[key]; ok {
            ch := make(chan ProviderEvent, 10)
            go func() {
                defer close(ch)
                for _, evt := range events {
                    ch <- evt
                }
            }()
            return ch, nil
        }
        return nil, fmt.Errorf("cassette missing scenario: %s", key)
    }
    
    // 录制模式：真打 GLM，记录响应
    ch, err := r.inner.StreamChat(ctx, p)
    if err != nil {
        return nil, err
    }
    
    out := make(chan ProviderEvent, 10)
    var recorded []ProviderEvent
    go func() {
        defer close(out)
        for evt := range ch {
            recorded = append(recorded, evt)
            out <- evt
        }
        r.tapes[key] = recorded
        r.saveCassette()  // 持久化到 JSON 文件
    }()
    return out, nil
}
```

**测试数据存放**：`backend/tests/copilot/cassettes/*.json`

**录制工作流**：
1. 首次运行设置 `COPILOT_RECORD_MODE=record`，真打 GLM 记录响应
2. 提交 cassette 文件到 git
3. 后续 CI 运行用 `mode=replay`，零成本 + 零延迟

## 3. 关键测试场景

### 3.1 Planner 测试

| 场景 | 验证点 |
|------|-------|
| 单意图清晰输入 | plan.intents 长度=1，无 missing_info |
| 单意图含代词 | input_rewrite 正确消解（"它" → bizKey） |
| 多意图（X 然后 Y） | plan.intents 长度=2，正确路由 writer/reader |
| 必填字段缺失 | missing_info 列出所有缺失字段 |
| 用户回答后再次调用 | DraftState 注入，plan 完整 |
| 配额超限 | 返回 ERR_COPILOT_QUOTA_EXCEEDED |
| 提示注入过滤 | 输入含 "ignore previous" 时被 REDACTED |
| 敏感字段过滤 | 输入含 password 时被 REDACTED |

### 3.2 Executor 测试

| 场景 | 验证点 |
|------|-------|
| writer 创建 MainItem | tool_call commit_create，form card 含正确字段 |
| writer 含 assignee 模糊 | 先 fuzzy_match_member，再 commit_create |
| reader 单记录 | query_result card 中 records[0].expanded=true |
| reader 多记录 | records 全部 expanded=false |
| reader 超过 20 条 | truncated=true |
| updater 状态变更 | 先 validate_transition，通过后 commit_update |
| updater 预校验失败 | card state=validation，errors.validTransitions 列出合法目标 |
| updater 跨 Team 拒绝 | tool_result status=error |
| mover 跨 Team | 拒绝（validate_source_target 失败） |
| mover 源 terminal | 拒绝 |

### 3.3 Orchestrator 测试

| 场景 | 验证点 |
|------|-------|
| 单意图完整流程 | Planner → Executor → form card → 中断 |
| 多意图中断 | intent_1 执行后中断，intent_2 丢弃 |
| clarify 流程 | missing_info → 等回答 → 重新调 Planner |
| 选候选 | select_candidate → Executor 注入 bizKey |
| 流式中断 | ctx.Done() 触发，已 persist 消息保留 |
| feature flag 关闭 | 返回 ERR_FEATURE_DISABLED |
| 配额预检 | 用户当日已达 50 次时拒绝 |

### 3.4 Handler 测试

| 场景 | 验证点 |
|------|-------|
| SSE 格式正确 | 每行一个 JSON，Content-Type 为 text/event-stream，无 event:/data: 前缀 |
| feature flag 关闭 | 503 + ERR_COPILOT_FEATURE_DISABLED |
| 权限校验 | 中间件拒绝时返回 403 |
| 配额预检 | 用户当日已达 50 次时 429 |
| commit 走 entity service | 不调 LLM，直接 entity service |
| commit 失败 | card state=failed + 字段错误 |
| PATCH 字段 | 更新 messages.card JSON |

### 3.5 Repository 测试

| 场景 | 验证点 |
|------|-------|
| Append + GetByBizKey | round-trip 一致 |
| ListBySession 分页 | limit/offset 正确 |
| ListByTurn 顺序 | 按 seq 升序 |
| UpdateStatus | JSON 字段（card / intent / trace）状态更新正确 |
| 软删除 | deleted_at 非空，查询过滤 |
| 配额查询 | COUNT(DATE(created_at) = today) 正确 |

## 4. 前端测试

### 4.1 Vitest + React Testing Library

```typescript
// frontend/src/__tests__/copilot/CopilotPanel.test.tsx

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CopilotPanel } from '../components/copilot/CopilotPanel';

describe('CopilotPanel', () => {
  it('renders float bubble by default', () => {
    render(<CopilotPanel />);
    expect(screen.getByLabelText('AI 助手')).toBeInTheDocument();
  });
  
  it('opens panel on bubble click', async () => {
    render(<CopilotPanel />);
    fireEvent.click(screen.getByLabelText('AI 助手'));
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });
  });
});
```

### 4.2 Mock fetch + ReadableStream

```typescript
// frontend/src/__tests__/copilot/sse-parser.test.ts

function createMockStream(chunks: string[]): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder();
  return new ReadableStream({
    start(controller) {
      chunks.forEach(chunk => controller.enqueue(encoder.encode(chunk)));
      controller.close();
    }
  });
}

describe('SSE parser', () => {
  it('parses multi-line stream correctly', async () => {
    const mockResponse = new Response(createMockStream([
      '{"event":"thinking","data":{"kind":"thinking","content":"..."\n',
      '}}\n',
      '{"event":"text_message","data":{"kind":"text","content":"hello"}}\n',
    ]));
    
    jest.spyOn(global, 'fetch').mockResolvedValue(mockResponse);
    
    const events: SSEEvent[] = [];
    await postSSE('/test', {}).then(async () => {
      // 事件已收集
    });
    
    // 验证事件被正确解析和分发
  });
});
```

### 4.3 前端关键测试场景

| 场景 | 验证点 |
|------|-------|
| SSE 解析 | 多行/单行/不完整行正确处理 |
| 事件分发 | 14 种事件类型正确路由到处理器 |
| TurnContext 缓存 | plan_ready 事件正确缓存（仅展示用） |
| 输入区双模式 | 文本 ↔ 选项组切换 |
| 卡片状态 | prefilled → submitting → submitted |
| 错误降级 | error 事件触发降级卡片 |
| feature flag 灰色态 | health 检查失败时 UF-1 灰色 |

## 5. 集成测试（API 层）

### 5.1 完整流程测试

```go
// backend/tests/copilot/integration_test.go

func TestPostMessage_FreeText_PlannerSuccess(t *testing.T) {
    // Setup: mock provider 返回固定 plan
    provider := newMockProvider()
    provider.When("glm-4-plus:create_main_item", []ProviderEvent{
        {Type: ProviderEventDelta, Delta: "thinking..."},
        {Type: ProviderEventDone, Usage: &ProviderUsage{InputTokens: 100, OutputTokens: 50}},
    })
    
    handler := newTestHandler(provider)
    req := httptest.NewRequest("POST", "/api/v1/copilot/sessions/sess_xxx/messages",
        strings.NewReader(`{"content":"创建 P1 事项"}`))
    req.Header.Set("Accept", "text/event-stream")
    
    rec := httptest.NewRecorder()
    handler.PostMessage(rec, req)
    
    // 验证响应是 SSE 格式
    assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
    
    // 解析每行 JSON
    lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
    assert.Greater(t, len(lines), 0)
    
    for _, line := range lines {
        var evt map[string]any
        require.NoError(t, json.Unmarshal([]byte(line), &evt))
        assert.Contains(t, evt, "event")
        assert.Contains(t, evt, "data")
    }
    
    // 验证消息已持久化
    msgs, _ := handler.msgRepo.ListBySession(ctx, "sess_xxx", 10, 0)
    assert.Greater(t, len(msgs), 0)
}
```

### 5.2 多意图中断测试

```go
func TestMultiIntent_InterruptsAtFirstFormCard(t *testing.T) {
    provider := newMockProvider()
    // Planner 返回 2 intent：writer + reader
    provider.When("glm-4-plus:multi_intent", []ProviderEvent{
        {Type: ProviderEventDelta, Delta: "two intents..."},
        {Type: ProviderEventDone, Usage: &ProviderUsage{}},
    })
    
    // Writer executor 返回 form card
    writerExec := newMockExecutor()
    writerExec.When("create_main_item", []Event{
        cardMessageEvent("form", "prefilled"),
    })
    
    handler := newTestHandler(provider, writerExec)
    
    // 发送多意图指令
    req := newPostMessageReq("创建 P1 事项 + 查我的 P0")
    rec := httptest.NewRecorder()
    handler.PostMessage(rec, req)
    
    // 验证：流中有 intent_1 完成 + intent_2 未执行
    lines := parseSSELines(rec.Body.String())

    hasIntent1Started := containsEvent(lines, "step_started", "intent_1")
    hasIntent2Started := containsEvent(lines, "step_started", "intent_2")
    
    assert.True(t, hasIntent1Started)
    assert.False(t, hasIntent2Started, "intent_2 should not start (interrupted)")
    
    // 验证 turn_phase_done 含 await_commit
    lastEvent := lines[len(lines)-1]
    assert.Equal(t, "turn_phase_done", lastEvent["event"])
    assert.Equal(t, "await_commit", 
        lastEvent["data"].(map[string]any)["nextAction"].(map[string]any)["type"])
}
```

### 5.3 录像回放测试

```go
func TestWithCassette_GLMRealResponse(t *testing.T) {
    // 使用预录的 GLM 响应（首次运行时录制）
    provider := NewRecorderProvider(
        NewGLMProvider(...),
        "tests/copilot/cassettes/glm_create_main_item.json",
        ModeReplay,
    )
    
    handler := newTestHandler(provider)
    
    req := newPostMessageReq("创建 P1 事项叫认证模块")
    rec := httptest.NewRecorder()
    handler.PostMessage(rec, req)
    
    // 验证真实 GLM 响应被正确解析
    lines := parseSSELines(rec.Body.String())
    assert.NotEmpty(t, lines)
    
    // 验证最终 card payload 结构正确
    cardEvent := findCardEvent(lines)
    require.NotNil(t, cardEvent)
    assert.Equal(t, "form", cardEvent["data"].(map[string]any)["cardType"])
}
```

## 6. 测试覆盖率目标

| 模块 | 覆盖率 | 说明 |
|------|-------|------|
| `agent/` | ≥ 85% | 核心业务逻辑（Planner / Executors） |
| `orchestrator/` | ≥ 85% | 路由 + 执行 + 重建 |
| `repository/` | ≥ 80% | CRUD |
| `handler/` | ≥ 75% | HTTP 适配 |
| `provider/glm_provider.go` | ≥ 70% | 用录像回放 |
| `prompt/` | ≥ 80% | ContextBuilder + 裁剪 |
| `tools/` | ≥ 85% | 工具执行 |
| `service/` | ≥ 80% | Dispatcher + 配额 |

## 7. CI 集成

### 7.1 GitHub Actions 工作流

```yaml
# .github/workflows/copilot-tests.yml
name: Copilot Tests

on: [push, pull_request]

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.26'
      
      - name: Run unit tests
        working-directory: backend
        run: go test ./internal/copilot/... -v -cover
      
      - name: Run integration tests (replay mode)
        working-directory: backend
        env:
          COPILOT_RECORD_MODE: replay
        run: go test ./tests/copilot/... -v
  
  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '20'
      
      - name: Install deps
        working-directory: frontend
        run: npm ci
      
      - name: Run tests
        working-directory: frontend
        run: npm test
```

### 7.2 录制模式（手动触发）

录制新 cassette 时（如新增测试场景或 GLM API 升级）：

```bash
# 设置真实 API Key
export GLM_API_KEY=your_real_key

# 录制模式运行
cd backend
COPILOT_RECORD_MODE=record go test ./tests/copilot/... -run TestRecord -v

# 提交 cassette 到 git
git add tests/copilot/cassettes/
git commit -m "test: update GLM cassettes"
```

## 8. 性能测试

### 8.1 延迟基准

| 操作 | P50 目标 | P95 目标 |
|------|---------|---------|
| 首字节（thinking 出现） | 500ms | 1s |
| plan_ready | 1.5s | 2.5s |
| 最终卡片（单意图） | 2s | 5s |
| 流式中断恢复 | — | 200ms |

### 8.2 负载测试

```go
func BenchmarkOrchestrator_ParallelRequests(b *testing.B) {
    provider := newMockProvider()
    orchestrator := newTestOrchestrator(provider)
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            turnCtx := newTestTurnContext()
            eventCh := make(chan sse.Event, 64)
            go func() {
                defer close(eventCh)
                orchestrator.HandleUserMessage(ctx, turnCtx, eventCh)
            }()
            for range eventCh {}
        }
    })
}
```

**目标**：20 并发用户下，P95 延迟不超过 5s（与 PRD 一致）。
