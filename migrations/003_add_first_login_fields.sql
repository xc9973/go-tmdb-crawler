-- Migration: Add First Login Fields to Sessions Table
-- Version: 003
-- Created: 2026-01-12
-- Description: Add fields to track first login and device fingerprint for session management

-- Add is_first_login field to track if this is the user's first login
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS is_first_login BOOLEAN DEFAULT true;

-- Add device_fingerprint field to identify the device/browser
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(255);

-- Add index for device_fingerprint for faster lookups
CREATE INDEX IF NOT EXISTS idx_sessions_device_fingerprint ON sessions(device_fingerprint);

-- Add index for is_first_login for queries filtering first-time sessions
CREATE INDEX IF NOT EXISTS idx_sessions_is_first_login ON sessions(is_first_login);

-- Add comment to document the new fields
COMMENT ON COLUMN sessions.is_first_login IS 'Indicates if this is the user''s first login session';
COMMENT ON COLUMN sessions.device_fingerprint IS 'Unique identifier for the device/browser (user agent + IP hash)';
