package authsvc

import (
	"encoding/json"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type RegisterResult struct {
	User        UserDTO `json:"user"`
	AccessToken string  `json:"access_token"`
	TokenType   string  `json:"token_type"`
}


func Register(name, email string) (RegisterResult, int, string) {
	userID := uuid.NewString()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return RegisterResult{}, 500, "failed to sign token"
	}

	return RegisterResult{
		User: UserDTO{
			ID:    userID,
			Name:  name,
			Email: email,
		},
		AccessToken: signed,
		TokenType:   "Bearer",
	}, 201, ""
}

// Marshal helper (optional) used by handlers
func Marshal(v any) ([]byte, error) { return json.Marshal(v) }
