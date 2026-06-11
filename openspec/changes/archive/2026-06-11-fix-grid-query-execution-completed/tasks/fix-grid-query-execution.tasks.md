# Tasks: Fix Grid Query Execution by ID

## Overview

Break down the SDD change "fix grid query execution by ID" into discrete, actionable implementation tasks.

---

## Task 1: Add ErrQueryNotFound Error

**Task ID:** T1  
**Title:** Add ErrQueryNotFound error  
**File(s):** `backend/internal/domain/grid/errors.go`  
**Status:** [x] Completed

**Description:**
Add a new sentinel error `ErrQueryNotFound` to the grid domain errors. This error is returned when `ExecuteQueryByID` cannot find the requested saved query.

**Changes:**
1. Open `backend/internal/domain/grid/errors.go`
2. Add `ErrQueryNotFound = errors.New("query not found")` to the existing `var` block alongside `ErrGridNotFound`

**Verification:**
- Unit test: `errors.Is(err, ErrQueryNotFound)` returns `true` when error is set
- Integration test: Calling `ExecuteQueryByID` with non-existent query ID returns error that satisfies `errors.Is(err, ErrQueryNotFound)`

**Estimated Lines:** +2

---

## Task 2: Verify Queries Repository GetByID

**Task ID:** T2  
**Title:** Verify queries repository GetByID  
**File(s):** `backend/internal/adapters/db/queries/query_repository.go`  
**Status:** [x] Completed (Verified - no changes needed)

**Description:**
Verify that `GetByID` implementation exists and correctly handles the "query not found" case. According to the design, this method already exists at line 66-97 and properly wraps `pgx.ErrNoRows` with "query not found" message. No code changes are needed if verification passes.

**Changes:**
1. Open `backend/internal/adapters/db/queries/query_repository.go`
2. Verify `GetByID` method:
   - Checks for `pgx.ErrNoRows` and returns `fmt.Errorf("query not found: %w", err)`
   - Correctly unmarshals `qry_query` JSONB column into `q.Query`
   - Returns the populated `SavedQuery` and any error
3. Verify interface in `backend/internal/domain/queries/repository.go` declares `GetByID(ctx context.Context, id string) (*SavedQuery, error)`

**Verification:**
- Review code: Error message contains "query not found" so service can detect via `strings.Contains`
- No test changes needed if existing implementation is correct
- If implementation differs from spec, update to match design

**Estimated Lines:** 0 (verification only)

---

## Task 3: Add queriesRepo Field and Constructor to Grid Service

**Task ID:** T3  
**Title:** Add queriesRepo to Grid Service  
**File(s):** `backend/internal/domain/grid/service.go`  
**Status:** [x] Completed

**Description:**
Add the `queriesRepo` field to the `Service` struct and create a new `NewServiceWithQueries` constructor. Update wire.go (or wherever dependency injection happens) to use the new constructor.

**Changes (service.go):**
1. Add `queriesdomain "github.com/nova/backend/internal/domain/queries"` import
2. Add `queriesRepo QueriesRepository` field to `Service` struct
3. Update `NewService` signature to optionally accept queries repo, or add `NewServiceWithQueries`
4. Add `strings` import for error checking

```go
type Service struct {
    repo        GridRepository
    fieldRepo   FieldsRepository
    queriesRepo QueriesRepository  // NEW
}

func NewServiceWithQueries(repo GridRepository, queriesRepo QueriesRepository) *Service {
    return &Service{repo: repo, queriesRepo: queriesRepo}
}
```

**Changes (wire.go or DI setup):**
1. Find where `NewService` is called
2. Update to pass queries repository: `NewServiceWithQueries(gridRepo, queriesRepo)`

**Verification:**
- Application starts without import/dependency errors
- Service struct has `queriesRepo` field when constructed with `NewServiceWithQueries`

**Estimated Lines:** +15

---

## Task 4: Implement ExecuteQueryByID Service Method

**Task ID:** T4  
**Title:** Implement ExecuteQueryByID service method  
**File(s):** `backend/internal/domain/grid/service.go`  
**Status:** [x] Completed

**Description:**
Implement the `ExecuteQueryByID` method that fetches the saved query by ID, looks up the grid to get `baseQuery`, and executes the query. Also add helper functions `convertSort` and `convertFilters`.

**Changes:**
1. Add `"strings"` import if not present
2. Add `ExecuteQueryByID` method:

```go
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
```

3. Add `convertSort` helper:

```go
func convertSort(sort []queriesdomain.QuerySort) []SortCondition {
    result := make([]SortCondition, len(sort))
    for i, s := range sort {
        result[i] = SortCondition{
            Field:     s.Field,
            Direction: s.Direction,
        }
    }
    return result
}
```

4. Add `convertFilters` helper:

