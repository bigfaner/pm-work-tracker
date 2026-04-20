package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainItemStatuses_Count(t *testing.T) {
	assert.Len(t, MainItemStatuses, 7)
}

func TestSubItemStatuses_Count(t *testing.T) {
	assert.Len(t, SubItemStatuses, 6)
}

func TestMainItemStatuses_ContainsAll(t *testing.T) {
	expected := []string{"pending", "progressing", "blocking", "pausing", "reviewing", "completed", "closed"}
	for _, code := range expected {
		_, ok := MainItemStatuses[code]
		assert.True(t, ok, "MainItemStatuses should contain %q", code)
	}
}

func TestSubItemStatuses_ContainsAll(t *testing.T) {
	expected := []string{"pending", "progressing", "blocking", "pausing", "completed", "closed"}
	for _, code := range expected {
		_, ok := SubItemStatuses[code]
		assert.True(t, ok, "SubItemStatuses should contain %q", code)
	}
}

func TestMainItemStatuses_Definitions(t *testing.T) {
	tests := []struct {
		code     string
		name     string
		terminal bool
	}{
		{"pending", "待开始", false},
		{"progressing", "进行中", false},
		{"blocking", "阻塞中", false},
		{"pausing", "已暂停", false},
		{"reviewing", "待验收", false},
		{"completed", "已完成", true},
		{"closed", "已关闭", true},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			def, ok := MainItemStatuses[tt.code]
			assert.True(t, ok)
			assert.Equal(t, tt.code, def.Code)
			assert.Equal(t, tt.name, def.Name)
			assert.Equal(t, tt.terminal, def.Terminal)
		})
	}
}

func TestSubItemStatuses_Definitions(t *testing.T) {
	tests := []struct {
		code     string
		name     string
		terminal bool
	}{
		{"pending", "待开始", false},
		{"progressing", "进行中", false},
		{"blocking", "阻塞中", false},
		{"pausing", "已暂停", false},
		{"completed", "已完成", true},
		{"closed", "已关闭", true},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			def, ok := SubItemStatuses[tt.code]
			assert.True(t, ok)
			assert.Equal(t, tt.code, def.Code)
			assert.Equal(t, tt.name, def.Name)
			assert.Equal(t, tt.terminal, def.Terminal)
		})
	}
}

func TestGetMainItemStatus(t *testing.T) {
	def, ok := GetMainItemStatus("pending")
	assert.True(t, ok)
	assert.Equal(t, "待开始", def.Name)

	_, ok = GetMainItemStatus("unknown")
	assert.False(t, ok)
}

func TestGetSubItemStatus(t *testing.T) {
	def, ok := GetSubItemStatus("progressing")
	assert.True(t, ok)
	assert.Equal(t, "进行中", def.Name)

	_, ok = GetSubItemStatus("reviewing")
	assert.False(t, ok, "SubItemStatuses should not contain reviewing")

	_, ok = GetSubItemStatus("unknown")
	assert.False(t, ok)
}

func TestIsMainTerminal(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"completed", true},
		{"closed", true},
		{"pending", false},
		{"progressing", false},
		{"blocking", false},
		{"pausing", false},
		{"reviewing", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.want, IsMainTerminal(tt.code))
		})
	}
}

func TestIsSubTerminal(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"completed", true},
		{"closed", true},
		{"pending", false},
		{"progressing", false},
		{"blocking", false},
		{"pausing", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.want, IsSubTerminal(tt.code))
		})
	}
}
