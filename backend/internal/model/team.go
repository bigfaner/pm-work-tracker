package model

import "time"

// Team represents a project team with a PM and members.
type Team struct {
	BaseModel
	TeamName string `gorm:"type:varchar(100);not null" json:"teamName"`
	TeamDesc string `gorm:"type:varchar(500)" json:"teamDesc"`
	PmKey    int64  `gorm:"not null" json:"pmKey"`
	Code     string `gorm:"column:team_code;type:varchar(6);not null;uniqueIndex:idx_teams_code" json:"code"`
	ItemSeq  uint   `gorm:"not null;default:0" json:"itemSeq"`
}

// TableName returns the database table name for Team.
func (Team) TableName() string {
	return "pmw_teams"
}

// TeamMember represents a user's membership in a team.
type TeamMember struct {
	BaseModel
	TeamKey  int64     `gorm:"not null" json:"teamKey"`
	UserKey  int64     `gorm:"not null" json:"userKey"`
	RoleKey  *int64    `json:"roleKey"`
	JoinedAt time.Time `gorm:"not null" json:"joinedAt"`
}

// TableName returns the database table name for TeamMember.
func (TeamMember) TableName() string {
	return "pmw_team_members"
}
