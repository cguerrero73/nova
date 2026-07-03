import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormControl } from '@angular/forms';
import { SelectField } from '../../models/layout-definition.model';

@Component({
  selector: 'app-field-select',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  template: `
    <div class="field-wrapper">
      <label [for]="field().name" class="field-label">
        {{ field().ui.label }}
        @if (isRequired) { <span class="required-mark">*</span> }
      </label>
      <select
        [id]="field().name"
        [formControl]="control()"
        class="field-select"
      >
        <option value="" disabled [selected]="!control().value">
          {{ field().ui.placeholder || 'Select an option' }}
        </option>
        @for (opt of field().options; track opt.value) {
          <option [ngValue]="opt.value">{{ opt.label }}</option>
        }
      </select>
      @if (field().ui.helpText) {
        <span class="field-help">{{ field().ui.helpText }}</span>
      }
      @if (control().invalid && control().touched) {
        <div class="field-errors">
          @if (control().hasError('required')) { <span class="field-error">This field is required</span> }
        </div>
      }
    </div>
  `,
  styles: [`
    .field-wrapper { display: flex; flex-direction: column; gap: 0.25rem; }
    .field-label { font-weight: 500; font-size: 0.875rem; }
    .required-mark { color: #dc2626; }
    .field-select { padding: 0.5rem; border: 1px solid #d1d5db; border-radius: 0.375rem; font-size: 0.875rem; background: white; }
    .field-select:disabled { background-color: #f3f4f6; cursor: not-allowed; }
    .field-help { font-size: 0.75rem; color: #6b7280; }
    .field-errors { display: flex; flex-direction: column; gap: 0.125rem; }
    .field-error { font-size: 0.75rem; color: #dc2626; }
  `],
})
export class FieldSelectComponent {
  field = input.required<SelectField>();
  control = input.required<FormControl>();
  isRequired = input(false);
}
