package model

import "time"

type Team struct {
	BaseModel
	TeamName string `gorm:"type:varchar(100);not null" json:"teamName"`
	TeamDesc string `gorm:"type:varchar(500)" json:"teamDesc"`
	PmKey    int64  `gorm:"not null" json:"pmKey"`
	Code     string `gorm:"column:team_code;type:varchar(6);not null;uniqueIndex:idx_teams_code" json:"code"`
	ItemSeq  uint   `gorm:"not null;default:0" json:"itemSeq"`
}

func (Team) TableName() string {
	return "pmw_teams"
}

type TeamMember struct {
	BaseModel
	TeamKey  int64     `gorm:"not null;uniqueIndex:idx_team_member" json:"teamKey"`
	UserKey  int64     `gorm:"not null;uniqueIndex:idx_team_member" json:"userKey"`
	RoleKey  *int64    `json:"roleKey"`
	JoinedAt time.Time `gorm:"not null" json:"joinedAt"`
}

func (TeamMember) TableName() string {
	return "pmw_team_members"
}
