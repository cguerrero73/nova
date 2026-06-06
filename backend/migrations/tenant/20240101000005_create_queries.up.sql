-- Nova EAM - Saved Queries Table
-- Timestamp: 20240101000005
-- Description: Creates eamqueries table for saved grid queries

CREATE TABLE eamqueries (
    qry_id VARCHAR(50) PRIMARY KEY,
    qry_grid_id INTEGER NOT NULL,
    qry_name VARCHAR(100) NOT NULL,
    qry_user_id VARCHAR(50),
    qry_is_public BOOLEAN DEFAULT false,
    qry_is_default BOOLEAN DEFAULT false,
    qry_query JSONB NOT NULL DEFAULT '{"fields":[],"sort":[],"filters":[]}',
    qry_created_at TIMESTAMP DEFAULT now(),
    qry_updated_at TIMESTAMP,
    qry_created_by VARCHAR(50),
    qry_updated_by VARCHAR(50)
);

-- Indexes
CREATE UNIQUE INDEX idx_qry_001 ON eamqueries(qry_id);
CREATE INDEX idx_qry_002 ON eamqueries(qry_grid_id);
CREATE INDEX idx_qry_003 ON eamqueries(qry_user_id);
CREATE INDEX idx_qry_004 ON eamqueries(qry_grid_id, qry_user_id);
CREATE INDEX idx_qry_005 ON eamqueries(qry_grid_id, qry_is_public);

-- Comments
COMMENT ON TABLE eamqueries IS 'Saved queries for grid views';
COMMENT ON COLUMN eamqueries.qry_grid_id IS 'Grid ID (1=users, etc)';
COMMENT ON COLUMN eamqueries.qry_is_public IS 'True = visible to all users in tenant';
COMMENT ON COLUMN eamqueries.qry_is_default IS 'True = auto-selected when grid opens';
COMMENT ON COLUMN eamqueries.qry_query IS 'JSON with fields[], sort[], filters[], pagination{}';