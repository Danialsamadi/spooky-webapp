package main

import (
	"flag"
	"fmt"
	"os"
	"webapp/database"
	"webapp/middleware"
	"webapp/utils"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// Initialize database
	database.InitDB()
	defer database.DB.Close()

	// Initialize logging
	utils.InitLogger()

	command := os.Args[1]

	switch command {
	case "createsuperuser":
		createSuperUser()
	case "cleanusers":
		cleanUsers()
	case "listusers":
		listUsers()
	case "generatecode":
		generateInviteCode()
	case "listcodes":
		listInviteCodes()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func createSuperUser() {
	var username, email, password string

	flag.StringVar(&username, "username", "admin", "Admin username")
	flag.StringVar(&email, "email", "admin@example.com", "Admin email")
	flag.StringVar(&password, "password", "", "Admin password")
	flag.Parse()

	if password == "" {
		fmt.Print("Enter password: ")
		fmt.Scanln(&password)
		if password == "" {
			fmt.Println("Password cannot be empty")
			return
		}
	}

	// Check if user already exists
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if count > 0 {
		fmt.Printf("User '%s' already exists\n", username)
		return
	}

	// Hash password
	hashPassword, err := middleware.HashPassword(password)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		return
	}

	// Create admin user
	_, err = database.DB.Exec(`
		INSERT INTO users (username, email, password, invitation_code, invited_by, is_admin) 
		VALUES (?, ?, ?, ?, ?, ?)
	`, username, email, hashPassword, "ADMIN-CREATED", nil, true)

	if err != nil {
		fmt.Printf("Error creating admin user: %v\n", err)
		return
	}

	fmt.Printf("✅ Superuser '%s' created successfully!\n", username)
	utils.LogInfo(fmt.Sprintf("Superuser created via CLI: %s", username))
}

func cleanUsers() {
	// Confirm action
	fmt.Print("Are you sure you want to delete all non-admin users? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" {
		fmt.Println("Operation cancelled")
		return
	}

	result, err := database.DB.Exec("DELETE FROM users WHERE is_admin = FALSE")
	if err != nil {
		fmt.Printf("Error cleaning users: %v\n", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ Cleaned %d users from database\n", rowsAffected)
	utils.LogInfo(fmt.Sprintf("Cleaned %d users via CLI", rowsAffected))
}

func listUsers() {
	rows, err := database.DB.Query(`
		SELECT id, username, email, is_admin, created_at 
		FROM users 
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("Error querying users: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n📋 Users List:")
	fmt.Println("ID | Username | Email | Role | Created")
	fmt.Println("---|----------|-------|------|--------")

	for rows.Next() {
		var id int
		var username, email string
		var isAdmin bool
		var createdAt string

		err := rows.Scan(&id, &username, &email, &isAdmin, &createdAt)
		if err != nil {
			continue
		}

		role := "User"
		if isAdmin {
			role = "Admin"
		}

		fmt.Printf("%-2d | %-8s | %-20s | %-5s | %s\n",
			id, username, email, role, createdAt)
	}
	fmt.Println()
}

func generateInviteCode() {
	var createdBy int
	flag.IntVar(&createdBy, "created-by", 1, "User ID who creates the code")
	flag.Parse()

	// Check if user exists and is admin
	var isAdmin bool
	err := database.DB.QueryRow("SELECT is_admin FROM users WHERE id = ?", createdBy).Scan(&isAdmin)
	if err != nil {
		fmt.Printf("Error: User with ID %d not found\n", createdBy)
		return
	}

	if !isAdmin {
		fmt.Printf("Error: User with ID %d is not an admin\n", createdBy)
		return
	}

	// Generate code
	code := fmt.Sprintf("INV-%d-%s", utils.GetCurrentTimestamp(), utils.GenerateRandomString(8))

	_, err = database.DB.Exec(`
		INSERT INTO invitation_codes (code, created_by, expires_at) 
		VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 30 DAY))
	`, code, createdBy)

	if err != nil {
		fmt.Printf("Error generating invite code: %v\n", err)
		return
	}

	fmt.Printf("✅ Invitation code generated: %s\n", code)
	fmt.Printf("   Expires in 30 days\n")
	utils.LogInfo(fmt.Sprintf("Invite code generated via CLI: %s", code))
}

func listInviteCodes() {
	rows, err := database.DB.Query(`
		SELECT id, code, created_by, is_used, expires_at, created_at, used_at
		FROM invitation_codes 
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("Error querying invite codes: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n🎫 Invitation Codes:")
	fmt.Println("ID | Code | Created By | Used | Expires | Created | Used At")
	fmt.Println("---|------|------------|------|---------|---------|--------")

	for rows.Next() {
		var id, createdBy int
		var code string
		var isUsed bool
		var expiresAt, createdAt, usedAt *string

		err := rows.Scan(&id, &code, &createdBy, &isUsed, &expiresAt, &createdAt, &usedAt)
		if err != nil {
			continue
		}

		used := "No"
		if isUsed {
			used = "Yes"
		}

		expires := "Never"
		if expiresAt != nil {
			expires = *expiresAt
		}

		usedAtStr := "-"
		if usedAt != nil {
			usedAtStr = *usedAt
		}

		fmt.Printf("%-2d | %-15s | %-10d | %-4s | %-8s | %s | %s\n",
			id, code, createdBy, used, expires, *createdAt, usedAtStr)
	}
	fmt.Println()
}

func printUsage() {
	fmt.Println("🚀 Webapp Management CLI")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println()
	fmt.Println("  createsuperuser")
	fmt.Println("    Creates a new admin user")
	fmt.Println("    Options:")
	fmt.Println("      -username string  Admin username (default: admin)")
	fmt.Println("      -email string     Admin email (default: admin@example.com)")
	fmt.Println("      -password string  Admin password (will prompt if not provided)")
	fmt.Println()
	fmt.Println("  cleanusers")
	fmt.Println("    Removes all non-admin users from database")
	fmt.Println()
	fmt.Println("  listusers")
	fmt.Println("    Lists all users in the database")
	fmt.Println()
	fmt.Println("  generatecode")
	fmt.Println("    Generates a new invitation code")
	fmt.Println("    Options:")
	fmt.Println("      -created-by int   User ID who creates the code (default: 1)")
	fmt.Println()
	fmt.Println("  listcodes")
	fmt.Println("    Lists all invitation codes")
	fmt.Println()
	fmt.Println("  help")
	fmt.Println("    Shows this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run manage.go createsuperuser")
	fmt.Println("  go run manage.go createsuperuser -username myadmin -email admin@mydomain.com")
	fmt.Println("  go run manage.go cleanusers")
	fmt.Println("  go run manage.go listusers")
	fmt.Println("  go run manage.go generatecode -created-by 1")
	fmt.Println("  go run manage.go listcodes")
}
