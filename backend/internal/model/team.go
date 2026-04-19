package model

import (
	"time"
)

type Team struct {
	BaseModel
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	PmID        uint   `gorm:"not null" json:"pm_id"`
}

func (Team) TableName() string {
	return "teams"
}

type TeamMember struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	TeamID    uint      `gorm:"not null;uniqueIndex:idx_team_user" json:"team_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_team_user" json:"user_id"`
	Role      string    `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	JoinedAt  time.Time `gorm:"not null" json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TeamMember) TableName() string {
	return "team_members"
}
