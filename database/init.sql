-- Database initialization script for Docker
-- This script will be executed when the MySQL container starts

-- Create database if not exists (already created by MYSQL_DATABASE)
-- USE blogdb;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create posts table
CREATE TABLE IF NOT EXISTS posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    author_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for better performance
-- Note: MySQL doesn't support CREATE INDEX IF NOT EXISTS
-- We'll create indexes without the IF NOT EXISTS clause
-- If they already exist, MySQL will show a warning but continue
CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_created ON posts(created_at);

-- Insert some sample data (optional)
-- INSERT INTO users (username, email, password) VALUES 
-- ('admin', 'admin@example.com', '$2a$14$example.hash.here'),
-- ('testuser', 'test@example.com', '$2a$14$example.hash.here');

-- INSERT INTO posts (title, content, author_id) VALUES
-- ('Welcome to the Haunted Blog', 'This is a spooky blog built with Go and Docker!', 1),
-- ('Ghost Mode Activated', 'The spirits are pleased with this dark theme.', 1);

-- Create invitation_codes table
CREATE TABLE IF NOT EXISTS invitation_codes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    created_by INT NOT NULL,
    used_by INT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMP NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (used_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Add invitation_code field to users table
ALTER TABLE users ADD COLUMN invitation_code VARCHAR(50) NULL;
ALTER TABLE users ADD COLUMN invited_by INT NULL;
ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE SET NULL;

-- Create first admin user
INSERT INTO users (username, email, password, invitation_code, invited_by, is_admin) 
VALUES ('admin', 'admin@example.com', '$2a$14$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'ADMIN-CREATED', NULL, TRUE);


