---
created: "2026-06-26"
updated: "2026-06-26"
parent: tech-design.md
---

# State Machines: 三层状态机与并发控制

> 返回 [`tech-design.md`](./tech-design.md)

本文档定义 Session / Turn / Message 三层独立状态机、它们的协调关系，以及基于 Turn.status 的并发拦截规则。

## 1. 概览

```
┌─────────────────────────────────────────────────────────────────┐
│ Session.status (宏观，用户级)                                    │
│   active → archived → expired                                   │
└─────────────────────────────────────────────────────────────────┘
              │
              ▼ 1:N
┌─────────────────────────────────────────────────────────────────┐
│ Turn.status (中观，流程级) — 10 个状态                           │
│                                                                 │
│   planning                                                      │
│     ↓                                                           │
│   awaiting_confirm_intent / awaiting_clarify                    │
│     ↓                                                           │
│   executing                                                     │
│     ↓                                                           │
│   awaiting_commit / awaiting_select_candidate                   │
│     ↓                                                           │
│   done / cancelled / superseded / failed                        │
│                                                                 │
│   核心作用：刷新页面恢复 UI、跨请求状态协调、并发拦截            │
└─────────────────────────────────────────────────────────────────┘
              │
              ▼ 1:N
┌─────────────────────────────────────────────────────────────────┐
│ Message.status (微观，单条消息级，按 type 多态)                  │
│   - text: sent                                                  │
│   - trace: streaming → done/failed                              │
│   - intent: awaiting_confirm → confirmed/adjusted/cancelled     │
│   - card/form: prefilled → submitting → submitted/failed        │
│   - card/disambig: awaiting_select → selected/discarded         │
└─────────────────────────────────────────────────────────────────┘
```

### 设计原则

1. **三层独立**——各自状态机，独立 UPDATE
2. **事务一致**——Turn + Message 状态变化在同一事务内
3. **Turn.status 是协调信源**——刷新页面恢复 UI、并发拦截都依赖此字段
4. **多卡片共存**——同会话允许多个 form card 同时存在（不同 turn）
5. **失败不阻塞**——failed 状态可重试，状态在 submitting/failed 间循环

---

## 2. Session 状态机

### 状态枚举（3 个）

| 状态 | 含义 |
|------|------|
| `active` | 活跃会话（默认） |
| `archived` | 用户归档（不再显示在主列表，但可查） |
| `expired` | 已过期（超过 30 天，等待物理清理） |

### 转换图

```
       ┌──────────────────────┐
       │      (created)        │
       └──────────┬───────────┘
                  ▼
            ┌──────────┐
      ┌────▶│  active  │◀──── 用户取消归档（unarchive）
      │     └────┬─────┘
      │          │
      │          ├──── 用户主动归档 ──▶┌──────────┐
      │          │                     │ archived │
      │          │                     └────┬─────┘
      │          │                          │
      │          │                          ├──── cron 30 天 ──▶┌──────────┐
      │          │                          │                    │ expired  │─▶ 物理清理
      │          │                          ├──── 用户恢复 ─────▶│ (active) │
      │          ▼                          │
      │     cron 30 天                     │
      │  (active 直接过期)                 │
      │          │                          │
      └──────────┴──────────────────────────┘
                  ▼
            ┌──────────┐
            │ expired  │ ──▶ cron 物理清理
            └──────────┘
```

### 转换矩阵

| 当前 | 触发 | 新状态 |
|------|------|-------|
| (创建) | 用户创建会话 | `active` |
| `active` | 用户主动归档 | `archived` |
| `active` | 30 天未活动（cron） | `expired` |
| `archived` | 用户取消归档 | `active` |
| `archived` | 30 天（cron） | `expired` |
| `expired` | cron 物理清理 | (删除行) |

---

## 3. Turn 状态机（核心）

### 状态枚举（10 个）

