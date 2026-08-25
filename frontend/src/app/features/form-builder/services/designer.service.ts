import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';
import { ApiService } from '@core/services/api.service';
import { LayoutDefinition } from '../models/layout-definition.model';
import {
  FormDefinition,
  FormLayout,
  LayoutVersion,
  AuditListResponse,
} from '../models/designer.model';

/**
 * HTTP client for the form-builder designer API.
 * Covers CRUD forms/layouts/drafts/publish/versions.
 * Backend wraps responses as { success: boolean, data: T }, so we extract data.
 */
@Injectable({ providedIn: 'root' })
export class FormDesignerService {
  private readonly api = inject(ApiService);

  // --- Forms ---

  listForms(): Observable<FormDefinition[]> {
    return this.api
      .get<FormDefinition[]>('/formbuilder/forms')
      .pipe(map((response) => response.data!));
  }

  getForm(formKey: string): Observable<FormDefinition> {
    return this.api
      .get<FormDefinition>(`/formbuilder/forms/${formKey}`)
      .pipe(map((response) => response.data!));
  }

  // --- Layouts ---

  listLayouts(formKey: string): Observable<FormLayout[]> {
    return this.api
      .get<FormLayout[]>(`/formbuilder/forms/${formKey}/layouts`)
      .pipe(map((response) => response.data!));
  }

  createLayout(
    formKey: string,
    body: { name: string; displayName: string; description?: string },
  ): Observable<FormLayout> {
    return this.api
      .post<FormLayout>(`/formbuilder/forms/${formKey}/layouts`, body)
      .pipe(map((response) => response.data!));
  }

  archiveLayout(formKey: string, layoutName: string): Observable<null> {
    return this.api
      .post<null>(`/formbuilder/forms/${formKey}/layouts/${layoutName}/archive`, {})
      .pipe(map((response) => response.data!));
  }

  // --- Draft ---

  getDraft(formKey: string, layoutName: string): Observable<LayoutDefinition> {
    return this.api
      .get<LayoutVersion>(
        `/formbuilder/forms/${formKey}/layouts/${layoutName}/draft`,
      )
      .pipe(map((response) => response.data!.flv_definition as LayoutDefinition));
  }

  saveDraft(
    formKey: string,
    layoutName: string,
    definition: LayoutDefinition,
  ): Observable<LayoutDefinition> {
    return this.api
      .putRaw<{ success: boolean; data: LayoutVersion }>(
        `/formbuilder/forms/${formKey}/layouts/${layoutName}/draft`,
        definition,
      )
      .pipe(map((response) => response.data!.flv_definition as LayoutDefinition));
  }

  // --- Publish & revert ---

  publishLayout(
    formKey: string,
    layoutName: string,
    description: string,
  ): Observable<LayoutVersion> {
    return this.api
      .post<LayoutVersion>(
        `/formbuilder/forms/${formKey}/layouts/${layoutName}/publish`,
        { description },
      )
      .pipe(map((response) => response.data!));
  }

  revertLayout(
    formKey: string,
    layoutName: string,
    versionNumber: number,
  ): Observable<LayoutDefinition> {
    return this.api
      .post<LayoutVersion>(
        `/formbuilder/forms/${formKey}/layouts/${layoutName}/revert`,
        { versionNumber },
      )
      .pipe(map((response) => response.data!.flv_definition as LayoutDefinition));
  }

  // --- Versions ---

  listVersions(
    formKey: string,
    layoutName: string,
  ): Observable<LayoutVersion[]> {
    return this.api
      .get<LayoutVersion[]>(
        `/formbuilder/forms/${formKey}/layouts/${layoutName}/versions`,
      )
      .pipe(map((response) => response.data!));
  }

  getVersion(
    formKey: string,
    layoutName: string,
    versionNumber: number,
  ): Observable<LayoutVersion> {
    return this.api
      .get<LayoutVersion>(
        `/formbuilder/forms/${formKey}/layouts/${layoutName}/versions/${versionNumber}`,
      )
      .pipe(map((response) => response.data!));
  }

  // --- Audit ---

  listAudit(
    formKey: string,
    params?: { page?: number; pageSize?: number; action?: string; entityType?: string },
  ): Observable<AuditListResponse> {
    const queryParts: string[] = [];
    if (params?.page) queryParts.push(`page=${params.page}`);
    if (params?.pageSize) queryParts.push(`pageSize=${params.pageSize}`);
    if (params?.action) queryParts.push(`action=${encodeURIComponent(params.action)}`);
    if (params?.entityType) queryParts.push(`entity_type=${encodeURIComponent(params.entityType)}`);
    const qs = queryParts.length ? `?${queryParts.join('&')}` : '';
    return this.api
      .get<AuditListResponse>(`/formbuilder/forms/${formKey}/audit${qs}`)
      .pipe(map((response) => response.data!));
  }
}
