# ER Diagram: AI Copilot 对话助手

> 3-tier data model: Session → Turn → Message
>
> 返回 [`tech-design.md`](./tech-design.md)

## 概览

```
┌─────────────────────┐
│ copilot_sessions    │  会话元数据（用户级）
│ status: active/     │
│   archived/expired  │
└──────────┬──────────┘
           │ 1
           │
           │ N
┌──────────┴──────────┐       ┌──────────────────────────┐
│ copilot_turns       │ 1   N │ copilot_messages         │
│ status: planning/   │──────▶│ status: 多态（按 type） │
│   awaiting_confirm/ │       │                          │
│   executing/done/...│       └──────────────────────────┘
│ summary: 摘要       │
│ confirmed_intent    │
└─────────────────────┘

┌─────────────────────┐
│copilot_agent_call_  │
│       logs          │  Agent 调用日志（配额+成本+调试）
│ status: success/    │
│   failed/timeout    │
└─────────────────────┘

┌─────────────────────┐
│   feature_flags     │  灰度/熔断开关
│ (独立，无外键)       │
└─────────────────────┘
```

## 三级数据模型

### 设计原则

1. **Session → Turn → Message 三级层级**——自然映射"用户操作 → 完整周期 → 单条产物"
2. **三层都有独立状态字段**——各自状态机，协调但独立。DB 列名按实体前缀避免 MySQL 关键字（`status`/`role`/`type`）：`session_status` / `turn_status` / `msg_status`（messages 另有 `msg_role` / `msg_type`） / `log_status`（agent_call_logs） / `idem_status`（idempotency_keys）。本文概念性 `X.status` 指对应状态字段
3. **消息单一来源**——所有对话内容（含用户原文）只在 messages 表存储，turns 表不冗余 user_message
4. **Turn 嵌入 summary**——合并原 turn_summaries 表，加速重建

## 实体属性

### copilot_sessions — 会话主表

```
┌─────────────────────────────────────────────────────┐
│ copilot_sessions                                    │
├─────────────────────────────────────────────────────┤
│ PK  id                  BIGINT                      │
│ UQ  biz_key             VARCHAR(36)                 │
│ IDX user_id              BIGINT (→ users.id)        │
│ IDX team_id              BIGINT (→ teams.id, NULL)  │
│     team_name            VARCHAR(100)               │
│     session_title       VARCHAR(100)               │
│ IDX current_turn_id     VARCHAR(36)                 │
│ ★  session_status       VARCHAR(32) DEFAULT 'active'│
│ IDX last_active_at      TIMESTAMP                   │
│ IDX expires_at          TIMESTAMP                   │
│     created_at          TIMESTAMP                   │
│     updated_at          TIMESTAMP                   │
│     deleted_at          TIMESTAMP (NULL)            │
└─────────────────────────────────────────────────────┘
```

**Status 状态机**：
- `active`：活跃会话（默认）
- `archived`：用户归档（不再显示在主列表，但可查）
- `expired`：已过期（超过 30 天，cron 标记）

### copilot_turns — Turn 主表（新增）

```
┌─────────────────────────────────────────────────────┐
│ copilot_turns                                       │
├─────────────────────────────────────────────────────┤
│ PK  id                  BIGINT                      │
│ UQ  biz_key             VARCHAR(36)                 │
│ IDX session_id           VARCHAR(36) (→ sessions)   │
│ IDX user_biz_key         VARCHAR(36)                │
│                                                      │
│ ★  turn_status          VARCHAR(32) DEFAULT 'planning'│
│                                                      │
│     user_query_short    VARCHAR(200)                │
│     turn_summary        VARCHAR(200)                │
│     intents_total       INT                         │
│     intents_done        INT                         │
│                                                      │
│     intent_message_id   VARCHAR(36) (NULL)          │
│     confirmed_intent    JSON (NULL)                 │
│                                                      │
│     started_at          TIMESTAMP                   │
│     last_active_at      TIMESTAMP                   │
│     completed_at        TIMESTAMP (NULL)            │
└─────────────────────────────────────────────────────┘
复合索引: (session_id, started_at), (user_biz_key, started_at)
```

