-- ============================================================================
-- CREATE SUPER SYSTEM ADMIN USER MIGRATION
-- Run this on the erp_master database to create the first SSA user
-- ============================================================================

-- IMPORTANT: Replace the password hash with an actual Argon2id hash for your password
-- You can generate it using the backend's security.HashPassword function

-- Example: Insert a Super System Admin
INSERT INTO super_system_admins (
    id,
    username,
    email,
    password_hash,
    full_name,
    is_active,
    two_factor_enabled,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    'superadmin',
    'superadmin@lco.org.ly',
    -- This is a placeholder hash. REPLACE WITH REAL Argon2id HASH!
    -- Use: security.HashPassword("your-password", nil) from Go code
    '$argon2id$v=19$m=131072,t=4,p=4$PLACEHOLDER$PLACEHOLDER',
    'Super System Administrator',
    true,
    false,  -- Start with 2FA disabled, enable after first login
    NOW(),
    NOW()
) ON CONFLICT (username) DO NOTHING;

-- Verify insertion
SELECT id, username, email, full_name, is_active, two_factor_enabled 
FROM super_system_admins;

-- ============================================================================
-- QUICK SSA USER CREATION SCRIPT (For Development Only)
-- ============================================================================
-- For development/testing, you can use this simplified approach:
-- 
-- Step 1: In your Go code, run this temporarily:
-- ```go
-- hash, _ := security.HashPassword("Admin@123456", nil)
-- fmt.Println(hash)
-- ```
--
-- Step 2: Copy the output hash and update the INSERT statement above
--
-- Step 3: Run the INSERT on erp_master database
-- ============================================================================
