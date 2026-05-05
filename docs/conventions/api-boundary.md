---
scope: backend
source: feature/bizkey-unification, feature/jlc-schema-alignment
---

# API Boundary Conventions

Rules for which ID type (int64 bizKey vs uint internal ID) crosses each layer boundary. These rules prevent the class of bugs where uint/internal IDs and int64/bizKey values are silently cast between each other.

## AB-001: Service Boundary Uses int64 BizKey

**Rule**: If a value enters the system via an HTTP request (URL param, request body, query string), it must be `int64` at the Service interface boundary.

**Why**: Mixing uint (internal ID) and int64 (bizKey) at the service layer caused three documented bugs: wrong `team_key` values in progress records, truncated role IDs in `isPMRole`, and silent data corruption from `uint(bizKey)` casts. Making the boundary explicit lets the compiler catch violations.

**Boundary map**:

| Value | Origin | Type at Service boundary |
|-------|--------|--------------------------|
| `teamBizKey` | URL param `:teamId` | `int64` |
| `roleBizKey` | Request body `roleKey` | `int64` |
| `pmID` / `callerID` / `authorID` | JWT auth context (internal user ID) | `uint` (unchanged) |
| `subItemID` / `poolItemID` | URL param (resolved by handler) | `uint` |

**Verification**:
```bash
# Should return zero results
grep -rn "uint(.*bizKey" backend/internal/service/ backend/internal/handler/
grep -rn "int64(.*teamID" backend/internal/service/ backend/internal/handler/
```

## AB-002: Repository Boundary Uses uint ID

**Rule**: Repository interfaces use `uint` ID for internal lookups (`FindByID`, `FindByIDs`). The `FindByBizKey(ctx, bizKey int64)` method is the external lookup entry point.

**Why**: Internal FK joins in the database use auto-increment `uint` IDs for performance. BizKey is only needed when resolving an external identifier to a record.

**Example**:
```go
type MainItemRepo interface {
    FindByID(ctx context.Context, id uint) (*model.MainItem, error)       // internal FK lookups
    FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.MainItem, error) // batch internal
    FindByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error)      // external lookup
    // ...
}
```

**Handler two-step delete pattern**:
```go
// Step 1: resolve external bizKey to record
item, err := h.svc.GetByBizKey(ctx, bizKey)
if err != nil { ... }

// Step 2: delete by internal uint ID
err = h.svc.Delete(ctx, item.ID)
```

## AB-003: Middleware Injects teamBizKey int64

**Rule**: `TeamScopeMiddleware` injects `teamBizKey int64` directly into the Gin context. It does NOT resolve the bizKey to an internal `uint` ID.

**Why**: Previously the middleware called `FindByBizKey` and then injected `team.ID` (uint), which required an extra DB query and lost the bizKey value. Injecting the bizKey directly ensures the service layer receives the correct snowflake identifier.

**Accessor function**:
```go
func GetTeamBizKey(c *gin.Context) int64 {
    val, exists := c.Get("teamBizKey")
    if !exists {
        return 0
    }
    bizKey, ok := val.(int64)
    if !ok {
        return 0
    }
    return bizKey
}
```

**Middleware flow**:
```
HTTP request (:teamId param)
  -> ParseInt -> int64 bizKey
  -> FindByBizKey validates team exists
  -> c.Set("teamBizKey", bizKey)   // inject int64, NOT team.ID
```

**Returning 0 on missing key is safe**: `TeamScopeMiddleware` always runs before handlers. A 0 reaching a service indicates a routing misconfiguration, not a user error. The downstream `FindByBizKey(0)` will return not-found.

## AB-004: Handler Parses BizKey from URL Params

**Rule**: Handlers parse bizKey values from URL path parameters as `int64` using `strconv.ParseInt`. The shared `ParseBizKeyParam` helper standardizes parsing and error responses.

**Why**: Centralizing parse logic prevents each handler from implementing its own error handling for invalid bizKey values. All parse failures respond with 400 Validation Error consistently.

**Helper signature**:
```go
// pkg/handler/bizkey.go
func ParseBizKeyParam(c *gin.Context, paramName string) (int64, bool)
func ResolveBizKey(c *gin.Context, paramName string, lookupFn func(ctx context.Context, bizKey int64) (uint, error)) (uint, bool)
```

**BizKey validation rules**:

| Rule | Failure condition | HTTP status |
|------|-------------------|-------------|
| Must be decimal integer string | Non-numeric characters | 400 |
| Must be positive | Parsed value <= 0 | 400 |
| Must match existing record | `FindByBizKey` returns not found | 404 |

**BizKey does NOT enforce snowflake digit length** -- only requires positive int64 and database existence.
