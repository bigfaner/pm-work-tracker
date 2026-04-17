package config

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	switch driver {
	case "sqlite":
		path := os.Getenv("DB_PATH")
		if path == "" {
			path = "./data/dev.db"
		}
		return gorm.Open(sqlite.Open(path), &gorm.Config{})
	case "mysql":
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			return nil, fmt.Errorf("DATABASE_URL env var is required for mysql driver")
		}
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER: %s (use 'sqlite' or 'mysql')", driver)
	}
}
