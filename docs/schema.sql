-- GIN Authentication Microservice Database Schema
-- PostgreSQL 12+
-- Created: May 30, 2026

-- ========================================
-- Users Table
-- ========================================
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Authentication Fields
    username VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(320) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    token_version INT DEFAULT 0,
    
    -- Email Verification
    is_email_verified BOOLEAN DEFAULT FALSE,
    verification_code VARCHAR(255),
    verification_expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Phone & Profile
    phone VARCHAR(20) UNIQUE,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female')),
    birthday DATE,
    
    -- Two-Factor Authentication (2FA)
    tfa_code VARCHAR(255),
    tfa_verifed BOOLEAN DEFAULT FALSE,
    login_codes TEXT,
    login_codes_set BOOLEAN DEFAULT FALSE,
    
    -- User Status & Role
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    role VARCHAR(20) CHECK (role IN ('admin', 'user', 'moderator', 'staff')) DEFAULT 'user',
    
    -- Profile Media
    profile_image_url VARCHAR(500),
    cover_image_url TEXT,
    bio TEXT,
    
    -- OAuth Support
    provider VARCHAR(50)
);

-- Create indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- ========================================
-- Device Record Table (Device Tracking)
-- ========================================
CREATE TABLE IF NOT EXISTS device_record (
    id SERIAL PRIMARY KEY,
    userid INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Geographic Information
    city VARCHAR(255),
    region VARCHAR(255),
    country VARCHAR(255),
    locale VARCHAR(10),
    lat FLOAT8,
    lon FLOAT8,
    zipcode VARCHAR(50),
    
    -- Device Information
    browser TEXT,
    last_login TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index for fast user device lookup
CREATE INDEX IF NOT EXISTS idx_device_record_userid ON device_record(userid);
CREATE INDEX IF NOT EXISTS idx_device_record_country ON device_record(country);
CREATE INDEX IF NOT EXISTS idx_device_record_last_login ON device_record(last_login);

-- ========================================
-- Old Password Table (Password History)
-- ========================================
CREATE TABLE IF NOT EXISTS old_password (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index for fast user password history lookup
CREATE INDEX IF NOT EXISTS idx_old_password_user_id ON old_password(user_id);
CREATE INDEX IF NOT EXISTS idx_old_password_created_at ON old_password(created_at);

-- ========================================
-- Comments for Documentation
-- ========================================

COMMENT ON TABLE users IS 'Core user accounts table with authentication and profile data';
COMMENT ON TABLE device_record IS 'Device tracking with geolocation information';
COMMENT ON TABLE old_password IS 'Password history to prevent password reuse';

COMMENT ON COLUMN users.id IS 'Primary key - unique user identifier';
COMMENT ON COLUMN users.username IS 'Unique username generated from name + UUID suffix';
COMMENT ON COLUMN users.email IS 'Unique email address - verified before use';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hash with pepper - never store plain text';
COMMENT ON COLUMN users.token_version IS 'Version number for JWT invalidation on password change';
COMMENT ON COLUMN users.is_email_verified IS 'Email verification status';
COMMENT ON COLUMN users.verification_code IS 'Temporary 6-digit code for email verification';
COMMENT ON COLUMN users.verification_expires_at IS 'Expiration time for verification code (15 minutes)';
COMMENT ON COLUMN users.tfa_verifed IS 'Two-factor authentication enabled status';
COMMENT ON COLUMN users.tfa_code IS 'TOTP secret for authenticator apps';
COMMENT ON COLUMN users.login_codes IS 'Comma-separated list of unused backup login codes';
COMMENT ON COLUMN users.login_codes_set IS 'Flag indicating backup codes have been generated';
COMMENT ON COLUMN users.is_active IS 'User active status - inactive requires re-login';
COMMENT ON COLUMN users.is_verified IS 'Full account verification status (may require admin approval)';
COMMENT ON COLUMN users.role IS 'User role for access control (admin/moderator/user/staff)';

COMMENT ON COLUMN device_record.userid IS 'Foreign key to users.id';
COMMENT ON COLUMN device_record.city IS 'City from geolocation API';
COMMENT ON COLUMN device_record.region IS 'Region/state from geolocation API';
COMMENT ON COLUMN device_record.country IS 'Country from geolocation API - used for access control';
COMMENT ON COLUMN device_record.lat IS 'Latitude coordinate';
COMMENT ON COLUMN device_record.lon IS 'Longitude coordinate';
COMMENT ON COLUMN device_record.browser IS 'Browser/User-Agent string';
COMMENT ON COLUMN device_record.last_login IS 'Last login timestamp for this device';

COMMENT ON COLUMN old_password.user_id IS 'Foreign key to users.id';
COMMENT ON COLUMN old_password.password IS 'Hashed old password - prevents recent password reuse';
COMMENT ON COLUMN old_password.created_at IS 'When the password was changed';

-- ========================================
-- Views for Analytics (Optional)
-- ========================================

-- Active users summary
CREATE OR REPLACE VIEW v_active_users AS
SELECT 
    COUNT(*) as total_users,
    COUNT(CASE WHEN is_active THEN 1 END) as active_users,
    COUNT(CASE WHEN is_email_verified THEN 1 END) as verified_users,
    COUNT(CASE WHEN tfa_verifed THEN 1 END) as tfa_enabled_users
FROM users
WHERE deleted_at IS NULL;

-- User activity by country
CREATE OR REPLACE VIEW v_user_activity_by_country AS
SELECT 
    dr.country,
    COUNT(DISTINCT dr.userid) as unique_users,
    COUNT(*) as total_logins,
    MAX(dr.last_login) as last_activity
FROM device_record dr
JOIN users u ON dr.userid = u.id
WHERE u.deleted_at IS NULL
GROUP BY dr.country
ORDER BY unique_users DESC;

-- Recent logins
CREATE OR REPLACE VIEW v_recent_logins AS
SELECT 
    u.username,
    u.email,
    dr.city,
    dr.country,
    dr.browser,
    dr.last_login
FROM device_record dr
JOIN users u ON dr.userid = u.id
WHERE u.deleted_at IS NULL
ORDER BY dr.last_login DESC
LIMIT 100;

-- ========================================
-- Functions for Common Operations
-- ========================================

-- Get user login count in last N days
CREATE OR REPLACE FUNCTION user_login_count(user_id INT, days INT DEFAULT 30)
RETURNS INT AS $$
SELECT COUNT(*)
FROM device_record
WHERE userid = user_id 
AND last_login > NOW() - INTERVAL '1 day' * days;
$$ LANGUAGE SQL STABLE;

-- Check if user has pending password history
CREATE OR REPLACE FUNCTION user_recent_passwords(user_id INT, limit_count INT DEFAULT 5)
RETURNS TABLE(password_hash VARCHAR, created_at TIMESTAMP WITH TIME ZONE) AS $$
SELECT password, created_at
FROM old_password
WHERE user_id = user_id
ORDER BY created_at DESC
LIMIT limit_count;
$$ LANGUAGE SQL STABLE;

-- ========================================
-- Sample Data (Optional - Remove in Production)
-- ========================================

-- This section is commented out for production.
-- Uncomment only for testing/development.

/*
INSERT INTO users (username, name, email, password_hash, is_email_verified, is_verified, is_active)
VALUES (
    'testuser_a1b2c3d4',
    'Test User',
    'test@gmail.com',
    '$2a$12$KIX...', -- bcrypt hash of 'TestPass123!@#'
    true,
    true,
    true
);

INSERT INTO device_record (userid, city, region, country, locale, lat, lon, zipcode, browser)
VALUES (
    1,
    'San Francisco',
    'California',
    'United States',
    'en-US',
    37.7749,
    -122.4194,
    '94105',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
);
*/

-- ========================================
-- Permissions and Security (PostgreSQL)
-- ========================================

-- Create application role (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
        CREATE ROLE app_user WITH PASSWORD 'change_me_in_production' LOGIN;
    END IF;
END $$;

-- Grant necessary permissions
-- GRANT CONNECT ON DATABASE book TO app_user;
-- GRANT USAGE ON SCHEMA public TO app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;

-- ========================================
-- Backup and Maintenance
-- ========================================

-- Backup this schema:
-- pg_dump -U postgres -d book --schema-only > schema_backup.sql

-- Restore this schema:
-- psql -U postgres -d book < schema_backup.sql

-- Full database backup:
-- pg_dump -U postgres -d book > full_backup.sql

-- Restore full backup:
-- psql -U postgres -d book < full_backup.sql

-- ========================================
-- End of Schema
-- ========================================
-- Version: 1.0.0
-- Last Updated: May 30, 2026
