# Proposal: funcionamiento-de-querybuilder

## Intent

Fix 14 bugs and usability issues in the QueryBuilder component (popup UI with Fields/Sort/Filter tabs) where three critical bugs prevent saved queries from loading correctly and the fields list from displaying after reset.

## Scope

### In Scope
- **Bug 1**: `reset()` clears `availableFieldsList` directly instead of calling `updateAvailableFields()` — fields list empty after reset
- **Bug 2**: `initialQuery` input exists but parent never passes selected query to `open()` — saved query never loads in layout
- **Bug 3**: `ExecuteQueryByID` passes `nil` for fields parameter — saved query field selection ignored by backend
- Add unit tests for QueryBuilder, QueryService, and ExecuteQueryByID
- Clear filter value when operator changes to IS_NULL/IS_NOT_NULL
- Extract hardcoded page size (20) into constant
- Add validation on filter values

### Out of Scope
- Redesign of the popup UI or tab structure
- Adding new filter operators beyond IS_NULL/IS_NOT_NULL
- Backend schema changes

## Capabilities

### Modified Capabilities
- `query-builder`: Fix reset behavior, initial query loading, and ExecuteQueryByID field handling

## Approach

1. **Phase 1 — Critical Bugs**: Fix `reset()` to use `updateAvailableFields()`, pass `initialQuery` via `open()`, fix `ExecuteQueryByID` to pass fields
2. **Phase 2 — Hardening**: Add tests, extract page size constant, clear filter on NULL operators
3. **Phase 3 — Validation**: Add filter value validation

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `src/components/QueryBuilder/` | Modified | reset(), open(), initialQuery handling |
| `src/services/QueryService.ts` | Modified | ExecuteQueryByID fields parameter |
| `src/services/GridService.ts` | Modified | Backend API call |
| `**/*.test.ts` | New | Unit tests for QueryBuilder, QueryService |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Backend field handling change breaks existing saved queries | Medium | Test with existing saved queries before merge; feature flag if needed |
| `updateAvailableFields()` has side effects | Low | Review implementation; add integration test |

## Rollback Plan

Revert `reset()` change: restore direct assignment `this.availableFieldsList = []`  
Revert `open()` change: remove `initialQuery` parameter handling  
Revert `ExecuteQueryByID`: restore `nil` fields parameter  

All changes are additive refactors with no schema migration.

## Dependencies

- None

## Success Criteria

- [ ] `reset()` calls `updateAvailableFields()` and fields list renders correctly
- [ ] `open(initialQuery)` loads saved query into layout when parent passes it
- [ ] `ExecuteQueryByID` passes selected fields to backend API
- [ ] Unit tests cover core QueryBuilder, QueryService, ExecuteQueryByID flows
- [ ] Filter value clears when operator changes to IS_NULL/IS_NOT_NULL