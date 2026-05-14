-- Nova EAM - Global Schema Migration
-- Timestamp: 20240101000000
-- Description: Creates global tables (tenants, tenant_customers)

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- eamtenants
-- ============================================
CREATE TABLE eamtenants (
    ten_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ten_code VARCHAR(50) UNIQUE NOT NULL,
    ten_name VARCHAR(255) NOT NULL,
    ten_is_active BOOLEAN DEFAULT true,
    ten_created_at TIMESTAMP DEFAULT now(),
    ten_updated_at TIMESTAMP
);

CREATE INDEX idx_ten_001 ON eamtenants(ten_code);

COMMENT ON TABLE eamtenants IS 'Global tenant catalog - each row represents a tenant organization';
COMMENT ON COLUMN eamtenants.ten_code IS 'Unique tenant identifier code';
COMMENT ON COLUMN eamtenants.ten_is_active IS 'Whether the tenant is active';

-- ============================================
-- eamtenant_customers
-- ============================================
CREATE TABLE eamtenant_customers (
    tcu_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tcu_code VARCHAR(50) UNIQUE NOT NULL,
    tcu_name VARCHAR(255) NOT NULL,
    tcu_tax_id VARCHAR(50),
    tcu_address TEXT,
    tcu_phone VARCHAR(50),
    tcu_email VARCHAR(255),
    tcu_created_at TIMESTAMP DEFAULT now(),
    tcu_updated_at TIMESTAMP
);

CREATE INDEX idx_tcu_001 ON eamtenant_customers(tcu_code);

COMMENT ON TABLE eamtenant_customers IS 'Customer information linked to tenants';
COMMENT ON COLUMN eamtenant_customers.tcu_tax_id IS 'Tax identification number (CUIT/RUC/etc)';