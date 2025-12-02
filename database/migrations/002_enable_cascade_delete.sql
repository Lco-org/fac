-- Migration: Enable CASCADE DELETE for financial containers
-- This allows deleting a container along with all its transactions
-- WARNING: This is a destructive operation - all transaction history will be lost

-- Step 1: Drop existing foreign key constraint
ALTER TABLE transactions 
DROP CONSTRAINT IF EXISTS transactions_container_id_fkey;

-- Step 2: Re-add foreign key with CASCADE DELETE
ALTER TABLE transactions 
ADD CONSTRAINT transactions_container_id_fkey 
FOREIGN KEY (container_id) 
REFERENCES financial_containers(id) 
ON DELETE CASCADE;

-- Verification: Check the new constraint
SELECT 
    tc.constraint_name, 
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name,
    rc.delete_rule
FROM information_schema.table_constraints AS tc 
JOIN information_schema.key_column_usage AS kcu
  ON tc.constraint_name = kcu.constraint_name
  AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
  ON ccu.constraint_name = tc.constraint_name
  AND ccu.table_schema = tc.table_schema
JOIN information_schema.referential_constraints AS rc
  ON rc.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY' 
  AND tc.table_name = 'transactions'
  AND kcu.column_name = 'container_id';
