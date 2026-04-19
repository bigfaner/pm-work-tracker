package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel replaces gorm.Model with proper JSON tags for API serialization.
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
