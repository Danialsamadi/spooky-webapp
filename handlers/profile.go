package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"webapp/database"
	"webapp/middleware"
	"webapp/models"
	"webapp/utils"
)

const (
	MaxUploadSize = 5 << 20 // 5 MB
	UploadPath    = "./uploads/profiles"
)

// Initialize upload directory
func init() {
	if err := os.MkdirAll(UploadPath, os.ModePerm); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}
}

// View user profile
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, loggedIn := middleware.GetSession(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	var userID int
	fmt.Sscanf(session.UserID, "%d", &userID)

	var bio, profileImage, location, website sql.NullString
	err := database.DB.QueryRow(`
		SELECT id, username, email, bio, profile_image, location, website, created_at, updated_at 
		FROM users WHERE id = ?`, userID).
		Scan(&user.ID, &user.Username, &user.Email, &bio, &profileImage,
			&location, &website, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		utils.LogError("User not found: " + err.Error())
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Convert sql.NullString to string
	user.Bio = bio.String
	user.ProfileImage = profileImage.String
	user.Location = location.String
	user.Website = website.String

	// Count user's posts
	var postCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE author_id = ?", user.ID).Scan(&postCount)

	// Log profile view
	clientIP := getClientIP(r)
	utils.LogInfo(fmt.Sprintf("Profile viewed - User: %s, IP: %s", user.Username, clientIP))

	// Create template with custom functions
	funcMap := template.FuncMap{
		"substr": func(s string, start, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"upper": strings.ToUpper,
	}

	tmpl := template.Must(template.New("profile.html").Funcs(funcMap).ParseFiles("templates/profile.html"))
	data := map[string]interface{}{
		"User":      user,
		"PostCount": postCount,
		"LoggedIn":  loggedIn,
	}
	tmpl.Execute(w, data)
}

// Edit profile form
func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, loggedIn := middleware.GetSession(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		var user models.User
		var userID int
		fmt.Sscanf(session.UserID, "%d", &userID)

		var bio, profileImage, location, website sql.NullString
		err := database.DB.QueryRow(`
			SELECT id, username, email, bio, profile_image, location, website 
			FROM users WHERE id = ?`, userID).
			Scan(&user.ID, &user.Username, &user.Email, &bio,
				&profileImage, &location, &website)

		if err != nil {
			utils.LogError("User not found for edit: " + err.Error())
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Convert sql.NullString to string
		user.Bio = bio.String
		user.ProfileImage = profileImage.String
		user.Location = location.String
		user.Website = website.String

		// Create template with custom functions
		funcMap := template.FuncMap{
			"substr": func(s string, start, length int) string {
				if start >= len(s) {
					return ""
				}
				end := start + length
				if end > len(s) {
					end = len(s)
				}
				return s[start:end]
			},
			"upper": strings.ToUpper,
		}

		tmpl := template.Must(template.New("edit_profile.html").Funcs(funcMap).ParseFiles("templates/edit_profile.html"))
		tmpl.Execute(w, user)
		return
	}

	if r.Method == "POST" {
		// Limit upload size
		r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

		if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
			utils.LogError("File too large: " + err.Error())
			http.Error(w, "File too large (max 5MB)", http.StatusBadRequest)
			return
		}

		// Get form values
		username := r.FormValue("username")
		bio := r.FormValue("bio")
		location := r.FormValue("location")
		website := r.FormValue("website")

		// Handle profile image upload
		var profileImagePath string
		file, handler, err := r.FormFile("profile_image")
		if err == nil {
			defer file.Close()

			// Validate file type
			if !isValidImageType(handler.Header.Get("Content-Type")) {
				utils.LogError("Invalid file type uploaded")
				http.Error(w, "Invalid file type. Only JPG, PNG, and GIF allowed", http.StatusBadRequest)
				return
			}

			// Generate unique filename
			filename := generateFilename(handler.Filename)
			filepath := filepath.Join(UploadPath, filename)

			// Save file
			dst, err := os.Create(filepath)
			if err != nil {
				utils.LogError("Failed to create file: " + err.Error())
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			written, err := io.Copy(dst, file)
			if err != nil {
				utils.LogError("Failed to copy file: " + err.Error())
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}

			// Save to database
			var userID int
			fmt.Sscanf(session.UserID, "%d", &userID)
			_, err = database.DB.Exec(`
				INSERT INTO profile_images (user_id, filename, original_name, file_path, file_size, mime_type) 
				VALUES (?, ?, ?, ?, ?, ?)`,
				userID, filename, handler.Filename, filepath, written, handler.Header.Get("Content-Type"))

			if err != nil {
				utils.LogError("Failed to save image record: " + err.Error())
			}

			profileImagePath = "/uploads/profiles/" + filename
		}

		// Update user profile
		var userID int
		fmt.Sscanf(session.UserID, "%d", &userID)
		query := `UPDATE users SET username = ?, bio = ?, location = ?, website = ?`
		args := []interface{}{username, bio, location, website}

		if profileImagePath != "" {
			query += `, profile_image = ?`
			args = append(args, profileImagePath)
		}

		query += ` WHERE id = ?`
		args = append(args, userID)

		_, err = database.DB.Exec(query, args...)
		if err != nil {
			utils.LogError("Failed to update profile: " + err.Error())
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}

		// Log profile update
		clientIP := getClientIP(r)
		utils.LogInfo(fmt.Sprintf("Profile updated - User: %s, IP: %s", username, clientIP))

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}

// Helper: Check if file is valid image type
func isValidImageType(mimeType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
	}

	for _, valid := range validTypes {
		if mimeType == valid {
			return true
		}
	}
	return false
}

// Helper: Generate unique filename
func generateFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes) + ext
}