| 状态 | 含义 | 是否终态 |
|------|------|---------|
| `planning` | Planner 流式中 | ❌ |
| `awaiting_clarify` | 等用户回答 missing_info | ❌ |
| `awaiting_confirm_intent` | 等用户"理解正确" | ❌ |
| `executing` | executor 链执行中 | ❌ |
| `awaiting_commit` | 等 form 提交 | ❌ |
| `awaiting_select_candidate` | 等歧义候选选定 | ❌ |
| `done` | 完成 | ✅ |
| `cancelled` | 用户主动取消（点"✗ 取消"） | ✅ |
| `superseded` | 用户发了新指令，原 turn 被替代 | ✅ |
| `failed` | 系统错误（LLM 失败、超时、panic） | ✅ |

### 终态语义对比

| 终态 | 触发主体 | 审计含义 |
|------|---------|---------|
| `done` | 系统 | 流程正常完成 |
| `cancelled` | 用户显式 | 用户主动点"取消"按钮 |
| `superseded` | 用户隐式 | 用户发了新指令，原 turn 被自然替代 |
| `failed` | 系统 | 异常终止 |

### 完整转换图

```
                                  ┌──────────────────────────────┐
                                  │                              │
                                  ▼                              │
                            ┌───────────┐                        │
       用户发指令 ─────────▶│ planning  │                        │
                            └─────┬─────┘                        │
                                  │                              │
                  ┌───────────────┼───────────────┐              │
                  │               │               │              │
                  ▼               ▼               ▼              │
            Planner 输出     Planner 输出     Planner 流失败      │
            missing_info     完整 plan                        │
                  │               │               │              │
                  ▼               ▼               ▼              │
        ┌─────────────────┐  ┌──────────────────────┐   ┌────────┴───┐
        │awaiting_clarify │  │awaiting_confirm_intent│   │  failed    │
        └────────┬────────┘  └───────────┬──────────┘   └────────────┘
                 │                       │
                 │ 用户回答               │ ├─ "理解正确"
                 │                       │ ├─ "我要调整" ──▶ 回到 planning
                 │                       │ ├─ "取消" ─────▶ cancelled
                 │                       │ └─ 用户发新指令 ──▶ superseded
                 │                       │
                 │                       ▼
                 │              ┌────────────────┐
                 └─────────────▶│   executing    │
                                └───────┬────────┘
                                        │
                          ┌─────────────┼─────────────┬──────────────┐
                          │             │             │              │
                     form 推送    query_result    disambig 推送   所有 intent 完成
                          │             │             │              │
                          ▼             ▼             ▼              ▼
                ┌────────────────┐  继续/完成  ┌──────────────────┐  ┌──────┐
                │awaiting_commit │            │awaiting_select_  │  │ done │
                └───────┬────────┘            │   candidate      │  └──────┘
                        │                     └────────┬─────────┘
                        │                              │
                ┌───────┼────────┐              用户选定│
                │       │        │                     │
            提交成功  提交失败  取消                    │
            (最后一个  (重试)    │                     │
             intent)    │        │                     │
                │       │        │                     │
                ▼       │        ▼                     │
              done     └─ executing  cancelled         │
              (或继续   (form=failed)                  │
               下一个                                    │
               intent)                                  │
                                                       ▼
                                                  回到 executing

  ★ 任何非终态 + 用户发新指令 ──▶ superseded
  ★ 任何非终态 + 1 小时无活动 ──▶ failed（cron 孤儿清理）
```

### 转换矩阵（22 条规则）

