-- Nova EAM - Global Schema Migration Down
-- Timestamp: 20240101000000
-- Description: Rollback global tables

DROP TABLE IF EXISTS eamtenant_customers;
DROP TABLE IF EXISTS eamtenants;
DROP EXTENSION IF EXISTS "uuid-ossp";