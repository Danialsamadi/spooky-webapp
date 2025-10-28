package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
	"webapp/database"
	"webapp/middleware"
	"webapp/models"
	"webapp/utils"
)

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/signup.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		Username := r.PostFormValue("username")
		Email := r.PostFormValue("email")
		Password := r.PostFormValue("password")
		InvitationCode := r.PostFormValue("invitation_code")
		clientIP := getClientIP(r)

		// Validate invitation code
		var invitation models.InvitationCode
		err := database.DB.QueryRow(`
			SELECT id, code, created_by, used_by, is_used, expires_at 
			FROM invitation_codes 
			WHERE code = ? AND is_used = FALSE 
			AND (expires_at IS NULL OR expires_at > NOW())
		`, InvitationCode).Scan(
			&invitation.ID, &invitation.Code, &invitation.CreatedBy,
			&invitation.UsedBy, &invitation.IsUsed, &invitation.ExpiresAt,
		)

		if err != nil {
			utils.LogError(fmt.Sprintf("Invalid invitation code %s for user %s: %v", InvitationCode, Username, err))
			http.Error(w, "Invalid or expired invitation code", http.StatusBadRequest)
			return
		}

		hashPassword, err := middleware.HashPassword(Password)
		if err != nil {
			utils.LogError(fmt.Sprintf("Password hashing failed for user %s: %v", Username, err))
			http.Error(w, "Error Creating Password", http.StatusInternalServerError)
			return
		}

		// Start transaction to ensure atomicity
		tx, err := database.DB.Begin()
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to start transaction for user %s: %v", Username, err))
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Insert user with invitation code
		result, err := tx.Exec(`
			INSERT INTO users (username, email, password, invitation_code, invited_by) 
			VALUES (?, ?, ?, ?, ?)
		`, Username, Email, hashPassword, InvitationCode, invitation.CreatedBy)

		if err != nil {
			utils.LogSignup(Username, Email, clientIP, false)
			utils.LogError(fmt.Sprintf("Signup failed for user %s: %v", Username, err))
			http.Error(w, "Username or Email already Exist", http.StatusBadRequest)
			return
		}

		// Get the new user ID
		userID, err := result.LastInsertId()
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to get user ID for %s: %v", Username, err))
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Mark invitation code as used
		_, err = tx.Exec(`
			UPDATE invitation_codes 
			SET is_used = TRUE, used_by = ?, used_at = NOW() 
			WHERE id = ?
		`, userID, invitation.ID)

		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to mark invitation code as used for user %s: %v", Username, err))
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			utils.LogError(fmt.Sprintf("Failed to commit transaction for user %s: %v", Username, err))
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Log successful signup
		utils.LogSignup(Username, Email, clientIP, true)
		utils.LogInfo(fmt.Sprintf("New user registered with invitation code: %s (%s)", Username, Email))

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		Username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		clientIP := getClientIP(r)

		// Debug: Print login attempt
		fmt.Printf("Login attempt - Username: '%s', Password: '%s'\n", Username, password)
		utils.LogInfo(fmt.Sprintf("Login attempt from IP %s for user: %s", clientIP, Username))

		var user models.User
		err := database.DB.QueryRow("SELECT id, username, password FROM users WHERE username = ?", Username).Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			// note for myself: Show custom spooky user not found page
			utils.LogLogin(Username, clientIP, false)
			utils.LogError(fmt.Sprintf("User not found: %s from IP %s", Username, clientIP))
			tmpl := template.Must(template.ParseFiles("templates/wrong_password.html"))
			tmpl.Execute(w, nil)
			return
		}

		// Debug: Print stored hash and input password
		fmt.Printf("Login - Stored hash: %s\n", user.Password)
		fmt.Printf("Login - Input password: '%s'\n", password)

		// note for myself: CheckPassword(password, hash) - password first, then hash
		// WHY: bcrypt.CompareHashAndPassword expects (hash, password) but our function signature is (password, hash)
		// if you swap them, bcrypt will fail with "hashedSecret too short" error because it tries to use plaintext as hash
		passwordMatch := middleware.CheckPassword(password, user.Password)
		fmt.Printf("Login - Password match result: %t\n", passwordMatch)

		if !passwordMatch {
			// note for myself: Show custom spooky wrong password page instead of generic error
			utils.LogLogin(Username, clientIP, false)
			utils.LogError(fmt.Sprintf("Wrong password for user: %s from IP %s", Username, clientIP))
			tmpl := template.Must(template.ParseFiles("templates/wrong_password.html"))
			tmpl.Execute(w, nil)
			return
		}

		// Session creation after successful password verification
		token := middleware.GenerateToken()
		middleware.Sessions[token] = &middleware.Session{
			Token:    token,
			UserID:   fmt.Sprintf("%d", user.ID),
			ExpireAt: time.Now().Add(time.Hour * 24),
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Expires:  time.Now().Add(time.Hour * 24),
			HttpOnly: true,
		})

		// Log successful login
		utils.LogLogin(Username, clientIP, true)
		utils.LogInfo(fmt.Sprintf("User %s logged in successfully from IP %s", Username, clientIP))

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)

	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Get username from session before deleting
		if session, exists := middleware.Sessions[cookie.Value]; exists {
			utils.LogLogout(session.UserID, clientIP)
			utils.LogInfo(fmt.Sprintf("User %s logged out from IP %s", session.UserID, clientIP))
		}
		delete(middleware.Sessions, cookie.Value)
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})

	// note for myself: Show spooky goodbye page instead of direct redirect
	// This gives users a nice farewell experience with ghost animations
	tmpl := template.Must(template.ParseFiles("templates/logout.html"))
	tmpl.Execute(w, nil)
}

