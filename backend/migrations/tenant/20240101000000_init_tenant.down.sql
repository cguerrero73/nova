-- Nova EAM - Tenant Schema Migration Down
-- Timestamp: 20240101000000
-- Description: Rollback all tenant tables

DROP TABLE IF EXISTS eamsessions;
DROP TABLE IF EXISTS eamsyscodes;
DROP TABLE IF EXISTS eamevents;
DROP TABLE IF EXISTS eambin_stocks;
DROP TABLE IF EXISTS eamstocks;
DROP TABLE IF EXISTS eamparts;
DROP TABLE IF EXISTS eambins;
DROP TABLE IF EXISTS eamstores;
DROP TABLE IF EXISTS eamstructure;
DROP TABLE IF EXISTS eamobjects;
DROP TABLE IF EXISTS eamuser_roles;
DROP TABLE IF EXISTS eamrole_permissions;
DROP TABLE IF EXISTS eamroles;
DROP TABLE IF EXISTS eamusers;
DROP TABLE IF EXISTS eamuser_organizations;
DROP TABLE IF EXISTS eamuoms;
DROP TABLE IF EXISTS eamorganizations;
