package vo

import (
	"time"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/pkg/status"
)

// MainItemVO is the frontend-facing view object for a main item.
type MainItemVO struct {
	ID              uint    `json:"id"`
	TeamKey         int64   `json:"teamKey"`
	Code            string  `json:"code"`
	Title           string  `json:"title"`
	Priority        string  `json:"priority"`
	ProposerKey     int64   `json:"proposerKey"`
	AssigneeKey     *int64  `json:"assigneeKey"`
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
	Code            string  `json:"code"`
	TeamKey         int64   `json:"teamKey"`
	MainItemKey     int64   `json:"mainItemKey"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Priority        string  `json:"priority"`
	AssigneeKey     *int64  `json:"assigneeKey"`
	StartDate       *string `json:"startDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
	Status          string  `json:"status"`
	StatusName      string  `json:"statusName"`
	Completion      float64 `json:"completion"`
	IsKeyItem       bool    `json:"isKeyItem"`
	Weight          float64 `json:"weight"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

// ProgressRecordVO is the frontend-facing view object for a progress record.
type ProgressRecordVO struct {
	ID          uint    `json:"id"`
	SubItemKey  int64   `json:"subItemKey"`
	TeamKey     int64   `json:"teamKey"`
	AuthorKey   int64   `json:"authorKey"`
	AuthorName  string  `json:"authorName"`
	Completion  float64 `json:"completion"`
	Achievement string  `json:"achievement"`
	Blocker     string  `json:"blocker"`
	Lesson      string  `json:"lesson"`
	IsPmCorrect int     `json:"isPmCorrect"`
	CreatedAt   string  `json:"createdAt"`
}

// SubItemSummaryVO is a lightweight sub-item summary for nesting in MainItemVO responses.
type SubItemSummaryVO struct {
	ID              uint    `json:"id"`
	Code            string  `json:"code"`
	Title           string  `json:"title"`
	Status          string  `json:"status"`
	StatusName      string  `json:"statusName"`
	Completion      float64 `json:"completion"`
	AssigneeKey     *int64  `json:"assigneeKey"`
	Priority        string  `json:"priority"`
	StartDate       *string `json:"startDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
}

// NewMainItemVO converts a model.MainItem to a MainItemVO.
func NewMainItemVO(m *model.MainItem) MainItemVO {
	statusName := ""
	if def, ok := status.GetMainItemStatus(m.ItemStatus); ok {
		statusName = def.Name
	}
	return MainItemVO{
		ID:              m.ID,
		TeamKey:         m.TeamKey,
		Code:            m.Code,
		Title:           m.Title,
		Priority:        m.Priority,
		ProposerKey:     m.ProposerKey,
		AssigneeKey:     m.AssigneeKey,
		StartDate:       dates.FormatTimePtr(m.PlanStartDate),
		ExpectedEndDate: dates.FormatTimePtr(m.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(m.ActualEndDate),
		Status:          m.ItemStatus,
		StatusName:      statusName,
		Completion:      m.Completion,
		IsKeyItem:       m.IsKeyItem,
		ArchivedAt:      dates.FormatTimePtr(m.ArchivedAt),
		CreatedAt:       m.CreateTime.Format(time.RFC3339),
		UpdatedAt:       m.DbUpdateTime.Format(time.RFC3339),
	}
}

// NewSubItemVO converts a model.SubItem to a SubItemVO.
func NewSubItemVO(m *model.SubItem) SubItemVO {
	statusName := ""
	if def, ok := status.GetSubItemStatus(m.ItemStatus); ok {
		statusName = def.Name
	}
	return SubItemVO{
		ID:              m.ID,
		Code:            m.Code,
		TeamKey:         m.TeamKey,
		MainItemKey:     m.MainItemKey,
		Title:           m.Title,
		Description:     m.ItemDesc,
		Priority:        m.Priority,
		AssigneeKey:     m.AssigneeKey,
		StartDate:       dates.FormatTimePtr(m.PlanStartDate),
		ExpectedEndDate: dates.FormatTimePtr(m.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(m.ActualEndDate),
		Status:          m.ItemStatus,
		StatusName:      statusName,
		Completion:      m.Completion,
		IsKeyItem:       m.IsKeyItem,
		Weight:          m.Weight,
		CreatedAt:       m.CreateTime.Format(time.RFC3339),
		UpdatedAt:       m.DbUpdateTime.Format(time.RFC3339),
	}
}

// NewProgressRecordVO converts a model.ProgressRecord to a ProgressRecordVO.
func NewProgressRecordVO(m *model.ProgressRecord, authorName string) ProgressRecordVO {
	return ProgressRecordVO{
		ID:          m.ID,
		SubItemKey:  m.SubItemKey,
		TeamKey:     m.TeamKey,
		AuthorKey:   m.AuthorKey,
		AuthorName:  authorName,
		Completion:  m.Completion,
		Achievement: m.Achievement,
		Blocker:     m.Blocker,
		Lesson:      m.Lesson,
		IsPmCorrect: m.IsPmCorrect,
		CreatedAt:   m.CreateTime.Format(time.RFC3339),
	}
}

// NewSubItemSummaryVOs converts model.SubItem slice to SubItemSummaryVO slice.
func NewSubItemSummaryVOs(items []*model.SubItem) []SubItemSummaryVO {
	result := make([]SubItemSummaryVO, 0, len(items))
	for _, si := range items {
		statusName := ""
		if def, ok := status.GetSubItemStatus(si.ItemStatus); ok {
			statusName = def.Name
		}
		result = append(result, SubItemSummaryVO{
			ID:              si.ID,
			Code:            si.Code,
			Title:           si.Title,
			Status:          si.ItemStatus,
			StatusName:      statusName,
			Completion:      si.Completion,
			AssigneeKey:     si.AssigneeKey,
			Priority:        si.Priority,
			StartDate:       dates.FormatTimePtr(si.PlanStartDate),
			ExpectedEndDate: dates.FormatTimePtr(si.ExpectedEndDate),
			ActualEndDate:   dates.FormatTimePtr(si.ActualEndDate),
		})
	}
	return result
}
