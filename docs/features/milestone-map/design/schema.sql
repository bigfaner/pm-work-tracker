-- ============================================================
-- Schema: 里程碑图 (Milestone Map)
-- Generated from: design/er-diagram.md
-- ============================================================

-- ============================================================
-- MySQL 8.0 Dialect
-- ============================================================

-- [NEW] pmw_milestone_maps
CREATE TABLE IF NOT EXISTS pmw_milestone_maps (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key         BIGINT          NOT NULL                      COMMENT '业务唯一键',
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据库更新时间',
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    team_key        BIGINT          NOT NULL                      COMMENT '所属团队 biz_key',
    creator_key     BIGINT          NOT NULL                      COMMENT '创建者 biz_key',
    assignee_key    BIGINT          NOT NULL                      COMMENT '负责人 biz_key',
    map_name        VARCHAR(100)    NOT NULL                      COMMENT '里程碑图名称',
    map_desc        VARCHAR(2000)   NOT NULL DEFAULT ''           COMMENT '里程碑图描述',
    map_status      VARCHAR(20)     NOT NULL DEFAULT 'planning'   COMMENT '状态：planning=规划中，reviewed=已评审，ready=待实施，executing=实施中，completed=已完成，cancelled=已取消',
    plan_start_date DATETIME                                   COMMENT '计划开始时间',
    expected_end_date   DATETIME                                   COMMENT '计划完成时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_milestone_maps_team_name_deleted (team_key, map_name, deleted_flag, deleted_time),
    KEY idx_milestone_maps_team_status (team_key, map_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='里程碑图表';

-- [NEW] pmw_milestones
CREATE TABLE IF NOT EXISTS pmw_milestones (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key                 BIGINT          NOT NULL                      COMMENT '业务唯一键',
    create_time             DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    db_update_time          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据库更新时间',
    deleted_flag            TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time            DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    team_key                BIGINT          NOT NULL                      COMMENT '所属团队 biz_key',
    milestone_map_key       BIGINT          NOT NULL                      COMMENT '所属里程碑图 biz_key',
    milestone_name          VARCHAR(100)    NOT NULL                      COMMENT '里程碑名称',
    milestone_desc          VARCHAR(2000)   NOT NULL DEFAULT ''           COMMENT '里程碑描述',
    expected_end_date       DATETIME                                      COMMENT '计划完成时间',
    milestone_status        VARCHAR(20)     NOT NULL DEFAULT 'not_started' COMMENT '状态：not_started=未开始，in_progress=进行中，completed=已完成，cancelled=已取消',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_milestones_map_name_deleted (milestone_map_key, milestone_name, deleted_flag, deleted_time),
    KEY idx_milestones_team_status (team_key, milestone_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='里程碑表';

-- [MODIFIED] pmw_main_items: add milestone_key column
-- ALTER TABLE pmw_main_items
--     ADD COLUMN milestone_key BIGINT COMMENT '所属里程碑 biz_key，NULL 表示未分配',
--     ADD KEY idx_main_items_milestone_key (milestone_key);

-- ============================================================
-- SQLite Dialect
-- ============================================================

-- [NEW] pmw_milestone_maps
CREATE TABLE IF NOT EXISTS pmw_milestone_maps (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key         INTEGER       NOT NULL,
    create_time     DATETIME      NOT NULL DEFAULT (datetime('now')),
    db_update_time  DATETIME      NOT NULL DEFAULT (datetime('now')),
    deleted_flag    INTEGER       NOT NULL DEFAULT 0,
    deleted_time    DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key        INTEGER       NOT NULL,
    creator_key     INTEGER       NOT NULL,
    assignee_key    INTEGER       NOT NULL,
    map_name        VARCHAR(100)  NOT NULL,
    map_desc        VARCHAR(2000) NOT NULL DEFAULT '',
    map_status      VARCHAR(20)   NOT NULL DEFAULT 'planning',
    plan_start_date DATETIME,
    expected_end_date   DATETIME
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_milestone_maps_biz_key ON pmw_milestone_maps(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_milestone_maps_team_name_deleted ON pmw_milestone_maps(team_key, map_name, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_milestone_maps_team_status ON pmw_milestone_maps(team_key, map_status);

-- [NEW] pmw_milestones
CREATE TABLE IF NOT EXISTS pmw_milestones (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    biz_key                 INTEGER       NOT NULL,
    create_time             DATETIME      NOT NULL DEFAULT (datetime('now')),
    db_update_time          DATETIME      NOT NULL DEFAULT (datetime('now')),
    deleted_flag            INTEGER       NOT NULL DEFAULT 0,
    deleted_time            DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key                INTEGER       NOT NULL,
    milestone_map_key       INTEGER       NOT NULL,
    milestone_name          VARCHAR(100)  NOT NULL,
    milestone_desc          VARCHAR(2000) NOT NULL DEFAULT '',
    expected_end_date       DATETIME,
    milestone_status        VARCHAR(20)   NOT NULL DEFAULT 'not_started'
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_milestones_biz_key ON pmw_milestones(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_milestones_map_name_deleted ON pmw_milestones(milestone_map_key, milestone_name, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_milestones_team_status ON pmw_milestones(team_key, milestone_status);

-- [MODIFIED] pmw_main_items: add milestone_key column
-- ALTER TABLE pmw_main_items ADD COLUMN milestone_key INTEGER;
-- CREATE INDEX IF NOT EXISTS idx_main_items_milestone_key ON pmw_main_items(milestone_key);
