package users

import (
	"context"
	"database/sql"
	"errors"
)

// Repo is the MySQL implementation of the Users repository.
type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) GetByEmail(ctx context.Context, email string) (User, error) {
	const q = `
		SELECT id, uuid, name, email, password_hash
		FROM users
		WHERE email = ?
		LIMIT 1
	`

	var u User
	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.UUID, &u.Name, &u.Email, &u.PasswordHash,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *Repo) Create(ctx context.Context, p CreateParams) (User, error) {
	const q = `
		INSERT INTO users (uuid, name, email, password_hash)
		VALUES (?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(ctx, q, p.UUID, p.Name, p.Email, p.PasswordHash)
	if err != nil {
		return User{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}

	return User{
		ID:           id,
		UUID:         p.UUID,
		Name:         p.Name,
		Email:        p.Email,
		PasswordHash: p.PasswordHash,
	}, nil
}

func (r *Repo) GetByUUID(ctx context.Context, userUUID string) (User, error) {
	const q = `
		SELECT id, uuid, name, email, password_hash
		FROM users
		WHERE uuid = ?
		LIMIT 1
	`

	var u User
	err := r.db.QueryRowContext(ctx, q, userUUID).Scan(
		&u.ID, &u.UUID, &u.Name, &u.Email, &u.PasswordHash,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}
