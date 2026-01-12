package clips

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var ErrNotFound = errors.New("clips: not found")

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, p CreateParams) (Clip, error) {
	durationSeconds := 0
	if p.EndMS > p.StartMS {
		durationSeconds = (p.EndMS - p.StartMS) / 1000
	}

	if p.Status == "" {
		p.Status = "draft"
	}

	const q = `
		INSERT INTO clips (user_id, recording_id, candidate_id, title, caption, start_ms, end_ms, duration_seconds, status, export_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(ctx, q,
		p.UserID, p.RecordingID, p.CandidateID, p.Title, p.Caption, p.StartMS, p.EndMS, durationSeconds, p.Status, p.ExportPath,
	)
	if err != nil {
		return Clip{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Clip{}, err
	}

	return r.GetByIDForUser(ctx, p.UserID, id)
}

func (r *Repo) GetByIDForUser(ctx context.Context, userID int64, id int64) (Clip, error) {
	const q = `
		SELECT id, user_id, recording_id, candidate_id, title, caption, start_ms, end_ms, duration_seconds, status, export_path, created_at, updated_at
		FROM clips
		WHERE user_id = ? AND id = ?
		LIMIT 1
	`

	var c Clip
	var cand sql.NullInt64
	var caption sql.NullString
	var export sql.NullString

	err := r.db.QueryRowContext(ctx, q, userID, id).Scan(
		&c.ID, &c.UserID, &c.RecordingID, &cand, &c.Title, &caption, &c.StartMS, &c.EndMS, &c.DurationSeconds,
		&c.Status, &export, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Clip{}, ErrNotFound
	}
	if err != nil {
		return Clip{}, err
	}

	if cand.Valid {
		v := cand.Int64
		c.CandidateID = &v
	}
	if caption.Valid {
		v := caption.String
		c.Caption = &v
	}
	if export.Valid {
		v := export.String
		c.ExportPath = &v
	}

	return c, nil
}

func (r *Repo) ListByUser(ctx context.Context, userID int64, recordingID *int64) ([]Clip, error) {
	var sb strings.Builder
	sb.WriteString(`
		SELECT id, user_id, recording_id, candidate_id, title, caption, start_ms, end_ms, duration_seconds, status, export_path, created_at, updated_at
		FROM clips
		WHERE user_id = ?
	`)
	args := []interface{}{userID}

	if recordingID != nil {
		sb.WriteString(" AND recording_id = ?")
		args = append(args, *recordingID)
	}

	sb.WriteString(" ORDER BY created_at DESC")

	rows, err := r.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Clip
	for rows.Next() {
		var c Clip
		var cand sql.NullInt64
		var caption sql.NullString
		var export sql.NullString

		if err := rows.Scan(
			&c.ID, &c.UserID, &c.RecordingID, &cand, &c.Title, &caption, &c.StartMS, &c.EndMS, &c.DurationSeconds,
			&c.Status, &export, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if cand.Valid {
			v := cand.Int64
			c.CandidateID = &v
		}
		if caption.Valid {
			v := caption.String
			c.Caption = &v
		}
		if export.Valid {
			v := export.String
			c.ExportPath = &v
		}

		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repo) UpdateByIDForUser(ctx context.Context, userID int64, id int64, p UpdateParams) (Clip, error) {
	// Build a dynamic UPDATE; only set provided fields.
	setParts := make([]string, 0, 6)
	args := make([]interface{}, 0, 8)

	if p.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *p.Title)
	}
	if p.Caption != nil {
		setParts = append(setParts, "caption = ?")
		args = append(args, *p.Caption)
	}
	if p.StartMS != nil {
		setParts = append(setParts, "start_ms = ?")
		args = append(args, *p.StartMS)
	}
	if p.EndMS != nil {
		setParts = append(setParts, "end_ms = ?")
		args = append(args, *p.EndMS)
	}
	if p.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *p.Status)
	}
	if p.ExportPath != nil {
		setParts = append(setParts, "export_path = ?")
		args = append(args, *p.ExportPath)
	}

	if len(setParts) == 0 {
		// Nothing to update; return current clip.
		return r.GetByIDForUser(ctx, userID, id)
	}

	q := `
		UPDATE clips
		SET ` + strings.Join(setParts, ", ") + `
		WHERE user_id = ? AND id = ?
		LIMIT 1
	`

	args = append(args, userID, id)

	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return Clip{}, err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return Clip{}, err
	}
	if aff == 0 {
		return Clip{}, ErrNotFound
	}

	// Recompute duration if start/end changed
	if p.StartMS != nil || p.EndMS != nil {
		cur, err := r.GetByIDForUser(ctx, userID, id)
		if err != nil {
			return Clip{}, err
		}
		durationSeconds := 0
		if cur.EndMS > cur.StartMS {
			durationSeconds = (cur.EndMS - cur.StartMS) / 1000
		}
		_, _ = r.db.ExecContext(ctx, `
			UPDATE clips SET duration_seconds = ? WHERE user_id = ? AND id = ? LIMIT 1
		`, durationSeconds, userID, id)
	}

	return r.GetByIDForUser(ctx, userID, id)
}

func (r *Repo) DeleteByIDForUser(ctx context.Context, userID int64, id int64) error {
	const q = `DELETE FROM clips WHERE user_id = ? AND id = ? LIMIT 1`
	res, err := r.db.ExecContext(ctx, q, userID, id)
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
