-- Nova EAM - Seed Default Roles and Organizations
-- Timestamp: 20240101000002
-- Description: Populates default roles and common organization
-- Note: Data migration - no rollback (reference data)

-- ============================================
-- Default Roles
-- ============================================
INSERT INTO eamroles (rol_code, rol_desc, rol_system) VALUES
('ADMIN', 'Administrator with full access to all modules', '+'),
('EMPTY', 'No permissions assigned', '+');

-- ============================================
-- Default Common Organization (*)
-- ============================================
INSERT INTO eamorganizations (org_code, org_name, org_common) VALUES
('*', 'Common Organization', '+');