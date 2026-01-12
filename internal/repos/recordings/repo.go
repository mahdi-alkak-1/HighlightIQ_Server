package recordings

import (
	"context"
	"database/sql"
	"errors"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, p CreateParams) (Recording, error) {
	const q = `
		INSERT INTO recordings (uuid, user_id, title, original_filename, storage_path, duration_seconds, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(ctx, q,
		p.UUID, p.UserID, p.Title, p.OriginalName, p.StoragePath, p.DurationSeconds, p.Status,
	)
	if err != nil {
		return Recording{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Recording{}, err
	}

	// Return the created row (simple: fetch by uuid+user)
	return r.GetByUUIDForUser(ctx, p.UserID, p.UUID, id)
}

func (r *Repo) GetByUUIDForUser(ctx context.Context, userID int64, recUUID string, fallbackID int64) (Recording, error) {
	// If fallbackID is 0, ignore it; otherwise allow either match (helps Create return)
	const q = `
		SELECT id, uuid, user_id, title, original_filename, storage_path, duration_seconds, status, created_at, updated_at
		FROM recordings
		WHERE user_id = ?
		  AND (uuid = ? OR (? <> 0 AND id = ?))
		LIMIT 1
	`

	var rec Recording
	err := r.db.QueryRowContext(ctx, q, userID, recUUID, fallbackID, fallbackID).Scan(
		&rec.ID,
		&rec.UUID,
		&rec.UserID,
		&rec.Title,
		&rec.OriginalName,
		&rec.StoragePath,
		&rec.DurationSeconds,
		&rec.Status,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Recording{}, ErrNotFound
	}
	if err != nil {
		return Recording{}, err
	}
	return rec, nil
}

func (r *Repo) ListByUser(ctx context.Context, userID int64) ([]Recording, error) {
	const q = `
		SELECT id, uuid, user_id, title, original_filename, storage_path, duration_seconds, status, created_at, updated_at
		FROM recordings
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Recording
	for rows.Next() {
		var rec Recording
		if err := rows.Scan(
			&rec.ID,
			&rec.UUID,
			&rec.UserID,
			&rec.Title,
			&rec.OriginalName,
			&rec.StoragePath,
			&rec.DurationSeconds,
			&rec.Status,
			&rec.CreatedAt,
			&rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) UpdateTitleByUUIDForUser(ctx context.Context, userID int64, recUUID string, title string) error {
	const q = `
		UPDATE recordings
		SET title = ?
		WHERE user_id = ? AND uuid = ?
	`

	res, err := r.db.ExecContext(ctx, q, title, userID, recUUID)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) DeleteByUUIDForUser(ctx context.Context, userID int64, recUUID string) (string, error) {
	// We return storage_path so service can delete the file from disk
	const sel = `
		SELECT storage_path
		FROM recordings
		WHERE user_id = ? AND uuid = ?
		LIMIT 1
	`
	var path string
	err := r.db.QueryRowContext(ctx, sel, userID, recUUID).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	const del = `
		DELETE FROM recordings
		WHERE user_id = ? AND uuid = ?
	`
	_, err = r.db.ExecContext(ctx, del, userID, recUUID)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (r *Repo) GetStoragePathByIDForUser(ctx context.Context, userID int64, recordingID int64) (string, error) {
	const q = `
		SELECT storage_path
		FROM recordings
		WHERE user_id = ? AND id = ?
		LIMIT 1
	`
	var path string
	err := r.db.QueryRowContext(ctx, q, userID, recordingID).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return path, err
}