**Status 状态机**（10 个状态）：

```
planning                    ← 用户发指令，Planner 流式中
    ↓
awaiting_confirm_intent     ← 意图消息已推送，等用户"理解正确"
    ├─ confirm → executing
    ├─ adjust → planning（重新调 Planner）
    ├─ cancel → cancelled
    └─ 用户新指令 → superseded
    ↓
awaiting_clarify            ← 意图含 missing_info，等用户回答
    ↓ (回答)
    planning                ← 重新调 Planner
    ↓
executing                   ← 按 plan.intents 执行
    ↓ (遇到 form card)
    awaiting_commit         ← 等用户提交
    ↓ (提交成功)
    executing               ← 继续下一个 intent（如有）
    ↓ (所有 intent 完成)
    done
    ↓
cancelled / superseded / failed   ← 终态（4 个）
```

**终态对比**：
- `done`：正常完成
- `cancelled`：用户主动取消
- `superseded`：用户发了新指令，原 turn 被替代
- `failed`：系统错误

详细状态机与转换矩阵见 [`state-machines.md`](./state-machines.md)。

**关键设计点**：
- **不存 user_message**——用户原文由 messages 表存（type=text, role=user）
- **user_query_short**：截断的简短摘要（用于会话列表显示，类似 Session.title）
- **summary**：规则提取的 outcome（"创建了 MainItem「认证模块」"）
- **confirmed_intent**：用户确认后从意图消息提取并冗余存储（加速重建）

### copilot_messages — 消息表

```
┌─────────────────────────────────────────────────────┐
│ copilot_messages                                    │
├─────────────────────────────────────────────────────┤
│ PK  id                  BIGINT                      │
│ UQ  biz_key             VARCHAR(36)                 │
│ IDX session_id           VARCHAR(36) (→ sessions)   │
│ FK  turn_id              VARCHAR(36) (→ turns)      │
│ IDX intent_id            VARCHAR(36) (NULL)         │
│     msg_seq              INT                        │
│     msg_role             VARCHAR(16)                │
│     msg_type             VARCHAR(16)                │
│ ★  msg_status           VARCHAR(32) DEFAULT 'sent'  │
│     msg_content          TEXT (NULL)                │
│     msg_trace            JSON (NULL)                │
│     card_type            VARCHAR(32) (NULL)         │
│     msg_card             JSON (NULL)                │
│     intent_meta          JSON (NULL)                │
│     created_at           TIMESTAMP                  │
│     updated_at           TIMESTAMP                  │
│     deleted_at           TIMESTAMP (NULL)           │
└─────────────────────────────────────────────────────┘
复合索引: (session_id, turn_id, msg_seq)
```

**msg_status 多态枚举（按 msg_type 解释）**：

| msg_type | msg_role | msg_status 取值 |
|------|------|------------|
| text | user | `sent`（始终） |
| text | ai | `sent` |
| text | system | `sent` |
| trace | ai | `streaming` / `done` / `failed` |
| intent | ai | `awaiting_confirm` / `info_complete` / `confirmed` / `adjusted` / `cancelled` |
| card (form) | ai | `prefilled` / `editing` / `validation` / `submitting` / `submitted` / `failed` / `discarded` / `permission` |
| card (query_result) | ai | `sent` |
| card (disambig) | ai | `awaiting_select` / `selected` / `discarded` |
| card (candidate_list) | ai | `awaiting_select` / `selected` / `discarded` |
| card (fallback) | ai | `sent` |

**关键设计点**：
- **统一 status 字段**——替代之前的 `card_state`，简化模型
- **status 多态**——按 type 解释语义，避免字段爆炸
- **type 多态**——text / trace / card / intent 共用一张表
- **turn_id 加外键**——指向 copilot_turns，级联删除

