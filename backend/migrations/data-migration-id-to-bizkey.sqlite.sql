-- Data migration: replace internal IDs with biz_keys in all foreign-key columns.
--
-- Background: columns named *_key were originally populated with the table's
-- internal auto-increment `id`. This script rewrites every such column to hold
-- the corresponding `biz_key` (snowflake ID) instead.
--
-- Run once against an existing SQLite database BEFORE deploying the new code.
-- Requires SQLite >= 3.25.0 (RENAME COLUMN support).
--
-- Idempotency:
--   DDL step (RENAME COLUMN) is NOT idempotent — it will error on a second run
--   because role_id no longer exists. Run this script exactly once.
--   Data UPDATE steps are idempotent: the WHERE EXISTS guard matches only rows
--   whose value is still a small auto-increment id. Snowflake biz_keys (~2e18)
--   never collide with auto-increment ids (typically < 1e6), so a second run
--   of the UPDATE steps is a safe no-op.
--
-- SQLite supports transactional DDL: if any statement fails the entire
-- transaction (including the RENAME COLUMN) is rolled back automatically.

BEGIN;

-- ── 1. pmw_role_permissions ──────────────────────────────────────────────────
-- DDL: rename column role_id → role_key (run once only — not idempotent)
ALTER TABLE pmw_role_permissions RENAME COLUMN role_id TO role_key;

-- Data: role_key was storing pmw_roles.id, replace with pmw_roles.biz_key
UPDATE pmw_role_permissions
SET role_key = (
    SELECT biz_key FROM pmw_roles WHERE id = pmw_role_permissions.role_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_roles WHERE id = pmw_role_permissions.role_key
);

-- ── 2. pmw_teams ─────────────────────────────────────────────────────────────
-- pm_key: pmw_users.id → pmw_users.biz_key
UPDATE pmw_teams
SET pm_key = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_teams.pm_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_teams.pm_key
);

-- ── 3. pmw_team_members ──────────────────────────────────────────────────────
-- team_key: pmw_teams.id → pmw_teams.biz_key
UPDATE pmw_team_members
SET team_key = (
    SELECT biz_key FROM pmw_teams WHERE id = pmw_team_members.team_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_teams WHERE id = pmw_team_members.team_key
);

-- user_key: pmw_users.id → pmw_users.biz_key
UPDATE pmw_team_members
SET user_key = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_team_members.user_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_team_members.user_key
);

-- role_key: pmw_roles.id → pmw_roles.biz_key (nullable)
UPDATE pmw_team_members
SET role_key = (
    SELECT biz_key FROM pmw_roles WHERE id = pmw_team_members.role_key
)
WHERE role_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM pmw_roles WHERE id = pmw_team_members.role_key
);

-- ── 4. pmw_main_items ────────────────────────────────────────────────────────
-- team_key: pmw_teams.id → pmw_teams.biz_key
UPDATE pmw_main_items
SET team_key = (
    SELECT biz_key FROM pmw_teams WHERE id = pmw_main_items.team_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_teams WHERE id = pmw_main_items.team_key
);

-- proposer_key: pmw_users.id → pmw_users.biz_key
UPDATE pmw_main_items
SET proposer_key = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_main_items.proposer_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_main_items.proposer_key
);

-- assignee_key: pmw_users.id → pmw_users.biz_key (nullable)
UPDATE pmw_main_items
SET assignee_key = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_main_items.assignee_key
)
WHERE assignee_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_main_items.assignee_key
);

-- ── 5. pmw_sub_items ─────────────────────────────────────────────────────────
-- team_key: pmw_teams.id → pmw_teams.biz_key
UPDATE pmw_sub_items
SET team_key = (
    SELECT biz_key FROM pmw_teams WHERE id = pmw_sub_items.team_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_teams WHERE id = pmw_sub_items.team_key
);

-- main_item_key: pmw_main_items.id → pmw_main_items.biz_key
-- NOTE: run after pmw_main_items is updated; lookup is by .id which never changes.
UPDATE pmw_sub_items
SET main_item_key = (
    SELECT biz_key FROM pmw_main_items WHERE id = pmw_sub_items.main_item_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_main_items WHERE id = pmw_sub_items.main_item_key
);

