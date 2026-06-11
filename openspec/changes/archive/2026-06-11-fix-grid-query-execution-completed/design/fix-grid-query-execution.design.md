# Design: Fix Grid Query Execution by ID

## Overview

Fix grid query execution so the frontend sends only `{ queryId, page, pageSize }`, while the backend fetches the saved query configuration, looks up the grid's base query, and constructs SQL dynamically.

**Problem**: `ExecuteQuery` in `service.go` (line 146) passes an empty string as `baseQuery`, causing all queries to fail.

**Root Cause**: The service was designed to receive `baseQuery` from the grid repository, but the handler never passed it—always sending `""`.

---

## Design 1: GridDataRequest + Handler

**File**: `backend/internal/adapters/api/grid/handler.go`

### Change 1.1: Add QueryID field to GridDataRequest (line 21-29)

```go
// GridDataRequest represents the request to execute a grid query
type GridDataRequest struct {
    GridID    int                          `json:"gridId"`
    QueryID   string                       `json:"queryId,omitempty"`  // NEW
    FirstLoad bool                         `json:"firstLoad"`
    Fields    []int                        `json:"fields"`
    Sort      []griddomain.SortCondition   `json:"sort"`
    Filters   []griddomain.FilterCondition `json:"filters"`
    Page      int                          `json:"page"`
    PageSize  int                          `json:"pageSize"`
}
```

### Change 1.2: Modify ExecuteData handler (line 92-146)

Replace the existing `ExecuteData` logic with routing based on presence of `QueryID`:

```go
// ExecuteData handles POST /grid/data
// Executes a grid query
func (h *GridHandler) ExecuteData(c *fiber.Ctx) error {
    var req GridDataRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(errors.ErrBadRequest)
    }

    // Default pagination
    page := req.Page
    if page < 1 {
        page = 1
    }
    pageSize := req.PageSize
    if pageSize < 1 {
        pageSize = 20
    }

    // Route based on presence of QueryID
    if req.QueryID != "" {
        // Execute by saved query ID
        result, err := h.gridService.ExecuteQueryByID(c.Context(), req.QueryID, page, pageSize)
        if err != nil {
            if err == griddomain.ErrQueryNotFound {
                return c.Status(404).JSON(errors.New("NOT_FOUND", "Query not found", 404))
            }
            if err == griddomain.ErrGridNotFound {
                return c.Status(404).JSON(errors.New("NOT_FOUND", "Grid not found", 404))
            }
            return c.Status(500).JSON(errors.ErrInternal)
        }

        totalPages := result.Total / result.PageSize
        if result.Total%result.PageSize > 0 {
            totalPages++
        }
        return c.JSON(griddomain.GridResponse{
            Success: true,
            Data:    convertToSliceAny(result.Data),
            Meta: griddomain.GridMeta{
                Page:       result.Page,
                PageSize:   result.PageSize,
                Total:      result.Total,
                TotalPages: totalPages,
            },
        })
    }

    // Fallback: validate gridId required when no QueryID
    if req.GridID == 0 {
        return c.Status(400).JSON(errors.New("BAD_REQUEST", "gridId is required when queryId not provided", 400))
    }

    // Existing flat-field execution (backward compatible)
    queryConfig := &griddomain.GridQueryConfig{
        Fields:  req.Fields,
        Sort:    req.Sort,
        Filters: req.Filters,
        Pagination: griddomain.Pagination{
            PageSize: pageSize,
        },
    }

    result, err := h.gridService.ExecuteQuery(c.Context(), req.GridID, queryConfig, page, pageSize)
    if err != nil {
        if err == griddomain.ErrGridNotFound {
            return c.Status(404).JSON(errors.New("NOT_FOUND", "Grid not found", 404))
        }
        return c.Status(500).JSON(errors.ErrInternal)
    }

    totalPages := result.Total / result.PageSize
    if result.Total%result.PageSize > 0 {
        totalPages++
    }

    return c.JSON(griddomain.GridResponse{
        Success: true,
        Data:    convertToSliceAny(result.Data),
        Meta: griddomain.GridMeta{
            Page:       result.Page,
            PageSize:   result.PageSize,
            Total:      result.Total,
            TotalPages: totalPages,
        },
    })
}
```

### Rollback Plan
- Remove `QueryID` field from `GridDataRequest`
- Restore original `ExecuteData` logic that always validates `gridId` and builds `GridQueryConfig` from flat fields

---

## Design 2: ExecuteQueryByID Service Method

