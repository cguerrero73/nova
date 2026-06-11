# Spec: coreccion de errores y deteccion de problemas para llegar a un app estable

## Overview

This spec details concrete, testable specifications for each bug fix identified in the proposal. Bugs are organized by phase in dependency order.

---

## Phase 1: Critical Nil Pointer Bugs (Auth Backend)

### Bug 1.1: `FindByRefreshToken` returns nil pointer on ErrNoRows

**What the bug is:** When `FindByRefreshToken` encounters `pgx.ErrNoRows`, it returns `nil, nil` instead of `nil, err`, causing nil pointer panics downstream.

**Where it is:** `backend/internal/adapters/db/auth/session_repository.go:69`

**Expected behavior:** When no session matches the refresh token, return `nil` with an error that wraps `pgx.ErrNoRows`.

**Actual behavior:** The closure returns `nil` for session when `pgx.ErrNoRows` occurs (line 62-63), and the outer function returns `session, err` where `session` is `nil` and `err` is `nil`.

**Fix description:** After the closure's `tx.QueryRow().Scan()` call, check if the error is `pgx.ErrNoRows` using `errors.Is(err, pgx.ErrNoRows)`. If true, wrap it with `fmt.Errorf("session not found: %w", err)` or use a custom sentinel error before returning. The closure should NOT swallow the error.

```go
// In closure at line 55-67:
// Change:
if err != nil {
    return err
}
session = &s
return nil

// To:
if errors.Is(err, pgx.ErrNoRows) {
    return fmt.Errorf("session not found for token: %w", err)
}
if err != nil {
    return err
}
session = &s
return nil
```

**Test scenario:** 
1. Call `FindByRefreshToken(ctx, "nonexistent-token")`
2. Verify the return is `nil, err` where `errors.Is(err, pgx.ErrNoRows)` is true
3. Verify calling code handles the error gracefully (no panic)

---

### Bug 1.2: `FindByEmail` returns nil pointer on ErrNoRows

**What the bug is:** When `FindByEmail` encounters `pgx.ErrNoRows`, it returns `nil, nil` instead of `nil, err`, causing nil pointer panics downstream.

**Where it is:** `backend/internal/adapters/db/auth/user_repository.go:44`

**Expected behavior:** When no user matches the email, return `nil` with an error that wraps `pgx.ErrNoRows`.

**Actual behavior:** Same pattern as Bug 1.1 — the closure returns `nil` for user when `pgx.ErrNoRows` occurs (line 37-39), and the outer function returns `user, err` where both are `nil`.

**Fix description:** Apply same pattern as Bug 1.1 — check for `pgx.ErrNoRows` after `tx.QueryRow().Scan()` and return a wrapped error instead of nil.

```go
// In closure at line 30-42:
// Change to check errors.Is(err, pgx.ErrNoRows) before assigning user = &u
```

**Test scenario:**
1. Call `FindByEmail(ctx, "nonexistent@example.com")`
2. Verify the return is `nil, err` where `errors.Is(err, pgx.ErrNoRows)` is true
3. Verify calling code handles the error gracefully (no panic)

---

### Bug 1.3: `FindByCode` returns nil pointer on ErrNoRows

**What the bug is:** When `FindByCode` encounters `pgx.ErrNoRows`, it returns `nil, nil` instead of `nil, err`, causing nil pointer panics downstream.

**Where it is:** `backend/internal/adapters/db/auth/user_repository.go:69`

**Expected behavior:** When no user matches the code, return `nil` with an error that wraps `pgx.ErrNoRows`.

**Actual behavior:** Same pattern as Bug 1.2 — closure at line 62-64 returns nil on error before assigning user.

**Fix description:** Apply same pattern as Bug 1.2.

**Test scenario:**
1. Call `FindByCode(ctx, "NONEXISTENT")`
2. Verify the return is `nil, err` where `errors.Is(err, pgx.ErrNoRows)` is true
3. Verify calling code handles the error gracefully (no panic)

---

## Phase 2: SQL Injection + Query Repository

### Bug 2.1: SQL Injection risk in `baseQuery` interpolation

**What the bug is:** The `baseQuery` string is interpolated directly into SQL strings (`countQuery`, `fullQuery`) without validation, allowing potential SQL injection if `baseQuery` contains malicious content.

**Where it is:** `backend/internal/adapters/db/grid/grid_repository.go:187` and `line 215`

**Expected behavior:** `baseQuery` should be validated against a whitelist of allowed table/table-alias names before interpolation, OR the query should be restructured to use only safe identifiers.

**Actual behavior:** 
- Line 187: `countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", baseQuery, whereClause)`
- Line 215: `fullQuery := fmt.Sprintf("SELECT %s FROM %s%s%s%s", selectClause, baseQuery, whereClause, orderClause, limitOffset)`

**Fix description:** Add validation that `baseQuery` contains only safe SQL identifiers (alphanumeric, underscores, spaces, and common SQL join syntax). Reject if it contains:
- SQL keywords (UNION, SELECT, INSERT, UPDATE, DELETE, DROP, --, /*, */, ;, etc.)
- Special characters beyond alphanumeric, underscore, space, dot (for schema.table), comma

