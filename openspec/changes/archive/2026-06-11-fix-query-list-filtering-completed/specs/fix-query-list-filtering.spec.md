# Spec: Fix Query List in Query Builder

## Overview

This document specifies the requirements for fixing the query list in the query builder dropdown to only show queries associated with the current grid.

---

## Spec 1: Query Dropdown Filters by Grid

**File**: `frontend/src/app/features/users/screens/user-list/user-list.component.html`

### What
The query dropdown should only display queries that belong to the current grid.

### Where
- Lines 30-42 in user-list.component.html

### Expected Behavior
```html
<select [value]="selectedQueryId()" (change)="onQuerySelect($event)">
  <option value="">{{ t['query.select'] || 'Seleccionar query...' }}</option>
  @for (q of queryService.queries(); track q.id) {
    <option [value]="q.id">{{ q.name }}{{ q.isDefault ? ' (default)' : '' }}</option>
  }
</select>
```

The dropdown should show only queries from `queryService.queries()` where `q.gridId` matches the current grid.

### Actual Behavior
The dropdown may show queries that don't belong to the current grid, or show an extra spurious option.

### Fix Description
1. Verify `queryService.queries()` only contains queries for the current grid
2. Verify `loadByGridId(gridId)` correctly filters on the backend
3. Ensure no hardcoded or mock queries appear in the list

### Test Scenario
```
Scenario: Query dropdown shows only grid's queries
Given grid "BMUSER" has3 associated queries: ["All Users", "Active Users", "Admins"]
When the user-list component loads
Then the dropdown shows:
  - "Seleccionar query..." (empty value)
  - "All Users" (default)
  - "Active Users"
  - "Admins"
And no other options appear
```

---

## Spec 2: Backend Filters Queries by gridId

**File**: `backend/internal/adapters/api/queries/handler.go`

### What
The `/queries` endpoint should filter queries by `gridId` parameter.

### Where
- `GET /queries?gridId=:id` handler

### Expected Behavior
```go
// Handler receives gridId from query params
gridID := c.Query("gridId")
queries, err := h.queriesRepo.ListByGridID(ctx, gridID)
```

### Actual Behavior
The handler may not be filtering by gridId, returning all queries.

### Fix Description
1. Verify handler extracts `gridId` from query parameters
2. Verify repository calls `ListByGridID(ctx, gridID)` not `FindAll()`
3. If filtering is missing, add it

### Test Scenario
```
Scenario: GET /queries?gridId=1 returns only grid 1's queries
Given queries exist for grid 1 and grid 2
When GET /queries?gridId=1 is called
Then only queries with qry_grid_id = 1 are returned
And queries with qry_grid_id = 2 are not included
```

---

## Spec 3: Empty Selection Clears and Loads Default

**File**: `frontend/src/app/features/users/screens/user-list/user-list.component.ts`

### What
When the user selects the empty "Seleccionar query..." option, the selection should clear and default data should load.

### Where
- `onQuerySelect` method (lines 470-488)

### Expected Behavior
```typescript
onQuerySelect(event: Event) {
  const queryId = (event.target as HTMLSelectElement).value;

  if (!queryId) {
    this.loadDefaultData();           // Load default grid data
    this.queryService.clearSelection(); // Clear query selection
    this.selectedQueryId.set('');     // Reset selected ID
    this.loadedPages.clear();         // Clear pagination cache
    return;
  }
  // ... execute selected query
}
```

### Actual Behavior
Clicking the empty option may not properly clear state or load default data.

### Fix Description
1. Verify `loadDefaultData()` is called when queryId is empty
2. Verify `clearSelection()` is called
3. Verify `loadedPages.clear()` resets pagination

### Test Scenario
```
Scenario: User clears query selection
Given a query is currently selected
When the user selects "Seleccionar query..." (empty option)
Then:
  - selectedQueryId is cleared to ''
  - default grid data is loaded
  - pagination is reset
  - no query is highlighted in dropdown
```
