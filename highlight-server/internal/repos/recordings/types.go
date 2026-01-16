package recordings

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("recordings: not found")

type Recording struct {
	ID              int64
	UUID            string
	UserID          int64
	Title           string
	OriginalName    string
	StoragePath     string
	DurationSeconds int
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateParams struct {
	UUID            string
	UserID          int64
	Title           string
	OriginalName    string
	StoragePath     string
	DurationSeconds int
	Status          string
}
