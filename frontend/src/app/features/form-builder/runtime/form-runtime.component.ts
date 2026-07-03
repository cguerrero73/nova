import {
  Component,
  input,
  output,
  OutputEmitterRef,
  OnInit,
  DestroyRef,
  inject,
  signal,
  computed,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  ReactiveFormsModule,
  FormBuilder,
  FormGroup,
  FormControl,
  Validators,
  ValidatorFn,
  ValidationErrors,
} from '@angular/forms';
import { LayoutDefinition, Section, FieldType, ValidatorKind, CrossFieldRule } from '../models/layout-definition.model';
import { FormSectionComponent } from './form-section.component';
import { evaluateCrossFieldRules } from './cross-field-validator';

@Component({
  selector: 'app-form-runtime',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, FormSectionComponent],
  template: `
    @if (layout(); as layoutDef) {
      <form [formGroup]="form" class="form-runtime" (ngSubmit)="onSubmit()">
        @for (section of sortedSections(); track section.name) {
          <app-form-section
            [section]="section"
            [group]="form"
            [hiddenFields]="hiddenFields()"
          />
        }
        @if (showActions()) {
          <div class="flex gap-3 mt-6 pt-4 border-t border-gray-200">
            <button type="submit" class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                    [disabled]="form.invalid">
              {{ submitLabel() }}
            </button>
            <button type="button" (click)="onCancel()" class="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300">
              Cancel
            </button>
          </div>
        }
      </form>
    } @else {
      <div class="form-loading">Loading form...</div>
    }
  `,
  styles: [`
    .form-runtime {
      max-width: 48rem;
      margin: 0 auto;
      padding: 1rem;
    }
    .form-loading {
      text-align: center;
      padding: 2rem;
      color: #6b7280;
    }
  `],
})
export class FormRuntimeComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly destroyRef = inject(DestroyRef);

  layout = input<LayoutDefinition | null>(null);
  initialValues = input<Record<string, any>>({});
  submitLabel = input<string>('Submit');
  showActions = input<boolean>(true);

  readonly formSubmit = output<Record<string, any>>();
  readonly formCancel = output<void>();

  form!: FormGroup;

  hiddenFields = signal<string[]>([]);

  sortedSections = computed<Section[]>(() => {
    const l = this.layout();
    if (!l) return [];
    return [...l.sections].sort((a, b) => a.order - b.order);
  });

  ngOnInit(): void {
    this.form = this.buildForm();
    this.setupCrossFieldRules();
  }

  onSubmit(): void {
    if (this.form.valid) {
      this.formSubmit.emit(this.form.value);
    } else {
      this.form.markAllAsTouched();
    }
  }

  onCancel(): void {
    this.formCancel.emit();
  }

  private buildForm(): FormGroup {
    const l = this.layout();
    if (!l) return this.fb.group({});

    const controls: Record<string, FormControl> = {};
    const initVals = this.initialValues();

    for (const section of l.sections) {
      for (const field of section.fields) {
        const validators = this.translateValidators(field.validators || []);
        const initialValue = initVals[field.name] !== undefined
          ? initVals[field.name]
          : this.defaultValue(field);
        const control = new FormControl(
          { value: initialValue, disabled: field.ui.readOnly ?? false },
          validators,
        );
        controls[field.name] = control;
      }
    }

    return this.fb.group(controls);
  }

  private translateValidators(validators: ValidatorKind[]): ValidatorFn[] {
    const fns: ValidatorFn[] = [];
    for (const v of validators) {
      switch (v.kind) {
        case 'required':
          fns.push(Validators.required);
          break;
        case 'minLength':
          fns.push(Validators.minLength(v.value));
          break;
        case 'maxLength':
          fns.push(Validators.maxLength(v.value));
          break;
        case 'pattern':
          fns.push(Validators.pattern(v.value));
          break;
        case 'email':
          fns.push(Validators.email);
          break;
        case 'min':
          fns.push(Validators.min(v.value));
          break;
        case 'max':
          fns.push(Validators.max(v.value));
          break;
      }
    }
    return fns;
  }

  private defaultValue(field: FieldType): unknown {
    switch (field.type) {
      case 'checkbox':
        return false;
      case 'multiselect':
        return [];
      default:
        return '';
    }
  }

  private setupCrossFieldRules(): void {
    const rules = this.layout()?.rules;
    if (!rules || rules.length === 0) return;

    const sub = this.form.valueChanges.subscribe(() => {
      this.evaluateRules(rules);
    });

    this.destroyRef.onDestroy(() => sub.unsubscribe());

    // Initial evaluation
    this.evaluateRules(rules);
  }

  private evaluateRules(rules: CrossFieldRule[]): void {
    const result = evaluateCrossFieldRules(rules, this.form);
    this.hiddenFields.set(result.hiddenFields);

    // Clear previous cross-field errors from all target controls
    const crossFieldErrorKeys = ['equals', 'notEquals', 'requiredIf'];
    const targetControls = new Set(rules.map((r) => r.target));
    for (const controlName of targetControls) {
      const ctrl = this.form.get(controlName);
      if (ctrl?.errors) {
        const cleaned = { ...ctrl.errors };
        for (const key of crossFieldErrorKeys) {
          delete cleaned[key];
        }
        ctrl.setErrors(Object.keys(cleaned).length > 0 ? cleaned : null, { emitEvent: false });
      }
    }

    // Set new cross-field errors (merge with existing built-in errors)
    for (const [controlName, error] of Object.entries(result.errors) as [string, ValidationErrors][]) {
      const ctrl = this.form.get(controlName);
      if (ctrl) {
        const merged = { ...(ctrl.errors || {}), ...error };
        ctrl.setErrors(merged, { emitEvent: false });
      }
    }
  }
}
