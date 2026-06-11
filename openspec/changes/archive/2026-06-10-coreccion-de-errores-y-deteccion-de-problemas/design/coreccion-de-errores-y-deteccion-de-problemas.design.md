# Design: coreccion de errores y deteccion de problemas para llegar a un app estable

## Overview

This design document specifies concrete implementation details for each bug fix identified in the spec. Each fix includes exact file changes, verification steps, and rollback procedures.

---

## Phase 1: Critical Nil Pointer Bugs (Auth Backend)

### Fix 1.1: `FindByRefreshToken` — Return error on `pgx.ErrNoRows`

**File:** `backend/internal/adapters/db/auth/session_repository.go`

**Lines affected:** 55–67 (closure body within `FindByRefreshToken`)

**Change needed:**

```go
// At line 55-67, replace the closure body:
// FROM:
err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
    var s auth.Session
    err := tx.QueryRow(ctx, query, token).Scan(
       &s.ID, &s.UserCode, &s.RefreshToken,
        &s.ExpiresAt, &s.IPAddress, &s.UserAgent,
        &s.CreatedAt, &s.RevokedAt,
    )
    if err != nil {
        return err
    }
    session = &s
    return nil
})

// TO:
err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
    var s auth.Session
    err := tx.QueryRow(ctx, query, token).Scan(
        &s.ID, &s.UserCode, &s.RefreshToken,
        &s.ExpiresAt, &s.IPAddress, &s.UserAgent,
        &s.CreatedAt, &s.RevokedAt,
    )
    if errors.Is(err, pgx.ErrNoRows) {
        return fmt.Errorf("session not found for token: %w", err)
    }
    if err != nil {
        return err
    }
    session = &s
    return nil
})
```

**Note:** Add `"errors"` and `"fmt"` to imports if not already present.

**Verification:**
1. Call `FindByRefreshToken(ctx, "nonexistent-token")`
2. Verify return is `(nil, err)` where `errors.Is(err, pgx.ErrNoRows)` is true
3. Verify calling code (e.g., refresh token flow) handles the error gracefully without panic

**Rollback:**
```bash
git checkout HEAD -- backend/internal/adapters/db/auth/session_repository.go
```

---

### Fix 1.2: `FindByEmail` — Return error on `pgx.ErrNoRows`

**File:** `backend/internal/adapters/db/auth/user_repository.go`

**Lines affected:** 30–42 (closure body within `FindByEmail`)

**Change needed:**

```go
// At line 30-42, replace the closure body:
// FROM:
err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
    var u auth.User
    err := tx.QueryRow(ctx, query, email).Scan(
        &u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
        &u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
        &u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
    )
    if err != nil {
        return err
    }
    user = &u
    return nil
})

// TO:
err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
    var u auth.User
    err := tx.QueryRow(ctx, query, email).Scan(
        &u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
        &u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
        &u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
    )
    if errors.Is(err, pgx.ErrNoRows) {
        return fmt.Errorf("user not found for email: %w", err)
    }
    if err != nil {
        return err
    }
    user = &u
    return nil
})
```

**Note:** Add `"errors"` and `"fmt"` to imports if not already present.

**Verification:**
1. Call `FindByEmail(ctx, "nonexistent@example.com")`
2. Verify return is `(nil, err)` where `errors.Is(err, pgx.ErrNoRows)` is true
3. Verify login flow handles the error gracefully

**Rollback:**
```bash
git checkout HEAD -- backend/internal/adapters/db/auth/user_repository.go
```

---

### Fix 1.3: `FindByCode` — Return error on `pgx.ErrNoRows`

**File:** `backend/internal/adapters/db/auth/user_repository.go`

**Lines affected:** 55–67 (closure body within `FindByCode`)

**Change needed:**

```go
// At line 55-67, replace the closure body:
// FROM:
err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
    var u auth.User
    err := tx.QueryRow(ctx, query, code).Scan(
        &u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
        &u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
        &u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
    )
    if err != nil {
        return err
    }
    user = &u
    return nil
})

// TO:
err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
    var u auth.User
    err := tx.QueryRow(ctx, query, code).Scan(
        &u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
        &u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
        &u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
    )
    if errors.Is(err, pgx.ErrNoRows) {
        return fmt.Errorf("user not found for code: %w", err)
    }
    if err != nil {
        return err
    }
    user = &u
    return nil
})
```

