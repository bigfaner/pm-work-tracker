package config

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB(cfg *DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "sqlite", "":
		path := cfg.Path
		if path == "" {
			path = "./data/dev.db"
		}
		dialector = sqlite.Open(path)
	case "mysql":
		if cfg.URL == "" {
			return nil, fmt.Errorf("mysql driver requires a non-empty url")
		}
		dialector = mysql.Open(cfg.URL)
	default:
		return nil, fmt.Errorf("unsupported driver: %s (use 'sqlite' or 'mysql')", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime.Value())

	return db, nil
}
