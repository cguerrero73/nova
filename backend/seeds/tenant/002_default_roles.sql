-- Nova EAM - Default Roles Seed
-- Version: 001

-- ============================================
-- Default Roles
-- ============================================
INSERT INTO eamroles (rol_name, rol_desc, rol_is_system, rol_permissions) VALUES
('ADMIN', 'Administrator with full access to all modules', true, '{"*": {"*": true}}'),
('EMPTY', 'No permissions assigned', true, '{}'),
('SUPERVISOR', 'Supervisor - can view and edit all but not delete', true, '{"*": {"list": true, "view": true, "create": true, "edit": true}}'),
('OPERATOR', 'Operator - can view and update work orders', true, '{"work_orders": {"list": true, "view": true, "edit": true}, "objects": {"list": true, "view": true}}');

-- ============================================
-- Default Common Organization (*)
-- ============================================
INSERT INTO eamorganizations (org_code, org_name, org_common) VALUES
('*', 'Common Organization', '+');