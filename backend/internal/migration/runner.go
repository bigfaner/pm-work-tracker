package migration

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

// RunSchema executes a single schema SQL file against the database.
// Statements are split by ";" and executed individually because
// MySQL driver does not support multi-statement execution via db.Exec.
func RunSchema(db *gorm.DB, schemaFile string) error {
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("read schema %s: %w", schemaFile, err)
	}

	stmts := splitStatements(string(content))
	for i, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("execute schema statement #%d: %w", i+1, err)
		}
	}

	return nil
}

// splitStatements splits SQL text into individual statements,
// ignoring comments and empty lines.
func splitStatements(sql string) []string {
	var stmts []string
	for _, line := range strings.Split(sql, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
	}
	// Split by semicolons and keep non-empty statements.
	for _, part := range strings.Split(sql, ";") {
		part = strings.TrimSpace(part)
		// Strip leading comment lines within each part.
		var lines []string
		for _, l := range strings.Split(part, "\n") {
			l = strings.TrimSpace(l)
			if l != "" && !strings.HasPrefix(l, "--") {
				lines = append(lines, l)
			}
		}
		if len(lines) > 0 {
			stmts = append(stmts, strings.Join(lines, "\n"))
		}
	}
	return stmts
}
