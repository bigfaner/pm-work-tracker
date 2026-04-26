# Testing Conventions

## TDD Flow

Follow red-green-refactor for all new code:

1. **RED** — write a failing test that describes the desired behavior
2. **GREEN** — write the minimal code to make the test pass
3. **REFACTOR** — clean up while keeping tests green

```go
// Step 1: RED — failing test
func TestMainItemService_Create(t *testing.T) {
    _, err := svc.Create(ctx, teamID, pmID, dto.MainItemCreateReq{Title: "test"})
    // fails: svc doesn't exist yet
}

// Step 2: GREEN — minimal implementation
func (s *mainItemService) Create(ctx context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error) {
    item := &model.MainItem{TeamID: teamID, Title: req.Title, Priority: req.Priority}
    return item, s.repo.Create(ctx, item)
}

// Step 3: REFACTOR — if needed
```

## Test File Naming

| Layer | Source file | Test file |
|-------|-----------|-----------|
| Go unit | `main_item_service.go` | `main_item_service_test.go` |
| Go unit | `main_item_handler.go` | `main_item_handler_test.go` |
| React unit | `badge.tsx` | `badge.test.tsx` |
| React E2E | — | `item-list.spec.ts` |

Test files are co-located: same directory as source file.

## Go Test Patterns

### Table-Driven Tests

Use table-driven tests for multiple input/output cases.

```go
// ✅ Correct
func TestApplyPaginationDefaults(t *testing.T) {
    tests := []struct {
        name              string
        page, pageSize    int
        wantOffset        int
        wantPage, wantSize int
    }{
        {"defaults", 0, 0, 0, 1, 20},
        {"negative", -1, -5, 0, 1, 20},
        {"normal", 3, 10, 20, 3, 10},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            offset, page, size := dto.ApplyPaginationDefaults(tt.page, tt.pageSize)
            assert.Equal(t, tt.wantOffset, offset)
            assert.Equal(t, tt.wantPage, page)
            assert.Equal(t, tt.wantSize, size)
        })
    }
}

// ❌ Wrong — separate test per case
func TestApplyPaginationDefaults_Defaults(t *testing.T) { ... }
func TestApplyPaginationDefaults_Negative(t *testing.T) { ... }
func TestApplyPaginationDefaults_Normal(t *testing.T) { ... }
```

### Handler Tests

Use `httptest` + Gin test context. Mock services via interfaces.

```go
// ✅ Correct
func TestMainItemHandler_Create(t *testing.T) {
    mockSvc := &mockMainItemService{createFn: func(...) (*model.MainItem, error) {
        return &model.MainItem{BaseModel: model.BaseModel{ID: 1}}, nil
    }}
    h := NewMainItemHandler(mockSvc, mockPoolRepo)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Set("teamID", uint(1))
    // ... set request body
    h.Create(c)
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Service Tests

Mock repositories via interfaces. Test business logic, not database queries.

```go
// ✅ Correct — mock repo, test service logic
type mockMainItemRepo struct {
    findByIDFn func(ctx context.Context, id uint) (*model.MainItem, error)
}

func TestMainItemService_GetByID_NotFound(t *testing.T) {
    mockRepo := &mockMainItemRepo{
        findByIDFn: func(ctx context.Context, id uint) (*model.MainItem, error) {
            return nil, gorm.ErrRecordNotFound
        },
    }
    svc := NewMainItemService(mockRepo, nil)
    _, err := svc.GetByID(context.Background(), 999)
    assert.ErrorIs(t, err, apperrors.ErrItemNotFound)
}
```

## React Test Patterns

### Component Tests

Use `@testing-library/react` + `vitest`.

```tsx
// ✅ Correct — badge.test.tsx
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Badge } from './badge';

describe('Badge', () => {
  it('renders children', () => {
    render(<Badge>P0</Badge>);
    expect(screen.getByText('P0')).toBeInTheDocument();
  });

  it('applies variant class', () => {
    const { container } = render(<Badge variant="error">P0</Badge>);
    expect(container.firstChild).toHaveClass('bg-error');
  });
});
```

### API Module Tests

Mock axios client, verify call patterns.

```ts
// ✅ Correct — modules.test.ts
vi.mock('./client', () => ({
  client: { get: vi.fn(), post: vi.fn() }
}));

it('calls correct endpoint for list', async () => {
  mockClient.get.mockResolvedValue({ data: { items: [], total: 0 } });
  await mainItemsApi.list(1);
  expect(mockClient.get).toHaveBeenCalledWith('/api/v1/teams/1/main-items', expect.any(Object));
});
```

## Running Tests

**Only run tests directly related to the files you changed.** Do not run `go test ./...` or `go test ./internal/...` unless explicitly asked.

```bash
# ✅ Correct — run only the package you modified
go test ./internal/service/team_service_test.go ./internal/service/team_service.go
go test ./internal/handler/ -run TestTeam
go test ./internal/model/ -run TestTeam

# ❌ Wrong — runs everything, surfaces unrelated failures
go test ./...
go test ./internal/...
```

Map changed files to their test targets:
- `model/team.go` → `go test ./internal/model/ -run TestTeam`
- `service/team_service.go` → `go test ./internal/service/ -run TestTeam`
- `handler/team_handler.go` → `go test ./internal/handler/ -run TestTeam`
- `repository/gorm/team_repo.go` → `go test ./internal/repository/gorm/ -run TestTeam`

## Test Assertions

Go: use `github.com/stretchr/testify` — `assert` for non-fatal, `require` for fatal (stops test on failure).

```go
require.NoError(t, err)   // stops test if err != nil
assert.Equal(t, want, got) // continues even if wrong
```
