package model

import (
	"time"

	"gorm.io/gorm"
)

// ItemPool status values: 待分配, 已分配, 已拒绝
type ItemPool struct {
	gorm.Model
	TeamID         uint       `gorm:"not null;index" json:"team_id"`
	Title          string     `gorm:"type:varchar(100);not null" json:"title"`
	Background     string     `gorm:"type:text" json:"background"`
	ExpectedOutput string     `gorm:"type:text" json:"expected_output"`
	SubmitterID    uint       `gorm:"not null" json:"submitter_id"`
	Status         string     `gorm:"type:varchar(20);not null;default:'待分配'" json:"status"`
	AssignedMainID *uint      `json:"assigned_main_id"`
	AssignedSubID  *uint      `json:"assigned_sub_id"`
	AssigneeID     *uint      `json:"assignee_id"`
	RejectReason   string     `gorm:"type:varchar(200)" json:"reject_reason"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
	ReviewerID     *uint      `json:"reviewer_id"`
}

func (ItemPool) TableName() string {
	return "item_pools"
}
