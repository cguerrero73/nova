import { FormGroup, ValidationErrors } from '@angular/forms';
import { CrossFieldRule } from '../models/layout-definition.model';

export interface CrossFieldResult {
  /** Control names that should be hidden (from hiddenIf rules). */
  hiddenFields: string[];
  /** Errors keyed by control name. */
  errors: Record<string, ValidationErrors>;
}

/**
 * Evaluates an array of cross-field rules against the current FormGroup values.
 * Returns hidden fields and validation errors to apply.
 */
export function evaluateCrossFieldRules(
  rules: CrossFieldRule[],
  form: FormGroup,
): CrossFieldResult {
  const hiddenFields: string[] = [];
  const errors: Record<string, ValidationErrors> = {};

  for (const rule of rules) {
    const sourceCtrl = form.get(rule.source);
    const targetCtrl = form.get(rule.target);

    if (!sourceCtrl || !targetCtrl) continue;

    const sourceValue = sourceCtrl.value;

    switch (rule.operator) {
      case 'hiddenIf':
        if (isTruthy(sourceValue)) {
          hiddenFields.push(rule.target);
        }
        break;

      case 'requiredIf':
        if (isTruthy(sourceValue)) {
          const targetValue = targetCtrl.value;
          if (isEmpty(targetValue)) {
            errors[rule.target] = {
              requiredIf: { message: rule.message ?? `${rule.target} is required` },
            };
          }
        }
        break;

      case 'equals': {
        const targetValue = targetCtrl.value;
        if (sourceValue !== targetValue) {
          errors[rule.target] = {
            equals: { message: rule.message ?? `${rule.target} must equal ${rule.source}` },
          };
        }
        break;
      }

      case 'notEquals': {
        const targetValue = targetCtrl.value;
        if (sourceValue === targetValue && !isEmpty(sourceValue)) {
          errors[rule.target] = {
            notEquals: { message: rule.message ?? `${rule.target} must not equal ${rule.source}` },
          };
        }
        break;
      }
    }
  }

  return { hiddenFields, errors };
}

function isTruthy(value: unknown): boolean {
  if (value === null || value === undefined) return false;
  if (typeof value === 'string') return value.length > 0;
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number') return value !== 0;
  if (Array.isArray(value)) return value.length > 0;
  return true;
}

function isEmpty(value: unknown): boolean {
  if (value === null || value === undefined) return true;
  if (typeof value === 'string') return value.trim().length === 0;
  if (Array.isArray(value)) return value.length === 0;
  return false;
}
