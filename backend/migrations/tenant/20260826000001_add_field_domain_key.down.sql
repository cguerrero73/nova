-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20260826000001
-- Description: Remove domain key mapping from field catalog

ALTER TABLE eamfields
    DROP COLUMN IF EXISTS fld_domain_key;
