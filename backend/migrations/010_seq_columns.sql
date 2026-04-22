-- 010_seq_columns.sql: Add sequence counter columns used for code generation
ALTER TABLE teams ADD COLUMN item_seq INTEGER NOT NULL DEFAULT 0;
ALTER TABLE main_items ADD COLUMN sub_item_seq INTEGER NOT NULL DEFAULT 0;