| # | 当前状态 | 触发 | 新状态 | 关联 Message 变化 |
|---|---------|------|-------|------------------|
| 1 | (创建) | 用户发指令（free_text） | `planning` | 创建 user text msg |
| 2 | `planning` | Planner 完成（无 missing_info） | `awaiting_confirm_intent` | 创建 trace + intent msg；记录 turn.intent_message_id |
| 3 | `planning` | Planner 完成（有 missing_info） | `awaiting_clarify` | 同上（intent msg 含 missingInfo） |
| 4 | `planning` | Planner 流失败 / 超时 | `failed` | trace.status=failed；推送 fallback card |
| 5 | `awaiting_clarify` | 用户回答（answer_clarify） | `planning` | 创建 user text msg；保留原 intent msg（awaiting_confirm） |
| 6 | `awaiting_confirm_intent` | 用户点"理解正确" | `executing` | intent.status=`confirmed`；turn.confirmed_intent 填充 |
| 7 | `awaiting_confirm_intent` | 用户点"我要调整" | `planning` | intent.status=`adjusted`；创建新 user msg；turn.intent_message_id=null |
| 8 | `awaiting_confirm_intent` | 用户点"取消" | `cancelled` | intent.status=`cancelled` |
| 9 | `executing` | executor 推 form card | `awaiting_commit` | 创建 trace + form card（status=`prefilled`） |
| 10 | `executing` | executor 推 disambig card | `awaiting_select_candidate` | 创建 trace + disambig card（status=`awaiting_select`） |
| 11 | `executing` | executor 推 query_result（无后续 intent） | `done` | 创建 trace + query_result card；turn.summary 更新 |
| 12 | `executing` | executor 推 query_result（有后续 intent） | `executing`（继续） | 同上 |
| 13 | `executing` | executor 失败 | `failed` | trace.status=failed |
| 14 | `awaiting_commit` | 提交成功（最后一个 intent） | `done` | form.status=`submitted`；创建 followup text msg；turn.summary/intents_done 更新 |
| 15 | `awaiting_commit` | 提交成功（仍有 intent） | `executing` | 同上，继续下一个 intent |
| 16 | `awaiting_commit` | 提交失败（后端校验） | `executing` | form.status=`failed`；保留字段值供重试 |
| 17 | `awaiting_commit` | 用户取消 | `cancelled` | form.status=`discarded` |
| 18 | `awaiting_select_candidate` | 用户选候选 | `executing` | disambig.status=`selected`；注入 bizKey 继续执行 |
| 19 | `awaiting_select_candidate` | 用户取消 | `cancelled` | disambig.status=`discarded` |
| **20** | **任何非终态** | **用户发新指令（free_text）** | **`superseded`** | **关联 message 状态变化：intent→cancelled / form→discarded / disambig→discarded** |
| 21 | `planning` / `executing` | cron 1 小时无活动 | `failed` | 孤儿 turn 清理；trace.status=failed |
| 22 | `awaiting_*`（非终态） | cron 24 小时无活动 | `superseded` | 长时间等用户操作，自动放弃 |

---

## 4. Message 状态机（按 type 多态）

### 4.1 简单消息（无状态迁移）

| type | role | status |
|------|------|--------|
| text | user | `sent`（始终） |
| text | ai | `sent` |
| text | system | `sent` |
| card (query_result) | ai | `sent` |
| card (fallback) | ai | `sent` |

→ 这些消息**创建即终态**，无状态迁移。

### 4.2 trace 状态机

```
   (创建)
      │
      ▼
┌───────────┐    流式完成    ┌──────┐
│ streaming │ ────────────▶ │ done │
└─────┬─────┘                └──────┘
      │
      │ 流式失败/中断
      ▼
┌──────────┐
│  failed  │
└──────────┘
```

### 4.3 intent 消息状态机

```
                   (Planner 推送)
                         │
                         ▼
                ┌─────────────────────┐
                │ awaiting_confirm    │◀──── clarify 后新版本也从此开始
                │ (含或不含 missingInfo)│
                └──────────┬──────────┘
                           │
            ┌──────────────┼──────────────┬──────────────┐
            │              │              │              │
       用户回答        用户"理解正确"   用户"我要调整"  用户"取消"/
       (clarify)                        (新指令)        
            │              │              │              │
            ▼              ▼              ▼              ▼
   (保留 awaiting_    ┌──────────┐  ┌──────────┐   ┌──────────┐
    confirm，被      │ confirmed│  │ adjusted │   │cancelled │
    新 msg 替代)     └──────────┘  └──────────┘   └──────────┘
```

**注**：clarify 流程中，原 intent msg 保留 `awaiting_confirm` 状态，turn.intent_message_id 指向新版本。前端通过 turn.intent_message_id 判断哪个是当前活跃的。

### 4.4 form card 状态机（最复杂）

