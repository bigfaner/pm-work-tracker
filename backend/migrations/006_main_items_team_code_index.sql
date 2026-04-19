-- 006: Change main_items.code unique index from global to per-team
DROP INDEX IF EXISTS idx_main_items_code;
CREATE UNIQUE INDEX IF NOT EXISTS idx_main_items_team_code ON main_items(team_id, code);
