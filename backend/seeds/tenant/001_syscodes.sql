-- Nova EAM - System Codes Seed
-- Version: 001

-- ============================================
-- Object Types (OBTP)
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('OBTP', 'MAC', 'MAC', 'Machine', '+'),
('OBTP', 'VEH', 'VEH', 'Vehicle', '+'),
('OBTP', 'SYS', 'SYS', 'System', '+'),
('OBTP', 'EQP', 'EQP', 'Equipment', '+'),
('OBTP', 'INS', 'INS', 'Instrument', '+'),
('OBTP', 'TOOL', 'TOOL', 'Tool', '+');

-- ============================================
-- Object Status (OBST)
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('OBST', 'ACT', 'ACT', 'Active', '+'),
('OBST', 'INACT', 'INACT', 'Inactive', '+'),
('OBST', 'MAINT', 'MAINT', 'Under Maintenance', '+'),
('OBST', 'RET', 'RET', 'Retired', '+'),
('OBST', 'SCRAP', 'SCRAP', 'Scrapped', '+');

-- ============================================
-- Event Types (JBTP) - Job Types
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('JBTP', 'PREV', 'PREV', 'Preventive Maintenance', '+'),
('JBTP', 'CORR', 'CORR', 'Corrective Maintenance', '+'),
('JBTP', 'INSP', 'INSP', 'Inspection', '+'),
('JBTP', 'CAL', 'CAL', 'Calibration', '+'),
('JBTP', 'INST', 'INST', 'Installation', '+'),
('JBTP', 'REM', 'REM', 'Removal', '+');

-- ============================================
-- Event Status (JBST) - Job Status
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('JBST', 'OP', 'OP', 'Open', '+'),
('JBST', 'WIP', 'WIP', 'Work In Progress', '+'),
('JBST', 'COMP', 'COMP', 'Completed', '+'),
('JBST', 'CL', 'CL', 'Closed', '+'),
('JBST', 'CANC', 'CANC', 'Cancelled', '+'),
('JBST', 'ONH', 'ONH', 'On Hold', '+');

-- ============================================
-- Event Priority
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('JBPR', 'U', 'U', 'Urgent', '+'),
('JBPR', 'H', 'H', 'High', '+'),
('JBPR', 'M', 'M', 'Medium', '+'),
('JBPR', 'L', 'L', 'Low', '+');

-- ============================================
-- User Status
-- ============================================
INSERT INTO eamsyscodes (sys_type, sys_code, sys_ucode, sys_desc, sys_system) VALUES
('USST', 'ACT', 'ACT', 'Active', '+'),
('USST', 'INACT', 'INACT', 'Inactive', '+'),
('USST', 'LOCK', 'LOCK', 'Locked', '+');