```
                    (Executor 推送)
                          │
                          ▼
                   ┌─────────────┐
                   │ prefilled   │◀──── 用户编辑后回到此态
                   └──────┬──────┘
                          │
                          ├──── 用户编辑字段 ───▶ editing
                          │
                          ├──── 权限校验失败 ───▶ permission（终态）
                          │
                          ├──── available-transitions 预校验失败 ──▶ validation
                          │
                          ▼
                   ┌─────────────┐
       用户点提交 ─▶│ submitting  │
                   └──────┬──────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
         后端成功    后端校验失败   用户取消/Turn superseded
              │           │           │
              ▼           ▼           ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │submitted │ │  failed  │ │discarded │
        └──────────┘ └─────┬────┘ └──────────┘
                           │
                           │ 用户重试 → 回到 submitting
                           ▼
                     (form 保持 failed 状态，
                      用户重试时再切 submitting)
```

### 4.5 disambig card 状态机

```
            (Executor 推送)
                  │
                  ▼
          ┌─────────────────┐
          │ awaiting_select │
          └────────┬────────┘
                   │
            ┌──────┼──────┐
            │             │
       用户选定       用户取消/Turn superseded
            │             │
            ▼             ▼
      ┌──────────┐  ┌──────────┐
      │ selected │  │discarded │
      └──────────┘  └──────────┘
```

---

## 5. 三层状态协调

### 协调原则

```
Session.status = active
    │
    └── Turn.status: planning → awaiting_confirm_intent → executing → awaiting_commit → done
        │
        └── Message.status:
            - user msg: sent
            - trace: streaming → done
            - intent card: awaiting_confirm → confirmed
            - form card: prefilled → submitted
```

**协调规则**：
- 三层状态独立但相关
- Session 状态宏观（活跃/归档）
- Turn 状态中观（流程进度）—— **是协调信源**
- Message 状态微观（单条消息生命周期）
- 状态变化时各层独立 UPDATE，事务保证一致性

### 协调示例

| 场景 | Turn.status | 关联 Message.status |
|------|------------|-------------------|
| Planner 流式中 | `planning` | trace: streaming |
| 等"理解正确" | `awaiting_confirm_intent` | intent: awaiting_confirm；trace: done |
| 执行 executor 中 | `executing` | trace: streaming（intent_id=intent_1） |
| 等表单提交 | `awaiting_commit` | form: prefilled；trace: done |
| 提交成功（最后 intent） | `done` | form: submitted；followup text: sent |
| 用户取消 | `cancelled` | intent/form: cancelled/discarded |
| 用户发新指令 | `superseded` | 关联 message: cancelled/discarded |

---

## 6. 并发控制（基于 Turn.status 的拦截）

### 拦截规则

**核心规则**：用 `Session.current_turn_id` + `Turn.status` 联合判断。

```go
func CanAcceptNewMessage(sessionID string) error {
    sess := sessionRepo.Get(sessionID)
    if sess.CurrentTurnID == "" {
        return nil  // 无当前 turn，允许
    }
    
    turn := turnRepo.Get(sess.CurrentTurnID)
    
    // LLM 实际调用中 → 拦截
    if turn.Status == "planning" || turn.Status == "executing" {
        return ErrTurnInFlight
    }
    
    // 终态或等用户操作 → 允许（宽松策略，支持多卡片共存）
    return nil
}
```

### 拦截矩阵（完整）

| 当前 turn 状态 | free_text | answer_clarify | confirm_intent | select_candidate | commit_card | cancel |
|--------------|-----------|----------------|----------------|------------------|-------------|--------|
| `planning` | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ✅ 允许 |
| `awaiting_clarify` | ⚠️ 允许（前 turn→superseded） | ✅ 允许 | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ✅ 允许 |
| `awaiting_confirm_intent` | ⚠️ 允许（前 turn→superseded） | ❌ 拦截 | ✅ 允许 | ❌ 拦截 | ❌ 拦截 | ✅ 允许 |
| `executing` | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ✅ 允许 |
| `awaiting_commit` | ⚠️ 允许（前 turn→superseded） | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ✅ 允许 | ✅ 允许 |
| `awaiting_select_candidate` | ⚠️ 允许（前 turn→superseded） | ❌ 拦截 | ❌ 拦截 | ✅ 允许 | ❌ 拦截 | ✅ 允许 |
| `done` / `cancelled` / `superseded` / `failed` | ✅ 允许 | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 | ❌ 拦截 |

