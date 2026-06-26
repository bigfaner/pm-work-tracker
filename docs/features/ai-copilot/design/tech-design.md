---
created: "2026-06-26"
updated: "2026-06-26"
prd: prd/prd-spec.md
ui: ui/ui-design.md
status: Draft
---

# Technical Design: AI Copilot 对话助手

> 本设计文档定义 AI Copilot 的技术实现方案，与 PRD 的"做什么"、UI Design 的"长什么样"分离。
>
> 上游文档：
> - PRD：[`prd/prd-spec.md`](../prd/prd-spec.md)
> - UI Design：[`ui/ui-design.md`](../ui/ui-design.md)
> - Entity Schemas：[`prd/entity-schemas.md`](../prd/entity-schemas.md)

## 详细设计索引

本主文档提供总体架构与跨模块视图。各模块详细设计见独立文档：

| 文档 | 内容 |
|------|------|
| [agent-architecture.md](./agent-architecture.md) | Planner & Executors Agent 架构 |
| [request-model.md](./request-model.md) | 请求模型 + 路由分流（一个请求 = 一个 turn = 一个 plan） |
| [interfaces.md](./interfaces.md) | 接口定义（Provider / ContextBuilder / Repository / Dispatcher） |
| [sse-protocol.md](./sse-protocol.md) | SSE 事件协议（`text/event-stream` + 无前缀 JSON） |
| [security.md](./security.md) | 安全策略 |
| [testing-strategy.md](./testing-strategy.md) | 测试策略 |
| [er-diagram.md](./er-diagram.md) | ER 图 |
| [schema.sql](./schema.sql) | 数据库 DDL |
| [api-handbook.md](./api-handbook.md) | API 手册 |

---

## 1. Overview

### 1.1 设计目标

在现有 Go (Gin/GORM) + React/TS 项目上**增量**引入 AI Copilot 模块，**不重构既有 CRUD/状态机/RBAC**。所有写操作仍走现有实体 service，AI 层只做"理解 + 调度"。

### 1.2 核心架构

```
┌─────────────────────────────────────────────────────────────────────┐
│ Frontend (React/TS)                                                  │
│  ├─ CopilotPanel (global overlay, fixed right)                       │
│  ├─ fetch + ReadableStream 解析 SSE 流                               │
│  └─ ENTITY_SCHEMAS + renderEntityCard (来自 UI design)                │
└──────────────────┬──────────────────────────────────────────────────┘
                   │
                   │ POST /sessions/:id/messages  (单一端点，按 type 分派)
                   │ PATCH /messages/:id          (卡片字段/状态更新)
                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│ Backend (Gin) — 新增 internal/copilot/ 子树                           │
│                                                                      │
│  Handler → Orchestrator → Agent Registry → LLM Provider              │
│                                                                      │
│  ┌────────────────────────────────────────────┐                      │
│  │  一个请求 = 一个 turn = 一个 plan            │                      │
│  │  Plan 不入库，不前端缓存，请求结束即释放      │                      │
│  └────────────────────────────────────────────┘                      │
│                                                                      │
│  复用：MainItemService/SubItemService/... (现有 service)              │
│       middleware.Permission / TeamScope (现有 RBAC)                   │
│       available-transitions endpoint (现有状态变更预校验)             │
└─────────────────────────────────────────────────────────────────────┘
              │
              ▼ 5 张新表（3 级数据模型：Session → Turn → Message）
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────────────┐
│ copilot_sessions │ │ copilot_turns    │ │ copilot_messages         │
│ (会话元数据)      │ │ (turn 周期+摘要) │ │ (消息不可变快照)          │
│ status: active/  │ │ status: planning/│ │ status: 多态（按 type）  │
│   archived/...   │ │   awaiting_.../  │ │                          │
└──────────────────┘ │   executing/...  │ └──────────────────────────┘
                     └──────────────────┘
┌──────────────────────────┐ ┌────────────────────────────────────────┐
│ copilot_agent_call_logs  │ │ feature_flags                           │
│ (Agent 调用元数据)        │ │ (Copilot 灰度/熔断开关，30s 热读取缓存)  │
└──────────────────────────┘ └────────────────────────────────────────┘
```

### 1.3 关键设计原则

