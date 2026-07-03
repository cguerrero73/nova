# Archive Report: fix-tenant-context-propagation

**Change**: `fix-tenant-context-propagation`
**Archived on**: 2026-06-19
**Archived to**: `openspec/changes/archive/2026-06-19-fix-tenant-context-propagation/`
**SDD cycle**: complete — explore → propose → design → tasks → apply → verify → archive
**Verdict inherited**: PASS WITH WARNINGS (per `verify-report.md`, Engram id=124)

---

## 1. Task Completion Gate

Inspected `openspec/changes/fix-tenant-context-propagation/tasks.md` before archive. All 12 tasks are checked `[x]`:

| # | Title | Status |
|---|-------|--------|
| 1 | `fix(tenant): use c.UserContext() in queries handler` | [x] |
| 2 | `fix(tenant): use c.UserContext() in users handler` | [x] |
| 3 | `fix(tenant): use c.UserContext() in auth handler` | [x] |
| 4 | `fix(tenant): use c.UserContext() in stores handler` | [x] |
| 5 | `fix(tenant): use c.UserContext() in events handler` | [x] |
| 6 | `fix(tenant): use c.UserContext() in objects handler` | [x] |
| 7 | `fix(tenant): use c.UserContext() in parts handler` | [x] |
| 8 | `fix(tenant): use c.UserContext() in stocks handler` | [x] |
| 9 | `fix(tenant): use c.UserContext() in syscodes handler` | [x] |
| 10 | `fix(tenant): use c.UserContext() in organizations handler` | [x] |
| 11 | `fix(tenant): use c.UserContext() in structure handler` | [x] |
| 12 | `test(tenant): regression test for RunInTenantTx context propagation` | [x] |

Gate passed. No reconciliation of stale checkboxes was required.

## 2. Spec Sync Decision

The proposal explicitly stated:

> **Modified Capabilities**: None — `openspec/specs/` does not exist in this project (no existing capability covers tenant propagation). Adding a spec for tenant propagation is out of scope for this bug-fix change and should be tracked separately if desired.

