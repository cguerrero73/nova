/**
 * TypeScript types mirroring the shared Zod schemas.
 * These are the runtime types used by Angular components.
 * The Zod schemas are the single source of truth for validation.
 */

export type FieldWidth = 'full' | 'half' | 'third';

export interface FieldUi {
  label: string;
  placeholder?: string;
  helpText?: string;
  readOnly?: boolean;
  width?: FieldWidth;
}

export interface FieldOption {
  label: string;
  value: string | number;
}

// --- Validator kinds ---

export type ValidatorKind =
  | { kind: 'required' }
  | { kind: 'minLength'; value: number }
  | { kind: 'maxLength'; value: number }
  | { kind: 'pattern'; value: string }
  | { kind: 'email' }
  | { kind: 'min'; value: number }
  | { kind: 'max'; value: number };

// --- Field types (discriminated union on `type`) ---

interface BaseField {
  name: string;
  ui: FieldUi;
  validators?: ValidatorKind[];
}

export interface TextField extends BaseField {
  type: 'text';
}

export interface TextareaField extends BaseField {
  type: 'textarea';
}

export interface NumberField extends BaseField {
  type: 'number';
}

export interface DateField extends BaseField {
  type: 'date';
}

export interface CheckboxField extends BaseField {
  type: 'checkbox';
}

export interface SelectField extends BaseField {
  type: 'select';
  options: FieldOption[];
}

export interface RadioField extends BaseField {
  type: 'radio';
  options: FieldOption[];
}

export interface MultiselectField extends BaseField {
  type: 'multiselect';
  options: FieldOption[];
}

export type FieldType =
  | TextField
  | TextareaField
  | NumberField
  | DateField
  | CheckboxField
  | SelectField
  | RadioField
  | MultiselectField;

// --- Cross-field rules ---

export type CrossFieldRule =
  | { operator: 'equals'; source: string; target: string; message?: string }
  | { operator: 'notEquals'; source: string; target: string; message?: string }
  | { operator: 'requiredIf'; source: string; target: string; message?: string }
  | { operator: 'hiddenIf'; source: string; target: string };

// --- Section & Layout ---

export interface Section {
  name: string;
  title?: string;
  order: number;
  fields: FieldType[];
}

export interface LayoutDefinition {
  formKey: string;
  layoutName: string;
  sections: Section[];
  rules?: CrossFieldRule[];
}
