-- Nova EAM - Seed Form Designer Role
-- Timestamp: 20260219000003
-- Description: Creates the form_designer role with formbuilder permissions.

-- ============================================
-- Create form_designer role
-- ============================================
INSERT INTO eamroles (rol_code, rol_desc, rol_system, rol_created_at, rol_created_by)
VALUES ('form_designer', 'Dynamic form builder designer', '-', now(), 'SYSTEM')
ON CONFLICT (rol_code) DO NOTHING;

-- ============================================
-- Seed permissions for form_designer
-- ============================================
-- view: runtime resolution (read published layouts)
INSERT INTO eamrole_permissions (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_by)
VALUES ('form_designer', 'formbuilder', 'view', true, 'SYSTEM')
ON CONFLICT (rpe_role, rpe_screen, rpe_action) DO NOTHING;

-- view_draft: read draft versions
INSERT INTO eamrole_permissions (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_by)
VALUES ('form_designer', 'formbuilder', 'view_draft', true, 'SYSTEM')
ON CONFLICT (rpe_role, rpe_screen, rpe_action) DO NOTHING;

-- design: create forms, layouts, save drafts
INSERT INTO eamrole_permissions (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_by)
VALUES ('form_designer', 'formbuilder', 'design', true, 'SYSTEM')
ON CONFLICT (rpe_role, rpe_screen, rpe_action) DO NOTHING;

-- publish: publish drafts, archive forms/layouts
INSERT INTO eamrole_permissions (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_by)
VALUES ('form_designer', 'formbuilder', 'publish', true, 'SYSTEM')
ON CONFLICT (rpe_role, rpe_screen, rpe_action) DO NOTHING;

-- assign: manage role-to-layout assignments
INSERT INTO eamrole_permissions (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_by)
VALUES ('form_designer', 'formbuilder', 'assign', true, 'SYSTEM')
ON CONFLICT (rpe_role, rpe_screen, rpe_action) DO NOTHING;

-- ============================================
-- Also grant ADMIN full formbuilder access via wildcard
-- ============================================
INSERT INTO eamrole_permissions (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_by)
VALUES ('ADMIN', 'formbuilder', '*', true, 'SYSTEM')
ON CONFLICT (rpe_role, rpe_screen, rpe_action) DO NOTHING;
