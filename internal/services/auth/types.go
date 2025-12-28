package auth

import "errors"

var (
	ErrEmailTaken = errors.New("auth: email already registered")
	ErrInvalidCredentials  = errors.New("auth: invalid credentials")
)

// RegisterInput is what the service needs (already validated by the request layer).
type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type UserDTO struct {
	ID    string `json:"id"` // public id (uuid)
	Name  string `json:"name"`
	Email string `json:"email"`
}

type RegisterOutput struct {
	User        UserDTO `json:"user"`
	AccessToken string  `json:"access_token"`
	TokenType   string  `json:"token_type"`
}
type LoginInput struct {
	Email    string
	Password string
}
