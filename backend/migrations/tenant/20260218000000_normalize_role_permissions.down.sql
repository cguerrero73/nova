-- Nova EAM - Rollback: Normalize Role Permissions
-- Timestamp: 20260218000000
-- Description: Revert from normalized (screen, action, allowed) back to legacy CRUD columns

-- ============================================
-- Step 1: Remove ADMIN wildcard seed
-- ============================================
DELETE FROM eamrole_permissions WHERE rpe_role = 'ADMIN' AND rpe_screen = '*' AND rpe_action = '*';

-- ============================================
-- Step 2: Recreate legacy table structure
-- ============================================
CREATE TABLE eamrole_permissions_legacy (
    rpe_id BIGSERIAL PRIMARY KEY,
    rpe_role VARCHAR(50) NOT NULL,
    rpe_screen VARCHAR(100) NOT NULL,
    rpe_select VARCHAR(1) NOT NULL DEFAULT '-',
    rpe_insert VARCHAR(1) NOT NULL DEFAULT '-',
    rpe_update VARCHAR(1) NOT NULL DEFAULT '-',
    rpe_delete VARCHAR(1) NOT NULL DEFAULT '-',
    rpe_print VARCHAR(1) NOT NULL DEFAULT '-',
    rpe_created_at TIMESTAMP DEFAULT now(),
    rpe_updated_at TIMESTAMP,
    rpe_created_by VARCHAR(50),
    rpe_updated_by VARCHAR(50),
    CONSTRAINT fk_rpe_legacy_001 FOREIGN KEY (rpe_role) REFERENCES eamroles(rol_code)
);

CREATE UNIQUE INDEX idx_rpe_legacy_001 ON eamrole_permissions_legacy(rpe_role, rpe_screen);

-- ============================================
-- Step 3: Pivot normalized data back to CRUD columns
-- ============================================
-- Group by (role, screen) and pivot actions into columns
INSERT INTO eamrole_permissions_legacy (rpe_role, rpe_screen, rpe_select, rpe_insert, rpe_update, rpe_delete, rpe_print, rpe_created_at, rpe_created_by)
SELECT 
    rpe_role,
    rpe_screen,
    COALESCE(MAX(CASE WHEN rpe_action = 'select' AND rpe_allowed THEN '+' ELSE '-' END), '-'),
    COALESCE(MAX(CASE WHEN rpe_action = 'insert' AND rpe_allowed THEN '+' ELSE '-' END), '-'),
    COALESCE(MAX(CASE WHEN rpe_action = 'update' AND rpe_allowed THEN '+' ELSE '-' END), '-'),
    COALESCE(MAX(CASE WHEN rpe_action = 'delete' AND rpe_allowed THEN '+' ELSE '-' END), '-'),
    COALESCE(MAX(CASE WHEN rpe_action = 'print' AND rpe_allowed THEN '+' ELSE '-' END), '-'),
    MIN(rpe_created_at),
    MIN(rpe_created_by)
FROM eamrole_permissions
GROUP BY rpe_role, rpe_screen;

-- ============================================
-- Step 4: Drop normalized table, rename legacy
-- ============================================
DROP TABLE eamrole_permissions;
ALTER TABLE eamrole_permissions_legacy RENAME TO eamrole_permissions;
ALTER INDEX idx_rpe_legacy_001 RENAME TO idx_rpe_001;
ALTER SEQUENCE eamrole_permissions_legacy_rpe_id_seq RENAME TO eamrole_permissions_rpe_id_seq;
ALTER TABLE eamrole_permissions RENAME CONSTRAINT fk_rpe_legacy_001 TO fk_rpe_001;

-- ============================================
-- Step 5: Remove ses_active_role from eamsessions
-- ============================================
DROP INDEX IF EXISTS idx_ses_active_role;
ALTER TABLE eamsessions DROP COLUMN IF EXISTS ses_active_role;
