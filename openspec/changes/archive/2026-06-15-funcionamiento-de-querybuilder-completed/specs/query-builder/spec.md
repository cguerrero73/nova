# Delta for QueryBuilder

## Purpose

Fix critical bugs in the QueryBuilder component and add query selector UI.

## Status: COMPLETED (2026-06-15)

## Bugs Fixed

### Bug 1: QueryBuilder reset() must recompute available fields

**File:** `query-builder.component.ts`

When the user resets the QueryBuilder to create a new query, the system SHALL recalculate the available fields list from the complete set of fields.

The `reset()` method MUST call `updateAvailableFields()` to recompute the available fields list based on the current selection state, rather than directly assigning an empty array.

**Status:** ✅ FIXED — `reset()` calls `updateAvailableFields()` (line 151)

---

### Bug 2: QueryBuilder open() must load initialQuery when no argument provided

**File:** `query-builder.component.ts`

When `open()` is called without a query argument and `initialQuery` input is set, the system SHALL populate the form with the saved query configuration from `initialQuery`.

The `open()` method MUST check `this.initialQuery()` as a fallback when no query parameter is passed.

**Status:** ✅ FIXED — `open()` has fallback at lines 117-120

---

### Bug 3: ExecuteQueryByID must pass selected fields to repository

**File:** `backend/internal/domain/grid/service.go`

When executing a saved query by ID, the system SHALL pass the saved query's field selection to the repository layer.

**Status:** ✅ FIXED — `config.Fields` now set to `savedQuery.Query.Fields` (IDs), and `columnNames` passed to `ExecuteQuery()`

**Note:** Original spec said pass `config.Fields` but `config.Fields` is `[]int` (IDs) while `ExecuteQuery` expects `[]string` (names). Fix passes `columnNames` (converted names) directly, which is functionally correct.

---

### Bug 4: availableFieldsList empty on first open (NEW - discovered during fix)

**File:** `query-builder.component.ts`

**Problem:** On first open, `loadFields()` is async but `reset()` runs synchronously, so `availableFieldsList` was computed with empty `fields`. On subsequent opens, `fields` was already cached so it worked.

**Solution:** Changed `availableFieldsList` from `signal` to `computed`:
```typescript
availableFieldsList = computed(() => {
  const allFields = this.fields();
  const selected = this.selectedFields();
  return allFields.filter(f => !selected.includes(f.id)).map(f => f.id);
});
```

**Status:** ✅ FIXED — `availableFieldsList` is now a computed that automatically recalculates when `fields()` or `selectedFields()` change

---

## Feature Added: Query Selector UI

**File:** `query-builder.component.ts`, `query-builder.component.html`

### Changes

1. **Dropdown de queries existentes** — El campo "Consulta" ahora es un `<select>` que muestra todas las queries guardadas (`queries` input)
2. **Botón "+ Nueva"** — Crea una query en blanco, limpia todo y muestra input para nombre
3. **Botón "Copiar"** — Duplica la config actual (fields, sort, filters) y permite asignar nuevo nombre
4. **Input binding** — Se agregó `[queries]="queryService.queries()"` en `user-list.component.html`

### UI Behavior

| Acción | Dropdown → Input | Resultado |
|--------|------------------|-----------|
| Seleccionar query del dropdown | Mantiene dropdown | Carga config de esa query |
| + Nueva | Reemplaza por input vacío | Limpia todo para crear |
| Copiar | Reemplaza por input con "nombre (copia)" | Copia config actual |

**Status:** ✅ IMPLEMENTED

---

## Verification Criteria

| Bug/Feature | Fix | Verification |
|-------------|-----|--------------|
| reset() field list | `reset()` calls `updateAvailableFields()` | After reset, availableFieldsList shows all unselected fields in UI |
| initialQuery loading | `open()` fallback to `initialQuery()` | Saved query pre-fills form when user opens builder |
| ExecuteQueryByID passthrough | Pass `columnNames` to `ExecuteQuery()` | Backend returns only selected fields |
| availableFieldsList empty first open | `availableFieldsList` is `computed` | Fields appear on first open |
| Query selector dropdown | `[queries]` input + dropdown UI | Dropdown shows saved queries |
| Nueva/Copiar buttons | `newQuery()` / `copyQuery()` methods | Creates/duplicates query correctly |

---

## Files Changed

| File | Change |
|------|--------|
| `frontend/src/app/shared/components/query-builder/query-builder.component.ts` | Bug 4 fix + Feature: query selector state and methods |
| `frontend/src/app/shared/components/query-builder/query-builder.component.html` | Feature: dropdown/input toggle + Nueva/Copiar buttons |
| `frontend/src/app/features/users/screens/user-list/user-list.component.html` | Added `[queries]="queryService.queries()"` input |
| `backend/internal/domain/grid/service.go` | Bug 3 fix: `config.Fields` set to saved IDs |
