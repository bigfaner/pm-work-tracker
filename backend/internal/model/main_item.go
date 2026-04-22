package model

import (
	"time"
)

type MainItem struct {
	BaseModel
	TeamID          uint       `gorm:"not null;uniqueIndex:idx_main_items_team_code" json:"teamId"`
	Code            string     `gorm:"type:varchar(12);not null;uniqueIndex:idx_main_items_team_code" json:"code"`
	Title           string     `gorm:"type:varchar(100);not null" json:"title"`
	Description     string     `gorm:"type:text;not null;default:''" json:"description"`
	Priority        string     `gorm:"type:varchar(5);not null" json:"priority"`
	ProposerID      uint       `gorm:"not null" json:"proposerId"`
	AssigneeID      *uint      `gorm:"index" json:"assigneeId"`
	StartDate       *time.Time `json:"startDate"`
	ExpectedEndDate *time.Time `gorm:"index" json:"expectedEndDate"`
	ActualEndDate   *time.Time `json:"actualEndDate"`
	Status          string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Completion      float64    `gorm:"default:0" json:"completion"`
	IsKeyItem       bool       `gorm:"not null;default:false" json:"isKeyItem"`
	ArchivedAt      *time.Time `json:"archivedAt"`
}

func (MainItem) TableName() string {
	return "main_items"
}