// Delete profile image
func DeleteProfileImageHandler(w http.ResponseWriter, r *http.Request) {
	session, loggedIn := middleware.GetSession(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get current profile image
	var userID int
	fmt.Sscanf(session.UserID, "%d", &userID)
	var profileImage sql.NullString
	err := database.DB.QueryRow("SELECT profile_image FROM users WHERE id = ?", userID).
		Scan(&profileImage)

	if err != nil || !profileImage.Valid || profileImage.String == "" {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Delete file from filesystem
	filename := strings.TrimPrefix(profileImage.String, "/uploads/profiles/")
	filePath := filepath.Join(UploadPath, filename)
	os.Remove(filePath)

	// Update database
	_, err = database.DB.Exec("UPDATE users SET profile_image = NULL WHERE id = ?", userID)
	if err != nil {
		utils.LogError("Failed to delete image: " + err.Error())
	}

	// Log image deletion
	clientIP := getClientIP(r)
	utils.LogInfo(fmt.Sprintf("Profile image deleted - User ID: %s, IP: %s", session.UserID, clientIP))

	http.Redirect(w, r, "/edit-profile", http.StatusSeeOther)
}

// View another user's profile (public)
func PublicProfileHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	session, loggedIn := middleware.GetSession(r)

	var user models.User
	var bio, profileImage, location, website sql.NullString
	err := database.DB.QueryRow(`
		SELECT id, username, email, bio, profile_image, location, website, created_at 
		FROM users WHERE username = ?`, username).
		Scan(&user.ID, &user.Username, &user.Email, &bio,
			&profileImage, &location, &website, &user.CreatedAt)

	if err != nil {
		utils.LogError("Public profile not found: " + err.Error())
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Convert sql.NullString to string
	user.Bio = bio.String
	user.ProfileImage = profileImage.String
	user.Location = location.String
	user.Website = website.String

	// Get user's posts
	rows, err := database.DB.Query(`
		SELECT id, title, content, created_at 
		FROM posts 
		WHERE author_id = ? 
		ORDER BY created_at DESC 
		LIMIT 10`, user.ID)

	if err != nil {
		utils.LogError("Failed to get user posts: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt)
		posts = append(posts, post)
	}

	// Log public profile view
	clientIP := getClientIP(r)
	utils.LogInfo(fmt.Sprintf("Public profile viewed - User: %s, Viewer IP: %s", username, clientIP))

	// Create template with custom functions
	funcMap := template.FuncMap{
		"substr": func(s string, start, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"upper": strings.ToUpper,
	}

	tmpl := template.Must(template.New("public_profile.html").Funcs(funcMap).ParseFiles("templates/public_profile.html"))
	data := map[string]interface{}{
		"User":         user,
		"Posts":        posts,
		"LoggedIn":     loggedIn,
		"IsOwnProfile": loggedIn && session.UserID == fmt.Sprintf("%d", user.ID),
	}
	tmpl.Execute(w, data)
}
