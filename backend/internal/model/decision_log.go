package model

import "time"

// DecisionLog is an append-only model for recording decisions on main items.
// No BaseModel embedding — no soft-delete needed. Follows ProgressRecord pattern.
type DecisionLog struct {
	ID          uint      `gorm:"primarykey;autoIncrement" json:"-"`
	BizKey      int64     `gorm:"not null" json:"bizKey"`
	MainItemKey int64     `gorm:"not null;index" json:"mainItemKey"`
	TeamKey     int64     `gorm:"not null" json:"teamKey"`
	Category    string    `gorm:"type:varchar(20);not null" json:"category"`
	Tags        string    `gorm:"type:text;not null;default:''" json:"tags"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	LogStatus   string    `gorm:"type:varchar(10);not null;default:'draft'" json:"logStatus"`
	CreatedBy   int64     `gorm:"not null" json:"createdBy"`
	CreateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
	UpdateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updateTime"`
}

func (DecisionLog) TableName() string { return "pmw_decision_logs" }
