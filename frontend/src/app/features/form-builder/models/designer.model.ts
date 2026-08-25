/**
 * Backend DTOs for the form designer — mirrors the Go domain structs.
 * These are the API response types used by the designer services.
 */

export interface FormDefinition {
  frm_id: number;
  frm_key: string;
  frm_name: string;
  frm_description: string;
  frm_status: string;
  frm_created_by: string;
  frm_created_at: string;
  frm_updated_at: string;
}

export interface FormLayout {
  fl_id: number;
  fl_form_id: number;
  fl_name: string;
  fl_display_name: string;
  fl_description: string;
  fl_status: string;
  fl_draft_version_id: number | null;
  fl_published_version_id: number | null;
  fl_created_by: string;
  fl_created_at: string;
  fl_updated_at: string;
}

export interface LayoutVersion {
  flv_id: number;
  flv_layout_id: number;
  flv_version_number: number;
  flv_kind: 'draft' | 'published' | 'archived';
  flv_description: string;
  flv_definition: unknown;
  flv_created_by: string;
  flv_created_at: string;
  flv_published_at: string | null;
}

export interface RoleAssignment {
  fra_role_name: string;
  fra_layout_name: string;
  fra_assigned_at: string;
}

export interface AuditEntry {
  id: number;
  actorUserId: string;
  action: string;
  entityType: string;
  entityId: number;
  metadata: Record<string, unknown>;
  note: string;
  createdAt: string;
}

export interface AuditListResponse {
  items: AuditEntry[];
  total: number;
  page: number;
  pageSize: number;
}

/** Palette item used to create a new field from the designer palette. */
export interface PaletteFieldType {
  type: 'text' | 'textarea' | 'number' | 'date' | 'checkbox' | 'select' | 'radio' | 'multiselect';
  label: string;
  icon: string;
}

export const PALETTE_FIELD_TYPES: PaletteFieldType[] = [
  { type: 'text', label: 'Text', icon: 'T' },
  { type: 'textarea', label: 'Textarea', icon: '¶' },
  { type: 'number', label: 'Number', icon: '#' },
  { type: 'date', label: 'Date', icon: '📅' },
  { type: 'checkbox', label: 'Checkbox', icon: '☑' },
  { type: 'select', label: 'Select', icon: '▼' },
  { type: 'radio', label: 'Radio', icon: '◉' },
  { type: 'multiselect', label: 'Multi-select', icon: '☰' },
];
