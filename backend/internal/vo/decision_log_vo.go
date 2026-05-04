package vo

import (
	"encoding/json"
	"time"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
)

// DecisionLogVO is the frontend-facing view object for a decision log.
type DecisionLogVO struct {
	BizKey      string   `json:"bizKey"`
	MainItemKey string   `json:"mainItemKey"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Content     string   `json:"content"`
	LogStatus   string   `json:"logStatus"`
	CreatedBy   string   `json:"createdBy"`
	CreatorName string   `json:"creatorName"`
	CreateTime  string   `json:"createTime"`
	UpdateTime  string   `json:"updateTime"`
}

// NewDecisionLogVO converts a model.DecisionLog to a DecisionLogVO.
// Tags are parsed from JSON string to []string. CreatorName must be resolved by caller.
func NewDecisionLogVO(m *model.DecisionLog, creatorName string) DecisionLogVO {
	var tags []string
	if m.Tags != "" {
		_ = json.Unmarshal([]byte(m.Tags), &tags)
	}
	if tags == nil {
		tags = []string{}
	}
	return DecisionLogVO{
		BizKey:      pkg.FormatID(m.BizKey),
		MainItemKey: pkg.FormatID(m.MainItemKey),
		Category:    m.Category,
		Tags:        tags,
		Content:     m.Content,
		LogStatus:   m.LogStatus,
		CreatedBy:   pkg.FormatID(m.CreatedBy),
		CreatorName: creatorName,
		CreateTime:  m.CreateTime.Format(time.RFC3339),
		UpdateTime:  m.UpdateTime.Format(time.RFC3339),
	}
}