**通用规则**：
- `cancel` 始终允许（用户可随时终止）
- 各 action_type 只在对应 awaiting_* 状态允许
- `free_text` 在非 LLM 处理中状态都允许（多卡片共存）

### Superseded 触发场景

用户在以下状态发新指令（free_text）时，前一个 turn 标记为 `superseded`：

```go
func (h *CopilotHandler) handleFreeText(c *gin.Context, req MessageRequest) {
    sess, _ := h.sessionRepo.Get(c, c.Param("id"))
    
    // 如果有未完成的 turn，标记为 superseded
    if sess.CurrentTurnID != "" {
        prevTurn, _ := h.turnRepo.Get(c, sess.CurrentTurnID)
        if !isTerminalStatus(prevTurn.Status) {
            // 事务：标记前 turn + 关联 message
            h.txManager.WithinTx(c, func(tx *gorm.DB) error {
                h.turnRepo.UpdateStatus(c, prevTurn.BizKey, TurnStatusSuperseded)
                h.markRelatedMessagesSuperseded(c, prevTurn.BizKey)
                return nil
            })
        }
    }
    
    // 创建新 turn（同流程 A）
    // ...
}

func markRelatedMessagesSuperseded(ctx, turnID) {
    msgs := msgRepo.ListByTurn(turnID)
    for _, m := range msgs {
        switch {
        case m.Type == "intent" && m.Status == "awaiting_confirm":
            msgRepo.UpdateStatus(m.BizKey, MsgStatusIntentCancelled)
        case m.Type == "card" && m.CardType == "form" && m.Status == "prefilled":
            msgRepo.UpdateStatus(m.BizKey, MsgStatusFormDiscarded)
        case m.Type == "card" && m.CardType == "disambig" && m.Status == "awaiting_select":
            msgRepo.UpdateStatus(m.BizKey, MsgStatusDisambigDiscarded)
        }
    }
}
```

### 双层防护

#### 第一层：前端 UI 禁用

UI design 已设计（Component 2 输入区双模式）：
- 思考态：输入框 + 发送按钮同时 disabled
- inline 提示："AI 处理中，请稍候…"

前端根据 SSE 流状态判断（有流 = 处理中）。

#### 第二层：后端中间件兜底

防止前端 bug 或恶意请求：

```go
// internal/copilot/handler/middleware.go

func (h *CopilotHandler) TurnInFlightGuard(c *gin.Context) {
    var req MessageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Next()
        return
    }
    
    // 只对会启动新 LLM 流的请求类型拦截
    if req.Type != "free_text" {
        c.Next()  // confirm/select/commit/cancel/answer_clarify 不拦截
        return
    }
    
    sess, err := h.sessionRepo.Get(c, c.Param("id"))
    if err != nil || sess.CurrentTurnID == "" {
        c.Next()
        return
    }
    
    turn, err := h.turnRepo.Get(c, sess.CurrentTurnID)
    if err != nil {
        c.Next()
        return
    }
    
    if turn.Status == "planning" || turn.Status == "executing" {
        c.JSON(http.StatusConflict, gin.H{
            "code":    "ERR_COPILOT_TURN_IN_FLIGHT",
            "message": "当前消息正在处理中，请稍候",
            "turnId":  turn.BizKey,
            "status":  turn.Status,
        })
        c.Abort()
        return
    }
    
    c.Next()
}
```

---

## 7. 核心流程时序

### 流程 A：单意图写操作（最常见）