**File**: `backend/internal/domain/grid/service.go`

### Change 2.1: Add queriesRepo and new errors (line 1-14)

Add `queriesRepo` field to Service struct and add new errors:

```go
type Service struct {
    repo        GridRepository
    fieldRepo   FieldsRepository
    queriesRepo QueriesRepository  // NEW
}
```

Update `NewService` and add new constructor:

```go
// NewService creates a new grid service
func NewService(repo GridRepository) *Service {
    return &Service{repo: repo}
}

// NewServiceWithQueries creates a new grid service with queries repository
func NewServiceWithQueries(repo GridRepository, queriesRepo QueriesRepository) *Service {
    return &Service{repo: repo, queriesRepo: queriesRepo}
}
```

### Change 2.2: Add error definitions

**File**: `backend/internal/domain/grid/errors.go`

```go
package grid

import "errors"

var (
    ErrGridNotFound  = errors.New("grid not found")
    ErrQueryNotFound = errors.New("query not found")  // NEW
)
```

### Change 2.3: Add ExecuteQueryByID method (after line 147)

```go
// ExecuteQueryByID fetches saved query, resolves grid, and executes
func (s *Service) ExecuteQueryByID(ctx context.Context, queryID string, page, pageSize int) (*GridResult, error) {
    // Default pagination
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 20
    }

    // 1. Fetch SavedQuery by ID
    if s.queriesRepo == nil {
        return nil, errors.New("queries repository not configured")
    }
    savedQuery, err := s.queriesRepo.GetByID(ctx, queryID)
    if err != nil {
        if strings.Contains(err.Error(), "query not found") {
            return nil, ErrQueryNotFound
        }
        return nil, err
    }

    // 2. Fetch Grid by gridId to get baseQuery
    grid, err := s.repo.FindByID(ctx, savedQuery.GridID)
    if err != nil {
        return nil, err
    }
    if grid == nil {
        return nil, ErrGridNotFound
    }

    // 3. Build config from saved query's fields/filters/sort
    config := &GridQueryConfig{
        Fields:  savedQuery.Query.Fields,
        Sort:    convertSort(savedQuery.Query.Sort),
        Filters: convertFilters(savedQuery.Query.Filters),
    }

    // 4. Call repo.ExecuteQuery with baseQuery from grid
    return s.repo.ExecuteQuery(ctx, grid.BaseQuery, config.Fields, config.Filters, config.Sort, page, pageSize)
}

// convertSort converts queries.QuerySort to grid.SortCondition
func convertSort(sort []queries.QuerySort) []SortCondition {
    result := make([]SortCondition, len(sort))
    for i, s := range sort {
        result[i] = SortCondition{
            Field:     s.Field,
            Direction: s.Direction,
        }
    }
    return result
}

// convertFilters converts queries.QueryFilter to grid.FilterCondition
func convertFilters(filters []queries.QueryFilter) []FilterCondition {
    result := make([]FilterCondition, len(filters))
    for i, f := range filters {
        result[i] = FilterCondition{
            Field:    f.Field,
            Operator: f.Operator,
            Value:    f.Value,
        }
    }
    return result
}
```

### Import additions (line 3-8)

```go
import (
    "context"
    "strings"

    fieldsdomain "github.com/nova/backend/internal/domain/fields"
    queriesdomain "github.com/nova/backend/internal/domain/queries"  // NEW
)
```

### Rollback Plan
- Remove `queriesRepo` field from Service struct
- Remove `NewServiceWithQueries` constructor
- Remove `ErrQueryNotFound` from errors.go
- Remove `ExecuteQueryByID` method and helper functions
- Restore `ExecuteQuery` to original state (still buggy for flat-field path—this is a separate issue)

---

## Design 3: Frontend Query Service

**File**: `frontend/src/app/core/services/query.service.ts`

### Change 3.1: Modify executeQuery method (line 29-47)

Change from flat-field `{ gridId, query, page }` to `{ queryId, page, pageSize }`:

```typescript
// Ejecutar query y obtener datos (using saved query ID)
executeQuery(
  queryId: string,
  page: number = 1,
  pageSize: number = 20
): Observable<PaginatedResponse<unknown>> {
  return this.api
    .postRaw<PaginatedResponse<unknown>>('/grid/data', {
      queryId,
      page,
      pageSize,
    })
    .pipe(
      map((response) => {
        this.currentMeta.set(response.meta);
        return response;
      })
    );
}
```

