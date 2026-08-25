import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';
import { ApiService } from '@core/services/api.service';
import { RoleAssignment } from '../models/designer.model';

/**
 * HTTP client for the form-builder assignment API.
 * Manages role-to-layout mappings.
 * Backend wraps responses as { success: boolean, data: T }, so we extract data.
 */
@Injectable({ providedIn: 'root' })
export class AssignmentService {
  private readonly api = inject(ApiService);

  listAssignments(formKey: string): Observable<RoleAssignment[]> {
    return this.api
      .get<RoleAssignment[]>(`/formbuilder/forms/${formKey}/assignments`)
      .pipe(map((response) => response.data!));
  }

  assignRole(
    formKey: string,
    roleName: string,
    layoutName: string,
  ): Observable<RoleAssignment> {
    return this.api
      .put<RoleAssignment>(
        `/formbuilder/forms/${formKey}/assignments`,
        roleName,
        { layoutName },
      )
      .pipe(map((response) => response.data!));
  }

  revokeAssignment(
    formKey: string,
    roleName: string,
  ): Observable<null> {
    return this.api
      .deleteRaw(`/formbuilder/forms/${formKey}/assignments/${roleName}`)
      .pipe(map(() => null));
  }
}