```go
func convertFilters(filters []queriesdomain.QueryFilter) []FilterCondition {
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

**Verification:**
- Unit test: `ExecuteQueryByID` with valid query ID calls `queriesRepo.GetByID`, then `repo.FindByID`, then `repo.ExecuteQuery`
- Unit test: `ExecuteQueryByID` with non-existent ID returns `ErrQueryNotFound`
- Unit test: `ExecuteQueryByID` when grid not found returns `ErrGridNotFound`
- Unit test: `convertSort` and `convertFilters` correctly map domain types

**Estimated Lines:** +65

---

## Task 5: Update Handler to Route on QueryID Presence

**Task ID:** T5  
**Title:** Add QueryID to GridDataRequest and update handler  
**File(s):** `backend/internal/adapters/api/grid/handler.go`  
**Status:** [x] Completed

**Description:**
Add the optional `QueryID` field to `GridDataRequest` struct and update `ExecuteData` handler to branch on whether `QueryID` is provided.

**Changes:**
1. Add `QueryID string` field to `GridDataRequest` struct with `json:"queryId,omitempty"`
2. In `ExecuteData` handler, after parsing request:
   - If `QueryID != ""`: call `ExecuteQueryByID(ctx, QueryID, page, pageSize)` and handle errors
   - If `QueryID == ""` and `GridID == 0`: return 400 error
   - If `QueryID == ""` and `GridID != 0`: use existing flat-field execution (backward compatible)
3. Add error handling for `ErrQueryNotFound` and `ErrGridNotFound` with 404 responses

**Verification:**
- POST `/grid/data` with `{"queryId": "abc123", "page": 1, "pageSize": 20}` routes to `ExecuteQueryByID`
- POST `/grid/data` with `{"gridId": 5, "fields": [1,2]}` routes to existing `ExecuteQuery` (backward compatible)
- POST `/grid/data` with `{"queryId": "nonexistent"}` returns 404
- POST `/grid/data` with no `queryId` and no `gridId` returns 400

**Estimated Lines:** +42

---

## Task 6: Update Frontend Query Service

**Task ID:** T6  
**Title:** Update frontend query.service.ts  
**File(s):** `frontend/src/app/core/services/query.service.ts`  
**Status:** [x] Completed

**Description:**
Change `executeQuery` to accept `(queryId, page, pageSize)` and send flat structure. Add `executeQueryDirect` for backward compatibility.

**Changes:**
1. Modify `executeQuery` signature:
   - Old: `executeQuery(gridId: number, query: GridQuery, page: number = 1)`
   - New: `executeQuery(queryId: string, page: number = 1, pageSize: number = 20)`
2. Send flat payload: `{ queryId, page, pageSize }` — no nested `query` object
3. Add new `executeQueryDirect` method for backward compat:
   - Signature: `executeQueryDirect(gridId: number, fields: number[], filters: FilterCondition[], sort: SortCondition[], page: number = 1, pageSize: number = 20)`
   - Sends full flat structure `{ gridId, fields, filters, sort, page, pageSize }`

**Verification:**
- `executeQuery("abc123", 1, 20)` sends POST with `{ queryId: "abc123", page: 1, pageSize: 20 }`
- `executeQueryDirect(5, [1,2], [], [], 1, 20)` sends POST with full flat structure
- Existing call sites that use the old signature need to be updated to use `executeQueryDirect`

**Estimated Lines:** +35

---

## Task 7: Update Grid Component to Pass QueryId

**Task ID:** T7  
**Title:** Update grid component to pass queryId  
**File(s):** `frontend/src/app/shared/components/data-grid/` (or wherever grid/query execution is triggered)  
**Status:** [x] Completed

**Description:**
Find where `executeQuery` is called in the grid component and update it to pass `queryId` from the selected query instead of `gridId` and nested query object.

**Changes:**
1. Find grid component or data-grid component that calls `queryService.executeQuery()`
2. Update the call to use new signature: `executeQuery(selectedQuery.id, page, pageSize)`
3. If the component currently passes `gridId` and `query` object, update to use `executeQueryDirect` instead

**Verification:**
- Grid component works when user selects a saved query and triggers data load
- Pagination works with the new signature
- Custom queries (without saved query) continue to work via `executeQueryDirect`

**Estimated Lines:** +15

---

## Implementation Order

| Order | Task ID | Title | Risk |
|-------|---------|-------|------|
| 1 | T1 | Add ErrQueryNotFound error | Low |
| 2 | T2 | Verify queries repository GetByID | Low (no change) |
| 3 | T3 | Add queriesRepo to Grid Service | Medium |
| 4 | T4 | Implement ExecuteQueryByID service method | Medium |
| 5 | T5 | Update Handler to route on QueryID presence | Medium |
| 6 | T6 | Update Frontend Query Service | Medium |
| 7 | T7 | Update Grid Component to pass queryId | Medium |

---

## File Summary

| Task | File | Lines | Type |
|------|------|-------|------|
| T1 | `backend/internal/domain/grid/errors.go` | +2 | Modified |
| T2 | `backend/internal/adapters/db/queries/query_repository.go` | 0 | Verified |
| T3 | `backend/internal/domain/grid/service.go` | +15 | Modified |
| T4 | `backend/internal/domain/grid/service.go` | +65 | Modified |
| T5 | `backend/internal/adapters/api/grid/handler.go` | +42 | Modified |
| T6 | `frontend/src/app/core/services/query.service.ts` | +35 | Modified |
| T7 | `frontend/src/app/shared/components/data-grid/` | +15 | Modified |
| **Total** | | **~174** | |

---

## Dependencies

- T3 (queriesRepo field) must be completed before T4 (ExecuteQueryByID uses queriesRepo)
- T4 (ExecuteQueryByID) must be completed before T5 (handler calls this method)
- T6 (frontend executeQuery change) should align with T5 (backend handler change)

---

## Rollback Summary

| Task | Rollback Action |
|------|-----------------|
| T1 | Remove ErrQueryNotFound from errors.go |
| T2 | No rollback needed |
| T3 | Remove queriesRepo field, revert NewServiceWithQueries |
| T4 | Remove ExecuteQueryByID method and helper functions |
| T5 | Remove QueryID field, restore original ExecuteData logic |
| T6 | Restore original executeQuery signature, remove executeQueryDirect |
| T7 | Revert grid component to use old executeQuery signature |