1. **AI 不直接写库**——AI 解析意图后调用现有实体 service 的 Create/Update/Transition，复用 RBAC + 状态机 + bizKey 体系
2. **Schema 单一来源**——后端把 [`entity-schemas.md`](../prd/entity-schemas.md) 编码为 Go 常量 + JSON，同时驱动 prompt 构造与前端渲染
3. **3 级数据模型**——Session → Turn → Message 自然层级；三层独立 status 状态机；消息单一来源（统一在 messages 表）
4. **后端完全无状态**——每个 HTTP 请求自包含，TurnContext 与 goroutine 同生命周期；跨请求重建从 turn 表读取
5. **意图作为 message 持久化**——意图是 LLM 对用户想法的推测（含文本 + 结构化确认体），作为 `type=intent` 消息持久化；用户确认后从消息重建 plan
6. **可熔断**——feature_flags 表 + agent_call_logs 配额表 + 月度成本累计三道闸

### 1.4 13 项关键决策汇总

| # | 决策 | 选择 | 详细文档 |
|---|------|------|---------|
| 1 | Agent 架构 | **Planner & Executors**：1 个 Planner + 4 个 Executor | [agent-architecture.md](./agent-architecture.md) |
| 2 | 输入改写 | **input_rewrite 嵌入 Planner**：不持久化，仅 agent_call_logs 留底 | [agent-architecture.md](./agent-architecture.md) |
| 3 | 历史摘要 | **规则提取**：零 LLM 调用，嵌入 turn.summary | [interfaces.md](./interfaces.md) §6 |
| 4 | Schema 注入 | **按需加载**：按 entity_type 加载单个 schema | [interfaces.md](./interfaces.md) §5 |
| 5 | Token 超限 | **FIFO 丢弃最早**：滑动窗口 | [interfaces.md](./interfaces.md) §5 |
| 6 | 端点设计 | **单一端点 POST /messages**，按请求体 `type` 字段分派 | [request-model.md](./request-model.md) §4 |
| 7 | 意图存储 | **作为 type=intent 消息持久化**：用户确认后从消息重建 plan，无需前端缓存 | [request-model.md](./request-model.md) §3 |
| 8 | 流式协议 | **SSE**（`text/event-stream` MIME + 无前缀 JSON 内容），泛型 `Event[T]` + kind 判别 | [sse-protocol.md](./sse-protocol.md) |
| 9 | 后端状态 | **完全无状态**：TurnContext 与 goroutine 同生命周期 | [request-model.md](./request-model.md) §3 |
| 10 | 意图确认 | **保留 UF-9 阶段 1**：意图消息推送后等用户"理解正确"确认，再执行 | [request-model.md](./request-model.md) §5 |
| 11 | LLM Provider | **可插拔接口 + Factory**：首版接 GLM，配置驱动切换 | [interfaces.md](./interfaces.md) §1-3 |
| 12 | 卡片字段更新 | **PATCH /messages/:id**（独立端点，RESTful 习惯） | [api-handbook.md](./api-handbook.md) |
| **13** | **数据模型** | **3 级层级 Session→Turn→Message** + 三层独立 status 状态机 + 消息单一来源 | [er-diagram.md](./er-diagram.md) |

### 1.5 PRD 偏离说明

本设计在一点上偏离 PRD：

| PRD 条款 | 偏离 | 理由 |
|---------|------|------|
| 单会话 50 轮上限 | 取消 | 改为 token 预算控制（决策 5） |

**UF-9 意图确认完整保留**：
- 阶段 1 意图回执（文本 + 结构化确认体）作为 `type=intent` 消息推送
- 用户点"✓ 理解正确"触发新请求，从意图消息重建 plan，调 executor
- 用户点"✎ 我要调整"切文本模式重新输入
- 用户点"✗ 取消"终止
- 阶段 2 主动澄清（missing_info）嵌入意图消息（结构化部分），用户回答后重新调 Planner

### 1.6 关键风险与对策

| 风险 | 对策 |
|------|------|
| AI 输出不可控 | Provider 接口 + 后端 schema 校验 + bizKey 权限校验三道闸 |
| 延迟失控 | SSE 流式让用户看到中间过程（首字节 < 1s），最终卡片 P95 < 5s |
| 成本失控 | 每日每用户 50 次硬上限 + 月度 $200 熔断 + 单日 1000 次异常告警 |
| 数据隐私 | 密钥仅后端持有；敏感字段黑名单 + 正则过滤；GLM 国内合规 |

