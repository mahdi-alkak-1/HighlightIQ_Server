package clipcandidates

import (
	"context"
	"database/sql"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateMany(ctx context.Context, items []CreateParams) (int64, error) {
	if len(items) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	const q = `
		INSERT INTO clip_candidates (recording_id, start_ms, end_ms, score, detected_signals, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var inserted int64
	for _, it := range items {
		if it.Status == "" {
			it.Status = "new"
		}
		_, err := tx.ExecContext(ctx, q,
			it.RecordingID, it.StartMS, it.EndMS, it.Score, it.DetectedJSON, it.Status,
		)
		if err != nil {
			return 0, err
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return inserted, nil
}

func (r *Repo) ListByRecordingID(ctx context.Context, recordingID int64) ([]Candidate, error) {
	const q = `
		SELECT id, recording_id, start_ms, end_ms, score, detected_signals, status, created_at, updated_at
		FROM clip_candidates
		WHERE recording_id = ?
		ORDER BY score DESC, start_ms ASC
	`

	rows, err := r.db.QueryContext(ctx, q, recordingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Candidate
	for rows.Next() {
		var c Candidate
		var detected sql.NullString
		if err := rows.Scan(
			&c.ID, &c.RecordingID, &c.StartMS, &c.EndMS, &c.Score, &detected, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if detected.Valid {
			s := detected.String
			c.DetectedJSON = &s
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repo) UpdateStatus(ctx context.Context, id int64, status string) error {
	const q = `
		UPDATE clip_candidates
		SET status = ?
		WHERE id = ?
		LIMIT 1
	`
	_, err := r.db.ExecContext(ctx, q, status, id)
	return err
}

func (r *Repo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM clip_candidates WHERE id = ? LIMIT 1`
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
