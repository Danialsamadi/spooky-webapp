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

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	session, loggedIn := middleware.GetSession(r)
	clientIP := getClientIP(r)

	// Log page view
	if loggedIn {
		utils.LogInfo(fmt.Sprintf("User %s viewed home page from IP %s", session.UserID, clientIP))
	} else {
		utils.LogInfo(fmt.Sprintf("Anonymous user viewed home page from IP %s", clientIP))
	}

	rows, err := database.DB.Query(`SELECT p.id, p.title, p.content, p.author_id, u.username, p.created_at, p.updated_at
		FROM posts p JOIN users u ON p.author_id = u.id
		ORDER BY p.created_at DESC`)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var createdAtStr, updatedAtStr string

		// note for myself: MySQL returns timestamps as []uint8, need to scan as string first
		// then parse using Go's reference time format "2006-01-02 15:04:05"
		err := rows.Scan(&post.Id, &post.Title, &post.Content, &post.AuthorId, &post.Username, &createdAtStr, &updatedAtStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse timestamp strings to time.Time
		post.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		post.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		posts = append(posts, post)
	}
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	// note for myself: Fixed syntax error - map[string]interface{} needs {} after interface
	// PROBLEM: "unexpected literal 'Posts', expected ~ term or type"
	// CAUSE: Missing {} after interface in map declaration
	data := map[string]interface{}{
		"Posts":    posts,
		"LoggedIn": loggedIn,
		"UserID":   session.UserID,
	}
	tmpl.Execute(w, data)
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	session, loggedIn := middleware.GetSession(r)
	clientIP := getClientIP(r)

	if !loggedIn {
		utils.LogError(fmt.Sprintf("Unauthorized post creation attempt from IP %s", clientIP))
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		utils.LogInfo(fmt.Sprintf("User %s accessed create post page from IP %s", session.UserID, clientIP))
		tmpl := template.Must(template.ParseFiles("templates/create_post.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")

		_, err := database.DB.Exec("INSERT INTO posts (title, content, author_id) VALUES (?, ?, ?)", title, content, session.UserID)
		if err != nil {
			utils.LogError(fmt.Sprintf("Post creation failed for user %s: %v", session.UserID, err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Log successful post creation
		utils.LogInfo(fmt.Sprintf("User %s created new post: '%s' from IP %s", session.UserID, title, clientIP))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	session, loggedIn := middleware.GetSession(r)
	clientIP := getClientIP(r)

	if !loggedIn {
		utils.LogError(fmt.Sprintf("Unauthorized edit attempt from IP %s", clientIP))
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postID := r.URL.Query().Get("id")
	var post models.Post

	err := database.DB.QueryRow("SELECT id, title, content, author_id FROM posts WHERE id = ?", postID).Scan(&post.Id, &post.Title, &post.Content, &post.AuthorId)
	if err != nil {
		utils.LogError(fmt.Sprintf("Post not found for edit: ID %s by user %s from IP %s", postID, session.UserID, clientIP))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if fmt.Sprintf("%d", post.AuthorId) != session.UserID {
		utils.LogError(fmt.Sprintf("Unauthorized edit attempt: User %s tried to edit post %s (owned by %d) from IP %s", session.UserID, postID, post.AuthorId, clientIP))
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		utils.LogInfo(fmt.Sprintf("User %s accessed edit page for post '%s' (ID: %s) from IP %s", session.UserID, post.Title, postID, clientIP))
		tmpl := template.Must(template.ParseFiles("templates/edit_post.html"))
		tmpl.Execute(w, post)
		return
	}

	if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")

		_, err := database.DB.Exec("UPDATE posts SET title = ?, content = ? WHERE id = ?", title, content, postID)
		if err != nil {
			utils.LogError(fmt.Sprintf("Post update failed for user %s, post %s: %v", session.UserID, postID, err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Log successful post update
		utils.LogInfo(fmt.Sprintf("User %s updated post '%s' (ID: %s) from IP %s", session.UserID, title, postID, clientIP))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	session, loggedIn := middleware.GetSession(r)
	clientIP := getClientIP(r)

	if !loggedIn {
		utils.LogError(fmt.Sprintf("Unauthorized delete attempt from IP %s", clientIP))
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postID := r.URL.Query().Get("id")
	var authorID string
	var postTitle string

	// Get post title for logging
	err := database.DB.QueryRow("SELECT author_id, title FROM posts WHERE id = ?", postID).Scan(&authorID, &postTitle)
	if err != nil {
		utils.LogError(fmt.Sprintf("Post not found for deletion: ID %s by user %s from IP %s", postID, session.UserID, clientIP))
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	if authorID != session.UserID {
		utils.LogError(fmt.Sprintf("Unauthorized delete attempt: User %s tried to delete post %s (owned by %s) from IP %s", session.UserID, postID, authorID, clientIP))
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	_, err = database.DB.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		utils.LogError(fmt.Sprintf("Post deletion failed for user %s, post %s: %v", session.UserID, postID, err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log successful deletion
	utils.LogInfo(fmt.Sprintf("User %s deleted post '%s' (ID: %s) from IP %s", session.UserID, postTitle, postID, clientIP))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
