package model

import "time"

// Milestone represents a milestone entity within a milestone map.
type Milestone struct {
	BaseModel
	TeamKey         int64      `gorm:"not null;index:idx_milestones_team_status" json:"teamKey"`
	MilestoneMapKey int64      `gorm:"not null;uniqueIndex:uk_milestones_map_name_deleted" json:"milestoneMapKey"`
	MilestoneName   string     `gorm:"type:varchar(100);not null;uniqueIndex:uk_milestones_map_name_deleted" json:"milestoneName"`
	MilestoneDesc   string     `gorm:"type:varchar(2000);not null;default:''" json:"milestoneDesc"`
	ExpectedEndDate *time.Time `json:"expectedEndDate"`
	MilestoneStatus string     `gorm:"type:varchar(20);not null;default:'not_started';index:idx_milestones_team_status" json:"milestoneStatus"`
}

// TableName returns the database table name for Milestone.
func (Milestone) TableName() string {
	return "pmw_milestones"
}
