package model

import "time"

// StatusHistory is an append-only log of status changes.
// Does NOT embed BaseModel — no biz_key or soft-delete needed.
type StatusHistory struct {
	ID         uint      `gorm:"primarykey;autoIncrement" json:"-"`
	ItemType   string    `gorm:"type:varchar(20);not null" json:"itemType"`
	ItemKey    int64     `gorm:"not null" json:"itemKey"`
	FromStatus string    `gorm:"type:varchar(20);not null" json:"fromStatus"`
	ToStatus   string    `gorm:"type:varchar(20);not null" json:"toStatus"`
	ChangedBy  int64     `gorm:"not null" json:"changedBy"`
	IsAuto     int       `gorm:"not null;default:0" json:"isAuto"`
	Remark     string    `gorm:"type:varchar(200)" json:"remark"`
	CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
}

func (StatusHistory) TableName() string {
	return "pmw_status_histories"
}
