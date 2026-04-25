-- schema.sql: MySQL 8.0 compatible schema (JLC standard)
-- Replaces SQLite schema. All tables use pmw_ prefix.

-- pmw_users
CREATE TABLE pmw_users (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key         BIGINT          NOT NULL,
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    username        VARCHAR(64)     NOT NULL,
    display_name    VARCHAR(64)     NOT NULL,
    password_hash   VARCHAR(255)    NOT NULL,
    is_super_admin  TINYINT(1)      NOT NULL DEFAULT 0,
    email           VARCHAR(100)             DEFAULT '',
    user_status     VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_username_deleted (username, deleted_flag, deleted_time),
    KEY idx_users_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- pmw_teams
CREATE TABLE pmw_teams (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key         BIGINT          NOT NULL,
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_name       VARCHAR(100)    NOT NULL,
    team_desc       VARCHAR(500),
    pm_key          BIGINT          NOT NULL,
    code            VARCHAR(6)      NOT NULL DEFAULT '',
    item_seq        INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_teams_code_deleted (code, deleted_flag, deleted_time),
    KEY idx_teams_pm_key (pm_key),
    KEY idx_teams_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队表';

-- pmw_team_members（有软删：成员可被移出团队）
CREATE TABLE pmw_team_members (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key         BIGINT          NOT NULL,
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key        BIGINT          NOT NULL,
    user_key        BIGINT          NOT NULL,
    role_key        BIGINT,
    joined_at       DATETIME        NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_team_user_deleted (team_key, user_key, deleted_flag, deleted_time),
    KEY idx_team_members_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队成员表';

-- pmw_main_items
CREATE TABLE pmw_main_items (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key           BIGINT          NOT NULL,
    create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag      TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time      DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key          BIGINT          NOT NULL,
    code              VARCHAR(12)     NOT NULL,
    title             VARCHAR(100)    NOT NULL,
    item_desc         VARCHAR(2000)   NOT NULL DEFAULT '',
    priority          VARCHAR(5)      NOT NULL,
    proposer_key      BIGINT          NOT NULL,
    assignee_key      BIGINT,
    plan_start_date   DATETIME,
    expected_end_date DATETIME,
    actual_end_date   DATETIME,
    item_status       VARCHAR(20)     NOT NULL DEFAULT '待开始',
    completion        DECIMAL(5,2)    NOT NULL DEFAULT 0.00,
    is_key_item       TINYINT(1)      NOT NULL DEFAULT 0,
    delay_count       INT             NOT NULL DEFAULT 0,
    archived_at       DATETIME,
    sub_item_seq      INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_main_items_team_code_deleted (team_key, code, deleted_flag, deleted_time),
    KEY idx_main_items_team_key (team_key),
    KEY idx_main_items_assignee_key (assignee_key),
    KEY idx_main_items_expected_end_date (expected_end_date),
    KEY idx_main_items_team_status (team_key, item_status),
    KEY idx_main_items_team_priority (team_key, priority),
    KEY idx_main_items_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主事项表';

-- pmw_sub_items
CREATE TABLE pmw_sub_items (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key           BIGINT          NOT NULL,
    create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag      TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time      DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key          BIGINT          NOT NULL,
    main_item_key     BIGINT          NOT NULL,
    code              VARCHAR(15)     NOT NULL DEFAULT '',
    title             VARCHAR(100)    NOT NULL,
    item_desc         VARCHAR(2000),
    priority          VARCHAR(5)      NOT NULL,
    assignee_key      BIGINT,
    plan_start_date   DATETIME,
    expected_end_date DATETIME,
    actual_end_date   DATETIME,
    item_status       VARCHAR(20)     NOT NULL DEFAULT '待开始',
    completion        DECIMAL(5,2)    NOT NULL DEFAULT 0.00,
    is_key_item       TINYINT(1)      NOT NULL DEFAULT 0,
    delay_count       INT             NOT NULL DEFAULT 0,
    weight            DECIMAL(5,2)    NOT NULL DEFAULT 1.00,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_sub_items_main_code (main_item_key, code),
    KEY idx_sub_items_main_item_key (main_item_key),
    KEY idx_sub_items_team_key (team_key),
    KEY idx_sub_items_assignee_key (assignee_key),
    KEY idx_sub_items_team_status (team_key, item_status),
    KEY idx_sub_items_team_priority (team_key, priority),
    KEY idx_sub_items_expected_end_date (expected_end_date),
    KEY idx_sub_items_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子事项表';

-- pmw_item_pools
CREATE TABLE pmw_item_pools (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key           BIGINT          NOT NULL,
    create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag      TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time      DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key          BIGINT          NOT NULL,
    title             VARCHAR(100)    NOT NULL,
    background        VARCHAR(2000),
    expected_output   VARCHAR(1000),
    submitter_key     BIGINT          NOT NULL,
    pool_status       VARCHAR(20)     NOT NULL DEFAULT '待分配',
    assigned_main_key BIGINT,
    assigned_sub_key  BIGINT,
    assignee_key      BIGINT,
    reject_reason     VARCHAR(200),
    reviewed_at       DATETIME,
    reviewer_key      BIGINT,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    KEY idx_item_pools_team_key (team_key),
    KEY idx_item_pools_team_status (team_key, pool_status),
    KEY idx_item_pools_submitter_key (submitter_key),
    KEY idx_item_pools_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='需求池表';

-- pmw_progress_records（append-only：无软删，无 biz_key）
CREATE TABLE pmw_progress_records (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    sub_item_key    BIGINT          NOT NULL,
    team_key        BIGINT          NOT NULL,
    author_key      BIGINT          NOT NULL,
    completion      DECIMAL(5,2)    NOT NULL,
    achievement     VARCHAR(1000),
    blocker         VARCHAR(1000),
    lesson          VARCHAR(1000),
    is_pm_correct   TINYINT(1)      NOT NULL DEFAULT 0,
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_progress_records_sub_item_key (sub_item_key),
    KEY idx_progress_records_sub_item_created (sub_item_key, create_time),
    KEY idx_progress_records_team_key (team_key),
    KEY idx_progress_records_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='进度记录表（追加写入）';

-- pmw_status_histories（append-only：无软删，无 biz_key）
CREATE TABLE pmw_status_histories (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    item_type    VARCHAR(20)     NOT NULL,
    item_key     BIGINT          NOT NULL,
    from_status  VARCHAR(20)     NOT NULL,
    to_status    VARCHAR(20)     NOT NULL,
    changed_by   BIGINT          NOT NULL,
    is_auto      TINYINT(1)      NOT NULL DEFAULT 0,
    remark       VARCHAR(200),
    create_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_item (item_type, item_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='状态变更历史表（追加写入）';
