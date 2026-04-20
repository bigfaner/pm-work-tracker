-- 007_status_histories.sql: Create status_histories table
-- Append-only log for status transition audit. No soft delete, no updated_at.

CREATE TABLE IF NOT EXISTS status_histories (
    id           INTEGER      PRIMARY KEY AUTOINCREMENT,
    item_type    VARCHAR(20)  NOT NULL,
    item_id      INTEGER      NOT NULL,
    from_status  VARCHAR(20)  NOT NULL,
    to_status    VARCHAR(20)  NOT NULL,
    changed_by   INTEGER      NOT NULL,
    is_auto      BOOLEAN      NOT NULL DEFAULT FALSE,
    remark       VARCHAR(200),
    created_at   DATETIME     NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_item ON status_histories(item_type, item_id);