### copilot_agent_call_logs — Agent 调用日志

```
┌─────────────────────────────────────────────────────┐
│ copilot_agent_call_logs                             │
├─────────────────────────────────────────────────────┤
│ PK  id                  BIGINT                      │
│ UQ  biz_key             VARCHAR(36)                 │
│ IDX session_id           VARCHAR(36) (→ sessions)   │
│ IDX turn_id              VARCHAR(36) (→ turns)      │
│ IDX step_id              VARCHAR(36) (NULL)         │
│ IDX intent_id            VARCHAR(36) (NULL)         │
│ IDX user_biz_key         VARCHAR(36)                │
│     agent_role           VARCHAR(32)                │
│     llm_provider         VARCHAR(32)                │
│     llm_model            VARCHAR(64)                │
│     input_tokens         INT                        │
│     output_tokens        INT                        │
│     duration_ms          INT                        │
│     cost_usd             DECIMAL(10,4)              │
│ ★  log_status            VARCHAR(16)                │
│     error_message        TEXT (NULL)                │
│     input_rewrite_payload JSON (NULL)               │
│     created_at           TIMESTAMP                  │
└─────────────────────────────────────────────────────┘
复合索引: (user_biz_key, created_at)  -- 配额检查
```

**Status 取值**：`success` / `failed` / `timeout`

### copilot_idempotency_keys — commit_card 幂等表

```
┌─────────────────────────────────────────────────────┐
│ copilot_idempotency_keys                            │
├─────────────────────────────────────────────────────┤
│ PK  id                  BIGINT                      │
│ UQ  request_id           VARCHAR(36)                │
│ IDX message_id           VARCHAR(36) (→ messages)   │
│     turn_id              VARCHAR(36) (→ turns)      │
│     session_id           VARCHAR(36) (→ sessions)   │
│     user_biz_key         VARCHAR(36)                │
│     result_biz_key       VARCHAR(36) (NULL)         │
│     idem_status          VARCHAR(16)                │
│     created_at           TIMESTAMP                  │
│     committed_at         TIMESTAMP (NULL)           │
└─────────────────────────────────────────────────────┘
复合索引: (request_id) UNIQUE
```

**Status 取值**：`pending` / `committed` / `failed`

**作用**：`commit_card` 幂等保护——同 `request_id` 重复提交只创建一次实体（见 [interfaces.md](./interfaces.md) §7.1）。命中 `committed` 直接返回 `result_biz_key`，不再调 entity service。

### feature_flags — Feature Flag 表

```
┌─────────────────────────────────────────────────────┐
│ feature_flags                                       │
├─────────────────────────────────────────────────────┤
│ PK  id                  BIGINT                      │
│     flag_key             VARCHAR(64)                │
│     flag_enabled         BOOLEAN                    │
│     scope_type           VARCHAR(32)                │
│     scope_id             VARCHAR(64)                │
│     flag_reason          VARCHAR(200)               │
│     created_at           TIMESTAMP                  │
│     updated_at           TIMESTAMP                  │
└─────────────────────────────────────────────────────┘
复合索引: (flag_key, scope_type, scope_id)
```

**作用域优先级**：user > team > global（最细匹配优先）

## 关系说明

### Session ↔ Turn（1:N）

- 一个 Session 有多个 Turn
- 通过 `session_id` 关联
- `copilot_sessions.current_turn_id` 指向当前活动 turn

### Turn ↔ Message（1:N）

- 一个 Turn 有多条 Message
- 通过 `turn_id` 外键关联
- 级联删除（删 turn 即删其所有 message）
- 按 `seq` 排序

### Turn ↔ AgentCallLog（1:N）

- 一个 Turn 有多次 Agent 调用
- Planner 调用：step_id = NULL
- Executor 调用：step_id = intent_id

### Session ↔ Message（间接，1:N via Turn）

