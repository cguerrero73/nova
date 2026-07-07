import { Injectable, inject } from '@angular/core';
import { Observable, tap, map } from 'rxjs';
import { ApiService } from '@core/services/api.service';
import { LayoutDefinition } from '../models/layout-definition.model';

/**
 * HTTP client for the form-builder runtime API.
 * Resolves a layout by formKey (assignment → default fallback → published version).
 */
@Injectable({ providedIn: 'root' })
export class FormRuntimeService {
  private readonly api = inject(ApiService);

  /**
   * GET /api/formbuilder/forms/:formKey
   * Returns the resolved layout definition for the current user's role.
   * Backend returns: { success: true, data: { formKey, layoutName, version, definition } }
   * We extract the definition (LayoutDefinition) from the response.
   */
  resolveForm(formKey: string): Observable<LayoutDefinition> {
    return this.api
      .getRaw<{
        success: boolean;
        data: { definition: LayoutDefinition };
      }>(`/formbuilder/forms/${formKey}`)
      .pipe(
        map((response) => response.data.definition),
        tap(() => {
          // Layout is ready for the renderer
        })
      );
  }
}
