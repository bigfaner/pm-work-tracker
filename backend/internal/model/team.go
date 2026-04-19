package model

import (
	"time"
)

type Team struct {
	BaseModel
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	PmID        uint   `gorm:"not null" json:"pmId"`
}

func (Team) TableName() string {
	return "teams"
}

type TeamMember struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	TeamID    uint      `gorm:"not null;uniqueIndex:idx_team_user" json:"teamId"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_team_user" json:"userId"`
	Role      string    `gorm:"-" json:"role"`
	RoleID    *uint     `gorm:"index" json:"roleId"`
	JoinedAt  time.Time `gorm:"not null" json:"joinedAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (TeamMember) TableName() string {
	return "team_members"
}
