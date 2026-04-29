-- Data migration: replace internal IDs with biz_keys in all foreign-key columns.
--
-- Background: columns named *_key were originally populated with the table's
-- internal auto-increment `id`. This script rewrites every such column to hold
-- the corresponding `biz_key` (snowflake ID) instead.
--
-- Run once against an existing MySQL database BEFORE deploying the new code.
-- Requires MySQL >= 8.0 (RENAME COLUMN support).
--
-- Idempotency:
--   The DDL step (RENAME COLUMN) is NOT idempotent — it will error on a second
--   run because role_id no longer exists. Run this script exactly once.
--   Data UPDATE steps are idempotent: the JOIN matches only rows whose value is
--   still a small auto-increment id. Snowflake biz_keys (~2e18) never collide
--   with auto-increment ids (typically < 1e6), so re-running the UPDATEs is safe.
--
-- !! IMPORTANT: MySQL DDL (ALTER TABLE) causes an implicit commit and cannot be
--    rolled back. The DDL step is executed first, outside any transaction.
--    If the script is interrupted after the DDL but before the data UPDATEs,
--    re-run from the START TRANSACTION block below (skip the ALTER TABLE).

-- ── 1. pmw_role_permissions — DDL (outside transaction, not rollback-safe) ───
-- RENAME COLUMN in MySQL 8.0 automatically updates all index references,
-- so no need to drop and recreate uk_role_permission.
ALTER TABLE pmw_role_permissions
    RENAME COLUMN role_id TO role_key,
    MODIFY COLUMN role_key BIGINT NOT NULL COMMENT '角色 biz_key（关联 pmw_roles.biz_key）';

-- ── Data migration — wrapped in a transaction ────────────────────────────────
START TRANSACTION;

-- role_key data: was storing pmw_roles.id, replace with pmw_roles.biz_key
UPDATE pmw_role_permissions rp
    JOIN pmw_roles r ON r.id = rp.role_key
SET rp.role_key = r.biz_key;

-- ── 2. pmw_teams ─────────────────────────────────────────────────────────────
-- pm_key: pmw_users.id → pmw_users.biz_key
UPDATE pmw_teams t
    JOIN pmw_users u ON u.id = t.pm_key
SET t.pm_key = u.biz_key;

-- ── 3. pmw_team_members ──────────────────────────────────────────────────────
-- team_key: pmw_teams.id → pmw_teams.biz_key
UPDATE pmw_team_members tm
    JOIN pmw_teams t ON t.id = tm.team_key
SET tm.team_key = t.biz_key;

-- user_key: pmw_users.id → pmw_users.biz_key
UPDATE pmw_team_members tm
    JOIN pmw_users u ON u.id = tm.user_key
SET tm.user_key = u.biz_key;

-- role_key: pmw_roles.id → pmw_roles.biz_key (nullable)
UPDATE pmw_team_members tm
    JOIN pmw_roles r ON r.id = tm.role_key
SET tm.role_key = r.biz_key
WHERE tm.role_key IS NOT NULL;

-- ── 4. pmw_main_items ────────────────────────────────────────────────────────
-- team_key
UPDATE pmw_main_items mi
    JOIN pmw_teams t ON t.id = mi.team_key
SET mi.team_key = t.biz_key;

-- proposer_key
UPDATE pmw_main_items mi
    JOIN pmw_users u ON u.id = mi.proposer_key
SET mi.proposer_key = u.biz_key;

-- assignee_key (nullable)
UPDATE pmw_main_items mi
    JOIN pmw_users u ON u.id = mi.assignee_key
SET mi.assignee_key = u.biz_key
WHERE mi.assignee_key IS NOT NULL;

-- ── 5. pmw_sub_items ─────────────────────────────────────────────────────────
-- team_key
UPDATE pmw_sub_items si
    JOIN pmw_teams t ON t.id = si.team_key
SET si.team_key = t.biz_key;

-- main_item_key: pmw_main_items.id → pmw_main_items.biz_key
-- NOTE: join on .id which is stable; run after pmw_main_items data is updated.
UPDATE pmw_sub_items si
    JOIN pmw_main_items mi ON mi.id = si.main_item_key
SET si.main_item_key = mi.biz_key;

-- assignee_key (nullable)
UPDATE pmw_sub_items si
    JOIN pmw_users u ON u.id = si.assignee_key
SET si.assignee_key = u.biz_key
WHERE si.assignee_key IS NOT NULL;

-- ── 6. pmw_item_pools ────────────────────────────────────────────────────────
-- team_key
UPDATE pmw_item_pools ip
    JOIN pmw_teams t ON t.id = ip.team_key
SET ip.team_key = t.biz_key;

-- submitter_key
UPDATE pmw_item_pools ip
    JOIN pmw_users u ON u.id = ip.submitter_key
SET ip.submitter_key = u.biz_key;

-- assignee_key (nullable)
UPDATE pmw_item_pools ip
    JOIN pmw_users u ON u.id = ip.assignee_key
SET ip.assignee_key = u.biz_key
WHERE ip.assignee_key IS NOT NULL;

-- reviewer_key (nullable)
UPDATE pmw_item_pools ip
    JOIN pmw_users u ON u.id = ip.reviewer_key
SET ip.reviewer_key = u.biz_key
WHERE ip.reviewer_key IS NOT NULL;

-- assigned_main_key (nullable)
UPDATE pmw_item_pools ip
    JOIN pmw_main_items mi ON mi.id = ip.assigned_main_key
SET ip.assigned_main_key = mi.biz_key
WHERE ip.assigned_main_key IS NOT NULL;

-- assigned_sub_key (nullable)
UPDATE pmw_item_pools ip
    JOIN pmw_sub_items si ON si.id = ip.assigned_sub_key
SET ip.assigned_sub_key = si.biz_key
WHERE ip.assigned_sub_key IS NOT NULL;

-- ── 7. pmw_progress_records ──────────────────────────────────────────────────
-- sub_item_key
UPDATE pmw_progress_records pr
    JOIN pmw_sub_items si ON si.id = pr.sub_item_key
SET pr.sub_item_key = si.biz_key;

-- team_key
UPDATE pmw_progress_records pr
    JOIN pmw_teams t ON t.id = pr.team_key
SET pr.team_key = t.biz_key;

-- author_key
UPDATE pmw_progress_records pr
    JOIN pmw_users u ON u.id = pr.author_key
SET pr.author_key = u.biz_key;

-- ── 8. pmw_status_histories ──────────────────────────────────────────────────
-- item_key for main_item rows
UPDATE pmw_status_histories sh
    JOIN pmw_main_items mi ON mi.id = sh.item_key
SET sh.item_key = mi.biz_key
WHERE sh.item_type = 'main_item';

-- item_key for sub_item rows
UPDATE pmw_status_histories sh
    JOIN pmw_sub_items si ON si.id = sh.item_key
SET sh.item_key = si.biz_key
WHERE sh.item_type = 'sub_item';

-- changed_by
UPDATE pmw_status_histories sh
    JOIN pmw_users u ON u.id = sh.changed_by
SET sh.changed_by = u.biz_key;

COMMIT;
