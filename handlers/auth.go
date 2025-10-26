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
		clientIP := getClientIP(r)

		hashPassword, err := middleware.HashPassword(Password)
		if err != nil {
			utils.LogError(fmt.Sprintf("Password hashing failed for user %s: %v", Username, err))
			http.Error(w, "Error Creating Password", http.StatusInternalServerError)
			return
		}

		// Debug: Print the original password and hash
		fmt.Printf("Signup - Original password: '%s'\n", Password)
		fmt.Printf("Signup - Generated hash: %s\n", hashPassword)

		_, err = database.DB.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
			Username, Email, hashPassword)
		if err != nil {
			utils.LogSignup(Username, Email, clientIP, false)
			utils.LogError(fmt.Sprintf("Signup failed for user %s: %v", Username, err))
			http.Error(w, "Username or Email already Exist", http.StatusBadRequest)
			return
		}

		// Log successful signup
		utils.LogSignup(Username, Email, clientIP, true)
		utils.LogInfo(fmt.Sprintf("New user registered: %s (%s)", Username, Email))

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
