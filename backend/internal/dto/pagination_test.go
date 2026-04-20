package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyPaginationDefaults_ZeroValues(t *testing.T) {
	offset, page, pageSize := ApplyPaginationDefaults(0, 0)
	assert.Equal(t, 0, offset)
	assert.Equal(t, 1, page)
	assert.Equal(t, 20, pageSize)
}

func TestApplyPaginationDefaults_NegativeValues(t *testing.T) {
	offset, page, pageSize := ApplyPaginationDefaults(-1, -5)
	assert.Equal(t, 0, offset)
	assert.Equal(t, 1, page)
	assert.Equal(t, 20, pageSize)
}

func TestApplyPaginationDefaults_NormalValues(t *testing.T) {
	offset, page, pageSize := ApplyPaginationDefaults(3, 10)
	assert.Equal(t, 20, offset)
	assert.Equal(t, 3, page)
	assert.Equal(t, 10, pageSize)
}

func TestApplyPaginationDefaults_FirstPage(t *testing.T) {
	offset, page, pageSize := ApplyPaginationDefaults(1, 50)
	assert.Equal(t, 0, offset)
	assert.Equal(t, 1, page)
	assert.Equal(t, 50, pageSize)
}

func TestApplyPaginationDefaults_NegativePage(t *testing.T) {
	offset, page, pageSize := ApplyPaginationDefaults(-10, 15)
	assert.Equal(t, 0, offset)
	assert.Equal(t, 1, page)
	assert.Equal(t, 15, pageSize)
}

func TestApplyPaginationDefaults_NegativePageSize(t *testing.T) {
	offset, page, pageSize := ApplyPaginationDefaults(5, -1)
	assert.Equal(t, 80, offset)
	assert.Equal(t, 5, page)
	assert.Equal(t, 20, pageSize)
}
