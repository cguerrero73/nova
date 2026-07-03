import { z } from 'zod';
import { ValidatorKindSchema } from './validators';

/**
 * UI configuration shared by all field types.
 */
export const FieldUiSchema = z.object({
  label: z.string(),
  placeholder: z.string().optional(),
  helpText: z.string().optional(),
  readOnly: z.boolean().optional().default(false),
  width: z.enum(['full', 'half', 'third']).optional().default('full'),
});

/**
 * Option for select / radio / multiselect fields.
 */
export const FieldOptionSchema = z.object({
  label: z.string(),
  value: z.union([z.string(), z.number()]),
});

/**
 * Discriminated union of the 8 supported field types.
 * Discriminator: `type`.
 */
export const FieldTypeSchema = z.discriminatedUnion('type', [
  // 1. Text
  z.object({
    type: z.literal('text'),
    name: z.string(),
    ui: FieldUiSchema,
    validators: z.array(ValidatorKindSchema).optional().default([]),
  }),
  // 2. Textarea
  z.object({
    type: z.literal('textarea'),
    name: z.string(),
    ui: FieldUiSchema,
    validators: z.array(ValidatorKindSchema).optional().default([]),
  }),
  // 3. Number
  z.object({
    type: z.literal('number'),
    name: z.string(),
    ui: FieldUiSchema,
    validators: z.array(ValidatorKindSchema).optional().default([]),
  }),
  // 4. Date
  z.object({
    type: z.literal('date'),
    name: z.string(),
    ui: FieldUiSchema,
    validators: z.array(ValidatorKindSchema).optional().default([]),
  }),
  // 5. Checkbox
  z.object({
    type: z.literal('checkbox'),
    name: z.string(),
    ui: FieldUiSchema,
    validators: z.array(ValidatorKindSchema).optional().default([]),
  }),
  // 6. Select
  z.object({
    type: z.literal('select'),
    name: z.string(),
    ui: FieldUiSchema,
    options: z.array(FieldOptionSchema),
    validators: z.array(ValidatorKindSchema).optional().default([]),
  }),
  // 7. Radio
  z.object({
    type: z.literal('radio'),
    name: z.string(),
    ui: FieldUiSchema,
    options: z.array(FieldOptionSchema),
    validators: z.array(ValidatorKindSchema).optional().default([]),
  }),
  // 8. Multiselect
  z.object({
    type: z.literal('multiselect'),
    name: z.string(),
    ui: FieldUiSchema,
    options: z.array(FieldOptionSchema),
    validators: z.array(ValidatorKindSchema).optional().default([]),
  }),
]);

export type FieldType = z.infer<typeof FieldTypeSchema>;
export type FieldUi = z.infer<typeof FieldUiSchema>;
export type FieldOption = z.infer<typeof FieldOptionSchema>;
