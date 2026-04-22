package gorm_test

import (
	"context"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

// setupConcurrentTestDB opens a file-based SQLite DB with MaxOpenConns=1
// to serialize concurrent access (SQLite table-lock guarantee).
func setupConcurrentTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	// Use a temp file so multiple goroutines share the same DB file
	db, err := gormlib.Open(sqlite.Open("file::memory:?cache=shared"), &gormlib.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1) // serialize concurrent queries through one connection

	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.MainItem{}, &model.SubItem{}))
	return db
}

func TestConcurrentNextCode(t *testing.T) {
	db := setupConcurrentTestDB(t)
	repo := gormrepo.NewGormMainItemRepo(db)
	ctx := context.Background()

	u := model.User{Username: "cc_pm", DisplayName: "CC PM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "CC Team", PmID: u.ID, Code: "CONC"}
	require.NoError(t, db.Create(&team).Error)

	results := make(chan string, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			code, err := repo.NextCode(ctx, team.ID)
			if err == nil {
				results <- code
			}
		}()
	}

	wg.Wait()
	close(results)

	var codes []string
	for c := range results {
		codes = append(codes, c)
	}

	require.Len(t, codes, 2, "both goroutines should return a code")
	assert.NotEmpty(t, codes[0])
	assert.NotEmpty(t, codes[1])
	assert.NotEqual(t, codes[0], codes[1], "concurrent NextCode calls must return different codes")
}

func TestConcurrentNextSubCode(t *testing.T) {
	db := setupConcurrentTestDB(t)
	subRepo := gormrepo.NewGormSubItemRepo(db)
	ctx := context.Background()

	u := model.User{Username: "cs_pm", DisplayName: "CS PM", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "CS Team", PmID: u.ID, Code: "CSUB"}
	require.NoError(t, db.Create(&team).Error)
	mi := model.MainItem{TeamID: team.ID, Code: "CSUB-00001", Title: "Main", Priority: "P1", ProposerID: u.ID, Status: "pending"}
	require.NoError(t, db.Create(&mi).Error)

	results := make(chan string, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			code, err := subRepo.NextSubCode(ctx, mi.ID)
			if err == nil {
				results <- code
			}
		}()
	}

	wg.Wait()
	close(results)

	var codes []string
	for c := range results {
		codes = append(codes, c)
	}

	require.Len(t, codes, 2, "both goroutines should return a sub code")
	assert.NotEmpty(t, codes[0])
	assert.NotEmpty(t, codes[1])
	assert.NotEqual(t, codes[0], codes[1], "concurrent NextSubCode calls must return different codes")
}
