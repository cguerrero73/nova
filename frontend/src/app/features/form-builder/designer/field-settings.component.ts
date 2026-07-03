import { Component, inject, computed, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup } from '@angular/forms';
import { FormDesignerStore } from '../state/designer.store';
import {
  FieldType,
  FieldUi,
  ValidatorKind,
  FieldOption,
} from '../models/layout-definition.model';

/**
 * Right panel — edit label, helpText, validators, placeholder, defaultValue,
 * ui.width, ui.readOnly for the selected field.
 */
@Component({
  selector: 'app-field-settings',
  standalone: true,
  imports: [CommonModule, FormsModule, ReactiveFormsModule],
  template: `
    <div class="field-settings p-4">
      <h3 class="text-sm font-semibold text-gray-700 mb-3">Field Settings</h3>

      @if (!store.selectedField()) {
        <p class="text-xs text-gray-400">Select a field to edit its properties.</p>
      } @else {
        <div class="space-y-3">
          <!-- Label -->
          <div>
            <label class="block text-xs font-medium text-gray-600 mb-1">Label</label>
            <input
              type="text"
              [ngModel]="store.selectedField()!.ui.label"
              (ngModelChange)="patchField({ ui: { ...currentUi(), label: $event } })"
              class="w-full px-2 py-1 text-sm border rounded"
            />
          </div>

          <!-- Placeholder -->
          <div>
            <label class="block text-xs font-medium text-gray-600 mb-1">Placeholder</label>
            <input
              type="text"
              [ngModel]="store.selectedField()!.ui.placeholder || ''"
              (ngModelChange)="patchField({ ui: { ...currentUi(), placeholder: $event || undefined } })"
              class="w-full px-2 py-1 text-sm border rounded"
            />
          </div>

          <!-- Help text -->
          <div>
            <label class="block text-xs font-medium text-gray-600 mb-1">Help Text</label>
            <input
              type="text"
              [ngModel]="store.selectedField()!.ui.helpText || ''"
              (ngModelChange)="patchField({ ui: { ...currentUi(), helpText: $event || undefined } })"
              class="w-full px-2 py-1 text-sm border rounded"
            />
          </div>

          <!-- Width -->
          <div>
            <label class="block text-xs font-medium text-gray-600 mb-1">Width</label>
            <select
              [ngModel]="store.selectedField()!.ui.width || 'full'"
              (ngModelChange)="patchField({ ui: { ...currentUi(), width: $event } })"
              class="w-full px-2 py-1 text-sm border rounded"
            >
              <option value="full">Full</option>
              <option value="half">Half</option>
              <option value="third">Third</option>
            </select>
          </div>

          <!-- Read-only -->
          <div class="flex items-center gap-2">
            <input
              type="checkbox"
              [ngModel]="store.selectedField()!.ui.readOnly || false"
              (ngModelChange)="patchField({ ui: { ...currentUi(), readOnly: $event } })"
              id="readOnly"
            />
            <label for="readOnly" class="text-xs text-gray-600">Read-only</label>
          </div>

          <!-- Validators -->
          <div class="border-t pt-3">
            <label class="block text-xs font-medium text-gray-600 mb-2">Validators</label>

            <!-- Required -->
            <div class="flex items-center gap-2 mb-2">
              <input
                type="checkbox"
                [ngModel]="hasValidator('required')"
                (ngModelChange)="toggleRequired($event)"
                id="required"
              />
              <label for="required" class="text-xs text-gray-600">Required</label>
            </div>

            <!-- MinLength -->
            <div class="flex items-center gap-2 mb-2">
              <label class="text-xs text-gray-600 w-16">Min length</label>
              <input
                type="number"
                [ngModel]="getValidatorValue('minLength')"
                (ngModelChange)="setValidator('minLength', $event)"
                class="flex-1 px-2 py-0.5 text-xs border rounded"
                min="0"
              />
            </div>

            <!-- MaxLength -->
            <div class="flex items-center gap-2 mb-2">
              <label class="text-xs text-gray-600 w-16">Max length</label>
              <input
                type="number"
                [ngModel]="getValidatorValue('maxLength')"
                (ngModelChange)="setValidator('maxLength', $event)"
                class="flex-1 px-2 py-0.5 text-xs border rounded"
                min="0"
              />
            </div>

            <!-- Pattern -->
            <div class="flex items-center gap-2 mb-2">
              <label class="text-xs text-gray-600 w-16">Pattern</label>
              <input
                type="text"
                [ngModel]="getValidatorValue('pattern') || ''"
                (ngModelChange)="setValidator('pattern', $event)"
                class="flex-1 px-2 py-0.5 text-xs border rounded"
                placeholder="regex"
              />
            </div>

            <!-- Email -->
            <div class="flex items-center gap-2 mb-2">
              <input
                type="checkbox"
                [ngModel]="hasValidator('email')"
                (ngModelChange)="toggleEmail($event)"
                id="email"
              />
              <label for="email" class="text-xs text-gray-600">Email</label>
            </div>

            <!-- Min -->
            <div class="flex items-center gap-2 mb-2">
              <label class="text-xs text-gray-600 w-16">Min</label>
              <input
                type="number"
                [ngModel]="getValidatorValue('min')"
                (ngModelChange)="setValidator('min', $event)"
                class="flex-1 px-2 py-0.5 text-xs border rounded"
              />
            </div>

            <!-- Max -->
            <div class="flex items-center gap-2 mb-2">
              <label class="text-xs text-gray-600 w-16">Max</label>
              <input
                type="number"
                [ngModel]="getValidatorValue('max')"
                (ngModelChange)="setValidator('max', $event)"
                class="flex-1 px-2 py-0.5 text-xs border rounded"
              />
            </div>
          </div>

          <!-- Options (for select/radio/multiselect) -->
          @if (hasOptions()) {
            <div class="border-t pt-3">
              <label class="block text-xs font-medium text-gray-600 mb-2">Options</label>
              @for (opt of getOptions(); track $index; let i = $index) {
                <div class="flex items-center gap-1 mb-1">
                  <input
                    type="text"
                    [ngModel]="opt.label"
                    (ngModelChange)="updateOption(i, 'label', $event)"
                    class="flex-1 px-2 py-0.5 text-xs border rounded"
                    placeholder="Label"
                  />
                  <input
                    type="text"
                    [ngModel]="opt.value"
                    (ngModelChange)="updateOption(i, 'value', $event)"
                    class="w-20 px-2 py-0.5 text-xs border rounded"
                    placeholder="Value"
                  />
                  <button
                    class="text-red-400 hover:text-red-600 text-xs px-1"
                    (click)="removeOption(i)"
                  >
                    ✕
                  </button>
                </div>
              }
              <button
                class="text-xs text-blue-600 hover:text-blue-800"
                (click)="addOption()"
              >
                + Add option
              </button>
            </div>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .field-settings {
      border-left: 1px solid #e5e7eb;
      min-height: 100%;
    }
  `],
})
export class FieldSettingsComponent {
  readonly store = inject(FormDesignerStore);

