-- 008_rollback.sql: Rollback for 008_item_code_redesign.sql
--
-- NOTE: SQLite < 3.35.0 does not support DROP COLUMN.
-- On SQLite, rolling back requires rebuilding the affected tables.
-- On MySQL/PostgreSQL, the statements below work directly.
--
-- Steps to rollback:
-- 1. Drop composite unique index on sub_items
-- 2. Remove sub_items.code column (rebuild table on SQLite)
-- 3. Drop unique index on teams
-- 4. Remove teams.code column (rebuild table on SQLite)

-- MySQL / PostgreSQL:
-- DROP INDEX idx_sub_items_main_code ON sub_items;
-- ALTER TABLE sub_items DROP COLUMN code;
-- DROP INDEX idx_teams_code ON teams;
-- ALTER TABLE teams DROP COLUMN code;

-- SQLite (3.35.0+):
-- DROP INDEX IF EXISTS idx_sub_items_main_code;
-- ALTER TABLE sub_items DROP COLUMN code;
-- DROP INDEX IF EXISTS idx_teams_code;
-- ALTER TABLE teams DROP COLUMN code;

-- SQLite (< 3.35.0): rebuild tables without the code column.
-- See https://www.sqlite.org/lang_altertable.html for the 12-step procedure.