**Verification:**
1. Call `FindByCode(ctx, "NONEXISTENT")`
2. Verify return is `(nil, err)` where `errors.Is(err, pgx.ErrNoRows)` is true
3. Verify any code-lookup flow handles the error gracefully

**Rollback:**
```bash
git checkout HEAD -- backend/internal/adapters/db/auth/user_repository.go
```

---

## Phase 2: SQL Injection + Query Repository

### Fix 2.1: `baseQuery` validation in `ExecuteQuery`

**File:** `backend/internal/adapters/db/grid/grid_repository.go`

**Lines affected:** Lines 121–261 (`ExecuteQuery` function), plus new validation function

**Change needed:**

1. **Add validation function** (add near top of file, after imports):

```go
// isValidTableName validates that baseQuery contains only safe SQL identifiers.
// It rejects SQL keywords, special characters, and common injection patterns.
func isValidTableName(name string) bool {
    if name == "" {
        return false
    }
    // Check for dangerous patterns
    dangerous := []string{
        "--", "/*", "*/", ";", "'", "\"",
        "UNION", "SELECT", "INSERT", "UPDATE", "DELETE", "DROP",
        "CREATE", "ALTER", "EXEC", "EXECUTE", "xp_", "sp_",
 }
    upper := strings.ToUpper(name)
    for _, d := range dangerous {
        if strings.Contains(upper, d) {
            return false
        }
    }
    // Allow only: alphanumeric, underscore, space, dot (schema.table), comma (JOIN)
    // Check for any other special characters
    for _, c := range name {
        if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == ' ' || c == '.' || c == ',') {
            return false
        }
    }
    return true
}
```

2. **Add validation call** at start of `ExecuteQuery` (after line 127, before building query):

```go
// Validate baseQuery before using it
if !isValidTableName(baseQuery) {
    return nil, fmt.Errorf("invalid base query: potentially dangerous SQL")
}
```

**Verification:**
1. Call `ExecuteQuery` with `baseQuery = "eamusers; DROP TABLE eamusers; --"` — verify rejection
2. Call `ExecuteQuery` with `baseQuery = "eamusers UNION SELECT * FROM eamusers"` — verify rejection
3. Call `ExecuteQuery` with `baseQuery = "eamusers"` (valid) — verify successful execution
4. Call `ExecuteQuery` with `baseQuery = "public.eamusers"` (valid with schema) — verify successful execution

**Rollback:**
```bash
git checkout HEAD -- backend/internal/adapters/db/grid/grid_repository.go
```

---

### Fix 2.2: `GetByID` return value

**File:** `backend/internal/adapters/db/queries/query_repository.go`

**Lines affected:** 78–79 (ErrNoRows handling) and 94 (return statement)

**Change needed:**

1. **Fix ErrNoRows handling** (lines 78–79):

```go
// FROM:
if errors.Is(err, pgx.ErrNoRows) {
    return nil
}

// TO:
if errors.Is(err, pgx.ErrNoRows) {
    return nil, fmt.Errorf("query not found: %w", err)
}
```

2. **Fix return on success** (line 94):

```go
// FROM:
return nil, err

// TO:
return&q, nil
```

**Note:** Add `"fmt"` to imports if not already present.

**Verification:**
1. Create a known query in the database
2. Call `GetByID(ctx, knownQueryID)`
3. Verify return is `(*queries.SavedQuery, nil)` with correct data
4. Call `GetByID(ctx, unknownID)`
5. Verify return is `(nil, error)` where error wraps `pgx.ErrNoRows`

**Rollback:**
```bash
git checkout HEAD -- backend/internal/adapters/db/queries/query_repository.go
```

---

## Phase 3: Frontend User CRUD + Error Handling

### Fix 3.1: Safe type assertion in `getUserCode`

**File:** `backend/internal/adapters/api/queries/handler.go`

**Lines affected:** 134–144 (`getUserCode` function)

**Change needed:**

```go
// FROM:
func getUserCode(c *fiber.Ctx) string {
    // Check for user claims set by auth middleware
    if user := c.Locals("user"); user != nil {
        if claims, ok := user.(map[string]interface{}); ok {
            if code, ok := claims["userCode"]; ok {
                return code.(string)
            }
        }
    }
    return ""
}

// TO:
func getUserCode(c *fiber.Ctx) string {
    // Check for user claims set by auth middleware
    if user := c.Locals("user"); user != nil {
        if claims, ok := user.(map[string]interface{}); ok {
            if code, ok := claims["userCode"].(string); ok {
                return code
            }
        }
    }
    return ""
}
```

