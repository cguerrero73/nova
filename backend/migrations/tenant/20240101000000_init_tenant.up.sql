-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20240101000000
-- Description: Creates all tenant-scoped tables

-- ============================================
-- eamorganizations
-- ============================================
CREATE TABLE eamorganizations (
    org_id BIGSERIAL PRIMARY KEY,
    org_code VARCHAR(50) NOT NULL,
    org_name VARCHAR(255) NOT NULL,
    org_common CHAR(1) DEFAULT NULL,
    org_notused CHAR(1) DEFAULT NULL,
    org_created_at TIMESTAMP DEFAULT now(),
    org_updated_at TIMESTAMP,
    org_created_by VARCHAR(50),
    org_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_org_001 ON eamorganizations(org_code);
CREATE INDEX idx_org_002 ON eamorganizations(org_common);

COMMENT ON TABLE eamorganizations IS 'Organizations (factories, business units, locations)';
COMMENT ON COLUMN eamorganizations.org_code IS 'Unique code (* for common org)';
COMMENT ON COLUMN eamorganizations.org_common IS '+ if this is a common org shared across all orgs';

-- ============================================
-- eamsyscodes
-- ============================================
CREATE TABLE eamsyscodes (
    sys_id BIGSERIAL PRIMARY KEY,
    sys_type VARCHAR(50) NOT NULL,
    sys_code VARCHAR(50) NOT NULL,
    sys_ucode VARCHAR(50) NOT NULL,
    sys_desc VARCHAR(255) NOT NULL,
    sys_system CHAR(1) DEFAULT NULL,
    sys_notused CHAR(1) DEFAULT NULL,
    sys_created_at TIMESTAMP DEFAULT now(),
    sys_updated_at TIMESTAMP,
    sys_created_by VARCHAR(50),
    sys_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_sys_001 ON eamsyscodes(sys_type, sys_code);
CREATE INDEX idx_sys_002 ON eamsyscodes(sys_ucode);
CREATE INDEX idx_sys_003 ON eamsyscodes(sys_type);
CREATE INDEX idx_sys_004 ON eamsyscodes(sys_notused);

COMMENT ON TABLE eamsyscodes IS 'System codes catalog (types, statuses)';
COMMENT ON COLUMN eamsyscodes.sys_code IS 'System code for logic (U, A, C, etc)';
COMMENT ON COLUMN eamsyscodes.sys_ucode IS 'User-displayable code (can be customized)';
COMMENT ON COLUMN eamsyscodes.sys_system IS '+ for system codes (not editable by user)';

-- ============================================
-- eamusers
-- ============================================
CREATE TABLE eamusers (
    usr_id BIGSERIAL PRIMARY KEY,
    usr_code VARCHAR(50) NOT NULL,
    usr_name VARCHAR(255) NOT NULL,
    usr_email VARCHAR(255) NOT NULL,
    usr_password VARCHAR(255) NOT NULL,
    usr_phone VARCHAR(50),
    usr_eam VARCHAR(1) DEFAULT '+',
    usr_default_org VARCHAR(50),
    usr_created_at TIMESTAMP DEFAULT now(),
    usr_updated_at TIMESTAMP,
    usr_created_by VARCHAR(50),
    usr_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_usr_001 ON eamusers(usr_code);
CREATE UNIQUE INDEX idx_usr_002 ON eamusers(usr_email);
CREATE INDEX idx_usr_003 ON eamusers(usr_eam);

COMMENT ON TABLE eamusers IS 'User accounts';
COMMENT ON COLUMN eamusers.usr_password IS 'BCrypt hashed password';

-- ============================================
-- eamroles
-- ============================================
CREATE TABLE eamroles (
    rol_id BIGSERIAL PRIMARY KEY,
    rol_code VARCHAR(50) NOT NULL,
    rol_desc VARCHAR(255) NOT NULL,
    rol_system CHAR(1),
    rol_notused CHAR(1) DEFAULT '-',
    rol_created_at TIMESTAMP DEFAULT now(),
    rol_updated_at TIMESTAMP,
    rol_created_by VARCHAR(50),
    rol_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_rol_001 ON eamroles(rol_code);

COMMENT ON TABLE eamroles IS 'Role definitions for permissions';

-- ============================================
-- eamuser_organizations
-- ============================================
CREATE TABLE eamuser_organizations (
    uog_id BIGSERIAL PRIMARY KEY,
    uog_user VARCHAR(50) NOT NULL,
    uog_org VARCHAR(50) NOT NULL,
    uog_role VARCHAR(50) NOT NULL,
    uog_default VARCHAR(1),
    uog_created_at TIMESTAMP DEFAULT now(),
    uog_updated_at TIMESTAMP,
    uog_created_by VARCHAR(50),
    uog_updated_by VARCHAR(50),
    CONSTRAINT fk_uog_001 FOREIGN KEY (uog_user) REFERENCES eamusers(usr_code),
    CONSTRAINT fk_uog_002 FOREIGN KEY (uog_org) REFERENCES eamorganizations(org_code),
    CONSTRAINT fk_uog_003 FOREIGN KEY (uog_role) REFERENCES eamroles(rol_code)
);

CREATE UNIQUE INDEX idx_uog_001 ON eamuser_organizations(uog_user, uog_org);
CREATE INDEX idx_uog_002 ON eamuser_organizations(uog_user);
CREATE INDEX idx_uog_003 ON eamuser_organizations(uog_org);

COMMENT ON TABLE eamuser_organizations IS 'Many-to-many relationship between users and organizations';

-- ============================================
-- eamrole_permissions
-- ============================================
CREATE TABLE eamrole_permissions (
    rpe_id BIGSERIAL PRIMARY KEY,
    rpe_role VARCHAR(50) NOT NULL,
    rpe_screen VARCHAR(100) NOT NULL,
    rpe_select VARCHAR(1) NOT NULL,
    rpe_insert VARCHAR(1) NOT NULL,
    rpe_update VARCHAR(1) NOT NULL,
    rpe_delete VARCHAR(1) NOT NULL,
    rpe_print VARCHAR(1) NOT NULL,
    rpe_created_at TIMESTAMP DEFAULT now(),
    rpe_updated_at TIMESTAMP,
    rpe_created_by VARCHAR(50),
    rpe_updated_by VARCHAR(50),
    CONSTRAINT fk_rpe_001 FOREIGN KEY (rpe_role) REFERENCES eamroles(rol_code)
);

CREATE UNIQUE INDEX idx_rpe_001 ON eamrole_permissions(rpe_role, rpe_screen);

COMMENT ON TABLE eamrole_permissions IS 'Detailed permissions per role';


-- ============================================
-- eamobjects
-- ============================================
CREATE TABLE eamobjects (
    obj_id BIGSERIAL PRIMARY KEY,
    obj_code VARCHAR(50) NOT NULL,
    obj_type VARCHAR(50),
    obj_desc VARCHAR(255),
    obj_serial VARCHAR(100),
    obj_status VARCHAR(50),
    obj_org VARCHAR(50) NOT NULL,
    obj_parent VARCHAR(50),
    obj_parent_org VARCHAR(50),
    obj_install_date DATE,
    obj_notused CHAR(1) DEFAULT NULL,
    obj_created_at TIMESTAMP DEFAULT now(),
    obj_updated_at TIMESTAMP,
    obj_created_by VARCHAR(50),
    obj_updated_by VARCHAR(50),
    CONSTRAINT fk_obj_001 FOREIGN KEY (obj_org) REFERENCES eamorganizations(org_code)
);

CREATE UNIQUE INDEX idx_obj_001 ON eamobjects(obj_code, obj_org);
CREATE INDEX idx_obj_002 ON eamobjects(obj_type);
CREATE INDEX idx_obj_003 ON eamobjects(obj_status);
CREATE INDEX idx_obj_004 ON eamobjects(obj_org);
CREATE INDEX idx_obj_005 ON eamobjects(obj_parent, obj_parent_org);
CREATE INDEX idx_obj_006 ON eamobjects(obj_notused);

COMMENT ON TABLE eamobjects IS 'Assets/Equipment within the organization';
COMMENT ON COLUMN eamobjects.obj_type IS 'References eamsyscodes.sys_code where sys_type=OBTP';
COMMENT ON COLUMN eamobjects.obj_status IS 'References eamsyscodes.sys_code where sys_type=OBST';

-- ============================================
-- eamstructure
-- ============================================
CREATE TABLE eamstructure (
    sct_id BIGSERIAL PRIMARY KEY,
    sct_parent_code VARCHAR(50) NOT NULL,
    sct_parent_org VARCHAR(50) NOT NULL,
    sct_child_code VARCHAR(50) NOT NULL,
    sct_child_org VARCHAR(50) NOT NULL,
    sct_cost CHAR(1) DEFAULT NULL,
    sct_meter CHAR(1) DEFAULT NULL,
    sct_created_at TIMESTAMP DEFAULT now(),
    sct_updated_at TIMESTAMP,
    sct_created_by VARCHAR(50),
    sct_updated_by VARCHAR(50),
    CONSTRAINT fk_sct_001 FOREIGN KEY (sct_parent_code, sct_parent_org) REFERENCES eamobjects(obj_code, obj_org),
    CONSTRAINT fk_sct_002 FOREIGN KEY (sct_child_code, sct_child_org) REFERENCES eamobjects(obj_code, obj_org)
);

CREATE UNIQUE INDEX idx_sct_001 ON eamstructure(sct_parent_code, sct_parent_org, sct_child_code, sct_child_org);
CREATE INDEX idx_sct_002 ON eamstructure(sct_parent_code, sct_parent_org);
CREATE INDEX idx_sct_003 ON eamstructure(sct_child_code, sct_child_org);

COMMENT ON TABLE eamstructure IS 'Parent-child relationships between objects (hierarchy)';
COMMENT ON COLUMN eamstructure.sct_cost IS '+ if costs flow up the hierarchy';
COMMENT ON COLUMN eamstructure.sct_meter IS '+ if meters flow down the hierarchy';

-- ============================================
-- eamstores
-- ============================================
CREATE TABLE eamstores (
    str_id BIGSERIAL PRIMARY KEY,
    str_code VARCHAR(50) NOT NULL,
    str_org VARCHAR(50) NOT NULL,
    str_desc VARCHAR(255) NOT NULL,
    str_notused CHAR(1) DEFAULT NULL,
    str_created_at TIMESTAMP DEFAULT now(),
    str_updated_at TIMESTAMP,
    str_created_by VARCHAR(50),
    str_updated_by VARCHAR(50),
    CONSTRAINT fk_str_001 FOREIGN KEY (str_org) REFERENCES eamorganizations(org_code)
);

CREATE UNIQUE INDEX idx_str_001 ON eamstores(str_code, str_org);
CREATE UNIQUE INDEX idx_str_002 ON eamstores(str_code);

-- ============================================
-- eambins
-- ============================================
CREATE TABLE eambins (
    bin_id BIGSERIAL PRIMARY KEY,
    bin_store VARCHAR(50) NOT NULL,
    bin_code VARCHAR(50) NOT NULL,
    bin_desc VARCHAR(255) NOT NULL,
    bin_notused CHAR(1) DEFAULT NULL,
    bin_created_at TIMESTAMP DEFAULT now(),
    bin_updated_at TIMESTAMP,
    bin_created_by VARCHAR(50),
    bin_updated_by VARCHAR(50),
    CONSTRAINT fk_bin_001 FOREIGN KEY (bin_store) REFERENCES eamstores(str_code)
);

CREATE UNIQUE INDEX idx_bin_001 ON eambins(bin_store, bin_code);
CREATE INDEX idx_bin_002 ON eambins(bin_notused);

-- ============================================
-- eamuoms
-- ============================================
CREATE TABLE eamuoms (
    uom_id BIGSERIAL PRIMARY KEY,
    uom_code VARCHAR(50) NOT NULL,
    uom_desc VARCHAR(255) NOT NULL,
    uom_notused CHAR(1) DEFAULT NULL,
    uom_created_at TIMESTAMP DEFAULT now(),
    uom_updated_at TIMESTAMP,
    uom_created_by VARCHAR(50),
    uom_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_uom_001 ON eamuoms(uom_code);

-- ============================================
-- eamparts
-- ============================================
CREATE TABLE eamparts (
    par_id BIGSERIAL PRIMARY KEY,
    par_org VARCHAR(50) NOT NULL,
    par_code VARCHAR(50) NOT NULL,
    par_desc VARCHAR(255) NOT NULL,
    par_uom VARCHAR(20) NOT NULL,
    par_notused CHAR(1) DEFAULT NULL,
    par_created_at TIMESTAMP DEFAULT now(),
    par_updated_at TIMESTAMP,
    par_created_by VARCHAR(50),
    par_updated_by VARCHAR(50),
    CONSTRAINT fk_par_001 FOREIGN KEY (par_org) REFERENCES eamorganizations(org_code),
    CONSTRAINT fk_par_002 FOREIGN KEY (par_uom) REFERENCES eamuoms(uom_code)
);

CREATE UNIQUE INDEX idx_par_001 ON eamparts(par_org, par_code);
CREATE INDEX idx_par_002 ON eamparts(par_org);
CREATE INDEX idx_par_003 ON eamparts(par_notused);

COMMENT ON TABLE eamparts IS 'Parts/Items in inventory';

-- ============================================
-- eamstocks
-- ============================================
CREATE TABLE eamstocks (
    sto_id BIGSERIAL PRIMARY KEY,
    sto_store VARCHAR(50) NOT NULL,
    sto_part VARCHAR(50) NOT NULL,
    sto_part_org VARCHAR(50) NOT NULL,
    sto_min_stock DECIMAL(24,6) DEFAULT 0,
    sto_reorder_qty DECIMAL(24,6) DEFAULT 0,
    sto_actual_qty DECIMAL(24,6) DEFAULT 0,
    sto_base_price DECIMAL(24,6) DEFAULT 0,
    sto_avg_price DECIMAL(24,6) DEFAULT 0,
    sto_last_price DECIMAL(24,6) DEFAULT 0,
    sto_std_price DECIMAL(24,6) DEFAULT 0,
    sto_notused CHAR(1) DEFAULT NULL,
    sto_created_at TIMESTAMP DEFAULT now(),
    sto_updated_at TIMESTAMP,
    sto_created_by VARCHAR(50),
    sto_updated_by VARCHAR(50),
    CONSTRAINT fk_sto_001 FOREIGN KEY (sto_store) REFERENCES eamstores(str_code),
    CONSTRAINT fk_sto_002 FOREIGN KEY (sto_part, sto_part_org) REFERENCES eamparts(par_code, par_org)
);

CREATE UNIQUE INDEX idx_sto_001 ON eamstocks(sto_store, sto_part, sto_part_org);
CREATE INDEX idx_sto_002 ON eamstocks(sto_part, sto_part_org);
CREATE INDEX idx_sto_003 ON eamstocks(sto_store);

-- ============================================
-- eambin_stocks
-- ============================================
CREATE TABLE eambin_stocks (
    bis_id BIGSERIAL PRIMARY KEY,
    bis_store VARCHAR(50) NOT NULL,
    bis_part VARCHAR(50) NOT NULL,
    bis_part_org VARCHAR(50) NOT NULL,
    bis_bin VARCHAR(50) NOT NULL,
    bis_quantity DECIMAL(24, 6) DEFAULT 0,
    bis_created_at TIMESTAMP DEFAULT now(),
    bis_updated_at TIMESTAMP,
    bis_created_by VARCHAR(50),
    bis_updated_by VARCHAR(50),
    CONSTRAINT fk_bis_001 FOREIGN KEY (bis_store, bis_bin) REFERENCES eambins(bin_store, bin_code),
    CONSTRAINT fk_bis_002 FOREIGN KEY (bis_part, bis_part_org) REFERENCES eamparts(par_code, par_org)
);

CREATE UNIQUE INDEX idx_bis_001 ON eambin_stocks(bis_part, bis_part_org, bis_store, bis_bin);
CREATE INDEX idx_bis_002 ON eambin_stocks(bis_part, bis_part_org, bis_store);
CREATE INDEX idx_bis_003 ON eambin_stocks(bis_bin);

-- ============================================
-- eamevents
-- ============================================
CREATE TABLE eamevents (
    evt_id BIGSERIAL PRIMARY KEY,
    evt_org VARCHAR(50) NOT NULL,
    evt_code VARCHAR(50) NOT NULL,
    evt_desc VARCHAR(255) NOT NULL,
    evt_type VARCHAR(50),
    evt_rtype VARCHAR(50),
    evt_status VARCHAR(50),
    evt_rstatus VARCHAR(50),
    evt_object VARCHAR(50),
    evt_object_org VARCHAR(50),
    evt_notused CHAR(1) DEFAULT NULL,
    evt_created_at TIMESTAMP DEFAULT now(),
    evt_updated_at TIMESTAMP,
    evt_created_by VARCHAR(50),
    evt_updated_by VARCHAR(50),
    CONSTRAINT fk_evt_001 FOREIGN KEY (evt_org) REFERENCES eamorganizations(org_code),
    CONSTRAINT fk_evt_002 FOREIGN KEY (evt_object, evt_object_org) REFERENCES eamobjects(obj_code, obj_org) 
);

CREATE UNIQUE INDEX idx_evt_001 ON eamevents(evt_code, evt_org);
CREATE INDEX idx_evt_002 ON eamevents(evt_org);
CREATE INDEX idx_evt_003 ON eamevents(evt_type);
CREATE INDEX idx_evt_004 ON eamevents(evt_status);
CREATE INDEX idx_evt_005 ON eamevents(evt_object, evt_object_org);

COMMENT ON TABLE eamevents IS 'Events (work orders, activities) associated with objects';

-- ============================================
-- eamsessions
-- ============================================
CREATE TABLE eamsessions (
    ses_id BIGSERIAL PRIMARY KEY,
    ses_user VARCHAR(50) NOT NULL,
    ses_expires_at TIMESTAMP NOT NULL,
    ses_ip_address VARCHAR(45),
    ses_created_at TIMESTAMP DEFAULT now(),
    ses_updated_at TIMESTAMP,
    ses_created_by VARCHAR(50),
    ses_updated_by VARCHAR(50)
);

CREATE INDEX idx_ses_001 ON eamsessions(ses_user);

COMMENT ON TABLE eamsessions IS 'User sessions for JWT refresh tokens';