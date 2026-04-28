package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

// seedProgressBenchData creates a dataset of n progress records for benchmarks.
func seedProgressBenchData(n int) (*mockProgressService, *trackingUserRepo) {
	records := make([]model.ProgressRecord, n)
	users := make(map[uint]*model.User, 50)

	for i := range records {
		records[i] = model.ProgressRecord{
			ID:          uint(i + 1),
			SubItemKey:   5,
			TeamKey: 10,
			AuthorKey: int64(i%50 + 1),
			Completion:  float64(i % 100),
			Achievement: fmt.Sprintf("Achievement %d", i),
			Blocker:     "",
			Lesson:      "",
			IsPmCorrect: 0,
			CreateTime:   time.Now().Add(-time.Duration(i) * time.Minute),
		}
		users[uint(i%50+1)] = &model.User{DisplayName: fmt.Sprintf("User %d", i%50+1)}
	}

	svc := &mockProgressService{}
	svc.listResult.records = records

	trackingRepo := &trackingUserRepo{users: users}
	return svc, trackingRepo
}

func BenchmarkProgressHandler_List(b *testing.B) {
	b.StopTimer()
	svc, trackingRepo := seedProgressBenchData(200)

	deps, _ := testDeps(b)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{ RoleKey: func() *int64 { v := int64(1); return &v }()}}
	deps.Progress = NewProgressHandler(svc, trackingRepo, &mockSubItemSvcForProgress{})
	r := SetupRouter(deps, nil)

	token := signTestToken(b, 3, "testuser")
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/5/progress", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", w.Code)
		}
	}
}

// ---------------------------------------------------------------------------
// Mock ProgressService for handler tests
// ---------------------------------------------------------------------------

type mockProgressService struct {
	appendResult struct {
		record *model.ProgressRecord
		err    error
	}
	listResult struct {
		records []model.ProgressRecord
		err     error
	}
	correctResult struct {
		err error
	}

	// capture calls
	appendCalled    bool
	lastTeamBizKey  int64
	lastAuthorID    uint
	lastSubItemID   uint
	lastCompletion  float64
	lastAchievement string
	lastBlocker     string
	lastLesson      string
	lastIsPM        bool

	listCalled   bool
	listSubItemID uint

	correctCalled bool
	correctID     uint
	correctValue  float64
}

func (m *mockProgressService) Append(_ context.Context, teamBizKey int64, authorID, subItemID uint, completion float64, achievement, blocker, lesson string, isPM bool) (*model.ProgressRecord, error) {
	m.appendCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastAuthorID = authorID
	m.lastSubItemID = subItemID
	m.lastCompletion = completion
	m.lastAchievement = achievement
	m.lastBlocker = blocker
	m.lastLesson = lesson
	m.lastIsPM = isPM
	return m.appendResult.record, m.appendResult.err
}

func (m *mockProgressService) CorrectCompletion(_ context.Context, teamBizKey int64, recordID uint, completion float64) error {
	m.correctCalled = true
	m.lastTeamBizKey = teamBizKey
	m.correctID = recordID
	m.correctValue = completion
	return m.correctResult.err
}

func (m *mockProgressService) GetByBizKey(_ context.Context, bizKey int64) (*model.ProgressRecord, error) {
	return &model.ProgressRecord{ID: uint(bizKey)}, nil
}

func (m *mockProgressService) List(_ context.Context, teamBizKey int64, subItemID uint) ([]model.ProgressRecord, error) {
	m.listCalled = true
	m.lastTeamBizKey = teamBizKey
	m.listSubItemID = subItemID
	return m.listResult.records, m.listResult.err
}

// ---------------------------------------------------------------------------
// Mock SubItemService for bizKey resolution in progress handler tests
// ---------------------------------------------------------------------------

type mockSubItemSvcForProgress struct{}

