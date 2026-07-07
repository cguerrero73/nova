-- Nova EAM - Rollback User Form Seed
-- Timestamp: 20260219000005

-- Remove published version pointer
UPDATE eamform_layouts
SET fl_published_version_id = NULL
WHERE fl_form_id = (SELECT frm_id FROM eamform_definitions WHERE frm_key = 'user-form')
  AND fl_name = 'default';

-- Remove published version
DELETE FROM eamform_layout_versions
WHERE flv_layout_id = (
    SELECT fl_id FROM eamform_layouts
    WHERE fl_form_id = (SELECT frm_id FROM eamform_definitions WHERE frm_key = 'user-form')
      AND fl_name = 'default'
);

-- Remove default layout
DELETE FROM eamform_layouts
WHERE fl_form_id = (SELECT frm_id FROM eamform_definitions WHERE frm_key = 'user-form')
  AND fl_name = 'default';

-- Remove form definition
DELETE FROM eamform_definitions
WHERE frm_key = 'user-form';
