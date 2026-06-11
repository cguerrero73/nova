# Spec: Fix Grid Query Execution by ID

## Overview

This document details the specifications for fixing grid query execution so the frontend only passes `queryId` and pagination parameters, while the backend fetches the saved query configuration, looks up the grid's base query, and constructs the SQL dynamically.

---

## Spec 1: GridDataRequest with QueryID

**File**: `backend/internal/adapters/api/grid/handler.go`

### What
Add an optional `QueryID` field to `GridDataRequest` struct to allow clients to execute a saved query by ID instead of providing flat field/filter/sort configuration.

### Where
- `GridDataRequest` struct (line 21-29)

### Expected Behavior
```go
type GridDataRequest struct {
    GridID    int                          `json:"gridId"`
    QueryID   string                       `json:"queryId,omitempty"`  // NEW: optional query identifier
    FirstLoad bool                         `json:"firstLoad"`
    Fields    []int                        `json:"fields"`
    Sort      []griddomain.SortCondition   `json:"sort"`
    Filters   []griddomain.FilterCondition `json:"filters"`
    Page      int                          `json:"page"`
    PageSize  int                          `json:"pageSize"`
}
```

### Actual Behavior
`GridDataRequest` does not have a `QueryID` field. The handler always builds `GridQueryConfig` from flat fields and calls `ExecuteQuery` with `gridID` only.

### Fix Description
1. Add `QueryID string` field to `GridDataRequest` with `json:"queryId,omitempty"`
2. In `ExecuteData` handler:
   - If `QueryID` is provided (non-empty) AND `GridID` is zero: call `ExecuteQueryByID(ctx, QueryID, page, pageSize)`
   - If `QueryID` is provided AND `GridID` is also provided: call `ExecuteQueryByID` with request-level overrides
   - If `QueryID` is empty: use existing flat-field behavior (backward compatible)
3. The handler should NOT build `GridQueryConfig` from flat fields when `QueryID` is provided

### Test Scenario
```
Scenario: Execute query using only QueryID
Given a saved query exists with ID "abc123" linked to grid ID 5
When POST /grid/data is sent with body:
  {
    "queryId": "abc123",
    "page": 1,
    "pageSize": 20
  }
Then the backend:
  1. Fetches SavedQuery "abc123" from queries repository
  2. Extracts GridID (5) from SavedQuery
  3. Fetches Grid by ID 5 to get baseQuery
  4. Extracts fields/filters/sort from SavedQuery.Query
  5. Executes SQL using baseQuery + merged configuration
  6. Returns paginated results

Scenario: Execute query with QueryID and request overrides
Given a saved query exists with ID "abc123"
When POST /grid/data is sent with body:
  {
    "queryId": "abc123",
    "page": 2,
    "pageSize": 50,
    "sort": [{"field": 1, "direction": 2}]
  }
Then the backend:
  1. Fetches SavedQuery "abc123"
  2. Uses saved query's fields/filters
  3. OVERRIDES sort with request-level sort (request takes precedence)
  4. Uses request-level page=2, pageSize=50
  5. Returns paginated results

Scenario: Backward compatibility (no QueryID)
Given no queryId is provided
When POST /grid/data is sent with body:
  {
    "gridId": 5,
    "fields": [1, 2, 3],
    "page": 1,
    "pageSize": 20
  }
Then the backend behaves as before:
  - Builds GridQueryConfig from flat fields
  - Calls ExecuteQuery(gridID, config, page, pageSize)
```

---

## Spec 2: ExecuteQuery Service Fix

**File**: `backend/internal/domain/grid/service.go`

### What
Modify `ExecuteQuery` (or add `ExecuteQueryByID`) to fetch `SavedQuery` by ID, then `Grid` by `SavedQuery.GridID` to obtain `baseQuery`, and execute the query with proper configuration.

### Where
- `Service` struct (line 11-14)
- `ExecuteQuery` method (line 136-147)

### Expected Behavior
```go
// ExecuteQueryByID fetches saved query, resolves grid, and executes
func (s *Service) ExecuteQueryByID(ctx context.Context, queryID string, page, pageSize int) (*GridResult, error) {
    // 1. Fetch SavedQuery by ID
    savedQuery, err := s.queriesRepo.GetByID(ctx, queryID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "query not found") {
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
    
    // 3. Convert saved query's Fields/Sort/Filters to domain types
    // SavedQuery.Query.Fields []int -> config.Fields
    // SavedQuery.Query.Filters []QueryFilter -> []FilterCondition
    // SavedQuery.Query.Sort []QuerySort -> []SortCondition
    
    // 4. Build final config (request-level overrides applied by handler)
    config := &GridQueryConfig{
        Fields:  savedQuery.Query.Fields,
        Sort:    convertSort(savedQuery.Query.Sort),
        Filters: convertFilters(savedQuery.Query.Filters),
    }
    
    // 5. Call repo.ExecuteQuery with baseQuery from grid
    return s.repo.ExecuteQuery(ctx, grid.BaseQuery, config.Fields, config.Filters, config.Sort, page, pageSize)
}
```