```
用户输入："创建 P1 事项叫认证模块"

请求 1: POST /messages {type: free_text}
─────────────────────────────────────────
T=0   Session.status = active
T=0   创建 Turn (status=planning, started_at=now)
T=0   创建 Message: user/text/sent (seq=1)
T=0   persist user msg
T=0.1 启动 Planner goroutine
T=0.2 SSE event: input_rewrite
T=0.5 SSE event: thinking
T=0.5 persist Message: ai/trace/streaming (seq=2)
T=0.8 Planner 输出 plan（无 missing_info）
T=0.8 UPDATE trace.status=done
T=0.8 persist Message: ai/card/intent/awaiting_confirm (seq=3, cardData={intents:[writer...]})
T=0.9 UPDATE Turn: status=awaiting_confirm_intent, intent_message_id=msg_003
T=1.0 SSE event: turn_phase_done (nextAction=await_confirm_intent)

请求 2: POST /messages {type: confirm_intent, intentMessageId: msg_003}
─────────────────────────────────────────
T=2.0 读取 turn (单行查询)
T=2.0 读取 intent msg 的 payload
T=2.1 UPDATE Turn: status=executing, confirmed_intent={...}
T=2.1 UPDATE Message (intent, msg_003): status=confirmed
T=2.2 启动 writer executor
T=2.3 SSE event: step_started (intent_1)
T=2.5 SSE event: thinking
T=2.5 persist Message: ai/trace/streaming (seq=4, intent_id=intent_1)
T=2.8 SSE event: tool_call fuzzy_match_member
T=3.0 SSE event: tool_result success
T=3.2 persist Message: ai/card/form/prefilled (seq=5, intent_id=intent_1)
T=3.3 UPDATE trace.status=done
T=3.3 UPDATE Turn: status=awaiting_commit
T=3.4 SSE event: turn_phase_done (nextAction=await_commit)

请求 3: POST /messages {type: commit_card, messageId: msg_005}
─────────────────────────────────────────
T=4.0 读取 form card payload
T=4.1 权限校验
T=4.2 调 MainItemService.Create（不走 LLM）
T=4.5 后端返回成功
T=4.6 UPDATE Message (form, msg_005): status=submitted
T=4.6 persist Message: ai/text/sent (seq=6, followup="已为你创建...")
T=4.7 UPDATE Turn: status=done, intents_done=1, summary="创建了 MainItem「认证模块」", completed_at=now
T=4.8 JSON 响应
```

### 流程 B：主动澄清 + 多轮 Q&A

```
请求 1: POST /messages {type: free_text, content: "创建一个 P1 事项"}
─────────────────────────────────────────
T=0   创建 Turn (status=planning)
T=0   persist user msg (seq=1)
T=0.8 Planner 输出（含 missing_info=[title]）
T=0.8 persist trace (seq=2)
T=0.8 persist Message: intent/awaiting_confirm (seq=3, missingInfo=[{title}])
T=0.9 UPDATE Turn: status=awaiting_clarify, intent_message_id=msg_003

请求 2: POST /messages {type: answer_clarify, intentMessageId: msg_003, answer: "认证模块"}
─────────────────────────────────────────
T=2.0 persist user msg (seq=4, content="认证模块")
T=2.0 UPDATE Turn: status=planning
T=2.1 重新调 Planner（注入历史 + 回答）
T=2.5 Planner 输出新 plan（无 missing_info）
T=2.5 persist trace (seq=5)
T=2.5 persist Message: intent/awaiting_confirm (seq=6, 完整字段)
T=2.6 UPDATE Turn: status=awaiting_confirm_intent, intent_message_id=msg_006
（原 msg_003 保持 awaiting_confirm 状态，但不再活跃）

请求 3: POST /messages {type: confirm_intent, intentMessageId: msg_006}
─────────────────────────────────────────
→ 进入执行流程，同流程 A 请求 2
```

### 流程 C：多意图中断

```
请求 1: POST /messages {type: free_text, content: "创建 P1 事项 + 查我的 P0"}
─────────────────────────────────────────
T=0   创建 Turn (status=planning, intents_total=2)
T=0.8 Planner 输出 2 个 intent
T=0.9 persist Message: intent/awaiting_confirm (含 2 intents)
T=1.0 UPDATE Turn: status=awaiting_confirm_intent

请求 2: POST /messages {type: confirm_intent, intentMessageId: msg_xxx}
─────────────────────────────────────────
T=2.0 UPDATE Turn: status=executing, confirmed_intent={2 intents}
T=2.1 执行 intent_1 (writer)
T=3.0 persist trace (intent_1)
T=3.2 persist form card (intent_1, prefilled)
T=3.3 UPDATE Turn: status=awaiting_commit
（intent_2 reader **未执行**，仍待用户操作后处理）

请求 3: POST /messages {type: commit_card, messageId: msg_form}
─────────────────────────────────────────
T=4.0 调 entity service，成功
T=4.5 UPDATE form.status=submitted
T=4.6 persist followup text
T=4.7 UPDATE Turn: status=done, intents_done=1
       summary="创建了 MainItem「认证模块」；intent_2（查询 P0）未执行"

→ 用户如想查 P0，需重新发指令（新 turn）
```

