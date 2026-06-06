-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20240101000010
-- Description: Fix NULL timestamps in eamfields

UPDATE eamfields SET fld_created_at = NOW(), fld_updated_at = NOW() WHERE fld_created_at IS NULL OR fld_updated_at IS NULL;
