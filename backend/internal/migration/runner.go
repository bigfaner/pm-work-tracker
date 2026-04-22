package migration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type Runner struct {
	db   *gorm.DB
	dir  string
}

func NewRunner(db *gorm.DB, migrationsDir string) *Runner {
	return &Runner{db: db, dir: migrationsDir}
}

func (r *Runner) Run() error {
	if err := r.ensureSchemaMigrationsTable(); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	files, err := r.readMigrationFiles()
	if err != nil {
		return fmt.Errorf("read migration files: %w", err)
	}

	for _, f := range files {
		applied, err := r.isApplied(f)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", f, err)
		}
		if applied {
			continue
		}

		content, err := os.ReadFile(filepath.Join(r.dir, f))
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}

		sql := string(content)
		if precond := parsePreconditionSkipIf(sql); precond != "" {
			var count int64
			if err := r.db.Raw(precond).Scan(&count).Error; err != nil {
				return fmt.Errorf("precondition check for %s: %w", f, err)
			}
			if count > 0 {
				log.Printf("skipping migration %s: precondition not met (%d blocking rows)", f, count)
				continue
			}
		}

		if err := r.apply(f, sql); err != nil {
			return fmt.Errorf("apply %s: %w", f, err)
		}
	}

	return nil
}

func (r *Runner) ensureSchemaMigrationsTable() error {
	sql := `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version  VARCHAR(255) PRIMARY KEY,
    applied  DATETIME NOT NULL DEFAULT (datetime('now'))
);
`
	return r.db.Exec(sql).Error
}

func (r *Runner) readMigrationFiles() ([]string, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}

func (r *Runner) isApplied(filename string) (bool, error) {
	var count int64
	err := r.db.Raw("SELECT count(*) FROM schema_migrations WHERE version = ?", filename).Scan(&count).Error
	return count > 0, err
}

func (r *Runner) apply(filename, sql string) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Exec(sql).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", filename).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// parsePreconditionSkipIf extracts the SQL query from a line of the form:
//
//	-- precondition-skip-if: SELECT count(*) FROM ...
//
// Returns empty string if no such directive is found.
func parsePreconditionSkipIf(content string) string {
	const prefix = "-- precondition-skip-if:"
	for _, line := range strings.SplitN(content, "\n", 20) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(line[len(prefix):])
		}
	}
	return ""
}
