---
title: "单元测试约定"
domains: [testing, unit]
---

# 单元测试约定

## T-003: Table-Driven Tests

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

## T-004: Handler Tests

Use `httptest` + Gin test context. Mock services via interfaces.

```go
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

## T-005: Service Tests

Mock repositories via interfaces. Test business logic, not database queries.

```go
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

## T-006: React Component Tests

Use `@testing-library/react` + `vitest`.

```tsx
// badge.test.tsx
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

## T-007: API Module Tests

Mock axios client, verify call patterns.

```ts
// modules.test.ts
vi.mock('./client', () => ({
  client: { get: vi.fn(), post: vi.fn() }
}));

it('calls correct endpoint for list', async () => {
  mockClient.get.mockResolvedValue({ data: { items: [], total: 0 } });
  await mainItemsApi.list(1);
  expect(mockClient.get).toHaveBeenCalledWith('/api/v1/teams/1/main-items', expect.any(Object));
});
```
