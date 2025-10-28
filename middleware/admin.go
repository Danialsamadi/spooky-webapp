package middleware

import (
	"net/http"
	"webapp/database"
	"webapp/models"
)

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, exists := GetSession(r)
		if !exists {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check if user is admin
		var user models.User
		err := database.DB.QueryRow("SELECT is_admin FROM users WHERE id = ?", session.UserID).Scan(&user.IsAdmin)
		if err != nil || !user.IsAdmin {
			http.Error(w, "Access denied. Admin privileges required.", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
