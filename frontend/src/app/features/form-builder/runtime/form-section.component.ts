import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormGroup } from '@angular/forms';
import { Section, FieldType } from '../models/layout-definition.model';
import { FieldRendererComponent } from './field-renderer.component';

@Component({
  selector: 'app-form-section',
  standalone: true,
  imports: [CommonModule, FieldRendererComponent],
  template: `
    <fieldset class="form-section">
      @if (section().title) {
        <legend class="section-title">{{ section().title }}</legend>
      }
      <div class="section-grid">
        @for (field of section().fields; track field.name) {
          @if (!hiddenFields().includes(field.name)) {
            <app-field-renderer
              [field]="field"
              [control]="getControl(field)"
              [isRequired]="hasRequiredValidator(field)"
              [width]="field.ui.width || 'full'"
              [hidden]="hiddenFields().includes(field.name)"
            />
          }
        }
      </div>
    </fieldset>
  `,
  styles: [`
    .form-section {
      border: 1px solid #e5e7eb;
      border-radius: 0.5rem;
      padding: 1rem;
      margin-bottom: 1rem;
    }
    .section-title {
      font-size: 1rem;
      font-weight: 600;
      padding: 0 0.5rem;
      color: #374151;
    }
    .section-grid {
      display: grid;
      grid-template-columns: repeat(12, 1fr);
      gap: 1rem;
    }
  `],
})
export class FormSectionComponent {
  section = input.required<Section>();
  group = input.required<FormGroup>();
  hiddenFields = input<string[]>([]);

  getControl(field: FieldType) {
    return this.group().get(field.name) as import('@angular/forms').FormControl;
  }

  hasRequiredValidator(field: FieldType): boolean {
    return field.validators?.some((v) => v.kind === 'required') ?? false;
  }
}
