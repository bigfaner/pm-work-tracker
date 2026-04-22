-- 008_item_code_redesign.sql: Add code columns and indexes for item code redesign

-- 1. teams: add code column and unique index
ALTER TABLE teams ADD COLUMN code VARCHAR(6) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_teams_code ON teams(code);

-- 2. sub_items: add code column (unique index added in 009 after rewrite_codes populates data)
ALTER TABLE sub_items ADD COLUMN code VARCHAR(15) NOT NULL DEFAULT '';
