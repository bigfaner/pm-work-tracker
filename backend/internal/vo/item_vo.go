package vo

import (
	"time"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/status"
)

// MainItemVO is the frontend-facing view object for a main item.
type MainItemVO struct {
	ID              uint    `json:"id"`
	TeamID          uint    `json:"teamId"`
	Code            string  `json:"code"`
	Title           string  `json:"title"`
	Priority        string  `json:"priority"`
	ProposerID      uint    `json:"proposerId"`
	AssigneeID      *uint   `json:"assigneeId"`
	StartDate       *string `json:"startDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
	Status          string  `json:"status"`
	StatusName      string  `json:"statusName"`
	Completion      float64 `json:"completion"`
	IsKeyItem       bool    `json:"isKeyItem"`
	ArchivedAt      *string `json:"archivedAt"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

// SubItemVO is the frontend-facing view object for a sub item.
type SubItemVO struct {
	ID              uint    `json:"id"`
	TeamID          uint    `json:"teamId"`
	MainItemID      uint    `json:"mainItemId"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Priority        string  `json:"priority"`
	AssigneeID      *uint   `json:"assigneeId"`
	StartDate       *string `json:"startDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
	Status          string  `json:"status"`
	Completion      float64 `json:"completion"`
	IsKeyItem       bool    `json:"isKeyItem"`
	Weight          float64 `json:"weight"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

// ProgressRecordVO is the frontend-facing view object for a progress record.
type ProgressRecordVO struct {
	ID          uint    `json:"id"`
	SubItemID   uint    `json:"subItemId"`
	TeamID      uint    `json:"teamId"`
	AuthorID    uint    `json:"authorId"`
	AuthorName  string  `json:"authorName"`
	Completion  float64 `json:"completion"`
	Achievement string  `json:"achievement"`
	Blocker     string  `json:"blocker"`
	Lesson      string  `json:"lesson"`
	IsPMCorrect bool    `json:"isPMCorrect"`
	CreatedAt   string  `json:"createdAt"`
}

// SubItemSummaryVO is a lightweight sub-item summary for nesting in MainItemVO responses.
type SubItemSummaryVO struct {
	ID              uint    `json:"id"`
	Title           string  `json:"title"`
	Status          string  `json:"status"`
	StatusName      string  `json:"statusName"`
	Completion      float64 `json:"completion"`
	AssigneeID      *uint   `json:"assigneeId"`
	Priority        string  `json:"priority"`
	StartDate       *string `json:"startDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
}

// NewMainItemVO converts a model.MainItem to a MainItemVO.
func NewMainItemVO(m *model.MainItem) MainItemVO {
	statusName := ""
	if def, ok := status.GetMainItemStatus(m.Status); ok {
		statusName = def.Name
	}
	return MainItemVO{
		ID:              m.ID,
		TeamID:          m.TeamID,
		Code:            m.Code,
		Title:           m.Title,
		Priority:        m.Priority,
		ProposerID:      m.ProposerID,
		AssigneeID:      m.AssigneeID,
		StartDate:       formatTimePtr(m.StartDate),
		ExpectedEndDate: formatTimePtr(m.ExpectedEndDate),
		ActualEndDate:   formatTimePtr(m.ActualEndDate),
		Status:          m.Status,
		StatusName:      statusName,
		Completion:      m.Completion,
		IsKeyItem:       m.IsKeyItem,
		ArchivedAt:      formatTimePtr(m.ArchivedAt),
		CreatedAt:       m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       m.UpdatedAt.Format(time.RFC3339),
	}
}

// NewSubItemVO converts a model.SubItem to a SubItemVO.
func NewSubItemVO(m *model.SubItem) SubItemVO {
	return SubItemVO{
		ID:              m.ID,
		TeamID:          m.TeamID,
		MainItemID:      m.MainItemID,
		Title:           m.Title,
		Description:     m.Description,
		Priority:        m.Priority,
		AssigneeID:      m.AssigneeID,
		StartDate:       formatTimePtr(m.StartDate),
		ExpectedEndDate: formatTimePtr(m.ExpectedEndDate),
		ActualEndDate:   formatTimePtr(m.ActualEndDate),
		Status:          m.Status,
		Completion:      m.Completion,
		IsKeyItem:       m.IsKeyItem,
		Weight:          m.Weight,
		CreatedAt:       m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       m.UpdatedAt.Format(time.RFC3339),
	}
}

// NewProgressRecordVO converts a model.ProgressRecord to a ProgressRecordVO.
func NewProgressRecordVO(m *model.ProgressRecord, authorName string) ProgressRecordVO {
	return ProgressRecordVO{
		ID:          m.ID,
		SubItemID:   m.SubItemID,
		TeamID:      m.TeamID,
		AuthorID:    m.AuthorID,
		AuthorName:  authorName,
		Completion:  m.Completion,
		Achievement: m.Achievement,
		Blocker:     m.Blocker,
		Lesson:      m.Lesson,
		IsPMCorrect: m.IsPMCorrect,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
	}
}

// NewSubItemSummaryVOs converts model.SubItem slice to SubItemSummaryVO slice.
func NewSubItemSummaryVOs(items []*model.SubItem) []SubItemSummaryVO {
	result := make([]SubItemSummaryVO, 0, len(items))
	for _, si := range items {
		statusName := ""
		if def, ok := status.GetSubItemStatus(si.Status); ok {
			statusName = def.Name
		}
		result = append(result, SubItemSummaryVO{
			ID:              si.ID,
			Title:           si.Title,
			Status:          si.Status,
			StatusName:      statusName,
			Completion:      si.Completion,
			AssigneeID:      si.AssigneeID,
			Priority:        si.Priority,
			StartDate:       formatTimePtr(si.StartDate),
			ExpectedEndDate: formatTimePtr(si.ExpectedEndDate),
			ActualEndDate:   formatTimePtr(si.ActualEndDate),
		})
	}
	return result
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}
