import { Injectable, signal, computed, inject } from '@angular/core';
import {
  LayoutDefinition,
  Section,
  FieldType,
} from '../models/layout-definition.model';
import {
  FormLayout,
  LayoutVersion,
  RoleAssignment,
} from '../models/designer.model';
import { FormDesignerService } from '../services/designer.service';
import { AssignmentService } from '../services/assignment.service';
import { LayoutDefinitionSchema } from '@form-schemas/layout-definition.schema';

/**
 * Signal-based store for the form designer.
 * Manages current layout, sections, selected field, dirty flag, and server state.
 */
@Injectable({ providedIn: 'root' })
export class FormDesignerStore {
  private readonly designerService = inject(FormDesignerService);
  private readonly assignmentService = inject(AssignmentService);

  // --- Form context ---
  private _formKey = signal<string>('');
  private _layouts = signal<FormLayout[]>([]);
  private _currentLayout = signal<FormLayout | null>(null);
  private _assignments = signal<RoleAssignment[]>([]);
  private _versions = signal<LayoutVersion[]>([]);

  // --- Designer state ---
  private _definition = signal<LayoutDefinition | null>(null);
  private _selectedFieldId = signal<string | null>(null);
  private _dirty = signal(false);
  private _saving = signal(false);
  private _loading = signal(false);
  private _error = signal<string | null>(null);
  private _validationErrors = signal<string[]>([]);
  private _lastSavedAt = signal<string | null>(null);

  // --- Public read-only signals ---
  readonly formKey = this._formKey.asReadonly();
  readonly layouts = this._layouts.asReadonly();
  readonly currentLayout = this._currentLayout.asReadonly();
  readonly assignments = this._assignments.asReadonly();
  readonly versions = this._versions.asReadonly();
  readonly definition = this._definition.asReadonly();
  readonly selectedFieldId = this._selectedFieldId.asReadonly();
  readonly dirty = this._dirty.asReadonly();
  readonly saving = this._saving.asReadonly();
  readonly loading = this._loading.asReadonly();
  readonly error = this._error.asReadonly();
  readonly validationErrors = this._validationErrors.asReadonly();
  readonly lastSavedAt = this._lastSavedAt.asReadonly();

  readonly sortedSections = computed<Section[]>(() => {
    const def = this._definition();
    if (!def) return [];
    return [...def.sections].sort((a, b) => a.order - b.order);
  });

  readonly selectedField = computed<FieldType | null>(() => {
    const id = this._selectedFieldId();
    if (!id) return null;
    const def = this._definition();
    if (!def) return null;
    for (const section of def.sections) {
      const field = section.fields.find((f) => f.name === id);
      if (field) return field;
    }
    return null;
  });

  readonly selectedFieldSection = computed<string | null>(() => {
    const id = this._selectedFieldId();
    if (!id) return null;
    const def = this._definition();
    if (!def) return null;
    for (const section of def.sections) {
      if (section.fields.some((f) => f.name === id)) return section.name;
    }
    return null;
  });

  // --- Actions ---

  init(formKey: string): void {
    this._formKey.set(formKey);
    this._loading.set(true);
    this._error.set(null);

    this.designerService.listLayouts(formKey).subscribe({
      next: (layouts) => {
        this._layouts.set(layouts);
        this._loading.set(false);
      },
      error: (err) => {
        this._error.set(err?.message ?? 'Failed to load layouts');
        this._loading.set(false);
      },
    });

    this.assignmentService.listAssignments(formKey).subscribe({
      next: (assignments) => this._assignments.set(assignments),
      error: () => {},
    });
  }

  selectLayout(layout: FormLayout): void {
    this._currentLayout.set(layout);
    this._selectedFieldId.set(null);
    this._dirty.set(false);
    this._validationErrors.set([]);
    this._loading.set(true);

    const formKey = this._formKey();
    this.designerService.getDraft(formKey, layout.fl_name).subscribe({
      next: (def) => {
        this._definition.set(def);
        this._loading.set(false);
        this._dirty.set(false);
      },
      error: () => {
        // No draft yet — start with empty definition
        this._definition.set({
          formKey,
          layoutName: layout.fl_name,
          sections: [],
          rules: [],
        });
        this._loading.set(false);
        this._dirty.set(false);
      },
    });

    this.designerService
      .listVersions(formKey, layout.fl_name)
      .subscribe({
        next: (versions) => this._versions.set(versions),
        error: () => {},
      });
  }

  selectField(fieldName: string | null): void {
    this._selectedFieldId.set(fieldName);
  }

