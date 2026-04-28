package gorm_test

import (
	"os"
	"testing"

	"pm-work-tracker/backend/internal/pkg/snowflake"
)

func TestMain(m *testing.M) {
	_ = snowflake.Init(1)
	os.Exit(m.Run())
}
