-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20260827000001
-- Description: Add user code field to user-form detail layout

WITH user_layout AS (
    SELECT l.fl_id
    FROM eamform_layouts l
    INNER JOIN eamform_definitions d ON l.fl_form_id = d.frm_id
    WHERE d.frm_key = 'user-form'
),
max_version AS (
    SELECT COALESCE(MAX(v.flv_version_number), 0) + 1 AS next_version
    FROM eamform_layout_versions v
    WHERE v.flv_layout_id = (SELECT fl_id FROM user_layout)
),
new_version AS (
    INSERT INTO eamform_layout_versions (
        flv_layout_id,
        flv_version_number,
        flv_kind,
        flv_description,
        flv_definition,
        flv_created_by,
        flv_published_at
    )
    SELECT
        (SELECT fl_id FROM user_layout),
        (SELECT next_version FROM max_version),
        'published',
        'Added user code field to the form',
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
                            "name": "code",
                            "type": "text",
                            "ui": {
                                "label": "Código",
                                "placeholder": "Ingrese el código del usuario",
                                "width": "full"
                            },
                            "validators": [
                                { "kind": "required" },
                                { "kind": "minLength", "value": 2 }
                            ]
                        },
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
                            "dataSource": {
                                "type": "syscodes",
                                "codeType": "USST"
                            },
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
        NOW()
    FROM eamform_layout_versions v
    WHERE v.flv_layout_id = (SELECT fl_id FROM user_layout)
      AND v.flv_kind = 'published'
    ORDER BY v.flv_version_number DESC
    LIMIT 1
    RETURNING flv_id
)
UPDATE eamform_layouts
SET fl_published_version_id = (SELECT flv_id FROM new_version)
WHERE fl_id = (SELECT fl_id FROM user_layout);
