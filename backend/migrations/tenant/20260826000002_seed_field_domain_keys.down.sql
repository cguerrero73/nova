-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20260826000002
-- Description: Clear seeded domain keys

UPDATE eamfields SET fld_domain_key = NULL;
