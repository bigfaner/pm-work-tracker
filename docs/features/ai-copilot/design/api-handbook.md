---
created: "2026-06-30"
updated: "2026-06-30"
parent: tech-design.md
---

# API Handbook: HTTP 端点契约

> 返回 [`tech-design.md`](./tech-design.md)

本文档列出 Copilot 模块所有 HTTP 端点（请求/响应 schema、错误码映射）。所有路径前缀：`/api/v1/copilot`。鉴权：JWT（既有 middleware），Team scope 由 `TeamScopeMiddleware` 注入（AB-003）。

## 1. 端点总览

| Method | Path | LLM | 响应模式 | 用途 |
|--------|------|-----|---------|------|
| POST   | `/sessions` | ✗ | JSON | 创建会话 |
| GET    | `/sessions` | ✗ | JSON | 列出会话（分页） |
| GET    | `/sessions/:id` | ✗ | JSON | 会话详情（含 `current_turn_id` + `current_turn_status`） |
| DELETE | `/sessions/:id` | ✗ | JSON | 软删除会话 |
| GET    | `/sessions/:id/messages` | ✗ | JSON | 历史消息（分页） |
| POST   | `/sessions/:id/messages` | 视 type | SSE/JSON | 单一入口，按 `type` 分派 |
| PATCH  | `/messages/:id` | ✗ | JSON | 卡片字段编辑 / 展开 / 应用 diff |
| GET    | `/health` | ✗ | JSON | 健康探针（UF-1 灰色态） |

## 2. POST /sessions/:id/messages（核心单一端点）

请求体 `MessageRequest`（见 [`interfaces.md`](./interfaces.md) §10）含 `type` 字段决定处理逻辑。

### 2.1 Request 类型与响应

| `type` | LLM | 响应 | 入参字段 |
|--------|-----|------|---------|
| `free_text` | Planner | SSE | `content: string` |
| `answer_clarify` | Planner | SSE | `answer: string`, `intentMessageId: string` |
| `adjust_intent` | Planner | SSE | `intentMessageId: string`, `newContent: string` |
| `confirm_intent` | Executor | SSE | `intentMessageId: string` |
| `select_candidate` | Executor | SSE | `messageId: string`, `candidateBizKey: string` |
| `commit_card` | ✗ | JSON | `messageId: string` |
| `cancel` | ✗ | JSON | `messageId?: string` |

### 2.2 SSE 响应（free_text / answer_clarify / adjust_intent / confirm_intent / select_candidate）

- Content-Type: `text/event-stream; charset=utf-8`
- 内容格式：每行一个 JSON 对象（无 `data:` 前缀）
- 事件类型见 [`sse-protocol.md`](./sse-protocol.md) §2

### 2.3 JSON 响应：`commit_card`

```json
{
  "messageId": "msg_003",
  "newState": "submitted",
  "createdEntity": { "bizKey": "MI-0023", "title": "认证模块", "bizCode": "MI-0023" },
  "followupMessageId": "msg_004",
  "followupContent": "已为你创建 P1 事项「认证模块」（MI-0023）。"
}
```

**`commit_card` 错误响应**（entity service 校验失败）：

| HTTP | code | 场景 |
|------|------|------|
| 422 | `ERR_COPILOT_VALIDATION_FAILED` | 必填字段缺失 / 状态变更预校验失败（payload 含 `errors.validTransitions`） |
| 403 | `ERR_COPILOT_PERMISSION_DENIED` | RBAC 拒绝 |
| 404 | `ERR_COPILOT_ENTITY_NOT_FOUND` | 目标实体已删除 |

`commit_card` 失败时 HTTP 状态码 = 实体 service 返回的映射；body 同时含 `messageId` + `newState:"failed"` + `errors`，前端据此把 form card 切到 `failed` 状态保留用户输入。详见 [`state-machines.md`](./state-machines.md) §4.4。

### 2.4 JSON 响应：`cancel`

```json
{ "cancelled": true, "turnStatus": "cancelled" }
```

## 3. PATCH /messages/:id

更新卡片字段、展开/折叠记录、应用/丢弃 diff。请求体：

```json
{
  "op": "edit_field" | "expand" | "apply_diff" | "discard_diff",
  "fieldName": "title",        // op=edit_field
  "value": "认证模块 v2",       // op=edit_field
  "recordBizKey": "MI-0007",   // op=expand
  "expanded": true,            // op=expand
  "diffId": "diff_001",        // op=apply_diff / discard_diff
  "accepted": true             // op=apply_diff：true=应用对方变更，false=保留本地
}
```

