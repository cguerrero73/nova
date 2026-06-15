# Tasks: funcionamiento-de-querybuilder

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 15–25 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | N/A |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: N/A
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Fix3 critical QueryBuilder bugs | PR 1 | All3 tasks in one PR; tests included |

## Phase 1: Core Bug Fixes

- [ ] 1.1 **T1**: In `reset()` (line ~145), replace `this.availableFieldsList.set([])` with `this.updateAvailableFields()`
- [ ] 1.2 **T2**: In `open()` (lines ~112-129), add fallback: `if (!query && this.initialQuery()) { query = this.initialQuery(); }` and populate `name`, `selectedFields`, `currentSort`, `filters` from query
- [ ] 1.3 **T3**: In `ExecuteQueryByID` (line ~216), change `nil` to `config.Fields` in `s.repo.ExecuteQuery(ctx, grid.BaseQuery, config.Fields, ...)`

## Phase 2: Testing

- [ ] 2.1 Write unit test: `reset()` calls `updateAvailableFields()` — verify `availableFieldsList` contains all unselected fields
- [ ] 2.2 Write unit test: `open()` with no arg falls back to `initialQuery` — verify form populated
- [ ] 2.3 Write unit test: `ExecuteQueryByID` passes `config.Fields` to repository — verify third arg

## Phase3: Verification

- [ ] 3.1 Verify reset shows all unselected fields in UI
- [ ] 3.2 Verify saved query pre-fills form when builder opens
- [ ] 3.3 Verify saved query field selection is applied by backend

## Task Details

### T1: Fix reset() field list
- **File**: `frontend/src/app/shared/components/query-builder/query-builder.component.ts`
- **Change**: Replace `this.availableFieldsList.set([])` with `this.updateAvailableFields()` in `reset()` method (~line 145)
- **Verification**: After reset, available fields list shows all unselected fields
- **Estimated lines**: 1

### T2: Fix initialQuery loading in open()
- **File**: `frontend/src/app/shared/components/query-builder/query-builder.component.ts`
- **Change**: In `open()` method, add fallback to `this.initialQuery()` and populate form fields from query
- **Verification**: Opening builder after selecting a query pre-fills the form
- **Estimated lines**: 8–12

### T3: Fix ExecuteQueryByID fields passthrough
- **File**: `backend/internal/domain/grid/service.go`
- **Change**: In `ExecuteQueryByID` (~line 216), pass `config.Fields` instead of `nil` to `s.repo.ExecuteQuery()`
- **Verification**: Saved query field selection is applied when executing
- **Estimated lines**: 1
