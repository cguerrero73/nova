-- Nova EAM - Rollback user column alignment
-- Timestamp: 20240101000003
-- Description: Reverts usr_status back to usr_eam, removes added columns

-- ============================================
-- Step 1: Drop added columns
-- ============================================
ALTER TABLE eamusers DROP COLUMN IF EXISTS usr_notused;
ALTER TABLE eamusers DROP COLUMN IF EXISTS usr_tenant_id;

-- ============================================
-- Step 2: Rename usr_status back to usr_eam
-- Restore values: 'ACT'/'INA' → '+'/NULL
-- ============================================
UPDATE eamusers SET usr_status = '+' WHERE usr_status = 'ACT';
UPDATE eamusers SET usr_status = NULL WHERE usr_status = 'INA';
UPDATE eamusers SET usr_status = '+' WHERE usr_status IS NOT NULL;

ALTER TABLE eamusers RENAME COLUMN usr_status TO usr_eam;

ALTER TABLE eamusers ALTER COLUMN usr_eam SET DEFAULT '+';
