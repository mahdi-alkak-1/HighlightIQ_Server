package users

import "errors"

var ErrNotFound = errors.New("users: not found")

// User represents a row in the users table.
type User struct {
	ID           int64
	UUID         string
	Name         string
	Email        string
	PasswordHash string
}

type CreateParams struct {
	UUID         string
	Name         string
	Email        string
	PasswordHash string
}
