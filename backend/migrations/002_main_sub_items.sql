-- 002_main_sub_items.sql: Add missing composite index on sub_items
-- SubItems needs (team_id, priority) composite index for filtered queries.
-- main_items already has both (team_id, status) and (team_id, priority) from 001.

CREATE INDEX IF NOT EXISTS idx_sub_items_team_priority ON sub_items(team_id, priority);
