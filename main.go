package main

import (
	"fmt"
	"log"
	"net/http"
	"webapp/database"
	"webapp/handlers"
	"webapp/utils"
	"webapp/middleware"
)

func main() {
	// Initialize logging system
	utils.InitLogger()
	utils.LogInfo("Application starting...")

	// Test logging system
	utils.TestLogging()

	database.InitDB()
	defer database.DB.Close()

	// Create first admin if none exists
	handlers.CreateFirstAdmin()

	// Static files handler for ghost.gif and uploaded files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads/"))))

	// Existing routes
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/signup", handlers.SignupHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/post/create", handlers.CreatePostHandler)
	http.HandleFunc("/post/edit", handlers.EditPostHandler)
	http.HandleFunc("/post/delete", handlers.DeletePostHandler)

	// Profile routes
	http.HandleFunc("/profile", handlers.ProfileHandler)
	http.HandleFunc("/edit-profile", handlers.EditProfileHandler)
	http.HandleFunc("/profile/delete-image", handlers.DeleteProfileImageHandler)
	http.HandleFunc("/user", handlers.PublicProfileHandler)

	// Add this import
	// Add these routes after your existing routes
	http.HandleFunc("/admin", middleware.RequireAdmin(handlers.AdminDashboardHandler))
	http.HandleFunc("/admin/generate-code", middleware.RequireAdmin(handlers.GenerateInviteCodeHandler))
	http.HandleFunc("/admin/users", middleware.RequireAdmin(handlers.AdminUsersHandler))

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
