import { z } from 'zod';

/**
 * Validator kinds for form fields.
 * Each validator maps to an Angular Validator at runtime.
 */
export const ValidatorKindSchema = z.discriminatedUnion('kind', [
  z.object({ kind: z.literal('required') }),
  z.object({ kind: z.literal('minLength'), value: z.number().int().positive() }),
  z.object({ kind: z.literal('maxLength'), value: z.number().int().positive() }),
  z.object({ kind: z.literal('pattern'), value: z.string() }),
  z.object({ kind: z.literal('email') }),
  z.object({ kind: z.literal('min'), value: z.number() }),
  z.object({ kind: z.literal('max'), value: z.number() }),
]);

export type ValidatorKind = z.infer<typeof ValidatorKindSchema>;
