package vo

import (
	"time"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/dates"
)

// ItemPoolVO is the frontend-facing view object for an item pool entry.
type ItemPoolVO struct {
	ID                uint    `json:"id"`
	TeamID            uint    `json:"teamId"`
	Title             string  `json:"title"`
	Background        string  `json:"background"`
	ExpectedOutput    string  `json:"expectedOutput"`
	SubmitterID       uint    `json:"submitterId"`
	SubmitterName     string  `json:"submitterName"`
	Status            string  `json:"status"`
	AssignedMainID    *uint   `json:"assignedMainId"`
	AssignedSubID     *uint   `json:"assignedSubId"`
	AssignedMainCode  string  `json:"assignedMainCode"`
	AssignedMainTitle string  `json:"assignedMainTitle"`
	AssigneeID        *uint   `json:"assigneeId"`
	RejectReason      string  `json:"rejectReason"`
	ReviewedAt        *string `json:"reviewedAt"`
	ReviewerID        *uint   `json:"reviewerId"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

// NewItemPoolVO converts a model.ItemPool to an ItemPoolVO.
func NewItemPoolVO(m *model.ItemPool, submitterName string) ItemPoolVO {
	return ItemPoolVO{
		ID:             m.ID,
		TeamID:         m.TeamID,
		Title:          m.Title,
		Background:     m.Background,
		ExpectedOutput: m.ExpectedOutput,
		SubmitterID:    m.SubmitterID,
		SubmitterName:  submitterName,
		Status:         m.Status,
		AssignedMainID: m.AssignedMainID,
		AssignedSubID:  m.AssignedSubID,
		AssigneeID:     m.AssigneeID,
		RejectReason:   m.RejectReason,
		ReviewedAt:     dates.FormatTimePtr(m.ReviewedAt),
		ReviewerID:     m.ReviewerID,
		CreatedAt:      m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      m.UpdatedAt.Format(time.RFC3339),
	}
}
