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
