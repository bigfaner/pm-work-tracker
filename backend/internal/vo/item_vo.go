package vo

import (
	"time"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/pkg/status"
)

// MainItemVO is the frontend-facing view object for a main item.
type MainItemVO struct {
	BizKey          string  `json:"bizKey"`
	TeamKey         string  `json:"teamKey"`
	Code            string  `json:"code"`
	Title           string  `json:"title"`
	ItemDesc        string  `json:"itemDesc"`
	Priority        string  `json:"priority"`
	ProposerKey     string  `json:"proposerKey"`
	AssigneeKey     *string `json:"assigneeKey"`
	PlanStartDate   *string `json:"planStartDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
	ItemStatus      string  `json:"itemStatus"`
	StatusName      string  `json:"statusName"`
	Completion      float64 `json:"completion"`
	IsKeyItem       bool    `json:"isKeyItem"`
	ArchivedAt      *string `json:"archivedAt"`
	CreateTime      string  `json:"createTime"`
	DbUpdateTime    string  `json:"dbUpdateTime"`
}

// SubItemVO is the frontend-facing view object for a sub item.
type SubItemVO struct {
	BizKey          string  `json:"bizKey"`
	Code            string  `json:"code"`
	TeamKey         string  `json:"teamKey"`
	MainItemKey     string  `json:"mainItemKey"`
	Title           string  `json:"title"`
	ItemDesc        string  `json:"itemDesc"`
	Priority        string  `json:"priority"`
	AssigneeKey     *string `json:"assigneeKey"`
	PlanStartDate   *string `json:"planStartDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
	ItemStatus      string  `json:"itemStatus"`
	StatusName      string  `json:"statusName"`
	Completion      float64 `json:"completion"`
	IsKeyItem       bool    `json:"isKeyItem"`
	Weight          float64 `json:"weight"`
	CreateTime      string  `json:"createTime"`
	DbUpdateTime    string  `json:"dbUpdateTime"`
}

// ProgressRecordVO is the frontend-facing view object for a progress record.
type ProgressRecordVO struct {
	BizKey      string  `json:"bizKey"`
	SubItemKey  string  `json:"subItemKey"`
	TeamKey     string  `json:"teamKey"`
	AuthorKey   string  `json:"authorKey"`
	AuthorName  string  `json:"authorName"`
	Completion  float64 `json:"completion"`
	Achievement string  `json:"achievement"`
	Blocker     string  `json:"blocker"`
	Lesson      string  `json:"lesson"`
	IsPmCorrect int     `json:"isPmCorrect"`
	CreateTime  string  `json:"createTime"`
}

// SubItemSummaryVO is a lightweight sub-item summary for nesting in MainItemVO responses.
type SubItemSummaryVO struct {
	BizKey          string  `json:"bizKey"`
	Code            string  `json:"code"`
	Title           string  `json:"title"`
	ItemStatus      string  `json:"itemStatus"`
	StatusName      string  `json:"statusName"`
	Completion      float64 `json:"completion"`
	AssigneeKey     *string `json:"assigneeKey"`
	Priority        string  `json:"priority"`
	PlanStartDate   *string `json:"planStartDate"`
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
		BizKey:          pkg.FormatID(m.BizKey),
		TeamKey:         pkg.FormatID(m.TeamKey),
		Code:            m.Code,
		Title:           m.Title,
		ItemDesc:        m.ItemDesc,
		Priority:        m.Priority,
		ProposerKey:     pkg.FormatID(m.ProposerKey),
		AssigneeKey:     pkg.FormatIDPtr(m.AssigneeKey),
		PlanStartDate:   dates.FormatTimePtr(m.PlanStartDate),
		ExpectedEndDate: dates.FormatTimePtr(m.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(m.ActualEndDate),
		ItemStatus:      m.ItemStatus,
		StatusName:      statusName,
		Completion:      m.Completion,
		IsKeyItem:       m.IsKeyItem,
		ArchivedAt:      dates.FormatTimePtr(m.ArchivedAt),
		CreateTime:      m.CreateTime.Format(time.RFC3339),
		DbUpdateTime:    m.DbUpdateTime.Format(time.RFC3339),
	}
}

// NewSubItemVO converts a model.SubItem to a SubItemVO.
func NewSubItemVO(m *model.SubItem) SubItemVO {
	statusName := ""
	if def, ok := status.GetSubItemStatus(m.ItemStatus); ok {
		statusName = def.Name
	}
	return SubItemVO{
		BizKey:          pkg.FormatID(m.BizKey),
		Code:            m.Code,
		TeamKey:         pkg.FormatID(m.TeamKey),
		MainItemKey:     pkg.FormatID(m.MainItemKey),
		Title:           m.Title,
		ItemDesc:        m.ItemDesc,
		Priority:        m.Priority,
		AssigneeKey:     pkg.FormatIDPtr(m.AssigneeKey),
		PlanStartDate:   dates.FormatTimePtr(m.PlanStartDate),
		ExpectedEndDate: dates.FormatTimePtr(m.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(m.ActualEndDate),
		ItemStatus:      m.ItemStatus,
		StatusName:      statusName,
		Completion:      m.Completion,
		IsKeyItem:       m.IsKeyItem,
		Weight:          m.Weight,
		CreateTime:      m.CreateTime.Format(time.RFC3339),
		DbUpdateTime:    m.DbUpdateTime.Format(time.RFC3339),
	}
}

// NewProgressRecordVO converts a model.ProgressRecord to a ProgressRecordVO.
func NewProgressRecordVO(m *model.ProgressRecord, authorName string) ProgressRecordVO {
	return ProgressRecordVO{
		BizKey:      pkg.FormatID(m.BizKey),
		SubItemKey:  pkg.FormatID(m.SubItemKey),
		TeamKey:     pkg.FormatID(m.TeamKey),
		AuthorKey:   pkg.FormatID(m.AuthorKey),
		AuthorName:  authorName,
		Completion:  m.Completion,
		Achievement: m.Achievement,
		Blocker:     m.Blocker,
		Lesson:      m.Lesson,
		IsPmCorrect: m.IsPmCorrect,
		CreateTime:  m.CreateTime.Format(time.RFC3339),
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
			BizKey:          pkg.FormatID(si.BizKey),
			Code:            si.Code,
			Title:           si.Title,
			ItemStatus:      si.ItemStatus,
			StatusName:      statusName,
			Completion:      si.Completion,
			AssigneeKey:     pkg.FormatIDPtr(si.AssigneeKey),
			Priority:        si.Priority,
			PlanStartDate:   dates.FormatTimePtr(si.PlanStartDate),
			ExpectedEndDate: dates.FormatTimePtr(si.ExpectedEndDate),
			ActualEndDate:   dates.FormatTimePtr(si.ActualEndDate),
		})
	}
	return result
}
