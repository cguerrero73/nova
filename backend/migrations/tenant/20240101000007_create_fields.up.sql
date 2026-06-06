-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20240101000007
-- Description: Creates fields catalog table

-- ============================================
-- eamfields
-- ============================================
CREATE TABLE eamfields (
    fld_id          SERIAL PRIMARY KEY,
    fld_fieldname   VARCHAR(80) NOT NULL,
    fld_datatype    VARCHAR(20) NOT NULL,
    fld_tablename   VARCHAR(80) NOT NULL,
    
    fld_created_at  TIMESTAMPTZ DEFAULT NOW(),
    fld_updated_at  TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(fld_tablename, fld_fieldname)
);

CREATE INDEX idx_fld_001 ON eamfields(fld_tablename);

COMMENT ON TABLE eamfields IS 'Field definitions for grids - maps field IDs to actual table columns';
COMMENT ON COLUMN eamfields.fld_fieldname IS 'Actual column name in database (e.g. usr_code)';
COMMENT ON COLUMN eamfields.fld_datatype IS 'Data type: string, number, date, boolean, select';
COMMENT ON COLUMN eamfields.fld_tablename IS 'Table name this field belongs to (e.g. eamusers)';
