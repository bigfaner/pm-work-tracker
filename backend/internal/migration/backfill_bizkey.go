package migration

import (
	"fmt"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/pkg/snowflake"
)

const bizKeyBackfillVersion = "bizkey_backfill_001"

// BackfillBizKeys assigns snowflake BizKeys to any record that has biz_key = 0.
// Each table is processed independently; unique indexes prevent duplicate biz_keys.
// Idempotent: tracked via schema_migrations.
func BackfillBizKeys(db *gorm.DB) error {
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	var count int64
	if err := db.Raw("SELECT count(*) FROM schema_migrations WHERE version = ?", bizKeyBackfillVersion).Scan(&count).Error; err != nil {
		return fmt.Errorf("check bizkey backfill migration status: %w", err)
	}
	if count > 0 {
		return nil
	}

	tables := []string{
		"pmw_sub_items",
		"pmw_progress_records",
		"pmw_item_pools",
		"pmw_team_members",
		"pmw_main_items",
		"pmw_teams",
		"pmw_users",
	}

	for _, table := range tables {
		if !tableExists(db, table) {
			continue
		}
		if err := backfillTableBizKeys(db, table); err != nil {
			return fmt.Errorf("backfill biz_key for %s: %w", table, err)
		}
	}

	if err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", bizKeyBackfillVersion).Error; err != nil {
		return fmt.Errorf("mark bizkey backfill migration: %w", err)
	}

	return nil
}

// backfillTableBizKeys finds all rows with biz_key = 0 and assigns a unique snowflake ID to each.
func backfillTableBizKeys(db *gorm.DB, table string) error {
	type idRow struct {
		ID uint
	}

	var rows []idRow
	if err := db.Table(table).Where("biz_key = 0").Find(&rows).Error; err != nil {
		return err
	}

	for _, row := range rows {
		if err := db.Table(table).Where("id = ?", row.ID).Update("biz_key", snowflake.Generate()).Error; err != nil {
			return err
		}
	}

	return nil
}