  updateDefinition(def: LayoutDefinition): void {
    this._definition.set(def);
    this._dirty.set(true);
    this._validationErrors.set([]);
  }

  updateSection(sectionName: string, updated: Section): void {
    const def = this._definition();
    if (!def) return;
    const sections = def.sections.map((s) =>
      s.name === sectionName ? updated : s,
    );
    this.updateDefinition({ ...def, sections });
  }

  addSection(section: Section): void {
    const def = this._definition();
    if (!def) return;
    this.updateDefinition({ ...def, sections: [...def.sections, section] });
  }

  removeSection(sectionName: string): void {
    const def = this._definition();
    if (!def) return;
    this.updateDefinition({
      ...def,
      sections: def.sections.filter((s) => s.name !== sectionName),
    });
    if (this.selectedFieldSection() === null) {
      this._selectedFieldId.set(null);
    }
  }

  updateField(sectionName: string, fieldName: string, updated: FieldType): void {
    const def = this._definition();
    if (!def) return;
    const sections = def.sections.map((s) => {
      if (s.name !== sectionName) return s;
      return {
        ...s,
        fields: s.fields.map((f) => (f.name === fieldName ? updated : f)),
      };
    });
    this.updateDefinition({ ...def, sections });
  }

  removeField(sectionName: string, fieldName: string): void {
    const def = this._definition();
    if (!def) return;
    const sections = def.sections.map((s) => {
      if (s.name !== sectionName) return s;
      return { ...s, fields: s.fields.filter((f) => f.name !== fieldName) };
    });
    this.updateDefinition({ ...def, sections });
    if (this._selectedFieldId() === fieldName) {
      this._selectedFieldId.set(null);
    }
  }

  /** Validate with shared Zod schema, then save draft. */
  saveDraft(): void {
    const def = this._definition();
    if (!def) return;

    const result = LayoutDefinitionSchema.safeParse(def);
    if (!result.success) {
      const errors = result.error.issues.map(
        (i) => `${i.path.join('.')}: ${i.message}`,
      );
      this._validationErrors.set(errors);
      return;
    }

    this._saving.set(true);
    this._validationErrors.set([]);
    const formKey = this._formKey();
    const layoutName = this._currentLayout()!.fl_name;

    this.designerService.saveDraft(formKey, layoutName, def).subscribe({
      next: (saved) => {
        this._definition.set(saved);
        this._dirty.set(false);
        this._saving.set(false);
        this._lastSavedAt.set(new Date().toISOString());
      },
      error: (err) => {
        this._error.set(err?.message ?? 'Failed to save draft');
        this._saving.set(false);
      },
    });
  }

  publishLayout(description: string): void {
    const formKey = this._formKey();
    const layoutName = this._currentLayout()!.fl_name;
    this._saving.set(true);

    this.designerService
      .publishLayout(formKey, layoutName, description)
      .subscribe({
        next: () => {
          this._saving.set(false);
          this._dirty.set(false);
          // Refresh versions after publish
          this.designerService
            .listVersions(formKey, layoutName)
            .subscribe({
              next: (versions) => this._versions.set(versions),
              error: () => {},
            });
        },
        error: (err) => {
          this._error.set(err?.message ?? 'Failed to publish');
          this._saving.set(false);
        },
      });
  }

  loadAssignments(): void {
    this.assignmentService
      .listAssignments(this._formKey())
      .subscribe({
        next: (a) => this._assignments.set(a),
        error: () => {},
      });
  }

  assign(roleName: string, layoutName: string): void {
    this.assignmentService
      .assignRole(this._formKey(), roleName, layoutName)
      .subscribe({
        next: () => this.loadAssignments(),
        error: (err) => this._error.set(err?.message ?? 'Failed to assign'),
      });
  }

  revoke(roleName: string): void {
    this.assignmentService
      .revokeAssignment(this._formKey(), roleName)
      .subscribe({
        next: () => this.loadAssignments(),
        error: (err) => this._error.set(err?.message ?? 'Failed to revoke'),
      });
  }

  reset(): void {
    this._formKey.set('');
    this._layouts.set([]);
    this._currentLayout.set(null);
    this._definition.set(null);
    this._selectedFieldId.set(null);
    this._dirty.set(false);
    this._saving.set(false);
    this._loading.set(false);
    this._error.set(null);
    this._validationErrors.set([]);
    this._lastSavedAt.set(null);
    this._assignments.set([]);
    this._versions.set([]);
  }
}