### Change 3.2: Add executeQueryDirect for backward compatibility (new method after executeQuery)

```typescript
// Execute direct query without saved query ID (backward compatible)
executeQueryDirect(
  gridId: number,
  fields: number[],
  filters: FilterCondition[],
  sort: SortCondition[],
  page: number = 1,
  pageSize: number = 20
): Observable<PaginatedResponse<unknown>> {
  return this.api
    .postRaw<PaginatedResponse<unknown>>('/grid/data', {
      gridId,
      fields,
      filters,
      sort,
      page,
      pageSize,
    })
    .pipe(
      map((response) => {
        this.currentMeta.set(response.meta);
        return response;
      })
    );
}
```

### Change 3.3: Update selectQuery to set query ID context (line 118-121)

The `selectQuery` method already sets `selectedQuery` and `currentQuery`. No change needed for the signal update, but callers of `executeQuery` now need to pass `queryId` instead of `gridId`.

### Required Model Updates

Add/update types if not present in `@core/models/query.model`:

```typescript
interface FilterCondition {
  field: number;
  operator: number;
  value: any;
}

interface SortCondition {
  field: number;
  direction: number; // 1=ASC, 2=DESC
}
```

### Rollback Plan
- Restore `executeQuery` to original signature `(gridId: number, query: GridQuery, page: number = 1)`
- Remove `executeQueryDirect` method

---

## Design 4: Queries Repository Verification

**File**: `backend/internal/adapters/db/queries/query_repository.go`

### Verification: GetByID already exists (line 66-97)

No code changes required. The existing implementation correctly:
1. Returns `fmt.Errorf("query not found: %w", err)` when `pgx.ErrNoRows` occurs (line 80-82)
2. Unmarshals JSONB `qry_query` into `q.Query` (line 87-91)
3. Returns proper error that the service layer can detect via `strings.Contains(err.Error(), "query not found")`

### Optional Enhancement (not required for this change)

Add a typed error wrapper in `domain/queries/errors.go` if stricter error handling is desired in the future. Not implementing now to minimize blast radius.

---

## Implementation Order

| Step | Task | File | Risk |
|------|------|------|------|
| 1 | Add `ErrQueryNotFound` to errors.go | `backend/internal/domain/grid/errors.go` | Low |
| 2 | Verify `GetByID` implementation | `backend/internal/adapters/db/queries/query_repository.go` | Low (no change) |
| 3 | Add `queriesRepo` field and `ExecuteQueryByID` to service | `backend/internal/domain/grid/service.go` | Medium |
| 4 | Update handler to route based on `QueryID` | `backend/internal/adapters/api/grid/handler.go` | Medium |
| 5 | Update frontend `query.service.ts` | `frontend/src/app/core/services/query.service.ts` | Medium |

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Breaking existing API consumers | `QueryID` is optional field; requests without it use existing flat-field flow |
| SavedQuery config merging | Request-level overrides NOT implemented in this change (future enhancement) |
| Query not found (404) | Service returns `ErrQueryNotFound`; handler returns 404 with "Query not found" |
| Grid not found (404) | Service returns `ErrGridNotFound`; handler returns 404 with "Grid not found" |
| Service panics if queriesRepo nil | Added nil check with descriptive error message |

---

## Testing Requirements

### Backend Unit Tests
- `ExecuteQueryByID` with valid query ID → returns GridResult
- `ExecuteQueryByID` with non-existent query ID → returns `ErrQueryNotFound`
- `ExecuteQueryByID` where grid does not exist → returns `ErrGridNotFound`
- Handler routes correctly based on presence of `QueryID`
- Backward compatibility: handler still works without `QueryID`

### Frontend
- `executeQuery("abc123", 1, 20)` sends `{ queryId, page, pageSize }`
- `executeQueryDirect(5, [1,2], [], [], 1, 20)` sends flat structure for backward compat

---

## File Summary

| File | Lines Changed | New Lines | Type |
|------|---------------|-----------|------|
| `backend/internal/domain/grid/errors.go` | 7 → 9 | +2 | Modified |
| `backend/internal/domain/grid/service.go` | 157 → ~220 | +63 | Modified |
| `backend/internal/adapters/api/grid/handler.go` | 168 → ~210 | +42 | Modified |
| `frontend/src/app/core/services/query.service.ts` | 175 → ~210 | +35 | Modified |
| `backend/internal/adapters/db/queries/query_repository.go` | 0 | 0 | Verified (no change) |

**Total new lines**: ~142
**Total files modified**: 4