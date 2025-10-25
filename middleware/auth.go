package middleware

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	Token    string
	UserID   string
	ExpireAt time.Time
}

var Sessions = make(map[string]*Session)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err

}

func CheckPassword(password, hash string) bool {
	fmt.Printf("CheckPassword - Comparing password: '%s' with hash: %s\n", password, hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	fmt.Printf("CheckPassword - bcrypt error: %v\n", err)
	result := err == nil
	fmt.Printf("CheckPassword - Final result: %t\n", result)
	return result
}

func GenerateToken() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func GetSession(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return Session{}, false
	}

	session, exists := Sessions[cookie.Value]
	if !exists || time.Now().After(session.ExpireAt) {
		return Session{}, false

	}
	return *session, true

}
