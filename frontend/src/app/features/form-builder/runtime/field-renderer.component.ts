import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormControl } from '@angular/forms';
import { FieldType } from '../models/layout-definition.model';

import { FieldTextComponent } from './field-renderers/field-text.component';
import { FieldTextareaComponent } from './field-renderers/field-textarea.component';
import { FieldNumberComponent } from './field-renderers/field-number.component';
import { FieldDateComponent } from './field-renderers/field-date.component';
import { FieldCheckboxComponent } from './field-renderers/field-checkbox.component';
import { FieldSelectComponent } from './field-renderers/field-select.component';
import { FieldRadioComponent } from './field-renderers/field-radio.component';
import { FieldMultiselectComponent } from './field-renderers/field-multiselect.component';

@Component({
  selector: 'app-field-renderer',
  standalone: true,
  imports: [
    CommonModule,
    FieldTextComponent,
    FieldTextareaComponent,
    FieldNumberComponent,
    FieldDateComponent,
    FieldCheckboxComponent,
    FieldSelectComponent,
    FieldRadioComponent,
    FieldMultiselectComponent,
  ],
  template: `
    @switch (field().type) {
      @case ('text') {
        <app-field-text [field]="$any(field())" [control]="control()" [isRequired]="isRequired()" />
      }
      @case ('textarea') {
        <app-field-textarea [field]="$any(field())" [control]="control()" [isRequired]="isRequired()" />
      }
      @case ('number') {
        <app-field-number [field]="$any(field())" [control]="control()" [isRequired]="isRequired()" />
      }
      @case ('date') {
        <app-field-date [field]="$any(field())" [control]="control()" [isRequired]="isRequired()" />
      }
      @case ('checkbox') {
        <app-field-checkbox [field]="$any(field())" [control]="control()" [isRequired]="isRequired()" />
      }
      @case ('select') {
        <app-field-select [field]="$any(field())" [control]="control()" [isRequired]="isRequired()" />
      }
      @case ('radio') {
        <app-field-radio [field]="$any(field())" [control]="control()" [isRequired]="isRequired()" />
      }
      @case ('multiselect') {
        <app-field-multiselect [field]="$any(field())" [control]="control()" [isRequired]="isRequired()" />
      }
    }
  `,
  host: {
    '[class.field-full]': 'width() === "full"',
    '[class.field-half]': 'width() === "half"',
    '[class.field-third]': 'width() === "third"',
    '[class.field-hidden]': 'hidden()',
  },
  styles: [`
    :host { display: block; }
    :host(.field-full) { grid-column: span 12 / span 12; }
    :host(.field-half) { grid-column: span 6 / span 6; }
    :host(.field-third) { grid-column: span 4 / span 4; }
    :host(.field-hidden) { display: none; }
  `],
})
export class FieldRendererComponent {
  field = input.required<FieldType>();
  control = input.required<FormControl>();
  isRequired = input(false);
  hidden = input(false);

  width = input<'full' | 'half' | 'third'>('full');
}
