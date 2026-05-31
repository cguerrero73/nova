-- Nova EAM - Rollback session column alignment
-- Timestamp: 20240101000004
-- Description: Reverts session table changes

DROP INDEX IF EXISTS idx_ses_refresh_token;

ALTER TABLE eamsessions DROP COLUMN IF EXISTS ses_revoked_at;
ALTER TABLE eamsessions DROP COLUMN IF EXISTS ses_user_agent;
ALTER TABLE eamsessions DROP COLUMN IF EXISTS ses_refresh_token;

ALTER TABLE eamsessions RENAME COLUMN ses_user_code TO ses_user;
