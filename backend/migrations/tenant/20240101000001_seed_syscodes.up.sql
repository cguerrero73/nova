-- Nova EAM - Seed System Codes
-- Timestamp: 20240101000001
-- Description: Populates system codes catalog (types, statuses)
-- Note: Data migration - no rollback (reference data)

-- ============================================
-- Object Types (OBTP)
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('OBTP', 'A', 'A', 'Activo', '+'),
('OBTP', 'S', 'S', 'Posición', '+'),
('OBTP', 'P', 'P', 'Sistema', '+');

-- ============================================
-- Object Status (OBST)
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('OBST', 'I', 'I', 'Instalado', '+'),
('OBST', 'D', 'D', 'Baja', '+'),
('OBST', 'C', 'C', 'En Alamcén', '+');

-- ============================================
-- Event Types (JBTP) - Job Types
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('JBTP', 'PM', 'PM', 'Preventive Maintenance', '+'),
('JBTP', 'BRKD0', 'BRKD', 'Corrective Maintenance', '+');
-- ============================================
-- Event Status (JBST) - Job Status
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('JBST', 'R', 'R', 'En Proceso', '+'),
('JBST', 'Q', 'Q', 'Solicitud', '+'),
('JBST', 'A', 'A', 'Esperando liberación', '+'),
('JBST', 'B', 'B', 'Omitido por anidación', '+'),
('JBST', 'C', 'C', 'Cerrado', '+');
