package migration

import (
	"fmt"
	"os"

	"gorm.io/gorm"
)

// RunSchema executes a single schema SQL file against the database.
// The file should use CREATE TABLE IF NOT EXISTS for idempotency.
func RunSchema(db *gorm.DB, schemaFile string) error {
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("read schema %s: %w", schemaFile, err)
	}

	if err := db.Exec(string(content)).Error; err != nil {
		return fmt.Errorf("execute schema %s: %w", schemaFile, err)
	}

	return nil
}
