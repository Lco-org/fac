-- Add missing 'details' column to notifications table
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS details JSONB DEFAULT '{}'::jsonb;