// Helper function to get client IP address
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check for X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := strings.Split(r.RemoteAddr, ":")[0]
	return ip
}

// GenerateInvitationCode creates a new invitation code
func GenerateInvitationCode(createdBy int, expiresAt *time.Time) (string, error) {
	// Generate a random code (you can customize this logic)
	code := fmt.Sprintf("INV-%d-%s", time.Now().Unix(), generateRandomString(8))

	_, err := database.DB.Exec(`
		INSERT INTO invitation_codes (code, created_by, expires_at) 
		VALUES (?, ?, ?)
	`, code, createdBy, expiresAt)

	return code, err
}

// generateRandomString creates a random string of specified length
func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ListInvitationCodes returns all invitation codes for a user
func ListInvitationCodes(createdBy int) ([]models.InvitationCode, error) {
	rows, err := database.DB.Query(`
		SELECT id, code, created_by, used_by, is_used, expires_at, created_at, used_at
		FROM invitation_codes 
		WHERE created_by = ?
		ORDER BY created_at DESC
	`, createdBy)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []models.InvitationCode
	for rows.Next() {
		var code models.InvitationCode
		err := rows.Scan(
			&code.ID, &code.Code, &code.CreatedBy, &code.UsedBy,
			&code.IsUsed, &code.ExpiresAt, &code.CreatedAt, &code.UsedAt,
		)
		if err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}

	return codes, nil
}

func InvitationCodeHandler(w http.ResponseWriter, r *http.Request) {
	// This would require authentication middleware to get the current user
	// For now, assuming you have a way to get the current user ID

	if r.Method == "POST" {
		// Generate new invitation code
		// You'll need to get the current user ID from session
		userID := 1 // Replace with actual user ID from session

		// Set expiration to 30 days from now
		expiresAt := time.Now().AddDate(0, 0, 30)

		code, err := GenerateInvitationCode(userID, &expiresAt)
		if err != nil {
			http.Error(w, "Failed to generate invitation code", http.StatusInternalServerError)
			return
		}

		// Return the code (you might want to show it in a template)
		fmt.Fprintf(w, "Generated invitation code: %s", code)
		return
	}

	// Show invitation code management page
	// You can create a template for this
}
