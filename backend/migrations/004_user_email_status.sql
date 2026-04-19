-- 004_user_email_status.sql: Add Email and Status fields to users table
ALTER TABLE users ADD COLUMN email VARCHAR(100) DEFAULT '';
ALTER TABLE users ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'enabled';
