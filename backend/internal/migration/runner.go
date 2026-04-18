package migration

import (
	"fmt"
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

		if err := r.apply(f, string(content)); err != nil {
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
