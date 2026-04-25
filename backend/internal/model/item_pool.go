package model

import "time"

// ItemPool status values: pending, assigned, rejected
type ItemPool struct {
	BaseModel
	TeamID          uint       `gorm:"not null;index" json:"teamId"`
	Title           string     `gorm:"type:varchar(100);not null" json:"title"`
	Background      string     `gorm:"type:text" json:"background"`
	ExpectedOutput  string     `gorm:"type:text" json:"expectedOutput"`
	SubmitterKey    int64      `gorm:"not null" json:"submitterKey"`
	PoolStatus      string     `gorm:"type:varchar(20);not null;default:'pending'" json:"poolStatus"`
	AssignedMainKey *int64     `json:"assignedMainKey"`
	AssignedSubKey  *int64     `json:"assignedSubKey"`
	AssigneeKey     *int64     `json:"assigneeKey"`
	RejectReason    string     `gorm:"type:varchar(200)" json:"rejectReason"`
	ReviewedAt      *time.Time `json:"reviewedAt"`
	ReviewerKey     *int64     `json:"reviewerKey"`
}

func (ItemPool) TableName() string {
	return "pmw_item_pools"
}