### Actual Behavior
```go
// Current ExecuteQuery (line 136-147)
func (s *Service) ExecuteQuery(ctx context.Context, gridID int, config *GridQueryConfig, page, pageSize int) (*GridResult, error) {
    // Default pagination
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 20
    }
    
    return s.repo.ExecuteQuery(ctx, "", nil, config.Filters, config.Sort, page, pageSize)
    // ^ BUG: passes empty string as baseQuery, causing all queries to fail
}
```

### Fix Description
1. Add `queriesRepo QueriesRepository` field to `Service` struct
2. Add `ErrQueryNotFound` error to `errors.go`
3. Add `ExecuteQueryByID(ctx context.Context, queryID string, page, pageSize int)` method
4. Add helper functions `convertSort()` and `convertFilters()` to convert from `queries.QuerySort/QueryFilter` to `grid.SortCondition/FilterCondition`
5. The existing `ExecuteQuery` method remains for backward compatibility when clients send flat fields directly

### Service Struct Changes
```go
type Service struct {
    repo        GridRepository
    fieldRepo   FieldsRepository
    queriesRepo QueriesRepository  // NEW: for fetching saved queries
}
```

### Test Scenario
```
Scenario: ExecuteQueryByID with valid query ID
Given SavedQuery exists with id="q1", gridId=5, query.fields=[1,2], query.filters=[], query.sort=[]
And Grid exists with id=5 and baseQuery="SELECT * FROM users"
When ExecuteQueryByID(ctx, "q1", 1, 20) is called
Then:
  1. queriesRepo.GetByID("q1") is called and returns SavedQuery
  2. repo.FindByID(5) is called and returns Grid
  3. repo.ExecuteQuery(ctx, "SELECT * FROM users", [1,2], [], [], 1, 20) is called
  4. Result is returned

Scenario: ExecuteQueryByID with non-existent query ID
Given no SavedQuery exists with id="nonexistent"
When ExecuteQueryByID(ctx, "nonexistent", 1, 20) is called
Then ErrQueryNotFound is returned with 404 status in handler

Scenario: ExecuteQueryByID where grid not found
Given SavedQuery exists with id="q1" but gridId=999 does not exist
When ExecuteQueryByID(ctx, "q1", 1, 20) is called
Then ErrGridNotFound is returned with 404 status in handler

Scenario: ExecuteQuery still works without QueryID (backward compat)
Given a GridQueryConfig with fields=[1,2], filters=[], sort=[]
When ExecuteQuery(ctx, 5, config, 1, 20) is called
Then repo.ExecuteQuery is called with proper baseQuery (still has bug, needs separate fix)
```

---

## Spec 3: Frontend Query Service

**File**: `frontend/src/app/core/services/query.service.ts`

### What
Change `executeQuery()` to send a simplified flat structure `{ queryId, page, pageSize }` instead of the nested `{ gridId, query, page }` structure.

### Where
- `executeQuery()` method (line 30-47)

### Expected Behavior
```typescript
// Execute query using saved query ID (new behavior)
executeQuery(
  queryId: string,
  page: number = 1,
  pageSize: number = 20
): Observable<PaginatedResponse<unknown>> {
  return this.api
    .postRaw<PaginatedResponse<unknown>>('/grid/data', {
      queryId,    // flat structure - no nested query object
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

// Alternative: also support direct queries without queryId (backward compat)
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

### Actual Behavior
```typescript
// Current executeQuery (line 30-47)
executeQuery(
  gridId: number,
  query: GridQuery,
  page: number = 1
): Observable<PaginatedResponse<unknown>> {
  return this.api
    .postRaw<PaginatedResponse<unknown>>('/grid/data', {
      gridId,
      query,  // nested object: { fields, sort, filters, pagination }
      page,
    })
    .pipe(...)
}
```

### Fix Description
1. Change `executeQuery()` signature to accept `queryId` as first parameter
2. Send flat structure `{ queryId, page, pageSize }` - no nested `query` object
3. Maintain backward compatibility by keeping `executeQueryDirect()` for cases without queryId
4. Update `selectQuery()` to call `executeQuery(selectedQuery.id)` instead of setting `currentQuery`

### Test Scenario
```
Scenario: Execute query with queryId
Given user has selected a saved query with id="abc123"
When executeQuery("abc123", 1, 20) is called
Then POST /grid/data is sent with body:
  {
    "queryId": "abc123",
    "page": 1,
    "pageSize": 20
  }
