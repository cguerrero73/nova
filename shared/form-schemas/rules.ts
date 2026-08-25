import { z } from 'zod';

/**
 * Cross-field rules evaluated at runtime against the FormGroup values.
 *
 * - `equals`    — target value must equal the source field value
 * - `notEquals` — target value must differ from the source field value
 * - `requiredIf` — target becomes required when source is truthy
 * - `hiddenIf`  — target is hidden when source is truthy
 */
export const CrossFieldRuleSchema = z.discriminatedUnion('operator', [
  z.object({
    operator: z.literal('equals'),
    source: z.string(),
    target: z.string(),
    message: z.string().optional(),
  }),
  z.object({
    operator: z.literal('notEquals'),
    source: z.string(),
    target: z.string(),
    message: z.string().optional(),
  }),
  z.object({
    operator: z.literal('requiredIf'),
    source: z.string(),
    target: z.string(),
    message: z.string().optional(),
  }),
  z.object({
    operator: z.literal('hiddenIf'),
    source: z.string(),
    target: z.string(),
  }),
]);

export type CrossFieldRule = z.infer<typeof CrossFieldRuleSchema>;
