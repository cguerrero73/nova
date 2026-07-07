-- Nova EAM - Seed User Form Definition
-- Timestamp: 20260219000005
-- Description: Creates the "user-form" form definition with a default layout
--              containing name, email, and status fields for the user management screen.

-- ============================================
-- Create form definition
-- ============================================
INSERT INTO eamform_definitions (frm_key, frm_name, frm_description, frm_status, frm_created_by)
VALUES (
    'user-form',
    'User Management Form',
    'Dynamic form for creating and editing users',
    'active',
    'SYSTEM'
)
ON CONFLICT (frm_key) DO NOTHING;

-- ============================================
-- Create default layout
-- ============================================
INSERT INTO eamform_layouts (fl_form_id, fl_name, fl_display_name, fl_description, fl_status, fl_created_by)
SELECT 
    frm_id,
    'default',
    'Default Layout',
    'Standard layout for user management',
    'active',
    'SYSTEM'
FROM eamform_definitions
WHERE frm_key = 'user-form'
ON CONFLICT (fl_form_id, fl_name) DO NOTHING;

-- ============================================
-- Create published version with layout definition
-- ============================================
INSERT INTO eamform_layout_versions (flv_layout_id, flv_version_number, flv_kind, flv_description, flv_definition, flv_created_by, flv_published_at)
SELECT 
    fl.fl_id,
    1,
    'published',
    'Initial version with name, email, and status fields',
    '{
        "formKey": "user-form",
        "layoutName": "default",
        "sections": [
            {
                "name": "basic-info",
                "title": "Información Básica",
                "order": 1,
                "fields": [
                    {
                        "name": "name",
                        "type": "text",
                        "ui": {
                            "label": "Nombre",
                            "placeholder": "Ingrese el nombre del usuario",
                            "width": "full"
                        },
                        "validators": [
                            { "kind": "required" },
                            { "kind": "minLength", "value": 2 }
                        ]
                    },
                    {
                        "name": "email",
                        "type": "text",
                        "ui": {
                            "label": "Email",
                            "placeholder": "usuario@ejemplo.com",
                            "width": "full"
                        },
                        "validators": [
                            { "kind": "required" },
                            { "kind": "email" }
                        ]
                    },
                    {
                        "name": "status",
                        "type": "select",
                        "ui": {
                            "label": "Estado",
                            "width": "full"
                        },
                        "options": [
                            { "label": "Activo", "value": "active" },
                            { "label": "Inactivo", "value": "inactive" }
                        ],
                        "validators": [
                            { "kind": "required" }
                        ]
                    }
                ]
            }
        ],
        "rules": []
    }'::jsonb,
    'SYSTEM',
    now()
FROM eamform_layouts fl
INNER JOIN eamform_definitions fd ON fl.fl_form_id = fd.frm_id
WHERE fd.frm_key = 'user-form' AND fl.fl_name = 'default'
AND NOT EXISTS (
    SELECT 1 FROM eamform_layout_versions
    WHERE flv_layout_id = fl.fl_id AND flv_kind = 'published'
);

-- ============================================
-- Update layout with published version pointer
-- ============================================
UPDATE eamform_layouts
SET fl_published_version_id = sub.flv_id
FROM (
    SELECT flv_id, flv_layout_id
    FROM eamform_layout_versions
    WHERE flv_kind = 'published' AND flv_version_number = 1
) sub
WHERE fl_id = sub.flv_layout_id
  AND fl_form_id = (SELECT frm_id FROM eamform_definitions WHERE frm_key = 'user-form')
  AND fl_name = 'default'
  AND fl_published_version_id IS NULL;
