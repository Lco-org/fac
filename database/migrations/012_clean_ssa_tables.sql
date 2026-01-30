-- =====================================================
-- Migration 012: Clean Duplicate SSA Tables from Tenants
-- Purpose: Remove SSA-only tables that should not exist in tenant databases
-- These tables belong ONLY in erp_master
-- =====================================================

-- Drop SSA tables from tenant databases (if they exist)
-- These tables are for Super System Admin only and should be in erp_master

DROP TABLE IF EXISTS domain_verifications CASCADE;
DROP TABLE IF EXISTS allowed_domains CASCADE;
DROP TABLE IF EXISTS super_admin_audit CASCADE;
DROP TABLE IF EXISTS super_system_admins CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;
DROP TABLE IF EXISTS org_deletion_requests CASCADE;

-- Note: 'organizations' table should NOT be dropped as some tenants might 
-- legitimately have their own 'organizations' table for different purposes.
-- If you need to drop it too, uncomment the line below:
-- DROP TABLE IF EXISTS organizations CASCADE;
