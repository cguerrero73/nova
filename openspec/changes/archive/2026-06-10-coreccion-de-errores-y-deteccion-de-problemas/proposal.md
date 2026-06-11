# Proposal: coreccion de errores y deteccion de problemas para llegar a un app estable

## Intent

Stabilize the Nova application by fixing critical and high-severity bugs, improving error handling consistency, and establishing baseline test scaffolding. The goal is a production-ready app with zero nil-pointer panics in auth flows, prevented SQL injection in grid queries, and functional user management.

## Scope

### In Scope
- Fix 4 critical nil-pointer bugs in session/user repositories and query repository
- Fix SQL injection risk in grid_repository.go
- Fix user code extraction in tenant context
- Fix tenant auth token mismatch
- Complete user CRUD operations on frontend
- Add error handling consistency across backend and frontend
- Add pagination validation
- Create test scaffolding for backend and frontend

### Out of Scope
- Low-severity issues (identical env files, debug log leak)
- New features or major architectural refactors
- Comprehensive test coverage (scaffolding only)

## Capabilities

### New Capabilities
- `error-handling-consistency`: Unified error handling pattern across backend repositories and frontend services
- `test-scaffolding`: Basic test infrastructure for backend (Go) and frontend (if applicable)

### Modified Capabilities
- `user-auth`: Nil-pointer fixes and error wrapping in session/user repositories
- `grid-queries`: SQL injection prevention via parameterized queries
- `user-management`: Frontend user CRUD completion (delete/save/actions)

## Approach

Fix bugs in place without major refactors. Each fix is small, focused, and verified manually. Prioritize by severity and dependencies: auth fixes first (Phase 1), then query repository (Phase 2), then frontend user management (Phase 3), finally test scaffolding (Phase 4).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/auth/session.go` | Modified | Add `errors.Is` checks for `pgx.ErrNoRows` |
| `backend/auth/user.go` | Modified | Add `errors.Is` checks for `pgx.ErrNoRows` |
| `backend/queries/repository.go:94` | Modified | Return `q, err` instead of `nil, err` |
| `backend/queries/handler.go:136` | Modified | Fix type assertion |
| `backend/grids/repository.go` | Modified | Sanitize `baseQuery` to prevent SQL injection |
| `backend/common/` | Modified | Standardize error wrapping |
| `frontend/user-list.component.ts` | Modified | Implement real delete/save/actions |
| `frontend/error interceptor` | Modified | Consistent error handling structure |
| `frontend/auth service` | Modified | Fix tenant auth token mismatch |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Breaking auth flow with nil-pointer fixes | Medium | Test each auth fix with login/logout flow |
| SQL injection fix breaking grid queries | Medium | Verify with actual grid data after fix |
| Frontend changes affecting user CRUD | Low | Test full CRUD cycle after fixes |

## Rollback Plan

Each fix is isolated. To rollback:
1. Revert the specific file to the previous commit
2. No database migrations needed (data integrity preserved)
3. Test auth flow and grid queries after rollback

## Dependencies

- None (all fixes are self-contained)
- Auth nil-pointer fixes should complete before frontend user CRUD testing

## Success Criteria

- [ ] Zero nil-pointer panics in auth flow (login/logout)
- [ ] SQL injection prevented in grid queries ( parameterized queries)
- [ ] User can log in, view user list, delete users, save users
- [ ] Grid configuration loads correctly
- [ ] Basic test scaffolding exists for backend
- [ ] All error responses are consistent (JSON with `error` field)

## Estimation

| Component | Lines Changed | Files |
|-----------|---------------|-------|
| Backend | ~150-200 | 5 |
| Frontend | ~100-150 | 3 |
| **Total** | **~250-350** | **8** |

## Open Questions

1. Should we add test scaffolding even if tests won't be comprehensive immediately?
2. Any bugs from the 18 found that should be deprioritized?