```go
// Add before line 187
func isValidTableName(name string) bool {
    // Allow: alphanumeric, underscore, space, dot (schema.table), comma (JOIN)
    // Reject: quotes, semicolons, dashes, parens, comments, keywords
    dangerous := []string{"--", "/*", "*/", ";", "'", "\"", "UNION", "SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER"}
    for _, d := range dangerous {
        if strings.Contains(strings.ToUpper(name), d) {
            return false
        }
    }
    return true
}

// Then before using baseQuery:
if !isValidTableName(baseQuery) {
    return nil, fmt.Errorf("invalid base query: potentially dangerous SQL")
}
```

**Test scenario:**
1. Attempt `ExecuteQuery` with `baseQuery = "eamusers; DROP TABLE eamusers; --"`
2. Verify the query is rejected before execution with an error
3. Attempt `ExecuteQuery` with `baseQuery = "eamusers UNION SELECT * FROM eamusers"`
4. Verify the query is rejected before execution with an error
5. Attempt `ExecuteQuery` with `baseQuery = "eamusers"` (valid)
6. Verify the query executes successfully

---

### Bug 2.2: `GetByID` returns nil instead of query object

**What the bug is:** `GetByID` returns `nil, err` on success instead of `&q, nil`, causing callers to receive a nil query object when one exists.

**Where it is:** `backend/internal/adapters/db/queries/query_repository.go:94`

**Expected behavior:** When a query is found by ID, return `&q, nil` (the populated query struct).

**Actual behavior:** Line 94 always returns `nil, err` — even when no error occurs, the query struct is never returned to the caller.

**Fix description:** Change line 94 from `return nil, err` to `return &q, nil`. Also ensure the `pgx.ErrNoRows` case at line 78-79 returns `nil` with an appropriate error (not just `nil`).

```go
// Line 78-79: change
if errors.Is(err, pgx.ErrNoRows) {
    return nil
}
// To:
if errors.Is(err, pgx.ErrNoRows) {
    return nil, fmt.Errorf("query not found: %w", err)
}

// Line 94: change
return nil, err
// To:
return &q, nil
```

**Test scenario:**
1. Create a known query in the database
2. Call `GetByID(ctx, knownQueryID)`
3. Verify the return is `(*queries.SavedQuery, nil)` with the correct data
4. Call `GetByID(ctx, unknownID)`
5. Verify the return is `(nil, error)` where error wraps `pgx.ErrNoRows`

---

## Phase 3: Frontend User CRUD + Error Handling

### Bug 3.1: Type assertion panic in `getUserCode`

