# Design: funcionamiento-de-querybuilder

## Technical Approach

Fix three critical bugs in the QueryBuilder component that prevent saved queries from loading correctly and cause the fields list to render empty after reset. All changes are additive refactors with no schema migration.

## Architecture Decisions

### Decision: reset() must recompute available fields via method call

**Choice**: Replace `this.availableFieldsList.set([])` with `this.updateAvailableFields()` call in `reset()`
**Alternatives considered**: 
- Keep direct assignment — simpler but breaks drag-drop computed state
- Pass fields list as parameter — requires API change to `updateAvailableFields()`
**Rationale**: `updateAvailableFields()` correctly computes available fields as `allFields.filter(f => !selected.includes(f.id))`. After `reset()` sets `selectedFields` to all field IDs, calling the method yields an empty available list — same result as direct assignment, but with correct internal state for drag-drop operations.

### Decision: open() must fall back to initialQuery input when no argument passed

**Choice**: Add `if (!query && this.initialQuery()) { query = this.initialQuery(); }` at the start of `open()`
**Alternatives considered**:
- Require parent to always pass query explicitly — breaks existing usage patterns
- Use a computed signal for query source — more invasive refactor
**Rationale**: The `initialQuery` input already exists but is never used. Adding a fallback inside `open()` allows parent components to bind the query via input without changing call sites. Explicit argument still takes precedence.

### Decision: ExecuteQueryByID must pass config.Fields to repository

**Choice**: Change `s.repo.ExecuteQuery(ctx, grid.BaseQuery, nil, ...)` to `s.repo.ExecuteQuery(ctx, grid.BaseQuery, config.Fields, ...)`
**Alternatives considered**:
- Keep `nil` — backend returns all columns regardless of saved query field selection
- Create new repository method — adds unnecessary surface area
**Rationale**: The saved query already contains field selections in `config.Fields`. Passing `nil` ignores the user's saved field preferences. The existing `ExecuteQuery` method already supports field filtering via its third parameter.

## Data Flow

```
Parent Component
    │
    │ (optional) initialQuery binding
    ▼
QueryBuilder.open() ──fallback──► this.initialQuery()
    │
    │ query present → populate form fields/sort/filters
    │ no query      → reset() → updateAvailableFields()
    ▼
QueryBuilder.save() ──saved output event──► Parent
    │
    ▼
GridService.ExecuteQueryByID(config.Fields, config.Filters, config.Sort)
    │
    ▼
Backend Repository (respects field selection)
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `frontend/src/app/shared/components/query-builder/query-builder.component.ts` | Modify | Design 1: line 145. Design 2: lines 112-129 |
| `backend/internal/domain/grid/service.go` | Modify | Design 3: line 216 — pass `config.Fields` instead of `nil` |

## Interfaces / Contracts

No new interfaces. Existing contracts preserved:

**QueryBuilderComponent**
```typescript
// Inputs (unchanged)
initialQuery = input<SavedQuery | null>(null);

// open() signature unchanged — fallback handled internally
open(query?: SavedQuery): void

// reset() behavior unchanged externally — updateAvailableFields() called internally
reset(): void
```

**GridService.ExecuteQueryByID**
```go
// Signature unchanged — third parameter (fields) now receives config.Fields instead of nil
ExecuteQueryByID(ctx context.Context, gridID int64, queryID string, page, pageSize int) (*QueryResult, error)
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `reset()` calls `updateAvailableFields()` | Mock `fields` signal, verify `availableFieldsList` matches computed expectation |
| Unit | `open()` falls back to `initialQuery` | Set `initialQuery` input, call `open()` with no arg, verify form populated |
| Unit | `ExecuteQueryByID` passes `config.Fields` | Mock repository, verify third arg to `ExecuteQuery` is `config.Fields` |

## Migration / Rollout

No migration required. All changes are additive:
- `reset()` behavior is functionally identical (empty available list) but with correct internal state
- `open()` fallback only activates when no query argument passed — backward compatible
- `ExecuteQueryByID` passing fields may change column set returned for existing saved queries — verify existing saved queries still load correctly

## Open Questions

- None — all three bugs have clear fixes with straightforward implementation paths.