-- 001_init.sql: Initial schema for PM Work Tracker
-- All tables use explicit SQL (no GORM AutoMigrate).

-- users
CREATE TABLE IF NOT EXISTS users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at      DATETIME,
    updated_at      DATETIME,
    deleted_at      DATETIME,
    username        VARCHAR(64)  NOT NULL,
    display_name    VARCHAR(64)  NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    is_super_admin  BOOLEAN      NOT NULL DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- teams
CREATE TABLE IF NOT EXISTS teams (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at   DATETIME,
    updated_at   DATETIME,
    deleted_at   DATETIME,
    name         VARCHAR(100) NOT NULL,
    description  VARCHAR(500),
    pm_id        INTEGER      NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_teams_pm_id ON teams(pm_id);
CREATE INDEX IF NOT EXISTS idx_teams_deleted_at ON teams(deleted_at);

-- team_members
CREATE TABLE IF NOT EXISTS team_members (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id    INTEGER   NOT NULL,
    user_id    INTEGER   NOT NULL,
    role       VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at  DATETIME  NOT NULL,
    created_at DATETIME,
    updated_at DATETIME,
    CONSTRAINT idx_team_user UNIQUE (team_id, user_id)
);

-- main_items
CREATE TABLE IF NOT EXISTS main_items (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at        DATETIME,
    updated_at        DATETIME,
    deleted_at        DATETIME,
    team_id           INTEGER      NOT NULL,
    code              VARCHAR(10)  NOT NULL,
    title             VARCHAR(100) NOT NULL,
    priority          VARCHAR(5)   NOT NULL,
    proposer_id       INTEGER      NOT NULL,
    assignee_id       INTEGER,
    start_date        DATETIME,
    expected_end_date DATETIME,
    actual_end_date   DATETIME,
    status            VARCHAR(20)  NOT NULL DEFAULT '待开始',
    completion        REAL         NOT NULL DEFAULT 0,
    is_key_item       BOOLEAN      NOT NULL DEFAULT FALSE,
    delay_count       INTEGER      NOT NULL DEFAULT 0,
    archived_at       DATETIME
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_main_items_code ON main_items(code);
CREATE INDEX IF NOT EXISTS idx_main_items_team_id ON main_items(team_id);
CREATE INDEX IF NOT EXISTS idx_main_items_assignee_id ON main_items(assignee_id);
CREATE INDEX IF NOT EXISTS idx_main_items_expected_end_date ON main_items(expected_end_date);
CREATE INDEX IF NOT EXISTS idx_main_items_team_status ON main_items(team_id, status);
CREATE INDEX IF NOT EXISTS idx_main_items_team_priority ON main_items(team_id, priority);
CREATE INDEX IF NOT EXISTS idx_main_items_deleted_at ON main_items(deleted_at);

-- sub_items
CREATE TABLE IF NOT EXISTS sub_items (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at        DATETIME,
    updated_at        DATETIME,
    deleted_at        DATETIME,
    team_id           INTEGER      NOT NULL,
    main_item_id      INTEGER      NOT NULL,
    title             VARCHAR(100) NOT NULL,
    description       TEXT,
    priority          VARCHAR(5)   NOT NULL,
    assignee_id       INTEGER,
    start_date        DATETIME,
    expected_end_date DATETIME,
    actual_end_date   DATETIME,
    status            VARCHAR(20)  NOT NULL DEFAULT '待开始',
    completion        REAL         NOT NULL DEFAULT 0,
    is_key_item       BOOLEAN      NOT NULL DEFAULT FALSE,
    delay_count       INTEGER      NOT NULL DEFAULT 0,
    weight            REAL         NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_sub_items_main_item_id ON sub_items(main_item_id);
CREATE INDEX IF NOT EXISTS idx_sub_items_team_id ON sub_items(team_id);
CREATE INDEX IF NOT EXISTS idx_sub_items_assignee_id ON sub_items(assignee_id);
CREATE INDEX IF NOT EXISTS idx_sub_items_team_status ON sub_items(team_id, status);
CREATE INDEX IF NOT EXISTS idx_sub_items_expected_end_date ON sub_items(expected_end_date);
CREATE INDEX IF NOT EXISTS idx_sub_items_deleted_at ON sub_items(deleted_at);

-- progress_records
CREATE TABLE IF NOT EXISTS progress_records (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    sub_item_id  INTEGER NOT NULL,
    team_id      INTEGER NOT NULL,
    author_id    INTEGER NOT NULL,
    completion   REAL    NOT NULL,
    achievement  TEXT,
    blocker      TEXT,
    lesson       TEXT,
    is_pm_correct BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_progress_records_sub_item_id ON progress_records(sub_item_id);
CREATE INDEX IF NOT EXISTS idx_progress_records_sub_item_created ON progress_records(sub_item_id, created_at);
CREATE INDEX IF NOT EXISTS idx_progress_records_team_id ON progress_records(team_id);
CREATE INDEX IF NOT EXISTS idx_progress_records_created_at ON progress_records(created_at);

-- item_pools
CREATE TABLE IF NOT EXISTS item_pools (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at       DATETIME,
    updated_at       DATETIME,
    deleted_at       DATETIME,
    team_id          INTEGER      NOT NULL,
    title            VARCHAR(100) NOT NULL,
    background       TEXT,
    expected_output  TEXT,
    submitter_id     INTEGER      NOT NULL,
    status           VARCHAR(20)  NOT NULL DEFAULT '待分配',
    assigned_main_id INTEGER,
    assigned_sub_id  INTEGER,
    assignee_id      INTEGER,
    reject_reason    VARCHAR(200),
    reviewed_at      DATETIME,
    reviewer_id      INTEGER
);
CREATE INDEX IF NOT EXISTS idx_item_pools_team_id ON item_pools(team_id);
CREATE INDEX IF NOT EXISTS idx_item_pools_team_status ON item_pools(team_id, status);
CREATE INDEX IF NOT EXISTS idx_item_pools_submitter_id ON item_pools(submitter_id);
CREATE INDEX IF NOT EXISTS idx_item_pools_deleted_at ON item_pools(deleted_at);
