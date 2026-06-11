package model

import "time"

// MilestoneMap represents a milestone map entity within a team.
type MilestoneMap struct {
	BaseModel
	TeamKey         int64      `gorm:"not null;uniqueIndex:uk_milestone_maps_team_name_deleted;index:idx_milestone_maps_team_status" json:"teamKey"`
	CreatorKey      int64      `gorm:"not null" json:"creatorKey"`
	AssigneeKey     int64      `gorm:"not null" json:"assigneeKey"`
	MapName         string     `gorm:"type:varchar(100);not null;uniqueIndex:uk_milestone_maps_team_name_deleted" json:"mapName"`
	MapDesc         string     `gorm:"type:varchar(2000);not null;default:''" json:"mapDesc"`
	MapStatus       string     `gorm:"type:varchar(20);not null;default:'planning';index:idx_milestone_maps_team_status" json:"mapStatus"`
	PlanStartDate   *time.Time `json:"planStartDate"`
	ExpectedEndDate *time.Time `json:"expectedEndDate"`
}

// TableName returns the database table name for MilestoneMap.
func (MilestoneMap) TableName() string {
	return "pmw_milestone_maps"
}