  currentUi(): FieldUi {
    return this.store.selectedField()!.ui;
  }

  patchField(patch: Partial<FieldType>): void {
    const field = this.store.selectedField();
    const sectionName = this.store.selectedFieldSection();
    if (!field || !sectionName) return;

    const updated = { ...field, ...patch } as FieldType;
    this.store.updateField(sectionName, field.name, updated);
  }

  hasValidator(kind: string): boolean {
    const field = this.store.selectedField();
    if (!field) return false;
    return field.validators?.some((v) => v.kind === kind) ?? false;
  }

  getValidatorValue(kind: string): number | string | null {
    const field = this.store.selectedField();
    if (!field) return null;
    const v = field.validators?.find((v) => v.kind === kind);
    if (!v) return null;
    if ('value' in v) return v.value;
    return null;
  }

  toggleRequired(checked: boolean): void {
    this.setSimpleValidator('required', checked);
  }

  toggleEmail(checked: boolean): void {
    this.setSimpleValidator('email', checked);
  }

  setValidator(kind: string, value: unknown): void {
    const field = this.store.selectedField();
    const sectionName = this.store.selectedFieldSection();
    if (!field || !sectionName) return;

    let validators = [...(field.validators || [])];

    if (value === null || value === '' || value === undefined) {
      validators = validators.filter((v) => v.kind !== kind);
    } else {
      const existing = validators.findIndex((v) => v.kind === kind);
      const entry = { kind, value } as ValidatorKind;
      if (existing >= 0) {
        validators[existing] = entry;
      } else {
        validators.push(entry);
      }
    }

    this.store.updateField(sectionName, field.name, { ...field, validators } as FieldType);
  }

  private setSimpleValidator(kind: string, enabled: boolean): void {
    const field = this.store.selectedField();
    const sectionName = this.store.selectedFieldSection();
    if (!field || !sectionName) return;

    let validators = [...(field.validators || [])];
    if (enabled) {
      if (!validators.some((v) => v.kind === kind)) {
        validators.push({ kind } as ValidatorKind);
      }
    } else {
      validators = validators.filter((v) => v.kind !== kind);
    }

    this.store.updateField(sectionName, field.name, { ...field, validators } as FieldType);
  }

  hasOptions(): boolean {
    const field = this.store.selectedField();
    if (!field) return false;
    return field.type === 'select' || field.type === 'radio' || field.type === 'multiselect';
  }

  getOptions(): FieldOption[] {
    const field = this.store.selectedField();
    if (!field || !this.hasOptions()) return [];
    return (field as { options: FieldOption[] }).options || [];
  }

  addOption(): void {
    const field = this.store.selectedField();
    const sectionName = this.store.selectedFieldSection();
    if (!field || !sectionName || !this.hasOptions()) return;

    const options = [...this.getOptions(), { label: '', value: '' }];
    this.store.updateField(sectionName, field.name, { ...field, options } as FieldType);
  }

  removeOption(index: number): void {
    const field = this.store.selectedField();
    const sectionName = this.store.selectedFieldSection();
    if (!field || !sectionName) return;

    const options = this.getOptions().filter((_, i) => i !== index);
    this.store.updateField(sectionName, field.name, { ...field, options } as FieldType);
  }

  updateOption(index: number, prop: 'label' | 'value', value: string): void {
    const field = this.store.selectedField();
    const sectionName = this.store.selectedFieldSection();
    if (!field || !sectionName) return;

    const options = this.getOptions().map((opt, i) => {
      if (i !== index) return opt;
      return { ...opt, [prop]: prop === 'value' && !isNaN(+value) ? +value : value };
    });
    this.store.updateField(sectionName, field.name, { ...field, options } as FieldType);
  }
}
