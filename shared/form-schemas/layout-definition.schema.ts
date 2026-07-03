import { z } from 'zod';
import { FieldTypeSchema } from './field-types';
import { CrossFieldRuleSchema } from './rules';

/**
 * A section groups related fields within a layout.
 */
export const SectionSchema = z.object({
  name: z.string(),
  title: z.string().optional(),
  order: z.number().int().min(0),
  fields: z.array(FieldTypeSchema),
});

/**
 * Complete layout definition — the JSON document stored as a published version.
 */
export const LayoutDefinitionSchema = z.object({
  formKey: z.string(),
  layoutName: z.string(),
  sections: z.array(SectionSchema),
  rules: z.array(CrossFieldRuleSchema).optional().default([]),
});

export type LayoutDefinition = z.infer<typeof LayoutDefinitionSchema>;
export type Section = z.infer<typeof SectionSchema>;
