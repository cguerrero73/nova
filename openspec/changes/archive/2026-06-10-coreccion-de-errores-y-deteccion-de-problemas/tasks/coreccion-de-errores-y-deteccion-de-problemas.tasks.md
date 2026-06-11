# Tasks: coreccion de errores y deteccion de problemas para llegar a un app estable

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~280–340 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR — all fixes are small, isolated, low-risk |
| Delivery strategy | single-pr |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | All bug fixes | PR 1 | 7 files, ~280–340 lines; all fixes independent and isolated |

---

## Phase 1: Critical Nil Pointer Bugs (Auth Backend)

- [ ] **1.1** — Fix `FindByRefreshToken` nil return on ErrNoRows
  - **File:** `backend/internal/adapters/db/auth/session_repository.go`
  - **Change:** Add `errors.Is(err, pgx.ErrNoRows)` check in closure (line 55–67). Return wrapped error instead of letting nil leak.
  - **Verification:** Call `FindByRefreshToken(ctx, "nonexistent-token")` → returns `(nil, err)` where `errors.Is(err, pgx.ErrNoRows)` is true.
  - **Est. lines:** 8

- [ ] **1.2** — Fix `FindByEmail` nil return on ErrNoRows
  - **File:** `backend/internal/adapters/db/auth/user_repository.go`
  - **Change:** Add `errors.Is(err, pgx.ErrNoRows)` check in closure (line 30–42). Return wrapped error instead of nil.
  - **Verification:** Call `FindByEmail(ctx, "nonexistent@example.com")` → returns `(nil, err)` where `errors.Is(err, pgx.ErrNoRows)` is true.
  - **Est. lines:** 8

- [ ] **1.3** — Fix `FindByCode` nil return on ErrNoRows
  - **File:** `backend/internal/adapters/db/auth/user_repository.go`
  - **Change:** Add `errors.Is(err, pgx.ErrNoRows)` check in closure (line 55–67). Return wrapped error instead of nil.
  - **Verification:** Call `FindByCode(ctx, "NONEXISTENT")` → returns `(nil, err)` where `errors.Is(err, pgx.ErrNoRows)` is true.
  - **Est. lines:** 8

---

## Phase 2: SQL Injection + Query Repository

- [ ] **2.1** — Add table name validation for `baseQuery` in `ExecuteQuery`
  - **File:** `backend/internal/adapters/db/grid/grid_repository.go`
  - **Change:** Add `isValidTableName()` validation function; call it before SQL interpolation at line ~127. Reject dangerous patterns (UNION, SELECT, --, ;, quotes, etc.).
  - **Verification:** `ExecuteQuery` with `baseQuery = "eamusers; DROP TABLE..."` → rejected with error. Valid `baseQuery = "eamusers"` → executes normally.
  - **Est. lines:** 20

- [ ] **2.2** — Fix `GetByID` return value
  - **File:** `backend/internal/adapters/db/queries/query_repository.go`
  - **Change:** (a) Line 78–79: return wrapped error on ErrNoRows. (b) Line 94: return `&q, nil` instead of `nil, err`.
  - **Verification:** `GetByID` with known ID → returns `(*SavedQuery, nil)`. With unknown ID → returns `(nil, error)` wrapping ErrNoRows.
  - **Est. lines:** 6

---

## Phase 3: Frontend User CRUD + Error Handling

- [ ] **3.1** — Fix safe type assertion in `getUserCode`
  - **File:** `backend/internal/adapters/api/queries/handler.go`
  - **Change:** Change `if code, ok := claims["userCode"]; ok { return code.(string) }` to `if code, ok := claims["userCode"].(string); ok { return code }` — safe assertion.
  - **Verification:** JWT with string `userCode` → works. JWT with numeric `userCode` → no panic, returns "". No `userCode` key → returns "".
  - **Est. lines:** 4

- [ ] **3.2** — Implement `onDelete` in user-list.component.ts
  - **File:** `frontend/src/app/features/users/screens/user-list/user-list.component.ts`
  - **Change:** Replace `console.log('Delete user:', selected.id)` with `this.userService.delete(selected.id).subscribe(...)` — remove user from list on success.
  - **Verification:** Select user → click delete → confirm → API called with correct ID → user removed from list.
  - **Est. lines:** 10

- [ ] **3.3** — Implement `onDetailSave` in user-list.component.ts
  - **File:** `frontend/src/app/features/users/screens/user-list/user-list.component.ts`
  - **Change:** Replace `setTimeout` mock with `userService.create()` or `userService.update()` call. Update list on success (add new or replace existing).
  - **Verification:** Create new user (no ID) → create API called → drawer closes, list updates. Update existing → update API called with correct ID.
  - **Est. lines:** 22

- [ ] **3.4** — Fix error message extraction in error.interceptor.ts
  - **File:** `frontend/src/app/core/interceptors/error.interceptor.ts`
  - **Change:** Replace `error.error?.error?.message || error.error?.message` with logic that handles both flat `{code, message}` and wrapped `{success: false, error: {code, message}}` formats.
  - **Verification:** Wrapped format error → notification shows inner message. Flat format error → notification shows direct message.
  - **Est. lines:** 12

---

## Summary

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1 | 3 | Nil-pointer bugs in auth repositories |
| Phase 2 | 2 | SQL injection prevention + query repo return value |
| Phase 3 | 4 | Frontend CRUD + error handling |
| **Total** | **9** | |

**Estimated total lines:** ~280–340 (7 files)

**Dependencies:** Phase 1 fixes are independent of each other but should complete before frontend auth testing. Phase 2 and Phase 3 fixes are independent from each other and from Phase 1.

**Implementation Order:** 1.1 → 1.2 → 1.3 → 2.2 → 2.1 → 3.1 → 3.4 → 3.2 → 3.3

**Next Step:** Ready for sdd-apply — all tasks are small, isolated, and independently verifiable.