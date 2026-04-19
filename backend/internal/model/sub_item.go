package model

import (
	"time"
)

// SubItem status values (8 total):
// 待开始, 进行中, 阻塞中, 挂起, 已延期, 待验收, 已完成, 已关闭
type SubItem struct {
	BaseModel
	TeamID          uint       `gorm:"not null;index" json:"teamId"`
	MainItemID      uint       `gorm:"not null;index" json:"mainItemId"`
	Title           string     `gorm:"type:varchar(100);not null" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	Priority        string     `gorm:"type:varchar(5);not null" json:"priority"`
	AssigneeID      *uint      `gorm:"index" json:"assigneeId"`
	StartDate       *time.Time `json:"startDate"`
	ExpectedEndDate *time.Time `gorm:"index" json:"expectedEndDate"`
	ActualEndDate   *time.Time `json:"actualEndDate"`
	Status          string     `gorm:"type:varchar(20);not null;default:'待开始'" json:"status"`
	Completion      float64    `gorm:"default:0" json:"completion"`
	IsKeyItem       bool       `gorm:"not null;default:false" json:"isKeyItem"`
	DelayCount      int        `gorm:"not null;default:0" json:"delayCount"`
	Weight          float64    `gorm:"default:1" json:"weight"`
}

func (SubItem) TableName() string {
	return "sub_items"
}
