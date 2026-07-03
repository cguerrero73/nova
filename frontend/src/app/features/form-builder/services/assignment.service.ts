import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '@core/services/api.service';
import { RoleAssignment } from '../models/designer.model';

/**
 * HTTP client for the form-builder assignment API.
 * Manages role-to-layout mappings.
 */
@Injectable({ providedIn: 'root' })
export class AssignmentService {
  private readonly api = inject(ApiService);

  listAssignments(formKey: string): Observable<RoleAssignment[]> {
    return this.api.getRaw<RoleAssignment[]>(
      `/formbuilder/forms/${formKey}/assignments`,
    );
  }

  assignRole(
    formKey: string,
    roleName: string,
    layoutName: string,
  ): Observable<RoleAssignment> {
    return this.api.putRaw<RoleAssignment>(
      `/formbuilder/forms/${formKey}/assignments/${roleName}`,
      { layoutName },
    );
  }

  revokeAssignment(
    formKey: string,
    roleName: string,
  ): Observable<null> {
    return this.api.deleteRaw(
      `/formbuilder/forms/${formKey}/assignments/${roleName}`,
    );
  }
}
