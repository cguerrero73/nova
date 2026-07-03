-- Nova EAM - Normalize Role Permissions
-- Timestamp: 20260218000000
-- Description: Migrate from legacy CRUD columns (rpe_select, rpe_insert, etc.) 
--              to normalized (rpe_screen, rpe_action, rpe_allowed) model.
--              This enables semantic actions like "formbuilder.design" instead of
--              fixed CRUD operations. Also adds ses_active_role to eamsessions.

-- ============================================
-- Step 1: Add ses_active_role to eamsessions
-- ============================================
ALTER TABLE eamsessions ADD COLUMN IF NOT EXISTS ses_active_role VARCHAR(50);
CREATE INDEX IF NOT EXISTS idx_ses_active_role ON eamsessions(ses_active_role);

COMMENT ON COLUMN eamsessions.ses_active_role IS 'Currently active role for the session';

-- ============================================
-- Step 2: Create normalized permissions table
-- ============================================
CREATE TABLE eamrole_permissions_new (
    rpe_id BIGSERIAL PRIMARY KEY,
    rpe_role VARCHAR(50) NOT NULL,
    rpe_screen VARCHAR(100) NOT NULL,
    rpe_action VARCHAR(50) NOT NULL,
    rpe_allowed BOOLEAN NOT NULL DEFAULT false,
    rpe_created_at TIMESTAMP DEFAULT now(),
    rpe_updated_at TIMESTAMP,
    rpe_created_by VARCHAR(50),
    rpe_updated_by VARCHAR(50),
    CONSTRAINT fk_rpe_new_001 FOREIGN KEY (rpe_role) REFERENCES eamroles(rol_code)
);

CREATE UNIQUE INDEX idx_rpe_new_001 ON eamrole_permissions_new(rpe_role, rpe_screen, rpe_action);
CREATE INDEX idx_rpe_new_002 ON eamrole_permissions_new(rpe_role);

COMMENT ON TABLE eamrole_permissions_new IS 'Normalized role permissions (screen, action, allowed)';

-- ============================================
-- Step 3: Migrate existing CRUD data
-- ============================================
-- Unpivot each legacy row into up to 5 normalized rows (one per action)
INSERT INTO eamrole_permissions_new (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_at, rpe_created_by)
SELECT rpe_role, rpe_screen, 'select', (rpe_select = '+'), rpe_created_at, rpe_created_by
FROM eamrole_permissions
WHERE rpe_select IS NOT NULL;

INSERT INTO eamrole_permissions_new (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_at, rpe_created_by)
SELECT rpe_role, rpe_screen, 'insert', (rpe_insert = '+'), rpe_created_at, rpe_created_by
FROM eamrole_permissions
WHERE rpe_insert IS NOT NULL;

INSERT INTO eamrole_permissions_new (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_at, rpe_created_by)
SELECT rpe_role, rpe_screen, 'update', (rpe_update = '+'), rpe_created_at, rpe_created_by
FROM eamrole_permissions
WHERE rpe_update IS NOT NULL;

INSERT INTO eamrole_permissions_new (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_at, rpe_created_by)
SELECT rpe_role, rpe_screen, 'delete', (rpe_delete = '+'), rpe_created_at, rpe_created_by
FROM eamrole_permissions
WHERE rpe_delete IS NOT NULL;

INSERT INTO eamrole_permissions_new (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_at, rpe_created_by)
SELECT rpe_role, rpe_screen, 'print', (rpe_print = '+'), rpe_created_at, rpe_created_by
FROM eamrole_permissions
WHERE rpe_print IS NOT NULL;

-- ============================================
-- Step 4: Drop old table, rename new
-- ============================================
DROP TABLE eamrole_permissions;
ALTER TABLE eamrole_permissions_new RENAME TO eamrole_permissions;
ALTER INDEX idx_rpe_new_001 RENAME TO idx_rpe_001;
ALTER INDEX idx_rpe_new_002 RENAME TO idx_rpe_002;
ALTER SEQUENCE eamrole_permissions_new_rpe_id_seq RENAME TO eamrole_permissions_rpe_id_seq;
ALTER TABLE eamrole_permissions RENAME CONSTRAINT fk_rpe_new_001 TO fk_rpe_001;

-- ============================================
-- Step 5: Seed ADMIN with wildcard permissions
-- ============================================
-- ADMIN gets (*) screen with (*) action = full access
INSERT INTO eamrole_permissions (rpe_role, rpe_screen, rpe_action, rpe_allowed, rpe_created_by)
VALUES ('ADMIN', '*', '*', true, 'SYSTEM')
ON CONFLICT (rpe_role, rpe_screen, rpe_action) DO NOTHING;

-- EMPTY role gets no permissions (explicit no-op, but documented)
-- The EMPTY role exists in eamroles but has no rows in eamrole_permissions.
