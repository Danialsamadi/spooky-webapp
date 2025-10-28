package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"webapp/database"
	"webapp/middleware"
	"webapp/models"
	"webapp/utils"
)

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user info
	session, _ := middleware.GetSession(r)
	userID, _ := strconv.Atoi(session.UserID)

	// Get invitation codes created by this admin
	codes, err := ListInvitationCodes(userID)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to get invitation codes: %v", err))
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get all users count
	var userCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)

	// Get all posts count
	var postCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM posts").Scan(&postCount)

	// Check for success message
	success := r.URL.Query().Get("success")
	code := r.URL.Query().Get("code")
	count := r.URL.Query().Get("count")

	data := struct {
		UserCount   int
		PostCount   int
		InviteCodes []models.InvitationCode
		Success     string
		Code        string
		Count       string
	}{
		UserCount:   userCount,
		PostCount:   postCount,
		InviteCodes: codes,
		Success:     success,
		Code:        code,
		Count:       count,
	}

	tmpl := template.Must(template.ParseFiles("templates/admin_dashboard.html"))
	tmpl.Execute(w, data)
}

func GenerateInviteCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, _ := middleware.GetSession(r)
	userID, _ := strconv.Atoi(session.UserID)

	// Set expiration to 30 days from now
	expiresAt := time.Now().AddDate(0, 0, 30)

	code, err := GenerateInvitationCode(userID, &expiresAt)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to generate invitation code: %v", err))
		http.Error(w, "Failed to generate invitation code", http.StatusInternalServerError)
		return
	}

	utils.LogInfo(fmt.Sprintf("Admin %s generated invitation code: %s", session.UserID, code))

	// Redirect back to admin dashboard with success message
	http.Redirect(w, r, "/admin?success=code_generated&code="+code, http.StatusSeeOther)
}

func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get all users
	rows, err := database.DB.Query(`
		SELECT id, username, email, is_admin, created_at, invited_by
		FROM users 
		ORDER BY created_at DESC
	`)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to get users: %v", err))
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin, &user.CreatedAt, &user.InvitedBy)
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to scan user: %v", err))
			continue
		}
		users = append(users, user)
	}

	tmpl := template.Must(template.ParseFiles("templates/admin_users.html"))
	tmpl.Execute(w, users)
}

// CreateFirstAdmin creates the first admin user if none exists
func CreateFirstAdmin() {
	// Check if any admin exists
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = TRUE").Scan(&count)

	if count == 0 {
		// No admin exists, create one
		password := "password" // Default password
		hashPassword, err := middleware.HashPassword(password)
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to hash admin password: %v", err))
			return
		}

		_, err = database.DB.Exec(`
			INSERT INTO users (username, email, password, invitation_code, invited_by, is_admin) 
			VALUES (?, ?, ?, ?, ?, ?)
		`, "admin", "admin@example.com", hashPassword, "ADMIN-CREATED", nil, true)

		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to create admin user: %v", err))
		} else {
			utils.LogInfo("First admin user created: admin / password")
		}
	}
}

// CleanAllUsersHandler removes all users except admins
func CleanAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delete all non-admin users
	result, err := database.DB.Exec("DELETE FROM users WHERE is_admin = FALSE")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to clean users: %v", err))
		http.Error(w, "Failed to clean users", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	utils.LogInfo(fmt.Sprintf("Cleaned %d users from database", rowsAffected))

	// Redirect back to admin dashboard
	http.Redirect(w, r, "/admin?success=users_cleaned&count="+fmt.Sprintf("%d", rowsAffected), http.StatusSeeOther)
}
