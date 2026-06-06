-- Nova EAM - Tenant Schema Seed
-- Timestamp: 20240101000009
-- Description: Seed field definitions for grids

-- ============================================
-- eamusers fields
-- ============================================
INSERT INTO eamfields (fld_id, fld_fieldname, fld_datatype, fld_tablename) VALUES
    (1, 'usr_id', 'VARCHAR', 'eamusers'),
    (2, 'usr_code', 'VARCHAR', 'eamusers'),
    (3, 'usr_name', 'VARCHAR', 'eamusers'),
    (4, 'usr_email', 'VARCHAR', 'eamusers'),
    (5, 'usr_status', 'VARCHAR', 'eamusers')
ON CONFLICT (fld_tablename, fld_fieldname) DO NOTHING;

-- ============================================
-- eamparts fields
-- ============================================
INSERT INTO eamfields (fld_id, fld_fieldname, fld_datatype, fld_tablename) VALUES
    (6, 'par_id', 'VARCHAR', 'eamparts'),
    (7, 'par_code', 'VARCHAR', 'eamparts'),
    (8, 'par_desc', 'VARCHAR', 'eamparts'),
    (9, 'par_uom', 'VARCHAR', 'eamparts'),
    (10, 'par_type', 'VARCHAR', 'eamparts'),
    (11, 'par_status', 'VARCHAR', 'eamparts')
ON CONFLICT (fld_tablename, fld_fieldname) DO NOTHING;

-- ============================================
-- eamstores fields
-- ============================================
INSERT INTO eamfields (fld_id, fld_fieldname, fld_datatype, fld_tablename) VALUES
    (12, 'str_id', 'VARCHAR', 'eamstores'),
    (13, 'str_code', 'VARCHAR', 'eamstores'),
    (14, 'str_desc', 'VARCHAR', 'eamstores'),
    (15, 'str_org', 'VARCHAR', 'eamstores')
ON CONFLICT (fld_tablename, fld_fieldname) DO NOTHING;

-- ============================================
-- eamobjects fields
-- ============================================
INSERT INTO eamfields (fld_id, fld_fieldname, fld_datatype, fld_tablename) VALUES
    (16, 'obj_id', 'VARCHAR', 'eamobjects'),
    (17, 'obj_code', 'VARCHAR', 'eamobjects'),
    (18, 'obj_desc', 'VARCHAR', 'eamobjects'),
    (19, 'obj_type', 'VARCHAR', 'eamobjects'),
    (20, 'obj_status', 'VARCHAR', 'eamobjects'),
    (21, 'obj_org', 'VARCHAR', 'eamobjects')
ON CONFLICT (fld_tablename, fld_fieldname) DO NOTHING;

-- ============================================
-- eamevents fields
-- ============================================
INSERT INTO eamfields (fld_id, fld_fieldname, fld_datatype, fld_tablename) VALUES
    (22, 'evt_id', 'VARCHAR', 'eamevents'),
    (23, 'evt_code', 'VARCHAR', 'eamevents'),
    (24, 'evt_desc', 'VARCHAR', 'eamevents'),
    (25, 'evt_type', 'VARCHAR', 'eamevents'),
    (26, 'evt_status', 'VARCHAR', 'eamevents'),
    (27, 'evt_org', 'VARCHAR', 'eamevents')
ON CONFLICT (fld_tablename, fld_fieldname) DO NOTHING;

-- ============================================
-- eamsyscodes fields
-- ============================================
INSERT INTO eamfields (fld_id, fld_fieldname, fld_datatype, fld_tablename) VALUES
    (28, 'sys_id', 'VARCHAR', 'eamsyscodes'),
    (29, 'sys_type', 'VARCHAR', 'eamsyscodes'),
    (30, 'sys_code', 'VARCHAR', 'eamsyscodes'),
    (31, 'sys_ucode', 'VARCHAR', 'eamsyscodes'),
    (32, 'sys_desc', 'VARCHAR', 'eamsyscodes'),
    (33, 'sys_system', 'VARCHAR', 'eamsyscodes')
ON CONFLICT (fld_tablename, fld_fieldname) DO NOTHING;

-- ============================================
-- eamstocks fields
-- ============================================
INSERT INTO eamfields (fld_id, fld_fieldname, fld_datatype, fld_tablename) VALUES
    (34, 'stc_id', 'VARCHAR', 'eamstocks'),
    (35, 'stc_part_code', 'VARCHAR', 'eamstocks'),
    (36, 'stc_store_code', 'VARCHAR', 'eamstocks'),
    (37, 'stc_min_stock', 'NUMBER', 'eamstocks'),
    (38, 'stc_actual_qty', 'NUMBER', 'eamstocks')
ON CONFLICT (fld_tablename, fld_fieldname) DO NOTHING;