**Verification:**
1. Set up a test with JWT claims where `userCode` is a string — verify it works
2. Set up a test with JWT claims where `userCode` is a number (e.g., `123`) — verify no panic, returns empty string
3. Set up a test with no `userCode` in claims — verify returns empty string

**Rollback:**
```bash
git checkout HEAD -- backend/internal/adapters/api/queries/handler.go
```

---

### Fix 3.2: `onDelete` — Implement actual delete API call

**File:** `frontend/src/app/features/users/screens/user-list/user-list.component.ts`

**Lines affected:** 518–526 (`onDelete` method)

**Change needed:**

```typescript
// FROM:
onDelete() {
    const selected = this.selected();
    const message = this.t['messages.delete_confirm'] || '¿Está seguro que desea eliminar?';
    if (selected && confirm(`${message} ${selected.name}?`)) {
      // TODO: call delete API
      console.log('Delete user:', selected.id);
      this.selected.set(null);
    }
  }

// TO:
onDelete() {
    const selected = this.selected();
    const message = this.t['messages.delete_confirm'] || '¿Está seguro que desea eliminar?';
    if (selected && confirm(`${message} ${selected.name}?`)) {
      this.userService.delete(selected.id).subscribe({
        next: () => {
          this.userService.users.update(users => users.filter(u => u.id !== selected.id));
          this.selected.set(null);
        },
        error: (err) => console.error('Error deleting user:', err)
      });
    }
  }
```

**Verification:**
1. Select a user in the UI
2. Click delete button and confirm
3. Verify the delete API is called with the correct user ID
4. Verify the user is removed from the list
5. Verify an error is shown if the API call fails

**Rollback:**
```bash
git checkout HEAD -- frontend/src/app/features/users/screens/user-list/user-list.component.ts
```

---

### Fix 3.3: `onDetailSave` — Implement actual save API call

**File:** `frontend/src/app/features/users/screens/user-list/user-list.component.ts`

**Lines affected:** 631–647 (`onDetailSave` method)

**Change needed:**

```typescript
// FROM:
onDetailSave(entity: unknown) {
    this.detailSaving.set(true);
    // TODO: Llamar API para guardar
    console.log('Saving user:', entity);
    setTimeout(() => {
      this.detailSaving.set(false);
      this.showDetail.set(false);
      this.hasUnsavedChanges.set(false);
      // Actualizar la lista con el entity guardado
      if (entity && typeof entity === 'object') {
        const savedUser = entity as User;
        this.userService.users.update((users) =>
          users.map((u) => (u.id === savedUser.id ? savedUser : u))
        );
      }
    }, 1000);
  }

// TO:
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

**Verification:**
1. Create a new user (no ID) and save — verify create API is called
2. Update an existing user and save — verify update API is called with correct ID
3. Verify success: drawer closes, list updates with new/updated user
4. Verify error: drawer shows loading but on error, stays open with error message

**Rollback:**
```bash
git checkout HEAD -- frontend/src/app/features/users/screens/user-list/user-list.component.ts
```

---

### Fix 3.4: Error interceptor message extraction

**File:** `frontend/src/app/core/interceptors/error.interceptor.ts`

**Lines affected:** 14–17 (message extraction logic)

**Change needed:**

```typescript
// FROM (lines 14-17):
// Extract server message supporting both formats:
// - flat:  {code: "...", message: "..."}        (AppError)
// - wrapped: {success: false, error: {code: "...", message: "..."}} (customErrorHandler)
const serverMsg = error.error?.error?.message || error.error?.message || '';

// TO (lines 14-24):
// Extract server message supporting both formats:
// - flat:  {code: "...", message: "..."}        (AppError)
// - wrapped: {success: false, error: {code: "...", message: "..."}} (customErrorHandler)
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

**Verification:**
1. Trigger a 500 error with wrapped format `{success: false, error: {code: "INTERNAL", message: "Internal server error"}}`
2. Verify the notification shows "Internal server error"
3. Trigger a 400 error with flat format `{code: "BAD_REQUEST", message: "Invalid input"}`
4. Verify the notification shows "Invalid input"

**Rollback:**
```bash
git checkout HEAD -- frontend/src/app/core/interceptors/error.interceptor.ts
```

