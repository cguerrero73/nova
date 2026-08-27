-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20260826000003
-- Description: Seed user status system codes for LOV support

INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
    ('USST', 'ACT', 'ACT', 'Activo', '+'),
    ('USST', 'INA', 'INA', 'Inactivo', '+'),
    ('USST', 'SUS', 'SUS', 'Suspendido', '+')
ON CONFLICT (sys_type, sys_code) DO NOTHING;