- 通过 Turn 间接关联
- 查询 session 所有消息：JOIN turns 或子查询

### Message ↔ AgentCallLog（间接）

- Message 记录"对话产物"（持久化到 messages 表）
- AgentCallLog 记录"调用元数据"
- 两者通过 `turn_id` + `step_id` 间接关联，无外键约束

### Message ↔ IdempotencyKey（1:N）

- 一个 form card 消息（`type=card, cardType=form`）可对应多条幂等行（每条 `commit_card` 请求一行）
- 通过 `copilot_idempotency_keys.message_id` 关联到 `copilot_messages.biz_key`
- 无 DB 外键约束（仅逻辑关联）

### FeatureFlag（独立）

- 独立表，不与其他表关联
- 用于灰度发布和熔断

## 三层状态机协调

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

**协调原则**：
- 三层状态独立但相关
- Session 状态宏观（活跃/归档）
- Turn 状态中观（流程进度）
- Message 状态微观（单条消息生命周期）
- 状态变化时各层独立 UPDATE，事务保证一致性

## 索引策略

| 表 | 索引 | 用途 |
|----|------|------|
| copilot_sessions | `idx_user (user_id)` | 用户会话列表查询 |
| copilot_sessions | `idx_session_status (session_status)` | 按状态过滤（如只看 active） |
| copilot_sessions | `idx_expires (expires_at)` | cron 过期清理 |
| copilot_turns | `idx_session_started (session_id, started_at)` | 会话内 turn 时间线 |
| copilot_turns | `idx_user_started (user_biz_key, started_at)` | 用户级 turn 查询 |
| copilot_turns | `idx_turn_status (turn_status)` | 按状态过滤（如 awaiting_commit） |
| copilot_messages | `idx_session_turn_seq (session_id, turn_id, msg_seq)` | 按 turn 顺序查询消息 |
| copilot_messages | `idx_intent (intent_id)` | 按 intent 聚合查询 |
| copilot_messages | `idx_msg_status (msg_status)` | 按状态过滤（如 awaiting_confirm 的意图） |
| copilot_agent_call_logs | `idx_user_date (user_biz_key, created_at)` | 每日配额检查 |
| feature_flags | `idx_key_scope (flag_key, scope_type, scope_id)` | 精确匹配 |

## 软删除策略

| 表 | 策略 | 原因 |
|----|------|------|
| copilot_sessions | 软删除 | 用户主动删除会话（可恢复） |
| copilot_turns | 硬删除（跟随 session） | 跟随 session 级联 |
| copilot_messages | 软删除 | 用户删除单条消息（可恢复） |
| copilot_agent_call_logs | 硬删除 + 30 天 cron | 数据量大，定期清理 |
| feature_flags | 硬删除 | 管理员操作，配置项 |

## 数据保留

| 数据 | 保留期 | 清理方式 |
|------|-------|---------|
| 会话元数据 | 用户主动删除或 30 天过期 | 软删除 + cron 物理清理 |
| Turn 数据 | 跟随会话 | 跟随会话清理 |
| 消息内容 | 跟随会话 | 跟随会话清理 |
| Agent 调用日志 | 30 天 | cron 清理 |
| Feature flag 变更 | 永久 | 不清理（审计） |

## 与之前模型的对比

| 维度 | 之前（两级） | 现在（三级） |
|------|------------|------------|
| 表数量 | 5 | 5（合并 turn_summaries 到 turns） |
| 重建上下文 | 扫描 messages + 反推 | 单次查询 turn 行 |
| 状态持久化 | 仅 messages.card_state | 三层独立 status |
| 刷新页面恢复 | 依赖反推 | turn.status 直接告知 |
| API 列表 | 分页 messages，前端分组 | 分页 turns，天然层级 |
| user_message 存储 | messages | messages（统一） |
| Turn 摘要 | 独立 turn_summaries 表 | 嵌入 turns.summary |