---

## 2. Architecture

### 2.1 Layer Placement

新增 `internal/copilot/` 子树，遵循项目既有分层：

```
backend/internal/copilot/
├── handler/                  # HTTP 层（Gin handler）
│   ├── copilot_handler.go
│   └── stream.go             # SSE 流辅助
├── orchestrator/             # 编排层（goroutine 内运行）
│   ├── orchestrator.go
│   ├── turn_context.go
│   └── summary.go            # 规则提取 turn 摘要
├── agent/                    # Agent 层
│   ├── agent.go
│   ├── registry.go
│   ├── planner.go
│   ├── writer_executor.go
│   ├── reader_executor.go
│   ├── updater_executor.go
│   └── mover_executor.go
├── prompt/                   # Prompt 工程
│   ├── context_builder.go
│   ├── system_prompts.go
│   ├── schema_loader.go
│   ├── token_counter.go
│   └── history_window.go
├── provider/                 # LLM Provider 层
│   ├── provider.go
│   ├── factory.go
│   ├── glm_provider.go
│   ├── mock_provider.go
│   └── recorder.go           # VCR 风格录像/回放
├── sse/                      # SSE 事件协议
│   ├── event.go
│   ├── payload.go
│   └── writer.go
├── model/                    # 数据模型
│   ├── session.go
│   ├── message.go
│   ├── turn_summary.go
│   ├── agent_call_log.go
│   └── feature_flag.go
├── repository/               # 数据访问
│   ├── interfaces.go
│   └── gorm/
├── service/                  # 业务服务
│   ├── dispatcher.go
│   ├── quota.go
│   └── feature_flag_cache.go
└── tools/                    # Agent 工具实现
    ├── tool_registry.go
    ├── fuzzy_match.go
    ├── query_entities.go
    ├── commit.go
    └── validate_transition.go
```

**不修改的现有代码**：
- `internal/service/main_item_service.go` 等实体 service
- `internal/handler/main_item_handler.go` 等 CRUD handler
- `internal/middleware/permission.go` / `team_scope.go`
- `internal/pkg/status/`、`internal/pkg/permissions/`

**仅添加**：以上 `internal/copilot/` 全部为新代码；新增 1 个路由前缀到 `router.go`。

### 2.2 Component Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│ Frontend (CopilotPanel)                                              │
│  fetch + ReadableStream (SSE)                                        │
└──────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────────┐
│ Backend Handler Layer (goroutine per request)                        │
│  PostMessage │ PostAction │ PostCommit │ PatchMessage                │
└──────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────────┐
│ Orchestrator (无状态，goroutine 内运行)                                │
│  - 创建 TurnContext                                                    │
│  - 按 action type 路由                                                 │
│  - 流式聚合 + persist                                                  │
└──────────────────────────────────────────────────────────────────────┘
              │                  │                  │
              ▼                  ▼                  ▼
        ┌──────────┐    ┌─────────────┐    ┌─────────────┐
        │ Planner  │    │ Writer      │    │ Reader      │
        │ Agent    │    │ Executor    │    │ Executor    │
        │          │    │             │    │             │
        │ Tools:   │    │ Tools:      │    │ Tools:      │
        │ - schema │    │ - commit    │    │ - query     │
        │ - members│    │ - fuzzy     │    │             │
        └────┬─────┘    └──────┬──────┘    └──────┬──────┘
             │                 │                  │
             └─────────────────┴──────────────────┘
                               │
                               ▼
                ┌────────────────────────────┐
                │ LLM Provider Factory       │
                │  - GLM (default)           │
                │  - Mock (测试用)            │
                └────────────┬───────────────┘
                             │ HTTPS
                             ▼
                      ┌──────────────┐
                      │ GLM API       │
                      └──────────────┘

