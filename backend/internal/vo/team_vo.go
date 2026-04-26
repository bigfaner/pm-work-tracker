package vo

import (
	"time"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
)

// TeamVO is the frontend-facing view object for a team.
type TeamVO struct {
	BizKey      string `json:"bizKey"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
	PmKey       string `json:"pmKey"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// NewTeamVO converts a model.Team to a TeamVO.
func NewTeamVO(t *model.Team) TeamVO {
	return TeamVO{
		BizKey:      pkg.FormatID(t.BizKey),
		Name:        t.TeamName,
		Description: t.TeamDesc,
		Code:        t.Code,
		PmKey:       pkg.FormatID(t.PmKey),
		CreatedAt:   t.CreateTime.Format(time.RFC3339),
		UpdatedAt:   t.DbUpdateTime.Format(time.RFC3339),
	}
}