And response is handled correctly with meta update

Scenario: Load queries and select one for execution
Given queries are loaded for gridId=5 with results including {id: "q1", name: "My Query"}
When user selects query q1
Then selectQuery(q1) is called
And currentQuery is set to q1.query
And later when executeQuery is called, it uses q1.id

Scenario: Execute direct query without queryId (backward compat)
Given user builds a custom query with fields=[1,2], filters=[], sort=[]
When executeQueryDirect(5, [1,2], [], [], 1, 20) is called
Then POST /grid/data is sent with flat structure:
  {
    "gridId": 5,
    "fields": [1, 2],
    "filters": [],
    "sort": [],
    "page": 1,
    "pageSize": 20
  }
```

---

## Spec 4: Query Repository GetByID

**File**: `backend/internal/adapters/db/queries/query_repository.go`

### What
Verify `GetByID` implementation exists and works correctly. It already exists at line 66-97, but verify it properly handles the "query not found" case.

### Where
- `PgQueryRepository.GetByID()` method (line 66-97)

### Expected Behavior
```go
func (r *PgQueryRepository) GetByID(ctx context.Context, id string) (*queries.SavedQuery, error) {
    var q queries.SavedQuery
    var queryJSON []byte

    err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
        sql := `
            SELECT qry_id, qry_grid_id, qry_name, qry_user_id, qry_is_public,
                   qry_is_default, qry_query, qry_created_at, qry_updated_at
            FROM eamqueries WHERE qry_id = $1`

        err := tx.QueryRow(ctx, sql, id).Scan(
            &q.ID, &q.GridID, &q.Name, &q.UserID, &q.IsPublic, &q.IsDefault,
            &queryJSON, &q.CreatedAt, &q.UpdatedAt,
        )
        if errors.Is(err, pgx.ErrNoRows) {
            return fmt.Errorf("query not found: %w", err)  // Must wrap with context
        }
        if err != nil {
            return err
        }

        if len(queryJSON) > 0 {
            if err := json.Unmarshal(queryJSON, &q.Query); err != nil {
                return err
            }
        }

        return nil
    })

    return &q, err
}
```

### Actual Behavior
The implementation exists and looks correct. The error message "query not found: %w" is properly wrapped. The service layer should check for this error pattern.

### Fix Description
No code changes needed. Verify:
1. `infraDB.RunInTenantTx` properly propagates errors
2. `pgx.ErrNoRows` is correctly detected and wrapped
3. JSON unmarshaling of `q.Query` works correctly
4. The returned error message contains "query not found" so service can detect it

### Test Scenario
```
Scenario: GetByID with existing query
Given a SavedQuery exists in database with qry_id="abc123"
When GetByID(ctx, "abc123") is called
Then SavedQuery is returned with all fields populated
And Query.Query contains the unmarshaled GridQuery object
And error is nil

Scenario: GetByID with non-existent query
Given no SavedQuery exists with qry_id="nonexistent"
When GetByID(ctx, "nonexistent") is called
Then error is returned
And error.Error() contains "query not found"
And error wraps pgx.ErrNoRows

Scenario: GetByID with corrupted JSON in qry_query
Given SavedQuery exists with qry_id="badjson" but qry_query contains invalid JSON
When GetByID(ctx, "badjson") is called
Then error is returned from json.Unmarshal
And SavedQuery is not returned
```

---

## Error Handling Summary

| Scenario | HTTP Status | Error Code | Message |
|----------|-------------|------------|---------|
| QueryID not found | 404 | NOT_FOUND | Query not found |
| Grid not found (from SavedQuery.GridID) | 404 | NOT_FOUND | Grid not found |
| Invalid request body | 400 | BAD_REQUEST | Bad request |
| GridID required (when QueryID not provided) | 400 | BAD_REQUEST | gridId is required |
| Internal error | 500 | INTERNAL | Internal server error |

---

## Acceptance Criteria

- [ ] Grid loads data when frontend sends only `{ queryId, page, pageSize }`
- [ ] Pagination (page, pageSize) works correctly
- [ ] Filters and sort from saved query are applied to SQL execution
- [ ] Request-level overrides take precedence over saved query defaults
- [ ] Existing calls without `queryId` continue to work (backward compatible)
- [ ] 404 returned if `queryId` does not exist
- [ ] 404 returned if grid referenced by saved query does not exist