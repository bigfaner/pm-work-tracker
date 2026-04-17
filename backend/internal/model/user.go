package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username       string `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	DisplayName    string `gorm:"type:varchar(64);not null" json:"display_name"`
	PasswordHash   string `gorm:"type:varchar(255);not null" json:"-"`
	IsSuperAdmin   bool   `gorm:"not null;default:false" json:"is_super_admin"`
	CanCreateTeam  bool   `gorm:"not null;default:false" json:"can_create_team"`
}

func (User) TableName() string {
	return "users"
}