复用现有 entity services (零改动):
MainItemService / SubItemService / MilestoneService / ...
```

### 2.3 Dependencies

**新增外部依赖**：无（GLM 通过标准 HTTPS 调用，不引入 SDK）

**新增内部依赖**：
- `internal/copilot/*` → `internal/service/*`（调用现有 entity service）
- `internal/copilot/*` → `internal/middleware/*`（RBAC + TeamScope）
- `internal/copilot/*` → `internal/pkg/snowflake`、`internal/pkg/apperrors`
- `internal/handler/router.go` → 新增 `/api/v1/copilot/*` 路由组

---

## 3. Data Models 概览

> **ER Diagram**: [`design/er-diagram.md`](./er-diagram.md)
> **SQL Schema**: [`design/schema.sql`](./schema.sql)
>
> **3 级数据模型**：Session → Turn → Message，三层独立 status 状态机。

### Field Quick Reference

| Model | Key Fields | Notes |
|-------|-----------|-------|
| `Session` | biz_key, user_id, team_id, current_turn_id, **status**, expires_at | 30 天过期；status: active/archived/expired |
| **`Turn`** | biz_key, session_id, **status**, user_query_short, summary, confirmed_intent, intents_total/done | turn 周期+状态+摘要；status: planning/awaiting_*/executing/done/... |
| `Message` | biz_key, turn_id (FK), intent_id, seq, role, type, **status**, content, trace, card | 统一 status 多态；type: text/trace/card/intent |
| `AgentCallLog` | session_id, turn_id, agent_role, provider, model, input_tokens, user_biz_key, input_rewrite_payload, status | 配额统计 + 调试；status: success/failed/timeout |
| `FeatureFlag` | flag_key, enabled, scope_type, scope_id | 30s 热读取缓存 |

### Go 结构概览

```go
// Session
type Session struct {
    BaseModel
    BizKey         string      `gorm:"type:varchar(36);uniqueIndex;not null"`
    UserID         uint        `gorm:"not null;index"`
    TeamID         *uint       `gorm:"index"`
    TeamName       string      `gorm:"type:varchar(100)"`
    Title          string      `gorm:"type:varchar(100)"`
    CurrentTurnID  string      `gorm:"type:varchar(36);index"`
    Status         SessionStatus `gorm:"type:varchar(32);default:active;index"` // active/archived/expired
    LastActiveAt   time.Time   `gorm:"index"`
    ExpiresAt      time.Time   `gorm:"index"`
}

// Turn（新增）
type Turn struct {
    BaseModel
    BizKey           string         `gorm:"type:varchar(36);uniqueIndex;not null"`
    SessionID        string         `gorm:"type:varchar(36);index;not null"`
    UserBizKey       string         `gorm:"type:varchar(36);index;not null"`
    Status           TurnStatus     `gorm:"type:varchar(32);default:planning;index"`
    UserQueryShort   string         `gorm:"type:varchar(200)"`
    Summary          string         `gorm:"type:varchar(200)"`
    IntentsTotal     int            `gorm:"default:0"`
    IntentsDone      int            `gorm:"default:0"`
    IntentMessageID  *string        `gorm:"type:varchar(36);index"`
    ConfirmedIntent  *IntentPayload `gorm:"type:json"`
    StartedAt        time.Time      `gorm:"not null"`
    LastActiveAt     time.Time      `gorm:"not null"`
    CompletedAt      *time.Time
}

// Message
type Message struct {
    BaseModel
    BizKey     string          `gorm:"type:varchar(36);uniqueIndex;not null"`
    SessionID  string          `gorm:"type:varchar(36);index;not null"`
    TurnID     string          `gorm:"type:varchar(36);index;not null"` // FK to turns
    IntentID   *string         `gorm:"type:varchar(36);index"`
    Seq        int             `gorm:"not null"`
    Role       MsgRole         `gorm:"type:varchar(16);not null"`
    Type       MsgType         `gorm:"type:varchar(16);not null"`
    Status     MsgStatus       `gorm:"type:varchar(32);default:sent;index"` // 多态，按 type 解释
    Content    string          `gorm:"type:text"`
    Trace      *TracePayload   `gorm:"type:json"`
    CardType   *CardType       `gorm:"type:varchar(32)"`
    Card       json.RawMessage `gorm:"type:json"`
    IntentMeta *IntentMeta     `gorm:"type:json"`
}
```

详细字段、关系、索引、状态机见 [`er-diagram.md`](./er-diagram.md) 与 [`schema.sql`](./schema.sql)。

---

## 4. Error Handling

### 4.1 错误类型与编码

