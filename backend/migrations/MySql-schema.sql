-- schema.sql: MySQL 8.0 compatible schema (JLC standard)
-- Replaces SQLite schema. All tables use pmw_ prefix.

-- pmw_users
CREATE TABLE IF NOT EXISTS pmw_users (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key         BIGINT          NOT NULL                      COMMENT '业务唯一键',
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据库更新时间',
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    username        VARCHAR(64)     NOT NULL                      COMMENT '登录用户名',
    display_name    VARCHAR(64)     NOT NULL                      COMMENT '显示名称',
    password_hash   VARCHAR(255)    NOT NULL                      COMMENT '密码哈希值',
    is_super_admin  TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '是否超级管理员：0=否，1=是',
    email           VARCHAR(100)             DEFAULT ''           COMMENT '邮箱地址',
    user_status     VARCHAR(20)     NOT NULL DEFAULT 'enabled'    COMMENT '用户状态：enabled=启用，disabled=禁用',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_username_deleted (username, deleted_flag, deleted_time),
    KEY idx_users_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- pmw_teams
CREATE TABLE IF NOT EXISTS pmw_teams (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key         BIGINT          NOT NULL                      COMMENT '业务唯一键',
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据库更新时间',
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    team_name       VARCHAR(100)    NOT NULL                      COMMENT '团队名称',
    team_desc       VARCHAR(500)                                  COMMENT '团队描述',
    pm_key          BIGINT          NOT NULL                      COMMENT '团队负责人 biz_key',
    team_code       VARCHAR(6)      NOT NULL DEFAULT ''           COMMENT '团队邀请码',
    item_seq        INT             NOT NULL DEFAULT 0            COMMENT '主事项序号计数器，用于生成事项编号',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_teams_code_deleted (team_code, deleted_flag, deleted_time),
    KEY idx_teams_pm_key (pm_key),
    KEY idx_teams_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队表';

-- pmw_team_members（有软删：成员可被移出团队）
CREATE TABLE IF NOT EXISTS pmw_team_members (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key         BIGINT          NOT NULL                      COMMENT '业务唯一键',
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据库更新时间',
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    team_key        BIGINT          NOT NULL                      COMMENT '所属团队 biz_key',
    user_key        BIGINT          NOT NULL                      COMMENT '成员用户 biz_key',
    role_key        BIGINT                                        COMMENT '成员角色 biz_key，NULL 表示未分配角色',
    joined_at       DATETIME        NOT NULL                      COMMENT '加入团队时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_team_user_deleted (team_key, user_key, deleted_flag, deleted_time),
    KEY idx_team_members_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队成员表';

-- pmw_main_items
CREATE TABLE IF NOT EXISTS pmw_main_items (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key           BIGINT          NOT NULL                      COMMENT '业务唯一键',
    create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    db_update_time    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据库更新时间',
    deleted_flag      TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time      DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    team_key          BIGINT          NOT NULL                      COMMENT '所属团队 biz_key',
    item_code         VARCHAR(12)     NOT NULL                      COMMENT '事项编号，团队内唯一',
    title             VARCHAR(100)    NOT NULL                      COMMENT '事项标题',
    item_desc         VARCHAR(2000)   NOT NULL DEFAULT ''           COMMENT '事项描述',
    priority          VARCHAR(5)      NOT NULL                      COMMENT '优先级：P0/P1/P2/P3',
    proposer_key      BIGINT          NOT NULL                      COMMENT '提出人 biz_key',
    assignee_key      BIGINT                                        COMMENT '负责人 biz_key，NULL 表示未分配',
    plan_start_date   DATETIME                                      COMMENT '计划开始日期',
    expected_end_date DATETIME                                      COMMENT '预计结束日期',
    actual_end_date   DATETIME                                      COMMENT '实际结束日期',
    item_status       VARCHAR(20)     NOT NULL DEFAULT '待开始'     COMMENT '事项状态：待开始/进行中/已完成/已暂停',
    completion_pct    DECIMAL(5,2)    NOT NULL DEFAULT 0.00         COMMENT '完成度百分比，0.00~100.00',
    is_key_item       TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '是否关键事项：0=否，1=是',
    delay_count       INT             NOT NULL DEFAULT 0            COMMENT '延期次数',
    archived_at       DATETIME                                      COMMENT '归档时间，NULL 表示未归档',
    sub_item_seq      INT             NOT NULL DEFAULT 0            COMMENT '子事项序号计数器，用于生成子事项编号',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_main_items_team_code_deleted (team_key, item_code, deleted_flag, deleted_time),
    KEY idx_main_items_team_key (team_key),
    KEY idx_main_items_assignee_key (assignee_key),
    KEY idx_main_items_expected_end_date (expected_end_date),
    KEY idx_main_items_team_status (team_key, item_status),
    KEY idx_main_items_team_priority (team_key, priority),
    KEY idx_main_items_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主事项表';

-- pmw_sub_items
CREATE TABLE IF NOT EXISTS pmw_sub_items (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key           BIGINT          NOT NULL                      COMMENT '业务唯一键',
    create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    db_update_time    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据库更新时间',
    deleted_flag      TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time      DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    team_key          BIGINT          NOT NULL                      COMMENT '所属团队 biz_key',
    main_item_key     BIGINT          NOT NULL                      COMMENT '所属主事项 biz_key',
    item_code         VARCHAR(15)     NOT NULL DEFAULT ''           COMMENT '子事项编号，主事项内唯一',
    title             VARCHAR(100)    NOT NULL                      COMMENT '子事项标题',
    item_desc         VARCHAR(2000)                                 COMMENT '子事项描述',
    priority          VARCHAR(5)      NOT NULL                      COMMENT '优先级：P0/P1/P2/P3',
    assignee_key      BIGINT                                        COMMENT '负责人 biz_key，NULL 表示未分配',
    plan_start_date   DATETIME                                      COMMENT '计划开始日期',
    expected_end_date DATETIME                                      COMMENT '预计结束日期',
    actual_end_date   DATETIME                                      COMMENT '实际结束日期',
    item_status       VARCHAR(20)     NOT NULL DEFAULT '待开始'     COMMENT '事项状态：待开始/进行中/已完成/已暂停',
    completion_pct    DECIMAL(5,2)    NOT NULL DEFAULT 0.00         COMMENT '完成度百分比，0.00~100.00',
    is_key_item       TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '是否关键子事项：0=否，1=是',
    delay_count       INT             NOT NULL DEFAULT 0            COMMENT '延期次数',
    weight            DECIMAL(5,2)    NOT NULL DEFAULT 1.00         COMMENT '权重，用于计算父事项完成度',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_sub_items_main_code (main_item_key, item_code, deleted_flag, deleted_time),
    KEY idx_sub_items_main_item_key (main_item_key),
    KEY idx_sub_items_team_key (team_key),
    KEY idx_sub_items_assignee_key (assignee_key),
    KEY idx_sub_items_team_status (team_key, item_status),
    KEY idx_sub_items_team_priority (team_key, priority),
    KEY idx_sub_items_expected_end_date (expected_end_date),
    KEY idx_sub_items_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子事项表';

-- pmw_item_pools
CREATE TABLE IF NOT EXISTS pmw_item_pools (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key           BIGINT          NOT NULL                      COMMENT '业务唯一键',
    create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    db_update_time    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据库更新时间',
    deleted_flag      TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time      DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    team_key          BIGINT          NOT NULL                      COMMENT '所属团队 biz_key',
    title             VARCHAR(100)    NOT NULL                      COMMENT '需求标题',
    background        VARCHAR(2000)                                 COMMENT '需求背景',
    expected_output   VARCHAR(1000)                                 COMMENT '预期产出',
    submitter_key     BIGINT          NOT NULL                      COMMENT '提交人 biz_key',
    pool_status       VARCHAR(20)     NOT NULL DEFAULT 'pending'    COMMENT '需求状态：pending=待评审，approved=已通过，rejected=已拒绝，converted=已转化',
    assigned_main_key BIGINT                                        COMMENT '转化后关联的主事项 biz_key',
    assigned_sub_key  BIGINT                                        COMMENT '转化后关联的子事项 biz_key',
    assignee_key      BIGINT                                        COMMENT '指派处理人 biz_key',
    reject_reason     VARCHAR(200)                                  COMMENT '拒绝原因',
    reviewed_at       DATETIME                                      COMMENT '评审时间',
    reviewer_key      BIGINT                                        COMMENT '评审人 biz_key',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    KEY idx_item_pools_team_key (team_key),
    KEY idx_item_pools_team_status (team_key, pool_status),
    KEY idx_item_pools_submitter_key (submitter_key),
    KEY idx_item_pools_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='需求池表';

-- pmw_progress_records（append-only：无软删）
CREATE TABLE IF NOT EXISTS pmw_progress_records (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key         BIGINT          NOT NULL                      COMMENT '业务唯一键',
    sub_item_key    BIGINT          NOT NULL                      COMMENT '所属子事项 biz_key',
    team_key        BIGINT          NOT NULL                      COMMENT '所属团队 biz_key',
    author_key      BIGINT          NOT NULL                      COMMENT '填写人 biz_key',
    completion_pct  DECIMAL(5,2)    NOT NULL                      COMMENT '本次填写的完成度百分比',
    achievement     VARCHAR(1000)                                 COMMENT '本周成果',
    blocker         VARCHAR(1000)                                 COMMENT '阻塞问题',
    lesson          VARCHAR(1000)                                 COMMENT '经验教训',
    is_pm_correct   TINYINT(1)      NOT NULL DEFAULT 0            COMMENT 'PM 是否已修正：0=否，1=是',
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    KEY idx_progress_records_sub_item_key (sub_item_key),
    KEY idx_progress_records_sub_item_created (sub_item_key, create_time),
    KEY idx_progress_records_team_key (team_key),
    KEY idx_progress_records_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='进度记录表（追加写入）';

-- pmw_status_histories（append-only：无软删，无 biz_key）
CREATE TABLE IF NOT EXISTS pmw_status_histories (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    item_type    VARCHAR(20)     NOT NULL                      COMMENT '事项类型：main_item=主事项，sub_item=子事项',
    item_key     BIGINT          NOT NULL                      COMMENT '事项 biz_key',
    from_status  VARCHAR(20)     NOT NULL                      COMMENT '变更前状态',
    to_status    VARCHAR(20)     NOT NULL                      COMMENT '变更后状态',
    changed_by   BIGINT          NOT NULL                      COMMENT '操作人 biz_key',
    is_auto      TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '是否系统自动变更：0=人工，1=自动',
    remark       VARCHAR(200)                                  COMMENT '变更备注',
    create_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    PRIMARY KEY (id),
    KEY idx_item (item_type, item_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='状态变更历史表（追加写入）';

-- pmw_roles (RBAC)
CREATE TABLE IF NOT EXISTS pmw_roles (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    biz_key         BIGINT          NOT NULL                      COMMENT '业务唯一键',
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据库更新时间',
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    role_name       VARCHAR(50)     NOT NULL                      COMMENT '角色名称',
    role_desc       VARCHAR(200)    NOT NULL DEFAULT ''           COMMENT '角色描述',
    is_preset       TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '是否预置角色：0=自定义，1=系统预置',
    PRIMARY KEY (id),
    UNIQUE KEY uk_roles_name (role_name, deleted_flag, deleted_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- pmw_role_permissions (RBAC)
CREATE TABLE IF NOT EXISTS pmw_role_permissions (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT       COMMENT '自增主键',
    deleted_flag     TINYINT(1)      NOT NULL DEFAULT 0            COMMENT '软删标志：0=正常，1=已删除',
    deleted_time     DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值',
    role_id          BIGINT UNSIGNED NOT NULL                      COMMENT '角色 id（关联 pmw_roles.id）',
    permission_code  VARCHAR(50)     NOT NULL                      COMMENT '权限码，如 item:create、item:delete',
    PRIMARY KEY (id),
    UNIQUE KEY uk_role_permission (role_id, permission_code, deleted_flag, deleted_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限表';
