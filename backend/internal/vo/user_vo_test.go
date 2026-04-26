package vo

import (
	"testing"

	"pm-work-tracker/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestNewUserVO(t *testing.T) {
	user := &model.User{
		BaseModel: model.BaseModel{
			ID:     1,
			BizKey: 42,
		},
		Username:     "alice",
		DisplayName:  "Alice",
		Email:        "alice@example.com",
		UserStatus:   "enabled",
		IsSuperAdmin: true,
	}

	result := NewUserVO(user)

	assert.Equal(t, "42", result.BizKey)
	assert.Equal(t, "alice", result.Username)
	assert.Equal(t, "Alice", result.DisplayName)
	assert.Equal(t, "alice@example.com", result.Email)
	assert.Equal(t, "enabled", result.Status)
	assert.True(t, result.IsSuperAdmin)
}