`openspec/specs/` confirmed absent at archive time (`test -d openspec/specs` → `DOES NOT EXIST`). The change ships **no delta spec**. There is therefore nothing to merge into a main spec. **This step was skipped by design** (the bug fix is mechanical, the regression test pins behavior, and the proposal's §Out-of-scope list explicitly excludes introducing a new capability spec for tenant propagation).

If a future change wants to formalize tenant-propagation as a tracked capability, it should be a separate SDD cycle that starts from scratch under `openspec/specs/`.

## 3. Move to Archive

```
openspec/changes/fix-tenant-context-propagation/
  → openspec/changes/archive/2026-06-19-fix-tenant-context-propagation/
```

`openspec/changes/archive/` already existed (5 prior archives present from earlier SDD cycles). No `mkdir -p` was strictly needed but was applied defensively.

## 4. Verification

- [x] Source folder removed from `openspec/changes/`.
- [x] Archive folder present at `openspec/changes/archive/2026-06-19-fix-tenant-context-propagation/`.
- [x] All 5 artifacts present: `explore.md`, `proposal.md`, `design.md`, `tasks.md`, `verify-report.md`.
- [x] Archived `tasks.md` has all 12 implementation tasks marked `[x]` (no stale checkboxes).
- [x] `openspec/specs/` remains absent (no delta to merge was promised).
- [x] The 12 commits remain on `main` (see §6).

## 5. Archive Contents

```
openspec/changes/archive/2026-06-19-fix-tenant-context-propagation/
├── explore.md         (20.3 KB, 216 lines)
├── proposal.md        ( 5.1 KB,  68 lines)
├── design.md          ( 6.7 KB, 119 lines)
├── tasks.md           (13.5 KB, 287 lines) — 12/12 tasks [x]
└── verify-report.md   (12.7 KB, 175 lines) — PASS WITH WARNINGS
```

## 6. Commit Lineage (12 commits, on `main`, not pushed by archive)

| # | SHA | Title |
|---|------|-------|
| 1 | `9791fbe` | fix(tenant): use c.UserContext() in queries handler |
| 2 | `c1105db` | fix(tenant): use c.UserContext() in users handler |
| 3 | `2af567c` | fix(tenant): use c.UserContext() in auth handler |
| 4 | `90eb654` | fix(tenant): use c.UserContext() in stores handler |
| 5 | `7b895a3` | fix(tenant): use c.UserContext() in events handler |
| 6 | `644f2d6` | fix(tenant): use c.UserContext() in objects handler |
| 7 | `02a0f31` | fix(tenant): use c.UserContext() in parts handler |
| 8 | `cb3f96f` | fix(tenant): use c.UserContext() in stocks handler |
| 9 | `b05f0b7` | fix(tenant): use c.UserContext() in syscodes handler |
| 10 | `af7a2d0` | fix(tenant): use c.UserContext() in organizations handler |
| 11 | `2a602b7` | fix(tenant): use c.UserContext() in structure handler |
| 12 | `aadd7f3` | test(tenant): regression test for RunInTenantTx context propagation |

`git log --oneline -12` at archive time confirms `aadd7f3` is the HEAD of the SDD series. Three later commits (`d734b4d`, `fb838e2`, `b41204a` — `.gitignore` + `pnpm v11 supply-chain check`) are unrelated to this change and are not part of the SDD lineage. **The 12 SDD commits are on `main` and have NOT been pushed by archive** (orchestrator handles push policy).

## 7. Engram Observation Lineage

| Artifact | Engram ID | Title |
|----------|-----------|-------|
| explore | 118 | `sdd/fix-tenant-context-propagation/explore` |
| proposal | 120 | `sdd/fix-tenant-context-propagation/proposal` |
| design | 121 | `sdd/fix-tenant-context-propagation/design` |
| tasks | 122 | `sdd/fix-tenant-context-propagation/tasks` |
| apply-progress | 123 | `sdd/fix-tenant-context-propagation/apply-progress` |
| verify-report | 124 | `sdd/fix-tenant-context-propagation/verify-report` |
| **archive-report** | **(this save)** | `sdd/fix-tenant-context-propagation/archive-report` |

## 8. Documented Deviations (carry-forward)

Carried forward verbatim from `verify-report.md §Deviations Recap` and `apply-progress` so future readers of this archive understand what shipped vs. what was originally designed.

### DEV-1 — `QueryEngine` interface invented in commit 12 (not pre-existing)

- The apply-progress observation (Engram id=123) described `QueryEngine` as "already declared" — that wording is inaccurate. `git show 1dc091a -- backend/internal/infrastructure/db/context.go` confirms the file at the prior tip had only `func RunInTenantTx(ctx, pool *pgxpool.Pool, …)`. The interface was added in the same commit (`aadd7f3`) that introduced the regression test.
- Impact: zero runtime behavior change. `*pgxpool.Pool` satisfies `QueryEngine` exactly (Begin/Exec/Query/QueryRow). All 11 mechanical handler commits are unaffected.
- Recommendation: treat the interface as **test-only scaffolding**. If a future contributor "generalizes" it, they should be aware it was introduced solely to make `pgxmock.NewPool()` driveable.

### DEV-2 — Regression test dropped the fasthttp fallback subtest

- Design specified three subtests; the apply shipped two (`tenant_wrapped_ctx_emits_set_search_path`, `no_tenant_uses_public_schema`). The third (`tenant_in_fasthttp_user_value`) was dropped because constructing a minimally viable `*fasthttp.RequestCtx` requires a `Server` + goroutine lifecycle that adds substantial noise for code that becomes inert in practice after the 11 mechanical fixes.
- Impact: the fasthttp fallback in `extractTenantCode` (context.go lines 183–188) remains untested. If a future contributor deletes it, no test will fail. Documented as an **accepted scope reduction** vs. the design.

### DEV-3 — Transitive dependency upgrades

- `pgxmock/v3 v3.4.0` transitively upgraded `pgx/v5` from `5.4.3` → `5.5.5` and `testify` from `1.8.1` → `1.9.0`. Both are minor/patch-range; full-tree `go build` / `go vet` / `go test` all pass.

## 9. Documented Warnings (from verify-report)

- **W1 (DEV-1)**: `QueryEngine` provenance — see §8 above.
- **W2 (DEV-2)**: Untested fasthttp fallback — see §8 above.
- **W3 (pre-existing)**: `GET /api/v1/queries/:id` is not registered in `cmd/server/main.go`. Already flagged as out-of-scope in `proposal.md §(b)`. Auth middleware overwrites `c.Locals("tenant", "")` from JWT claims — pre-existing, not caused by this change.

None of these block archive (no CRITICALs; warnings documented and accepted).

## 10. SDD Cycle Status

**COMPLETE.** The change has been:

1. Explored (root cause: handlers pass `c.Context()` instead of `c.UserContext()`, dropping the wrapped Go context that carries the tenant).
2. Proposed (mechanical 70-site replacement + regression test; explicitly opted out of `openspec/specs/` delta).
3. Designed (one PR, one commit per handler, pgxmock-based test).
4. Tasked (12 work-unit commits, independently revertible).
5. Applied (12 commits landed on `main`; `go build` / `go vet` / `go test` / `go test -race` all clean).
6. Verified (build, vet, full test suite, race detector, static guard, live smoke against tenant `acme` — all pass; HTTP 200 with `[TX] [acme] BEGIN + SET search_path TO tenant_acme, public` confirmed).
7. **Archived (this report)** — change folder moved to dated archive; no delta specs to merge; cycle complete.

The next change is unblocked.

## 11. Relevant Files (final state)

### Modified (11 handler files, 70 sites total)

- `backend/internal/adapters/api/queries/handler.go` — 6 sites
- `backend/internal/adapters/api/users/handler.go` — 5 sites
- `backend/internal/adapters/api/auth/handler.go` — 5 sites
- `backend/internal/adapters/api/stores/handler.go` — 9 sites
- `backend/internal/adapters/api/events/handler.go` — 7 sites
- `backend/internal/adapters/api/objects/handler.go` — 6 sites
- `backend/internal/adapters/api/parts/handler.go` — 5 sites
- `backend/internal/adapters/api/stocks/handler.go` — 11 sites
- `backend/internal/adapters/api/syscodes/handler.go` — 6 sites
- `backend/internal/adapters/api/organizations/handler.go` — 5 sites
- `backend/internal/adapters/api/structure/handler.go` — 5 sites

### Modified (signature widening + new test)

- `backend/internal/infrastructure/db/context.go` — `RunInTenantTx` widened from `*pgxpool.Pool` to `QueryEngine` interface; `extractTenantCode` unchanged
- `backend/internal/infrastructure/db/context_test.go` — new, 88 lines, 2 subtests

### Modified (dependencies)

- `backend/go.mod` / `backend/go.sum` — added `github.com/pashagolub/pgxmock/v3 v3.4.0`; transitive `pgx/v5 5.4.3 → 5.5.5`, `testify 1.8.1 → 1.9.0`

### NOT modified (intentionally)

- `backend/internal/infrastructure/middleware/tenant.go` — wraps ctx at 55–57; correct as-is.
- `backend/internal/infrastructure/db/context.go::extractTenantCode` — fasthttp fallback retained as defensive code.
- `backend/internal/adapters/api/grid/handler.go` — already correct from commit `1dc091a`.