---
paths:
  - "backend/**/*.go"
---

# Backend Go Patterns

## Handler Constructor Pattern

All handler constructors validate dependencies with panic-on-nil. No method-level nil checks.

```go
// ✅ Correct
func NewMainItemHandler(svc MainItemService, poolRepo repository.ItemPoolRepo) *MainItemHandler {
    if svc == nil {
        panic("main_item_handler: mainItemService must not be nil")
    }
    if poolRepo == nil {
        panic("main_item_handler: itemPoolRepo must not be nil")
    }
    return &MainItemHandler{svc: svc, poolRepo: poolRepo}
}

// ❌ Wrong — nil check inside handler method
func (h *MainItemHandler) Create(c *gin.Context) {
    if h.svc == nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": "service not initialized"})
        return
    }
}
```

Panic format: `<snake_case_handler>: <interfaceName> must not be nil`

## Handler Method Pattern

Parse request → call service → respond.

```go
// ✅ Correct
func (h *MainItemHandler) Create(c *gin.Context) {
    teamID := middleware.GetTeamID(c)
    var req dto.MainItemCreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        apperrors.RespondError(c, apperrors.ErrValidation)
        return
    }
    item, err := h.svc.Create(c.Request.Context(), teamID, pmID, req)
    if err != nil {
        apperrors.RespondError(c, err)
        return
    }
    apperrors.RespondOK(c, vo.NewMainItemVO(item))
}

// ❌ Wrong — business logic in handler
func (h *MainItemHandler) Create(c *gin.Context) {
    // no direct DB calls or business rules
    item := &model.MainItem{Title: req.Title}
    h.db.Create(item)
}
```

## Service Pattern

Service interfaces are defined in the service package. Methods receive DTOs, return models or DTOs.

```go
// ✅ Correct — interface + constructor + implementation
type MainItemService interface {
    Create(ctx context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error)
    List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)
}

func NewMainItemService(repo repository.MainItemRepo, subRepo repository.SubItemRepo) MainItemService {
    return &mainItemService{repo: repo, subRepo: subRepo}
}

// ❌ Wrong — service depends on *gorm.DB directly
func NewMainItemService(db *gorm.DB) MainItemService { ... }
```

## Error Mapping Pattern

Use `errors.MapNotFound` from `pkg/errors` for not-found mapping. Use `apperrors.RespondError` in handlers.

```go
// ✅ Correct
import pkgerrors "pm-work-tracker/internal/pkg/errors"

func (s *mainItemService) GetByID(ctx context.Context, id uint) (*model.MainItem, error) {
    item, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, pkgerrors.MapNotFound(err, apperrors.ErrItemNotFound)
    }
    return item, nil
}

// ❌ Wrong — inline not-found mapping
func (s *mainItemService) GetByID(ctx context.Context, id uint) (*model.MainItem, error) {
    item, err := s.repo.FindByID(ctx, id)
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperrors.ErrItemNotFound
    }
    return item, err
}
```

## Repository Pattern

Interface in `repository/`, GORM implementation in `repository/gorm/`.

```go
// repository/main_item_repo.go — interface
type MainItemRepo interface {
    Create(ctx context.Context, item *model.MainItem) error
    FindByID(ctx context.Context, id uint) (*model.MainItem, error)
    FindByTeamID(ctx context.Context, teamID uint, filter dto.MainItemFilter, offset, limit int) ([]model.MainItem, error)
}

// repository/gorm/main_item_repo.go — implementation
func NewGormMainItemRepo(db *gorm.DB) repository.MainItemRepo {
    return &gormMainItemRepo{db: db}
}
```

## Pagination Pattern

Use `dto.ApplyPaginationDefaults` in handlers only. Repos receive pre-computed offset/limit.

```go
// ✅ Correct — handler computes pagination
offset, page, pageSize := dto.ApplyPaginationDefaults(pageParam, pageSizeParam)
result, err := h.svc.List(ctx, teamID, filter, offset, pageSize)

// ❌ Wrong — pagination logic in repository
func (r *gormMainItemRepo) FindByTeamID(..., page, pageSize int) {
    if page <= 0 { page = 1 }
    offset := (page - 1) * pageSize
}
```

## VO Conversion Pattern

VOs convert models to API-friendly responses in handlers, not services.

```go
// ✅ Correct — handler converts
items := make([]MainItemVO, len(result.Items))
for i, item := range result.Items {
    items[i] = vo.NewMainItemVO(&item)
}

// ❌ Wrong — service returns VO
func (s *mainItemService) List(...) ([]vo.MainItemVO, error) { ... }
```
