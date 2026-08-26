-- Nova EAM - Tenant Schema Migration
-- Timestamp: 20260826000003
-- Description: Remove user status system codes

DELETE FROM eamsyscodes WHERE sys_type = 'USST';
