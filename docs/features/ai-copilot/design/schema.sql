-- AI Copilot Schema
-- Feature: ai-copilot
-- Created: 2026-06-26
--
-- 3-tier data model: Session → Turn → Message
-- 6 new tables (含 copilot_idempotency_keys). Does NOT modify any existing tables.
-- Run via internal/migration/runner.go auto-migrate.
--
-- Both SQLite and MySQL DDL kept in sync per project convention
-- (see backend/migrations/SQLite-schema.sql + MySql-schema.sql).

-- =============================================================================
-- copilot_sessions
-- =============================================================================
-- 会话元数据。一个用户可有多个会话。

-- SQLite
CREATE TABLE copilot_sessions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key         VARCHAR(36) NOT NULL UNIQUE,
    user_id         INTEGER NOT NULL,
    team_id         INTEGER,
    team_name       VARCHAR(100) NOT NULL DEFAULT '',
    session_title   VARCHAR(100) NOT NULL DEFAULT '',
    current_turn_id VARCHAR(36) NOT NULL DEFAULT '',
    session_status  VARCHAR(32) NOT NULL DEFAULT 'active',  -- active / archived / expired
    last_active_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE INDEX idx_copilot_sessions_user    ON copilot_sessions(user_id);
CREATE INDEX idx_copilot_sessions_team    ON copilot_sessions(team_id);
CREATE INDEX idx_copilot_sessions_turn    ON copilot_sessions(current_turn_id);
CREATE INDEX idx_copilot_sessions_session_status ON copilot_sessions(session_status);
CREATE INDEX idx_copilot_sessions_expires ON copilot_sessions(expires_at);
CREATE INDEX idx_copilot_sessions_deleted ON copilot_sessions(deleted_at);

