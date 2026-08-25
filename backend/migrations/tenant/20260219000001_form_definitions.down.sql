-- Nova EAM - Dynamic Form Builder: Rollback Core Tables
-- Timestamp: 20260219000001

-- Drop triggers first
DROP TRIGGER IF EXISTS trg_protect_default_layout ON eamform_definitions;
DROP TRIGGER IF EXISTS trg_no_delete_audit ON eamform_audit_log;
DROP TRIGGER IF EXISTS trg_no_update_audit ON eamform_audit_log;
DROP TRIGGER IF EXISTS trg_no_update_versions ON eamform_layout_versions;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS eamform_audit_log;
DROP TABLE IF EXISTS eamform_role_assignments;

-- Drop foreign keys from layouts before dropping versions
ALTER TABLE eamform_layouts DROP CONSTRAINT IF EXISTS fk_fl_draft_version;
ALTER TABLE eamform_layouts DROP CONSTRAINT IF EXISTS fk_fl_published_version;

DROP TABLE IF EXISTS eamform_layout_versions;
DROP TABLE IF EXISTS eamform_layouts;
DROP TABLE IF EXISTS eamform_definitions;

-- Drop helper functions
DROP FUNCTION IF EXISTS raise_immutable();
DROP FUNCTION IF EXISTS check_default_layout_exists();