---

## Implementation Order

### Recommended Order (by dependency and risk)

| Order | Phase | Fix | Reason |
|-------|-------|-----|--------|
| 1 | Phase 1 | Fix 1.1 (`session_repository.go`) | Auth flow is foundational; other fixes may depend on working auth |
| 2 | Phase 1 | Fix 1.2 (`user_repository.go` — `FindByEmail`) | Login flow depends on this |
| 3 | Phase 1 | Fix 1.3 (`user_repository.go` — `FindByCode`) | Lower risk; same pattern as 1.2 |
| 4 | Phase 2 | Fix 2.2 (`query_repository.go`) | Small, isolated change; return value fix |
| 5 | Phase 2 | Fix 2.1 (`grid_repository.go`) | SQL injection prevention; validate after other repo fixes |
| 6 | Phase 3 | Fix 3.1 (`handler.go`) | Backend type assertion fix; small and isolated |
| 7 | Phase 3 | Fix 3.4 (`error.interceptor.ts`) | Frontend error handling; independent |
| 8 | Phase 3 | Fix 3.2 (`user-list.component.ts` — `onDelete`) | Frontend CRUD; depends on UserService existing |
| 9 | Phase 3 | Fix 3.3 (`user-list.component.ts` — `onDetailSave`) | Frontend CRUD; depends on UserService and Fix 3.2 |

### Dependencies Summary

- **Phase 1 fixes are independent of each other** but should complete before Phase 3 frontend testing
- **Phase 2 fixes are independent** but Fix 2.1 (SQL injection) should be tested with actual grid data
- **Phase 3 fixes3.2 and 3.3 are independent** of each other but both depend on `UserService` existing and working
- **Fix3.1 (handler.go) is independent** and can be tested in isolation

---

## Risk Mitigation

### Phase 1: Nil Pointer Bugs

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Breaking auth flow with nil-pointer fixes | Medium | Test login/logout flow after each fix. Verify no panics in server logs. |
| Return of new error type breaks caller code | Low | Errors wrap `pgx.ErrNoRows`; callers using `errors.Is` will work unchanged. |

**Verification:** Run auth integration tests or manually test login/logout flow.

### Phase 2: SQL Injection + Query Repo

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| SQL injection fix breaking valid grid queries | Medium | Test with actual grid data. Whitelist approach is conservative — may reject edge-case but valid queries if they contain unusual characters. |
| `GetByID` return fix breaking caller | Low | Verify callers of `GetByID` handle non-nil return correctly. |

**Verification:** 
- For Fix 2.1: Test grid queries with valid `baseQuery` values; verify rejection of malicious input.
- For Fix 2.2: Test `GetByID` with known ID and unknown ID.

### Phase 3: Frontend User CRUD + Error Handling

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| `onDelete`/`onDetailSave` not wired to real API | Low | Verify `UserService.delete`, `UserService.create`, `UserService.update` exist and work. |
| Error interceptor change breaking error display | Low | Test both wrapped and flat error formats after change. |

**Verification:**
- Fix 3.2& 3.3: Test full CRUD cycle (create, read, update, delete) in browser.
- Fix 3.4: Trigger various error responses and verify notification text.

### General Rollback Strategy

If any fix causes issues:
1. Identify the failing fix by reverting files one at a time
2. Each fix is isolated — rollback is file-level, no database migrations needed
3. After rollback, test the specific flow that was broken

---

## File Summary

| Phase | File | Lines | Type |
|-------|------|-------|------|
| 1 | `backend/internal/adapters/db/auth/session_repository.go` | 55–67 | Go |
| 1 | `backend/internal/adapters/db/auth/user_repository.go` | 30–42, 55–67 | Go |
| 2 | `backend/internal/adapters/db/grid/grid_repository.go` | +15 (validation func), 127 (call site) | Go |
| 2 | `backend/internal/adapters/db/queries/query_repository.go` | 78–79, 94 | Go |
| 3 | `backend/internal/adapters/api/queries/handler.go` | 134–144 | Go |
| 3 | `frontend/src/app/features/users/screens/user-list/user-list.component.ts` | 518–526, 631–647 | TypeScript |
| 3 | `frontend/src/app/core/interceptors/error.interceptor.ts` | 14–24 | TypeScript |

**Total files:** 7 (4 backend, 3 frontend)
