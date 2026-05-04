package vo

import (
	"testing"
	"time"

	"pm-work-tracker/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestNewDecisionLogVO(t *testing.T) {
	now := time.Now()
	t.Run("parses tags from JSON string", func(t *testing.T) {
		m := &model.DecisionLog{
			BizKey:      123456789,
			MainItemKey: 987654321,
			Category:    "technical",
			Tags:        `["缓存策略","性能优化"]`,
			Content:     "采用 Redis 缓存策略",
			LogStatus:   "draft",
			CreatedBy:   111222333,
			CreateTime:  now,
			UpdateTime:  now,
		}
		vo := NewDecisionLogVO(m, "张三")
		assert.Equal(t, "123456789", vo.BizKey)
		assert.Equal(t, "987654321", vo.MainItemKey)
		assert.Equal(t, "technical", vo.Category)
		assert.Equal(t, []string{"缓存策略", "性能优化"}, vo.Tags)
		assert.Equal(t, "采用 Redis 缓存策略", vo.Content)
		assert.Equal(t, "draft", vo.LogStatus)
		assert.Equal(t, "111222333", vo.CreatedBy)
		assert.Equal(t, "张三", vo.CreatorName)
	})

	t.Run("empty tags returns empty slice", func(t *testing.T) {
		m := &model.DecisionLog{
			BizKey:      1,
			MainItemKey: 2,
			Tags:        "",
			CreateTime:  now,
			UpdateTime:  now,
		}
		vo := NewDecisionLogVO(m, "")
		assert.Equal(t, []string{}, vo.Tags)
	})

	t.Run("invalid JSON tags returns empty slice", func(t *testing.T) {
		m := &model.DecisionLog{
			BizKey:      1,
			MainItemKey: 2,
			Tags:        "not-json",
			CreateTime:  now,
			UpdateTime:  now,
		}
		vo := NewDecisionLogVO(m, "")
		assert.Equal(t, []string{}, vo.Tags)
	})

	t.Run("time formatted as RFC3339", func(t *testing.T) {
		ts := time.Date(2026, 5, 4, 10, 30, 0, 0, time.UTC)
		m := &model.DecisionLog{
			BizKey:     1,
			Tags:       "[]",
			CreateTime: ts,
			UpdateTime: ts,
		}
		vo := NewDecisionLogVO(m, "")
		assert.Equal(t, ts.Format(time.RFC3339), vo.CreateTime)
		assert.Equal(t, ts.Format(time.RFC3339), vo.UpdateTime)
	})
}
