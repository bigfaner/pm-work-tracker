package model

import (
	"time"
)

// ProgressRecord is append-only: no UpdatedAt, no DeletedAt.
// Only Completion and IsPMCorrect may be updated (PM correction).
type ProgressRecord struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SubItemID   uint      `gorm:"not null;index" json:"sub_item_id"`
	TeamID      uint      `gorm:"not null;index" json:"team_id"`
	AuthorID    uint      `gorm:"not null" json:"author_id"`
	Completion  float64   `gorm:"not null" json:"completion"`
	Achievement string    `gorm:"type:text" json:"achievement"`
	Blocker     string    `gorm:"type:text" json:"blocker"`
	Lesson      string    `gorm:"type:text" json:"lesson"`
	IsPMCorrect bool      `gorm:"not null;default:false" json:"is_pm_correct"`
	CreatedAt   time.Time `gorm:"not null;index" json:"created_at"`
}

func (ProgressRecord) TableName() string {
	return "progress_records"
}
