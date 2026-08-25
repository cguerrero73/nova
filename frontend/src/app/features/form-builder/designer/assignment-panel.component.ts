import { Component, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FormDesignerStore } from '../state/designer.store';
import { FormLayout } from '../models/designer.model';

/**
 * Lists roles and their assigned layouts, lets designer change mapping.
 * Roles without assignment show "→ uses default layout".
 */
@Component({
  selector: 'app-assignment-panel',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="assignment-panel">
      <h3 class="text-sm font-semibold text-gray-700 mb-2">Role Assignments</h3>

      @if (store.layouts().length === 0) {
        <p class="text-xs text-gray-500">No layouts available.</p>
      } @else {
        <div class="space-y-2">
          @for (role of allRoles(); track role) {
            <div class="flex items-center gap-2 text-sm">
              <span class="font-medium min-w-[80px]">{{ role }}</span>
              <select
                [ngModel]="getAssignedLayout(role)"
                (ngModelChange)="onAssignmentChange(role, $event)"
                class="flex-1 px-2 py-1 text-xs border rounded"
              >
                <option value="">→ uses default layout</option>
                @for (layout of store.layouts(); track layout.fl_id) {
                  @if (layout.fl_name !== 'default') {
                    <option [value]="layout.fl_name">
                      {{ layout.fl_display_name || layout.fl_name }}
                    </option>
                  }
                }
              </select>
            </div>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .assignment-panel {
      padding: 0.5rem;
    }
  `],
})
export class AssignmentPanelComponent {
  readonly store = inject(FormDesignerStore);

  /**
   * Known roles — in a real app this would come from a roles API.
   * For now, we derive from assignments + common defaults.
   */
  allRoles = computed<string[]>(() => {
    const assignments = this.store.assignments();
    const roleSet = new Set<string>();
    for (const a of assignments) {
      roleSet.add(a.fra_role_name);
    }
    // Always show common roles
    roleSet.add('ADMIN');
    roleSet.add('VIEWER');
    return Array.from(roleSet).sort();
  });

  getAssignedLayout(roleName: string): string {
    const assignment = this.store.assignments().find(
      (a) => a.fra_role_name === roleName,
    );
    return assignment?.fra_layout_name ?? '';
  }

  onAssignmentChange(roleName: string, layoutName: string): void {
    if (layoutName) {
      this.store.assign(roleName, layoutName);
    } else {
      this.store.revoke(roleName);
    }
  }
}
