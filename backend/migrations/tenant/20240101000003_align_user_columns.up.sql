-- Nova EAM - Align eamusers columns with Go entity
-- Timestamp: 20240101000003
-- Description: Renames usr_eam to usr_status, adds missing columns

-- ============================================
-- Step 1: Rename usr_eam to usr_status
-- ============================================
ALTER TABLE eamusers RENAME COLUMN usr_eam TO usr_status;

-- Widen column from VARCHAR(1) to VARCHAR(50) to store 'ACT'/'INA' values
ALTER TABLE eamusers ALTER COLUMN usr_status TYPE VARCHAR(50);

-- Update values to match entity convention (ACT/INA)
UPDATE eamusers SET usr_status = 'ACT' WHERE usr_status = '+';
UPDATE eamusers SET usr_status = 'INA' WHERE usr_status IS NULL OR usr_status = '';
UPDATE eamusers SET usr_status = 'INA' WHERE usr_status NOT IN ('ACT', 'INA', 'SUS');

ALTER TABLE eamusers ALTER COLUMN usr_status SET DEFAULT 'ACT';

-- ============================================
-- Step 2: Add usr_tenant_id column
-- ============================================
ALTER TABLE eamusers ADD COLUMN IF NOT EXISTS usr_tenant_id VARCHAR(50);

-- ============================================
-- Step 3: Add usr_notused column (matches other EAM tables)
-- ============================================
ALTER TABLE eamusers ADD COLUMN IF NOT EXISTS usr_notused CHAR(1) DEFAULT NULL;
