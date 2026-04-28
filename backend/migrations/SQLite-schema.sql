-- 001_init.sql: SQLite schema (converted from schema.sql)
-- All tables use pmw_ prefix.

-- pmw_users
CREATE TABLE IF NOT EXISTS pmw_users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    biz_key         INTEGER       NOT NULL,            -- 业务唯一键
    create_time     DATETIME      NOT NULL DEFAULT (datetime('now')), -- 创建时间
    db_update_time  DATETIME      NOT NULL DEFAULT (datetime('now')), -- 数据库更新时间
    deleted_flag    INTEGER       NOT NULL DEFAULT 0,  -- 软删标志：0=正常，1=已删除
    deleted_time    DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00', -- 软删时间，未删除时为固定占位值
    username        VARCHAR(64)   NOT NULL,            -- 登录用户名
    display_name    VARCHAR(64)   NOT NULL,            -- 显示名称
    password_hash   VARCHAR(255)  NOT NULL,            -- 密码哈希值
    is_super_admin  INTEGER       NOT NULL DEFAULT 0,  -- 是否超级管理员：0=否，1=是
    email           VARCHAR(100)  DEFAULT '',          -- 邮箱地址
    user_status     VARCHAR(20)   NOT NULL DEFAULT 'enabled' -- 用户状态：enabled=启用，disabled=禁用
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_biz_key ON pmw_users(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_username_deleted ON pmw_users(username, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_users_deleted_flag ON pmw_users(deleted_flag);

-- pmw_teams
CREATE TABLE IF NOT EXISTS pmw_teams (
    id              INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    biz_key         INTEGER        NOT NULL,           -- 业务唯一键
    create_time     DATETIME       NOT NULL DEFAULT (datetime('now')), -- 创建时间
    db_update_time  DATETIME       NOT NULL DEFAULT (datetime('now')), -- 数据库更新时间
    deleted_flag    INTEGER        NOT NULL DEFAULT 0, -- 软删标志：0=正常，1=已删除
    deleted_time    DATETIME       NOT NULL DEFAULT '1970-01-01 08:00:00', -- 软删时间，未删除时为固定占位值
    team_name       VARCHAR(100)   NOT NULL,           -- 团队名称
    team_desc       VARCHAR(500),                      -- 团队描述
    pm_key          INTEGER        NOT NULL,           -- 团队负责人 biz_key
    team_code       VARCHAR(6)     NOT NULL DEFAULT '', -- 团队邀请码
    item_seq        INTEGER        NOT NULL DEFAULT 0  -- 主事项序号计数器，用于生成事项编号
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_teams_biz_key ON pmw_teams(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_teams_code_deleted ON pmw_teams(team_code, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_teams_pm_key ON pmw_teams(pm_key);
CREATE INDEX IF NOT EXISTS idx_teams_deleted_flag ON pmw_teams(deleted_flag);

-- pmw_team_members
CREATE TABLE IF NOT EXISTS pmw_team_members (
    id              INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    biz_key         INTEGER    NOT NULL,               -- 业务唯一键
    create_time     DATETIME   NOT NULL DEFAULT (datetime('now')), -- 创建时间
    db_update_time  DATETIME   NOT NULL DEFAULT (datetime('now')), -- 数据库更新时间
    deleted_flag    INTEGER    NOT NULL DEFAULT 0,     -- 软删标志：0=正常，1=已删除
    deleted_time    DATETIME   NOT NULL DEFAULT '1970-01-01 08:00:00', -- 软删时间，未删除时为固定占位值
    team_key        INTEGER    NOT NULL,               -- 所属团队 biz_key
    user_key        INTEGER    NOT NULL,               -- 成员用户 biz_key
    role_key        INTEGER,                           -- 成员角色 biz_key，NULL 表示未分配角色
    joined_at       DATETIME   NOT NULL                -- 加入团队时间
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_team_members_biz_key ON pmw_team_members(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_team_user_deleted ON pmw_team_members(team_key, user_key, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_team_members_deleted_flag ON pmw_team_members(deleted_flag);

-- pmw_main_items
CREATE TABLE IF NOT EXISTS pmw_main_items (
    id                INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    biz_key           INTEGER       NOT NULL,            -- 业务唯一键
    create_time       DATETIME      NOT NULL DEFAULT (datetime('now')), -- 创建时间
    db_update_time    DATETIME      NOT NULL DEFAULT (datetime('now')), -- 数据库更新时间
    deleted_flag      INTEGER       NOT NULL DEFAULT 0,  -- 软删标志：0=正常，1=已删除
    deleted_time      DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00', -- 软删时间，未删除时为固定占位值
    team_key          INTEGER       NOT NULL,            -- 所属团队 biz_key
    item_code         VARCHAR(12)   NOT NULL,            -- 事项编号，团队内唯一
    title             VARCHAR(100)  NOT NULL,            -- 事项标题
    item_desc         VARCHAR(2000) NOT NULL DEFAULT '', -- 事项描述
    priority          VARCHAR(5)    NOT NULL,            -- 优先级：P0/P1/P2/P3
    proposer_key      INTEGER       NOT NULL,            -- 提出人 biz_key
    assignee_key      INTEGER,                           -- 负责人 biz_key，NULL 表示未分配
    plan_start_date   DATETIME,                          -- 计划开始日期
    expected_end_date DATETIME,                          -- 预计结束日期
    actual_end_date   DATETIME,                          -- 实际结束日期
    item_status       VARCHAR(20)   NOT NULL DEFAULT '待开始', -- 事项状态：待开始/进行中/已完成/已暂停
    completion_pct    REAL          NOT NULL DEFAULT 0.00, -- 完成度百分比，0.00~100.00
    is_key_item       INTEGER       NOT NULL DEFAULT 0,  -- 是否关键事项：0=否，1=是
    delay_count       INTEGER       NOT NULL DEFAULT 0,  -- 延期次数
    archived_at       DATETIME,                          -- 归档时间，NULL 表示未归档
    sub_item_seq      INTEGER       NOT NULL DEFAULT 0   -- 子事项序号计数器，用于生成子事项编号
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_main_items_biz_key ON pmw_main_items(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_main_items_team_code_deleted ON pmw_main_items(team_key, item_code, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_main_items_team_key ON pmw_main_items(team_key);
CREATE INDEX IF NOT EXISTS idx_main_items_assignee_key ON pmw_main_items(assignee_key);
CREATE INDEX IF NOT EXISTS idx_main_items_expected_end_date ON pmw_main_items(expected_end_date);
CREATE INDEX IF NOT EXISTS idx_main_items_team_status ON pmw_main_items(team_key, item_status);
CREATE INDEX IF NOT EXISTS idx_main_items_team_priority ON pmw_main_items(team_key, priority);
CREATE INDEX IF NOT EXISTS idx_main_items_deleted_flag ON pmw_main_items(deleted_flag);

-- pmw_sub_items
CREATE TABLE IF NOT EXISTS pmw_sub_items (
    id                INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    biz_key           INTEGER       NOT NULL,            -- 业务唯一键
    create_time       DATETIME      NOT NULL DEFAULT (datetime('now')), -- 创建时间
    db_update_time    DATETIME      NOT NULL DEFAULT (datetime('now')), -- 数据库更新时间
    deleted_flag      INTEGER       NOT NULL DEFAULT 0,  -- 软删标志：0=正常，1=已删除
    deleted_time      DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00', -- 软删时间，未删除时为固定占位值
    team_key          INTEGER       NOT NULL,            -- 所属团队 biz_key
    main_item_key     INTEGER       NOT NULL,            -- 所属主事项 biz_key
    item_code         VARCHAR(15)   NOT NULL DEFAULT '', -- 子事项编号，主事项内唯一
    title             VARCHAR(100)  NOT NULL,            -- 子事项标题
    item_desc         VARCHAR(2000),                     -- 子事项描述
    priority          VARCHAR(5)    NOT NULL,            -- 优先级：P0/P1/P2/P3
    assignee_key      INTEGER,                           -- 负责人 biz_key，NULL 表示未分配
    plan_start_date   DATETIME,                          -- 计划开始日期
    expected_end_date DATETIME,                          -- 预计结束日期
    actual_end_date   DATETIME,                          -- 实际结束日期
    item_status       VARCHAR(20)   NOT NULL DEFAULT '待开始', -- 事项状态：待开始/进行中/已完成/已暂停
    completion_pct    REAL          NOT NULL DEFAULT 0.00, -- 完成度百分比，0.00~100.00
    is_key_item       INTEGER       NOT NULL DEFAULT 0,  -- 是否关键子事项：0=否，1=是
    delay_count       INTEGER       NOT NULL DEFAULT 0,  -- 延期次数
    weight            REAL          NOT NULL DEFAULT 1.00 -- 权重，用于计算父事项完成度
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sub_items_biz_key ON pmw_sub_items(biz_key);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sub_items_main_code ON pmw_sub_items(main_item_key, item_code, deleted_flag, deleted_time);
CREATE INDEX IF NOT EXISTS idx_sub_items_main_item_key ON pmw_sub_items(main_item_key);
CREATE INDEX IF NOT EXISTS idx_sub_items_team_key ON pmw_sub_items(team_key);
CREATE INDEX IF NOT EXISTS idx_sub_items_assignee_key ON pmw_sub_items(assignee_key);
CREATE INDEX IF NOT EXISTS idx_sub_items_team_status ON pmw_sub_items(team_key, item_status);
CREATE INDEX IF NOT EXISTS idx_sub_items_team_priority ON pmw_sub_items(team_key, priority);
CREATE INDEX IF NOT EXISTS idx_sub_items_expected_end_date ON pmw_sub_items(expected_end_date);
CREATE INDEX IF NOT EXISTS idx_sub_items_deleted_flag ON pmw_sub_items(deleted_flag);

-- pmw_item_pools
CREATE TABLE IF NOT EXISTS pmw_item_pools (
    id                INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    biz_key           INTEGER       NOT NULL,            -- 业务唯一键
    create_time       DATETIME      NOT NULL DEFAULT (datetime('now')), -- 创建时间
    db_update_time    DATETIME      NOT NULL DEFAULT (datetime('now')), -- 数据库更新时间
    deleted_flag      INTEGER       NOT NULL DEFAULT 0,  -- 软删标志：0=正常，1=已删除
    deleted_time      DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00', -- 软删时间，未删除时为固定占位值
    team_key          INTEGER       NOT NULL,            -- 所属团队 biz_key
    title             VARCHAR(100)  NOT NULL,            -- 需求标题
    background        VARCHAR(2000),                     -- 需求背景
    expected_output   VARCHAR(1000),                     -- 预期产出
    submitter_key     INTEGER       NOT NULL,            -- 提交人 biz_key
    pool_status       VARCHAR(20)   NOT NULL DEFAULT 'pending', -- 需求状态：pending=待评审，approved=已通过，rejected=已拒绝，converted=已转化
    assigned_main_key INTEGER,                           -- 转化后关联的主事项 biz_key
    assigned_sub_key  INTEGER,                           -- 转化后关联的子事项 biz_key
    assignee_key      INTEGER,                           -- 指派处理人 biz_key
    reject_reason     VARCHAR(200),                      -- 拒绝原因
    reviewed_at       DATETIME,                          -- 评审时间
    reviewer_key      INTEGER                            -- 评审人 biz_key
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_item_pools_biz_key ON pmw_item_pools(biz_key);
CREATE INDEX IF NOT EXISTS idx_item_pools_team_key ON pmw_item_pools(team_key);
CREATE INDEX IF NOT EXISTS idx_item_pools_team_status ON pmw_item_pools(team_key, pool_status);
CREATE INDEX IF NOT EXISTS idx_item_pools_submitter_key ON pmw_item_pools(submitter_key);
CREATE INDEX IF NOT EXISTS idx_item_pools_deleted_flag ON pmw_item_pools(deleted_flag);

-- pmw_progress_records (append-only: no soft-delete)
CREATE TABLE IF NOT EXISTS pmw_progress_records (
    id              INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    biz_key         INTEGER       NOT NULL,            -- 业务唯一键
    sub_item_key    INTEGER       NOT NULL,            -- 所属子事项 id
    team_key        INTEGER       NOT NULL,            -- 所属团队 id
    author_key      INTEGER       NOT NULL,            -- 填写人 id
    completion_pct  REAL          NOT NULL,            -- 本次填写的完成度百分比
    achievement     VARCHAR(1000),                     -- 本周成果
    blocker         VARCHAR(1000),                     -- 阻塞问题
    lesson          VARCHAR(1000),                     -- 经验教训
    is_pm_correct   INTEGER       NOT NULL DEFAULT 0,  -- PM 是否已修正：0=否，1=是
    create_time     DATETIME      NOT NULL DEFAULT (datetime('now')) -- 记录创建时间
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_progress_records_biz_key ON pmw_progress_records(biz_key);
CREATE INDEX IF NOT EXISTS idx_progress_records_sub_item_key ON pmw_progress_records(sub_item_key);
CREATE INDEX IF NOT EXISTS idx_progress_records_sub_item_created ON pmw_progress_records(sub_item_key, create_time);
CREATE INDEX IF NOT EXISTS idx_progress_records_team_key ON pmw_progress_records(team_key);
CREATE INDEX IF NOT EXISTS idx_progress_records_create_time ON pmw_progress_records(create_time);

-- pmw_roles (RBAC)
CREATE TABLE IF NOT EXISTS pmw_roles (
    id              INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    biz_key         INTEGER       NOT NULL,            -- 业务唯一键
    create_time     DATETIME      NOT NULL DEFAULT (datetime('now')), -- 创建时间
    db_update_time  DATETIME      NOT NULL DEFAULT (datetime('now')), -- 数据库更新时间
    deleted_flag    INTEGER       NOT NULL DEFAULT 0,  -- 软删标志：0=正常，1=已删除
    deleted_time    DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00', -- 软删时间，未删除时为固定占位值
    role_name       VARCHAR(50)   NOT NULL,            -- 角色名称
    role_desc       VARCHAR(200)  NOT NULL DEFAULT '', -- 角色描述
    is_preset       INTEGER       NOT NULL DEFAULT 0   -- 是否预置角色：0=自定义，1=系统预置
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_roles_name ON pmw_roles(role_name, deleted_flag, deleted_time);

-- pmw_role_permissions (RBAC)
CREATE TABLE IF NOT EXISTS pmw_role_permissions (
    id               INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    deleted_flag     INTEGER      NOT NULL DEFAULT 0,   -- 软删标志：0=正常，1=已删除
    deleted_time     DATETIME     NOT NULL DEFAULT '1970-01-01 08:00:00', -- 软删时间，未删除时为固定占位值
    role_id          INTEGER      NOT NULL,             -- 角色 id（关联 pmw_roles.id）
    permission_code  VARCHAR(50)  NOT NULL,             -- 权限码，如 item:create、item:delete
    UNIQUE(role_id, permission_code, deleted_flag, deleted_time)
);

-- pmw_status_histories (append-only: no soft-delete, no biz_key)
CREATE TABLE IF NOT EXISTS pmw_status_histories (
    id           INTEGER PRIMARY KEY AUTOINCREMENT, -- 自增主键
    item_type    VARCHAR(20)  NOT NULL,             -- 事项类型：main_item=主事项，sub_item=子事项
    item_key     INTEGER      NOT NULL,             -- 事项 biz_key
    from_status  VARCHAR(20)  NOT NULL,             -- 变更前状态
    to_status    VARCHAR(20)  NOT NULL,             -- 变更后状态
    changed_by   INTEGER      NOT NULL,             -- 操作人 biz_key
    is_auto      INTEGER      NOT NULL DEFAULT 0,   -- 是否系统自动变更：0=人工，1=自动
    remark       VARCHAR(200),                      -- 变更备注
    create_time  DATETIME     NOT NULL DEFAULT (datetime('now')) -- 记录创建时间
);
CREATE INDEX IF NOT EXISTS idx_status_histories_item ON pmw_status_histories(item_type, item_key);
