-- 001_init.sql: SQLite schema (converted from schema.sql)
-- All tables use pmw_ prefix.

-- pmw_users
CREATE TABLE IF NOT EXISTS pmw_users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key         INTEGER       NOT NULL,
    create_time     DATETIME      NOT NULL DEFAULT (datetime('now')),
    db_update_time  DATETIME      NOT NULL DEFAULT (datetime('now')),
    deleted_flag    INTEGER       NOT NULL DEFAULT 0,
    deleted_time    DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00',
    username        VARCHAR(64)   NOT NULL,
    display_name    VARCHAR(64)   NOT NULL,
    password_hash   VARCHAR(255)  NOT NULL,
    is_super_admin  INTEGER       NOT NULL DEFAULT 0,
    email           VARCHAR(100)  DEFAULT '',
    user_status     VARCHAR(20)   NOT NULL DEFAULT 'enabled'
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_biz_key ON pmw_users(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_username_deleted ON pmw_users(username, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_users_deleted_flag ON pmw_users(deleted_flag);

-- pmw_teams
CREATE TABLE IF NOT EXISTS pmw_teams (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key         INTEGER        NOT NULL,
    create_time     DATETIME       NOT NULL DEFAULT (datetime('now')),
    db_update_time  DATETIME       NOT NULL DEFAULT (datetime('now')),
    deleted_flag    INTEGER        NOT NULL DEFAULT 0,
    deleted_time    DATETIME       NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_name       VARCHAR(100)   NOT NULL,
    team_desc       VARCHAR(500),
    pm_key          INTEGER        NOT NULL,
    code            VARCHAR(6)     NOT NULL DEFAULT '',
    item_seq        INTEGER        NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_teams_biz_key ON pmw_teams(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_teams_code_deleted ON pmw_teams(code, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_teams_pm_key ON pmw_teams(pm_key);
CREATE INDEX IF NOT EXISTS idx_teams_deleted_flag ON pmw_teams(deleted_flag);

-- pmw_team_members
CREATE TABLE IF NOT EXISTS pmw_team_members (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key         INTEGER    NOT NULL,
    create_time     DATETIME   NOT NULL DEFAULT (datetime('now')),
    db_update_time  DATETIME   NOT NULL DEFAULT (datetime('now')),
    deleted_flag    INTEGER    NOT NULL DEFAULT 0,
    deleted_time    DATETIME   NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key        INTEGER    NOT NULL,
    user_key        INTEGER    NOT NULL,
    role_key        INTEGER,
    joined_at       DATETIME   NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_team_members_biz_key ON pmw_team_members(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_team_user_deleted ON pmw_team_members(team_key, user_key, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_team_members_deleted_flag ON pmw_team_members(deleted_flag);

-- pmw_main_items
CREATE TABLE IF NOT EXISTS pmw_main_items (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key           INTEGER       NOT NULL,
    create_time       DATETIME      NOT NULL DEFAULT (datetime('now')),
    db_update_time    DATETIME      NOT NULL DEFAULT (datetime('now')),
    deleted_flag      INTEGER       NOT NULL DEFAULT 0,
    deleted_time      DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key          INTEGER       NOT NULL,
    code              VARCHAR(12)   NOT NULL,
    title             VARCHAR(100)  NOT NULL,
    item_desc         VARCHAR(2000) NOT NULL DEFAULT '',
    priority          VARCHAR(5)    NOT NULL,
    proposer_key      INTEGER       NOT NULL,
    assignee_key      INTEGER,
    plan_start_date   DATETIME,
    expected_end_date DATETIME,
    actual_end_date   DATETIME,
    item_status       VARCHAR(20)   NOT NULL DEFAULT '待开始',
    completion        REAL          NOT NULL DEFAULT 0.00,
    is_key_item       INTEGER       NOT NULL DEFAULT 0,
    delay_count       INTEGER       NOT NULL DEFAULT 0,
    archived_at       DATETIME,
    sub_item_seq      INTEGER       NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_main_items_biz_key ON pmw_main_items(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_main_items_team_code_deleted ON pmw_main_items(team_key, code, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_main_items_team_key ON pmw_main_items(team_key);
CREATE INDEX IF NOT EXISTS idx_main_items_assignee_key ON pmw_main_items(assignee_key);
CREATE INDEX IF NOT EXISTS idx_main_items_expected_end_date ON pmw_main_items(expected_end_date);
CREATE INDEX IF NOT EXISTS idx_main_items_team_status ON pmw_main_items(team_key, item_status);
CREATE INDEX IF NOT EXISTS idx_main_items_team_priority ON pmw_main_items(team_key, priority);
CREATE INDEX IF NOT EXISTS idx_main_items_deleted_flag ON pmw_main_items(deleted_flag);

-- pmw_sub_items
CREATE TABLE IF NOT EXISTS pmw_sub_items (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key           INTEGER       NOT NULL,
    create_time       DATETIME      NOT NULL DEFAULT (datetime('now')),
    db_update_time    DATETIME      NOT NULL DEFAULT (datetime('now')),
    deleted_flag      INTEGER       NOT NULL DEFAULT 0,
    deleted_time      DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key          INTEGER       NOT NULL,
    main_item_key     INTEGER       NOT NULL,
    code              VARCHAR(15)   NOT NULL DEFAULT '',
    title             VARCHAR(100)  NOT NULL,
    item_desc         VARCHAR(2000),
    priority          VARCHAR(5)    NOT NULL,
    assignee_key      INTEGER,
    plan_start_date   DATETIME,
    expected_end_date DATETIME,
    actual_end_date   DATETIME,
    item_status       VARCHAR(20)   NOT NULL DEFAULT '待开始',
    completion        REAL          NOT NULL DEFAULT 0.00,
    is_key_item       INTEGER       NOT NULL DEFAULT 0,
    delay_count       INTEGER       NOT NULL DEFAULT 0,
    weight            REAL          NOT NULL DEFAULT 1.00
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sub_items_biz_key ON pmw_sub_items(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sub_items_main_code ON pmw_sub_items(main_item_key, code);
CREATE INDEX IF NOT EXISTS idx_sub_items_main_item_key ON pmw_sub_items(main_item_key);
CREATE INDEX IF NOT EXISTS idx_sub_items_team_key ON pmw_sub_items(team_key);
CREATE INDEX IF NOT EXISTS idx_sub_items_assignee_key ON pmw_sub_items(assignee_key);
CREATE INDEX IF NOT EXISTS idx_sub_items_team_status ON pmw_sub_items(team_key, item_status);
CREATE INDEX IF NOT EXISTS idx_sub_items_team_priority ON pmw_sub_items(team_key, priority);
CREATE INDEX IF NOT EXISTS idx_sub_items_expected_end_date ON pmw_sub_items(expected_end_date);
CREATE INDEX IF NOT EXISTS idx_sub_items_deleted_flag ON pmw_sub_items(deleted_flag);

-- pmw_item_pools
CREATE TABLE IF NOT EXISTS pmw_item_pools (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key           INTEGER       NOT NULL,
    create_time       DATETIME      NOT NULL DEFAULT (datetime('now')),
    db_update_time    DATETIME      NOT NULL DEFAULT (datetime('now')),
    deleted_flag      INTEGER       NOT NULL DEFAULT 0,
    deleted_time      DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key          INTEGER       NOT NULL,
    title             VARCHAR(100)  NOT NULL,
    background        VARCHAR(2000),
    expected_output   VARCHAR(1000),
    submitter_key     INTEGER       NOT NULL,
    pool_status       VARCHAR(20)   NOT NULL DEFAULT 'pending',
    assigned_main_key INTEGER,
    assigned_sub_key  INTEGER,
    assignee_key      INTEGER,
    reject_reason     VARCHAR(200),
    reviewed_at       DATETIME,
    reviewer_key      INTEGER
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_item_pools_biz_key ON pmw_item_pools(biz_key);
CREATE INDEX IF NOT EXISTS idx_item_pools_team_key ON pmw_item_pools(team_key);
CREATE INDEX IF NOT EXISTS idx_item_pools_team_status ON pmw_item_pools(team_key, pool_status);
CREATE INDEX IF NOT EXISTS idx_item_pools_submitter_key ON pmw_item_pools(submitter_key);
CREATE INDEX IF NOT EXISTS idx_item_pools_deleted_flag ON pmw_item_pools(deleted_flag);

-- pmw_progress_records (append-only: no soft-delete)
CREATE TABLE IF NOT EXISTS pmw_progress_records (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key         INTEGER       NOT NULL,
    sub_item_key    INTEGER       NOT NULL,
    team_key        INTEGER       NOT NULL,
    author_key      INTEGER       NOT NULL,
    completion      REAL          NOT NULL,
    achievement     VARCHAR(1000),
    blocker         VARCHAR(1000),
    lesson          VARCHAR(1000),
    is_pm_correct   INTEGER       NOT NULL DEFAULT 0,
    create_time     DATETIME      NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_progress_records_biz_key ON pmw_progress_records(biz_key);
CREATE INDEX IF NOT EXISTS idx_progress_records_sub_item_key ON pmw_progress_records(sub_item_key);
CREATE INDEX IF NOT EXISTS idx_progress_records_sub_item_created ON pmw_progress_records(sub_item_key, create_time);
CREATE INDEX IF NOT EXISTS idx_progress_records_team_key ON pmw_progress_records(team_key);
CREATE INDEX IF NOT EXISTS idx_progress_records_create_time ON pmw_progress_records(create_time);

-- roles (RBAC)
CREATE TABLE IF NOT EXISTS roles (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key         INTEGER       NOT NULL,
    create_time     DATETIME      NOT NULL DEFAULT (datetime('now')),
    db_update_time  DATETIME      NOT NULL DEFAULT (datetime('now')),
    deleted_flag    INTEGER       NOT NULL DEFAULT 0,
    deleted_time    DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00',
    name            VARCHAR(50)   NOT NULL,
    description     VARCHAR(200)  NOT NULL DEFAULT '',
    is_preset       INTEGER       NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_roles_name ON roles(name);

-- role_permissions (RBAC)
CREATE TABLE IF NOT EXISTS role_permissions (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id          INTEGER      NOT NULL,
    permission_code  VARCHAR(50)  NOT NULL,
    UNIQUE(role_id, permission_code)
);

-- pmw_status_histories (append-only: no soft-delete, no biz_key)
CREATE TABLE IF NOT EXISTS pmw_status_histories (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    item_type    VARCHAR(20)  NOT NULL,
    item_key     INTEGER      NOT NULL,
    from_status  VARCHAR(20)  NOT NULL,
    to_status    VARCHAR(20)  NOT NULL,
    changed_by   INTEGER      NOT NULL,
    is_auto      INTEGER      NOT NULL DEFAULT 0,
    remark       VARCHAR(200),
    create_time  DATETIME     NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_status_histories_item ON pmw_status_histories(item_type, item_key);