| Error Code | Name | Description | HTTP Status |
|------------|------|-------------|-------------|
| `ERR_COPILOT_QUOTA_EXCEEDED` | QuotaExceededError | 用户当日 AI 调用达上限（50 次） | 429 |
| `ERR_COPILOT_FEATURE_DISABLED` | FeatureDisabledError | feature flag 关闭（熔断） | 503 |
| `ERR_COPILOT_AI_TIMEOUT` | AITimeoutError | AI 服务 > 10s 未响应 | 504 |
| `ERR_COPILOT_AI_UNAVAILABLE` | AIUnavailableError | AI 服务整体不可用 | 503 |
| `ERR_COPILOT_PARSE_FAILED` | ParseFailedError | AI 返回不可解析 | 502 |
| `ERR_COPILOT_VALIDATION_FAILED` | ValidationFailedError | available-transitions 预校验失败 | 422 |
| `ERR_COPILOT_PERMISSION_DENIED` | PermissionDeniedError | 用户对目标实体无权限 | 403 |
| `ERR_COPILOT_ENTITY_NOT_FOUND` | EntityNotFoundError | AI 返回的 bizKey 实体已删除 | 404 |
| `ERR_COPILOT_INPUT_TOO_LONG` | InputTooLongError | 输入 > 1000 字符（截断后继续） | — |
| `ERR_COPILOT_TURN_CANCELLED` | TurnCancelledError | 用户取消 | — |
| `ERR_COPILOT_SESSION_EXPIRED` | SessionExpiredError | 会话 > 30 天 | 410 |

### 4.2 错误传播策略

```
LLM Provider 错误
    │
    ▼ Orchestrator 捕获
    │
    ├─ 可恢复（如超时）→ 推送 error event，前端展示重试按钮
    │
    ├─ 不可恢复（如解析失败）→ 推送 error event + use_form fallback
    │
    └─ 系统级（如配额超限）→ 阻断当前流，关闭 SSE，HTTP 错误响应
```

**关键策略**：
- **流中错误**：以 `error` 事件推送，前端展示降级卡片（UF-6）
- **入口错误**：HTTP 错误响应（标准 apperrors 格式）
- **持久化错误**：日志记录，但不阻塞流（best-effort persist）
- **敏感字段过滤未触发**：fail-open（继续调用），但记录告警

### 4.3 流式中断处理

| 中断类型 | 处理 |
|---------|------|
| 客户端断开（fetch 中止） | `ctx.Done()` 触发，goroutine 退出，已 persist 的消息保留 |
| 服务端 panic | recover middleware 捕获，记录日志，关闭流 |
| LLM API 中断 | Provider 返回 error event，Orchestrator 转发并关闭流 |
| 后端 entity service 失败 | commit 端点返回错误，PATCH card state=failed |

---

## 5. Cross-Layer Data Map

数据流跨越 Frontend ↔ Backend ↔ LLM Provider ↔ Existing Entity Service 四层：

| 数据字段 | Frontend | Backend (Handler) | Backend (Orchestrator) | LLM Provider | Entity Service | DB |
|---------|----------|-------------------|----------------------|--------------|----------------|-----|
| user 原文 | textarea | PostMessage req | persist user msg | 注入到 ProviderMsg.user | — | messages.content |
| input_rewrite | UI 显示（trace） | 转发 event | 不 persist | Planner 生成 | — | agent_call_logs.input_rewrite_payload |
| thinking | UI 显示（trace） | 转发 event | 聚合到 trace.thinking | Provider delta | — | messages.trace |
| tool_call | UI 显示（trace） | 转发 event | 聚合到 trace.actions | Provider tool_call | — | messages.trace |
| form card | 渲染卡片 | 转发 event | persist card message | Executor final_card | — | messages.card |
| card.Fields[name] | 用户编辑 | PATCH req | UpdateCardField | — | — | messages.card (JSON) |
| commit result | UI 反馈 | PostCommit resp | — | — | MainItemService.Create 返回 | — |
| followup text | UI 显示 | PostCommit resp | persist followup msg | — | — | messages.content |
| agent call log | — | — | Append log | Provider usage | — | agent_call_logs.* |
| turn summary | — | — | Extract + persist | — | — | turns.summary (嵌入字段) |
| feature flag | UF-1 灰色态 | Health check | 检查（30s 缓存） | — | — | feature_flags.enabled |
| quota count | UI 显示剩余 | Pre-check | CountTodayByUser | — | — | agent_call_logs COUNT |

