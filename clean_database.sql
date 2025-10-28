-- Clean Database Script
-- This will remove all users and related data

-- Delete all users (posts will be deleted due to CASCADE)
DELETE FROM users;

-- Reset auto-increment counters
ALTER TABLE users AUTO_INCREMENT = 1;
ALTER TABLE invitation_codes AUTO_INCREMENT = 1;

-- Clean up invitation codes
DELETE FROM invitation_codes;

-- Recreate the admin user
INSERT INTO users (username, email, password, invitation_code, invited_by, is_admin) 
VALUES ('admin', 'admin@example.com', '$2a$14$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'ADMIN-CREATED', NULL, TRUE);

-- Verify the cleanup
SELECT COUNT(*) as user_count FROM users;
SELECT COUNT(*) as post_count FROM posts;
SELECT COUNT(*) as invite_code_count FROM invitation_codes;