-- assignee_key: pmw_users.id → pmw_users.biz_key (nullable)
UPDATE pmw_sub_items
SET assignee_key = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_sub_items.assignee_key
)
WHERE assignee_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_sub_items.assignee_key
);

-- ── 6. pmw_item_pools ────────────────────────────────────────────────────────
-- team_key
UPDATE pmw_item_pools
SET team_key = (
    SELECT biz_key FROM pmw_teams WHERE id = pmw_item_pools.team_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_teams WHERE id = pmw_item_pools.team_key
);

-- submitter_key
UPDATE pmw_item_pools
SET submitter_key = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_item_pools.submitter_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_item_pools.submitter_key
);

-- assignee_key (nullable)
UPDATE pmw_item_pools
SET assignee_key = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_item_pools.assignee_key
)
WHERE assignee_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_item_pools.assignee_key
);

-- reviewer_key (nullable)
UPDATE pmw_item_pools
SET reviewer_key = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_item_pools.reviewer_key
)
WHERE reviewer_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_item_pools.reviewer_key
);

-- assigned_main_key (nullable): pmw_main_items.id → pmw_main_items.biz_key
UPDATE pmw_item_pools
SET assigned_main_key = (
    SELECT biz_key FROM pmw_main_items WHERE id = pmw_item_pools.assigned_main_key
)
WHERE assigned_main_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM pmw_main_items WHERE id = pmw_item_pools.assigned_main_key
);

-- assigned_sub_key (nullable): pmw_sub_items.id → pmw_sub_items.biz_key
UPDATE pmw_item_pools
SET assigned_sub_key = (
    SELECT biz_key FROM pmw_sub_items WHERE id = pmw_item_pools.assigned_sub_key
)
WHERE assigned_sub_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM pmw_sub_items WHERE id = pmw_item_pools.assigned_sub_key
);

-- ── 7. pmw_progress_records ──────────────────────────────────────────────────
-- sub_item_key: pmw_sub_items.id → pmw_sub_items.biz_key
UPDATE pmw_progress_records
SET sub_item_key = (
    SELECT biz_key FROM pmw_sub_items WHERE id = pmw_progress_records.sub_item_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_sub_items WHERE id = pmw_progress_records.sub_item_key
);

-- team_key: pmw_teams.id → pmw_teams.biz_key
UPDATE pmw_progress_records
SET team_key = (
    SELECT biz_key FROM pmw_teams WHERE id = pmw_progress_records.team_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_teams WHERE id = pmw_progress_records.team_key
);

-- author_key: pmw_users.id → pmw_users.biz_key
UPDATE pmw_progress_records
SET author_key = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_progress_records.author_key
)
WHERE EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_progress_records.author_key
);

-- ── 8. pmw_status_histories ──────────────────────────────────────────────────
-- item_key references either main_item or sub_item depending on item_type
UPDATE pmw_status_histories
SET item_key = (
    SELECT biz_key FROM pmw_main_items WHERE id = pmw_status_histories.item_key
)
WHERE item_type = 'main_item'
  AND EXISTS (
    SELECT 1 FROM pmw_main_items WHERE id = pmw_status_histories.item_key
);

UPDATE pmw_status_histories
SET item_key = (
    SELECT biz_key FROM pmw_sub_items WHERE id = pmw_status_histories.item_key
)
WHERE item_type = 'sub_item'
  AND EXISTS (
    SELECT 1 FROM pmw_sub_items WHERE id = pmw_status_histories.item_key
);

-- changed_by: pmw_users.id → pmw_users.biz_key
UPDATE pmw_status_histories
SET changed_by = (
    SELECT biz_key FROM pmw_users WHERE id = pmw_status_histories.changed_by
)
WHERE EXISTS (
    SELECT 1 FROM pmw_users WHERE id = pmw_status_histories.changed_by
);

COMMIT;
