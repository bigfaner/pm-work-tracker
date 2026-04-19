package model

import (
	"time"
)

type MainItem struct {
	BaseModel
	TeamID          uint       `gorm:"not null;index" json:"team_id"`
	Code            string     `gorm:"type:varchar(10);not null;uniqueIndex" json:"code"`
	Title           string     `gorm:"type:varchar(100);not null" json:"title"`
	Description     string     `gorm:"type:text;not null;default:''" json:"description"`
	Priority        string     `gorm:"type:varchar(5);not null" json:"priority"`
	ProposerID      uint       `gorm:"not null" json:"proposer_id"`
	AssigneeID      *uint      `gorm:"index" json:"assignee_id"`
	StartDate       *time.Time `json:"start_date"`
	ExpectedEndDate *time.Time `gorm:"index" json:"expected_end_date"`
	ActualEndDate   *time.Time `json:"actual_end_date"`
	Status          string     `gorm:"type:varchar(20);not null;default:'待开始'" json:"status"`
	Completion      float64    `gorm:"default:0" json:"completion"`
	IsKeyItem       bool       `gorm:"not null;default:false" json:"is_key_item"`
	DelayCount      int        `gorm:"not null;default:0" json:"delay_count"`
	ArchivedAt      *time.Time `json:"archived_at"`
}

func (MainItem) TableName() string {
	return "main_items"
}
