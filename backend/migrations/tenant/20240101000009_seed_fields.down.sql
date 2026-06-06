-- Nova EAM - Tenant Schema Seed Down
-- Timestamp: 20240101000009
-- Description: Remove seed field definitions

DELETE FROM eamfields WHERE fld_tablename IN (
    'eamusers', 'eamparts', 'eamstores', 'eamobjects', 'eamevents', 'eamsyscodes', 'eamstocks'
);