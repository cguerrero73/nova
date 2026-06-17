import { test, expect } from '../fixtures/auth';
import { QueryBuilderPage } from '../helpers/query-builder';

/**
 * E2E regression test for the PUT /queries/:id bug.
 *
 * Before the fix:
 *   The query builder loaded saved queries whose sort[].field and
 *   filters[].field/operator arrived from the backend as strings. When the
 *   user saved the query unchanged, the PUT body serialized those strings
 *   back to the Go backend, which expected ints and rejected with 400.
 *
 * After the fix (commit 4f5729f):
 *   open() and onQueryChange() coerce field/operator to numbers, and
 *   addSort/addFilter do the same when creating new entries.
 *
 * This test reproduces the original flow against the real backend and asserts
 * the PUT body uses numeric types and the response is 2xx.
 */
test.describe('Query Builder — PUT /queries/:id', () => {
  test('updates an existing query without type errors', async ({ authenticatedPage }) => {
    // 1. Navigate to the users grid (the grid that uses BMUSER/queries)
    await authenticatedPage.goto('/users');
    await expect(authenticatedPage.locator('app-user-list')).toBeVisible();

    // 2. Capture the PUT request and its response
    const putResponsePromise = authenticatedPage.waitForResponse(
      (r) =>
        r.url().includes('/api/v1/queries/') &&
        r.request().method() === 'PUT' &&
        !r.url().endsWith('/queries'),
    );

    // 3. Open the query builder and load the seeded "Todos los registros" query
    const qb = new QueryBuilderPage(authenticatedPage);
    await qb.open();
    await qb.selectQueryByLabel('Todos los registros');

    // 4. Modify a sort entry (toggles the path that originally produced strings
    //    from the backend and triggered the bug)
    await qb.addSort('Código', 'Ascendente');

    // 5. Save and wait for the PUT to land
    await qb.save();
    const putResponse = await putResponsePromise;

    // 6. The backend must accept the payload (no more 400)
    expect(putResponse.status()).toBeLessThan(300);

    // 7. Defense in depth: verify the payload we sent had numeric types.
    //    This catches regressions even if the backend becomes lenient.
    const requestBody = putResponse.request().postDataJSON() as {
      query?: { sort?: Array<{ field: unknown }>; filters?: Array<{ field: unknown; operator: unknown }> };
    };
    for (const entry of requestBody.query?.sort ?? []) {
      expect(typeof entry.field).toBe('number');
    }
    for (const entry of requestBody.query?.filters ?? []) {
      expect(typeof entry.field).toBe('number');
      expect(typeof entry.operator).toBe('number');
    }
  });

  test('creates a fresh query with numeric sort/filter fields', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/users');
    const qb = new QueryBuilderPage(authenticatedPage);

    // Capture both POST (create) and the subsequent PUT (if any)
    const postResponsePromise = authenticatedPage.waitForResponse(
      (r) => r.url().endsWith('/api/v1/queries') && r.request().method() === 'POST',
    );

    await qb.open();
    // New query path: type a name and add a sort + filter
    await qb.nameInput.fill('TEST_E2E_QUERY');
    await qb.addSort('Código', 'Ascendente');
    await qb.addFilter('Código', '=', 'A1');
    await qb.save();

    const postResponse = await postResponsePromise;
    expect(postResponse.status()).toBeLessThan(300);

    const body = postResponse.request().postDataJSON() as {
      query?: { sort?: Array<{ field: unknown }>; filters?: Array<{ field: unknown; operator: unknown }> };
    };
    for (const entry of body.query?.sort ?? []) {
      expect(typeof entry.field).toBe('number');
    }
    for (const entry of body.query?.filters ?? []) {
      expect(typeof entry.field).toBe('number');
      expect(typeof entry.operator).toBe('number');
    }
  });
});
