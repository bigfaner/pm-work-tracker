-- ============================================================================
-- V2: 唯一索引扩展 — 加入 deleted_flag + deleted_time（支持软删后重建）
-- 时间范围：2026-04-27 18:00 ~ 2026-04-28
-- 涉及 3 张表，均为 DROP + ADD INDEX，大表请在低峰期执行
-- ============================================================================

-- 1. pmw_roles: (role_name) → (role_name, deleted_flag, deleted_time)
ALTER TABLE `pmw_roles`
    DROP INDEX `uk_roles_name`,
    ADD UNIQUE KEY `uk_roles_name` (`role_name`, `deleted_flag`, `deleted_time`);

-- 2. pmw_sub_items: (main_item_key, item_code) → (main_item_key, item_code, deleted_flag, deleted_time)
ALTER TABLE `pmw_sub_items`
    DROP INDEX `uk_sub_items_main_code`,
    ADD UNIQUE KEY `uk_sub_items_main_code` (`main_item_key`, `item_code`, `deleted_flag`, `deleted_time`);

-- 3. pmw_role_permissions: 新增软删列 + 扩展唯一索引
ALTER TABLE `pmw_role_permissions`
    ADD COLUMN `deleted_flag` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '软删标志：0=正常，1=已删除' AFTER `id`,
    ADD COLUMN `deleted_time` DATETIME NOT NULL DEFAULT '1970-01-01 08:00:00' COMMENT '软删时间，未删除时为固定占位值' AFTER `deleted_flag`;

ALTER TABLE `pmw_role_permissions`
    DROP INDEX `uk_role_permission`,
    ADD UNIQUE KEY `uk_role_permission` (`role_id`, `permission_code`, `deleted_flag`, `deleted_time`);
