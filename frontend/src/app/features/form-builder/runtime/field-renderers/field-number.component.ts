import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormControl } from '@angular/forms';
import { NumberField } from '../../models/layout-definition.model';

@Component({
  selector: 'app-field-number',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  template: `
    <div class="field-wrapper">
      <label [for]="field().name" class="field-label">
        {{ field().ui.label }}
        @if (isRequired()) { <span class="required-mark">*</span> }
      </label>
      <input
        [id]="field().name"
        type="number"
        [formControl]="control()"
        [placeholder]="field().ui.placeholder || ''"
        class="field-input"
      />
      @if (field().ui.helpText) {
        <span class="field-help">{{ field().ui.helpText }}</span>
      }
      @if (control().invalid && control().touched) {
        <div class="field-errors">
          @if (control().hasError('required')) { <span class="field-error">This field is required</span> }
          @if (control().hasError('min')) { <span class="field-error">Value is too small</span> }
          @if (control().hasError('max')) { <span class="field-error">Value is too large</span> }
        </div>
      }
    </div>
  `,
  styles: [`
    .field-wrapper { display: flex; flex-direction: column; gap: 0.25rem; }
    .field-label { font-weight: 500; font-size: 0.875rem; }
    .required-mark { color: #dc2626; }
    .field-input { padding: 0.5rem; border: 1px solid #d1d5db; border-radius: 0.375rem; font-size: 0.875rem; }
    .field-input:disabled { background-color: #f3f4f6; cursor: not-allowed; }
    .field-help { font-size: 0.75rem; color: #6b7280; }
    .field-errors { display: flex; flex-direction: column; gap: 0.125rem; }
    .field-error { font-size: 0.75rem; color: #dc2626; }
  `],
})
export class FieldNumberComponent {
  field = input.required<NumberField>();
  control = input.required<FormControl>();
  isRequired = input(false);
}
