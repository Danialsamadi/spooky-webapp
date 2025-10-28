package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
	"webapp/database"
	"webapp/middleware"
	"webapp/models"
	"webapp/utils"
)

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user info
	session, _ := middleware.GetSession(r)

	// Get invitation codes created by this admin
	codes, err := ListInvitationCodes(session.UserID)
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

	data := struct {
		UserCount   int
		PostCount   int
		InviteCodes []models.InvitationCode
	}{
		UserCount:   userCount,
		PostCount:   postCount,
		InviteCodes: codes,
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

	// Set expiration to 30 days from now
	expiresAt := time.Now().AddDate(0, 0, 30)

	code, err := GenerateInvitationCode(session.UserID, &expiresAt)
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
