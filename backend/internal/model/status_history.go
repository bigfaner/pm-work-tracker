package model

import "time"

// StatusHistory is an append-only log of status changes.
// Does NOT embed BaseModel — no UpdatedAt or soft delete needed.
type StatusHistory struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ItemType   string    `gorm:"type:varchar(20);not null;index:idx_item" json:"itemType"`
	ItemID     uint      `gorm:"not null;index:idx_item" json:"itemId"`
	FromStatus string    `gorm:"type:varchar(20);not null" json:"fromStatus"`
	ToStatus   string    `gorm:"type:varchar(20);not null" json:"toStatus"`
	ChangedBy  uint      `gorm:"not null" json:"changedBy"`
	IsAuto     bool      `gorm:"not null;default:false" json:"isAuto"`
	Remark     string    `gorm:"type:varchar(200)" json:"remark"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (StatusHistory) TableName() string {
	return "status_histories"
}