---

## 6. 配置与部署

### 6.1 新增配置（backend/config.yaml）

```yaml
copilot:
  enabled: true  # 总开关（feature flag 的 yaml fallback）
  
  provider:
    type: glm  # glm / deepseek / openai / mock
    api_key: ${GLM_API_KEY}
    base_url: https://open.bigmodel.cn/api/paas/v4
    model: glm-4-plus
    timeout: 30s
    max_retries: 2
  
  quota:
    daily_calls_per_user: 50
    monthly_cost_limit_usd: 200
    daily_global_call_limit: 1000
  
  feature_flag:
    cache_ttl: 30s
  
  message:
    retention_days: 30
  
  input:
    max_length: 1000
  
  timeout:
    ai_response: 10s
    first_byte: 1s
    plan_visible: 2s
```

### 6.2 环境变量

```bash
GLM_API_KEY=your_api_key_here
COPILOT_ENABLED=true  # 可选，覆盖 config.yaml
```

### 6.3 数据库迁移

新增 5 张表，通过 `internal/migration/runner.go` 自动迁移：

```go
func RunAutoMigrate(db *gorm.DB) error {
    // ... 现有 models ...
    
    // Copilot 新增（3 级数据模型 + 调用日志 + 开关）
    if err := db.AutoMigrate(
        &copilotmodel.Session{},
        &copilotmodel.Turn{},       // turn 主表（含 summary 嵌入字段）
        &copilotmodel.Message{},
        &copilotmodel.AgentCallLog{},
        &copilotmodel.FeatureFlag{},
    ); err != nil {
        return err
    }
    
    return nil
}
```

**SQLite + MySQL 同步**：按 CLAUDE.md 约定，`backend/migrations/SQLite-schema.sql` 和 `MySql-schema.sql` 必须同步新增这 5 张表的 DDL。

### 6.4 Seed 数据

```go
// 默认 feature flag：Copilot 全局关闭，按 Team 灰度开启
feature_flags:
  - flag_key: copilot.enabled
    enabled: false
    scope_type: global
    scope_id: ""
```

### 6.5 Cron 任务

```go
// 每日凌晨清理过期数据
// - copilot_sessions where expires_at < NOW() AND status != 'archived'
// - copilot_turns where session_id in (expired sessions) — 级联
// - copilot_messages where session_id in (expired sessions) — 级联
// - copilot_agent_call_logs where created_at < NOW() - INTERVAL 30 DAY
```

---

## 7. 监控与可观察性

### 7.1 关键指标

| 指标 | 来源 | 告警阈值 |
|------|------|---------|
| AI 响应 P95 延迟 | agent_call_logs.duration_ms | > 5000ms |
| AI 调用失败率 | agent_call_logs.status | > 5% |
| 每日 AI 调用量 | agent_call_logs COUNT | > 800（接近熔断） |
| 月度成本 | SUM(cost_usd) | > $180（接近熔断） |
| 配额超限拒绝率 | ERR_COPILOT_QUOTA_EXCEEDED 计数 | > 10% |
| Feature flag 状态变化 | feature_flags 表 | 任何变化（审计） |

### 7.2 日志结构

```json
{
  "level": "info",
  "module": "copilot.orchestrator",
  "session_id": "sess_xxx",
  "turn_id": "turn_001",
  "step_id": "intent_1",
  "agent_role": "writer",
  "event": "step_completed",
  "duration_ms": 1200,
  "tokens_input": 800,
  "tokens_output": 150,
  "cost_usd": 0.0006,
  "timestamp": "2026-06-26T14:30:00Z"
}
```

### 7.3 可观察性端点

- `GET /api/v1/copilot/health`：AI 服务健康（用于 UF-1 灰色态）
- `GET /api/v1/admin/copilot/metrics`：管理员看板数据
- `GET /api/v1/admin/copilot/feature-flags`：feature flag 状态

---

## 8. 实现里程碑（建议）

按依赖顺序分阶段实现：