**What the bug is:** The type assertion `code.(string)` at `handler.go:139` will panic if `code` is not a string (e.g., if it's an int or float64 from JSON unmarshaling).

**Where it is:** `backend/internal/adapters/api/queries/handler.go:139`

**Expected behavior:** The type assertion should be safe with a second return value indicating success, or the value should be converted properly.

**Actual behavior:** Direct type assertion without safety check: `return code.(string)` — if `userCode` is stored as a number in the JWT claims, this panics.

**Fix description:** Use the safe type assertion form with ok check:

```go
// Line 138-139: change
if code, ok := claims["userCode"]; ok {
    return code.(string)
// To:
if code, ok := claims["userCode"].(string); ok {
    return code
```

**Test scenario:**
1. Set up a test with JWT claims where `userCode` is a string — verify it works
2. Set up a test with JWT claims where `userCode` is a number (e.g., `123`) — verify no panic, returns empty string
3. Set up a test with no `userCode` in claims — verify returns empty string

---

### Bug 3.2: `onDelete` stub — delete not implemented

**What the bug is:** `onDelete` in `user-list.component.ts` logs to console instead of calling the delete API.

**Where it is:** `frontend/src/app/features/users/screens/user-list/user-list.component.ts:518-526`

**Expected behavior:** When user confirms deletion, call the delete API and remove the user from the list.

**Actual behavior:** Line 522-523 contains `// TODO: call delete API` and `console.log('Delete user:', selected.id)`.

**Fix description:** Inject `UserService` and call the delete method:

```typescript
// In onDelete(), change lines 521-525:
if (selected && confirm(`${message} ${selected.name}?`)) {
  // TODO: call delete API
  console.log('Delete user:', selected.id);
  this.selected.set(null);
}

// To:
if (selected && confirm(`${message} ${selected.name}?`)) {
  this.userService.delete(selected.id).subscribe({
    next: () => {
      this.userService.users.update(users => users.filter(u => u.id !== selected.id));
      this.selected.set(null);
    },
    error: (err) => console.error('Error deleting user:', err)
  });
}
```

**Test scenario:**
1. Select a user in the UI
2. Click delete button and confirm
3. Verify the delete API is called with the correct user ID
4. Verify the user is removed from the list
5. Verify an error is shown if the API call fails

---

### Bug 3.3: `onDetailSave` stub — save not implemented

**What the bug is:** `onDetailSave` in `user-list.component.ts` uses a setTimeout mock instead of calling the save API.

**Where it is:** `frontend/src/app/features/users/screens/user-list/user-list.component.ts:631-647`

**Expected behavior:** When saving, call the appropriate create or update API based on whether the entity has an ID.

**Actual behavior:** Lines 633-646 contain `// TODO: Llamar API para guardar` and a setTimeout mock that just closes the drawer.

**Fix description:** Implement actual save logic:

```typescript
// In onDetailSave(), change lines 631-647:
onDetailSave(entity: unknown) {
    this.detailSaving.set(true);
    // TODO: Llamar API para guardar
    console.log('Saving user:', entity);
    setTimeout(() => {
      this.detailSaving.set(false);
      this.showDetail.set(false);
      this.hasUnsavedChanges.set(false);
      if (entity && typeof entity === 'object') {
        const savedUser = entity as User;
        this.userService.users.update((users) =>
          users.map((u) => (u.id === savedUser.id ? savedUser : u))
        );
      }
    }, 1000);
}

// To:
onDetailSave(entity: unknown) {
    this.detailSaving.set(true);
    const user = entity as User;
    
    const operation = user.id 
      ? this.userService.update(user.id, user)
      : this.userService.create(user);
    
    operation.subscribe({
      next: (savedUser) => {
        this.detailSaving.set(false);
        this.showDetail.set(false);
        this.hasUnsavedChanges.set(false);
        if (user.id) {
          // Update existing
          this.userService.users.update((users) =>
            users.map((u) => (u.id === savedUser.id ? savedUser : u))
          );
        } else {
          // Add new
          this.userService.users.update((users) => [...users, savedUser]);
        }
      },
      error: (err) => {
        this.detailSaving.set(false);
        console.error('Error saving user:', err);
      }
    });
}
```

**Test scenario:**
1. Create a new user (no ID) and save — verify create API is called
2. Update an existing user and save — verify update API is called with correct ID
3. Verify success: drawer closes, list updates with new/updated user
4. Verify error: drawer shows loading but on error, stays open with error message

---

### Bug 3.4: Error interceptor response structure extraction

**What the bug is:** The error interceptor at line 17 tries `error.error?.error?.message` first, but many backend errors use `{success: false, error: {...}}` structure where the inner `error` is the actual error object.

**Where it is:** `frontend/src/app/core/interceptors/error.interceptor.ts:17`

**Expected behavior:** Extract the error message from both formats:
- `{code: "...", message: "..."}` (flat)
- `{success: false, error: {code: "...", message: "..."}}` (wrapped)

**Actual behavior:** The check `error.error?.error?.message` works for wrapped format but `error.error?.message` may not correctly prioritize. When backend returns `{success: false, error: {code: "INTERNAL", message: "..."}}`, the current extraction logic may fail.

**Fix description:** Refine the extraction logic:

```typescript
// Line 14-17: change
// Extract server message supporting both formats:
// - flat:  {code: "...", message: "..."}        (AppError)
// - wrapped: {success: false, error: {code: "...", message: "..."}} (customErrorHandler)
const serverMsg = error.error?.error?.message || error.error?.message || '';

// To:
const serverData = error.error;
let serverMsg = '';

// Handle wrapped format: {success: false, error: {code: "...", message: "..."}}
if (serverData && typeof serverData === 'object' && 'error' in serverData) {
  const inner = (serverData as any).error;
  serverMsg = inner?.message || '';
}
// Handle flat format: {code: "...", message: "..."}
if (!serverMsg && serverData && typeof serverData === 'object') {
  serverMsg = (serverData as any).message || '';
}
```

**Test scenario:**
1. Trigger a 500 error with wrapped format `{success: false, error: {code: "INTERNAL", message: "Internal server error"}}`
2. Verify the notification shows "Internal server error"
3. Trigger a 400 error with flat format `{code: "BAD_REQUEST", message: "Invalid input"}`
4. Verify the notification shows "Invalid input"

---

## Phase 4: Skipped (Test Scaffolding)

*No specs for this phase — marked as skipped in proposal.*

---

## Summary

| Phase | Bug ID | File | Line | Severity | Fix Required |
|-------|--------|------|------|----------|--------------|
| 1 | 1.1 | session_repository.go | 69 | Critical | Check `errors.Is(err, pgx.ErrNoRows)` before returning nil |
| 1 | 1.2 | user_repository.go | 44 | Critical | Same pattern as 1.1 for `FindByEmail` |
| 1 | 1.3 | user_repository.go | 69 | Critical | Same pattern as 1.1 for `FindByCode` |
| 2 | 2.1 | grid_repository.go | 187, 215 | High | Validate `baseQuery` before SQL interpolation |
| 2 | 2.2 | query_repository.go | 94 | High | Return `&q, nil` instead of `nil, err` on success |
| 3 | 3.1 | handler.go | 139 | Medium | Safe type assertion for `userCode` |
| 3 | 3.2 | user-list.component.ts | 518-526 | Medium | Implement actual delete API call |
| 3 | 3.3 | user-list.component.ts | 631-647 | Medium | Implement actual save API call |
| 3 | 3.4 | error.interceptor.ts | 17 | Medium | Fix error message extraction logic |