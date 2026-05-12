package model

import "time"

// MilestoneMap represents a milestone plan (里程碑图) belonging to a team.
type MilestoneMap struct {
	BaseModel
	TeamKey   int64  `gorm:"not null;index:idx_milestone_maps_team_status" json:"teamKey"`
	MapName   string `gorm:"type:varchar(100);not null" json:"mapName"`
	MapDesc   string `gorm:"type:text" json:"mapDesc"`
	MapStatus string `gorm:"type:varchar(20);not null;default:'planning';index:idx_milestone_maps_team_status" json:"mapStatus"`
}

// TableName returns the database table name for MilestoneMap.
func (MilestoneMap) TableName() string {
	return "pmw_milestone_maps"
}

// Milestone represents a single milestone within a milestone map.
type Milestone struct {
	BaseModel
	TeamKey         int64      `gorm:"not null;index:idx_milestones_team_key;index:idx_milestones_team_status" json:"teamKey"`
	MilestoneMapKey int64      `gorm:"not null;index:idx_milestones_milestone_map_key" json:"milestoneMapKey"`
	MilestoneName   string     `gorm:"type:varchar(100);not null" json:"milestoneName"`
	ExpectedEndDate *time.Time `json:"expectedEndDate"`
	MilestoneStatus string     `gorm:"type:varchar(20);not null;default:'not_started';index:idx_milestones_team_status" json:"milestoneStatus"`
}

// TableName returns the database table name for Milestone.
func (Milestone) TableName() string {
	return "pmw_milestones"
}
