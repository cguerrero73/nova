import { Injectable, signal, computed, inject } from '@angular/core';
import { Observable, tap } from 'rxjs';
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
   */
  resolveForm(formKey: string): Observable<LayoutDefinition> {
    return this.api
      .getRaw<LayoutDefinition>(`/formbuilder/forms/${formKey}`)
      .pipe(
        tap((layout) => {
          // Layout is ready for the renderer
        }),
      );
  }
}
