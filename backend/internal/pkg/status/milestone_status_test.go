package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMilestoneMapStatuses_Count(t *testing.T) {
	assert.Len(t, MilestoneMapStatuses, 5)
}

func TestMilestoneStatuses_Count(t *testing.T) {
	assert.Len(t, MilestoneStatuses, 4)
}

func TestMilestoneMapStatuses_ContainsAll(t *testing.T) {
	expected := []string{"planning", "reviewed", "ready", "executing", "completed"}
	for _, code := range expected {
		_, ok := MilestoneMapStatuses[code]
		assert.True(t, ok, "MilestoneMapStatuses should contain %q", code)
	}
}

func TestMilestoneStatuses_ContainsAll(t *testing.T) {
	expected := []string{"not_started", "in_progress", "completed", "cancelled"}
	for _, code := range expected {
		_, ok := MilestoneStatuses[code]
		assert.True(t, ok, "MilestoneStatuses should contain %q", code)
	}
}

func TestMilestoneMapStatuses_Definitions(t *testing.T) {
	tests := []struct {
		code     string
		name     string
		terminal bool
	}{
		{"planning", "规划中", false},
		{"reviewed", "已评审", false},
		{"ready", "就绪", false},
		{"executing", "执行中", false},
		{"completed", "已完成", true},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			def, ok := MilestoneMapStatuses[tt.code]
			assert.True(t, ok)
			assert.Equal(t, tt.code, def.Code)
			assert.Equal(t, tt.name, def.Name)
			assert.Equal(t, tt.terminal, def.Terminal)
		})
	}
}

func TestMilestoneStatuses_Definitions(t *testing.T) {
	tests := []struct {
		code     string
		name     string
		terminal bool
	}{
		{"not_started", "未开始", false},
		{"in_progress", "进行中", false},
		{"completed", "已完成", false},
		{"cancelled", "已取消", true},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			def, ok := MilestoneStatuses[tt.code]
			assert.True(t, ok)
			assert.Equal(t, tt.code, def.Code)
			assert.Equal(t, tt.name, def.Name)
			assert.Equal(t, tt.terminal, def.Terminal)
		})
	}
}

func TestGetMilestoneMapStatus(t *testing.T) {
	def, ok := GetMilestoneMapStatus("planning")
	assert.True(t, ok)
	assert.Equal(t, "规划中", def.Name)

	_, ok = GetMilestoneMapStatus("unknown")
	assert.False(t, ok)
}

func TestGetMilestoneStatus(t *testing.T) {
	def, ok := GetMilestoneStatus("not_started")
	assert.True(t, ok)
	assert.Equal(t, "未开始", def.Name)

	_, ok = GetMilestoneStatus("unknown")
	assert.False(t, ok)
}

func TestIsMilestoneMapTerminal(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"completed", true},
		{"planning", false},
		{"reviewed", false},
		{"ready", false},
		{"executing", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.want, IsMilestoneMapTerminal(tt.code))
		})
	}
}

func TestIsMilestoneTerminal(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"cancelled", true},
		{"not_started", false},
		{"in_progress", false},
		{"completed", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.want, IsMilestoneTerminal(tt.code))
		})
	}
}

func TestMilestoneCompletedIsNonTerminal(t *testing.T) {
	def, ok := MilestoneStatuses["completed"]
	assert.True(t, ok)
	assert.False(t, def.Terminal, "completed should be non-terminal for Milestone (can transition to cancelled)")
}
