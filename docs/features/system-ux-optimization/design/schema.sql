-- ============================================================
-- Schema: System UX Optimization Batch
-- Generated from: design/er-diagram.md
-- No DDL changes — seed data only
-- ============================================================
-- Note: Actual seed insertion is handled by Go migration code
-- (SyncPresetRoles with INSERT-IGNORE semantics). The SQL below
-- documents the expected seed data for reference only.
-- ============================================================

-- Seed: add delete permissions to pm preset role
-- These are added to the existing pm role's permission set.
-- The Go code in internal/pkg/permissions/codes.go must declare
-- the new codes before SyncPresetRoles runs.

-- Permission codes to add:
--   main_item:delete — PM can soft-delete main items (cascades to sub-items)
--   sub_item:delete  — PM can soft-delete individual sub-items

-- Idempotent INSERT (handled by SyncPresetRoles):
-- INSERT IGNORE INTO pmw_role_permissions (role_key, permission_code, deleted_flag, deleted_time)
--   SELECT biz_key, 'main_item:delete', 0, '1970-01-01 08:00:00'
--   FROM pmw_roles WHERE role_name = 'pm' AND deleted_flag = 0;
-- INSERT IGNORE INTO pmw_role_permissions (role_key, permission_code, deleted_flag, deleted_time)
--   SELECT biz_key, 'sub_item:delete', 0, '1970-01-01 08:00:00'
--   FROM pmw_roles WHERE role_name = 'pm' AND deleted_flag = 0;