-- MySQL
-- CREATE TABLE copilot_sessions (
--     id              BIGINT PRIMARY KEY AUTO_INCREMENT,
--     biz_key         VARCHAR(36) NOT NULL,
--     user_id         BIGINT NOT NULL,
--     team_id         BIGINT NULL,
--     team_name       VARCHAR(100) NOT NULL DEFAULT '',
--     session_title   VARCHAR(100) NOT NULL DEFAULT '',
--     current_turn_id VARCHAR(36) NOT NULL DEFAULT '',
--     session_status  VARCHAR(32) NOT NULL DEFAULT 'active',
--     last_active_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     expires_at      TIMESTAMP NOT NULL,
--     created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--     deleted_at      TIMESTAMP NULL,
--     UNIQUE KEY uk_biz_key (biz_key),
--     KEY idx_user (user_id),
--     KEY idx_team (team_id),
--     KEY idx_turn (current_turn_id),
--     KEY idx_session_status (session_status),
--     KEY idx_expires (expires_at),
--     KEY idx_deleted (deleted_at)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- copilot_turns
-- =============================================================================
-- 一个用户指令 = 一个 Turn。包含 turn 级状态、摘要、已确认意图。
-- 不存 user_message（统一由 messages 表存储）。

-- SQLite
CREATE TABLE copilot_turns (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key             VARCHAR(36) NOT NULL UNIQUE,
    session_id          VARCHAR(36) NOT NULL,
    user_biz_key        VARCHAR(36) NOT NULL,           -- 冗余，配额检查

    -- Turn 摘要（合并原 turn_summaries 表）
    user_query_short    VARCHAR(200) NOT NULL DEFAULT '',  -- 截断 16-80 字简短摘要
    turn_summary        VARCHAR(200) NOT NULL DEFAULT '',  -- 规则提取的 outcome
    intents_total       INTEGER NOT NULL DEFAULT 0,
    intents_done        INTEGER NOT NULL DEFAULT 0,

    -- 状态机
    turn_status         VARCHAR(32) NOT NULL DEFAULT 'planning',
    -- planning / awaiting_confirm_intent / awaiting_clarify
    -- / executing / awaiting_commit / awaiting_select_candidate
    -- / done / cancelled / superseded / failed

    -- 意图相关（加速重建）
    intent_message_id   VARCHAR(36),                  -- 关联的意图消息 biz_key
    confirmed_intent    TEXT,                          -- JSON，用户确认后填充

    -- 元数据
    started_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at        TIMESTAMP,

    FOREIGN KEY (session_id) REFERENCES copilot_sessions(biz_key) ON DELETE CASCADE
);

CREATE INDEX idx_copilot_turns_session   ON copilot_turns(session_id, started_at);
CREATE INDEX idx_copilot_turns_user      ON copilot_turns(user_biz_key, started_at);
CREATE INDEX idx_copilot_turns_turn_status   ON copilot_turns(turn_status);

-- MySQL
-- CREATE TABLE copilot_turns (
--     id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
--     biz_key             VARCHAR(36) NOT NULL,
--     session_id          VARCHAR(36) NOT NULL,
--     user_biz_key        VARCHAR(36) NOT NULL,
--     user_query_short    VARCHAR(200) NOT NULL DEFAULT '',
--     turn_summary        VARCHAR(200) NOT NULL DEFAULT '',
--     intents_total       INT NOT NULL DEFAULT 0,
--     intents_done        INT NOT NULL DEFAULT 0,
--     turn_status         VARCHAR(32) NOT NULL DEFAULT 'planning',
--     intent_message_id   VARCHAR(36) NULL,
--     confirmed_intent    JSON NULL,
--     started_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     last_active_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     completed_at        TIMESTAMP NULL,
--     UNIQUE KEY uk_biz_key (biz_key),
--     KEY idx_session_started (session_id, started_at),
--     KEY idx_user_started (user_biz_key, started_at),
--     KEY idx_turn_status (turn_status),
--     CONSTRAINT fk_copilot_turns_session FOREIGN KEY (session_id) REFERENCES copilot_sessions(biz_key) ON DELETE CASCADE
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- copilot_messages
-- =============================================================================
-- 消息存储。type 多态（text / trace / card / intent），status 按 type 解释。
-- 统一存储所有对话产物：user 原文、AI text、trace、card（含 intent / form / query_result / disambig / fallback）。

-- SQLite
CREATE TABLE copilot_messages (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key      VARCHAR(36) NOT NULL UNIQUE,
    session_id   VARCHAR(36) NOT NULL,
    turn_id      VARCHAR(36) NOT NULL,
    intent_id    VARCHAR(36),                          -- 同 turn 内的子分组（多意图场景）
    msg_seq      INTEGER NOT NULL,                     -- 会话内单调递增
    msg_role     VARCHAR(16) NOT NULL,                 -- user / ai / system
    msg_type     VARCHAR(16) NOT NULL,                 -- text / trace / card / intent
    msg_status   VARCHAR(32) NOT NULL DEFAULT 'sent',  -- 见下方状态枚举说明
    msg_content  TEXT,                                  -- type=text/system（含 intent.text 冗余）
    msg_trace    TEXT,                                  -- type=trace, JSON
    card_type    VARCHAR(32),                           -- type=card: intent/form/query_result/disambig/candidate_list/fallback
    msg_card     TEXT,                                  -- type=card, JSON (parsed by card_type)
    intent_meta  TEXT,                                  -- JSON: {label, seq, total}
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMP,

    FOREIGN KEY (turn_id) REFERENCES copilot_turns(biz_key) ON DELETE CASCADE
);

-- status 枚举（按 type 多态解释）：
--   type=text/user:     sent
--   type=text/ai:       sent
--   type=text/system:   sent
--   type=trace:         streaming / done / failed
--   type=intent:        awaiting_confirm / info_complete / confirmed / adjusted / cancelled
--   type=card (form):   prefilled / editing / validation / submitting / submitted / failed / discarded / permission
--   type=card (query_result): sent
--   type=card (disambig): awaiting_select / selected / discarded
--   type=card (fallback): sent

CREATE INDEX idx_copilot_messages_session_turn_seq ON copilot_messages(session_id, turn_id, seq);
CREATE INDEX idx_copilot_messages_intent           ON copilot_messages(intent_id);
CREATE INDEX idx_copilot_messages_msg_status       ON copilot_messages(msg_status);
CREATE INDEX idx_copilot_messages_deleted          ON copilot_messages(deleted_at);

-- MySQL
-- CREATE TABLE copilot_messages (
--     id           BIGINT PRIMARY KEY AUTO_INCREMENT,
--     biz_key      VARCHAR(36) NOT NULL,
--     session_id   VARCHAR(36) NOT NULL,
--     turn_id      VARCHAR(36) NOT NULL,
--     intent_id    VARCHAR(36) NULL,
--     msg_seq      INT NOT NULL,
--     msg_role     VARCHAR(16) NOT NULL,
--     msg_type     VARCHAR(16) NOT NULL,
--     msg_status   VARCHAR(32) NOT NULL DEFAULT 'sent',
--     msg_content  TEXT NULL,
--     msg_trace    JSON NULL,
--     card_type    VARCHAR(32) NULL,
--     msg_card     JSON NULL,
--     intent_meta  JSON NULL,
--     created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--     deleted_at   TIMESTAMP NULL,
--     UNIQUE KEY uk_biz_key (biz_key),
--     KEY idx_session_turn_seq (session_id, turn_id, seq),
--     KEY idx_intent (intent_id),
--     KEY idx_msg_status (msg_status),
--     KEY idx_deleted (deleted_at),
--     CONSTRAINT fk_copilot_messages_turn FOREIGN KEY (turn_id) REFERENCES copilot_turns(biz_key) ON DELETE CASCADE
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- copilot_idempotency_keys
-- =============================================================================
-- commit_card 请求幂等表。防 LLM 重试 / 前端网络抖动重试导致重复创建实体。
-- 客户端每次提交生成 requestId (UUID v4)；Dispatcher 在 entity service Create 前 INSERT 此表。

-- SQLite
CREATE TABLE copilot_idempotency_keys (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id      VARCHAR(36) NOT NULL UNIQUE,    -- 客户端生成的 UUID v4
    message_id      VARCHAR(36) NOT NULL,           -- 关联的 form card 消息
    turn_id         VARCHAR(36) NOT NULL,
    session_id      VARCHAR(36) NOT NULL,
    user_biz_key    VARCHAR(36) NOT NULL,
    result_biz_key  VARCHAR(36),                    -- entity service 返回的实体 bizKey（commit 成功后填）
    idem_status     VARCHAR(16) NOT NULL,           -- pending / committed / failed
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    committed_at    TIMESTAMP
);

CREATE UNIQUE INDEX idx_copilot_idempotency_request ON copilot_idempotency_keys(request_id);
CREATE INDEX idx_copilot_idempotency_message        ON copilot_idempotency_keys(message_id);

-- MySQL
-- CREATE TABLE copilot_idempotency_keys (
--     id              BIGINT PRIMARY KEY AUTO_INCREMENT,
--     request_id      VARCHAR(36) NOT NULL,
--     message_id      VARCHAR(36) NOT NULL,
--     turn_id         VARCHAR(36) NOT NULL,
--     session_id      VARCHAR(36) NOT NULL,
--     user_biz_key    VARCHAR(36) NOT NULL,
--     result_biz_key  VARCHAR(36) NULL,
--     idem_status     VARCHAR(16) NOT NULL,
--     created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     committed_at    TIMESTAMP NULL,
--     UNIQUE KEY uk_request (request_id),
--     KEY idx_message (message_id)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- copilot_agent_call_logs
-- =============================================================================
-- 每次 Agent 调用一条记录（Planner + Executors）。配额 + 成本 + 调试。

-- SQLite
CREATE TABLE copilot_agent_call_logs (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key               VARCHAR(36) NOT NULL UNIQUE,
    session_id            VARCHAR(36) NOT NULL,
    turn_id               VARCHAR(36) NOT NULL,
    step_id               VARCHAR(36),                  -- NULL = planner 调用
    intent_id             VARCHAR(36),
    user_biz_key          VARCHAR(36) NOT NULL,         -- 冗余：配额检查不需 JOIN
    agent_role            VARCHAR(32) NOT NULL,         -- planner / writer / reader / updater / mover
    llm_provider          VARCHAR(32) NOT NULL,         -- glm / deepseek / openai
    llm_model             VARCHAR(64) NOT NULL,
    input_tokens          INTEGER NOT NULL DEFAULT 0,
    output_tokens         INTEGER NOT NULL DEFAULT 0,
    duration_ms           INTEGER NOT NULL DEFAULT 0,
    cost_usd              DECIMAL(10,4) NOT NULL DEFAULT 0,
    log_status            VARCHAR(16) NOT NULL,         -- success / failed / timeout
    error_message         TEXT,
    input_rewrite_payload TEXT,                          -- JSON, planner only
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_copilot_agent_call_logs_user_date ON copilot_agent_call_logs(user_biz_key, created_at);
CREATE INDEX idx_copilot_agent_call_logs_session   ON copilot_agent_call_logs(session_id);
CREATE INDEX idx_copilot_agent_call_logs_turn      ON copilot_agent_call_logs(turn_id);
CREATE INDEX idx_copilot_agent_call_logs_created   ON copilot_agent_call_logs(created_at);

-- MySQL
-- CREATE TABLE copilot_agent_call_logs (
--     id                    BIGINT PRIMARY KEY AUTO_INCREMENT,
--     biz_key               VARCHAR(36) NOT NULL,
--     session_id            VARCHAR(36) NOT NULL,
--     turn_id               VARCHAR(36) NOT NULL,
--     step_id               VARCHAR(36) NULL,
--     intent_id             VARCHAR(36) NULL,
--     user_biz_key          VARCHAR(36) NOT NULL,
--     agent_role            VARCHAR(32) NOT NULL,
--     llm_provider          VARCHAR(32) NOT NULL,
--     llm_model             VARCHAR(64) NOT NULL,
--     input_tokens          INT NOT NULL DEFAULT 0,
--     output_tokens         INT NOT NULL DEFAULT 0,
--     duration_ms           INT NOT NULL DEFAULT 0,
--     cost_usd              DECIMAL(10,4) NOT NULL DEFAULT 0,
--     log_status            VARCHAR(16) NOT NULL,
--     error_message         TEXT NULL,
--     input_rewrite_payload JSON NULL,
--     created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     UNIQUE KEY uk_biz_key (biz_key),
--     KEY idx_user_date (user_biz_key, created_at),
--     KEY idx_session (session_id),
--     KEY idx_turn (turn_id),
--     KEY idx_created (created_at)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- feature_flags
-- =============================================================================
-- Copilot 灰度/熔断开关。30 秒热读取缓存。

-- SQLite
CREATE TABLE feature_flags (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    flag_key    VARCHAR(64) NOT NULL,            -- e.g. copilot.enabled
    flag_enabled BOOLEAN NOT NULL DEFAULT 0,
    scope_type  VARCHAR(32) NOT NULL DEFAULT 'global', -- global / team / user
    scope_id    VARCHAR(64) NOT NULL DEFAULT '',  -- team_biz_key / user_biz_key (empty for global)
    flag_reason VARCHAR(200) NOT NULL DEFAULT '',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_feature_flags_unique ON feature_flags(flag_key, scope_type, scope_id);

-- MySQL
-- CREATE TABLE feature_flags (
--     id          BIGINT PRIMARY KEY AUTO_INCREMENT,
--     flag_key    VARCHAR(64) NOT NULL,
--     flag_enabled BOOLEAN NOT NULL DEFAULT FALSE,
--     scope_type  VARCHAR(32) NOT NULL DEFAULT 'global',
--     scope_id    VARCHAR(64) NOT NULL DEFAULT '',
--     flag_reason VARCHAR(200) NOT NULL DEFAULT '',
--     created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--     UNIQUE KEY uk_key_scope (flag_key, scope_type, scope_id)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- =============================================================================
-- Seed: default feature flag (Copilot globally disabled at startup)
-- =============================================================================

INSERT INTO feature_flags (flag_key, enabled, scope_type, scope_id, reason)
VALUES ('copilot.enabled', 0, 'global', '', 'Initial seed: globally disabled, enable per-team for gradual rollout');