### 流程 D：歧义消解

```
请求 1: POST /messages {type: free_text, content: "把认证模块改成已完成"}
─────────────────────────────────────────
T=0   创建 Turn (status=planning)
T=1.0 UPDATE Turn: status=awaiting_confirm_intent

请求 2: POST /messages {type: confirm_intent}
─────────────────────────────────────────
T=2.0 UPDATE Turn: status=executing
T=2.5 updater executor 流式
T=2.8 tool_call query_entities("认证模块")
T=3.0 tool_result 返回 3 个候选
T=3.2 persist disambig card (awaiting_select, candidates=[...])
T=3.3 UPDATE Turn: status=awaiting_select_candidate

请求 3: POST /messages {type: select_candidate, messageId: msg_disambig, candidateBizKey: "MI-0023"}
─────────────────────────────────────────
T=4.0 UPDATE disambig.status=selected
T=4.0 UPDATE Turn: status=executing（继续）
T=4.1 updater 继续执行（注入 target_bizkey=MI-0023）
T=4.3 tool_call validate_transition(MI-0023, "completed")
T=4.5 tool_result 失败（validTransitions=[paused, cancelled]）
T=4.7 persist form card (status=validation, errors.validTransitions=[paused, cancelled])
T=4.8 UPDATE Turn: status=awaiting_commit（用户修改目标状态后提交）
```

### 流程 E：Superseded（用户发新指令）

```
请求 1: POST /messages {type: free_text, content: "创建 P1 事项"}
─────────────────────────────────────────
T=0   创建 Turn_001 (status=planning)
T=1.0 UPDATE Turn_001: status=awaiting_confirm_intent
T=1.0 persist intent msg_002 (awaiting_confirm)

请求 2 (用户改变主意): POST /messages {type: free_text, content: "查我的 P0 事项"}
─────────────────────────────────────────
T=2.0 中间件检查：Turn_001.status = awaiting_confirm_intent（非 planning/executing）→ 允许
T=2.1 事务：
       - UPDATE Turn_001: status=superseded, completed_at=now
       - UPDATE msg_002 (intent): status=cancelled
       - 创建 Turn_002 (status=planning)
       - persist user msg (seq=N+1, turn_id=Turn_002)
T=2.2 启动新 Planner（针对 Turn_002）
...

→ Turn_001 最终状态：superseded（审计可见，与 cancelled 区分）
→ Turn_002 正常进行
```

### 流程 F：流式中断恢复

```
请求 1: POST /messages {type: free_text}
─────────────────────────────────────────
T=0   创建 Turn (status=planning)
T=0.5 SSE event: thinking
T=0.5 persist trace (streaming)
T=2.0 网络断开
T=2.0 trace.status 流中保持 streaming
T=2.0 goroutine 退出（ctx.Done）
T=2.0 流关闭，前端看到 SSE 提前结束

T=10  cron 检查孤儿 turn
T=10  发现 turn.status=planning 且 last_active_at 超过 1 小时
T=10  UPDATE Turn: status=failed
T=10  UPDATE trace.status=failed
T=10  （可选）persist fallback card

用户回看：看到失败 turn + 半截 trace
用户重新发指令：新 turn 开始
```

---

## 8. 关键不变量

1. **Turn.status 是流程协调的唯一信源**——刷新页面后前端读 turn.status 决定 UI 模式
2. **三层状态独立但协调**——Session 宏观、Turn 中观、Message 微观，各自状态机
3. **事务保证一致性**——Turn + Message 状态变更在同一个事务内
4. **状态变化通过 UPDATE 而非 INSERT**——避免状态分裂（如意图消息的 confirmed/adjusted/cancelled 通过 PATCH 同一行）
5. **失败状态不阻塞**——form=failed 时用户可重试，状态在 submitting/failed 间循环
6. **终态不可恢复**——done/cancelled/superseded/failed 的 Turn 不再变化；submitted/discarded 的 Message 不再变化
7. **cancel 始终允许**——用户可随时终止当前流程
8. **多卡片共存**——同会话可同时存在多个 form card（不同 turn），互不干扰

