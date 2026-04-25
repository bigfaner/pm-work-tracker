package model

import "time"

type MainItem struct {
	BaseModel
	TeamKey         int64      `gorm:"not null;uniqueIndex:uk_main_items_team_code_deleted" json:"teamKey"`
	Code            string     `gorm:"type:varchar(12);not null;uniqueIndex:uk_main_items_team_code_deleted" json:"code"`
	Title           string     `gorm:"type:varchar(100);not null" json:"title"`
	ItemDesc        string     `gorm:"type:text;not null;default:''" json:"itemDesc"`
	Priority        string     `gorm:"type:varchar(5);not null" json:"priority"`
	ProposerKey     int64      `gorm:"not null" json:"proposerKey"`
	AssigneeKey     *int64     `gorm:"index" json:"assigneeKey"`
	PlanStartDate   *time.Time `json:"planStartDate"`
	ExpectedEndDate *time.Time `gorm:"index" json:"expectedEndDate"`
	ActualEndDate   *time.Time `json:"actualEndDate"`
	ItemStatus      string     `gorm:"type:varchar(20);not null;default:'pending'" json:"itemStatus"`
	Completion      float64    `gorm:"type:decimal(5,2);default:0" json:"completion"`
	IsKeyItem       bool       `gorm:"not null;default:false" json:"isKeyItem"`
	DelayCount      int        `gorm:"not null;default:0" json:"delayCount"`
	ArchivedAt      *time.Time `json:"archivedAt"`
	SubItemSeq      uint       `gorm:"not null;default:0" json:"subItemSeq"`
}

func (MainItem) TableName() string {
	return "pmw_main_items"
}