响应（200）：

```json
{ "messageId": "msg_003", "newState": "prefilled", "appliedAt": 1734900000000 }
```

并发编辑（last-write-wins + diff 展示）详见 §4 与 [`request-model.md`](./request-model.md) §1.4。

## 4. 并发编辑（last-write-wins + diff 展示）

PRD Story 7 要求：当用户直接编辑卡片字段与对话补充指令同时发生时，时间戳晚者胜出；但对话产生的增量变更在覆盖前先展示 diff。

**模型**：

- 每条 form card 消息携带 `cardData.fields[].version`（int64，每次 PATCH 递增）与 `cardData.lastEditedAt`（int64 ms）。
- `PATCH /messages/:id op=edit_field` 写入时携带客户端读出的 `version`；后端在事务内：
  1. 读取当前 DB version（`v_db`）与请求 version（`v_req`）
  2. 若 `v_req == v_db`：直接应用 → version+1，更新 lastEditedAt
  3. 若 `v_req < v_db`（已被对话路径覆盖）：返回 409 + 当前 DB 字段值，前端展示 `diffOverlay`（DiffOverlay 结构见 [`interfaces.md`](./interfaces.md) §11），用户选 `apply_diff` 或 `discard_diff`
- 对话路径触发字段变更（如 `adjust_intent` 触发 Planner 重新推断字段）写同一张表，同样递增 version；冲突检测一致。

**冲突响应**（409）：

```json
{
  "code": "ERR_COPILOT_FIELD_CONFLICT",
  "messageId": "msg_003",
  "fieldName": "title",
  "yourValue": "认证模块 v2",
  "currentValue": "认证模块（被对话更新）",
  "currentVersion": 5,
  "diffOverlay": { /* DiffOverlay */ }
}
```

错误码 `ERR_COPILOT_FIELD_CONFLICT` 加入 tech-design.md §4.1 表（HTTP 409）。

## 5. GET /health

```json
{ "status": "ok" | "degraded" | "down", "provider": "glm", "model": "glm-4-plus" }
```

- `ok`：provider 可用（最近 1 分钟内有成功调用）
- `degraded`：最近失败率 > 5%
- `down`：feature flag 关闭或最近 1 分钟全失败

前端 UF-1 灰色态由 `degraded`/`down` 触发。

## 6. 错误码总表（与 tech-design.md §4.1 对齐）

| HTTP | code | 说明 |
|------|------|------|
| 429 | `ERR_COPILOT_QUOTA_EXCEEDED` | 当日 AI 调用达上限（50 次） |
| 503 | `ERR_COPILOT_FEATURE_DISABLED` | feature flag 关闭（熔断） |
| 504 | `ERR_COPILOT_AI_TIMEOUT` | AI > 10s 未响应（SSE 流首字节前） |
| 503 | `ERR_COPILOT_AI_UNAVAILABLE` | AI 服务整体不可用（pre-flight） |
| 502 | `ERR_COPILOT_PARSE_FAILED` | AI 输出不可解析 |
| 422 | `ERR_COPILOT_VALIDATION_FAILED` | available-transitions 预校验失败 |
| 403 | `ERR_COPILOT_PERMISSION_DENIED` | RBAC 拒绝 |
| 404 | `ERR_COPILOT_ENTITY_NOT_FOUND` | bizKey 实体已删除 |
| 409 | `ERR_COPILOT_FIELD_CONFLICT` | 并发字段冲突（diff 预览） |
| 409 | `ERR_COPILOT_TURN_IN_FLIGHT` | 当前 turn 在 LLM 处理中 |
| 410 | `ERR_COPILOT_SESSION_EXPIRED` | 会话 > 30 天 |

**错误响应 body（标准 apperrors 格式）**：

```json
{
  "code": "ERR_COPILOT_QUOTA_EXCEEDED",
  "message": "今日 AI 调用已达上限（50 次）",
  "recoverable": true,
  "fallbackAction": "use_form"
}
```

## 7. 配额错误路径（pre-flight HTTP 429，不进 SSE 流）

配额检查在 SSE 流**打开之前**执行（security.md §7.1）。若超限：HTTP 429 + 上述 body，**不写 SSE 头**，前端按降级卡片渲染（fallbackAction=use_form 或 keyword_mode）。

> 在 SSE 流**打开之后**出现的错误（如 LLM 中途超时、流中 parse 失败）走 `error` 事件（sse-protocol.md §9）。配额错误不会以 SSE `error` 事件形式出现——配额在 pre-flight 已 gating。详见 [`security.md`](./security.md) §7.1。
