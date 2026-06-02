-- ============================================================
-- Schema: System UX Optimization Batch
-- Generated from: design/er-diagram.md
-- No DDL changes — seed data only
-- ============================================================

-- Seed data: add delete permissions to pm preset role
-- These INSERTs are idempotent (handled by SyncPresetRoles with INSERT-IGNORE semantics)

-- main_item:delete — allows PM users to soft-delete main items (cascades to sub-items)
-- INSERT IGNORE INTO pmw_role_permissions (role_key, permission_code, deleted_flag, deleted_time)
--   SELECT biz_key, 'main_item:delete', 0, '1970-01-01 08:00:00'
--   FROM pmw_roles WHERE role_name = 'pm' AND deleted_flag = 0;

-- sub_item:delete — allows PM users to soft-delete individual sub-items
-- INSERT IGNORE INTO pmw_role_permissions (role_key, permission_code, deleted_flag, deleted_time)
--   SELECT biz_key, 'sub_item:delete', 0, '1970-01-01 08:00:00'
--   FROM pmw_roles WHERE role_name = 'pm' AND deleted_flag = 0;

-- ============================================================
-- No new tables, no ALTER TABLE, no new indexes
-- ============================================================
