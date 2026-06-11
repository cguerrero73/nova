# Proposal: Fix Query List in Query Builder

## Intent

Fix the query list in the query builder dropdown so it only shows queries that belong to the current grid, removing any spurious options that don't belong to the grid's query array.

## Problem Statement

The query builder dropdown shows a "Seleccionar query..." option with an empty value, and potentially other options that don't belong to the current grid's associated queries. The dropdown should only display queries from `queryService.queries()` that are associated with the current grid via `gridId`.

## Scope

### In Scope
- Verify that the query dropdown only shows queries from the current grid's query array
- Ensure the "Seleccionar query..." option correctly clears selection and loads default data
- Verify backend `/queries` endpoint filters by `gridId` correctly
- Handle edge case when no queries exist for a grid

### Out of Scope
- Creating new queries
- Editing existing queries
- Deleting queries
- Query execution logic

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `frontend/src/app/features/users/screens/user-list/user-list.component.html` | Modified | Query dropdown template |
| `frontend/src/app/core/services/query.service.ts` | Modified | Query loading and state |
| `backend/internal/adapters/api/queries/handler.go` | Modified | Verify filtering by gridId |
| `backend/internal/adapters/db/queries/query_repository.go` | Modified | Verify SQL filtering by gridId |

## Approach

1. **Frontend audit**: Check how queries are loaded and displayed
2. **Backend audit**: Verify `/queries` endpoint correctly filters by `gridId`
3. **Fix**: Ensure only grid-associated queries appear in the dropdown

## Success Criteria

1. Query dropdown shows only queries where `q.gridId` matches the current grid
2. The "Seleccionar query..." option (empty value) works correctly to clear selection
3. No duplicate or spurious query options appear in the dropdown
4. When no queries exist for a grid, appropriate empty state is shown
