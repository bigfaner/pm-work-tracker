package model

import "time"

// BaseModel replaces gorm.Model. ID is internal only (json:"-").
// BizKey is the external identifier exposed via API.
type BaseModel struct {
	ID           uint      `gorm:"primarykey;autoIncrement" json:"-"`
	BizKey       int64     `gorm:"not null" json:"bizKey"`
	CreateTime   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
	DbUpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"dbUpdateTime"`
	DeletedFlag  int       `gorm:"not null;default:0;index" json:"-"`
	DeletedTime  time.Time `gorm:"not null;default:'1970-01-01 08:00:00'" json:"-"`
}
