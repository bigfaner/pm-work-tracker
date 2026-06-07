package vo

import (
	"time"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/pkg/status"
)

// MilestoneMapVO is the frontend-facing view object for a milestone map.
type MilestoneMapVO struct {
	BizKey          string  `json:"bizKey"`
	TeamKey         string  `json:"teamKey"`
	CreatorKey      string  `json:"creatorKey"`
	CreatorName     string  `json:"creatorName"`
	AssigneeKey     string  `json:"assigneeKey"`
	AssigneeName    string  `json:"assigneeName"`
	MapName         string  `json:"mapName"`
	MapDesc         string  `json:"mapDesc"`
	MapStatus       string  `json:"mapStatus"`
	StatusName      string  `json:"statusName"`
	PlanStartDate   *string `json:"planStartDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	MilestoneCount  int     `json:"milestoneCount"`
	ItemCount       int     `json:"itemCount"`
	OverallProgress float64 `json:"overallProgress"`
	CreateTime      string  `json:"createTime"`
	DbUpdateTime    string  `json:"dbUpdateTime"`
}

// MilestoneVO is the frontend-facing view object for a milestone.
type MilestoneVO struct {
	BizKey          string  `json:"bizKey"`
	TeamKey         string  `json:"teamKey"`
	MilestoneMapKey string  `json:"milestoneMapKey"`
	MilestoneName   string  `json:"milestoneName"`
	MilestoneDesc   string  `json:"milestoneDesc"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	MilestoneStatus string  `json:"milestoneStatus"`
	StatusName      string  `json:"statusName"`
	Completion      float64 `json:"completion"`
	RelatedMICount  int     `json:"relatedMICount"` //nolint:tagliatelle // MI is a domain acronym (Main Item); field name per tech-design spec
	CreateTime      string  `json:"createTime"`
	DbUpdateTime    string  `json:"dbUpdateTime"`
}

// NewMilestoneMapVO converts a model.MilestoneMap to a MilestoneMapVO.
// CreatorName and AssigneeName must be set by the caller via enrichment.
func NewMilestoneMapVO(m *model.MilestoneMap) MilestoneMapVO {
	statusName := ""
	if def, ok := status.GetMilestoneMapStatus(m.MapStatus); ok {
		statusName = def.Name
	}
	return MilestoneMapVO{
		BizKey:          pkg.FormatID(m.BizKey),
		TeamKey:         pkg.FormatID(m.TeamKey),
		CreatorKey:      pkg.FormatID(m.CreatorKey),
		AssigneeKey:     pkg.FormatID(m.AssigneeKey),
		MapName:         m.MapName,
		MapDesc:         m.MapDesc,
		MapStatus:       m.MapStatus,
		StatusName:      statusName,
		PlanStartDate:   dates.FormatTimePtr(m.PlanStartDate),
		ExpectedEndDate: dates.FormatTimePtr(m.ExpectedEndDate),
		CreateTime:      m.CreateTime.Format(time.RFC3339),
		DbUpdateTime:    m.DbUpdateTime.Format(time.RFC3339),
	}
}

// NewMilestoneVO converts a model.Milestone to a MilestoneVO.
// Completion and RelatedMICount must be set by the caller via enrichment.
func NewMilestoneVO(m *model.Milestone) MilestoneVO {
	statusName := ""
	if def, ok := status.GetMilestoneStatus(m.MilestoneStatus); ok {
		statusName = def.Name
	}
	return MilestoneVO{
		BizKey:          pkg.FormatID(m.BizKey),
		TeamKey:         pkg.FormatID(m.TeamKey),
		MilestoneMapKey: pkg.FormatID(m.MilestoneMapKey),
		MilestoneName:   m.MilestoneName,
		MilestoneDesc:   m.MilestoneDesc,
		ExpectedEndDate: dates.FormatTimePtr(m.ExpectedEndDate),
		MilestoneStatus: m.MilestoneStatus,
		StatusName:      statusName,
		CreateTime:      m.CreateTime.Format(time.RFC3339),
		DbUpdateTime:    m.DbUpdateTime.Format(time.RFC3339),
	}
}