func (m *mockSubItemSvcForProgress) Create(_ context.Context, _ int64, _ uint, _ dto.SubItemCreateReq) (*model.SubItem, error) {
	return nil, nil
}
func (m *mockSubItemSvcForProgress) Update(_ context.Context, _ int64, _ uint, _ dto.SubItemUpdateReq) error { return nil }
func (m *mockSubItemSvcForProgress) ChangeStatus(_ context.Context, _ int64, _, _ uint, _ string) (*service.SubItemChangeResult, error) {
	return nil, nil
}
func (m *mockSubItemSvcForProgress) Delete(_ context.Context, _ int64, _, _ uint) error { return nil }
func (m *mockSubItemSvcForProgress) Get(_ context.Context, _ int64, _ uint) (*model.SubItem, error) { return nil, nil }
func (m *mockSubItemSvcForProgress) GetByBizKey(_ context.Context, bizKey int64) (*model.SubItem, error) {
	return &model.SubItem{BaseModel: model.BaseModel{ID: uint(bizKey)}}, nil
}
func (m *mockSubItemSvcForProgress) List(_ context.Context, _ int64, _ *uint, _ dto.SubItemFilter, _ dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}
func (m *mockSubItemSvcForProgress) Assign(_ context.Context, _ int64, _, _, _ uint) error { return nil }
func (m *mockSubItemSvcForProgress) AvailableTransitions(_ context.Context, _ int64, _ uint) ([]string, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// depsWithProgressSvc wires a mock ProgressService with PM role.

func depsWithProgressSvc(t *testing.T, svc *mockProgressService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{ RoleKey: func() *int64 { v := int64(1); return &v }()}}
	deps.Progress = NewProgressHandler(svc, &mockUserRepoForHandler{}, &mockSubItemSvcForProgress{})
	return deps
}

// depsWithProgressSvcMemberRole wires with member (non-PM) role.

func depsWithProgressSvcMemberRole(t *testing.T, svc *mockProgressService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{ RoleKey: func() *int64 { v := int64(2); return &v }()}}
	deps.Progress = NewProgressHandler(svc, &mockUserRepoForHandler{}, &mockSubItemSvcForProgress{})
	return deps
}

// depsWithProgressSvcAndUser wires with a specific user repo.

func depsWithProgressSvcAndUser(t *testing.T, svc *mockProgressService, userRepo repository.UserRepo) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{ RoleKey: func() *int64 { v := int64(1); return &v }()}}
	deps.Progress = NewProgressHandler(svc, userRepo, &mockSubItemSvcForProgress{})
	return deps
}

// testProgressRecord creates a ProgressRecord model for tests.
func testProgressRecord(id uint, subItemID uint, authorID uint) *model.ProgressRecord {
	return &model.ProgressRecord{
		ID:          id,
		SubItemKey: int64(subItemID),
		TeamKey: 10,
		AuthorKey: int64(authorID),
		Completion:  60.0,
		Achievement: "completed feature",
		Blocker:     "none",
		Lesson:      "test early",
		IsPmCorrect: 0,
		CreateTime:   time.Now(),
	}
}

// trackingUserRepo tracks call counts to verify batch vs individual lookups.
type trackingUserRepo struct {
	users              map[uint]*model.User
	findByIDsCallCount int
	findByIDCallCount  int
}

func (t *trackingUserRepo) FindByID(_ context.Context, id uint) (*model.User, error) {
	t.findByIDCallCount++
	return t.users[id], nil
}
func (t *trackingUserRepo) FindByIDs(_ context.Context, ids []uint) (map[uint]*model.User, error) {
	t.findByIDsCallCount++
	result := make(map[uint]*model.User, len(ids))
	for _, id := range ids {
		if u, ok := t.users[id]; ok {
			result[id] = u
		}
	}
	return result, nil
}
func (t *trackingUserRepo) FindByUsername(_ context.Context, _ string) (*model.User, error)    { return nil, nil }
func (t *trackingUserRepo) List(_ context.Context) ([]*model.User, error)                       { return nil, nil }
func (t *trackingUserRepo) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
	return nil, 0, nil
}
func (t *trackingUserRepo) SearchAvailable(_ context.Context, _ int64, _ string, _ int) ([]*model.User, error) {
	return nil, nil
}
func (t *trackingUserRepo) Create(_ context.Context, _ *model.User) error { return nil }
func (t *trackingUserRepo) Update(_ context.Context, _ *model.User) error { return nil }
func (t *trackingUserRepo) SoftDelete(_ context.Context, _ *model.User) error { return nil }
func (t *trackingUserRepo) FindByBizKey(_ context.Context, _ int64) (*model.User, error) {
	return nil, nil
}

// compile-time check
var _ repository.UserRepo = (*trackingUserRepo)(nil)

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/sub-items/:subId/progress (Append)
// ---------------------------------------------------------------------------

func TestAppendProgress_Success(t *testing.T) {
	svc := &mockProgressService{}
	record := testProgressRecord(1, 5, 3)
	svc.appendResult.record = record

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":60,"achievement":"completed feature","blocker":"none","lesson":"test early"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(60), data["completion"])
	assert.True(t, svc.appendCalled)
	assert.Equal(t, int64(10), svc.lastTeamBizKey)
	assert.Equal(t, uint(3), svc.lastAuthorID)
	assert.Equal(t, uint(5), svc.lastSubItemID)
}

