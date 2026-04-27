package model

import "time"

// ProgressRecord is append-only: no soft-delete fields.
type ProgressRecord struct {
	ID          uint      `gorm:"primarykey;autoIncrement" json:"-"`
	BizKey      int64     `gorm:"not null" json:"bizKey"`
	SubItemKey  int64     `gorm:"not null" json:"subItemKey"`
	TeamKey     int64     `gorm:"not null" json:"teamKey"`
	AuthorKey   int64     `gorm:"not null" json:"authorKey"`
	Completion  float64   `gorm:"column:completion_pct;type:decimal(5,2);not null" json:"completion"`
	Achievement string    `gorm:"type:varchar(1000)" json:"achievement"`
	Blocker     string    `gorm:"type:varchar(1000)" json:"blocker"`
	Lesson      string    `gorm:"type:varchar(1000)" json:"lesson"`
	IsPmCorrect int       `gorm:"not null;default:0" json:"isPmCorrect"`
	CreateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
}

func (ProgressRecord) TableName() string {
	return "pmw_progress_records"
}
