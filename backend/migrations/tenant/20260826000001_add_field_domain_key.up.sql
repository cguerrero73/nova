-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20260826000001
-- Description: Add domain key mapping to field catalog

ALTER TABLE eamfields
    ADD COLUMN fld_domain_key VARCHAR(80);

COMMENT ON COLUMN eamfields.fld_domain_key IS 'Canonical domain key exposed to clients (e.g. id, name, email). Falls back to fld_fieldname when null.';
