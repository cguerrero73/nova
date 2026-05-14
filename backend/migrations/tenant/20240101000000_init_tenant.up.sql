-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20240101000000
-- Description: Creates all tenant-scoped tables

-- Note: uuid-ossp extension is created in global migration (public schema)

-- ============================================
-- eamorganizations
-- ============================================
CREATE TABLE eamorganizations (
    org_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_code VARCHAR(50) NOT NULL,
    org_name VARCHAR(255) NOT NULL,
    org_common CHAR(1) DEFAULT NULL,
    org_notused CHAR(1) DEFAULT NULL,
    org_tenant_id UUID NOT NULL,
    org_created_at TIMESTAMP DEFAULT now(),
    org_updated_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_org_001 ON eamorganizations(org_code, org_tenant_id);
CREATE INDEX idx_org_002 ON eamorganizations(org_tenant_id);
CREATE INDEX idx_org_003 ON eamorganizations(org_common);

COMMENT ON TABLE eamorganizations IS 'Organizations within a tenant (factories, business units, locations)';
COMMENT ON COLUMN eamorganizations.org_code IS 'Unique code within tenant (* for common org)';
COMMENT ON COLUMN eamorganizations.org_common IS '+ if this is a common org shared across all orgs';

-- ============================================
-- eamuser_organizations
-- ============================================
CREATE TABLE eamuser_organizations (
    uog_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uog_user VARCHAR(50) NOT NULL,
    uog_org VARCHAR(50) NOT NULL,
    uog_is_default BOOLEAN DEFAULT false,
    uog_tenant_id UUID NOT NULL,
    uog_assigned_at TIMESTAMP DEFAULT now(),
    uog_assigned_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_uog_001 ON eamuser_organizations(uog_user, uog_org);
CREATE INDEX idx_uog_002 ON eamuser_organizations(uog_user);
CREATE INDEX idx_uog_003 ON eamuser_organizations(uog_org);

COMMENT ON TABLE eamuser_organizations IS 'Many-to-many relationship between users and organizations';

-- ============================================
-- eamusers
-- ============================================
CREATE TABLE eamusers (
    usr_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usr_code VARCHAR(50) NOT NULL,
    usr_name VARCHAR(255) NOT NULL,
    usr_email VARCHAR(255) NOT NULL,
    usr_password VARCHAR(255) NOT NULL,
    usr_phone VARCHAR(50),
    usr_status VARCHAR(20) DEFAULT 'active',
    usr_default_org VARCHAR(50),
    usr_notused CHAR(1) DEFAULT NULL,
    usr_tenant_id UUID NOT NULL,
    usr_created_at TIMESTAMP DEFAULT now(),
    usr_updated_at TIMESTAMP,
    usr_created_by VARCHAR(50),
    usr_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_usr_001 ON eamusers(usr_code, usr_tenant_id);
CREATE UNIQUE INDEX idx_usr_002 ON eamusers(usr_email, usr_tenant_id);
CREATE INDEX idx_usr_003 ON eamusers(usr_tenant_id);
CREATE INDEX idx_usr_004 ON eamusers(usr_status);
CREATE INDEX idx_usr_005 ON eamusers(usr_notused);

COMMENT ON TABLE eamusers IS 'User accounts within a tenant';
COMMENT ON COLUMN eamusers.usr_password IS 'BCrypt hashed password';

-- ============================================
-- eamroles
-- ============================================
CREATE TABLE eamroles (
    rol_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rol_name VARCHAR(100) NOT NULL,
    rol_desc TEXT,
    rol_is_system BOOLEAN DEFAULT false,
    rol_permissions JSONB DEFAULT '{}',
    rol_notused CHAR(1) DEFAULT NULL,
    rol_tenant_id UUID NOT NULL,
    rol_created_at TIMESTAMP DEFAULT now(),
    rol_updated_at TIMESTAMP
);

CREATE INDEX idx_rol_001 ON eamroles(rol_tenant_id);
CREATE INDEX idx_rol_002 ON eamroles(rol_is_system);
CREATE INDEX idx_rol_003 ON eamroles(rol_notused);

COMMENT ON TABLE eamroles IS 'Role definitions with permissions';
COMMENT ON COLUMN eamroles.rol_permissions IS 'JSON object with screen/action permissions';

-- ============================================
-- eamrole_permissions
-- ============================================
CREATE TABLE eamrole_permissions (
    rpe_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rpe_role_id UUID NOT NULL,
    rpe_screen VARCHAR(100) NOT NULL,
    rpe_action VARCHAR(50) NOT NULL,
    rpe_allowed BOOLEAN DEFAULT false,
    rpe_tenant_id UUID NOT NULL,
    rpe_created_at TIMESTAMP DEFAULT now(),
    CONSTRAINT fk_rpe_001 FOREIGN KEY (rpe_role_id) REFERENCES eamroles(rol_id)
);

CREATE INDEX idx_rpe_001 ON eamrole_permissions(rpe_role_id);
CREATE INDEX idx_rpe_002 ON eamrole_permissions(rpe_tenant_id);

COMMENT ON TABLE eamrole_permissions IS 'Detailed permissions per role (alternative to JSON)';

-- ============================================
-- eamuser_roles
-- ============================================
CREATE TABLE eamuser_roles (
    urr_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    urr_user_id UUID NOT NULL,
    urr_role_id UUID NOT NULL,
    urr_tenant_id UUID NOT NULL,
    urr_assigned_at TIMESTAMP DEFAULT now(),
    urr_assigned_by VARCHAR(50),
    CONSTRAINT fk_urr_001 FOREIGN KEY (urr_user_id) REFERENCES eamusers(usr_id),
    CONSTRAINT fk_urr_002 FOREIGN KEY (urr_role_id) REFERENCES eamroles(rol_id)
);

CREATE UNIQUE INDEX idx_urr_001 ON eamuser_roles(urr_user_id, urr_role_id, urr_tenant_id);
CREATE INDEX idx_urr_002 ON eamuser_roles(urr_user_id);
CREATE INDEX idx_urr_003 ON eamuser_roles(urr_role_id);

-- ============================================
-- eamobjects
-- ============================================
CREATE TABLE eamobjects (
    obj_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    obj_code VARCHAR(50) NOT NULL,
    obj_type VARCHAR(50),
    obj_desc TEXT,
    obj_serial VARCHAR(100),
    obj_status VARCHAR(50),
    obj_org VARCHAR(50) NOT NULL,
    obj_parent_code VARCHAR(50),
    obj_parent_org VARCHAR(50),
    obj_install_date DATE,
    obj_notused CHAR(1) DEFAULT NULL,
    obj_tenant_id UUID NOT NULL,
    obj_created_at TIMESTAMP DEFAULT now(),
    obj_updated_at TIMESTAMP,
    obj_created_by VARCHAR(50),
    obj_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_obj_001 ON eamobjects(obj_code, obj_tenant_id);
CREATE INDEX idx_obj_002 ON eamobjects(obj_type);
CREATE INDEX idx_obj_003 ON eamobjects(obj_status);
CREATE INDEX idx_obj_004 ON eamobjects(obj_org);
CREATE INDEX idx_obj_005 ON eamobjects(obj_parent_code, obj_parent_org);
CREATE INDEX idx_obj_006 ON eamobjects(obj_notused);

COMMENT ON TABLE eamobjects IS 'Assets/Equipment within the organization';
COMMENT ON COLUMN eamobjects.obj_type IS 'References eamsyscodes.sys_code where sys_type=OBTP';
COMMENT ON COLUMN eamobjects.obj_status IS 'References eamsyscodes.sys_code where sys_type=OBST';

-- ============================================
-- eamstructure
-- ============================================
CREATE TABLE eamstructure (
    sct_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sct_parent_code VARCHAR(50) NOT NULL,
    sct_parent_org VARCHAR(50) NOT NULL,
    sct_child_code VARCHAR(50) NOT NULL,
    sct_child_org VARCHAR(50) NOT NULL,
    sct_cost CHAR(1) DEFAULT NULL,
    sct_meter CHAR(1) DEFAULT NULL,
    sct_tenant_id UUID NOT NULL,
    sct_created_at TIMESTAMP DEFAULT now(),
    sct_updated_at TIMESTAMP
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
    str_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    str_code VARCHAR(50) NOT NULL,
    str_name VARCHAR(255) NOT NULL,
    str_desc TEXT,
    str_org VARCHAR(50) NOT NULL,
    str_notused CHAR(1) DEFAULT NULL,
    str_tenant_id UUID NOT NULL,
    str_created_at TIMESTAMP DEFAULT now(),
    str_updated_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_str_001 ON eamstores(str_code, str_tenant_id);
CREATE INDEX idx_str_002 ON eamstores(str_org);
CREATE INDEX idx_str_003 ON eamstores(str_notused);

-- ============================================
-- eambins
-- ============================================
CREATE TABLE eambins (
    bin_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bin_code VARCHAR(50) NOT NULL,
    bin_desc TEXT,
    bin_org VARCHAR(50) NOT NULL,
    bin_notused CHAR(1) DEFAULT NULL,
    bin_tenant_id UUID NOT NULL,
    bin_created_at TIMESTAMP DEFAULT now(),
    bin_updated_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_bin_001 ON eambins(bin_code, bin_org, bin_tenant_id);
CREATE INDEX idx_bin_002 ON eambins(bin_org);
CREATE INDEX idx_bin_003 ON eambins(bin_notused);

-- ============================================
-- eamparts
-- ============================================
CREATE TABLE eamparts (
    par_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    par_code VARCHAR(50) NOT NULL,
    par_desc TEXT,
    par_notused CHAR(1) DEFAULT NULL,
    par_org VARCHAR(50) NOT NULL,
    par_tenant_id UUID NOT NULL,
    par_created_at TIMESTAMP DEFAULT now(),
    par_updated_at TIMESTAMP,
    par_created_by VARCHAR(50),
    par_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_par_001 ON eamparts(par_code, par_tenant_id);
CREATE INDEX idx_par_002 ON eamparts(par_org);
CREATE INDEX idx_par_003 ON eamparts(par_notused);

COMMENT ON TABLE eamparts IS 'Parts/Items in inventory';

-- ============================================
-- eamstocks
-- ============================================
CREATE TABLE eamstocks (
    stc_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stc_part_code VARCHAR(50) NOT NULL,
    stc_part_org VARCHAR(50) NOT NULL,
    stc_store_code VARCHAR(50) NOT NULL,
    stc_store_org VARCHAR(50) NOT NULL,
    stc_min_stock DECIMAL(10,2) DEFAULT 0,
    stc_reorder_qty DECIMAL(10,2) DEFAULT 0,
    stc_actual_qty DECIMAL(10,2) DEFAULT 0,
    stc_notused CHAR(1) DEFAULT NULL,
    stc_tenant_id UUID NOT NULL,
    stc_created_at TIMESTAMP DEFAULT now(),
    stc_updated_at TIMESTAMP,
    stc_created_by VARCHAR(50),
    stc_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_stc_001 ON eamstocks(stc_part_code, stc_part_org, stc_store_code, stc_store_org);
CREATE INDEX idx_stc_002 ON eamstocks(stc_part_code, stc_part_org);
CREATE INDEX idx_stc_003 ON eamstocks(stc_store_code, stc_store_org);
CREATE INDEX idx_stc_004 ON eamstocks(stc_notused);

-- ============================================
-- eambin_stocks
-- ============================================
CREATE TABLE eambin_stocks (
    bis_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bis_part_code VARCHAR(50) NOT NULL,
    bis_part_org VARCHAR(50) NOT NULL,
    bis_store_code VARCHAR(50) NOT NULL,
    bis_store_org VARCHAR(50) NOT NULL,
    bis_bin_code VARCHAR(50) NOT NULL,
    bis_bin_org VARCHAR(50) NOT NULL,
    bis_quantity DECIMAL(10,2) DEFAULT 0,
    bis_tenant_id UUID NOT NULL,
    bis_created_at TIMESTAMP DEFAULT now(),
    bis_updated_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_bis_001 ON eambin_stocks(bis_part_code, bis_part_org, bis_store_code, bis_store_org, bis_bin_code, bis_bin_org);
CREATE INDEX idx_bis_002 ON eambin_stocks(bis_part_code, bis_part_org, bis_store_code, bis_store_org);
CREATE INDEX idx_bis_003 ON eambin_stocks(bis_bin_code, bis_bin_org);

-- ============================================
-- eamevents
-- ============================================
CREATE TABLE eamevents (
    evt_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    evt_code VARCHAR(50) NOT NULL,
    evt_org VARCHAR(50) NOT NULL,
    evt_desc TEXT,
    evt_type VARCHAR(50),
    evt_rtype VARCHAR(50),
    evt_status VARCHAR(50),
    evt_rstatus VARCHAR(50),
    evt_object VARCHAR(50),
    evt_object_org VARCHAR(50),
    evt_notused CHAR(1) DEFAULT NULL,
    evt_tenant_id UUID NOT NULL,
    evt_created_at TIMESTAMP DEFAULT now(),
    evt_updated_at TIMESTAMP,
    evt_created_by VARCHAR(50),
    evt_updated_by VARCHAR(50)
);

CREATE UNIQUE INDEX idx_evt_001 ON eamevents(evt_code, evt_tenant_id);
CREATE INDEX idx_evt_002 ON eamevents(evt_org);
CREATE INDEX idx_evt_003 ON eamevents(evt_type);
CREATE INDEX idx_evt_004 ON eamevents(evt_status);
CREATE INDEX idx_evt_005 ON eamevents(evt_object, evt_object_org);
CREATE INDEX idx_evt_006 ON eamevents(evt_notused);

COMMENT ON TABLE eamevents IS 'Events (work orders, activities) associated with objects';

-- ============================================
-- eamsyscodes
-- ============================================
CREATE TABLE eamsyscodes (
    sys_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sys_type VARCHAR(50) NOT NULL,
    sys_code VARCHAR(50) NOT NULL,
    sys_ucode VARCHAR(50) NOT NULL,
    sys_desc TEXT,
    sys_system CHAR(1) DEFAULT NULL,
    sys_notused CHAR(1) DEFAULT NULL,
    sys_created_at TIMESTAMP DEFAULT now(),
    sys_updated_at TIMESTAMP
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
-- eamsessions
-- ============================================
CREATE TABLE eamsessions (
    ses_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ses_user_code VARCHAR(50) NOT NULL,
    ses_refresh_token VARCHAR(255) UNIQUE NOT NULL,
    ses_expires_at TIMESTAMP NOT NULL,
    ses_ip_address VARCHAR(45),
    ses_user_agent VARCHAR(255),
    ses_created_at TIMESTAMP DEFAULT now(),
    ses_revoked_at TIMESTAMP
);

CREATE INDEX idx_ses_001 ON eamsessions(ses_user_code);
CREATE INDEX idx_ses_002 ON eamsessions(ses_refresh_token);
CREATE INDEX idx_ses_003 ON eamsessions(ses_expires_at);

COMMENT ON TABLE eamsessions IS 'User sessions for JWT refresh tokens';