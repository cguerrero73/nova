import {
  Directive,
  input,
  OnInit,
  OnDestroy,
  inject,
  DestroyRef,
} from '@angular/core';
import { FormGroup } from '@angular/forms';
import { CrossFieldRule } from '../models/layout-definition.model';
import { evaluateCrossFieldRules } from './cross-field-validator';

/**
 * Structural directive that evaluates cross-field rules on a FormGroup.
 * Subscribes to valueChanges and applies hidden/error state.
 *
 * Usage:
 * <form [formGroup]="form" appCrossFieldValidator [rules]="layout.rules">
 */
@Directive({
  selector: '[appCrossFieldValidator]',
  standalone: true,
})
export class CrossFieldValidatorDirective implements OnInit, OnDestroy {
  private readonly destroyRef = inject(DestroyRef);

  /** The FormGroup to evaluate rules against. */
  form = input.required<FormGroup>({ alias: 'appCrossFieldValidator' });

  /** The cross-field rules to evaluate. */
  rules = input<CrossFieldRule[]>([]);

  /** Signal-compatible callback for hidden fields. */
  onHiddenFieldsChange: ((fields: string[]) => void) | null = null;

  private subscription: { unsubscribe: () => void } | null = null;

  ngOnInit(): void {
    const form = this.form();
    const rulesList = this.rules();

    if (!rulesList || rulesList.length === 0) return;

    // Initial evaluation
    this.evaluate(form, rulesList);

    // Subscribe to changes
    this.subscription = form.valueChanges.subscribe(() => {
      this.evaluate(form, this.rules());
    });

    this.destroyRef.onDestroy(() => {
      this.subscription?.unsubscribe();
    });
  }

  ngOnDestroy(): void {
    this.subscription?.unsubscribe();
  }

  private evaluate(form: FormGroup, rulesList: CrossFieldRule[]): void {
    const result = evaluateCrossFieldRules(rulesList, form);

    // Notify hidden fields change
    if (this.onHiddenFieldsChange) {
      this.onHiddenFieldsChange(result.hiddenFields);
    }

    // Apply errors
    for (const [controlName, error] of Object.entries(result.errors) as [string, import('@angular/forms').ValidationErrors][]) {
      const ctrl = form.get(controlName);
      if (ctrl) {
        ctrl.setErrors(error, { emitEvent: false });
      }
    }
  }
}
