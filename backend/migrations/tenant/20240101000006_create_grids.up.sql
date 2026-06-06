-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20240101000006
-- Description: Creates grids configuration table

-- ============================================
-- eamgrids
-- ============================================
CREATE TABLE eamgrids (
    grd_id              SERIAL PRIMARY KEY,
    grd_name            VARCHAR(60) NOT NULL UNIQUE,
    grd_desc            VARCHAR(120),
    
    grd_base_query      TEXT NOT NULL,
    
    grd_key_fields      VARCHAR(500),
    grd_filterable_list VARCHAR(1000),
    grd_sortable_list   VARCHAR(1000),
    grd_displayable_list VARCHAR(2000),
    
    grd_org_column   VARCHAR(50),
    grd_bot_function VARCHAR(50),
    grd_sec_entity   VARCHAR(50),
    grd_hints        VARCHAR(255),
    grd_type         SMALLINT DEFAULT 2,
    
    grd_created_at TIMESTAMPTZ DEFAULT NOW(),
    grd_updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_grd_001 ON eamgrids(grd_name);

COMMENT ON TABLE eamgrids IS 'Grid definitions with base queries and field configuration';
COMMENT ON COLUMN eamgrids.grd_base_query IS 'FROM clause with joins, conditions, security filters';
COMMENT ON COLUMN eamgrids.grd_key_fields IS 'Comma-separated field IDs for row identification';
COMMENT ON COLUMN eamgrids.grd_filterable_list IS 'Comma-separated field IDs that can be used in filters';
COMMENT ON COLUMN eamgrids.grd_sortable_list IS 'Comma-separated field IDs that can be sorted';
COMMENT ON COLUMN eamgrids.grd_displayable_list IS 'Comma-separated field IDs visible to user';
COMMENT ON COLUMN eamgrids.grd_hints IS 'SQL hints (e.g. top 101)';
