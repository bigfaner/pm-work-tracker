package config

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

// SeedAdmin creates the initial admin user if configured and not already present.
// It is idempotent: if the username is empty or the user already exists, it returns nil.
func SeedAdmin(db *gorm.DB, authCfg *AuthConfig) error {
	if authCfg.InitialAdmin.Username == "" {
		return nil
	}

	var existing model.User
	result := db.Where("username = ?", authCfg.InitialAdmin.Username).First(&existing)
	if result.RowsAffected > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(authCfg.InitialAdmin.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	user := model.User{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		Username:     authCfg.InitialAdmin.Username,
		DisplayName:  authCfg.InitialAdmin.Username,
		PasswordHash: string(hash),
		IsSuperAdmin: true,
	}

	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	return nil
}