func TestAppendProgress_MemberCanAppend(t *testing.T) {
	// Any team member can append progress
	svc := &mockProgressService{}
	record := testProgressRecord(1, 5, 3)
	svc.appendResult.record = record

	deps := depsWithProgressSvcMemberRole(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":60}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, svc.appendCalled)
}

func TestAppendProgress_InvalidBody(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.appendCalled)
}

func TestAppendProgress_CompletionOutOfRange_Negative(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":-1}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.appendCalled)
}

func TestAppendProgress_CompletionOutOfRange_Above100(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":101}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.appendCalled)
}

func TestAppendProgress_CompletionBoundary_0(t *testing.T) {
	svc := &mockProgressService{}
	record := testProgressRecord(1, 5, 3)
	record.Completion = 0
	svc.appendResult.record = record

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":0}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, svc.appendCalled)
}

func TestAppendProgress_CompletionBoundary_100(t *testing.T) {
	svc := &mockProgressService{}
	record := testProgressRecord(1, 5, 3)
	record.Completion = 100
	svc.appendResult.record = record

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":100}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, svc.appendCalled)
}

func TestAppendProgress_Regression_422(t *testing.T) {
	svc := &mockProgressService{}
	svc.appendResult.err = apperrors.ErrProgressRegression

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":30}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "PROGRESS_REGRESSION", resp["code"])
}

func TestAppendProgress_ServiceError(t *testing.T) {
	svc := &mockProgressService{}
	svc.appendResult.err = errors.New("unexpected")

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":60}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAppendProgress_InvalidSubID(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":60}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/abc/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.appendCalled)
}

