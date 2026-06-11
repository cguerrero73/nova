# Tasks: Fix Query List in Query Builder

## Overview

Break down the SDD change "fix query list in query builder" into discrete, actionable implementation tasks.

---

## Task 1: Audit Frontend Query Dropdown

**Task ID:** T1  
**Title:** Audit frontend query dropdown template  
**File(s):** `frontend/src/app/features/users/screens/user-list/user-list.component.html`  
**Status:** [x] Completed

**Findings:**
- Audit completed via delegation (agent: uptight-salmon-lobster)
- Signal singleton pattern verified (queryService providedIn: 'root')
- queries input on QueryBuilderComponent not used by parent
- Root cause: static placeholder option with empty value was showing
- Backend SQL filtering verified correct: `WHERE qry_grid_id = $1`

**Verification:**
- Dropdown shows only queries from the current grid
- No spurious options appear after fix

**Estimated Lines:** 0 (audit only)

---

## Task 2: Audit Backend /queries Endpoint

**Task ID:** T2  
**Title:** Audit backend /queries endpoint filtering  
**File(s):** `backend/internal/adapters/api/queries/handler.go`  
**Status:** [x] Completed (audit via delegation)

**Findings:**
- Handler correctly extracts gridId via QueryInt
- Calls service method with gridId parameter
- Backend filtering verified correct

**Verification:**
- GET `/queries?gridId=1` returns only queries for grid 1
- GET `/queries` (no gridId) returns empty list or error

**Estimated Lines:** 0 (audit only)

---

## Task 3: Audit Query Repository ListByGridID

**Task ID:** T3  
**Title:** Audit query repository ListByGridID  
**File(s):** `backend/internal/adapters/db/queries/query_repository.go`  
**Status:** [x] Completed (audit via delegation)

**Findings:**
- Repository SQL correctly filters: `WHERE qry_grid_id = $1`
- Also filters by user ownership: `(qry_is_public = true OR qry_user_id = $2)`
- No issues found

**Verification:**
- SQL query includes `WHERE qry_grid_id = $1`
- Only queries for the specified grid are returned

**Estimated Lines:** 0 (audit only)

---

## Task 4: Verify onQuerySelect Empty Handling

**Task ID:** T4  
**Title:** Verify onQuerySelect handles empty selection  
**File(s):** `frontend/src/app/features/users/screens/user-list/user-list.component.ts`  
**Status:** [x] Completed (audit via delegation)

**Findings:**
- onQuerySelect already handles empty queryId (line 473)
- Calls loadDefaultData(), clearSelection(), loadedPages.clear()
- No changes needed

**Verification:**
- Selecting empty option clears selection and loads default data
- No errors occur when clearing selection

**Estimated Lines:** 0 (audit only)

---

## Task 5: Fix Identified Issues

**Task ID:** T5  
**Title:** Fix any identified issues  
**File(s):** `frontend/src/app/features/users/screens/user-list/user-list.component.html`  
**Status:** [x] Completed

**Changes:**
- Removed static `<option value="">Seleccionar query...</option>` placeholder (line 38)
- Dropdown now shows only actual queries from `queryService.queries()`

**Verification:**
- All tests pass
- Query dropdown only shows correct options
- Commit: `fix(query-builder): remove placeholder option from query dropdown`

**Estimated Lines:** 1 (1 deletion)

---

## Implementation Order

| Order | Task ID | Title | Risk |
|-------|---------|-------|------|
| 1 | T1 | Audit frontend dropdown | Low |
| 2 | T2 | Audit backend handler | Low |
| 3 | T3 | Audit repository | Low |
| 4 | T4 | Audit onQuerySelect | Low |
| 5 | T5 | Fix identified issues | Medium |

---

## File Summary

| Task | File | Lines | Type |
|------|------|-------|------|
| T1 | `user-list.component.html` | 0 | Audit |
| T2 | `queries/handler.go` | 0 | Audit |
| T3 | `query_repository.go` | 0 | Audit |
| T4 | `user-list.component.ts` | 0 | Audit |
| T5 | TBD | TBD | Fix |
| **Total** | | **TBD** | |

---

## Dependencies

- T2 (handler audit) depends on T3 (repository audit)
- T5 (fix) depends on T1-T4 (all audits)

---

## Rollback Summary

| Task | Rollback Action |
|------|-----------------|
| T1 | Revert template changes if any |
| T2 | Revert handler changes if any |
| T3 | Revert repository changes if any |
| T4 | Revert onQuerySelect changes if any |
| T5 | Revert fixes applied in T5 |
