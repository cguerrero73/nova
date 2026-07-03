import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormControl } from '@angular/forms';
import { CheckboxField } from '../../models/layout-definition.model';

@Component({
  selector: 'app-field-checkbox',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  template: `
    <div class="field-wrapper field-checkbox-wrapper">
      <label class="checkbox-label">
        <input
          type="checkbox"
          [formControl]="control()"
          class="field-checkbox"
        />
        <span class="field-label">
          {{ field().ui.label }}
          @if (isRequired) { <span class="required-mark">*</span> }
        </span>
      </label>
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
    .field-checkbox-wrapper { flex-direction: row; align-items: center; }
    .checkbox-label { display: flex; align-items: center; gap: 0.5rem; cursor: pointer; }
    .field-label { font-weight: 500; font-size: 0.875rem; }
    .required-mark { color: #dc2626; }
    .field-checkbox { width: 1rem; height: 1rem; cursor: pointer; }
    .field-help { font-size: 0.75rem; color: #6b7280; }
    .field-errors { display: flex; flex-direction: column; gap: 0.125rem; }
    .field-error { font-size: 0.75rem; color: #dc2626; }
  `],
})
export class FieldCheckboxComponent {
  field = input.required<CheckboxField>();
  control = input.required<FormControl>();
  isRequired = input(false);
}
