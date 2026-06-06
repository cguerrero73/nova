-- Nova EAM - Tenant Schema Seed
-- Timestamp: 20240101000008
-- Description: Seed initial grid definitions

-- ============================================
-- Grids
-- ============================================
INSERT INTO eamgrids (grd_name, grd_desc, grd_base_query, grd_key_fields, grd_filterable_list, grd_sortable_list, grd_displayable_list, grd_hints, grd_type)
VALUES 
    ('BMUSER', 'User Management', 'FROM eamusers', '1', '1,2,3,4,5', '1,2,3,4,5', '1,2,3,4,5', null, 1),
    ('SMPART', 'Parts Management', 'FROM eamparts', '6', '6,7,8,9,10,11', '6,7,8,9,10,11', '6,7,8,9,10,11', null, 1),
    ('SMSTOR', 'Stores Management', 'FROM eamstores', '12', '12,13,14,15', '12,13,14,15', '12,13,14,15', null, 1),
    ('OMOBJA', 'Objects Management', 'FROM eamobjects', '16', '16,17,18,19,20,21', '16,17,18,19,20,21', '16,17,18,19,20,21', null, 1),
    ('WMJOBS', 'Work Orders', 'FROM eamevents', '22', '22,23,24,25,26,27', '22,23,24,25,26,27', '22,23,24,25,26,27', null, 1),
    ('BCCODE', 'System Codes', 'FROM eamsyscodes', '28', '28,29,30,31,32,33', '28,29,30,31,32,33', '28,29,30,31,32,33', null, 1),
    ('SMSTOC', 'Stocks Management', 'FROM eamstocks', '34', '34,35,36,37,38', '34,35,36,37,38', '34,35,36,37,38', null, 1)
ON CONFLICT (grd_name) DO NOTHING;
