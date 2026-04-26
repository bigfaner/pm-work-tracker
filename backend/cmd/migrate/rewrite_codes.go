package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/model"

	"gorm.io/gorm"
)

// errDryRun is a sentinel used to roll back the transaction in dry-run mode.
var errDryRun = fmt.Errorf("dry-run: rolling back (no changes committed)")

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	dryRun := flag.Bool("dry-run", false, "print planned changes without committing")
	flag.Parse()

	if err := run(*configPath, *dryRun); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func run(configPath string, dryRun bool) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	return migrate(context.Background(), db, dryRun)
}

// migrate performs the code rewrite migration inside a single transaction.
// If dryRun is true the transaction is always rolled back.
func migrate(ctx context.Context, db *gorm.DB, dryRun bool) error {
	// Pre-condition: every team must have a non-empty code.
	var teamsWithoutCode []model.Team
	if err := db.WithContext(ctx).Where("code = '' OR code IS NULL").Find(&teamsWithoutCode).Error; err != nil {
		return fmt.Errorf("check teams: %w", err)
	}
	if len(teamsWithoutCode) > 0 {
		return fmt.Errorf("pre-condition failed: %d team(s) have empty code — set team codes before running migration", len(teamsWithoutCode))
	}

	// Print record counts for manual confirmation.
	var mainItemCount, subItemCount int64
	if err := db.WithContext(ctx).Model(&model.MainItem{}).Count(&mainItemCount).Error; err != nil {
		return fmt.Errorf("count main_items: %w", err)
	}
	if err := db.WithContext(ctx).Model(&model.SubItem{}).Count(&subItemCount).Error; err != nil {
		return fmt.Errorf("count sub_items: %w", err)
	}
	log.Printf("records to process: %d main_items, %d sub_items", mainItemCount, subItemCount)
	if dryRun {
		log.Println("dry-run mode: changes will be printed but not committed")
	}

	txErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// newMainItemCodes tracks the computed new code for each main_item ID so
		// sub_item codes can be derived correctly even in dry-run mode.
		newMainItemCodes := make(map[uint]string)

		// Step 1: rewrite main_items.code — group by team, order by id.
		var teams []model.Team
		if err := tx.Order("id").Find(&teams).Error; err != nil {
			return fmt.Errorf("load teams: %w", err)
		}

		for _, team := range teams {
			var items []model.MainItem
			if err := tx.Where("team_key = ?", team.ID).Order("id").Find(&items).Error; err != nil {
				return fmt.Errorf("load main_items for team %d: %w", team.ID, err)
			}
			for i, item := range items {
				newCode := fmt.Sprintf("%s-%05d", team.Code, i+1)
				newMainItemCodes[item.ID] = newCode
				if dryRun {
					log.Printf("[dry-run] main_item id=%d: %q -> %q", item.ID, item.Code, newCode)
					continue
				}
				if err := tx.Model(&model.MainItem{}).Where("id = ?", item.ID).Update("code", newCode).Error; err != nil {
					return fmt.Errorf("update main_item id=%d: %w", item.ID, err)
				}
			}
		}

		// Step 2: generate sub_items.code — group by main_item, order by id.
		var mainItems []model.MainItem
		if err := tx.Order("id").Find(&mainItems).Error; err != nil {
			return fmt.Errorf("reload main_items: %w", err)
		}

		for _, mainItem := range mainItems {
			// Use the newly computed code (handles dry-run where DB wasn't updated).
			mainCode, ok := newMainItemCodes[mainItem.ID]
			if !ok {
				mainCode = mainItem.Code
			}

			var subs []model.SubItem
			if err := tx.Where("main_item_key = ?", mainItem.ID).Order("id").Find(&subs).Error; err != nil {
				return fmt.Errorf("load sub_items for main_item %d: %w", mainItem.ID, err)
			}
			for i, sub := range subs {
				newCode := fmt.Sprintf("%s-%02d", mainCode, i+1)
				if dryRun {
					log.Printf("[dry-run] sub_item id=%d: %q -> %q", sub.ID, sub.Code, newCode)
					continue
				}
				if err := tx.Model(&model.SubItem{}).Where("id = ?", sub.ID).Update("code", newCode).Error; err != nil {
					return fmt.Errorf("update sub_item id=%d: %w", sub.ID, err)
				}
			}
		}

		if dryRun {
			return errDryRun
		}

		// Post-migration validation.
		var oldFormatCount int64
		if err := tx.Model(&model.MainItem{}).Where("code LIKE 'MI-%'").Count(&oldFormatCount).Error; err != nil {
			return fmt.Errorf("validation (main_items old format): %w", err)
		}
		if oldFormatCount > 0 {
			return fmt.Errorf("validation failed: %d main_item(s) still have MI- prefix", oldFormatCount)
		}

		var emptySubCodeCount int64
		if err := tx.Model(&model.SubItem{}).Where("code IS NULL OR code = ''").Count(&emptySubCodeCount).Error; err != nil {
			return fmt.Errorf("validation (sub_items empty code): %w", err)
		}
		if emptySubCodeCount > 0 {
			return fmt.Errorf("validation failed: %d sub_item(s) have empty code", emptySubCodeCount)
		}

		log.Println("validation passed")
		return nil
	})

	if dryRun {
		log.Println("dry-run complete — transaction rolled back, no changes persisted")
		return nil
	}
	if txErr != nil {
		return fmt.Errorf("transaction rolled back: %w", txErr)
	}

	log.Println("migration complete")
	return nil
}
