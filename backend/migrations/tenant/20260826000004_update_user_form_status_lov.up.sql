-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20260826000004
-- Description: Update user-form status field to use DB-backed LOV
-- Note: published versions are immutable, so we create a new published version.

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
        'Status field now uses USST syscodes LOV',
        jsonb_set(
            v.flv_definition,
            '{sections,0,fields,2}',
            '{
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
            }'::jsonb
        ),
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
