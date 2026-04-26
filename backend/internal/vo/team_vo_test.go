package vo

import (
	"testing"
	"time"

	"pm-work-tracker/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestNewTeamVO(t *testing.T) {
	now := time.Now()
	team := &model.Team{
		BaseModel: model.BaseModel{
			ID:           1,
			BizKey:       12345,
			CreateTime:   now,
			DbUpdateTime: now,
		},
		TeamName: "Alpha",
		TeamDesc: "A team",
		Code:     "ALPHA",
		PmKey:    99,
	}

	result := NewTeamVO(team)

	assert.Equal(t, "12345", result.BizKey)
	assert.Equal(t, "Alpha", result.Name)
	assert.Equal(t, "A team", result.Description)
	assert.Equal(t, "ALPHA", result.Code)
	assert.Equal(t, "99", result.PmKey)
	assert.Equal(t, now.Format(time.RFC3339), result.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), result.UpdatedAt)
}
