package model

import (
	"time"
)

// ItemPool status values: pending, assigned, rejected
type ItemPool struct {
	BaseModel
	TeamID         uint       `gorm:"not null;index" json:"teamId"`
	Title          string     `gorm:"type:varchar(100);not null" json:"title"`
	Background     string     `gorm:"type:text" json:"background"`
	ExpectedOutput string     `gorm:"type:text" json:"expectedOutput"`
	SubmitterID    uint       `gorm:"not null" json:"submitterId"`
	Status         string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	AssignedMainID *uint      `json:"assignedMainId"`
	AssignedSubID  *uint      `json:"assignedSubId"`
	AssigneeID     *uint      `json:"assigneeId"`
	RejectReason   string     `gorm:"type:varchar(200)" json:"rejectReason"`
	ReviewedAt     *time.Time `json:"reviewedAt"`
	ReviewerID     *uint      `json:"reviewerId"`
}

func (ItemPool) TableName() string {
	return "item_pools"
}
