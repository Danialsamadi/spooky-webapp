package middleware

import "testing"

func TestHashPassword(t *testing.T) {
	password := "password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)

	}

	if !CheckPassword(password, hash) {
		t.Errorf("Password check failed")
	}

	if CheckPassword("wrong", hash) {
		t.Error("Wrong password was accepted ")
	}
	
}

func TestGenerateToken(t *testing.T) {
	token1 := GenerateToken()
	token2 := GenerateToken()

	if token1 == token2 {
		t.Errorf("Tokens are the same")
	}
}
