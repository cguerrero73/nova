-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20260826000002
-- Description: Seed domain keys for existing fields

UPDATE eamfields SET fld_domain_key = 'id'       WHERE fld_fieldname = 'usr_id';
UPDATE eamfields SET fld_domain_key = 'code'     WHERE fld_fieldname = 'usr_code';
UPDATE eamfields SET fld_domain_key = 'name'     WHERE fld_fieldname = 'usr_name';
UPDATE eamfields SET fld_domain_key = 'email'    WHERE fld_fieldname = 'usr_email';
UPDATE eamfields SET fld_domain_key = 'status'   WHERE fld_fieldname = 'usr_status';

UPDATE eamfields SET fld_domain_key = 'id'          WHERE fld_fieldname = 'par_id';
UPDATE eamfields SET fld_domain_key = 'code'        WHERE fld_fieldname = 'par_code';
UPDATE eamfields SET fld_domain_key = 'description' WHERE fld_fieldname = 'par_desc';
UPDATE eamfields SET fld_domain_key = 'uom'         WHERE fld_fieldname = 'par_uom';
UPDATE eamfields SET fld_domain_key = 'type'        WHERE fld_fieldname = 'par_type';
UPDATE eamfields SET fld_domain_key = 'status'      WHERE fld_fieldname = 'par_status';

UPDATE eamfields SET fld_domain_key = 'id'          WHERE fld_fieldname = 'str_id';
UPDATE eamfields SET fld_domain_key = 'code'        WHERE fld_fieldname = 'str_code';
UPDATE eamfields SET fld_domain_key = 'description' WHERE fld_fieldname = 'str_desc';
UPDATE eamfields SET fld_domain_key = 'organization' WHERE fld_fieldname = 'str_org';

UPDATE eamfields SET fld_domain_key = 'id'          WHERE fld_fieldname = 'obj_id';
UPDATE eamfields SET fld_domain_key = 'code'        WHERE fld_fieldname = 'obj_code';
UPDATE eamfields SET fld_domain_key = 'description' WHERE fld_fieldname = 'obj_desc';
UPDATE eamfields SET fld_domain_key = 'type'        WHERE fld_fieldname = 'obj_type';
UPDATE eamfields SET fld_domain_key = 'status'      WHERE fld_fieldname = 'obj_status';
UPDATE eamfields SET fld_domain_key = 'organization' WHERE fld_fieldname = 'obj_org';

UPDATE eamfields SET fld_domain_key = 'id'          WHERE fld_fieldname = 'evt_id';
UPDATE eamfields SET fld_domain_key = 'code'        WHERE fld_fieldname = 'evt_code';
UPDATE eamfields SET fld_domain_key = 'description' WHERE fld_fieldname = 'evt_desc';
UPDATE eamfields SET fld_domain_key = 'type'        WHERE fld_fieldname = 'evt_type';
UPDATE eamfields SET fld_domain_key = 'status'      WHERE fld_fieldname = 'evt_status';
UPDATE eamfields SET fld_domain_key = 'organization' WHERE fld_fieldname = 'evt_org';

UPDATE eamfields SET fld_domain_key = 'id'          WHERE fld_fieldname = 'sys_id';
UPDATE eamfields SET fld_domain_key = 'type'        WHERE fld_fieldname = 'sys_type';
UPDATE eamfields SET fld_domain_key = 'code'        WHERE fld_fieldname = 'sys_code';
UPDATE eamfields SET fld_domain_key = 'uniqueCode'  WHERE fld_fieldname = 'sys_ucode';
UPDATE eamfields SET fld_domain_key = 'description' WHERE fld_fieldname = 'sys_desc';
UPDATE eamfields SET fld_domain_key = 'system'      WHERE fld_fieldname = 'sys_system';

UPDATE eamfields SET fld_domain_key = 'id'          WHERE fld_fieldname = 'stc_id';
UPDATE eamfields SET fld_domain_key = 'partCode'    WHERE fld_fieldname = 'stc_part_code';
UPDATE eamfields SET fld_domain_key = 'storeCode'   WHERE fld_fieldname = 'stc_store_code';
UPDATE eamfields SET fld_domain_key = 'minStock'    WHERE fld_fieldname = 'stc_min_stock';
UPDATE eamfields SET fld_domain_key = 'actualQty'   WHERE fld_fieldname = 'stc_actual_qty';
