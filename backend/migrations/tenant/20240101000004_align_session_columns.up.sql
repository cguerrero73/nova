-- Nova EAM - Align eamsessions columns with Go entity
-- Timestamp: 20240101000004
-- Description: Adds missing columns for refresh tokens, user agent, revocation

-- ============================================
-- Step 1: Rename ses_user to ses_user_code
-- ============================================
ALTER TABLE eamsessions RENAME COLUMN ses_user TO ses_user_code;

-- ============================================
-- Step 2: Add missing columns
-- ============================================
ALTER TABLE eamsessions ADD COLUMN IF NOT EXISTS ses_refresh_token VARCHAR(255);
ALTER TABLE eamsessions ADD COLUMN IF NOT EXISTS ses_user_agent VARCHAR(255);
ALTER TABLE eamsessions ADD COLUMN IF NOT EXISTS ses_revoked_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_ses_refresh_token ON eamsessions(ses_refresh_token);
