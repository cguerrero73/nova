import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormControl } from '@angular/forms';
import { RadioField } from '../../models/layout-definition.model';

@Component({
  selector: 'app-field-radio',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  template: `
    <div class="field-wrapper">
      <span class="field-label">
        {{ field().ui.label }}
        @if (isRequired) { <span class="required-mark">*</span> }
      </span>
      <div class="radio-group" [attr.id]="field().name" role="radiogroup">
        @for (opt of field().options; track opt.value) {
          <label class="radio-label">
            <input
              type="radio"
              [formControl]="control()"
              [value]="opt.value"
              class="field-radio"
            />
            <span>{{ opt.label }}</span>
          </label>
        }
      </div>
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
    .radio-group { display: flex; flex-direction: column; gap: 0.375rem; }
    .radio-label { display: flex; align-items: center; gap: 0.5rem; cursor: pointer; font-size: 0.875rem; }
    .field-radio { width: 1rem; height: 1rem; cursor: pointer; }
    .field-help { font-size: 0.75rem; color: #6b7280; }
    .field-errors { display: flex; flex-direction: column; gap: 0.125rem; }
    .field-error { font-size: 0.75rem; color: #dc2626; }
  `],
})
export class FieldRadioComponent {
  field = input.required<RadioField>();
  control = input.required<FormControl>();
  isRequired = input(false);
}
