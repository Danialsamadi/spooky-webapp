package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() {
	// Get database configuration from environment variables
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "dandan1234")
	dbName := getEnv("DB_NAME", "blogdb")

	// For Docker, we don't need to create database as it's already created
	// Just connect directly to the database
	dsn := dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true"

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error connecting to database", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Error pinging database", err)
	}

	log.Println("Database connected successfully")
	createTables()
}

// Helper function to get environment variables with default values
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createTables() {
	usersTable := `CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	postsTable := `CREATE TABLE IF NOT EXISTS posts (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(200) NOT NULL,
		content TEXT NOT NULL,
		author_id INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE)`

	_, err := DB.Exec(usersTable)
	if err != nil {
		log.Fatal("Error creating users table", err)
	}
	log.Println("Users table created")

	_, err = DB.Exec(postsTable)
	if err != nil {
		log.Fatal("Error creating posts table", err)
	}
	log.Println("Posts table created")
	createProfileTables()
	createIndexes()
}

func createProfileTables() {
	// Add profile fields to existing users table
	alterUsers := []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_image VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS location VARCHAR(100)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS website VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
	}

	for _, query := range alterUsers {
		_, err := DB.Exec(query)
		if err != nil {
			// Ignore if column already exists
			log.Printf("Info: %v", err)
		}
	}

	// Create profile_images table
	profileImagesTable := `CREATE TABLE IF NOT EXISTS profile_images (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		filename VARCHAR(255) NOT NULL,
		original_name VARCHAR(255) NOT NULL,
		file_path VARCHAR(500) NOT NULL,
		file_size INT NOT NULL,
		mime_type VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`

	_, err := DB.Exec(profileImagesTable)
	if err != nil {
		log.Fatal("Error creating profile_images table:", err)
	}

	log.Println("Profile tables created successfully!")
}

func createIndexes() {
	// Check and create indexes - MySQL doesn't support IF NOT EXISTS for indexes
	// So we try to create and ignore if it already exists
	indexes := map[string]string{
		"idx_posts_author":  "CREATE INDEX idx_posts_author ON posts(author_id)",
		"idx_posts_created": "CREATE INDEX idx_posts_created ON posts(created_at)",
	}

	for indexName, indexQuery := range indexes {
		// Check if index exists
		var exists int
		err := DB.QueryRow("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = 'blogdb' AND table_name = 'posts' AND index_name = ?", indexName).Scan(&exists)

		if err != nil {
			log.Printf("Warning: Could not check index %s: %v", indexName, err)
			continue
		}

		// Create index only if it doesn't exist
		if exists == 0 {
			_, err = DB.Exec(indexQuery)
			if err != nil {
				log.Printf("Warning: Could not create index %s: %v", indexName, err)
			} else {
				log.Printf("Index %s created successfully!", indexName)

				// note for myself: Using log.Printf instead of log.Fatal for index creation
				// if index fails to create, app can still work (just slower queries)
				// log.Fatal would crash the entire app, but we want graceful degradation

			}
		}
	}
	log.Println("Indexes created")
}
