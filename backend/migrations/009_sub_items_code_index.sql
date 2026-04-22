-- 009_sub_items_code_index.sql: Add unique index on sub_items(main_item_id, code)
-- Run after rewrite_codes has populated all sub_item codes.
-- precondition-skip-if: SELECT count(*) FROM sub_items WHERE code = '' OR code IS NULL
CREATE UNIQUE INDEX IF NOT EXISTS idx_sub_items_main_code ON sub_items(main_item_id, code);
