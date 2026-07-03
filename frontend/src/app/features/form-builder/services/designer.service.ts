import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '@core/services/api.service';
import { LayoutDefinition } from '../models/layout-definition.model';
import {
  FormDefinition,
  FormLayout,
  LayoutVersion,
} from '../models/designer.model';

/**
 * HTTP client for the form-builder designer API.
 * Covers CRUD forms/layouts/drafts/publish/versions.
 */
@Injectable({ providedIn: 'root' })
export class FormDesignerService {
  private readonly api = inject(ApiService);

  // --- Forms ---

  listForms(): Observable<FormDefinition[]> {
    return this.api.getRaw<FormDefinition[]>('/formbuilder/forms');
  }

  getForm(formKey: string): Observable<FormDefinition> {
    return this.api.getRaw<FormDefinition>(`/formbuilder/forms/${formKey}`);
  }

  // --- Layouts ---

  listLayouts(formKey: string): Observable<FormLayout[]> {
    return this.api.getRaw<FormLayout[]>(
      `/formbuilder/forms/${formKey}/layouts`,
    );
  }

  createLayout(
    formKey: string,
    body: { name: string; displayName: string; description?: string },
  ): Observable<FormLayout> {
    return this.api.postRaw<FormLayout>(
      `/formbuilder/forms/${formKey}/layouts`,
      body,
    );
  }

  archiveLayout(formKey: string, layoutName: string): Observable<null> {
    return this.api.postRaw<null>(
      `/formbuilder/forms/${formKey}/layouts/${layoutName}/archive`,
      {},
    );
  }

  // --- Draft ---

  getDraft(formKey: string, layoutName: string): Observable<LayoutDefinition> {
    return this.api.getRaw<LayoutDefinition>(
      `/formbuilder/forms/${formKey}/layouts/${layoutName}/draft`,
    );
  }

  saveDraft(
    formKey: string,
    layoutName: string,
    definition: LayoutDefinition,
  ): Observable<LayoutDefinition> {
    return this.api.putRaw<LayoutDefinition>(
      `/formbuilder/forms/${formKey}/layouts/${layoutName}/draft`,
      definition,
    );
  }

  // --- Publish ---

  publishLayout(
    formKey: string,
    layoutName: string,
    description: string,
  ): Observable<LayoutVersion> {
    return this.api.postRaw<LayoutVersion>(
      `/formbuilder/forms/${formKey}/layouts/${layoutName}/publish`,
      { description },
    );
  }

  revertLayout(
    formKey: string,
    layoutName: string,
    versionNumber: number,
  ): Observable<LayoutDefinition> {
    return this.api.postRaw<LayoutDefinition>(
      `/formbuilder/forms/${formKey}/layouts/${layoutName}/revert`,
      { versionNumber },
    );
  }

  // --- Versions ---

  listVersions(
    formKey: string,
    layoutName: string,
  ): Observable<LayoutVersion[]> {
    return this.api.getRaw<LayoutVersion[]>(
      `/formbuilder/forms/${formKey}/layouts/${layoutName}/versions`,
    );
  }

  getVersion(
    formKey: string,
    layoutName: string,
    versionNumber: number,
  ): Observable<LayoutVersion> {
    return this.api.getRaw<LayoutVersion>(
      `/formbuilder/forms/${formKey}/layouts/${layoutName}/versions/${versionNumber}`,
    );
  }
}
