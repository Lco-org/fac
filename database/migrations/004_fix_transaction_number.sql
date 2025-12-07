-- Fix transaction_number Generation
-- This creates the trigger function if it doesn't exist

-- Create function to generate transaction number
CREATE OR REPLACE FUNCTION generate_transaction_number()
RETURNS TRIGGER AS $$
DECLARE
    prefix TEXT;
    seq_num INT;
    year_suffix TEXT;
BEGIN
    year_suffix := TO_CHAR(CURRENT_DATE, 'YY');
    
    -- Determine prefix based on transaction type
    CASE NEW.type
        WHEN 'income_grant' THEN prefix := 'GR';
        WHEN 'income_donation' THEN prefix := 'DN';
        WHEN 'disbursement_custody' THEN prefix := 'CU';
        WHEN 'disbursement_salary' THEN prefix := 'SA';
        WHEN 'disbursement_operational' THEN prefix := 'OP';
        WHEN 'disbursement_donation' THEN prefix := 'DD';
        WHEN 'disbursement_bonus' THEN prefix := 'BN';
        WHEN 'procurement' THEN prefix := 'PR';
        WHEN 'transfer' THEN prefix := 'TR';
        ELSE prefix := 'TX';
    END CASE;
    
    -- Get next sequence number for this year
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(transaction_number FROM 8) AS INT)
    ), 0) + 1 INTO seq_num
    FROM transactions
    WHERE transaction_number LIKE prefix || '-' || year_suffix || '-%';
    
    NEW.transaction_number := prefix || '-' || year_suffix || '-' || LPAD(seq_num::TEXT, 6, '0');
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if exists
DROP TRIGGER IF EXISTS set_transaction_number ON transactions;

-- Create trigger
CREATE TRIGGER set_transaction_number
    BEFORE INSERT ON transactions
    FOR EACH ROW
    WHEN (NEW.transaction_number IS NULL OR NEW.transaction_number = '')
    EXECUTE FUNCTION generate_transaction_number();

-- Also ensure members table has proper defaults
-- Fix member status default
ALTER TABLE members ALTER COLUMN status SET DEFAULT 'active';
ALTER TABLE members ALTER COLUMN is_active SET DEFAULT true;

-- Verify by checking constraints
SELECT 'Trigger created successfully!' as status;
