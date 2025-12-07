-- Add permissions column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions TEXT[] DEFAULT '{}';

-- Create index for faster permission lookups
CREATE INDEX IF NOT EXISTS idx_users_permissions ON users USING GIN (permissions);

SELECT 'Permissions column added successfully' as status;