func TestAppendProgress_IncludesAuthorName(t *testing.T) {
	svc := &mockProgressService{}
	record := testProgressRecord(1, 5, 3)
	svc.appendResult.record = record

	userRepo := &mockUserRepoForHandler{
		user: &model.User{DisplayName: "张三"},
	}

	deps := depsWithProgressSvcAndUser(t, svc, userRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := `{"completion":60}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/sub-items/5/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "张三", data["authorName"])
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/sub-items/:subId/progress (List)
// ---------------------------------------------------------------------------

func TestListProgress_Success(t *testing.T) {
	svc := &mockProgressService{}
	svc.listResult.records = []model.ProgressRecord{
		*testProgressRecord(1, 5, 3),
		*testProgressRecord(2, 5, 3),
	}
	svc.listResult.records[0].Completion = 30
	svc.listResult.records[1].Completion = 60

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/5/progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
	assert.True(t, svc.listCalled)
	assert.Equal(t, uint(5), svc.listSubItemID)
}

func TestListProgress_Empty(t *testing.T) {
	svc := &mockProgressService{}
	svc.listResult.records = []model.ProgressRecord{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/5/progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, data)
}

func TestListProgress_InvalidSubID(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/abc/progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.listCalled)
}

func TestListProgress_ServiceError(t *testing.T) {
	svc := &mockProgressService{}
	svc.listResult.err = errors.New("db error")

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/5/progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListProgress_IncludesAuthorNames(t *testing.T) {
	svc := &mockProgressService{}
	record1 := testProgressRecord(1, 5, 3)
	record2 := testProgressRecord(2, 5, 7)
	svc.listResult.records = []model.ProgressRecord{*record1, *record2}

	userRepo := &mockUserRepoForHandler{
		user: &model.User{DisplayName: "张三"},
	}

	deps := depsWithProgressSvcAndUser(t, svc, userRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/5/progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)

	// Both records get authorName from userRepo
	item0 := data[0].(map[string]interface{})
	assert.Equal(t, "张三", item0["authorName"])
}

func TestListProgress_UsesBatchLookup(t *testing.T) {
	// Verify progressRecordsToVOs uses FindByIDs (batch) not FindByID (N+1)
	svc := &mockProgressService{}
	svc.listResult.records = []model.ProgressRecord{
		*testProgressRecord(1, 5, 3),
		*testProgressRecord(2, 5, 7),
		*testProgressRecord(3, 5, 3),
	}

	trackingRepo := &trackingUserRepo{
		users: map[uint]*model.User{
			3: {DisplayName: "Alice"},
			7: {DisplayName: "Bob"},
		},
	}

	deps := depsWithProgressSvcAndUser(t, svc, trackingRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/5/progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, trackingRepo.findByIDsCallCount, "FindByIDs should be called exactly once")
	assert.Equal(t, 0, trackingRepo.findByIDCallCount, "FindByID should not be called in batch path")

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].([]interface{})
	item0 := data[0].(map[string]interface{})
	item1 := data[1].(map[string]interface{})
	item2 := data[2].(map[string]interface{})
	assert.Equal(t, "Alice", item0["authorName"])
	assert.Equal(t, "Bob", item1["authorName"])
	assert.Equal(t, "Alice", item2["authorName"])
}

// ---------------------------------------------------------------------------
// Tests: PATCH /teams/:teamId/progress/:recordId/completion (CorrectCompletion)
// ---------------------------------------------------------------------------

func TestCorrectCompletion_Success(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"completion":75}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/progress/100/completion", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.correctCalled)
	assert.Equal(t, uint(100), svc.correctID)
	assert.InDelta(t, 75.0, svc.correctValue, 0.001)
}

func TestCorrectCompletion_RequiresPM(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvcMemberRole(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"completion":75}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/progress/100/completion", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.correctCalled)
}

func TestCorrectCompletion_InvalidRecordID(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"completion":75}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/progress/abc/completion", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.correctCalled)
}

func TestCorrectCompletion_InvalidBody(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/progress/100/completion", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.correctCalled)
}

func TestCorrectCompletion_CompletionOutOfRange_Negative(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"completion":-1}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/progress/100/completion", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.correctCalled)
}

func TestCorrectCompletion_CompletionOutOfRange_Above100(t *testing.T) {
	svc := &mockProgressService{}

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"completion":101}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/progress/100/completion", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.correctCalled)
}

func TestCorrectCompletion_RecordNotFound(t *testing.T) {
	svc := &mockProgressService{}
	svc.correctResult.err = apperrors.ErrItemNotFound

	deps := depsWithProgressSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"completion":75}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/progress/999/completion", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCorrectCompletion_SuperAdminBypass(t *testing.T) {
	svc := &mockProgressService{}

	// Use member-role team, but superadmin token
	deps := depsWithProgressSvcMemberRole(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 1, "admin")
	body := `{"completion":75}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/progress/100/completion", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.correctCalled)
}

// ---------------------------------------------------------------------------
// Tests: Response shape matches Data Contract
// ---------------------------------------------------------------------------

func TestAppendProgress_ResponseShapeMatchesDataContract(t *testing.T) {
	svc := &mockProgressService{}
	now := time.Now()
	record := &model.ProgressRecord{
		ID:          100,
		BizKey: 100,
		SubItemKey:   10,
		TeamKey: 1,
		AuthorKey:    3,
		Completion:  60,
		Achievement: "completed SDK init",
		Blocker:     "certificate pending",
		Lesson:      "sandbox vs prod config diff",
		IsPmCorrect: 0,
		CreateTime:   now,
	}
	svc.appendResult.record = record

	userRepo := &mockUserRepoForHandler{
		user: &model.User{DisplayName: "李四"},
	}

	deps := depsWithProgressSvcAndUser(t, svc, userRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 3, "testuser")
	body := fmt.Sprintf(`{"completion":60,"achievement":"completed SDK init","blocker":"certificate pending","lesson":"sandbox vs prod config diff"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/1/sub-items/10/progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})

	// Verify all expected fields from ProgressRecord Object contract
	assert.Equal(t, "100", data["bizKey"])
	assert.Equal(t, "10", data["subItemKey"])
	assert.Equal(t, "3", data["authorKey"])
	assert.Equal(t, "李四", data["authorName"])
	assert.Equal(t, 60.0, data["completion"])
	assert.Equal(t, "completed SDK init", data["achievement"])
	assert.Equal(t, "certificate pending", data["blocker"])
	assert.Equal(t, float64(0), data["isPmCorrect"])
	assert.NotNil(t, data["createTime"])
}

// ---------------------------------------------------------------------------
// Tests: buildProgressRecordVOs unified function
// ---------------------------------------------------------------------------

func TestBuildProgressRecordVOs_EmptySlice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	result := buildProgressRecordVOs(nil, &mockUserRepoForHandler{}, c)
	assert.Equal(t, []vo.ProgressRecordVO{}, result)
}

func TestBuildProgressRecordVOs_SingleItem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	record := model.ProgressRecord{ID: 1, AuthorKey: 3, Completion: 50, Achievement: "did it"}
	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "Bob"}}

	result := buildProgressRecordVOs([]model.ProgressRecord{record}, userRepo, c)
	require.Len(t, result, 1)
	assert.Equal(t, "Bob", result[0].AuthorName)
	assert.Equal(t, 50.0, result[0].Completion)
}
