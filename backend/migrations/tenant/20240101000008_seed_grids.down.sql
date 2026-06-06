-- Nova EAM - Tenant Schema Seed Down
-- Timestamp: 20240101000008
-- Description: Remove seed grid definitions

DELETE FROM eamgrids WHERE grd_name IN ('BMUSER', 'SMPART', 'SMSTOR', 'OMOBJA', 'WMJOBS', 'BCCODE', 'SMSTOC');
