package vo

import (
	"time"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/dates"
)

// ItemPoolVO is the frontend-facing view object for an item pool entry.
type ItemPoolVO struct {
	BizKey            string  `json:"bizKey"`
	TeamKey           string  `json:"teamKey"`
	Title             string  `json:"title"`
	Background        string  `json:"background"`
	ExpectedOutput    string  `json:"expectedOutput"`
	SubmitterKey      string  `json:"submitterKey"`
	SubmitterName     string  `json:"submitterName"`
	PoolStatus        string  `json:"poolStatus"`
	AssignedMainKey   *string `json:"assignedMainKey"`
	AssignedSubKey    *string `json:"assignedSubKey"`
	AssignedMainCode  string  `json:"assignedMainCode"`
	AssignedMainTitle string  `json:"assignedMainTitle"`
	AssigneeKey       *string `json:"assigneeKey"`
	RejectReason      string  `json:"rejectReason"`
	ReviewedAt        *string `json:"reviewedAt"`
	ReviewerKey       *string `json:"reviewerKey"`
	CreateTime        string  `json:"createTime"`
	DbUpdateTime      string  `json:"dbUpdateTime"`
}

// NewItemPoolVO converts a model.ItemPool to an ItemPoolVO.
func NewItemPoolVO(m *model.ItemPool, submitterName string) ItemPoolVO {
	return ItemPoolVO{
		BizKey:           pkg.FormatID(m.BizKey),
		TeamKey:          pkg.FormatID(m.TeamKey),
		Title:            m.Title,
		Background:       m.Background,
		ExpectedOutput:   m.ExpectedOutput,
		SubmitterKey:     pkg.FormatID(m.SubmitterKey),
		SubmitterName:    submitterName,
		PoolStatus:       m.PoolStatus,
		AssignedMainKey:  pkg.FormatIDPtr(m.AssignedMainKey),
		AssignedSubKey:   pkg.FormatIDPtr(m.AssignedSubKey),
		AssigneeKey:      pkg.FormatIDPtr(m.AssigneeKey),
		RejectReason:     m.RejectReason,
		ReviewedAt:       dates.FormatTimePtr(m.ReviewedAt),
		ReviewerKey:      pkg.FormatIDPtr(m.ReviewerKey),
		CreateTime:       m.CreateTime.Format(time.RFC3339),
		DbUpdateTime:     m.DbUpdateTime.Format(time.RFC3339),
	}
}