| 阶段 | 内容 | 依赖 | 详细文档 |
|------|------|------|---------|
| M1 | Provider 接口 + GLM 实现 + Mock | 无 | [interfaces.md](./interfaces.md) §1-3 |
| M2 | 5 张表 + Repository + Migration | 无 | [schema.sql](./schema.sql) |
| M3 | ContextBuilder + Schema 加载 + Token 计数 | M1 | [interfaces.md](./interfaces.md) §4-5 |
| M4 | Planner Agent + input_rewrite + 工具实现 | M1, M3 | [agent-architecture.md](./agent-architecture.md) |
| M5 | 4 个 Executor + 工具实现 | M1, M3 | [agent-architecture.md](./agent-architecture.md) |
| M6 | Orchestrator + Routing + TurnContext | M4, M5 | [request-model.md](./request-model.md) |
| M7 | Handler + SSE 流 + Event 类型 | M6 | [sse-protocol.md](./sse-protocol.md) |
| M8 | Dispatcher（复用现有 entity service） | 无 | [interfaces.md](./interfaces.md) §7 |
| M9 | feature_flags + 配额检查 | M2 | [security.md](./security.md) §7 |
| M10 | 前端 CopilotPanel + fetch/ReadableStream | M7 | UI Design |
| M11 | 集成测试 + E2E | 全部 | [testing-strategy.md](./testing-strategy.md) |

---

## 附录 A：Plan 跨 turn 评估结论

**Plan 跨 turn 不必要**。详细评估：

| # | 场景 | 是否需要 plan 跨 turn | 替代机制 |
|---|------|---------------------|---------|
| 1 | 跨会话恢复未完成 plan | ❌ PRD 排除 | — |
| 2 | 新 turn 引用旧 turn 成果 | ❌ | messages 历史 + input_rewrite |
| 3 | 新 turn 修改旧 turn 成果 | ❌ | messages 历史 + input_rewrite |
| 4 | 长任务分多次完成 | ❌ PRD 排除 | — |
| 5 | 同会话内多 turn 累积草稿 | ❌ | messages 历史（DraftState 不跨） |
| 6 | Plan step 失败后跨 turn 重试 | ❌ PRD 排除 | — |
| 7 | 多意图 plan 中途放弃后续 step | ❌ PRD 排除 | — |
| 8 | Turn 中途插入新意图 | ❌ | 每 turn 独立 plan；多卡片共存 |

**核心理由**：
1. PRD 明确排除支持 plan 跨 turn 的所有场景
2. Messages 历史已足够承担"新 turn 感知旧 turn"的需求
3. 业界主流（ChatGPT/Claude/Cursor/LangGraph）都不持久化 plan
4. 持久化引入状态机 + 一致性 + 并发 + 隐私问题，复杂度高

---

## 附录 B：决策溯源（11 项决策讨论要点）

| # | 决策 | 讨论要点 |
|---|------|---------|
| 1 | Planner & Executors | 替代单 agent + function calling，关注点分离 |
| 2 | input_rewrite 嵌入 Planner | 改写不持久化（保真），仅 agent_call_logs 留底 |
| 3 | 历史摘要规则提取 | 零 LLM 调用，避免成本翻倍 |
| 4 | Schema 按需注入 | 按 entity_type 加载单个 schema，token 不浪费 |
| 5 | Token 超限 FIFO 丢弃 | 最简单，避免压缩任务复杂度 |
| 6 | 路由分流 | 结构化操作绕过 LLM（业界共识） |
| 7 | 意图作为消息持久化 | 意图 = LLM 对用户想法的推测（文本 + 结构化确认体）；用户确认后重建 plan，无需前端缓存 |
| 8 | SSE 流式协议 | `text/event-stream` MIME + 无前缀 JSON 内容；兼顾标准 MIME 与简洁格式 |
| 9 | 后端无状态 | TurnContext 与 goroutine 同生命周期，水平扩展天然支持 |
| 10 | 保留意图确认 | UF-9 阶段 1 不取消；意图消息推送后等用户确认再执行 |
| 11 | LLM Provider 可插拔 | 接口 + Factory，首版接 GLM |
| 12 | 单一端点 + type 分派 | POST /messages 统一入口，按 type 字段路由处理逻辑 |
| 13 | 3 级数据模型 | Session→Turn→Message 三级层级；三层独立 status 状态机；消息单一来源；turns 表不冗余 user_message |

---

> 本设计文档完成于 2026-06-26。所有 11 项关键决策经多轮讨论确认。
> 接下来进入 DB schema review gate（[`er-diagram.md`](./er-diagram.md) + [`schema.sql`](./schema.sql)）。
