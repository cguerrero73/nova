# Proposal: Fix Grid Query Execution by ID

## Intent

Fix grid query execution so the frontend only passes `queryId` and pagination parameters, while the backend fetches the saved query configuration, looks up the grid's base query, and constructs the SQL dynamically. Currently `ExecuteQuery` passes an empty `baseQuery` string to the repository, causing every query to fail.

## Scope

### In Scope
- Backend `GridDataRequest` accepts `QueryID` as optional field (backward compatible)
- Backend `ExecuteQuery` service method fetches `SavedQuery` by ID, then `Grid` by `SavedQuery.gridId` to obtain `baseQuery`
- Backend merges saved query's `fields`, `filters`, `sort` with request-level overrides
- Frontend `query.service.ts` sends simplified `{ queryId, page, pageSize }` payload

### Out of Scope
- Frontend query builder UI (creating/editing saved queries)
- New query persistence endpoints
- Query sharing or user-specific query visibility

## Capabilities

### New Capabilities
- `grid-query-execution-by-id`: Grid data endpoint accepts `queryId` to load saved query config and execute SQL

### Modified Capabilities
- `grid-data-execution`: Existing flat-field request contract extended with optional `queryId`; service layer now resolves query+grid before SQL execution

## Approach

1. **Handler** (`grid/handler.go`): Add `QueryID` field to `GridDataRequest`. If `QueryID` is provided, skip building `GridQueryConfig` from flat fields and instead pass `QueryID` to service.

2. **Service** (`grid/service.go`): New `ExecuteQueryByID` method (or parameterize existing `ExecuteQuery`):
   - Fetch `SavedQuery` by ID via queries repository
   - Fetch `Grid` by `SavedQuery.GridID` to get `baseQuery`
   - Convert saved query's `Fields/Sort/Filters` to domain types
   - Merge with any request-level overrides
   - Call `repo.ExecuteQuery(baseQuery, ...)`

3. **Repository** (`grid/grid_repository.go`): `ExecuteQuery` already accepts `baseQuery`; no changes needed.

4. **Queries Repository** (`queries/query_repository.go`): Add `FindByID` implementation if not already present.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/internal/adapters/api/grid/handler.go` | Modified | Accept `QueryID` in `GridDataRequest` |
| `backend/internal/domain/grid/service.go` | Modified | `ExecuteQuery` resolves query+grid before execution |
| `backend/internal/domain/queries/repository.go` | Modified | `GetByID` interface already exists; ensure implementation |
| `backend/internal/adapters/db/queries/query_repository.go` | Modified | Implement `GetByID` if missing |
| `frontend/src/services/query.service.ts` | Modified | Send `{ queryId, page, pageSize }` only |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Breaking existing API consumers | Low | Add `QueryID` as optional field; existing calls without it continue working |
| SavedQuery config merging logic | Medium | Define explicit precedence: request overrides > saved query defaults |
| Query not found (404) | Low | Return clear error "Query not found" with 404 status |

## Rollback Plan

- **Revert handler**: Remove `QueryID` field from `GridDataRequest`, restore flat-field parsing in `ExecuteData`
- **Revert service**: Restore `ExecuteQuery` to accept `baseQuery` directly (no query lookup)
- **Revert frontend**: Restore sending full `{ gridId, fields, sort, filters, page }` payload
- All changes are additive; rollback is straightforward

## Dependencies

- Queries repository `GetByID` must be implemented before service layer work

## Success Criteria

- [ ] Grid loads data when frontend sends only `{ queryId, page, pageSize }`
- [ ] Pagination (page, pageSize) works correctly
- [ ] Filters and sort from saved query are applied to SQL execution
- [ ] Existing calls without `queryId` continue to work (backward compatible)
- [ ] 404 returned if `queryId` does not exist