package model

import (
	"time"

	"gorm.io/gorm"
)

// SubItem status values (8 total):
// 待开始, 进行中, 阻塞中, 挂起, 已延期, 待验收, 已完成, 已关闭
type SubItem struct {
	gorm.Model
	TeamID          uint       `gorm:"not null;index" json:"team_id"`
	MainItemID      uint       `gorm:"not null;index" json:"main_item_id"`
	Title           string     `gorm:"type:varchar(100);not null" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	Priority        string     `gorm:"type:varchar(5);not null" json:"priority"`
	AssigneeID      *uint      `gorm:"index" json:"assignee_id"`
	StartDate       *time.Time `json:"start_date"`
	ExpectedEndDate *time.Time `gorm:"index" json:"expected_end_date"`
	ActualEndDate   *time.Time `json:"actual_end_date"`
	Status          string     `gorm:"type:varchar(20);not null;default:'待开始'" json:"status"`
	Completion      float64    `gorm:"default:0" json:"completion"`
	IsKeyItem       bool       `gorm:"not null;default:false" json:"is_key_item"`
	DelayCount      int        `gorm:"not null;default:0" json:"delay_count"`
	Weight          float64    `gorm:"default:1" json:"weight"`
}

func (SubItem) TableName() string {
	return "sub_items"
}
