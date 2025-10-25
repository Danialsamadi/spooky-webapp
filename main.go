package main

import (
	"fmt"
	"log"
	"net/http"
	"webapp/database"
	"webapp/handlers"
	"webapp/utils"
)

func main() {
	// Initialize logging system
	utils.InitLogger()
	utils.LogInfo("Application starting...")

	database.InitDB()
	defer database.DB.Close()

	// Static files handler for ghost.gif
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/signup", handlers.SignupHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/post/create", handlers.CreatePostHandler)
	http.HandleFunc("/post/edit", handlers.EditPostHandler)
	http.HandleFunc("/post/delete", handlers.DeletePostHandler)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