---

## 9. 状态变化与 UI 反馈

### 前端根据 Turn.status 切换 UI 模式

| Turn.status | 输入区模式 | 默认焦点 |
|------------|----------|---------|
| `planning` | 禁用 + "AI 思考中..." 提示 | — |
| `awaiting_clarify` | 文本模式（等用户回答） | textarea |
| `awaiting_confirm_intent` | 选项组（理解正确/调整/取消） | 第一个选项 |
| `executing` | 禁用 + "AI 执行中..." 提示 | — |
| `awaiting_commit` | 默认文本模式（可发新指令） | textarea（但不强制） |
| `awaiting_select_candidate` | 默认文本模式（可发新指令） | textarea（但不强制） |
| `done` / 终态 | 默认文本模式 | textarea |

### 刷新页面恢复

```
GET /api/v1/copilot/sessions/:id
→ 返回 current_turn_id + 当前 turn.status

前端根据 turn.status 切换输入区模式
→ 用户能继续未完成的流程（如点确认、提交、选候选）
```

---

## 10. 状态枚举汇总

### Turn 状态（10 个）

```go
type TurnStatus string
const (
    TurnStatusPlanning               TurnStatus = "planning"
    TurnStatusAwaitingClarify        TurnStatus = "awaiting_clarify"
    TurnStatusAwaitingConfirmIntent  TurnStatus = "awaiting_confirm_intent"
    TurnStatusExecuting              TurnStatus = "executing"
    TurnStatusAwaitingCommit         TurnStatus = "awaiting_commit"
    TurnStatusAwaitingSelectCandidate TurnStatus = "awaiting_select_candidate"
    TurnStatusDone                   TurnStatus = "done"
    TurnStatusCancelled              TurnStatus = "cancelled"
    TurnStatusSuperseded             TurnStatus = "superseded"
    TurnStatusFailed                 TurnStatus = "failed"
)

func isTerminalStatus(s TurnStatus) bool {
    return s == TurnStatusDone || s == TurnStatusCancelled || 
           s == TurnStatusSuperseded || s == TurnStatusFailed
}

func isLLMInFlightStatus(s TurnStatus) bool {
    return s == TurnStatusPlanning || s == TurnStatusExecuting
}
```

### Session 状态（3 个）

```go
type SessionStatus string
const (
    SessionStatusActive   SessionStatus = "active"
    SessionStatusArchived SessionStatus = "archived"
    SessionStatusExpired  SessionStatus = "expired"
)
```

### Message 状态（按 type 多态）

```go
type MsgStatus string
const (
    // 通用
    MsgStatusSent       MsgStatus = "sent"
    
    // trace
    MsgStatusStreaming  MsgStatus = "streaming"
    MsgStatusTraceDone  MsgStatus = "done"
    MsgStatusTraceFailed MsgStatus = "failed"
    
    // intent
    MsgStatusIntentAwaiting  MsgStatus = "awaiting_confirm"
    MsgStatusIntentComplete  MsgStatus = "info_complete"
    MsgStatusIntentConfirmed MsgStatus = "confirmed"
    MsgStatusIntentAdjusted  MsgStatus = "adjusted"
    MsgStatusIntentCancelled MsgStatus = "cancelled"
    
    // form card
    MsgStatusFormPrefilled   MsgStatus = "prefilled"
    MsgStatusFormEditing     MsgStatus = "editing"
    MsgStatusFormValidation  MsgStatus = "validation"
    MsgStatusFormSubmitting  MsgStatus = "submitting"
    MsgStatusFormSubmitted   MsgStatus = "submitted"
    MsgStatusFormFailed      MsgStatus = "failed"
    MsgStatusFormDiscarded   MsgStatus = "discarded"
    MsgStatusFormPermission  MsgStatus = "permission"
    
    // disambig card
    MsgStatusDisambigAwaiting MsgStatus = "awaiting_select"
    MsgStatusDisambigSelected MsgStatus = "selected"
    MsgStatusDisambigDiscarded MsgStatus = "discarded"
)
```
