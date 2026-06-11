# Design: Fix Query List in Query Builder

## Overview

This document describes the technical design for fixing the query list in the query builder dropdown.

---

## Design1: Frontend Query List Filtering

### Problem
The query dropdown in `user-list.component.html` uses `queryService.queries()` which should only contain queries for the current grid. However, if the backend returns all queries or if there's state pollution, spurious options may appear.

### Solution
Verify the complete flow from backend filtering to frontend display.

### Implementation

**Step 1: Verify query loading in user-list.component.ts**
```typescript
private loadGridQueries(gridId: number): void {
  this.queryService.loadByGridId(gridId).subscribe({
    next: (queries) => {
      // queries should only contain queries for this gridId
      console.log('[UserList] Queries loaded:', queries.length);
    }
  });
}
```

**Step 2: Verify dropdown template uses correct data**
```html
<select [value]="selectedQueryId()" (change)="onQuerySelect($event)">
  <option value="">{{ t['query.select'] || 'Seleccionar query...' }}</option>
  @for (q of queryService.queries(); track q.id) {
    <option [value]="q.id">{{ q.name }}{{ q.isDefault ? ' (default)' : '' }}</option>
  }
</select>
```

**Step 3: Verify queryService.queries state**
The `queries` signal should only be set by `loadByGridId` or `loadByGridName`, never by a blanket `loadAll` call.

---

## Design 2: Backend Query Filtering

### Problem
The backend `/queries` endpoint may not be filtering by `gridId`, returning all queries instead of only those belonging to the specified grid.

### Solution
Verify and fix the handler and repository to filter by `gridId`.

### Implementation

**Handler (queries/handler.go)**
```go
func (h *QueryHandler) List(c *fiber.Ctx) error {
  gridIDStr := c.Query("gridId")
  
  if gridIDStr != "" {
    gridID, err := strconv.Atoi(gridIDStr)
    if err != nil {
      return c.Status(400).JSON(errors.New("BAD_REQUEST", "invalid gridId", 400))
    }
    queries, err := h.queriesRepo.ListByGridID(c.Context(), gridID)
    if err != nil {
      return c.Status(500).JSON(errors.ErrInternal)
    }
    return c.JSON(fiber.Map{
      "success": true,
      "data":    queries,
    })
  }
  
  // If no gridId provided, return empty or error
  return c.JSON(fiber.Map{
    "success": true,
    "data":    []interface{}{},
  })
}
```

**Repository (queries/query_repository.go)**
```go
func (r *PgQueryRepository) ListByGridID(ctx context.Context, gridID int) ([]*queriesdomain.SavedQuery, error) {
  query := `
    SELECT qry_id, qry_grid_id, qry_name, qry_user_id, qry_is_public,
           qry_is_default, qry_query, qry_created_at, qry_updated_at
    FROM eamqueries
    WHERE qry_grid_id = $1
    ORDER BY qry_is_default DESC, qry_name ASC
  `
  
  rows, err := r.pool.Query(ctx, query, gridID)
  // ...
}
```

---

## Design 3: Empty Selection Handling

### Problem
When user selects the empty "Seleccionar query..." option, the state may not be properly cleared.

### Solution
Ensure `onQuerySelect` handles empty value correctly.

### Implementation

```typescript
onQuerySelect(event: Event) {
  const queryId = (event.target as HTMLSelectElement).value;

  if (!queryId) {
    // User cleared selection - load default grid data
    this.loadDefaultData();
    this.queryService.clearSelection();
    this.selectedQueryId.set('');
    this.loadedPages.clear();
    return;
  }

  // User selected a query - execute it
  const query = this.queryService.queries().find((q) => q.id === queryId);
  if (query) {
    this.queryService.selectQuery(query);
    this.selectedQueryId.set(queryId);
    this.loadedPages.clear();
    this.executeQuery(query.query);
  }
}
```

---

## File Summary

| File | Change | Lines |
|------|--------|-------|
| `frontend/src/app/features/users/screens/user-list/user-list.component.html` | Verify dropdown template | ~5 |
| `frontend/src/app/features/users/screens/user-list/user-list.component.ts` | Verify onQuerySelect | ~10 |
| `frontend/src/app/core/services/query.service.ts` | Verify loadByGridId | ~5 |
| `backend/internal/adapters/api/queries/handler.go` | Add/verify gridId filtering | ~15 |
| `backend/internal/adapters/db/queries/query_repository.go` | Add ListByGridID if missing | ~20 |
| **Total** | | **~55** |
