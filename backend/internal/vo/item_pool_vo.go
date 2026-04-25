package vo

import (
	"time"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/dates"
)

// ItemPoolVO is the frontend-facing view object for an item pool entry.
type ItemPoolVO struct {
	ID                uint    `json:"id"`
	TeamKey           int64   `json:"teamKey"`
	Title             string  `json:"title"`
	Background        string  `json:"background"`
	ExpectedOutput    string  `json:"expectedOutput"`
	SubmitterKey      int64   `json:"submitterKey"`
	SubmitterName     string  `json:"submitterName"`
	Status            string  `json:"status"`
	AssignedMainKey   *int64  `json:"assignedMainKey"`
	AssignedSubKey    *int64  `json:"assignedSubKey"`
	AssignedMainCode  string  `json:"assignedMainCode"`
	AssignedMainTitle string  `json:"assignedMainTitle"`
	AssigneeKey       *int64  `json:"assigneeKey"`
	RejectReason      string  `json:"rejectReason"`
	ReviewedAt        *string `json:"reviewedAt"`
	ReviewerKey       *int64  `json:"reviewerKey"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

// NewItemPoolVO converts a model.ItemPool to an ItemPoolVO.
func NewItemPoolVO(m *model.ItemPool, submitterName string) ItemPoolVO {
	return ItemPoolVO{
		ID:             m.ID,
		TeamKey:        m.TeamKey,
		Title:          m.Title,
		Background:     m.Background,
		ExpectedOutput: m.ExpectedOutput,
		SubmitterKey:   m.SubmitterKey,
		SubmitterName:  submitterName,
		Status:         m.PoolStatus,
		AssignedMainKey: m.AssignedMainKey,
		AssignedSubKey:  m.AssignedSubKey,
		AssigneeKey:     m.AssigneeKey,
		RejectReason:   m.RejectReason,
		ReviewedAt:     dates.FormatTimePtr(m.ReviewedAt),
		ReviewerKey:    m.ReviewerKey,
		CreatedAt:      m.CreateTime.Format(time.RFC3339),
		UpdatedAt:      m.DbUpdateTime.Format(time.RFC3339),
	}
}
